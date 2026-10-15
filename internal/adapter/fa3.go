package adapter

import (
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"strings"

	"rhea/internal/core"
)

// The FA(3) invoice as Rhea writes it (DIRECTION 2026-10-15, "Documents
// out"): the subset a sole trader's sales invoice needs, in the order the
// Ministry's schema fixes — header, seller, buyer, the body with its rate
// slots, the mandatory annotation block, the kind, the lines, the payment.
// The simulator writes the same shape for the invoices a persona receives,
// so what Rhea sends and what Rhea reads are one document.

// FA3Namespace is the schema namespace the files declare.
const FA3Namespace = "http://crd.gov.pl/wzor/2025/06/25/13775/"

// FA3Invoice is what a writer fills; amounts in minor units.
type FA3Invoice struct {
	Seller, Buyer FA3Party
	Currency      string
	IssueDate     string // P_1, YYYY-MM-DD
	Number        string // P_2
	WZ            []string
	RateCode      string // the pack's vat_rate code: 23, 8, 5, 0, zw
	Net, VAT      int64
	Gross         int64
	Lines         []FA3Line // empty: one line for the whole net
	Term          string    // payment term, YYYY-MM-DD; empty omits it
	Account       string    // the seller's IBAN; empty omits it
	Created       string    // DataWytworzeniaFa, RFC 3339
	SystemInfo    string
}

// FA3Party is one side: NIP as digits, name, country code, one address line.
type FA3Party struct {
	NIP, Name, Country, Address string
}

// FA3Line is one invoice line; the rate code names its P_12.
type FA3Line struct {
	Name, Unit, Qty string
	UnitPrice, Net  int64
	RateCode        string
}

// The schema, as encoding/xml sees it.
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
	SystemInfo        string `xml:"SystemInfo,omitempty"`
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

// faAmount is one P_13_k / P_14_k slot, named at writing so the one rate an
// invoice uses lands in its slot.
type faAmount struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

type faBody struct {
	KodWaluty     string        `xml:"KodWaluty"`
	P1            string        `xml:"P_1"` // issue date
	P2            string        `xml:"P_2"` // number
	WZ            []string      `xml:"WZ,omitempty"`
	Totals        []faAmount    // P_13_k, P_14_k
	P15           string        `xml:"P_15"` // gross
	Adnotacje     faAnnotations `xml:"Adnotacje"`
	RodzajFaktury string        `xml:"RodzajFaktury"`
	FaWiersz      []faLine      `xml:"FaWiersz"`
	Platnosc      *faPayment    `xml:"Platnosc,omitempty"`
}

// faAnnotations is the mandatory block: cash method, self-invoicing,
// reverse charge, split payment — each "2" for no — and the three choices
// (exemption, new means of transport, margin) each taken as "none" unless
// the invoice is exempt, which the exemption then declares.
type faAnnotations struct {
	P16        string      `xml:"P_16"`
	P17        string      `xml:"P_17"`
	P18        string      `xml:"P_18"`
	P18A       string      `xml:"P_18A"`
	Zwolnienie faExemption `xml:"Zwolnienie"`
	NoweSrodki faNone22    `xml:"NoweSrodkiTransportu"`
	P23        string      `xml:"P_23"`
	PMarzy     faNoneMarzy `xml:"PMarzy"`
}

type faExemption struct {
	P19  string `xml:"P_19,omitempty"`
	P19C string `xml:"P_19C,omitempty"` // the legal basis, free text
	P19N string `xml:"P_19N,omitempty"`
}

type faNone22 struct {
	P22N string `xml:"P_22N"`
}

type faNoneMarzy struct {
	PMarzyN string `xml:"P_PMarzyN"`
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
	Termin   *faTerm    `xml:"TerminPlatnosci,omitempty"`
	Rachunek *faAccount `xml:"RachunekBankowy,omitempty"`
}

type faTerm struct {
	Termin string `xml:"Termin"`
}

type faAccount struct {
	NrRB string `xml:"NrRB"`
}

// faRateSlot maps a pack rate code to the FA slot its net and VAT go in.
func faRateSlot(code string) (string, bool) {
	switch code {
	case "23", "22":
		return "1", true
	case "8", "7":
		return "2", true
	case "5":
		return "3", true
	case "0":
		return "6_1", true // domestic 0%: FA(3) splits the zero rate (6_1 home, 6_2 WDT, 6_3 export)
	case "zw":
		return "7", true
	}
	return "", false
}

// faRate is the P_12 spelling of a pack rate code.
func faRate(code string) string {
	if code == "0" {
		return "0 KR"
	}
	return code
}

// zeroRateSlot reports a slot that carries no VAT element: FA(3) has
// P_13_6_k but no P_14_6_k, and P_13_7 (exempt) likewise.
func zeroRateSlot(slot string) bool { return strings.HasPrefix(slot, "6") || slot == "7" }

// WriteFA3 renders an invoice as the FA(3) document KSeF validates: the
// XML prolog, UTF-8, no processing instructions.
func WriteFA3(inv FA3Invoice) ([]byte, error) {
	slot, ok := faRateSlot(inv.RateCode)
	if !ok {
		return nil, fmt.Errorf("invoice %s: no FA(3) slot for VAT rate %q", inv.Number, inv.RateCode)
	}
	if inv.Seller.NIP == "" || inv.Seller.Name == "" {
		return nil, fmt.Errorf("invoice %s: the seller needs a NIP and a name", inv.Number)
	}
	if inv.Buyer.NIP == "" {
		return nil, fmt.Errorf("invoice %s: the buyer needs a NIP", inv.Number)
	}
	totals := []faAmount{{XMLName: xml.Name{Local: "P_13_" + slot}, Value: core.FormatMoney(inv.Net)}}
	if !zeroRateSlot(slot) {
		totals = append(totals, faAmount{XMLName: xml.Name{Local: "P_14_" + slot}, Value: core.FormatMoney(inv.VAT)})
	}
	var lines []faLine
	for n, l := range inv.Lines {
		code := l.RateCode
		if code == "" {
			code = inv.RateCode
		}
		lines = append(lines, faLine{NrWierszaFa: n + 1, P7: l.Name, P8A: l.Unit, P8B: l.Qty,
			P9A: core.FormatMoney(l.UnitPrice), P11: core.FormatMoney(l.Net), P12: faRate(code)})
	}
	if len(lines) == 0 {
		lines = []faLine{{NrWierszaFa: 1, P7: "Sprzedaż wg faktury " + inv.Number, P8A: "szt.", P8B: "1",
			P9A: core.FormatMoney(inv.Net), P11: core.FormatMoney(inv.Net), P12: faRate(inv.RateCode)}}
	}
	annotations := faAnnotations{P16: "2", P17: "2", P18: "2", P18A: "2",
		Zwolnienie: faExemption{P19N: "1"}, NoweSrodki: faNone22{P22N: "1"}, P23: "2", PMarzy: faNoneMarzy{PMarzyN: "1"}}
	if slot == "7" {
		annotations.Zwolnienie = faExemption{P19: "1", P19C: "zwolnienie z VAT"}
	}
	var payment *faPayment
	if inv.Term != "" || inv.Account != "" {
		payment = &faPayment{}
		if inv.Term != "" {
			payment.Termin = &faTerm{Termin: inv.Term}
		}
		if inv.Account != "" {
			payment.Rachunek = &faAccount{NrRB: inv.Account}
		}
	}
	party := func(p FA3Party) faParty {
		country, address := p.Country, p.Address
		if country == "" {
			country = "PL"
		}
		if address == "" {
			address = p.Name
		}
		return faParty{Dane: faIdent{NIP: p.NIP, Nazwa: p.Name}, Adres: faAddress{KodKraju: country, AdresL1: address}}
	}
	doc := faInvoice{
		XMLNS: FA3Namespace,
		Naglowek: faHeader{
			KodFormularza:     faCode{KodSystemowy: "FA (3)", WersjaSchemy: "1-0E", Value: "FA"},
			WariantFormularza: 3,
			DataWytworzeniaFa: inv.Created,
			SystemInfo:        inv.SystemInfo,
		},
		Podmiot1: party(inv.Seller),
		Podmiot2: party(inv.Buyer),
		Fa: faBody{
			KodWaluty: inv.Currency, P1: inv.IssueDate, P2: inv.Number, WZ: inv.WZ,
			Totals: totals, P15: core.FormatMoney(inv.Gross), Adnotacje: annotations,
			RodzajFaktury: "VAT", FaWiersz: lines, Platnosc: payment,
		},
	}
	body, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), append(body, '\n')...), nil
}

// KSeFReference is the number KSeF assigns on acceptance, derived here from
// the invoice so a simulated persona and the fake client always yield the
// same one: <NIP>-<YYYYMMDD>-<12 hex>-<CRC-8>, the Ministry's shape with
// its checksum, so it passes KSeFNumberValid.
func KSeFReference(sellerNIP, issueDate, number string) string {
	h := sha256.Sum256([]byte(sellerNIP + "|" + issueDate + "|" + number))
	data := fmt.Sprintf("%s-%s-%012X", sellerNIP, strings.ReplaceAll(issueDate, "-", ""), h[:6])
	return fmt.Sprintf("%s-%02X", data, crc8([]byte(data)))
}

// NIPDigits strips the country prefix a vat_id carries: FA wants digits.
func NIPDigits(vatID string) string { return strings.TrimPrefix(strings.TrimSpace(vatID), "PL") }
