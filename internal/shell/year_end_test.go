package shell_test

import (
	"fmt"
	"testing"

	"rhea/internal/core"
)

// TestYearEndAndResult (the auditor's exam, 2026-10-19; EXAM.md 6.1.7, 6.1.2
// c, 9.1.3): a year's costs and revenues close into 860 on the balance sheet
// date; the statement's approval moves the result to 820 on the day of
// approval; the shareholders' resolution sends part of it to reserve
// capital and part to dividends; a dividend is paid net of the 19%
// withholding, which the company owes on 225; what is left stays on 820 as
// undistributed. A dividend at a 0% rate has no tax line.
func TestYearEndAndResult(t *testing.T) {
	w := newCITWorld(t)
	ctx, s, door := w.ctx, w.s, w.door
	door("declare_company_form", map[string]any{"company": w.self, "vat_status": "active", "books": "full", "income_tax": "cit",
		"from": "2026-01-01", "occurred_at": "2026-08-01"})
	w.sale("fv-1", "FV 1/09/2026", "2026-09-21", "1000.00", "230.00", "1230.00")
	w.sale("fv-2", "FV 2/09/2026", "2026-09-24", "500.00", "115.00", "615.00")
	w.buy("srv", "SRV/2026", "2026-09-16", "5260001246", "50.01", "11.50", "61.51")
	w.buy("ubz", "UBZ/2026", "2026-08-10", "5260001246", "600.00", "138.00", "738.00")
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
	before := sums()
	if before["700 credit"] != 150000 || before["403 debit"] != 65001 {
		t.Fatalf("the year's books: %v", before)
	}

	// The close: revenues 1 500 and costs 650.01 into 860 — a profit of 849.99.
	door("close_year", map[string]any{"year_end": "2026-12-31", "book": "pl-stat", "occurred_at": "2027-03-15"})
	after := sums()
	if after["700 debit"] != 150000 || after["860 credit"] != 150000 || after["403 credit"] != 65001 || after["860 debit"] != 65001 {
		t.Fatalf("the close: %v", after)
	}
	var lines []core.Object
	for _, typ := range []string{"year_close_cost", "year_close_revenue", "year_close_tax"} {
		of, _ := s.ObjectsByType(ctx, typ)
		lines = append(lines, of...)
	}
	if len(lines) == 0 {
		t.Fatal("no close lines")
	}
	for _, l := range lines {
		if l.State["year_end"] != "2026-12-31" {
			t.Fatalf("a close line is dated the balance sheet date: %v", l.State)
		}
	}

	// Approval on 30 April moves the profit to 820.
	door("approve_statement", map[string]any{"year_end": "2026-12-31", "approved_on": "2027-04-30", "book": "pl-stat", "occurred_at": "2027-04-30"})
	after = sums()
	if after["860 debit"] != 65001+84999 || after["820 credit"] != 84999 {
		t.Fatalf("approval: %v", after)
	}

	// The resolution: 68 to reserve capital (8%), 750 to dividends, nothing to cover; 31.99 stays undistributed.
	door("distribute_result", map[string]any{"year_end": "2026-12-31", "resolution_date": "2027-04-30", "to_reserve": "68.00", "to_dividends": "750.00",
		"loss_cover": "0.00", "book": "pl-stat", "occurred_at": "2027-04-30"})
	after = sums()
	if after["820 debit"] != 81800 || after["802 credit"] != 6800 || after["241 credit"] != 75000 {
		t.Fatalf("distribution: %v", after)
	}
	if after["820 credit"]-after["820 debit"] != 3199 {
		t.Fatalf("undistributed result: %d", after["820 credit"]-after["820 debit"])
	}

	// 700 paid to a person at 19%: 567 to the bank, 133 withheld; 50 to an exempt company: no tax line.
	door("pay_dividend", map[string]any{"date": "2027-05-10", "gross": "700.00", "withholding_rate": "19", "recipient": "Jan Kowalski", "occurred_at": "2027-05-10"})
	door("pay_dividend", map[string]any{"date": "2027-05-12", "gross": "50.00", "withholding_rate": "0", "recipient": "Helios sp. z o.o.", "occurred_at": "2027-05-12"})
	after = sums()
	if after["241 debit"] != 75000 || after["130 credit"] != 56700+5000 || after["225 credit"] != 13300 {
		t.Fatalf("dividends paid: %v", after)
	}
	dists, _ := s.ObjectsByType(ctx, "result_distribution")
	if len(dists) != 1 || dists[0].State["resolution_date"] != "2027-04-30" {
		t.Fatalf("distribution object: %+v", dists)
	}
}
