package shell_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/network"
	"rhea/internal/shell"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

// Rhea learns (DIRECTION, the network; KK 2026-10-08: "if it learns, it uses
// what it learned"). Four installations explain invoices; they publish the
// shapes of their rules — never their data; the network counts. A fifth
// installation's agent sees the learned answer with its real support, a
// proposal resting on it is verified and counted by Rhea, a claim she has
// not learned is refused, and the client's own rule wins and is counted too.
func TestRheaLearnsAcrossInstallations(t *testing.T) {
	ctx := context.Background()
	dsn, drop, err := store.ThrowawayDSN(ctx, "rhea_network_test")
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	defer drop()
	nw, err := network.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer nw.Close()

	invoice := core.ObjectType{Name: "invoice", Version: 1, Domain: "finance", IsDocument: true,
		Fields: []core.FieldDef{{Name: "customer", Type: "string", Required: true},
			{Name: "currency", Type: "string", Required: true}, {Name: "total", Type: "money", Required: true}}}
	posting := func(credit string) core.RuleSpec {
		return core.RuleSpec{
			Match: core.Match{EventType: core.EventObjectMaterialized,
				Where: []core.Condition{{Path: "$.object_type", Op: "eq", Value: "invoice"}}},
			Effect: core.Effect{Postings: &core.PostingsTemplate{Currency: "=$.state.currency",
				Lines: []core.PostingLine{{Account: "201", Debit: "=$.state.total"}, {Account: credit, Credit: "=$.state.total"}}}}}
	}
	install := func(name, credit, customer string) *store.Store {
		t.Helper()
		s := storetest.New(t)
		if err := s.InsertObjectType(ctx, invoice); err != nil {
			t.Fatal(err)
		}
		book := core.RuleSpec{Match: core.Match{EventType: "invoice.received",
			Where: []core.Condition{{Path: "$.customer", Op: "eq", Value: customer}}}, // business data: must not travel
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "invoice", Fields: map[string]string{
				"customer": "=$.customer", "currency": "=$.currency", "total": "=sum($.lines[*].amount)"}}}}
		for i, spec := range []core.RuleSpec{book, posting(credit)} {
			// Different ids, descriptions and priorities everywhere: the same
			// explanation is recognized by its shape, not its name.
			if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: name + "-rule-" + string(rune('a'+i)),
				Status: core.StatusActive, Priority: 100 + 10*i + len(name), EffectiveFrom: "2026-01-01",
				CreatedBy: "test", Description: name + " books its invoices", Spec: spec}); err != nil {
				t.Fatal(err)
			}
		}
		return s
	}
	publish := func(name string, s *store.Store) {
		t.Helper()
		rules, err := s.ActiveRules(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := nw.Publish(ctx, name, rules); err != nil {
			t.Fatal(err)
		}
	}
	for _, in := range []struct{ name, credit, customer string }{
		{"alfa", "702", "ACME Sp. z o.o."}, {"beta", "702", "Zeta Works"},
		{"gamma", "702", "Nabywca S.A."}, {"delta", "730", "Gamma SARL"},
	} {
		publish(in.name, install(in.name, in.credit, in.customer))
	}

	// What leaves an installation is explanations, never facts.
	var raw string
	if err := nw.Pool.QueryRow(ctx, `SELECT string_agg(shapes::text, ' ') FROM publication`).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"ACME", "Zeta", "Nabywca", "Gamma SARL", "books its invoices"} {
		if strings.Contains(raw, secret) {
			t.Fatalf("%q reached the network: business data must never leave an installation", secret)
		}
	}

	k, err := nw.Learn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	q := "object.materialized:invoice → postings:main"
	answers := k.Questions[q]
	if k.Installations != 4 || len(answers) != 2 || answers[0].Count != 3 || answers[0].Of != 4 ||
		!answers[0].Surface || answers[1].Count != 1 || answers[1].Surface {
		t.Fatalf("learned %s: %+v", q, answers)
	}
	if sup, ok := k.SupportFor(posting("702")); !ok || sup.Count != 3 || sup.Of != 4 {
		t.Fatalf("support for 201/702 = %+v %v", sup, ok)
	}
	if _, ok := k.SupportFor(posting("730")); ok {
		t.Fatal("one installation's answer surfaced: below the floor it is that client's own business")
	}

	// A fifth installation: the agent sees what Rhea learned, with real counts,
	// following the cascade from the raw event to the posting question.
	fifth := storetest.New(t)
	if err := fifth.InsertObjectType(ctx, invoice); err != nil {
		t.Fatal(err)
	}
	sample := core.Event{Type: "invoice.received", OccurredAt: "2026-10-08", Payload: json.RawMessage(`{"customer":"Kowalski","currency":"PLN"}`)}
	ask, err := shell.DraftAsk(ctx, fifth, &k, "book our invoices", sample, "")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range ask.Priors {
		if p.Question == q && p.Count == 3 && p.Of == 4 && strings.Contains(string(p.Rule), `"702"`) {
			found = true
		}
		if strings.Contains(string(p.Rule), `"730"`) {
			t.Fatalf("a below-floor answer reached the agent: %+v", p)
		}
	}
	if !found {
		t.Fatalf("priors = %+v", ask.Priors)
	}

	// The agent rests a proposal on the network: Rhea verifies and counts.
	learned := agent.BundleDraft{BundleID: "post-invoices", Description: "post invoices as most do",
		Warrant: &core.Warrant{Basis: "network"},
		Rules:   []agent.Draft{{RuleID: "post-invoice", Description: "201/702", Priority: 200, Spec: posting("702")}}}
	b, err := shell.StoreBundleDraft(ctx, fifth, &k, learned, "agent:test", "2026-01-01", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if b.Warrant.Support == nil || b.Warrant.Support.Count != 3 || b.Warrant.Support.Of != 4 {
		t.Fatalf("stored warrant = %+v", b.Warrant)
	}
	// A claim Rhea has not learned never lands.
	claim := learned
	claim.BundleID = "post-invoices-730"
	claim.Rules = []agent.Draft{{RuleID: "post-invoice-730", Description: "201/730", Priority: 200, Spec: posting("730")}}
	if _, err := shell.StoreBundleDraft(ctx, fifth, &k, claim, "agent:test", "2026-01-01", "", ""); err == nil ||
		!strings.Contains(err.Error(), "none of its rules is an answer Rhea has learned") {
		t.Fatalf("unlearned network claim: %v", err)
	}

	// The client's own rule wins — and is counted: 3 of 5 now.
	own := agent.BundleDraft{BundleID: "our-way", Description: "we book services to 700",
		Warrant: &core.Warrant{Basis: "client", Citations: []string{"our accountant's policy"}},
		Rules:   []agent.Draft{{RuleID: "post-invoice-700", Description: "201/700", Priority: 200, Spec: posting("700")}}}
	ob, err := shell.StoreBundleDraft(ctx, fifth, &k, own, "agent:test", "2026-01-01", "", "")
	if err != nil {
		t.Fatal(err)
	}
	x := &exec.Executor{Store: fifth}
	if _, _, _, err := x.ApproveBundle(ctx, ob.ID, "krzysztof", "2026-10-08"); err != nil {
		t.Fatal(err)
	}
	publish("epsilon", fifth)
	if k, err = nw.Learn(ctx); err != nil {
		t.Fatal(err)
	}
	if sup, ok := k.SupportFor(posting("702")); !ok || sup.Count != 3 || sup.Of != 5 {
		t.Fatalf("after the override: %+v %v — the deviation is counted, honestly", sup, ok)
	}
}
