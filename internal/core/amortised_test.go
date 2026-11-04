package core

import (
	"fmt"
	"math/big"
	"strings"
	"testing"
)

// The first equation (DIRECTION "Optimal decisions", KK 2026-11-04): a loan
// with a fee costs more than its rate says, and the accounting act values
// it at amortised cost with the effective interest rate (UoR art. 28 ust.
// 8a). The goldens come from an independent computation in exact fractions.

func row(r AmortisedRow) string {
	return fmt.Sprintf("%d %d %d %d %d %d %d", r.N, r.Opening, r.Interest, r.Nominal, r.Capital, r.Payment, r.Closing)
}

func TestAmortisedCostSolvesTheEffectiveRate(t *testing.T) {
	// 100 000 for three years at 10%, interest yearly and the capital at the
	// end, a 2 000 fee kept back: 98 000 received, 130 000 paid back. The
	// rate that discounts the three payments to 98 000 is 10.8158% a year.
	a, err := AmortisedCost(10000000, 200000, big.NewRat(10, 1), 36, 12, "bullet", "2027-01-02")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"1 9800000 1059949 1000000 0 1000000 9859949",
		"2 9859949 1066433 1000000 0 1000000 9926382",
		"3 9926382 1073618 1000000 10000000 11000000 0",
	}
	for i, r := range a.Rows {
		if row(r) != want[i] {
			t.Fatalf("row %d: %s, want %s", i+1, row(r), want[i])
		}
	}
	if a.PeriodPct().FloatString(8) != "10.81580553" || a.AnnualPct().FloatString(4) != "10.8158" {
		t.Fatalf("the effective rate: %s a period, %s a year", a.PeriodPct().FloatString(8), a.AnnualPct().FloatString(4))
	}
	var interest int64
	for _, r := range a.Rows {
		interest += r.Interest
	}
	if interest != 13000000-9800000 {
		t.Fatalf("the interest over the term is what was paid above what was received: %d", interest)
	}
	if a.Rows[0].Date != "2027-01-02" || a.Rows[0].Period != "2027-01" || a.Rows[2].Date != "2029-01-02" {
		t.Fatalf("yearly dates: %+v", a.Rows)
	}

	// Twelve equal monthly instalments at 10% with the same fee: 8 791.59 a
	// month, the last one 8 791.56; effective 1.1530% a month, 14.75% a year.
	b, err := AmortisedCost(10000000, 200000, big.NewRat(10, 1), 12, 1, "annuity", "2027-01-31")
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Rows) != 12 || row(b.Rows[0]) != "1 9800000 112996 83333 795826 879159 9033837" ||
		row(b.Rows[11]) != "12 869134 10022 7266 871890 879156 0" {
		t.Fatalf("the annuity: first %s, last %s", row(b.Rows[0]), row(b.Rows[11]))
	}
	if b.PeriodPct().FloatString(8) != "1.15301902" || b.AnnualPct().FloatString(4) != "14.7483" {
		t.Fatalf("the annuity's effective rate: %s a month, %s a year", b.PeriodPct().FloatString(8), b.AnnualPct().FloatString(4))
	}
	// the day clamps to the month: 31 January, then 28 February, 31 March
	if b.Rows[1].Date != "2027-02-28" || b.Rows[2].Date != "2027-03-31" || b.Rows[11].Period != "2027-12" {
		t.Fatalf("monthly dates clamp to the month's length: %s %s", b.Rows[1].Date, b.Rows[2].Date)
	}

	// Without a fee the contract's rate is the effective rate, exactly, and
	// every row's interest is the lender's own; the last instalment carries
	// the rounding of the eleven before it.
	c, err := AmortisedCost(10000000, 0, big.NewRat(10, 1), 12, 1, "annuity", "2027-01-01")
	if err != nil {
		t.Fatal(err)
	}
	if c.Rate.Cmp(big.NewRat(1, 120)) != 0 {
		t.Fatalf("no fee: the rate is the contract's, 10%%/12 = 1/120, got %s", c.Rate.RatString())
	}
	for _, r := range c.Rows {
		if r.Interest != r.Nominal {
			t.Fatalf("no fee: row %d accrues %d at the effective rate against %d at the contract's", r.N, r.Interest, r.Nominal)
		}
	}
	if row(c.Rows[11]) != "12 871890 7266 7266 871890 879156 0" {
		t.Fatalf("the last instalment completes the schedule: %s", row(c.Rows[11]))
	}

	// Equal capital parts, quarterly, 8% with a 1 000 fee: 25 000 of capital
	// each quarter and the interest falling with what is owed.
	d, err := AmortisedCost(10000000, 100000, big.NewRat(8, 1), 12, 3, "equal_capital", "2027-03-31")
	if err != nil {
		t.Fatal(err)
	}
	want = []string{
		"1 9900000 239528 200000 2500000 2700000 7439528",
		"2 7439528 179997 150000 2500000 2650000 4969525",
		"3 4969525 120236 100000 2500000 2600000 2489761",
		"4 2489761 60239 50000 2500000 2550000 0",
	}
	for i, r := range d.Rows {
		if row(r) != want[i] {
			t.Fatalf("quarterly row %d: %s, want %s", i+1, row(r), want[i])
		}
	}
	if d.Rows[1].Date != "2027-06-30" || d.Rows[3].Date != "2027-12-31" || d.AnnualPct().FloatString(4) != "10.0348" {
		t.Fatalf("quarterly dates and rate: %s %s %s", d.Rows[1].Date, d.Rows[3].Date, d.AnnualPct().FloatString(4))
	}

	// The same inputs give the same rows, every time (invariant 4).
	again, _ := AmortisedCost(10000000, 200000, big.NewRat(10, 1), 12, 1, "annuity", "2027-01-31")
	if again.Rate.Cmp(b.Rate) != 0 || fmt.Sprint(again.Rows) != fmt.Sprint(b.Rows) {
		t.Fatal("the schedule is not deterministic")
	}
}

func TestAmortisedCostRefusesWhatItCannotSchedule(t *testing.T) {
	for _, c := range []struct {
		nominal, fee  int64
		months, every int
		kind, first   string
		want          string
	}{
		{10000000, 10000000, 12, 1, "annuity", "2027-01-01", "must stay below the nominal"},
		{10000000, 0, 13, 3, "annuity", "2027-01-01", "not whole instalments"},
		{10000000, 0, 12, 5, "annuity", "2027-01-01", "1, 2, 3, 4, 6 or 12 months"},
		{10000000, 0, 12, 1, "balloon", "2027-01-01", "the kind must be one of"},
		{10000000, 0, 7200, 1, "annuity", "2027-01-01", "above the bound"},
		{10000000, 0, 12, 1, "annuity", "soon", "is not a date"},
		{0, 0, 12, 1, "annuity", "2027-01-01", "above zero"},
	} {
		_, err := AmortisedCost(c.nominal, c.fee, big.NewRat(10, 1), c.months, c.every, c.kind, c.first)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Fatalf("%+v: err = %v, want %q", c, err, c.want)
		}
	}
}

// loanTerms is a door's shape: what a person gives when scheduling a loan.
func loanTerms() *Env {
	terms := ObjectType{Name: "door:schedule_loan", Fields: []FieldDef{
		{Name: "nominal", Type: "money"}, {Name: "fee", Type: "money"}, {Name: "rate", Type: "decimal"},
		{Name: "months", Type: "int"}, {Name: "first", Type: "date"},
	}}
	return &Env{Typed: map[string]string{"$": terms.Name}, Types: func(name string) (ObjectType, bool) {
		if name == terms.Name {
			return terms, true
		}
		return ObjectType{}, false
	}, Raw: true}
}

func TestFormulaAmortisedCostIsRowsAFormulaWalks(t *testing.T) {
	env := loanTerms()
	p := `{"nominal": "100000.00", "fee": "2000.00", "rate": "10", "months": 36, "first": "2027-01-02"}`
	call := `amortised_cost($.nominal, $.fee, $.rate, $.months, 12, "bullet", $.first)`

	// The rows are a collection like any other: folds walk them, and their
	// fields are typed — the interest is money, the date a date.
	if v, _ := evalFormula(t, "=count(r in "+call+")", p, "int", env); v != int64(3) {
		t.Fatalf("three instalments, got %v", v)
	}
	if v, _ := evalFormula(t, "=sum(r in "+call+" : r.interest)", p, "money", env); v != int64(3200000) {
		t.Fatalf("the interest over the term: %v", v)
	}
	if v, _ := evalFormula(t, "=max(r in "+call+" : r.date)", p, "date", env); v != "2029-01-02" {
		t.Fatalf("the last instalment's day: %v", v)
	}
	if v, _ := evalFormula(t, `=effective_rate($.nominal, $.fee, $.rate, $.months, 12, "bullet")`, p, "decimal", env); v != "10.8158" {
		t.Fatalf("the effective rate a year, in per cent: %v", v)
	}
	// Each walks the rows as it walks a keyed collection: typed state per
	// element, so a rule materialises one instalment per row.
	tm, err := ParseTemplate("=" + call)
	if err != nil {
		t.Fatal(err)
	}
	if !tm.IsRows() || tm.IsCollection() {
		t.Fatal("a method's rows are a collection an each may walk")
	}
	items, err := tm.Collection(decode(t, p), env)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 || items[0].Type != "amortised_cost_row" || items[0].State["interest"] != int64(1059949) ||
		items[2].State["closing"] != int64(0) || items[0].State["rate"] != "10.81580553" {
		t.Fatalf("the rows as each sees them: %+v", items)
	}
	// Typed state reads: the row's money is money, through the prefix an
	// each sets.
	scope := `{"doc": ` + p + `, "line": {"object_type": "amortised_cost_row", "state": {"n": 1, "date": "2027-01-02", "period": "2027-01",
		"opening": 9800000, "interest": 1059949, "nominal_interest": 1000000, "capital": 0, "payment": 1000000, "closing": 9859949, "rate": "10.81580553"}}, "n": 1}`
	rows := &Env{Typed: map[string]string{"$.line.state": "amortised_cost_row"}, Raw: true}
	if v, _ := evalFormula(t, "=$.line.state.interest", scope, "money", rows); v != int64(1059949) {
		t.Fatalf("a row's interest copied into money: %v", v)
	}
	if v, _ := evalFormula(t, "=$.line.state.opening + $.line.state.interest", scope, "money", rows); v != int64(10859949) {
		t.Fatalf("a row's money computes as money: %v", v)
	}

	// The draft gate: the right number of arguments, a rows result where a
	// list is wanted, a decimal where a rate is.
	if _, err := ParseFormula(`amortised_cost($.nominal, $.fee)`); err == nil || !strings.Contains(err.Error(), "takes 7 argument") {
		t.Fatalf("arity: %v", err)
	}
	if _, err := ParseFormula(`effective_rate($.nominal, $.fee, $.rate, $.months, 12, "bullet", $.first)`); err == nil || !strings.Contains(err.Error(), "takes 6 argument") {
		t.Fatalf("arity: %v", err)
	}
	f, _ := ParseFormula("sum(r in " + call + " : r.interest)")
	if err := f.Check(env.Types, env.Typed, FieldDef{Name: "amount", Type: "money"}); err != nil {
		t.Fatalf("a sum over the rows is money: %v", err)
	}
	// the kernel declares the rows' type itself, so a misnamed field is
	// refused as the formula is read, before any catalog
	if _, err := ParseFormula("sum(r in " + call + " : r.nothing)"); err == nil || !strings.Contains(err.Error(), `no field "nothing"`) {
		t.Fatalf("a row's fields are known at draft time: %v", err)
	}
	f, _ = ParseFormula(call)
	if err := f.Check(env.Types, env.Typed, FieldDef{Name: "x", Type: "money"}); err == nil {
		t.Fatal("rows cannot be stored in a money field")
	}
	// A refusal names the method's law.
	wantErr(t, `=count(r in amortised_cost($.nominal, $.nominal, $.rate, $.months, 12, "bullet", $.first))`, p, "int", env, "must stay below the nominal")
}

func TestLeaseScheduleSplitsEachPaymentByTheInternalRate(t *testing.T) {
	// A car: 100 000 net, 10 000 paid on handover, 35 monthly instalments of
	// 2 700 and a 1 000 buy-out with the last: 95 500 paid for 90 000 owed,
	// 5 500 of interest in all, 4.0321% a year.
	a, err := LeaseSchedule(10000000, 1000000, 270000, 35, 1, 100000, "2027-02-15")
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Rows) != 35 || row(a.Rows[0]) != "1 9000000 29696 29696 240304 270000 8759696" ||
		row(a.Rows[1]) != "2 8759696 28903 28903 241097 270000 8518599" || row(a.Rows[34]) != "35 368784 1216 1216 368784 370000 0" {
		t.Fatalf("the lease: first %s, second %s, last %s", row(a.Rows[0]), row(a.Rows[1]), row(a.Rows[34]))
	}
	if a.PeriodPct().FloatString(8) != "0.32995434" || a.AnnualPct().FloatString(4) != "4.0321" {
		t.Fatalf("the lease's rate: %s a month, %s a year", a.PeriodPct().FloatString(8), a.AnnualPct().FloatString(4))
	}
	var interest int64
	for _, r := range a.Rows {
		interest += r.Interest
	}
	if interest != 550000 || a.Rows[0].Date != "2027-02-15" || a.Rows[34].Date != "2029-12-15" {
		t.Fatalf("5 500 of interest over the term, from February 2027 to December 2029: %d %s %s", interest, a.Rows[0].Date, a.Rows[34].Date)
	}
	// Payments that are exactly the value: no interest, every instalment capital.
	z, err := LeaseSchedule(1200000, 0, 100000, 12, 1, 0, "2027-01-01")
	if err != nil {
		t.Fatal(err)
	}
	if z.Rate.Sign() != 0 || z.Rows[0].Interest != 0 || z.Rows[11].Closing != 0 {
		t.Fatalf("a lease at no cost: %s %+v", z.Rate.RatString(), z.Rows[0])
	}
	for _, c := range []struct {
		value, initial, inst, residual int64
		want                           string
	}{
		{10000000, 10000000, 270000, 0, "must stay below the value"},
		{10000000, 0, 0, 0, "the instalment must be above zero"},
		{10000000, 0, 100000, 0, "are below the"},
	} {
		if _, err := LeaseSchedule(c.value, c.initial, c.inst, 12, 1, c.residual, "2027-01-01"); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Fatalf("%+v: err = %v, want %q", c, err, c.want)
		}
	}

	// As a formula: rows an each walks, the rate a year, the draft gate.
	terms := ObjectType{Name: "door:record_lease", Fields: []FieldDef{
		{Name: "value", Type: "money"}, {Name: "initial", Type: "money"}, {Name: "instalment", Type: "money"},
		{Name: "months", Type: "int"}, {Name: "residual", Type: "money"}, {Name: "first", Type: "date"},
	}}
	env := &Env{Typed: map[string]string{"$": terms.Name}, Types: func(name string) (ObjectType, bool) { return terms, name == terms.Name }, Raw: true}
	p := `{"value": "100000.00", "initial": "10000.00", "instalment": "2700.00", "months": 35, "residual": "1000.00", "first": "2027-02-15"}`
	call := `lease_schedule($.value, $.initial, $.instalment, $.months, 1, $.residual, $.first)`
	if v, _ := evalFormula(t, "=sum(r in "+call+" : r.interest)", p, "money", env); v != int64(550000) {
		t.Fatalf("the interest over the lease: %v", v)
	}
	if v, _ := evalFormula(t, "=sum(r in "+call+" : r.capital)", p, "money", env); v != int64(9000000) {
		t.Fatalf("the capital repaid is what was owed: %v", v)
	}
	if v, _ := evalFormula(t, `=lease_rate($.value, $.initial, $.instalment, $.months, 1, $.residual)`, p, "decimal", env); v != "4.0321" {
		t.Fatalf("the lease's rate a year: %v", v)
	}
	tm, _ := ParseTemplate("=" + call)
	if !tm.IsRows() {
		t.Fatal("a lease's rows are a collection an each may walk")
	}
	if _, err := ParseFormula(`lease_rate($.value, $.initial)`); err == nil || !strings.Contains(err.Error(), "takes 6 argument") {
		t.Fatalf("arity: %v", err)
	}
}
