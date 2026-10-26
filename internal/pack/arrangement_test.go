package pack_test

import (
	"encoding/json"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/pack"
)

// TestThePolandPackArrangesItsApps: the Poland pack's words arrange the
// Activity Bar as Polish accountants work (KK 2026-10-26) — Dziś, Sprzedaż,
// Koszty, Banki, Księgowość i podatki, Środki trwałe, Ustawienia at the far
// end — and every App of the packs a Polish company has (the core, the
// common packs, Poland, its warehouse) is placed, none left for the rest to
// catch; every item opens something those packs define: an App, a tab it
// has, a kind of record one of them keeps.
func TestThePolandPackArrangesItsApps(t *testing.T) {
	apps := map[string]core.App{}
	var order []string
	for _, p := range []string{"core/core.json", "common/people.json", "common/periods.json", "common/documents.json", "pl/pack.json", "pl/warehouse.json"} {
		_, as, err := pack.ReadFile(filepath.Join("..", "..", "packs", p))
		if err != nil {
			t.Fatalf("%s: %v", p, err)
		}
		for _, a := range as {
			if _, ok := apps[a.ID]; !ok {
				order = append(order, a.ID)
			}
			apps[a.ID] = a // a later pack's version of an App over the earlier's
		}
	}
	kinds := map[string]bool{}
	for _, a := range apps {
		for _, k := range a.Data {
			kinds[k] = true
		}
	}
	m, _, err := pack.ReadFile(filepath.Join("..", "..", "packs", "pl", "pack.json"))
	if err != nil {
		t.Fatal(err)
	}
	var w core.WordsSpec
	for _, v := range m.ViewDefs {
		if v.Notion == "words" {
			if err := json.Unmarshal(v.Spec, &w); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := w.Check(); err != nil {
		t.Fatalf("the words: %v", err)
	}
	var bar []string
	for _, a := range w.Activities {
		bar = append(bar, a.Title)
	}
	if got := strings.Join(bar, ", "); got != "Dziś, Sprzedaż, Koszty, Banki, Księgowość i podatki, Środki trwałe, Ustawienia" {
		t.Fatalf("the Activity Bar: %s", got)
	}
	if last := w.Activities[len(w.Activities)-1]; !last.End || last.ID != "settings" {
		t.Fatalf("Ustawienia sits at the bar's far end: %+v", last)
	}
	placed := map[string]bool{}
	var wrong []string
	for _, a := range w.Activities {
		for _, s := range a.Sections {
			for _, it := range s.Items {
				if it.Kind != "" {
					if !kinds[it.Kind] {
						wrong = append(wrong, a.ID+": no App keeps "+it.Kind)
					}
					continue
				}
				app, ok := apps[it.App]
				if !ok {
					wrong = append(wrong, a.ID+": no App "+it.App)
					continue
				}
				placed[it.App] = true
				if it.Tab == "" {
					continue
				}
				var tabs []string
				for _, tb := range app.Tabs {
					if tb.Desk != "" {
						tabs = append(tabs, "desk:"+tb.Desk)
					} else {
						tabs = append(tabs, tb.View)
					}
				}
				if !slices.Contains(tabs, it.Tab) {
					wrong = append(wrong, a.ID+": "+it.App+" has no tab "+it.Tab)
				}
			}
		}
	}
	for _, id := range order {
		if !placed[id] {
			wrong = append(wrong, "App "+id+" is placed nowhere")
		}
	}
	if len(wrong) > 0 {
		t.Fatalf("the arrangement: %s", strings.Join(wrong, "; "))
	}
}
