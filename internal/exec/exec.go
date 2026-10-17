// Package exec is the deterministic kernel: it evaluates active rules against
// events, emits derived events with full provenance, and maintains the object
// projection cache. It is the only writer of state (SPEC §3).
package exec

import (
	"errors"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"rhea/internal/core"
	"rhea/internal/store"
)

type Executor struct {
	Store *store.Store
	// MaxCollection caps a formula's linked collection read (objects());
	// zero means core.DefaultMaxCollection. Tests lower it to see the
	// refusal; nothing raises it per rule — the cap is the kernel's.
	MaxCollection int
	// draftTypes overlays a bundle's draft object types for its dry run:
	// set only on a simulation's private copy of the executor, never on the
	// one that books.
	draftTypes map[string]core.ObjectType
}

// objectType is the type the executor validates against: the newest active
// version, or a bundle draft while that bundle is being dry-run.
func (x *Executor) objectType(ctx context.Context, name string) (core.ObjectType, error) {
	if t, ok := x.draftTypes[name]; ok {
		return t, nil
	}
	return x.Store.GetObjectType(ctx, name)
}

// ProcessPending evaluates every unmatched raw event against the active rule
// set, in event order, and repeats until a pass books nothing: a firing can
// materialize an object (a company, say) that an earlier event's ref() was
// waiting for. Returns how many objects were materialized. An event that
// matches no rule simply stays in the worklist; an event whose matching rule
// fails to expand is reported (from the final pass) and left in place.
func (x *Executor) ProcessPending(ctx context.Context) (int, []error) {
	rules, err := x.Store.ActiveRules(ctx)
	if err != nil {
		return 0, []error{err}
	}
	var booked int
	for {
		events, err := x.Store.PendingEvents(ctx)
		if err != nil {
			return booked, []error{err}
		}
		pass := 0
		var errs []error
		for _, ev := range events {
			fired, err := x.evaluate(ctx, ev, rules)
			if err != nil {
				errs = append(errs, fmt.Errorf("event %d: %w", ev.ID, err))
				continue
			}
			if fired {
				pass++
			}
		}
		booked += pass
		if pass == 0 {
			return booked, errs
		}
	}
}

// chainNode is one would-be consequence in a cascade chain: a
// materialization, an amendment, or a match (found or not), the rule that
// produced it, and which chain entry's derived event caused it (-1 is the
// root raw event). Nodes book in chain order. Amendments end their branch:
// nothing fires on them. A match is a chain point like a materialization:
// the rules that explain it (amendments of open amounts, settlements,
// differences) fire on its event in the next generation.
type chainNode struct {
	mat       core.MaterializedObject
	amend     *core.AmendedObject
	matched   *core.MatchedObjects
	unmatched *core.UnmatchedObjects
	rule      core.Rule
	cause     int
	// booked is the derived event already in the log for this node: a
	// prior firing replayed into a backfill chain, never written again.
	booked int64
}

// priorFiring is one consequence already in the log, keyed in Prior by the
// point in a chain it fired at: the root event's id, or the causing object's.
type priorFiring struct {
	ruleID    string
	version   int
	eventID   int64
	mat       *core.MaterializedObject // nil for an amendment or a match
	matched   *core.MatchedObjects
	unmatched *core.UnmatchedObjects
}

// Prior is a root event's existing derivations, by chain point. A backfill
// expansion replays them from the log, baked state and all, and expands only
// the rules that never fired at that point; live firing passes nil.
type Prior map[string][]priorFiring

// objectID names the object this node touches — the unit the same-id
// conflict guard speaks about — or, for a match, its chain point.
func (n chainNode) objectID() string {
	if n.amend != nil {
		return n.amend.ObjectID
	}
	if n.matched != nil || n.unmatched != nil {
		return n.idBase()
	}
	return n.mat.ObjectID
}

func (n chainNode) objectType() string {
	if n.amend != nil {
		return n.amend.ObjectType
	}
	return n.mat.ObjectType
}

// idBase is the chain point cascaded firings build their ids on: the
// materialized object's id, or for a match the structure's name qualified
// by the trigger — rooted, like every id, in the root event through the
// trigger's own id.
func (n chainNode) idBase() string {
	switch {
	case n.matched != nil:
		return matchChainPoint(n.matched)
	case n.unmatched != nil:
		return MatchIDBase(n.unmatched.Matching, n.unmatched.Trigger.ObjectID)
	}
	return n.mat.ObjectID
}

// MatchIDBase is the chain point of one match: ids of the objects a match
// causes (settlements, differences) build on it, so one structure firing on
// one trigger mints them once and reproducibly. A reversal is its own
// chain point, named by the match it undoes, so what it causes never
// collides with what the match caused.
func MatchIDBase(matching, triggerID string) string { return matching + "-" + triggerID }

// matchChainPoint is a match event's chain point: the structure and the
// trigger, then the match's sequence on that trigger when it is not the
// first, then the event a reversal undoes.
func matchChainPoint(m *core.MatchedObjects) string {
	base := MatchIDBase(m.Matching, m.Trigger.ObjectID)
	if m.Seq > 0 {
		base = fmt.Sprintf("%s-%d", base, m.Seq)
	}
	if m.ReversalOf != 0 {
		return fmt.Sprintf("%s-undo-%d", base, m.ReversalOf)
	}
	return base
}

// derivedType is the event a node emits for the next generation.
func (n chainNode) derivedType() string {
	switch {
	case n.matched != nil:
		return core.EventObjectsMatched
	case n.unmatched != nil:
		return core.EventObjectsUnmatched
	}
	return core.EventObjectMaterialized
}

// derivedPayload is what the next generation's rules see: the
// materialization or the match as it will be written to the log, plus,
// under "root", the fact that started the chain (read-only, never written).
func (n chainNode) derivedPayload(root any) (map[string]any, error) {
	var evPayload map[string]any
	switch {
	case n.matched != nil || n.unmatched != nil:
		var v any = n.matched
		if n.unmatched != nil {
			v = n.unmatched
		}
		// The event's own encoding, decoded as every payload is: what a
		// rule reads is exactly what the log will hold.
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(b, &evPayload); err != nil {
			return nil, err
		}
	default:
		evPayload = map[string]any{
			"object_id":    n.mat.ObjectID,
			"object_type":  n.mat.ObjectType,
			"type_version": n.mat.TypeVersion,
			"state":        n.mat.State,
		}
	}
	evPayload["root"] = root
	return evPayload, nil
}

// maxCascadeDepth caps chain generations — the loop guard for cascaded rules,
// since cause-qualified identity mints a fresh id each generation and a cycle
// never self-collides. Its error names the looping rule and the id it keeps
// materializing; the id itself shows the loop, one type name per generation.
const maxCascadeDepth = 16

// evaluate fires every matching rule on the event and cascades: each firing's
// derived event is itself evaluated against the rule set (receipt → stock
// movement → valuation posting), and the whole chain books atomically — all
// of it or none, so an event is never half-explained. Two rules materializing
// the same object id anywhere in the chain is a conflict: the event is
// refused whole and waits for a human.
func (x *Executor) evaluate(ctx context.Context, ev core.Event, rules []core.Rule) (bool, error) {
	var payload any
	if err := json.Unmarshal(ev.Payload, &payload); err != nil {
		return false, fmt.Errorf("payload: %w", err)
	}
	// Lookups resolve against the object cache as of this point in the log;
	// resolved ids are baked into the derived events, so replay never
	// re-resolves and determinism (invariant 4) is untouched.
	lookup := func(typ, field string, value any) ([]string, error) {
		return x.Store.FindObjectIDsByField(ctx, typ, field, fmt.Sprintf("%v", value))
	}
	get := func(id string) (map[string]any, bool, error) {
		o, err := x.Store.GetObject(ctx, id)
		if err != nil {
			return nil, false, nil // not found; a lookup already vouched for real ids
		}
		return o.State, true, nil
	}
	nodes, err := x.expandChain(ctx, ev, rules, payload, lookup, get, nil, nil)
	var locked *PeriodLockedError
	if errors.As(err, &locked) {
		// A late document is never refused (UoR art. 20 ust. 1, art. 24
		// ust. 3): its entries land in the first open month of the same
		// year, while the fact keeps its own date in the log — the chain is
		// booked as backfill books a chain bound for a closed period.
		open, ferr := firstOpenMonth(lookup, locked.Book, locked.Month)
		if ferr != nil {
			return false, ferr
		}
		if open == "" {
			return false, fmt.Errorf("%w, and so is every later month of its year — a late document does not cross into the next year by itself", err)
		}
		forwarded := ev
		forwarded.OccurredAt = open + "-01"
		if nodes, err = x.expandChain(ctx, forwarded, rules, payload, lookup, get, nil, nil); err == nil && len(nodes) > 0 {
			return true, x.book(ctx, forwarded, nodes)
		}
	}
	if err != nil {
		return false, err
	}
	if len(nodes) == 0 {
		return false, nil
	}
	return true, x.book(ctx, ev, nodes)
}

// expandChain evaluates the rule set against the root event, then against the
// derived events its firings would emit, generation by generation until no
// rule fires. Identity needs no sequence state anywhere: a root firing's ids
// derive from the raw event (<type>-<root id>), a cascaded firing's from the
// causing object's id (<type>-<cause object id>), which is itself rooted — so
// simulation reproduces every cascaded id exactly, and one rule firing on two
// sibling derived events (two line movements of one receipt) mints distinct
// ids instead of colliding. Expansions see state as of before the root event
// plus the chain's earlier generations — the receipt's movement is visible to
// the valuation rule — never their own siblings. Shared verbatim by the live
// path and the simulator, so a dry run cannot drift from reality.
func (x *Executor) expandChain(ctx context.Context, root core.Event, rules []core.Rule, payload any, base core.Lookup, baseGet core.Getter, prior Prior, seed *chainNode) ([]chainNode, error) {
	var nodes []chainNode
	// A seed is a first node the root already carries — a person's
	// pairing, run by the kernel before the rules — from which the
	// generations cascade as from any firing.
	if seed != nil {
		nodes = append(nodes, *seed)
	}
	owner := map[string]string{} // object id → rule that claimed it
	// born marks objects materialized in this chain and not yet amended in
	// it: a later generation may amend what an earlier one materialized (a
	// bank line born, matched and settled by one fact) — a sequence, not a
	// fight. Two materializations or two amendments of one object in one
	// chain remain the conflict a human decides.
	born := map[string]bool{}
	visible := 0 // how many nodes earlier generations contributed
	lookup := func(typ, field string, value any) ([]string, error) {
		ids, err := base(typ, field, value)
		if err != nil {
			return nil, err
		}
		want := fmt.Sprintf("%v", value)
		for _, n := range nodes[:visible] {
			// A replayed prior node is already in the base world: counting
			// it twice would make every ref to it ambiguous.
			if n.amend == nil && n.booked == 0 && n.mat.ObjectType == typ && fmt.Sprintf("%v", n.mat.State[field]) == want {
				ids = append(ids, n.mat.ObjectID)
			}
		}
		return ids, nil
	}
	// State reads see what lookups see: the base world plus the chain's
	// earlier generations (a rate materialized earlier in this chain is
	// readable by a later firing), with earlier amendments merged in.
	get := func(id string) (map[string]any, bool, error) {
		var state map[string]any
		for _, n := range nodes[:visible] {
			if n.amend == nil && n.mat.ObjectID == id {
				state = n.mat.State
			}
		}
		if state == nil {
			s, ok, err := baseGet(id)
			if err != nil || !ok {
				return nil, false, err
			}
			state = s
		}
		for _, n := range nodes[:visible] {
			if n.amend != nil && n.amend.ObjectID == id {
				merged := make(map[string]any, len(state)+len(n.amend.Set))
				for k, v := range state {
					merged[k] = v
				}
				for k, v := range n.amend.Set {
					merged[k] = v
				}
				state = merged
			}
		}
		return state, true, nil
	}
	fire := func(ev core.Event, evPayload any, cause int) error {
		// A fact that came through a door has the door's shape: its ref
		// inputs, vouched by the door, read as links.
		ctx := withDoor(ctx, ev)
		idBase := fmt.Sprintf("%d", root.ID)
		if cause >= 0 {
			idBase = nodes[cause].idBase()
		}
		// What already fired here is replayed from the log, never expanded
		// again: its state is baked, and re-judging an old amendment against
		// today's status would refuse a transition that already happened.
		fired := map[string]bool{}
		for _, p := range prior[idBase] {
			fired[p.ruleID] = true
			switch {
			case p.mat != nil:
				owner[p.mat.ObjectID], born[p.mat.ObjectID] = p.ruleID, true
				nodes = append(nodes, chainNode{mat: *p.mat, cause: cause, booked: p.eventID,
					rule: core.Rule{ID: p.ruleID, Version: p.version}})
			case p.matched != nil || p.unmatched != nil:
				nodes = append(nodes, chainNode{matched: p.matched, unmatched: p.unmatched, cause: cause,
					booked: p.eventID, rule: core.Rule{ID: p.ruleID, Version: p.version}})
			default:
				// amendments end their branch: nothing cascades from them
			}
		}
		// Formula conditions read the world the effects read: the base world
		// plus the chain so far.
		matchEnv := x.env(ctx, evPayload, lookup, get)
		for _, r := range rules {
			if fired[r.ID] {
				continue
			}
			ok, err := matchRule(r, ev, evPayload, matchEnv)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			if r.Spec.Effect.Matching != nil {
				m, u, err := x.expandMatching(ctx, r, evPayload, lookup, get)
				if err != nil {
					return fmt.Errorf("rule %s v%d: %w", r.ID, r.Version, err)
				}
				nodes = append(nodes, chainNode{matched: m, unmatched: u, rule: r, cause: cause})
				continue
			}
			if r.Spec.Effect.Amend != nil {
				ams, err := x.expandAmend(ctx, r, evPayload, lookup, get)
				if err != nil {
					return err
				}
				for _, am := range ams {
					if prev, clash := owner[am.ObjectID]; clash && !born[am.ObjectID] {
						return fmt.Errorf("rules %s and %s both touch %s in one chain — conflicting rules need a human", prev, r.ID, am.ObjectID)
					}
					owner[am.ObjectID], born[am.ObjectID] = r.ID, false
					nodes = append(nodes, chainNode{amend: am, rule: r, cause: cause})
				}
				continue
			}
			mats, err := x.expandEffect(ctx, root, idBase, r, evPayload, lookup, get)
			if err != nil {
				return err
			}
			for _, m := range mats {
				if prev, clash := owner[m.ObjectID]; clash {
					return fmt.Errorf("rules %s and %s both materialize %s — conflicting rules need a human", prev, r.ID, m.ObjectID)
				}
				owner[m.ObjectID], born[m.ObjectID] = r.ID, true
				nodes = append(nodes, chainNode{mat: m, rule: r, cause: cause})
			}
		}
		return nil
	}
	if err := fire(root, payload, -1); err != nil {
		return nil, err
	}
	lo, hi := 0, len(nodes)
	for gen := 1; lo < hi; gen++ {
		if gen > maxCascadeDepth {
			last := nodes[len(nodes)-1]
			return nil, fmt.Errorf("cascade exceeded %d generations — rule %s keeps materializing %s",
				maxCascadeDepth, last.rule.ID, last.objectID())
		}
		visible = hi
		for i := lo; i < hi; i++ {
			n := nodes[i]
			if n.amend != nil {
				continue // amendments end their branch: nothing fires on them
			}
			derived := core.Event{
				Kind: core.KindDerived, Type: n.derivedType(),
				OccurredAt: root.OccurredAt, // business date is inherited down the chain
				RuleID:     n.rule.ID, RuleVersion: n.rule.Version,
			}
			// What a cascade rule sees is the materialization (or the match)
			// itself, shaped exactly as it will be written to the log — plus,
			// under "root", the fact that started the chain (read-only, never
			// written into the derived event): a document's lines live in the
			// delivery, not in the PZ header that cascades them (KK, 2026-10-08).
			evPayload, err := n.derivedPayload(payload)
			if err != nil {
				return nil, err
			}
			if err := fire(derived, evPayload, i); err != nil {
				return nil, err
			}
		}
		lo, hi = hi, len(nodes)
	}
	return nodes, nil
}

// matchRule reports whether rule r fires on ev (payload already decoded):
// event type, effective date, and every where condition — a formula
// condition evaluated in env, the state the rule's effect would read. Shared
// between the live executor and the simulator, so a dry run cannot drift
// from reality.
func matchRule(r core.Rule, ev core.Event, payload any, env *core.Env) (bool, error) {
	if r.Spec.Match.EventType != ev.Type {
		return false, nil
	}
	if r.EffectiveFrom > ev.OccurredAt { // YYYY-MM-DD compares lexically
		return false, nil
	}
	for _, c := range r.Spec.Match.Where {
		ok, err := c.Holds(payload, env)
		if err != nil {
			return false, fmt.Errorf("rule %s v%d condition: %w", r.ID, r.Version, err)
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

// book writes one event's whole cascade chain atomically: derived events in
// the log plus rows in the object cache — all of it or nothing, so an event
// is never half-explained. A cascaded entry names the derived event that
// caused it, so provenance walks back to the root through the log.
func (x *Executor) book(ctx context.Context, root core.Event, nodes []chainNode) error {
	tx, err := x.Store.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := x.bookTx(ctx, tx, root, nodes); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// bookTx is book inside a transaction the caller owns — a verb's reaction
// books the chain its own event seeds, in the transaction that appends it.
func (x *Executor) bookTx(ctx context.Context, tx store.Querier, root core.Event, nodes []chainNode) error {
	ids := make([]int64, len(nodes))
	for i, n := range nodes {
		if n.booked != 0 {
			ids[i] = n.booked // already in the log: a cause, never written again
			continue
		}
		cause := root.ID
		if n.cause >= 0 {
			cause = ids[n.cause]
		}
		if n.matched != nil || n.unmatched != nil {
			// A match is one event, baked: the allocations and differences
			// as the engine found them. It touches no object itself — the
			// rules that explain it do, in the nodes that follow.
			var v any = n.matched
			if n.unmatched != nil {
				v = n.unmatched
			}
			matchPayload, err := json.Marshal(v)
			if err != nil {
				return err
			}
			id, err := store.AppendEvent(ctx, tx, core.Event{
				Kind:         core.KindDerived,
				Type:         n.derivedType(),
				OccurredAt:   root.OccurredAt,
				Payload:      matchPayload,
				CauseEventID: &cause,
				RuleID:       n.rule.ID,
				RuleVersion:  n.rule.Version,
				Actor:        "kernel",
				Company:      root.Company,
			})
			if err != nil {
				return err
			}
			ids[i] = id
			continue
		}
		if n.amend != nil {
			amendPayload, err := json.Marshal(n.amend)
			if err != nil {
				return err
			}
			id, err := store.AppendEvent(ctx, tx, core.Event{
				Kind:         core.KindDerived,
				Type:         core.EventObjectAmended,
				OccurredAt:   root.OccurredAt,
				Payload:      amendPayload,
				CauseEventID: &cause,
				RuleID:       n.rule.ID,
				RuleVersion:  n.rule.Version,
				Actor:        "kernel",
				Company:      root.Company,
			})
			if err != nil {
				return err
			}
			ids[i] = id
			if err := store.AmendObject(ctx, tx, n.amend.ObjectID, n.amend.Set); err != nil {
				return err
			}
			continue
		}
		matPayload, err := json.Marshal(n.mat)
		if err != nil {
			return err
		}
		id, err := store.AppendEvent(ctx, tx, core.Event{
			Kind:         core.KindDerived,
			Type:         core.EventObjectMaterialized,
			OccurredAt:   root.OccurredAt, // derived events inherit the business date
			Payload:      matPayload,
			CauseEventID: &cause,
			RuleID:       n.rule.ID,
			RuleVersion:  n.rule.Version,
			Actor:        "kernel", // rule provenance explains the rest
			Company:      root.Company, // and the partition key
		})
		if err != nil {
			return err
		}
		ids[i] = id
		if err := store.InsertObject(ctx, tx, core.Object{
			ID: n.mat.ObjectID, Type: n.mat.ObjectType, TypeVersion: n.mat.TypeVersion,
			State: n.mat.State, SourceEventID: cause, RuleID: n.rule.ID, RuleVersion: n.rule.Version,
			Company: root.Company,
		}); err != nil {
			return err
		}
	}
	// The single writer is the single announcer: one notice per booked chain,
	// naming the touched types, delivered only if this transaction commits.
	touched, seen := []string{}, map[string]bool{}
	for _, n := range nodes {
		if n.booked != 0 || n.matched != nil || n.unmatched != nil {
			continue
		}
		if !seen[n.objectType()] {
			seen[n.objectType()] = true
			touched = append(touched, n.objectType())
		}
	}
	return store.NotifyProjection(ctx, tx, strings.Join(touched, ","))
}

// expandEffect turns a matched rule into the objects it materializes, using
// lookup for all state access so the simulator can reuse it verbatim. The
// payload may be the root event's or a cascaded derived event's; the business
// date always comes from the root, and ids build on idBase — the root event
// id for root firings, the causing object's id for cascaded ones — which
// keeps every id reproducible without any sequence state.
func (x *Executor) expandEffect(ctx context.Context, root core.Event, idBase string, r core.Rule, payload any, lookup core.Lookup, get core.Getter) ([]core.MaterializedObject, error) {
	if r.Spec.Effect.Postings != nil {
		postingType, err := x.objectType(ctx, PostingObjectType)
		if err != nil {
			return nil, fmt.Errorf("rule %s v%d: %w", r.ID, r.Version, err)
		}
		mats, err := ExpandPostings(r.Spec.Effect.Postings, postingType, root, idBase, r.ID, payload, x.env(ctx, payload, lookup, get))
		if err != nil {
			return nil, fmt.Errorf("rule %s v%d postings: %w", r.ID, r.Version, err)
		}
		return mats, nil
	}

	tmpl := r.Spec.Effect.Object
	objType, err := x.objectType(ctx, tmpl.Type)
	if err != nil {
		return nil, fmt.Errorf("rule %s v%d: %w", r.ID, r.Version, err)
	}
	if tmpl.Each == "" {
		state, calc, err := Expand(tmpl, objType, payload, x.env(ctx, payload, lookup, get))
		if err != nil {
			return nil, fmt.Errorf("rule %s v%d expand: %w", r.ID, r.Version, err)
		}
		// Deterministic object identity: derived from the root event (or the
		// causing object when cascaded), so replay and simulation reproduce
		// it exactly (invariant 4).
		return []core.MaterializedObject{{
			ObjectID:    fmt.Sprintf("%s-%s", objType.Name, idBase),
			ObjectType:  objType.Name,
			TypeVersion: objType.Version,
			State:       state,
			Calc:        calc,
		}}, nil
	}

	// Each: one object per element of the fan-out — the posting pattern
	// generalized to any type. Templates evaluate against {doc, line, n};
	// ids gain the line number, so two each-rules claiming the same type
	// collide like any other same-id ambiguity. An empty fan-out is a rule
	// error: a document with no lines is malformed, not silently explained.
	t, err := core.ParseTemplate(tmpl.Each)
	if err != nil {
		return nil, fmt.Errorf("rule %s v%d each: %w", r.ID, r.Version, err)
	}
	v, err := t.Eval(payload, core.FieldDef{}, nil)
	if err != nil {
		return nil, fmt.Errorf("rule %s v%d each: %w", r.ID, r.Version, err)
	}
	arr, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("rule %s v%d each: %s is not an array", r.ID, r.Version, tmpl.Each)
	}
	if len(arr) == 0 {
		return nil, fmt.Errorf("rule %s v%d each: %s matched no elements", r.ID, r.Version, tmpl.Each)
	}
	mats := make([]core.MaterializedObject, 0, len(arr))
	for i, el := range arr {
		scope := map[string]any{"doc": payload, "line": el, "n": i + 1}
		// The chain's root fact stays readable inside each, as it is on the
		// cascade payload itself: a line's date is the delivery's date.
		if m, ok := payload.(map[string]any); ok {
			if root, ok := m["root"]; ok {
				scope["root"] = root
			}
		}
		state, calc, err := Expand(tmpl, objType, scope, x.env(ctx, scope, lookup, get))
		if err != nil {
			return nil, fmt.Errorf("rule %s v%d line %d: %w", r.ID, r.Version, i+1, err)
		}
		mats = append(mats, core.MaterializedObject{
			ObjectID:    fmt.Sprintf("%s-%s-%d", objType.Name, idBase, i+1),
			ObjectType:  objType.Name,
			TypeVersion: objType.Version,
			State:       state,
			Calc:        calc,
		})
	}
	return mats, nil
}

// expandAmend turns a matched amend rule into the deltas it applies: target
// resolved and vouched (the object exists, and identity is typed by
// construction — <type>-… — so the id itself is the type check), values
// baked in the payload's canonical encoding, and the lifecycle law enforced
// where the kernel knows the current status: a Set touching the declared
// lifecycle field must move along a declared transition, judged the way
// balance is judged — at expansion, before anything books. With Each, one
// delta per element of the fan-out, templates evaluating against {doc, line,
// n} as an object fan-out's do: the items a match allocated, each moved.
func (x *Executor) expandAmend(ctx context.Context, r core.Rule, payload any, lookup core.Lookup, get core.Getter) ([]*core.AmendedObject, error) {
	tmpl := r.Spec.Effect.Amend
	objType, err := x.objectType(ctx, tmpl.Type)
	if err != nil {
		return nil, fmt.Errorf("rule %s v%d: %w", r.ID, r.Version, err)
	}
	if err := (core.RuleSpec{
		Match:  core.Match{EventType: "-"},
		Effect: core.Effect{Amend: tmpl},
	}).Validate(&objType); err != nil {
		return nil, fmt.Errorf("rule %s v%d: %w", r.ID, r.Version, err)
	}
	if tmpl.Each == "" {
		am, err := x.amendOne(ctx, r, objType, payload, lookup, get)
		if err != nil {
			return nil, err
		}
		return []*core.AmendedObject{am}, nil
	}
	t, err := core.ParseTemplate(tmpl.Each)
	if err != nil {
		return nil, fmt.Errorf("rule %s v%d each: %w", r.ID, r.Version, err)
	}
	v, err := t.Eval(payload, core.FieldDef{}, nil)
	if err != nil {
		return nil, fmt.Errorf("rule %s v%d each: %w", r.ID, r.Version, err)
	}
	arr, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("rule %s v%d each: %s is not an array", r.ID, r.Version, tmpl.Each)
	}
	if len(arr) == 0 {
		return nil, fmt.Errorf("rule %s v%d each: %s matched no elements", r.ID, r.Version, tmpl.Each)
	}
	out := make([]*core.AmendedObject, 0, len(arr))
	for i, el := range arr {
		scope := map[string]any{"doc": payload, "line": el, "n": i + 1}
		if m, ok := payload.(map[string]any); ok {
			if root, ok := m["root"]; ok {
				scope["root"] = root
			}
		}
		am, err := x.amendOne(ctx, r, objType, scope, lookup, get)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		out = append(out, am)
	}
	return out, nil
}

// amendOne expands one amendment against one scope.
func (x *Executor) amendOne(ctx context.Context, r core.Rule, objType core.ObjectType, payload any, lookup core.Lookup, get core.Getter) (*core.AmendedObject, error) {
	tmpl := r.Spec.Effect.Amend
	env := x.env(ctx, payload, lookup, get)
	t, err := core.ParseTemplate(tmpl.Target)
	if err != nil {
		return nil, fmt.Errorf("rule %s v%d target: %w", r.ID, r.Version, err)
	}
	v, _, err := t.Evaluate(payload, core.FieldDef{}, env)
	if err != nil {
		return nil, fmt.Errorf("rule %s v%d target: %w", r.ID, r.Version, err)
	}
	id, ok := v.(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("rule %s v%d target: %v is not an object id", r.ID, r.Version, v)
	}
	if !strings.HasPrefix(id, tmpl.Type+"-") {
		return nil, fmt.Errorf("rule %s v%d: %q is not a %s", r.ID, r.Version, id, tmpl.Type)
	}
	cur, found, err := get(id)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("rule %s v%d amend: no %s %q — the event waits for its object", r.ID, r.Version, tmpl.Type, id)
	}
	set := make(map[string]any, len(tmpl.Set))
	var calc map[string]core.Calc
	for name, raw := range tmpl.Set {
		fd, _ := objType.Field(name)
		tp, err := core.ParseTemplate(raw)
		if err != nil {
			return nil, fmt.Errorf("rule %s v%d set %q: %w", r.ID, r.Version, name, err)
		}
		val, c, err := tp.Evaluate(payload, fd, env)
		if err != nil {
			return nil, fmt.Errorf("rule %s v%d set %q: %w", r.ID, r.Version, name, err)
		}
		if err := vouchRef(fd, tp, val, get); err != nil {
			return nil, fmt.Errorf("rule %s v%d set %q: %w", r.ID, r.Version, name, err)
		}
		set[name] = val
		if c != nil {
			if calc == nil {
				calc = map[string]core.Calc{}
			}
			calc[name] = *c
		}
	}
	// Consent was judged per field by Validate; the lifecycle field, when the
	// type has one, additionally moves only along a declared transition. An
	// enrichment-only type (amendable fields, no lifecycle) has no law to apply.
	// A status set to what it already is is a transition like any other: a
	// type that lets an open invoice stay open under a partial payment
	// declares open → open (DECISIONS 2026-10-11, the matching attempt).
	lc := objType.Lifecycle
	if lc == nil {
		return &core.AmendedObject{ObjectID: id, ObjectType: tmpl.Type, Set: set, Calc: calc}, nil
	}
	if nv, touched := set[lc.Field]; touched {
		from, _ := cur[lc.Field].(string)
		to, _ := nv.(string)
		allowed := false
		for _, candidate := range lc.Transitions[from] {
			if candidate == to {
				allowed = true
			}
		}
		if !allowed {
			return nil, fmt.Errorf("rule %s v%d: %s %s→%s is not a declared transition of %s",
				r.ID, r.Version, lc.Field, from, to, objType.Name)
		}
	}
	return &core.AmendedObject{ObjectID: id, ObjectType: tmpl.Type, Set: set, Calc: calc}, nil
}

// Expand evaluates every field template against the payload, typed by the
// object type. Missing required fields, evaluation failures or unresolvable
// refs abort the whole expansion — a half-materialized object never exists.
// Computed fields return their Calc — formula and inputs — for the derived
// event to carry; nil when nothing was computed.
func Expand(tmpl core.ObjectTemplate, objType core.ObjectType, payload any, env *core.Env) (map[string]any, map[string]core.Calc, error) {
	if err := (core.RuleSpec{
		Match:  core.Match{EventType: "-"},
		Effect: core.Effect{Object: tmpl},
	}).Validate(&objType); err != nil {
		return nil, nil, err
	}
	state := make(map[string]any, len(tmpl.Fields))
	var calc map[string]core.Calc
	for name, raw := range tmpl.Fields {
		fd, _ := objType.Field(name)
		t, err := core.ParseTemplate(raw)
		if err != nil {
			return nil, nil, fmt.Errorf("field %q: %w", name, err)
		}
		v, c, err := t.Evaluate(payload, fd, env)
		if err != nil {
			return nil, nil, fmt.Errorf("field %q: %w", name, err)
		}
		if err := vouchRef(fd, t, v, env.Get); err != nil {
			return nil, nil, fmt.Errorf("field %q: %w", name, err)
		}
		state[name] = v
		if c != nil {
			if calc == nil {
				calc = map[string]core.Calc{}
			}
			calc[name] = *c
		}
	}
	return state, calc, nil
}

// env is the formula environment of one expansion: lookups and reads
// against the chain-visible world, the catalog that types object fields,
// and which payload prefixes hold typed object state — a cascade payload
// {object_type, state} types $.state, an each scope {doc: …} types
// $.doc.state. Formulas read money as money and follow refs through links
// because of this; every value they compute is baked into the event.
func (x *Executor) env(ctx context.Context, payload any, lookup core.Lookup, get core.Getter) *core.Env {
	// The catalog is read once per type per expansion: a formula asks for a
	// field's type on every read, and a fold over a thousand objects must
	// not turn that into a thousand store queries. Types do not change
	// inside one expansion, so the memo is exact.
	type known struct {
		t  core.ObjectType
		ok bool
	}
	memo := map[string]known{}
	typed := typedPrefixes(payload)
	shape, isDoor := x.doorShape(ctx)
	if isDoor {
		if typed == nil {
			typed = map[string]string{}
		}
		typed["$"] = shape.Name
		memo[shape.Name] = known{shape, true}
	}
	return &core.Env{
		Lookup: lookup,
		Get:    get,
		Types: func(name string) (core.ObjectType, bool) {
			if k, seen := memo[name]; seen {
				return k.t, k.ok
			}
			t, err := x.objectType(ctx, name)
			memo[name] = known{t, err == nil}
			return t, err == nil
		},
		Typed:         typed,
		MaxCollection: x.MaxCollection,
		Balance: func(book, account, from, to string) (int64, error) {
			return x.Store.AccountBalance(ctx, book, account, from, to)
		},
	}
}

// doorKey carries, on the context of one firing, the door a raw fact came
// through.
type doorKey struct{}

type doorRef struct {
	name    string
	version int
}

// withDoor marks a firing on a fact that came through a declared door. Facts
// rules derive, and facts with no door, carry nothing.
func withDoor(ctx context.Context, ev core.Event) context.Context {
	if ev.Kind != core.KindRaw || ev.ActivityName == "" {
		return ctx
	}
	return context.WithValue(ctx, doorKey{}, doorRef{ev.ActivityName, ev.ActivityVersion})
}

// doorShape is a door fact's payload as a type: the door's inputs as fields,
// so a ref input — an object the door vouched exists and is what the
// declaration says — reads as a link ($.operation.amount). A door whose
// payload is passed through (a json input: the open door) has no shape of
// its own.
func (x *Executor) doorShape(ctx context.Context) (core.ObjectType, bool) {
	d, ok := ctx.Value(doorKey{}).(doorRef)
	if !ok || x.Store == nil {
		return core.ObjectType{}, false
	}
	act, err := x.Store.GetActivityVersion(ctx, d.name, d.version)
	if err != nil {
		return core.ObjectType{}, false
	}
	t := core.ObjectType{Name: "door:" + d.name, Version: d.version}
	for _, in := range act.Spec.Inputs {
		if in.Type == "json" {
			return core.ObjectType{}, false
		}
		if in.Type == "file" || in.Name == "occurred_at" {
			continue
		}
		t.Fields = append(t.Fields, core.FieldDef{Name: in.Name, Type: in.Type, Values: in.Values})
	}
	return t, true
}

func typedPrefixes(payload any) map[string]string {
	m, ok := payload.(map[string]any)
	if !ok {
		return nil
	}
	if t, ok := m["object_type"].(string); ok {
		if _, ok := m["state"]; ok {
			return map[string]string{"$.state": t}
		}
	}
	if doc, ok := m["doc"].(map[string]any); ok {
		if t, ok := doc["object_type"].(string); ok {
			if _, ok := doc["state"]; ok {
				return map[string]string{"$.doc.state": t}
			}
		}
	}
	return nil
}

// vouchRef checks an object id a ref field carries by path (KK, 2026-10-08:
// a follow-up may point at exactly the thing that caused it): it is an id
// of the declared kind — identity is typed by construction, <type>-… — and
// the object exists, in the world or earlier in this chain. The same courtesy
// the amendment pays its target. Resolved refs (=ref) need no vouching.
func vouchRef(fd core.FieldDef, t core.Template, v any, get core.Getter) error {
	target, isRef := core.RefTarget(fd.Type)
	if !isRef || !t.CarriesID() {
		return nil
	}
	id, ok := v.(string)
	if !ok || !strings.HasPrefix(id, target+"-") {
		return fmt.Errorf("%v is not a %s id", v, target)
	}
	if get == nil {
		return nil // structural contexts without state
	}
	if _, found, err := get(id); err != nil {
		return err
	} else if !found {
		return fmt.Errorf("no %s %q", target, id)
	}
	return nil
}

// Replay rebuilds the object projection cache from the derived events in the
// log — the determinism invariant made executable. It returns the rebuilt
// objects in log order.
func (x *Executor) Replay(ctx context.Context) ([]core.Object, error) {
	derived, err := x.Store.EventsByKind(ctx, core.KindDerived)
	if err != nil {
		return nil, err
	}
	tx, err := x.Store.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM object`); err != nil {
		return nil, err
	}
	var out []core.Object
	for _, ev := range derived {
		switch ev.Type {
		case core.EventObjectMaterialized:
			var mat core.MaterializedObject
			if err := json.Unmarshal(ev.Payload, &mat); err != nil {
				return nil, fmt.Errorf("derived event %d: %w", ev.ID, err)
			}
			o := core.Object{
				ID: mat.ObjectID, Type: mat.ObjectType, TypeVersion: mat.TypeVersion,
				State: mat.State, SourceEventID: *ev.CauseEventID,
				RuleID: ev.RuleID, RuleVersion: ev.RuleVersion, Company: ev.Company,
			}
			if err := store.InsertObject(ctx, tx, o); err != nil {
				return nil, err
			}
			out = append(out, o)
		case core.EventObjectAmended:
			// Amendments re-apply in log order, deltas baked at firing time —
			// replay never re-evaluates, so determinism is by construction.
			var am core.AmendedObject
			if err := json.Unmarshal(ev.Payload, &am); err != nil {
				return nil, fmt.Errorf("derived event %d: %w", ev.ID, err)
			}
			if err := store.AmendObject(ctx, tx, am.ObjectID, am.Set); err != nil {
				return nil, fmt.Errorf("derived event %d: %w", ev.ID, err)
			}
		}
	}
	// A replay rewrites the whole cache; every live screen should look again.
	if err := store.NotifyProjection(ctx, tx, "replay"); err != nil {
		return nil, err
	}
	return out, tx.Commit(ctx)
}
