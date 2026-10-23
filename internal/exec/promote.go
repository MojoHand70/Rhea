package exec

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"rhea/internal/core"
	"rhea/internal/network"
	"rhea/internal/rights"
	"rhea/internal/store"
)

// A company's own reading becomes Rhea's when many companies hold it (KK,
// 2026-10-22): the same rule body held, active, by at least PromoteCompanies
// companies kept by at least PromoteKeepers offices or installations — a
// company no office keeps is its own keeper — and only Rhea's staff promote
// it. Below the floor a pattern can still be one firm's own business, so no
// one sees it.
const (
	PromoteCompanies = 10
	PromoteKeepers   = 3
)

// Candidate is a reading companies hold that has reached the floor: its
// body, the rule of Rhea's the companies' copies replace when they agree,
// and how many companies and keepers hold it. It names no company.
type Candidate struct {
	Shape     string        `json:"shape"`
	Spec      core.RuleSpec `json:"spec"`
	Replaces  string        `json:"replaces,omitempty"`
	Companies int           `json:"companies"`
	Keepers   int           `json:"keepers"`
	// Of is how many companies read the same fact their own way at all,
	// this way or another: the population the count is out of.
	Of int `json:"of"`
}

// shapeOf keys a rule body: the same match and effect, whoever holds it.
func shapeOf(spec core.RuleSpec) (string, error) {
	b, err := json.Marshal(spec)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:8]), nil
}

// Promotable counts every company's own active rules by body, across the
// installation as the kernel sees it, and returns the readings at or above
// the floor, most held first. Simulated companies are not counted.
func (x *Executor) Promotable(ctx context.Context) ([]Candidate, error) {
	owner := store.WithOwner(ctx)
	rules, err := x.Store.ActiveRules(owner)
	if err != nil {
		return nil, err
	}
	keepers, err := rights.Keepers(owner, x.Store, x.today())
	if err != nil {
		return nil, err
	}
	type group struct {
		c         Candidate
		companies map[string]bool
		keepers   map[string]bool
		replaces  map[string]bool
	}
	groups := map[string]*group{}
	readers := map[string]map[string]bool{} // event type → companies reading it their own way
	for _, r := range rules {
		if r.Company == "" || strings.HasPrefix(r.Company, network.SyntheticPrefix) {
			continue
		}
		shape, err := shapeOf(r.Spec)
		if err != nil {
			return nil, err
		}
		g := groups[shape]
		if g == nil {
			g = &group{c: Candidate{Shape: shape, Spec: r.Spec}, companies: map[string]bool{}, keepers: map[string]bool{}, replaces: map[string]bool{}}
			groups[shape] = g
		}
		keeper := keepers[r.Company]
		if keeper == "" {
			keeper = r.Company
		}
		g.companies[r.Company], g.keepers[keeper], g.replaces[r.Replaces] = true, true, true
		et := r.Spec.Match.EventType
		if readers[et] == nil {
			readers[et] = map[string]bool{}
		}
		readers[et][r.Company] = true
	}
	var out []Candidate
	for _, g := range groups {
		if len(g.companies) < PromoteCompanies || len(g.keepers) < PromoteKeepers {
			continue
		}
		g.c.Companies, g.c.Keepers, g.c.Of = len(g.companies), len(g.keepers), len(readers[g.c.Spec.Match.EventType])
		if len(g.replaces) == 1 {
			for r := range g.replaces {
				g.c.Replaces = r
			}
		}
		out = append(out, g.c)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Companies != out[j].Companies {
			return out[i].Companies > out[j].Companies
		}
		return out[i].Shape < out[j].Shape
	})
	return out, nil
}

// preparePromotion is promote_rule's reaction: the pattern must be at the
// floor now, and the rule written is Rhea's — a new version of the rule the
// companies' copies replace, or the rule named — booking every company from
// the promotion's day, its warrant the count and the auditor's opinion.
func (x *Executor) preparePromotion(ctx context.Context, ev *core.Event) (func(tx store.Querier) error, error) {
	var p struct {
		Shape      string `json:"shape"`
		RuleID     string `json:"rule_id"`
		Opinion    string `json:"opinion"`
		ApprovedBy string `json:"approved_by"`
	}
	if err := json.Unmarshal(ev.Payload, &p); err != nil {
		return nil, err
	}
	cands, err := x.Promotable(ctx)
	if err != nil {
		return nil, err
	}
	var c *Candidate
	for i := range cands {
		if cands[i].Shape == p.Shape {
			c = &cands[i]
		}
	}
	if c == nil {
		return nil, fmt.Errorf("no reading %s is held by %d companies kept by %d offices — nothing to promote", p.Shape, PromoteCompanies, PromoteKeepers)
	}
	id := p.RuleID
	if id == "" {
		id = c.Replaces
	}
	if id == "" {
		return nil, fmt.Errorf("the companies' rules replace no one rule of Rhea's — name the rule to write")
	}
	priority := 100
	if prev, err := x.Store.GetRule(store.WithOwner(ctx), id); err == nil {
		if prev.Company != "" {
			return nil, fmt.Errorf("%s is a company's rule, not Rhea's", id)
		}
		priority = prev.Priority
	}
	citations := []string{fmt.Sprintf("held by %d of the %d companies that read %s their own way, kept by %d offices or installations",
		c.Companies, c.Of, c.Spec.Match.EventType, c.Keepers)}
	if p.Opinion != "" {
		citations = append(citations, "opinion: "+p.Opinion)
	}
	r := core.Rule{ID: id, Status: core.StatusActive, Priority: priority, EffectiveFrom: ev.OccurredAt, CreatedBy: "staff:" + p.ApprovedBy,
		Description: fmt.Sprintf("A reading %d companies held as their own, promoted by Rhea's staff", c.Companies), Spec: c.Spec,
		Warrant: &core.Warrant{Basis: "network", Citations: citations,
			Support: &core.Support{Count: c.Companies, Of: c.Of, Population: fmt.Sprintf("companies reading %s their own way", c.Spec.Match.EventType)}}}
	record, _ := json.Marshal(map[string]any{"shape": c.Shape, "rule_id": id, "companies": c.Companies, "keepers": c.Keepers,
		"of": c.Of, "opinion": p.Opinion, "approved_by": p.ApprovedBy})
	ev.Payload = record
	ev.DedupKey = fmt.Sprintf("promote/%s/%s", c.Shape, id)
	return func(tx store.Querier) error {
		_, err := store.InsertRuleVersion(ctx, tx, r)
		return err
	}, nil
}
