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

// A rule's own warrant (DECISIONS 2026-10-16): a pack rule names the source
// its reading rests on; the warrant lands with the draft, rides along when
// approval appends the active version, and is what the provenance walk shows
// beside the rule that explains a fact. A changed citation is a changed rule
// — the pack upgrade drafts it — and a statute claimed without its article
// never lands.
func TestRuleWarrant(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	dir := t.TempDir()
	write := func(version int, warrant map[string]any) string {
		rule := map[string]any{"rule_id": "toy-vat", "priority": 100, "effective_from": "2026-01-01",
			"description": "the tax point of a sale",
			"spec": map[string]any{"match": map[string]any{"event_type": "toy.sold"},
				"effect": map[string]any{"object": map[string]any{"type": "toy", "fields": map[string]any{"name": "=$.name"}}}}}
		if warrant != nil {
			rule["warrant"] = warrant
		}
		b, _ := json.Marshal(map[string]any{
			"pack": "toy", "version": version, "description": "a toy market",
			"object_types": []any{map[string]any{"name": "toy", "version": 1, "domain": "work", "label_field": "name",
				"fields": []any{map[string]any{"name": "name", "type": "string", "required": true}}}},
			"rules": []any{rule},
		})
		p := filepath.Join(dir, "toy.json")
		if err := os.WriteFile(p, b, 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	statute := func(citation string) map[string]any {
		return map[string]any{"basis": "statute", "citations": []any{citation}, "scope": "pl"}
	}

	sum, err := pack.Load(ctx, s, write(1, statute("ustawa o VAT art. 19a ust. 1")), "test")
	if err != nil || sum.Rules != 1 {
		t.Fatalf("v1: %+v %v", sum, err)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-01-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve v1: %v %v", err, errs)
	}
	active, err := s.GetRule(ctx, "toy-vat")
	if err != nil || active.Status != core.StatusActive || active.Warrant == nil ||
		active.Warrant.Basis != "statute" || active.Warrant.Citations[0] != "ustawa o VAT art. 19a ust. 1" {
		t.Fatalf("the approved version keeps its warrant: %+v %v", active, err)
	}

	// The walk from a fact ends at the law.
	tr, err := x.TriggerActivity(ctx, "submit_event", map[string]any{
		"event_type": "toy.sold", "occurred_at": "2026-01-02", "payload": map[string]any{"name": "a sale"}}, "test", "")
	if err != nil || len(tr.Errors) != 0 {
		t.Fatalf("submit: %v %v", err, tr.Errors)
	}
	toys, _ := s.ObjectsByType(ctx, "toy")
	if len(toys) != 1 || toys[0].RuleID != "toy-vat" || toys[0].RuleVersion != active.Version {
		t.Fatalf("toys: %+v", toys)
	}
	notes, err := s.RuleNotes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if n := notes[store.RuleKey{ID: "toy-vat", Version: active.Version}]; n.Warrant == nil || n.Warrant.Citations[0] != "ustawa o VAT art. 19a ust. 1" {
		t.Fatalf("the walk's note on the explaining rule: %+v", n)
	}

	// A corrected citation is the pack's next version of the rule, behind the gate.
	sum, err = pack.Load(ctx, s, write(2, statute("ustawa o VAT art. 19a ust. 1 i ust. 5")), "test")
	if err != nil || sum.Upgraded != 1 {
		t.Fatalf("v2: %+v %v", sum, err)
	}
	if drafted, _ := s.GetRule(ctx, "toy-vat"); drafted.Status != core.StatusDraft || drafted.Warrant.Citations[0] != "ustawa o VAT art. 19a ust. 1 i ust. 5" {
		t.Fatalf("the upgrade draft: %+v", drafted)
	}
	// The same content again is already present.
	if sum, err = pack.Load(ctx, s, write(2, statute("ustawa o VAT art. 19a ust. 1 i ust. 5")), "test"); err != nil || sum.Upgraded != 0 || sum.Skipped != 2 {
		t.Fatalf("reload: %+v %v", sum, err)
	}

	// A statute without its article, or a support figure the pack claims, never lands.
	for _, bad := range []map[string]any{
		{"basis": "statute"},
		{"basis": "practice", "support": map[string]any{"count": 9, "of": 10, "population": "Polish firms"}},
	} {
		if _, err := pack.Load(ctx, s, write(3, bad), "test"); err == nil {
			t.Fatalf("warrant %v was accepted", bad)
		}
	}
}
