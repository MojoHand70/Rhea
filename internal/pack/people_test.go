package pack_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/store/storetest"
)

// People as data (DIRECTION, "The host is a concierge", KK 2026-10-10): a
// person is master data the way a contractor is — registered through a
// door, with a role, a deputy and a life — so that cases, notifications and
// disputes can refer to a human. The kernel does not know the type exists;
// the actor on an event resolves to a person by a lookup, not by code.
func TestPeoplePack(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}

	sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", "common", "people.json"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if sum.Types != 1 || sum.Rules != 3 || sum.Activities != 3 || sum.Bundle != "pack-common-people-v1" {
		t.Fatalf("load: %+v", sum)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-09-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve: %v %v", err, errs)
	}

	// Registration goes through the declared door, signed by the actor.
	register := func(name, handle, role string) (objectID string, eventID int64) {
		t.Helper()
		tr, err := x.TriggerActivity(ctx, "register_person", map[string]any{
			"name": name, "handle": handle, "role": role, "occurred_at": "2026-09-01",
		}, "shell:kk", "people/"+handle)
		if err != nil {
			t.Fatal(err)
		}
		if tr.Booked != 1 || len(tr.Errors) != 0 {
			t.Fatalf("register %s: booked %d, errors %v", handle, tr.Booked, tr.Errors)
		}
		ids, err := s.FindObjectIDsByField(ctx, "person", "handle", handle)
		if err != nil || len(ids) != 1 {
			t.Fatalf("person by handle %q: %v %v", handle, ids, err)
		}
		return ids[0], tr.EventID
	}
	kk, _ := register("Krzysztof Kowal", "kk", "owner")
	anna, annaEvent := register("Anna Nowak", "anna", "accountant")

	// The actor convention: a human signs events as "<door>:<handle>", and
	// the handle resolves to exactly one person — a lookup, never kernel code.
	ev, err := s.GetEvent(ctx, annaEvent)
	if err != nil {
		t.Fatal(err)
	}
	if ev.ActivityName != "register_person" || ev.Actor != "shell:kk" {
		t.Fatalf("door stamp: activity %q actor %q", ev.ActivityName, ev.Actor)
	}
	handle := ev.Actor[strings.IndexByte(ev.Actor, ':')+1:]
	if ids, _ := s.FindObjectIDsByField(ctx, "person", "handle", handle); len(ids) != 1 || ids[0] != kk {
		t.Fatalf("actor %q resolves to %v, want %s", ev.Actor, ids, kk)
	}
	o, err := s.GetObject(ctx, anna)
	if err != nil || o.State["status"] != "active" || o.State["role"] != "accountant" {
		t.Fatalf("anna: %v %v", o.State, err)
	}

	// A deputy is a person; the door vouches the ref is one.
	if _, err := x.TriggerActivity(ctx, "appoint_deputy", map[string]any{
		"person": anna, "deputy": "company-1",
	}, "shell:kk", "people/anna/deputy-wrong"); err == nil || !strings.Contains(err.Error(), "no object") {
		t.Fatalf("a non-person deputy must be refused at the door, got %v", err)
	}
	tr, err := x.TriggerActivity(ctx, "appoint_deputy", map[string]any{
		"person": anna, "deputy": kk, "occurred_at": "2026-09-02",
	}, "shell:kk", "people/anna/deputy")
	if err != nil || tr.Booked != 1 || len(tr.Errors) != 0 {
		t.Fatalf("appoint deputy: %v %+v", err, tr)
	}
	if o, _ = s.GetObject(ctx, anna); o.State["deputy"] != kk {
		t.Fatalf("anna.deputy = %v, want %s", o.State["deputy"], kk)
	}

	// A person has a life: active ⇄ away, either → left, and nothing back
	// from left — the kernel judges the move, the pack only declared it.
	move := func(status, dedup string) exec.Triggered {
		t.Helper()
		tr, err := x.TriggerActivity(ctx, "set_person_status", map[string]any{
			"person": anna, "status": status, "occurred_at": "2026-09-03",
		}, "shell:kk", dedup)
		if err != nil {
			t.Fatal(err)
		}
		return tr
	}
	for i, st := range []string{"away", "active", "left"} {
		if tr := move(st, "people/anna/"+st); tr.Booked != 1 || len(tr.Errors) != 0 {
			t.Fatalf("move %d to %s: %+v", i, st, tr)
		}
		if o, _ = s.GetObject(ctx, anna); o.State["status"] != st {
			t.Fatalf("after move to %s: status %v", st, o.State["status"])
		}
	}
	back := move("active", "people/anna/back")
	if back.Booked != 0 || len(back.Errors) == 0 {
		t.Fatalf("left → active must be refused, got %+v", back)
	}
	if o, _ = s.GetObject(ctx, anna); o.State["status"] != "left" {
		t.Fatalf("refused move changed state: %v", o.State["status"])
	}
	pending, err := s.PendingEvents(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var waits bool
	for _, p := range pending {
		waits = waits || p.ID == back.EventID
	}
	if !waits {
		t.Fatalf("the refused move must wait in the worklist, not vanish")
	}
}
