package exec

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"rhea/internal/core"
	"rhea/internal/rights"
	"rhea/internal/store"
)

// A GuardError is a chain the kinds of record it writes refuse: a record
// born in another state than its type begins in, or a second record of a
// type's key in one company (2026-10-27).
type GuardError struct{ Reason string }

func (e *GuardError) Error() string { return e.Reason }

// vetChain holds a chain to what the kinds of record it writes guard,
// whichever rule of whichever App wrote it (KK, 2026-10-27: the approval
// guards the table, not one App's screen): every record begins in its
// type's first state; a key names one record per company; a move a type
// gives to a group is made by a person in that group, never by the one who
// proposed the record while someone else could. A consequence already in
// the log (a backfill's prior firing) was vetted when it was written.
func (x *Executor) vetChain(ctx context.Context, root core.Event, nodes []chainNode, base core.Lookup, baseGet core.Getter) error {
	types := map[string]core.ObjectType{}
	typeOf := func(name string) (core.ObjectType, error) {
		if t, ok := types[name]; ok {
			return t, nil
		}
		t, err := x.objectType(ctx, name)
		if err == nil {
			types[name] = t
		}
		return t, err
	}
	born := map[string]map[string]any{} // the records this chain makes, as made
	keys := map[string]string{}         // type|key value → the record that holds it in this chain
	for _, n := range nodes {
		if n.booked != 0 {
			continue
		}
		switch {
		case n.amend != nil:
			am := n.amend
			t, err := typeOf(am.ObjectType)
			if err != nil {
				return err
			}
			lc := t.Lifecycle
			to, touched := am.Set[lcField(lc)].(string)
			if lc == nil || len(lc.Moves) == 0 || !touched {
				continue
			}
			cur := born[am.ObjectID]
			if cur == nil {
				if cur, _, err = baseGet(am.ObjectID); err != nil {
					return err
				}
			}
			from, _ := cur[lc.Field].(string)
			mv, guarded := lc.MoveOf(from, to)
			if !guarded {
				continue
			}
			proposer := ""
			if mv.NotBy != "" {
				proposer, _ = cur[mv.NotBy].(string)
			}
			reason, err := rights.MayMove(ctx, x.Store, root.Actor, root.Company, mv.Group, proposer, x.today())
			if err != nil {
				return err
			}
			if reason != "" {
				return &rights.ErrRefused{Person: root.Actor, Door: root.ActivityName, Company: root.Company,
					Reason: fmt.Sprintf("%s %s → %s: %s", t.Name, from, to, reason)}
			}
		case n.matched == nil && n.unmatched == nil && n.mat.ObjectID != "":
			m := n.mat
			born[m.ObjectID] = m.State
			t, err := typeOf(m.ObjectType)
			if err != nil {
				return err
			}
			if lc := t.Lifecycle; lc != nil && lc.Initial != "" {
				if st, _ := m.State[lc.Field].(string); st != lc.Initial {
					return &GuardError{fmt.Sprintf("a %s begins %s — rule %s made it %q", t.Name, lc.Initial, n.rule.ID, st)}
				}
			}
			if t.Key == "" {
				continue
			}
			v := core.Text(m.State[t.Key])
			if v == "" {
				continue
			}
			if other, dup := keys[t.Name+"|"+v]; dup {
				return &GuardError{fmt.Sprintf("%s %s %s is made twice in one chain (%s, %s)", t.Name, t.Key, v, other, m.ObjectID)}
			}
			keys[t.Name+"|"+v] = m.ObjectID
			ids, err := base(t.Name, t.Key, v)
			if err != nil {
				return err
			}
			for _, id := range ids {
				if id != m.ObjectID {
					return &GuardError{fmt.Sprintf("%s %s %s is registered already (%s)", t.Name, t.Key, v, id)}
				}
			}
		}
	}
	return nil
}

func lcField(lc *core.LifecycleDef) string {
	if lc == nil {
		return ""
	}
	return lc.Field
}

// precheck runs a door's fact through the chain it would start before the
// fact is written, when a rule it fires writes a kind of record that guards
// itself: a move refused, a second record of a key, a record born in the
// wrong state is the door's refusal, and nothing waits in the worklist for
// a person who may not. The chain is vetted again when it is booked.
func (x *Executor) precheck(ctx context.Context, ev core.Event) error {
	rules, err := x.Store.ActiveRules(ctx)
	if err != nil {
		return err
	}
	guarded := false
	for _, r := range rules {
		if r.Spec.Match.EventType != ev.Type {
			continue
		}
		name := r.Spec.Effect.Object.Type
		if a := r.Spec.Effect.Amend; a != nil {
			name = a.Type
		}
		if name == "" {
			continue
		}
		if t, err := x.objectType(ctx, name); err == nil && (t.Key != "" || t.Lifecycle != nil && (t.Lifecycle.Initial != "" || len(t.Lifecycle.Moves) > 0)) {
			guarded = true
			break
		}
	}
	if !guarded {
		return nil
	}
	ctx = scopeFor(ctx, ev.Company)
	var payload any
	if err := json.Unmarshal(ev.Payload, &payload); err != nil {
		return err
	}
	lookup := func(typ, field string, value any) ([]string, error) {
		return x.Store.FindObjectIDsByField(ctx, typ, field, core.Text(value))
	}
	get := func(id string) (map[string]any, bool, error) {
		o, err := x.Store.GetObject(ctx, id)
		if errors.Is(err, store.ErrNotFound) {
			return nil, false, nil
		}
		if err != nil {
			return nil, false, err
		}
		return o.State, true, nil
	}
	_, err = x.expandChain(ctx, ev, rules, payload, lookup, get, nil, nil)
	var refused *rights.ErrRefused
	var guard *GuardError
	if errors.As(err, &refused) || errors.As(err, &guard) {
		return err
	}
	return nil // anything else the fact meets, it meets in the worklist, as before
}
