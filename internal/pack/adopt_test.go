package pack_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

// A pack played again at the same version drafts nothing (idempotent); a
// pack that changed without a version bump is refused before it drafts a
// thing, and told to bump.
func TestPackChangedWithoutBumpIsRefused(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	src := filepath.Join("..", "..", "packs", "common", "people.json")
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "people.json")
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	first, err := pack.Load(ctx, s, path, "test")
	if err != nil || first.Bundle == "" {
		t.Fatalf("first load: %+v %v", first, err)
	}
	again, err := pack.Load(ctx, s, path, "test")
	if err != nil || again.Bundle != "" || again.Types != 0 {
		t.Fatalf("the same pack again: %+v %v", again, err)
	}
	var m pack.Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	m.ObjectTypes = append(m.ObjectTypes, core.ObjectType{Name: "zz_extra", Version: 1, Domain: "people", LabelField: "name",
		Fields: []core.FieldDef{{Name: "name", Type: "string", Required: true}}})
	changed, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, changed, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := pack.Load(ctx, s, path, "test"); err == nil || !strings.Contains(err.Error(), "bump") {
		t.Fatalf("a changed pack at the same version: %v", err)
	}
	if _, _, err := s.GetObjectTypeVersion(ctx, "zz_extra", 1); err == nil {
		t.Fatal("the refused load drafted the new type")
	}
}

// A type that changed at the same type version is refused whatever the pack
// version says: the stored type would never carry the change, and a rule
// using the new field would be drafted against a type never stored.
func TestPackTypeChangedAtSameVersionIsRefused(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	src := filepath.Join("..", "..", "packs", "common", "people.json")
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "people.json")
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := pack.Load(ctx, s, path, "test"); err != nil {
		t.Fatal(err)
	}
	var m pack.Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	m.Version++
	m.ObjectTypes[0].Fields = append(m.ObjectTypes[0].Fields, core.FieldDef{Name: "zz_extra", Type: "string"})
	changed, _ := json.Marshal(m)
	if err := os.WriteFile(path, changed, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := pack.Load(ctx, s, path, "test"); err == nil || !strings.Contains(err.Error(), "bump the type's version") {
		t.Fatalf("a changed type at the same version: %v", err)
	}
}
