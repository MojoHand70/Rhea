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

// A definition belongs to the packs whose files declare it (DIRECTION
// 2026-10-24, step 3): a pack that succeeds another may carry the other's
// rules forward unchanged — nothing is renamed, nothing drafted again — and
// from its approval the carried rules are its own: installed per company,
// they book only where it is installed, though their rows still name the
// pack that wrote them. What the old pack keeps and the new one does not
// carry stays the old pack's, for every company as before.
func TestADefinitionBelongsToThePacksThatDeclareIt(t *testing.T) {
	ctx := context.Background()
	s, x := withCore(t)
	rule := func(id, typ string) map[string]any {
		return map[string]any{"rule_id": id, "priority": 100, "effective_from": "2026-01-01", "description": id,
			"spec": map[string]any{"match": map[string]any{"event_type": "thing.recorded"},
				"effect": map[string]any{"object": map[string]any{"type": typ, "fields": map[string]any{"n": "=$.n"}}}}}
	}
	// today's pack, for every company: two rules
	approveFile(t, s, x, map[string]any{"pack": "old-test", "kind": "market", "install": "installation",
		"object_types": []map[string]any{
			{"name": "carried_thing", "version": 1, "fields": []map[string]any{{"name": "n", "type": "string"}}},
			{"name": "kept_thing", "version": 1, "fields": []map[string]any{{"name": "n", "type": "string"}}}},
		"rules": []map[string]any{rule("carried", "carried_thing"), rule("kept", "kept_thing")}}, "")
	// its successor, per company, carries one of them unchanged
	sum, err := loadFile(t, s, map[string]any{"pack": "new-test", "kind": "regime", "install": "company", "succeeds": []string{"old-test"},
		"object_types": []map[string]any{{"name": "carried_thing", "version": 1, "fields": []map[string]any{{"name": "n", "type": "string"}}}},
		"rules":        []map[string]any{rule("carried", "carried_thing")}})
	if err != nil {
		t.Fatal(err)
	}
	if sum.Rules+sum.Types+sum.Upgraded != 0 || len(sum.Conflicts) != 0 || sum.Bundle == "" {
		t.Fatalf("the successor drafted %+v — want nothing drafted again, a version to approve", sum)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-01-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve %s: %v %v", sum.Bundle, err, errs)
	}
	r, err := s.GetRule(ctx, "carried")
	if err != nil || r.CreatedBy != "pack:old-test" || r.Version != 2 {
		t.Fatalf("the carried rule was written again: %+v %v", r, err)
	}

	const a, b = "PL5250001111", "PL5260001246"
	if _, err := x.TriggerActivity(store.WithCompany(ctx, a), "install_pack", map[string]any{"pack": "new-test", "from": "2026-01-01", "installed_by": "krzysztof"}, "human", ""); err != nil {
		t.Fatal(err)
	}
	for _, c := range []string{a, b} {
		p, _ := json.Marshal(map[string]any{"n": c})
		if _, err := s.AppendEvent(store.WithOwner(ctx), core.Event{Kind: core.KindRaw, Type: "thing.recorded", OccurredAt: "2026-04-01",
			Payload: p, DedupKey: "thing/" + c, Company: c}); err != nil {
			t.Fatal(err)
		}
	}
	x.ProcessPending(store.WithOwner(ctx))
	var booked []string
	for _, typ := range []string{"carried_thing", "kept_thing"} {
		objs, err := s.ObjectsByType(store.WithOwner(ctx), typ)
		if err != nil {
			t.Fatal(err)
		}
		for _, o := range objs {
			booked = append(booked, typ+":"+o.Company)
		}
	}
	sort.Strings(booked)
	if got, want := fmt.Sprint(booked), fmt.Sprint([]string{"carried_thing:" + a, "kept_thing:" + a, "kept_thing:" + b}); got != want {
		t.Fatalf("booked %s, want %s — the carried rule where its new pack is installed, the kept one everywhere", got, want)
	}
	assertReplayIdentical(t, x)
}
