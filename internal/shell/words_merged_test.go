package shell_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"rhea/internal/pack"
)

// Words from every pack the company has, merged (DIRECTION 2026-10-24, step
// 3): the market's, then the regime's, the later naming a thing over the
// earlier; whichever words screen is read, it answers with the merge, and a
// company that has not installed the regime keeps the market's words alone.
func TestWordsAreMergedFromEveryPackTheCompanyHas(t *testing.T) {
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
	words := func(id string, spec map[string]any) map[string]any {
		return map[string]any{"view_id": id, "version": 1, "notion": "words", "title": "Słownik", "domain": "system", "function": "words",
			"spec": spec}
	}
	for _, p := range []map[string]any{
		{"pack": "market-test", "kind": "market", "market": "PL", "install": "installation",
			"object_types": []map[string]any{{"name": "thing_test", "version": 1, "domain": "work", "fields": []map[string]any{{"name": "n", "type": "string"}}}},
			"view_defs": []map[string]any{words("market-words", map[string]any{"lang": "pl",
				"types":  map[string]any{"thing_test": map[string]any{"label": "Rzecz", "shape": "card", "kind": "master"}},
				"fields": map[string]any{"n": "Numer"}})}},
		{"pack": "regime-test", "kind": "regime", "market": "PL", "install": "company",
			"view_defs": []map[string]any{words("regime-words", map[string]any{"lang": "pl",
				"types": map[string]any{"thing_test": map[string]any{"label": "Pozycja księgi", "shape": "card", "kind": "master"}},
				"notes": []string{"the regime's note"}})}},
	} {
		p["version"], p["stage"], p["disclaimer"], p["description"] = 1, "initial", "A test pack.", "a test pack"
		p["history"] = []map[string]any{{"versions": "1", "changes": "first"}}
		b, _ := json.Marshal(p)
		file := filepath.Join(t.TempDir(), "p.json")
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
	}
	a, b := "PL5250001111", "PL5260001246"
	w.person("Krzysztof Kowal", "kk", a, "company_manager")
	w.person("Olga Nowa", "olga", b, "company_manager")
	if code, out := w.door("kk", a, "install_pack", map[string]any{"pack": "regime-test", "from": w.today}); code != 200 {
		t.Fatalf("install: %d %v", code, out)
	}
	type served struct {
		Words struct {
			Types  map[string]struct{ Label string } `json:"types"`
			Fields map[string]string                 `json:"fields"`
			Notes  []string                          `json:"notes"`
		} `json:"words"`
	}
	read := func(person, company, id string) served {
		t.Helper()
		var out served
		if code := w.call(person, company, "GET", "/api/views/"+id, nil, &out); code != 200 {
			t.Fatalf("%s reads %s: %d", company, id, code)
		}
		return out
	}
	for _, id := range []string{"market-words", "regime-words"} {
		got := read("kk", a, id).Words
		if got.Types["thing_test"].Label != "Pozycja księgi" || got.Fields["n"] != "Numer" || len(got.Notes) != 1 {
			t.Fatalf("alfa's %s: %+v — want the regime's name over the market's, the market's field, the regime's note", id, got)
		}
	}
	if got := read("olga", b, "market-words").Words; got.Types["thing_test"].Label != "Rzecz" || len(got.Notes) != 0 {
		t.Fatalf("beta, without the regime, reads %+v — want the market's words alone", got)
	}
}
