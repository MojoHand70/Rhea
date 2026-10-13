package sim_test

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/sim"
)

func load(t *testing.T, name string) sim.Persona {
	t.Helper()
	p, _, err := sim.Load(filepath.Join("..", "..", "testdata", "customers", name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// The same persona always yields the same facts: the seed is the only
// randomness, so an eval run is reproducible and a demo dataset is stable.
func TestGenerateIsDeterministic(t *testing.T) {
	p := load(t, "nordwind")
	a, err := sim.Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := sim.Generate(p)
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	if string(ja) != string(jb) {
		t.Fatal("two generations of one persona differ")
	}
	p.Seed++
	c, _ := sim.Generate(p)
	jc, _ := json.Marshal(c)
	if string(ja) == string(jc) {
		t.Fatal("a different seed yields the same facts")
	}
}

// The expected state is consistent with the events — counted from them
// here a second way — and the warehouse never issues what it does not hold.
func TestFactsAreConsistent(t *testing.T) {
	p := load(t, "nordwind")
	f, err := sim.Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	kinds := map[string]int{}
	stock := map[string]int{}
	var pzValue int64
	lines, movesIn := 0, 0
	for i, ev := range f.Events {
		kinds[ev.EventType]++
		if i > 0 && ev.OccurredAt < f.Events[i-1].OccurredAt {
			t.Fatalf("events out of order at %d", i)
		}
		switch ev.EventType {
		case "delivery.received":
			for _, l := range ev.Payload["lines"].([]any) {
				line := l.(map[string]any)
				qty := line["qty"].(int)
				price, _ := core.ParseMoney(line["price"].(string))
				pzValue += int64(qty) * price
				stock[line["item"].(string)+"|"+ev.Payload["location"].(string)] += qty
				lines++
				movesIn++
			}
		case "goods.issued":
			key := ev.Payload["item"].(string) + "|" + ev.Payload["location"].(string)
			stock[key] -= ev.Payload["qty"].(int)
			if stock[key] < 0 {
				t.Fatalf("issued below zero: %s", key)
			}
		}
	}
	if f.Counts["sales_invoice"] != kinds["sales.invoice.issued"] || f.Counts["pz"] != kinds["delivery.received"] ||
		f.Counts["purchase_invoice"] != kinds["purchase.invoice.received"] || f.Counts["complaint"] != kinds["complaint.registered"] ||
		f.Counts["case"] != kinds["complaint.registered"] || f.Counts["goods_issue"] != kinds["goods.issued"] ||
		f.Counts["stock_level"] != kinds["stock.counted"] || f.Counts["pz_line"] != lines {
		t.Fatalf("counts %v vs events %v", f.Counts, kinds)
	}
	if f.Counts["stock_movement"] != movesIn+kinds["goods.issued"] {
		t.Fatalf("movements %d, want %d in + %d out", f.Counts["stock_movement"], movesIn, kinds["goods.issued"])
	}
	want := 3*f.Counts["sales_invoice"] + 3*f.Counts["purchase_invoice"] + 2*lines + 2*f.Counts["matching_difference"]
	if f.Counts["posting"] != want {
		t.Fatalf("postings %d, want %d", f.Counts["posting"], want)
	}
	if f.PZValue != pzValue {
		t.Fatalf("PZ value %d, want %d", f.PZValue, pzValue)
	}
	sums := map[string]string{}
	for _, s := range f.Sums {
		sums[s.Type+"."+s.Field] = s.Total
	}
	if sums["pz_line.value"] != core.FormatMoney(pzValue) || sums["sales_invoice.vat"] == "" || sums["stock_level.book"] == "" {
		t.Fatalf("sums = %v", sums)
	}
	bankStory(t, f)
	if kinds["stock.counted"] != 2*len(p.Locations)*3 {
		t.Fatalf("stock counts %d", kinds["stock.counted"])
	}
}

// Voices substitute the scripted answer; a missing voice is refused.
func TestCompileVoices(t *testing.T) {
	p := load(t, "helios")
	p.Months = 1
	c, _, err := sim.Compile(p, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	if c.Name != "sim:helios" || len(c.Tasks) != len(p.Interview) || c.Tasks[0].Intent != p.Interview[0].Answer {
		t.Fatalf("corpus = %s, %d tasks", c.Name, len(c.Tasks))
	}
	if _, _, err := sim.Compile(p, nil, 1); err == nil || !strings.Contains(err.Error(), "run paraphrase first") {
		t.Fatalf("missing voice: %v", err)
	}
	v := sim.Voices{Model: "test"}
	for range p.Interview {
		v.Answers = append(v.Answers, []string{"first way", "second way"})
	}
	c, _, err = sim.Compile(p, &v, 2)
	if err != nil || c.Tasks[0].Intent != "second way" || string(c.Tasks[0].Reference) != string(p.Interview[0].Reference) {
		t.Fatalf("voice 2: %v %q", err, c.Tasks[0].Intent)
	}
}

// A paraphrase keeps every fact: numbers, codes, status words.
func TestPreserved(t *testing.T) {
	orig := "Each sales invoice is posted in the book pl-stat: debit 201 with the gross, credit 700 with the net; the status starts open."
	if m := sim.Preserved(orig, "Post every sales invoice to pl-stat — 201 debit for the gross, 700 credit for the net — and it starts out open."); len(m) != 0 {
		t.Fatalf("faithful paraphrase reported %v", m)
	}
	m := sim.Preserved(orig, "Post every sales invoice: debit 201 gross, credit 702 net; it starts open.")
	if len(m) != 2 || m[0] != "pl-stat" || m[1] != "700" {
		t.Fatalf("missing = %v", m)
	}
	// a hyphen alone is wording; a named book or type is a fact
	if m := sim.Preserved("Every complaint opens a follow-up case in the book pl-stat.", "Each complaint opens a case to track it, in the book pl-stat."); len(m) != 0 {
		t.Fatalf("wording reported as fact: %v", m)
	}
	if m := sim.Preserved("A new type named pz-doc holds it in the book pl-stat.", "A new type called receipt holds it in the main book."); len(m) != 2 || m[0] != "pz-doc" || m[1] != "pl-stat" {
		t.Fatalf("missing = %v", m)
	}
	// the rounding stance is a phrase in any spelling; "book quantity" and
	// "dated up to" are wording
	if m := sim.Preserved("rounded half up to the grosz; the book quantity, dated up to the count date", "half-up rounding to the grosz; the ledger amount through the count date"); len(m) != 0 {
		t.Fatalf("wording reported as fact: %v", m)
	}
	if m := sim.Preserved("rounded half up to the grosz", "rounded to the grosz"); len(m) != 1 || m[0] != "half up" {
		t.Fatalf("dropped rounding not reported: %v", m)
	}
	// a small number spelled as a word is the same fact
	if m := sim.Preserved("the fewest invoices that add up to it, at most 4", "the fewest invoices adding up to it, at most four"); len(m) != 0 {
		t.Fatalf("four is 4: %v", m)
	}
	if m := sim.Preserved("at most 4 invoices", "the fewest invoices"); len(m) != 1 || m[0] != "4" {
		t.Fatalf("a dropped number is reported: %v", m)
	}
}

// The bank story's identity, counted from the events a second way: money
// in is the gross of the sales invoices it paid, less the differences
// written off, plus what it paid of invoices still open, plus the one stray
// line; money out is the gross of the purchase invoices it paid. And the
// customer pays in every way the policy names, after every fall-back.
func bankStory(t *testing.T, f sim.Facts) {
	t.Helper()
	var credits, debits int64
	bankLines := 0
	for _, ev := range f.Events {
		if ev.EventType == "bank.statement.line" {
			bankLines++
			amount, _ := core.ParseMoney(ev.Payload["amount"].(string))
			if ev.Payload["direction"] == "debit" {
				debits += amount
			} else {
				credits += amount
			}
		}
	}
	if credits != f.BankIn || debits != f.BankOut || bankLines != f.Counts["bank_line"] {
		t.Fatalf("lines: in %d vs %d, out %d vs %d, %d vs %d", credits, f.BankIn, debits, f.BankOut, bankLines, f.Counts["bank_line"])
	}
	if f.BankIn != f.PaidSalesGross-f.Differences+f.PartialOpen+f.Stray || f.BankOut != f.PaidPurchaseGross {
		t.Fatalf("money in %d ≠ paid %d − differences %d + partial %d + stray %d, or out %d ≠ paid %d",
			f.BankIn, f.PaidSalesGross, f.Differences, f.PartialOpen, f.Stray, f.BankOut, f.PaidPurchaseGross)
	}
	for _, way := range []string{"named", "exact", "subset", "short", "named_part", "oldest_part", "stray"} {
		if f.Ways[way] == 0 {
			t.Fatalf("the customer pays in every way: %v lacks %s", f.Ways, way)
		}
	}
	if f.Counts["matching_difference"] != f.Ways["short"] || f.Counts["settlement"] <= bankLines-1 {
		t.Fatalf("differences %d vs short %d, settlements %d of %d lines", f.Counts["matching_difference"], f.Ways["short"], f.Counts["settlement"], bankLines)
	}
}

// Both personas pay in every way over their nine months.
func TestEveryWayOfPaying(t *testing.T) {
	for _, name := range []string{"helios", "nordwind"} {
		f, err := sim.Generate(load(t, name))
		if err != nil {
			t.Fatal(err)
		}
		bankStory(t, f)
		t.Logf("%s pays: %v", name, f.Ways)
	}
}
