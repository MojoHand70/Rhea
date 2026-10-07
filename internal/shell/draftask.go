package shell

import (
	"context"
	"slices"

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/store"
)

// referenceCap bounds the master data shown per type: the agent needs to see
// what a ref() can resolve against, not the ledger.
const referenceCap = 30

// DraftAsk gathers the read-only context the agent drafts against: the whole
// type catalog, the active rules, and master data for every type some field
// references (exactly the set a ref() or a posting's account can resolve
// against). The agent has no store access; this is what it is shown.
func DraftAsk(ctx context.Context, st *store.Store, intent string, sample core.Event, hint string) (agent.Ask, error) {
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
	return agent.Ask{Intent: intent, Sample: sample, Types: types, Hint: hint,
		Rules: rules, Reference: reference}, nil
}
