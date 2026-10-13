package core

import (
	"encoding/json"
	"fmt"
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
type NarrativeSpec struct {
	Facts []NarrativeFact `json:"facts"`
	Voice string          `json:"voice,omitempty"`
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
