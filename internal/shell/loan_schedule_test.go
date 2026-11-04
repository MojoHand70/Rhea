package shell_test

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"rhea/internal/core"
)

// TestLoanAtAmortisedCost (DIRECTION "Optimal decisions", KK 2026-11-04: an
// equation with one answer is a kernel method; EXAM.md 6.1.1 a, 6.1.8 f): a
// loan taken with a fee costs more than its rate says. The office enters the
// contract once — fee, rate, term, instalment, kind, first date — and Rhea
// computes the instalments, the effective interest rate and each period's
// interest at amortised cost (UoR art. 28 ust. 8a). The fee lowers the debt
// instead of being the day's cost; the interest enters the books month by
// month as the clock opens the next one; a schedule entered late gives the
// months that passed in the month it was entered; a second schedule for
// the same loan books nothing; replay reproduces it all.
func TestLoanAtAmortisedCost(t *testing.T) {
	w := newCITWorld(t)
	ctx, s, x, door := w.ctx, w.s, w.x, w.door
	door("declare_company_form", map[string]any{"company": w.self, "vat_status": "active", "books": "full", "income_tax": "cit",
		"from": "2026-01-01", "occurred_at": "2026-01-01"})
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
	monthOpens := func(date string) {
		t.Helper()
		if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: core.EventMonthOpened, OccurredAt: date,
			Payload: []byte(fmt.Sprintf(`{"month":%q,"date":%q}`, date[:7], date)), DedupKey: "time/month/" + date[:7]}); err != nil {
			t.Fatal(err)
		}
		if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
			t.Fatalf("month opens %s: %v", date, errs)
		}
	}
	loan := func(title string) core.Object {
		t.Helper()
		for _, l := range mustObjects(t, s, ctx, "loan") {
			if l.State["title"] == title {
				return l
			}
		}
		t.Fatalf("no loan %q", title)
		return core.Object{}
	}
	money := func(o core.Object, field string) int64 {
		n, _ := o.State[field].(float64)
		return int64(n)
	}
	instalments := func(loanID string) []core.Object {
		t.Helper()
		var out []core.Object
		for _, i := range mustObjects(t, s, ctx, "loan_instalment") {
			if i.State["loan"] == loanID {
				out = append(out, i)
			}
		}
		sort.Slice(out, func(a, b int) bool { return money(out[a], "n") < money(out[b], "n") })
		return out
	}
	interests := func(loanID string) []core.Object {
		t.Helper()
		var out []core.Object
		all, _ := s.ObjectsByType(ctx, "loan_interest")
		for _, i := range all {
			if i.State["loan"] == loanID {
				out = append(out, i)
			}
		}
		sort.Slice(out, func(a, b int) bool { return fmt.Sprint(out[a].State["date"]) < fmt.Sprint(out[b].State["date"]) })
		return out
	}
	row := func(i core.Object) string {
		return fmt.Sprintf("%d %d %d %d %d %d %d %v", money(i, "n"), money(i, "opening"), money(i, "interest"), money(i, "nominal_interest"),
			money(i, "capital"), money(i, "payment"), money(i, "closing"), i.State["period"])
	}

	// A working-capital loan: 100 000 paid out on 2 January at 10%, twelve
	// equal monthly instalments from 2 February, a 2 000 fee kept back.
	door("receive_loan", map[string]any{"title": "Kredyt obrotowy", "date": "2026-01-02", "nominal": "100000.00",
		"interest_rate": "10", "occurred_at": "2026-01-02"})
	k := loan("Kredyt obrotowy")
	door("schedule_loan", map[string]any{"loan": k.ID, "fee": "2000.00", "months": 12, "every": 1, "kind": "annuity",
		"first_instalment": "2026-02-02", "date": "2026-01-02", "occurred_at": "2026-01-02"})
	rows := instalments(k.ID)
	if len(rows) != 12 || row(rows[0]) != "1 9800000 112996 83333 795826 879159 9033837 2026-02" ||
		row(rows[11]) != "12 869134 10022 7266 871890 879156 0 2027-01" {
		t.Fatalf("the schedule: %d rows, first %s, last %s", len(rows), row(rows[0]), row(rows[len(rows)-1]))
	}
	if rows[0].State["rate"] != "1.15301902" || rows[0].State["date"] != "2026-02-02" || rows[11].State["date"] != "2027-01-02" {
		t.Fatalf("the rate a month and the dates: %v", rows[0].State)
	}
	// The loan is measured at amortised cost: 98 000 owed, the fee kept,
	// the effective rate 14.7483% a year beside the contract's 10%, the
	// first interest due in February.
	if k = loan("Kredyt obrotowy"); k.State["measurement"] != "amortised_cost" || money(k, "open") != 9800000 || money(k, "fee") != 200000 ||
		k.State["effective_rate"] != "14.7483" || k.State["kind"] != "annuity" || money(k, "months") != 12 || k.State["next_period"] != "2026-02" {
		t.Fatalf("the loan at amortised cost: %v", k.State)
	}
	// The books: the loan at its nominal, the fee lowering it, no cost yet.
	after := sums()
	if after["130 debit"] != 10000000 || after["240 credit"] != 10000000 || after["240 debit"] != 200000 || after["130 credit"] != 200000 || after["751 debit"] != 0 {
		t.Fatalf("the fee lowers the debt, Wn 240 / Ma 130: %v", after)
	}

	// February opens: no instalment fell in January, nothing accrues.
	monthOpens("2026-02-01")
	if got := interests(k.ID); len(got) != 0 {
		t.Fatalf("January had no instalment: %v", got)
	}
	// March opens: February's instalment accrues its effective interest,
	// 1 129.96 on 98 000, dated the instalment's day: Wn 751 / Ma 240.
	monthOpens("2026-03-01")
	got := interests(k.ID)
	if len(got) != 1 || got[0].State["period"] != "2026-02" || got[0].State["date"] != "2026-02-02" || money(got[0], "amount") != 112996 {
		t.Fatalf("February's interest: %v", got)
	}
	after = sums()
	if after["751 debit"] != 112996 || after["240 credit"] != 10000000+112996 {
		t.Fatalf("effective interest, Wn 751 / Ma 240: %v", after)
	}
	if k = loan("Kredyt obrotowy"); money(k, "open") != 9800000+112996 {
		t.Fatalf("the interest joins the amount owed: %v", k.State)
	}
	if k.State["next_period"] != "2026-03" || money(got[0], "instalments") != 1 {
		t.Fatalf("the first instalment is in the books, March is next: %v", k.State)
	}
	// The bank took the instalment on 2 February: what is owed is the
	// schedule's closing amount for the first period.
	door("repay_loan", map[string]any{"loan": k.ID, "amount": "8791.59", "date": "2026-02-02", "occurred_at": "2026-02-02"})
	if k = loan("Kredyt obrotowy"); money(k, "open") != 9033837 {
		t.Fatalf("after the instalment the loan stands at the schedule's closing amount: %v", k.State)
	}
	// April opens: March's interest on 90 338.37.
	monthOpens("2026-04-01")
	if got = interests(k.ID); len(got) != 2 || money(got[1], "amount") != 104162 || got[1].State["date"] != "2026-03-02" {
		t.Fatalf("March's interest: %v", got)
	}
	// A second schedule for the same loan is nobody's rule: nothing books.
	before := sums()
	tr, err := x.TriggerActivity(ctx, "schedule_loan", map[string]any{"loan": k.ID, "fee": "2000.00", "months": 12, "every": 1, "kind": "annuity",
		"first_instalment": "2026-02-02", "date": "2026-04-02", "occurred_at": "2026-04-02"}, "krzysztof", "")
	if err != nil || tr.Booked != 0 {
		t.Fatalf("a second schedule: %+v %v", tr, err)
	}
	if after = sums(); !reflect.DeepEqual(before, after) || len(instalments(k.ID)) != 12 {
		t.Fatalf("a second schedule changed the books: %v", after)
	}

	// A shareholder's loan paid out on 1 February, 12 000 at 12%, interest
	// monthly and the capital at the end of May, entered only on 15 April:
	// the March and April instalments belong to April and accrue when May
	// opens — March's on the last day of April, April's on its own day.
	door("receive_loan", map[string]any{"title": "Pożyczka od wspólnika", "date": "2026-02-01", "nominal": "12000.00",
		"interest_rate": "12", "occurred_at": "2026-02-01"})
	p := loan("Pożyczka od wspólnika")
	door("schedule_loan", map[string]any{"loan": p.ID, "months": 3, "every": 1, "kind": "bullet",
		"first_instalment": "2026-03-01", "date": "2026-04-15", "occurred_at": "2026-04-15"})
	rows = instalments(p.ID)
	if len(rows) != 3 || row(rows[0]) != "1 1200000 12000 12000 0 12000 1200000 2026-04" ||
		row(rows[1]) != "2 1200000 12000 12000 0 12000 1200000 2026-04" || row(rows[2]) != "3 1200000 12000 12000 1200000 1212000 0 2026-05" {
		t.Fatalf("a late schedule: %v", rows)
	}
	if p = loan("Pożyczka od wspólnika"); money(p, "open") != 1200000 || money(p, "fee") != 0 || p.State["effective_rate"] != "12.6825" || p.State["next_period"] != "2026-04" {
		t.Fatalf("no fee: the contract's rate is the effective rate, 12%% a year compounding monthly: %v", p.State)
	}
	monthOpens("2026-05-01")
	if got = interests(p.ID); len(got) != 1 || got[0].State["date"] != "2026-04-30" || money(got[0], "amount") != 24000 || money(got[0], "instalments") != 2 {
		t.Fatalf("the months that passed accrue in April, on its last day: %v", got)
	}
	if p = loan("Pożyczka od wspólnika"); money(p, "open") != 1224000 || p.State["next_period"] != "2026-05" {
		t.Fatalf("two months of interest owed, May next: %v", p.State)
	}
	// June opens: May's interest on its own day; the schedule is spent.
	monthOpens("2026-06-01")
	if got = interests(p.ID); len(got) != 2 || got[1].State["date"] != "2026-05-01" || money(got[1], "amount") != 12000 {
		t.Fatalf("May's interest: %v", got)
	}
	if p = loan("Pożyczka od wspólnika"); money(p, "open") != 1236000 || p.State["next_period"] != "" {
		t.Fatalf("the last interest: nothing is next: %v", p.State)
	}

	// A loan granted with a fee: 100 000 at 8% on 30 June, the capital in
	// four equal quarterly parts from 30 September, 1 000 kept back. The
	// fee lowers the claim (Wn 130 / Ma 030); the first quarter's effective
	// interest, 2 395.28 on 99 000, is financial revenue (Wn 030 / Ma 750).
	door("grant_loan", map[string]any{"title": "Pożyczka dla spółki zależnej", "date": "2026-06-30", "nominal": "100000.00",
		"interest_rate": "8", "occurred_at": "2026-06-30"})
	g := loan("Pożyczka dla spółki zależnej")
	door("schedule_loan", map[string]any{"loan": g.ID, "fee": "1000.00", "months": 12, "every": 3, "kind": "equal_capital",
		"first_instalment": "2026-09-30", "date": "2026-06-30", "occurred_at": "2026-06-30"})
	if rows = instalments(g.ID); len(rows) != 4 || row(rows[0]) != "1 9900000 239528 200000 2500000 2700000 7439528 2026-09" {
		t.Fatalf("the quarterly schedule: %v", rows)
	}
	after = sums()
	if after["130 debit"] != 10000000+1200000+100000 || after["030 credit"] != 100000 || after["030 debit"] != 10000000 {
		t.Fatalf("the fee lowers the claim, Wn 130 / Ma 030: %v", after)
	}
	for _, d := range []string{"2026-07-01", "2026-08-01", "2026-09-01", "2026-10-01"} {
		monthOpens(d)
	}
	after = sums()
	if after["750 credit"] != 239528 || after["030 debit"] != 10000000+239528 {
		t.Fatalf("the quarter's effective interest is revenue, Wn 030 / Ma 750: %v", after)
	}
	if g = loan("Pożyczka dla spółki zależnej"); money(g, "open") != 9900000+239528 {
		t.Fatalf("the claim at amortised cost: %v", g.State)
	}

	// Replay reproduces every row, every posting and every amount owed.
	all, _ := s.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	again, _ := s.AllObjects(ctx)
	if !reflect.DeepEqual(all, again) {
		t.Fatal("replay does not reproduce the schedules")
	}
}
