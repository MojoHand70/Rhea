package pack_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

// withCore is a database holding the core the kernel ships, approved.
func withCore(t *testing.T) (*store.Store, *exec.Executor) {
	t.Helper()
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	sums, err := pack.LoadCore(ctx, s, "init")
	if err != nil {
		t.Fatal(err)
	}
	for _, sum := range sums {
		if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-01-01"); err != nil || len(errs) != 0 {
			t.Fatalf("approve %s: %v %v", sum.Bundle, err, errs)
		}
	}
	return s, x
}

// packFile writes a small pack to a file: what it says it is, and one type of
// its own so that it has something to approve.
func packFile(t *testing.T, header map[string]any, typ string) string {
	t.Helper()
	m := map[string]any{"version": 1, "description": "a test pack", "stage": "initial",
		"disclaimer": "A test pack.", "history": []map[string]any{{"versions": "1", "changes": "first"}},
		"object_types": []map[string]any{}}
	if typ != "" {
		m["object_types"] = []map[string]any{{"name": typ, "version": 1, "fields": []map[string]any{{"name": "n", "type": "string"}}}}
	}
	for k, v := range header {
		m[k] = v
	}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), header["pack"].(string)+".json")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// A pack says what it is (DIRECTION 2026-10-24, step 2): its kind, its market,
// the packs it needs and whether it is installed for the whole installation or
// per company — and the approved version keeps it, so the installation knows
// its catalog. A pack whose needs are not available, or that says it is
// something no pack can be, is refused at loading.
func TestAPackVersionKeepsWhatThePackIs(t *testing.T) {
	ctx := context.Background()
	s, x := withCore(t)

	kpir := packFile(t, map[string]any{"pack": "kpir-test", "kind": "regime", "market": "PL",
		"needs": []string{"base-test"}, "install": "company"}, "kpir_entry_test")
	if _, err := pack.Load(ctx, s, kpir, "test"); err == nil || !strings.Contains(err.Error(), "base-test") {
		t.Fatalf("a pack whose needs are not available loaded: %v", err)
	}

	base, err := pack.Load(ctx, s, packFile(t, map[string]any{"pack": "base-test", "kind": "market", "market": "PL",
		"install": "company"}, "base_thing_test"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, base.Bundle, "krzysztof", "2026-01-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve %s: %v %v", base.Bundle, err, errs)
	}
	b, err := s.GetBundle(ctx, base.Bundle)
	if err != nil {
		t.Fatal(err)
	}
	want := &core.PackHeader{Pack: "base-test", Version: 1, Kind: "market", Market: "PL", Install: "company", Description: "a test pack"}
	if b.Status != core.StatusActive || !reflect.DeepEqual(b.Pack, want) {
		t.Fatalf("approved bundle = %s %+v, want active %+v", b.Status, b.Pack, want)
	}

	sum, err := pack.Load(ctx, s, kpir, "test")
	if err != nil {
		t.Fatalf("a pack whose needs are approved did not load: %v", err)
	}
	if b, err := s.GetBundle(ctx, sum.Bundle); err != nil || b.Pack == nil || !reflect.DeepEqual(b.Pack.Needs, []string{"base-test"}) {
		t.Fatalf("the draft does not keep what it needs: %+v %v", b.Pack, err)
	}

	for _, bad := range []map[string]any{
		{"pack": "odd-kind", "kind": "religion", "install": "company"},
		{"pack": "odd-install", "kind": "module", "install": "everywhere"},
	} {
		if _, err := pack.Load(ctx, s, packFile(t, bad, strings.ReplaceAll(bad["pack"].(string), "-", "_")), "test"); err == nil {
			t.Fatalf("pack %v loaded", bad)
		}
	}
}
