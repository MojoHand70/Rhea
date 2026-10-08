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
	want := 3*f.Counts["sales_invoice"] + 3*f.Counts["purchase_invoice"] + 2*lines
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
	paid := 0
	for _, w := range f.Where {
		if w.Type == "sales_invoice" && w.Value == "paid" {
			paid = w.Count
		}
	}
	if paid != kinds["bank.statement.line"] || paid == 0 {
		t.Fatalf("paid %d, bank lines %d", paid, kinds["bank.statement.line"])
	}
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
}
