package shell_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"rhea/internal/pack"
)

// The catalog for Ustawienia → Pakiety (DIRECTION 2026-10-24, step 2): every
// approved pack with what it says it is and what it brings, and for the
// company acted for whether it is installed there — from when, until when,
// the record to switch it off by — or there for every company. One company's
// installs are not another's.
func TestTheCatalogSaysWhatIsInstalledHere(t *testing.T) {
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
	for _, p := range []map[string]any{
		{"pack": "base-test", "kind": "market", "market": "PL", "install": "company", "description": "the base"},
		{"pack": "kpir-test", "kind": "regime", "market": "PL", "install": "company", "needs": []string{"base-test"}, "description": "the book"},
	} {
		p["version"], p["stage"], p["disclaimer"] = 1, "initial", "A test pack."
		p["history"] = []map[string]any{{"versions": "1", "changes": "first"}}
		p["object_types"] = []map[string]any{{"name": p["pack"].(string)[:4] + "_thing_test", "version": 1, "domain": "work",
			"fields": []map[string]any{{"name": "n", "type": "string"}}}}
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
	w.person("Krzysztof Kowal", "kk", a, "owner")
	w.person("Olga Nowa", "olga", b, "owner")
	if code, out := w.door("kk", a, "install_pack", map[string]any{"pack": "base-test", "from": w.today}); code != 200 {
		t.Fatalf("install: %d %v", code, out)
	}

	type entry struct {
		Pack         string              `json:"pack"`
		Version      int                 `json:"version"`
		Kind         string              `json:"kind"`
		Install      string              `json:"install"`
		Needs        []string            `json:"needs"`
		Description  string              `json:"description"`
		EveryCompany bool                `json:"every_company"`
		Brings       struct{ Types int } `json:"brings"`
		Installed    *struct {
			Object string `json:"object"`
			From   string `json:"from"`
			To     string `json:"to"`
			Status string `json:"status"`
		} `json:"installed"`
	}
	catalog := func(person, company string) map[string]entry {
		t.Helper()
		var list []entry
		if code := w.call(person, company, "GET", "/api/packs", nil, &list); code != 200 {
			t.Fatalf("catalog for %s: %d", company, code)
		}
		out := map[string]entry{}
		for _, e := range list {
			out[e.Pack] = e
		}
		return out
	}
	ca := catalog("kk", a)
	if e := ca["base-test"]; e.Kind != "market" || e.Install != "company" || e.Description != "the base" || e.Brings.Types != 1 ||
		e.Installed == nil || e.Installed.From != w.today || e.Installed.Status != "installed" || e.Installed.Object == "" {
		t.Fatalf("alfa's base-test: %+v", e)
	}
	if e := ca["kpir-test"]; e.Installed != nil || len(e.Needs) != 1 || e.Needs[0] != "base-test" || e.EveryCompany {
		t.Fatalf("alfa's kpir-test: %+v", e)
	}
	if e := ca["core"]; !e.EveryCompany || e.Kind != "core" || e.Installed != nil {
		t.Fatalf("alfa's core: %+v", e)
	}
	if e := catalog("olga", b)["base-test"]; e.Installed != nil {
		t.Fatalf("beta's catalog shows alfa's install: %+v", e)
	}
}
