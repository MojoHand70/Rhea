// Package agent is the AI layer: it drafts rules as strict JSON from a plain-
// language intent, a sample event and the language as it stands. It has no
// store access and no side effects — the caller gathers the read-only context
// (Ask), the kernel validates the output and stores it as a draft (invariant
// 3). Model output is data, never executed.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
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

// Ask is everything the agent may see while drafting one rule: the whole
// vocabulary, not one pre-chosen target — choosing the effect and its type is
// the authoring. The caller fills it from the store; the agent only reads it.
type Ask struct {
	Intent string
	Sample core.Event
	Types  []core.ObjectType // every declared object type
	// Hint names the type the human pointed at, if any — a preference, not a
	// constraint: a posting or an amendment may answer the intent better.
	Hint string
	// Rules are the active rules: what is already explained, what cascades
	// from what, which priorities are taken.
	Rules []core.Rule
	// Reference is existing master data, per type: the states of objects a
	// ref() or a posting's account code can resolve against. The caller caps
	// it; the agent sees samples, never the whole ledger.
	Reference map[string][]map[string]any
	// Residue is the worklist as the agent sees it: every unexplained event
	// type with its count and one sample — what a bundle can answer at once.
	Residue []Cluster
}

// Cluster is one unexplained event shape in the worklist.
type Cluster struct {
	EventType string          `json:"event_type"`
	Count     int             `json:"count"`
	Sample    json.RawMessage `json:"sample"`
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
			MaxTokens: 4000,
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

const systemPrompt = `You draft rules for Rhea, a rule-driven ERP kernel. Every fact in Rhea is an
event in an append-only log; a rule explains an event by deriving state from it.
Nothing appears in state without a rule. You answer with ONE JSON object and
nothing else - no prose, no markdown fences.

The JSON shape:
` + ruleShape + ruleGrammar + `Never invent a type that is not in the catalog.`

// ruleShape is one rule as JSON — alone, or as an element of a bundle's rules.
const ruleShape = `{
  "rule_id": "kebab-case-slug",
  "description": "one sentence, what the rule does and why",
  "priority": 100,
  "spec": {
    "match": {
      "event_type": "<event type>",
      "where": [ {"path": "$.field", "op": "eq|ne|exists|gt|lt", "value": ...} ]
    },
    "effect": { exactly ONE of "object", "postings", "amend" - see below }
  }
}

`

// ruleGrammar is the rule language: templates, the three effects, cascade.
const ruleGrammar = `Use "where" conditions only for constraints the user's intent actually implies.
Priority orders firing among rules matching the same event: lower fires first.
Pick one not taken by the active rules listed in the request.

TEMPLATES - every field value in an effect is a template string:
- a plain string is a literal
- "=$.path.to.value" copies a value from the event payload (indexing: $.lines[0].x)
- "=sum($.lines[*].amount)" sums decimal-string amounts into a money value
- "=ref(T, field, $.path)" resolves a reference: the id of the single existing object
  of type T whose field equals the payload value (usually T's label field)
There is no other syntax: no arithmetic, no string building, no conditionals.
Money fields MUST use "=$.path" to a decimal string or "=sum(...)". Dates are copied
as strings. A field typed "ref<T>" MUST use "=ref(T, ...)", and the referenced object
must exist (see the master data in the request). A field typed "enum" only accepts
one of its declared "values".

EFFECT 1 - "object": materialize one object of a declared type.
  "effect": {"object": {"type": "invoice", "fields": {"<field>": "<template>", ...}}}
Every required field of the target type must be present.
For MULTI-LINE events (one object per array element), add "each":
  "effect": {"object": {"type": "stock_movement", "each": "=$.lines[*]", "fields": {
    "item": "=ref(item, sku, $.line.item)", "qty": "=$.line.qty", "date": "=$.doc.date"}}}
Inside an "each" rule, templates see {"doc": the whole event payload, "line": one
array element, "n": the 1-based line number} instead of the payload.

EFFECT 2 - "postings": post one balanced journal entry (accounting rules).
  "effect": {"postings": {"currency": "=$.currency", "lines": [
    {"account": "201", "debit": "=sum($.lines[*].amount)"},
    {"account": "702", "credit": "=sum($.lines[*].amount)"}
  ]}}
Each line names an account code (literal or "=$.path"; it must be the "code" of an
existing account object) and exactly ONE of debit or credit. Debits must equal
credits; the kernel rejects anything else. Optional beside "currency":
- "book": the ledger book (literal or template); omitted means "main". Parallel
  accounting standards are parallel rule-books, one book per entry.
- "convert": {"to": "PLN", "date": "=$.issue_date", "rounding": "half_up",
  "rounding_account": "756"} books the entry in a functional currency at the fx_rate
  for (from, to, date). Rounding must be declared; "half_up" is the only method.
  rounding_account takes the plug when per-line rounding breaks the balance.

EFFECT 3 - "amend": move an EXISTING object along its declared lifecycle.
  "effect": {"amend": {"type": "case", "target": "=$.case",
    "set": {"status": "resolved", "resolution": "=$.resolution"}}}
Only types with a "lifecycle" may be amended, and a status change must follow a
declared transition. "target" is "=$.path" carrying an object id, or
"=ref(T, field, $.path)" with T equal to the amended type - never a literal id.

CASCADE - rules may match derived events, not only raw ones. When a rule
materializes an object, the kernel emits "object.materialized" with payload
  {"object_id": "...", "object_type": "<type>", "state": {<the object's fields>}}
and rules matching it fire in the same atomic chain. Match it with
  "where": [{"path": "$.object_type", "op": "eq", "value": "<type>"}]
and read fields as "=$.state.<field>" and the new object's id as "=$.object_id".
This is how one fact grows its consequences: a document materializes, a cascade
rule posts it, another raises a follow-up case. Amendments end chains: never match
"object.amended".

LIMITS of the language as it stands - draft around them:
- In a cascade rule, a ref<T> field cannot be filled from "$.state.<field>": the state
  holds an object id, and ref() resolves by a field value. When a second object of one
  event needs refs, write a second rule matching the same RAW event instead - every rule
  matching an event fires on it, in priority order.
- An event is explained once: when its rules fire, later rules never revisit it.
  Every consequence of an event must be in place when its explanation activates.

Choose the effect the intent asks for. Prefer reusing declared types.
`

// DraftRule asks the model for a rule and strictly validates the answer
// against the language before returning it: the spec's own gate, plus the
// catalog — an effect may only target a declared type.
func (a *Agent) DraftRule(ctx context.Context, ask Ask) (Draft, error) {
	raw, err := a.Complete(ctx, systemPrompt, userMessage(ask))
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
	target, err := targetType(d.Spec, ask.Types)
	if err != nil {
		return Draft{}, fmt.Errorf("draft failed validation: %w", err)
	}
	if err := d.Spec.Validate(target); err != nil {
		return Draft{}, fmt.Errorf("draft failed validation: %w", err)
	}
	return d, nil
}

// targetType finds the declared type an object or amend effect names. A
// postings effect targets the kernel's posting machinery, validated
// structurally (nil target).
func targetType(spec core.RuleSpec, types []core.ObjectType) (*core.ObjectType, error) {
	name := spec.Effect.Object.Type
	if spec.Effect.Amend != nil {
		name = spec.Effect.Amend.Type
	}
	if name == "" {
		return nil, nil
	}
	i := slices.IndexFunc(types, func(t core.ObjectType) bool { return t.Name == name })
	if i < 0 {
		return nil, fmt.Errorf("effect targets %q, which is not a declared object type", name)
	}
	return &types[i], nil
}

// userMessage renders the Ask: the catalog, the running rules, master data,
// then the evidence and the intent — context before question.
func userMessage(ask Ask) string {
	var sb strings.Builder
	sb.WriteString("Object type catalog:\n")
	for _, t := range ask.Types {
		b, _ := json.Marshal(t)
		sb.Write(b)
		sb.WriteByte('\n')
	}
	sb.WriteString("\nActive rules (id, priority, matches, description):\n")
	if len(ask.Rules) == 0 {
		sb.WriteString("(none)\n")
	}
	for _, r := range ask.Rules {
		match, _ := json.Marshal(r.Spec.Match)
		fmt.Fprintf(&sb, "- %s p%d %s: %s\n", r.ID, r.Priority, match, r.Description)
	}
	sb.WriteString("\nExisting master data (samples, per type):\n")
	if len(ask.Reference) == 0 {
		sb.WriteString("(none)\n")
	}
	names := make([]string, 0, len(ask.Reference))
	for n := range ask.Reference {
		names = append(names, n)
	}
	slices.Sort(names)
	for _, n := range names {
		b, _ := json.Marshal(ask.Reference[n])
		fmt.Fprintf(&sb, "%s: %s\n", n, b)
	}
	if len(ask.Residue) > 0 {
		sb.WriteString("\nUnexplained events waiting in the worklist (type, count, one sample payload):\n")
		for _, c := range ask.Residue {
			fmt.Fprintf(&sb, "- %s ×%d: %s\n", c.EventType, c.Count, c.Sample)
		}
	}
	if ask.Sample.Type != "" {
		fmt.Fprintf(&sb, "\nSample event (type %q, occurred %s, payload):\n%s\n",
			ask.Sample.Type, ask.Sample.OccurredAt, string(ask.Sample.Payload))
	}
	if ask.Hint != "" {
		fmt.Fprintf(&sb, "\nThe user points at object type %q.\n", ask.Hint)
	}
	fmt.Fprintf(&sb, "\nUser intent: %s\n", ask.Intent)
	return sb.String()
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
