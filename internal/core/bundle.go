package core

import (
	"fmt"
	"slices"
	"time"
)

// Bundle is the scope of one approval (DIRECTION, implementation by
// interview): a set of draft definitions that answer one question together —
// a type, the rules that explain events into it, the view that shows it —
// activated whole or not at all (KK, 2026-10-08: no partial approval). The
// members stay ordinary versioned definitions with their own provenance; the
// bundle adds only the approval's scope. A pack install is one bundle.
// Versioned like everything: draft → active on approval, draft → superseded
// on rejection, the reason in the rejection event.
type Bundle struct {
	ID          string   `json:"bundle_id"`
	Version     int      `json:"version"`
	Status      string   `json:"status"` // draft | active | superseded
	Description string   `json:"description"`
	Members     []Member `json:"members"`
	// Warrant says why this is the proposal: where the knowledge comes from
	// (DIRECTION, "Rhea suggests standards"). Nil on hand-made bundles.
	Warrant   *Warrant  `json:"warrant,omitempty"`
	// The conversation, on the record: Question is the plain-language ask
	// this bundle answers; After names the rejected bundle it redrafts.
	Question  string    `json:"question,omitempty"`
	After     string    `json:"after,omitempty"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// Warrant is a suggestion's provenance, as provenance is a fact's: its
// basis, the citations behind it, and the scope it holds in. Support — how
// many approved explanations across the network agree — is Rhea's to count
// from what she has actually learned; whoever proposes may never supply it.
type Warrant struct {
	Basis     string   `json:"basis"`               // see WarrantBases
	Citations []string `json:"citations,omitempty"` // statute, standard, textbook, pack
	Scope     string   `json:"scope,omitempty"`     // market, industry, size: "PL, wholesale"
	Support   *Support `json:"support,omitempty"`   // counted by Rhea, never proposed
}

// Support is a real count over approved rules: Count of Of installations in
// Population explain this shape this way.
type Support struct {
	Count      int    `json:"count"`
	Of         int    `json:"of"`
	Population string `json:"population"`
}

// WarrantBases are the kinds of knowledge a suggestion may rest on. "model"
// is the agent's general training (books, the web) — honest about being
// unverified; "network" is learned consensus and comes only with Support.
var WarrantBases = []string{"statute", "standard", "practice", "pack", "model", "client", "network"}

// ValidateProposed checks a warrant as a proposer wrote it: a known basis,
// and no support figures — those are counted, not claimed. A proposer may
// claim the "network" basis; Rhea verifies the claim against what she has
// learned and counts the support herself before the bundle is stored.
func (w Warrant) ValidateProposed() error {
	if !slices.Contains(WarrantBases, w.Basis) {
		return fmt.Errorf("warrant basis %q is not one of %v", w.Basis, WarrantBases)
	}
	if w.Support != nil {
		return fmt.Errorf("support figures are counted by Rhea from approved rules, never proposed")
	}
	if (w.Basis == "statute" || w.Basis == "standard") && len(w.Citations) == 0 {
		return fmt.Errorf("a %s warrant needs its citation", w.Basis)
	}
	return nil
}

// ValidateStored checks a warrant as it is stored: a network warrant must
// carry the support Rhea counted — a claim she could not verify never lands.
func (w Warrant) ValidateStored() error {
	if !slices.Contains(WarrantBases, w.Basis) {
		return fmt.Errorf("warrant basis %q is not one of %v", w.Basis, WarrantBases)
	}
	if w.Basis == "network" && w.Support == nil {
		return fmt.Errorf("a network warrant needs the support Rhea counted")
	}
	return nil
}

// Member points at one draft definition row.
type Member struct {
	Kind    string `json:"kind"` // rule | activity | object_type | view_def
	Name    string `json:"name"` // rule_id, activity name, type name, view_id
	Version int    `json:"version"`
}

// Definition kinds a bundle can carry: the four versioned vocabularies.
const (
	KindRule       = "rule"
	KindActivity   = "activity"
	KindObjectType = "object_type"
	KindViewDef    = "view_def"
)

// Validate checks a bundle's shape: known kinds, no member twice, at least
// one member. Whether each member is a draft awaiting approval is the
// store's to judge.
func (b Bundle) Validate() error {
	if b.ID == "" {
		return fmt.Errorf("bundle needs an id")
	}
	if len(b.Members) == 0 {
		return fmt.Errorf("bundle %s has no members", b.ID)
	}
	seen := map[[2]string]bool{}
	for _, m := range b.Members {
		if !slices.Contains([]string{KindRule, KindActivity, KindObjectType, KindViewDef}, m.Kind) {
			return fmt.Errorf("bundle %s: unknown member kind %q", b.ID, m.Kind)
		}
		if m.Name == "" || m.Version < 1 {
			return fmt.Errorf("bundle %s: member %s needs a name and a version", b.ID, m.Kind)
		}
		k := [2]string{m.Kind, m.Name}
		if seen[k] {
			return fmt.Errorf("bundle %s: %s %s appears twice", b.ID, m.Kind, m.Name)
		}
		seen[k] = true
	}
	return nil
}
