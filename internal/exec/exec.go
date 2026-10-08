// Package exec is the deterministic kernel: it evaluates active rules against
// events, emits derived events with full provenance, and maintains the object
// projection cache. It is the only writer of state (SPEC §3).
package exec

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"rhea/internal/core"
	"rhea/internal/store"
)

type Executor struct {
	Store *store.Store
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
// materialization or an amendment, the rule that produced it, and which
// chain entry's derived event caused it (-1 is the root raw event). Nodes
// book in chain order. Amendments end their branch: nothing fires on them.
type chainNode struct {
	mat   core.MaterializedObject
	amend *core.AmendedObject
	rule  core.Rule
	cause int
	// booked is the derived event already in the log for this node: a
	// prior firing replayed into a backfill chain, never written again.
	booked int64
}

// priorFiring is one consequence already in the log, keyed in Prior by the
// point in a chain it fired at: the root event's id, or the causing object's.
type priorFiring struct {
	ruleID  string
	version int
	eventID int64
	mat     *core.MaterializedObject // nil for an amendment
}

// Prior is a root event's existing derivations, by chain point. A backfill
// expansion replays them from the log, baked state and all, and expands only
// the rules that never fired at that point; live firing passes nil.
type Prior map[string][]priorFiring

// objectID names the object this node touches — the unit the same-id
// conflict guard speaks about, for both kinds of consequence.
func (n chainNode) objectID() string {
	if n.amend != nil {
		return n.amend.ObjectID
	}
	return n.mat.ObjectID
}

func (n chainNode) objectType() string {
	if n.amend != nil {
		return n.amend.ObjectType
	}
	return n.mat.ObjectType
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
	nodes, err := x.expandChain(ctx, ev, rules, payload, lookup, get, nil)
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
func (x *Executor) expandChain(ctx context.Context, root core.Event, rules []core.Rule, payload any, base core.Lookup, baseGet core.Getter, prior Prior) ([]chainNode, error) {
	var nodes []chainNode
	owner := map[string]string{} // object id → rule that claimed it
	visible := 0                 // how many nodes earlier generations contributed
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
		idBase := fmt.Sprintf("%d", root.ID)
		if cause >= 0 {
			idBase = nodes[cause].mat.ObjectID
		}
		// What already fired here is replayed from the log, never expanded
		// again: its state is baked, and re-judging an old amendment against
		// today's status would refuse a transition that already happened.
		fired := map[string]bool{}
		for _, p := range prior[idBase] {
			fired[p.ruleID] = true
			if p.mat == nil {
				continue // amendments end their branch: nothing cascades from them
			}
			owner[p.mat.ObjectID] = p.ruleID
			nodes = append(nodes, chainNode{mat: *p.mat, cause: cause, booked: p.eventID,
				rule: core.Rule{ID: p.ruleID, Version: p.version}})
		}
		for _, r := range rules {
			if fired[r.ID] {
				continue
			}
			ok, err := matchRule(r, ev, evPayload)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			if r.Spec.Effect.Amend != nil {
				am, err := x.expandAmend(ctx, r, evPayload, lookup, get)
				if err != nil {
					return err
				}
				if prev, clash := owner[am.ObjectID]; clash {
					return fmt.Errorf("rules %s and %s both touch %s in one chain — conflicting rules need a human", prev, r.ID, am.ObjectID)
				}
				owner[am.ObjectID] = r.ID
				nodes = append(nodes, chainNode{amend: am, rule: r, cause: cause})
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
				owner[m.ObjectID] = r.ID
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
				Kind: core.KindDerived, Type: core.EventObjectMaterialized,
				OccurredAt: root.OccurredAt, // business date is inherited down the chain
				RuleID:     n.rule.ID, RuleVersion: n.rule.Version,
			}
			// What a cascade rule sees is the materialization itself, shaped
			// exactly as it will be written to the log — plus, under "root",
			// the fact that started the chain (read-only, never written into
			// the derived event): a document's lines live in the delivery,
			// not in the PZ header that cascades them (KK, 2026-10-08).
			evPayload := map[string]any{
				"object_id":    n.mat.ObjectID,
				"object_type":  n.mat.ObjectType,
				"type_version": n.mat.TypeVersion,
				"state":        n.mat.State,
				"root":         payload,
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
// event type, effective date, and every where condition. Shared between the
// live executor and the simulator, so a dry run cannot drift from reality.
func matchRule(r core.Rule, ev core.Event, payload any) (bool, error) {
	if r.Spec.Match.EventType != ev.Type {
		return false, nil
	}
	if r.EffectiveFrom > ev.OccurredAt { // YYYY-MM-DD compares lexically
		return false, nil
	}
	for _, c := range r.Spec.Match.Where {
		ok, err := core.EvalCondition(payload, c)
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
		})
		if err != nil {
			return err
		}
		ids[i] = id
		if err := store.InsertObject(ctx, tx, core.Object{
			ID: n.mat.ObjectID, Type: n.mat.ObjectType, TypeVersion: n.mat.TypeVersion,
			State: n.mat.State, SourceEventID: cause, RuleID: n.rule.ID, RuleVersion: n.rule.Version,
		}); err != nil {
			return err
		}
	}
	// The single writer is the single announcer: one notice per booked chain,
	// naming the touched types, delivered only if this transaction commits.
	touched, seen := []string{}, map[string]bool{}
	for _, n := range nodes {
		if n.booked != 0 {
			continue
		}
		if !seen[n.objectType()] {
			seen[n.objectType()] = true
			touched = append(touched, n.objectType())
		}
	}
	if err := store.NotifyProjection(ctx, tx, strings.Join(touched, ",")); err != nil {
		return err
	}
	return tx.Commit(ctx)
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
		mats, err := ExpandPostings(r.Spec.Effect.Postings, postingType, root, idBase, r.ID, payload, lookup, get)
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
		state, err := Expand(tmpl, objType, payload, lookup, get)
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
		state, err := Expand(tmpl, objType, scope, lookup, get)
		if err != nil {
			return nil, fmt.Errorf("rule %s v%d line %d: %w", r.ID, r.Version, i+1, err)
		}
		mats = append(mats, core.MaterializedObject{
			ObjectID:    fmt.Sprintf("%s-%s-%d", objType.Name, idBase, i+1),
			ObjectType:  objType.Name,
			TypeVersion: objType.Version,
			State:       state,
		})
	}
	return mats, nil
}

// expandAmend turns a matched amend rule into the delta it applies: target
// resolved and vouched (the object exists, and identity is typed by
// construction — <type>-… — so the id itself is the type check), values
// baked in the payload's canonical encoding, and the lifecycle law enforced
// where the kernel knows the current status: a Set touching the declared
// lifecycle field must move along a declared transition, judged the way
// balance is judged — at expansion, before anything books.
func (x *Executor) expandAmend(ctx context.Context, r core.Rule, payload any, lookup core.Lookup, get core.Getter) (*core.AmendedObject, error) {
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
	t, err := core.ParseTemplate(tmpl.Target)
	if err != nil {
		return nil, fmt.Errorf("rule %s v%d target: %w", r.ID, r.Version, err)
	}
	v, err := t.Eval(payload, core.FieldDef{}, lookup)
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
	for name, raw := range tmpl.Set {
		fd, _ := objType.Field(name)
		tp, err := core.ParseTemplate(raw)
		if err != nil {
			return nil, fmt.Errorf("rule %s v%d set %q: %w", r.ID, r.Version, name, err)
		}
		val, err := tp.Eval(payload, fd, lookup)
		if err != nil {
			return nil, fmt.Errorf("rule %s v%d set %q: %w", r.ID, r.Version, name, err)
		}
		if err := vouchRef(fd, tp, val, get); err != nil {
			return nil, fmt.Errorf("rule %s v%d set %q: %w", r.ID, r.Version, name, err)
		}
		set[name] = val
	}
	// Consent was judged per field by Validate; the lifecycle field, when the
	// type has one, additionally moves only along a declared transition. An
	// enrichment-only type (amendable fields, no lifecycle) has no law to apply.
	lc := objType.Lifecycle
	if lc == nil {
		return &core.AmendedObject{ObjectID: id, ObjectType: tmpl.Type, Set: set}, nil
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
	return &core.AmendedObject{ObjectID: id, ObjectType: tmpl.Type, Set: set}, nil
}

// Expand evaluates every field template against the payload, typed by the
// object type. Missing required fields, evaluation failures or unresolvable
// refs abort the whole expansion — a half-materialized object never exists.
func Expand(tmpl core.ObjectTemplate, objType core.ObjectType, payload any, lookup core.Lookup, get core.Getter) (map[string]any, error) {
	if err := (core.RuleSpec{
		Match:  core.Match{EventType: "-"},
		Effect: core.Effect{Object: tmpl},
	}).Validate(&objType); err != nil {
		return nil, err
	}
	state := make(map[string]any, len(tmpl.Fields))
	for name, raw := range tmpl.Fields {
		fd, _ := objType.Field(name)
		t, err := core.ParseTemplate(raw)
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", name, err)
		}
		v, err := t.Eval(payload, fd, lookup)
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", name, err)
		}
		if err := vouchRef(fd, t, v, get); err != nil {
			return nil, fmt.Errorf("field %q: %w", name, err)
		}
		state[name] = v
	}
	return state, nil
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
				RuleID: ev.RuleID, RuleVersion: ev.RuleVersion,
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

