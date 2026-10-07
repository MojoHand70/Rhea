package exec_test

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

// The bundle (DIRECTION, implementation by interview; KK 2026-10-08): one
// approval activates a set of drafts, whole or not at all. The forcing case
// came out of the eval: goods.received needs a receipt AND a stock movement.
// Approved one rule at a time, the first explains the event and the second
// only ever sees the future; approved together, both fire on the waiting
// event, and the receipt's type itself arrives in the same bundle.
func TestBundleTwoConsequences(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}

	// The vocabulary already live: master data and the movement ledger.
	for _, ot := range []core.ObjectType{
		{Name: "item", Version: 1, Domain: "warehouse", LabelField: "name", Fields: []core.FieldDef{
			{Name: "sku", Type: "string", Required: true}, {Name: "name", Type: "string", Required: true}}},
		{Name: "location", Version: 1, Domain: "warehouse", LabelField: "name", Fields: []core.FieldDef{
			{Name: "code", Type: "string", Required: true}, {Name: "name", Type: "string", Required: true}}},
		{Name: "stock_movement", Version: 1, Domain: "warehouse", Fields: []core.FieldDef{
			{Name: "item", Type: "ref<item>", Required: true},
			{Name: "location", Type: "ref<location>", Required: true},
			{Name: "direction", Type: "enum", Values: []string{"in", "out"}, Required: true},
			{Name: "qty", Type: "int", Required: true}}},
	} {
		if err := s.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	active := func(id string, spec core.RuleSpec) {
		t.Helper()
		if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: id, Status: core.StatusActive, Priority: 10,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: spec}); err != nil {
			t.Fatal(err)
		}
	}
	active("register-item", core.RuleSpec{Match: core.Match{EventType: "item.created"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "item",
			Fields: map[string]string{"sku": "=$.sku", "name": "=$.name"}}}})
	active("register-location", core.RuleSpec{Match: core.Match{EventType: "location.created"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "location",
			Fields: map[string]string{"code": "=$.code", "name": "=$.name"}}}})
	submit := func(typ, payload string) {
		t.Helper()
		var p map[string]any
		json.Unmarshal([]byte(payload), &p)
		if _, err := x.TriggerActivity(ctx, "submit_event", map[string]any{
			"event_type": typ, "occurred_at": "2026-09-10", "payload": p,
		}, "test", ""); err != nil {
			t.Fatal(err)
		}
	}
	submit("item.created", `{"sku":"WID-1","name":"Widget"}`)
	submit("location.created", `{"code":"MAIN","name":"Main"}`)
	submit("goods.received", `{"item":"WID-1","location":"MAIN","qty":25}`)

	// The answer, drafted: a new document type, its rule, the movement rule
	// for the same event, and a view.
	receipt := core.ObjectType{Name: "goods_receipt", Version: 1, Domain: "warehouse", IsDocument: true,
		Fields: []core.FieldDef{
			{Name: "item", Type: "ref<item>", Required: true},
			{Name: "location", Type: "ref<location>", Required: true},
			{Name: "qty", Type: "int", Required: true}}}
	if err := store.InsertObjectTypeRow(ctx, s.Pool, receipt, core.StatusDraft); err != nil {
		t.Fatal(err)
	}
	view := core.ViewDef{ID: "receipt-list", Version: 1, Notion: "list", Title: "Receipts",
		Domain: "warehouse", Function: "inbound",
		Spec: json.RawMessage(`{"object_type":"goods_receipt","columns":[{"field":"qty","label":"Qty"}]}`)}
	if err := store.InsertViewDefRow(ctx, s.Pool, view, core.StatusDraft); err != nil {
		t.Fatal(err)
	}
	refs := map[string]string{"item": "=ref(item, sku, $.item)", "location": "=ref(location, code, $.location)", "qty": "=$.qty"}
	moveFields := map[string]string{"direction": "in"}
	for k, v := range refs {
		moveFields[k] = v
	}
	for id, eff := range map[string]core.ObjectTemplate{
		"book-goods-receipt": {Type: "goods_receipt", Fields: refs},
		"move-stock-in":      {Type: "stock_movement", Fields: moveFields},
	} {
		if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: id, Status: core.StatusDraft, Priority: 100,
			EffectiveFrom: "2026-01-01", CreatedBy: "agent:test", Description: id,
			Spec: core.RuleSpec{Match: core.Match{EventType: "goods.received"}, Effect: core.Effect{Object: eff}}}); err != nil {
			t.Fatal(err)
		}
	}
	b, err := s.InsertBundle(ctx, core.Bundle{ID: "inbound-goods", CreatedBy: "agent:test",
		Description: "Goods received become receipts and move stock in",
		Members: []core.Member{
			{Kind: core.KindObjectType, Name: "goods_receipt", Version: 1},
			{Kind: core.KindViewDef, Name: "receipt-list", Version: 1},
			{Kind: core.KindRule, Name: "book-goods-receipt", Version: 1},
			{Kind: core.KindRule, Name: "move-stock-in", Version: 1},
		}})
	if err != nil {
		t.Fatal(err)
	}

	// A draft type is not language yet: invisible to the kernel and the shell.
	if _, err := s.GetObjectType(ctx, "goods_receipt"); err == nil {
		t.Fatal("draft type visible before its bundle is approved")
	}
	// No partial approval: the single-rule door refuses a bundle member.
	if _, _, _, err := x.ApproveRule(ctx, "move-stock-in", "krzysztof", "2026-09-10"); err == nil ||
		!strings.Contains(err.Error(), "approve the bundle") {
		t.Fatalf("member approved alone: %v", err)
	}

	// The dry run sees the bundle as approval would land it, draft type included.
	diff, err := x.SimulateBundle(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(diff.Errors) != 0 || len(diff.Added) != 2 ||
		len(diff.UnexplainedBefore)-len(diff.UnexplainedAfter) != 1 {
		t.Fatalf("bundle dry run: errors %v, added %d, unexplained %v → %v",
			diff.Errors, len(diff.Added), diff.UnexplainedBefore, diff.UnexplainedAfter)
	}

	got, booked, procErrs, err := x.ApproveBundle(ctx, "inbound-goods", "krzysztof", "2026-09-10")
	if err != nil || len(procErrs) != 0 {
		t.Fatalf("approve: %v %v", err, procErrs)
	}
	if got.Status != core.StatusActive || booked != 1 { // one event explained, twice
		t.Fatalf("bundle %s, booked %d", got.Status, booked)
	}
	// Both consequences of the one event, each explained by its own rule.
	receipts, _ := s.ObjectsByType(ctx, "goods_receipt")
	moves, _ := s.ObjectsByType(ctx, "stock_movement")
	if len(receipts) != 1 || len(moves) != 1 || receipts[0].SourceEventID != moves[0].SourceEventID {
		t.Fatalf("receipts %v, moves %v", receipts, moves)
	}
	if _, err := s.GetObjectType(ctx, "goods_receipt"); err != nil {
		t.Fatalf("type not active after approval: %v", err)
	}
	if vds, _ := s.LatestViewDefs(ctx); len(vds) != 1 || vds[0].ID != "receipt-list" {
		t.Fatalf("views = %v", vds)
	}
	// The approval is one event naming its members (invariant 2), and a
	// decided bundle cannot be decided again.
	approval, ok, _ := s.LatestEventOfType(ctx, "bundle.approved")
	if !ok || !strings.Contains(string(approval.Payload), `"move-stock-in"`) || approval.ActivityName != "approve_bundle" {
		t.Fatalf("approval event = %+v", approval)
	}
	if _, _, _, err := x.ApproveBundle(ctx, "inbound-goods", "krzysztof", "2026-09-10"); err == nil {
		t.Fatal("bundle approved twice")
	}

	// Invariant 4 holds through a bundle: replay reproduces the state.
	before, _ := s.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := s.AllObjects(ctx)
	strip := func(objs []core.Object) map[string]map[string]any {
		m := map[string]map[string]any{}
		for _, o := range objs {
			m[o.ID] = o.State
		}
		return m
	}
	if !reflect.DeepEqual(strip(before), strip(after)) {
		t.Fatal("replay diverged after a bundle approval")
	}
}

// A rejection supersedes the bundle and keeps the reason on the record; its
// drafts stay rejected drafts, locked by membership: a redraft is a new
// version in a new bundle.
func TestBundleRejection(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	seed(t, x)
	draftRule(t, x)
	if _, err := s.InsertBundle(ctx, core.Bundle{ID: "invoices", CreatedBy: "agent:test",
		Description: "invoices", Members: []core.Member{{Kind: core.KindRule, Name: "book-pln-invoice", Version: 1}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := x.RejectBundle(ctx, "invoices", "", "krzysztof", "2026-09-15"); err == nil {
		t.Fatal("rejection without a reason accepted")
	}
	b, err := x.RejectBundle(ctx, "invoices", "EUR invoices must book too", "krzysztof", "2026-09-15")
	if err != nil || b.Status != core.StatusSuperseded {
		t.Fatalf("reject: %v, %+v", err, b)
	}
	ev, ok, _ := s.LatestEventOfType(ctx, "bundle.rejected")
	if !ok || !strings.Contains(string(ev.Payload), "EUR invoices must book too") {
		t.Fatalf("rejection event = %+v", ev)
	}
	if _, _, _, err := x.ApproveBundle(ctx, "invoices", "krzysztof", "2026-09-15"); err == nil {
		t.Fatal("rejected bundle approved")
	}
	if _, _, _, err := x.ApproveRule(ctx, "book-pln-invoice", "krzysztof", "2026-09-15"); err == nil ||
		!strings.Contains(err.Error(), "redraft") {
		t.Fatalf("rejected draft approved alone: %v", err)
	}
	if _, err := s.InsertBundle(ctx, core.Bundle{ID: "invoices-again", CreatedBy: "agent:test",
		Description: "same draft", Members: []core.Member{{Kind: core.KindRule, Name: "book-pln-invoice", Version: 1}}}); err == nil {
		t.Fatal("a rejected draft joined a second bundle")
	}
	// The redraft: a new version, a new bundle, and it activates.
	r, _ := s.GetRule(ctx, "book-pln-invoice")
	r.Description = "Invoices in any currency become Invoice documents"
	r.Spec.Match.Where = nil
	r, err = s.InsertRuleVersion(ctx, r)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.InsertBundle(ctx, core.Bundle{ID: "invoices-v2", CreatedBy: "agent:test",
		Description: "redraft", Members: []core.Member{{Kind: core.KindRule, Name: r.ID, Version: r.Version}}}); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := x.ApproveBundle(ctx, "invoices-v2", "krzysztof", "2026-09-15"); err != nil {
		t.Fatal(err)
	}
	if active, _ := s.ActiveRules(ctx); len(active) != 1 || active[0].Spec.Match.Where != nil {
		t.Fatalf("active = %+v", active)
	}
}
