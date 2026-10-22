package store

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"rhea/internal/core"
)

// Backup and restore (DIRECTION 2026-10-15, "The plan after the statutory
// floor", item 1): the system of record is the append-only tables — the
// definitions in every version with their statuses, the bundles, the files
// under their hashes, and the event log with its ids, causes, actors and
// doors. That is the whole backup. The object table is not in it: it is
// the projection replay rebuilds (invariants 4 and 7), and a restore ends
// with a replay. The format is JSON lines — a header, then one line per
// row as Postgres itself spells the row — so a backup reads without Rhea
// and restores through Rhea alone. Since format 2 a backup ends with what it
// carried — each table's rows and the last event's id — so a file cut short
// or edited is told from a whole one (review 2026-10-21, G2).

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
	End       *backupEnd      `json:"rhea_backup_end,omitempty"`
}

// backupFormat is the format a backup is written in: 2 ends with backupEnd.
const backupFormat = 2

// backupEnd is a backup's last line: what it carried.
type backupEnd struct {
	Counts     BackupCounts `json:"counts"`
	MaxEventID int64        `json:"max_event_id"`
}

// Backup writes the system of record to w.
func (s *Store) Backup(ctx context.Context, w io.Writer) (BackupCounts, error) {
	bw := bufio.NewWriter(w)
	enc := json.NewEncoder(bw)
	if err := enc.Encode(backupLine{Backup: backupFormat, CreatedAt: time.Now().UTC().Format(time.RFC3339)}); err != nil {
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
	end := backupEnd{Counts: counts}
	if err := tx.QueryRow(ctx, "SELECT COALESCE(max(event_id), 0) FROM event").Scan(&end.MaxEventID); err != nil {
		return nil, err
	}
	if err := enc.Encode(backupLine{End: &end}); err != nil {
		return nil, err
	}
	return counts, bw.Flush()
}

// VerifyBackup reads a backup without a database and says whether it is
// whole: its end line is there and names exactly what it carried. A
// format-1 backup has no end line and cannot be verified.
func VerifyBackup(r io.Reader) (BackupCounts, error) {
	counts := BackupCounts{}
	var maxID int64
	var end *backupEnd
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 64<<20)
	first := true
	for sc.Scan() {
		if len(sc.Bytes()) == 0 {
			continue
		}
		var l backupLine
		if err := json.Unmarshal(sc.Bytes(), &l); err != nil {
			return nil, fmt.Errorf("backup line %d: %w", countLines(counts)+1, err)
		}
		switch {
		case first:
			first = false
			if l.Backup != backupFormat {
				return nil, fmt.Errorf("a format-%d backup carries no end line to verify against (format %d does)", l.Backup, backupFormat)
			}
		case end != nil:
			return nil, fmt.Errorf("a line after the backup's end")
		case l.End != nil:
			end = l.End
		default:
			counts[l.Table]++
			if l.Table == "event" {
				var ev struct {
					ID int64 `json:"event_id"`
				}
				if err := json.Unmarshal(l.Row, &ev); err != nil {
					return nil, fmt.Errorf("backup event row %d: %w", counts["event"], err)
				}
				maxID = max(maxID, ev.ID)
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return counts, checkEnd(end, counts, maxID)
}

// checkEnd compares what a backup says it carried with what it carried.
func checkEnd(end *backupEnd, counts BackupCounts, maxID int64) error {
	if end == nil {
		return fmt.Errorf("the backup stops before its end line: it was cut short")
	}
	for t, n := range end.Counts {
		if counts[t] != n {
			return fmt.Errorf("the backup says it carries %d %s rows and carries %d", n, t, counts[t])
		}
	}
	for t, n := range counts {
		if end.Counts[t] != n {
			return fmt.Errorf("the backup carries %d %s rows it does not name", n, t)
		}
	}
	if end.MaxEventID != maxID {
		return fmt.Errorf("the backup says its last event is %d and carries %d", end.MaxEventID, maxID)
	}
	return nil
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
	return s.restore(ctx, r, nil)
}

// Stamped says what a restore for one company did with the facts it found
// carrying no company: how many took a company and how many stayed the
// installation's.
type Stamped struct{ Company, Installation int }

// RestoreFor restores a backup written before the company key — every fact
// acting for no company, so under isolation one company's books are the
// installation's rows (DECISIONS 2026-10-21) — for the one company whose
// books they are: each fact with no company comes back as if it had been
// written acting for that company. A door's fact takes the key, unless the
// door is installation-wide in any version the backup holds (a person, a
// sign-in); a fact through no door is the pack's master data, which the
// loader writes as the installation's, and a fact signed "kernel" is the
// kernel's own (the seed's approvals, a refusal) — both stay blank; a derived
// fact takes its cause's company, as the executor stamps a chain from its
// root. A fact that names a company keeps it. Nothing is rewritten: the log
// is appended once, into an empty system of record, with the keys its facts
// would have carried.
func (s *Store) RestoreFor(ctx context.Context, r io.Reader, company string) (BackupCounts, Stamped, error) {
	if company == "" || strings.ContainsAny(company, ", \t\r\n") {
		return nil, Stamped{}, fmt.Errorf("a restore stamps one company's key, not %q", company)
	}
	st := &stamper{company: company, installation: map[string]bool{}, of: map[int64]string{}}
	counts, err := s.restore(ctx, r, st)
	return counts, st.n, err
}

// stamper decides, row by row in the backup's order (the doors before the
// facts, a cause before its effect), whose a fact with no company is.
type stamper struct {
	company      string
	installation map[string]bool  // doors whose facts are the installation's
	of           map[int64]string // each restored fact's company, for its consequences
	n            Stamped
}

func (st *stamper) door(row json.RawMessage) error {
	var a struct {
		Name string `json:"name"`
		Spec struct {
			Installation bool `json:"installation"`
		} `json:"spec"`
	}
	if err := json.Unmarshal(row, &a); err != nil {
		return err
	}
	if a.Spec.Installation {
		st.installation[a.Name] = true
	}
	return nil
}

func (st *stamper) fact(row json.RawMessage) (json.RawMessage, error) {
	var ev struct {
		ID       int64   `json:"event_id"`
		Kind     string  `json:"kind"`
		Cause    *int64  `json:"cause_event_id"`
		Actor    *string `json:"actor"`
		Activity *string `json:"activity_name"`
		Company  *string `json:"company"`
	}
	if err := json.Unmarshal(row, &ev); err != nil {
		return nil, err
	}
	if ev.Company != nil && *ev.Company != "" {
		st.of[ev.ID] = *ev.Company
		return row, nil
	}
	company := ""
	switch {
	case ev.Kind == core.KindDerived:
		if ev.Cause == nil {
			return nil, fmt.Errorf("derived event %d names no cause", ev.ID)
		}
		c, ok := st.of[*ev.Cause]
		if !ok {
			return nil, fmt.Errorf("derived event %d: its cause %d is not before it in the backup", ev.ID, *ev.Cause)
		}
		company = c
	case ev.Activity == nil || *ev.Activity == "": // the pack's master data
	case ev.Actor != nil && *ev.Actor == "kernel": // the kernel's own
	case st.installation[*ev.Activity]: // an installation-wide door
	default:
		company = st.company
	}
	st.of[ev.ID] = company
	if company == "" {
		st.n.Installation++
		return row, nil
	}
	st.n.Company++
	var f map[string]json.RawMessage
	if err := json.Unmarshal(row, &f); err != nil {
		return nil, err
	}
	f["company"], _ = json.Marshal(company)
	return json.Marshal(f)
}

func (s *Store) restore(ctx context.Context, r io.Reader, st *stamper) (BackupCounts, error) {
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
	columns := map[string]map[string]bool{} // each table's columns in this database
	var format int
	var maxID int64
	var end *backupEnd
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
			if format = l.Backup; format != 1 && format != backupFormat {
				return nil, fmt.Errorf("not a Rhea backup this Rhea reads: the first line says rhea_backup %d", l.Backup)
			}
			continue
		}
		if end != nil {
			return nil, fmt.Errorf("a line after the backup's end")
		}
		if l.End != nil {
			end = l.End
			continue
		}
		if !allowed[l.Table] {
			return nil, fmt.Errorf("backup names a table a restore does not take: %q", l.Table)
		}
		// Every field of a row is a column here: a row this database cannot
		// hold whole is refused, never restored without what it lacks.
		if columns[l.Table] == nil {
			if columns[l.Table], err = tableColumns(ctx, tx, l.Table); err != nil {
				return nil, err
			}
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(l.Row, &fields); err != nil {
			return nil, fmt.Errorf("restore %s row %d: %w", l.Table, counts[l.Table]+1, err)
		}
		for f := range fields {
			if !columns[l.Table][f] {
				return nil, fmt.Errorf("restore %s row %d: its %q is no column of this database's %s — a backup from another Rhea; this one would restore it without it", l.Table, counts[l.Table]+1, f, l.Table)
			}
		}
		if l.Table == "event" {
			var id int64
			if err := json.Unmarshal(fields["event_id"], &id); err != nil {
				return nil, fmt.Errorf("restore event row %d: %w", counts["event"]+1, err)
			}
			maxID = max(maxID, id)
		}
		if st != nil {
			switch l.Table {
			case "activity":
				if err := st.door(l.Row); err != nil {
					return nil, fmt.Errorf("restore activity row %d: %w", counts[l.Table]+1, err)
				}
			case "event":
				if l.Row, err = st.fact(l.Row); err != nil {
					return nil, fmt.Errorf("restore event row %d: %w", counts[l.Table]+1, err)
				}
			}
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
	if format >= 2 {
		if err := checkEnd(end, counts, maxID); err != nil {
			return nil, err
		}
	}
	if counts["event"] > 0 {
		if _, err := tx.Exec(ctx, "SELECT setval(pg_get_serial_sequence('event', 'event_id'), (SELECT max(event_id) FROM event))"); err != nil {
			return nil, fmt.Errorf("restore: the event sequence: %w", err)
		}
	}
	return counts, tx.Commit(ctx)
}

// tableColumns is a table's columns in this database.
func tableColumns(ctx context.Context, tx pgx.Tx, table string) (map[string]bool, error) {
	rows, err := tx.Query(ctx, `SELECT column_name FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = $1`, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols := map[string]bool{}
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		cols[c] = true
	}
	return cols, rows.Err()
}

func countLines(c BackupCounts) int {
	n := 0
	for _, v := range c {
		n += v
	}
	return n
}
