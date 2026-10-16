package pl_test

import (
	"context"
	"encoding/xml"
	"os"
	"strings"
	"testing"

	"rhea/internal/adapter"
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
	// The company says who it is through the Poland pack's door: a legal
	// person, its tax office and e-mail.
	companies, _ := s.ObjectsByType(ctx, "company")
	var selfID string
	for _, c := range companies {
		if c.State["kind"] == "self" {
			selfID = c.ID
		}
	}
	if tr, err := x.TriggerActivity(ctx, "identify_company", map[string]any{"company": selfID, "email": "biuro@rhea-test.pl", "tax_office": "1471", "legal_form": "legal_entity", "occurred_at": "2026-09-01"}, "krzysztof", ""); err != nil || tr.Booked != 1 {
		t.Fatalf("identify_company: %+v %v", tr, err)
	}
	// Before the close there is no settlement, and an original JPK_V7M
	// without its declaration is no filing: the writer refuses it.
	if _, err := adapter.Render(ctx, s, "jpk_v7m", map[string]string{"period": "2026-09", "created": "2026-10-05T10:00:00Z"}); err == nil || !strings.Contains(err.Error(), "close the month first") {
		t.Fatalf("before the close: %v", err)
	}
	// The close: the settlement, then the declaration.
	if tr, err := x.TriggerActivity(ctx, "close_period", map[string]any{"period": "2026-09-30", "book": "pl-stat", "occurred_at": "2026-10-05"}, "krzysztof", ""); err != nil || tr.Booked != 1 {
		t.Fatalf("close: %+v %v", tr, err)
	}
	f, err := adapter.Render(ctx, s, "jpk_v7m", map[string]string{"period": "2026-09", "created": "2026-10-05T10:00:00Z"})
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
	// A natural person's return: the owner says who they are through the
	// people pack's door, the company names them as its owner through the
	// Poland pack's door, and the taxpayer block is the person.
	if tr, err := x.TriggerActivity(ctx, "register_person", map[string]any{"name": "Jan Kowalski", "handle": "jan", "role": "owner", "occurred_at": "2026-09-01"}, "krzysztof", ""); err != nil || tr.Booked != 1 {
		t.Fatalf("register_person: %+v %v", tr, err)
	}
	persons, _ := s.ObjectsByType(ctx, "person")
	if tr, err := x.TriggerActivity(ctx, "identify_person", map[string]any{"person": persons[0].ID, "given_name": "Jan", "family_name": "Kowalski", "birth_date": "1977-07-07", "occurred_at": "2026-09-01"}, "krzysztof", ""); err != nil || tr.Booked != 1 {
		t.Fatalf("identify_person: %+v %v", tr, err)
	}
	if tr, err := x.TriggerActivity(ctx, "identify_company", map[string]any{"company": selfID, "email": "jan@kowalski.pl", "tax_office": "1471", "legal_form": "natural_person", "owner": persons[0].ID, "occurred_at": "2026-09-01"}, "krzysztof", ""); err != nil || tr.Booked != 1 {
		t.Fatalf("identify_company: %+v %v", tr, err)
	}
	f, err = adapter.Render(ctx, s, "jpk_v7m", map[string]string{"period": "2026-09", "created": "2026-10-05T10:00:00Z"})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"<OsobaFizyczna>", "<NIP>8058083055</NIP>", "<ImiePierwsze>Jan</ImiePierwsze>", "<Nazwisko>Kowalski</Nazwisko>", "<DataUrodzenia>1977-07-07</DataUrodzenia>", "<Email>jan@kowalski.pl</Email>"} {
		if !strings.Contains(string(f.Bytes), want) {
			t.Fatalf("the natural person's return lacks %s:\n%s", want, f.Bytes)
		}
	}
	if strings.Contains(string(f.Bytes), "OsobaNiefizyczna") {
		t.Fatal("a natural person is not also a legal person")
	}
	if path := os.Getenv("RHEA_JPK_DUMP"); path != "" {
		if err := os.WriteFile(path, f.Bytes, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
