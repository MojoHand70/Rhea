package shell_test

import (
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"rhea/internal/core"
	"rhea/internal/pack"
)

// TestFXRevaluation (the auditor's exam, 2026-10-20, lesson I-28 of
// financial accounting 2025 generalised from shares to every asset and
// liability in a foreign currency; EXAM.md 6.1.1 c, 6.1.3 a): at the balance
// date the open receivables and payables in a currency are measured at the
// NBP average rate in force that day (UoR art. 30 ust. 1 pkt 1), the
// difference against the documents' rates is financial revenue or cost (ust.
// 4), and on the first day of the next year the storno gives the documents
// their rates back so that a payment's realised difference is measured from
// the document. Złoty items never enter; a currency with nothing open books
// nothing.
func TestFXRevaluation(t *testing.T) {
	w := newCITWorld(t)
	ctx, s, x, door, submit := w.ctx, w.s, w.x, w.door, w.submit
	door("declare_company_form", map[string]any{"company": w.self, "vat_status": "active", "books": "full", "income_tax": "cit",
		"from": "2026-01-01", "occurred_at": "2026-08-01"})
	// The NBP channel's rule registers rates; the test publishes them.
	sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", "pl", "nbp.json"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-09-01"); err != nil {
		t.Fatal(err)
	}
	rate := func(code, day, rate string) {
		t.Helper()
		submit("fx.rate.published", day, "nbp-"+code+"-"+day, map[string]any{"code": code + "/PLN/" + day, "base": code, "quote": "PLN", "date": day, "rate": rate, "market": "pl"})
	}
	rate("EUR", "2026-10-02", "4.30")
	rate("EUR", "2026-11-15", "4.40")
	rate("EUR", "2027-01-01", "4.50") // the table of 31 December: the channel keeps it under the next day
	rate("USD", "2027-01-01", "3.90")
	sums := func() map[string]int64 {
		t.Helper()
		accounts, _ := s.ObjectsByType(ctx, "account")
		code := map[string]string{}
		for _, a := range accounts {
			code[a.ID] = fmt.Sprint(a.State["code"])
		}
		ps, _ := s.ObjectsByType(ctx, "posting")
		out := map[string]int64{}
		for _, p := range ps {
			n, _ := p.State["amount"].(float64)
			out[code[fmt.Sprint(p.State["account"])]+" "+fmt.Sprint(p.State["side"])] += int64(n)
		}
		return out
	}
	delta := func(before, after map[string]int64) map[string]int64 {
		d := map[string]int64{}
		for k, v := range after {
			if v-before[k] != 0 {
				d[k] = v - before[k]
			}
		}
		return d
	}
	money := func(o core.Object, f string) int64 {
		n, _ := o.State[f].(float64)
		return int64(n)
	}
	sale := func(key, number, date, net, vat, gross, currency string) {
		t.Helper()
		submit("sales.invoice.issued", date, key, map[string]any{"number": number, "issue_date": date, "market": "pl", "intragroup": "no",
			"seller_nip": "5250001111", "buyer_nip": "5260001246", "net": net, "vat": vat, "gross": gross, "vat_rate": "23", "currency": currency})
	}
	// A sale of 1 230 EUR at 4,30 on 2 October, a purchase of 3 690 EUR at 4,40
	// in November, and a sale in złoty that the revaluation must not touch.
	sale("s-eur", "FV 7/2026", "2026-10-02", "1000.00", "230.00", "1230.00", "EUR")
	submit("purchase.invoice.received", "2026-11-15", "p-eur", map[string]any{"number": "INV-EU-11", "invoice_date": "2026-11-15", "market": "pl",
		"supplier_nip": "5260001246", "buyer_nip": "5250001111", "net": "3000.00", "vat": "690.00", "gross": "3690.00", "vat_rate": "23", "currency": "EUR"})
	sale("s-pln", "FV 8/2026", "2026-12-10", "500.00", "115.00", "615.00", "PLN")

	before := sums()
	door("revalue_fx_balances", map[string]any{"balance_date": "2026-12-31", "currency": "EUR", "occurred_at": "2026-12-31"})
	revs := mustObjects(t, s, ctx, "fx_revaluation")
	if len(revs) != 1 {
		t.Fatalf("one revaluation per currency: %v", revs)
	}
	r := revs[0]
	if money(r, "receivables_open") != 123000 || money(r, "receivables_booked_pln") != 528900 || money(r, "receivables_closing_pln") != 553500 || money(r, "receivables_difference") != 24600 {
		t.Fatalf("the receivables side at 4,50 against 4,30: %v", r.State)
	}
	if money(r, "payables_open") != 369000 || money(r, "payables_booked_pln") != 1623600 || money(r, "payables_closing_pln") != 1660500 || money(r, "payables_difference") != 36900 {
		t.Fatalf("the payables side at 4,50 against 4,40: %v", r.State)
	}
	if r.State["storno_date"] != "2027-01-01" || r.State["balance_date"] != "2026-12-31" {
		t.Fatalf("the dates: %v", r.State)
	}
	// The rate rose: a gain on what customers owe, a loss on what we owe —
	// and each booked back on 1 January.
	want := map[string]int64{"201 debit": 24600, "750 credit": 24600, "750 debit": 24600, "201 credit": 24600,
		"751 debit": 36900, "202 credit": 36900, "202 debit": 36900, "751 credit": 36900}
	if got := delta(before, sums()); !reflect.DeepEqual(got, want) {
		t.Fatalf("revaluation and storno: got %v want %v", got, want)
	}
	ps, _ := s.ObjectsByType(ctx, "posting")
	dated := map[string]int{}
	for _, p := range ps {
		if p.State["date"] == "2026-12-31" || p.State["date"] == "2027-01-01" {
			dated[fmt.Sprint(p.State["date"])]++
		}
	}
	if dated["2026-12-31"] != 4 || dated["2027-01-01"] != 4 {
		t.Fatalf("four lines on the balance date and four on the storno day: %v", dated)
	}
	// Złoty is not revalued: nothing booked, no object.
	before = sums()
	x.TriggerActivity(ctx, "revalue_fx_balances", map[string]any{"balance_date": "2026-12-31", "currency": "PLN", "occurred_at": "2026-12-31"}, "krzysztof", "")
	if got := delta(before, sums()); len(got) != 0 || len(mustObjects(t, s, ctx, "fx_revaluation")) != 1 {
		t.Fatalf("złoty revalued: %v", got)
	}
	// A currency with nothing open: a revaluation of zeros, nothing booked.
	door("revalue_fx_balances", map[string]any{"balance_date": "2026-12-31", "currency": "USD", "occurred_at": "2026-12-31"})
	if got := delta(before, sums()); len(got) != 0 {
		t.Fatalf("an empty currency booked something: %v", got)
	}
	for _, o := range mustObjects(t, s, ctx, "fx_revaluation") {
		if o.State["currency"] == "USD" && (money(o, "receivables_open") != 0 || money(o, "payables_difference") != 0) {
			t.Fatalf("USD: %v", o.State)
		}
	}
}
