// Package sim is the simulated customer (DIRECTION, the growth plan: the
// whole demo story as one unattended eval run). A persona is a business as
// data — its parties, catalogue, chart, rates, volumes, and the answers its
// owner gives in the implementation interview, each with a known-good
// reference. A deterministic generator turns the persona into a year of
// documents and the state a correct implementation ends in, independently
// of any rule: counts, lifecycle moves, sums of computed fields. The model
// plays no part in the facts; it only paraphrases the interview answers
// (paraphrase.go), so phrasing stays human while ground truth stays exact.
// Compile turns all of it into the corpus the eval already runs.
//
// Synthetic installations never count in the real network: their
// publications are marked and excluded from every figure a real client sees
// (network.SyntheticPrefix).
package sim

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"rhea/internal/core"
	"rhea/internal/eval"
	"rhea/internal/network"
)

// Persona is a simulated customer as data.
type Persona struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Seeds       []string `json:"seeds"`           // definition files, relative to the persona file
	Packs       []string `json:"packs,omitempty"` // pack files, relative to the persona file: the documents pack, the market's readers
	Seed        uint64   `json:"seed"`            // the generator's randomness, fixed
	From        string   `json:"from"`            // first month, YYYY-MM
	Months      int      `json:"months"`

	Company   Party     `json:"company"`
	Customers []Party   `json:"customers"`
	Suppliers []Party   `json:"suppliers"`
	Items     []Item    `json:"items"`
	Locations []Place   `json:"locations,omitempty"`
	Accounts  []Account `json:"accounts"`
	VATRates  []VATRate `json:"vat_rates"`
	Volume    Volume    `json:"volume"`
	// Postings is the owner's booking policy as a count: how many posting
	// lines one document of each kind books — what the interview answers
	// promise, stated once so the generator can expect it.
	Postings map[string]int `json:"postings"`
	Book     string         `json:"book"` // the statutory book the postings land in

	Interview []Answer `json:"interview"`
}

type Party struct {
	Name    string `json:"name"`
	VATID   string `json:"vat_id"`
	Country string `json:"country"`
	IBAN    string `json:"iban,omitempty"` // the account statements and invoices name; derived when absent
}

// Item is a catalogue line: goods with a purchase cost and stock, or a
// service (no cost, never stocked).
type Item struct {
	SKU   string `json:"sku"`
	Name  string `json:"name"`
	Unit  string `json:"unit"`
	Price string `json:"price"`          // selling price, decimal
	Cost  string `json:"cost,omitempty"` // purchase price on deliveries; empty for services
}

type Place struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type Account struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type VATRate struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	Percent int    `json:"percent"`
}

// Volume is how much happens per month, and when the warehouse is counted.
type Volume struct {
	SalesInvoices    int      `json:"sales_invoices"`
	Deliveries       int      `json:"deliveries"`
	PurchaseInvoices int      `json:"purchase_invoices"` // for deliveries, or costs when there are none
	Issues           int      `json:"issues"`
	Complaints       int      `json:"complaints"`
	PaidShare        int      `json:"paid_share"`   // percent of sales invoices paid within the period
	PaymentDays      int      `json:"payment_days"` // the term printed on sales invoices
	StockCounts      []string `json:"stock_counts,omitempty"`
}

// Answer is one thing the owner says in the interview, with the reference
// the eval's recorded agent answers with. The model never sees the
// reference; it may rephrase the answer (paraphrase.go).
type Answer struct {
	Answer     string          `json:"answer"`
	Mode       string          `json:"mode,omitempty"`
	SampleType string          `json:"sample_type,omitempty"`
	Reference  json.RawMessage `json:"reference"`
}

// Voices are a persona's interview answers rephrased by the model, kept as
// a file beside the persona so runs are reproducible: Answers[i] holds the
// variants of interview answer i.
type Voices struct {
	Model   string     `json:"model"`
	Answers [][]string `json:"answers"`
}

// Load reads a persona file.
func Load(path string) (Persona, string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Persona{}, "", err
	}
	var p Persona
	if err := json.Unmarshal(b, &p); err != nil {
		return Persona{}, "", fmt.Errorf("%s: %w", path, err)
	}
	if err := p.Validate(); err != nil {
		return Persona{}, "", fmt.Errorf("%s: %w", path, err)
	}
	return p, filepath.Dir(path), nil
}

// IsPersona reports whether a file is a persona rather than a corpus.
func IsPersona(path string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var probe struct {
		Interview []json.RawMessage `json:"interview"`
	}
	return json.Unmarshal(b, &probe) == nil && len(probe.Interview) > 0
}

// VoicesPath is where a persona's paraphrases live.
func VoicesPath(personaPath string) string {
	return strings.TrimSuffix(personaPath, ".json") + ".voices.json"
}

// LoadVoices reads the paraphrase file if there is one.
func LoadVoices(personaPath string) (Voices, bool, error) {
	b, err := os.ReadFile(VoicesPath(personaPath))
	if os.IsNotExist(err) {
		return Voices{}, false, nil
	}
	if err != nil {
		return Voices{}, false, err
	}
	var v Voices
	if err := json.Unmarshal(b, &v); err != nil {
		return Voices{}, false, fmt.Errorf("%s: %w", VoicesPath(personaPath), err)
	}
	return v, true, nil
}

func (p Persona) Validate() error {
	switch {
	case p.Name == "":
		return fmt.Errorf("persona needs a name")
	case strings.ContainsAny(p.Name, " :/"):
		return fmt.Errorf("persona name %q: letters, digits and dashes only", p.Name)
	case p.Months <= 0:
		return fmt.Errorf("persona %s: months must be positive", p.Name)
	case len(p.Interview) == 0:
		return fmt.Errorf("persona %s: no interview answers", p.Name)
	case p.Company.Name == "" || p.Company.VATID == "":
		return fmt.Errorf("persona %s: the company needs a name and a vat_id", p.Name)
	case len(p.VATRates) == 0:
		return fmt.Errorf("persona %s: no VAT rates", p.Name)
	}
	if _, err := time.Parse("2006-01", p.From); err != nil {
		return fmt.Errorf("persona %s: from must be YYYY-MM", p.Name)
	}
	for _, it := range p.Items {
		if _, err := core.ParseMoney(it.Price); err != nil {
			return fmt.Errorf("persona %s: item %s price: %w", p.Name, it.SKU, err)
		}
		if it.Cost != "" {
			if _, err := core.ParseMoney(it.Cost); err != nil {
				return fmt.Errorf("persona %s: item %s cost: %w", p.Name, it.SKU, err)
			}
		}
	}
	if p.Volume.Deliveries > 0 && (len(p.Locations) == 0 || len(p.Suppliers) == 0) {
		return fmt.Errorf("persona %s: deliveries need locations and suppliers", p.Name)
	}
	if p.Volume.SalesInvoices > 0 && (len(p.Customers) == 0 || len(p.Items) == 0) {
		return fmt.Errorf("persona %s: sales need customers and items", p.Name)
	}
	for i, a := range p.Interview {
		if a.Answer == "" || len(a.Reference) == 0 {
			return fmt.Errorf("persona %s: interview answer %d needs an answer and a reference", p.Name, i+1)
		}
	}
	return nil
}

// Event is one generated fact.
type Event struct {
	EventType  string         `json:"event_type"`
	OccurredAt string         `json:"occurred_at"`
	Payload    map[string]any `json:"payload"`
}

// Facts is what the generator knows to be true after a run — the ground
// truth the eval judges by, computed without any rule.
type Facts struct {
	Events []Event
	Counts map[string]int // object type → count
	Where  []eval.Where
	Sums   []eval.Sum
	// Totals for the report, in minor units.
	SalesNet, SalesVAT, PurchaseVAT, PZValue int64
	// The bank story's identity, in minor units: money in equals the gross
	// of the sales invoices it paid, less the differences written off, plus
	// what it paid of invoices still open, plus the stray line nobody can
	// match; money out equals the gross of the purchase invoices it paid.
	BankIn, BankOut, PaidSalesGross, PaidPurchaseGross, Differences, PartialOpen, Stray int64
	// Ways counts the payments by the way they were made, after every
	// fall-back: named, exact, subset, short, named_part, oldest_part, stray.
	Ways map[string]int
	// Documents are the same facts as files, the way channels deliver them
	// (documents.go); each names the events it carries.
	Documents []Document
}

// Generate runs the persona through its months. The same persona always
// yields the same facts: randomness comes from Seed alone.
func Generate(p Persona) (Facts, error) {
	if err := p.Validate(); err != nil {
		return Facts{}, err
	}
	g := &gen{p: p, rng: rand.New(rand.NewPCG(p.Seed, p.Seed^0x9e3779b97f4a7c15)),
		stock: map[string]int{}, counts: map[string]int{}, ways: map[string]int{}}
	g.masterData()
	start, _ := time.Parse("2006-01", p.From)
	for m := 0; m < p.Months; m++ {
		g.month(start.AddDate(0, m, 0))
	}
	g.stockCounts()
	g.payments()
	// Facts in time order: a stable sort keeps the generator's order within a day.
	sort.SliceStable(g.events, func(i, j int) bool { return g.events[i].OccurredAt < g.events[j].OccurredAt })
	f := g.facts()
	f.Documents = g.documents(f.Events)
	return f, nil
}

type gen struct {
	p      Persona
	rng    *rand.Rand
	events []Event
	stock  map[string]int // sku|location → quantity
	counts map[string]int

	salesNo, deliveryNo, purchaseNo, complaintNo int
	sales                                        []invoiceFact
	purchases                                    []invoiceFact
	salesNet, salesVAT, purchaseVAT, pzValue     int64
	pzLines, movesIn, movesOut, resolved         int
	// the bank story as the generator predicts it (payments)
	ways                                                                                         map[string]int
	settlements                                                                                  int
	bankIn, bankOut, paidSalesGross, paidPurchaseGross, differences, partialOpen, stray, settled int64
	levels                                                                                       []levelFact
}

// invoiceFact is a generated invoice the bank may settle; the party is the
// customer who pays us or the supplier we pay.
type invoiceFact struct {
	number string
	date   time.Time
	gross  int64
	open   int64 // what payments have not yet settled, as the generator predicts it
	paid   bool
	party  Party
}

type levelFact struct{ book, counted int }

func (g *gen) emit(typ, date string, payload map[string]any) {
	g.events = append(g.events, Event{EventType: typ, OccurredAt: date, Payload: payload})
}

// day picks a day of the month: deliveries early, everything else after,
// so stock is never issued before it arrived in date order.
func (g *gen) day(month time.Time) time.Time {
	return month.AddDate(0, 0, 9+g.rng.IntN(19))
}

func (g *gen) earlyDay(month time.Time) time.Time {
	return month.AddDate(0, 0, g.rng.IntN(9))
}

func ymd(t time.Time) string { return t.Format("2006-01-02") }

func (g *gen) masterData() {
	start, _ := time.Parse("2006-01", g.p.From)
	day := ymd(start) // the first day: rules take effect with the period
	for _, a := range g.p.Accounts {
		g.emit("account.created", day, map[string]any{"code": a.Code, "name": a.Name, "type": a.Type})
	}
	// Every party carries its account: the key a bank line is matched by.
	g.emit("company.registered", day, map[string]any{"name": g.p.Company.Name, "kind": "self",
		"vat_id": g.p.Company.VATID, "country": g.p.Company.Country, "iban": iban(g.p.Company)})
	for _, c := range g.p.Customers {
		g.emit("company.registered", day, map[string]any{"name": c.Name, "kind": "customer", "vat_id": c.VATID, "country": c.Country, "iban": iban(c)})
	}
	for _, s := range g.p.Suppliers {
		g.emit("company.registered", day, map[string]any{"name": s.Name, "kind": "supplier", "vat_id": s.VATID, "country": s.Country, "iban": iban(s)})
	}
	for _, it := range g.p.Items {
		g.emit("item.created", day, map[string]any{"sku": it.SKU, "name": it.Name, "unit": it.Unit})
	}
	for _, l := range g.p.Locations {
		g.emit("location.created", day, map[string]any{"code": l.Code, "name": l.Name})
	}
	for _, r := range g.p.VATRates {
		g.emit("vat_rate.defined", day, map[string]any{"code": r.Code, "name": r.Name, "percent": r.Percent})
	}
	g.counts["account"] = len(g.p.Accounts)
	g.counts["company"] = 1 + len(g.p.Customers) + len(g.p.Suppliers)
	g.counts["item"] = len(g.p.Items)
	g.counts["location"] = len(g.p.Locations)
	g.counts["vat_rate"] = len(g.p.VATRates)
}

func (g *gen) pick(n int) int { return g.rng.IntN(n) }

func (g *gen) goods() []Item {
	var out []Item
	for _, it := range g.p.Items {
		if it.Cost != "" {
			out = append(out, it)
		}
	}
	return out
}

func (g *gen) month(start time.Time) {
	year := start.Year()
	// deliveries first: stock before sales
	for i := 0; i < g.p.Volume.Deliveries; i++ {
		g.deliveryNo++
		date := g.earlyDay(start)
		sup := g.p.Suppliers[g.pick(len(g.p.Suppliers))]
		loc := g.p.Locations[g.pick(len(g.p.Locations))]
		goods := g.goods()
		n := 1 + g.pick(min(3, len(goods)))
		var lines []any
		var net int64
		for _, it := range g.rng.Perm(len(goods))[:n] {
			item := goods[it]
			qty := 5 + g.pick(20)
			cost, _ := core.ParseMoney(item.Cost)
			lines = append(lines, map[string]any{"item": item.SKU, "qty": qty, "price": item.Cost})
			g.stock[item.SKU+"|"+loc.Code] += qty
			net += int64(qty) * cost
			g.pzLines++
			g.movesIn++
		}
		doc := fmt.Sprintf("WZ %d/%d", g.deliveryNo, year)
		g.emit("delivery.received", ymd(date), map[string]any{"document": doc, "supplier": sup.Name,
			"location": loc.Code, "date": ymd(date), "currency": "PLN", "lines": lines})
		g.pzValue += net
		g.counts["pz"]++
		if i < g.p.Volume.PurchaseInvoices {
			g.purchaseInvoice(date.AddDate(0, 0, 1+g.pick(5)), sup, net, doc)
		}
	}
	// costs without goods: a services business buys too
	if g.p.Volume.Deliveries == 0 {
		for i := 0; i < g.p.Volume.PurchaseInvoices; i++ {
			sup := g.p.Suppliers[g.pick(len(g.p.Suppliers))]
			net := int64(50000 + g.pick(400)*100)
			g.purchaseInvoice(g.day(start), sup, net, "")
		}
	}
	for i := 0; i < g.p.Volume.SalesInvoices; i++ {
		g.salesNo++
		date := g.day(start)
		cust := g.p.Customers[g.pick(len(g.p.Customers))]
		n := 1 + g.pick(min(3, len(g.p.Items)))
		var lines []any
		var net int64
		for _, ix := range g.rng.Perm(len(g.p.Items))[:n] {
			item := g.p.Items[ix]
			qty := 1 + g.pick(6)
			price, _ := core.ParseMoney(item.Price)
			lines = append(lines, map[string]any{"item": item.SKU, "qty": qty, "price": item.Price})
			net += int64(qty) * price
		}
		rate := g.p.VATRates[0]
		vat := halfUp(net*int64(rate.Percent), 100)
		number := fmt.Sprintf("FS %d/%d", g.salesNo, year)
		g.emit("sales.invoice.issued", ymd(date), map[string]any{"number": number, "issue_date": ymd(date),
			"seller_nip": g.p.Company.VATID, "buyer_nip": cust.VATID, "buyer": cust.Name,
			"net": core.FormatMoney(net), "vat_rate": rate.Code, "currency": "PLN",
			"payment_days": g.p.Volume.PaymentDays, "market": "pl", "lines": lines})
		g.salesNet += net
		g.salesVAT += vat
		g.sales = append(g.sales, invoiceFact{number: number, date: date, gross: net + vat, party: cust})
		g.counts["sales_invoice"]++
	}
	for i := 0; i < g.p.Volume.Issues; i++ {
		goods := g.goods()
		if len(goods) == 0 || len(g.p.Locations) == 0 {
			break
		}
		item := goods[g.pick(len(goods))]
		loc := g.p.Locations[g.pick(len(g.p.Locations))]
		have := g.stock[item.SKU+"|"+loc.Code]
		if have < 1 {
			continue // nothing to issue; the warehouse is honest
		}
		qty := 1 + g.pick(min(have, 8))
		g.stock[item.SKU+"|"+loc.Code] -= qty
		date := ymd(g.day(start))
		g.emit("goods.issued", date, map[string]any{"item": item.SKU, "location": loc.Code,
			"qty": qty, "date": date, "reason": "sale"})
		g.counts["goods_issue"]++
		g.movesOut++
	}
	for i := 0; i < g.p.Volume.Complaints; i++ {
		g.complaintNo++
		date := g.day(start)
		cust := g.p.Customers[g.pick(len(g.p.Customers))]
		number := fmt.Sprintf("R-%d", g.complaintNo)
		g.emit("complaint.registered", ymd(date), map[string]any{"number": number, "customer": cust.Name,
			"details": []string{"wrong amount on the invoice", "late delivery", "damaged goods", "missing document"}[g.pick(4)]})
		g.counts["complaint"]++
		g.counts["case"]++
		if g.pick(2) == 0 {
			g.emit("complaint.withdrawn", ymd(date.AddDate(0, 0, 2+g.pick(10))), map[string]any{"number": number})
			g.resolved++
		}
	}
}

func (g *gen) purchaseInvoice(date time.Time, sup Party, net int64, forDocument string) {
	g.purchaseNo++
	rate := g.p.VATRates[0]
	vat := halfUp(net*int64(rate.Percent), 100)
	number := fmt.Sprintf("%s/%d/%d", strings.ToUpper(sup.Name[:3]), g.purchaseNo, date.Year())
	payload := map[string]any{"number": number,
		"supplier": sup.Name, "supplier_nip": sup.VATID, "buyer_nip": g.p.Company.VATID, "invoice_date": ymd(date),
		"net": core.FormatMoney(net), "vat": core.FormatMoney(vat), "gross": core.FormatMoney(net + vat),
		"vat_rate": rate.Code, "currency": "PLN", "market": "pl"}
	if forDocument != "" {
		payload["for_document"] = forDocument
	}
	g.emit("purchase.invoice.received", ymd(date), payload)
	g.purchaseVAT += vat
	g.purchases = append(g.purchases, invoiceFact{number: number, date: date, gross: net + vat, party: sup})
	g.counts["purchase_invoice"]++
}

func (g *gen) stockCounts() {
	for _, date := range g.p.Volume.StockCounts {
		// the book quantity at a count is the stock after every movement up
		// to that day; movements are generated day by day, so recount them
		// from the events rather than trusting the running total
		book := map[string]int{}
		for _, ev := range g.events {
			if ev.OccurredAt > date {
				continue
			}
			switch ev.EventType {
			case "delivery.received":
				for _, l := range ev.Payload["lines"].([]any) {
					line := l.(map[string]any)
					book[line["item"].(string)+"|"+ev.Payload["location"].(string)] += line["qty"].(int)
				}
			case "goods.issued":
				book[ev.Payload["item"].(string)+"|"+ev.Payload["location"].(string)] -= ev.Payload["qty"].(int)
			}
		}
		for _, it := range g.goods() {
			for _, loc := range g.p.Locations {
				key := it.SKU + "|" + loc.Code
				counted := book[key] + g.pick(5) - 2
				if counted < 0 {
					counted = 0
				}
				g.emit("stock.counted", date, map[string]any{"item": it.SKU, "location": loc.Code, "counted": counted, "date": date})
				g.levels = append(g.levels, levelFact{book: book[key], counted: counted})
				g.counts["stock_level"]++
			}
		}
	}
}

// payWay is one of the ways a customer pays — the ways the owner's matching
// policy names (DIRECTION 2026-10-11): the simulated customer pays in every
// one of them.
type payWay int

const (
	payNamed      payWay = iota // the invoice named in the title, whole
	payExact                    // unnamed, the exact amount
	paySubset                   // unnamed, several invoices of one customer in one line
	payShort                    // named, two grosz short — the difference written off
	payNamedPart                // named, in part; the rest named later
	payOldestPart               // unnamed, in part: the customer's oldest open invoice; the rest named later
)

type plannedPayment struct {
	date     time.Time
	way      payWay
	invs     []*invoiceFact
	purchase bool
}

// payments: a share of sales invoices is paid to us, in every way the
// matching policy names, and a share of purchase invoices is paid by us,
// whole, naming the invoice. The generator predicts what the policy
// settles without any rule: it walks the planned payments in date order
// with the open amount of every invoice, and a scenario the policy could
// read two ways (an unnamed amount two open invoices share, a sum two
// subsets reach) falls back to a payment naming the invoice. In the eval
// every invoice of the period exists before any line is matched, so "open"
// counts them all; a live run sees only the past ones, and the scenarios
// hold either way — the oldest open invoice is a past one, and fewer
// candidates make nothing ambiguous. One stray line, from a payer nobody
// knows, stays open for the person. A line says what a statement says: the
// direction, the account, the counterparty and its account.
func (g *gen) payments() {
	end, _ := time.Parse("2006-01", g.p.From)
	end = end.AddDate(0, g.p.Months, -1)
	later := func(from time.Time) time.Time {
		return from.AddDate(0, 0, 3+g.pick(max(1, g.p.Volume.PaymentDays+10)))
	}
	var plans []plannedPayment
	members := map[string][]*invoiceFact{} // customer → invoices to be paid together
	for i := range g.sales {
		s := &g.sales[i]
		s.open = s.gross
		if g.pick(100) >= g.p.Volume.PaidShare {
			continue
		}
		way := payNamed
		switch d := g.pick(20); {
		case d < 8:
			way = payNamed
		case d < 11:
			way = payExact
		case d < 15:
			way = paySubset
		case d < 17:
			way = payShort
		case d < 18:
			way = payNamedPart
		default:
			way = payOldestPart
		}
		if way == paySubset {
			members[s.party.Name] = append(members[s.party.Name], s)
			continue
		}
		plans = append(plans, plannedPayment{date: later(s.date), way: way, invs: []*invoiceFact{s}})
	}
	names := make([]string, 0, len(members))
	for name := range members {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		group := members[name]
		for len(group) >= 2 {
			k := min(len(group), 2+g.pick(2))
			together := group[:k]
			group = group[k:]
			latest := together[0].date
			for _, m := range together {
				if m.date.After(latest) {
					latest = m.date
				}
			}
			plans = append(plans, plannedPayment{date: later(latest), way: paySubset, invs: together})
		}
		for _, m := range group { // a lone member pays alone, named
			plans = append(plans, plannedPayment{date: later(m.date), way: payNamed, invs: []*invoiceFact{m}})
		}
	}
	for i := range g.purchases {
		pf := &g.purchases[i]
		pf.open = pf.gross
		if g.pick(100) >= g.p.Volume.PaidShare {
			continue
		}
		plans = append(plans, plannedPayment{date: later(pf.date), way: payNamed, invs: []*invoiceFact{pf}, purchase: true})
	}
	// The walk: earliest first, follow-ups (the rest of a part payment)
	// joining the queue at their own date.
	for len(plans) > 0 {
		sort.SliceStable(plans, func(i, j int) bool { return plans[i].date.Before(plans[j].date) })
		pl := plans[0]
		plans = plans[1:]
		if pl.date.After(end) {
			continue
		}
		if rest, ok := g.apply(pl, later); ok {
			plans = append(plans, rest)
		}
	}
	// The stray line: a payer no company carries the account of.
	mid, _ := time.Parse("2006-01", g.p.From)
	mid = mid.AddDate(0, g.p.Months/2, 14)
	g.line(mid, 10000, "credit", "PRZELEW", Party{Name: "Nieznany Płatnik", IBAN: "PL00000000000000000000000000"})
	g.stray += 10000
	g.ways["stray"]++
}

// line emits one bank statement line.
func (g *gen) line(date time.Time, amount int64, direction, reference string, who Party) {
	g.emit("bank.statement.line", ymd(date), map[string]any{"date": ymd(date), "amount": core.FormatMoney(amount),
		"currency": "PLN", "direction": direction, "reference": reference, "counterparty": who.Name,
		"counterparty_account": iban(who), "account": iban(g.p.Company)})
	g.counts["bank_line"]++
	if direction == "credit" {
		g.bankIn += amount
	} else {
		g.bankOut += amount
	}
}

// openOf lists a party's invoices with something open, the way the engine
// orders them: by date, then the order they were issued in.
func (g *gen) openOf(party string, purchase bool) []*invoiceFact {
	all := g.sales
	if purchase {
		all = g.purchases
	}
	var out []*invoiceFact
	for i := range all {
		if all[i].party.Name == party && all[i].open > 0 {
			out = append(out, &all[i])
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].date.Before(out[j].date) })
	return out
}

// apply emits the line of one planned payment and records what the policy
// will settle; a scenario the policy could read two ways falls back to
// naming the invoice. It returns the follow-up a part payment schedules.
func (g *gen) apply(pl plannedPayment, later func(time.Time) time.Time) (plannedPayment, bool) {
	x := pl.invs[0]
	direction := "credit"
	if pl.purchase {
		direction = "debit"
	}
	settle := func(inv *invoiceFact, amount int64) {
		before := inv.open
		inv.open -= amount
		g.settlements++
		g.settled += amount
		if inv.open == 0 {
			inv.paid = true
			g.partialOpen -= inv.gross - before // what earlier part payments had left open
			if pl.purchase {
				g.paidPurchaseGross += inv.gross
			} else {
				g.paidSalesGross += inv.gross
			}
		}
	}
	named := func(inv *invoiceFact) {
		g.line(pl.date, inv.open, direction, inv.number, inv.party)
		settle(inv, inv.open)
		g.ways["named"]++
	}
	// An invoice another line settled meanwhile (a part payment's rest
	// after its own named payment, a subset member paid in part) has
	// nothing left for this plan.
	var live []*invoiceFact
	for _, inv := range pl.invs {
		if inv.open > 0 {
			live = append(live, inv)
		}
	}
	if len(live) == 0 {
		return plannedPayment{}, false
	}
	pl.invs, x = live, live[0]
	if pl.way == paySubset && len(live) < 2 {
		named(x)
		return plannedPayment{}, false
	}
	open := g.openOf(x.party.Name, pl.purchase)
	switch pl.way {
	case payExact:
		if countOpen(open, x.open) != 1 {
			named(x)
			return plannedPayment{}, false
		}
		g.line(pl.date, x.open, direction, "PRZELEW", x.party)
		settle(x, x.open)
		g.ways["exact"]++
	case paySubset:
		var sum int64
		for _, inv := range pl.invs {
			sum += inv.open
		}
		// the fewest items first: no single invoice, no smaller subset,
		// and exactly one subset of this size
		if countOpen(open, sum) != 0 || subsets(open, sum, len(pl.invs)-1) != 0 || subsets(open, sum, len(pl.invs)) != 1 {
			for _, inv := range pl.invs {
				named(inv)
			}
			return plannedPayment{}, false
		}
		g.line(pl.date, sum, direction, "PRZELEW", x.party)
		for _, inv := range pl.invs {
			settle(inv, inv.open)
		}
		g.ways["subset"]++
	case payShort:
		g.line(pl.date, x.open-2, direction, x.number, x.party)
		g.differences += 2
		g.counts["matching_difference"]++
		settle(x, x.open)
		g.ways["short"]++
	case payNamedPart, payOldestPart:
		target := x
		if pl.way == payOldestPart {
			target = open[0] // the customer's oldest open invoice, whoever it is
		}
		part := target.open / 2
		// an unnamed part must match nothing whole: no single invoice, no
		// subset the policy would find first; a named part only needs its
		// invoice open
		if part < 1 || (pl.way == payOldestPart && (countOpen(open, part) != 0 || subsets(open, part, 4) != 0)) {
			named(x)
			return plannedPayment{}, false
		}
		reference, way := target.number, "named_part"
		if pl.way == payOldestPart {
			reference, way = "PRZELEW", "oldest_part"
		}
		g.line(pl.date, part, direction, reference, target.party)
		settle(target, part)
		g.partialOpen += part
		g.ways[way]++
		return plannedPayment{date: later(pl.date), way: payNamed, invs: []*invoiceFact{target}, purchase: pl.purchase}, true
	default:
		named(x)
	}
	return plannedPayment{}, false
}

func countOpen(open []*invoiceFact, amount int64) int {
	n := 0
	for _, inv := range open {
		if inv.open == amount {
			n++
		}
	}
	return n
}

// subsets counts the subsets of exactly k open invoices whose open amounts
// sum to the target — the generator's check that the policy's subset step
// reads a line one way only. Stops counting at two.
func subsets(open []*invoiceFact, target int64, k int) int {
	if k <= 0 {
		return 0
	}
	n := 0
	var walk func(i, left int, sum int64)
	walk = func(i, left int, sum int64) {
		if n >= 2 || sum > target {
			return
		}
		if left == 0 {
			if sum == target {
				n++
			}
			return
		}
		for j := i; j <= len(open)-left; j++ {
			walk(j+1, left-1, sum+open[j].open)
		}
	}
	walk(0, k, 0)
	return n
}

func (g *gen) facts() Facts {
	f := Facts{Events: g.events, Counts: map[string]int{}, SalesNet: g.salesNet, SalesVAT: g.salesVAT,
		PurchaseVAT: g.purchaseVAT, PZValue: g.pzValue,
		BankIn: g.bankIn, BankOut: g.bankOut, PaidSalesGross: g.paidSalesGross, PaidPurchaseGross: g.paidPurchaseGross,
		Differences: g.differences, PartialOpen: g.partialOpen, Stray: g.stray, Ways: g.ways}
	for k, v := range g.counts {
		if v > 0 {
			f.Counts[k] = v
		}
	}
	if g.pzLines > 0 {
		f.Counts["pz_line"] = g.pzLines
	}
	if g.movesIn+g.movesOut > 0 {
		f.Counts["stock_movement"] = g.movesIn + g.movesOut
		f.Where = append(f.Where, eval.Where{Type: "stock_movement", Field: "direction", Value: "in", Count: g.movesIn})
	}
	if g.resolved > 0 {
		f.Where = append(f.Where, eval.Where{Type: "case", Field: "status", Value: "resolved", Count: g.resolved})
	}
	paid := 0
	for _, s := range g.sales {
		if s.paid {
			paid++
		}
	}
	if paid > 0 {
		f.Where = append(f.Where, eval.Where{Type: "sales_invoice", Field: "status", Value: "paid", Count: paid})
	}
	paidPurchases := 0
	for _, pf := range g.purchases {
		if pf.paid {
			paidPurchases++
		}
	}
	if paidPurchases > 0 {
		f.Where = append(f.Where, eval.Where{Type: "purchase_invoice", Field: "status", Value: "paid", Count: paidPurchases})
	}
	// The bank story: every line but the stray one settles; a settlement
	// per invoice a match allocated; a difference per short payment.
	if lines := g.counts["bank_line"]; lines > 0 {
		f.Where = append(f.Where,
			eval.Where{Type: "bank_line", Field: "status", Value: "settled", Count: lines - 1},
			eval.Where{Type: "bank_line", Field: "status", Value: "open", Count: 1})
		f.Counts["settlement"] = g.settlements
		f.Sums = append(f.Sums, eval.Sum{Type: "settlement", Field: "amount", Total: core.FormatMoney(g.settled)})
	}
	if g.counts["matching_difference"] > 0 {
		f.Sums = append(f.Sums, eval.Sum{Type: "matching_difference", Field: "amount", Total: core.FormatMoney(g.differences)})
	}
	postings := 0
	for kind, lines := range g.p.Postings {
		postings += lines * f.Counts[kind]
	}
	if postings > 0 {
		f.Counts["posting"] = postings
		if g.p.Book != "" {
			f.Where = append(f.Where, eval.Where{Type: "posting", Field: "book", Value: g.p.Book, Count: postings})
		}
	}
	if f.Counts["sales_invoice"] > 0 {
		f.Sums = append(f.Sums,
			eval.Sum{Type: "sales_invoice", Field: "vat", Total: core.FormatMoney(g.salesVAT)},
			eval.Sum{Type: "sales_invoice", Field: "gross", Total: core.FormatMoney(g.salesNet + g.salesVAT)})
	}
	if f.Counts["purchase_invoice"] > 0 {
		f.Sums = append(f.Sums, eval.Sum{Type: "purchase_invoice", Field: "vat", Total: core.FormatMoney(g.purchaseVAT)})
	}
	if g.pzLines > 0 {
		f.Sums = append(f.Sums, eval.Sum{Type: "pz_line", Field: "value", Total: core.FormatMoney(g.pzValue)})
	}
	if len(g.levels) > 0 {
		book, diff := 0, 0
		for _, l := range g.levels {
			book += l.book
			diff += l.counted - l.book
		}
		f.Sums = append(f.Sums,
			eval.Sum{Type: "stock_level", Field: "book", Total: fmt.Sprint(book)},
			eval.Sum{Type: "stock_level", Field: "difference", Total: fmt.Sprint(diff)})
	}
	sort.Slice(f.Where, func(i, j int) bool { return f.Where[i].Type+f.Where[i].Field < f.Where[j].Type+f.Where[j].Field })
	return f
}

// halfUp divides non-negative integers rounding half away from zero — the
// statutory method the personas book VAT with, computed here independently
// of the kernel so the two can disagree.
func halfUp(n, d int64) int64 {
	q, r := n/d, n%d
	if 2*r >= d {
		q++
	}
	return q
}

// Compile turns a persona into the corpus the eval runs: seeds, the
// generated intake, the interview as tasks (voice 0 is the scripted answer,
// voice k the k-th paraphrase), and the expected state. The corpus is named
// with the synthetic prefix so a networked run is marked.
func Compile(p Persona, voices *Voices, voice int) (eval.Corpus, Facts, error) {
	return CompileDocuments(p, voices, voice, nil)
}

// Readers names the channel that reads each document source a persona can
// emit. A source without a reader stays JSON through the open door —
// exactly "what no channel reads yet".
var Readers = map[string]string{"ksef": "ksef_inbound", "bank": "bank_mt940", "pdf": "pdf_invoice"}

// CompileDocuments compiles with the documents of the named sources fed as
// files instead of the facts they carry — documents in, state out
// (DIRECTION 2026-10-10). The expected state gains the received documents,
// read by their channel.
func CompileDocuments(p Persona, voices *Voices, voice int, sources []string) (eval.Corpus, Facts, error) {
	facts, err := Generate(p)
	if err != nil {
		return eval.Corpus{}, Facts{}, err
	}
	c := eval.Corpus{Name: network.SyntheticPrefix + p.Name, Seeds: p.Seeds, Packs: p.Packs, Expect: facts.Counts,
		ExpectWhere: facts.Where, ExpectSum: facts.Sums}
	covered := map[int]bool{}
	for _, src := range sources {
		reader, ok := Readers[src]
		if !ok {
			return eval.Corpus{}, Facts{}, fmt.Errorf("no reader for %q documents yet — they stay JSON", src)
		}
		if !slices.Contains(c.Readers, reader) {
			c.Readers = append(c.Readers, reader) // in the order given: invoices before the lines that pay them
		}
		for _, d := range facts.Documents {
			if d.Source != src {
				continue
			}
			c.Files = append(c.Files, eval.File{Name: filepath.Base(d.Path), MediaType: d.MediaType,
				Source: d.Source, Date: d.Date, ContentBase64: base64.StdEncoding.EncodeToString(d.Bytes)})
			for _, i := range d.Events {
				covered[i] = true
			}
		}
	}
	if len(c.Files) > 0 {
		c.Expect = make(map[string]int, len(facts.Counts)+1)
		for k, v := range facts.Counts {
			c.Expect[k] = v
		}
		c.Expect["received_document"] = len(c.Files)
		c.ExpectWhere = append(append([]eval.Where{}, facts.Where...),
			eval.Where{Type: "received_document", Field: "status", Value: "read", Count: len(c.Files)})
	}
	for i, ev := range facts.Events {
		if covered[i] {
			continue
		}
		b, err := json.Marshal(ev)
		if err != nil {
			return eval.Corpus{}, Facts{}, err
		}
		c.Events = append(c.Events, b)
	}
	for i, a := range p.Interview {
		intent := a.Answer
		if voice > 0 {
			if voices == nil || i >= len(voices.Answers) || voice > len(voices.Answers[i]) {
				return eval.Corpus{}, Facts{}, fmt.Errorf("persona %s: no voice %d for answer %d — run paraphrase first", p.Name, voice, i+1)
			}
			intent = voices.Answers[i][voice-1]
		}
		c.Tasks = append(c.Tasks, eval.Task{Intent: intent, Mode: a.Mode, SampleType: a.SampleType, Reference: a.Reference})
	}
	return c, facts, nil
}

// Summary renders what a persona generates, for the terminal.
func Summary(p Persona, f Facts) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "persona %s — %s\n%d months from %s: %d events, %d interview answers\n",
		p.Name, p.Description, p.Months, p.From, len(f.Events), len(p.Interview))
	kinds := map[string]int{}
	for _, ev := range f.Events {
		kinds[ev.EventType]++
	}
	names := make([]string, 0, len(kinds))
	for k := range kinds {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		fmt.Fprintf(&sb, "  %-26s %4d\n", k, kinds[k])
	}
	fmt.Fprintf(&sb, "expected: sales net %s, VAT %s; purchase VAT %s; PZ value %s\n",
		core.FormatMoney(f.SalesNet), core.FormatMoney(f.SalesVAT), core.FormatMoney(f.PurchaseVAT), core.FormatMoney(f.PZValue))
	docs := map[string]int{}
	covered := map[int]bool{}
	for _, d := range f.Documents {
		docs[d.Source]++
		for _, i := range d.Events {
			covered[i] = true
		}
	}
	foreign := 0
	for i, ev := range f.Events {
		if ev.EventType == "purchase.invoice.received" && !covered[i] {
			foreign++
		}
	}
	ways := make([]string, 0, len(f.Ways))
	for w := range f.Ways {
		ways = append(ways, w)
	}
	sort.Strings(ways)
	sb.WriteString("payments by way:")
	for _, w := range ways {
		fmt.Fprintf(&sb, " %s %d", w, f.Ways[w])
	}
	sb.WriteByte('\n')
	fmt.Fprintf(&sb, "documents: %d KSeF invoices, %d PDF invoices from foreign suppliers, %d bank statements", docs["ksef"], docs["pdf"], docs["bank"])
	if foreign > 0 {
		fmt.Fprintf(&sb, "; %d purchase invoices from foreign suppliers still arrive as JSON", foreign)
	}
	sb.WriteByte('\n')
	types := make([]string, 0, len(f.Counts))
	for t := range f.Counts {
		types = append(types, t)
	}
	sort.Strings(types)
	sb.WriteString("expected objects:")
	for _, t := range types {
		fmt.Fprintf(&sb, " %s %d", t, f.Counts[t])
	}
	sb.WriteByte('\n')
	return sb.String()
}
