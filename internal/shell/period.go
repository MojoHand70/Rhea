package shell

import (
	"context"
	"net/url"
	"strconv"
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
	month := s.Clock.Today()[:7]
	locks, err := s.Store.ObjectsByType(ctx, "period_lock")
	if err != nil {
		return core.Period{}, 500, err
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
	p, err := core.MonthPeriod(month, start)
	if err != nil {
		return p, 500, err
	}
	return p, 0, nil
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
