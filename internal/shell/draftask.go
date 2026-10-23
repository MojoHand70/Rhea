package shell

import (
	"context"
	"encoding/json"
	"slices"

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/network"
	"rhea/internal/store"
)

// referenceCap bounds the master data shown per type: the agent needs to see
// what a ref() can resolve against, not the ledger.
const referenceCap = 30

// DraftAsk gathers the read-only context the agent drafts against: the whole
// type catalog, the active rules, and master data for every type some field
// references (exactly the set a ref() or a posting's account can resolve
// against). The agent has no store access; this is what it is shown.
func DraftAsk(ctx context.Context, st *store.Store, k *network.Knowledge, intent string, sample core.Event, hint string) (agent.Ask, error) {
	types, err := st.ListObjectTypes(ctx)
	if err != nil {
		return agent.Ask{}, err
	}
	rules, err := st.ActiveRules(ctx)
	if err != nil {
		return agent.Ask{}, err
	}
	// the author drafts for the company against what it is offered: a pack
	// it has not installed is not shown (DIRECTION 2026-10-24, step 2)
	offer, err := exec.Offered(ctx, st, store.CompanyOf(ctx), "")
	if err != nil {
		return agent.Ask{}, err
	}
	types = slices.DeleteFunc(types, func(t core.ObjectType) bool { return !offer.Type(t.Name) })
	rules = slices.DeleteFunc(rules, func(r core.Rule) bool { return !offer.Rule(r) })
	var targets []string
	for _, t := range types {
		for _, f := range t.Fields {
			if ref, ok := core.RefTarget(f.Type); ok && !slices.Contains(targets, ref) {
				targets = append(targets, ref)
			}
		}
	}
	reference := map[string][]map[string]any{}
	for _, name := range targets {
		objs, err := st.ObjectsByType(ctx, name)
		if err != nil {
			return agent.Ask{}, err
		}
		for i, o := range objs {
			if i == referenceCap {
				break
			}
			reference[name] = append(reference[name], o.State)
		}
	}
	residue, err := Residue(ctx, st)
	if err != nil {
		return agent.Ask{}, err
	}
	corrections, err := Corrections(ctx, st, rules)
	if err != nil {
		return agent.Ask{}, err
	}
	return agent.Ask{Intent: intent, Sample: sample, Types: types, Hint: hint,
		Rules: rules, Reference: reference, Residue: residue, Priors: priors(k, sample, residue),
		Corrections: corrections}, nil
}

// Corrections lists, for every active rule, the readings its owner replaced:
// each approved version that a later approval superseded, with the later
// approval's reason — read from the rule.approved events, the record of
// every activation. The author never proposes an overruled reading again.
func Corrections(ctx context.Context, st *store.Store, active []core.Rule) ([]agent.Correction, error) {
	raws, err := st.EventsByKind(ctx, core.KindRaw)
	if err != nil {
		return nil, err
	}
	type approval struct {
		version int
		by      string
		reason  string
	}
	approvals := map[string][]approval{}
	for _, ev := range raws {
		if ev.Type != "rule.approved" {
			continue
		}
		var p struct {
			RuleID  string `json:"rule_id"`
			Version int    `json:"approved_version"`
			By      string `json:"approved_by"`
			Reason  string `json:"reason"`
		}
		if json.Unmarshal(ev.Payload, &p) != nil || p.RuleID == "" {
			continue
		}
		approvals[p.RuleID] = append(approvals[p.RuleID], approval{p.Version, p.By, p.Reason})
	}
	var out []agent.Correction
	for _, r := range active {
		as := approvals[r.ID]
		for i := 0; i+1 < len(as); i++ {
			old, err := st.GetRuleVersion(ctx, r.ID, as[i].version)
			if err != nil {
				continue
			}
			out = append(out, agent.Correction{RuleID: r.ID, Version: as[i].version, Description: old.Description,
				ReplacedBy: as[i+1].version, By: as[i+1].by, Reason: as[i+1].reason})
		}
	}
	return out, nil
}

// priors turns what the network learned into the agent's evidence: answers
// to questions about the events in front of it, with their real counts.
func priors(k *network.Knowledge, sample core.Event, residue []agent.Cluster) []agent.Prior {
	if k == nil {
		return nil
	}
	var types []string
	if sample.Type != "" {
		types = append(types, sample.Type)
	}
	for _, c := range residue {
		types = append(types, c.EventType)
	}
	var out []agent.Prior
	for _, a := range k.Priors(types) {
		rule, _ := json.Marshal(a.Shape.Spec)
		out = append(out, agent.Prior{Question: a.Shape.Key, Count: a.Count, Of: a.Of, Rule: rule})
	}
	return out
}

// Residue clusters the worklist by event type, in first-seen order: the
// shape of what is still unexplained, one sample each. Time passing is not
// residue (an uneventful day is normal).
func Residue(ctx context.Context, st *store.Store) ([]agent.Cluster, error) {
	events, err := st.UnmatchedRawEvents(ctx)
	if err != nil {
		return nil, err
	}
	var out []agent.Cluster
	at := map[string]int{}
	for _, e := range events {
		if core.TimeEventType(e.Type) {
			continue
		}
		if i, ok := at[e.Type]; ok {
			out[i].Count++
			continue
		}
		at[e.Type] = len(out)
		out = append(out, agent.Cluster{EventType: e.Type, Count: 1, Sample: e.Payload})
	}
	return out, nil
}
