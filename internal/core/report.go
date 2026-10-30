package core

import (
	"fmt"
	"regexp"
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

// Parameter is what a person sets before a report is computed (KK
// 2026-10-30, the context selector): its name, as the queries read it; its
// kind; its label in the client's language; its default; for a choice, the
// values offered. A period is read as {{name_from}}, {{name_to}} and
// {{name_year_from}} (the one named "period" also as {{period_from}},
// {{period_to}} and {{year_from}}, as reports read it since 2026-10-27);
// any other kind as {{name}}. The kernel puts in values it has checked —
// a firm or an account the scope holds, a choice among the values, a date,
// a period of two days in order — never a request's text.
type Parameter struct {
	Name    string   `json:"name"`
	Kind    string   `json:"kind"`
	Label   string   `json:"label,omitempty"`
	Default string   `json:"default,omitempty"` // a period: "open-month", "2026-09" or "2026-09-01..2026-09-30"; a firm, an account: its id; a choice, a text: a value; a date: a day
	Values  []string `json:"values,omitempty"`  // a choice: the values a person may pick
}

// ParameterKinds are the kinds a parameter may be.
var ParameterKinds = []string{"period", "firm", "account", "choice", "date", "text"}

// OpenMonth is the period default that means the first month the books
// have not closed, as a trial balance opens (2026-10-27).
const OpenMonth = "open-month"

var (
	parameterName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	placeholder   = regexp.MustCompile(`\{\{([a-z0-9_]+)\}\}`)
	periodDefault = regexp.MustCompile(`^(open-month|\d{4}-\d{2}|\d{4}-\d{2}-\d{2}\.\.\d{4}-\d{2}-\d{2})$`)
	dayDefault    = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

// Placeholders are the names a query reads a parameter by.
func (p Parameter) Placeholders() []string {
	if p.Kind != "period" {
		return []string{p.Name}
	}
	out := []string{p.Name + "_from", p.Name + "_to", p.Name + "_year_from"}
	if p.Name == "period" {
		out = append(out, "year_from")
	}
	return out
}

// Check refuses a parameter the kernel could not offer or put in.
func (p Parameter) Check() error {
	if !parameterName.MatchString(p.Name) {
		return fmt.Errorf("parameter %q: its name is lowercase letters, digits and underscores, starting with a letter", p.Name)
	}
	if !slices.Contains(ParameterKinds, p.Kind) {
		return fmt.Errorf("parameter %s: kind %q is not one of %v", p.Name, p.Kind, ParameterKinds)
	}
	switch p.Kind {
	case "period":
		if p.Default != "" && !periodDefault.MatchString(p.Default) {
			return fmt.Errorf("parameter %s: a period's default is %s, a month (2026-09) or two days (2026-09-01..2026-09-30), not %q", p.Name, OpenMonth, p.Default)
		}
	case "date":
		if p.Default != "" && !dayDefault.MatchString(p.Default) {
			return fmt.Errorf("parameter %s: a date's default is a day (2026-09-30), not %q", p.Name, p.Default)
		}
	case "choice":
		if len(p.Values) == 0 {
			return fmt.Errorf("parameter %s: a choice offers values", p.Name)
		}
		if p.Default != "" && !slices.Contains(p.Values, p.Default) {
			return fmt.Errorf("parameter %s: its default %q is not among its values", p.Name, p.Default)
		}
	}
	if p.Kind != "choice" && len(p.Values) > 0 {
		return fmt.Errorf("parameter %s: only a choice offers values", p.Name)
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
	// Parameters are what a person sets before the report is computed (KK
	// 2026-10-30, the context selector), each read by its queries.
	Parameters []Parameter `json:"parameters,omitempty"`
	// Period says the report is computed for a period a person picks, as
	// an analysis view is (AnalysisSpec.Period): the parameter named
	// "period", of kind period, opening on the first month not closed —
	// written so since 2026-10-30 as a parameter, kept here for the reports
	// kept before (KK 2026-10-30: only the accounting period changes).
	Period bool `json:"period,omitempty"`
}

// Params are the report's parameters as a face offers them: those written,
// and the period of a report that says it takes one without writing it.
func (r ReportSpec) Params() []Parameter {
	out := slices.Clone(r.Parameters)
	if r.Period && !slices.ContainsFunc(out, func(p Parameter) bool { return p.Name == "period" }) {
		out = append(out, Parameter{Name: "period", Kind: "period", Default: OpenMonth})
	}
	return out
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
	// every parameter is one the kernel can offer, read by some query; every
	// name a query reads is a parameter's
	known := map[string]bool{}
	names := map[string]bool{}
	for _, p := range r.Params() {
		if err := p.Check(); err != nil {
			return fmt.Errorf("report: %w", err)
		}
		if names[p.Name] {
			return fmt.Errorf("report: parameter %s named twice", p.Name)
		}
		names[p.Name] = true
		read := false
		for _, ph := range p.Placeholders() {
			known[ph] = true
			for _, c := range r.Components {
				if strings.Contains(c.SQL, "{{"+ph+"}}") {
					read = true
				}
			}
		}
		if !read {
			return fmt.Errorf("report: no query reads parameter %s (as {{%s}})", p.Name, strings.Join(p.Placeholders(), "}}, {{"))
		}
	}
	for i, c := range r.Components {
		for _, m := range placeholder.FindAllStringSubmatch(c.SQL, -1) {
			if !known[m[1]] {
				return fmt.Errorf("report: component %d reads {{%s}}, which no parameter gives", i+1, m[1])
			}
		}
	}
	return nil
}
