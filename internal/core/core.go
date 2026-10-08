// Package core defines the language of ERP: events, object types, objects,
// rules and view definitions. Everything here is data — storable, versioned,
// loadable into a running kernel. See SPEC.md §2.
package core

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"
)

// Event is an immutable fact, the only source of state change.
type Event struct {
	ID           int64           `json:"event_id"`
	Kind         string          `json:"kind"` // "raw" | "derived"
	Type         string          `json:"event_type"`
	OccurredAt   string          `json:"occurred_at"` // business date, YYYY-MM-DD
	RecordedAt   time.Time       `json:"recorded_at"`
	Payload      json.RawMessage `json:"payload"`
	CauseEventID *int64          `json:"cause_event_id,omitempty"`
	RuleID       string          `json:"rule_id,omitempty"`
	RuleVersion  int             `json:"rule_version,omitempty"`
	DedupKey     string          `json:"dedup_key,omitempty"`
	// Actor is who caused the event: "cli:<user>", "shell:<user>",
	// "agent:<model>", "kernel". Empty on events from before actors existed.
	Actor string `json:"actor,omitempty"`
	// ActivityName/ActivityVersion stamp the declared door a raw event came
	// through — the provenance symmetry to (RuleID, RuleVersion) on derived
	// events: every event names its door. Empty on adapter and pack events
	// (their door is the actor and the dedup key) and on history from before
	// doors existed.
	ActivityName    string `json:"activity_name,omitempty"`
	ActivityVersion int    `json:"activity_version,omitempty"`
}

const (
	KindRaw     = "raw"
	KindDerived = "derived"

	// EventObjectMaterialized is the derived event the executor emits when a
	// rule fires; its payload is a MaterializedObject.
	EventObjectMaterialized = "object.materialized"

	// EventObjectAmended is the derived event an amend effect emits: an
	// existing object's declared lifecycle moving, as a logged, provenanced
	// delta — status is a projection, never an update (DECISIONS 2026-10-05,
	// the amendment). Its payload is an AmendedObject.
	EventObjectAmended = "object.amended"

	// Time passing is a fact in the log (SPEC M5): the clock adapter appends
	// these, rules match them, and replay reproduces when the system knew
	// time had passed. Rules never read the wall clock.
	EventDayOpened   = "time.day_opened"   // payload {date}
	EventMonthOpened = "time.month_opened" // payload {month, date}
)

// TimeEventType reports whether an event is time passing. Not reserved —
// rules match time like any fact — but infrastructure: a day no rule reacts
// to is normal, so the worklist and the simulator's unexplained count leave
// the time.* namespace out, while the processing path still fires on it.
func TimeEventType(t string) bool { return strings.HasPrefix(t, "time.") }

// MaterializedObject is the payload of an object.materialized event and the
// unit from which all object projections (Postgres cache, DuckDB) are built.
type MaterializedObject struct {
	ObjectID    string         `json:"object_id"`
	ObjectType  string         `json:"object_type"`
	TypeVersion int            `json:"type_version"`
	State       map[string]any `json:"state"`
	// Calc explains every computed field: the formula and the inputs it
	// read, baked at firing so replay never recomputes and the walk can show
	// the arithmetic (invariant 5 for computed values). Absent on copies.
	Calc map[string]Calc `json:"calc,omitempty"`
}

// AmendedObject is the payload of an object.amended event: the delta a rule
// firing applied to an existing object, values baked at firing time so
// replay never re-evaluates. The delta is the explanation; the object's
// current state is the projection of its materialization plus every
// amendment, in log order.
type AmendedObject struct {
	ObjectID   string          `json:"object_id"`
	ObjectType string          `json:"object_type"`
	Set        map[string]any  `json:"set"`
	Calc       map[string]Calc `json:"calc,omitempty"`
}

// ObjectType is a schema as data.
type ObjectType struct {
	Name       string     `json:"name"`
	Version    int        `json:"version"`
	Domain     string     `json:"domain"`
	IsDocument bool       `json:"is_document"`
	LabelField string     `json:"label_field,omitempty"` // field shown when another object references this one
	Fields     []FieldDef `json:"fields"`
	// Lifecycle declares that objects of this type have a life: which enum
	// field carries it and which transitions are legal. Declaring it is the
	// type's consent to amendment — a rule may amend only types with a
	// lifecycle, and the kernel validates every move of the lifecycle field
	// against the declared transitions, the way it validates balance.
	Lifecycle *LifecycleDef `json:"lifecycle,omitempty"`
	// Amendable lists the fields rules may set after materialization —
	// consent per field (DECISIONS 2026-10-08, enrichment): an IBAN
	// legitimately arrives later, an account number never changes. The
	// lifecycle field is amendable implicitly, under its transition law;
	// every other field is fixed at birth unless listed here.
	Amendable []string `json:"amendable,omitempty"`
}

// CanAmend reports whether rules may set a field after materialization.
func (t ObjectType) CanAmend(field string) bool {
	return (t.Lifecycle != nil && t.Lifecycle.Field == field) || slices.Contains(t.Amendable, field)
}

// LifecycleDef: the status field and its allowed moves, as data. The initial
// status at materialization is unconstrained beyond the enum's value set.
type LifecycleDef struct {
	Field       string              `json:"field"`
	Transitions map[string][]string `json:"transitions"` // from → allowed to
}

type FieldDef struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"` // string | int | decimal | date | money | enum | ref<type>
	Required bool     `json:"required"`
	Values   []string `json:"values,omitempty"` // enum: the allowed value set
}

// RefTarget extracts the target type from a ref field type:
// "ref<company>" → ("company", true).
func RefTarget(fieldType string) (string, bool) {
	if strings.HasPrefix(fieldType, "ref<") && strings.HasSuffix(fieldType, ">") {
		target := fieldType[4 : len(fieldType)-1]
		return target, target != ""
	}
	return "", false
}

// Validate checks an object type definition: known field types, enum value
// sets, syntactically sound ref targets, and a label_field that exists. Ref
// targets are not checked for existence — definitions may load in any order.
func (t ObjectType) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("object type needs a name")
	}
	if len(t.Fields) == 0 {
		return fmt.Errorf("object type %q has no fields", t.Name)
	}
	for _, f := range t.Fields {
		switch f.Type {
		case "string", "int", "decimal", "date", "money":
		case "enum":
			if len(f.Values) == 0 {
				return fmt.Errorf("field %q: enum needs values", f.Name)
			}
		default:
			if _, ok := RefTarget(f.Type); !ok {
				return fmt.Errorf("field %q: unknown type %q", f.Name, f.Type)
			}
		}
		if f.Type != "enum" && len(f.Values) > 0 {
			return fmt.Errorf("field %q: values only belong on enum fields", f.Name)
		}
	}
	if t.LabelField != "" {
		if _, ok := t.Field(t.LabelField); !ok {
			return fmt.Errorf("label_field %q is not a field of %q", t.LabelField, t.Name)
		}
	}
	for _, f := range t.Amendable {
		if _, ok := t.Field(f); !ok {
			return fmt.Errorf("amendable %q is not a field of %q", f, t.Name)
		}
	}
	if lc := t.Lifecycle; lc != nil {
		fd, ok := t.Field(lc.Field)
		if !ok {
			return fmt.Errorf("lifecycle field %q is not a field of %q", lc.Field, t.Name)
		}
		if fd.Type != "enum" {
			return fmt.Errorf("lifecycle field %q must be an enum", lc.Field)
		}
		if len(lc.Transitions) == 0 {
			return fmt.Errorf("lifecycle of %q declares no transitions", t.Name)
		}
		for from, tos := range lc.Transitions {
			if !slices.Contains(fd.Values, from) {
				return fmt.Errorf("lifecycle transition from %q: not a value of %q", from, lc.Field)
			}
			for _, to := range tos {
				if !slices.Contains(fd.Values, to) {
					return fmt.Errorf("lifecycle transition %q→%q: not a value of %q", from, to, lc.Field)
				}
			}
		}
	}
	return nil
}

func (t ObjectType) Field(name string) (FieldDef, bool) {
	for _, f := range t.Fields {
		if f.Name == name {
			return f, true
		}
	}
	return FieldDef{}, false
}

// Object is a projection of events, cached for serving, rebuildable by replay.
// Provenance (SourceEventID, RuleID, RuleVersion) explains every field: in M0
// an object is materialized wholly by one rule firing on one event.
type Object struct {
	ID            string         `json:"object_id"`
	Type          string         `json:"object_type"`
	TypeVersion   int            `json:"type_version"`
	State         map[string]any `json:"state"`
	SourceEventID int64          `json:"source_event_id"`
	RuleID        string         `json:"rule_id"`
	RuleVersion   int            `json:"rule_version"`
}

// Rule is the unit of system behavior. Versioned and append-only: any change,
// including a status transition, is a new version row.
type Rule struct {
	ID            string    `json:"rule_id"`
	Version       int       `json:"version"`
	Status        string    `json:"status"` // draft | approved | active | superseded
	Priority      int       `json:"priority"`
	EffectiveFrom string    `json:"effective_from"` // YYYY-MM-DD
	CreatedBy     string    `json:"created_by"`     // "agent" | "human"
	Description   string    `json:"description"`
	Spec          RuleSpec  `json:"spec"`
	CreatedAt     time.Time `json:"created_at"`
}

const (
	StatusDraft      = "draft"
	StatusApproved   = "approved"
	StatusActive     = "active"
	StatusSuperseded = "superseded"
)

// RuleSpec = match (predicate over an event) + effect (what to materialize).
type RuleSpec struct {
	Match  Match  `json:"match"`
	Effect Effect `json:"effect"`
}

type Match struct {
	EventType string      `json:"event_type"`
	Where     []Condition `json:"where,omitempty"`
}

type Condition struct {
	Path  string `json:"path"` // e.g. $.customer, $.lines[*].amount
	Op    string `json:"op"`   // eq | ne | exists | gt | lt
	Value any    `json:"value,omitempty"`
}

// Effect is what a firing rule does: materialize one object, post one
// balanced journal entry (the double-entry sub-language, SPEC M1), or amend
// an existing object's declared lifecycle (the amendment, DECISIONS
// 2026-10-05). Exactly one of the three; Object's zero value means absent.
type Effect struct {
	Object   ObjectTemplate    `json:"object,omitempty"`
	Postings *PostingsTemplate `json:"postings,omitempty"`
	Amend    *AmendTemplate    `json:"amend,omitempty"`
}

// AmendTemplate moves an existing object: Target resolves the object's id
// (a =$.path carrying an id the door or a prior resolution vouched, or a
// =ref() resolution — never a literal), Set maps fields to value templates.
// The amended type must declare a lifecycle — amendment is consent-based —
// and a Set touching the lifecycle field must move along a declared
// transition. An amendment whose target does not exist is a rule error and
// the event waits in the worklist; nothing half-applies.
type AmendTemplate struct {
	Type   string            `json:"type"`
	Target string            `json:"target"`
	Set    map[string]string `json:"set"`
}

// PostingsTemplate expands into the lines of one journal entry, materialized
// as `posting` objects. The kernel enforces, after expansion and before
// booking (the Sunbeetle lessons): one currency per entry, debits equal to
// credits, accounts resolved by code, and no posting into a period locked
// for the entry's book.
type PostingsTemplate struct {
	// Book names the ledger book this entry belongs to — parallel accounting
	// as parallel rule-books over one log (DECISIONS 2026-10-04, E1). A
	// literal or template; empty means "main". Declared at entry level, so an
	// entry cannot straddle books by construction.
	Book     string           `json:"book,omitempty"`
	Currency string           `json:"currency"` // template, e.g. "PLN" or "=$.currency"
	Convert  *ConvertTemplate `json:"convert,omitempty"`
	Lines    []PostingLine    `json:"lines"`
}

// ConvertTemplate books the entry in a functional currency (DECISIONS
// 2026-10-04, E3): line amounts evaluate and balance in the transaction
// currency, then book converted at the fx_rate object for (from, to, date),
// each posting carrying tx_amount/tx_currency. `to` equal to the transaction
// currency is the identity conversion — one rule explains domestic and
// foreign documents alike. Deterministic method vocabulary lives in the
// kernel; this clause is only the choice and its parameters.
type ConvertTemplate struct {
	To   string `json:"to"`   // template: the functional currency
	Date string `json:"date"` // template: the rate's application date
	// Rounding is statutory, so the stance is declared, never implied.
	// "half_up" is the one admitted method; others join by proof.
	Rounding string `json:"rounding"`
	// RoundingAccount takes the plug line when per-line rounding breaks the
	// functional balance — the residue stays visible and the ledger stays
	// balanced by invariant. Without it, a broken balance refuses the entry.
	RoundingAccount string `json:"rounding_account,omitempty"`
}

// PostingLine is one side of value movement: an account code (literal or
// template) and exactly one of debit/credit as a money template.
type PostingLine struct {
	Account string `json:"account"`
	Debit   string `json:"debit,omitempty"`
	Credit  string `json:"credit,omitempty"`
}

// ObjectTemplate maps object fields to values. A value starting with "=" is an
// expression (see expr.go); anything else is a literal string. With Each set —
// a "=$.path" fan-out into an array — the rule materializes one object per
// element (a multi-line document's lines, the posting pattern generalized),
// and field templates evaluate against {"doc": <the event payload>, "line":
// <the element>, "n": <1-based line number>} instead of the payload alone.
type ObjectTemplate struct {
	Type   string            `json:"type"`
	Each   string            `json:"each,omitempty"`
	Fields map[string]string `json:"fields"`
}

// Validate checks structural soundness of a rule spec against known ops and,
// when the target ObjectType is supplied, its field set. This is the gate all
// rules pass — agent-drafted and hand-written alike.
func (s RuleSpec) Validate(target *ObjectType) error {
	if s.Match.EventType == "" {
		return fmt.Errorf("match.event_type is required")
	}
	// Amendments end their branch of the chain: nothing fires on them, so a
	// rule matching object.amended would be silently dead — refused instead.
	if s.Match.EventType == EventObjectAmended {
		return fmt.Errorf("rules cannot match %s — amendments end chains", EventObjectAmended)
	}
	for _, c := range s.Match.Where {
		switch c.Op {
		case "eq", "ne", "exists", "gt", "lt":
		default:
			return fmt.Errorf("unknown condition op %q", c.Op)
		}
		if _, err := parsePath(c.Path); err != nil {
			return fmt.Errorf("condition path %q: %w", c.Path, err)
		}
	}
	effects := 0
	if s.Effect.Object.Type != "" || len(s.Effect.Object.Fields) > 0 {
		effects++
	}
	if s.Effect.Postings != nil {
		effects++
	}
	if s.Effect.Amend != nil {
		effects++
	}
	if effects > 1 {
		return fmt.Errorf("effect has more than one of object/postings/amend; a rule does one")
	}
	if s.Effect.Postings != nil {
		return s.Effect.Postings.validate()
	}
	if s.Effect.Amend != nil {
		return s.Effect.Amend.validate(target)
	}
	if s.Effect.Object.Type == "" {
		return fmt.Errorf("effect.object.type is required")
	}
	if len(s.Effect.Object.Fields) == 0 {
		return fmt.Errorf("effect.object.fields is empty")
	}
	if s.Effect.Object.Each != "" {
		pt, err := ParseTemplate(s.Effect.Object.Each)
		if err != nil {
			return fmt.Errorf("effect.object.each: %w", err)
		}
		if pt.kind != "path" {
			return fmt.Errorf("effect.object.each wants a =$.path fan-out, got %q", s.Effect.Object.Each)
		}
	}
	for name, tmpl := range s.Effect.Object.Fields {
		if err := checkFieldTemplate(target, name, tmpl); err != nil {
			return err
		}
	}
	if target != nil {
		if target.Name != s.Effect.Object.Type {
			return fmt.Errorf("effect targets %q but validated against %q", s.Effect.Object.Type, target.Name)
		}
		for _, f := range target.Fields {
			if f.Required {
				if _, ok := s.Effect.Object.Fields[f.Name]; !ok {
					return fmt.Errorf("required field %q missing from template", f.Name)
				}
			}
		}
	}
	return nil
}

// checkFieldTemplate validates one field template, against the target type
// when supplied: ref fields and ref() templates must pair up with matching
// targets (referential integrity only through resolution), and literal enum
// values must be in the declared set. Shared by materialization fields and
// amendment sets.
func checkFieldTemplate(target *ObjectType, name, tmpl string) error {
	pt, err := ParseTemplate(tmpl)
	if err != nil {
		return fmt.Errorf("field %q template: %w", name, err)
	}
	if target == nil {
		return nil
	}
	fd, ok := target.Field(name)
	if !ok {
		return fmt.Errorf("field %q not in object type %q", name, target.Name)
	}
	refTarget, isRef := RefTarget(fd.Type)
	switch {
	case isRef && pt.kind != "ref" && pt.kind != "path":
		return fmt.Errorf("field %q is %s and must use =ref(%s, <field>, $.path) or a =$.path carrying a %s id — never a literal or a computed id", name, fd.Type, refTarget, refTarget)
	case !isRef && pt.kind == "ref":
		return fmt.Errorf("field %q is %s, not a ref", name, fd.Type)
	case isRef && pt.kind == "ref" && pt.refType != refTarget:
		return fmt.Errorf("field %q is %s but template resolves a %q", name, fd.Type, pt.refType)
	}
	if err := pt.Check(nil, nil, fd); err != nil {
		return fmt.Errorf("field %q: %w", name, err)
	}
	if fd.Type == "enum" && pt.kind == "literal" && !slices.Contains(fd.Values, pt.raw) {
		return fmt.Errorf("field %q: %q is not one of %v", name, pt.raw, fd.Values)
	}
	return nil
}

// validate checks an amend template structurally; the lifecycle transition
// itself is judged at expansion, where the object's current status is known.
func (a *AmendTemplate) validate(target *ObjectType) error {
	if a.Type == "" {
		return fmt.Errorf("amend.type is required")
	}
	pt, err := ParseTemplate(a.Target)
	if err != nil {
		return fmt.Errorf("amend.target: %w", err)
	}
	switch pt.kind {
	case "path":
	case "ref":
		if pt.refType != a.Type {
			return fmt.Errorf("amend.target resolves a %q but the amendment is of %q", pt.refType, a.Type)
		}
	default:
		return fmt.Errorf("amend.target wants =$.path or =ref(...), never a literal id")
	}
	if len(a.Set) == 0 {
		return fmt.Errorf("amend.set is empty")
	}
	for name, tmpl := range a.Set {
		if err := checkFieldTemplate(target, name, tmpl); err != nil {
			return err
		}
	}
	if target != nil {
		if target.Name != a.Type {
			return fmt.Errorf("amend targets %q but validated against %q", a.Type, target.Name)
		}
		// Amendment is consent-based, per field: the lifecycle field moves
		// under its transition law, the declared amendable fields may be set,
		// everything else is fixed at birth.
		if target.Lifecycle == nil && len(target.Amendable) == 0 {
			return fmt.Errorf("type %q declares no lifecycle and no amendable fields — amendment refused", target.Name)
		}
		for name := range a.Set {
			if !target.CanAmend(name) {
				return fmt.Errorf("field %q of %q is fixed at birth — not its lifecycle field, not declared amendable", name, target.Name)
			}
		}
	}
	return nil
}

// validate checks a postings template structurally. Balance cannot be judged
// from templates, so it is enforced at expansion — an unbalanced expansion is
// a rule error and nothing books.
func (p *PostingsTemplate) validate() error {
	if p.Book != "" {
		if _, err := ParseTemplate(p.Book); err != nil {
			return fmt.Errorf("postings.book: %w", err)
		}
	}
	if p.Currency == "" {
		return fmt.Errorf("postings.currency is required")
	}
	if _, err := ParseTemplate(p.Currency); err != nil {
		return fmt.Errorf("postings.currency: %w", err)
	}
	if c := p.Convert; c != nil {
		if c.To == "" || c.Date == "" {
			return fmt.Errorf("postings.convert needs to and date")
		}
		if _, err := ParseTemplate(c.To); err != nil {
			return fmt.Errorf("postings.convert.to: %w", err)
		}
		if _, err := ParseTemplate(c.Date); err != nil {
			return fmt.Errorf("postings.convert.date: %w", err)
		}
		if c.Rounding != "half_up" {
			return fmt.Errorf("postings.convert.rounding must be declared; %q is not an admitted method", c.Rounding)
		}
		if c.RoundingAccount != "" {
			if _, err := ParseTemplate(c.RoundingAccount); err != nil {
				return fmt.Errorf("postings.convert.rounding_account: %w", err)
			}
		}
	}
	if len(p.Lines) < 2 {
		return fmt.Errorf("a journal entry needs at least two lines")
	}
	for i, l := range p.Lines {
		if l.Account == "" {
			return fmt.Errorf("line %d: account is required", i+1)
		}
		if _, err := ParseTemplate(l.Account); err != nil {
			return fmt.Errorf("line %d account: %w", i+1, err)
		}
		if (l.Debit == "") == (l.Credit == "") {
			return fmt.Errorf("line %d: exactly one of debit or credit", i+1)
		}
		amount := l.Debit + l.Credit // one of them is empty
		if _, err := ParseTemplate(amount); err != nil {
			return fmt.Errorf("line %d amount: %w", i+1, err)
		}
	}
	return nil
}

// ViewDef binds data to a view notion. The shell renders notions generically.
type ViewDef struct {
	ID       string          `json:"view_id"`
	Version  int             `json:"version"`
	Notion   string          `json:"notion"` // list | detail | action | analysis
	Title    string          `json:"title"`
	Domain   string          `json:"domain"`
	Function string          `json:"function"` // submenu grouping
	Spec     json.RawMessage `json:"spec"`
}

// ListSpec: spec of a `list` notion.
type ListSpec struct {
	ObjectType string       `json:"object_type"`
	Columns    []ColumnSpec `json:"columns"`
}

type ColumnSpec struct {
	Field string `json:"field"`
	Label string `json:"label"`
	// Type declares the column's field type for renderers (money, int, date,
	// string, enum, ref<...>). List and detail views derive types from the
	// ObjectType and ignore this; analysis columns have no ObjectType, so the
	// view declares them. Empty means undeclared: renderers get a bare string.
	Type string `json:"type,omitempty"`
}

// DetailSpec: spec of a `detail` notion.
type DetailSpec struct {
	ObjectType string        `json:"object_type"`
	Sections   []SectionSpec `json:"sections"`
}

type SectionSpec struct {
	Title  string   `json:"title"`
	Fields []string `json:"fields"`
}

// AnalysisSpec: spec of an `analysis` notion; SQL runs on the DuckDB read side.
type AnalysisSpec struct {
	SQL     string       `json:"sql"`
	Columns []ColumnSpec `json:"columns"`
}

// SchedulingSpec: spec of a `scheduling` notion (SPEC M5) — time-axis
// placement of objects: which type, and which date field places each one.
// Label and status derive from the ObjectType (label_field, lifecycle), the
// way derived views do: the axis is the view's only opinion.
type SchedulingSpec struct {
	ObjectType string `json:"object_type"`
	DateField  string `json:"date_field"`
}

// Check is the typed draft gate: Validate against the catalog's target type,
// then every formula checked with the catalog — reads through links name
// real fields, kinds combine lawfully, results fit their fields. The typed
// payload prefixes come from the match: a cascade on a materialized type T
// reads T's state under $.state (and under $.doc.state inside each).
func (s RuleSpec) Check(types Catalog) error {
	var target *ObjectType
	name := s.Effect.Object.Type
	if s.Effect.Amend != nil {
		name = s.Effect.Amend.Type
	}
	if name != "" {
		t, ok := types(name)
		if !ok {
			return fmt.Errorf("effect targets %q, which is not a declared object type", name)
		}
		target = &t
	}
	if err := s.Validate(target); err != nil {
		return err
	}
	typed := s.TypedPrefixes()
	check := func(fields map[string]string) error {
		for name, raw := range fields {
			pt, err := ParseTemplate(raw)
			if err != nil {
				return err
			}
			var fd FieldDef
			if target != nil {
				fd, _ = target.Field(name)
			}
			if err := pt.Check(types, typed, fd); err != nil {
				return fmt.Errorf("field %q: %w", name, err)
			}
		}
		return nil
	}
	switch {
	case s.Effect.Postings != nil:
		for i, l := range s.Effect.Postings.Lines {
			pt, _ := ParseTemplate(l.Debit + l.Credit)
			if err := pt.Check(types, typed, FieldDef{Name: "amount", Type: "money"}); err != nil {
				return fmt.Errorf("line %d amount: %w", i+1, err)
			}
		}
		return nil
	case s.Effect.Amend != nil:
		return check(s.Effect.Amend.Set)
	}
	return check(s.Effect.Object.Fields)
}

// TypedPrefixes names the payload prefixes a rule's templates read typed
// object state under: $.state for a cascade on a materialized type (the
// where condition names it), $.doc.state inside an each over that cascade.
func (s RuleSpec) TypedPrefixes() map[string]string {
	if s.Match.EventType != EventObjectMaterialized {
		return nil
	}
	for _, c := range s.Match.Where {
		if c.Path == "$.object_type" && c.Op == "eq" {
			if t, ok := c.Value.(string); ok {
				if s.Effect.Object.Each != "" {
					return map[string]string{"$.doc.state": t}
				}
				return map[string]string{"$.state": t}
			}
		}
	}
	return nil
}
