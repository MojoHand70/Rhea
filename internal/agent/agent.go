// Package agent is the AI layer: it drafts rules as strict JSON from a plain-
// language intent, a sample event and the language as it stands. It has no
// store access and no side effects — the caller gathers the read-only context
// (Ask), the kernel validates the output and stores it as a draft (invariant
// 3). Model output is data, never executed.
package agent

import (
	"context"
	"encoding/base64"
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
	// Priors are what Rhea has learned across installations: for questions
	// like this one, how many explain it which way — real counts only.
	Priors []Prior
	// Rejections are the earlier answers to this very question that were
	// refused, oldest first: a redraft answers the latest refusal instead of
	// starting over.
	Rejections []Rejection
	// Corrections are readings the owner replaced: an earlier version of an
	// active rule, superseded by the next with a reason. The author sees
	// them so an overruled reading never returns (the host remembers).
	Corrections []Correction
}

// Correction is one superseded reading of a rule and why it was replaced.
type Correction struct {
	RuleID      string `json:"rule_id"`
	Version     int    `json:"version"`     // the reading that was replaced
	Description string `json:"description"` // what it said
	ReplacedBy  int    `json:"replaced_by"` // the version that replaced it
	By          string `json:"by"`
	Reason      string `json:"reason,omitempty"`
}

// Prior is one learned answer: Count of Of installations answer Question
// with Rule (anonymized: "?" marks values each business fills itself).
type Prior struct {
	Question string          `json:"question"`
	Count    int             `json:"count"`
	Of       int             `json:"of"`
	Rule     json.RawMessage `json:"rule"`
}

// Rejection is one earlier answer to this question that was refused, with
// the reason: the conversation so far. Proposal is the refused bundle as it
// was proposed; By names who refused it — a person, or the approver that
// read the dry run.
type Rejection struct {
	Proposal json.RawMessage `json:"proposal,omitempty"`
	Reason   string          `json:"reason"`
	By       string          `json:"by"`
}

// Cluster is one unexplained event shape in the worklist.
type Cluster struct {
	EventType string          `json:"event_type"`
	Count     int             `json:"count"`
	Sample    json.RawMessage `json:"sample"`
}

// Agent drafts rules. Complete is the single LLM call, injectable for tests.
// Read is the model reading a document — a PDF's bytes with an instruction
// — for the model-read channels; nil when the agent cannot read files.
// Speak is the host speaking — the same call with the model's thinking
// switched off: the words are daily and easy, and a local model then answers
// in seconds instead of a minute (DECISIONS 2026-10-15). Nil falls back to
// Complete, so a test's injected Complete hears the host too.
type Agent struct {
	Complete func(ctx context.Context, system, user string) (string, error)
	Read     func(ctx context.Context, system string, pdf []byte, user string) (string, error)
	Speak    func(ctx context.Context, system, user string) (string, error)
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
	ask := func(ctx context.Context, think bool, system string, blocks ...anthropic.ContentBlockParamUnion) (string, error) {
		params := anthropic.MessageNewParams{
			Model:     anthropic.Model(Model()),
			MaxTokens: 16000,
			System:    []anthropic.TextBlockParam{{Text: system}},
			Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(blocks...)},
		}
		if !think {
			// The model's default stands when thinking is wanted; switched
			// off explicitly, a local model skips its reasoning pass (Ollama
			// honours the same parameter on its Anthropic endpoint).
			params.Thinking = anthropic.ThinkingConfigParamUnion{OfDisabled: &anthropic.ThinkingConfigDisabledParam{}}
		}
		msg, err := client.Messages.New(ctx, params)
		if err != nil {
			return "", err
		}
		var sb strings.Builder
		for _, block := range msg.Content {
			sb.WriteString(block.Text)
		}
		return sb.String(), nil
	}
	return &Agent{
		Complete: func(ctx context.Context, system, user string) (string, error) {
			return ask(ctx, true, system, anthropic.NewTextBlock(user))
		},
		// The document goes to the model as it is — the bytes, base64 — with
		// the instruction beside it; the model's text comes back as data.
		Read: func(ctx context.Context, system string, pdf []byte, user string) (string, error) {
			doc := anthropic.NewDocumentBlock(anthropic.Base64PDFSourceParam{Data: base64.StdEncoding.EncodeToString(pdf)})
			return ask(ctx, true, system, doc, anthropic.NewTextBlock(user))
		},
		Speak: func(ctx context.Context, system, user string) (string, error) {
			return ask(ctx, false, system, anthropic.NewTextBlock(user))
		},
	}
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
      "where": [ {"path": "$.field", "op": "eq|ne|exists|absent|gt|lt", "value": ...}
                 or {"formula": "=<an expression that is true or false>"} ]
    },
    "effect": { exactly ONE of "object", "postings", "amend", "matching" - see below }
  }
}

`

// ruleGrammar is the rule language: templates, the three effects, cascade.
const ruleGrammar = `Use "where" conditions only for constraints the user's intent actually implies.
A condition is a path test, or a formula when it must read through a link to a record the event
names: {"formula": "=ref(company, vat_id, $.seller_nip).vat_status = \"active\""} holds when that
company says so (a field the record lacks makes the formula false; a record that does not exist
makes the event wait).
Priority orders firing among rules matching the same event: lower fires first.
Pick one not taken by the active rules listed in the request.

TEMPLATES - every field value in an effect is a template string:
- a plain string is a literal
- "=$.path.to.value" copies a value from the event payload (indexing: $.lines[0].x)
- "=ref(T, field, $.path)" resolves a reference: the id of the single existing object
  of type T whose field equals the payload value (usually T's label field). The value
  may be a quoted literal: our own company is "=ref(company, kind, \"self\")" - a bare
  word is a name, not a value
- "=ref(T, field, ref(U, field2, $.path))" looks up one step through a link: the T whose
  field holds the U found by field2 (the case of complaint R-1:
  "=ref(case, complaint, ref(complaint, number, $.number))")
- any other "=" template is a FORMULA, computed at firing and baked into the event
  with its inputs, so every number explains itself:
  * arithmetic + - * and comparisons = <> < <= > >=, and/or/not, if(c, a, b),
    abs, min(a, b), max(a, b); numbers are exact, never floats
  * DIVISION ONLY UNDER round: "=round($.net * 23 / 100, 2, half_up)". Rounding is
    declared, never implied: round(x, places, method), method one of half_up,
    half_even, down, up. A money field holds two places: a product that has more
    (a 3-place unit cost) must be rounded on purpose, or the firing is refused.
  * reads through links: "$.state.item.std_cost" reads the std_cost of the item the
    causing object links to; "ref(item, sku, $.line.item).std_cost" reads it after
    a lookup. Values are read by their declared types (money as money). A raw payload
    value is text, never a link: "$.vat_rate.percent" fails - resolve first,
    "ref(vat_rate, code, $.vat_rate).percent".
  * dates: "=$.state.registered_on + 14" (days), date - date (days),
    end_of_month(d), add_months(d, n), month(d) — "YYYY-MM", the key a period is named by
  * folds over the event's own lines: "=sum(l in $.lines: l.qty * l.price)",
    count(l in $.lines where l.qty > 0), min/max/any/all(l in $.lines: ...),
    "=sum($.lines[*].amount)" sums a fanned-out path
  * folds over linked state, keyed and capped: objects(T, field, value) is the
    list of T objects whose field equals value (log order, at most 1000):
    "=sum(m in objects(stock_movement, item, $.state.item) where m.location = $.state.location:
      if(m.direction = \"in\", m.qty, -m.qty))" is the book quantity of an item in a location
  * no loops, no variables, no string building. Algorithms (FIFO, allocation) are not
    formulas: they arrive as kernel methods.
Money fields take "=$.path" to a decimal string, "=sum(...)" or a formula. Dates are
copied as strings or computed. A field typed "ref<T>" links to an object of type T, in
one of two ways: "=ref(T, field, $.path)" looks it up by a field value (the referenced
object must exist - see the master data in the request), or "=$.path" carries the exact
object id - in a cascade, "=$.object_id" is the object that caused it and
"=$.state.<ref field>" a link it holds. The kernel checks a carried id is a T and
exists. Never a literal id, never a computed one.
A field typed "enum" only accepts one of its declared "values". A field typed
"decimal" holds an exact decimal string (a rate, a unit cost). When the question
names fields, use exactly those names.

EFFECT 1 - "object": materialize one object of a declared type.
  "effect": {"object": {"type": "invoice", "fields": {"<field>": "<template>", ...}}}
Every required field of the target type must be present. Leave out optional fields
the event does not carry - never fill a field, least of all a ref, with an empty string.
A rule may set only fields its target type declares: check every field you set against
the catalog, or against the new version of the type your bundle declares.
For MULTI-LINE events (one object per array element), add "each":
  "effect": {"object": {"type": "stock_movement", "each": "=$.lines[*]", "fields": {
    "item": "=ref(item, sku, $.line.item)", "qty": "=$.line.qty", "date": "=$.doc.date"}}}
Inside an "each" rule, templates see {"doc": the whole event payload, "line": one
array element, "n": the 1-based line number} instead of the payload; in a cascade,
"$.root" (the raw event that started the chain) stays readable inside each too.

EFFECT 2 - "postings": post one balanced journal entry (accounting rules).
  "effect": {"postings": {"currency": "=$.currency", "lines": [
    {"account": "201", "debit": "=sum($.lines[*].amount)"},
    {"account": "702", "credit": "=sum($.lines[*].amount)"}
  ]}}
Each line names an account code (literal or "=$.path"; it must be the "code" of an
existing account object - see the master data in the request; never an account the
chart does not have) and exactly ONE of debit or credit. Post only what the question
asks to post: a question that says nothing about booking gets no postings. Debits must equal
credits; the kernel rejects anything else. Optional beside "currency":
- "book": the ledger book (literal or template); omitted means "main". Parallel
  accounting standards are parallel rule-books, one book per entry. When the
  question names a book (pl-stat, group), EVERY entry it describes carries
  "book": "<that name>" - naming it in the rule id does not book there.
- "convert": {"to": "PLN", "date": "=$.issue_date", "rounding": "half_up",
  "rounding_account": "756"} books the entry in a functional currency at the fx_rate
  for (from, to, date). Rounding must be declared; "half_up" is the only method.
  rounding_account takes the plug when per-line rounding breaks the balance.

EFFECT 3 - "amend": move an EXISTING object along its declared lifecycle.
  "effect": {"amend": {"type": "case", "target": "=$.case",
    "set": {"status": "resolved", "resolution": "=$.resolution"}}}
Amendment is consent-based, per field: a rule may set a type's lifecycle field (along
a declared transition) and the fields the type lists in "amendable"; every other field
is fixed at birth. Late information about an existing thing (an IBAN for an account)
is an amendment of an amendable field. "target" is "=$.path" carrying an object id, or
"=ref(T, field, $.path)" with T equal to the amended type - never a literal id.

CASCADE - rules may match derived events, not only raw ones. When a rule
materializes an object, the kernel emits "object.materialized" with payload
  {"object_id": "...", "object_type": "<type>", "state": {<the object's fields>}}
and rules matching it fire in the same atomic chain. Match it with
  "where": [{"path": "$.object_type", "op": "eq", "value": "<type>"}]
and read fields as "=$.state.<field>", the new object's id as "=$.object_id", and the
raw event that started the chain as "=$.root.<field>". A document with lines: one rule
materializes the header from the raw event, a cascade on the header fans out its lines
with "each": "=$.root.lines[*]", each line pointing at its header with
"=$.doc.object_id" (inside "each", "doc" is the cascade payload).
This is how one fact grows its consequences: a document materializes, a cascade
rule posts it, another raises a follow-up case. Amendments end chains: never match
"object.amended", and an amended status never re-materializes - a rule matching
"object.materialized" with a status set later by an amendment never fires. React to
the raw event instead (complaint.withdrawn), and find the object through what the
event names.

EFFECT 4 - "matching": match the object that just materialized against other sides
under a structure and an ordered policy - the kernel's matching engine (bank lines
against open invoices, receipts against orders, intercompany pairs). The rule matches
"object.materialized" of the trigger side's type:
  "effect": {"matching": {"name": "incoming-payments", "trigger": "payment",
    "sides": [
      {"name": "payment", "type": "bank_line", "open": {"money": "open"},
       "currency": "currency", "date": "date", "reference": "reference"},
      {"name": "invoice", "type": "sales_invoice",
       "items": "=objects(sales_invoice, buyer, ref(company, iban, $.state.counterparty_account))",
       "open": {"money": "open"}, "currency": "currency", "date": "issue_date", "reference": "number"}],
    "policy": [{"strategy": "reference"}, {"strategy": "exact"},
      {"strategy": "subset", "max_items": 4},
      {"strategy": "exact", "tolerance": {"money": "0.02"}, "difference_account": "755"},
      {"strategy": "oldest_first"}]}}
A side names its "type"; "items" is the keyed objects(...) read of its candidates, written
against the trigger's state (the trigger side has none - its item is the object that
fired); "join": {"<candidate field>": "<trigger field>"} narrows them to those agreeing
with the trigger (the order's lines of the same item: "join": {"item": "item"}); "open"
maps each measure to the AMENDABLE field holding what is not yet allocated
(every side declares the same measures); optional "compare" maps a measure to a field that
must agree but is never consumed (a price); "currency" (when a measure is money), "date",
and "reference" (needed by the reference strategy) name fields. Strategies are tried in
order and the first feasible wins: "reference" (the items the trigger's reference names;
"partial": true settles what it can), "exact" (one item agreeing on every measure),
"subset" (the fewest items summing to the trigger, "max_items" up to 8), "oldest_first"
(settles the earliest open items as far as the trigger reaches; the remainder stays open).
Any step may add "tolerance" or "tolerance_percent" per measure with a "difference_account"
(the difference is absorbed and recorded against it) and "within_days". The policy is the
company's choice: the order and the bounds are what the question says, nothing more.
Sides in different currencies take a "convert" clause beside "policy": {"functional": "PLN",
"date_side": "payment", "rounding": "half_up", "rate_up_account": "750", "rate_down_account":
"755"} - every candidate is read in the trigger's currency at the fx_rate of the item on
date_side's date, allocations are written back in each item's own currency, and the rate's
move since the item's own date is a difference of kind "fx", direction "up" or "down", in the
functional currency. Netting (kompensata) needs no clause: a structure whose trigger is our
purchase invoice and whose other side is the same party's open sales invoices.
The outcome is ONE event of type "objects.matched". For the structure above, a payment
settling two invoices with two grosz written off reads, for example:
  {"matching": "incoming-payments", "strategy": "subset", "step": 3,
   "trigger": {"side": "payment", "object_id": "bank_line-41"},
   "kinds": {"money": "money"},
   "sides": {"payment": [{"object_id": "bank_line-41", "object_type": "bank_line",
                          "allocated": {"money": 29998}, "open_before": {"money": 29998},
                          "open_after": {"money": 0}, "currency": "PLN", "date": "2026-09-14"}],
             "invoice": [{"object_id": "sales_invoice-12", "object_type": "sales_invoice",
                          "allocated": {"money": 10000}, "open_before": {"money": 10000},
                          "open_after": {"money": 0}, "currency": "PLN", "reference": "FV/1"},
                         {"object_id": "sales_invoice-13", "object_type": "sales_invoice",
                          "allocated": {"money": 20000}, "open_before": {"money": 20000},
                          "open_after": {"money": 0}, "currency": "PLN", "reference": "FV/2"}]},
   "differences": [{"kind": "tolerance", "measure": "money", "side": "invoice", "amount": 2,
                    "direction": "short", "currency": "PLN", "account": "755"}]}
Money amounts in it are minor units. It changes nothing by itself: ordinary rules in the
SAME answer explain it, matching "objects.matched" with
"where": [{"path": "$.matching", "op": "eq", "value": "incoming-payments"}] -
  the open amounts move by an amendment over the items:
  {"amend": {"type": "sales_invoice", "each": "=$.sides.invoice[*]", "target": "=$.line.object_id",
    "set": {"open": "=$.line.open_after.money",
            "status": "=if($.line.open_after.money = 0, \"paid\", \"open\")"}}}
  (a type whose invoices may stay open under a part payment declares the transition
  open -> open); a settlement object per item the same way ("object" with "each":
  "=$.sides.invoice[*]", "payment": "=$.doc.trigger.object_id", "amount":
  "=$.line.allocated.money", "date": "=$.root.date"); the differences as objects
  ("each": "=$.differences[*]", with "where": [{"path": "$.differences[0]", "op": "exists"}])
  that a cascade posts. When no step fits, the kernel emits "objects.unmatched" with every
  step's reason, and the item stays open for a person.

A follow-up points at exactly the thing that caused it: a case raised from a complaint
carries "complaint": "=$.object_id", and a later event naming only the complaint's number
finds that case through the link (see the nested ref above). A cascade may also copy a
link the causing object holds, e.g. "item": "=$.state.item". Create only the objects the
question asks for.

Put every consequence of an event in the same answer: rules approved later reach events
already explained only through a separate, human-approved backfill.

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
	// A refused draft comes back with its error: the gate's verdict is
	// final, but what was proposed is evidence for whoever reads the run.
	if _, err := targetType(d.Spec, ask.Types); err != nil {
		return d, fmt.Errorf("draft failed validation: %w", err)
	}
	if err := d.Spec.Check(CatalogOf(ask.Types)); err != nil {
		return d, fmt.Errorf("draft failed validation: %w", err)
	}
	return d, nil
}

// CatalogOf turns a type list into the catalog the typed draft gate reads:
// formulas are checked against the fields they read through links and the
// fields they write.
func CatalogOf(types []core.ObjectType) core.Catalog {
	return func(name string) (core.ObjectType, bool) {
		i := slices.IndexFunc(types, func(t core.ObjectType) bool { return t.Name == name })
		if i < 0 {
			return core.ObjectType{}, false
		}
		return types[i], true
	}
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
	if len(ask.Corrections) > 0 {
		sb.WriteString("\nReadings the owner corrected before - superseded, never to be proposed again:\n")
		for _, c := range ask.Corrections {
			fmt.Fprintf(&sb, "- %s v%d (%q) was replaced by v%d by %s", c.RuleID, c.Version, c.Description, c.ReplacedBy, c.By)
			if c.Reason != "" {
				fmt.Fprintf(&sb, ": %s", c.Reason)
			}
			sb.WriteString("\n")
		}
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
	if len(ask.Priors) > 0 {
		sb.WriteString("\nWhat Rhea has learned from other installations (question, support, anonymized rule):\n")
		for _, p := range ask.Priors {
			fmt.Fprintf(&sb, "- %s: %d of %d installations: %s\n", p.Question, p.Count, p.Of, p.Rule)
		}
	}
	if ask.Sample.Type != "" {
		fmt.Fprintf(&sb, "\nSample event (type %q, occurred %s, payload):\n%s\n",
			ask.Sample.Type, ask.Sample.OccurredAt, string(ask.Sample.Payload))
	}
	if ask.Hint != "" {
		fmt.Fprintf(&sb, "\nThe user points at object type %q.\n", ask.Hint)
	}
	if len(ask.Rejections) > 0 {
		sb.WriteString("\nThe conversation so far - earlier answers to this question, refused, oldest first:\n")
		for i, rj := range ask.Rejections {
			fmt.Fprintf(&sb, "%d. Refused by %s: %s\n", i+1, rj.By, rj.Reason)
			if len(rj.Proposal) > 0 {
				fmt.Fprintf(&sb, "   The refused proposal was: %s\n", rj.Proposal)
			}
		}
		sb.WriteString("Answer the latest refusal: change what it objects to, keep what it does not, and use a new bundle_id.\n")
	}
	fmt.Fprintf(&sb, "\nUser intent: %s\n", ask.Intent)
	return sb.String()
}

// extractJSON tolerates a model that wraps its answer despite instructions:
// it decodes the first complete JSON object in the text, ignoring prose or
// code fences around it.
func extractJSON(s string) string {
	// The answer is the object the model meant: the largest object in the
	// text that looks like an answer (a model that prefaces its bundle with
	// prose carrying a snippet — even one with a rule_id in it — still wrote
	// the bundle as the biggest thing); failing that, the first object when
	// it parses, so the gate says what is missing; and when the first object
	// fails to parse and nothing else looks like an answer, the text from
	// that first brace — so the caller reports the real syntax error where
	// it is, never "no envelope" for a bundle that broke in the middle.
	first := strings.IndexByte(s, '{')
	if first < 0 {
		return s
	}
	looksLikeAnswer := func(v string) bool {
		return strings.Contains(v, `"bundle_id"`) || strings.Contains(v, `"rule_id"`)
	}
	var firstObj string
	firstOK := false
	best := ""
	for i := first; i < len(s); i++ {
		if s[i] != '{' {
			continue
		}
		var v json.RawMessage
		dec := json.NewDecoder(strings.NewReader(s[i:]))
		err := dec.Decode(&v)
		if i == first && err == nil {
			firstObj, firstOK = string(v), true
		}
		if err != nil {
			continue
		}
		if looksLikeAnswer(string(v)) && len(v) > len(best) {
			best = string(v)
		}
		i += int(dec.InputOffset()) - 1
	}
	switch {
	case best != "":
		return best // the largest object that looks like an answer: the bundle, never a snippet of one
	case firstOK:
		return firstObj // a lone object that is no answer: the gate says what is missing
	}
	return s[first:] // the first object does not parse: the real error, with its place
}

// NarrateSystem is what the host is told before it speaks (DIRECTION
// 2026-10-10, the concierge): phrase the facts, never compute, never
// repeat yesterday's words.
const NarrateSystem = `You are the host of a business's accounting system: discreet, precise, never obsequious,
like the best concierge. You are handed facts the system computed — tables with their
columns and rows — and you phrase them for the owner in a few sentences: what was done,
what waits for a decision, what stands out. Rules: every number you say must appear in
the facts exactly as given (same digits, same decimals) — you never compute, round, add
up or estimate; you never invent a fact; you never deny or pass over one — every fact
with an amount in it is spoken, with at least one of its amounts exactly as given, and
only an empty fact may be called empty; you do not repeat what you said before; you
do not greet with a template. Speak in the voice the owner asked for. Answer with the
words only - no JSON, no markdown, no headings.`

// Narrate asks the host to phrase facts: the facts as JSON, the owner's
// voice, what was said before, and the date. The words come back as data
// for the caller to check and log.
func (a *Agent) Narrate(ctx context.Context, today, voice string, facts json.RawMessage, before []string) (string, error) {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Today is %s.\n", today)
	if voice != "" {
		fmt.Fprintf(&sb, "The owner's voice: %s\n", voice)
	}
	if len(before) > 0 {
		sb.WriteString("\nWhat you said before (do not repeat it):\n")
		for _, b := range before {
			fmt.Fprintf(&sb, "- %s\n", b)
		}
	}
	fmt.Fprintf(&sb, "\nThe facts:\n%s\n\nSpeak.", facts)
	speak := a.Speak
	if speak == nil {
		speak = a.Complete
	}
	return speak(ctx, NarrateSystem, sb.String())
}
