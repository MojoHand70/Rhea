package core

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestWordsSayWhatAThingIs: a market's words say what each type is — a
// document, an entry, or master data — so a face can file its lists without
// knowing the business, and give the title, in the market's language, of a
// view another pack brought in another (DIRECTION 2026-10-22, "The desk in
// order"). A kind outside the three, or an empty title, is refused.
func TestWordsSayWhatAThingIs(t *testing.T) {
	var w WordsSpec
	if err := json.Unmarshal([]byte(`{"lang": "pl",
		"types": {"sales_invoice": {"label": "Faktura sprzedaży", "shape": "paper", "kind": "document"},
		          "posting": {"label": "Zapis księgowy", "kind": "entry"}, "vat_rate": {"label": "Stawka VAT", "kind": "master"}},
		"views": {"trial-balance": "Zestawienie obrotów i sald"}}`), &w); err != nil {
		t.Fatal(err)
	}
	if err := w.Check(); err != nil {
		t.Fatalf("words with kinds and titles: %v", err)
	}
	if w.Types["sales_invoice"].Kind != "document" || w.Types["posting"].Kind != "entry" || w.Types["vat_rate"].Kind != "master" {
		t.Fatalf("kinds: %+v", w.Types)
	}
	if w.Views["trial-balance"] != "Zestawienie obrotów i sald" {
		t.Fatalf("titles: %+v", w.Views)
	}
	w.Types["posting"] = TypeWords{Label: "Zapis księgowy", Kind: "thing"}
	if err := w.Check(); err == nil || !strings.Contains(err.Error(), "thing") {
		t.Fatalf("a kind outside the three is refused: %v", err)
	}
	w.Types["posting"] = TypeWords{Label: "Zapis księgowy", Kind: "entry"}
	w.Views["journal"] = " "
	if err := w.Check(); err == nil {
		t.Fatal("an empty title is refused")
	}
}

// TestWordsNameAnApp: a market's words give, in its language, the title of an
// App another pack brought in another (2026-10-26) — the core's Journal is
// Dziennik where the Poland pack is — as they give a view's; a later pack's
// words name it over an earlier's, and an empty title is refused.
func TestWordsNameAnApp(t *testing.T) {
	var w WordsSpec
	if err := json.Unmarshal([]byte(`{"lang": "pl", "apps": {"journal": "Dziennik", "trial-balance": "Zestawienie obrotów i sald"}}`), &w); err != nil {
		t.Fatal(err)
	}
	if err := w.Check(); err != nil || w.Apps["journal"] != "Dziennik" {
		t.Fatalf("an App named: %+v %v", w.Apps, err)
	}
	base := WordsSpec{Lang: "en", Apps: map[string]string{"journal": "Journal", "packs": "Packs"}}
	base.Merge(w)
	if base.Apps["journal"] != "Dziennik" || base.Apps["packs"] != "Packs" || base.Apps["trial-balance"] != "Zestawienie obrotów i sald" {
		t.Fatalf("the later pack names it over the earlier: %+v", base.Apps)
	}
	w.Apps["packs"] = " "
	if err := w.Check(); err == nil || !strings.Contains(err.Error(), "packs") {
		t.Fatalf("an empty title is refused: %v", err)
	}
}

// TestWordsArrangeTheApps: a market's words arrange its Apps on a face as its
// people work (KK 2026-10-26): the Activity Bar's activities in order, each
// with its Nav Bar's sections, each section its items — an App, an App on one
// of its tabs, or the list of one kind of record. A later pack's words add to
// an activity of the same id; an arrangement a face could not draw is refused.
func TestWordsArrangeTheApps(t *testing.T) {
	var w WordsSpec
	if err := json.Unmarshal([]byte(`{"lang": "pl", "activities": [
		{"id": "sales", "title": "Sprzedaż", "icon": "sales", "key": "S", "sections": [
			{"items": [{"app": "sales-invoices"}, {"app": "sales-invoices", "tab": "desk:flow", "title": "Status faktur"}]},
			{"id": "reports", "title": "Raporty", "items": [{"app": "reports", "tab": "naleznosci-wiek"}]}]},
		{"id": "books", "title": "Księgowość i podatki", "icon": "books", "sections": [
			{"id": "parameters", "title": "Słowniki i parametry", "items": [{"app": "vat-rates"}, {"kind": "cit_limit"}]},
			{"id": "rest", "title": "Pozostałe", "rest": true}]},
		{"id": "settings", "title": "Ustawienia", "icon": "settings", "end": true, "sections": [{"items": [{"app": "people"}]}]}]}`), &w); err != nil {
		t.Fatal(err)
	}
	if err := w.Check(); err != nil {
		t.Fatalf("an arrangement: %v", err)
	}
	if len(w.Activities) != 3 || w.Activities[0].Sections[0].Items[1].Tab != "desk:flow" || !w.Activities[2].End || !w.Activities[1].Sections[1].Rest {
		t.Fatalf("read: %+v", w.Activities)
	}
	// a section with an id and a title may hold no item of its own: the Apps
	// that say they sit there fill it (KK 2026-10-30, the kept reports'
	// Raporty); one with no name for Apps to point at holds nothing
	empty := w
	empty.Activities = append([]NavActivity{{ID: "books", Title: "Księgi", Sections: []NavSection{{ID: "reports", Title: "Raporty"}}}}, w.Activities[2:]...)
	if err := empty.Check(); err != nil {
		t.Fatalf("a named section standing empty for Apps: %v", err)
	}
	empty.Activities[0].Sections[0].Title = ""
	if err := empty.Check(); err == nil {
		t.Fatal("a nameless empty section was taken")
	}

	// a later pack adds to an activity of the same id, to a section of the same
	// id, and brings an activity of its own; what is there already stays once
	later := WordsSpec{Lang: "pl", Activities: []NavActivity{
		{ID: "sales", Title: "Sprzedaż", Sections: []NavSection{{ID: "reports", Items: []NavItem{{App: "reports", Tab: "naleznosci-wiek"}, {App: "sales-by-customer"}}}}},
		{ID: "warehouse", Title: "Magazyn", Icon: "assets", Sections: []NavSection{{Items: []NavItem{{App: "goods-receipts"}}}}}}}
	if err := later.Check(); err != nil {
		t.Fatalf("a later pack's arrangement: %v", err)
	}
	before := len(w.Activities[0].Sections[1].Items)
	w.Merge(later)
	sales := w.Activities[0]
	if sales.Title != "Sprzedaż" || sales.Icon != "sales" || len(sales.Sections) != 2 {
		t.Fatalf("the later pack keeps what it does not name: %+v", sales)
	}
	if got := sales.Sections[1].Items; len(got) != 2 || got[1].App != "sales-by-customer" {
		t.Fatalf("the later pack's item joins the section, the one already there stays once: %+v", got)
	}
	if len(later.Activities[0].Sections[0].Items) != 2 || before != 1 {
		t.Fatal("merging changes neither pack's own words")
	}
	if w.Activities[3].ID != "warehouse" {
		t.Fatalf("a new activity joins the bar: %+v", w.Activities)
	}

	refused := map[string]string{
		"two activities of one id":      `[{"id": "sales", "title": "A", "sections": [{"items": [{"app": "x"}]}]}, {"id": "sales", "title": "B", "sections": [{"items": [{"app": "y"}]}]}]`,
		"an activity with no title":     `[{"id": "sales", "title": " ", "sections": [{"items": [{"app": "x"}]}]}]`,
		"an id that is no name":         `[{"id": "Sprzedaż", "title": "Sprzedaż", "sections": [{"items": [{"app": "x"}]}]}]`,
		"an activity with nothing":      `[{"id": "sales", "title": "Sprzedaż"}]`,
		"a section with nothing":        `[{"id": "sales", "title": "Sprzedaż", "sections": [{"title": "Raporty"}]}]`,
		"an item naming nothing":        `[{"id": "sales", "title": "Sprzedaż", "sections": [{"items": [{"title": "Coś"}]}]}]`,
		"an item both App and kind":     `[{"id": "sales", "title": "Sprzedaż", "sections": [{"items": [{"app": "x", "kind": "y"}]}]}]`,
		"a tab with no App":             `[{"id": "sales", "title": "Sprzedaż", "sections": [{"items": [{"kind": "y", "tab": "z"}]}]}]`,
		"a key of two letters":          `[{"id": "sales", "title": "Sprzedaż", "key": "SP", "sections": [{"items": [{"app": "x"}]}]}]`,
		"two sections for the rest":     `[{"id": "a", "title": "A", "sections": [{"rest": true}]}, {"id": "b", "title": "B", "sections": [{"rest": true}]}]`,
		"one item twice in its section": `[{"id": "sales", "title": "Sprzedaż", "sections": [{"items": [{"app": "x"}, {"app": "x"}]}]}]`,
	}
	for what, acts := range refused {
		var bad WordsSpec
		if err := json.Unmarshal([]byte(`{"lang": "pl", "activities": `+acts+`}`), &bad); err != nil {
			t.Fatalf("%s: %v", what, err)
		}
		if err := bad.Check(); err == nil {
			t.Errorf("%s: not refused", what)
		}
	}
}

// TestWordsOfferQuestionsToStart: the questions a face offers before a
// person has asked any are a market's words, in its language, by what is in
// front (APPS.md, step 5) — the face knows none. A later pack's join an
// earlier's, each once; an empty question is refused.
func TestWordsOfferQuestionsToStart(t *testing.T) {
	var w WordsSpec
	if err := json.Unmarshal([]byte(`{"lang": "pl", "asks": {"today": ["Co dziś czeka na moją decyzję?"], "list:sales_invoice": ["Które faktury są po terminie?"]}}`), &w); err != nil {
		t.Fatal(err)
	}
	if err := w.Check(); err != nil || len(w.Asks["today"]) != 1 {
		t.Fatalf("questions to start with: %+v %v", w.Asks, err)
	}
	w.Merge(WordsSpec{Lang: "pl", Asks: map[string][]string{"today": {"Co dziś czeka na moją decyzję?", "Ile mam na rachunku?"}}})
	if got := w.Asks["today"]; len(got) != 2 || got[1] != "Ile mam na rachunku?" || len(w.Asks["list:sales_invoice"]) != 1 {
		t.Fatalf("a later pack's questions join, each once: %+v", w.Asks)
	}
	for what, bad := range map[string]WordsSpec{
		"an empty question":         {Lang: "pl", Asks: map[string][]string{"today": {" "}}},
		"questions about no screen": {Lang: "pl", Asks: map[string][]string{"": {"Ile mam na rachunku?"}}},
	} {
		if err := bad.Check(); err == nil {
			t.Errorf("%s: not refused", what)
		}
	}
}
