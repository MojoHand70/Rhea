package project

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"rhea/internal/core"
)

// The collection (DIRECTION 2026-10-17, "Reporting"): the business's own
// databases of objects — its sales, its stock, its workforce, a real-estate
// firm's buildings — as DuckDB schemas, one per domain, and a typed table per
// object type, so a query reads like the business. Every table has the
// object's id, a column per field the type declares, and the provenance
// beside it; what the type does not declare stays in the row's _state.
type Table struct {
	Schema  string   `json:"schema"`
	Name    string   `json:"name"`
	Type    string   `json:"object_type"`
	Columns []Column `json:"columns"`
}

// Column is one column of a collection table.
type Column struct {
	Name  string `json:"name"`
	SQL   string `json:"sql"`             // its DuckDB type
	Field string `json:"field,omitempty"` // the field it holds; empty for the provenance
	Kind  string `json:"kind,omitempty"`  // the field's declared type: money, date, ref<company>, enum…
}

// Provenance columns, the same on every table: where each row comes from
// (invariant 5), named so no field can collide with them.
var provenance = []Column{
	{Name: "_type_version", SQL: "INTEGER"},
	{Name: "_event", SQL: "BIGINT"},
	{Name: "_rule", SQL: "VARCHAR"},
	{Name: "_rule_version", SQL: "INTEGER"},
	{Name: "_occurred_at", SQL: "DATE"},
	{Name: "_root_event", SQL: "BIGINT"},
	{Name: "_state", SQL: "VARCHAR"},
}

var sqlName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// reserved schema names a domain may not take.
var reservedSchemas = []string{"main", "temp", "system", "information_schema", "pg_catalog"}

// SchemaOf names a domain's schema: the domain as it is when it is a plain
// name, "general" when a type has none or a name DuckDB keeps for itself.
func SchemaOf(domain string) string {
	d := strings.ToLower(strings.TrimSpace(domain))
	d = regexp.MustCompile(`[^a-z0-9_]+`).ReplaceAllString(d, "_")
	if !sqlName.MatchString(d) || slices.Contains(reservedSchemas, d) {
		return "general"
	}
	return d
}

// Collection describes the tables the catalog makes: one per type whose
// name is a plain SQL name, in the domain's schema.
func Collection(types []core.ObjectType) []Table {
	var out []Table
	for _, t := range types {
		if !sqlName.MatchString(t.Name) {
			continue
		}
		tb := Table{Schema: SchemaOf(t.Domain), Name: t.Name, Type: t.Name, Columns: []Column{{Name: "id", SQL: "VARCHAR"}}}
		for _, f := range t.Fields {
			if !sqlName.MatchString(f.Name) || f.Name == "id" {
				continue // stays in _state
			}
			tb.Columns = append(tb.Columns, Column{Name: f.Name, SQL: sqlTypeOf(f.Type), Field: f.Name, Kind: f.Type})
		}
		tb.Columns = append(tb.Columns, provenance...)
		out = append(out, tb)
	}
	slices.SortFunc(out, func(a, b Table) int { return strings.Compare(a.Schema+"."+a.Name, b.Schema+"."+b.Name) })
	return out
}

// sqlTypeOf maps a field's declared type to its DuckDB type: money as
// DECIMAL(18,2) (invariant 6), dates as DATE, links as the linked id.
func sqlTypeOf(fieldType string) string {
	switch fieldType {
	case "money":
		return "DECIMAL(18,2)"
	case "int":
		return "BIGINT"
	case "decimal":
		return "DECIMAL(38,10)"
	case "date":
		return "DATE"
	}
	return "VARCHAR" // string, enum, ref<…>
}

func quoteIdent(s string) string { return `"` + strings.ReplaceAll(s, `"`, `""`) + `"` }

// ddl creates one table of the collection.
func (t Table) ddl() string {
	cols := make([]string, len(t.Columns))
	for i, c := range t.Columns {
		cols[i] = quoteIdent(c.Name) + " " + c.SQL
	}
	return fmt.Sprintf("CREATE TABLE %s.%s (%s)", quoteIdent(t.Schema), quoteIdent(t.Name), strings.Join(cols, ", "))
}

func (t Table) insert() string {
	ph := strings.TrimSuffix(strings.Repeat("?, ", len(t.Columns)), ", ")
	return fmt.Sprintf("INSERT INTO %s.%s VALUES (%s)", quoteIdent(t.Schema), quoteIdent(t.Name), ph)
}

// cell turns a state value into what its column holds; nil when it cannot
// be read as the field's type (the value is still in _state).
func cell(kind string, v any) any {
	if v == nil {
		return nil
	}
	switch kind {
	case "money":
		switch x := v.(type) {
		case float64:
			return core.FormatMoney(int64(math.Round(x)))
		case int64:
			return core.FormatMoney(x)
		case int:
			return core.FormatMoney(int64(x))
		case string:
			if m, err := core.ParseMoney(x); err == nil {
				return core.FormatMoney(m)
			}
		}
		return nil
	case "int":
		switch x := v.(type) {
		case float64:
			return int64(x)
		case int64, int:
			return x
		case string:
			if n, err := strconv.ParseInt(x, 10, 64); err == nil {
				return n
			}
		}
		return nil
	case "decimal":
		switch x := v.(type) {
		case float64:
			return strconv.FormatFloat(x, 'f', -1, 64)
		case string:
			if _, err := strconv.ParseFloat(x, 64); err == nil {
				return x
			}
		}
		return nil
	case "date":
		if s, ok := v.(string); ok && len(s) >= 10 && dateLike.MatchString(s[:10]) {
			return s[:10]
		}
		return nil
	}
	switch x := v.(type) {
	case string:
		return x
	case float64, int64, int, bool:
		return core.Text(x)
	}
	b, _ := json.Marshal(v)
	return string(b)
}

var dateLike = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
