package core

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// Derived default views (DIRECTION 2026-10-03/05): an ObjectType already says
// enough — fields, types, label_field, is_document — for a serviceable list
// and detail. Objects *have* their visualization, by derivation; stored
// ViewDefs are the exceptions, authored only where seeing is genuinely a
// point of view (localization, role, emphasis), never boilerplate.
//
// Derived views are computed, never stored: version 0, ids under "derived:".

const derivedPrefix = "derived:"

func DerivedListID(typeName string) string   { return derivedPrefix + "list:" + typeName }
func DerivedDetailID(typeName string) string { return derivedPrefix + "detail:" + typeName }

// derivedFunction groups a type's views in the submenu the way every pack so
// far has by hand: documents under "documents", the rest as master data.
func derivedFunction(t ObjectType) string {
	if t.IsDocument {
		return "documents"
	}
	return "master data"
}

// humanize turns a type name into a title: "sales_invoice" → "Sales invoice".
func humanize(name string) string {
	s := strings.ReplaceAll(name, "_", " ")
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func DerivedListView(t ObjectType) ViewDef {
	cols := make([]ColumnSpec, len(t.Fields))
	for i, f := range t.Fields {
		cols[i] = ColumnSpec{Field: f.Name, Label: f.Name}
	}
	spec, err := json.Marshal(ListSpec{ObjectType: t.Name, Columns: cols})
	if err != nil {
		panic(fmt.Sprintf("marshal derived list spec for %q: %v", t.Name, err)) // a struct of strings cannot fail
	}
	return ViewDef{
		ID: DerivedListID(t.Name), Version: 0, Notion: "list",
		Title: humanize(t.Name), Domain: t.Domain, Function: derivedFunction(t),
		Spec: spec,
	}
}

func DerivedDetailView(t ObjectType) ViewDef {
	fields := make([]string, len(t.Fields))
	for i, f := range t.Fields {
		fields[i] = f.Name
	}
	spec, err := json.Marshal(DetailSpec{ObjectType: t.Name,
		Sections: []SectionSpec{{Title: humanize(t.Name), Fields: fields}}})
	if err != nil {
		panic(fmt.Sprintf("marshal derived detail spec for %q: %v", t.Name, err))
	}
	return ViewDef{
		ID: DerivedDetailID(t.Name), Version: 0, Notion: "detail",
		Title: humanize(t.Name), Domain: t.Domain, Function: derivedFunction(t),
		Spec: spec,
	}
}

// NarrativeSpec: spec of a `narrative` notion (DIRECTION 2026-10-10, "The host
// is a concierge": the briefing). Facts names the views the host speaks
// from — analysis views and matching grids, computed by the kernel — and
// Voice is the owner's instruction on how to speak (the language, the
// length, the manner). The host phrases the facts fresh each time and never
// computes a number: every number in its words must be one the facts
// carry, or the words are refused and the facts are shown alone. Its words
// are logged through the host_said door, so it never repeats itself.
// For names the roles it speaks to (KK 2026-10-17: the owner wants the
// result, the accountant what is open): a person whose role a narrative
// names hears that one, anyone else the narrative that names no role.
type NarrativeSpec struct {
	Facts []NarrativeFact `json:"facts"`
	Voice string          `json:"voice,omitempty"`
	For   []string        `json:"for,omitempty"`
}

// NarrativeFact is one view a narrative draws on, under the name the host
// may call it by.
type NarrativeFact struct {
	View string `json:"view"`
	As   string `json:"as,omitempty"`
}

// EventHostSaid is the raw event the host_said door appends: what the host
// said, on which view, on what facts — the host's words as facts in the
// log (never a templated welcome: the next words know the last).
const EventHostSaid = "host.said"

// EventHostUntaught is the raw event the note_untaught door appends: what the
// host was asked and had not been taught (DIRECTION 2026-10-17) — the
// person's words, her general description, and whether it reached the house.
const EventHostUntaught = "host.untaught"

// EventHostAnswered is the raw event the host_answered door appends: a
// question answered from the books now (DIRECTION 2026-10-17, "Reporting") —
// what was asked, the query she wrote, what it counts, how it was shown, and
// the first rows as they were told, so what she told someone can be read back.
const EventHostAnswered = "host.answered"

// QueueSpec: spec of a `queue` notion (DIRECTION 2026-10-10, "The host is a
// concierge": what to do today). The objects of one type that match the
// conditions — the open ones — ranked by a formula over each object's
// state, highest first, with today in the payload ($.today) so "days
// overdue" is a formula and not code; the columns to show; and, when asked,
// the approvals waiting at the gate beside them. The verbs a person can
// take on an item are not declared here: every active activity with an
// input of the type is offered on every item, inline.
type QueueSpec struct {
	ObjectType string      `json:"object_type"`
	Where      []Condition `json:"where,omitempty"` // over {"state": …}: $.state.status eq open
	Rank       string      `json:"rank"`            // a formula: =$.today - $.state.due_date
	Columns    []string    `json:"columns,omitempty"`
	Limit      int         `json:"limit,omitempty"`
	Approvals  bool        `json:"approvals,omitempty"`
}

// WordsSpec: spec of a `words` notion — how a market names what the kernel
// holds. The kernel names nothing in any language; a pack does: each type
// (and the shape it takes on a face), each field, the values of enums, the
// events, the doors, the groups of its chart of accounts. A face shows the
// words; one without them shows the kernel's names.
type WordsSpec struct {
	Lang   string               `json:"lang"`
	Types  map[string]TypeWords `json:"types,omitempty"`
	Fields map[string]string    `json:"fields,omitempty"` // field name, or "type.field" for one type's
	// Values names the values of an enum field (or "type.field"), and of the
	// few kernel words that are not fields (a warrant's basis).
	Values        map[string]map[string]string `json:"values,omitempty"`
	Events        map[string]string            `json:"events,omitempty"`         // event type → its name
	Doors         map[string]string            `json:"doors,omitempty"`          // activity → its name
	AccountGroups map[string]string            `json:"account_groups,omitempty"` // a code's first digits → the group's name
	// Views gives, in this language, the title of a view another pack brought
	// in another (2026-10-22): a face shows the market's word, never the
	// base pack's English.
	Views map[string]string `json:"views,omitempty"`
	// Notes say how this market's books are read — what a payment to social
	// security is, which amount is in the home currency, which month an
	// invoice counts in (2026-10-18): market knowledge for whoever reads the
	// books, the reporting author first, so the kernel names none of it.
	Notes []string `json:"notes,omitempty"`
}

// Merge lays later words over these (DIRECTION 2026-10-24, step 3): a
// company's words are its packs' together — the core's, then the market's,
// the regime's, the modules', the business packs' — the later naming a
// thing over the earlier; notes are kept from every pack, once each.
func (w *WordsSpec) Merge(o WordsSpec) {
	if o.Lang != "" {
		w.Lang = o.Lang
	}
	if len(o.Types) > 0 && w.Types == nil {
		w.Types = map[string]TypeWords{}
	}
	for k, v := range o.Types {
		w.Types[k] = v
	}
	over := func(dst *map[string]string, src map[string]string) {
		if len(src) > 0 && *dst == nil {
			*dst = map[string]string{}
		}
		for k, v := range src {
			(*dst)[k] = v
		}
	}
	over(&w.Fields, o.Fields)
	over(&w.Events, o.Events)
	over(&w.Doors, o.Doors)
	over(&w.AccountGroups, o.AccountGroups)
	over(&w.Views, o.Views)
	for f, vals := range o.Values {
		if w.Values == nil {
			w.Values = map[string]map[string]string{}
		}
		inner := w.Values[f]
		over(&inner, vals)
		w.Values[f] = inner
	}
	for _, n := range o.Notes {
		if !slices.Contains(w.Notes, n) {
			w.Notes = append(w.Notes, n)
		}
	}
}

// TypeWords names one type and says which shape it takes and what kind of
// thing it is.
type TypeWords struct {
	Label  string `json:"label"`
	Plural string `json:"plural,omitempty"`
	Shape  string `json:"shape,omitempty"`
	Kind   string `json:"kind,omitempty"`
}

// Kinds are what a thing in the books is (DIRECTION 2026-10-22, "The desk in
// order"): a document that came in or went out, an entry the books made of
// what happened, or master data the work refers to. A face files its lists by
// them, knowing no business.
var Kinds = []string{"document", "entry", "master"}

// Shapes are the few forms an object takes on a face (DIRECTION, the
// face): a document as a document, a card, a ledger, a statement.
var Shapes = []string{"paper", "card", "ledger", "statement"}

// Check refuses words a face could not use: a language is named, every
// label says something, every shape is one of Shapes.
func (w WordsSpec) Check() error {
	if w.Lang == "" {
		return fmt.Errorf("words: no language")
	}
	for t, tw := range w.Types {
		if strings.TrimSpace(tw.Label) == "" {
			return fmt.Errorf("words: type %s has no label", t)
		}
		if tw.Shape != "" && !slices.Contains(Shapes, tw.Shape) {
			return fmt.Errorf("words: type %s takes shape %q, not one of %v", t, tw.Shape, Shapes)
		}
		if tw.Kind != "" && !slices.Contains(Kinds, tw.Kind) {
			return fmt.Errorf("words: type %s is of kind %q, not one of %v", t, tw.Kind, Kinds)
		}
	}
	for _, m := range []map[string]string{w.Fields, w.Events, w.Doors, w.AccountGroups, w.Views} {
		for k, v := range m {
			if strings.TrimSpace(v) == "" {
				return fmt.Errorf("words: %s has an empty label", k)
			}
		}
	}
	for i, n := range w.Notes {
		if strings.TrimSpace(n) == "" {
			return fmt.Errorf("words: note %d is empty", i+1)
		}
	}
	for f, vals := range w.Values {
		for k, v := range vals {
			if strings.TrimSpace(v) == "" {
				return fmt.Errorf("words: value %s of %s has an empty label", k, f)
			}
		}
	}
	return nil
}

// StatementSpec: spec of a `statement` notion — one object's figures laid
// out as a statement: sections of rows, each a field with the operation
// that joins it to the rows above it. The figures are the object's own,
// computed by its rules; the statement orders and signs them and computes
// nothing (a declaration rounded to whole złoty does not add up to the
// grosze above it, and must not be made to).
type StatementSpec struct {
	ObjectType string             `json:"object_type"`
	Sections   []StatementSection `json:"sections"`
}

type StatementSection struct {
	Title string         `json:"title,omitempty"`
	Rows  []StatementRow `json:"rows"`
}

type StatementRow struct {
	Op      string `json:"op,omitempty"`      // "", "+", "-" or "="
	Field   string `json:"field"`             // a money field of the type
	Note    string `json:"note,omitempty"`    // a word under the row: "w pełnych złotych"
	NonZero bool   `json:"nonzero,omitempty"` // shown only when the figure is not zero
}

// Check against the type: every row names a money field it has.
func (s StatementSpec) Check(t ObjectType) error {
	if s.ObjectType != t.Name {
		return fmt.Errorf("statement over %s checked against %s", s.ObjectType, t.Name)
	}
	if len(s.Sections) == 0 {
		return fmt.Errorf("statement over %s: no sections", s.ObjectType)
	}
	for _, sec := range s.Sections {
		for _, r := range sec.Rows {
			if !slices.Contains([]string{"", "+", "-", "="}, r.Op) {
				return fmt.Errorf("statement over %s: row %s: op %q is not +, -, = or none", s.ObjectType, r.Field, r.Op)
			}
			fd, ok := t.Field(r.Field)
			if !ok || fd.Type != "money" {
				return fmt.Errorf("statement over %s: %s is not a money field of the type", s.ObjectType, r.Field)
			}
		}
	}
	return nil
}

// TodayInSQL is how an analysis view names the business date: the kernel
// puts the clock's day there as a DATE before the query runs, so a report
// that depends on the day (receivables by age) reads the kernel's day, not
// the machine's.
const TodayInSQL = "{{today}}"
