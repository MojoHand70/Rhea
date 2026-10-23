package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"sort"
	"strings"
	"time"

	"rhea/internal/core"
)

// PDFInvoice is the reader of the first model-read channel (DIRECTION
// 2026-10-10, "Documents in", step 6): a supplier's invoice as a PDF — the
// German supplier after KSeF, the real reason a Polish company still gets
// PDFs. The reading is an interpretation with an author: the model reads
// the file and answers with the invoice's fields and how sure it is; the
// reader validates the answer against the event's shape and its own
// arithmetic, and the channel's confidence policy decides: at or above the
// threshold the reading becomes purchase.invoice.received, carrying the
// reader and its confidence; below it a document.reading_proposed waits in
// the worklist for a person, who confirms it through a door and becomes the
// reading's author. The model's answer is data, never state (invariant 3);
// replay never re-reads (invariant 4): the reading is a raw event in the
// log with its author on it.
//
// Without a model (ModelReader nil) the reader reads the text layer — the
// labelled lines the simulator writes — and says so (reader "text-layer",
// confidence 1.00): the bench runs the channel without a model call.
type PDFInvoice struct {
	Today func() string
	// threshold comes from the channel declaration at binding time.
	threshold *big.Rat
	policy    string
}

const pdfInvoiceCodeVersion = 1

// EventDocumentReadingProposed is a reading below the channel's threshold:
// the proposed event, its confidence and the reason, waiting for a person.
const EventDocumentReadingProposed = "document.reading_proposed"

// ModelReader is the one call a model-read channel makes: a system prompt,
// the document's bytes, the instruction — the model's text back. The
// binary wires it to the agent (no store access); nil means no model, and
// the text layer is read instead.
var ModelReader func(ctx context.Context, system string, pdf []byte, user string) (string, error)

func init() {
	Register(Implemented{Name: "pdf_invoice", CodeVersion: pdfInvoiceCodeVersion,
		Format: "PDF supplier invoice from the document store, read by the model with a confidence (the text layer without one)"})
	readers["pdf_invoice"] = func() Adapter { return &PDFInvoice{threshold: big.NewRat(1, 1)} }
}

func (*PDFInvoice) Name() string     { return "pdf_invoice" }
func (*PDFInvoice) CodeVersion() int { return pdfInvoiceCodeVersion }

// Bound receives the channel the pass speaks through: its confidence
// policy is the reader's threshold.
func (r *PDFInvoice) Bound(ch core.Activity) {
	if ch.Spec.Channel == nil {
		return
	}
	r.policy = ch.Spec.Channel.Confidence.Policy
	// The gate is never off: a model reader with no threshold, or one the
	// decimal reader refuses, asks a person about every reading (1.00) —
	// a nil threshold booked every reading, at confidence 0.00 too.
	r.threshold = big.NewRat(1, 1)
	if t := ch.Spec.Channel.Confidence.Threshold; t != "" {
		if th, err := core.ParseDecimal(t); err == nil && th.Sign() > 0 && th.Cmp(big.NewRat(1, 1)) <= 0 {
			r.threshold = th
		}
	}
}

// proposedDocuments lists the documents a reader has proposed a reading
// of: left received until a person decides, they are not read again.
func proposedDocuments(ctx context.Context, reads Reads, channel string) (map[string]bool, error) {
	er, ok := reads.(interface {
		EventsOfTypes(ctx context.Context, types []string) ([]core.Event, error)
	})
	out := map[string]bool{}
	if !ok {
		return out, nil
	}
	evs, err := er.EventsOfTypes(ctx, []string{EventDocumentReadingProposed})
	if err != nil {
		return nil, err
	}
	for _, ev := range evs {
		var p struct {
			Document string `json:"document"`
			Channel  string `json:"channel"`
		}
		if json.Unmarshal(ev.Payload, &p) == nil && p.Channel == channel {
			out[p.Document] = true
		}
	}
	return out, nil
}

// ReadingSystem is what the model is told before it reads an invoice.
const ReadingSystem = `You read supplier invoices for an accounting system. You answer with ONE JSON object and
nothing else - no prose, no markdown fences. Every value is exactly as printed on the
document: numbers with a dot as the decimal separator and two decimals, dates as
YYYY-MM-DD, never guessed. If a value is not on the document, leave the key out.`

// ReadingPrompt is the instruction that goes with the document.
const ReadingPrompt = `Read this invoice and answer with this JSON:
{"number": "the invoice number", "invoice_date": "YYYY-MM-DD", "supplier": "the seller's name",
 "supplier_vat_id": "the seller's VAT id with its country prefix (DE..., PL...)",
 "buyer_nip": "the buyer's VAT id with its country prefix", "net": "0.00", "vat": "0.00",
 "gross": "0.00", "vat_rate": "the VAT percent as an integer string", "currency": "ISO code",
 "payment_term": "YYYY-MM-DD", "supplier_account": "the seller's IBAN",
 "for_document": "the delivery note it settles, if named",
 "confidence": "a decimal between 0 and 1: how sure you are that EVERY value above is exactly as printed"}`

// reading is the model's (or the text layer's) answer.
type reading struct {
	Number          string `json:"number"`
	InvoiceDate     string `json:"invoice_date"`
	Supplier        string `json:"supplier"`
	SupplierVATID   string `json:"supplier_vat_id"`
	BuyerNIP        string `json:"buyer_nip"`
	Net             string `json:"net"`
	VAT             string `json:"vat"`
	Gross           string `json:"gross"`
	VATRate         string `json:"vat_rate"`
	Currency        string `json:"currency"`
	PaymentTerm     string `json:"payment_term"`
	SupplierAccount string `json:"supplier_account"`
	ForDocument     string `json:"for_document"`
	Confidence      string `json:"confidence"`
}

func (r *PDFInvoice) Pass(ctx context.Context, reads Reads) ([]core.Event, error) {
	docs, err := reads.ObjectsByType(ctx, ReceivedDocumentType)
	if err != nil {
		return nil, err
	}
	// A document with a reading proposed waits for the person: read again
	// each pass, a later reading crossing the threshold booked the invoice,
	// and the person's confirmation of the earlier proposal booked it twice.
	proposed, err := proposedDocuments(ctx, reads, r.Name())
	if err != nil {
		return nil, err
	}
	logged, err := LoggedInvoices(ctx, reads)
	if err != nil {
		return nil, err
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].ID < docs[j].ID })
	var out []core.Event
	for _, d := range docs {
		status, _ := d.State["status"].(string)
		media, _ := d.State["media_type"].(string)
		hash, _ := d.State["hash"].(string)
		name, _ := d.State["name"].(string)
		if status != "received" || hash == "" || !strings.Contains(media, "pdf") || proposed[hash] {
			continue
		}
		b, err := reads.GetBlob(ctx, hash)
		if err != nil {
			return out, fmt.Errorf("%s names file %s: %w", d.ID, hash, err)
		}
		if !bytes.HasPrefix(b.Bytes, []byte("%PDF")) {
			continue // not a PDF whatever its label says: someone else's
		}
		answer, who, err := r.read(ctx, b.Bytes)
		if err != nil {
			out = append(out, r.unreadable(hash, name, err.Error()))
			continue
		}
		fact, confidence, err := r.validate(answer)
		if err != nil {
			out = append(out, r.unreadable(hash, name, err.Error()))
			continue
		}
		fact["document"] = hash
		fact["reader"] = who
		fact["confidence"] = confidence.FloatString(2)
		// An invoice outside KSeF is received the day its file reached the
		// store; a file from before the store kept that day, on its own date.
		fact["received_date"] = fact["invoice_date"]
		if on, ok := d.State["received_on"].(string); ok && on != "" {
			fact["received_date"] = on
		}
		payload, err := json.Marshal(fact)
		if err != nil {
			return out, err
		}
		date, _ := fact["invoice_date"].(string)
		// a resent scan of the same invoice is the same invoice, and so is
		// the one KSeF delivered: the fact's identity is the supplier and the
		// number, whatever the bytes and whichever the door (D2)
		dedup := "pdf_invoice/" + hash
		nip, _ := fact["supplier_nip"].(string)
		number, _ := fact["number"].(string)
		if key := PurchaseInvoiceKey(nip, number); key != "" {
			dedup = key
			if logged[key] {
				read, _ := json.Marshal(map[string]any{"document": hash, "channel": r.Name(), "event_type": "purchase.invoice.received", "logged_before": key})
				out = append(out, core.Event{Type: EventDocumentRead, OccurredAt: date, Payload: read, DedupKey: "pdf_invoice/read/" + hash})
				continue
			}
		}
		if confidence.Cmp(r.threshold) < 0 {
			// Doubt goes to the worklist: the reading as proposed, for a
			// person to confirm through the confirm_reading door — under
			// the reader's identity key, so a confirmation and a later
			// reading of the same invoice are one fact.
			proposal, _ := json.Marshal(map[string]any{
				"document": hash, "name": name, "channel": r.Name(), "event_type": "purchase.invoice.received",
				"payload": json.RawMessage(payload), "confidence": confidence.FloatString(2), "dedup_key": dedup,
				"reason": fmt.Sprintf("the reader is %s sure; the channel asks for %s", confidence.FloatString(2), r.threshold.FloatString(2)),
			})
			out = append(out, core.Event{Type: EventDocumentReadingProposed, OccurredAt: date,
				Payload: proposal, DedupKey: "pdf_invoice/proposed/" + hash})
			continue
		}
		read, _ := json.Marshal(map[string]any{"document": hash, "channel": r.Name(), "event_type": "purchase.invoice.received"})
		out = append(out,
			core.Event{Type: "purchase.invoice.received", OccurredAt: date, Payload: payload, DedupKey: dedup},
			core.Event{Type: EventDocumentRead, OccurredAt: date, Payload: read, DedupKey: "pdf_invoice/read/" + hash})
	}
	return out, nil
}

// read asks the model when there is one, the text layer otherwise, and
// names the reader.
func (r *PDFInvoice) read(ctx context.Context, pdf []byte) (reading, string, error) {
	var rd reading
	if ModelReader != nil {
		raw, err := ModelReader(ctx, ReadingSystem, pdf, ReadingPrompt)
		if err != nil {
			return rd, "", fmt.Errorf("the model could not read it: %v", err)
		}
		if err := json.Unmarshal([]byte(extractJSON(raw)), &rd); err != nil {
			return rd, "", fmt.Errorf("the model's reading is not the JSON asked for: %v", err)
		}
		return rd, "model", nil
	}
	rd, err := readTextLayer(pdf)
	if err != nil {
		return rd, "", err
	}
	return rd, "text-layer", nil
}

// validate turns a reading into the event's fields: money as decimal
// strings that parse, the invoice's own arithmetic (net plus VAT is the
// gross), a date, a confidence in [0, 1].
func (r *PDFInvoice) validate(rd reading) (map[string]any, *big.Rat, error) {
	switch {
	case rd.Number == "":
		return nil, nil, fmt.Errorf("no invoice number read")
	case rd.SupplierVATID == "":
		return nil, nil, fmt.Errorf("no seller VAT id read")
	case rd.Currency == "":
		return nil, nil, fmt.Errorf("no currency read")
	}
	if _, err := time.Parse("2006-01-02", rd.InvoiceDate); err != nil {
		return nil, nil, fmt.Errorf("invoice date %q is not a date", rd.InvoiceDate)
	}
	net, err := core.ParseMoney(rd.Net)
	if err != nil {
		return nil, nil, fmt.Errorf("net %q: %v", rd.Net, err)
	}
	vat, err := core.ParseMoney(rd.VAT)
	if err != nil {
		return nil, nil, fmt.Errorf("vat %q: %v", rd.VAT, err)
	}
	gross, err := core.ParseMoney(rd.Gross)
	if err != nil {
		return nil, nil, fmt.Errorf("gross %q: %v", rd.Gross, err)
	}
	if net+vat != gross {
		return nil, nil, fmt.Errorf("the reading contradicts itself: net %s + VAT %s is not the gross %s", rd.Net, rd.VAT, rd.Gross)
	}
	confidence, err := core.ParseDecimal(rd.Confidence)
	if err != nil || confidence.Sign() < 0 || confidence.Cmp(big.NewRat(1, 1)) > 0 {
		return nil, nil, fmt.Errorf("confidence %q is not a decimal between 0 and 1", rd.Confidence)
	}
	fact := map[string]any{
		"number": rd.Number, "supplier": rd.Supplier, "supplier_nip": rd.SupplierVATID, "invoice_date": rd.InvoiceDate,
		"net": core.FormatMoney(net), "vat": core.FormatMoney(vat), "gross": core.FormatMoney(gross), "currency": rd.Currency,
	}
	for k, v := range map[string]string{"buyer_nip": rd.BuyerNIP, "vat_rate": rd.VATRate, "payment_term": rd.PaymentTerm,
		"supplier_account": rd.SupplierAccount, "for_document": rd.ForDocument} {
		if v != "" {
			fact[k] = v
		}
	}
	return fact, confidence, nil
}

func (r *PDFInvoice) unreadable(hash, name, reason string) core.Event {
	payload, _ := json.Marshal(map[string]any{"document": hash, "name": name, "channel": r.Name(), "reason": reason})
	return core.Event{Type: EventDocumentUnreadable, OccurredAt: r.today(), Payload: payload, DedupKey: "pdf_invoice/unreadable/" + hash}
}

func (r *PDFInvoice) today() string {
	if r.Today != nil {
		return r.Today()
	}
	return core.Today()
}

// The text layer: the strings a PDF's content stream shows, as the
// simulator writes them — one "Label: value" per line. A real scan has no
// text layer, and a real PDF's stream is usually compressed: that is what
// the model is for.
var pdfTextRe = regexp.MustCompile(`\(((?:[^()\\]|\\.)*)\)\s*Tj`)

// PDFTextLines returns the text lines of an uncompressed PDF's content.
func PDFTextLines(pdf []byte) []string {
	var lines []string
	for _, m := range pdfTextRe.FindAllSubmatch(pdf, -1) {
		s := string(m[1])
		s = strings.NewReplacer(`\(`, "(", `\)`, ")", `\\`, `\`).Replace(s)
		lines = append(lines, s)
	}
	return lines
}

// readTextLayer reads the simulator's labelled lines into a reading, sure
// of every value it found: the text says what it says.
func readTextLayer(pdf []byte) (reading, error) {
	lines := PDFTextLines(pdf)
	if len(lines) == 0 {
		return reading{}, fmt.Errorf("no text layer: a scan needs the model")
	}
	rd := reading{Confidence: "1.00"}
	amount := func(v string) (string, string) { // "1234.00 PLN" → amount, currency
		parts := strings.Fields(v)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
		return v, ""
	}
	for _, l := range lines {
		label, value, ok := strings.Cut(l, ":")
		if !ok {
			continue
		}
		label, value = strings.TrimSpace(label), strings.TrimSpace(value)
		switch {
		case label == "Rechnungsnummer" || label == "Invoice number":
			rd.Number = value
		case label == "Rechnungsdatum" || label == "Invoice date":
			rd.InvoiceDate = value
		case label == "Verkaeufer" || label == "Seller":
			rd.Supplier = value
		case label == "USt-IdNr." || label == "VAT id":
			rd.SupplierVATID = value
		case label == "Kaeufer NIP" || label == "Buyer NIP":
			rd.BuyerNIP = value
		case label == "Lieferschein" || label == "Delivery note":
			rd.ForDocument = value
		case label == "Nettobetrag" || label == "Net":
			rd.Net, rd.Currency = amount(value)
		case strings.HasPrefix(label, "MwSt.") || strings.HasPrefix(label, "VAT "):
			rd.VAT, _ = amount(value)
			rd.VATRate = strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(label, "MwSt."), "VAT")), "%")
		case label == "Bruttobetrag" || label == "Gross":
			rd.Gross, _ = amount(value)
		case label == "Zahlungsziel" || label == "Payment term":
			rd.PaymentTerm = value
		case label == "IBAN":
			rd.SupplierAccount = value
		}
	}
	return rd, nil
}

// extractJSON takes the first complete JSON object out of a model's text.
func extractJSON(s string) string {
	start := strings.Index(s, "{")
	if start < 0 {
		return s
	}
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return s[start:]
}
