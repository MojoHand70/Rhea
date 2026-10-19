package pl_test

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"rhea/internal/adapter"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
	"rhea/markets/pl"
)

// installPoland seeds the finance base the Poland pack layers on, installs
// the people pack and the Poland pack, approved, and a rule that registers
// companies — what the story test does by hand.
func installPoland(t *testing.T, s *store.Store, x *exec.Executor) {
	t.Helper()
	ctx := context.Background()
	for _, file := range []string{"finance.json", "finance_v2.json", "finance_v3.json", "finance_v4.json", "finance_v5.json", "finance_v6.json", "finance_v7.json", "finance_v8.json"} {
		b, err := os.ReadFile(filepath.Join("..", "..", "testdata", "seed", file))
		if err != nil {
			t.Fatal(err)
		}
		var defs struct {
			ObjectTypes []core.ObjectType `json:"object_types"`
			ViewDefs    []core.ViewDef    `json:"view_defs"`
		}
		if err := json.Unmarshal(b, &defs); err != nil {
			t.Fatal(err)
		}
		for _, ot := range defs.ObjectTypes {
			if err := s.InsertObjectType(ctx, ot); err != nil {
				t.Fatal(err)
			}
		}
		for _, vd := range defs.ViewDefs {
			if err := s.InsertViewDef(ctx, vd); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, p := range []string{filepath.Join("..", "..", "packs", "common", "people.json"), filepath.Join("..", "..", "packs", "common", "periods.json"), filepath.Join("..", "..", "packs", "pl", "pack.json")} {
		sum, err := pack.Load(ctx, s, p, "test")
		if err != nil {
			t.Fatal(err)
		}
		if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-01-01"); err != nil || len(errs) != 0 {
			t.Fatalf("approve %s: %v %v", sum.Bundle, err, errs)
		}
	}
	if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: "register-company", Status: core.StatusActive, Priority: 100,
		EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: "companies", Spec: core.RuleSpec{
			Match: core.Match{EventType: "company.registered"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "company", Fields: map[string]string{
				"name": "=$.name", "kind": "=$.kind", "vat_id": "=$.vat_id", "country": "=$.country", "iban": "=$.iban", "address": "=$.address"}}}}}); err != nil {
		t.Fatal(err)
	}
}

func append1(t *testing.T, s *store.Store, x *exec.Executor, typ, date, dedup string, payload map[string]any) {
	t.Helper()
	ctx := context.Background()
	raw, _ := json.Marshal(payload)
	if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: typ, OccurredAt: date, Payload: raw, DedupKey: dedup}); err != nil {
		t.Fatal(err)
	}
	if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
		t.Fatalf("%s: %v", typ, errs)
	}
}

// Documents out (DIRECTION 2026-10-15): the adapter writes every unsent
// sales invoice as FA(3) from the objects, sends them in one session, and
// the evidence comes back as events the pack books — an acceptance into a
// ksef_submission with the KSeF number and the UPO, a rejection into a
// ksef_submission with the Ministry's words and a task for today. A second
// pass sends nothing: the accepted invoice is done, the rejected XML is not
// sent twice.
func TestKSeFOutbound(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	installPoland(t, s, x)
	append1(t, s, x, "company.registered", "2026-09-01", "self", map[string]any{"name": "Rhea Test sp. z o.o.", "kind": "self",
		"vat_id": "PL8058083055", "country": "PL", "iban": "PL83101010230000261395100000", "address": "ul. Testowa 1, 00-001 Warszawa"})
	append1(t, s, x, "company.registered", "2026-09-01", "buyer", map[string]any{"name": "Nabywca S.A.", "kind": "customer",
		"vat_id": "PL5265877635", "country": "PL", "iban": "PL60102010260000042270201111", "address": "ul. Próbna 2, 00-002 Warszawa"})
	for i, number := range []string{"FV/1/2026", "FV/2/2026"} {
		append1(t, s, x, "sales.invoice.issued", "2026-09-21", fmt.Sprintf("inv-%d", i), map[string]any{
			"number": number, "issue_date": "2026-09-21", "seller_nip": "PL8058083055", "buyer_nip": "PL5265877635",
			"net": "1000.00", "vat": "230.00", "gross": "1230.00", "vat_rate": "23", "currency": "PLN", "intragroup": "no", "market": "pl"})
	}
	fake := pl.FakeKSeF{Reject: map[string]string{"FV/2/2026": "Błąd weryfikacji semantyki dokumentu faktury"}}
	ksef := pl.KSeF{Client: fake, Today: func() string { return "2026-09-22" }}
	run, err := adapter.Run(ctx, s, x, ksef)
	if err != nil || run.PassErr != nil {
		t.Fatalf("pass: %v / %v", err, run.PassErr)
	}
	if run.Submitted != 2 || len(run.Errors) != 0 || run.Channel != "ksef" || run.Version != 2 {
		t.Fatalf("pass: %+v", run)
	}
	// The evidence: what was sent is in the event, hashed; the UPO is there.
	ev, ok, err := s.LatestEventOfType(ctx, pl.EventKSeFSubmitted)
	if err != nil || !ok {
		t.Fatal(err, ok)
	}
	var sub map[string]any
	json.Unmarshal(ev.Payload, &sub)
	if sub["number"] != "FV/1/2026" || !pl.KSeFNumberValid(sub["ksef_ref"].(string)) || sub["upo"] != "FAKE-SESSION" ||
		!strings.Contains(sub["upo_xml"].(string), "<NumerFaktury>FV/1/2026</NumerFaktury>") || sub["invoice_hash"] == "" || sub["market"] != "pl" ||
		!strings.Contains(sub["invoice_xml"].(string), "<P_2>FV/1/2026</P_2>") || sub["sent_at"] != "2026-09-21T08:00:00+02:00" || sub["accepted_at"] != "2026-09-21T08:00:01+02:00" {
		t.Fatalf("the acceptance: %v", sub)
	}
	// The document sent is kept whole, and its fingerprint is the one the
	// receipt acknowledges.
	sum := sha256.Sum256([]byte(sub["invoice_xml"].(string)))
	if base64.StdEncoding.EncodeToString(sum[:]) != sub["invoice_hash"] {
		t.Fatal("the kept document is not the fingerprinted one")
	}
	// The objects: one accepted, one rejected, and the rejection's task.
	subs, _ := s.ObjectsByType(ctx, "ksef_submission")
	if len(subs) != 2 {
		t.Fatalf("submissions: %+v", subs)
	}
	byStatus := map[string]core.Object{}
	for _, o := range subs {
		byStatus[o.State["status"].(string)] = o
	}
	acc, rej := byStatus["accepted"], byStatus["rejected"]
	if acc.State["ksef_ref"] != sub["ksef_ref"] || acc.State["upo"] != "FAKE-SESSION" || acc.State["invoice_hash"] != sub["invoice_hash"] ||
		acc.State["sent_at"] != "2026-09-21T08:00:00+02:00" || acc.State["accepted_at"] != "2026-09-21T08:00:01+02:00" {
		t.Fatalf("accepted: %+v", acc.State)
	}
	if rej.State["reason"] != "KSeF 450: Błąd weryfikacji semantyki dokumentu faktury" || rej.State["ksef_ref"] != nil || rej.State["date"] != "2026-09-22" {
		t.Fatalf("rejected: %+v", rej.State)
	}
	tasks, _ := s.ObjectsByType(ctx, "task")
	if len(tasks) != 1 || tasks[0].State["status"] != "open" || tasks[0].State["priority"] != "high" || tasks[0].State["category"] != "invoices" ||
		tasks[0].State["due_date"] != "2026-09-22" || tasks[0].State["description"] != rej.State["reason"] {
		t.Fatalf("the rejection's task: %+v", tasks)
	}
	// The second pass: nothing owed — not even to a Ministry that would now accept.
	run, err = adapter.Run(ctx, s, x, pl.KSeF{Client: pl.FakeKSeF{}, Today: ksef.Today})
	if err != nil || run.PassErr != nil || run.Submitted != 0 || run.Duplicates != 0 {
		t.Fatalf("second pass: %+v %v", run, err)
	}
	// A later invoice goes alone: the accepted one is done, the rejected XML
	// is not sent twice. (A ref that resolves to nothing never reaches the
	// adapter — the kernel refuses it at expansion — so the writer's own
	// refusals are the writer test's to show.)
	append1(t, s, x, "sales.invoice.issued", "2026-09-23", "inv-4", map[string]any{
		"number": "FV/4/2026", "issue_date": "2026-09-23", "seller_nip": "PL8058083055", "buyer_nip": "PL5265877635",
		"net": "200.00", "vat": "16.00", "gross": "216.00", "vat_rate": "8", "currency": "PLN", "intragroup": "no", "market": "pl"})
	run, err = adapter.Run(ctx, s, x, pl.KSeF{Client: pl.FakeKSeF{}, Today: ksef.Today})
	if err != nil || run.PassErr != nil || run.Submitted != 1 {
		t.Fatalf("third pass: %+v %v %v", run, err, run.PassErr)
	}
	if subs, _ := s.ObjectsByType(ctx, "ksef_submission"); len(subs) != 3 {
		t.Fatalf("submissions after the third pass: %d", len(subs))
	}
	// A receipt that acknowledges a different document is refused: nothing
	// booked, the pass error names the invoice.
	append1(t, s, x, "sales.invoice.issued", "2026-09-24", "inv-5", map[string]any{
		"number": "FV/5/2026", "issue_date": "2026-09-24", "seller_nip": "PL8058083055", "buyer_nip": "PL5265877635",
		"net": "10.00", "vat": "2.30", "gross": "12.30", "vat_rate": "23", "currency": "PLN", "intragroup": "no", "market": "pl"})
	run, _ = adapter.Run(ctx, s, x, pl.KSeF{Client: lyingKSeF{}, Today: ksef.Today})
	if run.Submitted != 0 || run.PassErr == nil || !strings.Contains(run.PassErr.Error(), "FV/5/2026: the UPO acknowledges a different document") {
		t.Fatalf("a lying receipt: %+v %v", run, run.PassErr)
	}
	if subs, _ := s.ObjectsByType(ctx, "ksef_submission"); len(subs) != 3 {
		t.Fatalf("a lying receipt booked something: %d", len(subs))
	}
}

// The FA(3) the adapter writes from objects says what the objects say,
// in the schema's order, with the mandatory annotation block, the rate in
// its slot, the exemption declared when the rate is zw.
func TestWriteFA3(t *testing.T) {
	body, err := pl.WriteFA3(pl.FA3Invoice{
		Seller:   pl.FA3Party{NIP: "8058083055", Name: "Rhea Test sp. z o.o.", Address: "ul. Testowa 1"},
		Buyer:    pl.FA3Party{NIP: "5265877635", Name: "Nabywca S.A."},
		Currency: "PLN", IssueDate: "2026-09-21", Number: "FV/1/2026", RateCode: "23",
		Net: 100000, VAT: 23000, Gross: 123000, Account: "PL83101010230000261395100000",
		Created: "2026-09-21T08:00:00Z", SystemInfo: "Rhea",
	})
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, want := range []string{
		`<?xml version="1.0" encoding="UTF-8"?>`, `<Faktura xmlns="` + pl.FA3Namespace + `">`,
		`<KodFormularza kodSystemowy="FA (3)" wersjaSchemy="1-0E">FA</KodFormularza>`, `<WariantFormularza>3</WariantFormularza>`,
		`<NIP>8058083055</NIP>`, `<AdresL1>ul. Testowa 1</AdresL1>`, `<AdresL1>Nabywca S.A.</AdresL1>`, `<KodKraju>PL</KodKraju>`, `<JST>2</JST>`, `<GV>2</GV>`,
		`<P_13_1>1000.00</P_13_1>`, `<P_14_1>230.00</P_14_1>`, `<P_15>1230.00</P_15>`,
		`<Adnotacje>`, `<P_16>2</P_16>`, `<P_19N>1</P_19N>`, `<P_22N>1</P_22N>`, `<P_PMarzyN>1</P_PMarzyN>`, `<RodzajFaktury>VAT</RodzajFaktury>`,
		`<P_7>Sprzedaż wg faktury FV/1/2026</P_7>`, `<P_12>23</P_12>`, `<NrRB>PL83101010230000261395100000</NrRB>`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %s in:\n%s", want, text)
		}
	}
	// Order: the schema's.
	order := []string{"<Naglowek>", "<Podmiot1>", "<Podmiot2>", "<JST>", "<GV>", "<Fa>", "<KodWaluty>", "<P_1>", "<P_2>", "<P_13_1>", "<P_15>", "<Adnotacje>", "<RodzajFaktury>", "<FaWiersz>", "<Platnosc>"}
	last := -1
	for _, tag := range order {
		i := strings.Index(text, tag)
		if i < last {
			t.Fatalf("%s out of order", tag)
		}
		last = i
	}
	if strings.Contains(text, "TerminPlatnosci") {
		t.Fatal("no term given, none written")
	}
	var doc struct {
		XMLName xml.Name
	}
	if err := xml.Unmarshal(body, &doc); err != nil || doc.XMLName.Space != pl.FA3Namespace {
		t.Fatalf("well-formed in the namespace: %v %v", err, doc.XMLName)
	}
	// Exempt: the slot without VAT, P_12 zw, the exemption declared by the
	// provision the seller is exempt under (art. 106e ust. 1 pkt 19).
	exempt := pl.FA3Invoice{Seller: pl.FA3Party{NIP: "8058083055", Name: "X"}, Buyer: pl.FA3Party{NIP: "5265877635"},
		Currency: "PLN", IssueDate: "2026-09-21", Number: "FV/5/2026", RateCode: "zw", Net: 5000, Gross: 5000, Created: "2026-09-21T08:00:00Z"}
	if _, err := pl.WriteFA3(exempt); err == nil || !strings.Contains(err.Error(), "P_19A") {
		t.Fatalf("an exempt invoice without its provision: %v", err)
	}
	exempt.ExemptionBasis = "art. 113 ust. 1 ustawy o VAT"
	body, err = pl.WriteFA3(exempt)
	if err != nil || !strings.Contains(string(body), "<P_13_7>50.00</P_13_7>") || strings.Contains(string(body), "P_14_7") ||
		!strings.Contains(string(body), "<P_12>zw</P_12>") || !strings.Contains(string(body), "<P_19>1</P_19>") ||
		!strings.Contains(string(body), "<P_19A>art. 113 ust. 1 ustawy o VAT</P_19A>") || strings.Contains(string(body), "P_19C") || strings.Contains(string(body), "P_19N") {
		t.Fatalf("exempt: %v\n%s", err, body)
	}
	// The sale date: P_6 when the sale was on another day, nothing when the same.
	sale := pl.FA3Invoice{Seller: pl.FA3Party{NIP: "8058083055", Name: "X"}, Buyer: pl.FA3Party{NIP: "5265877635"},
		Currency: "PLN", IssueDate: "2026-10-03", SaleDate: "2026-09-30", Number: "FV/7/2026", RateCode: "23", Net: 10000, VAT: 2300, Gross: 12300, Created: "2026-10-03T08:00:00Z"}
	if body, err = pl.WriteFA3(sale); err != nil || !strings.Contains(string(body), "</P_2>\n    <P_6>2026-09-30</P_6>") {
		t.Fatalf("sale date: %v\n%s", err, body)
	}
	sale.SaleDate = sale.IssueDate
	if body, _ = pl.WriteFA3(sale); strings.Contains(string(body), "P_6") {
		t.Fatalf("a sale on the issue day writes no P_6:\n%s", body)
	}
	// Zero rate: 0 KR in the line, the 6_1 slot, no VAT element.
	body, _ = pl.WriteFA3(pl.FA3Invoice{Seller: pl.FA3Party{NIP: "8058083055", Name: "X"}, Buyer: pl.FA3Party{NIP: "5265877635"},
		Currency: "PLN", IssueDate: "2026-09-21", Number: "FV/6/2026", RateCode: "0", Net: 5000, Gross: 5000, Created: "2026-09-21T08:00:00Z"})
	if !strings.Contains(string(body), "<P_13_6_1>50.00</P_13_6_1>") || !strings.Contains(string(body), "<P_12>0 KR</P_12>") {
		t.Fatalf("zero rate:\n%s", body)
	}
	// Refusals: an unknown rate, a seller without a name.
	if _, err := pl.WriteFA3(pl.FA3Invoice{Seller: pl.FA3Party{NIP: "1", Name: "X"}, Buyer: pl.FA3Party{NIP: "2"}, Number: "n", RateCode: "np"}); err == nil || !strings.Contains(err.Error(), "no FA(3) slot") {
		t.Fatalf("unknown rate: %v", err)
	}
	if _, err := pl.WriteFA3(pl.FA3Invoice{Seller: pl.FA3Party{NIP: "1"}, Buyer: pl.FA3Party{NIP: "2"}, Number: "n", RateCode: "23"}); err == nil || !strings.Contains(err.Error(), "seller needs") {
		t.Fatalf("nameless seller: %v", err)
	}
}

// The live check against the Ministry's test environment (DIRECTION
// 2026-10-15): one invoice, written by Rhea, sent through a real session
// with the token the owner pasted; what came back is printed. Run by hand:
// RHEA_LIVE_KSEF=1 with RHEA_KSEF_ENV=test, RHEA_KSEF_NIP and RHEA_KSEF_TOKEN
// set. The test environment is the validator of record for the FA(3).
func TestKSeFLive(t *testing.T) {
	if os.Getenv("RHEA_LIVE_KSEF") == "" {
		t.Skip("set RHEA_LIVE_KSEF=1 (with RHEA_KSEF_ENV, RHEA_KSEF_NIP, RHEA_KSEF_TOKEN) to send one invoice to the Ministry's test environment")
	}
	env := os.Getenv("RHEA_KSEF_ENV")
	if env == "" {
		env = "test"
	}
	if env == "prod" {
		t.Fatal("the live check never runs against production")
	}
	nip, token := os.Getenv("RHEA_KSEF_NIP"), os.Getenv("RHEA_KSEF_TOKEN")
	sendLive(t, &pl.KSeF2{BaseURL: pl.KSeFEnvironments[env], NIP: nip, Token: token,
		Log: func(format string, args ...any) { t.Logf(format, args...) }})
}

// TestKSeFCertificateLive logs in to the Ministry's TEST environment with a
// certificate instead of the token — a company seal Rhea makes for the NIP,
// which only TEST accepts — and sends one invoice through the same session.
// Run by hand: RHEA_LIVE_KSEF_CERT=1, with RHEA_KSEF_NIP (default 8058083055,
// the made-up NIP of the token check).
func TestKSeFCertificateLive(t *testing.T) {
	if os.Getenv("RHEA_LIVE_KSEF_CERT") == "" {
		t.Skip("set RHEA_LIVE_KSEF_CERT=1 to log in to the Ministry's TEST environment with a self-made certificate")
	}
	nip := os.Getenv("RHEA_KSEF_NIP")
	if nip == "" {
		nip = "8058083055"
	}
	seal, err := pl.SelfSignedSeal(nip, "Rhea Test sp. z o.o.", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	sendLive(t, &pl.KSeF2{BaseURL: pl.KSeFEnvironments["test"], NIP: nip, Signer: seal,
		Log: func(format string, args ...any) { t.Logf(format, args...) }})
}

// TestKSeFCertificateEnrollLive rehearses on TEST what Novi will do once:
// log in with a signature (here a self-made seal), ask KSeF for its own login
// certificate and an Offline one, then log in with the issued login
// certificate and send an invoice. Each run creates certificates in KSeF's
// records, so it is gated apart: RHEA_LIVE_KSEF_ENROLL=1.
func TestKSeFCertificateEnrollLive(t *testing.T) {
	if os.Getenv("RHEA_LIVE_KSEF_ENROLL") == "" {
		t.Skip("set RHEA_LIVE_KSEF_ENROLL=1 to ask the Ministry's TEST environment for KSeF certificates")
	}
	nip := os.Getenv("RHEA_KSEF_NIP")
	if nip == "" {
		nip = "8058083055"
	}
	logf := func(format string, args ...any) { t.Logf(format, args...) }
	seal, err := pl.SelfSignedSeal(nip, "Rhea Test sp. z o.o.", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	first := &pl.KSeF2{BaseURL: pl.KSeFEnvironments["test"], NIP: nip, Signer: seal, Log: logf}
	access, err := first.Login(ctx)
	if err != nil {
		t.Fatal(err)
	}
	login, err := first.EnrollCertificate(ctx, access, pl.CertAuthentication, "Rhea logowanie test")
	if err != nil {
		t.Fatal(err)
	}
	offline, err := first.EnrollCertificate(ctx, access, pl.CertOffline, "Rhea offline test")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("login certificate: subject %s, issuer %s, key usage %v", login.Cert.Subject, login.Cert.Issuer, login.Cert.KeyUsage)
	t.Logf("offline certificate: subject %s, key usage %v", offline.Cert.Subject, offline.Cert.KeyUsage)
	if offline.Cert.KeyUsage&x509.KeyUsageContentCommitment == 0 {
		t.Errorf("the Offline certificate does not carry non-repudiation: %v", offline.Cert.KeyUsage)
	}
	// The issued certificate is the way in now.
	sendLive(t, &pl.KSeF2{BaseURL: pl.KSeFEnvironments["test"], NIP: nip, Signer: login, Log: logf})
}

// sendLive sends one invoice through client and reads the receipt back.
func sendLive(t *testing.T, client *pl.KSeF2) {
	t.Helper()
	nip := client.NIP
	number := "RHEA/TEST/" + time.Now().UTC().Format("20060102-150405")
	today := time.Now().UTC().Format("2006-01-02")
	body, err := pl.WriteFA3(pl.FA3Invoice{
		Seller:   pl.FA3Party{NIP: nip, Name: "Rhea Test sp. z o.o.", Address: "ul. Testowa 1, 00-001 Warszawa"},
		Buyer:    pl.FA3Party{NIP: "5265877635", Name: "Nabywca Testowy S.A.", Address: "ul. Próbna 2, 00-002 Warszawa"},
		Currency: "PLN", IssueDate: today, Number: number, RateCode: "23",
		Net: 100000, VAT: 23000, Gross: 123000, Created: time.Now().UTC().Format(time.RFC3339), SystemInfo: "Rhea",
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	subs, err := client.Submit(ctx, []pl.Outbound{{ID: "live", Number: number, XML: body}})
	if err != nil {
		t.Fatalf("the Ministry's test environment: %v", err)
	}
	for _, sub := range subs {
		t.Logf("invoice %s: accepted=%v code=%d %s details=%v ksef=%s session=%s", sub.Number, sub.Accepted, sub.Code, sub.Description, sub.Details, sub.KSeFNumber, sub.Session)
		if sub.Accepted {
			t.Logf("KSeF number valid: %v\nUPO:\n%s", pl.KSeFNumberValid(sub.KSeFNumber), sub.UPO)
			upo, err := pl.ParseUPO(sub.UPO)
			if err != nil || upo.DocumentHash != sub.InvoiceHash || upo.KSeFNumber != sub.KSeFNumber || upo.ReceivedAt == "" || upo.NumberedAt == "" {
				t.Errorf("the receipt read back: %+v %v (our fingerprint %s)", upo, err, sub.InvoiceHash)
			} else {
				t.Logf("the receipt acknowledges our document: fingerprint %s, received %s, numbered %s", upo.DocumentHash, upo.ReceivedAt, upo.NumberedAt)
			}
		} else {
			t.Errorf("rejected: %s", sub.Reason())
		}
	}
}

// lyingKSeF accepts everything with a receipt whose fingerprint is not of
// the document sent.
type lyingKSeF struct{}

func (lyingKSeF) Submit(ctx context.Context, invoices []pl.Outbound) ([]pl.Submission, error) {
	subs, err := pl.FakeKSeF{}.Submit(ctx, invoices)
	for i := range subs {
		subs[i].UPO = strings.Replace(subs[i].UPO, "<SkrotDokumentu>"+subs[i].InvoiceHash, "<SkrotDokumentu>AAAA", 1)
	}
	return subs, err
}

// TestDumpFA3 writes the writer's sample document to RHEA_FA3_DUMP, for a
// schema validator outside the test suite (none is installed here; the
// scratch validator uses lxml against the Ministry's XSD).
func TestDumpFA3(t *testing.T) {
	path := os.Getenv("RHEA_FA3_DUMP")
	if path == "" {
		t.Skip("set RHEA_FA3_DUMP=<file> to write the sample FA(3)")
	}
	body, err := pl.WriteFA3(pl.FA3Invoice{
		Seller:   pl.FA3Party{NIP: "8058083055", Name: "Rhea Test sp. z o.o.", Address: "ul. Testowa 1, 00-001 Warszawa"},
		Buyer:    pl.FA3Party{NIP: "5265877635", Name: "Nabywca S.A.", Address: "ul. Próbna 2, 00-002 Warszawa"},
		Currency: "PLN", IssueDate: "2026-09-21", Number: "FV/1/2026", RateCode: "23",
		Net: 100000, VAT: 23000, Gross: 123000, Account: "PL83101010230000261395100000",
		Created: "2026-09-21T08:00:00Z", SystemInfo: "Rhea",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
}

// The FA(3) golden file validated against the Ministry's schema on
// 2026-10-15 (lxml, the national XSDs) and was accepted by the test
// environment the same day; the writer must keep producing it.
func TestFA3Golden(t *testing.T) {
	body, err := pl.WriteFA3(pl.FA3Invoice{
		Seller:   pl.FA3Party{NIP: "8058083055", Name: "Rhea Test sp. z o.o.", Address: "ul. Testowa 1, 00-001 Warszawa"},
		Buyer:    pl.FA3Party{NIP: "5265877635", Name: "Nabywca S.A.", Address: "ul. Próbna 2, 00-002 Warszawa"},
		Currency: "PLN", IssueDate: "2026-09-21", Number: "FV/1/2026", RateCode: "23",
		Net: 100000, VAT: 23000, Gross: 123000, Account: "PL83101010230000261395100000",
		Created: "2026-09-21T08:00:00Z", SystemInfo: "Rhea",
	})
	if err != nil {
		t.Fatal(err)
	}
	golden, err := os.ReadFile("testdata/fa3_sample.xml")
	if err != nil {
		t.Fatal(err)
	}
	if string(golden) != string(body) {
		t.Fatalf("the FA(3) differs from the schema-validated golden file:\n%s", body)
	}
}
