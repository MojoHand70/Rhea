package store_test

import (
	"context"
	"testing"

	"rhea/internal/core"
	"rhea/internal/store"
)

// Invariant 9 for definitions: a company's session writes its own rules and
// bundles, never Rhea's, which every company runs and reads — only the
// installation's own scope writes those (the house's approval of Rhea's
// definitions goes through it). The rule and bundle policies once checked a
// new row with the read test alone, which lets every scope see Rhea's rows:
// a session acting for one company could insert a draft for everyone
// (readiness review 2026-10-23, §0.2).
func TestACompanyWritesNoRuleForEveryone(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	alfa := store.WithCompany(ctx, "PL5250001111")
	installation := store.WithInstallation(ctx)
	spec := core.RuleSpec{
		Match: core.Match{EventType: "invoice.received"},
		Effect: core.Effect{Object: core.ObjectTemplate{
			Type: "invoice", Fields: map[string]string{"customer": "=$.customer"}}},
	}
	rule := func(id, company string) core.Rule {
		return core.Rule{ID: id, Status: core.StatusDraft, Priority: 100, EffectiveFrom: "2026-01-01",
			CreatedBy: "agent", Description: "draft", Spec: spec, Company: company}
	}

	if _, err := s.InsertRuleVersion(alfa, rule("for-everyone", "")); err == nil {
		t.Fatal("a session acting for alfa wrote a rule of Rhea's")
	}
	if _, err := s.Pool.Exec(alfa, `INSERT INTO rule (rule_id, version, status, priority, effective_from, created_by, description, spec)
		VALUES ('for-everyone-sql', 1, 'draft', 100, '2026-01-01', 'agent', 'draft', '{}')`); err == nil {
		t.Fatal("a session acting for alfa wrote a rule of Rhea's through raw SQL")
	}
	own, err := s.InsertRuleVersion(alfa, rule("alfas@PL5250001111", "PL5250001111"))
	if err != nil {
		t.Fatalf("alfa could not draft its own rule: %v", err)
	}
	if _, err := s.InsertRuleVersion(installation, rule("rheas", "")); err != nil {
		t.Fatalf("the installation could not draft a rule of Rhea's: %v", err)
	}

	member := []core.Member{{Kind: core.KindRule, Name: own.ID, Version: own.Version}}
	if _, err := s.InsertBundle(alfa, core.Bundle{ID: "for-everyone", Description: "d", Members: member, CreatedBy: "agent"}); err == nil {
		t.Fatal("a session acting for alfa wrote a bundle of Rhea's")
	}
	if _, err := s.Pool.Exec(alfa, `INSERT INTO bundle (bundle_id, version, status, description, members, created_by)
		VALUES ('for-everyone-sql', 1, 'draft', 'd', '[]', 'agent')`); err == nil {
		t.Fatal("a session acting for alfa wrote a bundle of Rhea's through raw SQL")
	}
	if _, err := s.InsertBundle(alfa, core.Bundle{ID: "alfas@PL5250001111", Description: "d", Members: member,
		CreatedBy: "agent", Company: "PL5250001111"}); err != nil {
		t.Fatalf("alfa could not bundle its own draft: %v", err)
	}

	// what alfa tried for everyone is nowhere: the owner sees the whole table
	var n int
	if err := s.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM rule WHERE rule_id LIKE 'for-everyone%'`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("rules for everyone written by alfa: %d, %v", n, err)
	}
	if err := s.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM bundle WHERE bundle_id LIKE 'for-everyone%'`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("bundles for everyone written by alfa: %d, %v", n, err)
	}
}
