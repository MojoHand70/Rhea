package network_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/network"
	"rhea/internal/store"
)

// A simulated customer never counts for a real client. Its publication is
// stored marked; Learn leaves it out, and only a simulated run asks for it.
func TestSyntheticInstallationsNeverCount(t *testing.T) {
	ctx := context.Background()
	if _, err := store.Open(ctx, store.DSN()); err != nil {
		if os.Getenv("RHEA_CI") != "" {
			t.Fatalf("postgres unavailable and RHEA_CI is set: %v", err)
		}
		t.Skipf("postgres unavailable: %v", err)
	}
	dsn, drop, err := store.ThrowawayDSN(ctx, "rhea_network_test")
	if err != nil {
		t.Fatal(err)
	}
	defer drop()
	nw, err := network.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer nw.Close()

	rule := core.Rule{ID: "book", Status: core.StatusActive, Spec: core.RuleSpec{
		Match:  core.Match{EventType: "invoice.received"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "invoice", Fields: map[string]string{"total": "=$.total"}}}}}
	for _, inst := range []string{"acme", "sim:nordwind", "sim:helios"} {
		if _, err := nw.Publish(ctx, inst, []core.Rule{rule}); err != nil {
			t.Fatal(err)
		}
	}
	real, err := nw.Learn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if real.Installations != 1 {
		t.Fatalf("real knowledge counts %d installations, want the one real client", real.Installations)
	}
	if _, ok := real.SupportFor(rule.Spec); ok {
		t.Fatal("one real client's shape surfaced on the strength of simulated ones")
	}
	all, err := nw.LearnIncludingSynthetic(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if sup, ok := all.SupportFor(rule.Spec); !ok || sup.Count != 3 || all.Installations != 3 {
		t.Fatalf("synthetic knowledge = %+v, %v", sup, ok)
	}
}

// The network carries general rules only, with no company's values (KK,
// 2026-10-27: "Shared knowledge are general rules that relate to general law
// or accepted practice. No particular company data in it!"). The shape blanked
// a condition's value and nothing else, and every rule went: a company's own
// rule, a NIP written in a formula, a tenant's name and street written as a
// field's value all left the installation (review 2026-10-27, high). Now a
// company's rule is never published, and in Rhea's every value a rule writes
// — text, a long number, a plain field value — is a "?".
func TestOnlyGeneralRulesLeave(t *testing.T) {
	ctx := context.Background()
	if _, err := store.Open(ctx, store.DSN()); err != nil {
		if os.Getenv("RHEA_CI") != "" {
			t.Fatalf("postgres unavailable and RHEA_CI is set: %v", err)
		}
		t.Skipf("postgres unavailable: %v", err)
	}
	dsn, drop, err := store.ThrowawayDSN(ctx, "rhea_network_test")
	if err != nil {
		t.Fatal(err)
	}
	defer drop()
	nw, err := network.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer nw.Close()

	companys := core.Rule{ID: "own-shape@PL5250001111", Company: "PL5250001111", Status: core.StatusActive, Spec: core.RuleSpec{
		Match:  core.Match{EventType: "memo.left"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "memo", Fields: map[string]string{"text": "=$.text"}}}}}
	rheas := core.Rule{ID: "rent", Status: core.StatusActive, Spec: core.RuleSpec{
		Match: core.Match{EventType: "bank.statement.line", Where: []core.Condition{{Formula: `=$.counterparty_nip = "9512345678"`}}},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "rent", Fields: map[string]string{
			"text": "Czynsz Jan Kowalski ul. Polna 5", "tenant": `=if($.nip = 9512345678, "Kowalski", "other")`, "amount": "=$.amount"}}}}}
	n, err := nw.Publish(ctx, "acme", []core.Rule{companys, rheas})
	if err != nil || n != 1 {
		t.Fatalf("published %d shapes, %v — Rhea's rule alone", n, err)
	}
	k, err := nw.Learn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(k)
	for _, secret := range []string{"memo.left", "9512345678", "Kowalski", "Polna"} {
		if strings.Contains(string(b), secret) {
			t.Fatalf("%q left the installation: %s", secret, b)
		}
	}
	if !strings.Contains(string(b), "bank.statement.line") || !strings.Contains(string(b), "=$.amount") {
		t.Fatalf("the rule's shape is gone with its values: %s", b)
	}
}
