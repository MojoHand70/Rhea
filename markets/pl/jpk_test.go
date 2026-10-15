package pl_test

import (
	"context"
	"encoding/xml"
	"os"
	"strings"
	"testing"

	"rhea/internal/adapter"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store/storetest"
	"rhea/markets/pl"
)

// The JPK_V7M(3) writer renders what the kernel derived: the declaration
// from the month's vat_settlement in whole złoty, the records from the
// month's invoices with their KSeF numbers, the taxpayer from the company —
// and refuses to guess what the company has not said.
func TestJPKV7M(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	installPoland(t, s, x)
	append1(t, s, x, "company.registered", "2026-09-01", "self", map[string]any{"name": "Rhea Test sp. z o.o.", "kind": "self",
		"vat_id": "PL8058083055", "country": "PL", "iban": "PL83101010230000261395100000", "address": "ul. Testowa 1, 00-001 Warszawa"})
	append1(t, s, x, "company.registered", "2026-09-01", "buyer", map[string]any{"name": "Nabywca S.A.", "kind": "customer",
		"vat_id": "PL5265877635", "country": "PL", "iban": "PL60102010260000042270201111", "address": "ul. Próbna 2, 00-002 Warszawa"})
	append1(t, s, x, "sales.invoice.issued", "2026-09-21", "inv-1", map[string]any{
		"number": "FV/1/2026", "issue_date": "2026-09-21", "seller_nip": "PL8058083055", "buyer_nip": "PL5265877635",
		"net": "1000.00", "vat": "230.00", "gross": "1230.00", "vat_rate": "23", "currency": "PLN", "intragroup": "no", "market": "pl"})
	append1(t, s, x, "sales.invoice.issued", "2026-09-22", "inv-2", map[string]any{
		"number": "FV/2/2026", "issue_date": "2026-09-22", "seller_nip": "PL8058083055", "buyer_nip": "PL5265877635",
		"net": "200.00", "vat": "16.00", "gross": "216.00", "vat_rate": "8", "currency": "PLN", "intragroup": "no", "market": "pl"})
	// A purchase from KSeF carries its number; one from a scan does not.
	append1(t, s, x, "purchase.invoice.received", "2026-09-10", "pur-1", map[string]any{
		"number": "Z/77/2026", "invoice_date": "2026-09-10", "supplier_nip": "PL5265877635", "buyer_nip": "PL8058083055",
		"net": "300.00", "vat": "69.00", "gross": "369.00", "vat_rate": "23", "currency": "PLN", "market": "pl",
		"ksef_ref": "5265877635-20260910-0100001AF629-AF"})
	append1(t, s, x, "purchase.invoice.received", "2026-09-12", "pur-2", map[string]any{
		"number": "P/12", "invoice_date": "2026-09-12", "supplier_nip": "PL5265877635", "buyer_nip": "PL8058083055",
		"net": "100.00", "vat": "23.00", "gross": "123.00", "vat_rate": "23", "currency": "PLN", "market": "pl"})
	// FV/1 went to KSeF, FV/2 did not.
	if run, err := adapter.Run(ctx, s, x, pl.KSeF{Client: pl.FakeKSeF{Reject: map[string]string{"FV/2/2026": "odrzucona"}}, Today: func() string { return "2026-09-23" }}); err != nil || run.Submitted != 2 {
		t.Fatalf("ksef: %+v %v", run, err)
	}
	// Without the taxpayer's e-mail and tax office, no file: the writer
	// does not guess.
	if _, err := adapter.Render(ctx, s, "jpk_v7m", map[string]string{"period": "2026-09"}); err == nil || !strings.Contains(err.Error(), "tax office") {
		t.Fatalf("a company that has not said who it is: %v", err)
	}
	if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: "identify-company", Status: core.StatusActive, Priority: 100,
		EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: "the company says who it is to the tax office", Spec: core.RuleSpec{
			Match: core.Match{EventType: "company.identified"},
			Effect: core.Effect{Amend: &core.AmendTemplate{Type: "company", Target: "=ref(company, vat_id, $.vat_id)",
				Set: map[string]string{"email": "=$.email", "tax_office": "=$.tax_office", "legal_form": "=$.legal_form"}}}}}); err != nil {
		t.Fatal(err)
	}
	append1(t, s, x, "company.identified", "2026-09-01", "self-id", map[string]any{"vat_id": "PL8058083055", "email": "biuro@rhea-test.pl", "tax_office": "1471", "legal_form": "legal_entity"})
	// Before the close there is no settlement: records only, no declaration.
	f, err := adapter.Render(ctx, s, "jpk_v7m", map[string]string{"period": "2026-09", "created": "2026-10-05T10:00:00Z"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(f.Bytes), "<Deklaracja>") || !strings.Contains(string(f.Bytes), "<LiczbaWierszySprzedazy>2</LiczbaWierszySprzedazy>") {
		t.Fatalf("before the close:\n%s", f.Bytes)
	}
	// The close: the settlement, then the declaration.
	if tr, err := x.TriggerActivity(ctx, "close_period", map[string]any{"period": "2026-09-30", "book": "pl-stat", "occurred_at": "2026-10-05"}, "krzysztof", ""); err != nil || tr.Booked != 1 {
		t.Fatalf("close: %+v %v", tr, err)
	}
	f, err = adapter.Render(ctx, s, "jpk_v7m", map[string]string{"period": "2026-09", "created": "2026-10-05T10:00:00Z"})
	if err != nil {
		t.Fatal(err)
	}
	text := string(f.Bytes)
	if f.Name != "JPK_V7M_8058083055_202609.xml" || f.MediaType != "application/xml" {
		t.Fatalf("file: %s %s", f.Name, f.MediaType)
	}
	for _, want := range []string{
		`<JPK xmlns="` + pl.JPKNamespace + `"`, `<DataWytworzeniaJPK>2026-10-05T10:00:00Z</DataWytworzeniaJPK>`, `<CelZlozenia poz="P_7">1</CelZlozenia>`,
		`<Podmiot1 rola="Podatnik">`, `<OsobaNiefizyczna>`, `<NIP>8058083055</NIP>`, `<Email>biuro@rhea-test.pl</Email>`,
		`<KodFormularzaDekl kodSystemowy="VAT-7 (23)" kodPodatku="VAT" rodzajZobowiazania="Z" wersjaSchemy="1-0E">VAT-7</KodFormularzaDekl>`, `<WariantFormularzaDekl>23</WariantFormularzaDekl>`,
		`<P_17>200</P_17>`, `<P_18>16</P_18>`, `<P_19>1000</P_19>`, `<P_20>230</P_20>`, `<P_37>1200</P_37>`, `<P_38>246</P_38>`,
		`<P_42>400</P_42>`, `<P_43>92</P_43>`, `<P_48>92</P_48>`, `<P_51>154</P_51>`, `<Pouczenia>1</Pouczenia>`,
		`<NrKSeF>8058083055-20260921-`, `<BFK>1</BFK>`, `<K_17>200.00</K_17>`, `<K_18>16.00</K_18>`, `<PodatekNalezny>246.00</PodatekNalezny>`,
		`<NrKSeF>5265877635-20260910-0100001AF629-AF</NrKSeF>`, `<DowodZakupu>P/12</DowodZakupu>`, `<LiczbaWierszyZakupow>2</LiczbaWierszyZakupow>`, `<PodatekNaliczony>92.00</PodatekNaliczony>`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("the JPK lacks %s:\n%s", want, text)
		}
	}
	// The sales records: FV/1 with its KSeF number, FV/2 marked BFK.
	var doc struct {
		Sales []struct {
			Number string `xml:"DowodSprzedazy"`
			KSeF   string `xml:"NrKSeF"`
			BFK    string `xml:"BFK"`
		} `xml:"Ewidencja>SprzedazWiersz"`
	}
	if err := xml.Unmarshal(f.Bytes, &doc); err != nil || len(doc.Sales) != 2 || doc.Sales[0].KSeF == "" || doc.Sales[0].BFK != "" || doc.Sales[1].BFK != "1" || doc.Sales[1].KSeF != "" {
		t.Fatalf("the sales records: %+v %v", doc.Sales, err)
	}
	// Rendered twice, the same bytes: the file is a function of the objects.
	again, _ := adapter.Render(ctx, s, "jpk_v7m", map[string]string{"period": "2026-09", "created": "2026-10-05T10:00:00Z"})
	if string(again.Bytes) != text {
		t.Fatal("the writer is not deterministic")
	}
	// A month that is not a month, a period that does not exist.
	if _, err := adapter.Render(ctx, s, "jpk_v7m", map[string]string{"period": "2026-13"}); err == nil {
		t.Fatal("2026-13 is not a month")
	}
	if _, err := adapter.Render(ctx, s, "nobody", nil); err == nil || !strings.Contains(err.Error(), "jpk_v7m") {
		t.Fatalf("an unknown writer names the known ones: %v", err)
	}
	// The golden file: this exact document validated against the Ministry's
	// JPK_V7M(3) schema on 2026-10-15 (lxml, the national XSDs); a change
	// here is a change to re-validate, not a free edit.
	golden, err := os.ReadFile("testdata/jpk_v7m_2026-09.xml")
	if err != nil {
		t.Fatal(err)
	}
	if string(golden) != text {
		t.Fatalf("the JPK differs from the schema-validated golden file:\n%s", text)
	}
	if path := os.Getenv("RHEA_JPK_DUMP"); path != "" {
		if err := os.WriteFile(path, f.Bytes, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
