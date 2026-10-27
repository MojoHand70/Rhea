package shell_test

import (
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"rhea/internal/core"
	"rhea/internal/pack"
)

// A report that takes a period is computed for the period a person picks
// (KK, 2026-10-27: "in some reports like Trial Balance there is no
// Accounting Period column and should not be"): the kernel puts the
// period's days into its query, the year counted from the first of the
// month the company says its financial year begins in, and serves the
// period beside the rows. Asked for none, it is the first month the books
// have not closed. Each company's year and closes are its own: Alfa's year
// begins in April, Beta's says nothing and is the calendar year; Alfa's
// closes move Alfa's month alone. A period that is not two days in order is
// refused, so no request's text reaches a query.
func TestAReportIsComputedForThePeriodAsked(t *testing.T) {
	w, a, b := newBooksWorld(t)
	spec, _ := json.Marshal(core.AnalysisSpec{Period: true,
		SQL:     `SELECT CAST({{period_from}} AS VARCHAR) AS f, CAST({{period_to}} AS VARCHAR) AS t, CAST({{year_from}} AS VARCHAR) AS y`,
		Columns: []core.ColumnSpec{{Field: "f", Label: "From", Type: "date"}, {Field: "t", Label: "To", Type: "date"}, {Field: "y", Label: "Year from", Type: "date"}}})
	if err := w.s.InsertViewDef(w.ctx, core.ViewDef{ID: "period-test", Version: 1, Notion: "analysis", Title: "The period",
		Domain: "finance", Function: "analysis", Spec: spec}); err != nil {
		t.Fatal(err)
	}
	w.mustDoor("kk", a, "set_financial_year", map[string]any{"financial_year_start": "04"})

	type served struct {
		Rows   [][]struct{ V string } `json:"rows"`
		Period core.Period            `json:"period"`
	}
	read := func(person, company, query string, want int) served {
		t.Helper()
		var got served
		if code := w.call(person, company, "GET", "/api/views/period-test"+query, nil, &got); code != want {
			t.Fatalf("%s at %s, %q: %d, want %d", person, company, query, code, want)
		}
		if want == 200 && (len(got.Rows) != 1 || got.Rows[0][0].V != got.Period.From || got.Rows[0][1].V != got.Period.To || got.Rows[0][2].V != got.Period.YearFrom) {
			t.Fatalf("%s at %s, %q: the query was not computed for the period it says: %+v", person, company, query, got)
		}
		return got
	}
	september := "?from=2026-09-01&to=2026-09-30"
	if p := read("kk", a, september, 200).Period; p != (core.Period{From: "2026-09-01", To: "2026-09-30", YearFrom: "2026-04-01"}) {
		t.Fatalf("Alfa's September counts from April: %+v", p)
	}
	if p := read("olga", b, september, 200).Period; p.YearFrom != "2026-01-01" {
		t.Fatalf("Beta says nothing of its year, which is the calendar year: %+v", p)
	}
	if p := read("kk", a, "?from=2027-01-01&to=2027-03-31", 200).Period; p.YearFrom != "2026-04-01" {
		t.Fatalf("Alfa's first quarter of 2027 is in the year from April 2026: %+v", p)
	}
	for _, q := range []string{"?from=2026-09-30&to=2026-09-01", "?from=2026-09-01", "?to=2026-09-30", "?from=2026-09-31&to=2026-10-31",
		"?from=" + url.QueryEscape("2026-09-01' OR 1=1 --") + "&to=2026-09-30"} {
		read("kk", a, q, 400)
	}

	// Asked for no period: the month after the last one closed, never
	// later than today's — the current month while nothing is closed.
	today, _ := time.Parse("2006-01-02", w.today)
	month := func(n int) string {
		return time.Date(today.Year(), today.Month()+time.Month(n), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	}
	if p := read("kk", a, "", 200).Period; p.From != month(0) {
		t.Fatalf("nothing closed: %+v, want the month from %s", p, month(0))
	}
	w.mustDoor("kk", a, "close_period", map[string]any{"period": month(-2), "book": "main"})
	if p := read("kk", a, "", 200).Period; p.From != month(-1) || p.To != time.Date(today.Year(), today.Month(), 0, 0, 0, 0, 0, time.UTC).Format("2006-01-02") {
		t.Fatalf("the month after the last closed: %+v, want the month from %s", p, month(-1))
	}
	if p := read("olga", b, "", 200).Period; p.From != month(0) {
		t.Fatalf("Alfa's close moved Beta's month: %+v", p)
	}
	w.mustDoor("kk", a, "close_period", map[string]any{"period": month(-1), "book": "main"})
	w.mustDoor("kk", a, "close_period", map[string]any{"period": month(0), "book": "main"})
	if p := read("kk", a, "", 200).Period; p.From != month(0) {
		t.Fatalf("today's month closed, the next is not yet: %+v, want the month from %s", p, month(0))
	}
}

// newBooksWorld is the core installed and approved, with two companies, each
// with its own record: Alfa, whose owner is kk, and Beta, whose owner is olga.
func newBooksWorld(t *testing.T) (w *grantsWorld, alfa, beta string) {
	t.Helper()
	w = newGrantsWorld(t)
	sums, err := pack.LoadCore(w.ctx, w.s, "init")
	if err != nil {
		t.Fatal(err)
	}
	for _, sum := range sums {
		if sum.Bundle == "" {
			continue
		}
		if _, _, errs, err := w.x.ApproveBundle(w.ctx, sum.Bundle, "operator", w.today); err != nil || len(errs) != 0 {
			t.Fatalf("approve %s: %v %v", sum.Bundle, err, errs)
		}
	}
	if _, err := w.s.InsertRuleVersion(w.ctx, core.Rule{ID: "test-own-company", Status: core.StatusActive, Priority: 100, EffectiveFrom: "2026-01-01",
		CreatedBy: "test", Description: "the company's own record", Spec: core.RuleSpec{Match: core.Match{EventType: "test.company.entered"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "company", Fields: map[string]string{"name": "=$.name", "kind": "self"}}}}}); err != nil {
		t.Fatal(err)
	}
	alfa, beta = "PL5250001111", "PL5260001246"
	w.person("Krzysztof Kowal", "kk", alfa, "owner")
	w.person("Olga Nowa", "olga", beta, "owner")
	for _, c := range [][3]string{{"kk", alfa, "Alfa"}, {"olga", beta, "Beta"}} {
		if code := w.call(c[0], c[1], "POST", "/api/events", map[string]any{"event_type": "test.company.entered", "occurred_at": "2026-01-01",
			"dedup_key": "own-" + c[1], "payload": map[string]any{"name": c[2]}}, nil); code != 200 {
			t.Fatalf("%s enters %s: %d", c[0], c[2], code)
		}
	}
	return w, alfa, beta
}
