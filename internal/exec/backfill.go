package exec

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"rhea/internal/core"
)

// Ruled backfill (DIRECTION, late understanding; adopted 2026-10-08):
// understanding that arrives after the facts. An explained event is never
// revisited by the live path, so a rule approved today cannot see September.
// Backfill is the deliberate, gated act that lets it: every past event is
// re-expanded under the active rules with its existing derivations replayed
// from the log, and only the consequences that never happened are new —
// additive pairs, never a rewrite. Each chain books atomically, caused by its
// original event, dated by it while its period is open; a chain that would
// post into a closed (book, month) is offered forward, dated on the backfill
// and linked to its cause — the korekta. Time events stay out: a late rule on
// time is the clock's catch-up, behind its own gate.

// BackfillChain is one past event's missing consequences.
type BackfillChain struct {
	EventID    int64    `json:"event_id"`
	EventType  string   `json:"event_type"`
	OccurredAt string   `json:"occurred_at"` // the fact's business date
	BookedAt   string   `json:"booked_at"`   // the date its consequences carry
	Forwarded  string   `json:"forwarded,omitempty"`
	Rules      []string `json:"rules"`
	root       core.Event
	nodes      []chainNode
}

// BackfillPlan is what a backfill would do, computed read-only: the chains,
// the diff a human reads before approving, and the events whose missing
// consequences fail to expand (they stay as they are, reported).
type BackfillPlan struct {
	Date   string          `json:"date"`
	Chains []BackfillChain `json:"chains"`
	Diff   SimDiff         `json:"diff"`
	Errors []string        `json:"errors"`
}

// PlanBackfill computes the backfill against the current log and rules.
// date is the backfill's own business date: where a chain bound for a closed
// period is offered.
func (x *Executor) PlanBackfill(ctx context.Context, date string) (BackfillPlan, error) {
	plan := BackfillPlan{Date: date}
	rules, err := x.Store.ActiveRules(ctx)
	if err != nil {
		return plan, err
	}
	pending, err := x.Store.PendingEvents(ctx)
	if err != nil {
		return plan, err
	}
	waiting := map[int64]bool{}
	for _, e := range pending {
		waiting[e.ID] = true // unexplained events are the worklist's, not the past's
	}
	raws, err := x.Store.EventsByKind(ctx, core.KindRaw)
	if err != nil {
		return plan, err
	}
	lookup := func(typ, field string, value any) ([]string, error) {
		return x.Store.FindObjectIDsByField(ctx, typ, field, fmt.Sprintf("%v", value))
	}
	get := func(id string) (map[string]any, bool, error) {
		o, err := x.Store.GetObject(ctx, id)
		if err != nil {
			return nil, false, nil
		}
		return o.State, true, nil
	}
	for _, ev := range raws {
		if waiting[ev.ID] || core.ReservedEventType(ev.Type) || core.TimeEventType(ev.Type) {
			continue
		}
		chain, err := x.Store.ChainOf(ctx, ev.ID)
		if err != nil {
			return plan, err
		}
		if len(chain) == 0 {
			continue
		}
		prior, err := priorOf(ev, chain)
		if err != nil {
			return plan, err
		}
		var payload any
		if err := json.Unmarshal(ev.Payload, &payload); err != nil {
			plan.Errors = append(plan.Errors, fmt.Sprintf("event %d payload: %v", ev.ID, err))
			continue
		}
		root, forwarded := ev, ""
		nodes, err := x.expandChain(ctx, root, rules, payload, lookup, get, prior)
		var locked *PeriodLockedError
		if errors.As(err, &locked) {
			// Never into a closed period: offered forward, dated on the
			// backfill, still caused by the original fact.
			forwarded = locked.Error()
			root.OccurredAt = date
			nodes, err = x.expandChain(ctx, root, rules, payload, lookup, get, prior)
		}
		if err != nil {
			plan.Errors = append(plan.Errors, fmt.Sprintf("event %d: %v", ev.ID, err))
			continue
		}
		c := BackfillChain{EventID: ev.ID, EventType: ev.Type, OccurredAt: ev.OccurredAt,
			BookedAt: root.OccurredAt, Forwarded: forwarded, root: root}
		seen := map[string]bool{}
		for _, n := range nodes {
			if n.booked != 0 {
				continue
			}
			c.nodes = append(c.nodes, n)
			if !seen[n.rule.ID] {
				seen[n.rule.ID] = true
				c.Rules = append(c.Rules, n.rule.ID)
			}
		}
		if len(c.nodes) == 0 {
			continue
		}
		c.nodes = nodes // booking needs the replayed nodes too, as causes
		plan.Chains = append(plan.Chains, c)
	}
	plan.Diff, err = x.backfillDiff(ctx, plan.Chains)
	return plan, err
}

// priorOf indexes a root's existing derivations by the chain point they fired
// at: the root's id for root firings, the causing object's id for cascades.
func priorOf(root core.Event, chain []core.Event) (Prior, error) {
	objectOf := map[int64]string{} // derived event id → the object it materialized
	prior := Prior{}
	for _, d := range chain {
		key := fmt.Sprintf("%d", root.ID)
		if d.CauseEventID != nil && *d.CauseEventID != root.ID {
			key = objectOf[*d.CauseEventID]
		}
		p := priorFiring{ruleID: d.RuleID, version: d.RuleVersion, eventID: d.ID}
		if d.Type == core.EventObjectMaterialized {
			var m core.MaterializedObject
			if err := json.Unmarshal(d.Payload, &m); err != nil {
				return nil, fmt.Errorf("derived event %d: %w", d.ID, err)
			}
			p.mat = &m
			objectOf[d.ID] = m.ObjectID
		}
		prior[key] = append(prior[key], p)
	}
	return prior, nil
}

// backfillDiff shapes the plan for the reader: new objects as added, moved
// objects as changed — the same diff every approval in Rhea renders.
func (x *Executor) backfillDiff(ctx context.Context, chains []BackfillChain) (SimDiff, error) {
	diff := SimDiff{RuleID: "backfill"}
	added := map[string]int{}
	for _, c := range chains {
		for _, n := range c.nodes {
			if n.booked != 0 {
				continue
			}
			if n.amend == nil {
				added[n.mat.ObjectID] = len(diff.Added)
				diff.Added = append(diff.Added, core.Object{ID: n.mat.ObjectID, Type: n.mat.ObjectType,
					TypeVersion: n.mat.TypeVersion, State: n.mat.State, SourceEventID: c.EventID,
					RuleID: n.rule.ID, RuleVersion: n.rule.Version})
				continue
			}
			if i, ok := added[n.amend.ObjectID]; ok { // moved within its own new chain
				for k, v := range n.amend.Set {
					diff.Added[i].State[k] = v
				}
				continue
			}
			before, err := x.Store.GetObject(ctx, n.amend.ObjectID)
			if err != nil {
				return diff, err
			}
			after := before
			after.State = make(map[string]any, len(before.State)+len(n.amend.Set))
			for k, v := range before.State {
				after.State[k] = v
			}
			for k, v := range n.amend.Set {
				after.State[k] = v
			}
			diff.Changed = append(diff.Changed, SimChange{Before: before, After: after})
		}
	}
	return diff, nil
}

// executeBackfill books an approved plan, chain by chain, each atomic.
func (x *Executor) executeBackfill(ctx context.Context, plan *BackfillPlan) (int, []error) {
	var n int
	var errs []error
	for _, c := range plan.Chains {
		if err := x.book(ctx, c.root, c.nodes); err != nil {
			errs = append(errs, fmt.Errorf("backfill event %d: %w", c.EventID, err))
			continue
		}
		n++
	}
	return n, errs
}

// ApproveBackfill promotes the plan through the approve_backfill door: the
// approval is a raw event naming every chain (invariant 2), then the chains
// book. date is the backfill's business date — where forwarded chains land.
func (x *Executor) ApproveBackfill(ctx context.Context, approvedBy, date string) (int, []error, error) {
	t, err := x.TriggerActivity(ctx, "approve_backfill", map[string]any{
		"approved_by": approvedBy, "occurred_at": date,
	}, approvedBy, "")
	if err != nil {
		return 0, nil, err
	}
	return t.Backfilled, t.Errors, nil
}
