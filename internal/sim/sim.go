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
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
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
	Seeds       []string `json:"seeds"` // definition files, relative to the persona file
	Seed        uint64   `json:"seed"`  // the generator's randomness, fixed
	From        string   `json:"from"`  // first month, YYYY-MM
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
}

// Generate runs the persona through its months. The same persona always
// yields the same facts: randomness comes from Seed alone.
func Generate(p Persona) (Facts, error) {
	if err := p.Validate(); err != nil {
		return Facts{}, err
	}
	g := &gen{p: p, rng: rand.New(rand.NewPCG(p.Seed, p.Seed^0x9e3779b97f4a7c15)),
		stock: map[string]int{}, counts: map[string]int{}}
	g.masterData()
	start, _ := time.Parse("2006-01", p.From)
	for m := 0; m < p.Months; m++ {
		g.month(start.AddDate(0, m, 0))
	}
	g.stockCounts()
	g.payments()
	// Facts in time order: a stable sort keeps the generator's order within a day.
	sort.SliceStable(g.events, func(i, j int) bool { return g.events[i].OccurredAt < g.events[j].OccurredAt })
	return g.facts(), nil
}

type gen struct {
	p      Persona
	rng    *rand.Rand
	events []Event
	stock  map[string]int // sku|location → quantity
	counts map[string]int

	salesNo, deliveryNo, purchaseNo, complaintNo int
	sales                                        []saleFact
	salesNet, salesVAT, purchaseVAT, pzValue     int64
	pzLines, movesIn, movesOut, resolved         int
	levels                                       []levelFact
}

type saleFact struct {
	number string
	date   time.Time
	gross  int64
	paid   bool
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
	g.emit("company.registered", day, map[string]any{"name": g.p.Company.Name, "kind": "self",
		"vat_id": g.p.Company.VATID, "country": g.p.Company.Country})
	for _, c := range g.p.Customers {
		g.emit("company.registered", day, map[string]any{"name": c.Name, "kind": "customer", "vat_id": c.VATID, "country": c.Country})
	}
	for _, s := range g.p.Suppliers {
		g.emit("company.registered", day, map[string]any{"name": s.Name, "kind": "supplier", "vat_id": s.VATID, "country": s.Country})
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
		g.sales = append(g.sales, saleFact{number: number, date: date, gross: net + vat})
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
	payload := map[string]any{"number": fmt.Sprintf("%s/%d/%d", strings.ToUpper(sup.Name[:3]), g.purchaseNo, date.Year()),
		"supplier": sup.Name, "supplier_nip": sup.VATID, "invoice_date": ymd(date),
		"net": core.FormatMoney(net), "vat_rate": rate.Code, "currency": "PLN"}
	if forDocument != "" {
		payload["for_document"] = forDocument
	}
	g.emit("purchase.invoice.received", ymd(date), payload)
	g.purchaseVAT += vat
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

// payments: a share of sales invoices is paid, by a bank line naming the
// invoice, some days after issue and inside the period.
func (g *gen) payments() {
	end, _ := time.Parse("2006-01", g.p.From)
	end = end.AddDate(0, g.p.Months, -1)
	for i := range g.sales {
		s := &g.sales[i]
		if g.pick(100) >= g.p.Volume.PaidShare {
			continue
		}
		paid := s.date.AddDate(0, 0, 3+g.pick(max(1, g.p.Volume.PaymentDays+10)))
		if paid.After(end) {
			continue
		}
		s.paid = true
		g.emit("bank.statement.line", ymd(paid), map[string]any{"date": ymd(paid), "amount": core.FormatMoney(s.gross),
			"currency": "PLN", "reference": s.number, "counterparty": "customer transfer"})
	}
}

func (g *gen) facts() Facts {
	f := Facts{Events: g.events, Counts: map[string]int{}, SalesNet: g.salesNet, SalesVAT: g.salesVAT,
		PurchaseVAT: g.purchaseVAT, PZValue: g.pzValue}
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
	facts, err := Generate(p)
	if err != nil {
		return eval.Corpus{}, Facts{}, err
	}
	c := eval.Corpus{Name: network.SyntheticPrefix + p.Name, Seeds: p.Seeds, Expect: facts.Counts,
		ExpectWhere: facts.Where, ExpectSum: facts.Sums}
	for _, ev := range facts.Events {
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
