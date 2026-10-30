package shell

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/network"
	"rhea/internal/store"
)

// StoreBundleDraft lands an agent's bundle as drafts: types and views at
// their next version, rules and verbs as new draft versions, and the bundle
// over all of them. Validation already ran (agent.BundleDraft.Validate);
// versions are the kernel's to assign, and a taken bundle id gets a suffix —
// a redraft is a new bundle, naming in `after` the rejected one it answers,
// and every bundle keeps the question it answers.
//
// Drafted acting for a company, the bundle and its rules are the company's
// own (KK, 2026-10-22: company rules and Rhea's global rules): their ids end
// "@company", a rule named after one of Rhea's replaces Rhea's for the
// company alone, and the question is seen only where the company is. Acting
// for no company — the installation's own scope or the owner — they are
// Rhea's; a scope over several companies acting for none drafts nothing.
func StoreBundleDraft(ctx context.Context, st *store.Store, k *network.Knowledge, bd agent.BundleDraft, createdBy, effectiveFrom, question, after string) (core.Bundle, error) {
	company, known, err := draftingFor(ctx, st)
	if err != nil {
		return core.Bundle{}, err
	}
	// A proposal resting on the network is verified here, and its support is
	// counted by Rhea from what she learned — never taken from the proposer.
	if bd.Warrant != nil && bd.Warrant.Basis == "network" {
		w := *bd.Warrant
		if k == nil {
			return core.Bundle{}, fmt.Errorf("bundle %s claims network knowledge, but no network is connected", bd.BundleID)
		}
		for _, d := range bd.Rules {
			if sup, ok := k.SupportFor(d.Spec); ok && (w.Support == nil || sup.Count > w.Support.Count) {
				w.Support = &sup
			}
		}
		if w.Support == nil {
			return core.Bundle{}, fmt.Errorf("bundle %s claims network knowledge, but none of its rules is an answer Rhea has learned", bd.BundleID)
		}
		bd.Warrant = &w
	}
	// Whole or not at all: types, views, rules, activities and the bundle
	// land in one transaction, so a rule refused third leaves no orphan
	// drafts in append-only tables, bundled by nothing.
	tx, err := st.Pool.Begin(ctx)
	if err != nil {
		return core.Bundle{}, err
	}
	defer tx.Rollback(ctx)
	var members []core.Member
	for _, t := range bd.ObjectTypes {
		v, err := st.NextObjectTypeVersion(ctx, t.Name)
		if err != nil {
			return core.Bundle{}, err
		}
		t.Version = v
		if err := store.InsertObjectTypeRow(ctx, tx, t, core.StatusDraft); err != nil {
			return core.Bundle{}, fmt.Errorf("type %s: %w", t.Name, err)
		}
		members = append(members, core.Member{Kind: core.KindObjectType, Name: t.Name, Version: v})
	}
	owned := map[string]string{} // a view's id as drafted → as named for the company
	for _, view := range bd.ViewDefs {
		// a company's view is its own, named for it as its rules are (KK,
		// 2026-10-27: saved reports are company own)
		drafted := view.ID
		view.ID, view.Company = core.OwnedID(view.ID, company), company
		owned[drafted] = view.ID
		v, err := st.NextViewDefVersion(ctx, view.ID)
		if err != nil {
			return core.Bundle{}, err
		}
		view.Version = v
		if err := store.InsertViewDefRow(ctx, tx, view, core.StatusDraft); err != nil {
			return core.Bundle{}, fmt.Errorf("view %s: %w", view.ID, err)
		}
		members = append(members, core.Member{Kind: core.KindViewDef, Name: view.ID, Version: v})
	}
	for _, d := range bd.Rules {
		id, replaces := ownRule(known, company, d.RuleID)
		r, err := store.InsertRuleVersion(ctx, tx, core.Rule{
			ID: id, Status: core.StatusDraft, Priority: d.Priority,
			EffectiveFrom: effectiveFrom, CreatedBy: createdBy,
			Description: d.Description, Spec: d.Spec,
			Company: company, Replaces: replaces,
		})
		if err != nil {
			return core.Bundle{}, fmt.Errorf("rule %s: %w", d.RuleID, err)
		}
		members = append(members, core.Member{Kind: core.KindRule, Name: r.ID, Version: r.Version})
	}
	for _, ad := range bd.Activities {
		a, err := store.InsertActivityVersion(ctx, tx, core.Activity{
			Name: ad.Name, Status: core.StatusDraft, Domain: ad.Domain,
			Description: ad.Description, Spec: ad.Spec, CreatedBy: createdBy,
		})
		if err != nil {
			return core.Bundle{}, fmt.Errorf("activity %s: %w", ad.Name, err)
		}
		members = append(members, core.Member{Kind: core.KindActivity, Name: a.Name, Version: a.Version})
	}
	for _, a := range bd.Apps {
		// a company's App is its own, as its views are; its tabs name the
		// bundle's views as the company's, or views already on the record
		a.ID, a.Company = core.OwnedID(a.ID, company), company
		a.Tabs = slices.Clone(a.Tabs)
		for i, t := range a.Tabs {
			if t.View == "" {
				continue
			}
			if own, ok := owned[t.View]; ok {
				a.Tabs[i].View = own
				continue
			}
			if ok, err := st.HasViewDef(ctx, t.View); err != nil {
				return core.Bundle{}, err
			} else if !ok {
				return core.Bundle{}, fmt.Errorf("app %s: its tab names view %s, which neither this bundle nor the record has", a.ID, t.View)
			}
		}
		v, err := st.NextAppVersion(ctx, a.ID)
		if err != nil {
			return core.Bundle{}, err
		}
		a.Version = v
		if err := store.InsertAppRow(ctx, tx, a, core.StatusDraft); err != nil {
			return core.Bundle{}, fmt.Errorf("app %s: %w", a.ID, err)
		}
		members = append(members, core.Member{Kind: core.KindApp, Name: a.ID, Version: v})
	}
	id := core.OwnedID(bd.BundleID, company)
	for n := 2; ; n++ {
		_, err := st.GetBundle(ctx, id)
		if errors.Is(err, pgx.ErrNoRows) {
			break // free
		}
		if err != nil {
			return core.Bundle{}, err // a store failure is not a free id
		}
		id = core.OwnedID(fmt.Sprintf("%s-%d", strings.TrimSuffix(bd.BundleID, "@"+company), n), company)
	}
	// The members are this transaction's own fresh drafts: no other bundle
	// can hold them, so the bundle's guard has nothing to check here.
	b := core.Bundle{ID: id, Status: core.StatusDraft, Description: bd.Description,
		Members: members, CreatedBy: createdBy, Warrant: bd.Warrant, Question: question, After: after, Company: company}
	if err := b.Validate(); err != nil {
		return core.Bundle{}, err
	}
	if b.Warrant != nil {
		if err := b.Warrant.ValidateStored(); err != nil {
			return core.Bundle{}, fmt.Errorf("bundle %s: %w", b.ID, err)
		}
	}
	b, err = store.InsertBundleVersion(ctx, tx, b)
	if err != nil {
		return core.Bundle{}, err
	}
	return b, tx.Commit(ctx)
}

// draftingFor is the company a draft is made for — the one acted for, empty
// for Rhea's own (the installation's scope, the owner) — and, for a company,
// the rules its drafts are named against: its own and Rhea's, by id. A desk
// over several companies acting for none drafts nothing: whose would it be?
func draftingFor(ctx context.Context, st *store.Store) (string, map[string]core.Rule, error) {
	company := store.CompanyOf(ctx)
	if set, scoped := store.ScopeOf(ctx); scoped && company == "" && len(set) > 0 {
		return "", nil, fmt.Errorf("act for one of your companies to draft its rules")
	}
	known := map[string]core.Rule{}
	if company != "" {
		rules, err := st.LatestRules(ctx)
		if err != nil {
			return "", nil, err
		}
		for _, r := range rules {
			known[r.ID] = r
		}
	}
	return company, known, nil
}

// ownRule names a rule drafted for company (KK, 2026-10-22): the company's
// own id, and the rule of Rhea's it replaces — a redraft of the company's own
// keeps what that replaces, a rule named after one of Rhea's replaces it for
// the company alone. Drafted for no company, the name is Rhea's as drafted.
func ownRule(known map[string]core.Rule, company, drafted string) (id, replaces string) {
	id = core.OwnedID(drafted, company)
	if company == "" {
		return id, ""
	}
	if own, ok := known[id]; ok {
		return id, own.Replaces
	}
	if rheas, ok := known[strings.TrimSuffix(drafted, "@"+company)]; ok && rheas.Company == "" {
		return id, rheas.ID
	}
	return id, ""
}
