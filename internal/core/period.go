package core

import (
	"fmt"
	"strings"
	"time"
)

// A report that takes a period (AnalysisSpec.Period) names it in its SQL
// with these: the kernel puts there, as DATEs, the period's first and last
// day and the first day of the financial year its last day falls in —
// dates it has checked itself, never a request's text.
const (
	PeriodFromInSQL = "{{period_from}}"
	PeriodToInSQL   = "{{period_to}}"
	YearFromInSQL   = "{{year_from}}"
)

// Period is the stretch of the books a report is computed for: its first and
// last day, and the first day of the financial year the last day falls in —
// where the opening balances stand and "from the start of the year" counts
// (UoR art. 18 ust. 1 asks both of a trial balance; KK, 2026-10-27).
type Period struct {
	From     string `json:"from"`
	To       string `json:"to"`
	YearFrom string `json:"year_from"`
}

// NewPeriod is the period from one day to another, its financial year
// beginning on the first of startMonth (1 to 12; 0 is a company that names
// none, the calendar year). The first day may not be after the last.
func NewPeriod(from, to string, startMonth int) (Period, error) {
	if startMonth == 0 {
		startMonth = 1
	}
	if startMonth < 1 || startMonth > 12 {
		return Period{}, fmt.Errorf("a financial year begins in a month, 1 to 12, not %d", startMonth)
	}
	f, err := time.Parse("2006-01-02", from)
	if err != nil {
		return Period{}, fmt.Errorf("a period's first day is a date (2026-09-01), not %q", from)
	}
	t, err := time.Parse("2006-01-02", to)
	if err != nil {
		return Period{}, fmt.Errorf("a period's last day is a date (2026-09-30), not %q", to)
	}
	if f.After(t) {
		return Period{}, fmt.Errorf("a period's first day %s is after its last %s", from, to)
	}
	y := time.Date(t.Year(), time.Month(startMonth), 1, 0, 0, 0, 0, time.UTC)
	if y.After(t) {
		y = y.AddDate(-1, 0, 0)
	}
	return Period{From: from, To: to, YearFrom: y.Format("2006-01-02")}, nil
}

// MonthPeriod is a calendar month ("2026-09") as a period.
func MonthPeriod(month string, startMonth int) (Period, error) {
	m, err := time.Parse("2006-01", month)
	if err != nil {
		return Period{}, fmt.Errorf("a month is written 2026-09, not %q", month)
	}
	return NewPeriod(m.Format("2006-01-02"), m.AddDate(0, 1, -1).Format("2006-01-02"), startMonth)
}

// SQL puts the period into a report's query.
func (p Period) SQL(q string) string {
	return strings.NewReplacer(PeriodFromInSQL, "DATE '"+p.From+"'", PeriodToInSQL, "DATE '"+p.To+"'",
		YearFromInSQL, "DATE '"+p.YearFrom+"'").Replace(q)
}

// CheckPeriod refuses a report that reads a period it does not say it takes,
// or says it takes one it never reads: a face offers the period by what the
// report says, and the query must be computed for what the face shows.
func (a AnalysisSpec) CheckPeriod() error {
	reads := strings.Contains(a.SQL, PeriodFromInSQL) || strings.Contains(a.SQL, PeriodToInSQL) || strings.Contains(a.SQL, YearFromInSQL)
	switch {
	case reads && !a.Period:
		return fmt.Errorf("the report reads a period (%s, %s, %s) but does not say it takes one (\"period\": true)", PeriodFromInSQL, PeriodToInSQL, YearFromInSQL)
	case a.Period && !reads:
		return fmt.Errorf("the report says it takes a period but its query reads none of %s, %s, %s", PeriodFromInSQL, PeriodToInSQL, YearFromInSQL)
	}
	return nil
}
