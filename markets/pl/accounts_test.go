package pl_test

import (
	"context"
	"fmt"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

// Where purchases and penny differences land (the Poland pack review, FB-06,
// FB-14, MON-19; pack v20): a service or an expense is a cost — on the
// supplier's declared cost account, or 403 when none is declared — and only
// a purchase with a delivery note waits on 300 for its receipt; a shortfall
// within tolerance is another operating cost (761), an overpayment another
// operating income (760), and undoing the match takes the difference back
// through the account it came in on.
func TestWherePurchasesAndPenniesLand(t *testing.T) {
	ctx := context.Background()
	at := store.WithCompany(ctx, selfCo) // the books are a company's: its account approved by its manager
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	installPoland(t, s, x)
	registerPartiesAt(t, s, x, selfCo)
	ownAccount(t, s, x, selfCo, "PL83101010230000261395100000", "PLN") // its statement lines wait for nothing (Poland pack v54)
	appendAt(t, s, x, selfCo, "company.registered", "2026-09-01", "fuel", map[string]any{"name": "Stacja Paliw sp. z o.o.", "kind": "supplier",
		"vat_id": "PL7740001454", "country": "PL", "iban": "PL27114020040000300201355387", "address": "ul. Płocka 1, 09-400 Płock"})
	ids := map[string]string{}
	companies, _ := s.ObjectsByType(ctx, "company")
	for _, c := range companies {
		ids[c.State["vat_id"].(string)] = c.ID
	}

	// The fuel station's invoices are fuel: 402. An account outside the
	// chart is not taken — the declaration waits.
	if tr, err := x.TriggerActivity(at, "set_cost_account", map[string]any{"company": ids["PL7740001454"], "account": "999", "occurred_at": "2026-09-01"}, "krzysztof", ""); err != nil || tr.Booked != 0 {
		t.Fatalf("an account outside the chart: %+v %v", tr, err)
	}
	if tr, err := x.TriggerActivity(at, "set_cost_account", map[string]any{"company": ids["PL7740001454"], "account": "402", "occurred_at": "2026-09-01"}, "krzysztof", ""); err != nil || tr.Booked != 1 {
		t.Fatalf("set_cost_account: %+v %v", tr, err)
	}
	purchase := func(dedup, supplierNIP, number string, extra map[string]any) {
		t.Helper()
		p := map[string]any{"number": number, "invoice_date": "2026-09-10", "supplier_nip": supplierNIP, "buyer_nip": "PL8058083055",
			"net": "100.00", "vat": "23.00", "gross": "123.00", "vat_rate": "23", "currency": "PLN", "market": "pl"}
		for k, v := range extra {
			p[k] = v
		}
		appendAt(t, s, x, selfCo, "purchase.invoice.received", "2026-09-10", dedup, p)
	}
	purchase("fuel-1", "PL7740001454", "PAL/1", nil)                                     // declared: 402
	purchase("svc-1", "PL5265877635", "USL/1", nil)                                      // undeclared: 403
	purchase("goods-1", "PL5265877635", "TOW/1", map[string]any{"for_document": "WZ 7"}) // a delivery: 300
	if got := debits(t, s); got["402"] != 10000 || got["403"] != 10000 || got["300"] != 10000 || got["221"] != 6900 {
		t.Fatalf("purchase debits by account: %v", got)
	}

	// Two sales and their payments: one a grosz short, one two grosze over.
	for i, gross := range []string{"100.00", "50.00"} {
		appendAt(t, s, x, selfCo, "sales.invoice.issued", "2026-09-12", fmt.Sprintf("fv-%d", i), map[string]any{
			"number": fmt.Sprintf("FV/%d/2026", i+1), "issue_date": "2026-09-12", "seller_nip": "PL8058083055", "buyer_nip": "PL5265877635",
			"net": gross, "vat": "0.00", "gross": gross, "vat_rate": "zw", "currency": "PLN", "intragroup": "no", "market": "pl"})
	}
	for i, paid := range []string{"99.99", "50.02"} {
		appendAt(t, s, x, selfCo, "bank.statement.line", "2026-09-20", fmt.Sprintf("pay-%d", i), map[string]any{
			"date": "2026-09-20", "amount": paid, "currency": "PLN", "direction": "credit", "reference": fmt.Sprintf("FV/%d/2026", i+1),
			"counterparty": "Nabywca S.A.", "counterparty_account": "PL60102010260000042270201111", "account": "PL83101010230000261395100000", "market": "pl"})
	}
	if c, d := credits(t, s), debits(t, s); d["761"] != 1 || c["760"] != 2 || c["756"] != 0 || d["756"] != 0 {
		t.Fatalf("penny differences: debits %v, credits %v", d, c)
	}

	// Undo the short payment's match: its grosz leaves 761 the way it came.
	derived, _ := s.EventsByKind(ctx, core.KindDerived)
	var short int64
	for _, ev := range derived {
		if ev.Type == core.EventObjectsMatched && short == 0 {
			short = ev.ID
		}
	}
	if _, err := x.TriggerActivity(at, "unpair_items", map[string]any{"match_event": short, "reason": "wrong invoice"}, "shell:krzysztof", ""); err != nil {
		t.Fatal(err)
	}
	if c, d := credits(t, s), debits(t, s); d["761"] != 1 || c["761"] != 1 || c["760"] != 2 {
		t.Fatalf("after the unpairing: debits %v, credits %v", d, c)
	}
}

func debits(t *testing.T, s *store.Store) map[string]int64  { return bySide(t, s, "debit") }
func credits(t *testing.T, s *store.Store) map[string]int64 { return bySide(t, s, "credit") }

// bySide sums the statutory book's postings on one side by account code,
// in minor units.
func bySide(t *testing.T, s *store.Store, side string) map[string]int64 {
	t.Helper()
	ctx := context.Background()
	accounts, _ := s.ObjectsByType(ctx, "account")
	code := map[string]string{}
	for _, a := range accounts {
		code[a.ID] = fmt.Sprint(a.State["code"])
	}
	postings, _ := s.ObjectsByType(ctx, exec.PostingObjectType)
	out := map[string]int64{}
	for _, p := range postings {
		if p.State["side"] != side || p.State["book"] != "pl-stat" {
			continue
		}
		amount, ok := p.State["amount"].(float64)
		if !ok {
			t.Fatalf("posting amount %v", p.State["amount"])
		}
		out[code[fmt.Sprint(p.State["account"])]] += int64(amount)
	}
	return out
}
