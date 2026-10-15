package sim

// Documents: the generator emits what a channel would deliver (DIRECTION
// 2026-10-10, "Documents in", step 3) — a FA-shaped XML per purchase invoice
// from a Polish supplier, as KSeF hands it to the buyer, and one MT940
// statement per month for the company's account, as a bank exports it. Each
// document stands for facts the generator also knows as events, named by
// index, so the eval can feed the file instead of the JSON and expect the
// same state. The model plays no part: the bytes are exact.
//
// What the files are not: a certified rendering of either standard. The FA
// subset carries the elements a purchase-invoice reader needs (parties,
// dates, number, net and VAT by rate, gross, lines, payment term, the WZ it
// settles) under the FA(3) namespace; the MT940 is the common tag set
// (:20:, :25:, :28C:, :60F:, :61:, :86:, :62F:) with the :86: line in one
// bank's dialect — "sim-bank": title, counterparty and account separated by
// "|" — because every bank writes that line differently and a reader is
// expected to keep it raw. The namespace and the rate slots must be checked
// against the published FA(3) schema before a real KSeF pull (step 4).

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"rhea/internal/adapter"
	"rhea/internal/core"
)

// Document is one generated file: a channel's delivery, with the facts it
// carries named by index into Facts.Events.
type Document struct {
	Path      string `json:"path"` // relative: <YYYY-MM>/ksef/<ref>.xml, <YYYY-MM>/bank/<iban>-<YYYY-MM>.sta
	MediaType string `json:"media_type"`
	Source    string `json:"source"` // the channel it would arrive through: ksef, bank
	Date      string `json:"date"`   // the business date of receipt
	Bytes     []byte `json:"-"`
	Events    []int  `json:"events"` // indexes into Facts.Events
}

// FA3Namespace and KSeFReference are the adapter's: the simulator writes
// the same FA(3) shape Rhea sends (internal/adapter/fa3.go), so what a
// persona receives and what Rhea issues are one document.
const FA3Namespace = adapter.FA3Namespace

// KSeFReference is the number KSeF assigns on acceptance, derived from the
// invoice so the same persona always yields the same reference.
func KSeFReference(sellerNIP, issueDate, number string) string {
	return adapter.KSeFReference(sellerNIP, issueDate, number)
}

// nip strips the country prefix a persona's vat_id carries: FA wants digits.
func nip(vatID string) string { return adapter.NIPDigits(vatID) }

// iban is a party's account, or a deterministic stand-in when the persona
// names none: "PL" and 26 digits drawn from the VAT id.
func iban(p Party) string {
	if p.IBAN != "" {
		return p.IBAN
	}
	h := sha256.Sum256([]byte(p.VATID))
	var sb strings.Builder
	sb.WriteString("PL")
	for i := 0; sb.Len() < 28; i++ {
		sb.WriteByte('0' + h[i%len(h)]%10)
	}
	return sb.String()
}

// documents turns the sorted events into files: one FA per purchase invoice
// from a Polish supplier, one PDF per purchase invoice from a foreign one,
// one MT940 per month.
func (g *gen) documents(events []Event) []Document {
	var out []Document
	out = append(out, g.faInvoices(events)...)
	out = append(out, g.pdfInvoices(events)...)
	out = append(out, g.statements(events)...)
	return out
}

// pdfInvoices: a foreign supplier is not in KSeF, so its invoice arrives as
// a PDF (DIRECTION 2026-10-10, the channel map: email, read by the model
// with a confidence). The file is a one-page text invoice in the supplier's
// words — a German layout for a German supplier — with a text layer, so a
// parser can read it too. Each document names the event it carries.
func (g *gen) pdfInvoices(events []Event) []Document {
	parties := map[string]Party{}
	for _, s := range g.p.Suppliers {
		parties[s.VATID] = s
	}
	items := map[string]Item{}
	for _, it := range g.p.Items {
		items[it.SKU] = it
	}
	var out []Document
	for i, ev := range events {
		if ev.EventType != "purchase.invoice.received" {
			continue
		}
		supplierNIP, _ := ev.Payload["supplier_nip"].(string)
		if strings.HasPrefix(supplierNIP, "PL") {
			continue // a Polish supplier's invoice is a KSeF document
		}
		sup := parties[supplierNIP]
		number, _ := ev.Payload["number"].(string)
		date, _ := ev.Payload["invoice_date"].(string)
		netStr, _ := ev.Payload["net"].(string)
		vatStr, _ := ev.Payload["vat"].(string)
		grossStr, _ := ev.Payload["gross"].(string)
		currency, _ := ev.Payload["currency"].(string)
		rateCode, _ := ev.Payload["vat_rate"].(string)
		percent := 0
		for _, r := range g.p.VATRates {
			if r.Code == rateCode {
				percent = r.Percent
			}
		}
		issued, _ := time.Parse("2006-01-02", date)
		lines := []string{
			"RECHNUNG",
			"Rechnungsnummer: " + number,
			"Rechnungsdatum: " + date,
			"Verkaeufer: " + sup.Name,
			"USt-IdNr.: " + sup.VATID,
			"Kaeufer: " + g.p.Company.Name,
			"Kaeufer NIP: " + g.p.Company.VATID,
		}
		if doc, ok := ev.Payload["for_document"].(string); ok && doc != "" {
			lines = append(lines, "Lieferschein: "+doc)
		}
		lines = append(lines,
			fmt.Sprintf("Nettobetrag: %s %s", netStr, currency),
			fmt.Sprintf("MwSt. %d%%: %s %s", percent, vatStr, currency),
			fmt.Sprintf("Bruttobetrag: %s %s", grossStr, currency),
			"Zahlungsziel: "+ymd(issued.AddDate(0, 0, g.p.Volume.PaymentDays)),
			"IBAN: "+iban(sup),
		)
		out = append(out, Document{
			Path:      filepath.Join(date[:7], "pdf", strings.ReplaceAll(number, "/", "-")+".pdf"),
			MediaType: "application/pdf", Source: "pdf", Date: date,
			Bytes: WritePDF(lines), Events: []int{i},
		})
	}
	return out
}

// WritePDF writes one page of text lines as a PDF with a text layer — no
// compression, one Type1 font, a correct cross-reference table — so a
// person, a text extractor and the model all read the same thing.
func WritePDF(lines []string) []byte {
	var content bytes.Buffer
	content.WriteString("BT\n/F1 11 Tf\n14 TL\n50 790 Td\n")
	for _, l := range lines {
		fmt.Fprintf(&content, "(%s) Tj T*\n", pdfEscape(l))
	}
	content.WriteString("ET\n")
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", content.Len(), content.String()),
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>",
	}
	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects)+1)
	for i, o := range objects {
		offsets[i+1] = out.Len()
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", i+1, o)
	}
	xref := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for i := 1; i <= len(objects); i++ {
		fmt.Fprintf(&out, "%010d 00000 n \n", offsets[i])
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)
	return out.Bytes()
}

// pdfEscape writes a string for a PDF literal: the three characters that
// need escaping, and anything outside ASCII as a question mark — a text
// layer is kept plain so every reader agrees on it.
func pdfEscape(s string) string {
	var sb strings.Builder
	for _, r := range s {
		switch {
		case r == '\\' || r == '(' || r == ')':
			sb.WriteByte('\\')
			sb.WriteRune(r)
		case r > 126 || r < 32:
			sb.WriteByte('?')
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func (g *gen) faInvoices(events []Event) []Document {
	parties := map[string]Party{}
	for _, s := range g.p.Suppliers {
		parties[s.VATID] = s
	}
	items := map[string]Item{}
	for _, it := range g.p.Items {
		items[it.SKU] = it
	}
	deliveries := map[string]Event{}
	for _, ev := range events {
		if ev.EventType == "delivery.received" {
			deliveries[ev.Payload["document"].(string)] = ev
		}
	}
	var out []Document
	for i, ev := range events {
		if ev.EventType != "purchase.invoice.received" {
			continue
		}
		supplierNIP, _ := ev.Payload["supplier_nip"].(string)
		if !strings.HasPrefix(supplierNIP, "PL") {
			continue // a foreign supplier's invoice is a PDF, not a KSeF document
		}
		sup := parties[supplierNIP]
		number, _ := ev.Payload["number"].(string)
		date, _ := ev.Payload["invoice_date"].(string)
		netStr, _ := ev.Payload["net"].(string)
		net, _ := core.ParseMoney(netStr)
		rateCode, _ := ev.Payload["vat_rate"].(string)
		percent := 0
		for _, r := range g.p.VATRates {
			if r.Code == rateCode {
				percent = r.Percent
			}
		}
		vat := halfUp(net*int64(percent), 100)
		issued, _ := time.Parse("2006-01-02", date)

		var lines []adapter.FA3Line
		var wz []string
		if doc, ok := ev.Payload["for_document"].(string); ok && doc != "" {
			wz = append(wz, doc)
			if d, ok := deliveries[doc]; ok {
				for _, l := range d.Payload["lines"].([]any) {
					line := l.(map[string]any)
					qty := line["qty"].(int)
					price, _ := core.ParseMoney(line["price"].(string))
					it := items[line["item"].(string)]
					lines = append(lines, adapter.FA3Line{Name: it.Name, Unit: it.Unit, Qty: fmt.Sprint(qty),
						UnitPrice: price, Net: int64(qty) * price, RateCode: rateCode})
				}
			}
		}
		if len(lines) == 0 {
			lines = append(lines, adapter.FA3Line{Name: "Usługa", Unit: "usł.", Qty: "1", UnitPrice: net, Net: net, RateCode: rateCode})
		}
		body, err := adapter.WriteFA3(adapter.FA3Invoice{
			Seller:   adapter.FA3Party{NIP: nip(sup.VATID), Name: sup.Name, Country: sup.Country},
			Buyer:    adapter.FA3Party{NIP: nip(g.p.Company.VATID), Name: g.p.Company.Name, Country: g.p.Company.Country},
			Currency: "PLN", IssueDate: date, Number: number, WZ: wz, RateCode: rateCode,
			Net: net, VAT: vat, Gross: net + vat, Lines: lines,
			Term:    issued.AddDate(0, 0, g.p.Volume.PaymentDays).Format("2006-01-02"),
			Account: iban(sup), Created: issued.Format("2006-01-02") + "T08:00:00Z", SystemInfo: "Rhea sim " + g.p.Name,
		})
		if err != nil {
			panic(err) // the persona's rates are the pack's; a slot always exists
		}
		ref := KSeFReference(nip(sup.VATID), date, number)
		out = append(out, Document{
			Path:      filepath.Join(date[:7], "ksef", ref+".xml"),
			MediaType: "application/xml", Source: "ksef", Date: date,
			Bytes:  body,
			Events: []int{i},
		})
	}
	return out
}

// StatementOpening is the balance every persona's account starts the first
// month with, in minor units.
const StatementOpening int64 = 1_000_000

// statements: one MT940 per month, every bank line of the month inside,
// balances carried from month to month.
func (g *gen) statements(events []Event) []Document {
	start, _ := time.Parse("2006-01", g.p.From)
	byMonth := map[string][]int{}
	for i, ev := range events {
		if ev.EventType == "bank.statement.line" {
			byMonth[ev.OccurredAt[:7]] = append(byMonth[ev.OccurredAt[:7]], i)
		}
	}
	account := iban(g.p.Company)
	balance := StatementOpening
	var out []Document
	for m := 0; m < g.p.Months; m++ {
		first := start.AddDate(0, m, 0)
		last := first.AddDate(0, 1, -1)
		month := first.Format("2006-01")
		idx := byMonth[month]
		sort.Ints(idx)
		var sb bytes.Buffer
		w := func(f string, a ...any) { fmt.Fprintf(&sb, f+"\r\n", a...) }
		w(":20:STMT%s", first.Format("200601"))
		w(":25:%s", account)
		w(":28C:%d/%d", m+1, first.Year())
		w(":60F:%s", mt940Balance(balance, first))
		for n, i := range idx {
			ev := events[i]
			amount, _ := core.ParseMoney(ev.Payload["amount"].(string))
			day, _ := time.Parse("2006-01-02", ev.OccurredAt)
			ref, _ := ev.Payload["reference"].(string)
			who, _ := ev.Payload["counterparty"].(string)
			whoAccount, _ := ev.Payload["counterparty_account"].(string)
			mark, kind := "C", "PRZELEW PRZYCHODZACY"
			if ev.Payload["direction"] == "debit" {
				mark, kind, amount = "D", "PRZELEW WYCHODZACY", -amount
			}
			w(":61:%s%s%s%sNTRFNONREF//%d", day.Format("060102"), day.Format("0102"), mark, mt940Amount(core.Abs(amount)), n+1)
			w(":86:%s|%s|%s|%s", kind, ref, who, whoAccount)
			balance += amount
		}
		w(":62F:%s", mt940Balance(balance, last))
		w("-")
		out = append(out, Document{
			Path:      filepath.Join(month, "bank", account+"-"+month+".sta"),
			MediaType: "text/plain", Source: "bank", Date: last.Format("2006-01-02"),
			Bytes: sb.Bytes(), Events: idx,
		})
	}
	return out
}

// mt940Amount writes minor units the way the format does: a comma decimal.
func mt940Amount(minor int64) string {
	return strings.ReplaceAll(core.FormatMoney(minor), ".", ",")
}

func mt940Balance(minor int64, day time.Time) string {
	sign := "C"
	if minor < 0 {
		sign, minor = "D", -minor
	}
	return sign + day.Format("060102") + "PLN" + mt940Amount(minor)
}

// WriteDocuments writes a persona's documents under dir/<persona>/ and
// returns how many files landed.
func WriteDocuments(dir string, p Persona, docs []Document) (int, error) {
	for _, d := range docs {
		path := filepath.Join(dir, p.Name, d.Path)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return 0, err
		}
		if err := os.WriteFile(path, d.Bytes, 0o644); err != nil {
			return 0, err
		}
	}
	return len(docs), nil
}
