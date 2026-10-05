package shell

import (
	"context"
	"encoding/json"
	"net/http"

	"rhea/internal/core"
)

// The language explaining itself (DECISIONS 2026-10-05): GET /api/types
// serves every object type the kernel recognizes, each with its explanation —
// which rules produce it (and from which event types), which views render it,
// what it references, what references it, and how many instances the cache
// holds. A system that learns by explanation must be able to explain its own
// vocabulary. Definitions are not objects, so this is a native surface like
// worklist and rules, not a ViewDef.

type typeRuleRef struct {
	RuleID      string `json:"rule_id"`
	Version     int    `json:"version"`
	Status      string `json:"status"`
	EventType   string `json:"event_type"` // what the rule consumes
	Description string `json:"description,omitempty"`
}

type typeViewRef struct {
	ViewID  string `json:"view_id"`
	Notion  string `json:"notion"`
	Title   string `json:"title"`
	Derived bool   `json:"derived"` // version 0: computed from the type, not stored
}

// typeEdge is one ref field seen as a relation: from_type.field → to_type.
type typeEdge struct {
	FromType string `json:"from_type"`
	Field    string `json:"field"`
	ToType   string `json:"to_type"`
}

type typeExplanation struct {
	core.ObjectType
	ProducedBy   []typeRuleRef `json:"produced_by"`
	Views        []typeViewRef `json:"views"`
	References   []typeEdge    `json:"references"`
	ReferencedBy []typeEdge    `json:"referenced_by"`
	Instances    int           `json:"instances"`
}

// producesType says whether a rule's effect materializes objects of the named
// type. An object effect names its type; a postings effect always
// materializes `posting` lines.
func producesType(spec core.RuleSpec, name string) bool {
	if spec.Effect.Object.Type == name {
		return true
	}
	return spec.Effect.Postings != nil && name == "posting"
}

// viewObjectType pulls the object_type out of a list/detail spec; analysis
// specs carry SQL instead and return "".
func viewObjectType(vd core.ViewDef) string {
	var s struct {
		ObjectType string `json:"object_type"`
	}
	if err := json.Unmarshal(vd.Spec, &s); err != nil {
		return ""
	}
	return s.ObjectType
}

func (s *Server) explainTypes(ctx context.Context) ([]typeExplanation, error) {
	types, err := s.Store.ListObjectTypes(ctx)
	if err != nil {
		return nil, err
	}
	rules, err := s.Store.LatestRules(ctx)
	if err != nil {
		return nil, err
	}
	views, err := s.effectiveViewDefs(ctx)
	if err != nil {
		return nil, err
	}
	counts, err := s.Store.CountObjectsByType(ctx)
	if err != nil {
		return nil, err
	}

	// Every ref field across the vocabulary, once, from both ends.
	var edges []typeEdge
	for _, t := range types {
		for _, f := range t.Fields {
			if target, ok := core.RefTarget(f.Type); ok {
				edges = append(edges, typeEdge{FromType: t.Name, Field: f.Name, ToType: target})
			}
		}
	}

	out := make([]typeExplanation, 0, len(types))
	for _, t := range types {
		ex := typeExplanation{ObjectType: t,
			ProducedBy: []typeRuleRef{}, Views: []typeViewRef{},
			References: []typeEdge{}, ReferencedBy: []typeEdge{}}
		for _, r := range rules {
			if producesType(r.Spec, t.Name) {
				ex.ProducedBy = append(ex.ProducedBy, typeRuleRef{
					RuleID: r.ID, Version: r.Version, Status: r.Status,
					EventType: r.Spec.Match.EventType, Description: r.Description,
				})
			}
		}
		for _, vd := range views {
			if viewObjectType(vd) == t.Name {
				ex.Views = append(ex.Views, typeViewRef{
					ViewID: vd.ID, Notion: vd.Notion, Title: vd.Title,
					Derived: vd.Version == 0,
				})
			}
		}
		for _, e := range edges {
			if e.FromType == t.Name {
				ex.References = append(ex.References, e)
			}
			if e.ToType == t.Name {
				ex.ReferencedBy = append(ex.ReferencedBy, e)
			}
		}
		ex.Instances = counts[t.Name]
		out = append(out, ex)
	}
	return out, nil
}

func (s *Server) handleTypes(w http.ResponseWriter, r *http.Request) {
	ex, err := s.explainTypes(r.Context())
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, ex)
}
