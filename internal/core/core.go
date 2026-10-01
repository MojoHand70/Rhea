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
}

const (
	KindRaw     = "raw"
	KindDerived = "derived"

	// EventObjectMaterialized is the derived event the executor emits when a
	// rule fires; its payload is a MaterializedObject.
	EventObjectMaterialized = "object.materialized"
)

// MaterializedObject is the payload of an object.materialized event and the
// unit from which all object projections (Postgres cache, DuckDB) are built.
type MaterializedObject struct {
	ObjectID    string         `json:"object_id"`
	ObjectType  string         `json:"object_type"`
	TypeVersion int            `json:"type_version"`
	State       map[string]any `json:"state"`
}

// ObjectType is a schema as data.
type ObjectType struct {
	Name       string     `json:"name"`
	Version    int        `json:"version"`
	Domain     string     `json:"domain"`
	IsDocument bool       `json:"is_document"`
	LabelField string     `json:"label_field,omitempty"` // field shown when another object references this one
	Fields     []FieldDef `json:"fields"`
}

type FieldDef struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"` // string | int | date | money | enum | ref<type>
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
		case "string", "int", "date", "money":
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

type Effect struct {
	Object ObjectTemplate `json:"object"`
}

// ObjectTemplate maps object fields to values. A value starting with "=" is an
// expression (see expr.go); anything else is a literal string.
type ObjectTemplate struct {
	Type   string            `json:"type"`
	Fields map[string]string `json:"fields"`
}

// Validate checks structural soundness of a rule spec against known ops and,
// when the target ObjectType is supplied, its field set. This is the gate all
// rules pass — agent-drafted and hand-written alike.
func (s RuleSpec) Validate(target *ObjectType) error {
	if s.Match.EventType == "" {
		return fmt.Errorf("match.event_type is required")
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
	if s.Effect.Object.Type == "" {
		return fmt.Errorf("effect.object.type is required")
	}
	if len(s.Effect.Object.Fields) == 0 {
		return fmt.Errorf("effect.object.fields is empty")
	}
	for name, tmpl := range s.Effect.Object.Fields {
		pt, err := ParseTemplate(tmpl)
		if err != nil {
			return fmt.Errorf("field %q template: %w", name, err)
		}
		if target == nil {
			continue
		}
		fd, ok := target.Field(name)
		if !ok {
			return fmt.Errorf("field %q not in object type %q", name, target.Name)
		}
		// Ref fields and ref() templates must pair up, with matching targets:
		// referential integrity is only guaranteed through resolution.
		refTarget, isRef := RefTarget(fd.Type)
		switch {
		case isRef && pt.kind != "ref":
			return fmt.Errorf("field %q is %s and must use =ref(%s, <field>, $.path)", name, fd.Type, refTarget)
		case !isRef && pt.kind == "ref":
			return fmt.Errorf("field %q is %s, not a ref", name, fd.Type)
		case isRef && pt.refType != refTarget:
			return fmt.Errorf("field %q is %s but template resolves a %q", name, fd.Type, pt.refType)
		}
		if fd.Type == "enum" && pt.kind == "literal" && !slices.Contains(fd.Values, pt.raw) {
			return fmt.Errorf("field %q: %q is not one of %v", name, pt.raw, fd.Values)
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
