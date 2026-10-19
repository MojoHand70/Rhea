package shell_test

import (
	"fmt"
	"testing"

	"rhea/internal/core"
)

// TestFixedAssets (the auditor's exam, 2026-10-19: financial accounting
// 2025, question I-4; EXAM.md 6.1.2 a): a fixed asset accepted for use is
// booked Wn 010 / Ma 080 on the day of acceptance and depreciated from the
// following month at the annual rate given, one charge per month when the
// next month opens, cumulative — a month skipped is caught up at once —
// until the charges equal the initial value, when the asset is depreciated
// and charged no more. A disposal moves the accumulated depreciation off
// 071 and the net value into other operating costs, and stops the
// schedule.
func TestFixedAssets(t *testing.T) {
	w := newCITWorld(t)
	ctx, s, x, door := w.ctx, w.s, w.x, w.door
	monthOpens := func(date string) {
		t.Helper()
		if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: core.EventMonthOpened, OccurredAt: date,
			Payload: []byte(fmt.Sprintf(`{"month":%q,"date":%q}`, date[:7], date)), DedupKey: "time/month/" + date}); err != nil {
			t.Fatal(err)
		}
		if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
			t.Fatalf("month opens %s: %v", date, errs)
		}
	}
	// the books by account and side, in minor units, with the dates of the lines
	postings := func() (sums map[string]int64, dates map[string][]string) {
		t.Helper()
		accounts, _ := s.ObjectsByType(ctx, "account")
		code := map[string]string{}
		for _, a := range accounts {
			code[a.ID] = fmt.Sprint(a.State["code"])
		}
		ps, _ := s.ObjectsByType(ctx, "posting")
		sums, dates = map[string]int64{}, map[string][]string{}
		for _, p := range ps {
			k := code[fmt.Sprint(p.State["account"])] + " " + fmt.Sprint(p.State["side"])
			n, _ := p.State["amount"].(float64)
			sums[k] += int64(n)
			dates[k] = append(dates[k], fmt.Sprint(p.State["date"]))
		}
		return sums, dates
	}
	assets := func() map[string]core.Object {
		t.Helper()
		all, _ := s.ObjectsByType(ctx, "fixed_asset")
		by := map[string]core.Object{}
		for _, a := range all {
			by[fmt.Sprint(a.State["name"])] = a
		}
		return by
	}
	charges := func(assetID string) (n int, amounts []int64) {
		t.Helper()
		all, _ := s.ObjectsByType(ctx, "depreciation_charge")
		for _, c := range all {
			if c.State["asset"] == assetID {
				v, _ := c.State["amount"].(float64)
				if v > 0 {
					n++
					amounts = append(amounts, int64(v))
				}
			}
		}
		return n, amounts
	}

	// A laptop, 12 000 at 20% a year, accepted on 15 September: 200 a month from October.
	door("accept_fixed_asset", map[string]any{"name": "Laptop", "number": "ST/1", "kst": "491", "initial_value": "12000.00",
		"accepted_date": "2026-09-15", "rate": "20", "occurred_at": "2026-09-15"})
	// A printer, 100 at 600% a year (an absurd rate that ends the schedule in two months): 50, then 50.
	door("accept_fixed_asset", map[string]any{"name": "Drukarka", "initial_value": "100.00", "accepted_date": "2026-09-20", "rate": "600", "occurred_at": "2026-09-20"})
	by := assets()
	laptop, printer := by["Laptop"], by["Drukarka"]
	if laptop.State["start"] != "2026-10-01" || laptop.State["status"] != "active" || laptop.State["annual_rate"] == nil {
		t.Fatalf("laptop: %v", laptop.State)
	}
	sums, dates := postings()
	if sums["010 debit"] != 1210000 || sums["080 credit"] != 1210000 || dates["010 debit"][0] != "2026-09-15" {
		t.Fatalf("acceptance postings: %v %v", sums, dates["010 debit"])
	}

	monthOpens("2026-10-01") // September closed: nothing yet, the schedules start in October
	if n, _ := charges(laptop.ID); n != 0 {
		t.Fatalf("no charge before the schedule starts, got %d", n)
	}
	monthOpens("2026-11-01") // October closed: 200 and 50
	monthOpens("2026-12-01") // November closed: 200 and 50 — the printer is now fully depreciated
	if n, amounts := charges(laptop.ID); n != 2 || amounts[0] != 20000 || amounts[1] != 20000 {
		t.Fatalf("laptop charges: %d %v", n, amounts)
	}
	if n, amounts := charges(printer.ID); n != 2 || amounts[0] != 5000 || amounts[1] != 5000 {
		t.Fatalf("printer charges: %d %v", n, amounts)
	}
	if assets()["Drukarka"].State["status"] != "depreciated" {
		t.Fatalf("the printer should be depreciated: %v", assets()["Drukarka"].State)
	}
	sums, dates = postings()
	if sums["401 debit"] != 50000 || sums["071 credit"] != 50000 {
		t.Fatalf("depreciation postings: %v", sums)
	}
	if d := dates["401 debit"]; len(d) != 4 || d[0] != "2026-10-31" {
		t.Fatalf("a charge is dated the last day of its month: %v", d)
	}

	// January's opening is missed; when February opens, December and January are caught up at once: 400.
	monthOpens("2027-02-01")
	if n, amounts := charges(laptop.ID); n != 3 || amounts[2] != 40000 {
		t.Fatalf("catch-up: %d %v", n, amounts)
	}
	if n, _ := charges(printer.ID); n != 2 {
		t.Fatalf("a depreciated asset gets no more charges: %d", n)
	}

	// The laptop is scrapped on 10 February: 800 off 071, 11 200 into 761, 12 000 off 010; no charge for February.
	door("dispose_fixed_asset", map[string]any{"asset": laptop.ID, "date": "2027-02-10", "reason": "liquidation", "occurred_at": "2027-02-10"})
	if a := assets()["Laptop"]; a.State["status"] != "disposed" || a.State["liquidation_date"] != "2027-02-10" || a.State["liquidation_reason"] != "liquidation" {
		t.Fatalf("disposal: %v", a.State)
	}
	sums, _ = postings()
	if sums["071 debit"] != 80000 || sums["761 debit"] != 1120000 || sums["010 credit"] != 1200000 {
		t.Fatalf("disposal postings: %v", sums)
	}
	monthOpens("2027-03-01")
	if n, _ := charges(laptop.ID); n != 3 {
		t.Fatalf("a disposed asset gets no more charges: %d", n)
	}

	// The fully depreciated printer is scrapped: 071 and 010 only, no zero cost line.
	door("dispose_fixed_asset", map[string]any{"asset": printer.ID, "date": "2027-03-05", "reason": "liquidation", "occurred_at": "2027-03-05"})
	sums, _ = postings()
	if sums["071 debit"] != 90000 || sums["010 credit"] != 1210000 || sums["761 debit"] != 1120000 {
		t.Fatalf("disposal of a depreciated asset: %v", sums)
	}
}
