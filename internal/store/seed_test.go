package store_test

import (
	"context"
	"testing"

	"rhea/internal/core"
	"rhea/internal/store/storetest"
)

// TestTheKernelsDoorsMoveWithIt (DECISIONS 2026-10-18): init seeds the kernel's
// own doors; a door the kernel changed gets its next version, approved by the
// kernel, once — and a door unchanged is never written again.
func TestTheKernelsDoorsMoveWithIt(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	if err := s.Init(ctx); err != nil {
		t.Fatal(err)
	}
	for _, a := range core.BuiltinActivities() {
		cur, err := s.GetActivity(ctx, a.Name)
		if err != nil || cur.Version != 1 {
			t.Fatalf("%s: a second init wrote version %d (%v)", a.Name, cur.Version, err)
		}
	}
	changed := core.BuiltinActivities()
	for i := range changed {
		if changed[i].Name == "note_untaught" {
			changed[i].Spec.Inputs = append(changed[i].Spec.Inputs, core.FieldDef{Name: "extra", Type: "string"})
		}
	}
	for range 2 {
		if err := s.SeedActivities(ctx, changed); err != nil {
			t.Fatal(err)
		}
	}
	cur, err := s.GetActivity(ctx, "note_untaught")
	if err != nil || cur.Version != 2 || cur.Status != core.StatusActive {
		t.Fatalf("the changed door: v%d %s %v", cur.Version, cur.Status, err)
	}
	if cur.Spec.Inputs[len(cur.Spec.Inputs)-1].Name != "extra" {
		t.Fatalf("the changed door's inputs: %+v", cur.Spec.Inputs)
	}
	acts, err := s.ActiveActivities(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range acts {
		if a.Name == "note_untaught" && a.Version != 2 {
			t.Fatalf("the offered door is v%d", a.Version)
		}
	}
	// The kernel's own definition again: one more version, back to it.
	if err := s.Init(ctx); err != nil {
		t.Fatal(err)
	}
	if cur, _ := s.GetActivity(ctx, "note_untaught"); cur.Version != 3 {
		t.Fatalf("back to the kernel's definition: v%d", cur.Version)
	}
}
