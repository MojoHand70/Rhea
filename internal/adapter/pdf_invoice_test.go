package adapter_test

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"rhea/internal/adapter"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/sim"
	"rhea/internal/store/storetest"
)

// The first model-read channel (DIRECTION 2026-10-10, step 6): a foreign
// supplier's PDF invoice. Without a model the reader reads the text layer
// and says so; with a model that is not sure enough, the reading waits as a
// proposal and a person confirms it through a door, becoming its author —
// the typed event caused by the proposal, which leaves the worklist. A
// reading that contradicts itself is unreadable, never booked.
func TestPDFInvoiceChannel(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	for _, pk := range []string{"common/documents.json", "pl/pdf-invoice.json"} {
		sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", pk), "test")
		if err != nil {
			t.Fatal(err)
		}
		if _, _, _, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-09-30"); err != nil {
			t.Fatal(err)
		}
	}
	p, _, err := sim.Load(filepath.Join("..", "..", "testdata", "customers", "nordwind.json"))
	if err != nil {
		t.Fatal(err)
	}
	facts, err := sim.Generate(p) // the whole year: the German supplier's invoices are a share of the purchases
	if err != nil {
		t.Fatal(err)
	}
	var pdfs []sim.Document
	for _, d := range facts.Documents {
		if d.Source == "pdf" {
			pdfs = append(pdfs, d)
		}
	}
	if len(pdfs) < 3 {
		t.Fatalf("Nordwind's German supplier sends %d PDFs; the test needs three", len(pdfs))
	}
	receive := func(d sim.Document) {
		t.Helper()
		if _, err := x.TriggerActivity(ctx, "receive_document", map[string]any{
			"file":   map[string]any{"name": filepath.Base(d.Path), "media_type": d.MediaType, "bytes": d.Bytes},
			"source": "email", "occurred_at": d.Date}, "shell:krzysztof", ""); err != nil {
			t.Fatal(err)
		}
	}
	reader, _ := adapter.Lookup("pdf_invoice")

	// 1. The text layer, read exactly: every PDF becomes the invoice it carries.
	adapter.ModelReader = nil
	receive(pdfs[0])
	res, err := adapter.Run(ctx, s, x, reader)
	if err != nil || res.PassErr != nil {
		t.Fatalf("text-layer pass: %v %v", err, res.PassErr)
	}
	if res.Submitted != 2 {
		t.Fatalf("one invoice and one document.read: %+v", res)
	}
	fact := facts.Events[pdfs[0].Events[0]].Payload
	var read core.Event
	for _, ev := range res.Events {
		if ev.Type == "purchase.invoice.received" {
			read = ev
		}
	}
	var got map[string]any
	json.Unmarshal(read.Payload, &got)
	for _, k := range []string{"number", "invoice_date", "net", "vat", "gross", "currency", "supplier_nip", "buyer_nip", "vat_rate"} {
		if fmt.Sprint(got[k]) != fmt.Sprint(fact[k]) {
			t.Fatalf("%s: read %v, the fact says %v", k, got[k], fact[k])
		}
	}
	if got["reader"] != "text-layer" || got["confidence"] != "1.00" || got["market"] != "pl" || got["document"] == nil || read.ActivityName != "pdf_invoice" {
		t.Fatalf("the reading names its author, its confidence, its market and its file: %v %s", got, read.ActivityName)
	}
	docs, _ := s.ObjectsByType(ctx, adapter.ReceivedDocumentType)
	if len(docs) != 1 || docs[0].State["status"] != "read" {
		t.Fatalf("the document is read: %v", docs)
	}

	// 2. A model that is not sure enough: the reading waits as a proposal.
	unsure := func(ctx context.Context, system string, pdf []byte, user string) (string, error) {
		rd, _ := readTextLayerFor(pdf)
		rd["confidence"] = "0.60"
		b, _ := json.Marshal(rd)
		return "Here is the invoice:\n" + string(b), nil
	}
	adapter.ModelReader = unsure
	receive(pdfs[1])
	res, err = adapter.Run(ctx, s, x, reader)
	if err != nil || res.PassErr != nil || res.Submitted != 1 || res.Events[0].Type != adapter.EventDocumentReadingProposed {
		t.Fatalf("an unsure reading proposes: %+v %v %v", res, err, res.PassErr)
	}
	var proposal core.Event
	for _, ev := range res.Events {
		proposal = ev
	}
	var proposed map[string]any
	json.Unmarshal(proposal.Payload, &proposed)
	if !strings.Contains(proposed["reason"].(string), "0.60 sure; the channel asks for 0.90") {
		t.Fatalf("the proposal says why: %v", proposed)
	}
	// The proposal waits in the worklist (beside the first invoice, which
	// no rule books in this bare world).
	left, _ := s.UnmatchedRawEvents(ctx)
	var proposalID int64
	for _, e := range left {
		if e.Type == adapter.EventDocumentReadingProposed {
			proposalID = e.ID
		}
	}
	if proposalID == 0 {
		t.Fatalf("the proposal waits in the worklist: %v", left)
	}
	docs, _ = s.ObjectsByType(ctx, adapter.ReceivedDocumentType)
	if docs[1].State["status"] != "received" {
		t.Fatalf("a proposed reading leaves the document unread: %v", docs[1].State)
	}

	// 3. A person confirms: the typed event, the person its author, the
	// proposal its cause and gone from the worklist, the document read.
	conf, err := x.TriggerActivity(ctx, "confirm_reading", map[string]any{"proposal": proposalID}, "shell:krzysztof", "")
	if err != nil {
		t.Fatal(err)
	}
	ev, _ := s.GetEvent(ctx, conf.EventID)
	json.Unmarshal(ev.Payload, &got)
	if ev.Type != "purchase.invoice.received" || ev.Actor != "shell:krzysztof" || ev.CauseEventID == nil || *ev.CauseEventID != proposalID ||
		got["reader"] != "person" || fmt.Sprint(got["confirmed_from"]) != fmt.Sprint(proposalID) || got["number"] != facts.Events[pdfs[1].Events[0]].Payload["number"] {
		t.Fatalf("the confirmed reading: %+v %v", ev, got)
	}
	left, _ = s.UnmatchedRawEvents(ctx)
	for _, e := range left {
		if e.ID == proposalID {
			t.Fatal("a confirmed proposal still waits")
		}
	}
	docs, _ = s.ObjectsByType(ctx, adapter.ReceivedDocumentType)
	if docs[1].State["status"] != "read" {
		t.Fatalf("the confirmed document is read: %v", docs[1].State)
	}
	if _, err := x.TriggerActivity(ctx, "confirm_reading", map[string]any{"proposal": proposalID}, "shell:krzysztof", ""); err == nil || !strings.Contains(err.Error(), "already confirmed") {
		t.Fatalf("confirming twice: %v", err)
	}

	// 4. A reading that contradicts itself is unreadable, never booked.
	adapter.ModelReader = func(ctx context.Context, system string, pdf []byte, user string) (string, error) {
		return `{"number":"X","invoice_date":"2026-01-02","supplier_vat_id":"DE1","net":"100.00","vat":"23.00","gross":"124.00","currency":"PLN","confidence":"0.99"}`, nil
	}
	receive(pdfs[2])
	res, _ = adapter.Run(ctx, s, x, reader)
	if res.Submitted != 1 || res.Events[0].Type != adapter.EventDocumentUnreadable || !strings.Contains(string(res.Events[0].Payload), "contradicts itself") {
		t.Fatalf("a contradiction is unreadable: %+v", res)
	}
	adapter.ModelReader = nil
}

// readTextLayerFor reads the simulator's labelled lines the way the reader's
// text-layer path does, as a map a fake model can edit.
func readTextLayerFor(pdf []byte) (map[string]any, error) {
	out := map[string]any{}
	for _, l := range adapter.PDFTextLines(pdf) {
		label, value, ok := strings.Cut(l, ":")
		if !ok {
			continue
		}
		label, value = strings.TrimSpace(label), strings.TrimSpace(value)
		amount := func(v string) string { return strings.Fields(v)[0] }
		switch {
		case label == "Rechnungsnummer":
			out["number"] = value
		case label == "Rechnungsdatum":
			out["invoice_date"] = value
		case label == "Verkaeufer":
			out["supplier"] = value
		case label == "USt-IdNr.":
			out["supplier_vat_id"] = value
		case label == "Kaeufer NIP":
			out["buyer_nip"] = value
		case label == "Nettobetrag":
			out["net"], out["currency"] = amount(value), strings.Fields(value)[1]
		case strings.HasPrefix(label, "MwSt."):
			out["vat"] = amount(value)
			out["vat_rate"] = strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(label, "MwSt.")), "%")
		case label == "Bruttobetrag":
			out["gross"] = amount(value)
		case label == "IBAN":
			out["supplier_account"] = value
		}
	}
	return out, nil
}
