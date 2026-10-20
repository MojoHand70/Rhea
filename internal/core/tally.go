package core

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// TallySpec: an analysis a person can teach in words and the author can
// draft (DIRECTION 2026-10-17, "Teaching her a figure"). It says what is
// added up — the objects of one type and which amount of each — which of
// them, grouped how, and in which lines. The kernel writes the query
// itself from these parts, each checked against the catalog: a name is
// one the catalog holds, a value is a quoted literal, so a drafted tally
// is data and never code (invariant 3). An analysis view carries either
// its own SQL (a pack's, written by hand) or a tally, not both.
type TallySpec struct {
	Of    string       `json:"of"`    // the type added up: a document type, or the bookings
	Value string       `json:"value"` // a money field of it; on a type with side and amount, "debit-credit" or "credit-debit"
	Where []TallyCond  `json:"where,omitempty"`
	By    []TallyGroup `json:"by,omitempty"`    // at most two groupings
	Lines []TallyLine  `json:"lines,omitempty"` // none: one amount per group
	// Labels of the columns the kernel makes beside the groupings: the
	// kernel names nothing in any language, so the tally does.
	LineLabel   string `json:"line_label,omitempty"`
	AmountLabel string `json:"amount_label,omitempty"`
}

// TallyCond keeps an object in the tally (or in one line of it). Field is a
// field of the type, or "link.field": a field of the object a ref field
// names, one step through the link ("account.type").
type TallyCond struct {
	Field string `json:"field"`
	Op    string `json:"op"`              // eq | ne | in | exists | absent
	Value any    `json:"value,omitempty"` // a scalar; a list for in
}

// TallyGroup is one grouping: a field (or link.field), its value as it is,
// or the month of a date field.
type TallyGroup struct {
	Field string `json:"field"`
	By    string `json:"by,omitempty"` // "" or "month"
	Label string `json:"label"`
}

// TallyLine is one named line. Op joins it to the lines above it: "" and
// "+" add, "-" subtracts, and "=" is the lines above it so joined — a
// line of its own with no where and no value.
type TallyLine struct {
	Line  string      `json:"line"`
	Op    string      `json:"op,omitempty"`
	Where []TallyCond `json:"where,omitempty"`
	Value string      `json:"value,omitempty"` // this line's amount, when not the tally's
}

// The signed amounts of a booking: the kernel's postings carry a side and an
// unsigned amount, and which side counts as plus is the tally's to say.
const (
	TallyDebitCredit = "debit-credit"
	TallyCreditDebit = "credit-debit"
)

const tallyMaxLines = 24

var tallyName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// tallyPlan is a checked tally: every path resolved against the catalog.
type tallyPlan struct {
	spec  TallySpec
	links []string // ref fields joined, in order of first use
}

// Check refuses a tally the kernel could not write a query for: a type or a
// field the catalog does not hold, an amount that is not money, a month of
// something that is not a date, an operation it does not know, a line
// without a name.
func (s TallySpec) Check(cat Catalog) error {
	_, err := s.plan(cat)
	return err
}

func (s TallySpec) plan(cat Catalog) (*tallyPlan, error) {
	t, ok := cat(s.Of)
	if !ok {
		return nil, fmt.Errorf("tally: %q is not a type in the catalog", s.Of)
	}
	p := &tallyPlan{spec: s}
	field := func(path string) (FieldDef, error) {
		parts := strings.Split(path, ".")
		if len(parts) > 2 {
			return FieldDef{}, fmt.Errorf("tally: %q is a field or link.field", path)
		}
		for _, n := range parts {
			if !tallyName.MatchString(n) {
				return FieldDef{}, fmt.Errorf("tally: %q is not a field name", path)
			}
		}
		f, ok := t.Field(parts[0])
		if !ok {
			return FieldDef{}, fmt.Errorf("tally: %s has no field %s", s.Of, parts[0])
		}
		if len(parts) == 1 {
			return f, nil
		}
		target, ok := RefTarget(f.Type)
		if !ok {
			return FieldDef{}, fmt.Errorf("tally: %s.%s is not a link, so %q reads nothing through it", s.Of, parts[0], path)
		}
		lt, ok := cat(target)
		if !ok {
			return FieldDef{}, fmt.Errorf("tally: %s.%s links to %s, which is not in the catalog", s.Of, parts[0], target)
		}
		lf, ok := lt.Field(parts[1])
		if !ok {
			return FieldDef{}, fmt.Errorf("tally: %s has no field %s", target, parts[1])
		}
		if !slices.Contains(p.links, parts[0]) {
			p.links = append(p.links, parts[0])
		}
		return lf, nil
	}
	value := func(v string) error {
		if v == TallyDebitCredit || v == TallyCreditDebit {
			side, ok1 := t.Field("side")
			amount, ok2 := t.Field("amount")
			if !ok1 || !ok2 || side.Type != "enum" || !slices.Contains(side.Values, "debit") || !slices.Contains(side.Values, "credit") || amount.Type != "money" {
				return fmt.Errorf("tally: %s has no side and amount, so %q says nothing of it", s.Of, v)
			}
			return nil
		}
		if strings.Contains(v, ".") {
			return fmt.Errorf("tally: the amount %q is a money field of %s itself", v, s.Of)
		}
		f, err := field(v)
		if err != nil {
			return err
		}
		if f.Type != "money" {
			return fmt.Errorf("tally: %s.%s is %s, not money", s.Of, v, f.Type)
		}
		return nil
	}
	conds := func(cs []TallyCond) error {
		for _, c := range cs {
			if _, err := field(c.Field); err != nil {
				return err
			}
			switch c.Op {
			case "eq", "ne":
				if _, err := tallyLiteral(c.Value); err != nil {
					return fmt.Errorf("tally: %s %s: %w", c.Field, c.Op, err)
				}
			case "in":
				vs, ok := c.Value.([]any)
				if !ok || len(vs) == 0 {
					return fmt.Errorf("tally: %s in needs a list of values", c.Field)
				}
				for _, v := range vs {
					if _, err := tallyLiteral(v); err != nil {
						return fmt.Errorf("tally: %s in: %w", c.Field, err)
					}
				}
			case "exists", "absent":
				if c.Value != nil {
					return fmt.Errorf("tally: %s %s takes no value", c.Field, c.Op)
				}
			default:
				return fmt.Errorf("tally: %s: op %q is not eq, ne, in, exists or absent", c.Field, c.Op)
			}
		}
		return nil
	}

	if s.Value == "" {
		return nil, fmt.Errorf("tally: no amount (value)")
	}
	if err := value(s.Value); err != nil {
		return nil, err
	}
	if err := conds(s.Where); err != nil {
		return nil, err
	}
	if len(s.By) > 2 {
		return nil, fmt.Errorf("tally: at most two groupings, not %d", len(s.By))
	}
	seen := map[string]bool{}
	for _, g := range s.By {
		f, err := field(g.Field)
		if err != nil {
			return nil, err
		}
		switch g.By {
		case "":
		case "month":
			if f.Type != "date" {
				return nil, fmt.Errorf("tally: the month of %s, which is %s, not a date", g.Field, f.Type)
			}
		default:
			return nil, fmt.Errorf("tally: group %s by %q: only by month", g.Field, g.By)
		}
		if strings.TrimSpace(g.Label) == "" {
			return nil, fmt.Errorf("tally: the grouping by %s has no label", g.Field)
		}
		if n := g.column(); seen[n] {
			return nil, fmt.Errorf("tally: two groupings by %s", n)
		} else {
			seen[n] = true
		}
	}
	if len(s.Lines) > tallyMaxLines {
		return nil, fmt.Errorf("tally: at most %d lines", tallyMaxLines)
	}
	names := map[string]bool{}
	for i, l := range s.Lines {
		if strings.TrimSpace(l.Line) == "" {
			return nil, fmt.Errorf("tally: line %d has no name", i+1)
		}
		if names[l.Line] {
			return nil, fmt.Errorf("tally: two lines named %q", l.Line)
		}
		names[l.Line] = true
		switch l.Op {
		case "", "+", "-":
			if l.Value != "" {
				if err := value(l.Value); err != nil {
					return nil, err
				}
			}
			if err := conds(l.Where); err != nil {
				return nil, err
			}
		case "=":
			if i == 0 {
				return nil, fmt.Errorf("tally: line %q sums the lines above it, and there are none", l.Line)
			}
			if l.Value != "" || len(l.Where) > 0 {
				return nil, fmt.Errorf("tally: line %q sums the lines above it; it has no where or value of its own", l.Line)
			}
		default:
			return nil, fmt.Errorf("tally: line %q: op %q is not +, -, = or none", l.Line, l.Op)
		}
	}
	return p, nil
}

// column is the name of a grouping's column: the month, or the field's
// path with the step through a link spelled with an underscore.
func (g TallyGroup) column() string {
	if g.By == "month" {
		return "month"
	}
	return strings.ReplaceAll(g.Field, ".", "_")
}

// tallyLiteral quotes a scalar for SQL: the only text of a tally that
// reaches the query unnamed by the catalog, so it is always a literal.
func tallyLiteral(v any) (string, error) {
	var s string
	switch x := v.(type) {
	case string:
		s = x
	case bool, float64, int, int64:
		s = Text(x)
	default:
		return "", fmt.Errorf("value %v is not a string, number or yes/no", v)
	}
	if strings.ContainsRune(s, 0) {
		return "", fmt.Errorf("value contains a NUL")
	}
	return "'" + strings.ReplaceAll(s, "'", "''") + "'", nil
}

// Compile writes the tally's query over the read side's objects table and
// the columns it yields: the groupings, then op, line and amount (or the
// amount alone without lines), with the row-lineage column naming each
// row's objects. Rows come grouped in order, lines in the tally's order.
func (s TallySpec) Compile(cat Catalog) (string, []ColumnSpec, error) {
	p, err := s.plan(cat)
	if err != nil {
		return "", nil, err
	}
	alias := map[string]string{}
	for i, l := range p.links {
		alias[l] = fmt.Sprintf("l%d", i)
	}
	expr := func(path string) string {
		parts := strings.Split(path, ".")
		if len(parts) == 1 {
			return fmt.Sprintf("json_extract_string(x.state,'$.%s')", parts[0])
		}
		return fmt.Sprintf("json_extract_string(%s.state,'$.%s')", alias[parts[0]], parts[1])
	}
	amount := func(v string) string {
		signed := "CASE WHEN json_extract_string(x.state,'$.side') = 'debit' THEN CAST(json_extract(x.state,'$.amount') AS BIGINT) ELSE -CAST(json_extract(x.state,'$.amount') AS BIGINT) END"
		switch v {
		case TallyDebitCredit:
			return signed
		case TallyCreditDebit:
			return "-(" + signed + ")"
		}
		return fmt.Sprintf("CAST(json_extract(x.state,'$.%s') AS BIGINT)", v)
	}
	cond := func(c TallyCond) string {
		e := expr(c.Field)
		switch c.Op {
		case "eq":
			v, _ := tallyLiteral(c.Value)
			return e + " = " + v
		case "ne":
			v, _ := tallyLiteral(c.Value)
			return "(" + e + " IS NULL OR " + e + " <> " + v + ")"
		case "in":
			var vs []string
			for _, x := range c.Value.([]any) {
				v, _ := tallyLiteral(x)
				vs = append(vs, v)
			}
			return e + " IN (" + strings.Join(vs, ", ") + ")"
		case "exists":
			return e + " IS NOT NULL"
		}
		return e + " IS NULL"
	}
	all := func(cs []TallyCond) string {
		if len(cs) == 0 {
			return "TRUE"
		}
		var out []string
		for _, c := range cs {
			out = append(out, cond(c))
		}
		return strings.Join(out, " AND ")
	}

	// The objects, each with its groupings, and its amount and membership
	// for every line that adds something up.
	lines := s.Lines
	plain := len(lines) == 0
	if plain {
		lines = []TallyLine{{}}
	}
	sel := []string{"x.object_id AS id"}
	var gs []string
	for i, g := range s.By {
		e := expr(g.Field)
		if g.By == "month" {
			e = "substr(" + e + ", 1, 7)"
		}
		sel = append(sel, fmt.Sprintf("%s AS g%d", e, i))
		gs = append(gs, fmt.Sprintf("g%d", i))
	}
	for i, l := range lines {
		if l.Op == "=" {
			continue
		}
		v := s.Value
		if l.Value != "" {
			v = l.Value
		}
		sel = append(sel, fmt.Sprintf("%s AS v%d", amount(v), i), fmt.Sprintf("(%s) AS f%d", all(l.Where), i))
	}
	from := "objects x"
	for _, l := range p.links {
		from += fmt.Sprintf(" LEFT JOIN objects %s ON %s.object_id = json_extract_string(x.state,'$.%s')", alias[l], alias[l], l)
	}
	of, _ := tallyLiteral(s.Of)
	where := "x.object_type = " + of
	if len(s.Where) > 0 {
		where += " AND " + all(s.Where)
	}
	sql := fmt.Sprintf("WITH t AS (SELECT %s FROM %s WHERE %s) ", strings.Join(sel, ", "), from, where)

	groupSel, groupBy, order := "", "", "k"
	if len(gs) > 0 {
		groupSel = strings.Join(gs, ", ") + ", "
		groupBy = " GROUP BY " + strings.Join(gs, ", ")
		order = strings.Join(gs, ", ") + ", k"
	}
	sign := func(op string) string {
		if op == "-" {
			return "-"
		}
		return "+"
	}
	var parts []string
	for k, l := range lines {
		var amt, ids string
		if l.Op == "=" {
			var terms, member []string
			for j, above := range lines[:k] {
				if above.Op == "=" {
					continue
				}
				terms = append(terms, fmt.Sprintf("%s COALESCE(SUM(v%d) FILTER (WHERE f%d), 0)", sign(above.Op), j, j))
				member = append(member, fmt.Sprintf("f%d", j))
			}
			amt = strings.Join(terms, " ")
			ids = fmt.Sprintf("string_agg(id, ',' ORDER BY id) FILTER (WHERE %s)", strings.Join(member, " OR "))
		} else {
			amt = fmt.Sprintf("COALESCE(SUM(v%d) FILTER (WHERE f%d), 0)", k, k)
			ids = fmt.Sprintf("string_agg(id, ',' ORDER BY id) FILTER (WHERE f%d)", k)
		}
		op, _ := tallyLiteral(l.Op)
		name, _ := tallyLiteral(l.Line)
		parts = append(parts, fmt.Sprintf("SELECT %s%d AS k, %s AS op, %s AS line, %s AS amount, %s AS ids FROM t%s", groupSel, k, op, name, amt, ids, groupBy))
	}
	cols := make([]ColumnSpec, 0, len(s.By)+3)
	for _, g := range s.By {
		cols = append(cols, ColumnSpec{Field: g.column(), Label: g.Label, Type: "string"})
	}
	am := ColumnSpec{Field: "amount", Label: s.AmountLabel, Type: "money"}
	// A plain sum of one money field checks itself on "why?": the parts make the cell.
	if !strings.Contains(s.Value, "-") && !slices.ContainsFunc(s.Lines, func(l TallyLine) bool { return l.Op == "-" || l.Op == "=" || (l.Value != "" && l.Value != s.Value) }) {
		am.SumOf = []string{s.Value}
	}
	if plain {
		sql += fmt.Sprintf("SELECT %samount, ids AS %s FROM (%s) ORDER BY %s", groupSel, RowsOfColumn, parts[0], order)
		return sql, append(cols, am), nil
	}
	sql += fmt.Sprintf("SELECT %sop, line, amount, ids AS %s FROM (%s) ORDER BY %s", groupSel, RowsOfColumn, strings.Join(parts, " UNION ALL "), order)
	cols = append(cols, ColumnSpec{Field: "op", Type: "string"}, ColumnSpec{Field: "line", Label: s.LineLabel, Type: "string"}, am)
	return sql, cols, nil
}
