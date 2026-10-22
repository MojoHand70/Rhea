// Package project builds the DuckDB read side: projections of the Postgres
// event log that serve analysis views (SPEC §4). The DuckDB file is derived
// state — deletable, rebuildable, never the system of record.
package project

import (
	"context"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/duckdb/duckdb-go/v2"

	"rhea/internal/core"
	"rhea/internal/store"
)

func Path() string {
	if p := os.Getenv("RHEA_DUCKDB"); p != "" {
		return p
	}
	return "data/rhea.duckdb"
}

var (
	//go:embed objects_create.sql
	objectsCreateSQL string
	//go:embed objects_insert.sql
	objectsInsertSQL string
	//go:embed schemas_list.sql
	schemasListSQL string
	//go:embed meta_create.sql
	metaCreateSQL string
	//go:embed meta_insert.sql
	metaInsertSQL string
	//go:embed meta_read.sql
	metaReadSQL string
	//go:embed check_select.sql
	checkSelectSQL string
)

// shape names the copy's layout; a change to what Rebuild lays down bumps
// it, so an older copy is rebuilt rather than read.
const shape = "collection-1"

// Rebuild replays the derived events of the log into the read side: the
// `objects` table, state as JSON text for the packs' hand-written SQL
// (json_extract_string etc.), and the collection — a schema per domain and
// a typed table per object type (DIRECTION 2026-10-17, "Reporting"). Beside
// the immediate provenance (source_event_id, rule), every row carries
// root_event_id — the raw event its cascade chain started from — so analysis
// can group both sides of an intercompany position by the one fact that
// caused them (E5: eliminations match on provenance, not on heuristics).
// The copy is rebuilt only when the log, the scope or the catalog moved
// since it was last built.
func Rebuild(ctx context.Context, s *store.Store, duckPath string) (int, error) {
	types, err := s.ListObjectTypes(ctx)
	if err != nil {
		return 0, err
	}
	last, count, err := s.LogMark(ctx)
	if err != nil {
		return 0, err
	}
	cat, _ := json.Marshal(types)
	sum := sha256.Sum256(cat)
	// The copy is the whole log's: no scope in its key, or two tabs on two
	// companies would rebuild it on every alternate request.
	builtFrom := fmt.Sprintf("%s|%d|%d|%s", shape, last, count, hex.EncodeToString(sum[:8]))

	if err := os.MkdirAll(filepath.Dir(duckPath), 0o755); err != nil {
		return 0, err
	}
	db, err := sql.Open("duckdb", duckPath)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	var was string
	var n int
	if err := db.QueryRowContext(ctx, metaReadSQL).Scan(&was, &n); err == nil && was == builtFrom {
		return n, nil // the log has not moved: the copy is the log's
	}

	// The copy is built from what the caller's scope sees — under a scope,
	// the database shows that scope's rows and the installation's — into
	// that scope's own file (PathFor), so an office's analysis holds its
	// clients and nothing else, and a pack's SQL needs no filter of its own.
	derived, err := s.EventsByKind(ctx, core.KindDerived)
	if err != nil {
		return 0, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, objectsCreateSQL); err != nil {
		return 0, err
	}
	// The collection is laid down anew: every schema an earlier catalog
	// made goes, and this catalog's tables are created empty.
	old, err := tx.QueryContext(ctx, schemasListSQL)
	if err != nil {
		return 0, err
	}
	var drop []string
	for old.Next() {
		var name string
		if err := old.Scan(&name); err != nil {
			old.Close()
			return 0, err
		}
		drop = append(drop, name)
	}
	old.Close()
	for _, name := range drop {
		if _, err := tx.ExecContext(ctx, "DROP SCHEMA "+quoteIdent(name)+" CASCADE"); err != nil {
			return 0, err
		}
	}
	tables := Collection(types)
	byType := map[string]Table{}
	for _, t := range tables {
		if _, err := tx.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS "+quoteIdent(t.Schema)); err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx, t.ddl()); err != nil {
			return 0, fmt.Errorf("collection %s.%s: %w", t.Schema, t.Name, err)
		}
		byType[t.Type] = t
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
	inserts := map[string]*sql.Stmt{}
	for _, t := range tables {
		st, err := tx.PrepareContext(ctx, t.insert())
		if err != nil {
			return 0, err
		}
		defer st.Close()
		inserts[t.Type] = st
	}
	n = 0
	for _, id := range order {
		r := rows[id]
		state, err := json.Marshal(r.mat.State)
		if err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx, objectsInsertSQL,
			r.mat.ObjectID, r.mat.ObjectType, r.mat.TypeVersion, string(state),
			*r.ev.CauseEventID, r.ev.RuleID, r.ev.RuleVersion, r.ev.OccurredAt,
			r.root, r.ev.Company); err != nil {
			return 0, err
		}
		if t, ok := byType[r.mat.ObjectType]; ok {
			vals := make([]any, 0, len(t.Columns))
			for _, c := range t.Columns {
				switch {
				case c.Name == "id":
					vals = append(vals, r.mat.ObjectID)
				case c.Field != "":
					vals = append(vals, cell(c.Kind, r.mat.State[c.Field]))
				}
			}
			vals = append(vals, r.mat.TypeVersion, *r.ev.CauseEventID, r.ev.RuleID, r.ev.RuleVersion, r.ev.OccurredAt, r.root, r.ev.Company, string(state))
			if _, err := inserts[t.Type].ExecContext(ctx, vals...); err != nil {
				return 0, fmt.Errorf("collection %s.%s, %s: %w", t.Schema, t.Name, r.mat.ObjectID, err)
			}
		}
		n++
	}
	if _, err := tx.ExecContext(ctx, metaCreateSQL); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, metaInsertSQL, builtFrom, n); err != nil {
		return 0, err
	}
	return n, tx.Commit()
}

// The locked session (DIRECTION, security rule 1; invariant 3): the copy
// opened read-only, no files, no network, no extensions to install or load,
// the configuration frozen before any query arrives, bounded memory.
const locked = "?access_mode=read_only&enable_external_access=false&autoload_known_extensions=false" +
	"&autoinstall_known_extensions=false&lock_configuration=true&memory_limit=1GB&threads=2&max_temp_directory_size=256MB"

// Limits on one query in the locked session.
const (
	QueryTimeout = 30 * time.Second
	MaxRows      = 10000
)

// ErrNotOneSelect refuses a query that is not exactly one SELECT.
var ErrNotOneSelect = errors.New("a query is one SELECT")

// Query runs SQL in the locked session and returns column names plus rows
// rendered as strings, ready for the shell's generic table renderer — the
// packs' analysis views and the queries she writes alike. More than MaxRows
// rows is refused, not cut short: a figure must not be read from part of
// its rows.
func Query(ctx context.Context, duckPath, query string) ([]string, [][]string, error) {
	cols, rows, more, err := Run(ctx, duckPath, query, MaxRows)
	if err == nil && more {
		err = fmt.Errorf("the query returns more than %d rows", MaxRows)
	}
	return cols, rows, err
}

// Run is Query with a row limit of the caller's: it reads at most limit
// rows and says whether there were more.
func Run(ctx context.Context, duckPath, query string, limit int) ([]string, [][]string, bool, error) {
	cols, _, rows, more, err := RunTyped(ctx, duckPath, query, limit)
	return cols, rows, more, err
}

// RunTyped is Run with each column's DuckDB type beside its name
// (DECIMAL(18,2), DATE, VARCHAR…), so a face knows money from a count.
func RunTyped(ctx context.Context, duckPath, query string, limit int) ([]string, []string, [][]string, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()
	db, err := sql.Open("duckdb", duckPath+locked)
	if err != nil {
		return nil, nil, nil, false, err
	}
	defer db.Close()
	if err := oneSelect(ctx, db, query); err != nil {
		return nil, nil, nil, false, err
	}
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, nil, nil, false, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, nil, false, err
	}
	cts, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, nil, false, err
	}
	types := make([]string, len(cts))
	for i, ct := range cts {
		types[i] = ct.DatabaseTypeName()
	}
	var out [][]string
	for rows.Next() {
		if len(out) == limit {
			return cols, types, out, true, nil
		}
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, nil, nil, false, err
		}
		row := make([]string, len(cols))
		for i, v := range vals {
			row[i] = text(v)
		}
		out = append(out, row)
	}
	return cols, types, out, false, rows.Err()
}

// oneSelect asks DuckDB how it reads the query: exactly one statement, and
// that one a SELECT (with its CTEs, joins, window functions and PIVOT …
// FOR … IN). Anything else is refused before it runs — on top of a session
// that could not write or reach anything anyway.
func oneSelect(ctx context.Context, db *sql.DB, query string) error {
	var raw string
	if err := db.QueryRowContext(ctx, checkSelectSQL, query).Scan(&raw); err != nil {
		return err
	}
	var parsed struct {
		Error      bool              `json:"error"`
		Message    string            `json:"error_message"`
		Statements []json.RawMessage `json:"statements"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return err
	}
	if parsed.Error {
		return fmt.Errorf("%w: %s", ErrNotOneSelect, strings.TrimSuffix(parsed.Message, "!"))
	}
	if len(parsed.Statements) != 1 {
		return fmt.Errorf("%w, not %d statements", ErrNotOneSelect, len(parsed.Statements))
	}
	return nil
}

// text renders one value as the shell's canonical string: a decimal with
// all its places (0.10, not 0.1), a date as a date, NULL as empty.
func text(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case duckdb.Decimal:
		return decimalText(x)
	case time.Time:
		if x.Hour() == 0 && x.Minute() == 0 && x.Second() == 0 && x.Nanosecond() == 0 {
			return x.Format("2006-01-02")
		}
		return x.Format(time.RFC3339)
	case []byte:
		return string(x)
	}
	return core.Text(v)
}

func decimalText(d duckdb.Decimal) string {
	if d.Value == nil {
		return ""
	}
	neg := d.Value.Sign() < 0
	digits := new(big.Int).Abs(d.Value).String()
	scale := int(d.Scale)
	if scale > 0 {
		if len(digits) <= scale {
			digits = strings.Repeat("0", scale-len(digits)+1) + digits
		}
		digits = digits[:len(digits)-scale] + "." + digits[len(digits)-scale:]
	}
	if neg {
		return "-" + digits
	}
	return digits
}

// Wipe deletes the copy, so the next Rebuild lays it down from the log
// whatever it was built from: what `rhea replay` means by from scratch.
func Wipe(duckPath string) error {
	for _, p := range []string{duckPath, duckPath + ".wal"} {
		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

// PathOf is the file of the copy a context reads: its scope's, and within a
// wider scope the company it acts for, because the copy is filtered by the
// company acted for — one file for a desk acting for alfa and then for beta
// served alfa's figures as beta's (review 2026-10-21, B2). A context acting
// for its scope's one company keeps the scope's file.
func PathOf(ctx context.Context, base string) string {
	scope, scoped := store.ScopeOf(ctx)
	if c := store.CompanyOf(ctx); c != "" && !(len(scope) == 1 && scope[0] == c) {
		scope = append(append([]string{}, scope...), "acting:"+c)
	}
	return PathFor(base, scope, scoped)
}

// PathFor is the analysis file of a scope: the base path for the unscoped
// kernel, a file of its own per scope — the copy is what the scope sees.
func PathFor(base string, scope []string, scoped bool) string {
	if !scoped {
		return base
	}
	sum := sha256.Sum256([]byte(strings.Join(scope, ",")))
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext) + "-" + hex.EncodeToString(sum[:6]) + ext
}
