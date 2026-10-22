package exec_test

import (
	"bytes"
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

// A chain runs under its event's company whoever runs the executor: two
// companies each with complaint C-1, and company B's withdrawal — run by
// the owner, as the clock and the CLI run — closes B's complaint and
// never A's; their derived facts carry their own companies.
func TestChainsRunUnderTheEventsCompany(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	lc := &core.LifecycleDef{Field: "status", Transitions: map[string][]string{"open": {"closed"}}}
	if err := s.InsertObjectType(ctx, core.ObjectType{Name: "complaint", Version: 1, Domain: "sales", LabelField: "number",
		Fields: []core.FieldDef{{Name: "number", Type: "string", Required: true},
			{Name: "status", Type: "enum", Values: []string{"open", "closed"}, Required: true}}, Lifecycle: lc}); err != nil {
		t.Fatal(err)
	}
	for _, r := range []struct {
		id   string
		spec core.RuleSpec
	}{
		{"book-complaint", core.RuleSpec{Match: core.Match{EventType: "complaint.registered"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "complaint", Fields: map[string]string{"number": "=$.number", "status": "open"}}}}},
		{"withdraw-complaint", core.RuleSpec{Match: core.Match{EventType: "complaint.withdrawn"},
			Effect: core.Effect{Amend: &core.AmendTemplate{Type: "complaint", Target: "=ref(complaint, number, $.number)", Set: map[string]string{"status": "closed"}}}}},
	} {
		if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: r.id, Status: core.StatusActive, Priority: 100,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: r.id, Spec: r.spec}); err != nil {
			t.Fatal(err)
		}
	}
	alfa, beta := store.WithCompany(ctx, "PL5250001111"), store.WithCompany(ctx, "PL5260001246")
	for _, c := range []context.Context{alfa, beta} {
		if _, err := s.AppendEvent(c, core.Event{Kind: core.KindRaw, Type: "complaint.registered", OccurredAt: "2026-09-01",
			Payload: []byte(`{"number":"C-1"}`), DedupKey: "c-1"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.AppendEvent(beta, core.Event{Kind: core.KindRaw, Type: "complaint.withdrawn", OccurredAt: "2026-09-02",
		Payload: []byte(`{"number":"C-1"}`), DedupKey: "c-1-w"}); err != nil {
		t.Fatal(err)
	}
	// the owner runs everything, as the clock does
	if n, errs := x.ProcessPending(store.WithOwner(ctx)); n != 3 || len(errs) != 0 {
		t.Fatalf("owner's pass: fired %d, %v", n, errs)
	}
	status := func(c context.Context) string {
		t.Helper()
		objs, err := s.ObjectsByType(c, "complaint")
		if err != nil || len(objs) != 1 {
			t.Fatalf("complaints: %v %v", objs, err)
		}
		return objs[0].State["status"].(string)
	}
	if status(alfa) != "open" || status(beta) != "closed" {
		t.Fatalf("alfa %s, beta %s", status(alfa), status(beta))
	}
	derived, _ := s.EventsByKind(ctx, core.KindDerived)
	for _, ev := range derived {
		if ev.Company == "" {
			t.Fatalf("a derived fact with no company: %+v", ev)
		}
	}
	assertReplayIdentical(t, x)
}

// A company's own rule books that company's facts and no other's (KK,
// 2026-10-22: company rules and Rhea's global rules). Company alfa has a rule
// for memos; beta's memo — run by the owner too, as the clock and the CLI
// run — still waits in beta's worklist, while Rhea's rule for notes books for
// both. Beta's scope does not see alfa's rule, the installation's sees only
// Rhea's, a rule's id never changes company, and replay keeps it all.
func TestACompanyRuleBooksOnlyItsCompany(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	if err := s.InsertObjectType(ctx, core.ObjectType{Name: "note", Version: 1, Domain: "work", LabelField: "text",
		Fields: []core.FieldDef{{Name: "text", Type: "string", Required: true}}}); err != nil {
		t.Fatal(err)
	}
	const alfaKey, betaKey = "PL5250001111", "PL5260001246"
	rule := func(id, company, eventType string) core.Rule {
		return core.Rule{ID: id, Company: company, Status: core.StatusActive, Priority: 100,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: core.RuleSpec{
				Match:  core.Match{EventType: eventType},
				Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text"}}}}}
	}
	for _, r := range []core.Rule{rule("note", "", "note.taken"), rule("memo", alfaKey, "memo.left")} {
		if _, err := s.InsertRuleVersion(ctx, r); err != nil {
			t.Fatal(err)
		}
	}
	alfa, beta := store.WithCompany(ctx, alfaKey), store.WithCompany(ctx, betaKey)
	submit := func(scope context.Context, eventType, text string) {
		t.Helper()
		if _, err := x.TriggerActivity(scope, "submit_event", map[string]any{
			"event_type": eventType, "occurred_at": "2026-09-01", "payload": map[string]any{"text": text}}, "test", ""); err != nil {
			t.Fatal(err)
		}
	}
	submit(alfa, "memo.left", "alfa's memo")
	submit(beta, "memo.left", "beta's memo")
	submit(beta, "note.taken", "beta's note")
	if _, errs := x.ProcessPending(ctx); len(errs) != 0 {
		t.Fatal(errs)
	}

	texts := func(scope context.Context) []string {
		t.Helper()
		objs, err := s.ObjectsByType(scope, "note")
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, o := range objs {
			out = append(out, o.State["text"].(string))
		}
		return out
	}
	if !reflect.DeepEqual(texts(alfa), []string{"alfa's memo"}) || !reflect.DeepEqual(texts(beta), []string{"beta's note"}) {
		t.Fatalf("alfa %v, beta %v", texts(alfa), texts(beta))
	}
	if left, _ := s.UnmatchedRawEvents(beta); len(left) != 1 {
		t.Fatalf("beta's memo should wait for a rule of beta's own: %v", left)
	}

	ids := func(scope context.Context) []string {
		t.Helper()
		rules, err := s.ActiveRules(scope)
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, r := range rules {
			out = append(out, r.ID+"@"+r.Company)
		}
		return out
	}
	if got := ids(alfa); !reflect.DeepEqual(got, []string{"memo@" + alfaKey, "note@"}) {
		t.Fatalf("alfa's rules: %v", got)
	}
	if got := ids(beta); !reflect.DeepEqual(got, []string{"note@"}) {
		t.Fatalf("beta sees: %v", got)
	}
	if got := ids(store.WithInstallation(ctx)); !reflect.DeepEqual(got, []string{"note@"}) {
		t.Fatalf("the installation sees: %v", got)
	}
	if _, err := s.GetRule(beta, "memo"); err == nil {
		t.Fatal("beta reads alfa's rule")
	}

	// A rule's id keeps its company: no other company's version of alfa's
	// rule, no company's version of Rhea's; and beta writes no rule of alfa's.
	if _, err := s.InsertRuleVersion(ctx, rule("memo", betaKey, "memo.left")); err == nil {
		t.Fatal("alfa's rule took a version of beta's")
	}
	if _, err := s.InsertRuleVersion(ctx, rule("note", alfaKey, "note.taken")); err == nil {
		t.Fatal("Rhea's rule took a version of alfa's")
	}
	if _, err := s.InsertRuleVersion(beta, rule("memo-2", alfaKey, "memo.left")); err == nil {
		t.Fatal("beta wrote a rule of alfa's")
	}
	assertReplayIdentical(t, x)
	if got := texts(beta); !reflect.DeepEqual(got, []string{"beta's note"}) {
		t.Fatalf("replay booked beta's memo by alfa's rule: %v", got)
	}
	// A backup carries each rule's company: restored, alfa's rule is alfa's
	// still, never Rhea's for every company.
	var buf bytes.Buffer
	if _, err := s.Backup(ctx, &buf); err != nil {
		t.Fatal(err)
	}
	restored := storetest.NewBare(t)
	if _, err := restored.Restore(ctx, &buf); err != nil {
		t.Fatal(err)
	}
	if r, err := restored.GetRule(ctx, "memo"); err != nil || r.Company != alfaKey {
		t.Fatalf("restored memo rule: %+v, %v", r, err)
	}
}
