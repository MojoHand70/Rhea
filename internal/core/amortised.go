package core

import (
	"fmt"
	"math/big"
	"slices"
	"time"
)

// Amortised cost with the effective interest rate: the first equation Rhea
// solves (DIRECTION "Optimal decisions", KK 2026-11-04). A loan taken with a
// fee, or a bond bought below its face value, costs more than its stated
// rate says: the fee is interest paid up front. The accounting act values
// such a debt or claim at amortised cost (ustawa o rachunkowości art. 28
// ust. 1 pkt 7a and 8a, ust. 8a) — what was received, less what was repaid,
// plus the interest accrued at the one rate that discounts every instalment
// back to what was received: the effective interest rate (rozporządzenie
// Ministra Finansów w sprawie instrumentów finansowych § 3 pkt 11 and 12).
// A finance lease is the same equation stated the other way round: the
// contract names the instalments, not the rate, and the rate that discounts
// them to the asset's value less the first payment is the lease's internal
// rate of return (KSR 5, the lessee's split of each payment into capital and
// interest).
//
// This is a calculation with one right answer, so it is a kernel method:
// exact rationals throughout, the rate found by a fixed number of
// bisection steps (never "until it converges"), amounts rounded to the
// minor unit half up where the method says so, and the last instalment
// completing the schedule to the amount — the remainder lands on the last
// row, as an allocation's does. Replay recomputes it and gets the same
// rows (invariant 4); no float touches a figure (invariant 6).

// AmortisedRow is one instalment of the schedule, amounts in minor units.
type AmortisedRow struct {
	N        int    // the instalment's number, from 1
	Date     string // the day it falls due
	Period   string // the month it falls in, "YYYY-MM"
	Opening  int64  // the carrying amount before it: amortised cost
	Interest int64  // the period's interest at the effective rate
	Nominal  int64  // the period's interest at the contract's rate; the effective one where the contract names none
	Capital  int64  // the capital repaid with it
	Payment  int64  // the cash that leaves: nominal interest plus capital
	Closing  int64  // the carrying amount after it
}

// Amortised is the schedule and the rate that explains it.
type Amortised struct {
	// Rate is the effective interest rate per period, exact: the contract's
	// own rate when nothing was paid up front, else the root found by
	// bisection to 2^-64.
	Rate *big.Rat
	// Every is the instalment's length in months.
	Every int
	Rows  []AmortisedRow
}

// AmortisedKinds are the repayment shapes a contract takes: equal
// instalments (raty równe, annuitetowe), equal capital parts with falling
// interest (raty malejące), or interest each period with the capital at the
// end (odsetki okresowo, kapitał na koniec).
var AmortisedKinds = []string{"annuity", "equal_capital", "bullet"}

// maxInstalments bounds the schedule: fifty years of monthly instalments.
const maxInstalments = 600

// bisectionSteps fixes the search for the rate: a bounded iteration, never
// a loop that runs until a condition holds.
const bisectionSteps = 64

// AmortisedCost builds the schedule of a loan or a bond: nominal is the face
// value, fee what was kept back on day one (a bank's prowizja, a bond's
// discount), ratePct the contract's annual rate in per cent, months the
// term, every the instalment's length in months (a divisor of twelve),
// kind one of AmortisedKinds, first the day the first instalment falls due.
// The contract's instalments come first, from its own rate; then the one
// rate that discounts them to nominal less fee; then each period's
// interest at that rate on the carrying amount, the last period completing
// the schedule to zero.
func AmortisedCost(nominal, fee int64, ratePct *big.Rat, months, every int, kind, first string) (Amortised, error) {
	switch {
	case nominal <= 0:
		return Amortised{}, fmt.Errorf("amortised_cost: the nominal must be above zero, got %s", FormatMoney(nominal))
	case fee < 0:
		return Amortised{}, fmt.Errorf("amortised_cost: the fee cannot be negative, got %s", FormatMoney(fee))
	case fee >= nominal:
		return Amortised{}, fmt.Errorf("amortised_cost: the fee %s must stay below the nominal %s", FormatMoney(fee), FormatMoney(nominal))
	case ratePct == nil || ratePct.Sign() < 0:
		return Amortised{}, fmt.Errorf("amortised_cost: the rate cannot be negative")
	case !slices.Contains(AmortisedKinds, kind):
		return Amortised{}, fmt.Errorf("amortised_cost: the kind must be one of %v, not %q", AmortisedKinds, kind)
	}
	n, start, err := term("amortised_cost", months, every, first)
	if err != nil {
		return Amortised{}, err
	}
	// the contract's rate per period: per cent, per year, for every months
	i := new(big.Rat).Mul(new(big.Rat).Quo(ratePct, big.NewRat(100, 1)), big.NewRat(int64(every), 12))

	// The contract's instalments, as the lender computes them: interest on
	// what is still owed at the contract's rate, rounded to the minor unit;
	// the capital by the kind; the last instalment repays the rest.
	payments, nominals := make([]int64, n), make([]int64, n)
	remaining := nominal
	var annuity int64
	if kind == "annuity" && i.Sign() > 0 {
		// A = N · i / (1 − (1+i)^−n), the equal instalment
		one := big.NewRat(1, 1)
		q := new(big.Rat).Inv(new(big.Rat).Add(one, i))
		a := new(big.Rat).Quo(new(big.Rat).Mul(ratInt(nominal), i), new(big.Rat).Sub(one, ratPow(q, n)))
		annuity = minorHalfUp(a)
	}
	for t := 1; t <= n; t++ {
		interest := minorHalfUp(new(big.Rat).Mul(ratInt(remaining), i))
		var capital int64
		switch {
		case t == n:
			capital = remaining
		case kind == "bullet":
			capital = 0
		case kind == "equal_capital" || annuity == 0:
			capital = nominal / int64(n)
		default:
			capital = annuity - interest
		}
		if capital < 0 || capital > remaining {
			return Amortised{}, fmt.Errorf("amortised_cost: instalment %d would repay %s of %s owed — the rate outruns the instalment", t, FormatMoney(capital), FormatMoney(remaining))
		}
		payments[t-1], nominals[t-1] = interest+capital, interest
		remaining -= capital
	}
	// With nothing kept back the contract's rate is the effective rate;
	// otherwise it is solved from the instalments.
	var rate *big.Rat
	if fee == 0 {
		rate = i
	}
	return amortise("amortised_cost", nominal-fee, payments, nominals, rate, start, every)
}

// LeaseSchedule builds the lessee's schedule of a finance lease (KSR 5): the
// contract names the asset's value, the first payment made on handover,
// the instalment, the term, the instalment's length in months, the final
// payment that buys the asset out, and the first instalment's day. The
// rate is the one that discounts every instalment and the buy-out to the
// value less the first payment — the lease's internal rate of return — and
// each period's interest at that rate is the financial cost, the rest of
// the instalment the capital repaid. A contract states no rate of its own,
// so the nominal interest of a row is its effective interest.
func LeaseSchedule(value, initial, instalment int64, months, every int, residual int64, first string) (Amortised, error) {
	switch {
	case value <= 0:
		return Amortised{}, fmt.Errorf("lease_schedule: the asset's value must be above zero, got %s", FormatMoney(value))
	case initial < 0:
		return Amortised{}, fmt.Errorf("lease_schedule: the first payment cannot be negative, got %s", FormatMoney(initial))
	case initial >= value:
		return Amortised{}, fmt.Errorf("lease_schedule: the first payment %s must stay below the value %s", FormatMoney(initial), FormatMoney(value))
	case instalment <= 0:
		return Amortised{}, fmt.Errorf("lease_schedule: the instalment must be above zero, got %s", FormatMoney(instalment))
	case residual < 0:
		return Amortised{}, fmt.Errorf("lease_schedule: the buy-out cannot be negative, got %s", FormatMoney(residual))
	}
	n, start, err := term("lease_schedule", months, every, first)
	if err != nil {
		return Amortised{}, err
	}
	payments := make([]int64, n)
	for t := range payments {
		payments[t] = instalment
	}
	payments[n-1] += residual
	return amortise("lease_schedule", value-initial, payments, nil, nil, start, every)
}

// term checks a contract's term and reads its first day.
func term(method string, months, every int, first string) (int, time.Time, error) {
	switch {
	case every <= 0 || 12%every != 0:
		return 0, time.Time{}, fmt.Errorf("%s: an instalment is 1, 2, 3, 4, 6 or 12 months long, not %d", method, every)
	case months <= 0 || months%every != 0:
		return 0, time.Time{}, fmt.Errorf("%s: the term of %d months is not whole instalments of %d", method, months, every)
	case months/every > maxInstalments:
		return 0, time.Time{}, fmt.Errorf("%s: %d instalments is above the bound of %d", method, months/every, maxInstalments)
	}
	start, err := time.Parse("2006-01-02", first)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("%s: the first instalment's date %q is not a date", method, first)
	}
	return months / every, start, nil
}

// amortise is the equation itself: received is the carrying amount on day
// one, payments the cash that leaves each period, nominals the contract's
// interest in each (nil where the contract names no rate), rate the
// effective rate per period when it is known, else solved. The rate is
// the r with Σ payment_t / (1+r)^t = received — zero when the payments
// are exactly what was received, otherwise bracketed between zero and one
// hundred per cent a period and halved a fixed number of times. Then each
// period's interest at that rate on the carrying amount, the last period
// completing the schedule to zero.
func amortise(method string, received int64, payments, nominals []int64, rate *big.Rat, start time.Time, every int) (Amortised, error) {
	if rate == nil {
		discounted := func(r *big.Rat) *big.Rat {
			q := new(big.Rat).Inv(new(big.Rat).Add(big.NewRat(1, 1), r))
			acc, p := new(big.Rat), big.NewRat(1, 1)
			for _, pay := range payments {
				p = new(big.Rat).Mul(p, q)
				acc.Add(acc, new(big.Rat).Mul(ratInt(pay), p))
			}
			return acc.Sub(acc, ratInt(received))
		}
		lo, hi := new(big.Rat), big.NewRat(1, 1)
		switch atZero := discounted(lo).Sign(); {
		case atZero < 0:
			return Amortised{}, fmt.Errorf("%s: the payments, %s in all, are below the %s received", method, FormatMoney(sumInt64(payments)), FormatMoney(received))
		case atZero == 0:
			rate = new(big.Rat)
		case discounted(hi).Sign() > 0:
			return Amortised{}, fmt.Errorf("%s: the effective rate is above one hundred per cent a period", method)
		default:
			for step := 0; step < bisectionSteps; step++ {
				mid := new(big.Rat).Quo(new(big.Rat).Add(lo, hi), big.NewRat(2, 1))
				if discounted(mid).Sign() > 0 {
					lo = mid
				} else {
					hi = mid
				}
			}
			rate = new(big.Rat).Quo(new(big.Rat).Add(lo, hi), big.NewRat(2, 1))
		}
	}
	rows := make([]AmortisedRow, len(payments))
	opening := received
	for t := range rows {
		row := &rows[t]
		row.N, row.Payment, row.Opening = t+1, payments[t], opening
		if t == len(rows)-1 {
			row.Interest = row.Payment - opening
		} else {
			row.Interest = minorHalfUp(new(big.Rat).Mul(ratInt(opening), rate))
		}
		if nominals != nil {
			row.Nominal = nominals[t]
		} else {
			row.Nominal = row.Interest
		}
		row.Capital = row.Payment - row.Nominal
		row.Closing = opening + row.Interest - row.Payment
		due := addMonths(start, t*every)
		row.Date = due.Format("2006-01-02")
		row.Period = due.Format("2006-01")
		opening = row.Closing
	}
	return Amortised{Rate: rate, Every: every, Rows: rows}, nil
}

func sumInt64(xs []int64) int64 {
	var s int64
	for _, x := range xs {
		s += x
	}
	return s
}

// AnnualPct is the effective rate per year in per cent, (1+r)^(12/every) − 1,
// rounded half even to four places: the figure a person compares with the
// contract's.
func (a Amortised) AnnualPct() *big.Rat {
	one := big.NewRat(1, 1)
	annual := new(big.Rat).Sub(ratPow(new(big.Rat).Add(one, a.Rate), 12/a.Every), one)
	return roundRat(new(big.Rat).Mul(annual, big.NewRat(100, 1)), 4, "half_even")
}

// PeriodPct is the effective rate per period in per cent, rounded half even
// to eight places: what a row shows beside its interest.
func (a Amortised) PeriodPct() *big.Rat {
	return roundRat(new(big.Rat).Mul(a.Rate, big.NewRat(100, 1)), 8, "half_even")
}

// States renders the rows as typed state, the shape a formula's each walks:
// money in minor units, the rate as its decimal text.
func (a Amortised) States() []map[string]any {
	rate, _ := canonical(a.PeriodPct())
	out := make([]map[string]any, 0, len(a.Rows))
	for _, r := range a.Rows {
		out = append(out, map[string]any{
			"n": int64(r.N), "date": r.Date, "period": r.Period,
			"opening": r.Opening, "interest": r.Interest, "nominal_interest": r.Nominal,
			"capital": r.Capital, "payment": r.Payment, "closing": r.Closing, "rate": rate,
		})
	}
	return out
}

// amortisedRowType is the shape of a row amortised_cost() and
// lease_schedule() return — a type the kernel itself declares, served
// before any pack's (catalogType), so a formula reads a row's interest as
// money and its date as a date wherever the rows go: a fold's binder, an
// each's $.line.state.
var amortisedRowType = ObjectType{Name: "amortised_cost_row", Version: 1, Fields: []FieldDef{
	{Name: "n", Type: "int", Required: true},
	{Name: "date", Type: "date", Required: true},
	{Name: "period", Type: "string", Required: true},
	{Name: "opening", Type: "money", Required: true},
	{Name: "interest", Type: "money", Required: true},
	{Name: "nominal_interest", Type: "money", Required: true},
	{Name: "capital", Type: "money", Required: true},
	{Name: "payment", Type: "money", Required: true},
	{Name: "closing", Type: "money", Required: true},
	{Name: "rate", Type: "decimal", Required: true},
}}

// methodRows names the functions whose result is a list of typed rows an
// each may walk, by the type of their rows.
var methodRows = map[string]string{"amortised_cost": amortisedRowType.Name, "lease_schedule": amortisedRowType.Name}

// MethodType is a type the kernel declares for a method's rows.
func MethodType(name string) (ObjectType, bool) {
	if name == amortisedRowType.Name {
		return amortisedRowType, true
	}
	return ObjectType{}, false
}

// catalogType finds a type: the kernel's own first, then the catalog's.
func catalogType(types Catalog, name string) (ObjectType, bool) {
	if t, ok := MethodType(name); ok {
		return t, true
	}
	if types == nil {
		return ObjectType{}, false
	}
	return types(name)
}

// callAmortised runs one of the methods' functions on checked arguments:
// amortised_cost / effective_rate (nominal, fee, rate, months, every, kind[,
// first]) and lease_schedule / lease_rate (value, initial, instalment,
// months, every, residual[, first]). The schedule functions return the rows
// typed by the kernel's own type; the rate functions the rate a year.
func callAmortised(name string, args []Value) (Value, error) {
	amount := func(i int, what string) (int64, error) {
		n, err := toMinor(args[i])
		if err != nil {
			return 0, fmt.Errorf("%s: %s: %w", name, what, err)
		}
		return n, nil
	}
	whole := func(i int, what string) (int, error) {
		n, err := toInt(args[i])
		if err != nil {
			return 0, fmt.Errorf("%s: %s: %w", name, what, err)
		}
		return int(n), nil
	}
	rows := name == "amortised_cost" || name == "lease_schedule"
	first := "2000-01-01" // the rate does not depend on the dates
	if rows {
		d, err := toDate(args[6])
		if err != nil {
			return Value{}, fmt.Errorf("%s: the first instalment's date: %w", name, err)
		}
		first = d.Format("2006-01-02")
	}
	var a Amortised
	var origin string
	switch name {
	case "amortised_cost", "effective_rate":
		nominal, err := amount(0, "the nominal")
		if err != nil {
			return Value{}, err
		}
		fee, err := amount(1, "the fee")
		if err != nil {
			return Value{}, err
		}
		rate, k, err := toNum(args[2])
		if err != nil || k == KMoney {
			return Value{}, fmt.Errorf("%s: the rate is a number in per cent, not %s", name, args[2].Kind)
		}
		months, err := whole(3, "the months")
		if err != nil {
			return Value{}, err
		}
		every, err := whole(4, "the instalment's months")
		if err != nil {
			return Value{}, err
		}
		if args[5].Kind != KString {
			return Value{}, fmt.Errorf("%s: the kind is a word, one of %v", name, AmortisedKinds)
		}
		if a, err = AmortisedCost(nominal, fee, rate, months, every, args[5].Str, first); err != nil {
			return Value{}, err
		}
		origin = fmt.Sprintf("amortised_cost(%s, %s, %s, %d, %d, %s, %s)", FormatMoney(nominal), FormatMoney(fee), rate.RatString(), months, every, args[5].Str, first)
	default:
		value, err := amount(0, "the asset's value")
		if err != nil {
			return Value{}, err
		}
		initial, err := amount(1, "the first payment")
		if err != nil {
			return Value{}, err
		}
		instalment, err := amount(2, "the instalment")
		if err != nil {
			return Value{}, err
		}
		months, err := whole(3, "the months")
		if err != nil {
			return Value{}, err
		}
		every, err := whole(4, "the instalment's months")
		if err != nil {
			return Value{}, err
		}
		residual, err := amount(5, "the buy-out")
		if err != nil {
			return Value{}, err
		}
		if a, err = LeaseSchedule(value, initial, instalment, months, every, residual, first); err != nil {
			return Value{}, err
		}
		origin = fmt.Sprintf("lease_schedule(%s, %s, %s, %d, %d, %s, %s)", FormatMoney(value), FormatMoney(initial), FormatMoney(instalment), months, every, FormatMoney(residual), first)
	}
	if !rows {
		return Value{Kind: KDecimal, Num: a.AnnualPct()}, nil
	}
	out := Value{Kind: KList, Type: amortisedRowType.Name, Origin: origin}
	raws := make([]any, 0, len(a.Rows))
	for i, state := range a.States() {
		out.List = append(out.List, Value{Kind: KObject, Raw: state, Type: amortisedRowType.Name, Origin: fmt.Sprintf("%s[%d]", origin, i)})
		raws = append(raws, state)
	}
	out.Raw = raws
	return out, nil
}

// toMinor reads an amount a method takes as minor units: money, or a number
// with at most two places.
func toMinor(v Value) (int64, error) {
	r, _, err := toNum(v)
	if err != nil {
		return 0, err
	}
	if !fits(r, 2) {
		return 0, fmt.Errorf("%s does not fit an amount (two places)", describe(r))
	}
	minor := new(big.Int).Quo(new(big.Int).Mul(r.Num(), big.NewInt(100)), r.Denom())
	if !minor.IsInt64() {
		return 0, fmt.Errorf("%s overflows money", describe(r))
	}
	return minor.Int64(), nil
}

// minorHalfUp rounds an amount in minor units to a whole number, half up.
func minorHalfUp(r *big.Rat) int64 {
	x := roundRat(r, 0, "half_up")
	return new(big.Int).Quo(x.Num(), x.Denom()).Int64()
}

// ratPow is r^n for a whole n ≥ 0, exact.
func ratPow(r *big.Rat, n int) *big.Rat {
	out := big.NewRat(1, 1)
	for ; n > 0; n-- {
		out = new(big.Rat).Mul(out, r)
	}
	return out
}

// addMonths moves a date n months, the day clamping to the target month's
// length: 31 January plus one month is 28 February.
func addMonths(d time.Time, n int) time.Time {
	first := time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, n, 0)
	last := first.AddDate(0, 1, -1).Day()
	day := d.Day()
	if day > last {
		day = last
	}
	return time.Date(first.Year(), first.Month(), day, 0, 0, 0, 0, time.UTC)
}
