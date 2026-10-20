package shell_test

import (
	"fmt"
	"reflect"
	"testing"

	"rhea/internal/core"
)

// TestDeferredTax (the auditor's exam, 2026-10-20, lesson I-22 of financial
// accounting 2025; EXAM.md 6.1.8 d): a temporary difference between the
// carrying value and the tax value of an item is a deferred tax — an asset
// worth more in the books than for tax gives a provision (UoR art. 37 ust.
// 5), a liability worth more in the books than for tax, or a tax loss to
// deduct, gives a deferred tax asset (ust. 4), at the rate of the year the
// obligation arises (ust. 6); the reversal books the opposite (ust. 8); equal
// values are no difference and book nothing.
func TestDeferredTax(t *testing.T) {
	w := newCITWorld(t)
	ctx, s, x, door := w.ctx, w.s, w.x, w.door
	door("declare_company_form", map[string]any{"company": w.self, "vat_status": "active", "books": "full", "income_tax": "cit",
		"from": "2026-01-01", "occurred_at": "2026-08-01"})
	sums := func() map[string]int64 {
		t.Helper()
		accounts, _ := s.ObjectsByType(ctx, "account")
		code := map[string]string{}
		for _, a := range accounts {
			code[a.ID] = fmt.Sprint(a.State["code"])
		}
		ps, _ := s.ObjectsByType(ctx, "posting")
		out := map[string]int64{}
		for _, p := range ps {
			n, _ := p.State["amount"].(float64)
			out[code[fmt.Sprint(p.State["account"])]+" "+fmt.Sprint(p.State["side"])] += int64(n)
		}
		return out
	}
	byTitle := func(title string) core.Object {
		t.Helper()
		for _, d := range mustObjects(t, s, ctx, "deferred_tax") {
			if d.State["title"] == title {
				return d
			}
		}
		t.Fatalf("no deferred tax %q", title)
		return core.Object{}
	}
	money := func(o core.Object, f string) int64 {
		n, _ := o.State[f].(float64)
		return int64(n)
	}

	// The production line: 500 000 in the books, 420 000 for tax — a
	// positive difference of 80 000, a provision of 15 200 at 19%.
	door("recognise_deferred_tax", map[string]any{"title": "Linia produkcyjna", "item": "Linia produkcyjna L-1", "side": "asset",
		"carrying_value": "500000.00", "tax_value": "420000.00", "rate": "19", "date": "2026-12-31", "occurred_at": "2026-12-31"})
	line := byTitle("Linia produkcyjna")
	if line.State["kind"] != "liability" || money(line, "difference") != 8000000 || money(line, "amount") != 1520000 || line.State["status"] != "open" {
		t.Fatalf("a positive difference is a provision: %v", line.State)
	}
	after := sums()
	if after["870 debit"] != 1520000 || after["835 credit"] != 1520000 {
		t.Fatalf("the provision: Wn 870 / Ma 835: %v", after)
	}
	// A provision for a lawsuit: 30 000 in the books, nothing for tax until
	// paid — a liability worth more in the books is a negative difference.
	door("recognise_deferred_tax", map[string]any{"title": "Rezerwa na proces", "item": "Rezerwa na sprawę sądową", "side": "liability",
		"carrying_value": "30000.00", "tax_value": "0.00", "rate": "19", "date": "2026-12-31", "occurred_at": "2026-12-31"})
	suit := byTitle("Rezerwa na proces")
	if suit.State["kind"] != "asset" || money(suit, "amount") != 570000 {
		t.Fatalf("a liability above its tax value is a deferred tax asset: %v", suit.State)
	}
	// A tax loss of 100 000 to deduct in the years to come.
	door("recognise_deferred_tax", map[string]any{"title": "Strata podatkowa 2026", "item": "Strata podatkowa za 2026", "side": "asset",
		"carrying_value": "0.00", "tax_value": "100000.00", "rate": "19", "date": "2026-12-31", "occurred_at": "2026-12-31"})
	loss := byTitle("Strata podatkowa 2026")
	if loss.State["kind"] != "asset" || money(loss, "amount") != 1900000 {
		t.Fatalf("a tax loss is a deferred tax asset: %v", loss.State)
	}
	after = sums()
	if after["650 debit"] != 570000+1900000 || after["870 credit"] != 570000+1900000 {
		t.Fatalf("the assets: Wn 650 / Ma 870: %v", after)
	}
	// Equal values: no difference, nothing in the books, no object.
	before := sums()
	x.TriggerActivity(ctx, "recognise_deferred_tax", map[string]any{"title": "Grunt", "item": "Grunt", "side": "asset",
		"carrying_value": "1000.00", "tax_value": "1000.00", "rate": "19", "date": "2026-12-31", "occurred_at": "2026-12-31"}, "krzysztof", "")
	if after = sums(); !reflect.DeepEqual(before, after) || len(mustObjects(t, s, ctx, "deferred_tax")) != 3 {
		t.Fatalf("equal values booked something: %v", after)
	}
	// The line is sold the next year: the difference reverses.
	door("reverse_deferred_tax", map[string]any{"deferred_tax": line.ID, "date": "2027-06-30", "occurred_at": "2027-06-30"})
	after = sums()
	if after["835 debit"] != 1520000 || after["870 credit"] != 570000+1900000+1520000 {
		t.Fatalf("the reversal: Wn 835 / Ma 870: %v", after)
	}
	if line = byTitle("Linia produkcyjna"); line.State["status"] != "reversed" || line.State["reversed_on"] != "2027-06-30" {
		t.Fatalf("reversed: %v", line.State)
	}
	// The loss deducted: the asset leaves the books the other way.
	door("reverse_deferred_tax", map[string]any{"deferred_tax": loss.ID, "date": "2027-12-31", "occurred_at": "2027-12-31"})
	after = sums()
	if after["870 debit"] != 1520000+1900000 || after["650 credit"] != 1900000 {
		t.Fatalf("the asset's reversal: Wn 870 / Ma 650: %v", after)
	}
}
