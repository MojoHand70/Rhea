package pack_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"rhea/internal/core"
	"rhea/internal/store"
)

// An install dated back meets the facts the company recorded since its day
// (DIRECTION 2026-10-24, step 2) in two ways. A fact no rule explained waits
// in the worklist, and the install books it at once, as the arrival of any
// rule books what waits for it. A fact another rule already explained — a
// bank line the core books, which the new pack would book further — waits
// for the backfill the company approves for itself, which books exactly the
// pack's part of those since the day: never a fact dated before it, never
// another company's; a second backfill has nothing left to book.
func TestABackDatedInstallBooksThePastWhenApproved(t *testing.T) {
	ctx := context.Background()
	s, x := withCore(t)
	// every company's log keeps each thing recorded (a pack for the whole installation)
	approveFile(t, s, x, map[string]any{"pack": "log-test", "kind": "module", "install": "installation",
		"rules": []map[string]any{{"rule_id": "log-test-entry", "priority": 100, "effective_from": "2026-01-01",
			"description": "every thing recorded is logged",
			"spec": map[string]any{"match": map[string]any{"event_type": "thing.recorded"},
				"effect": map[string]any{"object": map[string]any{"type": "log_test_entry", "fields": map[string]any{"n": "=$.n"}}}}}}},
		"log_test_entry")
	approveFile(t, s, x, map[string]any{"pack": "base-test", "kind": "market", "market": "PL", "install": "company",
		"rules": []map[string]any{{"rule_id": "base-test-thing", "priority": 100, "effective_from": "2026-01-01",
			"description": "a recorded thing is a thing",
			"spec": map[string]any{"match": map[string]any{"event_type": "thing.recorded"},
				"effect": map[string]any{"object": map[string]any{"type": "base_test_thing", "fields": map[string]any{"n": "=$.n"}}}}},
			{"rule_id": "base-test-tally", "priority": 100, "effective_from": "2026-01-01",
				"description": "a recorded tally is a thing too",
				"spec": map[string]any{"match": map[string]any{"event_type": "tally.recorded"},
					"effect": map[string]any{"object": map[string]any{"type": "base_test_thing", "fields": map[string]any{"n": "=$.n"}}}}}}},
		"base_test_thing")
	const a, b = "PL5250001111", "PL5260001246"
	record := func(company, typ, n, date string) {
		t.Helper()
		p, _ := json.Marshal(map[string]any{"n": n})
		if _, err := s.AppendEvent(store.WithOwner(ctx), core.Event{Kind: core.KindRaw, Type: typ, OccurredAt: date, Payload: p,
			DedupKey: fmt.Sprintf("thing/%s/%s", company, n), Company: company}); err != nil {
			t.Fatal(err)
		}
	}
	record(a, "thing.recorded", "before-the-day", "2026-02-15")
	record(a, "thing.recorded", "since-the-day", "2026-03-10")
	record(b, "thing.recorded", "beta-since-the-day", "2026-03-10")
	record(a, "tally.recorded", "waiting", "2026-03-12") // no rule explains it yet
	x.ProcessPending(store.WithOwner(ctx))
	alfa := store.WithCompany(ctx, a)
	if _, err := x.TriggerActivity(alfa, "install_pack", map[string]any{"pack": "base-test", "from": "2026-03-01", "installed_by": "krzysztof"}, "human", ""); err != nil {
		t.Fatal(err)
	}
	things := func() []string {
		t.Helper()
		objs, err := s.ObjectsByType(store.WithOwner(ctx), "base_test_thing")
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, o := range objs {
			out = append(out, o.Company+":"+o.State["n"].(string))
		}
		sort.Strings(out)
		return out
	}
	if got := fmt.Sprint(things()); got != "["+a+":waiting]" {
		t.Fatalf("installing booked %s, want only alfa's fact that waited for a rule", got)
	}
	plan, err := x.PlanBackfill(alfa, "2026-10-24")
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Chains) != 1 || plan.Chains[0].OccurredAt != "2026-03-10" {
		t.Fatalf("alfa's past to book: %+v", plan.Chains)
	}
	if n, errs, err := x.ApproveBackfillPlan(alfa, "krzysztof", "2026-10-24", plan.Fingerprint); err != nil || len(errs) != 0 || n != 1 {
		t.Fatalf("approve alfa's past: %d %v %v", n, errs, err)
	}
	if got := fmt.Sprint(things()); got != "["+a+":since-the-day "+a+":waiting]" {
		t.Fatalf("booked %s, want alfa's facts since the day alone", got)
	}
	if again, err := x.PlanBackfill(alfa, "2026-10-24"); err != nil || len(again.Chains) != 0 {
		t.Fatalf("a second backfill still has %v to book (%v)", again.Chains, err)
	}
	assertReplayIdentical(t, x)
}
