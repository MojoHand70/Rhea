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
