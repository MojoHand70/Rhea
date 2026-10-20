package pl

import (
	"context"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	"rhea/internal/adapter"
	"rhea/internal/core"
)

// JPK_V7M(3): the monthly VAT return with its records, as the Ministry's
// schema of 19 December 2025 lays it out (in force for periods from
// February 2026). A writer, not a rule (DIRECTION 2026-10-15, "Month end"):
// every number here is read from objects the kernel derived — the month's
// vat_settlement for the declaration, the invoices for the records, the
// company and its owner for the taxpayer, the KSeF submissions for the
// numbers — and rendered in the form's units. The one arithmetic the
// writer does is the form's own: declaration positions are whole złoty
// (tax ordinance, art. 63 §1), records keep their grosze.

const (
	JPKNamespace = "http://crd.gov.pl/wzor/2025/12/19/14090/"
)

func init() {
	adapter.ProvideWriter("jpk_v7m", WriteJPKV7M)
}

type jpk struct {
	XMLName    xml.Name    `xml:"JPK"`
	XMLNS      string      `xml:"xmlns,attr"`
	Naglowek   jpkHeader   `xml:"Naglowek"`
	Podmiot1   jpkTaxpayer `xml:"Podmiot1"`
	Deklaracja *jpkReturn  `xml:"Deklaracja,omitempty"`
	Ewidencja  jpkRecords  `xml:"Ewidencja"`
}

type jpkHeader struct {
	KodFormularza      jpkFormCode `xml:"KodFormularza"`
	WariantFormularza  int         `xml:"WariantFormularza"`
	DataWytworzeniaJPK string      `xml:"DataWytworzeniaJPK"`
	NazwaSystemu       string      `xml:"NazwaSystemu"`
	CelZlozenia        jpkPurpose  `xml:"CelZlozenia"`
	KodUrzedu          string      `xml:"KodUrzedu"`
	Rok                int         `xml:"Rok"`
	Miesiac            int         `xml:"Miesiac"`
}

type jpkFormCode struct {
	KodSystemowy string `xml:"kodSystemowy,attr"`
	WersjaSchemy string `xml:"wersjaSchemy,attr"`
	Value        string `xml:",chardata"`
}

type jpkPurpose struct {
	Poz   string `xml:"poz,attr"`
	Value int    `xml:",chardata"`
}

type jpkTaxpayer struct {
	Rola             string            `xml:"rola,attr"`
	OsobaFizyczna    *jpkNaturalPerson `xml:"OsobaFizyczna,omitempty"`
	OsobaNiefizyczna *jpkLegalPerson   `xml:"OsobaNiefizyczna,omitempty"`
}

type jpkNaturalPerson struct {
	NIP           string `xml:"NIP"`
	ImiePierwsze  string `xml:"ImiePierwsze"`
	Nazwisko      string `xml:"Nazwisko"`
	DataUrodzenia string `xml:"DataUrodzenia"`
	Email         string `xml:"Email"`
}

type jpkLegalPerson struct {
	NIP        string `xml:"NIP"`
	PelnaNazwa string `xml:"PelnaNazwa"`
	Email      string `xml:"Email"`
}

type jpkReturn struct {
	Naglowek           jpkReturnHeader `xml:"Naglowek"`
	PozycjeSzczegolowe jpkPositions    `xml:"PozycjeSzczegolowe"`
	Pouczenia          int             `xml:"Pouczenia"`
}

type jpkReturnHeader struct {
	KodFormularzaDekl     jpkReturnCode `xml:"KodFormularzaDekl"`
	WariantFormularzaDekl int           `xml:"WariantFormularzaDekl"`
}

type jpkReturnCode struct {
	KodSystemowy       string `xml:"kodSystemowy,attr"`
	KodPodatku         string `xml:"kodPodatku,attr"`
	RodzajZobowiazania string `xml:"rodzajZobowiazania,attr"`
	WersjaSchemy       string `xml:"wersjaSchemy,attr"`
	Value              string `xml:",chardata"`
}

// jpkPositions: the declaration in whole złoty; an optional position left
// empty is omitted, as the brochure says ("w przypadku braku — pole
// pozostaje puste"). The schema's pairs (base and VAT of one rate) stand or
// fall together.
type jpkPositions struct {
	P10 *int64 `xml:"P_10,omitempty"` // exempt, base
	P13 *int64 `xml:"P_13,omitempty"` // 0% domestic, base
	P15 *int64 `xml:"P_15,omitempty"` // 5% base
	P16 *int64 `xml:"P_16,omitempty"` // 5% VAT
	P17 *int64 `xml:"P_17,omitempty"` // 8% base
	P18 *int64 `xml:"P_18,omitempty"` // 8% VAT
	P19 *int64 `xml:"P_19,omitempty"` // 23% base
	P20 *int64 `xml:"P_20,omitempty"` // 23% VAT
	P37 *int64 `xml:"P_37,omitempty"` // bases, total
	P38 int64  `xml:"P_38"`           // output VAT, total
	P39 *int64 `xml:"P_39,omitempty"` // excess carried in
	P42 *int64 `xml:"P_42,omitempty"` // other purchases, net
	P43 *int64 `xml:"P_43,omitempty"` // other purchases, input VAT
	P48 *int64 `xml:"P_48,omitempty"` // input VAT, total
	P51 int64  `xml:"P_51"`           // to pay
	P53 *int64 `xml:"P_53,omitempty"` // excess of input over output
	P62 *int64 `xml:"P_62,omitempty"` // carried forward
}

type jpkRecords struct {
	SprzedazWiersz []jpkSale       `xml:"SprzedazWiersz"`
	SprzedazCtrl   jpkSaleCtrl     `xml:"SprzedazCtrl"`
	ZakupWiersz    []jpkPurchase   `xml:"ZakupWiersz"`
	ZakupCtrl      jpkPurchaseCtrl `xml:"ZakupCtrl"`
}

type jpkSale struct {
	LpSprzedazy      int    `xml:"LpSprzedazy"`
	NrKontrahenta    string `xml:"NrKontrahenta"`
	NazwaKontrahenta string `xml:"NazwaKontrahenta"`
	DowodSprzedazy   string `xml:"DowodSprzedazy"`
	DataWystawienia  string `xml:"DataWystawienia"`
	DataSprzedazy    string `xml:"DataSprzedazy,omitempty"` // the sale date, when not the issue date
	NrKSeF           string `xml:"NrKSeF,omitempty"`
	BFK              string `xml:"BFK,omitempty"`
	K10              string `xml:"K_10,omitempty"`
	K13              string `xml:"K_13,omitempty"`
	K15              string `xml:"K_15,omitempty"`
	K16              string `xml:"K_16,omitempty"`
	K17              string `xml:"K_17,omitempty"`
	K18              string `xml:"K_18,omitempty"`
	K19              string `xml:"K_19,omitempty"`
	K20              string `xml:"K_20,omitempty"`
}

type jpkSaleCtrl struct {
	LiczbaWierszySprzedazy int    `xml:"LiczbaWierszySprzedazy"`
	PodatekNalezny         string `xml:"PodatekNalezny"`
}

type jpkPurchase struct {
	LpZakupu      int    `xml:"LpZakupu"`
	NrDostawcy    string `xml:"NrDostawcy"`
	NazwaDostawcy string `xml:"NazwaDostawcy"`
	DowodZakupu   string `xml:"DowodZakupu"`
	DataZakupu    string `xml:"DataZakupu"`           // the purchase document's issue date
	DataWplywu    string `xml:"DataWplywu,omitempty"` // its receipt, when not that day
	NrKSeF        string `xml:"NrKSeF,omitempty"`
	BFK           string `xml:"BFK,omitempty"`
	K42           string `xml:"K_42,omitempty"`
	K43           string `xml:"K_43,omitempty"`
}

type jpkPurchaseCtrl struct {
	LiczbaWierszyZakupow int    `xml:"LiczbaWierszyZakupow"`
	PodatekNaliczony     string `xml:"PodatekNaliczony"`
}

// WriteJPKV7M renders the month named by params["period"] ("2026-09"). The
// taxpayer is the company of kind self; a natural person needs its owner
// (a person with given and family names and a birth date), a legal person
// its full name; both need the tax office code and an e-mail. Without a
// settlement for the month there is no declaration, only the records —
// the file says what the log can derive and nothing more.
func WriteJPKV7M(ctx context.Context, reads adapter.Reads, params map[string]string) (adapter.File, error) {
	period := params["period"]
	if len(period) != 7 || period[4] != '-' {
		return adapter.File{}, fmt.Errorf("jpk_v7m wants period=YYYY-MM, got %q", period)
	}
	year, _ := strconv.Atoi(period[:4])
	month, _ := strconv.Atoi(period[5:])
	if year == 0 || month < 1 || month > 12 {
		return adapter.File{}, fmt.Errorf("jpk_v7m: %q is not a month", period)
	}
	companies, err := byID(ctx, reads, "company")
	if err != nil {
		return adapter.File{}, err
	}
	persons, err := byID(ctx, reads, "person")
	if err != nil {
		return adapter.File{}, err
	}
	rates, err := byID(ctx, reads, "vat_rate")
	if err != nil {
		return adapter.File{}, err
	}
	var self map[string]any
	selves := 0
	for _, c := range companies {
		if c["kind"] == "self" {
			self = c
			selves++
		}
	}
	if self == nil {
		return adapter.File{}, fmt.Errorf("jpk_v7m: no company of kind self — whose return is this?")
	}
	if selves > 1 {
		return adapter.File{}, fmt.Errorf("jpk_v7m: %d companies of kind self — whose return is this? one taxpayer per file", selves)
	}
	if self["vat_status"] == "exempt" {
		return adapter.File{}, fmt.Errorf("jpk_v7m: %s is exempt from VAT — an exempt company files no JPK_V7", text(self["name"]))
	}
	taxpayer, err := jpkTaxpayerOf(self, persons)
	if err != nil {
		return adapter.File{}, err
	}
	// The writer contract: the same state and parameters give the same
	// bytes, so the creation time is a parameter, else the period's last
	// instant — never the wall clock, or every rendering would read as a
	// correction.
	created := params["created"]
	if created == "" {
		created = periodEnd(period) + "T23:59:59Z"
	}
	doc := jpk{
		XMLNS: JPKNamespace,
		Naglowek: jpkHeader{
			KodFormularza:     jpkFormCode{KodSystemowy: "JPK_V7M (3)", WersjaSchemy: "1-0E", Value: "JPK_VAT"},
			WariantFormularza: 3, DataWytworzeniaJPK: created, NazwaSystemu: "Rhea",
			CelZlozenia: jpkPurpose{Poz: "P_7", Value: 1},
			KodUrzedu:   fmt.Sprint(self["tax_office"]), Rok: year, Miesiac: month,
		},
		Podmiot1: taxpayer,
	}
	// The records: the month's invoices, in log order.
	sales, err := reads.ObjectsByType(ctx, "sales_invoice")
	if err != nil {
		return adapter.File{}, err
	}
	subs, err := reads.ObjectsByType(ctx, "ksef_submission")
	if err != nil {
		return adapter.File{}, err
	}
	ksefOf := map[string]string{}
	for _, s := range subs {
		if s.State["status"] == "accepted" {
			inv, _ := s.State["invoice"].(string)
			ref, _ := s.State["ksef_ref"].(string)
			ksefOf[inv] = ref
		}
	}
	var outputVAT int64
	for _, inv := range sales {
		if inv.State["vat_period"] != period {
			continue
		}
		buyerID, _ := inv.State["buyer"].(string)
		buyer := companies[buyerID]
		row := jpkSale{
			LpSprzedazy:      len(doc.Ewidencja.SprzedazWiersz) + 1,
			NrKontrahenta:    nipOf(buyer, inv.State["buyer_nip"]),
			NazwaKontrahenta: text(buyer["name"]),
			DowodSprzedazy:   text(inv.State["number"]),
			DataWystawienia:  text(inv.State["issue_date"]),
		}
		if sale := text(inv.State["sale_date"]); sale != "" && sale != row.DataWystawienia {
			row.DataSprzedazy = sale
		}
		if ref := ksefOf[inv.ID]; ref != "" {
			row.NrKSeF = ref
		} else {
			row.BFK = "1" // issued outside KSeF
		}
		net, vat := plnAmounts(inv.State)
		code := text(rates[text(inv.State["vat_rate"])]["code"])
		switch code {
		case "23", "22":
			row.K19, row.K20 = core.FormatMoney(net), core.FormatMoney(vat)
		case "8", "7":
			row.K17, row.K18 = core.FormatMoney(net), core.FormatMoney(vat)
		case "5":
			row.K15, row.K16 = core.FormatMoney(net), core.FormatMoney(vat)
		case "0":
			row.K13 = core.FormatMoney(net)
		case "zw":
			row.K10 = core.FormatMoney(net)
		default:
			return adapter.File{}, fmt.Errorf("jpk_v7m: invoice %s carries VAT rate %q, which has no column in the records", row.DowodSprzedazy, code)
		}
		outputVAT += vat
		doc.Ewidencja.SprzedazWiersz = append(doc.Ewidencja.SprzedazWiersz, row)
	}
	doc.Ewidencja.SprzedazCtrl = jpkSaleCtrl{LiczbaWierszySprzedazy: len(doc.Ewidencja.SprzedazWiersz), PodatekNalezny: core.FormatMoney(outputVAT)}
	purchases, err := reads.ObjectsByType(ctx, "purchase_invoice")
	if err != nil {
		return adapter.File{}, err
	}
	var inputVAT int64
	for _, inv := range purchases {
		if inv.State["vat_period"] != period {
			continue
		}
		supplierID, _ := inv.State["supplier"].(string)
		supplier := companies[supplierID]
		net, vat := plnAmounts(inv.State)
		row := jpkPurchase{
			LpZakupu:      len(doc.Ewidencja.ZakupWiersz) + 1,
			NrDostawcy:    nipOf(supplier, nil),
			NazwaDostawcy: text(supplier["name"]),
			DowodZakupu:   text(inv.State["number"]),
			DataZakupu:    text(inv.State["invoice_date"]),
			K42:           core.FormatMoney(net), K43: core.FormatMoney(vat),
		}
		if got := text(inv.State["received_date"]); got != "" && got != row.DataZakupu {
			row.DataWplywu = got
		}
		if ref := text(inv.State["ksef_ref"]); ref != "" {
			row.NrKSeF = ref
		} else {
			row.BFK = "1"
		}
		inputVAT += vat
		doc.Ewidencja.ZakupWiersz = append(doc.Ewidencja.ZakupWiersz, row)
	}
	doc.Ewidencja.ZakupCtrl = jpkPurchaseCtrl{LiczbaWierszyZakupow: len(doc.Ewidencja.ZakupWiersz), PodatekNaliczony: core.FormatMoney(inputVAT)}

	// The declaration: the settlement the close derived, in whole złoty.
	settlements, err := reads.ObjectsByType(ctx, "vat_settlement")
	if err != nil {
		return adapter.File{}, err
	}
	for _, st := range settlements {
		if st.State["period"] != period {
			continue
		}
		// The original is the month as the close settled it. A document
		// booked late into a settled month — its entries in an open month,
		// its VAT in this one — makes the records disagree with the
		// declaration: what is owed then is a correction, not this file.
		settledOut, _ := minorUnits(st.State["output_vat"])
		settledIn, _ := minorUnits(st.State["purchase_vat"])
		if settledOut != outputVAT || settledIn != inputVAT {
			return adapter.File{}, fmt.Errorf("jpk_v7m: %s changed after its close — the records now say VAT %s due and %s to deduct, the settlement %s and %s; a late document owes a correction of the return (CelZlozenia 2), which Rhea does not write yet",
				period, core.FormatMoney(outputVAT), core.FormatMoney(inputVAT), core.FormatMoney(settledOut), core.FormatMoney(settledIn))
		}
		p := jpkPositions{}
		whole := func(field string) int64 {
			m, _ := minorUnits(st.State[field])
			return wholeZloty(m)
		}
		opt := func(field string) *int64 {
			if v := whole(field); v != 0 {
				return &v
			}
			return nil
		}
		pair := func(base, vat string) (*int64, *int64) {
			b, v := whole(base), whole(vat)
			if b == 0 && v == 0 {
				return nil, nil
			}
			return &b, &v
		}
		p.P10, p.P13 = opt("sales_net_exempt"), opt("sales_net_0")
		p.P15, p.P16 = pair("sales_net_5", "sales_vat_5")
		p.P17, p.P18 = pair("sales_net_8", "sales_vat_8")
		p.P19, p.P20 = pair("sales_net_23", "sales_vat_23")
		// The totals are sums of the rounded positions, as the Ministry
		// checks them: rounding the total on its own gave 111 where 101
		// and 11 rounded make 112, and the file failed by one złoty.
		bases := whole("sales_net_exempt") + whole("sales_net_0") + whole("sales_net_5") + whole("sales_net_8") + whole("sales_net_23")
		if bases != 0 {
			p.P37 = &bases
		}
		p.P38 = whole("sales_vat_5") + whole("sales_vat_8") + whole("sales_vat_23")
		p.P39 = opt("carried_in")
		p.P42, p.P43 = pair("purchase_net", "purchase_vat")
		input := whole("carried_in") + whole("purchase_vat")
		if input != 0 {
			p.P48 = &input
		}
		if due := p.P38 - input; due > 0 {
			p.P51 = due
		} else if carry := input - p.P38; carry > 0 {
			p.P53, p.P62 = &carry, &carry
		}
		doc.Deklaracja = &jpkReturn{
			Naglowek: jpkReturnHeader{KodFormularzaDekl: jpkReturnCode{KodSystemowy: "VAT-7 (23)", KodPodatku: "VAT", RodzajZobowiazania: "Z", WersjaSchemy: "1-0E", Value: "VAT-7"},
				WariantFormularzaDekl: 23},
			PozycjeSzczegolowe: p, Pouczenia: 1,
		}
	}
	// An original JPK_V7M is the records and the declaration together; the
	// declaration is the close's, so a month not yet closed has none to file.
	if doc.Deklaracja == nil {
		return adapter.File{}, fmt.Errorf("jpk_v7m: %s has no VAT settlement — close the month first; the declaration is what the close derives", period)
	}
	body, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return adapter.File{}, err
	}
	name := fmt.Sprintf("JPK_V7M_%s_%s.xml", nipOf(self, nil), strings.ReplaceAll(period, "-", ""))
	return adapter.File{Name: name, MediaType: "application/xml", Bytes: append([]byte(xml.Header), append(body, '\n')...)}, nil
}

// jpkTaxpayerOf reads the taxpayer block from the company of kind self.
func jpkTaxpayerOf(self map[string]any, persons map[string]map[string]any) (jpkTaxpayer, error) {
	nip := nipOf(self, nil)
	email := text(self["email"])
	office := text(self["tax_office"])
	if nip == "" || email == "" || office == "" {
		return jpkTaxpayer{}, fmt.Errorf("jpk_v7m: the company needs a NIP, an e-mail and a tax office code (vat_id, email, tax_office)")
	}
	t := jpkTaxpayer{Rola: "Podatnik"}
	if self["legal_form"] == "natural_person" {
		owner := persons[text(self["owner"])]
		given, family, born := text(owner["given_name"]), text(owner["family_name"]), text(owner["birth_date"])
		if given == "" || family == "" || born == "" {
			return jpkTaxpayer{}, fmt.Errorf("jpk_v7m: a natural person's return needs the owner's given name, family name and birth date (company.owner → person)")
		}
		t.OsobaFizyczna = &jpkNaturalPerson{NIP: nip, ImiePierwsze: given, Nazwisko: family, DataUrodzenia: born, Email: email}
		return t, nil
	}
	t.OsobaNiefizyczna = &jpkLegalPerson{NIP: nip, PelnaNazwa: text(self["name"]), Email: email}
	return t, nil
}

// wholeZloty rounds minor units to whole złoty the way the tax ordinance
// does (art. 63 §1): below 50 groszy down, 50 and above up.
func wholeZloty(minor int64) int64 {
	if minor < 0 {
		return -wholeZloty(-minor)
	}
	return (minor + 50) / 100
}

func nipOf(c map[string]any, fallback any) string {
	if v := text(c["vat_id"]); v != "" {
		return NIPDigits(v)
	}
	return NIPDigits(text(fallback))
}

// plnAmounts are an invoice's net and VAT in złoty: the amounts it carries
// in złoty (an invoice in another currency, converted at its booking), or its
// own when it was issued in złoty or before it carried them.
func plnAmounts(state map[string]any) (net, vat int64) {
	net, okN := minorUnits(state["net_pln"])
	vat, okV := minorUnits(state["vat_pln"])
	if !okN || !okV {
		net, _ = minorUnits(state["net"])
		vat, _ = minorUnits(state["vat"])
	}
	return net, vat
}

func text(v any) string {
	s, _ := v.(string)
	return s
}

// periodEnd is the last day of a YYYY-MM period.
func periodEnd(period string) string {
	t, err := time.Parse("2006-01", period)
	if err != nil {
		return period + "-01"
	}
	return t.AddDate(0, 1, -1).Format("2006-01-02")
}
