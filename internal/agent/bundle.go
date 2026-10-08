package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"rhea/internal/core"
)

// BundleDraft is the agent's answer to one interview question (DIRECTION,
// implementation by interview): every definition the answer needs — new
// types, the rules explaining events into them, the verbs and views around
// them — proposed together, approved together or not at all. Versions are
// the kernel's to assign; the agent proposes content.
type BundleDraft struct {
	BundleID    string            `json:"bundle_id"`
	Description string            `json:"description"`
	ObjectTypes []core.ObjectType `json:"object_types,omitempty"`
	ViewDefs    []core.ViewDef    `json:"view_defs,omitempty"`
	Rules       []Draft           `json:"rules,omitempty"`
	Activities  []ActivityDraft   `json:"activities,omitempty"`
	// Warrant: why this is the proposal. Required — Rhea speaks first with
	// a standard, and says where it comes from.
	Warrant *core.Warrant `json:"warrant"`
}

// ActivityDraft is a proposed verb: a declared door with typed inputs.
type ActivityDraft struct {
	Name        string            `json:"name"`
	Domain      string            `json:"domain"`
	Description string            `json:"description"`
	Spec        core.ActivitySpec `json:"spec"`
}

const bundlePrompt = `You draft definitions for Rhea, a rule-driven ERP kernel, answering one
question from an implementation interview. Every fact in Rhea is an event in an
append-only log; rules explain events by deriving state from them. Nothing appears
in state without a rule. Your answer is a BUNDLE: every definition the answer needs,
which a human approves as a whole or rejects with a reason. Answer with ONE JSON
object and nothing else - no prose, no markdown fences.

The JSON shape:
{
  "bundle_id": "kebab-case-slug naming the answer",
  "description": "one or two sentences: what this bundle makes the system do",
  "object_types": [ ... new types only, see OBJECT TYPES ... ],
  "rules": [ ... one or more rules, each shaped as below ... ],
  "view_defs": [ ... optional, see VIEWS ... ],
  "activities": [ ... optional, see ACTIVITIES ... ],
  "warrant": { ... required, see WARRANT ... }
}

A rule:
` + ruleShape + ruleGrammar + `
OBJECT TYPES - declare a new type only when nothing in the catalog fits:
  {"name": "complaint", "domain": "sales", "is_document": true, "label_field": "number",
   "fields": [{"name": "number", "type": "string", "required": true},
              {"name": "status", "type": "enum", "values": ["open", "resolved"], "required": true}],
   "lifecycle": {"field": "status", "transitions": {"open": ["resolved"]}}}
Field types: string, int, date, money, enum (with "values"), ref<type>. Omit "version";
the kernel assigns it. Declare a "lifecycle" only if objects of the type have a status
that moves; its field must be an enum and the transitions name its values. List in
"amendable" the other fields rules may set after the object exists (details that
legitimately arrive later, a resolution); everything else is fixed at birth.
To extend an existing type, declare it again with the same name and the new fields:
it becomes the next version, and existing objects keep the version they were born
under.
Rules in the bundle may target the bundle's own new types.

VIEWS - every type already gets a derived list and detail view. Add a view_def only
when the question asks for a particular way of seeing:
  {"view_id": "open-complaints", "notion": "list", "title": "Complaints", "domain": "sales",
   "function": "complaints", "spec": {"object_type": "complaint",
   "columns": [{"field": "number", "label": "No."}]}}
Notions: "list" (spec: object_type, columns[{field,label}]), "detail" (spec: object_type,
sections[{title, fields[]}]).

ACTIVITIES - a declared verb a human uses to record something, emitting a raw event:
  {"name": "register_complaint", "domain": "sales", "description": "...",
   "spec": {"inputs": [{"name": "number", "type": "string", "required": true},
                       {"name": "occurred_at", "type": "date", "required": true}],
            "emits": "complaint.registered", "who": ["human"]}}
Add one only when the question is about something people will record by hand.

WARRANT - Rhea suggests standards: when the question leaves practice open ("it depends",
"everyone does it a bit differently"), propose the customary standard of the market rather
than leaving it out, and say where it comes from. The client may always replace it.
  "warrant": {"basis": "practice", "citations": ["dokument PZ dla każdej dostawy; Wn 330 / Ma 300"],
              "scope": "PL, trade"}
When the request shows what Rhea has learned from other installations, that is your
strongest evidence: reuse a learned rule (fill its "?" from this client's events) and use
basis "network" - Rhea verifies the match and adds the real count herself. Deviate when
this client's events or words call for it, and say so in the description.
Basis: "network" (a learned rule, as above), "statute" or "standard" (cite the act or
standard), "practice" (customary practice -
name it), "pack" (the market pack's default), "model" (your general knowledge, when nothing
firmer applies), "client" (what the client said). Never state percentages, counts or shares
of companies: Rhea adds real support figures herself, from what she has learned. Cite only
what you are sure of; a vague true citation beats a precise invented one.

FIRST check the active rules in the request: if they already answer the question (perhaps
because an earlier answer took up what Rhea had learned), propose nothing - return
{"bundle_id": "...", "description": "already answered by <rule ids>: <how>"} with no
definitions. Never propose a second rule doing what an active rule already does.

Explain the events the question is about - the unexplained events in the request show
what is waiting. One event may need several rules (a document AND a ledger posting AND a
stock movement); put them all in the bundle - they activate together. Propose only what
the question needs.`

// ErrAlreadyAnswered is the agent's answer when the active rules already
// explain what the question asks — often because a learned answer was taken
// up earlier. Not a failure: there is nothing to approve. The bundle's
// description says which rules answer it.
var ErrAlreadyAnswered = errors.New("already answered by the active rules")

// DraftBundle asks the model for a bundle and validates every member against
// the language: types by the type gate, rules against the catalog overlaid
// with the bundle's own types, views against the fields they show, verbs by
// the activity gate. Nothing is stored here.
func (a *Agent) DraftBundle(ctx context.Context, ask Ask) (BundleDraft, error) {
	raw, err := a.Complete(ctx, bundlePrompt, userMessage(ask))
	if err != nil {
		return BundleDraft{}, fmt.Errorf("model call: %w", err)
	}
	var b BundleDraft
	if err := json.Unmarshal([]byte(extractJSON(raw)), &b); err != nil {
		return BundleDraft{}, fmt.Errorf("model returned unparseable JSON: %w", err)
	}
	if b.Description != "" && len(b.ObjectTypes)+len(b.Rules)+len(b.ViewDefs)+len(b.Activities) == 0 {
		return b, ErrAlreadyAnswered
	}
	if err := b.Validate(ask.Types); err != nil {
		return BundleDraft{}, fmt.Errorf("bundle failed validation: %w", err)
	}
	return b, nil
}

// Validate checks a bundle draft against a catalog. Exported so recorded and
// hand-written bundles pass the same gate as the model's.
func (b *BundleDraft) Validate(catalog []core.ObjectType) error {
	if b.BundleID == "" || b.Description == "" {
		return fmt.Errorf("bundle needs bundle_id and description")
	}
	if len(b.ObjectTypes)+len(b.Rules)+len(b.ViewDefs)+len(b.Activities) == 0 {
		return fmt.Errorf("bundle %s proposes nothing", b.BundleID)
	}
	if b.Warrant == nil {
		return fmt.Errorf("bundle %s has no warrant: say where the proposal comes from", b.BundleID)
	}
	if err := b.Warrant.ValidateProposed(); err != nil {
		return err
	}
	overlay := slices.Clone(catalog)
	seen := map[string]bool{}
	for _, t := range b.ObjectTypes {
		if seen["type:"+t.Name] {
			return fmt.Errorf("type %s declared twice", t.Name)
		}
		seen["type:"+t.Name] = true
		if err := t.Validate(); err != nil {
			return err
		}
		overlay = slices.DeleteFunc(overlay, func(o core.ObjectType) bool { return o.Name == t.Name })
		overlay = append(overlay, t)
	}
	for _, t := range b.ObjectTypes {
		for _, f := range t.Fields {
			if ref, ok := core.RefTarget(f.Type); ok && !slices.ContainsFunc(overlay, func(o core.ObjectType) bool { return o.Name == ref }) {
				return fmt.Errorf("type %s field %s references undeclared type %q", t.Name, f.Name, ref)
			}
		}
	}
	for i := range b.Rules {
		d := &b.Rules[i]
		if d.RuleID == "" || d.Description == "" {
			return fmt.Errorf("rule %d missing rule_id or description", i+1)
		}
		if seen["rule:"+d.RuleID] {
			return fmt.Errorf("rule %s proposed twice", d.RuleID)
		}
		seen["rule:"+d.RuleID] = true
		if d.Priority == 0 {
			d.Priority = 100
		}
		target, err := targetType(d.Spec, overlay)
		if err != nil {
			return fmt.Errorf("rule %s: %w", d.RuleID, err)
		}
		if err := d.Spec.Validate(target); err != nil {
			return fmt.Errorf("rule %s: %w", d.RuleID, err)
		}
	}
	for _, v := range b.ViewDefs {
		if seen["view:"+v.ID] {
			return fmt.Errorf("view %s proposed twice", v.ID)
		}
		seen["view:"+v.ID] = true
		if err := validateView(v, overlay); err != nil {
			return err
		}
	}
	for _, ad := range b.Activities {
		if seen["activity:"+ad.Name] {
			return fmt.Errorf("activity %s proposed twice", ad.Name)
		}
		seen["activity:"+ad.Name] = true
		if core.ReservedEventType(ad.Spec.Emits) {
			return fmt.Errorf("activity %s emits %q, a system verb", ad.Name, ad.Spec.Emits)
		}
		act := core.Activity{Name: ad.Name, Status: core.StatusDraft, Domain: ad.Domain,
			Description: ad.Description, Spec: ad.Spec}
		if err := act.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// validateView checks a list or detail view against the type it shows: the
// notions an agent may author, every field one the type declares.
func validateView(v core.ViewDef, types []core.ObjectType) error {
	if v.ID == "" || v.Title == "" || v.Domain == "" || v.Function == "" {
		return fmt.Errorf("view needs view_id, title, domain and function")
	}
	var spec struct {
		ObjectType string             `json:"object_type"`
		Columns    []core.ColumnSpec  `json:"columns"`
		Sections   []core.SectionSpec `json:"sections"`
	}
	if err := json.Unmarshal(v.Spec, &spec); err != nil {
		return fmt.Errorf("view %s spec: %w", v.ID, err)
	}
	i := slices.IndexFunc(types, func(t core.ObjectType) bool { return t.Name == spec.ObjectType })
	if i < 0 {
		return fmt.Errorf("view %s shows %q, which is not a declared object type", v.ID, spec.ObjectType)
	}
	var fields []string
	switch v.Notion {
	case "list":
		for _, c := range spec.Columns {
			fields = append(fields, c.Field)
		}
	case "detail":
		for _, s := range spec.Sections {
			fields = append(fields, s.Fields...)
		}
	default:
		return fmt.Errorf("view %s: notion %q is not one the agent authors (list, detail)", v.ID, v.Notion)
	}
	if len(fields) == 0 {
		return fmt.Errorf("view %s shows no fields", v.ID)
	}
	for _, f := range fields {
		if _, ok := types[i].Field(f); !ok {
			return fmt.Errorf("view %s: %q is not a field of %s", v.ID, f, spec.ObjectType)
		}
	}
	return nil
}

// Summary is a one-line account of what a bundle proposes.
func (b BundleDraft) Summary() string {
	var parts []string
	for _, p := range []struct {
		n    int
		noun string
	}{{len(b.ObjectTypes), "type"}, {len(b.Rules), "rule"}, {len(b.ViewDefs), "view"}, {len(b.Activities), "verb"}} {
		if p.n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s(s)", p.n, p.noun))
		}
	}
	return strings.Join(parts, ", ")
}
