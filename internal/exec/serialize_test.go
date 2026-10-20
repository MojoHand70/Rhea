package exec_test

import (
	"context"
	"sync"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

// The executor runs one chain at a time, across handlers and processes: an
// amend-only chain has no object key to collide on, so without the lock two
// passes over one pending event both book an amendment and the open amount
// moves twice. Eight passes at once over one withdrawal leave one amendment.
func TestExecutorSerialized(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	if err := s.InsertObjectType(ctx, core.ObjectType{Name: "complaint", Version: 1, Domain: "work", LabelField: "number",
		Fields: []core.FieldDef{{Name: "number", Type: "string", Required: true},
			{Name: "status", Type: "enum", Values: []string{"open", "closed"}, Required: true}},
		Lifecycle: &core.LifecycleDef{Field: "status", Transitions: map[string][]string{"open": {"closed"}}}}); err != nil {
		t.Fatal(err)
	}
	for _, r := range []struct {
		id   string
		spec core.RuleSpec
	}{
		{"book-complaint", core.RuleSpec{Match: core.Match{EventType: "complaint.registered"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "complaint",
				Fields: map[string]string{"number": "=$.number", "status": "open"}}}}},
		{"withdraw-complaint", core.RuleSpec{Match: core.Match{EventType: "complaint.withdrawn"},
			Effect: core.Effect{Amend: &core.AmendTemplate{Type: "complaint",
				Target: "=ref(complaint, number, $.number)", Set: map[string]string{"status": "closed"}}}}},
	} {
		if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: r.id, Status: core.StatusActive, Priority: 100,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: r.id, Spec: r.spec}); err != nil {
			t.Fatal(err)
		}
	}
	submit := func(dedup, typ, payload string) {
		t.Helper()
		if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: typ, OccurredAt: "2026-09-01",
			Payload: []byte(payload), DedupKey: dedup}); err != nil {
			t.Fatal(err)
		}
	}
	submit("c-1", "complaint.registered", `{"number":"C-1"}`)
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("registered: booked %d, %v", n, errs)
	}
	submit("c-1-w", "complaint.withdrawn", `{"number":"C-1"}`)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, errs := x.ProcessPending(ctx); len(errs) != 0 {
				t.Errorf("pass: %v", errs)
			}
		}()
	}
	wg.Wait()
	derived, _ := s.EventsByKind(ctx, core.KindDerived)
	amended := 0
	for _, ev := range derived {
		if ev.Type == core.EventObjectAmended {
			amended++
		}
	}
	if amended != 1 {
		t.Fatalf("one withdrawal booked %d amendments", amended)
	}
	// A trigger's own worklist pass runs under the lock it already holds.
	if _, err := x.TriggerActivity(ctx, "submit_event", map[string]any{
		"event_type": "complaint.registered", "occurred_at": "2026-09-02", "payload": map[string]any{"number": "C-2"}}, "test", ""); err != nil {
		t.Fatal(err)
	}
	if ids, _ := s.FindObjectIDsByField(ctx, "complaint", "number", "C-2"); len(ids) != 1 {
		t.Fatalf("C-2 after a trigger under the lock: %v", ids)
	}
}
