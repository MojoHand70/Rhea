package pack_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"rhea/internal/core"
	"rhea/internal/pack"
	"rhea/internal/store"
)

// The parts of the notes pack, as a pack writes them today.
var (
	noteType = map[string]any{"name": "note", "version": 1, "domain": "work", "is_document": false, "label_field": "text",
		"amendable": []string{"text", "amount"},
		"fields":    []map[string]any{{"name": "text", "type": "string", "required": true}, {"name": "amount", "type": "money", "required": true}}}
	noteList = map[string]any{"view_id": "note-list", "version": 1, "notion": "list", "title": "Notatki", "domain": "work", "function": "documents",
		"spec": map[string]any{"object_type": "note", "columns": []map[string]any{{"field": "text", "label": "Treść"}, {"field": "amount", "label": "Kwota"}}}}
	notesSQL   = "SELECT COUNT(*) AS notes,\n       SUM(CAST(json_extract(state, '$.amount') AS BIGINT)) AS total\nFROM objects WHERE object_type = 'note'"
	notesTotal = func(sql map[string]any) map[string]any {
		spec := map[string]any{"columns": []map[string]any{{"field": "notes", "label": "Notatki", "type": "int"}, {"field": "total", "label": "Razem", "type": "money"}}}
		for k, v := range sql {
			spec[k] = v
		}
		return map[string]any{"view_id": "notes-total", "version": 2, "notion": "analysis", "title": "Notatki razem", "domain": "work", "function": "briefing", "spec": spec}
	}
	noteByText = map[string]any{"view_id": "notes-by-text", "version": 1, "notion": "analysis", "title": "Notatki według treści", "domain": "work", "function": "analysis",
		"spec": map[string]any{"sql": "SELECT json_extract_string(state, '$.text') AS text FROM objects WHERE object_type = 'note'",
			"columns": []map[string]any{{"field": "text", "label": "Treść"}}}}
	noteRule = map[string]any{"rule_id": "note", "priority": 100, "effective_from": "2026-01-01", "description": "Notatka staje się zapisem note",
		"spec": map[string]any{"match": map[string]any{"event_type": "note.taken"},
			"effect": map[string]any{"object": map[string]any{"type": "note", "fields": map[string]string{"text": "=$.text", "amount": "=$.amount"}}}}}
	noteDoor = map[string]any{"name": "take_note", "domain": "work", "description": "Zapisz notatkę",
		"spec": map[string]any{"inputs": []map[string]any{{"name": "text", "type": "string", "required": true}, {"name": "amount", "type": "money", "required": true}},
			"emits": "note.taken", "who": []string{"human"}}}
	noteEvent = map[string]any{"event_type": "note.category.declared", "occurred_at": "2026-01-01", "dedup_key": "notes/category/misc", "payload": map[string]any{"name": "misc"}}
)

func notesHeader(apps bool) map[string]any {
	m := map[string]any{"pack": "notes", "version": 1, "description": "Notatki", "stage": "initial", "disclaimer": "Pakiet testowy.",
		"history": []map[string]any{{"versions": "1", "changes": "pierwsza"}}, "events": []any{noteEvent}}
	if apps {
		m["apps"] = []string{"notatki", "podsumowanie"}
	} else {
		m["object_types"] = []any{noteType}
		m["view_defs"] = []any{noteList, notesTotal(map[string]any{"sql": notesSQL}), noteByText}
		m["rules"] = []any{noteRule}
		m["activities"] = []any{noteDoor}
	}
	return m
}

// The same parts as two Apps: Notatki keeps the note, its list, its total
// (the query in a file beside it), a tab the desk draws, the door and the
// rule; Podsumowanie only reads notes.
func notesApps() map[string]any {
	return map[string]any{
		"notatki": map[string]any{"app": "notatki", "version": 1, "title": "Notatki", "layer": "business",
			"data":    []any{noteType},
			"tabs":    []any{noteList, notesTotal(map[string]any{"sql_file": "notes-total.sql"}), map[string]any{"desk": "flow", "title": "Przepływ"}},
			"record":  map[string]any{"of": "note", "tabs": []any{map[string]any{"desk": "document-sheet", "title": "Notatka"}}},
			"actions": []any{noteDoor}, "rules": []any{noteRule},
			"place": map[string]any{"activity": "costs", "section": "notes"}, "shows": []string{"desk", "phone"}},
		"podsumowanie": map[string]any{"app": "podsumowanie", "version": 1, "title": "Podsumowanie notatek", "layer": "business",
			"uses": []string{"note"}, "tabs": []any{noteByText}, "shows": []string{"desk"}},
	}
}

// writePack writes a pack file and, beside it, its Apps and their queries.
func writePack(t *testing.T, header map[string]any, apps map[string]any) string {
	t.Helper()
	dir := t.TempDir()
	write := func(name string, v any) {
		b, ok := v.([]byte)
		if !ok {
			var err error
			if b, err = json.MarshalIndent(v, "", " "); err != nil {
				t.Fatal(err)
			}
		}
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("pack.json", header)
	for name, a := range apps {
		write("apps/"+name+".json", a)
	}
	if apps != nil {
		write("apps/notes-total.sql", []byte(notesSQL+"\n"))
	}
	return filepath.Join(dir, "pack.json")
}

// what a database holds of the notes pack, timestamps aside
type notesDefs struct {
	Types []core.ObjectType
	Views []core.ViewDef
	Rules []core.Rule
	Doors []core.Activity
}

func readNotes(t *testing.T, s *store.Store) notesDefs {
	t.Helper()
	ctx := context.Background()
	var d notesDefs
	types, err := s.ListObjectTypes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range types {
		if x.Name == "note" {
			d.Types = append(d.Types, x)
		}
	}
	views, err := s.LatestViewDefs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range views {
		if strings.HasPrefix(v.ID, "note") {
			var spec any // the same spec, however it was spelled
			if err := json.Unmarshal(v.Spec, &spec); err != nil {
				t.Fatal(err)
			}
			v.Spec, _ = json.Marshal(spec)
			d.Views = append(d.Views, v)
		}
	}
	rules, err := s.LatestRules(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rules {
		if r.ID == "note" {
			r.CreatedAt = time.Time{}
			d.Rules = append(d.Rules, r)
		}
	}
	doors, err := s.LatestActivities(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range doors {
		if a.Name == "take_note" {
			a.CreatedAt = time.Time{}
			d.Doors = append(d.Doors, a)
		}
	}
	return d
}

// A pack written as Apps unpacks into exactly the definitions its flat twin
// loads today (KK 2026-10-26, APPS.md "Order within step 1", step 2): the
// same types, views, rules and doors, the query read from its file; approved
// in one bundle with the Apps, which bind them by name. A reload drafts
// nothing; an App changed under the same version is refused.
func TestAppsUnpackIntoTodaysDefinitions(t *testing.T) {
	ctx := context.Background()
	flatS, flatX := withCore(t)
	flat, err := pack.Load(ctx, flatS, writePack(t, notesHeader(false), nil), "test")
	if err != nil {
		t.Fatal(err)
	}
	appsS, appsX := withCore(t)
	path := writePack(t, notesHeader(true), notesApps())
	sum, err := pack.Load(ctx, appsS, path, "test")
	if err != nil {
		t.Fatal(err)
	}
	if sum.Apps != 2 || sum.Types != flat.Types || sum.Views != flat.Views || sum.Rules != flat.Rules || sum.Activities != flat.Activities || sum.Events != flat.Events {
		t.Fatalf("the Apps load what the flat pack loads, and themselves: %+v against %+v", sum, flat)
	}
	if _, _, errs, err := flatX.ApproveBundle(ctx, flat.Bundle, "krzysztof", "2026-01-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve flat: %v %v", err, errs)
	}
	if _, _, errs, err := appsX.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-01-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve apps: %v %v", err, errs)
	}
	want, got := readNotes(t, flatS), readNotes(t, appsS)
	if len(want.Types) != 1 || len(want.Views) != 3 || len(want.Rules) != 1 || len(want.Doors) != 1 {
		t.Fatalf("the flat pack loaded: %+v", want)
	}
	if !reflect.DeepEqual(want, got) {
		wb, _ := json.MarshalIndent(want, "", " ")
		gb, _ := json.MarshalIndent(got, "", " ")
		t.Fatalf("unpacked differs from today's:\nwant %s\ngot  %s", wb, gb)
	}

	all, err := appsS.LatestApps(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var apps []core.App // the notes pack's, beside the core's own
	for _, a := range all {
		if a.ID == "notatki" || a.ID == "podsumowanie" {
			apps = append(apps, a)
		}
	}
	wantApps := []core.App{
		{ID: "notatki", Version: 1, Title: "Notatki", Layer: "business", Data: []string{"note"},
			Tabs:    []core.AppTab{{View: "note-list"}, {View: "notes-total"}, {Desk: "flow", Title: "Przepływ"}},
			Record:  &core.AppRecord{Of: "note", Tabs: []core.AppTab{{Desk: "document-sheet", Title: "Notatka"}}},
			Actions: []string{"take_note"}, Rules: []string{"note"},
			Place: &core.AppPlace{Activity: "costs", Section: "notes"}, Shows: []string{"desk", "phone"}},
		{ID: "podsumowanie", Version: 1, Title: "Podsumowanie notatek", Layer: "business", Uses: []string{"note"},
			Tabs: []core.AppTab{{View: "notes-by-text"}}, Shows: []string{"desk"}},
	}
	if !reflect.DeepEqual(apps, wantApps) {
		t.Fatalf("the Apps, approved: %+v", apps)
	}
	if flatApps, _ := flatS.LatestApps(ctx); slices.ContainsFunc(flatApps, func(a core.App) bool { return a.ID == "notatki" || a.ID == "podsumowanie" }) {
		t.Fatalf("a flat pack brings no App: %+v", flatApps)
	}
	decl, err := appsS.PackDeclarations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(decl["notes"], core.Member{Kind: core.KindApp, Name: "notatki", Version: 1}) {
		t.Fatalf("the pack declares its Apps: %+v", decl["notes"])
	}

	again, err := pack.Load(ctx, appsS, path, "test")
	if err != nil || again.Bundle != "" || again.Apps != 0 {
		t.Fatalf("a reload drafts nothing: %+v %v", again, err)
	}
	changed := notesApps()
	changed["notatki"].(map[string]any)["title"] = "Notatki firmy"
	if _, err := pack.Load(ctx, appsS, writePack(t, notesHeader(true), changed), "test"); err == nil || !strings.Contains(err.Error(), "bump") {
		t.Fatalf("an App changed under its version is refused: %v", err)
	}
}

// What a pack of Apps may not be: each refused before it drafts a thing.
func TestAppsPackRefusals(t *testing.T) {
	ctx := context.Background()
	s, _ := withCore(t)
	cases := []struct {
		why  string
		edit func(h map[string]any, apps map[string]any)
		says string
	}{
		{"a part in two Apps", func(h, apps map[string]any) {
			apps["podsumowanie"].(map[string]any)["tabs"] = []any{noteByText, noteList}
		}, "note-list"},
		{"a kind kept by two Apps", func(h, apps map[string]any) {
			p := apps["podsumowanie"].(map[string]any)
			p["data"], p["uses"] = []any{noteType}, nil
		}, "note"},
		{"a kind used that nothing keeps", func(h, apps map[string]any) {
			apps["podsumowanie"].(map[string]any)["uses"] = []string{"note", "invoice_line"}
		}, "invoice_line"},
		{"a rule writing a kind its App neither keeps nor uses", func(h, apps map[string]any) {
			n := apps["notatki"].(map[string]any)
			p := apps["podsumowanie"].(map[string]any)
			p["data"], p["uses"] = n["data"], nil
			n["data"], n["record"] = nil, nil
		}, "rule note"},
		{"a file naming another App", func(h, apps map[string]any) {
			apps["podsumowanie"].(map[string]any)["app"] = "raporty"
		}, "raporty"},
		{"an App listed with no file", func(h, apps map[string]any) {
			h["apps"] = []string{"notatki", "podsumowanie", "kasa"}
		}, "kasa"},
		{"loose parts beside the Apps", func(h, apps map[string]any) {
			h["view_defs"] = []any{noteByText}
		}, "in its Apps"},
		{"a part the format does not know", func(h, apps map[string]any) {
			apps["podsumowanie"].(map[string]any)["screens"] = []any{}
		}, "screens"},
		{"a query file that is not there", func(h, apps map[string]any) {
			apps["notatki"].(map[string]any)["tabs"] = []any{noteList, notesTotal(map[string]any{"sql_file": "missing.sql"})}
		}, "missing.sql"},
		{"a query twice", func(h, apps map[string]any) {
			apps["notatki"].(map[string]any)["tabs"] = []any{noteList, notesTotal(map[string]any{"sql_file": "notes-total.sql", "sql": "SELECT 1"})}
		}, "sql_file"},
		{"a tab both a view and the desk's", func(h, apps map[string]any) {
			v := map[string]any{"desk": "flow"}
			for k, x := range noteByText {
				v[k] = x
			}
			apps["podsumowanie"].(map[string]any)["tabs"] = []any{v}
		}, "both"},
	}
	for _, c := range cases {
		h, apps := notesHeader(true), notesApps()
		c.edit(h, apps)
		_, err := pack.Load(ctx, s, writePack(t, h, apps), "test")
		if err == nil || !strings.Contains(err.Error(), c.says) {
			t.Errorf("%s: want a refusal naming %q, got %v", c.why, c.says, err)
		}
	}
	if b, err := s.LatestBundles(ctx); err != nil || slices.ContainsFunc(b, func(x core.Bundle) bool { return x.ID == pack.BundleID("notes", 1) }) {
		t.Fatalf("nothing drafted: %v", err)
	}
	// a pack of Apps comes from its file, its Apps beside it
	b, _ := json.Marshal(notesHeader(true))
	if _, err := pack.LoadBytes(ctx, s, b, "test"); err == nil || !strings.Contains(err.Error(), "from its file") {
		t.Fatalf("bytes alone: %v", err)
	}
}

// Outside the loader, a pack written as Apps reads as one written before them
// (pack.ReadFile; rhea pack -flatten for the face's tools): the same parts,
// each where a pack keeps it, and its Apps beside them.
func TestReadFileFlattensApps(t *testing.T) {
	flat, none, err := pack.ReadFile(writePack(t, notesHeader(false), nil))
	if err != nil || len(none) != 0 {
		t.Fatalf("a flat pack: %v %v", none, err)
	}
	got, apps, err := pack.ReadFile(writePack(t, notesHeader(true), notesApps()))
	if err != nil {
		t.Fatal(err)
	}
	if len(apps) != 2 || apps[0].ID != "notatki" || apps[1].ID != "podsumowanie" {
		t.Fatalf("its Apps, in order: %+v", apps)
	}
	same := func(a, b any) bool {
		var x, y any
		ja, _ := json.Marshal(a)
		jb, _ := json.Marshal(b)
		json.Unmarshal(ja, &x)
		json.Unmarshal(jb, &y)
		return reflect.DeepEqual(x, y)
	}
	type parts struct {
		T, V, R, A, E any
	}
	viewSpecs := func(vs []core.ViewDef) []any {
		var out []any
		for _, v := range vs {
			var spec any
			json.Unmarshal(v.Spec, &spec)
			out = append(out, map[string]any{"id": v.ID, "version": v.Version, "notion": v.Notion, "title": v.Title, "domain": v.Domain, "function": v.Function, "spec": spec})
		}
		return out
	}
	if !same(parts{flat.ObjectTypes, viewSpecs(flat.ViewDefs), flat.Rules, flat.Activities, flat.Events},
		parts{got.ObjectTypes, viewSpecs(got.ViewDefs), got.Rules, got.Activities, got.Events}) {
		t.Fatalf("the parts differ from the flat pack's:\n%+v\n%+v", flat, got)
	}
}

// A pack's App may show another pack's screen by naming it, {"view": id}: the
// screen stays the other pack's, declared by it alone, and the App lists it
// among its tabs. Refused: a name nothing has defined, a name of a screen of
// the same pack (that screen is one of its Apps' own), a name carrying a
// view's parts as well.
func TestAppNamesAnotherPacksScreen(t *testing.T) {
	ctx := context.Background()
	s, x := withCore(t)
	apps := notesApps()
	apps["podsumowanie"].(map[string]any)["tabs"] = []any{noteByText, map[string]any{"view": "account-list"}}
	sum, err := pack.Load(ctx, s, writePack(t, notesHeader(true), apps), "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-01-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve: %v %v", err, errs)
	}
	a, _, err := s.GetAppVersion(ctx, "podsumowanie", 1)
	if err != nil || len(a.Tabs) != 2 || a.Tabs[1] != (core.AppTab{View: "account-list"}) {
		t.Fatalf("the App lists the core's screen: %+v %v", a, err)
	}
	decl, err := s.PackDeclarations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(decl["notes"], core.Member{Kind: core.KindViewDef, Name: "account-list", Version: 1}) {
		t.Fatalf("a screen named is not declared by the pack naming it: %+v", decl["notes"])
	}

	s2, _ := withCore(t)
	for _, c := range []struct {
		why  string
		tab  map[string]any
		says string
	}{
		{"a screen nothing defines", map[string]any{"view": "no-such-list"}, "no-such-list"},
		{"a screen of the same pack", map[string]any{"view": "note-list"}, "note-list"},
		{"a name carrying a view's parts", map[string]any{"view": "account-list", "notion": "list"}, "account-list"},
	} {
		apps := notesApps()
		apps["podsumowanie"].(map[string]any)["tabs"] = []any{noteByText, c.tab}
		if _, err := pack.Load(ctx, s2, writePack(t, notesHeader(true), apps), "test"); err == nil || !strings.Contains(err.Error(), c.says) {
			t.Errorf("%s: want a refusal naming %q, got %v", c.why, c.says, err)
		}
	}
}
