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
	Bundle   *core.Bundle   // approve_bundle / reject_bundle: the new version
	// Backfilled counts the past events approve_backfill explained further.
	Backfilled int
	backfill   *BackfillPlan // planned before the append, booked after commit
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

	if out.backfill != nil {
		out.Backfilled, out.Errors = x.executeBackfill(ctx, out.backfill)
	}
	booked, errs := x.ProcessPending(ctx)
	out.Booked, out.Errors = booked, append(out.Errors, errs...)
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
		if err := x.refuseMember(ctx, core.Member{Kind: core.KindRule, Name: r.ID, Version: r.Version}); err != nil {
			return nil, err
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
		if err := x.refuseMember(ctx, core.Member{Kind: core.KindActivity, Name: a.Name, Version: a.Version}); err != nil {
			return nil, err
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

	case "bundle.approved":
		var p struct {
			BundleID   string `json:"bundle_id"`
			ApprovedBy string `json:"approved_by"`
		}
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return nil, err
		}
		b, err := x.openBundle(ctx, p.BundleID)
		if err != nil {
			return nil, err
		}
		// Gather every member's draft before anything is appended: a member
		// that moved since the bundle was drafted refuses the whole approval.
		var writes []func(tx store.Querier) error
		for _, m := range b.Members {
			if state, err := x.Store.MemberStatus(ctx, m); err != nil {
				return nil, err
			} else if state != core.StatusDraft {
				return nil, fmt.Errorf("bundle %s: %s %s v%d is %s — nothing activates", b.ID, m.Kind, m.Name, m.Version, state)
			}
			switch m.Kind {
			case core.KindRule:
				r, err := x.Store.GetRule(ctx, m.Name)
				if err != nil {
					return nil, err
				}
				writes = append(writes, func(tx store.Querier) error {
					r.Status = core.StatusActive
					_, err := store.InsertRuleVersion(ctx, tx, r)
					return err
				})
			case core.KindActivity:
				a, err := x.Store.GetActivity(ctx, m.Name)
				if err != nil {
					return nil, err
				}
				writes = append(writes, func(tx store.Querier) error {
					a.Status = core.StatusActive
					_, err := store.InsertActivityVersion(ctx, tx, a)
					return err
				})
			case core.KindObjectType:
				t, _, err := x.Store.GetObjectTypeVersion(ctx, m.Name, m.Version)
				if err != nil {
					return nil, err
				}
				writes = append(writes, func(tx store.Querier) error {
					return store.InsertObjectTypeRow(ctx, tx, t, core.StatusActive)
				})
			case core.KindViewDef:
				v, _, err := x.Store.GetViewDefVersion(ctx, m.Name, m.Version)
				if err != nil {
					return nil, err
				}
				writes = append(writes, func(tx store.Querier) error {
					return store.InsertViewDefRow(ctx, tx, v, core.StatusActive)
				})
			}
		}
		approval, _ := json.Marshal(map[string]any{
			"bundle_id": b.ID, "approved_version": b.Version + 1,
			"approved_by": p.ApprovedBy, "members": b.Members,
		})
		ev.Payload = approval
		ev.DedupKey = fmt.Sprintf("bundle-approve/%s/%d", b.ID, b.Version+1)
		return func(tx store.Querier) error {
			for _, w := range writes {
				if err := w(tx); err != nil {
					return err
				}
			}
			b.Status = core.StatusActive
			b, err = store.InsertBundleVersion(ctx, tx, b)
			out.Bundle = &b
			return err
		}, nil

	case "backfill.approved":
		var p struct {
			ApprovedBy string `json:"approved_by"`
		}
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return nil, err
		}
		plan, err := x.PlanBackfill(ctx, ev.OccurredAt)
		if err != nil {
			return nil, err
		}
		if len(plan.Chains) == 0 {
			return nil, fmt.Errorf("nothing to backfill: every past event is explained as far as the active rules go")
		}
		// The approval names exactly what it promotes: every chain, its
		// rules, and whether it was offered forward out of a closed period.
		approval, _ := json.Marshal(map[string]any{
			"approved_by": p.ApprovedBy, "chains": plan.Chains, "refused": plan.Errors,
		})
		ev.Payload = approval
		out.backfill = &plan
		return nil, nil

	case "bundle.rejected":
		var p struct {
			BundleID   string `json:"bundle_id"`
			Reason     string `json:"reason"`
			RejectedBy string `json:"rejected_by"`
		}
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return nil, err
		}
		b, err := x.openBundle(ctx, p.BundleID)
		if err != nil {
			return nil, err
		}
		// The members are left as they are: rejected drafts, locked by their
		// membership. Superseding a draft rule would retire the active
		// version it was redrafting.
		rejection, _ := json.Marshal(map[string]any{
			"bundle_id": b.ID, "rejected_version": b.Version + 1,
			"reason": p.Reason, "rejected_by": p.RejectedBy, "members": b.Members,
		})
		ev.Payload = rejection
		ev.DedupKey = fmt.Sprintf("bundle-reject/%s/%d", b.ID, b.Version+1)
		return func(tx store.Querier) error {
			b.Status = core.StatusSuperseded
			b, err = store.InsertBundleVersion(ctx, tx, b)
			out.Bundle = &b
			return err
		}, nil
	}
	return nil, nil
}

// openBundle returns a bundle still awaiting its decision.
func (x *Executor) openBundle(ctx context.Context, id string) (core.Bundle, error) {
	b, err := x.Store.GetBundle(ctx, id)
	if err != nil {
		return b, err
	}
	if b.Status != core.StatusDraft {
		return b, fmt.Errorf("bundle %s is %s, only drafts can be decided", b.ID, b.Status)
	}
	return b, nil
}

// refuseMember keeps the single-definition doors from approving part of a
// bundle (KK, 2026-10-08: no partial approval).
func (x *Executor) refuseMember(ctx context.Context, m core.Member) error {
	id, status, ok, err := x.Store.BundleOf(ctx, m)
	if err != nil || !ok {
		return err
	}
	if status == core.StatusDraft {
		return fmt.Errorf("%s %s v%d belongs to bundle %s — approve the bundle, never part of it", m.Kind, m.Name, m.Version, id)
	}
	return fmt.Errorf("%s %s v%d belongs to bundle %s, which is %s — redraft it", m.Kind, m.Name, m.Version, id, status)
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

// ApproveBundle activates a bundle through the approve_bundle door: every
// member in one transaction, then one worklist pass under the enlarged
// vocabulary — so an event needing two of the bundle's rules gets both.
func (x *Executor) ApproveBundle(ctx context.Context, id, approvedBy, businessDate string) (core.Bundle, int, []error, error) {
	t, err := x.TriggerActivity(ctx, "approve_bundle", map[string]any{
		"bundle_id": id, "approved_by": approvedBy, "occurred_at": businessDate,
	}, approvedBy, "")
	if err != nil {
		return core.Bundle{}, 0, nil, err
	}
	return *t.Bundle, t.Booked, t.Errors, nil
}

// RejectBundle records a rejection with its reason through reject_bundle.
func (x *Executor) RejectBundle(ctx context.Context, id, reason, rejectedBy, businessDate string) (core.Bundle, error) {
	t, err := x.TriggerActivity(ctx, "reject_bundle", map[string]any{
		"bundle_id": id, "reason": reason, "rejected_by": rejectedBy, "occurred_at": businessDate,
	}, rejectedBy, "")
	if err != nil {
		return core.Bundle{}, err
	}
	return *t.Bundle, nil
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
