package pl_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"rhea/internal/adapter"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
	"rhea/markets/pl"
)

// Currencies (the Poland pack review, VAT-18, KSEF-10, MON-16, MON-17,
// FB-13; pack v22): an invoice in euro carries its net and VAT in złoty at
// the NBP average of the last working day before the sale; the books convert
// at that same rate, a payment at the rate of its own day, and the gap is a
// realised exchange difference — a gain on what a customer owes when the
// rate rose (750), a loss on what we owe (751) — which an unpairing takes
// back through the same account. The settlement, the JPK and the invoice
// sent to KSeF count the VAT in złoty.
func TestAnInvoiceInEuro(t *testing.T) {
	ctx := context.Background()
	at := store.WithCompany(ctx, selfCo) // the books are a company's: its account approved by its manager
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	installPoland(t, s, x)
	sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", "pl", "nbp.json"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-09-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve: %v %v", err, errs)
	}
	selfID := registerPartiesAt(t, s, x, selfCo)
	ownAccount(t, s, x, selfCo, "PL83101010230000261395100000", "EUR") // its statement lines wait for nothing (Poland pack v54)
	if tr, err := x.TriggerActivity(at, "identify_company", map[string]any{"company": selfID, "email": "biuro@rhea-test.pl", "tax_office": "1471", "legal_form": "legal_entity", "occurred_at": "2026-09-01"}, "krzysztof", ""); err != nil || tr.Booked != 1 {
		t.Fatalf("identify_company: %+v %v", tr, err)
	}
	eur := func(no, date, mid string) pl.NBPTable {
		return pl.NBPTable{No: no, EffectiveDate: date, Rates: []pl.NBPRate{{Code: "EUR", Mid: jsonNumber(mid)}}}
	}
	nbp := pl.NBP{Client: pl.FakeNBP{Tables_: []pl.NBPTable{
		eur("187/A/NBP/2026", "2026-09-25", "4.2600"), eur("188/A/NBP/2026", "2026-09-28", "4.2650"),
		eur("189/A/NBP/2026", "2026-09-29", "4.2700"), eur("190/A/NBP/2026", "2026-09-30", "4.2800"),
	}}, Today: func() string { return "2026-10-01" }, From: "2026-09-26", Currencies: []string{"EUR"}}
	if res, err := adapter.Run(ctx, s, x, nbp); err != nil || res.PassErr != nil || res.Booked != 6 {
		t.Fatalf("rates: %+v %v", res, err)
	}

	// A sale on the 29th converts at Monday's average, 4.2650.
	appendAt(t, s, x, selfCo, "sales.invoice.issued", "2026-09-29", "fv-e1", map[string]any{
		"number": "FV/E1/2026", "issue_date": "2026-09-29", "seller_nip": "PL8058083055", "buyer_nip": "PL5265877635",
		"net": "1000.00", "vat": "230.00", "gross": "1230.00", "vat_rate": "23", "currency": "EUR", "intragroup": "no", "market": "pl"})
	// A purchase of the same day, from the same party.
	appendAt(t, s, x, selfCo, "purchase.invoice.received", "2026-09-29", "fz-e1", map[string]any{
		"number": "Z/E1/2026", "invoice_date": "2026-09-29", "supplier_nip": "PL5265877635", "buyer_nip": "PL8058083055",
		"net": "100.00", "vat": "23.00", "gross": "123.00", "vat_rate": "23", "currency": "EUR", "market": "pl"})
	sales, _ := s.ObjectsByType(ctx, "sales_invoice")
	if st := sales[0].State; st["net_pln"] != float64(426500) || st["vat_pln"] != float64(98095) {
		t.Fatalf("the sale in złoty: %v", st)
	}
	if d, c := debits(t, s), credits(t, s); d["201"] != 524595 || c["700"] != 426500 || c["222"] != 98095 ||
		d["221"] != 9810 || c["202"] != 52460 {
		t.Fatalf("booked at 4.2650: debits %v, credits %v", d, c)
	}

	// The customer pays on the 30th (4.2700): a gain of 6.15 zł on what it
	// owed. We pay the supplier on 1 October (4.2800): a loss of 1.84 zł.
	pay := func(dedup, date, direction, amount, title string) {
		t.Helper()
		appendAt(t, s, x, selfCo, "bank.statement.line", date, dedup, map[string]any{
			"date": date, "amount": amount, "currency": "EUR", "direction": direction, "reference": title,
			"counterparty": "Nabywca S.A.", "counterparty_account": "PL60102010260000042270201111",
			"account": "PL83101010230000261395100000", "code": "NTRF", "market": "pl"})
	}
	pay("in-1", "2026-09-30", "credit", "1230.00", "FV/E1/2026")
	pay("out-1", "2026-10-01", "debit", "123.00", "Z/E1/2026")
	d, c := debits(t, s), credits(t, s)
	if c["750"] != 615 || d["751"] != 184 {
		t.Fatalf("realised differences: debits %v, credits %v", d, c)
	}
	if d["201"]-c["201"] != 0 || c["202"]-d["202"] != 0 {
		t.Fatalf("the receivable and the payable close in złoty: 201 %d/%d, 202 %d/%d", d["201"], c["201"], d["202"], c["202"])
	}

	// September in złoty: the settlement, the records, the declaration.
	if tr, err := x.TriggerActivity(at, "close_period", map[string]any{"period": "2026-09-30", "book": "pl-stat", "occurred_at": "2026-10-05"}, "krzysztof", ""); err != nil || tr.Booked != 1 {
		t.Fatalf("close: %+v %v", tr, err)
	}
	st, _ := s.ObjectsByType(ctx, "vat_settlement")
	if len(st) != 1 || st[0].State["output_vat"] != float64(98095) || st[0].State["purchase_vat"] != float64(9810) {
		t.Fatalf("the settlement in złoty: %v", st)
	}
	f, err := adapter.Render(ctx, s, "jpk_v7m", map[string]string{"period": "2026-09", "created": "2026-10-05T10:00:00Z"})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"<K_19>4265.00</K_19>", "<K_20>980.95</K_20>", "<K_42>426.50</K_42>", "<K_43>98.10</K_43>", "<P_19>4265</P_19>", "<P_20>981</P_20>"} {
		if !strings.Contains(string(f.Bytes), want) {
			t.Fatalf("the JPK lacks %s:\n%s", want, f.Bytes)
		}
	}

	// The invoice as sent states its VAT in złoty.
	if res, err := adapter.Run(ctx, s, x, pl.KSeF{Client: pl.FakeKSeF{}, Today: func() string { return "2026-09-29" }}); err != nil || res.Submitted != 1 {
		t.Fatalf("ksef: %+v %v", res, err)
	}
	raws, _ := s.EventsByKind(ctx, core.KindRaw)
	var sent string
	for _, ev := range raws {
		if ev.Type == pl.EventKSeFSubmitted {
			var m map[string]any
			jsonUnmarshal(ev.Payload, &m)
			sent, _ = m["invoice_xml"].(string)
		}
	}
	vat, vatPLN := strings.Index(sent, "<P_14_1>230.00</P_14_1>"), strings.Index(sent, "<P_14_1W>980.95</P_14_1W>")
	if !strings.Contains(sent, "<KodWaluty>EUR</KodWaluty>") || vat < 0 || vatPLN < vat {
		t.Fatalf("the invoice as sent:\n%s", sent)
	}
	if path := os.Getenv("RHEA_FA3_EUR_DUMP"); path != "" {
		os.WriteFile(path, []byte(sent), 0o644)
	}

	// Undoing the customer's match takes the gain back out of 750.
	derived, _ := s.EventsByKind(ctx, core.KindDerived)
	var incoming int64
	for _, ev := range derived {
		if ev.Type == core.EventObjectsMatched && strings.Contains(string(ev.Payload), "pl-incoming-payments") {
			incoming = ev.ID
		}
	}
	if _, err := x.TriggerActivity(at, "unpair_items", map[string]any{"match_event": incoming, "reason": "wrong invoice"}, "shell:krzysztof", ""); err != nil {
		t.Fatal(err)
	}
	if d, c := debits(t, s), credits(t, s); c["750"] != 615 || d["750"] != 615 || d["751"] != 184 {
		t.Fatalf("after the unpairing: debits %v, credits %v", d, c)
	}
}
