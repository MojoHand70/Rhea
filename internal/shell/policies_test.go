package shell_test

import (
	"fmt"
	"reflect"
	"testing"

	"rhea/internal/core"
)

// TestPoliciesAndResult (the auditor's exam, 2026-10-20, lessons I-20, I-23
// and I-1 of financial accounting 2025; EXAM.md 6.1.7 c, 6.1.8 g, 6.1.1 b):
// finished development work is capitalised against the clearing account and
// written off over its life (UoR art. 33 ust. 2–3), the clearing account's
// balance reaches the year's result, and while the work is not written off
// the profit that may go to dividends is capped (KSH art. 191 § 4); a change
// of accounting policy corrects the opening balance through 820 (art. 8 ust.
// 2, KSR 7 pkt 3.12); without going concern an asset is measured at its net
// selling price not above its carrying value, with a provision for the costs
// of ceasing, both against the revaluation capital (art. 29).
func TestPoliciesAndResult(t *testing.T) {
	w := newCITWorld(t)
	ctx, s, x, door, submit := w.ctx, w.s, w.x, w.door, w.submit
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
	delta := func(before, after map[string]int64) map[string]int64 {
		d := map[string]int64{}
		for k, v := range after {
			if v-before[k] != 0 {
				d[k] = v - before[k]
			}
		}
		return d
	}
	money := func(o core.Object, f string) int64 {
		n, _ := o.State[f].(float64)
		return int64(n)
	}

	// Development work finished in March: 60 000 of the year's costs become an
	// intangible; 12 000 written off at the year end.
	before := sums()
	door("capitalise_development_costs", map[string]any{"title": "Receptura R-7", "amount": "60000.00", "date": "2026-03-31", "useful_life_years": 5, "occurred_at": "2026-03-31"})
	dev := mustObjects(t, s, ctx, "development_costs")[0]
	if dev.State["status"] != "active" || money(dev, "unamortised") != 6000000 {
		t.Fatalf("capitalised: %v", dev.State)
	}
	if got := delta(before, sums()); !reflect.DeepEqual(got, map[string]int64{"020 debit": 6000000, "490 credit": 6000000}) {
		t.Fatalf("capitalisation: Wn 020 / Ma 490: %v", got)
	}
	door("amortise_development_costs", map[string]any{"development_costs": dev.ID, "amount": "12000.00", "date": "2026-12-31", "occurred_at": "2026-12-31"})
	if dev = mustObjects(t, s, ctx, "development_costs")[0]; money(dev, "unamortised") != 4800000 || dev.State["status"] != "active" {
		t.Fatalf("after the write-off: %v", dev.State)
	}
	after := sums()
	if after["401 debit"] != 1200000 || after["075 credit"] != 1200000 {
		t.Fatalf("the write-off: Wn 401 / Ma 075: %v", after)
	}
	// A sale of 100 000, the year closed and the statement approved: the
	// result is 100 000 + 60 000 from the clearing account − 12 000.
	submit("sales.invoice.issued", "2026-06-30", "s-1", map[string]any{"number": "FV 1/2026", "issue_date": "2026-06-30", "market": "pl", "intragroup": "no",
		"seller_nip": "5250001111", "buyer_nip": "5260001246", "net": "100000.00", "vat": "23000.00", "gross": "123000.00", "vat_rate": "23", "currency": "PLN"})
	door("close_year", map[string]any{"year_end": "2026-12-31", "book": "pl-stat", "occurred_at": "2026-12-31"})
	door("approve_statement", map[string]any{"year_end": "2026-12-31", "approved_on": "2027-03-31", "book": "pl-stat", "occurred_at": "2027-03-31"})
	if after = sums(); after["820 credit"] != 14800000 {
		t.Fatalf("the profit on 820 with the clearing account closed: %v", after)
	}
	// 48 000 of the work is not written off and there is no reserve capital:
	// at most 100 000 of the 148 000 may go to dividends. 120 000 is refused.
	resolution := map[string]any{"year_end": "2026-12-31", "resolution_date": "2027-04-15", "to_reserve": "0.00", "to_dividends": "120000.00", "loss_cover": "0.00", "book": "pl-stat", "occurred_at": "2027-04-15"}
	before = sums()
	x.TriggerActivity(ctx, "distribute_result", resolution, "krzysztof", "")
	if got := delta(before, sums()); len(got) != 0 || len(mustObjects(t, s, ctx, "result_distribution")) != 0 {
		t.Fatalf("a distribution past the cap booked something: %v", got)
	}
	resolution["to_dividends"] = "100000.00"
	door("distribute_result", resolution)
	if got := delta(before, sums()); !reflect.DeepEqual(got, map[string]int64{"820 debit": 10000000, "241 credit": 10000000}) {
		t.Fatalf("the distribution within the cap: %v", got)
	}
	// A change of policy from 1 January: goods under the new method are
	// worth 5 000 more — the opening balance, through retained earnings.
	before = sums()
	door("adjust_for_policy_change", map[string]any{"title": "Wycena rozchodu towarów: z LIFO na średnią ważoną", "account": "330", "side": "debit", "amount": "5000.00",
		"effective_from": "2027-01-01", "basis": "art. 8 ust. 2 UoR", "occurred_at": "2027-01-01"})
	if got := delta(before, sums()); !reflect.DeepEqual(got, map[string]int64{"330 debit": 500000, "820 credit": 500000}) {
		t.Fatalf("the policy change: Wn 330 / Ma 820: %v", got)
	}
	// Liquidation decided: the stock is worth 15 000 to sell, 20 000 in the
	// books — the difference and the provision for ceasing go to 813.
	before = sums()
	door("measure_without_going_concern", map[string]any{"item": "Zapas towarów", "account": "330", "carrying_value": "20000.00", "net_selling_price": "15000.00",
		"date": "2027-06-30", "reason": "liquidation", "occurred_at": "2027-06-30"})
	if got := delta(before, sums()); !reflect.DeepEqual(got, map[string]int64{"813 debit": 500000, "330 credit": 500000}) {
		t.Fatalf("measurement without going concern: Wn 813 / Ma 330: %v", got)
	}
	if lm := mustObjects(t, s, ctx, "liquidation_measurement"); len(lm) != 1 || money(lm[0], "write_down") != 500000 || lm[0].State["reason"] != "liquidation" {
		t.Fatalf("the measurement: %v", lm)
	}
	before = sums()
	x.TriggerActivity(ctx, "measure_without_going_concern", map[string]any{"item": "Maszyna", "account": "075", "carrying_value": "10000.00", "net_selling_price": "12000.00",
		"date": "2027-06-30", "reason": "liquidation", "occurred_at": "2027-06-30"}, "krzysztof", "")
	if got := delta(before, sums()); len(got) != 0 || len(mustObjects(t, s, ctx, "liquidation_measurement")) != 1 {
		t.Fatalf("a price above the carrying value changed something: %v", got)
	}
	door("provide_for_cessation", map[string]any{"title": "Koszty likwidacji", "amount": "8000.00", "date": "2027-06-30", "occurred_at": "2027-06-30"})
	if got := delta(before, sums()); !reflect.DeepEqual(got, map[string]int64{"813 debit": 800000, "830 credit": 800000}) {
		t.Fatalf("the provision for ceasing: Wn 813 / Ma 830: %v", got)
	}
}
