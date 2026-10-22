package exec

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"rhea/internal/core"
	"rhea/internal/store"
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
	// Fingerprint names this plan — the events and the rules of every
	// chain — so an approval can say which dry run it read, and a plan
	// that moved since (another tab approved a rule) is refused.
	Fingerprint string `json:"plan"`
}

// fingerprint is a plan's identity: the chains' events, dates and rules, in
// order, and what each new consequence books — events and rules alone let a
// dry run whose amounts had moved pass for the plan approved (review
// 2026-10-21, B3).
func (p *BackfillPlan) fingerprint() string {
	h := sha256.New()
	for _, c := range p.Chains {
		fmt.Fprintf(h, "%d@%s>%s:", c.EventID, c.BookedAt, c.Forwarded)
		for _, n := range c.nodes {
			fmt.Fprintf(h, "%s/%d;", n.rule.ID, n.rule.Version)
			if n.booked == 0 {
				b, _ := json.Marshal([]any{n.mat, n.amend, n.matched, n.unmatched})
				h.Write(b)
			}
		}
		fmt.Fprint(h, "\n")
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
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
	// The chains are planned in log order, each against the world the
	// chains before it leave: a month's advance reads the advances of the
	// months planned before it, as the live path would have (KK's Helios,
	// 2026-10-18: closed before the rule existed, January to August).
	planned := newPlannedWorld()
	ctx = context.WithValue(ctx, plannedKey{}, planned)
	// Each chain is planned in its own company's view, as the live path
	// books it, and sees the bookings planned before it in that company and
	// the installation's: in the caller's view, an operator's backfill summed
	// every company's balances into each one's, and a company's own could
	// not write a chain rooted in the installation's facts (review
	// 2026-10-21, B3). A backfill run for a company is that company's past.
	_, scoped := store.ScopeOf(ctx)
	acting := store.CompanyOf(ctx)
	for _, ev := range raws {
		if waiting[ev.ID] || core.ReservedEventType(ev.Type) || core.TimeEventType(ev.Type) {
			continue
		}
		if scoped && ev.Company != acting {
			continue
		}
		company := ev.Company
		cctx := scopeFor(ctx, company)
		get := func(id string) (map[string]any, bool, error) {
			if st, ok := planned.stateOf(id, company); ok {
				return st, true, nil
			}
			o, err := x.Store.GetObject(cctx, id)
			if errors.Is(err, store.ErrNotFound) {
				return nil, false, nil
			}
			if err != nil {
				return nil, false, err
			}
			return o.State, true, nil
		}
		lookup := func(typ, field string, value any) ([]string, error) {
			ids, err := x.Store.FindObjectIDsByField(cctx, typ, field, core.Text(value))
			if err != nil {
				return nil, err
			}
			return planned.lookup(ids, typ, field, value, company), nil
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
		nodes, err := x.expandChain(cctx, root, rules, payload, lookup, get, prior, nil)
		var locked *PeriodLockedError
		if errors.As(err, &locked) {
			// Never into a closed period: offered forward, dated on the
			// backfill, still caused by the original fact.
			forwarded = locked.Error()
			root.OccurredAt = date
			nodes, err = x.expandChain(cctx, root, rules, payload, lookup, get, prior, nil)
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
		planned.add(c.nodes, get, company)
		c.nodes = nodes // booking needs the replayed nodes too, as causes
		plan.Chains = append(plan.Chains, c)
	}
	plan.Diff, err = x.backfillDiff(ctx, plan.Chains)
	plan.Fingerprint = plan.fingerprint()
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
		switch d.Type {
		case core.EventObjectMaterialized:
			var m core.MaterializedObject
			if err := json.Unmarshal(d.Payload, &m); err != nil {
				return nil, fmt.Errorf("derived event %d: %w", d.ID, err)
			}
			p.mat = &m
			objectOf[d.ID] = m.ObjectID
		case core.EventObjectsMatched:
			var m core.MatchedObjects
			if err := json.Unmarshal(d.Payload, &m); err != nil {
				return nil, fmt.Errorf("derived event %d: %w", d.ID, err)
			}
			p.matched = &m
			objectOf[d.ID] = matchChainPoint(&m)
		case core.EventObjectsUnmatched:
			var u core.UnmatchedObjects
			if err := json.Unmarshal(d.Payload, &u); err != nil {
				return nil, fmt.Errorf("derived event %d: %w", d.ID, err)
			}
			p.unmatched = &u
			objectOf[d.ID] = MatchIDBase(u.Matching, u.Trigger.ObjectID)
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
			if n.booked != 0 || n.matched != nil || n.unmatched != nil {
				continue
			}
			if n.amend == nil {
				added[n.mat.ObjectID] = len(diff.Added)
				// The diff shows the chain's end state, so an in-chain
				// amendment merges into this copy — never into the
				// materialization bookTx writes under its own rule's name.
				state := make(map[string]any, len(n.mat.State))
				for k, v := range n.mat.State {
					state[k] = v
				}
				diff.Added = append(diff.Added, core.Object{ID: n.mat.ObjectID, Type: n.mat.ObjectType,
					TypeVersion: n.mat.TypeVersion, State: state, SourceEventID: c.EventID,
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

// ApproveBackfill promotes the plan through the approve_backfill door: the
// approval is a raw event naming every chain (invariant 2), then the chains
// book. date is the backfill's business date — where forwarded chains land.
func (x *Executor) ApproveBackfill(ctx context.Context, approvedBy, date string) (int, []error, error) {
	return x.ApproveBackfillPlan(ctx, approvedBy, date, "")
}

// ApproveBackfillPlan approves the plan a person read: plan is the dry
// run's fingerprint, and a past that moved since is refused; empty skips
// the check (the CLI's own approval).
func (x *Executor) ApproveBackfillPlan(ctx context.Context, approvedBy, date, plan string) (int, []error, error) {
	inputs := map[string]any{"approved_by": approvedBy, "occurred_at": date}
	if plan != "" {
		inputs["plan"] = plan
	}
	t, err := x.TriggerActivity(ctx, "approve_backfill", inputs, approvedBy, "")
	if err != nil {
		return 0, nil, err
	}
	return t.Backfilled, t.Errors, nil
}

// plannedWorld is what the chains planned so far would leave: the objects
// they materialize and the states they amend, read before the store's, and
// their postings, which the ledger's balance read adds to the booked ones.
// Each planned object is its chain's company's: a chain sees its own
// company's and the installation's, as the store shows it.
type plannedWorld struct {
	state   map[string]map[string]any
	typ     map[string]string
	company map[string]string
	order   []string
}

type plannedKey struct{}

func newPlannedWorld() *plannedWorld {
	return &plannedWorld{state: map[string]map[string]any{}, typ: map[string]string{}, company: map[string]string{}}
}

// visible reports whether a chain of company sees a planned object.
func (w *plannedWorld) visible(id, company string) bool {
	c := w.company[id]
	return c == "" || c == company
}

// stateOf is a planned object's state as a chain of company sees it.
func (w *plannedWorld) stateOf(id, company string) (map[string]any, bool) {
	st, ok := w.state[id]
	if !ok || !w.visible(id, company) {
		return nil, false
	}
	return st, true
}

// add takes in a planned chain's new consequences, as its company's.
func (w *plannedWorld) add(nodes []chainNode, get core.Getter, company string) {
	for _, n := range nodes {
		if n.booked != 0 {
			continue
		}
		switch {
		case n.mat.ObjectID != "":
			w.put(n.mat.ObjectID, n.mat.ObjectType, n.mat.State, company)
		case n.amend != nil:
			base, ok, _ := get(n.amend.ObjectID)
			st := map[string]any{}
			if ok {
				for k, v := range base {
					st[k] = v
				}
			}
			for k, v := range n.amend.Set {
				st[k] = v
			}
			w.put(n.amend.ObjectID, n.amend.ObjectType, st, company)
		}
	}
}

func (w *plannedWorld) put(id, typ string, st map[string]any, company string) {
	if _, seen := w.state[id]; !seen {
		w.order = append(w.order, id)
		w.company[id] = company
	}
	w.state[id], w.typ[id] = st, typ
}

// lookup is the store's answer as the planned chains change it: an object
// whose planned state no longer matches drops out, a planned one that
// matches joins, after the booked ones — where it will stand once booked.
func (w *plannedWorld) lookup(stored []string, typ, field string, value any, company string) []string {
	want := core.Text(value)
	out := make([]string, 0, len(stored))
	in := map[string]bool{}
	for _, id := range stored {
		in[id] = true
		if st, ok := w.stateOf(id, company); ok && core.Text(st[field]) != want {
			continue
		}
		out = append(out, id)
	}
	for _, id := range w.order {
		if !in[id] && w.visible(id, company) && w.typ[id] == typ && core.Text(w.state[id][field]) == want {
			out = append(out, id)
		}
	}
	return out
}

// balance is the planned postings' part of an account's balance, as a
// chain of company sees them.
func (w *plannedWorld) balance(book, account, from, to, company string) int64 {
	var n int64
	for _, id := range w.order {
		st := w.state[id]
		if w.typ[id] != PostingObjectType || core.Text(st["account"]) != account || !w.visible(id, company) {
			continue
		}
		b := core.Text(st["book"])
		if b == "" {
			b = "main"
		}
		d := core.Text(st["date"])
		if b != book || d < from || d > to {
			continue
		}
		amt, err := strconv.ParseInt(core.Text(st["amount"]), 10, 64)
		if err != nil {
			continue
		}
		if st["side"] == "debit" {
			n += amt
		} else {
			n -= amt
		}
	}
	return n
}
