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
	"rhea/internal/store"
	"rhea/internal/store/storetest"
	"rhea/markets/pl"
)

// The dates that decide the VAT month (the Poland pack review, VAT-14,
// VAT-08, KSEF-13, JPK-04, JPK-10): a sale belongs to the month of the sale,
// not of its invoice; a purchase is deductible from the later of the
// supplier's tax point and its receipt, and a KSeF invoice is received the
// day KSeF numbered it. The reader takes both dates from the document, the
// close settles each month with what belongs to it, and the JPK records
// carry the sale and receipt dates where they differ from the document's.
func TestTheDatesThatDecideTheVATMonth(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	installPoland(t, s, x)
	installIntake(t, s, x)
	selfID := registerParties(t, s, x)
	if tr, err := x.TriggerActivity(ctx, "identify_company", map[string]any{"company": selfID, "email": "biuro@rhea-test.pl", "tax_office": "1471", "legal_form": "legal_entity", "occurred_at": "2026-09-01"}, "krzysztof", ""); err != nil || tr.Booked != 1 {
		t.Fatalf("identify_company: %+v %v", tr, err)
	}

	// A September sale invoiced on 3 October, and one invoiced the day it
	// was made.
	append1(t, s, x, "sales.invoice.issued", "2026-10-03", "fv-9", map[string]any{
		"number": "FV/9/2026", "issue_date": "2026-10-03", "sale_date": "2026-09-29", "seller_nip": "PL8058083055", "buyer_nip": "PL5265877635",
		"net": "1000.00", "vat": "230.00", "gross": "1230.00", "vat_rate": "23", "currency": "PLN", "intragroup": "no", "market": "pl"})
	append1(t, s, x, "sales.invoice.issued", "2026-09-25", "fv-10", map[string]any{
		"number": "FV/10/2026", "issue_date": "2026-09-25", "seller_nip": "PL8058083055", "buyer_nip": "PL5265877635",
		"net": "500.00", "vat": "115.00", "gross": "615.00", "vat_rate": "23", "currency": "PLN", "intragroup": "no", "market": "pl"})

	// Two September purchases that arrive in October: one through KSeF,
	// numbered on 2 October, with the sale date stated; one by e-mail on 6
	// October, the file named by nobody's number.
	supplier, buyer := pl.FA3Party{NIP: "5265877635", Name: "Nabywca S.A.", Country: "PL", Address: "ul. Próbna 2, 00-002 Warszawa"},
		pl.FA3Party{NIP: "8058083055", Name: "Rhea Test sp. z o.o.", Country: "PL", Address: "ul. Testowa 1, 00-001 Warszawa"}
	viaKSeF, err := pl.WriteFA3(pl.FA3Invoice{Seller: supplier, Buyer: buyer, Currency: "PLN", IssueDate: "2026-09-30", SaleDate: "2026-09-28",
		Number: "Z/1/2026", RateCode: "23", Net: 10000, VAT: 2300, Gross: 12300, Created: "2026-09-30T08:00:00Z", SystemInfo: "dostawca"})
	if err != nil {
		t.Fatal(err)
	}
	byMail, err := pl.WriteFA3(pl.FA3Invoice{Seller: supplier, Buyer: buyer, Currency: "PLN", IssueDate: "2026-09-30",
		Number: "Z/2/2026", RateCode: "23", Net: 20000, VAT: 4600, Gross: 24600, Created: "2026-09-30T08:00:00Z", SystemInfo: "dostawca"})
	if err != nil {
		t.Fatal(err)
	}
	receive(t, x, "5265877635-20261002-0100001AF629-AF.xml", "application/xml", "ksef", "2026-10-02", viaKSeF)
	receive(t, x, "faktura-z-maila.xml", "application/xml", "email", "2026-10-06", byMail)
	if res, err := adapter.Run(ctx, s, x, pl.KSeFInbound{}); err != nil || res.PassErr != nil || res.Booked != 4 {
		t.Fatalf("the reader: %+v %v", res, err)
	}
	purchases, _ := s.ObjectsByType(ctx, "purchase_invoice")
	byNumber := map[string]core.Object{}
	for _, p := range purchases {
		byNumber[p.State["number"].(string)] = p
	}
	if z := byNumber["Z/1/2026"].State; z["sale_date"] != "2026-09-28" || z["received_date"] != "2026-10-02" || z["vat_period"] != "2026-10" || z["ksef_ref"] == nil {
		t.Fatalf("the KSeF purchase: %v", z)
	}
	if z := byNumber["Z/2/2026"].State; z["sale_date"] != "2026-09-30" || z["received_date"] != "2026-10-06" || z["vat_period"] != "2026-10" {
		t.Fatalf("the mailed purchase: %v", z)
	}
	sales, _ := s.ObjectsByType(ctx, "sales_invoice")
	for _, inv := range sales {
		if inv.State["vat_period"] != "2026-09" {
			t.Fatalf("%v belongs to September: %v", inv.State["number"], inv.State)
		}
	}
	if sales[0].State["sale_date"] != "2026-09-29" || sales[1].State["sale_date"] != "2026-09-25" {
		t.Fatalf("sale dates: %v / %v", sales[0].State["sale_date"], sales[1].State["sale_date"])
	}

	// September settles both sales and no purchase; October the purchases.
	settle := func(period, closedOn string) map[string]any {
		t.Helper()
		if tr, err := x.TriggerActivity(ctx, "close_period", map[string]any{"period": period + "-01", "book": "pl-stat", "occurred_at": closedOn}, "krzysztof", ""); err != nil || tr.Booked != 1 {
			t.Fatalf("close %s: %+v %v", period, tr, err)
		}
		all, _ := s.ObjectsByType(ctx, "vat_settlement")
		for _, st := range all {
			if st.State["period"] == period {
				return st.State
			}
		}
		t.Fatalf("no settlement for %s", period)
		return nil
	}
	if sep := settle("2026-09", "2026-10-07"); sep["output_vat"] != float64(34500) || sep["input_vat"] != float64(0) || sep["vat_due"] != float64(34500) {
		t.Fatalf("September: %v", sep)
	}
	if oct := settle("2026-10", "2026-11-05"); oct["output_vat"] != float64(0) || oct["input_vat"] != float64(6900) || oct["vat_carry"] != float64(6900) {
		t.Fatalf("October: %v", oct)
	}

	render := func(period string) string {
		t.Helper()
		f, err := adapter.Render(ctx, s, "jpk_v7m", map[string]string{"period": period, "created": "2026-11-05T10:00:00Z"})
		if err != nil {
			t.Fatal(err)
		}
		if dir := os.Getenv("RHEA_JPK_DATES_DUMP"); dir != "" {
			os.WriteFile(filepath.Join(dir, f.Name), f.Bytes, 0o644)
		}
		return string(f.Bytes)
	}
	sep := render("2026-09")
	if !strings.Contains(sep, "<DataWystawienia>2026-10-03</DataWystawienia>\n      <DataSprzedazy>2026-09-29</DataSprzedazy>") ||
		strings.Count(sep, "<DataSprzedazy>") != 1 || strings.Contains(sep, "<ZakupWiersz>") {
		t.Fatalf("September's records:\n%s", sep)
	}
	oct := render("2026-10")
	if !strings.Contains(oct, "<DataZakupu>2026-09-30</DataZakupu>\n      <DataWplywu>2026-10-02</DataWplywu>\n      <NrKSeF>5265877635-20261002-0100001AF629-AF</NrKSeF>") ||
		!strings.Contains(oct, "<DataZakupu>2026-09-30</DataZakupu>\n      <DataWplywu>2026-10-06</DataWplywu>\n      <BFK>1</BFK>") ||
		strings.Contains(oct, "<SprzedazWiersz>") {
		t.Fatalf("October's records:\n%s", oct)
	}
}

// An exempt sole trader (VAT-01, VAT-03, KSEF-11, KPIR-08): the company
// says what it is through declare_company_form — exempt under art. 113,
// KPiR, the scale — and the pack follows. Its sale books without a VAT line,
// its purchase VAT is part of the cost, its invoice names the provision it
// is exempt under, its month closes without a VAT settlement and it files
// no JPK. A declaration the law does not allow matches no rule and waits.
func TestAnExemptSoleTrader(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	installPoland(t, s, x)
	selfID := registerParties(t, s, x)

	// The law does not allow these: ryczałt income tax kept in a KPiR, and
	// an exemption without the provision it rests on. Each waits.
	for _, bad := range []map[string]any{
		{"vat_status": "active", "books": "kpir", "income_tax": "ryczalt"},
		{"vat_status": "exempt", "books": "kpir", "income_tax": "skala"},
	} {
		in := map[string]any{"company": selfID, "from": "2026-01-01", "occurred_at": "2026-09-01"}
		for k, v := range bad {
			in[k] = v
		}
		if tr, err := x.TriggerActivity(ctx, "declare_company_form", in, "krzysztof", ""); err != nil || tr.Booked != 0 {
			t.Fatalf("declaration %v: %+v %v", bad, tr, err)
		}
	}
	if wl, _ := s.UnmatchedRawEvents(ctx); len(wl) != 2 {
		t.Fatalf("the two impossible declarations wait: %d", len(wl))
	}
	if tr, err := x.TriggerActivity(ctx, "declare_company_form", map[string]any{"company": selfID, "vat_status": "exempt",
		"vat_exemption_basis": "art. 113 ust. 1 ustawy o VAT", "books": "kpir", "income_tax": "skala", "from": "2026-01-01", "occurred_at": "2026-09-01"}, "krzysztof", ""); err != nil || tr.Booked != 1 {
		t.Fatalf("declare_company_form: %+v %v", tr, err)
	}
	self, _ := s.GetObject(ctx, selfID)
	if self.State["vat_status"] != "exempt" || self.State["vat_exemption_basis"] != "art. 113 ust. 1 ustawy o VAT" ||
		self.State["books"] != "kpir" || self.State["income_tax"] != "skala" || self.State["form_from"] != "2026-01-01" {
		t.Fatalf("the company as declared: %v", self.State)
	}

	append1(t, s, x, "sales.invoice.issued", "2026-09-21", "fv-1", map[string]any{
		"number": "FV/1/2026", "issue_date": "2026-09-21", "seller_nip": "PL8058083055", "buyer_nip": "PL5265877635",
		"net": "1000.00", "vat": "0.00", "gross": "1000.00", "vat_rate": "zw", "currency": "PLN", "intragroup": "no", "market": "pl"})
	append1(t, s, x, "purchase.invoice.received", "2026-09-22", "z-1", map[string]any{
		"number": "Z/1/2026", "invoice_date": "2026-09-22", "supplier_nip": "PL5265877635", "buyer_nip": "PL8058083055",
		"net": "100.00", "vat": "23.00", "gross": "123.00", "vat_rate": "23", "currency": "PLN", "market": "pl"})
	postings, _ := s.ObjectsByType(ctx, exec.PostingObjectType)
	byRule := map[string]int{}
	for _, p := range postings {
		byRule[p.RuleID]++
	}
	if byRule["pl-post-sales-invoice-no-vat"] != 2 || byRule["pl-post-purchase-invoice-gross"] != 2 ||
		byRule["pl-post-sales-invoice"] != 0 || byRule["pl-post-purchase-invoice"] != 0 {
		t.Fatalf("postings by rule: %v", byRule)
	}

	// The invoice names the provision.
	if res, err := adapter.Run(ctx, s, x, pl.KSeF{Client: pl.FakeKSeF{}, Today: func() string { return "2026-09-21" }}); err != nil || res.Submitted != 1 {
		t.Fatalf("ksef: %+v %v", res, err)
	}
	// The acceptance keeps the document exactly as sent.
	raws, _ := s.EventsByKind(ctx, core.KindRaw)
	var sent string
	for _, ev := range raws {
		if ev.Type == pl.EventKSeFSubmitted {
			var m map[string]any
			if err := jsonUnmarshal(ev.Payload, &m); err != nil {
				t.Fatal(err)
			}
			sent, _ = m["invoice_xml"].(string)
		}
	}
	if !strings.Contains(sent, "<P_19A>art. 113 ust. 1 ustawy o VAT</P_19A>") || strings.Contains(sent, "P_19C") {
		t.Fatalf("the exempt invoice as sent:\n%s", sent)
	}

	// The month closes; there is nothing to settle and no return to file.
	if tr, err := x.TriggerActivity(ctx, "close_period", map[string]any{"period": "2026-09-30", "book": "pl-stat", "occurred_at": "2026-10-05"}, "krzysztof", ""); err != nil || tr.Booked != 1 {
		t.Fatalf("close: %+v %v", tr, err)
	}
	if st, _ := s.ObjectsByType(ctx, "vat_settlement"); len(st) != 0 {
		t.Fatalf("an exempt company settles no VAT: %v", st)
	}
	if _, err := adapter.Render(ctx, s, "jpk_v7m", map[string]string{"period": "2026-09"}); err == nil || !strings.Contains(err.Error(), "exempt") {
		t.Fatalf("an exempt company files no JPK_V7: %v", err)
	}
}

// registerParties registers the company of kind self and one counterparty,
// and returns the self company's id.
func registerParties(t *testing.T, s *store.Store, x *exec.Executor) string {
	t.Helper()
	append1(t, s, x, "company.registered", "2026-09-01", "self", map[string]any{"name": "Rhea Test sp. z o.o.", "kind": "self",
		"vat_id": "PL8058083055", "country": "PL", "iban": "PL83101010230000261395100000", "address": "ul. Testowa 1, 00-001 Warszawa"})
	append1(t, s, x, "company.registered", "2026-09-01", "buyer", map[string]any{"name": "Nabywca S.A.", "kind": "customer",
		"vat_id": "PL5265877635", "country": "PL", "iban": "PL60102010260000042270201111", "address": "ul. Próbna 2, 00-002 Warszawa"})
	companies, _ := s.ObjectsByType(context.Background(), "company")
	for _, c := range companies {
		if c.State["kind"] == "self" {
			return c.ID
		}
	}
	t.Fatal("no company of kind self")
	return ""
}
