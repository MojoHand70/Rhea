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
