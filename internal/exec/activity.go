package exec

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"rhea/internal/core"
	"rhea/internal/match"
	"rhea/internal/rights"
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
	// Proposed is the bundle the host drafted on this pairing, when a
	// person has paired the same way twice: the standard as a policy step,
	// waiting for approval.
	Proposed *core.Bundle
	// Clients are the companies approve_for_clients taught: its write
	// reaches their books, and what waited there is booked after it.
	Clients []string
}

// matchesOn counts the matches a structure made on a trigger so far — the
// policy's, pairings and reversals alike — the next match's sequence.
func matchesOn(derived []core.Event, matching, triggerID string) int {
	n := 0
	for _, ev := range derived {
		if ev.Type != core.EventObjectsMatched {
			continue
		}
		var m core.MatchedObjects
		if json.Unmarshal(ev.Payload, &m) == nil && m.Matching == matching && m.Trigger.ObjectID == triggerID {
			n++
		}
	}
	return n
}

// withDate adds the business date to a verb's payload under "date" when the
// inputs carry none: the fact a person's act makes is dated like a
// document's, so rules reading $.root.date explain it the same way.
func withDate(payload json.RawMessage, date string) (json.RawMessage, error) {
	var m map[string]any
	if err := json.Unmarshal(payload, &m); err != nil {
		return nil, err
	}
	if _, has := m["date"]; has {
		return payload, nil
	}
	m["date"] = date
	return json.Marshal(m)
}

// learnedProposal decides, before the pairing is appended, whether the host
// proposes a standard: one of the pairing's standards that the policy lacks,
// that an earlier pairing of this structure was also explained by, and that
// was never proposed before — a rejected proposal never returns. It returns
// the write that drafts the rule's next version with the step appended,
// bundled with a client warrant naming both pairings, or nil. A company's
// people teach the company: paired acting for a company, the lesson is the
// company's own rule — replacing Rhea's for it alone when the pairing ran
// on Rhea's — in a bundle of the company's (KK, 2026-10-22).
func (x *Executor) learnedProposal(ctx context.Context, rule core.Rule, company string, matched *core.MatchedObjects, derived []core.Event) (func(tx store.Querier, eventID int64) (*core.Bundle, error), error) {
	tmpl := rule.Spec.Effect.Matching
	if len(matched.Standards) == 0 {
		return nil, nil
	}
	for _, standard := range matched.Standards {
		if hasStandard(tmpl.Policy, standard) {
			continue
		}
		if _, err := x.Store.GetBundle(ctx, core.OwnedID(core.LearnedBundleID(tmpl.Name, standard), company)); err == nil {
			continue // proposed already: open, approved, or overruled — never again
		}
		var earlier int64
		for _, ev := range derived {
			if ev.Type != core.EventObjectsMatched {
				continue
			}
			var m core.MatchedObjects
			if json.Unmarshal(ev.Payload, &m) != nil || m.Matching != tmpl.Name || m.Strategy != match.Selected {
				continue
			}
			for _, s := range m.Standards {
				if s == standard && ev.CauseEventID != nil {
					earlier = *ev.CauseEventID
				}
			}
		}
		if earlier == 0 {
			continue // the first time: noted in the event, nothing proposed yet
		}
		step := core.MatchingStep{Strategy: standard}
		switch standard {
		case "reference partial":
			step = core.MatchingStep{Strategy: match.Reference, Partial: true}
		case match.Subset:
			step.MaxItems = 0
			for name, allocs := range matched.Sides {
				if name != tmpl.Trigger && len(allocs) > step.MaxItems {
					step.MaxItems = len(allocs)
				}
			}
			if step.MaxItems < match.DefaultMaxItems {
				step.MaxItems = match.DefaultMaxItems
			}
		}
		return func(tx store.Querier, eventID int64) (*core.Bundle, error) {
			next := rule
			if company != "" && rule.Company == "" {
				next.ID, next.Company, next.Replaces = core.OwnedID(rule.ID, company), company, rule.ID
			}
			next.Status = core.StatusDraft
			next.CreatedBy = "kernel"
			next.Description = fmt.Sprintf("%s; learned: %s, after the owner paired this way twice", rule.Description, standard)
			spec := *rule.Spec.Effect.Matching
			spec.Policy = append(append([]core.MatchingStep{}, rule.Spec.Effect.Matching.Policy...), step)
			next.Spec.Effect.Matching = &spec
			next, err := store.InsertRuleVersion(ctx, tx, next)
			if err != nil {
				return nil, err
			}
			b := core.Bundle{ID: core.OwnedID(core.LearnedBundleID(tmpl.Name, standard), company), Company: company, Status: core.StatusDraft,
				Description: fmt.Sprintf("The owner paired %s by hand the %s way twice: add %s as the last step of the policy.", tmpl.Name, standard, standard),
				Members:     []core.Member{{Kind: core.KindRule, Name: next.ID, Version: next.Version}},
				Warrant: &core.Warrant{Basis: "client",
					Citations: []string{fmt.Sprintf("paired by hand: event #%d", earlier), fmt.Sprintf("paired by hand: event #%d", eventID)}},
				Question:  fmt.Sprintf("Should %s also match the %s way?", tmpl.Name, standard),
				CreatedBy: "kernel"}
			b, err = store.InsertBundleVersion(ctx, tx, b)
			if err != nil {
				return nil, err
			}
			return &b, nil
		}, nil
	}
	return nil, nil
}

// companysOwn reports whether a door that is Rhea staff's when it changes how
// every company's books are kept acts here on the acting company's own alone
// (KK, 2026-10-22: company rules and Rhea's global rules): approving its own
// rule; approving or rejecting its own bundle, whose rules are all its own and
// whose types, views and doors are new words nobody had — a new name changes
// no other company's books, a new version of one every company shares does;
// a backfill acting for it, which plans and books its own past alone. Then the
// company's right is enough. Anything else, or anything it cannot read, stays
// staff's.
func (x *Executor) companysOwn(ctx context.Context, door string, inputs map[string]any, company string) bool {
	if company == "" {
		return false
	}
	own := func(rule string) bool {
		r, err := x.Store.GetRule(ctx, rule)
		return err == nil && r.Company == company
	}
	switch door {
	case "approve_rule":
		id, _ := inputs["rule_id"].(string)
		return own(id)
	case "approve_bundle", "reject_bundle":
		id, _ := inputs["bundle_id"].(string)
		b, err := x.Store.GetBundle(ctx, id)
		if err != nil || b.Company != company {
			return false
		}
		for _, m := range b.Members {
			if m.Kind == core.KindRule && !own(m.Name) || m.Kind != core.KindRule && m.Version != 1 {
				return false
			}
		}
		return true
	case "approve_backfill":
		return true
	}
	return false
}

// announced are the doors that change Rhea's own definitions when they do not
// act on a company's own: their facts are the installation's, so every
// company hears of the change.
var announced = map[string]bool{"approve_rule": true, "approve_activity": true, "approve_bundle": true, "reject_bundle": true}

// TriggerActivity fires a declared verb: input validation, ref existence,
// the stamped append, the reserved reaction, then the ordinary worklist
// re-evaluation. dedupKey is the caller's idempotency handle (the CLI's file
// hash); the approval reactions set their own.
func (x *Executor) TriggerActivity(ctx context.Context, name string, inputs map[string]any, actor, dedupKey string) (Triggered, error) {
	ctx, release, err := x.serialize(ctx)
	if err != nil {
		return Triggered{}, err
	}
	defer release()
	caller := ctx // the worklist pass after the fact runs under the caller's scope
	act, err := x.Store.GetActiveActivity(ctx, name)
	if err != nil {
		return Triggered{}, err
	}
	// The face says who, the kernel says whether (DIRECTION 2026-10-19,
	// "People and rights"): a signed-in person is checked against the
	// door's Who and their grants in the company they act for, the door
	// signs its "…_by" inputs with them, and a refusal is a recorded fact
	// before it is an error. A signed-in person is always the one checked
	// and the one who signs, whatever actor the caller passed: the actor
	// string is the caller's to write, and "kernel" in a request body once
	// skipped the check (review 2026-10-21, A1). The kernel speaking for
	// itself under a person's request sets the person aside in code first
	// (rights.WithoutPrincipal).
	own := act.Spec.Everywhere && x.companysOwn(ctx, act.Name, inputs, store.CompanyOf(ctx))
	if p := rights.PrincipalOf(ctx); p != nil {
		actor = p.Handle
		gate := act
		if own {
			gate.Spec.Everywhere = false // the company's own alone: its right in the company is enough
		}
		signed, err := rights.Authorize(ctx, x.Store, *p, gate, inputs, store.CompanyOf(ctx), x.today())
		if err != nil {
			var refused *rights.ErrRefused
			if errors.As(err, &refused) {
				if rerr := x.RecordRefusal(ctx, refused); rerr != nil {
					return Triggered{}, rerr
				}
			}
			return Triggered{}, err
		}
		inputs = signed
	}
	_, scoped := store.ScopeOf(ctx)
	if act.Spec.Installation || scoped && announced[act.Name] && !own {
		// the installation's fact, whoever acts for whom: written under the
		// installation's own scope, the one that may write such rows — and a
		// change to Rhea's own definitions, which every company reads (KK,
		// 2026-10-22: a change of Rhea's rules is announced to each company)
		ctx = store.WithInstallation(ctx)
	}
	built, err := act.BuildEvent(inputs, x.today())
	if err != nil {
		return Triggered{}, fmt.Errorf("activity %s: %w", name, err)
	}
	if p := rights.PrincipalOf(ctx); p != nil && act.Spec.Passthrough() {
		if err := x.mayEmit(ctx, *p, act.Name, built); err != nil {
			return Triggered{}, err
		}
	}
	// A document's fact carries its business date in the payload, as a
	// verb's does: the day a file arrived is when an invoice was received,
	// and the readers that turn the file into its typed event read it there.
	if len(built.Files) > 0 {
		if built.Payload, err = withDate(built.Payload, built.OccurredAt); err != nil {
			return Triggered{}, err
		}
	}
	// Ref inputs name objects; the door vouches they exist and are what the
	// declaration says — the same courtesy ref() pays at rule expansion.
	for _, ref := range built.Refs {
		o, err := x.Store.GetObject(ctx, ref.ObjectID)
		if errors.Is(err, store.ErrNotFound) {
			return Triggered{}, fmt.Errorf("activity %s input %s: no object %q", name, ref.Input, ref.ObjectID)
		}
		if err != nil {
			return Triggered{}, err // a store failure is not a missing object
		}
		if o.Type != ref.TargetType {
			return Triggered{}, fmt.Errorf("activity %s input %s: %q is a %s, not a %s", name, ref.Input, ref.ObjectID, o.Type, ref.TargetType)
		}
	}

	// A file's hash is its identity: the same file through the same door is
	// the same fact, unless the caller names the fact otherwise. A key the
	// caller chose names what the caller did, under its door's name, where
	// no reader keys anything: taken as it came, it pre-claimed the key a
	// channel's fact would carry, and the fact was skipped (review
	// 2026-10-21, C2).
	if dedupKey != "" {
		dedupKey = "door/" + act.Name + "/" + dedupKey
	} else if built.Identity != "" {
		dedupKey = act.Name + "/" + built.Identity
	}
	ev := core.Event{
		Kind: core.KindRaw, Type: built.Type, OccurredAt: built.OccurredAt,
		Payload: built.Payload, DedupKey: dedupKey, Actor: actor,
		ActivityName: act.Name, ActivityVersion: act.Version,
		Company: store.CompanyOf(ctx), // the caller's scope: the fact's company
	}
	var out Triggered
	apply, err := x.prepareReaction(ctx, &ev, &out)
	if err != nil {
		return Triggered{}, err
	}

	// An office teaching its clients writes into their books too: the
	// transaction's scope widens to exactly the clients the person was let to
	// teach, still acting for the company it acts for.
	txCtx := ctx
	if set, scoped := store.ScopeOf(ctx); scoped && len(out.Clients) > 0 {
		txCtx = store.ActFor(store.WithScope(ctx, append(append([]string{}, set...), out.Clients...)), store.CompanyOf(ctx))
	}
	tx, err := x.Store.Pool.Begin(txCtx)
	if err != nil {
		return Triggered{}, err
	}
	defer tx.Rollback(ctx)
	// The file is the fact, so the file comes first: its bytes are kept
	// under their hash in the same transaction, and the event naming them
	// cannot exist without them.
	for _, f := range built.Files {
		if _, err := store.PutBlob(ctx, tx, f.Hash, f.MediaType, f.Bytes); err != nil {
			return Triggered{}, err
		}
	}
	if out.EventID, err = store.AppendEvent(ctx, tx, ev); err != nil {
		if store.IsDuplicate(err) {
			return Triggered{}, fmt.Errorf("activity %s: done already — %s is one fact, logged before: %w", name, dedupKey, err)
		}
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

	booked, errs := x.ProcessPending(caller)
	for _, c := range out.Clients { // what waited in a taught client's books may book now
		n, e := x.ProcessPending(store.WithCompany(caller, c))
		booked, errs = booked+n, append(errs, e...)
	}
	out.Booked, out.Errors = booked, append(out.Errors, errs...)
	return out, nil
}

// mayEmit decides a fact a person sends through the open door: a type some
// door or channel declares comes through here only from someone one of them
// admits, as if through it — the open door refused the kernel's own
// namespaces and nothing else, so an accountant, whose door:* reaches it and
// never the people doors, submitted membership.granted and owned another
// company's books, and moved the clock every company reads (review
// 2026-10-21, C1). A fact nothing declares still comes in.
func (x *Executor) mayEmit(ctx context.Context, p rights.Principal, door string, built core.BuiltEvent) error {
	acts, err := x.Store.ActiveActivities(ctx)
	if err != nil {
		return err
	}
	var inputs map[string]any
	_ = json.Unmarshal(built.Payload, &inputs)
	var owners []string
	for _, a := range acts {
		if a.Spec.Passthrough() || !a.Spec.Allows(built.Type) {
			continue
		}
		owners = append(owners, a.Name)
		_, err := rights.Authorize(ctx, x.Store, p, a, inputs, store.CompanyOf(ctx), x.today())
		var refused *rights.ErrRefused
		switch {
		case err == nil:
			return nil
		case !errors.As(err, &refused):
			return err
		}
	}
	if len(owners) == 0 {
		return nil
	}
	refused := &rights.ErrRefused{Person: p.Handle, Door: door, Company: store.CompanyOf(ctx),
		Reason: fmt.Sprintf("%s comes through %s, which no role of yours opens", built.Type, strings.Join(owners, ", "))}
	if err := x.RecordRefusal(ctx, refused); err != nil {
		return err
	}
	return refused
}

// RecordRefusal appends the refusal as a fact under the company it was
// refused in: who, at which door, why — the audit an owner reads.
func (x *Executor) RecordRefusal(ctx context.Context, r *rights.ErrRefused) error {
	payload, _ := json.Marshal(map[string]string{"person": r.Person, "door": r.Door, "company": r.Company, "reason": r.Reason})
	// the kernel's own record, written as the owner: it names the company
	// refused (or none), which the refused person's scope need not reach
	_, err := x.Store.AppendEvent(store.WithOwner(ctx), core.Event{
		Kind: core.KindRaw, Type: core.EventAccessRefused, OccurredAt: x.today(),
		Payload: payload, Actor: "kernel", Company: r.Company,
	})
	return err
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
			Reason     string `json:"reason"`
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
		record := map[string]any{"rule_id": r.ID, "approved_version": r.Version + 1, "approved_by": p.ApprovedBy}
		if p.Reason != "" {
			record["reason"] = p.Reason // the correction's reason: why this reading replaces the last
		}
		approval, _ := json.Marshal(record)
		ev.Payload = approval
		ev.DedupKey = fmt.Sprintf("approve/%s/%d", r.ID, r.Version+1)
		return func(tx store.Querier) error {
			r.Status = core.StatusActive
			r, err = store.InsertRuleVersion(ctx, tx, r)
			out.Rule = &r
			return err
		}, nil

	case "rule.promoted":
		return x.preparePromotion(ctx, ev)

	case "pack.installed":
		return nil, x.checkInstall(ctx, *ev)

	case "pack.switched_off":
		return nil, x.checkSwitchOff(ctx, *ev)

	case "rule.retired":
		var p struct {
			RuleID string `json:"rule_id"`
		}
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return nil, err
		}
		r, err := x.Store.GetRule(ctx, p.RuleID)
		if err != nil {
			return nil, err
		}
		if r.Company == "" || r.Company != ev.Company {
			return nil, fmt.Errorf("rule %s is not the company's own to switch off", p.RuleID)
		}
		if r.Status != core.StatusActive {
			return nil, fmt.Errorf("rule %s is %s — only a rule that books can be switched off", p.RuleID, r.Status)
		}
		ev.DedupKey = fmt.Sprintf("retire/%s/%d", r.ID, r.Version+1)
		return func(tx store.Querier) error {
			r.Status = core.StatusSuperseded
			r, err = store.InsertRuleVersion(ctx, tx, r)
			out.Rule = &r
			return err
		}, nil

	case "rule.approved_for_clients":
		var p struct {
			RuleID     string `json:"rule_id"`
			Clients    string `json:"clients"`
			ApprovedBy string `json:"approved_by"`
		}
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return nil, err
		}
		r, err := x.Store.GetRule(ctx, p.RuleID)
		if err != nil {
			return nil, err
		}
		if r.Company == "" || r.Company != ev.Company || r.Status != core.StatusActive {
			return nil, fmt.Errorf("rule %s is no active rule of the company acted for — an office teaches its clients what one of them was taught", p.RuleID)
		}
		seen, clients := map[string]bool{r.Company: true}, []string{}
		for _, c := range strings.Split(p.Clients, ",") {
			if c = strings.TrimSpace(c); c != "" && !seen[c] {
				seen[c] = true
				clients = append(clients, c)
			}
		}
		if len(clients) == 0 {
			return nil, fmt.Errorf("name the clients to teach, other than %s", r.Company)
		}
		sort.Strings(clients)
		// Whole or nothing: the person approves rules in every client named,
		// or no client is taught.
		if pr := rights.PrincipalOf(ctx); pr != nil {
			for _, c := range clients {
				if _, ok := pr.May(c, "door", "approve_rule", "system"); !ok {
					refused := &rights.ErrRefused{Person: pr.Handle, Door: ev.ActivityName, Company: c, Reason: "no role of yours approves rules here"}
					if err := x.RecordRefusal(ctx, refused); err != nil {
						return nil, err
					}
					return nil, refused
				}
			}
		}
		// A client that switched the lesson off said no: the office does not
		// teach it again over the client's veto.
		base := strings.TrimSuffix(r.ID, "@"+r.Company)
		for _, c := range clients {
			if prev, err := x.Store.GetRule(store.WithCompany(ctx, c), core.OwnedID(base, c)); err == nil && prev.Status == core.StatusSuperseded {
				return nil, fmt.Errorf("%s switched %s off — teaching it again is %s's to ask for", c, prev.ID, c)
			}
		}
		ev.DedupKey = fmt.Sprintf("approve-for/%s/%d/%s", r.ID, r.Version, strings.Join(clients, ","))
		out.Clients = clients
		return func(tx store.Querier) error {
			for _, c := range clients {
				own := r
				own.ID, own.Company, own.Status = core.OwnedID(base, c), c, core.StatusActive
				own, err := store.InsertRuleVersion(ctx, tx, own)
				if err != nil {
					return err
				}
				fact, _ := json.Marshal(map[string]any{"rule_id": own.ID, "approved_version": own.Version,
					"approved_by": p.ApprovedBy, "taught_from": r.ID})
				if _, err := store.AppendEvent(ctx, tx, core.Event{Kind: core.KindRaw, Type: ev.Type, OccurredAt: ev.OccurredAt,
					Payload: fact, Actor: ev.Actor, Company: c, ActivityName: ev.ActivityName, ActivityVersion: ev.ActivityVersion,
					DedupKey: fmt.Sprintf("approve-for/%s/%d", own.ID, own.Version)}); err != nil {
					return err
				}
			}
			return nil
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
		// A rule's type is approved already or comes in this bundle: approved
		// ahead of the bundle bringing its type — the core's parts are drafted
		// together at init — the rule would run over a definition nobody
		// approved yet.
		brings := map[string]bool{}
		for _, m := range b.Members {
			if m.Kind == core.KindObjectType {
				brings[m.Name] = true
			}
		}
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
				if t := r.Spec.Effect.Object.Type; t != "" && !brings[t] {
					if _, err := x.Store.GetObjectType(ctx, t); err != nil {
						return nil, fmt.Errorf("bundle %s: rule %s writes %s, which nothing approved defines yet — approve the bundle that brings it first: %w", b.ID, r.ID, t, err)
					}
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
			Plan       string `json:"plan"`
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
		if p.Plan != "" && p.Plan != plan.Fingerprint {
			return nil, fmt.Errorf("the past moved since the dry run you read (another approval landed) — read it again before approving")
		}
		// The approval names exactly what it promotes: every chain, its
		// rules, and whether it was offered forward out of a closed period.
		approval, _ := json.Marshal(map[string]any{
			"approved_by": p.ApprovedBy, "chains": plan.Chains, "refused": plan.Errors,
		})
		ev.Payload = approval
		// The chains book in the approval's own transaction, as a bundle's
		// members activate: an approval naming chains that never booked,
		// or chains planned on balances of earlier ones that failed, is
		// not a record.
		return func(tx store.Querier) error {
			for _, c := range plan.Chains {
				if err := x.bookTx(ctx, tx, c.root, c.nodes); err != nil {
					return fmt.Errorf("backfill event %d: %w", c.EventID, err)
				}
			}
			out.Backfilled = len(plan.Chains)
			return nil
		}, nil

	case core.EventMatchingPaired:
		// The pairing verb (DIRECTION 2026-10-11, the grid): the kernel runs
		// the matching rule's structure on the person's selection and books
		// the same chain the policy would have — the match, then the rules
		// that explain it — in the transaction that appends the pairing.
		var p struct {
			Matching string   `json:"matching"`
			Trigger  string   `json:"trigger"`
			Items    []string `json:"items"`
		}
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return nil, err
		}
		if p.Matching == "" || p.Trigger == "" || len(p.Items) == 0 {
			return nil, fmt.Errorf("a pairing names the matching rule, the trigger item and the items it settles")
		}
		// A verb's fact carries its business date in the payload, as every
		// document fact does: the rules that explain the match date the
		// settlement by the root's date.
		dated, err := withDate(ev.Payload, ev.OccurredAt)
		if err != nil {
			return nil, err
		}
		ev.Payload = dated
		rules, err := x.Store.ActiveRules(ctx)
		if err != nil {
			return nil, err
		}
		var rule *core.Rule
		for _, r := range forCompany(rules, ev.Company, ev.OccurredAt) { // another company's rule pairs nothing here
			if r.ID == p.Matching {
				rule = &r
			}
		}
		if rule == nil || rule.Spec.Effect.Matching == nil {
			return nil, fmt.Errorf("no active matching rule %q", p.Matching)
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
		matched, err := x.pairSelection(ctx, *rule, p.Trigger, p.Items, lookup, get)
		if err != nil {
			return nil, err
		}
		derived, err := x.Store.EventsByKind(ctx, core.KindDerived)
		if err != nil {
			return nil, err
		}
		matched.Seq = matchesOn(derived, matched.Matching, matched.Trigger.ObjectID)
		// What the host learns: which named standards would have found this
		// pairing over the sides as the policy reads them. After a person
		// has paired the same way twice, the host proposes the standard as a
		// step of the policy — a draft bundle behind the gate (DIRECTION
		// 2026-10-11, the brief; "the host remembers").
		if matched.Standards, err = x.standardsFor(ctx, rule.Spec.Effect.Matching, matched, lookup, get); err != nil {
			return nil, err
		}
		proposal, err := x.learnedProposal(ctx, *rule, ev.Company, matched, derived)
		if err != nil {
			return nil, err
		}
		var payload any
		if err := json.Unmarshal(ev.Payload, &payload); err != nil {
			return nil, err
		}
		seed := chainNode{matched: matched, rule: *rule, cause: -1}
		nodes, err := x.expandChain(ctx, *ev, rules, payload, lookup, get, nil, &seed)
		if err != nil {
			return nil, err
		}
		return func(tx store.Querier) error {
			root := *ev
			root.ID = out.EventID
			if err := x.bookTx(ctx, tx, root, nodes); err != nil {
				return err
			}
			if proposal != nil {
				b, err := proposal(tx, out.EventID)
				if err != nil {
					return err
				}
				out.Proposed = b
			}
			return nil
		}, nil

	case core.EventMatchingUnpaired:
		// The unpairing verb: one match undone by a correcting event the
		// same rules explain. The reversal is seeded into the chain under
		// the match's own rule version — the structure that explains both.
		var p struct {
			MatchEvent int64  `json:"match_event"`
			Reason     string `json:"reason"`
		}
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return nil, err
		}
		orig, err := x.Store.GetEvent(ctx, p.MatchEvent)
		if err != nil {
			return nil, fmt.Errorf("no event #%d to unpair", p.MatchEvent)
		}
		if orig.Type != core.EventObjectsMatched {
			return nil, fmt.Errorf("event #%d is a %s, not a match", p.MatchEvent, orig.Type)
		}
		var m core.MatchedObjects
		if err := json.Unmarshal(orig.Payload, &m); err != nil {
			return nil, err
		}
		if m.ReversalOf != 0 {
			return nil, fmt.Errorf("event #%d is itself a reversal; pair again instead", p.MatchEvent)
		}
		if ev.Payload, err = withDate(ev.Payload, ev.OccurredAt); err != nil {
			return nil, err
		}
		derived, err := x.Store.EventsByKind(ctx, core.KindDerived)
		if err != nil {
			return nil, err
		}
		for _, d := range derived {
			if d.Type != core.EventObjectsMatched {
				continue
			}
			var r core.MatchedObjects
			if json.Unmarshal(d.Payload, &r) == nil && r.ReversalOf == p.MatchEvent {
				return nil, fmt.Errorf("match #%d was already undone by event #%d", p.MatchEvent, d.ID)
			}
		}
		rule, err := x.Store.GetRuleVersion(ctx, orig.RuleID, orig.RuleVersion)
		if err != nil {
			return nil, err
		}
		rules, err := x.Store.ActiveRules(ctx)
		if err != nil {
			return nil, err
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
		reversal, err := reverseMatch(m, p.MatchEvent, p.Reason, get)
		if err != nil {
			return nil, fmt.Errorf("unpair #%d: %w", p.MatchEvent, err)
		}
		reversal.Seq = matchesOn(derived, m.Matching, m.Trigger.ObjectID)
		var payload any
		if err := json.Unmarshal(ev.Payload, &payload); err != nil {
			return nil, err
		}
		seed := chainNode{matched: reversal, rule: rule, cause: -1}
		nodes, err := x.expandChain(ctx, *ev, rules, payload, lookup, get, nil, &seed)
		if err != nil {
			return nil, fmt.Errorf("unpair #%d: %w", p.MatchEvent, err)
		}
		return func(tx store.Querier) error {
			root := *ev
			root.ID = out.EventID
			return x.bookTx(ctx, tx, root, nodes)
		}, nil

	case core.EventDocumentReadingConfirmed:
		// A person is a reader too (DIRECTION 2026-10-10): confirming a
		// doubtful reading makes the proposed event the person's own —
		// the typed event, with the person as actor, caused by the
		// proposal (which thereby leaves the worklist), and the document
		// read. Nothing of the model's doubt is lost: the proposal stays.
		var p struct {
			Proposal int64 `json:"proposal"`
		}
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return nil, err
		}
		prop, err := x.Store.GetEvent(ctx, p.Proposal)
		if err != nil {
			return nil, fmt.Errorf("no event #%d to confirm", p.Proposal)
		}
		if prop.Type != "document.reading_proposed" {
			return nil, fmt.Errorf("event #%d is a %s, not a proposed reading", p.Proposal, prop.Type)
		}
		var proposed struct {
			Document  string          `json:"document"`
			Channel   string          `json:"channel"`
			EventType string          `json:"event_type"`
			Payload   json.RawMessage `json:"payload"`
			DedupKey  string          `json:"dedup_key"` // the reader's identity for the fact
		}
		if err := json.Unmarshal(prop.Payload, &proposed); err != nil {
			return nil, err
		}
		if proposed.EventType == "" || len(proposed.Payload) == 0 {
			return nil, fmt.Errorf("event #%d proposes nothing to confirm", p.Proposal)
		}
		// Only a channel proposes, and only what it declares: a "proposal"
		// written through the open door named objects.matched, and its
		// confirmation booked a raw match (review 2026-10-21, C1).
		if ch, err := x.Store.GetActiveActivity(ctx, prop.ActivityName); err != nil || ch.Spec.Channel == nil ||
			core.ReservedEventType(proposed.EventType) || !ch.Spec.Allows(proposed.EventType) {
			return nil, fmt.Errorf("event #%d was not proposed by a channel declaring %s — a reading is confirmed only as its channel's own fact", p.Proposal, proposed.EventType)
		}
		raws, err := x.Store.EventsByKind(ctx, core.KindRaw)
		if err != nil {
			return nil, err
		}
		for _, r := range raws {
			if r.CauseEventID != nil && *r.CauseEventID == p.Proposal {
				return nil, fmt.Errorf("proposal #%d was already confirmed by event #%d", p.Proposal, r.ID)
			}
		}
		var fact map[string]any
		if err := json.Unmarshal(proposed.Payload, &fact); err != nil {
			return nil, err
		}
		fact["reader"] = "person"
		fact["confirmed_from"] = p.Proposal
		payload, err := json.Marshal(fact)
		if err != nil {
			return nil, err
		}
		cause := p.Proposal
		ev.Type, ev.Payload, ev.CauseEventID = proposed.EventType, payload, &cause
		ev.OccurredAt = prop.OccurredAt
		// the fact's identity is the reader's: the same invoice read again
		// and booked meanwhile is the same fact, not a second one
		ev.DedupKey = fmt.Sprintf("confirm_reading/%d", p.Proposal)
		if proposed.DedupKey != "" {
			ev.DedupKey = proposed.DedupKey
		}
		read, _ := json.Marshal(map[string]any{"document": proposed.Document, "channel": "confirm_reading", "event_type": proposed.EventType})
		readEvent := core.Event{Kind: core.KindRaw, Type: "document.read", OccurredAt: prop.OccurredAt, Payload: read,
			DedupKey: fmt.Sprintf("confirm_reading/read/%d", p.Proposal), Actor: ev.Actor, CauseEventID: &cause,
			ActivityName: ev.ActivityName, ActivityVersion: ev.ActivityVersion}
		return func(tx store.Querier) error {
			_, err := store.AppendEvent(ctx, tx, readEvent)
			return err
		}, nil

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
	return fmt.Errorf("%s %s v%d belongs to bundle %s, which is %s — bundle it anew or redraft it", m.Kind, m.Name, m.Version, id, status)
}

// ApproveRule records a human approval through the approve_rule door: the
// approval becomes a raw event (invariant 2), the kernel reaction writes the
// new active version row in the same transaction, and the pending worklist
// is re-evaluated under the enlarged rule set. Failures from that
// re-evaluation are worklist conditions, not approval failures — they come
// back as data, separate from the hard error.
func (x *Executor) ApproveRule(ctx context.Context, ruleID, approvedBy, businessDate string) (core.Rule, int, []error, error) {
	return x.ApproveRuleWithReason(ctx, ruleID, approvedBy, businessDate, "")
}

// ApproveRuleWithReason approves with the correction's reason on the record:
// why this version replaces the reading before it. The author's ask shows
// the superseded reading with that reason, so the overruled rule never
// returns (DIRECTION, "The host remembers").
func (x *Executor) ApproveRuleWithReason(ctx context.Context, ruleID, approvedBy, businessDate, reason string) (core.Rule, int, []error, error) {
	inputs := map[string]any{"rule_id": ruleID, "approved_by": approvedBy, "occurred_at": businessDate}
	if reason != "" {
		inputs["reason"] = reason
	}
	t, err := x.TriggerActivity(ctx, "approve_rule", inputs, approvedBy, "")
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
