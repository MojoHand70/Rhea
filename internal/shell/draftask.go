package shell

import (
	"context"
	"encoding/json"
	"slices"

	"rhea/internal/agent"
	"rhea/internal/core"
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
	return agent.Ask{Intent: intent, Sample: sample, Types: types, Hint: hint,
		Rules: rules, Reference: reference, Residue: residue, Priors: priors(k, sample, residue)}, nil
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
