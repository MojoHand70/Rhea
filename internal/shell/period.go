package shell

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"rhea/internal/core"
	"rhea/internal/store"
)

// periodOf is the period a report that takes one is computed for (KK,
// 2026-10-27): the one a person asked for (?from=&to=), else the first month
// the books have not closed — the month after the last month closed, the one
// the accountant is working on, and never later than today's; the current
// month while nothing is closed. Its financial year begins as the company
// says. A report that takes no period gets none. The status is 400 when the
// asked period is not two days in order.
func (s *Server) periodOf(ctx context.Context, spec core.AnalysisSpec, q url.Values) (core.Period, int, error) {
	if !spec.Period {
		return core.Period{}, 0, nil
	}
	start, err := s.yearStartMonth(ctx)
	if err != nil {
		return core.Period{}, 500, err
	}
	if from, to := q.Get("from"), q.Get("to"); from != "" || to != "" {
		p, err := core.NewPeriod(from, to, start)
		if err != nil {
			return p, 400, err
		}
		return p, 0, nil
	}
	p, err := s.openMonth(ctx, start)
	if err != nil {
		return p, 500, err
	}
	return p, 0, nil
}

// openMonth is the first month the books have not closed — the month after
// the last month closed, the one the accountant is working on, never later
// than today's; the current month while nothing is closed — as a period of
// a financial year beginning in month start.
func (s *Server) openMonth(ctx context.Context, start int) (core.Period, error) {
	month := s.Clock.Today()[:7]
	locks, err := s.Store.ObjectsByType(ctx, "period_lock")
	if err != nil {
		return core.Period{}, err
	}
	last := ""
	for _, l := range locks {
		if m, _ := l.State["month"].(string); ownRow(ctx, l) && m > last {
			last = m
		}
	}
	if last != "" {
		if m, err := time.Parse("2006-01", last); err == nil {
			if next := m.AddDate(0, 1, 0).Format("2006-01"); next < month {
				month = next
			}
		}
	}
	return core.MonthPeriod(month, start)
}

// parameterOut is a parameter as a face offers it, with the value the
// report was computed for: a period as its days, any other as its text.
type parameterOut struct {
	core.Parameter
	Value any `json:"value"`
}

// parameterValues is what the kernel puts into a report's queries for its
// parameters (KK 2026-10-30, the context selector): from the request
// (?name=value — a period as a month "2026-09" or two days "2026-09-01..
// 2026-09-30", the one named period also as ?from=&to=), else the default.
// Every value is checked before it nears a query: a firm or an account is
// a record of that kind the scope holds, a choice one of the values, a
// date a day, a period two days in order; a text is a quoted literal at
// most 200 characters. The status says whose fault a refusal is.
func (s *Server) parameterValues(ctx context.Context, params []core.Parameter, q url.Values) (map[string]string, []parameterOut, int, error) {
	sql := map[string]string{}
	var out []parameterOut
	quote := func(v string) string { return "'" + strings.ReplaceAll(v, "'", "''") + "'" }
	for _, p := range params {
		val := q.Get(p.Name)
		if val == "" && p.Kind == "period" && p.Name == "period" && (q.Get("from") != "" || q.Get("to") != "") {
			val = q.Get("from") + ".." + q.Get("to")
		}
		if val == "" {
			val = p.Default
		}
		po := parameterOut{Parameter: p, Value: val}
		switch p.Kind {
		case "period":
			start, err := s.yearStartMonth(ctx)
			if err != nil {
				return nil, nil, 500, err
			}
			var per core.Period
			switch {
			case val == "" || val == core.OpenMonth:
				if per, err = s.openMonth(ctx, start); err != nil {
					return nil, nil, 500, err
				}
			case strings.Contains(val, ".."):
				from, to, _ := strings.Cut(val, "..")
				if per, err = core.NewPeriod(from, to, start); err != nil {
					return nil, nil, 400, fmt.Errorf("parameter %s: %w", p.Name, err)
				}
			default:
				if per, err = core.MonthPeriod(val, start); err != nil {
					return nil, nil, 400, fmt.Errorf("parameter %s: %w", p.Name, err)
				}
			}
			sql[p.Name+"_from"], sql[p.Name+"_to"], sql[p.Name+"_year_from"] = "DATE '"+per.From+"'", "DATE '"+per.To+"'", "DATE '"+per.YearFrom+"'"
			if p.Name == "period" {
				sql["year_from"] = sql["period_year_from"]
			}
			po.Value = per
		case "firm", "account":
			kind := map[string]string{"firm": "company", "account": "account"}[p.Kind]
			if val == "" {
				return nil, nil, 400, fmt.Errorf("parameter %s needs a %s", p.Name, p.Kind)
			}
			o, err := s.Store.GetObject(ctx, val)
			if err != nil || o.Type != kind || !ownRow(ctx, o) {
				return nil, nil, 400, fmt.Errorf("parameter %s: %q is no %s of these books", p.Name, val, p.Kind)
			}
			sql[p.Name] = quote(val)
		case "choice":
			if !slices.Contains(p.Values, val) {
				return nil, nil, 400, fmt.Errorf("parameter %s: %q is not one of %s", p.Name, val, strings.Join(p.Values, ", "))
			}
			sql[p.Name] = quote(val)
		case "date":
			if _, err := time.Parse("2006-01-02", val); err != nil {
				return nil, nil, 400, fmt.Errorf("parameter %s: a day is written 2026-09-30, not %q", p.Name, val)
			}
			sql[p.Name] = "DATE '" + val + "'"
		case "text":
			if len([]rune(val)) > 200 {
				return nil, nil, 400, fmt.Errorf("parameter %s: at most 200 characters", p.Name)
			}
			sql[p.Name] = quote(val)
		}
		out = append(out, po)
	}
	return sql, out, 0, nil
}

// withParameters puts the parameters' values into a query.
func withParameters(sql string, values map[string]string) string {
	pairs := make([]string, 0, 2*len(values))
	for k, v := range values {
		pairs = append(pairs, "{{"+k+"}}", v)
	}
	return strings.NewReplacer(pairs...).Replace(sql)
}

// yearStartMonth is the month the financial year of the company a request
// acts for begins in: financial_year_start on its own record (the company
// of kind "self" in its own books), January when the record names none.
// A scope acting for no company takes the month its companies agree on,
// January when they differ.
func (s *Server) yearStartMonth(ctx context.Context) (int, error) {
	objs, err := s.Store.ObjectsByType(ctx, "company")
	if err != nil {
		return 0, err
	}
	months := map[int]bool{}
	for _, o := range objs {
		if o.State["kind"] != "self" || !ownRow(ctx, o) {
			continue
		}
		m, _ := o.State["financial_year_start"].(string)
		n, err := strconv.Atoi(m)
		if err != nil || n < 1 || n > 12 {
			n = 1
		}
		months[n] = true
	}
	if len(months) == 1 {
		for n := range months {
			return n, nil
		}
	}
	return 1, nil
}

// ownRow says a row belongs to the company the request acts for; acting
// for none, every row the scope sees does.
func ownRow(ctx context.Context, o core.Object) bool {
	c := store.CompanyOf(ctx)
	return c == "" || o.Company == c
}
