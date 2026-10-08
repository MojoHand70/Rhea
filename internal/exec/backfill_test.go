package exec_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

// Can we add to the past when the past is closed? (KK, 2026-10-08.) The log
// never closes; periods do. Invoices from a September that is closed for the
// statutory book get a statutory posting rule only in October: backfill never
// writes into the closed month — it offers those chains forward, dated on the
// backfill, still caused by their September facts. The open book's September
// takes its consequences where they belong. And a late amendment — the
// eval's forcing case — reaches the case that an earlier rule already left.
func TestBackfillOffersClosedPeriodsForward(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	for _, ot := range []core.ObjectType{
		{Name: "account", Version: 1, Domain: "finance", LabelField: "name", Fields: []core.FieldDef{
			{Name: "code", Type: "string", Required: true}, {Name: "name", Type: "string", Required: true}}},
		{Name: "posting", Version: 2, Domain: "finance", Fields: []core.FieldDef{
			{Name: "entry", Type: "string", Required: true}, {Name: "line", Type: "int", Required: true},
			{Name: "book", Type: "string", Required: true}, {Name: "account", Type: "ref<account>", Required: true},
			{Name: "side", Type: "enum", Values: []string{"debit", "credit"}, Required: true},
			{Name: "amount", Type: "money", Required: true}, {Name: "currency", Type: "string", Required: true},
			{Name: "date", Type: "date", Required: true}}},
		{Name: "period_lock", Version: 2, Domain: "finance", Fields: []core.FieldDef{
			{Name: "month", Type: "string", Required: true}, {Name: "book", Type: "string", Required: true}}},
		{Name: "invoice", Version: 1, Domain: "finance", IsDocument: true, Fields: []core.FieldDef{
			{Name: "number", Type: "string", Required: true}, {Name: "currency", Type: "string", Required: true},
			{Name: "total", Type: "money", Required: true}}},
	} {
		if err := s.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	activate := func(id string, priority int, spec core.RuleSpec) {
		t.Helper()
		if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: id, Status: core.StatusDraft, Priority: priority,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: spec}); err != nil {
			t.Fatal(err)
		}
		if _, _, errs, err := x.ApproveRule(ctx, id, "krzysztof", "2026-09-01"); err != nil || len(errs) > 0 {
			t.Fatalf("approve %s: %v %v", id, err, errs)
		}
	}
	submit := func(typ, date, payload string) {
		t.Helper()
		var p map[string]any
		json.Unmarshal([]byte(payload), &p)
		if _, err := x.TriggerActivity(ctx, "submit_event", map[string]any{
			"event_type": typ, "occurred_at": date, "payload": p}, "test", ""); err != nil {
			t.Fatal(err)
		}
	}
	activate("register-account", 10, core.RuleSpec{Match: core.Match{EventType: "account.created"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "account",
			Fields: map[string]string{"code": "=$.code", "name": "=$.name"}}}})
	activate("lock-period", 10, core.RuleSpec{Match: core.Match{EventType: "period.locked"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "period_lock",
			Fields: map[string]string{"month": "=$.month", "book": "=$.book"}}}})
	activate("book-invoice", 100, core.RuleSpec{Match: core.Match{EventType: "invoice.received"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "invoice",
			Fields: map[string]string{"number": "=$.number", "currency": "=$.currency", "total": "=$.total"}}}})
	submit("account.created", "2026-09-01", `{"code":"201","name":"Receivables"}`)
	submit("account.created", "2026-09-01", `{"code":"702","name":"Revenue"}`)
	submit("invoice.received", "2026-09-15", `{"number":"FV-1","currency":"PLN","total":"100.00"}`)
	submit("invoice.received", "2026-09-20", `{"number":"FV-2","currency":"PLN","total":"250.00"}`)
	submit("period.locked", "2026-09-30", `{"month":"2026-09","book":"pl-stat"}`)

	// October: two posting rules arrive, cascading off every invoice.
	post := func(book string) core.RuleSpec {
		return core.RuleSpec{
			Match: core.Match{EventType: core.EventObjectMaterialized, Where: []core.Condition{
				{Path: "$.object_type", Op: "eq", Value: "invoice"}}},
			Effect: core.Effect{Postings: &core.PostingsTemplate{Book: book, Currency: "=$.state.currency",
				Lines: []core.PostingLine{{Account: "201", Debit: "=$.state.total"}, {Account: "702", Credit: "=$.state.total"}}}}}
	}
	activate("post-invoice-main", 200, post("main"))
	activate("post-invoice-stat", 210, post("pl-stat"))
	if postings, _ := s.ObjectsByType(ctx, "posting"); len(postings) != 0 {
		t.Fatalf("postings = %d before backfill", len(postings))
	}

	plan, err := x.PlanBackfill(ctx, "2026-10-08")
	if err != nil {
		t.Fatal(err)
	}
	// Per invoice: main is open in September, pl-stat is closed — one chain
	// carries both, so the chain is offered forward whole and says why.
	if len(plan.Chains) != 2 || len(plan.Errors) != 0 {
		t.Fatalf("plan: %+v", plan)
	}
	for _, c := range plan.Chains {
		if c.BookedAt != "2026-10-08" || !strings.Contains(c.Forwarded, "2026-09 is locked for book pl-stat") {
			t.Fatalf("chain not offered forward: %+v", c)
		}
	}
	n, errs, err := x.ApproveBackfill(ctx, "krzysztof", "2026-10-08")
	if err != nil || n != 2 || len(errs) != 0 {
		t.Fatalf("backfill: %d %v %v", n, errs, err)
	}
	postings, _ := s.ObjectsByType(ctx, "posting")
	if len(postings) != 8 {
		t.Fatalf("postings = %d, want 2 invoices × 2 books × 2 lines", len(postings))
	}
	for _, p := range postings {
		if p.State["date"] != "2026-10-08" {
			t.Fatalf("posting into the closed past: %v", p.State)
		}
		// The October entry still walks back to its September fact.
		root, _ := s.GetEvent(ctx, p.SourceEventID)
		for root.CauseEventID != nil {
			root, _ = s.GetEvent(ctx, *root.CauseEventID)
		}
		if root.Type != "invoice.received" || !strings.HasPrefix(root.OccurredAt, "2026-09") {
			t.Fatalf("posting caused by %s on %s", root.Type, root.OccurredAt)
		}
	}
	assertReplayIdentical(t, x)
}

// The eval's forcing case: the withdrawal was explained once (the complaint
// marked withdrawn); the case-resolving rule approved later reaches it only
// through backfill — and the earlier amendment is replayed, not re-judged.
func TestBackfillLateAmendment(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	lc := &core.LifecycleDef{Field: "status", Transitions: map[string][]string{"open": {"closed"}}}
	for _, ot := range []core.ObjectType{
		{Name: "complaint", Version: 1, Domain: "sales", Fields: []core.FieldDef{
			{Name: "number", Type: "string", Required: true},
			{Name: "status", Type: "enum", Values: []string{"open", "closed"}, Required: true}}, Lifecycle: lc},
		{Name: "case", Version: 1, Domain: "work", Fields: []core.FieldDef{
			{Name: "subject", Type: "string", Required: true},
			{Name: "status", Type: "enum", Values: []string{"open", "closed"}, Required: true}}, Lifecycle: lc},
	} {
		if err := s.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	rules := []struct {
		id   string
		prio int
		spec core.RuleSpec
	}{
		{"book-complaint", 100, core.RuleSpec{Match: core.Match{EventType: "complaint.registered"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "complaint",
				Fields: map[string]string{"number": "=$.number", "status": "open"}}}}},
		{"raise-case", 200, core.RuleSpec{Match: core.Match{EventType: core.EventObjectMaterialized,
			Where: []core.Condition{{Path: "$.object_type", Op: "eq", Value: "complaint"}}},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "case",
				Fields: map[string]string{"subject": "=$.state.number", "status": "open"}}}}},
		{"withdraw-complaint", 300, core.RuleSpec{Match: core.Match{EventType: "complaint.withdrawn"},
			Effect: core.Effect{Amend: &core.AmendTemplate{Type: "complaint",
				Target: "=ref(complaint, number, $.number)", Set: map[string]string{"status": "closed"}}}}},
	}
	for _, r := range rules {
		if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: r.id, Status: core.StatusActive, Priority: r.prio,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: r.id, Spec: r.spec}); err != nil {
			t.Fatal(err)
		}
	}
	for _, e := range [][3]string{
		{"complaint.registered", "2026-09-27", `{"number":"R-1"}`},
		{"complaint.withdrawn", "2026-09-29", `{"number":"R-1"}`},
	} {
		var p map[string]any
		json.Unmarshal([]byte(e[2]), &p)
		if _, err := x.TriggerActivity(ctx, "submit_event", map[string]any{
			"event_type": e[0], "occurred_at": e[1], "payload": p}, "test", ""); err != nil {
			t.Fatal(err)
		}
	}
	// Understanding arrives: a withdrawal also closes the case.
	if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: "close-case-on-withdrawal", Status: core.StatusActive,
		Priority: 310, EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: "close the case",
		Spec: core.RuleSpec{Match: core.Match{EventType: "complaint.withdrawn"},
			Effect: core.Effect{Amend: &core.AmendTemplate{Type: "case",
				Target: "=ref(case, subject, $.number)", Set: map[string]string{"status": "closed"}}}}}); err != nil {
		t.Fatal(err)
	}
	plan, err := x.PlanBackfill(ctx, "2026-10-08")
	if err != nil || len(plan.Chains) != 1 || len(plan.Diff.Changed) != 1 || len(plan.Errors) != 0 {
		t.Fatalf("plan: %+v, %v", plan, err)
	}
	if _, _, err := x.ApproveBackfill(ctx, "krzysztof", "2026-10-08"); err != nil {
		t.Fatal(err)
	}
	cases, _ := s.ObjectsByType(ctx, "case")
	if len(cases) != 1 || cases[0].State["status"] != "closed" {
		t.Fatalf("case = %+v", cases)
	}
	assertReplayIdentical(t, x)
}
