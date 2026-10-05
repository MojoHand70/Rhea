package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Activity is a declared verb (SPEC §2): a named action a human or agent can
// take — typed inputs, the raw event it emits, who may trigger it. The schema
// belongs to the verb, not the event (DIRECTION 2026-10-05): inputs validate
// at trigger time, before the event exists; the emitted event is a raw event
// like any other — schemaless in the log, still needing a rule to explain it —
// and activities are never consulted during replay. Activities have the rule
// lifecycle, where `active` means *offered and triggerable*, never *executes
// in replay*: superseding an activity never touches the events it emitted.
type Activity struct {
	Name        string       `json:"name"`
	Version     int          `json:"version"`
	Status      string       `json:"status"` // draft | approved | active | superseded
	Domain      string       `json:"domain"`
	Description string       `json:"description"`
	Spec        ActivitySpec `json:"spec"`
	CreatedBy   string       `json:"created_by"`
	CreatedAt   time.Time    `json:"created_at"`
}

type ActivitySpec struct {
	Inputs []FieldDef `json:"inputs"`
	// Emits is the raw event type the activity appends: a literal, or the
	// passthrough form "=<input>" taking the type from a string input — the
	// open door declared as data (submit_event).
	Emits string `json:"emits"`
	// Who may trigger: actor-pattern strings ("human", "agent"), declared and
	// displayed, never enforced — the seed of authz-in-the-language, which
	// stays a falsifiability test of its own (DIRECTION). Attribution is the
	// actor column; this field is vocabulary.
	Who []string `json:"who,omitempty"`
}

// ReservedEventType reports whether an event type belongs to the kernel: the
// materialization namespace, and the rule.* / activity.* system-verb
// namespaces that only declared system activities may speak in.
func ReservedEventType(t string) bool {
	return t == EventObjectMaterialized ||
		strings.HasPrefix(t, "rule.") || strings.HasPrefix(t, "activity.")
}

// emitsSource returns the input name a passthrough emits template reads:
// "=event_type" → ("event_type", true).
func (sp ActivitySpec) emitsSource() (string, bool) {
	if strings.HasPrefix(sp.Emits, "=") {
		return sp.Emits[1:], true
	}
	return "", false
}

// envelopeInput reports whether an input is consumed by the event envelope
// rather than the payload: the emits source and the business date.
func (sp ActivitySpec) envelopeInput(name string) bool {
	if src, ok := sp.emitsSource(); ok && name == src {
		return true
	}
	return name == "occurred_at"
}

// Validate checks an activity declaration. Beyond the ObjectType field
// vocabulary, inputs admit "json": the schemaless passthrough. A passthrough
// activity has nothing else to say about the payload — the json input must be
// the only payload-contributing one — so declared verbs stay typed and the
// escape hatch stays a hatch, not a loophole.
func (a Activity) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("activity needs a name")
	}
	switch a.Status {
	case StatusDraft, StatusApproved, StatusActive, StatusSuperseded:
	case "":
	default:
		return fmt.Errorf("activity %q: unknown status %q", a.Name, a.Status)
	}
	sp := a.Spec
	if sp.Emits == "" {
		return fmt.Errorf("activity %q: emits is required", a.Name)
	}
	jsonInputs := 0
	byName := map[string]FieldDef{}
	for _, f := range sp.Inputs {
		if f.Name == "" {
			return fmt.Errorf("activity %q: input needs a name", a.Name)
		}
		if _, dup := byName[f.Name]; dup {
			return fmt.Errorf("activity %q: duplicate input %q", a.Name, f.Name)
		}
		byName[f.Name] = f
		switch f.Type {
		case "string", "int", "date", "money":
		case "enum":
			if len(f.Values) == 0 {
				return fmt.Errorf("activity %q input %q: enum needs values", a.Name, f.Name)
			}
		case "json":
			jsonInputs++
		default:
			if _, ok := RefTarget(f.Type); !ok {
				return fmt.Errorf("activity %q input %q: unknown type %q", a.Name, f.Name, f.Type)
			}
		}
		if f.Type != "enum" && len(f.Values) > 0 {
			return fmt.Errorf("activity %q input %q: values only belong on enum inputs", a.Name, f.Name)
		}
	}
	if src, ok := sp.emitsSource(); ok {
		fd, declared := byName[src]
		if !declared || fd.Type != "string" || !fd.Required {
			return fmt.Errorf("activity %q: emits %q wants a required string input %q", a.Name, sp.Emits, src)
		}
	}
	if fd, ok := byName["occurred_at"]; ok && fd.Type != "date" {
		return fmt.Errorf("activity %q: occurred_at is the business date and must be a date input", a.Name)
	}
	if jsonInputs > 1 {
		return fmt.Errorf("activity %q: at most one json input — one event, one payload", a.Name)
	}
	if jsonInputs == 1 {
		for _, f := range sp.Inputs {
			if f.Type != "json" && !sp.envelopeInput(f.Name) {
				return fmt.Errorf("activity %q: a passthrough activity has nothing else to say about the payload; input %q must go", a.Name, f.Name)
			}
		}
	}
	return nil
}

// RefInput is a declared ref input and the id a trigger handed it; existence
// and type are the caller's to verify against state — core stays pure.
type RefInput struct {
	Input      string
	TargetType string
	ObjectID   string
}

// BuiltEvent is what a validated trigger becomes: the envelope of the raw
// event the door will append. The payload carries the canonical boundary
// encoding — money as decimal strings (invariant 6), refs as object ids —
// indistinguishable from the same fact submitted free-form.
type BuiltEvent struct {
	Type       string
	OccurredAt string
	Payload    json.RawMessage
	Refs       []RefInput
}

// BuildEvent validates trigger inputs against the declaration and shapes the
// event: the door's courtesy, spent entirely before the event exists. Unknown
// inputs are refused — a typed verb is strict or it is not typed. `today` is
// the business date when the activity declares no occurred_at input.
func (a Activity) BuildEvent(in map[string]any, today string) (BuiltEvent, error) {
	sp := a.Spec
	declared := map[string]bool{}
	for _, f := range sp.Inputs {
		declared[f.Name] = true
	}
	for name := range in {
		if !declared[name] {
			return BuiltEvent{}, fmt.Errorf("activity %q has no input %q", a.Name, name)
		}
	}

	out := BuiltEvent{Type: sp.Emits, OccurredAt: today}
	payload := map[string]any{}
	for _, f := range sp.Inputs {
		v, given := in[f.Name]
		if !given || v == nil || v == "" {
			if f.Required {
				return BuiltEvent{}, fmt.Errorf("input %q is required", f.Name)
			}
			continue
		}
		cv, err := coerceInput(f, v)
		if err != nil {
			return BuiltEvent{}, err
		}
		if target, isRef := RefTarget(f.Type); isRef {
			out.Refs = append(out.Refs, RefInput{Input: f.Name, TargetType: target, ObjectID: cv.(string)})
		}
		if f.Name == "occurred_at" {
			out.OccurredAt = cv.(string)
			continue
		}
		if src, ok := sp.emitsSource(); ok && f.Name == src {
			out.Type = cv.(string)
			continue
		}
		if f.Type == "json" {
			// The passthrough: this value IS the payload, so an event through
			// the open door reads exactly like one a rule will be asked to
			// explain — never nested under the input's name.
			b, err := json.Marshal(cv)
			if err != nil {
				return BuiltEvent{}, err
			}
			out.Payload = b
			continue
		}
		payload[f.Name] = cv
	}

	if _, ok := sp.emitsSource(); ok && ReservedEventType(out.Type) {
		return BuiltEvent{}, fmt.Errorf("event type %q is a kernel namespace — the open door does not speak it", out.Type)
	}
	if out.OccurredAt == "" {
		return BuiltEvent{}, fmt.Errorf("activity %q: no business date — declare an occurred_at input or let the door stamp today", a.Name)
	}
	if _, err := time.Parse("2006-01-02", out.OccurredAt); err != nil {
		return BuiltEvent{}, fmt.Errorf("occurred_at %q: want YYYY-MM-DD", out.OccurredAt)
	}
	if out.Payload == nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return BuiltEvent{}, err
		}
		out.Payload = b
	}
	return out, nil
}

// coerceInput checks one value against its declared type and returns it in
// the payload's canonical encoding.
func coerceInput(f FieldDef, v any) (any, error) {
	switch f.Type {
	case "string":
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("input %q wants a string, got %T", f.Name, v)
		}
		return s, nil
	case "int":
		switch n := v.(type) {
		case int:
			return int64(n), nil
		case int64:
			return n, nil
		case float64:
			if n != float64(int64(n)) {
				return nil, fmt.Errorf("input %q wants an integer, got %v", f.Name, n)
			}
			return int64(n), nil
		}
		return nil, fmt.Errorf("input %q wants an integer, got %T", f.Name, v)
	case "date":
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("input %q wants a YYYY-MM-DD date, got %T", f.Name, v)
		}
		if _, err := time.Parse("2006-01-02", s); err != nil {
			return nil, fmt.Errorf("input %q: %q is not a YYYY-MM-DD date", f.Name, s)
		}
		return s, nil
	case "money":
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("input %q wants money as a decimal string, got %T", f.Name, v)
		}
		if _, err := ParseMoney(s); err != nil {
			return nil, fmt.Errorf("input %q: %w", f.Name, err)
		}
		return s, nil // canonical boundary encoding: the string crosses, int64 lives in Go
	case "enum":
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("input %q wants one of %v, got %T", f.Name, f.Values, v)
		}
		for _, allowed := range f.Values {
			if s == allowed {
				return s, nil
			}
		}
		return nil, fmt.Errorf("input %q: %q is not one of %v", f.Name, s, f.Values)
	case "json":
		m, ok := v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("input %q wants a JSON object, got %T", f.Name, v)
		}
		return m, nil
	default: // ref<T>, vetted by Validate
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("input %q wants an object id, got %T", f.Name, v)
		}
		return s, nil
	}
}

// BuiltinActivities are the kernel's own verbs, seeded active at init with
// their activation events in the log — the gate's own birth is recorded. The
// retrofit test (DIRECTION 2026-10-05): the shell's hardcoded verbs expressed
// as the same declared data every future verb will be.
func BuiltinActivities() []Activity {
	return []Activity{
		{
			Name: "submit_event", Domain: "system",
			Description: "The open door: submit any raw event. Fail-as-data depends on it — what no rule explains yet waits in the worklist, and the residue is the completeness gauge.",
			Spec: ActivitySpec{
				Inputs: []FieldDef{
					{Name: "event_type", Type: "string", Required: true},
					{Name: "occurred_at", Type: "date", Required: true},
					{Name: "payload", Type: "json", Required: true},
				},
				Emits: "=event_type",
				Who:   []string{"human", "agent"},
			},
		},
		{
			Name: "approve_rule", Domain: "system",
			Description: "Activate a draft rule. The approval is recorded as rule.approved (invariant 2); the kernel reacts by writing the new active version row.",
			Spec: ActivitySpec{
				Inputs: []FieldDef{
					{Name: "rule_id", Type: "string", Required: true}, // a definition ref; FieldDef refs point at ObjectTypes (recorded contortion, exit at ref-to-definition)
					{Name: "approved_by", Type: "string", Required: true},
					{Name: "occurred_at", Type: "date", Required: true},
				},
				Emits: "rule.approved",
				Who:   []string{"human"},
			},
		},
		{
			Name: "approve_activity", Domain: "system",
			Description: "Activate a draft activity — the gate applied to the gate; nothing in Rhea needs a second governance mechanism.",
			Spec: ActivitySpec{
				Inputs: []FieldDef{
					{Name: "activity", Type: "string", Required: true},
					{Name: "approved_by", Type: "string", Required: true},
					{Name: "occurred_at", Type: "date", Required: true},
				},
				Emits: "activity.approved",
				Who:   []string{"human"},
			},
		},
		{
			Name: "draft_rule", Domain: "system",
			Description: "Ask for a rule in plain language. Records the request as rule.draft_requested; the agent answers as an actor outside the kernel — a reaction must be deterministic, and drafting is not.",
			Spec: ActivitySpec{
				Inputs: []FieldDef{
					{Name: "intent", Type: "string", Required: true},
					{Name: "sample_event_id", Type: "int", Required: true},
					{Name: "object_type", Type: "string", Required: true},
				},
				Emits: "rule.draft_requested",
				Who:   []string{"human"},
			},
		},
	}
}
