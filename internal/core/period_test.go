package core

import (
	"encoding/json"
	"strings"
	"testing"
)

// A period's financial year begins on the first of the month the company
// names (KK, 2026-10-27), the calendar year when it names none: September
// 2026 counts from 1 January in a calendar year and from 1 April 2026 in a
// year that begins in April, February 2027 from 1 April 2026 too — the
// year is the one the period's last day falls in.
func TestAPeriodCountsFromItsFinancialYear(t *testing.T) {
	for _, c := range []struct {
		from, to string
		start    int
		year     string
	}{
		{"2026-09-01", "2026-09-30", 1, "2026-01-01"},
		{"2026-09-01", "2026-09-30", 4, "2026-04-01"},
		{"2027-02-01", "2027-02-28", 4, "2026-04-01"},
		{"2026-03-01", "2026-03-31", 4, "2025-04-01"},
		{"2026-04-01", "2026-04-30", 4, "2026-04-01"},
		{"2025-12-01", "2026-01-31", 1, "2026-01-01"}, // across a year's end: the year of its last day
		{"2026-09-15", "2026-09-15", 0, "2026-01-01"}, // a company that names no month keeps the calendar year
	} {
		p, err := NewPeriod(c.from, c.to, c.start)
		if err != nil || p.From != c.from || p.To != c.to || p.YearFrom != c.year {
			t.Errorf("%s..%s from month %d: %+v %v, want the year from %s", c.from, c.to, c.start, p, err, c.year)
		}
	}
	m, err := MonthPeriod("2026-02", 1)
	if err != nil || m.From != "2026-02-01" || m.To != "2026-02-28" || m.YearFrom != "2026-01-01" {
		t.Errorf("February as a period: %+v %v", m, err)
	}
}

// A period a request asks for is two days, the first not after the last —
// nothing else reaches a query, so a request's text never becomes SQL.
func TestAPeriodIsTwoDays(t *testing.T) {
	for _, c := range [][2]string{
		{"2026-09-30", "2026-09-01"},
		{"2026-09-01", ""},
		{"", "2026-09-30"},
		{"2026-09-01", "2026-09-31"},
		{"2026-09-01'; DROP TABLE objects; --", "2026-09-30"},
		{"1 września", "2026-09-30"},
	} {
		if p, err := NewPeriod(c[0], c[1], 1); err == nil {
			t.Errorf("%q..%q was taken as a period: %+v", c[0], c[1], p)
		}
	}
	if _, err := MonthPeriod("2026-13", 1); err == nil {
		t.Error("month 13 was taken")
	}
	if _, err := NewPeriod("2026-09-01", "2026-09-30", 13); err == nil {
		t.Error("a financial year from month 13 was taken")
	}
}

// The kernel puts the period into a report's query as DATEs; a report says
// it takes a period, and one that reads a period it never declared — or
// declares one it never reads — is refused before it runs.
func TestAReportSaysItTakesAPeriod(t *testing.T) {
	p, _ := NewPeriod("2026-09-01", "2026-09-30", 4)
	got := p.SQL("SELECT 1 WHERE d BETWEEN " + PeriodFromInSQL + " AND " + PeriodToInSQL + " AND d >= " + YearFromInSQL)
	if got != "SELECT 1 WHERE d BETWEEN DATE '2026-09-01' AND DATE '2026-09-30' AND d >= DATE '2026-04-01'" {
		t.Fatalf("the period in the query: %s", got)
	}
	for _, c := range []struct {
		spec string
		ok   bool
	}{
		{`{"sql": "SELECT 1 WHERE d <= {{period_to}}", "period": true}`, true},
		{`{"sql": "SELECT 1 WHERE d <= {{period_to}}"}`, false},
		{`{"sql": "SELECT 1", "period": true}`, false},
		{`{"sql": "SELECT 1 WHERE d <= {{today}}"}`, true},
	} {
		var a AnalysisSpec
		if err := json.Unmarshal([]byte(c.spec), &a); err != nil {
			t.Fatal(err)
		}
		if err := a.CheckPeriod(); (err == nil) != c.ok {
			t.Errorf("%s: %v, want ok=%v", c.spec, err, c.ok)
		} else if err != nil && !strings.Contains(err.Error(), "period") {
			t.Errorf("%s: %v does not say it is the period", c.spec, err)
		}
	}
}
