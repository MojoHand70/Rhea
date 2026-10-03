// Package agent is the AI layer: it drafts rules as strict JSON from a plain-
// language intent and a sample event. It has no store access and no side
// effects — the kernel validates its output and stores it as a draft
// (invariant 3). Model output is data, never executed.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"rhea/internal/core"
)

// Draft is what the agent proposes. The kernel decides what to do with it.
type Draft struct {
	RuleID      string        `json:"rule_id"`
	Description string        `json:"description"`
	Priority    int           `json:"priority"`
	Spec        core.RuleSpec `json:"spec"`
}

// Agent drafts rules. Complete is the single LLM call, injectable for tests.
type Agent struct {
	Complete func(ctx context.Context, system, user string) (string, error)
}

func Model() string {
	if m := os.Getenv("RHEA_MODEL"); m != "" {
		return m
	}
	return "claude-sonnet-4-6"
}

// New returns an Agent backed by the Anthropic API (key from ANTHROPIC_API_KEY).
// A user-scoped key (sk-ant-usr…) is not bound to a workspace, so the API then
// requires the workspace id with every request: ANTHROPIC_WORKSPACE_ID.
func New() *Agent {
	var opts []option.RequestOption
	if ws := os.Getenv("ANTHROPIC_WORKSPACE_ID"); ws != "" {
		opts = append(opts, option.WithHeader("anthropic-workspace-id", ws))
	}
	client := anthropic.NewClient(opts...)
	return &Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
		msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.Model(Model()),
			MaxTokens: 2000,
			System:    []anthropic.TextBlockParam{{Text: system}},
			Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(user))},
		})
		if err != nil {
			return "", err
		}
		var sb strings.Builder
		for _, block := range msg.Content {
			sb.WriteString(block.Text)
		}
		return sb.String(), nil
	}}
}

const systemPrompt = `You draft rules for Rhea, a rule-driven ERP kernel. You answer with ONE JSON object and nothing else - no prose, no markdown fences.

The JSON shape:
{
  "rule_id": "kebab-case-slug",
  "description": "one sentence, what the rule books and why",
  "priority": 100,
  "spec": {
    "match": {
      "event_type": "<the sample event's type>",
      "where": [ {"path": "$.field", "op": "eq|ne|exists|gt|lt", "value": ...} ]
    },
    "effect": {
      "object": {
        "type": "<target object type name>",
        "fields": { "<field>": "<template>", ... }
      }
    }
  }
}

Template language for field values:
- a plain string is a literal
- "=$.path.to.value" copies a value from the event payload (indexing: $.lines[0].x)
- "=sum($.lines[*].amount)" sums decimal-string amounts into a money value
- "=ref(T, field, $.path)" resolves a reference: the id of the single existing object
  of type T whose field equals the payload value (usually T's label field, e.g. name)
Money fields MUST use "=$.path" to a decimal string or "=sum(...)". Dates are copied as strings.
A field typed "ref<T>" MUST use "=ref(T, ...)". A field typed "enum" only accepts one of
its declared "values".

Every required field of the target object type must be present in "fields".
Use "where" conditions only for constraints the user's intent actually implies.

For MULTI-LINE events (one object per array element), add "each" to the object:
  "effect": {"object": {"type": "stock_movement", "each": "=$.lines[*]", "fields": {
    "item": "=ref(item, sku, $.line.item)", "qty": "=$.line.qty", "date": "=$.doc.date"}}}
Inside an "each" rule, field templates see {"doc": the whole event payload,
"line": one array element, "n": the 1-based line number} instead of the payload.

For ACCOUNTING rules (booking to a chart of accounts), "effect" instead posts one
balanced journal entry — no "object":
  "effect": {"postings": {"currency": "=$.currency", "lines": [
    {"account": "201", "debit": "=sum($.lines[*].amount)"},
    {"account": "702", "credit": "=sum($.lines[*].amount)"}
  ]}}
Each line names an account code (literal or "=$.path") and exactly ONE of debit or
credit (a money template). Debits must equal credits; the kernel rejects anything else.`

// DraftRule asks the model for a rule and strictly validates the answer
// against the language and the target object type before returning it.
func (a *Agent) DraftRule(ctx context.Context, intent string, sample core.Event, objType core.ObjectType) (Draft, error) {
	otJSON, _ := json.MarshalIndent(objType, "", "  ")
	user := fmt.Sprintf(
		"Target object type:\n%s\n\nSample event (type %q, payload):\n%s\n\nUser intent: %s",
		otJSON, sample.Type, string(sample.Payload), intent)

	raw, err := a.Complete(ctx, systemPrompt, user)
	if err != nil {
		return Draft{}, fmt.Errorf("model call: %w", err)
	}
	var d Draft
	if err := json.Unmarshal([]byte(extractJSON(raw)), &d); err != nil {
		return Draft{}, fmt.Errorf("model returned unparseable JSON: %w", err)
	}
	if d.RuleID == "" || d.Description == "" {
		return Draft{}, fmt.Errorf("draft missing rule_id or description")
	}
	if d.Priority == 0 {
		d.Priority = 100
	}
	if err := d.Spec.Validate(&objType); err != nil {
		return Draft{}, fmt.Errorf("draft failed validation: %w", err)
	}
	return d, nil
}

// extractJSON tolerates a model that wraps its answer despite instructions:
// it cuts from the first '{' to the last '}'.
func extractJSON(s string) string {
	start := strings.IndexByte(s, '{')
	end := strings.LastIndexByte(s, '}')
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}
