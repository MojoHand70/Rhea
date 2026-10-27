package pl_test

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"rhea/internal/adapter"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/sim"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
	"rhea/markets/pl"
)

// installs the documents pack and the KSeF inbound channel, approved.
func installIntake(t *testing.T, s *store.Store, x *exec.Executor) {
	t.Helper()
	ctx := context.Background()
	for _, p := range []string{filepath.Join("..", "..", "packs", "common", "documents.json"),
		filepath.Join("..", "..", "packs", "pl", "ksef-inbound.json")} {
		sum, err := pack.Load(ctx, s, p, "test")
		if err != nil {
			t.Fatal(err)
		}
		if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-01-01"); err != nil || len(errs) != 0 {
			t.Fatalf("approve %s: %v %v", sum.Bundle, err, errs)
		}
	}
}

func receive(t *testing.T, x *exec.Executor, name, media, source, date string, data []byte) {
	t.Helper()
	tr, err := x.TriggerActivity(context.Background(), "receive_document", map[string]any{
		"file": map[string]any{"name": name, "media_type": media, "bytes": data}, "source": source, "occurred_at": date,
	}, "test", "")
	if err != nil || tr.Booked != 1 || len(tr.Errors) != 0 {
		t.Fatalf("receive %s: %+v %v", name, tr, err)
	}
}

// The reader (DIRECTION 2026-10-10, step 4): a KSeF invoice in the store
// becomes a purchase.invoice.received saying what the generator's fact
// says, naming the file; the document moves to read; a second pass finds
// nothing; a document that contradicts itself is reported, not read; an
// XML that is not an invoice is left alone.
func TestKSeFInboundReadsTheStore(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	installIntake(t, s, x)

	p, _, err := sim.Load(filepath.Join("..", "..", "testdata", "customers", "nordwind.json"))
	if err != nil {
		t.Fatal(err)
	}
	p.Months = 1
	f, err := sim.Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	var ksef []sim.Document
	for _, d := range f.Documents {
		if d.Source == "ksef" {
			ksef = append(ksef, d)
			receive(t, x, filepath.Base(d.Path), d.MediaType, d.Source, d.Date, d.Bytes)
		}
	}
	receive(t, x, "notes.xml", "application/xml", "email", "2026-01-20", []byte(`<?xml version="1.0"?><Notes><Note>not an invoice</Note></Notes>`))

	res, err := adapter.Run(ctx, s, x, pl.KSeFInbound{})
	if err != nil || res.PassErr != nil {
		t.Fatalf("pass: %v %v", err, res.PassErr)
	}
	if res.Channel != "ksef_inbound" || res.Submitted != 2*len(ksef) || res.Booked != len(ksef) {
		t.Fatalf("pass = %+v (ksef files %d)", res, len(ksef))
	}
	// Each invoice fact says what the generator said, and names the file.
	raws, _ := s.EventsByKind(ctx, core.KindRaw)
	byNumber := map[string]core.Event{}
	for _, ev := range raws {
		if ev.Type == "purchase.invoice.received" {
			var m map[string]any
			if err := jsonUnmarshal(ev.Payload, &m); err != nil {
				t.Fatal(err)
			}
			byNumber[m["number"].(string)] = ev
		}
	}
	for _, d := range ksef {
		fact := f.Events[d.Events[0]]
		ev, ok := byNumber[fact.Payload["number"].(string)]
		if !ok {
			t.Fatalf("invoice %v not read", fact.Payload["number"])
		}
		var m map[string]any
		jsonUnmarshal(ev.Payload, &m)
		hash := core.HashBytes(d.Bytes)
		if ev.ActivityName != "ksef_inbound" || ev.Actor != "adapter:ksef_inbound" || ev.OccurredAt != fact.OccurredAt ||
			m["document"] != hash || m["net"] != fact.Payload["net"] || m["supplier_nip"] != fact.Payload["supplier_nip"] ||
			m["supplier"] != fact.Payload["supplier"] || m["invoice_date"] != fact.Payload["invoice_date"] ||
			m["vat_rate"] != fact.Payload["vat_rate"] || m["currency"] != "PLN" || m["market"] != "pl" ||
			m["buyer_nip"] != fact.Payload["buyer_nip"] ||
			m["ksef_ref"] != strings.TrimSuffix(filepath.Base(d.Path), ".xml") {
			t.Fatalf("read %v, generated %v (event %+v)", m, fact.Payload, ev)
		}
		if wz, ok := fact.Payload["for_document"]; ok && m["for_document"] != wz {
			t.Fatalf("for_document = %v, want %v", m["for_document"], wz)
		}
		ids, _ := s.FindObjectIDsByField(ctx, "received_document", "hash", hash)
		if o, err := s.GetObject(ctx, ids[0]); err != nil || o.State["status"] != "read" {
			t.Fatalf("document %s status = %v %v", hash[:16], o.State["status"], err)
		}
	}
	// The invoice facts wait for rules; nothing else does.
	wl, _ := s.UnmatchedRawEvents(ctx)
	if len(wl) != len(ksef) {
		t.Fatalf("worklist = %d, want the %d invoice facts", len(wl), len(ksef))
	}
	// The notes are not ours: still received, untouched.
	ids, _ := s.FindObjectIDsByField(ctx, "received_document", "name", "notes.xml")
	if o, _ := s.GetObject(ctx, ids[0]); o.State["status"] != "received" {
		t.Fatalf("notes.xml status = %v", o.State["status"])
	}

	// A second pass: every invoice is read, nothing is owed.
	res, err = adapter.Run(ctx, s, x, pl.KSeFInbound{})
	if err != nil || res.Submitted != 0 || res.Duplicates != 0 {
		t.Fatalf("second pass = %+v, %v", res, err)
	}

	// A document that contradicts itself is reported for a person and
	// stays received.
	lying := bytes.Replace(ksef[0].Bytes, []byte("<P_15>"), []byte("<P_15>9"), 1)
	receive(t, x, "lying.xml", "application/xml", "email", "2026-01-21", lying)
	res, err = adapter.Run(ctx, s, x, pl.KSeFInbound{Today: func() string { return "2026-01-22" }})
	if err != nil || res.Submitted != 1 {
		t.Fatalf("lying pass = %+v, %v", res, err)
	}
	ev, ok, _ := s.LatestEventOfType(ctx, adapter.EventDocumentUnreadable)
	if !ok || !bytes.Contains(ev.Payload, []byte("contradicts itself")) || !bytes.Contains(ev.Payload, []byte(core.HashBytes(lying))) {
		t.Fatalf("unreadable = %+v", ev)
	}
	if wl, _ := s.UnmatchedRawEvents(ctx); len(wl) != len(ksef)+1 {
		t.Fatalf("the refusal is not in the worklist: %d", len(wl))
	}
	ids, _ = s.FindObjectIDsByField(ctx, "received_document", "hash", core.HashBytes(lying))
	if o, _ := s.GetObject(ctx, ids[0]); o.State["status"] != "received" {
		t.Fatalf("lying.xml status = %v", o.State["status"])
	}

	// The buyer's NIP carries the buyer's own country, and is absent when
	// the document names none — never a bare "PL".
	// (other invoice numbers: the same seller's same number is the same
	// invoice whatever its bytes, and would deduplicate)
	numbered := func(b []byte, suffix string) []byte {
		m := invoiceNumber.FindSubmatch(b)
		return bytes.Replace(b, m[0], []byte("<P_2>"+string(m[1])+suffix+"</P_2>"), 1)
	}
	deBuyer := numbered(inBuyer(ksef[0].Bytes, "<KodKraju>PL</KodKraju>", "<KodKraju>DE</KodKraju>"), "-DE")
	noBuyer := numbered(inBuyer(ksef[0].Bytes, "<NIP>5260301234</NIP>", ""), "-NB")
	receive(t, x, "de-buyer.xml", "application/xml", "ksef", "2026-01-23", deBuyer)
	receive(t, x, "no-buyer.xml", "application/xml", "ksef", "2026-01-23", noBuyer)
	if res, err := adapter.Run(ctx, s, x, pl.KSeFInbound{}); err != nil || res.Submitted != 4 {
		t.Fatalf("buyer pass = %+v, %v", res, err)
	}
	// The same invoice in other bytes (CRLF, a manual upload beside the
	// pull) is the same invoice: its reading is new, its fact a duplicate —
	// never a second purchase invoice and input VAT twice.
	invoicesBefore, _ := s.EventsOfTypes(ctx, []string{"purchase.invoice.received"})
	receive(t, x, "again.xml", "application/xml", "email", "2026-01-24", bytes.ReplaceAll(ksef[0].Bytes, []byte("\n"), []byte("\r\n")))
	if res, err := adapter.Run(ctx, s, x, pl.KSeFInbound{}); err != nil || res.Submitted != 1 || res.Duplicates < 1 {
		t.Fatalf("the same invoice again = %+v, %v", res, err)
	}
	if invoicesAfter, _ := s.EventsOfTypes(ctx, []string{"purchase.invoice.received"}); len(invoicesAfter) != len(invoicesBefore) {
		t.Fatalf("the same invoice in other bytes booked again: %d → %d", len(invoicesBefore), len(invoicesAfter))
	}
	raws, _ = s.EventsByKind(ctx, core.KindRaw)
	factOf := func(hash string) map[string]any {
		for _, ev := range raws {
			var m map[string]any
			if ev.Type == "purchase.invoice.received" && jsonUnmarshal(ev.Payload, &m) == nil && m["document"] == hash {
				return m
			}
		}
		t.Fatalf("no invoice fact names %s", hash)
		return nil
	}
	if m := factOf(core.HashBytes(deBuyer)); m["buyer_nip"] != "DE5260301234" || m["market"] != "pl" {
		t.Fatalf("German buyer = %v", m)
	}
	if m := factOf(core.HashBytes(noBuyer)); m["buyer_nip"] != nil {
		t.Fatalf("no buyer NIP, yet buyer_nip = %v", m["buyer_nip"])
	}
}

// inBuyer edits the buyer party (Podmiot2) of a Faktura, leaving the seller alone.
func inBuyer(b []byte, from, to string) []byte {
	i := bytes.Index(b, []byte("<Podmiot2>"))
	return append(append([]byte{}, b[:i]...), bytes.Replace(b[i:], []byte(from), []byte(to), 1)...)
}

// ParseFA on the generator's file agrees with the generator.
func TestParseFA(t *testing.T) {
	p, _, _ := sim.Load(filepath.Join("..", "..", "testdata", "customers", "nordwind.json"))
	p.Months = 1
	f, _ := sim.Generate(p)
	for _, d := range f.Documents {
		if d.Source != "ksef" {
			continue
		}
		if !pl.IsFA(d.Bytes) {
			t.Fatalf("%s is not recognized", d.Path)
		}
		inv, err := pl.ParseFA(d.Bytes)
		if err != nil {
			t.Fatalf("%s: %v", d.Path, err)
		}
		fact := f.Events[d.Events[0]]
		net, _ := core.ParseMoney(fact.Payload["net"].(string))
		if inv.Net != net || inv.Net+inv.VAT != inv.Gross || inv.RateCode != "23" || inv.Number != fact.Payload["number"] ||
			"PL"+inv.SellerNIP != fact.Payload["supplier_nip"] || inv.BuyerNIP != "5260301234" || inv.BuyerCountry != "PL" || len(inv.Lines) == 0 {
			t.Fatalf("%s: %+v vs %v", d.Path, inv, fact.Payload)
		}
	}
	if pl.IsFA([]byte(`<Other/>`)) || pl.IsFA([]byte(`not xml`)) {
		t.Fatal("IsFA sees invoices everywhere")
	}
	if _, err := pl.ParseFA([]byte(`<Faktura><Fa><P_2>X</P_2></Fa></Faktura>`)); err == nil {
		t.Fatal("an empty invoice parsed")
	}
	// A zero-rated sale sits in P_13_6_1 (FA(3) splits the zero rate three
	// ways: domestic, intra-EU, export) and has no VAT element at all.
	zero := `<Faktura xmlns="http://crd.gov.pl/wzor/2025/06/25/13775/"><Podmiot1><DaneIdentyfikacyjne><NIP>5260301234</NIP><Nazwa>A</Nazwa></DaneIdentyfikacyjne></Podmiot1>` +
		`<Podmiot2><DaneIdentyfikacyjne><NIP>5272001122</NIP><Nazwa>B</Nazwa></DaneIdentyfikacyjne></Podmiot2>` +
		`<Fa><KodWaluty>PLN</KodWaluty><P_1>2026-09-02</P_1><P_2>FS 0/2026</P_2><P_13_6_1>100.00</P_13_6_1><P_15>100.00</P_15>` +
		`<FaWiersz><NrWierszaFa>1</NrWierszaFa><P_7>x</P_7><P_8B>1</P_8B><P_9A>100.00</P_9A><P_11>100.00</P_11><P_12>0</P_12></FaWiersz></Fa></Faktura>`
	inv, err := pl.ParseFA([]byte(zero))
	if err != nil || inv.RateCode != "0" || inv.Net != 10000 || inv.VAT != 0 || inv.Gross != 10000 {
		t.Fatalf("zero-rated FA(3): %+v %v", inv, err)
	}
	// The writer spells a zero-rated line "0 KR" (P_12), as KSeF wants; the
	// reader maps it back to the pack's "0", so the round trip holds and
	// ref(vat_rate, code, $.vat_rate) finds the rate.
	inv, err = pl.ParseFA([]byte(strings.Replace(zero, "<P_12>0</P_12>", "<P_12>0 KR</P_12>", 1)))
	if err != nil || inv.RateCode != "0" {
		t.Fatalf("a 0 KR line: %+v %v", inv, err)
	}
	// Slot 4 is the 4% flat rate, with VAT of its own: never 0% with VAT.
	flat := strings.Replace(strings.Replace(zero, "<P_13_6_1>100.00</P_13_6_1><P_15>100.00</P_15>",
		"<P_13_4>100.00</P_13_4><P_14_4>4.00</P_14_4><P_15>104.00</P_15>", 1), "<P_12>0</P_12>", "<P_12>4</P_12>", 1)
	inv, err = pl.ParseFA([]byte(flat))
	if err != nil || inv.RateCode != "4" || inv.VAT != 400 || inv.Gross != 10400 {
		t.Fatalf("the 4%% slot: %+v %v", inv, err)
	}
}

func jsonUnmarshal(b []byte, v any) error { return json.Unmarshal(b, v) }

var invoiceNumber = regexp.MustCompile(`<P_2>([^<]*)</P_2>`)

// One purchase invoice is one fact, whichever door it came through (review
// 2026-10-21, D2): the KSeF reader keyed an invoice "ksef_inbound/…" and the
// PDF reader "pdf_invoice/…", so a supplier's invoice pulled from KSeF and
// e-mailed as a PDF booked twice — the cost, the input VAT and the payable
// twice, and two rows in the JPK. Now both key it on the seller's tax id and
// the number, written one way; an invoice logged under either old key is
// known by those two fields, so it is not logged again either.
func TestOneInvoiceThroughTwoDoorsIsOneFact(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	installIntake(t, s, x)
	sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", "pl", "pdf-invoice.json"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-01-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve: %v %v", err, errs)
	}
	p, _, err := sim.Load(filepath.Join("..", "..", "testdata", "customers", "nordwind.json"))
	if err != nil {
		t.Fatal(err)
	}
	p.Months = 1
	f, err := sim.Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	var ksef []sim.Document
	for _, d := range f.Documents {
		if d.Source == "ksef" {
			ksef = append(ksef, d)
		}
	}
	if len(ksef) < 2 {
		t.Fatalf("the test needs two KSeF invoices, the month has %d", len(ksef))
	}
	// The same invoice as a PDF, its seller's NIP written as people write it.
	pdfOf := func(d sim.Document) []byte {
		fact := f.Events[d.Events[0]].Payload
		nip := strings.TrimPrefix(fact["supplier_nip"].(string), "PL")
		return sim.WritePDF([]string{"RECHNUNG",
			"Rechnungsnummer: " + fact["number"].(string),
			"Rechnungsdatum: " + fact["invoice_date"].(string),
			"Verkaeufer: " + fact["supplier"].(string),
			"USt-IdNr.: " + nip[:3] + "-" + nip[3:6] + "-" + nip[6:8] + "-" + nip[8:],
			"Kaeufer NIP: " + fact["buyer_nip"].(string),
			"Nettobetrag: " + fact["net"].(string) + " PLN",
			"MwSt. 23%: " + fact["vat"].(string) + " PLN",
			"Bruttobetrag: " + fact["gross"].(string) + " PLN"})
	}
	invoices := func() int {
		t.Helper()
		evs, err := s.EventsOfTypes(ctx, []string{"purchase.invoice.received"})
		if err != nil {
			t.Fatal(err)
		}
		return len(evs)
	}
	pdf, _ := adapter.Lookup("pdf_invoice")
	adapter.ModelReader = nil // the text layer, read exactly

	// From KSeF, then the e-mailed PDF of the same invoice.
	first := ksef[0]
	receive(t, x, filepath.Base(first.Path), first.MediaType, first.Source, first.Date, first.Bytes)
	if res, err := adapter.Run(ctx, s, x, pl.KSeFInbound{}); err != nil || res.PassErr != nil {
		t.Fatalf("ksef pass: %+v %v", res, err)
	}
	receive(t, x, "scan-1.pdf", "application/pdf", "email", first.Date, pdfOf(first))
	if res, err := adapter.Run(ctx, s, x, pdf); err != nil || res.PassErr != nil {
		t.Fatalf("pdf pass: %+v %v", res, err)
	}
	if n := invoices(); n != 1 {
		t.Fatalf("the invoice from KSeF and its PDF: %d facts", n)
	}

	// An invoice a PDF reading logged under the old key, then pulled from KSeF.
	second := ksef[1]
	fact := f.Events[second.Events[0]].Payload
	old, _ := json.Marshal(map[string]any{"number": fact["number"], "supplier": fact["supplier"], "supplier_nip": fact["supplier_nip"],
		"invoice_date": fact["invoice_date"], "net": fact["net"], "vat": fact["vat"], "gross": fact["gross"], "currency": "PLN"})
	if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: "purchase.invoice.received", OccurredAt: second.Date, Payload: old,
		DedupKey: "pdf_invoice/" + fact["supplier_nip"].(string) + "/" + fact["number"].(string), Actor: "adapter:pdf_invoice",
		ActivityName: "pdf_invoice", ActivityVersion: 1}); err != nil {
		t.Fatal(err)
	}
	receive(t, x, filepath.Base(second.Path), second.MediaType, second.Source, second.Date, second.Bytes)
	if res, err := adapter.Run(ctx, s, x, pl.KSeFInbound{}); err != nil || res.PassErr != nil {
		t.Fatalf("ksef pass: %+v %v", res, err)
	}
	if n := invoices(); n != 2 {
		t.Fatalf("two invoices, each once: %d facts", n)
	}
}

// An FA(3) amount has up to two decimals, and a line may state no net
// (review 2026-10-21, E5): the schema's TKwotowy takes "1000", "1000.5" and
// "1000.50" alike, and P_11 on an invoice line is optional — the reader
// asked every amount for exactly two decimals and every line for its net,
// so a valid invoice from KSeF came back unreadable for a person to retype.
// Now the reader takes what the schema takes, and a line without its net
// leaves the totals to say what the invoice is.
func TestFAAmountsAsTheSchemaWritesThem(t *testing.T) {
	p, _, _ := sim.Load(filepath.Join("..", "..", "testdata", "customers", "nordwind.json"))
	p.Months = 1
	f, _ := sim.Generate(p)
	var doc []byte
	for _, d := range f.Documents {
		if d.Source == "ksef" {
			doc = d.Bytes
			break
		}
	}
	want, err := pl.ParseFA(doc)
	if err != nil {
		t.Fatal(err)
	}
	// Every amount ending in a zero written with one decimal, as a valid
	// file may write it, and the lines' P_11 dropped.
	short := regexp.MustCompile(`>([0-9]+\.[0-9])0<`).ReplaceAll(doc, []byte(">$1<"))
	short = regexp.MustCompile(`<P_11>[^<]*</P_11>`).ReplaceAll(short, nil)
	if bytes.Equal(short, doc) || bytes.Contains(short, []byte("<P_11>")) {
		t.Fatal("the test changed nothing")
	}
	got, err := pl.ParseFA(short)
	if err != nil {
		t.Fatalf("a valid FA(3) refused: %v", err)
	}
	if got.Net != want.Net || got.VAT != want.VAT || got.Gross != want.Gross || got.RateCode != want.RateCode || len(got.Lines) != len(want.Lines) {
		t.Fatalf("read %+v, want %+v", got, want)
	}
}

// A purchase invoice is one the company bought (review 2026-10-27, high):
// the reader never asked whose invoice it read, so an office uploading a
// client's whole KSeF export — the client's own sales invoices in it — made
// each a purchase.invoice.received, and the pack deducted its VAT as input
// tax: the return was wrong. Now, where the company's own record says its
// NIP, an invoice the company sold, or one bought by someone else, is
// reported unreadable for a person, with the reason, and books nothing.
func TestTheReaderBooksOnlyWhatTheCompanyBought(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	installIntake(t, s, x)
	if err := s.InsertObjectType(ctx, core.ObjectType{Name: "company", Version: 1, Domain: "parties", LabelField: "name",
		Fields: []core.FieldDef{{Name: "name", Type: "string", Required: true}, {Name: "kind", Type: "string"}, {Name: "vat_id", Type: "string"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: "register-company", Status: core.StatusActive, Priority: 100,
		EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: "companies", Spec: core.RuleSpec{
			Match: core.Match{EventType: "company.registered"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "company", Fields: map[string]string{
				"name": "=$.name", "kind": "=$.kind", "vat_id": "=$.vat_id"}}}}}); err != nil {
		t.Fatal(err)
	}
	self, _ := json.Marshal(map[string]any{"name": "Helios sp. z o.o.", "kind": "self", "vat_id": "PL8058083055"})
	if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: "company.registered", OccurredAt: "2026-09-01", Payload: self, DedupKey: "self"}); err != nil {
		t.Fatal(err)
	}
	if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
		t.Fatal(errs)
	}
	invoice := func(number, seller, buyer string) []byte {
		body, err := pl.WriteFA3(pl.FA3Invoice{
			Seller:   pl.FA3Party{NIP: seller, Name: "Sprzedawca " + seller, Address: "ul. Testowa 1, 00-001 Warszawa"},
			Buyer:    pl.FA3Party{NIP: buyer, Name: "Nabywca " + buyer},
			Currency: "PLN", IssueDate: "2026-09-21", Number: number, RateCode: "23",
			Net: 100000, VAT: 23000, Gross: 123000, Created: "2026-09-21T08:00:00Z", SystemInfo: "test",
		})
		if err != nil {
			t.Fatal(err)
		}
		return body
	}
	receive(t, x, "ours.xml", "application/xml", "ksef", "2026-09-21", invoice("FS/1/2026", "8058083055", "5265877635"))
	receive(t, x, "theirs.xml", "application/xml", "ksef", "2026-09-21", invoice("FZ/9/2026", "5265877635", "7010001234"))
	receive(t, x, "bought.xml", "application/xml", "ksef", "2026-09-21", invoice("FZ/1/2026", "5265877635", "8058083055"))
	if res, err := adapter.Run(ctx, s, x, pl.KSeFInbound{}); err != nil || res.PassErr != nil {
		t.Fatalf("pass: %+v %v", res, err)
	}
	raws, _ := s.EventsByKind(ctx, core.KindRaw)
	var bought []string
	unreadable := 0
	for _, ev := range raws {
		var m map[string]any
		_ = json.Unmarshal(ev.Payload, &m)
		switch ev.Type {
		case "purchase.invoice.received":
			bought = append(bought, m["number"].(string))
		case adapter.EventDocumentUnreadable:
			unreadable++
		}
	}
	if len(bought) != 1 || bought[0] != "FZ/1/2026" || unreadable != 2 {
		t.Fatalf("bought %v, unreadable %d", bought, unreadable)
	}
}
