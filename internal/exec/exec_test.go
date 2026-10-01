package exec_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

const invoicePayload = `{
	"customer": "ACME Sp. z o.o.",
	"currency": "PLN",
	"issue_date": "2026-09-15",
	"lines": [
		{"desc": "Widget", "qty": 2, "amount": "200.00"},
		{"desc": "Gadget", "qty": 1, "amount": "150.50"}
	]
}`

func seed(t *testing.T, x *exec.Executor) {
	t.Helper()
	ctx := context.Background()
	if err := x.Store.InsertObjectType(ctx, core.ObjectType{
		Name: "invoice", Version: 1, Domain: "finance", IsDocument: true,
		Fields: []core.FieldDef{
			{Name: "customer", Type: "string", Required: true},
			{Name: "issue_date", Type: "date", Required: true},
			{Name: "currency", Type: "string", Required: true},
			{Name: "total", Type: "money", Required: true},
		},
	}); err != nil {
		t.Fatal(err)
	}
}

func draftRule(t *testing.T, x *exec.Executor) core.Rule {
	t.Helper()
	r, err := x.Store.InsertRuleVersion(context.Background(), core.Rule{
		ID: "book-pln-invoice", Status: core.StatusDraft, Priority: 100,
		EffectiveFrom: "2026-01-01", CreatedBy: "agent",
		Description: "PLN invoices become Invoice documents with summed total",
		Spec: core.RuleSpec{
			Match: core.Match{EventType: "invoice.received", Where: []core.Condition{
				{Path: "$.currency", Op: "eq", Value: "PLN"},
			}},
			Effect: core.Effect{Object: core.ObjectTemplate{
				Type: "invoice",
				Fields: map[string]string{
					"customer":   "=$.customer",
					"issue_date": "=$.issue_date",
					"currency":   "=$.currency",
					"total":      "=sum($.lines[*].amount)",
				},
			}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// TestVerticalSliceKernel walks the kernel half of the M0 demo story:
// event → worklist → approval → materialized object → replay reproduces state.
func TestVerticalSliceKernel(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	seed(t, x)

	evID, err := x.Store.AppendEvent(ctx, core.Event{
		Kind: core.KindRaw, Type: "invoice.received", OccurredAt: "2026-09-15",
		Payload: json.RawMessage(invoicePayload), DedupKey: "inv-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	// No active rule yet: the event sits in the worklist, nothing books.
	if n, errs := x.ProcessPending(ctx); n != 0 || len(errs) != 0 {
		t.Fatalf("premature booking: %d, %v", n, errs)
	}
	if wl, _ := x.Store.UnmatchedRawEvents(ctx); len(wl) != 1 {
		t.Fatalf("worklist = %d, want 1", len(wl))
	}

	// Draft + approve: approval re-evaluates the worklist and books.
	draftRule(t, x)
	r, booked, err := x.ApproveRule(ctx, "book-pln-invoice", "krzysztof", "2026-09-15")
	if err != nil {
		t.Fatal(err)
	}
	if r.Version != 2 || r.Status != core.StatusActive || booked != 1 {
		t.Fatalf("approve: v%d %s booked=%d", r.Version, r.Status, booked)
	}

	// The worklist is clear; the object exists with provenance and the total
	// summed in minor units.
	if wl, _ := x.Store.UnmatchedRawEvents(ctx); len(wl) != 0 {
		t.Fatalf("worklist not cleared: %d", len(wl))
	}
	objs, err := x.Store.ObjectsByType(ctx, "invoice")
	if err != nil || len(objs) != 1 {
		t.Fatalf("objects: %v, %v", objs, err)
	}
	o := objs[0]
	if o.SourceEventID != evID || o.RuleID != "book-pln-invoice" || o.RuleVersion != 2 {
		t.Fatalf("provenance wrong: %+v", o)
	}
	if o.State["total"] != float64(35050) && o.State["total"] != int64(35050) {
		t.Fatalf("total = %v (%T), want 35050 minor units", o.State["total"], o.State["total"])
	}
	if o.State["customer"] != "ACME Sp. z o.o." {
		t.Fatalf("customer = %v", o.State["customer"])
	}

	// Determinism (invariant 4): wipe the cache, replay the log, compare.
	before, _ := x.Store.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatalf("replay: %v", err)
	}
	after, _ := x.Store.AllObjects(ctx)
	if !reflect.DeepEqual(normalize(t, before), normalize(t, after)) {
		t.Fatalf("replay diverged:\nbefore %+v\nafter  %+v", before, after)
	}

	// An EUR invoice does not match the PLN rule and stays in the worklist.
	if _, err := x.Store.AppendEvent(ctx, core.Event{
		Kind: core.KindRaw, Type: "invoice.received", OccurredAt: "2026-09-16",
		Payload: json.RawMessage(`{"customer":"X","currency":"EUR","issue_date":"2026-09-16","lines":[{"amount":"10.00"}]}`),
		DedupKey: "inv-2",
	}); err != nil {
		t.Fatal(err)
	}
	if n, _ := x.ProcessPending(ctx); n != 0 {
		t.Fatal("EUR invoice booked by PLN rule")
	}

	// A backdated event before the rule's effective_from is not booked.
	if _, err := x.Store.AppendEvent(ctx, core.Event{
		Kind: core.KindRaw, Type: "invoice.received", OccurredAt: "2025-12-31",
		Payload: json.RawMessage(invoicePayload), DedupKey: "inv-3",
	}); err != nil {
		t.Fatal(err)
	}
	if n, _ := x.ProcessPending(ctx); n != 0 {
		t.Fatal("event before effective_from booked")
	}
}

// normalize round-trips object state through JSON so int64/float64 encoding
// differences between live expansion and replay do not count as divergence.
func normalize(t *testing.T, objs []core.Object) []core.Object {
	t.Helper()
	b, err := json.Marshal(objs)
	if err != nil {
		t.Fatal(err)
	}
	var out []core.Object
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestApproveRejectsNonDraft(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	seed(t, x)
	draftRule(t, x)
	if _, _, err := x.ApproveRule(ctx, "book-pln-invoice", "k", "2026-09-15"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := x.ApproveRule(ctx, "book-pln-invoice", "k", "2026-09-15"); err == nil {
		t.Fatal("approving an active rule succeeded")
	}
}
