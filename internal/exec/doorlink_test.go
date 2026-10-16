package exec_test

import (
	"context"
	"encoding/json"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

// A door's ref input reads as a link (DECISIONS 2026-10-16): the door vouches
// that the object it names exists and is of the declared type, so a rule on
// the door's fact may read through it — a person says "this bank line is the
// owner's deposit" and the rule books the line's own amount, in a condition
// as in an effect. A raw fact that came through no door, or through the
// open door, reads as before; replay reproduces what was read.
func TestADoorsRefInputReadsAsALink(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	for _, ot := range []core.ObjectType{
		{Name: "line", Version: 1, Domain: "finance", LabelField: "title",
			Fields: []core.FieldDef{{Name: "title", Type: "string", Required: true}, {Name: "amount", Type: "money", Required: true},
				{Name: "direction", Type: "enum", Values: []string{"credit", "debit"}, Required: true}}},
		{Name: "owner_move", Version: 1, Domain: "finance", LabelField: "title",
			Fields: []core.FieldDef{{Name: "title", Type: "string", Required: true}, {Name: "amount", Type: "money", Required: true}}},
	} {
		if err := x.Store.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := x.Store.InsertActivityVersion(ctx, core.Activity{
		Name: "say_owner_move", Status: core.StatusActive, Domain: "finance", CreatedBy: "test",
		Description: "a person says a bank line is the owner's own money",
		Spec: core.ActivitySpec{
			Inputs: []core.FieldDef{{Name: "line", Type: "ref<line>", Required: true}, {Name: "occurred_at", Type: "date"}},
			Emits:  "line.owner_move.said", Who: []string{"human"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	activateRule(t, x, "record-line", 100, core.RuleSpec{Match: core.Match{EventType: "line.recorded"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "line",
			Fields: map[string]string{"title": "=$.title", "amount": "=$.amount", "direction": "=$.direction"}}}})
	// Only a deposit (a credit) is the owner's money in; the amount is the
	// line's own.
	activateRule(t, x, "owner-deposit", 100, core.RuleSpec{
		Match: core.Match{EventType: "line.owner_move.said", Where: []core.Condition{{Formula: `=$.line.direction = "credit"`}}},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "owner_move",
			Fields: map[string]string{"title": "=$.line.title", "amount": "=$.line.amount"}}}})

	submitRaw(t, x, "l1", "line.recorded", "2026-10-01", `{"title":"wpłata własna","amount":"500.00","direction":"credit"}`)
	submitRaw(t, x, "l2", "line.recorded", "2026-10-01", `{"title":"przelew","amount":"80.00","direction":"debit"}`)
	if n, errs := x.ProcessPending(ctx); n != 2 || len(errs) != 0 {
		t.Fatalf("lines: %d %v", n, errs)
	}
	lines, _ := x.Store.ObjectsByType(ctx, "line")
	tr, err := x.TriggerActivity(ctx, "say_owner_move", map[string]any{"line": lines[0].ID, "occurred_at": "2026-10-02"}, "krzysztof", "")
	if err != nil || tr.Booked != 1 || len(tr.Errors) != 0 {
		t.Fatalf("the deposit: %+v %v", tr, err)
	}
	moves, _ := x.Store.ObjectsByType(ctx, "owner_move")
	if len(moves) != 1 || moves[0].State["title"] != "wpłata własna" || moves[0].State["amount"] != float64(50000) {
		t.Fatalf("the owner's move: %+v", moves)
	}
	// The debit is not a deposit: the condition reads its direction and
	// nothing matches.
	if tr, err := x.TriggerActivity(ctx, "say_owner_move", map[string]any{"line": lines[1].ID, "occurred_at": "2026-10-02"}, "krzysztof", ""); err != nil || tr.Booked != 0 {
		t.Fatalf("the debit: %+v %v", tr, err)
	}

	before, _ := x.Store.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := x.Store.AllObjects(ctx)
	bj, _ := json.Marshal(before)
	aj, _ := json.Marshal(after)
	if string(bj) != string(aj) {
		t.Fatalf("replay diverged")
	}
}
