// Package core defines the language of ERP: events, object types, objects,
// rules and view definitions. Everything here is data — storable, versioned,
// loadable into a running kernel. See SPEC.md §2.
package core

import (
	"encoding/json"
	"fmt"
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
	Fields     []FieldDef `json:"fields"`
}

type FieldDef struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // string | int | date | money
	Required bool   `json:"required"`
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
		if _, err := ParseTemplate(tmpl); err != nil {
			return fmt.Errorf("field %q template: %w", name, err)
		}
		if target != nil {
			if _, ok := target.Field(name); !ok {
				return fmt.Errorf("field %q not in object type %q", name, target.Name)
			}
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
