package pack_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/store/storetest"
)

// A pack upgrade (DECISIONS 2026-10-15): a newer pack moves its own
// definitions forward — a changed rule or activity becomes the next draft
// version in an upgrade bundle behind the gate — leaves an unchanged one
// alone, and reports a definition a person evolved since as a conflict
// rather than overriding it. Approving the bundle supersedes the old
// versions the way a rule approval always has.
func TestPackUpgrade(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	dir := t.TempDir()
	write := func(version int, ruleBDesc, doorDesc string) string {
		m := map[string]any{
			"pack": "toy", "version": version, "description": "a toy market",
			"object_types": []any{map[string]any{"name": "toy", "version": 1, "domain": "work", "label_field": "name",
				"fields": []any{map[string]any{"name": "name", "type": "string", "required": true}}}},
			"rules": []any{
				map[string]any{"rule_id": "toy-a", "priority": 100, "effective_from": "2026-01-01", "description": "a stays",
					"spec": map[string]any{"match": map[string]any{"event_type": "toy.a"}, "effect": map[string]any{"object": map[string]any{"type": "toy", "fields": map[string]any{"name": "=$.name"}}}}},
				map[string]any{"rule_id": "toy-b", "priority": 100, "effective_from": "2026-01-01", "description": ruleBDesc,
					"spec": map[string]any{"match": map[string]any{"event_type": "toy.b"}, "effect": map[string]any{"object": map[string]any{"type": "toy", "fields": map[string]any{"name": "=$.name"}}}}},
			},
			"activities": []any{map[string]any{"name": "toy_door", "domain": "work", "description": doorDesc,
				"spec": map[string]any{"inputs": []any{map[string]any{"name": "name", "type": "string", "required": true}}, "emits": "toy.a", "who": []any{"human"}}}},
		}
		b, _ := json.Marshal(m)
		p := filepath.Join(dir, "toy.json")
		os.WriteFile(p, b, 0o644)
		return p
	}
	// v1 installs.
	sum, err := pack.Load(ctx, s, write(1, "b as shipped", "the door"), "test")
	if err != nil || sum.Types != 1 || sum.Rules != 2 || sum.Activities != 1 || sum.Upgraded != 0 || sum.Bundle != "pack-toy-v1" {
		t.Fatalf("v1: %+v %v", sum, err)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-01-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve v1: %v %v", err, errs)
	}
	// Approval appends the active version; the pack's own authorship rides
	// along, which is what lets the pack upgrade it later.
	shipped, _ := s.GetRule(ctx, "toy-b")
	if shipped.Status != core.StatusActive || shipped.Description != "b as shipped" || shipped.CreatedBy != "pack:toy" {
		t.Fatalf("toy-b as installed: %+v", shipped)
	}
	// v2 changes rule b and the door: both are the pack's own, so they are
	// drafted as the next versions in the upgrade bundle; a is unchanged.
	sum, err = pack.Load(ctx, s, write(2, "b improved", "the door, explained"), "test")
	if err != nil || sum.Rules != 0 || sum.Activities != 0 || sum.Upgraded != 2 || sum.Skipped != 2 || len(sum.Conflicts) != 0 || sum.Bundle != "pack-toy-v2" {
		t.Fatalf("v2: %+v %v", sum, err)
	}
	drafted, err := s.GetRule(ctx, "toy-b")
	if err != nil || drafted.Version != shipped.Version+1 || drafted.Status != core.StatusDraft || drafted.Description != "b improved" {
		t.Fatalf("toy-b after the upgrade load: %+v %v", drafted, err)
	}
	if active, _ := s.ActiveRules(ctx); len(active) != 2 || ruleVersion(active, "toy-b") != shipped.Version {
		t.Fatalf("the upgrade waits at the gate; the shipped version still runs: %+v", active)
	}
	bundle, _ := s.GetBundle(ctx, "pack-toy-v2")
	if len(bundle.Members) != 2 || bundle.Description[:len("Upgrade to")] != "Upgrade to" {
		t.Fatalf("the upgrade bundle: %+v", bundle)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, "pack-toy-v2", "krzysztof", "2026-02-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve v2: %v %v", err, errs)
	}
	if active, _ := s.ActiveRules(ctx); ruleVersion(active, "toy-b") <= shipped.Version || ruleDescription(active, "toy-b") != "b improved" {
		t.Fatalf("after approval: %+v", active)
	}
	if act, err := s.GetActiveActivity(ctx, "toy_door"); err != nil || act.Description != "the door, explained" {
		t.Fatalf("the door after approval: %+v %v", act, err)
	}
	// A reload of v2 changes nothing.
	sum, err = pack.Load(ctx, s, write(2, "b improved", "the door, explained"), "test")
	if err != nil || sum.Upgraded != 0 || sum.Rules != 0 || sum.Bundle != "" || sum.Skipped != 4 {
		t.Fatalf("reload: %+v %v", sum, err)
	}
	// A person evolves rule b locally; v3 of the pack changes b again — the
	// pack reports the conflict and leaves the person's version alone.
	local, err := s.InsertRuleVersion(ctx, core.Rule{ID: "toy-b", Status: core.StatusDraft, Priority: 90, EffectiveFrom: "2026-03-01",
		CreatedBy: "krzysztof", Description: "b, our way", Spec: drafted.Spec})
	if err != nil {
		t.Fatal(err)
	}
	sum, err = pack.Load(ctx, s, write(3, "b improved again", "the door, explained"), "test")
	if err != nil || sum.Upgraded != 0 || len(sum.Conflicts) != 1 || sum.Conflicts[0] != fmt.Sprintf("rule toy-b v%d (by krzysztof)", local.Version) || sum.Bundle != "" {
		t.Fatalf("v3 over a local evolution: %+v %v", sum, err)
	}
	if latest, _ := s.GetRule(ctx, "toy-b"); latest.Version != local.Version || latest.Description != "b, our way" {
		t.Fatalf("the person's version stands: %+v", latest)
	}
}

func ruleDescription(rules []core.Rule, id string) string {
	for _, r := range rules {
		if r.ID == id {
			return r.Description
		}
	}
	return ""
}

func ruleVersion(rules []core.Rule, id string) int {
	for _, r := range rules {
		if r.ID == id {
			return r.Version
		}
	}
	return 0
}
