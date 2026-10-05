package exec

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"rhea/internal/core"
	"rhea/internal/store"
)

// The declared door (DIRECTION 2026-10-05, activities as data). A trigger
// validates its inputs against the active activity — the door's courtesy,
// spent before the event exists — then appends one raw event stamped
// (activity_name, activity_version): every event names its door. For the
// kernel's own verbs the emitted type carries a reaction: deterministic
// kernel code that runs in the same transaction as the append (the approval
// flip). A reaction must be deterministic — which is exactly why drafting is
// not one: rule.draft_requested is a request to another actor, consumed
// outside the kernel.

// Triggered reports what one trigger did: the appended event, what the
// worklist re-evaluation booked, and — for the approval verbs — the version
// row the reaction activated.
type Triggered struct {
	EventID  int64
	Booked   int
	Errors   []error
	Rule     *core.Rule     // approve_rule: the newly active version
	Activity *core.Activity // approve_activity: the newly active version
}

// TriggerActivity fires a declared verb: input validation, ref existence,
// the stamped append, the reserved reaction, then the ordinary worklist
// re-evaluation. dedupKey is the caller's idempotency handle (the CLI's file
// hash); the approval reactions set their own.
func (x *Executor) TriggerActivity(ctx context.Context, name string, inputs map[string]any, actor, dedupKey string) (Triggered, error) {
	act, err := x.Store.GetActiveActivity(ctx, name)
	if err != nil {
		return Triggered{}, err
	}
	built, err := act.BuildEvent(inputs, time.Now().UTC().Format("2006-01-02"))
	if err != nil {
		return Triggered{}, fmt.Errorf("activity %s: %w", name, err)
	}
	// Ref inputs name objects; the door vouches they exist and are what the
	// declaration says — the same courtesy ref() pays at rule expansion.
	for _, ref := range built.Refs {
		o, err := x.Store.GetObject(ctx, ref.ObjectID)
		if err != nil {
			return Triggered{}, fmt.Errorf("activity %s input %s: no object %q", name, ref.Input, ref.ObjectID)
		}
		if o.Type != ref.TargetType {
			return Triggered{}, fmt.Errorf("activity %s input %s: %q is a %s, not a %s", name, ref.Input, ref.ObjectID, o.Type, ref.TargetType)
		}
	}

	ev := core.Event{
		Kind: core.KindRaw, Type: built.Type, OccurredAt: built.OccurredAt,
		Payload: built.Payload, DedupKey: dedupKey, Actor: actor,
		ActivityName: act.Name, ActivityVersion: act.Version,
	}
	var out Triggered
	apply, err := x.prepareReaction(ctx, &ev, &out)
	if err != nil {
		return Triggered{}, err
	}

	tx, err := x.Store.Pool.Begin(ctx)
	if err != nil {
		return Triggered{}, err
	}
	defer tx.Rollback(ctx)
	if out.EventID, err = store.AppendEvent(ctx, tx, ev); err != nil {
		return Triggered{}, err
	}
	if apply != nil {
		if err := apply(tx); err != nil {
			return Triggered{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Triggered{}, err
	}

	out.Booked, out.Errors = x.ProcessPending(ctx)
	return out, nil
}

// prepareReaction checks a system verb's preconditions before anything is
// appended and returns the write that runs inside the trigger's transaction.
// Only builtins emit into the reserved namespaces (the store refuses them on
// declared activities), so this switch is the closed list of kernel verbs.
func (x *Executor) prepareReaction(ctx context.Context, ev *core.Event, out *Triggered) (func(tx store.Querier) error, error) {
	switch ev.Type {
	case "rule.approved":
		var p struct {
			RuleID     string `json:"rule_id"`
			ApprovedBy string `json:"approved_by"`
		}
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return nil, err
		}
		r, err := x.Store.GetRule(ctx, p.RuleID)
		if err != nil {
			return nil, err
		}
		if r.Status != core.StatusDraft {
			return nil, fmt.Errorf("rule %s is %s, only drafts can be approved", p.RuleID, r.Status)
		}
		approval, _ := json.Marshal(map[string]any{
			"rule_id": r.ID, "approved_version": r.Version + 1, "approved_by": p.ApprovedBy,
		})
		ev.Payload = approval
		ev.DedupKey = fmt.Sprintf("approve/%s/%d", r.ID, r.Version+1)
		return func(tx store.Querier) error {
			r.Status = core.StatusActive
			r, err = store.InsertRuleVersion(ctx, tx, r)
			out.Rule = &r
			return err
		}, nil

	case "activity.approved":
		var p struct {
			Activity   string `json:"activity"`
			ApprovedBy string `json:"approved_by"`
		}
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return nil, err
		}
		a, err := x.Store.GetActivity(ctx, p.Activity)
		if err != nil {
			return nil, err
		}
		if a.Status != core.StatusDraft {
			return nil, fmt.Errorf("activity %s is %s, only drafts can be approved", p.Activity, a.Status)
		}
		approval, _ := json.Marshal(map[string]any{
			"activity": a.Name, "approved_version": a.Version + 1, "approved_by": p.ApprovedBy,
		})
		ev.Payload = approval
		ev.DedupKey = fmt.Sprintf("activity-approve/%s/%d", a.Name, a.Version+1)
		return func(tx store.Querier) error {
			a.Status = core.StatusActive
			a, err = store.InsertActivityVersion(ctx, tx, a)
			out.Activity = &a
			return err
		}, nil
	}
	return nil, nil
}

// ApproveRule records a human approval through the approve_rule door: the
// approval becomes a raw event (invariant 2), the kernel reaction writes the
// new active version row in the same transaction, and the pending worklist
// is re-evaluated under the enlarged rule set. Failures from that
// re-evaluation are worklist conditions, not approval failures — they come
// back as data, separate from the hard error.
func (x *Executor) ApproveRule(ctx context.Context, ruleID, approvedBy, businessDate string) (core.Rule, int, []error, error) {
	t, err := x.TriggerActivity(ctx, "approve_rule", map[string]any{
		"rule_id": ruleID, "approved_by": approvedBy, "occurred_at": businessDate,
	}, approvedBy, "")
	if err != nil {
		return core.Rule{}, 0, nil, err
	}
	return *t.Rule, t.Booked, t.Errors, nil
}

// ApproveActivity is approve_rule's twin for the second vocabulary: the gate
// applied to the gate.
func (x *Executor) ApproveActivity(ctx context.Context, name, approvedBy, businessDate string) (core.Activity, int, []error, error) {
	t, err := x.TriggerActivity(ctx, "approve_activity", map[string]any{
		"activity": name, "approved_by": approvedBy, "occurred_at": businessDate,
	}, approvedBy, "")
	if err != nil {
		return core.Activity{}, 0, nil, err
	}
	return *t.Activity, t.Booked, t.Errors, nil
}
