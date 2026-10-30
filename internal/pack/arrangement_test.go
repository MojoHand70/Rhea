package pack_test

import (
	"encoding/json"
	"fmt"
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
	// and every App says where it sits and which faces show it (KK 2026-10-30),
	// agreeing with the arrangement: its activity holds an item of it, and
	// its section, where it names one, holds an item of it too
	for _, id := range order {
		a := apps[id]
		switch {
		case a.Place == nil:
			wrong = append(wrong, "App "+id+" says not where it sits")
			continue
		case len(a.Shows) == 0:
			wrong = append(wrong, "App "+id+" says not which faces show it")
		}
		inActivity, inSection := false, a.Place.Section == ""
		for _, act := range w.Activities {
			if act.ID != a.Place.Activity {
				continue
			}
			for _, sec := range act.Sections {
				for _, it := range sec.Items {
					if it.App != id {
						continue
					}
					inActivity = true
					if sec.ID == a.Place.Section {
						inSection = true
					}
				}
			}
		}
		if !inActivity || !inSection {
			wrong = append(wrong, fmt.Sprintf("App %s says it sits at %s/%s, where the arrangement does not put it", id, a.Place.Activity, a.Place.Section))
		}
	}
	// a section with no item of its own is a place Apps fill: the kept reports' Raporty
	var empty []string
	for _, act := range w.Activities {
		for _, sec := range act.Sections {
			if len(sec.Items) == 0 && !sec.Rest {
				empty = append(empty, act.ID+"/"+sec.ID)
			}
		}
	}
	if strings.Join(empty, ",") != "books/reports" {
		wrong = append(wrong, "the sections standing empty for Apps that place themselves: "+strings.Join(empty, ","))
	}
	if len(wrong) > 0 {
		t.Fatalf("the arrangement: %s", strings.Join(wrong, "; "))
	}

	// and the words name, in Polish, what those Apps keep and do: every kind
	// (what it is and its shape), field, value and door — a Polish company's
	// screens and forms never show the kernel's names (2026-10-27)
	var unnamed []string
	for _, id := range order {
		a := apps[id]
		for _, k := range a.Data {
			tw, ok := w.Types[k]
			if !ok || tw.Kind == "" || tw.Shape == "" {
				unnamed = append(unnamed, "kind "+k)
			}
		}
		for _, d := range a.Actions {
			if w.Doors[d] == "" {
				unnamed = append(unnamed, "door "+d)
			}
		}
	}
	for _, p := range []string{"core/core.json", "common/people.json", "common/periods.json", "common/documents.json", "pl/pack.json", "pl/warehouse.json"} {
		m, _, err := pack.ReadFile(filepath.Join("..", "..", "packs", p))
		if err != nil {
			t.Fatal(err)
		}
		for _, ty := range m.ObjectTypes {
			for _, f := range ty.Fields {
				if w.Fields[ty.Name+"."+f.Name] == "" && w.Fields[f.Name] == "" {
					unnamed = append(unnamed, "field "+ty.Name+"."+f.Name)
				}
				for _, v := range f.Values {
					if w.Values[ty.Name+"."+f.Name][v] == "" && w.Values[f.Name][v] == "" {
						unnamed = append(unnamed, "value "+ty.Name+"."+f.Name+"="+v)
					}
				}
			}
		}
	}
	if len(unnamed) > 0 {
		t.Fatalf("the Poland pack's words leave unnamed: %s", strings.Join(unnamed, ", "))
	}
}

// Under Ustawienia › Użytkownicy a person finds three screens, named as
// Polish business software names them — Użytkownicy, Grupy, Uprawnienia —
// and nothing else (KK 2026-10-30: seven entries, "Grupy" twice, "Miejsca w
// grupach", "Umocowania biur": "all the naming is confusing"). The raw
// record lists of the people pack are named in a hidden section, so the
// rest catches none of them and no Nav Bar draws them.
func TestTheUsersMenuIsThreeScreens(t *testing.T) {
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
	_, people, err := pack.ReadFile(filepath.Join("..", "..", "packs", "common", "people.json"))
	if err != nil || len(people) != 1 {
		t.Fatalf("the people pack's App: %v %d", err, len(people))
	}
	var settings *core.NavActivity
	for i := range w.Activities {
		if w.Activities[i].ID == "settings" {
			settings = &w.Activities[i]
		}
	}
	if settings == nil {
		t.Fatal("no Ustawienia")
	}
	var shown []string
	hidden := map[string]bool{}
	for _, s := range settings.Sections {
		for _, it := range s.Items {
			switch {
			case s.Hidden && it.Kind != "":
				hidden[it.Kind] = true
			case s.Hidden:
				t.Fatalf("a hidden section holds a screen, not a raw list: %+v", it)
			case s.ID == "users":
				shown = append(shown, it.App+" "+it.Tab+" → "+w.Views[it.Tab])
			case it.Kind != "":
				t.Fatalf("a raw list of records in the Nav Bar: %s under %s", it.Kind, s.Title)
			}
		}
	}
	if got := strings.Join(shown, "; "); got != "people desk:access → Użytkownicy; people desk:groups → Grupy; people who-can-do-what → Uprawnienia" {
		t.Fatalf("Użytkownicy holds: %s", got)
	}
	for _, kind := range people[0].Data {
		if !hidden[kind] {
			t.Fatalf("the people pack's %s is named in no hidden section: the rest would catch it", kind)
		}
	}
	for _, tb := range people[0].Tabs {
		if tb.Desk == "invites" {
			t.Fatal("invitations are a screen of their own; they belong on Użytkownicy")
		}
	}
}
