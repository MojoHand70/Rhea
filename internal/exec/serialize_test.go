package exec_test

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

// The executor runs one chain at a time, across handlers and processes: an
// amend-only chain has no object key to collide on, so without the lock two
// passes over one pending event both book an amendment and the open amount
// moves twice. Many passes at once over one withdrawal leave one amendment,
// and more passes than pool connections do not deadlock.
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

	// More callers than the pool has connections: each used to hold one
	// while waiting on Postgres, and the holder starved.
	var wg sync.WaitGroup
	for i := 0; i < 2*runtime.NumCPU()+8; i++ {
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

// One approved rule cannot stall the kernel: a rule on its own type with an
// each over three items triples every generation. The chain is refused by
// size, naming the rule, long before the depth cap and within seconds —
// the review clocked the depth-only cap at twenty-eight seconds under the
// executor's lock, and more items would have exhausted memory.
func TestCascadeBoundedBySize(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	if err := s.InsertObjectType(ctx, core.ObjectType{Name: "thing", Version: 1, Domain: "work", LabelField: "name",
		Fields: []core.FieldDef{{Name: "name", Type: "string", Required: true}}}); err != nil {
		t.Fatal(err)
	}
	for _, r := range []struct {
		id   string
		spec core.RuleSpec
	}{
		{"first-thing", core.RuleSpec{Match: core.Match{EventType: "thing.named"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "thing", Fields: map[string]string{"name": "=$.name"}}}}},
		{"fan-out", core.RuleSpec{Match: core.Match{EventType: core.EventObjectMaterialized,
			Where: []core.Condition{{Path: "$.object_type", Op: "eq", Value: "thing"}}},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "thing", Each: "=$.root.items[*]",
				Fields: map[string]string{"name": "=$.line.name"}}}}},
	} {
		if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: r.id, Status: core.StatusActive, Priority: 100,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: r.id, Spec: r.spec}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: "thing.named", OccurredAt: "2026-09-01",
		Payload: []byte(`{"name":"a","items":[{"name":"p"},{"name":"q"},{"name":"r"}]}`), DedupKey: "t-1"}); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	n, errs := x.ProcessPending(ctx)
	if n != 0 || len(errs) != 1 || !strings.Contains(errs[0].Error(), "fan-out") || !strings.Contains(errs[0].Error(), "consequences") {
		t.Fatalf("booked %d, errs %v", n, errs)
	}
	if time.Since(start) > 20*time.Second {
		t.Fatalf("the size cap took %s", time.Since(start))
	}
}

// A wait for the executor ends (review 2026-10-21, G5): the in-process half
// of the lock was a plain mutex — a caller queued behind a long run (a
// replay, a backfill, a slow chain) waited there whatever became of its
// request, and requests piled up while every door froze. Now the wait
// gives way to the caller's context, and an executor given a limit
// answers "busy" when it passes.
func TestAWaitForTheExecutorEnds(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	// Someone else holds the books: the lock taken from another session.
	held, err := s.Pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := held.Exec(ctx, `SELECT pg_advisory_lock($1)`, 0x52484541); err != nil {
		t.Fatal(err)
	}
	var freed sync.Once
	free := func() {
		freed.Do(func() {
			_, _ = held.Exec(ctx, `SELECT pg_advisory_unlock($1)`, 0x52484541)
			held.Release()
		})
	}
	t.Cleanup(free) // a failing test frees the books before its database goes
	// One caller waits on that session's lock, holding the process's turn.
	first := make(chan error, 1)
	go func() {
		_, errs := x.ProcessPending(ctx)
		first <- errors.Join(errs...)
	}()
	time.Sleep(200 * time.Millisecond)

	// A caller whose request is gone stops waiting.
	gone, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, errs := x.ProcessPending(gone)
		done <- errors.Join(errs...)
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("a caller behind a held lock booked")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("a caller whose request was gone still waited")
	}
	// An executor with a limit answers busy when it passes.
	limited := &exec.Executor{Store: s, LockWait: 300 * time.Millisecond}
	if _, err := limited.TriggerActivity(ctx, "submit_event", map[string]any{"event_type": "note.taken", "occurred_at": "2026-10-22", "payload": map[string]any{}}, "test", ""); err == nil || !strings.Contains(err.Error(), "busy") {
		t.Fatalf("past its limit: %v", err)
	}
	// The books freed, the first caller goes on.
	free()
	if err := <-first; err != nil {
		t.Fatalf("the first caller: %v", err)
	}
}
