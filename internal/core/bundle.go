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
	ID          string    `json:"bundle_id"`
	Version     int       `json:"version"`
	Status      string    `json:"status"` // draft | active | superseded
	Description string    `json:"description"`
	Members     []Member  `json:"members"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
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
