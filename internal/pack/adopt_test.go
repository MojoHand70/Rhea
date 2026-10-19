package pack_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

// A load that fails partway leaves drafts on the record with no bundle; a
// rejection frees a bundle's drafts the same way. The next load of the pack
// adopts such orphans into its bundle instead of skipping them, so the
// pack still installs whole — and a draft a pending bundle holds is left
// to that bundle.
func TestALoadAdoptsOrphanDrafts(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	path := filepath.Join("..", "..", "packs", "common", "people.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var m pack.Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	// the orphan: the pack's first type, drafted by a load that then failed
	orphan := m.ObjectTypes[0]
	if err := store.InsertObjectTypeRow(ctx, s.Pool, orphan, core.StatusDraft); err != nil {
		t.Fatal(err)
	}
	sum, err := pack.Load(ctx, s, path, "test")
	if err != nil {
		t.Fatal(err)
	}
	if sum.Adopted != 1 || sum.Types != len(m.ObjectTypes)-1 {
		t.Fatalf("adopted %d, types drafted %d: %+v", sum.Adopted, sum.Types, sum)
	}
	bundle, err := s.GetBundle(ctx, sum.Bundle)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, mem := range bundle.Members {
		if mem.Kind == core.KindObjectType && mem.Name == orphan.Name && mem.Version == orphan.Version {
			found = true
		}
	}
	if !found {
		t.Fatalf("the orphan is not in the bundle: %+v", bundle.Members)
	}
	// a second load: everything is a draft a pending bundle holds — nothing adopted, no new bundle
	again, err := pack.Load(ctx, s, path, "test")
	if err != nil {
		t.Fatal(err)
	}
	if again.Adopted != 0 || again.Bundle != "" {
		t.Fatalf("a pending bundle's drafts are its own: %+v", again)
	}
	// approving installs the orphan with the rest
	x := &exec.Executor{Store: s}
	if _, _, _, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-09-01"); err != nil {
		t.Fatal(err)
	}
	if _, status, err := s.GetObjectTypeVersion(ctx, orphan.Name, orphan.Version); err != nil || status != core.StatusActive {
		t.Fatalf("the orphan after approval: %s %v", status, err)
	}
}
