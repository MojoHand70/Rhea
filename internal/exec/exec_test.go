package exec_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
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
	r, booked, procErrs, err := x.ApproveRule(ctx, "book-pln-invoice", "krzysztof", "2026-09-15")
	if len(procErrs) > 0 {
		t.Fatalf("processing errors: %v", procErrs)
	}
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

	// Actor attribution: the approval names the human, the derived event the kernel.
	raws, _ := x.Store.EventsByKind(ctx, core.KindRaw)
	for _, ev := range raws {
		if ev.Type == "rule.approved" && ev.Actor != "krzysztof" {
			t.Fatalf("approval actor = %q, want krzysztof", ev.Actor)
		}
	}
	derived, _ := x.Store.EventsByKind(ctx, core.KindDerived)
	if len(derived) == 0 || derived[0].Actor != "kernel" {
		t.Fatalf("derived actor wrong: %+v", derived)
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
		Payload:  json.RawMessage(`{"customer":"X","currency":"EUR","issue_date":"2026-09-16","lines":[{"amount":"10.00"}]}`),
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
	if _, _, _, err := x.ApproveRule(ctx, "book-pln-invoice", "k", "2026-09-15"); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := x.ApproveRule(ctx, "book-pln-invoice", "k", "2026-09-15"); err == nil {
		t.Fatal("approving an active rule succeeded")
	}
}

// TestSimulateRule proves the dry run (SPEC M1): a draft rule replayed in
// memory shows what it would change — refs resolving against the simulated
// world — while writing nothing.
func TestSimulateRule(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	if err := x.Store.InsertObjectType(ctx, core.ObjectType{
		Name: "company", Version: 1, Domain: "finance", LabelField: "name",
		Fields: []core.FieldDef{
			{Name: "name", Type: "string", Required: true},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if err := x.Store.InsertObjectType(ctx, core.ObjectType{
		Name: "invoice", Version: 1, Domain: "finance", IsDocument: true,
		Fields: []core.FieldDef{
			{Name: "customer", Type: "ref<company>", Required: true},
			{Name: "total", Type: "money", Required: true},
		},
	}); err != nil {
		t.Fatal(err)
	}
	// The invoice arrives before the company is registered.
	for i, ev := range []core.Event{
		{Kind: core.KindRaw, Type: "invoice.received", OccurredAt: "2026-09-15",
			Payload: json.RawMessage(`{"customer":"ACME","lines":[{"amount":"10.00"}]}`)},
		{Kind: core.KindRaw, Type: "company.registered", OccurredAt: "2026-09-16",
			Payload: json.RawMessage(`{"name":"ACME"}`)},
	} {
		ev.DedupKey = fmt.Sprintf("sim-%d", i)
		if _, err := x.Store.AppendEvent(ctx, ev); err != nil {
			t.Fatal(err)
		}
	}
	draft := func(id string, spec core.RuleSpec) {
		t.Helper()
		if _, err := x.Store.InsertRuleVersion(ctx, core.Rule{
			ID: id, Status: core.StatusDraft, Priority: 100,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: spec,
		}); err != nil {
			t.Fatal(err)
		}
	}
	draft("register-company", core.RuleSpec{
		Match: core.Match{EventType: "company.registered"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "company",
			Fields: map[string]string{"name": "=$.name"}}},
	})

	// Simulating the company draft: one object appears, one event stays
	// unexplained (the invoice has no rule yet) — and nothing is written.
	diff, err := x.SimulateRule(ctx, "register-company")
	if err != nil {
		t.Fatal(err)
	}
	if len(diff.Added) != 1 || diff.Added[0].Type != "company" {
		t.Fatalf("added = %+v", diff.Added)
	}
	if len(diff.UnexplainedBefore) != 2 || len(diff.UnexplainedAfter) != 1 {
		t.Fatalf("unexplained %v -> %v", diff.UnexplainedBefore, diff.UnexplainedAfter)
	}
	if objs, _ := x.Store.AllObjects(ctx); len(objs) != 0 {
		t.Fatalf("simulation wrote %d objects", len(objs))
	}
	if evs, _ := x.Store.EventsByKind(ctx, core.KindDerived); len(evs) != 0 {
		t.Fatalf("simulation wrote %d derived events", len(evs))
	}

	// With the company rule active, simulating the invoice draft resolves
	// the ref against the simulated company via the in-memory lookup.
	if _, _, _, err := x.ApproveRule(ctx, "register-company", "krzysztof", "2026-09-16"); err != nil {
		t.Fatal(err)
	}
	draft("book-invoice", core.RuleSpec{
		Match: core.Match{EventType: "invoice.received"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "invoice",
			Fields: map[string]string{
				"customer": "=ref(company, name, $.customer)",
				"total":    "=sum($.lines[*].amount)"}}},
	})
	diff, err = x.SimulateRule(ctx, "book-invoice")
	if err != nil {
		t.Fatal(err)
	}
	if len(diff.Added) != 1 || diff.Added[0].Type != "invoice" {
		t.Fatalf("added = %+v", diff.Added)
	}
	if ref, _ := diff.Added[0].State["customer"].(string); !strings.HasPrefix(ref, "company-") {
		t.Fatalf("simulated ref = %v", diff.Added[0].State["customer"])
	}
	if len(diff.UnexplainedAfter) != 0 || len(diff.Changed) != 0 || len(diff.Removed) != 0 {
		t.Fatalf("diff = %+v", diff)
	}
}

// TestDoubleEntry walks the M1 sub-language: chart of accounts as master
// data, a posting rule booking a balanced entry, the balance invariant, the
// period lock, simulation with postings, and replay reproducing the ledger.
func TestDoubleEntry(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	for _, ot := range []core.ObjectType{
		{Name: "account", Version: 1, Domain: "finance", LabelField: "name",
			Fields: []core.FieldDef{
				{Name: "code", Type: "string", Required: true},
				{Name: "name", Type: "string", Required: true},
				{Name: "type", Type: "enum", Required: true,
					Values: []string{"asset", "liability", "equity", "revenue", "expense", "debtor", "creditor", "tax", "bank", "clearing"}},
			}},
		{Name: "posting", Version: 1, Domain: "finance",
			Fields: []core.FieldDef{
				{Name: "entry", Type: "string", Required: true},
				{Name: "line", Type: "int", Required: true},
				{Name: "account", Type: "ref<account>", Required: true},
				{Name: "side", Type: "enum", Values: []string{"debit", "credit"}, Required: true},
				{Name: "amount", Type: "money", Required: true},
				{Name: "currency", Type: "string", Required: true},
				{Name: "date", Type: "date", Required: true},
			}},
		{Name: "period_lock", Version: 2, Domain: "finance",
			Fields: []core.FieldDef{
				{Name: "month", Type: "string", Required: true},
				{Name: "book", Type: "string", Required: true}}},
	} {
		if err := x.Store.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}

	submit := func(dedup, evType, date, payload string) int64 {
		t.Helper()
		id, err := x.Store.AppendEvent(ctx, core.Event{
			Kind: core.KindRaw, Type: evType, OccurredAt: date,
			Payload: json.RawMessage(payload), DedupKey: dedup,
		})
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	activate := func(id string, spec core.RuleSpec) {
		t.Helper()
		if _, err := x.Store.InsertRuleVersion(ctx, core.Rule{
			ID: id, Status: core.StatusDraft, Priority: 100,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: spec,
		}); err != nil {
			t.Fatal(err)
		}
		if _, _, procErrs, err := x.ApproveRule(ctx, id, "krzysztof", "2026-09-30"); err != nil {
			t.Fatal(err)
		} else if len(procErrs) > 0 {
			t.Fatalf("approve %s: %v", id, procErrs)
		}
	}

	// Chart of accounts and the period-lock activity, both as plain rules.
	activate("register-account", core.RuleSpec{
		Match: core.Match{EventType: "account.created"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "account",
			Fields: map[string]string{"code": "=$.code", "name": "=$.name", "type": "=$.type"}}},
	})
	activate("lock-period", core.RuleSpec{
		Match: core.Match{EventType: "period.locked"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "period_lock",
			Fields: map[string]string{"month": "=$.month", "book": "main"}}},
	})
	submit("acc-201", "account.created", "2026-09-01", `{"code":"201","name":"Receivables","type":"debtor"}`)
	submit("acc-702", "account.created", "2026-09-01", `{"code":"702","name":"Sales revenue","type":"revenue"}`)
	if n, errs := x.ProcessPending(ctx); n != 2 || len(errs) != 0 {
		t.Fatalf("accounts: booked %d, errs %v", n, errs)
	}

	// Draft the posting rule and simulate it first: two balanced lines.
	postSpec := core.RuleSpec{
		Match: core.Match{EventType: "invoice.received",
			Where: []core.Condition{{Path: "$.currency", Op: "eq", Value: "PLN"}}},
		Effect: core.Effect{Postings: &core.PostingsTemplate{
			Currency: "=$.currency",
			Lines: []core.PostingLine{
				{Account: "201", Debit: "=sum($.lines[*].amount)"},
				{Account: "702", Credit: "=sum($.lines[*].amount)"},
			},
		}},
	}
	invoiceEv := submit("inv-1", "invoice.received", "2026-09-15",
		`{"currency":"PLN","lines":[{"amount":"200.00"},{"amount":"150.50"}]}`)
	if _, err := x.Store.InsertRuleVersion(ctx, core.Rule{
		ID: "post-pln-invoice", Status: core.StatusDraft, Priority: 100,
		EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: "post", Spec: postSpec,
	}); err != nil {
		t.Fatal(err)
	}
	diff, err := x.SimulateRule(ctx, "post-pln-invoice")
	if err != nil {
		t.Fatal(err)
	}
	if len(diff.Added) != 2 {
		t.Fatalf("simulated postings = %+v", diff.Added)
	}
	if objs, _ := x.Store.ObjectsByType(ctx, "posting"); len(objs) != 0 {
		t.Fatal("simulation wrote postings")
	}

	// Approve: the entry books, balanced, with resolved account refs.
	if _, booked, procErrs, err := x.ApproveRule(ctx, "post-pln-invoice", "krzysztof", "2026-09-30"); err != nil || booked != 1 || len(procErrs) != 0 {
		t.Fatalf("approve: booked %d, %v, %v", booked, procErrs, err)
	}
	postings, _ := x.Store.ObjectsByType(ctx, "posting")
	if len(postings) != 2 {
		t.Fatalf("postings = %d", len(postings))
	}
	var d, c int64
	for _, p := range postings {
		amt := int64(p.State["amount"].(float64))
		if p.State["side"] == "debit" {
			d += amt
		} else {
			c += amt
		}
		if acc, _ := p.State["account"].(string); !strings.HasPrefix(acc, "account-") {
			t.Fatalf("account not resolved: %v", p.State["account"])
		}
		if p.State["entry"] != fmt.Sprintf("entry-%d-post-pln-invoice", invoiceEv) {
			t.Fatalf("entry key = %v", p.State["entry"])
		}
	}
	if d != 35050 || c != 35050 {
		t.Fatalf("not balanced: D %d C %d", d, c)
	}

	// Balance invariant: an unbalanced expansion books nothing.
	activate("bad-fee", core.RuleSpec{
		Match: core.Match{EventType: "fee.charged"},
		Effect: core.Effect{Postings: &core.PostingsTemplate{
			Currency: "PLN",
			Lines: []core.PostingLine{
				{Account: "201", Debit: "10.00"},
				{Account: "702", Credit: "20.00"},
			},
		}},
	})
	submit("fee-1", "fee.charged", "2026-09-16", `{}`)
	if n, errs := x.ProcessPending(ctx); n != 0 || len(errs) == 0 {
		t.Fatalf("unbalanced entry: booked %d, errs %v", n, errs)
	}
	if objs, _ := x.Store.ObjectsByType(ctx, "posting"); len(objs) != 2 {
		t.Fatal("unbalanced entry left postings behind")
	}

	// Period lock: lock September, then a September invoice is refused.
	submit("lock-09", "period.locked", "2026-09-30", `{"month":"2026-09"}`)
	if n, _ := x.ProcessPending(ctx); n != 1 {
		t.Fatal("lock did not book")
	}
	submit("inv-2", "invoice.received", "2026-09-20",
		`{"currency":"PLN","lines":[{"amount":"10.00"}]}`)
	if n, errs := x.ProcessPending(ctx); n != 0 || len(errs) == 0 || !strings.Contains(errs[0].Error(), "locked") {
		t.Fatalf("locked period: booked %d, errs %v", n, errs)
	}
	// An October invoice still books.
	submit("inv-3", "invoice.received", "2026-10-02",
		`{"currency":"PLN","lines":[{"amount":"10.00"}]}`)
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) == 0 {
		// the September invoice keeps failing in the same pass — that error stays
		t.Fatalf("open period: booked %d, errs %v", n, errs)
	}

	// Determinism: replay reproduces the ledger exactly.
	before, _ := x.Store.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := x.Store.AllObjects(ctx)
	if !reflect.DeepEqual(normalize(t, before), normalize(t, after)) {
		t.Fatal("replay diverged")
	}
}

// TestCascade: rules match derived events (DECISIONS 2026-10-02, the intended
// end-state). A goods receipt becomes a stock movement, whose materialization
// event fires a valuation rule — postings and a note referencing the movement
// created in the same chain — all booked atomically with the root event.
// Cascaded identity builds on the causing object's id, itself rooted in the
// raw event, so simulation reproduces it with no sequence state; a cyclic
// rule set runs into the depth cap, which names the looping rule.
func TestCascade(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	for _, ot := range []core.ObjectType{
		{Name: "item", Version: 1, Domain: "warehouse", LabelField: "name",
			Fields: []core.FieldDef{
				{Name: "sku", Type: "string", Required: true},
				{Name: "name", Type: "string", Required: true},
			}},
		{Name: "location", Version: 1, Domain: "warehouse", LabelField: "name",
			Fields: []core.FieldDef{
				{Name: "code", Type: "string", Required: true},
				{Name: "name", Type: "string", Required: true},
			}},
		{Name: "stock_movement", Version: 1, Domain: "warehouse",
			Fields: []core.FieldDef{
				{Name: "item", Type: "ref<item>", Required: true},
				{Name: "location", Type: "ref<location>", Required: true},
				{Name: "direction", Type: "enum", Values: []string{"in", "out"}, Required: true},
				{Name: "qty", Type: "int", Required: true},
				{Name: "date", Type: "date", Required: true},
				{Name: "value", Type: "money", Required: true},
			}},
		{Name: "valuation_note", Version: 1, Domain: "warehouse",
			Fields: []core.FieldDef{
				{Name: "movement", Type: "ref<stock_movement>", Required: true},
				{Name: "value", Type: "money", Required: true},
			}},
		{Name: "account", Version: 1, Domain: "finance",
			Fields: []core.FieldDef{
				{Name: "code", Type: "string", Required: true},
				{Name: "name", Type: "string", Required: true},
			}},
		{Name: "posting", Version: 1, Domain: "finance",
			Fields: []core.FieldDef{
				{Name: "entry", Type: "string", Required: true},
				{Name: "line", Type: "int", Required: true},
				{Name: "account", Type: "ref<account>", Required: true},
				{Name: "side", Type: "enum", Values: []string{"debit", "credit"}, Required: true},
				{Name: "amount", Type: "money", Required: true},
				{Name: "currency", Type: "string", Required: true},
				{Name: "date", Type: "date", Required: true},
			}},
		{Name: "period_lock", Version: 2, Domain: "finance",
			Fields: []core.FieldDef{
				{Name: "month", Type: "string", Required: true},
				{Name: "book", Type: "string", Required: true}}},
	} {
		if err := x.Store.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}

	submit := func(dedup, evType, date, payload string) int64 {
		t.Helper()
		id, err := x.Store.AppendEvent(ctx, core.Event{
			Kind: core.KindRaw, Type: evType, OccurredAt: date,
			Payload: json.RawMessage(payload), DedupKey: dedup,
		})
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	draft := func(id string, priority int, spec core.RuleSpec) {
		t.Helper()
		if _, err := x.Store.InsertRuleVersion(ctx, core.Rule{
			ID: id, Status: core.StatusDraft, Priority: priority,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: spec,
		}); err != nil {
			t.Fatal(err)
		}
	}
	activate := func(id string, priority int, spec core.RuleSpec) {
		t.Helper()
		draft(id, priority, spec)
		if _, _, procErrs, err := x.ApproveRule(ctx, id, "krzysztof", "2026-09-01"); err != nil {
			t.Fatalf("approve %s: %v", id, err)
		} else if len(procErrs) > 0 {
			t.Fatalf("approve %s: %v", id, procErrs)
		}
	}
	obj := func(typ string, fields map[string]string) core.Effect {
		return core.Effect{Object: core.ObjectTemplate{Type: typ, Fields: fields}}
	}

	// Master data and the movement rule — the M2 state of the world, plus a
	// value carried on the movement for the valuation to post from.
	activate("register-item", 10, core.RuleSpec{
		Match:  core.Match{EventType: "item.created"},
		Effect: obj("item", map[string]string{"sku": "=$.sku", "name": "=$.name"})})
	activate("register-location", 10, core.RuleSpec{
		Match:  core.Match{EventType: "location.created"},
		Effect: obj("location", map[string]string{"code": "=$.code", "name": "=$.name"})})
	activate("register-account", 10, core.RuleSpec{
		Match:  core.Match{EventType: "account.created"},
		Effect: obj("account", map[string]string{"code": "=$.code", "name": "=$.name"})})
	activate("lock-period", 10, core.RuleSpec{
		Match:  core.Match{EventType: "period.locked"},
		Effect: obj("period_lock", map[string]string{"month": "=$.month", "book": "main"})})
	activate("move-stock-in", 100, core.RuleSpec{
		Match: core.Match{EventType: "goods.received"},
		Effect: obj("stock_movement", map[string]string{
			"item": "=ref(item, sku, $.item)", "location": "=ref(location, code, $.location)",
			"direction": "in", "qty": "=$.qty", "date": "=$.date", "value": "=$.value"})})
	submit("item-1", "item.created", "2026-09-01", `{"sku":"WID-1","name":"Widget"}`)
	submit("loc-1", "location.created", "2026-09-01", `{"code":"MAIN","name":"Main"}`)
	submit("acc-310", "account.created", "2026-09-01", `{"code":"310","name":"Materials"}`)
	submit("acc-300", "account.created", "2026-09-01", `{"code":"300","name":"GR/IR clearing"}`)
	if n, errs := x.ProcessPending(ctx); n != 4 || len(errs) != 0 {
		t.Fatalf("masters: booked %d, errs %v", n, errs)
	}

	// A receipt before any cascade rule exists: movement only.
	submit("gr-1", "goods.received", "2026-09-10",
		`{"item":"WID-1","location":"MAIN","qty":25,"date":"2026-09-10","value":"1250.00"}`)
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("receipt 1: booked %d, errs %v", n, errs)
	}

	// The valuation rule matches the movement's own materialization event.
	valueSpec := core.RuleSpec{
		Match: core.Match{EventType: core.EventObjectMaterialized, Where: []core.Condition{
			{Path: "$.object_type", Op: "eq", Value: "stock_movement"},
			{Path: "$.state.direction", Op: "eq", Value: "in"},
		}},
		Effect: core.Effect{Postings: &core.PostingsTemplate{
			Currency: "PLN",
			Lines: []core.PostingLine{
				{Account: "310", Debit: "=$.state.value"},
				{Account: "300", Credit: "=$.state.value"},
			},
		}},
	}

	// Simulated first: the dry run cascades too, showing the postings the
	// historical receipt would have produced — while writing nothing.
	draft("value-stock-in", 200, valueSpec)
	diff, err := x.SimulateRule(ctx, "value-stock-in")
	if err != nil {
		t.Fatal(err)
	}
	if len(diff.Added) != 2 || diff.Added[0].Type != "posting" {
		t.Fatalf("simulated cascade added %+v", diff.Added)
	}
	if objs, _ := x.Store.ObjectsByType(ctx, "posting"); len(objs) != 0 {
		t.Fatal("simulation wrote postings")
	}

	// Approving it books nothing: receipt 1 is already explained, and a rule
	// approved later never re-fires old events (no retroactivity).
	if _, booked, procErrs, err := x.ApproveRule(ctx, "value-stock-in", "krzysztof", "2026-09-11"); err != nil || booked != 0 || len(procErrs) != 0 {
		t.Fatalf("approve valuation: booked %d, %v, %v", booked, procErrs, err)
	}

	// A second cascade rule whose ref resolves against the chain itself: the
	// movement it references does not exist in the cache while it expands.
	activate("note-valuation", 300, core.RuleSpec{
		Match: core.Match{EventType: core.EventObjectMaterialized, Where: []core.Condition{
			{Path: "$.object_type", Op: "eq", Value: "stock_movement"},
			{Path: "$.state.direction", Op: "eq", Value: "in"},
		}},
		Effect: obj("valuation_note", map[string]string{
			"movement": "=ref(stock_movement, value, $.state.value)",
			"value":    "=$.state.value"})})

	// The full chain: receipt → movement → postings + note, one transaction.
	gr2 := submit("gr-2", "goods.received", "2026-09-12",
		`{"item":"WID-1","location":"MAIN","qty":16,"date":"2026-09-12","value":"800.00"}`)
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("receipt 2: booked %d, errs %v", n, errs)
	}
	movements, _ := x.Store.ObjectsByType(ctx, "stock_movement")
	postings, _ := x.Store.ObjectsByType(ctx, "posting")
	notes, _ := x.Store.ObjectsByType(ctx, "valuation_note")
	if len(movements) != 2 || len(postings) != 2 || len(notes) != 1 {
		t.Fatalf("movements %d, postings %d, notes %d", len(movements), len(postings), len(notes))
	}
	if notes[0].State["movement"] != fmt.Sprintf("stock_movement-%d", gr2) {
		t.Fatalf("chain ref = %v", notes[0].State["movement"])
	}
	var d, c int64
	for _, p := range postings {
		// A cascaded entry is keyed by the causing object, not the root event.
		if p.State["entry"] != fmt.Sprintf("entry-stock_movement-%d-value-stock-in", gr2) {
			t.Fatalf("entry key = %v", p.State["entry"])
		}
		amt := int64(p.State["amount"].(float64))
		if p.State["side"] == "debit" {
			d += amt
		} else {
			c += amt
		}
	}
	if d != 80000 || c != 80000 {
		t.Fatalf("not balanced: D %d C %d", d, c)
	}

	// Provenance walks the chain through the log: the movement's cause is the
	// receipt, the postings' and note's cause is the movement's own
	// materialization event — and the object cache agrees.
	derived, _ := x.Store.EventsByKind(ctx, core.KindDerived)
	var movementEv int64
	for _, ev := range derived {
		if ev.CauseEventID != nil && *ev.CauseEventID == gr2 {
			if movementEv != 0 {
				t.Fatalf("more than one event caused directly by receipt %d", gr2)
			}
			movementEv = ev.ID
		}
	}
	cascaded := 0
	for _, ev := range derived {
		if ev.CauseEventID != nil && *ev.CauseEventID == movementEv {
			cascaded++
		}
	}
	if movementEv == 0 || cascaded != 3 {
		t.Fatalf("cause chain: movement event %d, cascaded %d (want 3)", movementEv, cascaded)
	}
	if postings[0].SourceEventID != movementEv || notes[0].SourceEventID != movementEv {
		t.Fatalf("cascaded provenance: posting %d, note %d, want %d",
			postings[0].SourceEventID, notes[0].SourceEventID, movementEv)
	}

	// A locked period refuses the whole chain: no postings means no movement
	// either — the event is never half-explained.
	submit("lock-10", "period.locked", "2026-09-30", `{"month":"2026-10"}`)
	if n, _ := x.ProcessPending(ctx); n != 1 {
		t.Fatal("lock did not book")
	}
	submit("gr-3", "goods.received", "2026-10-05",
		`{"item":"WID-1","location":"MAIN","qty":5,"date":"2026-10-05","value":"250.00"}`)
	n, errs := x.ProcessPending(ctx)
	if n != 0 || len(errs) == 0 || !strings.Contains(errs[0].Error(), "locked") {
		t.Fatalf("locked chain: booked %d, errs %v", n, errs)
	}
	if ms, _ := x.Store.ObjectsByType(ctx, "stock_movement"); len(ms) != 2 {
		t.Fatal("refused chain left a movement behind")
	}

	// A cyclic rule — a note firing on the note's own materialization — mints
	// a fresh cause-qualified id each generation, so the depth cap is what
	// stops it, naming the looping rule and the id that shows the loop.
	// (Approved directly: the locked receipt still reports from the worklist.)
	draft("echo-note", 400, core.RuleSpec{
		Match: core.Match{EventType: core.EventObjectMaterialized, Where: []core.Condition{
			{Path: "$.object_type", Op: "eq", Value: "valuation_note"},
		}},
		Effect: obj("valuation_note", map[string]string{
			"movement": "=ref(stock_movement, value, $.state.value)",
			"value":    "=$.state.value"})})
	if _, _, _, err := x.ApproveRule(ctx, "echo-note", "krzysztof", "2026-09-20"); err != nil {
		t.Fatal(err)
	}
	submit("gr-4", "goods.received", "2026-09-20",
		`{"item":"WID-1","location":"MAIN","qty":3,"date":"2026-09-20","value":"999.00"}`)
	n, errs = x.ProcessPending(ctx)
	joined := fmt.Sprintf("%v", errs)
	if n != 0 || !strings.Contains(joined, "exceeded") || !strings.Contains(joined, "echo-note") {
		t.Fatalf("cycle: booked %d, errs %v", n, errs)
	}
	if ns, _ := x.Store.ObjectsByType(ctx, "valuation_note"); len(ns) != 1 {
		t.Fatal("cyclic chain booked notes")
	}
	if wl, _ := x.Store.UnmatchedRawEvents(ctx); len(wl) != 2 {
		t.Fatalf("worklist = %d, want the locked and the cyclic receipt", len(wl))
	}

	// The derived namespace cannot be submitted from outside.
	if _, err := x.Store.AppendEvent(ctx, core.Event{
		Kind: core.KindRaw, Type: core.EventObjectMaterialized, OccurredAt: "2026-09-21",
		Payload: json.RawMessage(`{"object_type":"stock_movement","state":{"direction":"in","value":100}}`),
	}); err == nil {
		t.Fatal("raw object.materialized accepted")
	}

	// Determinism (invariant 4) holds across cascades.
	before, _ := x.Store.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := x.Store.AllObjects(ctx)
	if !reflect.DeepEqual(normalize(t, before), normalize(t, after)) {
		t.Fatal("replay diverged")
	}
}

// TestEachEffect: multi-line documents, the M2 simplification lifted. An
// "each" object effect fans out over an array in the payload and materializes
// one object per element — the posting pattern generalized to any type — with
// per-line provenance, line-numbered rooted ids, and the same all-or-nothing
// booking as every other firing.
func TestEachEffect(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	for _, ot := range []core.ObjectType{
		{Name: "item", Version: 1, Domain: "warehouse", LabelField: "name",
			Fields: []core.FieldDef{
				{Name: "sku", Type: "string", Required: true},
				{Name: "name", Type: "string", Required: true},
			}},
		{Name: "location", Version: 1, Domain: "warehouse", LabelField: "name",
			Fields: []core.FieldDef{
				{Name: "code", Type: "string", Required: true},
				{Name: "name", Type: "string", Required: true},
			}},
		{Name: "goods_receipt", Version: 1, Domain: "warehouse", IsDocument: true,
			Fields: []core.FieldDef{
				{Name: "grn", Type: "string", Required: true},
				{Name: "location", Type: "ref<location>", Required: true},
				{Name: "date", Type: "date", Required: true},
			}},
		{Name: "stock_movement", Version: 1, Domain: "warehouse",
			Fields: []core.FieldDef{
				{Name: "grn", Type: "string", Required: true}, // the parent document's business key
				{Name: "item", Type: "ref<item>", Required: true},
				{Name: "location", Type: "ref<location>", Required: true},
				{Name: "direction", Type: "enum", Values: []string{"in", "out"}, Required: true},
				{Name: "qty", Type: "int", Required: true},
				{Name: "date", Type: "date", Required: true},
			}},
	} {
		if err := x.Store.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}

	submit := func(dedup, evType, date, payload string) int64 {
		t.Helper()
		id, err := x.Store.AppendEvent(ctx, core.Event{
			Kind: core.KindRaw, Type: evType, OccurredAt: date,
			Payload: json.RawMessage(payload), DedupKey: dedup,
		})
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	obj := func(typ string, fields map[string]string) core.Effect {
		return core.Effect{Object: core.ObjectTemplate{Type: typ, Fields: fields}}
	}
	activate := func(id string, priority int, spec core.RuleSpec) {
		t.Helper()
		if _, err := x.Store.InsertRuleVersion(ctx, core.Rule{
			ID: id, Status: core.StatusDraft, Priority: priority,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: spec,
		}); err != nil {
			t.Fatal(err)
		}
		if _, _, procErrs, err := x.ApproveRule(ctx, id, "krzysztof", "2026-09-01"); err != nil || len(procErrs) > 0 {
			t.Fatalf("approve %s: %v %v", id, procErrs, err)
		}
	}

	activate("register-item", 10, core.RuleSpec{
		Match:  core.Match{EventType: "item.created"},
		Effect: obj("item", map[string]string{"sku": "=$.sku", "name": "=$.name"})})
	activate("register-location", 10, core.RuleSpec{
		Match:  core.Match{EventType: "location.created"},
		Effect: obj("location", map[string]string{"code": "=$.code", "name": "=$.name"})})
	activate("book-receipt", 100, core.RuleSpec{
		Match: core.Match{EventType: "goods.received"},
		Effect: obj("goods_receipt", map[string]string{
			"grn": "=$.grn", "location": "=ref(location, code, $.location)", "date": "=$.date"})})
	submit("item-w", "item.created", "2026-09-01", `{"sku":"WID-1","name":"Widget"}`)
	submit("item-g", "item.created", "2026-09-01", `{"sku":"GAD-1","name":"Gadget"}`)
	submit("loc-1", "location.created", "2026-09-01", `{"code":"MAIN","name":"Main"}`)
	if n, errs := x.ProcessPending(ctx); n != 3 || len(errs) != 0 {
		t.Fatalf("masters: booked %d, errs %v", n, errs)
	}

	// A two-line receipt while only the document rule runs: header books,
	// lines wait for their rule.
	gr1 := submit("gr-1", "goods.received", "2026-09-15",
		`{"grn":"GRN-7","location":"MAIN","date":"2026-09-15","lines":[{"item":"WID-1","qty":10},{"item":"GAD-1","qty":4}]}`)
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("receipt 1: booked %d, errs %v", n, errs)
	}

	// The lines rule: one movement per line, templates scoped {doc, line, n}.
	linesSpec := core.RuleSpec{
		Match: core.Match{EventType: "goods.received"},
		Effect: core.Effect{Object: core.ObjectTemplate{
			Type: "stock_movement", Each: "=$.lines[*]",
			Fields: map[string]string{
				"grn":      "=$.doc.grn",
				"item":     "=ref(item, sku, $.line.item)",
				"location": "=ref(location, code, $.doc.location)",
				"direction": "in", "qty": "=$.line.qty", "date": "=$.doc.date",
			}}}}

	// Simulated as a draft first: the dry run fans out too — two movements
	// for the historical receipt — and writes nothing.
	if _, err := x.Store.InsertRuleVersion(ctx, core.Rule{
		ID: "move-lines", Status: core.StatusDraft, Priority: 200,
		EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: "move-lines", Spec: linesSpec,
	}); err != nil {
		t.Fatal(err)
	}
	diff, err := x.SimulateRule(ctx, "move-lines")
	if err != nil {
		t.Fatal(err)
	}
	if len(diff.Added) != 2 || diff.Added[0].ID != fmt.Sprintf("stock_movement-%d-1", gr1) {
		t.Fatalf("simulated lines = %+v", diff.Added)
	}
	if objs, _ := x.Store.ObjectsByType(ctx, "stock_movement"); len(objs) != 0 {
		t.Fatal("simulation wrote movements")
	}
	if _, booked, procErrs, err := x.ApproveRule(ctx, "move-lines", "krzysztof", "2026-09-16"); err != nil || booked != 0 || len(procErrs) != 0 {
		t.Fatalf("approve move-lines: booked %d, %v, %v", booked, procErrs, err) // no retroactivity
	}

	// A second receipt books header and both lines in one transaction.
	gr2 := submit("gr-2", "goods.received", "2026-09-17",
		`{"grn":"GRN-8","location":"MAIN","date":"2026-09-17","lines":[{"item":"WID-1","qty":5},{"item":"GAD-1","qty":3}]}`)
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("receipt 2: booked %d, errs %v", n, errs)
	}
	movements, _ := x.Store.ObjectsByType(ctx, "stock_movement")
	if len(movements) != 2 {
		t.Fatalf("movements = %d", len(movements))
	}
	for i, m := range movements {
		if m.ID != fmt.Sprintf("stock_movement-%d-%d", gr2, i+1) {
			t.Fatalf("line id = %s", m.ID)
		}
		if m.State["grn"] != "GRN-8" || m.RuleID != "move-lines" {
			t.Fatalf("line %d: %+v", i+1, m)
		}
		if ref, _ := m.State["item"].(string); !strings.HasPrefix(ref, "item-") {
			t.Fatalf("line %d item not resolved: %v", i+1, m.State["item"])
		}
	}
	if movements[0].State["qty"] != float64(5) || movements[1].State["qty"] != float64(3) {
		t.Fatalf("line qtys = %v, %v", movements[0].State["qty"], movements[1].State["qty"])
	}

	// A receipt with no lines is malformed, not silently explained: the whole
	// event is refused — no header either — and waits for a human.
	submit("gr-3", "goods.received", "2026-09-18",
		`{"grn":"GRN-9","location":"MAIN","date":"2026-09-18","lines":[]}`)
	n, errs := x.ProcessPending(ctx)
	if n != 0 || len(errs) == 0 || !strings.Contains(errs[0].Error(), "no elements") {
		t.Fatalf("empty lines: booked %d, errs %v", n, errs)
	}
	if docs, _ := x.Store.ObjectsByType(ctx, "goods_receipt"); len(docs) != 2 {
		t.Fatalf("empty-lines receipt half-booked: %d docs", len(docs))
	}
	if wl, _ := x.Store.UnmatchedRawEvents(ctx); len(wl) != 1 {
		t.Fatalf("worklist = %d", len(wl))
	}

	// Determinism holds across fan-outs.
	before, _ := x.Store.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := x.Store.AllObjects(ctx)
	if !reflect.DeepEqual(normalize(t, before), normalize(t, after)) {
		t.Fatal("replay diverged")
	}
}

// TestCascadePerLine: the combination that forced cause-qualified identity —
// an each-effect fans a receipt into one movement per line, and the valuation
// cascade fires once per movement, each firing booking its own balanced
// journal entry keyed by its movement's id. One raw event, two lines, two
// entries, four postings, one transaction.
func TestCascadePerLine(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	for _, ot := range []core.ObjectType{
		{Name: "item", Version: 1, Domain: "warehouse", LabelField: "name",
			Fields: []core.FieldDef{
				{Name: "sku", Type: "string", Required: true},
				{Name: "name", Type: "string", Required: true},
			}},
		{Name: "stock_movement", Version: 1, Domain: "warehouse",
			Fields: []core.FieldDef{
				{Name: "grn", Type: "string", Required: true},
				{Name: "item", Type: "ref<item>", Required: true},
				{Name: "direction", Type: "enum", Values: []string{"in", "out"}, Required: true},
				{Name: "qty", Type: "int", Required: true},
				{Name: "date", Type: "date", Required: true},
				{Name: "value", Type: "money", Required: true},
			}},
		{Name: "account", Version: 1, Domain: "finance",
			Fields: []core.FieldDef{
				{Name: "code", Type: "string", Required: true},
				{Name: "name", Type: "string", Required: true},
			}},
		{Name: "posting", Version: 1, Domain: "finance",
			Fields: []core.FieldDef{
				{Name: "entry", Type: "string", Required: true},
				{Name: "line", Type: "int", Required: true},
				{Name: "account", Type: "ref<account>", Required: true},
				{Name: "side", Type: "enum", Values: []string{"debit", "credit"}, Required: true},
				{Name: "amount", Type: "money", Required: true},
				{Name: "currency", Type: "string", Required: true},
				{Name: "date", Type: "date", Required: true},
			}},
	} {
		if err := x.Store.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	submit := func(dedup, evType, date, payload string) int64 {
		t.Helper()
		id, err := x.Store.AppendEvent(ctx, core.Event{
			Kind: core.KindRaw, Type: evType, OccurredAt: date,
			Payload: json.RawMessage(payload), DedupKey: dedup,
		})
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	activate := func(id string, priority int, spec core.RuleSpec) {
		t.Helper()
		if _, err := x.Store.InsertRuleVersion(ctx, core.Rule{
			ID: id, Status: core.StatusDraft, Priority: priority,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: spec,
		}); err != nil {
			t.Fatal(err)
		}
		if _, _, procErrs, err := x.ApproveRule(ctx, id, "krzysztof", "2026-09-01"); err != nil || len(procErrs) > 0 {
			t.Fatalf("approve %s: %v %v", id, procErrs, err)
		}
	}
	obj := func(typ string, fields map[string]string) core.Effect {
		return core.Effect{Object: core.ObjectTemplate{Type: typ, Fields: fields}}
	}

	activate("register-item", 10, core.RuleSpec{
		Match:  core.Match{EventType: "item.created"},
		Effect: obj("item", map[string]string{"sku": "=$.sku", "name": "=$.name"})})
	activate("register-account", 10, core.RuleSpec{
		Match:  core.Match{EventType: "account.created"},
		Effect: obj("account", map[string]string{"code": "=$.code", "name": "=$.name"})})
	activate("move-lines", 100, core.RuleSpec{
		Match: core.Match{EventType: "goods.received"},
		Effect: core.Effect{Object: core.ObjectTemplate{
			Type: "stock_movement", Each: "=$.lines[*]",
			Fields: map[string]string{
				"grn": "=$.doc.grn", "item": "=ref(item, sku, $.line.item)",
				"direction": "in", "qty": "=$.line.qty", "date": "=$.doc.date",
				"value": "=$.line.value",
			}}}})
	activate("value-movement", 200, core.RuleSpec{
		Match: core.Match{EventType: core.EventObjectMaterialized, Where: []core.Condition{
			{Path: "$.object_type", Op: "eq", Value: "stock_movement"},
			{Path: "$.state.direction", Op: "eq", Value: "in"},
		}},
		Effect: core.Effect{Postings: &core.PostingsTemplate{
			Currency: "PLN",
			Lines: []core.PostingLine{
				{Account: "310", Debit: "=$.state.value"},
				{Account: "300", Credit: "=$.state.value"},
			},
		}}})
	submit("item-w", "item.created", "2026-09-01", `{"sku":"WID-1","name":"Widget"}`)
	submit("item-g", "item.created", "2026-09-01", `{"sku":"GAD-1","name":"Gadget"}`)
	submit("acc-310", "account.created", "2026-09-01", `{"code":"310","name":"Materials"}`)
	submit("acc-300", "account.created", "2026-09-01", `{"code":"300","name":"GR/IR clearing"}`)
	if n, errs := x.ProcessPending(ctx); n != 4 || len(errs) != 0 {
		t.Fatalf("masters: booked %d, errs %v", n, errs)
	}

	gr := submit("gr-1", "goods.received", "2026-09-15",
		`{"grn":"GRN-7","date":"2026-09-15","lines":[
			{"item":"WID-1","qty":10,"value":"100.00"},
			{"item":"GAD-1","qty":5,"value":"50.00"}]}`)
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("receipt: booked %d, errs %v", n, errs)
	}

	movements, _ := x.Store.ObjectsByType(ctx, "stock_movement")
	postings, _ := x.Store.ObjectsByType(ctx, "posting")
	if len(movements) != 2 || len(postings) != 4 {
		t.Fatalf("movements %d, postings %d", len(movements), len(postings))
	}

	// Each movement got its own journal entry, keyed by the movement's id,
	// and each entry balances on its own line's value.
	byEntry := map[string][2]int64{} // entry key → [debits, credits]
	for _, p := range postings {
		k := p.State["entry"].(string)
		dc := byEntry[k]
		amt := int64(p.State["amount"].(float64))
		if p.State["side"] == "debit" {
			dc[0] += amt
		} else {
			dc[1] += amt
		}
		byEntry[k] = dc
	}
	wantEntries := map[string][2]int64{
		fmt.Sprintf("entry-stock_movement-%d-1-value-movement", gr): {10000, 10000},
		fmt.Sprintf("entry-stock_movement-%d-2-value-movement", gr): {5000, 5000},
	}
	if !reflect.DeepEqual(byEntry, wantEntries) {
		t.Fatalf("entries = %v, want %v", byEntry, wantEntries)
	}

	// Per-line provenance: the two entries trace to two different causing
	// events — each movement's own materialization.
	causes := map[int64]bool{}
	for _, p := range postings {
		causes[p.SourceEventID] = true
	}
	if len(causes) != 2 {
		t.Fatalf("posting causes = %v, want one per movement", causes)
	}

	// Determinism holds across per-line cascades.
	before, _ := x.Store.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := x.Store.AllObjects(ctx)
	if !reflect.DeepEqual(normalize(t, before), normalize(t, after)) {
		t.Fatal("replay diverged")
	}
}

// TestMultiRuleFiring: every matching rule fires, atomically per event — one
// invoice event becomes a document AND a balanced ledger entry. Two rules
// claiming the same object id are a conflict: the event is refused whole.
func TestMultiRuleFiring(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	seed(t, x) // invoice v1 object type
	for _, ot := range []core.ObjectType{
		{Name: "account", Version: 1, Domain: "finance",
			Fields: []core.FieldDef{
				{Name: "code", Type: "string", Required: true},
				{Name: "name", Type: "string", Required: true},
			}},
		{Name: "posting", Version: 1, Domain: "finance",
			Fields: []core.FieldDef{
				{Name: "entry", Type: "string", Required: true},
				{Name: "line", Type: "int", Required: true},
				{Name: "account", Type: "ref<account>", Required: true},
				{Name: "side", Type: "enum", Values: []string{"debit", "credit"}, Required: true},
				{Name: "amount", Type: "money", Required: true},
				{Name: "currency", Type: "string", Required: true},
				{Name: "date", Type: "date", Required: true},
			}},
	} {
		if err := x.Store.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	activateRule := func(id string, priority int, spec core.RuleSpec) {
		t.Helper()
		if _, err := x.Store.InsertRuleVersion(ctx, core.Rule{
			ID: id, Status: core.StatusDraft, Priority: priority,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: spec,
		}); err != nil {
			t.Fatal(err)
		}
		if _, _, procErrs, err := x.ApproveRule(ctx, id, "k", "2026-09-15"); err != nil || len(procErrs) > 0 {
			t.Fatalf("approve %s: %v %v", id, procErrs, err)
		}
	}
	activateRule("register-account", 10, core.RuleSpec{
		Match: core.Match{EventType: "account.created"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "account",
			Fields: map[string]string{"code": "=$.code", "name": "=$.name"}}},
	})
	for _, acc := range []string{`{"code":"201","name":"Receivables"}`, `{"code":"702","name":"Revenue"}`} {
		if _, err := x.Store.AppendEvent(ctx, core.Event{
			Kind: core.KindRaw, Type: "account.created", OccurredAt: "2026-09-01",
			Payload: json.RawMessage(acc), DedupKey: acc[9:12],
		}); err != nil {
			t.Fatal(err)
		}
	}
	if n, errs := x.ProcessPending(ctx); n != 2 || len(errs) != 0 {
		t.Fatalf("accounts: %d %v", n, errs)
	}

	// Document rule and ledger rule, both matching invoice.received.
	activateRule("book-invoice-document", 100, core.RuleSpec{
		Match: core.Match{EventType: "invoice.received"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "invoice",
			Fields: map[string]string{
				"customer": "=$.customer", "issue_date": "=$.issue_date",
				"currency": "=$.currency", "total": "=sum($.lines[*].amount)"}}},
	})
	activateRule("post-invoice-ledger", 200, core.RuleSpec{
		Match: core.Match{EventType: "invoice.received"},
		Effect: core.Effect{Postings: &core.PostingsTemplate{
			Currency: "=$.currency",
			Lines: []core.PostingLine{
				{Account: "201", Debit: "=sum($.lines[*].amount)"},
				{Account: "702", Credit: "=sum($.lines[*].amount)"},
			},
		}},
	})
	evID, err := x.Store.AppendEvent(ctx, core.Event{
		Kind: core.KindRaw, Type: "invoice.received", OccurredAt: "2026-09-15",
		Payload: json.RawMessage(invoicePayload), DedupKey: "multi-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("multi-fire: booked %d, errs %v", n, errs)
	}
	if wl, _ := x.Store.UnmatchedRawEvents(ctx); len(wl) != 0 {
		t.Fatalf("worklist not cleared: %d", len(wl))
	}
	invoices, _ := x.Store.ObjectsByType(ctx, "invoice")
	postings, _ := x.Store.ObjectsByType(ctx, "posting")
	if len(invoices) != 1 || len(postings) != 2 {
		t.Fatalf("got %d invoices, %d postings; want 1 and 2", len(invoices), len(postings))
	}
	if invoices[0].RuleID != "book-invoice-document" || postings[0].RuleID != "post-invoice-ledger" {
		t.Fatalf("provenance: %s / %s", invoices[0].RuleID, postings[0].RuleID)
	}
	if postings[0].State["entry"] != fmt.Sprintf("entry-%d-post-invoice-ledger", evID) {
		t.Fatalf("entry key = %v", postings[0].State["entry"])
	}

	// Replay reproduces the multi-fired state.
	before, _ := x.Store.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := x.Store.AllObjects(ctx)
	if !reflect.DeepEqual(normalize(t, before), normalize(t, after)) {
		t.Fatal("replay diverged")
	}

	// Conflict: a second rule claiming the invoice document of the same
	// event. The whole event is refused — no document, no postings.
	activateRule("book-invoice-again", 300, core.RuleSpec{
		Match: core.Match{EventType: "invoice.received"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "invoice",
			Fields: map[string]string{
				"customer": "x", "issue_date": "=$.issue_date",
				"currency": "=$.currency", "total": "=sum($.lines[*].amount)"}}},
	})
	if _, err := x.Store.AppendEvent(ctx, core.Event{
		Kind: core.KindRaw, Type: "invoice.received", OccurredAt: "2026-09-16",
		Payload: json.RawMessage(invoicePayload), DedupKey: "multi-2",
	}); err != nil {
		t.Fatal(err)
	}
	n, errs := x.ProcessPending(ctx)
	if n != 0 || len(errs) == 0 || !strings.Contains(errs[0].Error(), "both materialize") {
		t.Fatalf("conflict: booked %d, errs %v", n, errs)
	}
	if objs, _ := x.Store.ObjectsByType(ctx, "posting"); len(objs) != 2 {
		t.Fatal("conflicting event booked postings")
	}
	if wl, _ := x.Store.UnmatchedRawEvents(ctx); len(wl) != 1 {
		t.Fatalf("conflicting event not left in worklist: %d", len(wl))
	}
}
