package core

import (
	"strings"
	"testing"
)

// An App as the kernel keeps it (KK 2026-10-26, APPS.md "The App definition"):
// what binds its parts, each part by name; tabs one level deep, a tab either
// a view or a screen the desk draws, a record shown as tabs of its own.
func TestAppValidate(t *testing.T) {
	good := App{ID: "faktury-sprzedazy", Version: 2, Title: "Faktury sprzedaży", Layer: "law",
		Data: []string{"sales_invoice", "ksef_submission"}, Uses: []string{"company", "posting"},
		Tabs:    []AppTab{{View: "sales-invoice-list"}, {Desk: "flow", Title: "Przepływ"}, {View: "naleznosci-do-pilnowania"}},
		Record:  &AppRecord{Of: "sales_invoice", Tabs: []AppTab{{Desk: "document-sheet", Title: "Faktura"}, {View: "sales-invoice-detail"}}},
		Actions: []string{"ksef"}, Rules: []string{"pl-book-sales-invoice"},
		Place: &AppPlace{Activity: "sales"}, Shows: []string{"desk", "phone"}}
	if err := good.Validate(); err != nil {
		t.Fatalf("a whole App: %v", err)
	}
	// where it sits and where it shows (KK 2026-10-30): a place names an
	// activity and may name a section; shows names faces; none named is both
	if err := (App{ID: "raporty", Version: 1, Title: "Raporty", Layer: "law", Tabs: []AppTab{{View: "wynik"}},
		Place: &AppPlace{Activity: "books", Section: "statements"}}).Validate(); err != nil {
		t.Fatalf("a place with its section: %v", err)
	}
	if a := (App{Shows: []string{"desk"}}); a.ShownOn("phone") || !a.ShownOn("desk") || !(App{}).ShownOn("phone") {
		t.Fatal("shows: the faces named, every face when none is")
	}
	// a company's own App (a kept report) carries its owner in its id
	if err := (App{ID: "answer-12@company-95", Version: 1, Title: "RZiS", Layer: "business", Tabs: []AppTab{{View: "answer-12@company-95"}}, Company: "company-95"}).Validate(); err != nil {
		t.Fatalf("an owned App: %v", err)
	}
	// an App with only its data opens on its kind of record's list
	if err := (App{ID: "stawki-vat", Version: 1, Title: "Stawki VAT", Layer: "law", Data: []string{"vat_rate"}}).Validate(); err != nil {
		t.Fatalf("data alone: %v", err)
	}
	// a market's App shows the record of a kind the core keeps, and one with
	// rules alone opens on the list of the first kind it uses
	if err := (App{ID: "rechnungen", Version: 1, Title: "Ausgangsrechnungen", Layer: "law", Uses: []string{"sales_invoice"},
		Record: &AppRecord{Of: "sales_invoice", Tabs: []AppTab{{View: "rechnung-detail"}}}}).Validate(); err != nil {
		t.Fatalf("the record of a kind used: %v", err)
	}
	if err := (App{ID: "kontenrahmen", Version: 1, Title: "Kontenrahmen", Layer: "law", Uses: []string{"account"}, Rules: []string{"de-register-account"}}).Validate(); err != nil {
		t.Fatalf("rules over a kind used: %v", err)
	}

	bad := []struct {
		why  string
		edit func(a *App)
	}{
		{"no id", func(a *App) { a.ID = "" }},
		{"an id that is not a name", func(a *App) { a.ID = "Faktury sprzedaży" }},
		{"no version", func(a *App) { a.Version = 0 }},
		{"no title", func(a *App) { a.Title = " " }},
		{"an unknown layer", func(a *App) { a.Layer = "market" }},
		{"nothing to show", func(a *App) { a.Tabs, a.Record, a.Data, a.Uses = nil, nil, nil, nil }},
		{"a tab both a view and the desk's", func(a *App) { a.Tabs[0].Desk = "flow" }},
		{"a tab neither", func(a *App) { a.Tabs[0] = AppTab{Title: "Lista"} }},
		{"a desk tab without its title", func(a *App) { a.Tabs[1].Title = "" }},
		{"a desk screen that is not a name", func(a *App) { a.Tabs[1].Desk = "the flow" }},
		{"a view twice", func(a *App) { a.Record.Tabs[1].View = "sales-invoice-list" }},
		{"a record of a kind it neither keeps nor uses", func(a *App) { a.Record.Of = "vat_rate" }},
		{"a record with no tab", func(a *App) { a.Record.Tabs = nil }},
		{"a kind kept and used", func(a *App) { a.Uses = append(a.Uses, "sales_invoice") }},
		{"a kind kept twice", func(a *App) { a.Data = append(a.Data, "sales_invoice") }},
		{"a rule twice", func(a *App) { a.Rules = append(a.Rules, "pl-book-sales-invoice") }},
		{"a door twice", func(a *App) { a.Actions = append(a.Actions, "ksef") }},
		{"an empty name", func(a *App) { a.Uses = append(a.Uses, "") }},
		{"a place with no activity", func(a *App) { a.Place = &AppPlace{Section: "receivables"} }},
		{"a place whose activity is not an id", func(a *App) { a.Place = &AppPlace{Activity: "Sprzedaż"} }},
		{"a place whose section is not an id", func(a *App) { a.Place = &AppPlace{Activity: "sales", Section: "Należności"} }},
		{"a face it does not know", func(a *App) { a.Shows = []string{"desk", "watch"} }},
		{"a face twice", func(a *App) { a.Shows = []string{"desk", "desk"} }},
		{"an owner that is not a name", func(a *App) { a.ID = "answer-12@Helios S.A." }},
	}
	for _, c := range bad {
		a := good
		a.Data = append([]string(nil), good.Data...)
		a.Uses = append([]string(nil), good.Uses...)
		a.Tabs = append([]AppTab(nil), good.Tabs...)
		rec := *good.Record
		rec.Tabs = append([]AppTab(nil), good.Record.Tabs...)
		a.Record = &rec
		a.Actions = append([]string(nil), good.Actions...)
		a.Rules = append([]string(nil), good.Rules...)
		c.edit(&a)
		err := a.Validate()
		if err == nil {
			t.Errorf("%s: must be refused", c.why)
			continue
		}
		if !strings.Contains(err.Error(), "app") {
			t.Errorf("%s: the refusal names the App: %v", c.why, err)
		}
	}
}

// Its views in the order a person meets them: the App's tabs, then its
// record's.
func TestAppViews(t *testing.T) {
	a := App{Tabs: []AppTab{{View: "a"}, {Desk: "flow", Title: "F"}, {View: "b"}},
		Record: &AppRecord{Of: "x", Tabs: []AppTab{{View: "c"}}}}
	if got := strings.Join(a.Views(), ","); got != "a,b,c" {
		t.Fatalf("views: %s", got)
	}
}
