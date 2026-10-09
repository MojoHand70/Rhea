package exec_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

// "Yes, absolutely we need a PZ document" (KK, 2026-10-08). A delivery gets
// one PZ — the goods received note — with one line per delivered item, each
// line pointing at its PZ and moving that item's stock in. The lines live in
// the delivery, not in the PZ header that cascades them: a cascade rule reads
// the chain's root fact under $.root, and carried links tie line to header.
func TestDocumentWithLines(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	for _, ot := range []core.ObjectType{
		{Name: "item", Version: 1, Domain: "warehouse", LabelField: "sku", Fields: []core.FieldDef{
			{Name: "sku", Type: "string", Required: true}}},
		{Name: "location", Version: 1, Domain: "warehouse", LabelField: "code", Fields: []core.FieldDef{
			{Name: "code", Type: "string", Required: true}}},
		{Name: "pz", Version: 1, Domain: "warehouse", IsDocument: true, LabelField: "supplier_document", Fields: []core.FieldDef{
			{Name: "supplier_document", Type: "string", Required: true},
			{Name: "location", Type: "ref<location>", Required: true},
			{Name: "date", Type: "date", Required: true}}},
		{Name: "pz_line", Version: 1, Domain: "warehouse", Fields: []core.FieldDef{
			{Name: "document", Type: "ref<pz>", Required: true}, {Name: "line", Type: "int", Required: true},
			{Name: "item", Type: "ref<item>", Required: true}, {Name: "qty", Type: "int", Required: true},
			{Name: "received", Type: "date", Required: true}}},
		{Name: "stock_movement", Version: 1, Domain: "warehouse", Fields: []core.FieldDef{
			{Name: "item", Type: "ref<item>", Required: true}, {Name: "location", Type: "ref<location>", Required: true},
			{Name: "qty", Type: "int", Required: true}, {Name: "date", Type: "date", Required: true}}},
	} {
		if err := s.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	materialized := func(typ string) core.Match {
		return core.Match{EventType: core.EventObjectMaterialized,
			Where: []core.Condition{{Path: "$.object_type", Op: "eq", Value: typ}}}
	}
	for _, r := range []struct {
		id   string
		prio int
		spec core.RuleSpec
	}{
		{"register-item", 10, core.RuleSpec{Match: core.Match{EventType: "item.created"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "item", Fields: map[string]string{"sku": "=$.sku"}}}}},
		{"register-location", 10, core.RuleSpec{Match: core.Match{EventType: "location.created"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "location", Fields: map[string]string{"code": "=$.code"}}}}},
		{"pz-from-delivery", 100, core.RuleSpec{Match: core.Match{EventType: "delivery.received"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "pz", Fields: map[string]string{
				"supplier_document": "=$.document", "location": "=ref(location, code, $.location)", "date": "=$.date"}}}}},
		// The header cascades its lines from the delivery the chain began with.
		{"pz-lines", 200, core.RuleSpec{Match: materialized("pz"),
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "pz_line", Each: "=$.root.lines[*]", Fields: map[string]string{
				"document": "=$.doc.object_id", "line": "=$.n",
				"item": "=ref(item, sku, $.line.item)", "qty": "=$.line.qty",
				// the chain's root stays readable inside each: the line's
				// date is the delivery's
				"received": "=$.root.date"}}}}},
		{"pz-line-moves-stock", 300, core.RuleSpec{Match: materialized("pz_line"),
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "stock_movement", Fields: map[string]string{
				"item": "=$.state.item", "qty": "=$.state.qty",
				"location": "=ref(location, code, $.root.location)", "date": "=$.root.date"}}}}},
	} {
		if err := r.spec.Validate(nil); err != nil {
			t.Fatalf("%s: %v", r.id, err)
		}
		if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: r.id, Status: core.StatusActive, Priority: r.prio,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: r.id, Spec: r.spec}); err != nil {
			t.Fatal(err)
		}
	}
	for _, e := range [][2]string{
		{"item.created", `{"sku":"WID-1"}`}, {"item.created", `{"sku":"GAD-1"}`},
		{"location.created", `{"code":"MAIN"}`},
		{"delivery.received", `{"document":"WZ 77/2026","location":"MAIN","date":"2026-09-14",
			"lines":[{"item":"WID-1","qty":10},{"item":"GAD-1","qty":4}]}`},
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
		t.Fatalf("pz %d, lines %d, movements %d — want one document, two lines, two movements", len(pzs), len(lines), len(moves))
	}
	for _, l := range lines {
		if l.State["document"] != pzs[0].ID || l.State["received"] != "2026-09-14" {
			t.Fatalf("line %v does not point at its PZ %s with its date", l.State, pzs[0].ID)
		}
	}
	for _, m := range moves {
		if m.State["location"] != pzs[0].State["location"] || m.State["date"] != "2026-09-14" {
			t.Fatalf("movement %v", m.State)
		}
	}
	// $.root is what a cascade may read, never what it writes: the log's
	// derived events carry the materialization only.
	derived, _ := s.EventsByKind(ctx, core.KindDerived)
	for _, d := range derived {
		if strings.Contains(string(d.Payload), `"root"`) {
			t.Fatalf("root leaked into the log: %s", d.Payload)
		}
	}
	assertReplayIdentical(t, x)
}
