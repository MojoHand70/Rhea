package shell_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"rhea/internal/pack"
	"rhea/internal/store"
)

// What a company is served (DIRECTION 2026-10-24, step 2): a pack installed
// per company is offered to the companies that installed it — its screens and
// types, its rules on the rules page, its doors — and to no other; a door of
// it is refused, before anyone's rights are asked, where it is not installed.
// Switched off, the pack's screens and types stay (what it booked stays
// readable) and its doors go.
func TestAPackNotInstalledIsNotServed(t *testing.T) {
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
	tally := map[string]any{"pack": "tally-test", "version": 1, "kind": "module", "install": "company",
		"description": "a test pack", "stage": "initial", "disclaimer": "A test pack.",
		"history": []map[string]any{{"versions": "1", "changes": "first"}},
		"object_types": []map[string]any{{"name": "tally_test", "version": 1, "domain": "work", "label_field": "what",
			"fields": []map[string]any{{"name": "what", "type": "string", "required": true}}}},
		"view_defs": []map[string]any{{"view_id": "tally-list", "version": 1, "notion": "list", "title": "Tallies", "domain": "work",
			"function": "list", "spec": map[string]any{"object_type": "tally_test", "columns": []map[string]any{{"field": "what", "label": "What"}}}}},
		"rules": []map[string]any{{"rule_id": "tally-counted", "priority": 100, "effective_from": "2026-01-01", "description": "a tally",
			"spec": map[string]any{"match": map[string]any{"event_type": "tally.counted"},
				"effect": map[string]any{"object": map[string]any{"type": "tally_test", "fields": map[string]any{"what": "=$.what"}}}}}},
		"activities": []map[string]any{{"name": "count_tally", "domain": "work", "description": "count a tally",
			"spec": map[string]any{"inputs": []map[string]any{{"name": "what", "type": "string", "required": true}},
				"emits": "tally.counted", "who": []string{"human"}}}},
	}
	b, _ := json.Marshal(tally)
	file := filepath.Join(t.TempDir(), "tally.json")
	if err := os.WriteFile(file, b, 0o644); err != nil {
		t.Fatal(err)
	}
	sum, err := pack.Load(w.ctx, w.s, file, "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, errs, err := w.x.ApproveBundle(w.ctx, sum.Bundle, "operator", w.today); err != nil || len(errs) != 0 {
		t.Fatalf("approve: %v %v", err, errs)
	}
	a, b2 := "PL5250001111", "PL5260001246"
	w.person("Krzysztof Kowal", "kk", a, "company_manager")
	w.person("Olga Nowa", "olga", b2, "company_manager")
	day, _ := time.Parse("2006-01-02", w.today)
	yesterday := day.AddDate(0, 0, -1).Format("2006-01-02")
	if code, out := w.door("kk", a, "install_pack", map[string]any{"pack": "tally-test", "from": day.AddDate(0, 0, -7).Format("2006-01-02")}); code != 200 {
		t.Fatalf("kk installs the pack in alfa: %d %v", code, out)
	}

	served := func(person, company string) (views, types, rules, doors []string) {
		t.Helper()
		var nav []struct {
			Functions []struct {
				Views []struct {
					ID string `json:"view_id"`
				} `json:"views"`
			} `json:"functions"`
		}
		w.call(person, company, "GET", "/api/nav", nil, &nav)
		for _, d := range nav {
			for _, f := range d.Functions {
				for _, v := range f.Views {
					views = append(views, v.ID)
				}
			}
		}
		var lang []struct {
			Name       string `json:"name"`
			ProducedBy []struct {
				ID string `json:"rule_id"`
			} `json:"produced_by"`
		}
		w.call(person, company, "GET", "/api/types", nil, &lang)
		for _, ty := range lang {
			types = append(types, ty.Name)
			for _, r := range ty.ProducedBy {
				rules = append(rules, r.ID)
			}
		}
		var acts []struct {
			Name    string `json:"name"`
			Offered bool   `json:"offered"`
		}
		w.call(person, company, "GET", "/api/activities", nil, &acts)
		for _, ac := range acts {
			if ac.Offered {
				doors = append(doors, ac.Name)
			}
		}
		return
	}
	has := func(xs []string, x string) bool { return slices.Contains(xs, x) }

	v, ty, r, d := served("olga", b2)
	if has(v, "tally-list") || has(ty, "tally_test") || has(r, "tally-counted") || has(d, "count_tally") {
		t.Fatalf("beta, which did not install the pack, is served it: views %v types %v rules %v doors %v", v, ty, r, d)
	}
	if code := w.call("olga", b2, "GET", "/api/views/tally-list", nil, nil); code != 404 {
		t.Fatalf("beta reads the pack's screen: %d", code)
	}
	if code, out := w.door("olga", b2, "count_tally", map[string]any{"what": "one"}); code == 200 || !strings.Contains(strings.ToLower(jsonText(out)), "tally-test") {
		t.Fatalf("beta opens the pack's door: %d %v", code, out)
	}

	v, ty, r, d = served("kk", a)
	if !has(v, "tally-list") || !has(ty, "tally_test") || !has(r, "tally-counted") || !has(d, "count_tally") {
		t.Fatalf("alfa, which installed the pack, is not served it: views %v types %v rules %v doors %v", v, ty, r, d)
	}
	w.mustDoor("kk", a, "count_tally", map[string]any{"what": "one"})

	installs, err := w.s.ObjectsByType(store.WithCompany(w.ctx, a), "installed_pack")
	if err != nil || len(installs) != 1 {
		t.Fatalf("alfa's installs: %v %v", installs, err)
	}
	if code, out := w.door("kk", a, "switch_off_pack", map[string]any{"installed": installs[0].ID, "to": yesterday}); code != 200 {
		t.Fatalf("kk switches the pack off: %d %v", code, out)
	}
	v, ty, _, d = served("kk", a)
	if !has(v, "tally-list") || !has(ty, "tally_test") || has(d, "count_tally") {
		t.Fatalf("switched off, alfa is served views %v types %v doors %v — want the screens and types, not the door", v, ty, d)
	}
}

func jsonText(v any) string { b, _ := json.Marshal(v); return string(b) }
