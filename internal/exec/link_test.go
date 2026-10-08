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

// A follow-up may point at exactly the thing that caused it (KK, 2026-10-08,
// option A — the model's first reach four times over). A ref field takes a
// carried id: the causing object ($.object_id) or a ref it holds
// ($.state.item). The kernel vouches every carried id — right kind, existing —
// the courtesy the amendment already pays its target; a literal id is never a
// link.
func TestCarriedLinks(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	for _, ot := range []core.ObjectType{
		{Name: "item", Version: 1, Domain: "warehouse", LabelField: "sku", Fields: []core.FieldDef{
			{Name: "sku", Type: "string", Required: true}}},
		{Name: "goods_receipt", Version: 1, Domain: "warehouse", IsDocument: true, Fields: []core.FieldDef{
			{Name: "item", Type: "ref<item>", Required: true}, {Name: "qty", Type: "int", Required: true}}},
		{Name: "stock_movement", Version: 1, Domain: "warehouse", Fields: []core.FieldDef{
			{Name: "item", Type: "ref<item>", Required: true}, {Name: "qty", Type: "int", Required: true},
			{Name: "receipt", Type: "ref<goods_receipt>", Required: true}}},
		{Name: "complaint", Version: 1, Domain: "sales", IsDocument: true, LabelField: "number", Fields: []core.FieldDef{
			{Name: "number", Type: "string", Required: true}}},
		{Name: "case", Version: 1, Domain: "work", Fields: []core.FieldDef{
			{Name: "complaint", Type: "ref<complaint>", Required: true},
			{Name: "status", Type: "enum", Values: []string{"open", "closed"}, Required: true}},
			Lifecycle: &core.LifecycleDef{Field: "status", Transitions: map[string][]string{"open": {"closed"}}}},
	} {
		if err := s.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	rule := func(id string, prio int, spec core.RuleSpec) {
		t.Helper()
		if err := spec.Validate(nil); err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: id, Status: core.StatusActive, Priority: prio,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: spec}); err != nil {
			t.Fatal(err)
		}
	}
	materialized := func(typ string) core.Match {
		return core.Match{EventType: core.EventObjectMaterialized,
			Where: []core.Condition{{Path: "$.object_type", Op: "eq", Value: typ}}}
	}
	obj := func(typ string, fields map[string]string) core.Effect {
		return core.Effect{Object: core.ObjectTemplate{Type: typ, Fields: fields}}
	}
	rule("register-item", 10, core.RuleSpec{Match: core.Match{EventType: "item.created"},
		Effect: obj("item", map[string]string{"sku": "=$.sku"})})
	rule("book-receipt", 100, core.RuleSpec{Match: core.Match{EventType: "goods.received"},
		Effect: obj("goods_receipt", map[string]string{"item": "=ref(item, sku, $.item)", "qty": "=$.qty"})})
	// The cascade the model kept reaching for: the movement copies the
	// receipt's item link and points at the receipt itself.
	rule("move-stock-in", 200, core.RuleSpec{Match: materialized("goods_receipt"),
		Effect: obj("stock_movement", map[string]string{
			"item": "=$.state.item", "qty": "=$.state.qty", "receipt": "=$.object_id"})})
	rule("book-complaint", 100, core.RuleSpec{Match: core.Match{EventType: "complaint.registered"},
		Effect: obj("complaint", map[string]string{"number": "=$.number"})})
	rule("raise-case", 200, core.RuleSpec{Match: materialized("complaint"),
		Effect: obj("case", map[string]string{"complaint": "=$.object_id", "status": "open"})})
	// The reading side: a withdrawal names only the complaint's number, and
	// the case is found one step through its link.
	rule("close-case-on-withdrawal", 300, core.RuleSpec{Match: core.Match{EventType: "complaint.withdrawn"},
		Effect: core.Effect{Amend: &core.AmendTemplate{Type: "case",
			Target: "=ref(case, complaint, ref(complaint, number, $.number))",
			Set:    map[string]string{"status": "closed"}}}})

	submit := func(typ, payload string) (int, []error) {
		t.Helper()
		var p map[string]any
		json.Unmarshal([]byte(payload), &p)
		tr, err := x.TriggerActivity(ctx, "submit_event", map[string]any{
			"event_type": typ, "occurred_at": "2026-10-08", "payload": p}, "test", "")
		if err != nil {
			t.Fatal(err)
		}
		return tr.Booked, tr.Errors
	}
	submit("item.created", `{"sku":"WID-1"}`)
	if n, errs := submit("goods.received", `{"item":"WID-1","qty":25}`); n != 1 || len(errs) != 0 {
		t.Fatalf("receipt: %d %v", n, errs)
	}
	if n, errs := submit("complaint.registered", `{"number":"R-1"}`); n != 1 || len(errs) != 0 {
		t.Fatalf("complaint: %d %v", n, errs)
	}

	items, _ := s.ObjectsByType(ctx, "item")
	receipts, _ := s.ObjectsByType(ctx, "goods_receipt")
	moves, _ := s.ObjectsByType(ctx, "stock_movement")
	if len(moves) != 1 || moves[0].State["item"] != items[0].ID || moves[0].State["receipt"] != receipts[0].ID {
		t.Fatalf("movement = %+v", moves)
	}
	complaints, _ := s.ObjectsByType(ctx, "complaint")
	cases, _ := s.ObjectsByType(ctx, "case")
	if len(cases) != 1 || cases[0].State["complaint"] != complaints[0].ID {
		t.Fatalf("case = %+v — it must point at exactly its complaint", cases)
	}
	if n, errs := submit("complaint.withdrawn", `{"number":"R-1"}`); n != 1 || len(errs) != 0 {
		t.Fatalf("withdrawal: %d %v", n, errs)
	}
	if c, _ := s.GetObject(ctx, cases[0].ID); c.State["status"] != "closed" {
		t.Fatalf("case not closed through its link: %v", c.State)
	}

	// Guard rails. A carried id of the wrong kind is refused: a complaint is
	// not an item. The event waits; nothing half-books.
	rule("mislink", 300, core.RuleSpec{Match: core.Match{EventType: "complaint.escalated"},
		Effect: obj("stock_movement", map[string]string{"item": "=$.complaint", "qty": "=$.qty", "receipt": "=$.receipt"})})
	if _, errs := submit("complaint.escalated",
		`{"complaint":"`+complaints[0].ID+`","qty":1,"receipt":"`+receipts[0].ID+`"}`); len(errs) != 1 ||
		!strings.Contains(errs[0].Error(), "is not a item id") {
		t.Fatalf("wrong kind accepted: %v", errs)
	}
	// An id of the right kind that names nothing is refused too.
	rule("ghost-case", 300, core.RuleSpec{Match: core.Match{EventType: "complaint.reopened"},
		Effect: obj("case", map[string]string{"complaint": "=$.complaint", "status": "open"})})
	if _, errs := submit("complaint.reopened", `{"complaint":"complaint-999"}`); len(errs) == 0 ||
		!strings.Contains(errs[len(errs)-1].Error(), `no complaint "complaint-999"`) {
		t.Fatalf("missing object accepted: %v", errs)
	}
	// A literal id is never a link: a rule cannot hard-code what it points at.
	caseType := core.ObjectType{Name: "case", Fields: []core.FieldDef{{Name: "complaint", Type: "ref<complaint>", Required: true}}}
	if err := (core.RuleSpec{Match: core.Match{EventType: "x"},
		Effect: obj("case", map[string]string{"complaint": "complaint-1"})}).Validate(&caseType); err == nil {
		t.Fatal("literal id accepted as a link")
	}
	assertReplayIdentical(t, x)
}
