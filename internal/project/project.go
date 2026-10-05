// Package project builds the DuckDB read side: projections of the Postgres
// event log that serve analysis views (SPEC §4). The DuckDB file is derived
// state — deletable, rebuildable, never the system of record.
package project

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/duckdb/duckdb-go/v2"

	"rhea/internal/core"
	"rhea/internal/store"
)

func Path() string {
	if p := os.Getenv("RHEA_DUCKDB"); p != "" {
		return p
	}
	return "data/rhea.duckdb"
}

// Rebuild replays the derived events of the log into a fresh `objects` table.
// State lands as JSON text; analysis SQL reaches into it with DuckDB's JSON
// functions (json_extract_string etc.). Beside the immediate provenance
// (source_event_id, rule), every row carries root_event_id — the raw event
// its cascade chain started from — so analysis can group both sides of an
// intercompany position by the one fact that caused them (E5: eliminations
// match on provenance, not on heuristics).
func Rebuild(ctx context.Context, s *store.Store, duckPath string) (int, error) {
	derived, err := s.EventsByKind(ctx, core.KindDerived)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(filepath.Dir(duckPath), 0o755); err != nil {
		return 0, err
	}
	db, err := sql.Open("duckdb", duckPath)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
		DROP TABLE IF EXISTS objects;
		CREATE TABLE objects (
			object_id       VARCHAR PRIMARY KEY,
			object_type     VARCHAR NOT NULL,
			type_version    INTEGER NOT NULL,
			state           VARCHAR NOT NULL,
			source_event_id BIGINT NOT NULL,
			rule_id         VARCHAR NOT NULL,
			rule_version    INTEGER NOT NULL,
			occurred_at     DATE NOT NULL,
			root_event_id   BIGINT NOT NULL
		)`); err != nil {
		return 0, err
	}
	// A cause that is itself a derived event chains onward; causes precede
	// their effects in the log, so one ordered pass resolves every root.
	// Rows build in memory first: amendments merge into the state their
	// materialization produced, in log order, and only the final projection
	// lands in DuckDB — same replay, different sink.
	root := map[int64]int64{}
	rootOf := func(cause int64) int64 {
		if r, ok := root[cause]; ok {
			return r
		}
		return cause // a raw event: the chain's root
	}
	type row struct {
		mat  core.MaterializedObject
		ev   core.Event
		root int64
	}
	var order []string
	rows := map[string]*row{}
	for _, ev := range derived {
		root[ev.ID] = rootOf(*ev.CauseEventID)
		switch ev.Type {
		case core.EventObjectMaterialized:
			var mat core.MaterializedObject
			if err := json.Unmarshal(ev.Payload, &mat); err != nil {
				return 0, fmt.Errorf("derived event %d: %w", ev.ID, err)
			}
			if _, seen := rows[mat.ObjectID]; !seen {
				order = append(order, mat.ObjectID)
			}
			rows[mat.ObjectID] = &row{mat: mat, ev: ev, root: root[ev.ID]}
		case core.EventObjectAmended:
			var am core.AmendedObject
			if err := json.Unmarshal(ev.Payload, &am); err != nil {
				return 0, fmt.Errorf("derived event %d: %w", ev.ID, err)
			}
			r, ok := rows[am.ObjectID]
			if !ok {
				return 0, fmt.Errorf("derived event %d amends unknown object %q", ev.ID, am.ObjectID)
			}
			for k, v := range am.Set {
				r.mat.State[k] = v
			}
		}
	}
	n := 0
	for _, id := range order {
		r := rows[id]
		state, err := json.Marshal(r.mat.State)
		if err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO objects VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			r.mat.ObjectID, r.mat.ObjectType, r.mat.TypeVersion, string(state),
			*r.ev.CauseEventID, r.ev.RuleID, r.ev.RuleVersion, r.ev.OccurredAt,
			r.root); err != nil {
			return 0, err
		}
		n++
	}
	return n, tx.Commit()
}

// Query runs an analysis view's SQL read-only and returns column names plus
// rows rendered as strings, ready for the shell's generic table renderer.
func Query(ctx context.Context, duckPath, query string) ([]string, [][]string, error) {
	db, err := sql.Open("duckdb", duckPath+"?access_mode=read_only")
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	var out [][]string
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, nil, err
		}
		row := make([]string, len(cols))
		for i, v := range vals {
			if v == nil {
				row[i] = ""
				continue
			}
			row[i] = fmt.Sprintf("%v", v)
		}
		out = append(out, row)
	}
	return cols, out, rows.Err()
}
