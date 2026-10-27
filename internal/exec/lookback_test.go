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

// TestTheLookBackSaysWhatItAdds (KK 2026-10-27): a person asks Rhea to look
// back over the documents already in the books under the rules approved
// since, and decides on what she says it would do — so the dry run says it
// the way a person decides it: how many records of each kind it adds, and
// the entries it would book, by month and how much, or none. A rule that
// only recognises customers in old invoices adds two customers and books
// nothing; a posting rule approved later books three entries in two months
// and adds no record a person would count.
func TestTheLookBackSaysWhatItAdds(t *testing.T) {
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
		{Name: "invoice", Version: 1, Domain: "finance", IsDocument: true, Fields: []core.FieldDef{
			{Name: "number", Type: "string", Required: true}, {Name: "buyer", Type: "string", Required: true},
			{Name: "currency", Type: "string", Required: true}, {Name: "total", Type: "money", Required: true}}},
		{Name: "customer", Version: 1, Domain: "sales", LabelField: "name", Fields: []core.FieldDef{
			{Name: "name", Type: "string", Required: true}}},
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
		if _, _, errs, err := x.ApproveRule(ctx, id, "krzysztof", "2026-10-01"); err != nil || len(errs) > 0 {
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
	activate("book-invoice", 100, core.RuleSpec{Match: core.Match{EventType: "invoice.received"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "invoice",
			Fields: map[string]string{"number": "=$.number", "buyer": "=$.buyer", "currency": "=$.currency", "total": "=$.total"}}}})
	submit("account.created", "2026-09-01", `{"code":"201","name":"Receivables"}`)
	submit("account.created", "2026-09-01", `{"code":"702","name":"Revenue"}`)
	submit("invoice.received", "2026-09-15", `{"number":"FV-1","buyer":"Agencja Nova","currency":"PLN","total":"100.00"}`)
	submit("invoice.received", "2026-09-20", `{"number":"FV-2","buyer":"Agencja Nova","currency":"PLN","total":"250.00"}`)
	submit("invoice.received", "2026-10-02", `{"number":"FV-3","buyer":"Pod Lipą","currency":"PLN","total":"40.00"}`)

	// A rule that recognises a firm's first invoice as a new customer.
	activate("customer-from-invoice", 200, core.RuleSpec{
		Match: core.Match{EventType: core.EventObjectMaterialized, Where: []core.Condition{
			{Path: "$.object_type", Op: "eq", Value: "invoice"},
			{Formula: "=count(c in objects(customer, name, $.state.buyer)) = 0"}}},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "customer", Fields: map[string]string{"name": "=$.state.buyer"}}}})
	plan, err := x.PlanBackfill(ctx, "2026-10-27")
	if err != nil {
		t.Fatal(err)
	}
	if want := (exec.BackfillSummary{Records: []exec.RecordCount{{Type: "customer", Added: 2}}, Entries: []exec.MonthEntries{}}); !reflect.DeepEqual(plan.Summary, want) {
		t.Fatalf("records only: %+v, want %+v", plan.Summary, want)
	}
	if n, errs, err := x.ApproveBackfillPlan(ctx, "krzysztof", "2026-10-27", plan.Fingerprint); err != nil || len(errs) != 0 || n != 2 {
		t.Fatalf("approve the records: %d %v %v", n, errs, err)
	}

	// A posting rule approved later books the old invoices: entries, by month.
	activate("post-invoice", 300, core.RuleSpec{
		Match: core.Match{EventType: core.EventObjectMaterialized, Where: []core.Condition{
			{Path: "$.object_type", Op: "eq", Value: "invoice"}}},
		Effect: core.Effect{Postings: &core.PostingsTemplate{Book: "main", Currency: "=$.state.currency",
			Lines: []core.PostingLine{{Account: "201", Debit: "=$.state.total"}, {Account: "702", Credit: "=$.state.total"}}}}})
	if plan, err = x.PlanBackfill(ctx, "2026-10-27"); err != nil {
		t.Fatal(err)
	}
	want := exec.BackfillSummary{Records: []exec.RecordCount{}, Entries: []exec.MonthEntries{
		{Month: "2026-09", Book: "main", Currency: "PLN", Entries: 2, Debit: "350.00"},
		{Month: "2026-10", Book: "main", Currency: "PLN", Entries: 1, Debit: "40.00"}}}
	if !reflect.DeepEqual(plan.Summary, want) {
		t.Fatalf("entries: %+v, want %+v", plan.Summary, want)
	}
	if n, errs, err := x.ApproveBackfillPlan(ctx, "krzysztof", "2026-10-27", plan.Fingerprint); err != nil || len(errs) != 0 || n != 3 {
		t.Fatalf("approve the entries: %d %v %v", n, errs, err)
	}
	if plan, err = x.PlanBackfill(ctx, "2026-10-27"); err != nil || len(plan.Chains) != 0 || len(plan.Summary.Records)+len(plan.Summary.Entries) != 0 {
		t.Fatalf("looked back twice: %+v %v", plan.Summary, err)
	}
	assertReplayIdentical(t, x)
}
