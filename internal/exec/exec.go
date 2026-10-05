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
		events, err := x.Store.UnmatchedRawEvents(ctx)
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

// chainNode is one would-be materialization in a cascade chain: the object a
// rule firing produced, and which chain entry's derived event caused it
// (-1 is the root raw event). Nodes book in chain order.
type chainNode struct {
	mat   core.MaterializedObject
	rule  core.Rule
	cause int
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
	nodes, err := x.expandChain(ctx, ev, rules, payload, lookup, get)
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
func (x *Executor) expandChain(ctx context.Context, root core.Event, rules []core.Rule, payload any, base core.Lookup, baseGet core.Getter) ([]chainNode, error) {
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
			if n.mat.ObjectType == typ && fmt.Sprintf("%v", n.mat.State[field]) == want {
				ids = append(ids, n.mat.ObjectID)
			}
		}
		return ids, nil
	}
	// State reads see what lookups see: the base world plus the chain's
	// earlier generations (a rate materialized earlier in this chain is
	// readable by a later firing).
	get := func(id string) (map[string]any, bool, error) {
		for _, n := range nodes[:visible] {
			if n.mat.ObjectID == id {
				return n.mat.State, true, nil
			}
		}
		return baseGet(id)
	}
	fire := func(ev core.Event, evPayload any, cause int) error {
		idBase := fmt.Sprintf("%d", root.ID)
		if cause >= 0 {
			idBase = nodes[cause].mat.ObjectID
		}
		for _, r := range rules {
			ok, err := matchRule(r, ev, evPayload)
			if err != nil {
				return err
			}
			if !ok {
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
				maxCascadeDepth, last.rule.ID, last.mat.ObjectID)
		}
		visible = hi
		for i := lo; i < hi; i++ {
			n := nodes[i]
			derived := core.Event{
				Kind: core.KindDerived, Type: core.EventObjectMaterialized,
				OccurredAt: root.OccurredAt, // business date is inherited down the chain
				RuleID:     n.rule.ID, RuleVersion: n.rule.Version,
			}
			// What a cascade rule sees is the materialization itself, shaped
			// exactly as it will be written to the log.
			evPayload := map[string]any{
				"object_id":    n.mat.ObjectID,
				"object_type":  n.mat.ObjectType,
				"type_version": n.mat.TypeVersion,
				"state":        n.mat.State,
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
		matPayload, err := json.Marshal(n.mat)
		if err != nil {
			return err
		}
		cause := root.ID
		if n.cause >= 0 {
			cause = ids[n.cause]
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
		if !seen[n.mat.ObjectType] {
			seen[n.mat.ObjectType] = true
			touched = append(touched, n.mat.ObjectType)
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
		postingType, err := x.Store.GetObjectType(ctx, PostingObjectType)
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
	objType, err := x.Store.GetObjectType(ctx, tmpl.Type)
	if err != nil {
		return nil, fmt.Errorf("rule %s v%d: %w", r.ID, r.Version, err)
	}
	if tmpl.Each == "" {
		state, err := Expand(tmpl, objType, payload, lookup)
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
		state, err := Expand(tmpl, objType, scope, lookup)
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

// Expand evaluates every field template against the payload, typed by the
// object type. Missing required fields, evaluation failures or unresolvable
// refs abort the whole expansion — a half-materialized object never exists.
func Expand(tmpl core.ObjectTemplate, objType core.ObjectType, payload any, lookup core.Lookup) (map[string]any, error) {
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
		state[name] = v
	}
	return state, nil
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
		if ev.Type != core.EventObjectMaterialized {
			continue
		}
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
	}
	// A replay rewrites the whole cache; every live screen should look again.
	if err := store.NotifyProjection(ctx, tx, "replay"); err != nil {
		return nil, err
	}
	return out, tx.Commit(ctx)
}

// ApproveRule records a human approval: the approval itself becomes a raw
// event in the log, the rule gets a new active version row, and the pending
// worklist is re-evaluated under the enlarged rule set. Failures from that
// re-evaluation are worklist conditions, not approval failures — they come
// back as data, separate from the hard error.
func (x *Executor) ApproveRule(ctx context.Context, ruleID, approvedBy, businessDate string) (core.Rule, int, []error, error) {
	r, err := x.Store.GetRule(ctx, ruleID)
	if err != nil {
		return core.Rule{}, 0, nil, err
	}
	if r.Status != core.StatusDraft {
		return core.Rule{}, 0, nil, fmt.Errorf("rule %s is %s, only drafts can be approved", ruleID, r.Status)
	}
	approval, _ := json.Marshal(map[string]any{
		"rule_id": r.ID, "approved_version": r.Version + 1, "approved_by": approvedBy,
	})
	if _, err := x.Store.AppendEvent(ctx, core.Event{
		Kind: core.KindRaw, Type: "rule.approved", OccurredAt: businessDate,
		Payload: approval, DedupKey: fmt.Sprintf("approve/%s/%d", r.ID, r.Version+1),
		Actor: approvedBy,
	}); err != nil {
		return core.Rule{}, 0, nil, err
	}
	r.Status = core.StatusActive
	r, err = x.Store.InsertRuleVersion(ctx, r)
	if err != nil {
		return core.Rule{}, 0, nil, err
	}
	booked, errs := x.ProcessPending(ctx)
	return r, booked, errs, nil
}
