package network_test

import (
	"context"
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
