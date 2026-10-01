package agent_test

import (
	"context"
	"encoding/json"
	"os"
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

func fixtureAgent(answer string) *agent.Agent {
	return &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
		return answer, nil
	}}
}

func TestDraftRuleFixture(t *testing.T) {
	d, err := fixtureAgent(fixtureAnswer).DraftRule(
		context.Background(), "book PLN invoices", sampleEvent(), invoiceType())
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
	}
	for i, answer := range bad {
		if _, err := fixtureAgent(answer).DraftRule(
			context.Background(), "intent", sampleEvent(), invoiceType()); err == nil {
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
		"Invoices in PLN should become invoice documents; total is the sum of line amounts.",
		sampleEvent(), invoiceType())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("live draft: %+v", d)
}
