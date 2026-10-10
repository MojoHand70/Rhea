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
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

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

// FA3Namespace is the schema namespace the files declare.
const FA3Namespace = "http://crd.gov.pl/wzor/2025/06/25/13775/"

// The FA subset, as encoding/xml sees it.
type faInvoice struct {
	XMLName  xml.Name `xml:"Faktura"`
	XMLNS    string   `xml:"xmlns,attr"`
	Naglowek faHeader `xml:"Naglowek"`
	Podmiot1 faParty  `xml:"Podmiot1"`
	Podmiot2 faParty  `xml:"Podmiot2"`
	Fa       faBody   `xml:"Fa"`
}

type faHeader struct {
	KodFormularza     faCode `xml:"KodFormularza"`
	WariantFormularza int    `xml:"WariantFormularza"`
	DataWytworzeniaFa string `xml:"DataWytworzeniaFa"`
	SystemInfo        string `xml:"SystemInfo"`
}

type faCode struct {
	KodSystemowy string `xml:"kodSystemowy,attr"`
	WersjaSchemy string `xml:"wersjaSchemy,attr"`
	Value        string `xml:",chardata"`
}

type faParty struct {
	Dane  faIdent   `xml:"DaneIdentyfikacyjne"`
	Adres faAddress `xml:"Adres"`
}

type faIdent struct {
	NIP   string `xml:"NIP"`
	Nazwa string `xml:"Nazwa"`
}

type faAddress struct {
	KodKraju string `xml:"KodKraju"`
	AdresL1  string `xml:"AdresL1"`
}

// faAmount is one P_13_k / P_14_k slot: net and VAT by rate, named at
// generation so the one rate a persona uses lands in its slot.
type faAmount struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

type faBody struct {
	KodWaluty     string     `xml:"KodWaluty"`
	P1            string     `xml:"P_1"` // issue date
	P2            string     `xml:"P_2"` // number
	WZ            []string   `xml:"WZ,omitempty"`
	Totals        []faAmount // P_13_k, P_14_k
	P15           string     `xml:"P_15"` // gross
	RodzajFaktury string     `xml:"RodzajFaktury"`
	FaWiersz      []faLine   `xml:"FaWiersz"`
	Platnosc      faPayment  `xml:"Platnosc"`
}

type faLine struct {
	NrWierszaFa int    `xml:"NrWierszaFa"`
	P7          string `xml:"P_7"`  // name
	P8A         string `xml:"P_8A"` // unit
	P8B         string `xml:"P_8B"` // quantity
	P9A         string `xml:"P_9A"` // unit price net
	P11         string `xml:"P_11"` // line net
	P12         string `xml:"P_12"` // rate
}

type faPayment struct {
	Termin   faTerm    `xml:"TerminPlatnosci"`
	Rachunek faAccount `xml:"RachunekBankowy"`
}

type faTerm struct {
	Termin string `xml:"Termin"`
}

type faAccount struct {
	NrRB string `xml:"NrRB"`
}

// faRateSlot maps a VAT percent to the FA slot its net and VAT go in.
func faRateSlot(percent int) (string, bool) {
	switch percent {
	case 23, 22:
		return "1", true
	case 8, 7:
		return "2", true
	case 5:
		return "3", true
	case 0:
		return "6", true
	}
	return "", false
}

// nip strips the country prefix a persona's vat_id carries: FA wants digits.
func nip(vatID string) string { return strings.TrimPrefix(vatID, "PL") }

// KSeFReference is the number KSeF assigns on acceptance, derived here from
// the invoice so the same persona always yields the same reference:
// <NIP>-<YYYYMMDD>-<12 hex>-<2 hex>.
func KSeFReference(sellerNIP, issueDate, number string) string {
	h := sha256.Sum256([]byte(sellerNIP + "|" + issueDate + "|" + number))
	return fmt.Sprintf("%s-%s-%012X-%02X", sellerNIP, strings.ReplaceAll(issueDate, "-", ""), h[:6], h[6])
}

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
// from a Polish supplier, one MT940 per month.
func (g *gen) documents(events []Event) []Document {
	var out []Document
	out = append(out, g.faInvoices(events)...)
	out = append(out, g.statements(events)...)
	return out
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
		slot, _ := faRateSlot(percent)
		issued, _ := time.Parse("2006-01-02", date)

		var lines []faLine
		var wz []string
		if doc, ok := ev.Payload["for_document"].(string); ok && doc != "" {
			wz = append(wz, doc)
			if d, ok := deliveries[doc]; ok {
				for n, l := range d.Payload["lines"].([]any) {
					line := l.(map[string]any)
					qty := line["qty"].(int)
					price, _ := core.ParseMoney(line["price"].(string))
					it := items[line["item"].(string)]
					lines = append(lines, faLine{NrWierszaFa: n + 1, P7: it.Name, P8A: it.Unit,
						P8B: fmt.Sprint(qty), P9A: core.FormatMoney(price),
						P11: core.FormatMoney(int64(qty) * price), P12: fmt.Sprint(percent)})
				}
			}
		}
		if len(lines) == 0 {
			lines = append(lines, faLine{NrWierszaFa: 1, P7: "Usługa", P8A: "usł.", P8B: "1",
				P9A: core.FormatMoney(net), P11: core.FormatMoney(net), P12: fmt.Sprint(percent)})
		}
		inv := faInvoice{
			XMLNS: FA3Namespace,
			Naglowek: faHeader{
				KodFormularza:     faCode{KodSystemowy: "FA (3)", WersjaSchemy: "1-0E", Value: "FA"},
				WariantFormularza: 3,
				DataWytworzeniaFa: issued.Format("2006-01-02") + "T08:00:00Z",
				SystemInfo:        "Rhea sim " + g.p.Name,
			},
			Podmiot1: faParty{Dane: faIdent{NIP: nip(sup.VATID), Nazwa: sup.Name},
				Adres: faAddress{KodKraju: sup.Country, AdresL1: sup.Name}},
			Podmiot2: faParty{Dane: faIdent{NIP: nip(g.p.Company.VATID), Nazwa: g.p.Company.Name},
				Adres: faAddress{KodKraju: g.p.Company.Country, AdresL1: g.p.Company.Name}},
			Fa: faBody{
				KodWaluty: "PLN", P1: date, P2: number, WZ: wz,
				Totals: []faAmount{
					{XMLName: xml.Name{Local: "P_13_" + slot}, Value: core.FormatMoney(net)},
					{XMLName: xml.Name{Local: "P_14_" + slot}, Value: core.FormatMoney(vat)},
				},
				P15: core.FormatMoney(net + vat), RodzajFaktury: "VAT", FaWiersz: lines,
				Platnosc: faPayment{
					Termin:   faTerm{Termin: issued.AddDate(0, 0, g.p.Volume.PaymentDays).Format("2006-01-02")},
					Rachunek: faAccount{NrRB: iban(sup)},
				},
			},
		}
		body, err := xml.MarshalIndent(inv, "", "  ")
		if err != nil {
			panic(err) // a struct of strings cannot fail to marshal
		}
		ref := KSeFReference(nip(sup.VATID), date, number)
		out = append(out, Document{
			Path:      filepath.Join(date[:7], "ksef", ref+".xml"),
			MediaType: "application/xml", Source: "ksef", Date: date,
			Bytes:  append([]byte(xml.Header), append(body, '\n')...),
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
	customers := map[string]Party{}
	for _, c := range g.p.Customers {
		customers[c.Name] = c
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
			w(":61:%s%sC%sNTRFNONREF//%d", day.Format("060102"), day.Format("0102"), mt940Amount(amount), n+1)
			w(":86:PRZELEW PRZYCHODZACY|%s|%s|%s", ref, who, iban(customers[who]))
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
