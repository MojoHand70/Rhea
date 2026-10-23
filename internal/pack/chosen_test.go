package pack_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/store"
)

// Rules chosen by install (DIRECTION 2026-10-24, step 2): a rule of a pack
// installed per company books a company's fact only if the company had the
// pack installed on the fact's date — not before the install, not after it
// was switched off, never for a company that did not install it — and what
// it booked stays when the pack is switched off. The installation's own facts
// (master data, no company) meet every approved pack's rules, as before. A
// replay reproduces it all.
func TestRulesAreChosenByWhatTheCompanyInstalled(t *testing.T) {
	ctx := context.Background()
	s, x := withCore(t)
	approveFile(t, s, x, map[string]any{"pack": "base-test", "kind": "market", "market": "PL", "install": "company",
		"rules": []map[string]any{{"rule_id": "base-test-thing", "priority": 100, "effective_from": "2026-01-01",
			"description": "a recorded thing is a thing",
			"spec": map[string]any{"match": map[string]any{"event_type": "thing.recorded"},
				"effect": map[string]any{"object": map[string]any{"type": "base_test_thing", "fields": map[string]any{"n": "=$.n"}}}}}}},
		"base_test_thing")
	const a, b = "PL5250001111", "PL5260001246"
	alfa := store.WithCompany(ctx, a)
	if _, err := x.TriggerActivity(alfa, "install_pack", map[string]any{"pack": "base-test", "from": "2026-03-01", "installed_by": "krzysztof"}, "human", ""); err != nil {
		t.Fatal(err)
	}
	objs, err := s.ObjectsByType(alfa, "installed_pack")
	if err != nil || len(objs) != 1 {
		t.Fatalf("installs: %v %v", objs, err)
	}
	if _, err := x.TriggerActivity(alfa, "switch_off_pack", map[string]any{"installed": objs[0].ID, "to": "2026-06-30", "switched_off_by": "krzysztof"}, "human", ""); err != nil {
		t.Fatal(err)
	}

	record := func(company, n, date string) {
		t.Helper()
		p, _ := json.Marshal(map[string]any{"n": n})
		ev := core.Event{Kind: core.KindRaw, Type: "thing.recorded", OccurredAt: date, Payload: p,
			DedupKey: fmt.Sprintf("thing/%s/%s", company, n), Company: company}
		if _, err := s.AppendEvent(store.WithOwner(ctx), ev); err != nil {
			t.Fatal(err)
		}
	}
	record(a, "before-install", "2026-02-28")
	record(a, "installed", "2026-04-15")
	record(a, "last-day", "2026-06-30")
	record(a, "after-switch-off", "2026-07-01")
	record(b, "not-installed", "2026-04-15")
	record("", "installation", "2026-04-15")
	if _, errs := x.ProcessPending(store.WithOwner(ctx)); len(errs) != 0 {
		t.Fatalf("processing: %v", errs)
	}

	booked := func() []string {
		t.Helper()
		things, err := s.ObjectsByType(store.WithOwner(ctx), "base_test_thing")
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, o := range things {
			out = append(out, o.Company+":"+o.State["n"].(string))
		}
		sort.Strings(out)
		return out
	}
	want := fmt.Sprint([]string{":installation", a + ":installed", a + ":last-day"})
	if got := fmt.Sprint(booked()); got != want {
		t.Fatalf("booked %s, want %s", got, want)
	}
	assertReplayIdentical(t, x)
	if got := fmt.Sprint(booked()); got != want {
		t.Fatalf("after replay booked %s, want %s", got, want)
	}
}

// approveFile loads a small pack and approves it.
func approveFile(t *testing.T, s *store.Store, x *exec.Executor, header map[string]any, typ string) {
	t.Helper()
	ctx := context.Background()
	sum, err := pack.Load(ctx, s, packFile(t, header, typ), "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-01-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve %s: %v %v", sum.Bundle, err, errs)
	}
}

func assertReplayIdentical(t *testing.T, x *exec.Executor) {
	t.Helper()
	ctx := context.Background()
	state := func() string {
		objs, err := x.Store.AllObjects(ctx)
		if err != nil {
			t.Fatal(err)
		}
		m := map[string]map[string]any{}
		for _, o := range objs {
			m[o.ID] = o.State
		}
		b, _ := json.Marshal(m)
		return string(b)
	}
	before := state()
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	if after := state(); after != before {
		t.Fatalf("replay diverged:\n%s\n%s", before, after)
	}
}
