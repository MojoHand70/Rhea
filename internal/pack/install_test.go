package pack_test

import (
	"context"
	"strings"
	"testing"

	"rhea/internal/pack"
	"rhea/internal/store"
)

// An install is a fact in the company's log (DIRECTION 2026-10-24, step 2):
// "pack P installed from D", "switched off from E" — through the core's two
// doors, checked by the kernel against the catalog: a pack is installed in a
// company only when it is approved, installed per company, and what it needs
// is installed there; never twice at once; never switched off while a pack
// that needs it stays. Another company sees none of it.
func TestAnInstallIsAFactOfTheCompany(t *testing.T) {
	ctx := context.Background()
	s, x := withCore(t)
	for _, h := range []map[string]any{
		{"pack": "base-test", "kind": "market", "market": "PL", "install": "company"},
		{"pack": "kpir-test", "kind": "regime", "market": "PL", "needs": []string{"base-test"}, "install": "company"},
		{"pack": "whole-test", "kind": "module", "install": "installation"},
	} {
		sum, err := pack.Load(ctx, s, packFile(t, h, strings.ReplaceAll(h["pack"].(string), "-", "_")+"_thing"), "test")
		if err != nil {
			t.Fatal(err)
		}
		if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-01-01"); err != nil || len(errs) != 0 {
			t.Fatalf("approve %s: %v %v", sum.Bundle, err, errs)
		}
	}
	alfa, beta := store.WithCompany(ctx, "PL5250001111"), store.WithCompany(ctx, "PL5260001246")
	install := func(c context.Context, p, from string) error {
		_, err := x.TriggerActivity(c, "install_pack", map[string]any{"pack": p, "from": from, "installed_by": "krzysztof"}, "human", "")
		return err
	}
	installed := func(c context.Context) map[string]string {
		t.Helper()
		objs, err := s.ObjectsByType(c, "installed_pack")
		if err != nil {
			t.Fatal(err)
		}
		out := map[string]string{}
		for _, o := range objs {
			out[o.State["pack"].(string)] = o.State["status"].(string) + " " + o.ID
		}
		return out
	}

	for _, c := range []struct{ pack, why string }{
		{"kpir-test", "base-test"},      // what it needs is not installed here
		{"whole-test", "every company"}, // installed for the whole installation already
		{"core", "every company"},
		{"no-such-pack", "no-such-pack"},
	} {
		if err := install(alfa, c.pack, "2026-01-01"); err == nil || !strings.Contains(err.Error(), c.why) {
			t.Fatalf("install %s: %v, want a refusal naming %q", c.pack, err, c.why)
		}
	}
	if err := install(alfa, "base-test", "2026-01-01"); err != nil {
		t.Fatal(err)
	}
	if err := install(alfa, "base-test", "2026-01-01"); err == nil {
		t.Fatal("the same install twice made two facts")
	}
	if err := install(alfa, "base-test", "2026-02-01"); err == nil || !strings.Contains(err.Error(), "installed") {
		t.Fatalf("a second install while the first stands: %v", err)
	}
	if err := install(alfa, "kpir-test", "2026-03-01"); err != nil {
		t.Fatal(err)
	}
	got := installed(alfa)
	if !strings.HasPrefix(got["base-test"], "installed ") || !strings.HasPrefix(got["kpir-test"], "installed ") || len(got) != 2 {
		t.Fatalf("alfa's installs: %v", got)
	}
	if b := installed(beta); len(b) != 0 {
		t.Fatalf("beta sees alfa's installs: %v", b)
	}

	off := func(p, to string) error {
		_, err := x.TriggerActivity(alfa, "switch_off_pack", map[string]any{"installed": strings.Fields(got[p])[1], "to": to, "switched_off_by": "krzysztof"}, "human", "")
		return err
	}
	if err := off("base-test", "2026-06-30"); err == nil || !strings.Contains(err.Error(), "kpir-test") {
		t.Fatalf("switched off what an installed pack needs: %v", err)
	}
	if err := off("kpir-test", "2026-06-30"); err != nil {
		t.Fatal(err)
	}
	if s := installed(alfa)["kpir-test"]; !strings.HasPrefix(s, "switched_off ") {
		t.Fatalf("kpir-test after switching off: %q", s)
	}
	if err := off("base-test", "2026-06-30"); err != nil {
		t.Fatalf("base-test, needed by nothing installed any more: %v", err)
	}
}
