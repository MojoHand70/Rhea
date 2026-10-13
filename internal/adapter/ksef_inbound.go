package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"rhea/internal/core"
)

// KSeFInbound is the reader of the KSeF channel (DIRECTION 2026-10-10,
// "Documents in", step 4): every received document whose bytes are a
// Faktura becomes one purchase.invoice.received naming the file by its
// hash, and one document.read that moves the document on. The reading is
// exact — the numbers are the document's own, and a document that
// contradicts itself (net plus VAT is not the gross) is not read but
// reported, as document.unreadable, for a person. The pull from the KSeF
// API joins this reader when a client exists; until then files reach the
// store through the door.
type KSeFInbound struct {
	Today func() string // the business date of a refusal; defaults to the wall clock
}

const (
	ksefInboundCodeVersion = 1

	// ReceivedDocumentType is the documents pack's conventional vocabulary
	// (packs/common/documents.json) the readers scan — a type the pack
	// ships, like task_schedule for the clock; the kernel does not know it.
	ReceivedDocumentType = "received_document"

	// EventDocumentRead says a channel turned a received document into
	// typed events; the documents pack moves the document to "read" on it.
	EventDocumentRead = "document.read"
	// EventDocumentUnreadable says a channel could not: the reason is for a
	// person, so the event waits in the worklist.
	EventDocumentUnreadable = "document.unreadable"
)

func init() {
	Register(Implemented{Name: "ksef_inbound", CodeVersion: ksefInboundCodeVersion,
		Format: "FA(3) XML, the subset a purchase invoice needs, from the document store"})
	readers["ksef_inbound"] = func() Adapter { return KSeFInbound{} }
}

func (KSeFInbound) Name() string     { return "ksef_inbound" }
func (KSeFInbound) CodeVersion() int { return ksefInboundCodeVersion }

func (k KSeFInbound) Pass(ctx context.Context, reads Reads) ([]core.Event, error) {
	docs, err := reads.ObjectsByType(ctx, ReceivedDocumentType)
	if err != nil {
		return nil, err
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].ID < docs[j].ID })
	var out []core.Event
	for _, d := range docs {
		status, _ := d.State["status"].(string)
		media, _ := d.State["media_type"].(string)
		source, _ := d.State["source"].(string)
		hash, _ := d.State["hash"].(string)
		name, _ := d.State["name"].(string)
		if status != "received" || hash == "" {
			continue
		}
		if !strings.Contains(media, "xml") && source != "ksef" {
			continue
		}
		b, err := reads.GetBlob(ctx, hash)
		if err != nil {
			return out, fmt.Errorf("%s names file %s: %w", d.ID, hash, err)
		}
		if !IsFA(b.Bytes) {
			continue // an XML that is not an invoice is someone else's
		}
		inv, err := ParseFA(b.Bytes)
		if err != nil {
			payload, _ := json.Marshal(map[string]any{
				"document": hash, "name": name, "channel": k.Name(), "reason": err.Error(),
			})
			out = append(out, core.Event{Type: EventDocumentUnreadable, OccurredAt: k.today(),
				Payload: payload, DedupKey: "ksef_inbound/unreadable/" + hash})
			continue
		}
		fact := map[string]any{
			"number": inv.Number, "supplier": inv.SellerName, "supplier_nip": inv.SellerCountry + inv.SellerNIP,
			"invoice_date": inv.IssueDate, "net": core.FormatMoney(inv.Net), "vat": core.FormatMoney(inv.VAT),
			"gross": core.FormatMoney(inv.Gross), "currency": inv.Currency, "document": hash,
			"market": "pl",
		}
		if inv.BuyerNIP != "" {
			fact["buyer_nip"] = inv.BuyerCountry + inv.BuyerNIP
		}
		if inv.RateCode != "" {
			fact["vat_rate"] = inv.RateCode
		}
		if len(inv.WZ) > 0 {
			fact["for_document"] = inv.WZ[0]
		}
		if inv.Term != "" {
			fact["payment_term"] = inv.Term
		}
		if inv.Account != "" {
			fact["supplier_account"] = inv.Account
		}
		if ref := strings.TrimSuffix(name, ".xml"); ksefRefPattern.MatchString(ref) {
			fact["ksef_ref"] = ref
		}
		payload, err := json.Marshal(fact)
		if err != nil {
			return out, err
		}
		read, _ := json.Marshal(map[string]any{
			"document": hash, "channel": k.Name(), "event_type": "purchase.invoice.received",
		})
		out = append(out,
			core.Event{Type: "purchase.invoice.received", OccurredAt: inv.IssueDate,
				Payload: payload, DedupKey: "ksef_inbound/" + hash},
			core.Event{Type: EventDocumentRead, OccurredAt: inv.IssueDate,
				Payload: read, DedupKey: "ksef_inbound/read/" + hash})
	}
	return out, nil
}

func (k KSeFInbound) today() string {
	if k.Today != nil {
		return k.Today()
	}
	return time.Now().UTC().Format("2006-01-02")
}

var ksefRefPattern = regexp.MustCompile(`^\d{10}-\d{8}-[0-9A-F]{12}-[0-9A-F]{2}$`)

// Invoice is what the reader takes from a Faktura.
type Invoice struct {
	Number, IssueDate                    string
	SellerNIP, SellerName, SellerCountry string
	BuyerNIP, BuyerCountry               string // the country defaults to PL, as the seller's does
	Currency                             string
	Net, VAT, Gross                      int64
	RateCode                             string // the one VAT rate, or "" when the invoice mixes rates
	WZ                                   []string
	Term, Account                        string
	Lines                                []InvoiceLine
}

type InvoiceLine struct {
	Name, Unit, Qty, UnitPrice string
	Net                        int64
	Rate                       string
}

// fa is the FA subset as encoding/xml reads it: the slots net and VAT land
// in by rate (P_13_k, P_14_k), the parties, the lines, the payment.
type fa struct {
	XMLName xml.Name `xml:"Faktura"`
	Seller  faSide   `xml:"Podmiot1"`
	Buyer   faSide   `xml:"Podmiot2"`
	Body    struct {
		Currency  string   `xml:"KodWaluty"`
		IssueDate string   `xml:"P_1"`
		Number    string   `xml:"P_2"`
		WZ        []string `xml:"WZ"`
		Net1      string   `xml:"P_13_1"`
		Net2      string   `xml:"P_13_2"`
		Net3      string   `xml:"P_13_3"`
		Net4      string   `xml:"P_13_4"`
		Net5      string   `xml:"P_13_5"`
		Net6Home  string   `xml:"P_13_6_1"` // 0%: domestic
		Net6EU    string   `xml:"P_13_6_2"` // 0%: intra-EU supply
		Net6Exp   string   `xml:"P_13_6_3"` // 0%: export
		Net7      string   `xml:"P_13_7"`   // exempt
		VAT1      string   `xml:"P_14_1"`
		VAT2      string   `xml:"P_14_2"`
		VAT3      string   `xml:"P_14_3"`
		VAT4      string   `xml:"P_14_4"`
		VAT5      string   `xml:"P_14_5"`
		Gross     string   `xml:"P_15"`
		Lines     []struct {
			Name      string `xml:"P_7"`
			Unit      string `xml:"P_8A"`
			Qty       string `xml:"P_8B"`
			UnitPrice string `xml:"P_9A"`
			Net       string `xml:"P_11"`
			Rate      string `xml:"P_12"`
		} `xml:"FaWiersz"`
		Term    string `xml:"Platnosc>TerminPlatnosci>Termin"`
		Account string `xml:"Platnosc>RachunekBankowy>NrRB"`
	} `xml:"Fa"`
}

type faSide struct {
	NIP     string `xml:"DaneIdentyfikacyjne>NIP"`
	Name    string `xml:"DaneIdentyfikacyjne>Nazwa"`
	Country string `xml:"Adres>KodKraju"`
}

// IsFA reports whether bytes are a KSeF invoice: an XML document whose root
// element is Faktura.
func IsFA(b []byte) bool {
	dec := xml.NewDecoder(bytes.NewReader(b))
	for {
		tok, err := dec.Token()
		if err != nil {
			return false
		}
		if start, ok := tok.(xml.StartElement); ok {
			return start.Name.Local == "Faktura"
		}
	}
}

// ParseFA decodes a Faktura and checks that it agrees with itself: the net
// slots and the VAT slots add up to the gross, and the lines add up to the
// net. A reader reports what the document says; a document that lies is
// refused with the reason.
func ParseFA(b []byte) (Invoice, error) {
	var f fa
	if err := xml.Unmarshal(b, &f); err != nil {
		return Invoice{}, fmt.Errorf("not a readable Faktura: %w", err)
	}
	inv := Invoice{
		Number: strings.TrimSpace(f.Body.Number), IssueDate: strings.TrimSpace(f.Body.IssueDate),
		SellerNIP: strings.TrimSpace(f.Seller.NIP), SellerName: strings.TrimSpace(f.Seller.Name),
		SellerCountry: strings.TrimSpace(f.Seller.Country), BuyerNIP: strings.TrimSpace(f.Buyer.NIP),
		BuyerCountry: strings.TrimSpace(f.Buyer.Country),
		Currency:     strings.TrimSpace(f.Body.Currency), WZ: f.Body.WZ,
		Term: strings.TrimSpace(f.Body.Term), Account: strings.TrimSpace(f.Body.Account),
	}
	switch {
	case inv.Number == "":
		return inv, fmt.Errorf("no invoice number (P_2)")
	case inv.SellerNIP == "":
		return inv, fmt.Errorf("no seller NIP")
	case inv.Currency == "":
		return inv, fmt.Errorf("no currency (KodWaluty)")
	}
	if _, err := time.Parse("2006-01-02", inv.IssueDate); err != nil {
		return inv, fmt.Errorf("issue date (P_1) %q is not a date", inv.IssueDate)
	}
	if inv.SellerCountry == "" {
		inv.SellerCountry = "PL"
	}
	if inv.BuyerCountry == "" {
		inv.BuyerCountry = "PL"
	}
	slots := []struct{ net, vat, code string }{
		{f.Body.Net1, f.Body.VAT1, "23"}, {f.Body.Net2, f.Body.VAT2, "8"}, {f.Body.Net3, f.Body.VAT3, "5"},
		{f.Body.Net4, f.Body.VAT4, "0"}, {f.Body.Net5, f.Body.VAT5, "0"}, {f.Body.Net6Home, "", "0"}, {f.Body.Net6EU, "", "0"}, {f.Body.Net6Exp, "", "0"}, {f.Body.Net7, "", "zw"},
	}
	used := 0
	for _, s := range slots {
		if strings.TrimSpace(s.net) == "" {
			continue
		}
		used++
		net, err := core.ParseMoney(strings.TrimSpace(s.net))
		if err != nil {
			return inv, fmt.Errorf("net %q: %w", s.net, err)
		}
		inv.Net += net
		inv.RateCode = s.code
		if strings.TrimSpace(s.vat) != "" {
			vat, err := core.ParseMoney(strings.TrimSpace(s.vat))
			if err != nil {
				return inv, fmt.Errorf("VAT %q: %w", s.vat, err)
			}
			inv.VAT += vat
		}
	}
	if used == 0 {
		return inv, fmt.Errorf("no net amount in any rate slot (P_13_k)")
	}
	if used > 1 {
		inv.RateCode = ""
	}
	gross, err := core.ParseMoney(strings.TrimSpace(f.Body.Gross))
	if err != nil {
		return inv, fmt.Errorf("gross (P_15) %q: %w", f.Body.Gross, err)
	}
	inv.Gross = gross
	if inv.Net+inv.VAT != inv.Gross {
		return inv, fmt.Errorf("the document contradicts itself: net %s + VAT %s is not the gross %s",
			core.FormatMoney(inv.Net), core.FormatMoney(inv.VAT), core.FormatMoney(inv.Gross))
	}
	var linesNet int64
	rates := map[string]bool{}
	for _, l := range f.Body.Lines {
		net, err := core.ParseMoney(strings.TrimSpace(l.Net))
		if err != nil {
			return inv, fmt.Errorf("line %q net %q: %w", l.Name, l.Net, err)
		}
		linesNet += net
		rate := strings.TrimSpace(l.Rate)
		if rate != "" {
			rates[rate] = true
		}
		inv.Lines = append(inv.Lines, InvoiceLine{Name: l.Name, Unit: l.Unit, Qty: l.Qty, UnitPrice: l.UnitPrice, Net: net, Rate: rate})
	}
	if len(inv.Lines) > 0 && linesNet != inv.Net {
		return inv, fmt.Errorf("the lines add up to %s, the net says %s", core.FormatMoney(linesNet), core.FormatMoney(inv.Net))
	}
	if len(rates) == 1 {
		for r := range rates {
			inv.RateCode = r
		}
	}
	return inv, nil
}
