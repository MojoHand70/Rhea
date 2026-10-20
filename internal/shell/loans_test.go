package shell_test

import (
	"fmt"
	"reflect"
	"testing"

	"rhea/internal/core"
)

// TestLoans (the auditor's exam, 2026-10-20, lesson I-15 of financial
// accounting 2025; EXAM.md 6.1.2 d, 6.1.2 a, 6.1.6): a loan received or
// granted enters the books at its nominal value on the day it arises (UoR
// art. 28 ust. 11 pkt 2); repayments bring the amount due down and close
// the loan at zero, a repayment past the amount due books nothing; interest
// is financial cost or revenue (art. 42 ust. 3), paid through the account or
// accrued onto the amount due (art. 28 ust. 1 pkt 7 and 8).
func TestLoans(t *testing.T) {
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
	loan := func(direction string) core.Object {
		t.Helper()
		for _, l := range mustObjects(t, s, ctx, "loan") {
			if l.State["direction"] == direction {
				return l
			}
		}
		t.Fatalf("no %s loan", direction)
		return core.Object{}
	}
	open := func(l core.Object) int64 {
		n, _ := l.State["open"].(float64)
		return int64(n)
	}

	// A loan from a shareholder: 100 000 arrives on 1 March.
	door("receive_loan", map[string]any{"title": "Pożyczka od wspólnika", "date": "2026-03-01", "nominal": "100000.00", "due_date": "2027-03-01",
		"interest_rate": "6", "occurred_at": "2026-03-01"})
	got := loan("received")
	if got.State["status"] != "open" || open(got) != 10000000 {
		t.Fatalf("the loan at its nominal value: %v", got.State)
	}
	after := sums()
	if after["130 debit"] != 10000000 || after["240 credit"] != 10000000 {
		t.Fatalf("a loan received: Wn 130 / Ma 240: %v", after)
	}
	// 40 000 repaid at midsummer: the amount due falls, the loan stays open.
	door("repay_loan", map[string]any{"loan": got.ID, "amount": "40000.00", "date": "2026-06-30", "occurred_at": "2026-06-30"})
	after = sums()
	if after["240 debit"] != 4000000 || after["130 credit"] != 4000000 {
		t.Fatalf("a repayment: Wn 240 / Ma 130: %v", after)
	}
	if got = loan("received"); got.State["status"] != "open" || open(got) != 6000000 {
		t.Fatalf("after the repayment: %v", got.State)
	}
	// Interest for the year accrued at the balance date, not yet paid: a
	// financial cost, and it joins the amount due.
	door("charge_loan_interest", map[string]any{"loan": got.ID, "amount": "2000.00", "date": "2026-12-31", "settlement": "accrued", "occurred_at": "2026-12-31"})
	after = sums()
	if after["751 debit"] != 200000 || after["240 credit"] != 10200000 {
		t.Fatalf("accrued interest: Wn 751 / Ma 240: %v", after)
	}
	if got = loan("received"); open(got) != 6200000 {
		t.Fatalf("interest joins the amount due: %v", got.State)
	}
	// A repayment past the amount due is nobody's rule: the books do not move.
	before := sums()
	x.TriggerActivity(ctx, "repay_loan", map[string]any{"loan": got.ID, "amount": "70000.00", "date": "2027-01-15", "occurred_at": "2027-01-15"}, "krzysztof", "")
	if after = sums(); !reflect.DeepEqual(before, after) {
		t.Fatalf("a repayment past the amount due booked something: %v", after)
	}
	if got = loan("received"); open(got) != 6200000 || got.State["status"] != "open" {
		t.Fatalf("a repayment past the amount due changed the loan: %v", got.State)
	}
	// The rest, interest included, repaid in January: the loan closes.
	door("repay_loan", map[string]any{"loan": got.ID, "amount": "62000.00", "date": "2027-01-15", "occurred_at": "2027-01-15"})
	after = sums()
	if after["240 debit"] != 10200000 || after["240 credit"] != 10200000 {
		t.Fatalf("the loan account settles to zero: %v", after)
	}
	if got = loan("received"); got.State["status"] != "repaid" || open(got) != 0 {
		t.Fatalf("repaid: %v", got.State)
	}

	// A loan granted to a subsidiary: an investment, interest is revenue.
	door("grant_loan", map[string]any{"title": "Pożyczka dla spółki zależnej", "date": "2026-04-01", "nominal": "50000.00", "occurred_at": "2026-04-01"})
	g := loan("granted")
	after = sums()
	if after["030 debit"] != 5000000 || after["130 credit"] != 4000000+6200000+5000000 || open(g) != 5000000 {
		t.Fatalf("a loan granted: Wn 030 / Ma 130: %v %v", after, g.State)
	}
	door("charge_loan_interest", map[string]any{"loan": g.ID, "amount": "1000.00", "date": "2026-12-31", "settlement": "paid", "occurred_at": "2026-12-31"})
	door("charge_loan_interest", map[string]any{"loan": g.ID, "amount": "500.00", "date": "2026-12-31", "settlement": "accrued", "occurred_at": "2026-12-31"})
	after = sums()
	if after["750 credit"] != 150000 || after["130 debit"] != 10000000+100000 || after["030 debit"] != 5050000 {
		t.Fatalf("interest on a loan granted: paid Wn 130 / Ma 750, accrued Wn 030 / Ma 750: %v", after)
	}
	if g = loan("granted"); open(g) != 5050000 {
		t.Fatalf("accrued interest joins the amount due: %v", g.State)
	}
	door("repay_loan", map[string]any{"loan": g.ID, "amount": "50500.00", "date": "2027-03-31", "occurred_at": "2027-03-31"})
	after = sums()
	if after["030 credit"] != 5050000 || after["130 debit"] != 10000000+100000+5050000 {
		t.Fatalf("the borrower repays with interest: Wn 130 / Ma 030: %v", after)
	}
	if g = loan("granted"); g.State["status"] != "repaid" || open(g) != 0 {
		t.Fatalf("the loan granted is repaid: %v", g.State)
	}
}
