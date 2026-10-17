package project_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/duckdb/duckdb-go/v2"

	"rhea/internal/core"
	"rhea/internal/project"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

func TestRebuildAndQuery(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	duck := filepath.Join(t.TempDir(), "test.duckdb")

	// Two materialized invoices in the log.
	for i, c := range []struct{ cust, total string }{{"ACME", "350.50"}, {"Beta", "100.00"}} {
		raw, err := s.AppendEvent(ctx, core.Event{
			Kind: core.KindRaw, Type: "invoice.received", OccurredAt: "2026-09-15",
			Payload: json.RawMessage(`{}`), DedupKey: string(rune('a' + i)),
		})
		if err != nil {
			t.Fatal(err)
		}
		totalMinor, _ := core.ParseMoney(c.total)
		mat, _ := json.Marshal(core.MaterializedObject{
			ObjectID: "invoice-" + string(rune('a'+i)), ObjectType: "invoice", TypeVersion: 1,
			State: map[string]any{"customer": c.cust, "currency": "PLN", "total": totalMinor},
		})
		if _, err := s.AppendEvent(ctx, core.Event{
			Kind: core.KindDerived, Type: core.EventObjectMaterialized,
			OccurredAt: "2026-09-15", Payload: mat,
			CauseEventID: &raw, RuleID: "r", RuleVersion: 1,
		}); err != nil {
			t.Fatal(err)
		}
	}

	// A second-generation materialization: caused by invoice-b's own
	// materialization event, so its root must chase back to b's raw event.
	derived, _ := s.EventsByKind(ctx, core.KindDerived)
	causeB := derived[len(derived)-1].ID
	mat, _ := json.Marshal(core.MaterializedObject{
		ObjectID: "note-b", ObjectType: "note", TypeVersion: 1,
		State: map[string]any{"text": "cascaded"},
	})
	if _, err := s.AppendEvent(ctx, core.Event{
		Kind: core.KindDerived, Type: core.EventObjectMaterialized,
		OccurredAt: "2026-09-15", Payload: mat,
		CauseEventID: &causeB, RuleID: "r2", RuleVersion: 1,
	}); err != nil {
		t.Fatal(err)
	}

	n, err := project.Rebuild(ctx, s, duck)
	if err != nil || n != 3 {
		t.Fatalf("rebuild = %d, %v", n, err)
	}

	// Provenance reaches the read side: the cascaded note's root is the raw
	// event behind invoice-b, not the materialization that caused it.
	_, roots, err := project.Query(ctx, duck, `
		SELECT object_id, CAST(source_event_id AS VARCHAR), CAST(root_event_id AS VARCHAR)
		FROM objects WHERE object_id IN ('invoice-b', 'note-b') ORDER BY object_id`)
	if err != nil {
		t.Fatal(err)
	}
	if len(roots) != 2 || roots[0][1] != roots[0][2] || // invoice-b: cause is the root
		roots[1][1] == roots[1][2] || roots[1][2] != roots[0][2] { // note-b: chained to b's root
		t.Fatalf("provenance rows = %v", roots)
	}

	// Money stays integer all the way; display formatting is integer math too.
	cols, rows, err := project.Query(ctx, duck, `
		SELECT json_extract_string(state, '$.customer') AS customer,
		       printf('%d.%02d',
		              CAST(json_extract(state, '$.total') AS BIGINT) // 100,
		              CAST(json_extract(state, '$.total') AS BIGINT) % 100) AS total
		FROM objects WHERE object_type = 'invoice' ORDER BY customer`)
	if err != nil {
		t.Fatal(err)
	}
	if len(cols) != 2 || len(rows) != 2 {
		t.Fatalf("cols=%v rows=%v", cols, rows)
	}
	if rows[0][0] != "ACME" || rows[0][1] != "350.50" {
		t.Fatalf("row 0 = %v", rows[0])
	}
}

var _ = store.DSN // keep the import honest if helpers shift

// TestTheCollection (DIRECTION 2026-10-17, "Reporting"): the business's own
// databases of objects as typed tables, one schema per domain — money exact,
// dates as dates, links as the linked id, the provenance beside each row —
// and a copy rebuilt only when the log or the catalog moved.
func TestTheCollection(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	duck := filepath.Join(t.TempDir(), "test.duckdb")
	for _, ot := range []core.ObjectType{
		{Name: "invoice", Version: 1, Domain: "sales", IsDocument: true, Fields: []core.FieldDef{
			{Name: "number", Type: "string"}, {Name: "net", Type: "money"}, {Name: "issued", Type: "date"},
			{Name: "buyer", Type: "ref<company>"}, {Name: "qty", Type: "int"}, {Name: "status", Type: "enum", Values: []string{"open", "paid"}}}},
		{Name: "building", Version: 1, Domain: "Real Estate", Fields: []core.FieldDef{{Name: "name", Type: "string"}, {Name: "area", Type: "decimal"}}},
	} {
		if err := s.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	add := func(key, id, typ string, state map[string]any) {
		t.Helper()
		raw, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: "x.happened", OccurredAt: "2026-09-15", Payload: json.RawMessage(`{}`), DedupKey: key})
		if err != nil {
			t.Fatal(err)
		}
		mat, _ := json.Marshal(core.MaterializedObject{ObjectID: id, ObjectType: typ, TypeVersion: 1, State: state})
		if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindDerived, Type: core.EventObjectMaterialized, OccurredAt: "2026-09-15",
			Payload: mat, CauseEventID: &raw, RuleID: "r", RuleVersion: 1}); err != nil {
			t.Fatal(err)
		}
	}
	add("a", "invoice-a", "invoice", map[string]any{"number": "FV 1", "net": 35050, "issued": "2026-09-15", "buyer": "company-1", "qty": 3, "status": "open"})
	add("b", "invoice-b", "invoice", map[string]any{"number": "FV 2", "net": 10, "issued": "not a date", "buyer": "company-2", "extra": "kept"})
	add("c", "building-c", "building", map[string]any{"name": "Kamienica", "area": "412.5"})
	if n, err := project.Rebuild(ctx, s, duck); err != nil || n != 3 {
		t.Fatalf("rebuild: %d %v", n, err)
	}
	cols, rows, err := project.Query(ctx, duck, `SELECT id, net, typeof(net), issued, buyer, qty, status, _rule, json_extract_string(_state, '$.extra') FROM sales.invoice ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	if len(cols) != 9 || fmt.Sprint(rows) != "[[invoice-a 350.50 DECIMAL(18,2) 2026-09-15 company-1 3 open r ] [invoice-b 0.10 DECIMAL(18,2)  company-2   r kept]]" {
		t.Fatalf("sales.invoice: %v", rows)
	}
	_, rows, err = project.Query(ctx, duck, `SELECT name, area + 0.5 FROM real_estate.building`)
	if err != nil || fmt.Sprint(rows) != "[[Kamienica 413.0000000000]]" {
		t.Fatalf("real_estate.building: %v %v", rows, err)
	}
	_, rows, err = project.Query(ctx, duck, `SELECT sum(net) FROM sales.invoice`)
	if err != nil || rows[0][0] != "350.60" {
		t.Fatalf("money adds up exactly: %v %v", rows, err)
	}

	// The log has not moved: the copy is not rebuilt (a mark left in it stays).
	mark := func() int {
		db, err := sql.Open("duckdb", duck)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		var n int
		db.QueryRow(`SELECT count(*) FROM objects WHERE object_id = 'mark'`).Scan(&n)
		return n
	}
	func() {
		db, _ := sql.Open("duckdb", duck)
		defer db.Close()
		if _, err := db.Exec(`INSERT INTO objects VALUES ('mark', 'x', 1, '{}', 1, 'r', 1, '2026-01-01', 1)`); err != nil {
			t.Fatal(err)
		}
	}()
	if _, err := project.Rebuild(ctx, s, duck); err != nil || mark() != 1 {
		t.Fatalf("rebuilt though nothing moved: %v", err)
	}
	add("d", "invoice-d", "invoice", map[string]any{"number": "FV 3", "net": 100})
	if n, err := project.Rebuild(ctx, s, duck); err != nil || n != 4 || mark() != 0 {
		t.Fatalf("not rebuilt after the log moved: %d %v", n, err)
	}
	// A type the business teaches her is a table the next time.
	if err := s.InsertObjectType(ctx, core.ObjectType{Name: "employee", Version: 1, Domain: "hr", Fields: []core.FieldDef{{Name: "salary", Type: "money"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := project.Rebuild(ctx, s, duck); err != nil {
		t.Fatal(err)
	}
	if _, rows, err := project.Query(ctx, duck, `SELECT count(*) FROM hr.employee`); err != nil || rows[0][0] != "0" {
		t.Fatalf("hr.employee: %v %v", rows, err)
	}
}

// TestTheLockedSession: a query she writes reads the copy and nothing else —
// one SELECT, no writing, no files, no network, no settings — within a time
// and a row limit (invariant 3, reworded 2026-10-17).
func TestTheLockedSession(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	dir := t.TempDir()
	duck := filepath.Join(dir, "test.duckdb")
	if _, err := project.Rebuild(ctx, s, duck); err != nil {
		t.Fatal(err)
	}
	secret := filepath.Join(dir, "secret.csv")
	os.WriteFile(secret, []byte("a\n1\n"), 0o644)
	for _, q := range []string{
		`SELECT 1; SELECT 2`, `CREATE TABLE x (a INT)`, `INSERT INTO objects VALUES ('x','x',1,'{}',1,'r',1,'2026-01-01',1)`,
		`DELETE FROM objects`, `COPY objects TO '` + filepath.Join(dir, "out.csv") + `'`, `SELECT * FROM read_csv('` + secret + `')`,
		`ATTACH '` + filepath.Join(dir, "y.duckdb") + `' AS y`, `INSTALL httpfs`, `SET enable_external_access = true`, `PRAGMA version`,
	} {
		if _, _, err := project.Query(ctx, duck, q); err == nil {
			t.Errorf("ran: %s", q)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "out.csv")); err == nil {
		t.Fatalf("a query wrote a file")
	}
	for _, q := range []string{
		`WITH t AS (SELECT range AS n, range % 3 AS g FROM range(10)) SELECT g, sum(n) OVER (PARTITION BY g ORDER BY n) FROM t`,
		`SELECT * FROM (SELECT 'a' AS k, 1 AS v UNION ALL SELECT 'b', 2) PIVOT (sum(v) FOR k IN ('a', 'b'))`,
	} {
		if _, _, err := project.Query(ctx, duck, q); err != nil {
			t.Errorf("%s: %v", q, err)
		}
	}
	short, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, _, err := project.Query(short, duck, `SELECT count(*) FROM range(100000000) a, range(100000000) b`); err == nil || time.Since(start) > 5*time.Second {
		t.Fatalf("a runaway query was not stopped: %v after %s", err, time.Since(start))
	}
	if _, rows, more, err := project.Run(ctx, duck, `SELECT * FROM range(10)`, 5); err != nil || len(rows) != 5 || !more {
		t.Fatalf("row cap: %d %v %v", len(rows), more, err)
	}
	if _, _, err := project.Query(ctx, duck, `SELECT * FROM range(20000)`); err == nil {
		t.Fatalf("a figure read from part of its rows")
	}
}
