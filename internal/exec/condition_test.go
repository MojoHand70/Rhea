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

// A condition that reads through a link (DECISIONS 2026-10-16): the Poland
// pack serves a sole trader and a company by what the company says it is, so
// a rule must be able to ask the record the event names — "the seller keeps
// KPiR" — not only the event. The condition is a formula over the state the
// effect reads, at the same point in the log: a company that changes its
// books changes which rule explains the next sale, every earlier sale keeps
// its explanation, and replay decides each condition the same way.
func TestConditionReadsThroughALink(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	for _, ot := range []core.ObjectType{
		{Name: "company", Version: 1, Domain: "finance", LabelField: "name",
			Fields: []core.FieldDef{
				{Name: "name", Type: "string", Required: true},
				{Name: "vat_id", Type: "string", Required: true},
				{Name: "books", Type: "enum", Values: []string{"full", "kpir"}},
			},
			Amendable: []string{"books"}},
		{Name: "ledger_sale", Version: 1, Domain: "finance", LabelField: "number",
			Fields: []core.FieldDef{{Name: "number", Type: "string", Required: true}}},
		{Name: "kpir_entry", Version: 1, Domain: "finance", LabelField: "number",
			Fields: []core.FieldDef{{Name: "number", Type: "string", Required: true}}},
	} {
		if err := x.Store.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	// A company that never declared its books keeps them full — the formula
	// names that stand-in; a field read without one would be an error.
	keeps := func(books string) core.Condition {
		return core.Condition{Formula: `=coalesce(ref(company, vat_id, $.seller_nip).books, "full") = "` + books + `"`}
	}
	specs := map[string]core.RuleSpec{
		"register-company": {Match: core.Match{EventType: "company.registered"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "company",
				Fields: map[string]string{"name": "=$.name", "vat_id": "=$.vat_id", "books": "=$.books"}}}},
		"list-company": {Match: core.Match{EventType: "company.listed"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "company",
				Fields: map[string]string{"name": "=$.name", "vat_id": "=$.vat_id"}}}},
		"declare-books": {Match: core.Match{EventType: "company.books.declared"},
			Effect: core.Effect{Amend: &core.AmendTemplate{Type: "company",
				Target: "=ref(company, vat_id, $.vat_id)", Set: map[string]string{"books": "=$.books"}}}},
		"sale-in-the-ledger": {Match: core.Match{EventType: "sale.recorded", Where: []core.Condition{keeps("full")}},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "ledger_sale",
				Fields: map[string]string{"number": "=$.number"}}}},
		"sale-in-kpir": {Match: core.Match{EventType: "sale.recorded", Where: []core.Condition{keeps("kpir")}},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "kpir_entry",
				Fields: map[string]string{"number": "=$.number"}}}},
	}
	catalog := func(name string) (core.ObjectType, bool) {
		ot, err := x.Store.GetObjectType(ctx, name)
		return ot, err == nil
	}
	for id, spec := range specs {
		if err := spec.Check(catalog); err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		activateRule(t, x, id, 100, spec)
	}

	submitRaw(t, x, "c", "company.registered", "2026-01-02", `{"name":"Nowak","vat_id":"5250000001","books":"full"}`)
	submitRaw(t, x, "s1", "sale.recorded", "2026-03-10", `{"number":"FV/1","seller_nip":"5250000001"}`)
	submitRaw(t, x, "d", "company.books.declared", "2026-04-01", `{"vat_id":"5250000001","books":"kpir"}`)
	submitRaw(t, x, "s2", "sale.recorded", "2026-04-10", `{"number":"FV/2","seller_nip":"5250000001"}`)
	// A seller nobody registered: the link does not resolve, so the sale waits
	// rather than land in either book.
	submitRaw(t, x, "s3", "sale.recorded", "2026-04-11", `{"number":"FV/3","seller_nip":"9999999999"}`)
	// A company that never said: the stand-in decides.
	submitRaw(t, x, "l", "company.listed", "2026-04-12", `{"name":"Kowalska","vat_id":"5250000002"}`)
	submitRaw(t, x, "s4", "sale.recorded", "2026-04-13", `{"number":"K/1","seller_nip":"5250000002"}`)
	if n, errs := x.ProcessPending(ctx); n != 6 || len(errs) != 1 || !strings.Contains(errs[0].Error(), "no company") {
		t.Fatalf("booked %d, errs %v", n, errs)
	}
	ledger, _ := x.Store.ObjectsByType(ctx, "ledger_sale")
	kpir, _ := x.Store.ObjectsByType(ctx, "kpir_entry")
	if len(ledger) != 2 || ledger[0].State["number"] != "FV/1" || ledger[0].RuleID != "sale-in-the-ledger" || ledger[1].State["number"] != "K/1" ||
		len(kpir) != 1 || kpir[0].State["number"] != "FV/2" || kpir[0].RuleID != "sale-in-kpir" {
		t.Fatalf("ledger %+v, kpir %+v", ledger, kpir)
	}
	if wl, _ := x.Store.UnmatchedRawEvents(ctx); len(wl) != 1 {
		t.Fatalf("the unregistered seller's sale waits: %v", wl)
	}

	// Determinism (invariant 4): the conditions read the state of their
	// point in the log, so replay puts each sale in the same book.
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

	// The laws: a condition is true or false, a formula or a path, and what
	// it reads through a link must be a field the record declares.
	for _, bad := range []core.Condition{
		{Formula: `=ref(company, vat_id, $.seller_nip).name`},
		{Formula: `=1 + 2`},
		{Formula: `=$.x = 1`, Path: "$.x", Op: "eq"},
		{Formula: `ref(company, vat_id, $.seller_nip).books = "kpir"`},
	} {
		spec := core.RuleSpec{Match: core.Match{EventType: "sale.recorded", Where: []core.Condition{bad}},
			Effect: specs["sale-in-kpir"].Effect}
		if err := spec.Check(catalog); err == nil {
			t.Fatalf("condition %+v was admitted", bad)
		}
	}
	spec := core.RuleSpec{Match: core.Match{EventType: "sale.recorded",
		Where: []core.Condition{{Formula: `=ref(company, vat_id, $.seller_nip).method = "cash"`}}},
		Effect: specs["sale-in-kpir"].Effect}
	if err := spec.Check(catalog); err == nil || !strings.Contains(err.Error(), "method") {
		t.Fatalf("a field the company does not declare: %v", err)
	}
}
