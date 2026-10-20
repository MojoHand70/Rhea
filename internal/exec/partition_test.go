package exec_test

import (
	"context"
	"reflect"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

// The partition key (DIRECTION 2026-10-08, "Scale and security are welded",
// scale rule 1 and security rule 2): the one item cheap now and expensive
// later. Every fact carries the company it belongs to — stamped from the
// caller's scope at the door, inherited by everything derived from it, kept
// by replay — and a scoped read sees one company's facts alone, an unscoped
// read the whole log. Two companies in one log, each seeing only its own.
func TestPartitionKey(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	if err := s.InsertObjectType(ctx, core.ObjectType{Name: "note", Version: 1, Domain: "work", LabelField: "text",
		Fields: []core.FieldDef{{Name: "text", Type: "string", Required: true}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: "note", Status: core.StatusActive, Priority: 100,
		EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: "notes", Spec: core.RuleSpec{
			Match:  core.Match{EventType: "note.taken"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text"}}}}}); err != nil {
		t.Fatal(err)
	}
	alfa, beta := store.WithCompany(ctx, "PL5250001111"), store.WithCompany(ctx, "PL5260001246")
	note := func(scope context.Context, text string) {
		t.Helper()
		if _, err := x.TriggerActivity(scope, "submit_event", map[string]any{
			"event_type": "note.taken", "occurred_at": "2026-09-01", "payload": map[string]any{"text": text}}, "test", ""); err != nil {
			t.Fatal(err)
		}
	}
	note(alfa, "alfa 1")
	note(beta, "beta 1")
	note(alfa, "alfa 2")

	// The door stamped each fact with its company; the chain inherited it.
	raws, _ := s.EventsByKind(ctx, core.KindRaw)
	derived, _ := s.EventsByKind(ctx, core.KindDerived)
	byCompany := map[string]int{}
	for _, ev := range append(raws, derived...) {
		if ev.Type == "note.taken" || ev.Type == core.EventObjectMaterialized {
			byCompany[ev.Company]++
		}
	}
	if byCompany["PL5250001111"] != 4 || byCompany["PL5260001246"] != 2 || byCompany[""] != 0 {
		t.Fatalf("facts by company: %v", byCompany)
	}

	// A scoped read sees one company's facts alone; an unscoped read, the log.
	texts := func(scope context.Context) []string {
		t.Helper()
		objs, err := s.ObjectsByType(scope, "note")
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, o := range objs {
			out = append(out, o.State["text"].(string))
			if o.Company == "" {
				t.Fatalf("%s carries no company", o.ID)
			}
		}
		return out
	}
	if !reflect.DeepEqual(texts(alfa), []string{"alfa 1", "alfa 2"}) || !reflect.DeepEqual(texts(beta), []string{"beta 1"}) || len(texts(ctx)) != 3 {
		t.Fatalf("alfa %v, beta %v, all %v", texts(alfa), texts(beta), texts(ctx))
	}
	if ids, _ := s.FindObjectIDsByField(beta, "note", "text", "alfa 1"); len(ids) != 0 {
		t.Fatalf("beta can see alfa's note: %v", ids)
	}
	if counts, _ := s.CountObjectsByType(alfa); counts["note"] != 2 {
		t.Fatalf("alfa's count: %v", counts)
	}
	// beta's facts, beside the installation's own (the kernel's seeded
	// verbs carry no company and are every scope's): one note.
	evs, _ := s.EventsByKind(beta, core.KindRaw)
	betas := 0
	for _, ev := range evs {
		if ev.Company == "PL5260001246" {
			betas++
		} else if ev.Company != "" {
			t.Fatalf("beta's facts hold %s's", ev.Company)
		}
	}
	if betas != 1 {
		t.Fatalf("beta's facts: %d", betas)
	}

	// The worklist is scoped too: a fact no rule explains waits in its
	// company's worklist, not the other's.
	if _, err := x.TriggerActivity(beta, "submit_event", map[string]any{
		"event_type": "memo.left", "occurred_at": "2026-09-02", "payload": map[string]any{"text": "?"}}, "test", ""); err != nil {
		t.Fatal(err)
	}
	if left, _ := s.UnmatchedRawEvents(alfa); len(left) != 0 {
		t.Fatalf("alfa's worklist holds beta's memo: %v", left)
	}
	if left, _ := s.UnmatchedRawEvents(beta); len(left) != 1 {
		t.Fatalf("beta's worklist: %v", left)
	}

	// Replay keeps the key on every object.
	before, _ := s.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := s.AllObjects(ctx)
	if !reflect.DeepEqual(before, after) || after[0].Company == "" {
		t.Fatal("replay lost the partition key")
	}
	// Replay is the whole log's whatever scope it runs under: it wipes the
	// one cache every company shares, so it rebuilds every company's.
	if _, err := x.Replay(alfa); err != nil {
		t.Fatal(err)
	}
	again, _ := s.AllObjects(ctx)
	if diffs := exec.StateDiff(before, again); len(diffs) != 0 {
		t.Fatalf("replay under one company's scope lost the others: %v", diffs)
	}
	if diffs := exec.StateDiff(before, before[:1]); len(diffs) != 2 {
		t.Fatalf("StateDiff misses lost objects: %v", diffs)
	}
}
