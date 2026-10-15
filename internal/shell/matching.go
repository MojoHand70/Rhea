package shell

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"rhea/internal/core"
)

// The matching notion (DIRECTION 2026-10-11, "the remainder is the grid"):
// the N-sided view of what no strategy settled — every side's open items
// side by side, each trigger item carrying why the policy left it — with
// the pairing verb. The view names a matching rule; the rule's structure
// says what the sides are and which fields matter; the view adds no opinion
// of its own. Reconciliation, bank matching and intercompany matching are
// this one notion.

type matchingItem struct {
	ObjectID string             `json:"object_id"`
	Label    string             `json:"label"`
	Detail   string             `json:"detail,omitempty"`
	Cells    map[string]cellOut `json:"cells"`
	// Why is what the policy tried and found wanting, for a trigger item
	// still open: the latest objects.unmatched for it, step by step.
	Why []string `json:"why,omitempty"`
}

type matchingSide struct {
	Name    string         `json:"name"`
	Type    string         `json:"type"`
	Trigger bool           `json:"trigger"`
	Columns []columnOut    `json:"columns"`
	Items   []matchingItem `json:"items"`
}

func (s *Server) serveMatching(ctx context.Context, w http.ResponseWriter, vd *core.ViewDef) {
	data, err := s.matchingData(ctx, vd)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, data)
}

// matchingData computes a matching view: the sides and their open items.
func (s *Server) matchingData(ctx context.Context, vd *core.ViewDef) (map[string]any, error) {
	var spec core.MatchingSpec
	if err := json.Unmarshal(vd.Spec, &spec); err != nil {
		return nil, err
	}
	rules, err := s.Store.ActiveRules(ctx)
	if err != nil {
		return nil, err
	}
	var rule *core.Rule
	for i := range rules {
		if rules[i].ID == spec.Rule && rules[i].Spec.Effect.Matching != nil {
			rule = &rules[i]
		}
	}
	if rule == nil {
		return nil, fmt.Errorf("view %s: no active matching rule %q", vd.ID, spec.Rule)
	}
	tmpl := rule.Spec.Effect.Matching
	vds, err := s.effectiveViewDefs(ctx)
	if err != nil {
		return nil, err
	}
	detailFor := detailViewByType(vds)
	why, err := s.unmatchedReasons(ctx, tmpl.Name)
	if err != nil {
		return nil, err
	}
	var sides []matchingSide
	for _, side := range tmpl.Sides {
		ot, err := s.Store.GetObjectType(ctx, side.Type)
		if err != nil {
			return nil, err
		}
		objs, err := s.Store.ObjectsByType(ctx, side.Type)
		if err != nil {
			return nil, err
		}
		// The columns the structure names, in a fixed order: the label,
		// the reference, the date, what is open, its currency, what is
		// compared.
		var fields []string
		add := func(f string) {
			if f == "" {
				return
			}
			for _, have := range fields {
				if have == f {
					return
				}
			}
			fields = append(fields, f)
		}
		add(ot.LabelField)
		add(side.Reference)
		add(side.Date)
		for _, m := range tmpl.AllocatedMeasures() {
			add(side.Open[m])
		}
		add(side.Currency)
		for _, m := range tmpl.ComparedMeasures() {
			add(side.Compare[m])
		}
		out := matchingSide{Name: side.Name, Type: side.Type, Trigger: side.Name == tmpl.Trigger, Items: []matchingItem{}}
		for _, f := range fields {
			fd, _ := ot.Field(f)
			out.Columns = append(out.Columns, columnOut{Field: f, Label: humanizeField(f), Type: fd.Type})
		}
		for _, o := range objs {
			open := true
			for _, m := range tmpl.AllocatedMeasures() {
				if !positive(o.State[side.Open[m]]) {
					open = false
				}
			}
			if !open {
				continue
			}
			if out.Trigger && !triggersRule(rule, o) {
				// The trigger side lists what the rule would fire on: an open
				// object the rule's where excludes (a credit on the outgoing
				// grid) is another rule's to pair, not this grid's.
				continue
			}
			item := matchingItem{ObjectID: o.ID, Label: o.ID, Detail: detailFor[side.Type], Cells: map[string]cellOut{}}
			if ot.LabelField != "" {
				if l, ok := o.State[ot.LabelField].(string); ok && l != "" {
					item.Label = l
				}
			}
			for _, f := range fields {
				fd, _ := ot.Field(f)
				item.Cells[f] = s.typedCell(ctx, o.State[f], fd, detailFor)
			}
			if out.Trigger {
				item.Why = why[o.ID]
			}
			out.Items = append(out.Items, item)
		}
		sides = append(sides, out)
	}
	return map[string]any{
		"view": vd, "rule": rule.ID, "matching": tmpl.Name, "trigger": tmpl.Trigger,
		"verb": "pair_items", "sides": sides,
	}, nil
}

// triggersRule reports whether an object, materialized, would meet the
// rule's where — the same conditions the executor applies to the
// object.materialized event, evaluated on the same payload shape.
func triggersRule(rule *core.Rule, o core.Object) bool {
	payload := map[string]any{"object_id": o.ID, "object_type": o.Type, "state": o.State}
	for _, c := range rule.Spec.Match.Where {
		ok, err := core.EvalCondition(payload, c)
		if err != nil || !ok {
			return false
		}
	}
	return true
}

// unmatchedReasons indexes the latest objects.unmatched of one structure by
// trigger: the explanation the grid shows beside an item still open.
func (s *Server) unmatchedReasons(ctx context.Context, matching string) (map[string][]string, error) {
	derived, err := s.Store.EventsByKind(ctx, core.KindDerived)
	if err != nil {
		return nil, err
	}
	out := map[string][]string{}
	for _, ev := range derived {
		if ev.Type != core.EventObjectsUnmatched {
			continue
		}
		var u core.UnmatchedObjects
		if json.Unmarshal(ev.Payload, &u) != nil || u.Matching != matching {
			continue
		}
		var reasons []string
		if u.Reason != "" {
			reasons = append(reasons, u.Reason)
		}
		for _, a := range u.Tried {
			reasons = append(reasons, "step "+strconv.Itoa(a.Step)+" "+a.Strategy+": "+a.Reason)
		}
		out[u.Trigger.ObjectID] = reasons // log order: the latest wins
	}
	return out, nil
}

// positive reads an open amount as state holds it: a number, or a decimal
// string at the boundary.
func positive(v any) bool {
	switch n := v.(type) {
	case float64:
		return n > 0
	case int64:
		return n > 0
	case int:
		return n > 0
	case string:
		r, err := core.ParseDecimal(n)
		return err == nil && r.Sign() > 0
	}
	return false
}

func humanizeField(f string) string {
	if f == "" {
		return f
	}
	out := []byte(f)
	for i := range out {
		if out[i] == '_' {
			out[i] = ' '
		}
	}
	if out[0] >= 'a' && out[0] <= 'z' {
		out[0] -= 'a' - 'A'
	}
	return string(out)
}
