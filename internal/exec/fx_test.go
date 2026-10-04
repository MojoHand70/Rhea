package exec_test

// E3 of the enterprise-structure ladder (DIRECTION): foreign-currency events
// posting in transaction and functional currency. This file is the recorded
// fail-as-data attempt (DECISIONS 2026-10-04): the only way to book a EUR
// invoice into a PLN ledger with today's vocabulary is the precomputed-
// amounts contortion DIRECTION's network note already warns about — the
// event carries gross_pln/net_pln/vat_pln computed outside the system. It
// "works", and the explanation is hollow: the system's own rate table is
// decorative. Here it knows EUR/PLN = 4.3215 for the invoice date, the
// event claims amounts computed at 4.50, and the ledger agrees with the
// claim without a murmur. Provenance points at a raw event that carries its
// own conversion; replay reproduces the numbers by copying, not by
// explaining; no invariant can tie the PLN entry to the EUR document.

import (
	"context"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

func TestCurrencyAsPureData(t *testing.T) {
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
		{Name: "posting", Version: 2, Domain: "finance",
			Fields: []core.FieldDef{
				{Name: "entry", Type: "string", Required: true},
				{Name: "line", Type: "int", Required: true},
				{Name: "book", Type: "string", Required: true},
				{Name: "account", Type: "ref<account>", Required: true},
				{Name: "side", Type: "enum", Values: []string{"debit", "credit"}, Required: true},
				{Name: "amount", Type: "money", Required: true},
				{Name: "currency", Type: "string", Required: true},
				{Name: "date", Type: "date", Required: true},
			}},
		{Name: "fx_rate", Version: 1, Domain: "finance", LabelField: "code",
			Fields: []core.FieldDef{
				{Name: "code", Type: "string", Required: true}, // "EUR/PLN/2026-09-21"
				{Name: "base", Type: "string", Required: true},
				{Name: "quote", Type: "string", Required: true},
				{Name: "date", Type: "date", Required: true},
				{Name: "rate", Type: "string", Required: true},
			}},
	} {
		if err := x.Store.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}

	submit := func(dedup, evType, date, payload string) {
		t.Helper()
		if _, err := x.Store.AppendEvent(ctx, core.Event{
			Kind: core.KindRaw, Type: evType, OccurredAt: date,
			Payload: []byte(payload), DedupKey: dedup,
		}); err != nil {
			t.Fatal(err)
		}
	}
	activate := func(id string, spec core.RuleSpec) {
		t.Helper()
		if _, err := x.Store.InsertRuleVersion(ctx, core.Rule{
			ID: id, Status: core.StatusDraft, Priority: 100,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: spec,
		}); err != nil {
			t.Fatal(err)
		}
		if _, _, procErrs, err := x.ApproveRule(ctx, id, "krzysztof", "2026-08-31"); err != nil {
			t.Fatal(err)
		} else if len(procErrs) > 0 {
			t.Fatalf("approve %s: %v", id, procErrs)
		}
	}

	activate("register-account", core.RuleSpec{
		Match: core.Match{EventType: "account.created"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "account",
			Fields: map[string]string{"code": "=$.code", "name": "=$.name", "type": "=$.type"}}},
	})
	activate("register-fx-rate", core.RuleSpec{
		Match: core.Match{EventType: "fx.rate.published"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "fx_rate",
			Fields: map[string]string{"code": "=$.code", "base": "=$.base",
				"quote": "=$.quote", "date": "=$.date", "rate": "=$.rate"}}},
	})
	submit("acc-201", "account.created", "2026-09-01", `{"code":"201","name":"Rozrachunki","type":"debtor"}`)
	submit("acc-700", "account.created", "2026-09-01", `{"code":"700","name":"Sprzedaż","type":"revenue"}`)
	submit("acc-222", "account.created", "2026-09-01", `{"code":"222","name":"VAT należny","type":"tax"}`)
	// The system knows the NBP rate for the invoice date.
	submit("nbp-0921", "fx.rate.published", "2026-09-21",
		`{"code":"EUR/PLN/2026-09-21","base":"EUR","quote":"PLN","date":"2026-09-21","rate":"4.3215"}`)
	if n, errs := x.ProcessPending(ctx); n != 4 || len(errs) != 0 {
		t.Fatalf("master data: booked %d, errs %v", n, errs)
	}

	// The contortion: the functional entry books from amounts the event
	// brought with it. Nothing in the template references any rate.
	activate("pl-stat-post-eur", core.RuleSpec{
		Match: core.Match{EventType: "sale.recorded",
			Where: []core.Condition{{Path: "$.currency", Op: "eq", Value: "EUR"}}},
		Effect: core.Effect{Postings: &core.PostingsTemplate{
			Book: "pl-stat", Currency: "PLN",
			Lines: []core.PostingLine{
				{Account: "201", Debit: "=$.gross_pln"},
				{Account: "700", Credit: "=$.net_pln"},
				{Account: "222", Credit: "=$.vat_pln"},
			},
		}},
	})

	// A EUR invoice whose PLN amounts were computed outside the system — at
	// 4.50, contradicting the rate the system itself holds for that date.
	submit("sale-eur", "sale.recorded", "2026-09-21",
		`{"currency":"EUR","net":"200.00","vat":"46.00","gross":"246.00",
		  "net_pln":"900.00","vat_pln":"207.00","gross_pln":"1107.00"}`)
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("eur sale: booked %d, errs %v", n, errs)
	}

	// The ledger agrees with the event's claim, not with the system's own
	// knowledge: 1107.00 PLN booked while the known rate says 1063.09.
	postings, _ := x.Store.ObjectsByType(ctx, "posting")
	var grossPLN int64
	for _, p := range postings {
		if p.State["side"] == "debit" {
			grossPLN += int64(p.State["amount"].(float64))
		}
	}
	if grossPLN != 110700 {
		t.Fatalf("booked %d, expected the event's unverified claim 110700", grossPLN)
	}
	// And the EUR document amounts are nowhere in the ledger: the entry has
	// no transaction-currency trace, so no invariant can ever tie 1107.00
	// PLN back to 246.00 EUR, let alone to a rate.
	for _, p := range postings {
		if p.State["currency"] != "PLN" {
			t.Fatalf("posting currency %v", p.State["currency"])
		}
		if _, has := p.State["tx_amount"]; has {
			t.Fatal("posting carries a transaction amount — the vocabulary grew; rewrite this attempt")
		}
	}
}
