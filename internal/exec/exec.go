// Package exec is the deterministic kernel: it evaluates active rules against
// events, emits derived events with full provenance, and maintains the object
// projection cache. It is the only writer of state (SPEC §3).
package exec

import (
	"context"
	"encoding/json"
	"fmt"

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

// evaluate fires the first matching rule (rules arrive in firing order:
// priority ascending, rule_id tiebreak) whose effective_from covers the
// event's business date.
func (x *Executor) evaluate(ctx context.Context, ev core.Event, rules []core.Rule) (bool, error) {
	var payload any
	if err := json.Unmarshal(ev.Payload, &payload); err != nil {
		return false, fmt.Errorf("payload: %w", err)
	}
	for _, r := range rules {
		if r.Spec.Match.EventType != ev.Type {
			continue
		}
		if r.EffectiveFrom > ev.OccurredAt { // YYYY-MM-DD compares lexically
			continue
		}
		matched := true
		for _, c := range r.Spec.Match.Where {
			ok, err := core.EvalCondition(payload, c)
			if err != nil {
				return false, fmt.Errorf("rule %s v%d condition: %w", r.ID, r.Version, err)
			}
			if !ok {
				matched = false
				break
			}
		}
		if !matched {
			continue
		}
		return true, x.fire(ctx, ev, r, payload)
	}
	return false, nil
}

// fire expands the rule's template and books the result atomically: one
// derived event in the log plus one row in the object cache, or nothing.
func (x *Executor) fire(ctx context.Context, ev core.Event, r core.Rule, payload any) error {
	tmpl := r.Spec.Effect.Object
	objType, err := x.Store.GetObjectType(ctx, tmpl.Type)
	if err != nil {
		return fmt.Errorf("rule %s v%d: %w", r.ID, r.Version, err)
	}
	// ref() resolves against the object cache as of this point in the log;
	// the resolved id is baked into the derived event, so replay never
	// re-resolves and determinism (invariant 4) is untouched.
	lookup := func(typ, field string, value any) (string, error) {
		return x.Store.FindObjectIDByField(ctx, typ, field, fmt.Sprintf("%v", value))
	}
	state, err := Expand(tmpl, objType, payload, lookup)
	if err != nil {
		return fmt.Errorf("rule %s v%d expand: %w", r.ID, r.Version, err)
	}

	// Deterministic object identity: derived from the causing event, so replay
	// reproduces it exactly (invariant 4).
	mat := core.MaterializedObject{
		ObjectID:    fmt.Sprintf("%s-%d", objType.Name, ev.ID),
		ObjectType:  objType.Name,
		TypeVersion: objType.Version,
		State:       state,
	}
	matPayload, err := json.Marshal(mat)
	if err != nil {
		return err
	}

	tx, err := x.Store.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	derivedID, err := store.AppendEvent(ctx, tx, core.Event{
		Kind:         core.KindDerived,
		Type:         core.EventObjectMaterialized,
		OccurredAt:   ev.OccurredAt, // derived events inherit the business date
		Payload:      matPayload,
		CauseEventID: &ev.ID,
		RuleID:       r.ID,
		RuleVersion:  r.Version,
	})
	if err != nil {
		return err
	}
	if err := store.InsertObject(ctx, tx, core.Object{
		ID: mat.ObjectID, Type: mat.ObjectType, TypeVersion: mat.TypeVersion,
		State: mat.State, SourceEventID: ev.ID, RuleID: r.ID, RuleVersion: r.Version,
	}); err != nil {
		return err
	}
	_ = derivedID
	return tx.Commit(ctx)
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
	return out, tx.Commit(ctx)
}

// ApproveRule records a human approval: the approval itself becomes a raw
// event in the log, the rule gets a new active version row, and the pending
// worklist is re-evaluated under the enlarged rule set.
func (x *Executor) ApproveRule(ctx context.Context, ruleID, approvedBy, businessDate string) (core.Rule, int, error) {
	r, err := x.Store.GetRule(ctx, ruleID)
	if err != nil {
		return core.Rule{}, 0, err
	}
	if r.Status != core.StatusDraft {
		return core.Rule{}, 0, fmt.Errorf("rule %s is %s, only drafts can be approved", ruleID, r.Status)
	}
	approval, _ := json.Marshal(map[string]any{
		"rule_id": r.ID, "approved_version": r.Version + 1, "approved_by": approvedBy,
	})
	if _, err := x.Store.AppendEvent(ctx, core.Event{
		Kind: core.KindRaw, Type: "rule.approved", OccurredAt: businessDate,
		Payload: approval, DedupKey: fmt.Sprintf("approve/%s/%d", r.ID, r.Version+1),
	}); err != nil {
		return core.Rule{}, 0, err
	}
	r.Status = core.StatusActive
	r, err = x.Store.InsertRuleVersion(ctx, r)
	if err != nil {
		return core.Rule{}, 0, err
	}
	booked, errs := x.ProcessPending(ctx)
	if len(errs) > 0 {
		return r, booked, errs[0]
	}
	return r, booked, nil
}
