package pack_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/store/storetest"
)

// The PZ as a market standard (DIRECTION, "Rhea suggests standards"): the
// pl-warehouse pack proposes the customary Polish practice with its warrant —
// a delivery becomes a PZ with lines, each line takes stock in and books
// Wn 330 / Ma 300 at purchase price — and one approval installs it.
func TestPolishPZPack(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	for _, f := range []string{"finance.json", "finance_v2.json", "finance_v3.json", "finance_v4.json",
		"finance_v5.json", "finance_v6.json", "finance_v7.json", "finance_v8.json", "finance_v9.json", "warehouse.json"} {
		b, err := os.ReadFile(filepath.Join("..", "..", "testdata", "seed", f))
		if err != nil {
			t.Fatal(err)
		}
		var defs struct {
			ObjectTypes []core.ObjectType `json:"object_types"`
		}
		if err := json.Unmarshal(b, &defs); err != nil {
			t.Fatal(err)
		}
		for _, ot := range defs.ObjectTypes {
			if err := s.InsertObjectType(ctx, ot); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, p := range []string{"pack.json", "warehouse.json"} {
		sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", "pl", p), "test")
		if err != nil {
			t.Fatalf("%s: %v", p, err)
		}
		bundle, err := s.GetBundle(ctx, sum.Bundle)
		if err != nil {
			t.Fatal(err)
		}
		if bundle.Warrant == nil || len(bundle.Warrant.Citations) == 0 || bundle.Warrant.Support != nil {
			t.Fatalf("%s warrant = %+v", p, bundle.Warrant)
		}
		if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-09-01"); err != nil || len(errs) != 0 {
			t.Fatalf("approve %s: %v %v", sum.Bundle, err, errs)
		}
	}

	// Master data the warehouse base does not ship rules for.
	for id, spec := range map[string]core.RuleSpec{
		"register-company": {Match: core.Match{EventType: "company.registered"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "company",
				Fields: map[string]string{"name": "=$.name", "kind": "=$.kind"}}}},
		"register-item": {Match: core.Match{EventType: "item.created"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "item",
				Fields: map[string]string{"sku": "=$.sku", "name": "=$.name"}}}},
		"register-location": {Match: core.Match{EventType: "location.created"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "location",
				Fields: map[string]string{"code": "=$.code", "name": "=$.name"}}}},
	} {
		if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: id, Status: core.StatusActive, Priority: 10,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: spec}); err != nil {
			t.Fatal(err)
		}
	}
	for _, e := range [][2]string{
		{"company.registered", `{"name":"Omikron GmbH","kind":"supplier"}`},
		{"item.created", `{"sku":"WID-1","name":"Widget"}`},
		{"item.created", `{"sku":"GAD-1","name":"Gadget"}`},
		{"location.created", `{"code":"MAIN","name":"Magazyn główny"}`},
		{"delivery.received", `{"supplier":"Omikron GmbH","document":"WZ 77/2026","location":"MAIN",
			"date":"2026-09-14","currency":"PLN",
			"lines":[{"item":"WID-1","qty":10,"value":"250.00"},{"item":"GAD-1","qty":4,"value":"120.40"}]}`},
	} {
		var p map[string]any
		if err := json.Unmarshal([]byte(e[1]), &p); err != nil {
			t.Fatal(err)
		}
		tr, err := x.TriggerActivity(ctx, "submit_event", map[string]any{
			"event_type": e[0], "occurred_at": "2026-09-14", "payload": p}, "test", "")
		if err != nil || len(tr.Errors) != 0 {
			t.Fatalf("%s: %v %v", e[0], err, tr.Errors)
		}
	}

	pzs, _ := s.ObjectsByType(ctx, "pz")
	lines, _ := s.ObjectsByType(ctx, "pz_line")
	moves, _ := s.ObjectsByType(ctx, "stock_movement")
	if len(pzs) != 1 || len(lines) != 2 || len(moves) != 2 {
		t.Fatalf("pz %d, lines %d, movements %d", len(pzs), len(lines), len(moves))
	}
	postings, _ := s.ObjectsByType(ctx, "posting")
	var debit330, credit300 int64
	for _, p := range postings {
		acc, _ := s.GetObject(ctx, p.State["account"].(string))
		// Money is int64 minor units; read it as an integer, never a float.
		amount, err := strconv.ParseInt(fmt.Sprint(p.State["amount"]), 10, 64)
		if err != nil {
			t.Fatalf("amount %v: %v", p.State["amount"], err)
		}
		switch {
		case acc.State["code"] == "330" && p.State["side"] == "debit" && p.State["book"] == "pl-stat":
			debit330 += amount
		case acc.State["code"] == "300" && p.State["side"] == "credit" && p.State["book"] == "pl-stat":
			credit300 += amount
		}
	}
	if debit330 != 37040 || credit300 != 37040 {
		t.Fatalf("Wn 330 %d / Ma 300 %d, want 370.40 both sides", debit330, credit300)
	}
}
