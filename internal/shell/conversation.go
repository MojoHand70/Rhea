package shell

import (
	"context"
	"encoding/json"
	"fmt"

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/store"
)

// conversationDepth caps how far back a redraft reads: the last refusals
// are what it answers; older ones are on the record for anyone to read.
const conversationDepth = 3

// Conversation gathers what was already said about one question: the chain
// of rejected bundles behind `after`, oldest first, each with its reason as
// the reject_bundle door recorded it and the proposal as the agent shaped
// it — so a redraft answers the refusal instead of starting over. It also
// returns the question the chain began with, for a re-ask that brings no
// new words. Only a rejected bundle can be redrafted.
func Conversation(ctx context.Context, st *store.Store, after string) ([]agent.Rejection, string, error) {
	var chain []core.Bundle
	for id := after; id != "" && len(chain) < conversationDepth; {
		b, err := st.GetBundle(ctx, id)
		if err != nil {
			return nil, "", err
		}
		chain = append(chain, b)
		id = b.After
	}
	if len(chain) == 0 {
		return nil, "", nil
	}
	question := chain[len(chain)-1].Question
	var out []agent.Rejection
	for i := len(chain) - 1; i >= 0; i-- {
		b := chain[i]
		rj, ok, err := st.BundleRejection(ctx, b.ID)
		if err != nil {
			return nil, "", err
		}
		if !ok {
			return nil, "", fmt.Errorf("bundle %s was not rejected — only a rejected bundle is redrafted", b.ID)
		}
		proposal, err := proposalOf(ctx, st, b)
		if err != nil {
			return nil, "", err
		}
		out = append(out, agent.Rejection{Proposal: proposal, Reason: rj.Reason, By: rj.By})
	}
	return out, question, nil
}

// proposalOf reads a stored bundle back in the shape the agent proposes it,
// from its members' exact versions — what was refused, as it was said.
func proposalOf(ctx context.Context, st *store.Store, b core.Bundle) (json.RawMessage, error) {
	bd := agent.BundleDraft{BundleID: b.ID, Description: b.Description}
	if b.Warrant != nil {
		w := *b.Warrant
		w.Support = nil // counted by Rhea, never part of a proposal
		bd.Warrant = &w
	}
	for _, m := range b.Members {
		switch m.Kind {
		case core.KindRule:
			r, err := st.GetRuleVersion(ctx, m.Name, m.Version)
			if err != nil {
				return nil, err
			}
			bd.Rules = append(bd.Rules, agent.Draft{RuleID: r.ID, Description: r.Description, Priority: r.Priority, Spec: r.Spec})
		case core.KindObjectType:
			t, _, err := st.GetObjectTypeVersion(ctx, m.Name, m.Version)
			if err != nil {
				return nil, err
			}
			t.Version = 0 // the kernel's to assign, as in a proposal
			bd.ObjectTypes = append(bd.ObjectTypes, t)
		case core.KindViewDef:
			v, _, err := st.GetViewDefVersion(ctx, m.Name, m.Version)
			if err != nil {
				return nil, err
			}
			v.Version = 0
			bd.ViewDefs = append(bd.ViewDefs, v)
		case core.KindActivity:
			a, err := st.GetActivity(ctx, m.Name)
			if err != nil {
				return nil, err
			}
			bd.Activities = append(bd.Activities, agent.ActivityDraft{Name: a.Name, Domain: a.Domain, Description: a.Description, Spec: a.Spec})
		}
	}
	return json.Marshal(bd)
}
