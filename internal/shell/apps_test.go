package shell_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"rhea/internal/pack"
)

// A company's Apps (KK 2026-10-26, APPS.md "The App definition"): GET
// /api/apps serves each App of the packs the company has, with the tabs the
// person may see, each named, and where the App sits — its first tab's
// domain and function. A company that did not install the pack is served
// none of its Apps; a person who may see the App's first tab sees the App,
// without the tabs their role does not reach; an App whose first tab they
// may not see is not theirs at all.
func TestAppsAreTheCompanysAndThePersons(t *testing.T) {
	w := newGrantsWorld(t)
	sums, err := pack.LoadCore(w.ctx, w.s, "init")
	if err != nil {
		t.Fatal(err)
	}
	for _, sum := range sums {
		if sum.Bundle == "" {
			continue
		}
		if _, _, errs, err := w.x.ApproveBundle(w.ctx, sum.Bundle, "operator", w.today); err != nil || len(errs) != 0 {
			t.Fatalf("approve %s: %v %v", sum.Bundle, err, errs)
		}
	}
	tallyType := map[string]any{"name": "tally_test", "version": 1, "domain": "work", "label_field": "what",
		"fields": []map[string]any{{"name": "what", "type": "string", "required": true}}}
	list := map[string]any{"view_id": "tally-list", "version": 1, "notion": "list", "title": "Zliczenia", "domain": "work",
		"function": "documents", "spec": map[string]any{"object_type": "tally_test", "columns": []map[string]any{{"field": "what", "label": "Co"}}}}
	total := map[string]any{"view_id": "tally-total", "version": 1, "notion": "analysis", "title": "Zliczenia razem", "domain": "work",
		"function": "analysis", "spec": map[string]any{"sql": "SELECT COUNT(*) AS n FROM objects WHERE object_type = 'tally_test'",
			"columns": []map[string]any{{"field": "n", "label": "Ile", "type": "int"}}}}
	byWhat := map[string]any{"view_id": "tally-by-what", "version": 1, "notion": "analysis", "title": "Zliczenia według rodzaju", "domain": "work",
		"function": "analysis", "spec": map[string]any{"sql": "SELECT json_extract_string(state, '$.what') AS what FROM objects WHERE object_type = 'tally_test'",
			"columns": []map[string]any{{"field": "what", "label": "Co"}}}}
	apps := map[string]any{
		"liczenie": map[string]any{"app": "liczenie", "version": 1, "title": "Liczenie", "layer": "business",
			"data":   []any{tallyType},
			"tabs":   []any{list, total, map[string]any{"desk": "flow", "title": "Przepływ"}},
			"record": map[string]any{"of": "tally_test", "tabs": []any{map[string]any{"desk": "document-sheet", "title": "Zliczenie"}}},
			"actions": []any{map[string]any{"name": "count_tally", "domain": "work", "description": "Policz",
				"spec": map[string]any{"inputs": []map[string]any{{"name": "what", "type": "string", "required": true}},
					"emits": "tally.counted", "who": []string{"human"}}}},
			"rules": []any{map[string]any{"rule_id": "tally-counted", "priority": 100, "effective_from": "2026-01-01", "description": "Zliczenie",
				"spec": map[string]any{"match": map[string]any{"event_type": "tally.counted"},
					"effect": map[string]any{"object": map[string]any{"type": "tally_test", "fields": map[string]any{"what": "=$.what"}}}}}}},
		"liczenie-raport": map[string]any{"app": "liczenie-raport", "version": 1, "title": "Raport zliczeń", "layer": "business",
			"uses": []string{"tally_test"}, "tabs": []any{byWhat}},
	}
	dir := t.TempDir()
	write := func(name string, v any) {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("pack.json", map[string]any{"pack": "tally-apps", "version": 1, "kind": "module", "install": "company",
		"description": "a test pack of Apps", "stage": "initial", "disclaimer": "A test pack.",
		"history": []map[string]any{{"versions": "1", "changes": "first"}}, "apps": []string{"liczenie", "liczenie-raport"}})
	for name, a := range apps {
		write("apps/"+name+".json", a)
	}
	sum, err := pack.Load(w.ctx, w.s, filepath.Join(dir, "pack.json"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, errs, err := w.x.ApproveBundle(w.ctx, sum.Bundle, "operator", w.today); err != nil || len(errs) != 0 {
		t.Fatalf("approve: %v %v", err, errs)
	}

	a, b := "PL5250001111", "PL5260001246"
	w.person("Krzysztof Kowal", "kk", a, "owner")
	w.person("Olga Nowa", "olga", b, "owner")
	day, _ := time.Parse("2006-01-02", w.today)
	if code, out := w.door("kk", a, "install_pack", map[string]any{"pack": "tally-apps", "from": day.AddDate(0, 0, -7).Format("2006-01-02")}); code != 200 {
		t.Fatalf("kk installs the pack in alfa: %d %v", code, out)
	}
	w.mustDoor("kk", a, "declare_role", map[string]any{"name": "tally_clerk", "rights": "view:tally-list"})
	clerk := w.person("Celina Clerk", "celina", "", "tally_clerk")
	w.mustDoor("kk", a, "grant_membership", map[string]any{"person": clerk, "company": a, "role": "tally_clerk", "from": "2026-01-01", "granted_by": "kk"})

	type tab struct {
		View  string `json:"view"`
		Desk  string `json:"desk"`
		Title string `json:"title"`
	}
	type served struct {
		App      string `json:"app"`
		Title    string `json:"title"`
		Domain   string `json:"domain"`
		Function string `json:"function"`
		Tabs     []tab  `json:"tabs"`
		Record   *struct {
			Of   string `json:"of"`
			Tabs []tab  `json:"tabs"`
		} `json:"record"`
		Actions []string `json:"actions"`
		Rules   []string `json:"rules"`
	}
	appsOf := func(person, company string) map[string]served {
		t.Helper()
		var raw json.RawMessage
		if code := w.call(person, company, "GET", "/api/apps", nil, &raw); code != 200 {
			t.Fatalf("%s's Apps in %s: %d %s", person, company, code, raw)
		}
		var list []served
		if err := json.Unmarshal(raw, &list); err != nil || list == nil {
			t.Fatalf("%s's Apps in %s are a list: %s %v", person, company, raw, err)
		}
		out := map[string]served{}
		for _, x := range list {
			out[x.App] = x
		}
		return out
	}

	if got := appsOf("olga", b); len(got) != 0 {
		t.Fatalf("beta, which did not install the pack, is served its Apps: %+v", got)
	}

	got := appsOf("kk", a)
	l, ok := got["liczenie"]
	if !ok || l.Title != "Liczenie" || l.Domain != "work" || l.Function != "documents" {
		t.Fatalf("alfa's owner is served Liczenie, placed by its first tab: %+v", got)
	}
	wantTabs := []tab{{View: "tally-list", Title: "Zliczenia"}, {View: "tally-total", Title: "Zliczenia razem"}, {Desk: "flow", Title: "Przepływ"}}
	if !equalTabs(l.Tabs, wantTabs) || l.Record == nil || l.Record.Of != "tally_test" ||
		!equalTabs(l.Record.Tabs, []tab{{Desk: "document-sheet", Title: "Zliczenie"}}) {
		t.Fatalf("its tabs in order, each named, and its record's: %+v %+v", l.Tabs, l.Record)
	}
	if len(l.Actions) != 1 || l.Actions[0] != "count_tally" || len(l.Rules) != 1 || l.Rules[0] != "tally-counted" {
		t.Fatalf("what it consists of: %+v", l)
	}
	if r, ok := got["liczenie-raport"]; !ok || r.Function != "analysis" || !equalTabs(r.Tabs, []tab{{View: "tally-by-what", Title: "Zliczenia według rodzaju"}}) {
		t.Fatalf("and the report App: %+v", got)
	}

	got = appsOf("celina", a)
	if _, ok := got["liczenie-raport"]; ok {
		t.Fatalf("a clerk who may not see the report's first tab is served the report: %+v", got)
	}
	if !equalTabs(got["liczenie"].Tabs, []tab{{View: "tally-list", Title: "Zliczenia"}, {Desk: "flow", Title: "Przepływ"}}) {
		t.Fatalf("a clerk sees Liczenie without the tab her role does not reach: %+v", got["liczenie"].Tabs)
	}
}

func equalTabs[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
