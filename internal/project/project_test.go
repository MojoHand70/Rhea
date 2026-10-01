package project_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

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

	n, err := project.Rebuild(ctx, s, duck)
	if err != nil || n != 2 {
		t.Fatalf("rebuild = %d, %v", n, err)
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
