package exec_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

// The cases falsifiability test (DIRECTION: "express Alpha's case model as
// pure data"). A case is a work item as an object — kind, subject, status,
// deadline, resolving verbs — raised by rules from events, never typed in,
// closing itself when the world moves on.
//
// This file is the attempt in two halves, the E1 discipline: first what
// carries with the language as it stood (raising, the subject contortion,
// the resolving verb), then the recorded failure that earned the amendment
// (see DECISIONS 2026-10-05, "cases fail as pure data at the close").

func seedCaseWorld(t *testing.T, x *exec.Executor) {
	t.Helper()
	ctx := context.Background()
	for _, ot := range []core.ObjectType{
		{
			Name: "complaint", Version: 1, Domain: "crm", IsDocument: true,
			LabelField: "customer",
			Fields: []core.FieldDef{
				{Name: "customer", Type: "string", Required: true},
				{Name: "details", Type: "string"},
				{Name: "registered_on", Type: "date", Required: true},
			},
		},
		{
			// The case type is conventional data, not kernel vocabulary: kind,
			// subject, status, deadline. The subject is the recorded union-ref
			// contortion — (subject_type, subject_id) as strings, because
			// ref<T> names one target type and a case's subject is any object.
			// Exit named at union refs (SPEC §7, parked).
			Name: "case", Version: 1, Domain: "work", LabelField: "title",
			Fields: []core.FieldDef{
				{Name: "kind", Type: "string", Required: true},
				{Name: "subject_type", Type: "string", Required: true},
				{Name: "subject_id", Type: "string", Required: true},
				{Name: "title", Type: "string", Required: true},
				{Name: "status", Type: "enum", Required: true, Values: []string{"open", "resolved"}},
				{Name: "due_date", Type: "date"},
			},
		},
		{
			Name: "case_closure", Version: 1, Domain: "work",
			Fields: []core.FieldDef{
				{Name: "case", Type: "ref<case>", Required: true},
				{Name: "resolution", Type: "string", Required: true},
			},
		},
	} {
		if err := x.Store.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
}

func activateRule(t *testing.T, x *exec.Executor, id string, priority int, spec core.RuleSpec) {
	t.Helper()
	ctx := context.Background()
	if _, err := x.Store.InsertRuleVersion(ctx, core.Rule{
		ID: id, Status: core.StatusActive, Priority: priority,
		EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: spec,
	}); err != nil {
		t.Fatal(err)
	}
}

func submitRaw(t *testing.T, x *exec.Executor, dedup, typ, date, payload string) int64 {
	t.Helper()
	id, err := x.Store.AppendEvent(context.Background(), core.Event{
		Kind: core.KindRaw, Type: typ, OccurredAt: date,
		Payload: json.RawMessage(payload), DedupKey: dedup,
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// TestCasesRaiseAsPureData: the half that carried on the first attempt.
// A complaint event books the document, and a cascade rule raises the case
// from the document's own materialization — raised by the system from
// events, never typed in, with provenance walking back to the root fact.
func TestCasesRaiseAsPureData(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	seedCaseWorld(t, x)

	activateRule(t, x, "book-complaint", 100, core.RuleSpec{
		Match: core.Match{EventType: "complaint.registered"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "complaint",
			Fields: map[string]string{
				"customer": "=$.customer", "details": "=$.details",
				"registered_on": "=$.registered_on",
			}}},
	})
	activateRule(t, x, "raise-complaint-case", 200, core.RuleSpec{
		Match: core.Match{EventType: core.EventObjectMaterialized, Where: []core.Condition{
			{Path: "$.object_type", Op: "eq", Value: "complaint"},
		}},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "case",
			Fields: map[string]string{
				"kind":         "complaint_followup",
				"subject_type": "complaint",
				"subject_id":   "=$.object_id",
				"title":        "=$.state.customer",
				"status":       "open",
				// No date arithmetic in templates ("+14 days" is impossible):
				// the due date is the registered date, a recorded strain whose
				// exit is the arithmetic sub-language (DIRECTION 2026-10-03).
				"due_date": "=$.state.registered_on",
			}}},
	})

	rootID := submitRaw(t, x, "c-1", "complaint.registered", "2026-10-05",
		`{"customer":"ACME Sp. z o.o.","details":"wrong amount on FV 12/2026","registered_on":"2026-10-05"}`)
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("booked %d, errs %v", n, errs)
	}

	complaints, _ := x.Store.ObjectsByType(ctx, "complaint")
	cases, _ := x.Store.ObjectsByType(ctx, "case")
	if len(complaints) != 1 || len(cases) != 1 {
		t.Fatalf("complaints %d, cases %d", len(complaints), len(cases))
	}
	c := cases[0]
	if c.State["status"] != "open" || c.State["subject_id"] != complaints[0].ID ||
		c.State["kind"] != "complaint_followup" {
		t.Fatalf("case = %+v", c.State)
	}
	// Provenance: the case chains to the complaint's materialization, which
	// chains to the root raw fact — the walk answers "why does this work
	// item exist".
	ev, err := x.Store.GetEvent(ctx, c.SourceEventID)
	if err != nil || *ev.CauseEventID != rootID {
		t.Fatalf("case provenance does not chain to the root: %+v, %v", ev, err)
	}
}

// TestCaseClosureFailsAsPureData is the kept negative proof: with
// materialization as the only effect, a case cannot close.
//
// Two facts, both asserted below, located the failure precisely:
//  1. Identity physics forbids even touching the existing case. Ids are
//     cause-qualified, so a rule matching the resolution event mints
//     case-<new event> — a second case — and can never name the first.
//  2. A closure *marker* object leaves the case's own status field lying
//     ("open", forever); the truth retreats into an anti-join on the read
//     side — the hollow-explanation contortion, E3's shape. State that can
//     never tell the truth is not explained state (invariant 5).
//
// The exit is the amendment: status is a projection, never an update
// (DIRECTION 2026-10-03), now forced by fail-as-data.
func TestCaseClosureFailsAsPureData(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	seedCaseWorld(t, x)

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

	submitRaw(t, x, "c-1", "complaint.registered", "2026-10-05",
		`{"customer":"ACME Sp. z o.o.","registered_on":"2026-10-05"}`)
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("booked %d, errs %v", n, errs)
	}
	cases, _ := x.Store.ObjectsByType(ctx, "case")
	caseID := cases[0].ID

	// Attempt (i): a rule on the resolution event that materializes `case`,
	// hoping to overwrite. It cannot: cause-qualified identity mints a fresh
	// id, and the world ends up with TWO cases, both open.
	activateRule(t, x, "close-by-rematerializing", 300, core.RuleSpec{
		Match: core.Match{EventType: "complaint.resolved"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "case",
			Fields: map[string]string{
				"kind": "complaint_followup", "subject_type": "complaint",
				"subject_id": "=$.complaint", "title": "closed?",
				"status": "resolved",
			}}},
	})
	// Attempt (ii): the closure marker, referencing the case properly
	// (through resolution, E4's lesson — the payload carries the complaint
	// id, the ref resolves the case by its subject).
	activateRule(t, x, "close-by-marker", 400, core.RuleSpec{
		Match: core.Match{EventType: "complaint.resolved"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "case_closure",
			Fields: map[string]string{
				"case":       "=ref(case, subject_id, $.complaint)",
				"resolution": "=$.resolution",
			}}},
	})

	complaints, _ := x.Store.ObjectsByType(ctx, "complaint")
	submitRaw(t, x, "r-1", "complaint.resolved", "2026-10-05",
		fmt.Sprintf(`{"complaint":"%s","resolution":"credited the difference"}`, complaints[0].ID))
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("resolution: booked %d, errs %v", n, errs)
	}

	// The lie, measured: the original case is still open, a duplicate case
	// exists, and only the marker knows the truth.
	cases, _ = x.Store.ObjectsByType(ctx, "case")
	if len(cases) != 2 {
		t.Fatalf("cases = %d, want the duplicate the attempt predicts", len(cases))
	}
	for _, c := range cases {
		if c.ID == caseID && c.State["status"] != "open" {
			t.Fatalf("the original case changed — some effect can touch state, the attempt is stale")
		}
	}
	closures, _ := x.Store.ObjectsByType(ctx, "case_closure")
	if len(closures) != 1 || closures[0].State["case"] != caseID {
		t.Fatalf("closures = %+v", closures)
	}
}
