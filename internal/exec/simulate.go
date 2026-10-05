// Simulation: SPEC M1's dry run. A draft rule is replayed against the whole
// historical log entirely in memory — same match, same expansion, same ref
// resolution as live firing, but nothing is written and ref() resolves
// against the simulated world. The approval moment then shows a diff of what
// the rule would change instead of hoping the JSON looks right.
package exec

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"rhea/internal/core"
)

// SimRun is one in-memory replay of every raw business event through a rule set.
type SimRun struct {
	Objects     []core.Object `json:"objects"`     // in materialization order
	Unexplained []int64       `json:"unexplained"` // business event ids no rule fired on
	Errors      []string      `json:"errors"`      // failures from the final pass
}

// SimChange pairs the same object id under the two rule sets.
type SimChange struct {
	Before core.Object `json:"before"`
	After  core.Object `json:"after"`
}

// SimDiff is what approving the rule would change, by replaying history under
// the active rules with and without it.
type SimDiff struct {
	RuleID            string        `json:"rule_id"`
	Version           int           `json:"version"`
	Added             []core.Object `json:"added"`
	Changed           []SimChange   `json:"changed"`
	Removed           []core.Object `json:"removed"`
	UnexplainedBefore []int64       `json:"unexplained_before"`
	UnexplainedAfter  []int64       `json:"unexplained_after"`
	Errors            []string      `json:"errors"`
}

// SimulateRule dry-runs the latest version of a rule (typically a draft) as
// if it were active, against the full event log. Read-only.
func (x *Executor) SimulateRule(ctx context.Context, ruleID string) (SimDiff, error) {
	r, err := x.Store.GetRule(ctx, ruleID)
	if err != nil {
		return SimDiff{}, err
	}
	active, err := x.Store.ActiveRules(ctx)
	if err != nil {
		return SimDiff{}, err
	}
	// Candidate set: the active rules with this rule's latest version in play
	// (replacing any active version of itself), in firing order.
	cand := make([]core.Rule, 0, len(active)+1)
	for _, a := range active {
		if a.ID != r.ID {
			cand = append(cand, a)
		}
	}
	cand = append(cand, r)
	sort.Slice(cand, func(i, j int) bool {
		if cand[i].Priority != cand[j].Priority {
			return cand[i].Priority < cand[j].Priority
		}
		return cand[i].ID < cand[j].ID
	})

	base, err := x.simulate(ctx, active)
	if err != nil {
		return SimDiff{}, err
	}
	after, err := x.simulate(ctx, cand)
	if err != nil {
		return SimDiff{}, err
	}

	diff := SimDiff{
		RuleID: r.ID, Version: r.Version,
		UnexplainedBefore: base.Unexplained, UnexplainedAfter: after.Unexplained,
		Errors: after.Errors,
	}
	baseByID := map[string]core.Object{}
	for _, o := range base.Objects {
		baseByID[o.ID] = o
	}
	afterIDs := map[string]bool{}
	for _, o := range after.Objects {
		afterIDs[o.ID] = true
		b, existed := baseByID[o.ID]
		if !existed {
			diff.Added = append(diff.Added, o)
			continue
		}
		bj, _ := json.Marshal(b)
		aj, _ := json.Marshal(o)
		if string(bj) != string(aj) {
			diff.Changed = append(diff.Changed, SimChange{Before: b, After: o})
		}
	}
	for _, o := range base.Objects {
		if !afterIDs[o.ID] {
			diff.Removed = append(diff.Removed, o)
		}
	}
	return diff, nil
}

// simulate replays every raw business event through the given rules, in
// event order with the same fixpoint behavior as ProcessPending, building
// objects in memory only. Already-explained events are replayed too: the
// result is the complete would-be world under this rule set.
func (x *Executor) simulate(ctx context.Context, rules []core.Rule) (SimRun, error) {
	events, err := x.Store.EventsByKind(ctx, core.KindRaw)
	if err != nil {
		return SimRun{}, err
	}

	objects := map[string]core.Object{}
	var order []string
	lookup := func(typ, field string, value any) ([]string, error) {
		want := fmt.Sprintf("%v", value)
		var ids []string
		for _, id := range order {
			o := objects[id]
			if o.Type == typ && fmt.Sprintf("%v", o.State[field]) == want {
				ids = append(ids, id)
			}
		}
		return ids, nil
	}

	get := func(id string) (map[string]any, bool, error) {
		o, ok := objects[id]
		if !ok {
			return nil, false, nil
		}
		return o.State, true, nil
	}

	explained := map[int64]bool{}
	var errs []string
	for {
		pass := 0
		errs = errs[:0]
		for _, ev := range events {
			// System verbs (rule.*, activity.*) are not business events: the
			// kernel speaks them, no rule explains them — same as the worklist.
			if explained[ev.ID] || core.ReservedEventType(ev.Type) {
				continue
			}
			var payload any
			if err := json.Unmarshal(ev.Payload, &payload); err != nil {
				errs = append(errs, fmt.Sprintf("event %d payload: %v", ev.ID, err))
				continue
			}
			// The exact chain expansion live firing runs — cascade included —
			// with refs resolving against the simulated world.
			nodes, err := x.expandChain(ctx, ev, rules, payload, lookup, get)
			if err != nil {
				errs = append(errs, fmt.Sprintf("event %d: %v", ev.ID, err))
				continue
			}
			if len(nodes) == 0 {
				continue
			}
			for _, n := range nodes {
				if n.amend != nil {
					// The simulated world amends in memory the way the live
					// path amends the cache; the object exists — expansion
					// read it through the same overlay.
					o := objects[n.amend.ObjectID]
					merged := make(map[string]any, len(o.State)+len(n.amend.Set))
					for k, v := range o.State {
						merged[k] = v
					}
					for k, v := range n.amend.Set {
						merged[k] = v
					}
					o.State = merged
					objects[o.ID] = o
					continue
				}
				// Cascaded objects are attributed to the root event here: the
				// intermediate derived events only get ids when really booked.
				o := core.Object{
					ID: n.mat.ObjectID, Type: n.mat.ObjectType, TypeVersion: n.mat.TypeVersion,
					State: n.mat.State, SourceEventID: ev.ID, RuleID: n.rule.ID, RuleVersion: n.rule.Version,
				}
				if _, seen := objects[o.ID]; !seen {
					order = append(order, o.ID)
				}
				objects[o.ID] = o
			}
			explained[ev.ID] = true
			pass++
		}
		if pass == 0 {
			break
		}
	}

	run := SimRun{Errors: errs}
	for _, id := range order {
		run.Objects = append(run.Objects, objects[id])
	}
	for _, ev := range events {
		if !explained[ev.ID] && !core.ReservedEventType(ev.Type) {
			run.Unexplained = append(run.Unexplained, ev.ID)
		}
	}
	return run, nil
}
