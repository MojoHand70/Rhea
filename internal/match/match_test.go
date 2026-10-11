package match

import (
	"math/big"
	"reflect"
	"strings"
	"testing"
)

func rat(s string) *big.Rat {
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		panic(s)
	}
	return r
}

func money(name string) Measure { return Measure{Name: name, Allocated: true, Money: true, Places: 2} }

func item(id, date, open, ref string, seq int) Item {
	return Item{ID: id, Open: map[string]*big.Rat{"money": rat(open)}, Currency: "PLN", Date: date, Reference: ref, Seq: seq}
}

func invoices() Side {
	return Side{Name: "invoice", Items: []Item{
		item("FV/1", "2026-09-02", "100.00", "FV/1", 1),
		item("FV/2", "2026-09-03", "250.00", "FV/2", 2),
		item("FV/3", "2026-09-04", "300.00", "FV/3", 3),
		item("FV/4", "2026-09-05", "400.00", "FV/4", 4),
		item("FV/12", "2026-09-06", "100.00", "FV/12", 5),
	}}
}

func bank(open, ref string, policy ...Step) Problem {
	return Problem{TriggerSide: "payment", Trigger: item("L", "2026-09-20", open, ref, 9),
		Sides: []Side{invoices()}, Measures: []Measure{money("money")}, Policy: policy}
}

func ids(allocs []Allocation) []string {
	var out []string
	for _, a := range allocs {
		out = append(out, a.ID)
	}
	return out
}

func amount(allocs []Allocation, id string) (allocated, after string) {
	for _, a := range allocs {
		if a.ID == id {
			return a.Allocated["money"].FloatString(2), a.OpenAfter["money"].FloatString(2)
		}
	}
	return "", ""
}

func solve(t *testing.T, p Problem) (Result, bool) {
	t.Helper()
	r, ok, err := Solve(p)
	if err != nil {
		t.Fatal(err)
	}
	return r, ok
}

func TestReferenceNamesWholeTokens(t *testing.T) {
	r, ok := solve(t, bank("100.00", "zapłata za FV/1 dziękujemy", Step{Strategy: Reference}))
	if !ok || !reflect.DeepEqual(ids(r.Sides["invoice"]), []string{"FV/1"}) {
		t.Fatalf("want FV/1 by reference, got %v %v", ok, r)
	}
	// FV/1 is not named by "FV/12"; FV/12 is.
	r, ok = solve(t, bank("100.00", "FV/12", Step{Strategy: Reference}))
	if !ok || !reflect.DeepEqual(ids(r.Sides["invoice"]), []string{"FV/12"}) {
		t.Fatalf("want FV/12, got %v %v", ok, r)
	}
	// Named but the amount disagrees: strict reference declines, with the numbers.
	r, ok = solve(t, bank("60.00", "FV/1", Step{Strategy: Reference}))
	if ok || !strings.Contains(r.Tried[0].Reason, `"FV/1" names 100.00 on invoice, the payment brings 60.00`) {
		t.Fatalf("strict reference on a short payment: %v %v", ok, r.Tried)
	}
	// Several named: the sum must agree.
	r, ok = solve(t, bank("350.00", "FV/1 FV/2", Step{Strategy: Reference}))
	if !ok || !reflect.DeepEqual(ids(r.Sides["invoice"]), []string{"FV/1", "FV/2"}) {
		t.Fatalf("two named invoices: %v %v", ok, r)
	}
}

func TestReferencePartial(t *testing.T) {
	// A short payment settles part of the named invoice; the rest stays open.
	r, ok := solve(t, bank("60.00", "FV/1", Step{Strategy: Reference, Partial: true}))
	if !ok {
		t.Fatal(r.Tried)
	}
	if got, after := amount(r.Sides["invoice"], "FV/1"); got != "60.00" || after != "40.00" {
		t.Fatalf("FV/1 after 60.00: allocated %s, open %s", got, after)
	}
	if got, after := amount(r.Sides["payment"], "L"); got != "60.00" || after != "0.00" {
		t.Fatalf("payment: allocated %s, open %s", got, after)
	}
	// An over-payment settles the named invoice whole and keeps the rest open on the line.
	r, _ = solve(t, bank("130.00", "FV/1", Step{Strategy: Reference, Partial: true}))
	if got, after := amount(r.Sides["payment"], "L"); got != "100.00" || after != "30.00" {
		t.Fatalf("over-payment: allocated %s, open %s", got, after)
	}
	if len(r.Differences) != 0 {
		t.Fatalf("partial allocation absorbs nothing, got %v", r.Differences)
	}
}

func TestExactPrefersTheOldest(t *testing.T) {
	r, ok := solve(t, bank("100.00", "", Step{Strategy: Reference}, Step{Strategy: Exact}))
	if !ok || r.Step != 2 || r.Strategy != Exact {
		t.Fatalf("want exact at step 2, got %v %v", ok, r)
	}
	if !reflect.DeepEqual(ids(r.Sides["invoice"]), []string{"FV/1"}) {
		t.Fatalf("two invoices of 100.00: the oldest wins, got %v", ids(r.Sides["invoice"]))
	}
	if r.Tried[0].Reason != "the payment names no reference" {
		t.Fatalf("the reference step explains itself: %q", r.Tried[0].Reason)
	}
	// Nothing of 150.00: the reason counts the candidates.
	r, ok = solve(t, bank("150.00", "", Step{Strategy: Exact}))
	if ok || !strings.Contains(r.Tried[0].Reason, "no single item on invoice opens at 150.00 (5 candidates)") {
		t.Fatalf("want the explanation, got %v %q", ok, r.Tried[0].Reason)
	}
}

func TestExactExcludesWhatCannotMatch(t *testing.T) {
	p := bank("100.00", "", Step{Strategy: Exact, WithinDays: 10})
	p.Sides[0].Items[0].Currency = "EUR"            // FV/1 in another currency
	p.Sides[0].Items[4].Open["money"] = rat("0.00") // FV/12 settled
	p.Sides[0].Items[2].Date = "2026-08-01"         // FV/3 too old for the window
	r, ok := solve(t, p)
	if ok {
		t.Fatalf("nothing should match: %v", r)
	}
	reason := r.Tried[0].Reason
	for _, want := range []string{"FV/1: in EUR, the payment in PLN", "FV/12: nothing open", "FV/3: 50 days from the payment, the step allows 10"} {
		if !strings.Contains(reason, want) {
			t.Fatalf("reason %q lacks %q", reason, want)
		}
	}
}

func TestSubsetIsDeterministicAndCapped(t *testing.T) {
	policy := []Step{{Strategy: Exact}, {Strategy: Subset}}
	r, ok := solve(t, bank("750.00", "", policy...))
	if !ok || r.Strategy != Subset {
		t.Fatalf("want a subset, got %v %v", ok, r)
	}
	// 750 = 100 + 250 + 400, found including the oldest first; not 350 + 400.
	if !reflect.DeepEqual(ids(r.Sides["invoice"]), []string{"FV/1", "FV/2", "FV/4"}) {
		t.Fatalf("got %v", ids(r.Sides["invoice"]))
	}
	again, _ := solve(t, bank("750.00", "", policy...))
	if !reflect.DeepEqual(ids(again.Sides["invoice"]), ids(r.Sides["invoice"])) {
		t.Fatal("the same problem must yield the same subset")
	}
	// Three items needed, two allowed.
	r, ok = solve(t, bank("750.00", "", Step{Strategy: Subset, MaxItems: 2}))
	if ok || !strings.Contains(r.Tried[0].Reason, "no subset of up to 2 items on invoice opens at 750.00") {
		t.Fatalf("want the size cap, got %v %q", ok, r.Tried[0].Reason)
	}
	// Above the node budget: to the person, never a long search.
	p := bank("750.00", "", Step{Strategy: Subset})
	p.Budget = 2
	r, ok = solve(t, p)
	if ok || !strings.Contains(r.Tried[0].Reason, "above its cap") {
		t.Fatalf("want the budget refusal, got %v %q", ok, r.Tried[0].Reason)
	}
	// Nothing sums to 751.
	r, ok = solve(t, bank("751.00", "", Step{Strategy: Subset}))
	if ok {
		t.Fatalf("751.00 has no subset, got %v", r)
	}
}

func TestOldestFirstLeavesTheRemainderOpen(t *testing.T) {
	side := Side{Name: "invoice", Items: []Item{
		item("FV/9", "2026-09-02", "500.00", "", 1), item("FV/10", "2026-09-03", "600.00", "", 2)}}
	p := Problem{TriggerSide: "payment", Trigger: item("L", "2026-09-20", "700.00", "", 9),
		Sides: []Side{side}, Measures: []Measure{money("money")},
		Policy: []Step{{Strategy: Exact}, {Strategy: Subset}, {Strategy: OldestFirst}}}
	r, ok := solve(t, p)
	if !ok || r.Strategy != OldestFirst || r.Step != 3 {
		t.Fatalf("want oldest first at step 3, got %v %v", ok, r)
	}
	if got, after := amount(r.Sides["invoice"], "FV/9"); got != "500.00" || after != "0.00" {
		t.Fatalf("FV/9: %s open %s", got, after)
	}
	if got, after := amount(r.Sides["invoice"], "FV/10"); got != "200.00" || after != "400.00" {
		t.Fatalf("FV/10: %s open %s", got, after)
	}
	if got, after := amount(r.Sides["payment"], "L"); got != "700.00" || after != "0.00" {
		t.Fatalf("payment: %s open %s", got, after)
	}
	// More than everything open: the line keeps the rest.
	p.Trigger = item("L", "2026-09-20", "2000.00", "", 9)
	r, _ = solve(t, p)
	if got, after := amount(r.Sides["payment"], "L"); got != "1100.00" || after != "900.00" {
		t.Fatalf("payment above the open total: %s open %s", got, after)
	}
}

func TestToleranceAbsorbsAndRecords(t *testing.T) {
	strict := Step{Strategy: Exact}
	loose := Step{Strategy: Exact, Tolerance: map[string]*big.Rat{"money": rat("0.02")}, DifferenceAccount: "755"}
	r, ok := solve(t, bank("99.98", "", strict, loose))
	if !ok || r.Step != 2 {
		t.Fatalf("99.98 against 100.00 within two grosz: %v %v", ok, r)
	}
	if got, after := amount(r.Sides["invoice"], "FV/1"); got != "100.00" || after != "0.00" {
		t.Fatalf("the invoice settles whole: %s open %s", got, after)
	}
	if len(r.Differences) != 1 {
		t.Fatalf("want one difference, got %v", r.Differences)
	}
	d := r.Differences[0]
	if d.Kind != "tolerance" || d.Amount.FloatString(2) != "0.02" || d.Direction != "short" || d.Account != "755" || d.Side != "invoice" {
		t.Fatalf("difference %+v", d)
	}
	r, _ = solve(t, bank("100.01", "", loose))
	if r.Differences[0].Direction != "over" || r.Differences[0].Amount.FloatString(2) != "0.01" {
		t.Fatalf("over-payment difference %+v", r.Differences[0])
	}
	if _, ok := solve(t, bank("99.97", "", loose)); ok {
		t.Fatal("three grosz is outside two")
	}
	// A relative bound: one percent of 100.00 admits 99.50.
	pct := Step{Strategy: Exact, TolerancePercent: map[string]*big.Rat{"money": rat("1")}, DifferenceAccount: "755"}
	if _, ok := solve(t, bank("99.50", "", pct)); !ok {
		t.Fatal("99.50 is within one percent of 100.00")
	}
}

func TestThreeSides(t *testing.T) {
	qty := Measure{Name: "qty", Allocated: true, Places: 0}
	price := Measure{Name: "price", Places: 2}
	line := func(id, date, q, p string, seq int) Item {
		it := Item{ID: id, Open: map[string]*big.Rat{"qty": rat(q)}, Date: date, Seq: seq}
		if p != "" {
			it.Compare = map[string]*big.Rat{"price": rat(p)}
		}
		return it
	}
	p := Problem{TriggerSide: "invoice", Trigger: line("INV-L1", "2026-09-20", "100", "10.20", 9),
		Sides: []Side{
			{Name: "order", Items: []Item{line("PO-L1", "2026-09-01", "100", "10.00", 1), line("PO-L2", "2026-09-01", "100", "12.00", 2)}},
			{Name: "receipt", Items: []Item{line("PZ-L1", "2026-09-05", "60", "", 3), line("PZ-L2", "2026-09-08", "40", "", 4)}},
		},
		Measures: []Measure{qty, price},
		Policy: []Step{
			{Strategy: Exact},
			{Strategy: Subset, TolerancePercent: map[string]*big.Rat{"price": rat("5")}},
		}}
	r, ok := solve(t, p)
	if !ok || r.Strategy != Subset {
		t.Fatalf("want the subset step, got %v %v", ok, r)
	}
	// The order line within five percent of the invoice's price, not the one at 12.00;
	// the two receipts summing to the quantity.
	if !reflect.DeepEqual(ids(r.Sides["order"]), []string{"PO-L1"}) || !reflect.DeepEqual(ids(r.Sides["receipt"]), []string{"PZ-L1", "PZ-L2"}) {
		t.Fatalf("order %v receipt %v", ids(r.Sides["order"]), ids(r.Sides["receipt"]))
	}
	if len(r.Differences) != 1 || r.Differences[0].Measure != "price" || r.Differences[0].Item != "PO-L1" || r.Differences[0].Amount.FloatString(2) != "0.20" || r.Differences[0].Direction != "over" {
		t.Fatalf("the price gap is recorded on the order line: %+v", r.Differences)
	}
	// The strict step explains why the 12.00 line was left out.
	if !strings.Contains(r.Tried[0].Reason, "PO-L1: price 10.00, the invoice says 10.20") {
		t.Fatalf("strict reason: %q", r.Tried[0].Reason)
	}
}

func TestProblemsAreValidated(t *testing.T) {
	p := bank("100.00", "", Step{Strategy: "guess"})
	if _, _, err := Solve(p); err == nil || !strings.Contains(err.Error(), `unknown strategy "guess"`) {
		t.Fatalf("want the strategy refusal, got %v", err)
	}
	p = bank("100.00", "", Step{Strategy: OldestFirst})
	p.Measures = append(p.Measures, Measure{Name: "qty", Allocated: true})
	p.Trigger.Open["qty"] = rat("1")
	if _, _, err := Solve(p); err == nil || !strings.Contains(err.Error(), "oldest_first allocates one measure") {
		t.Fatalf("want the one-measure law, got %v", err)
	}
	p = bank("100.00", "", Step{Strategy: Subset, MaxItems: 9})
	if _, _, err := Solve(p); err == nil || !strings.Contains(err.Error(), "max_items 9 is above the cap of 8") {
		t.Fatalf("want the size cap, got %v", err)
	}
	// A settled trigger has nothing to match.
	r, ok := solve(t, bank("0.00", "", Step{Strategy: Exact}))
	if ok || r.Tried[0].Reason != "the payment has nothing open" {
		t.Fatalf("got %v %v", ok, r.Tried)
	}
}
