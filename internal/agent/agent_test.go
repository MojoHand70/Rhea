package agent_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"rhea/internal/agent"
	"rhea/internal/core"
)

func invoiceType() core.ObjectType {
	return core.ObjectType{
		Name: "invoice", Version: 1, Domain: "finance", IsDocument: true,
		Fields: []core.FieldDef{
			{Name: "customer", Type: "string", Required: true},
			{Name: "issue_date", Type: "date", Required: true},
			{Name: "currency", Type: "string", Required: true},
			{Name: "total", Type: "money", Required: true},
		},
	}
}

func sampleEvent() core.Event {
	return core.Event{
		Type: "invoice.received", OccurredAt: "2026-09-15",
		Payload: json.RawMessage(`{"customer":"ACME","currency":"PLN","issue_date":"2026-09-15","lines":[{"amount":"200.00"},{"amount":"150.50"}]}`),
	}
}

const fixtureAnswer = "```json\n" + `{
  "rule_id": "book-pln-invoice",
  "description": "PLN invoices become invoice documents with a summed total.",
  "priority": 100,
  "spec": {
    "match": {"event_type": "invoice.received",
              "where": [{"path": "$.currency", "op": "eq", "value": "PLN"}]},
    "effect": {"object": {"type": "invoice", "fields": {
      "customer": "=$.customer", "issue_date": "=$.issue_date",
      "currency": "=$.currency", "total": "=sum($.lines[*].amount)"}}}
  }
}` + "\n```"

func caseType() core.ObjectType {
	return core.ObjectType{
		Name: "case", Version: 1, Domain: "work",
		Fields: []core.FieldDef{
			{Name: "title", Type: "string", Required: true},
			{Name: "status", Type: "enum", Required: true, Values: []string{"open", "resolved"}},
			{Name: "resolution", Type: "string"},
		},
		Lifecycle: &core.LifecycleDef{Field: "status",
			Transitions: map[string][]string{"open": {"resolved"}}},
	}
}

func ask(intent string) agent.Ask {
	return agent.Ask{Intent: intent, Sample: sampleEvent(),
		Types: []core.ObjectType{invoiceType(), caseType()}}
}

func fixtureAgent(answer string) *agent.Agent {
	return &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
		return answer, nil
	}}
}

func TestDraftRuleFixture(t *testing.T) {
	d, err := fixtureAgent(fixtureAnswer).DraftRule(
		context.Background(), ask("book PLN invoices"))
	if err != nil {
		t.Fatal(err)
	}
	if d.RuleID != "book-pln-invoice" || d.Spec.Match.EventType != "invoice.received" {
		t.Fatalf("draft = %+v", d)
	}
}

func TestDraftRuleRejectsBadOutput(t *testing.T) {
	bad := []string{
		"I think you should use a rule here.",         // no JSON
		`{"rule_id":"x","description":"d","spec":{}}`, // empty spec
		`{"rule_id":"","description":"d"}`,            // missing id
		`{"rule_id":"x","description":"d","spec":{"match":{"event_type":"e"},"effect":{"object":{"type":"invoice","fields":{"bogus":"=$.x"}}}}}`, // unknown field
		`{"rule_id":"x","description":"d","spec":{"match":{"event_type":"e"},"effect":{"object":{"type":"receipt","fields":{"a":"=$.x"}}}}}`,     // undeclared type
		`{"rule_id":"x","description":"d","spec":{"match":{"event_type":"e"},"effect":{"amend":{"type":"invoice","target":"=$.id","set":{"total":"=$.x"}}}}}`, // no lifecycle, no consent
		`{"rule_id":"x","description":"d","spec":{"match":{"event_type":"e"},"effect":{"amend":{"type":"case","target":"c-1","set":{"status":"resolved"}}}}}`, // literal target
	}
	for i, answer := range bad {
		if _, err := fixtureAgent(answer).DraftRule(
			context.Background(), ask("intent")); err == nil {
			t.Errorf("case %d: bad model output accepted", i)
		}
	}
}

// TestDraftRuleLive exercises the real API. Gated: runs only with RHEA_LIVE=1
// and an ANTHROPIC_API_KEY present.
func TestDraftRuleLive(t *testing.T) {
	if os.Getenv("RHEA_LIVE") != "1" || os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("live test disabled (set RHEA_LIVE=1 and ANTHROPIC_API_KEY)")
	}
	d, err := agent.New().DraftRule(context.Background(),
		ask("Invoices in PLN should become invoice documents; total is the sum of line amounts."))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("live draft: %+v", d)
}

// The whole closed class is draftable: cascade off a derived event, an
// amendment along a declared lifecycle, a converted posting into a book.
func TestDraftRuleWholeGrammar(t *testing.T) {
	good := []string{
		`{"rule_id":"raise-invoice-case","description":"d","spec":{"match":{"event_type":"object.materialized","where":[{"path":"$.object_type","op":"eq","value":"invoice"}]},"effect":{"object":{"type":"case","fields":{"title":"=$.state.customer","status":"open"}}}}}`,
		`{"rule_id":"resolve-case","description":"d","spec":{"match":{"event_type":"case.resolved"},"effect":{"amend":{"type":"case","target":"=$.case","set":{"status":"resolved","resolution":"=$.resolution"}}}}}`,
		`{"rule_id":"post-invoice","description":"d","spec":{"match":{"event_type":"invoice.received"},"effect":{"postings":{"book":"group","currency":"=$.currency","convert":{"to":"EUR","date":"=$.issue_date","rounding":"half_up"},"lines":[{"account":"201","debit":"=sum($.lines[*].amount)"},{"account":"702","credit":"=sum($.lines[*].amount)"}]}}}}`,
	}
	for i, answer := range good {
		if _, err := fixtureAgent(answer).DraftRule(context.Background(), ask("intent")); err != nil {
			t.Errorf("case %d: %v", i, err)
		}
	}
}

// The agent sees the language as it stands: catalog, running rules, master
// data, the evidence and the human's pointer.
func TestDraftRuleSeesTheLanguage(t *testing.T) {
	var seen string
	a := &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
		seen = user
		return fixtureAnswer, nil
	}}
	in := ask("book PLN invoices")
	in.Hint = "invoice"
	in.Rules = []core.Rule{{ID: "register-company", Priority: 50, Description: "companies",
		Spec: core.RuleSpec{Match: core.Match{EventType: "company.created"}}}}
	in.Reference = map[string][]map[string]any{"company": {{"name": "ACME Sp. z o.o."}}}
	if _, err := a.DraftRule(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"name":"case"`, `"lifecycle"`, "register-company p50",
		"ACME Sp. z o.o.", `points at object type "invoice"`, "invoice.received"} {
		if !strings.Contains(seen, want) {
			t.Errorf("user message lacks %q", want)
		}
	}
}
