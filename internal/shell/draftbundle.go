package shell

import (
	"context"
	"fmt"

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
func StoreBundleDraft(ctx context.Context, st *store.Store, k *network.Knowledge, bd agent.BundleDraft, createdBy, effectiveFrom, question, after string) (core.Bundle, error) {
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
	var members []core.Member
	for _, t := range bd.ObjectTypes {
		v, err := st.NextObjectTypeVersion(ctx, t.Name)
		if err != nil {
			return core.Bundle{}, err
		}
		t.Version = v
		if err := store.InsertObjectTypeRow(ctx, st.Pool, t, core.StatusDraft); err != nil {
			return core.Bundle{}, fmt.Errorf("type %s: %w", t.Name, err)
		}
		members = append(members, core.Member{Kind: core.KindObjectType, Name: t.Name, Version: v})
	}
	for _, view := range bd.ViewDefs {
		v, err := st.NextViewDefVersion(ctx, view.ID)
		if err != nil {
			return core.Bundle{}, err
		}
		view.Version = v
		if err := store.InsertViewDefRow(ctx, st.Pool, view, core.StatusDraft); err != nil {
			return core.Bundle{}, fmt.Errorf("view %s: %w", view.ID, err)
		}
		members = append(members, core.Member{Kind: core.KindViewDef, Name: view.ID, Version: v})
	}
	for _, d := range bd.Rules {
		r, err := st.InsertRuleVersion(ctx, core.Rule{
			ID: d.RuleID, Status: core.StatusDraft, Priority: d.Priority,
			EffectiveFrom: effectiveFrom, CreatedBy: createdBy,
			Description: d.Description, Spec: d.Spec,
		})
		if err != nil {
			return core.Bundle{}, fmt.Errorf("rule %s: %w", d.RuleID, err)
		}
		members = append(members, core.Member{Kind: core.KindRule, Name: r.ID, Version: r.Version})
	}
	for _, ad := range bd.Activities {
		a, err := st.InsertActivityVersion(ctx, core.Activity{
			Name: ad.Name, Status: core.StatusDraft, Domain: ad.Domain,
			Description: ad.Description, Spec: ad.Spec, CreatedBy: createdBy,
		})
		if err != nil {
			return core.Bundle{}, fmt.Errorf("activity %s: %w", ad.Name, err)
		}
		members = append(members, core.Member{Kind: core.KindActivity, Name: a.Name, Version: a.Version})
	}
	id := bd.BundleID
	for n := 2; ; n++ {
		if _, err := st.GetBundle(ctx, id); err != nil {
			break
		}
		id = fmt.Sprintf("%s-%d", bd.BundleID, n)
	}
	return st.InsertBundle(ctx, core.Bundle{ID: id, Description: bd.Description,
		Members: members, CreatedBy: createdBy, Warrant: bd.Warrant, Question: question, After: after})
}
