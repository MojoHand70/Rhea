package core

import (
	"fmt"
	"slices"
	"strings"
)

// Visual says how a query's rows are shown (DIRECTION 2026-10-17,
// "Reporting"): the query computes, the face lays out. It names the columns
// that play each part — never a figure of its own.
type Visual struct {
	Kind  string `json:"kind"`            // kpi | table | statement | bar | line | pivot
	Title string `json:"title,omitempty"` // in the client's language
	// kpi: one number, a label, a figure to compare it with
	Value   string `json:"value,omitempty"`
	Label   string `json:"label,omitempty"`
	Compare string `json:"compare,omitempty"`
	// statement: lines joined by op (+, -, =), sections when a column groups them
	Line    string `json:"line,omitempty"`
	Amount  string `json:"amount,omitempty"`
	Op      string `json:"op,omitempty"`
	Section string `json:"section,omitempty"`
	// bar, line: the axis, the measures, a column splitting them into series
	X      string   `json:"x,omitempty"`
	Y      []string `json:"y,omitempty"`
	Series string   `json:"series,omitempty"`
	// table, pivot: the columns shown, in order (all but the lineage when empty)
	Columns []string `json:"columns,omitempty"`
	// Formats: a column's format where its type does not say it — money,
	// percent, int, date, month or text.
	Formats map[string]string `json:"formats,omitempty"`
}

// VisualKinds are the shapes a face draws a query's rows in.
var VisualKinds = []string{"kpi", "table", "statement", "bar", "line", "pivot"}

var visualFormats = []string{"money", "percent", "int", "date", "month", "text"}

// Check refuses a visual that names a column the rows do not have, or lacks
// the part its kind needs.
func (v Visual) Check(columns []string) error {
	if !slices.Contains(VisualKinds, v.Kind) {
		return fmt.Errorf("visual: kind %q is not one of %v", v.Kind, VisualKinds)
	}
	has := func(part, c string, need bool) error {
		if c == "" {
			if need {
				return fmt.Errorf("visual %s: needs a %s column", v.Kind, part)
			}
			return nil
		}
		if !slices.Contains(columns, c) {
			return fmt.Errorf("visual %s: the %s column %q is not among the rows' columns (%s)", v.Kind, part, c, strings.Join(columns, ", "))
		}
		return nil
	}
	var errs []error
	switch v.Kind {
	case "kpi":
		errs = append(errs, has("value", v.Value, true), has("label", v.Label, false), has("compare", v.Compare, false))
	case "statement":
		errs = append(errs, has("line", v.Line, true), has("amount", v.Amount, true), has("op", v.Op, false), has("section", v.Section, false))
	case "bar", "line":
		errs = append(errs, has("x", v.X, true), has("series", v.Series, false))
		if len(v.Y) == 0 {
			errs = append(errs, fmt.Errorf("visual %s: needs y, the measures", v.Kind))
		}
		for _, y := range v.Y {
			errs = append(errs, has("y", y, true))
		}
	case "table", "pivot":
		for _, c := range v.Columns {
			errs = append(errs, has("shown", c, true))
		}
	}
	for c, f := range v.Formats {
		errs = append(errs, has("formatted", c, true))
		if !slices.Contains(visualFormats, f) {
			errs = append(errs, fmt.Errorf("visual: format %q of %s is not one of %v", f, c, visualFormats))
		}
	}
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}

// ReportSpec: spec of a `report` notion (DIRECTION 2026-10-17, "Reporting") —
// a page of components, each a query over the collection and a visual: KPI
// tiles beside a statement beside a chart. The kernel runs each component's
// query in the locked session when the report is read; the face lays the
// rows out. A report is kept through the gate like any definition.
type ReportSpec struct {
	Components []ReportComponent `json:"components"`
	// Period says the report is computed for a period a person picks, as
	// an analysis view is (AnalysisSpec.Period): its components' queries
	// read the period's days (PeriodFromInSQL…), the kernel puts the period
	// in, a face offers the choice. A report kept from a question asked
	// for a month takes the month as its parameter (KK 2026-10-30: only
	// the accounting period changes).
	Period bool `json:"period,omitempty"`
}

// ReadsPeriod says whether a query names a period's days.
func ReadsPeriod(sql string) bool {
	return strings.Contains(sql, PeriodFromInSQL) || strings.Contains(sql, PeriodToInSQL) || strings.Contains(sql, YearFromInSQL)
}

// ReportComponent is one part of a report.
type ReportComponent struct {
	Title   string `json:"title"`
	Reading string `json:"reading"` // what the query counts, in plain words
	SQL     string `json:"sql"`
	Visual  Visual `json:"visual"`
}

// MaxComponents bounds a report: a page, not a warehouse.
const MaxComponents = 12

// Check refuses a report the kernel could not show: no components, too many,
// one without a query, a reading or a visual kind it knows. Whether each
// query runs, and whether its visual names its columns, is seen when it runs.
func (r ReportSpec) Check() error {
	if len(r.Components) == 0 || len(r.Components) > MaxComponents {
		return fmt.Errorf("report: 1 to %d components, not %d", MaxComponents, len(r.Components))
	}
	for i, c := range r.Components {
		switch {
		case strings.TrimSpace(c.SQL) == "":
			return fmt.Errorf("report: component %d has no query", i+1)
		case strings.TrimSpace(c.Reading) == "":
			return fmt.Errorf("report: component %d says nothing of what it counts", i+1)
		case !slices.Contains(VisualKinds, c.Visual.Kind):
			return fmt.Errorf("report: component %d: visual kind %q is not one of %v", i+1, c.Visual.Kind, VisualKinds)
		}
	}
	reads := slices.ContainsFunc(r.Components, func(c ReportComponent) bool { return ReadsPeriod(c.SQL) })
	switch {
	case reads && !r.Period:
		return fmt.Errorf("report: a component reads a period (%s, %s, %s) but the report does not say it takes one (\"period\": true)", PeriodFromInSQL, PeriodToInSQL, YearFromInSQL)
	case r.Period && !reads:
		return fmt.Errorf("report: it says it takes a period but no component reads %s, %s or %s", PeriodFromInSQL, PeriodToInSQL, YearFromInSQL)
	}
	return nil
}
