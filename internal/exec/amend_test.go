package exec_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

// The amendment's admission tests (DECISIONS 2026-10-05): earned by the
// cases attempt's fail-as-data proof, kept honest here by its invariant —
// amendment is consent-based (the type declares its life) and every move of
// the lifecycle field runs along declared transitions — and by the
// determinism proof: replay re-applies baked deltas and reproduces state.

func seedLifecycleWorld(t *testing.T, x *exec.Executor) {
	t.Helper()
	ctx := context.Background()
	for _, ot := range []core.ObjectType{
		{
			Name: "complaint", Version: 1, Domain: "crm", IsDocument: true,
			LabelField: "customer",
			Fields: []core.FieldDef{
				{Name: "customer", Type: "string", Required: true},
				{Name: "registered_on", Type: "date", Required: true},
			},
		},
		{
			Name: "case", Version: 1, Domain: "work", LabelField: "title",
			Fields: []core.FieldDef{
				{Name: "kind", Type: "string", Required: true},
				{Name: "subject_type", Type: "string", Required: true},
				{Name: "subject_id", Type: "string", Required: true},
				{Name: "title", Type: "string", Required: true},
				{Name: "status", Type: "enum", Required: true, Values: []string{"open", "resolved"}},
				{Name: "resolution", Type: "string"},
			},
			// The consent: this type has a life, and it moves one way.
			Lifecycle: &core.LifecycleDef{
				Field:       "status",
				Transitions: map[string][]string{"open": {"resolved"}},
			},
			// ...and the one field besides its status a rule may set later.
			Amendable: []string{"resolution"},
		},
	} {
		if err := x.Store.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	activateRule(t, x, "book-complaint", 100, core.RuleSpec{
		Match: core.Match{EventType: "complaint.registered"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "complaint",
			Fields: map[string]string{
				"customer": "=$.customer", "registered_on": "=$.registered_on",
			}}},
	})
	activateRule(t, x, "raise-complaint-case", 200, core.RuleSpec{
		Match: core.Match{EventType: core.EventObjectMaterialized, Where: []core.Condition{
			{Path: "$.object_type", Op: "eq", Value: "complaint"},
		}},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "case",
			Fields: map[string]string{
				"kind": "complaint_followup", "subject_type": "complaint",
				"subject_id": "=$.object_id", "title": "=$.state.customer",
				"status": "open",
			}}},
	})
}

var resolveCaseSpec = core.RuleSpec{
	Match: core.Match{EventType: "complaint.resolved"},
	Effect: core.Effect{Amend: &core.AmendTemplate{
		Type:   "case",
		Target: "=ref(case, subject_id, $.complaint)",
		Set: map[string]string{
			"status":     "resolved",
			"resolution": "=$.resolution",
		},
	}},
}

func TestAmendmentMovesDeclaredLifecycle(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	seedLifecycleWorld(t, x)
	activateRule(t, x, "resolve-case", 300, resolveCaseSpec)

	submitRaw(t, x, "c-1", "complaint.registered", "2026-10-05",
		`{"customer":"ACME Sp. z o.o.","registered_on":"2026-10-05"}`)
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("booked %d, errs %v", n, errs)
	}
	cases, _ := x.Store.ObjectsByType(ctx, "case")
	complaints, _ := x.Store.ObjectsByType(ctx, "complaint")
	caseID := cases[0].ID

	resID := submitRaw(t, x, "r-1", "complaint.resolved", "2026-10-06",
		fmt.Sprintf(`{"complaint":"%s","resolution":"credited the difference"}`, complaints[0].ID))
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("resolution: booked %d, errs %v", n, errs)
	}

	// The case moved — one case, status resolved, the resolution carried.
	cases, _ = x.Store.ObjectsByType(ctx, "case")
	if len(cases) != 1 || cases[0].State["status"] != "resolved" ||
		cases[0].State["resolution"] != "credited the difference" {
		t.Fatalf("case = %+v", cases[0].State)
	}
	// The move is in the log with full provenance: an object.amended derived
	// event, caused by the resolution, naming the rule version, delta baked.
	derived, _ := x.Store.EventsByKind(ctx, core.KindDerived)
	var amended *core.Event
	for i := range derived {
		if derived[i].Type == core.EventObjectAmended {
			amended = &derived[i]
		}
	}
	if amended == nil || *amended.CauseEventID != resID ||
		amended.RuleID != "resolve-case" || amended.RuleVersion != 1 {
		t.Fatalf("amendment event = %+v", amended)
	}
	var am core.AmendedObject
	json.Unmarshal(amended.Payload, &am)
	if am.ObjectID != caseID || am.Set["status"] != "resolved" {
		t.Fatalf("amendment payload = %+v", am)
	}
	// The resolution event is explained — out of the worklist.
	if wl, _ := x.Store.UnmatchedRawEvents(ctx); len(wl) != 0 {
		t.Fatalf("worklist = %v", wl)
	}

	// Determinism (invariant 4): wipe the cache, replay the log — the baked
	// deltas re-apply in order and identical state reappears.
	before, _ := x.Store.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := x.Store.AllObjects(ctx)
	bj, _ := json.Marshal(before)
	aj, _ := json.Marshal(after)
	if string(bj) != string(aj) {
		t.Fatalf("replay diverged:\n%s\n%s", bj, aj)
	}
}

// TestAmendmentLaws: the kernel refuses what the declarations do not allow —
// undeclared transitions, types without a lifecycle, missing targets (which
// wait, fail-as-data, and apply once the object arrives), ambiguous double
// amendment, and raw object.amended forgeries.
func TestAmendmentLaws(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	seedLifecycleWorld(t, x)
	activateRule(t, x, "resolve-case", 300, resolveCaseSpec)

	// Out-of-order arrival: the resolution lands before the complaint. The
	// first pass cannot find the case; the same ProcessPending's fixpoint
	// books the complaint, raises the case, then applies the amendment.
	submitRaw(t, x, "r-1", "complaint.resolved", "2026-10-06",
		`{"complaint":"__PLACEHOLDER__","resolution":"early"}`)
	// The placeholder resolves nothing; it waits as data.
	if n, errs := x.ProcessPending(ctx); n != 0 || len(errs) != 1 {
		t.Fatalf("early resolution: booked %d, errs %v", n, errs)
	}
	if wl, _ := x.Store.UnmatchedRawEvents(ctx); len(wl) != 1 {
		t.Fatalf("the unresolvable resolution should wait in the worklist")
	}

	// The orderly story, out of order: a resolution naming the real subject
	// arrives before its complaint, then the complaint. One call settles both.
	submitRaw(t, x, "c-1", "complaint.registered", "2026-10-05",
		`{"customer":"Beta GmbH","registered_on":"2026-10-05"}`)
	complaintID := "complaint-0"
	if evs, _ := x.Store.UnmatchedRawEvents(ctx); true {
		// the complaint event id is the newest unmatched one
		complaintID = fmt.Sprintf("complaint-%d", evs[len(evs)-1].ID)
	}
	submitRaw(t, x, "r-2", "complaint.resolved", "2026-10-06",
		fmt.Sprintf(`{"complaint":"%s","resolution":"replaced"}`, complaintID))
	if n, errs := x.ProcessPending(ctx); n != 2 || len(errs) != 1 {
		t.Fatalf("fixpoint: booked %d, errs %v (the placeholder keeps failing)", n, errs)
	}
	cases, _ := x.Store.ObjectsByType(ctx, "case")
	if len(cases) != 1 || cases[0].State["status"] != "resolved" {
		t.Fatalf("case = %+v", cases)
	}
	caseID := cases[0].ID

	// An undeclared transition refuses: resolved → resolved is not in the
	// transitions, so a second resolution fails as data and waits.
	submitRaw(t, x, "r-3", "complaint.resolved", "2026-10-07",
		fmt.Sprintf(`{"complaint":"%s","resolution":"again"}`, complaintID))
	_, errs := x.ProcessPending(ctx)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "not a declared transition") {
			found = true
		}
	}
	if !found {
		t.Fatalf("undeclared transition did not refuse: %v", errs)
	}
	cases, _ = x.Store.ObjectsByType(ctx, "case")
	if cases[0].State["resolution"] != "replaced" {
		t.Fatalf("a refused amendment half-applied: %+v", cases[0].State)
	}

	// Consent: amending a type with no lifecycle refuses at expansion.
	activateRule(t, x, "mutate-complaint", 400, core.RuleSpec{
		Match: core.Match{EventType: "complaint.edited"},
		Effect: core.Effect{Amend: &core.AmendTemplate{
			Type: "complaint", Target: "=$.complaint",
			Set: map[string]string{"customer": "=$.customer"},
		}},
	})
	submitRaw(t, x, "e-1", "complaint.edited", "2026-10-07",
		fmt.Sprintf(`{"complaint":"%s","customer":"Mallory"}`, complaintID))
	_, errs = x.ProcessPending(ctx)
	found = false
	for _, e := range errs {
		if strings.Contains(e.Error(), "declares no lifecycle") {
			found = true
		}
	}
	if !found {
		t.Fatalf("lifecycle consent not enforced: %v", errs)
	}

	// Ambiguity: two rules amending the same case on one event is a
	// conflict — the event refuses whole and waits for a human.
	activateRule(t, x, "resolve-case-again", 500, resolveCaseSpec)
	submitRaw(t, x, "c-2", "complaint.registered", "2026-10-08",
		`{"customer":"Gamma sp. j.","registered_on":"2026-10-08"}`)
	// Three refused events keep failing honestly: the placeholder, the
	// undeclared re-resolution, the consentless edit.
	if _, errs := x.ProcessPending(ctx); len(errs) != 3 {
		t.Fatalf("setup errs = %v", errs)
	}
	complaints, _ := x.Store.ObjectsByType(ctx, "complaint")
	var gamma string
	for _, c := range complaints {
		if c.State["customer"] == "Gamma sp. j." {
			gamma = c.ID
		}
	}
	submitRaw(t, x, "r-4", "complaint.resolved", "2026-10-08",
		fmt.Sprintf(`{"complaint":"%s","resolution":"twice?"}`, gamma))
	_, errs = x.ProcessPending(ctx)
	found = false
	for _, e := range errs {
		if strings.Contains(e.Error(), "both touch") {
			found = true
		}
	}
	if !found {
		t.Fatalf("double amendment did not conflict: %v", errs)
	}

	// Forgery: a raw object.amended is refused at the door and at append.
	if _, err := x.Store.AppendEvent(ctx, core.Event{
		Kind: core.KindRaw, Type: core.EventObjectAmended, OccurredAt: "2026-10-08",
		Payload: json.RawMessage(fmt.Sprintf(`{"object_id":"%s","set":{"status":"open"}}`, caseID)),
	}); err == nil {
		t.Fatal("raw object.amended accepted")
	}
}

// TestAmendmentSimulates: the approval gate sees the move before it is
// real — a drafted amend rule dry-runs into a field-level changed object,
// and nothing books.
func TestAmendmentSimulates(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	seedLifecycleWorld(t, x)

	submitRaw(t, x, "c-1", "complaint.registered", "2026-10-05",
		`{"customer":"ACME Sp. z o.o.","registered_on":"2026-10-05"}`)
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("booked %d, errs %v", n, errs)
	}
	complaints, _ := x.Store.ObjectsByType(ctx, "complaint")
	submitRaw(t, x, "r-1", "complaint.resolved", "2026-10-06",
		fmt.Sprintf(`{"complaint":"%s","resolution":"credited"}`, complaints[0].ID))
	x.ProcessPending(ctx) // stays unexplained: no resolve rule is active yet

	if _, err := x.Store.InsertRuleVersion(ctx, core.Rule{
		ID: "resolve-case", Status: core.StatusDraft, Priority: 300,
		EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: "resolve the case",
		Spec: resolveCaseSpec,
	}); err != nil {
		t.Fatal(err)
	}
	diff, err := x.SimulateRule(ctx, "resolve-case")
	if err != nil {
		t.Fatal(err)
	}
	if len(diff.Changed) != 1 || diff.Changed[0].After.State["status"] != "resolved" ||
		diff.Changed[0].Before.State["status"] != "open" {
		t.Fatalf("simulated change = %+v", diff.Changed)
	}
	if len(diff.UnexplainedBefore) != 1 || len(diff.UnexplainedAfter) != 0 {
		t.Fatalf("unexplained %v -> %v", diff.UnexplainedBefore, diff.UnexplainedAfter)
	}
	cases, _ := x.Store.ObjectsByType(ctx, "case")
	if cases[0].State["status"] != "open" {
		t.Fatal("simulation wrote an amendment")
	}
}
