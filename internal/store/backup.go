package store

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/jackc/pgx/v5"
)

// Backup and restore (DIRECTION 2026-10-15, "The plan after the statutory
// floor", item 1): the system of record is the append-only tables — the
// definitions in every version with their statuses, the bundles, the files
// under their hashes, and the event log with its ids, causes, actors and
// doors. That is the whole backup. The object table is not in it: it is
// the projection replay rebuilds (invariants 4 and 7), and a restore ends
// with a replay. The format is JSON lines — a header, then one line per
// row as Postgres itself spells the row — so a backup reads without Rhea
// and restores through Rhea alone.

// backupTables, in the order a restore can insert them: definitions and
// bundles first, the files, then the events (a cause precedes its effect
// in id order).
var backupTables = []struct{ name, order string }{
	{"object_type", "name, version"},
	{"view_def", "view_id, version"},
	{"rule", "rule_id, version"},
	{"activity", "name, version"},
	{"bundle", "bundle_id, version"},
	{"blob", "hash"},
	{"event", "event_id"},
}

// BackupCounts says how many rows of each table a backup or restore carried.
type BackupCounts map[string]int

type backupLine struct {
	Backup    int             `json:"rhea_backup,omitempty"`
	CreatedAt string          `json:"created_at,omitempty"`
	Table     string          `json:"t,omitempty"`
	Row       json.RawMessage `json:"r,omitempty"`
}

// Backup writes the system of record to w.
func (s *Store) Backup(ctx context.Context, w io.Writer) (BackupCounts, error) {
	bw := bufio.NewWriter(w)
	enc := json.NewEncoder(bw)
	if err := enc.Encode(backupLine{Backup: 1, CreatedAt: time.Now().UTC().Format(time.RFC3339)}); err != nil {
		return nil, err
	}
	// One snapshot: every table read at one point in the log, so no event
	// in the backup names a rule version the backup lacks.
	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	counts := BackupCounts{}
	for _, t := range backupTables {
		rows, err := tx.Query(ctx, "SELECT row_to_json(t) FROM "+t.name+" t ORDER BY "+t.order)
		if err != nil {
			return nil, fmt.Errorf("backup %s: %w", t.name, err)
		}
		for rows.Next() {
			var raw json.RawMessage
			if err := rows.Scan(&raw); err != nil {
				rows.Close()
				return nil, err
			}
			if err := enc.Encode(backupLine{Table: t.name, Row: raw}); err != nil {
				rows.Close()
				return nil, err
			}
			counts[t.name]++
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}
	return counts, bw.Flush()
}

// InitSchema runs the schema without seeding the builtin verbs: the empty
// database a restore lands in, whose doors arrive with the backup.
func (s *Store) InitSchema(ctx context.Context) error {
	_, err := s.Pool.Exec(ctx, schemaSQL)
	return err
}

// Restore reads a backup into an empty system of record — the schema
// present, no event and no definition yet — in one transaction, keeping
// every id and timestamp, and moves the event sequence past the last id
// so the log continues where it stopped. The caller replays afterwards.
func (s *Store) Restore(ctx context.Context, r io.Reader) (BackupCounts, error) {
	var events, defs int
	if err := s.Pool.QueryRow(ctx, "SELECT count(*) FROM event").Scan(&events); err != nil {
		return nil, fmt.Errorf("restore wants the schema in place (rhea init without the seed, or restore into a fresh database): %w", err)
	}
	if err := s.Pool.QueryRow(ctx, "SELECT (SELECT count(*) FROM rule) + (SELECT count(*) FROM object_type) + (SELECT count(*) FROM activity)").Scan(&defs); err != nil {
		return nil, err
	}
	if events > 0 || defs > 0 {
		return nil, fmt.Errorf("restore wants an empty system of record; this one holds %d events and %d definitions", events, defs)
	}
	allowed := map[string]bool{}
	for _, t := range backupTables {
		allowed[t.name] = true
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	counts := BackupCounts{}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 64<<20) // a blob row carries its bytes
	first := true
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var l backupLine
		if err := json.Unmarshal(line, &l); err != nil {
			return nil, fmt.Errorf("backup line %d: %w", countLines(counts)+1, err)
		}
		if first {
			first = false
			if l.Backup != 1 {
				return nil, fmt.Errorf("not a Rhea backup: the first line should say rhea_backup 1")
			}
			continue
		}
		if !allowed[l.Table] {
			return nil, fmt.Errorf("backup names a table a restore does not take: %q", l.Table)
		}
		sql := "INSERT INTO " + l.Table + " SELECT * FROM json_populate_record(NULL::" + l.Table + ", $1)"
		if l.Table == "event" {
			sql = "INSERT INTO event OVERRIDING SYSTEM VALUE SELECT * FROM json_populate_record(NULL::event, $1)"
		}
		if _, err := tx.Exec(ctx, sql, string(l.Row)); err != nil {
			return nil, fmt.Errorf("restore %s row %d: %w", l.Table, counts[l.Table]+1, err)
		}
		counts[l.Table]++
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if counts["event"] > 0 {
		if _, err := tx.Exec(ctx, "SELECT setval(pg_get_serial_sequence('event', 'event_id'), (SELECT max(event_id) FROM event))"); err != nil {
			return nil, fmt.Errorf("restore: the event sequence: %w", err)
		}
	}
	return counts, tx.Commit(ctx)
}

func countLines(c BackupCounts) int {
	n := 0
	for _, v := range c {
		n += v
	}
	return n
}
