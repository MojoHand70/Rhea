package exec_test

import (
	"context"
	"encoding/json"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

// TestTheExecutorRemembersWhatWaits (2026-10-27, the 100,000-invoice
// measurement): the executor remembers which facts still wait for a rule
// and reads only what the log gained since, instead of the whole log after
// every fact. What it remembers must never be wrong: a waiting fact another
// process books is not booked again here; a fact another process records
// without booking it (as the clock records a day) is found by the next pass
// here; and a company's pass books none of another company's waiting facts.
func TestTheExecutorRemembersWhatWaits(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	// another process: connections of its own to the same database
	s2, err := store.Open(ctx, s.Pool.Config().ConnString())
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()
	cli := &exec.Executor{Store: s2}
	if err := s.InsertObjectType(ctx, core.ObjectType{Name: "note", Version: 1, Domain: "work", LabelField: "text",
		Fields: []core.FieldDef{{Name: "text", Type: "string", Required: true}}}); err != nil {
		t.Fatal(err)
	}
	a, b := store.WithCompany(ctx, "PL5250001111"), store.WithCompany(ctx, "PL5260001246")
	submit := func(x *exec.Executor, at context.Context, typ, text string) exec.Triggered {
		t.Helper()
		tr, err := x.TriggerActivity(at, "submit_event", map[string]any{"event_type": typ, "occurred_at": "2026-10-27",
			"payload": map[string]any{"text": text}}, "test", text)
		if err != nil || len(tr.Errors) != 0 {
			t.Fatalf("%s: %+v %v", text, tr, err)
		}
		return tr
	}
	rule := func(id, typ, status string) {
		t.Helper()
		if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: id, Status: status, Priority: 100, EffectiveFrom: "2026-01-01", CreatedBy: "test",
			Description: id, Spec: core.RuleSpec{Match: core.Match{EventType: typ},
				Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text"}}}}}); err != nil {
			t.Fatal(err)
		}
	}
	explanations := func(ev int64) int {
		t.Helper()
		got, err := s.Explained(ctx, []int64{ev})
		if err != nil {
			t.Fatal(err)
		}
		chain, err := s.ChainOf(ctx, ev)
		if err != nil {
			t.Fatal(err)
		}
		if !got[ev] {
			return 0
		}
		return len(chain)
	}

	// No rule for memos yet: the first waits, and this process remembers it.
	first := submit(x, a, "memo.taken", "first")
	if first.Booked != 0 {
		t.Fatalf("a memo with no rule booked: %+v", first)
	}
	// The other process is taught the rule, and its pass books the memo.
	rule("memo-note", "memo.taken", core.StatusDraft)
	if _, _, errs, err := cli.ApproveRule(ctx, "memo-note", "krzysztof", "2026-10-27"); err != nil || len(errs) != 0 {
		t.Fatalf("approve: %v %v", err, errs)
	}
	if n := explanations(first.EventID); n != 1 {
		t.Fatalf("the other process booked the first memo: %d consequences", n)
	}
	// This process's next pass books its own memo and not the first again.
	if second := submit(x, a, "memo.taken", "second"); second.Booked != 1 || explanations(first.EventID) != 1 {
		t.Fatalf("booked again what another process booked: %+v, first memo's consequences %d", second, explanations(first.EventID))
	}
	// The other process records a memo without booking it, as the clock
	// records a day; this process's next pass finds it.
	third, err := s2.AppendEvent(a, core.Event{Kind: core.KindRaw, Type: "memo.taken", OccurredAt: "2026-10-27",
		Payload: json.RawMessage(`{"text":"third"}`), DedupKey: "third"})
	if err != nil {
		t.Fatal(err)
	}
	if fourth := submit(x, a, "memo.taken", "fourth"); fourth.Booked != 2 || explanations(third) != 1 {
		t.Fatalf("the fact the other process recorded: %+v, its consequences %d", fourth, explanations(third))
	}

	// A parcel of company A waits; its rule arrives with no pass of its own.
	parcel := submit(x, a, "parcel.arrived", "parcel")
	rule("parcel-note", "parcel.arrived", core.StatusActive)
	// Company B's pass books none of A's waiting facts...
	submit(x, b, "memo.taken", "b-memo")
	if n := explanations(parcel.EventID); n != 0 {
		t.Fatalf("company B's pass booked company A's parcel: %d consequences", n)
	}
	// ...A's next pass books it.
	if tr := submit(x, a, "memo.taken", "fifth"); tr.Booked != 2 || explanations(parcel.EventID) != 1 {
		t.Fatalf("A's next pass: %+v, the parcel's consequences %d", tr, explanations(parcel.EventID))
	}

	// An accounting office sees both clients and acts for one at a time:
	// what it remembers acting for A is not what it books acting for B.
	office := store.WithScope(ctx, []string{"PL5250001111", "PL5260001246"})
	forA, forB := store.ActFor(office, "PL5250001111"), store.ActFor(office, "PL5260001246")
	box := submit(x, forA, "box.delivered", "box")
	rule("box-note", "box.delivered", core.StatusActive)
	submit(x, forB, "memo.taken", "office-b-memo")
	if n := explanations(box.EventID); n != 0 {
		t.Fatalf("the office acting for B booked A's box: %d consequences", n)
	}
	if tr := submit(x, forA, "memo.taken", "office-a-memo"); tr.Booked != 2 || explanations(box.EventID) != 1 {
		t.Fatalf("the office acting for A: %+v, the box's consequences %d", tr, explanations(box.EventID))
	}
	assertReplayIdentical(t, x)
}
