// Package pack loads market packs (SPEC M3): bundles of object types, views,
// rules and master-data events that teach a running kernel a market — a data
// file, not a code fork. A pack installs nothing: its types, views, rules and
// activities arrive as drafts gathered into one bundle behind the same human
// approval gate as everything else (KK, 2026-10-08: everything drafts), and
// its events wait in the worklist until the bundle is approved — approving
// the pack's bundle IS the installation. Loading is idempotent: definitions
// and events already present are skipped. A newer pack upgrades its own
// definitions — a changed rule or activity whose latest version is the
// pack's becomes the next draft version in the upgrade bundle — and leaves
// alone what a person or the agent evolved locally since, reporting it as a
// conflict (DECISIONS 2026-10-15).
package pack

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"rhea/internal/core"
	"rhea/internal/store"
)

// Manifest is a pack file: one JSON document holding everything the market
// needs. Requires names object types the pack builds on (the finance base's
// account and posting, say) without shipping them — base definitions belong
// to the base, and a second market proves the boundary by reusing them.
type Manifest struct {
	Pack        string            `json:"pack"`
	Version     int               `json:"version"`
	Description string            `json:"description"`
	Requires    []string          `json:"requires,omitempty"`
	ObjectTypes []core.ObjectType `json:"object_types,omitempty"`
	ViewDefs    []core.ViewDef    `json:"view_defs,omitempty"`
	Rules       []Rule            `json:"rules,omitempty"`
	Activities  []Activity        `json:"activities,omitempty"`
	Events      []Event           `json:"events,omitempty"`
	// Warrant: why the pack proposes what it does. Defaults to the pack's
	// own authority ("pack" basis, its description as citation).
	Warrant *core.Warrant `json:"warrant,omitempty"`
	// Stage, Disclaimer and History say what the pack is to whoever reads it
	// (KK 2026-10-18): "initial" — prepared by AI, not yet reviewed by the
	// profession, its version read as 0.<version>; the disclaimer, in the
	// market's language, says so first; the history, newest first, shows
	// the pack developing. TestEveryPackSaysWhatItIs holds every pack to it.
	Stage      string         `json:"stage,omitempty"`
	Disclaimer string         `json:"disclaimer,omitempty"`
	History    []HistoryEntry `json:"history,omitempty"`
}

// HistoryEntry is what changed in one version of a pack, or in a run of
// early versions ("1–18"), in the market's language.
type HistoryEntry struct {
	Versions string `json:"versions"`
	Date     string `json:"date,omitempty"`
	Changes  string `json:"changes"`
}

// Stages a pack may be at: prepared by AI and not yet reviewed, or reviewed
// by the profession.
var Stages = []string{"initial", "reviewed"}

// Rule is a pack-shipped rule. It always lands as a draft; status is not a
// pack's to set.
type Rule struct {
	RuleID        string        `json:"rule_id"`
	Priority      int           `json:"priority"`
	EffectiveFrom string        `json:"effective_from"`
	Description   string        `json:"description"`
	Spec          core.RuleSpec `json:"spec"`
	// Warrant: the rule's own source, beside the pack's. Travels with the
	// rule's versions; a changed citation is a changed rule.
	Warrant *core.Warrant `json:"warrant,omitempty"`
}

// Activity is a pack-shipped verb. Like a rule it always lands as a draft:
// what humans can do is approved by humans, never installed.
type Activity struct {
	Name        string            `json:"name"`
	Domain      string            `json:"domain"`
	Description string            `json:"description"`
	Spec        core.ActivitySpec `json:"spec"`
}

// Event is pack-shipped master data (a chart of accounts, VAT rates) as raw
// events. Every event must carry its own dedup key — that key is what makes
// reloading a pack a no-op instead of a double-booking.
type Event struct {
	EventType  string          `json:"event_type"`
	OccurredAt string          `json:"occurred_at"`
	DedupKey   string          `json:"dedup_key"`
	Payload    json.RawMessage `json:"payload"`
}

// Summary says what a load did; Skipped counts everything already present.
// Upgraded counts the pack's own definitions it drafted a new version of
// (a pack upgrade, 2026-10-15); Conflicts names the ones it left alone
// because a person or the agent evolved them locally since.
type Summary struct {
	Pack       string   `json:"pack"`
	Types      int      `json:"types"`
	Views      int      `json:"views"`
	Rules      int      `json:"rules"`
	Activities int      `json:"activities"`
	Events     int      `json:"events"`
	Skipped    int      `json:"skipped"`
	Upgraded   int      `json:"upgraded"`
	Conflicts  []string `json:"conflicts,omitempty"`
	// Bundle is the approval scope holding every draft this load added;
	// empty when the load added none (a reload).
	Bundle string `json:"bundle,omitempty"`
}

// A pack upgrade (DECISIONS 2026-10-15): a known rule or activity is the
// pack's to move forward when its latest version is the pack's own — then
// a changed definition is drafted as the next version, into the upgrade
// bundle behind the gate like everything else. A definition a person or
// the agent evolved since is not the pack's to touch: it is reported as a
// conflict and left alone, for a person to resolve. An unchanged one is
// already present. Object types and views were always keyed by version,
// so their new versions upgrade by the same path they install.
func packOwns(createdBy, pack string) bool { return createdBy == "pack:"+pack }

func sameRule(have core.Rule, want Rule) bool {
	a, _ := json.Marshal(have.Spec)
	b, _ := json.Marshal(want.Spec)
	wa, _ := json.Marshal(have.Warrant)
	wb, _ := json.Marshal(want.Warrant)
	return have.Priority == want.Priority && have.EffectiveFrom == want.EffectiveFrom &&
		have.Description == want.Description && string(a) == string(b) && string(wa) == string(wb)
}

func sameActivity(have core.Activity, want Activity) bool {
	a, _ := json.Marshal(have.Spec)
	b, _ := json.Marshal(want.Spec)
	return have.Domain == want.Domain && have.Description == want.Description && string(a) == string(b)
}

// BundleID names a pack version's bundle.
func BundleID(pack string, version int) string { return fmt.Sprintf("pack-%s-v%d", pack, version) }

// Load reads a pack file and loads it into a running kernel. The actor is
// whoever runs the load — pack provenance lives in the dedup keys and the
// rules' created_by. Order matters only once: types land before rules so
// rule validation can see its targets.
func Load(ctx context.Context, s *store.Store, path, actor string) (Summary, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Summary{}, err
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return Summary{}, err
	}
	if m.Pack == "" {
		return Summary{}, fmt.Errorf("pack file needs a pack name")
	}
	sum := Summary{Pack: m.Pack}

	for _, name := range m.Requires {
		if _, err := s.GetObjectType(ctx, name); err != nil {
			return sum, fmt.Errorf("pack %s requires object type %q — load its base definitions first", m.Pack, name)
		}
	}

	var members []core.Member
	packTypes := map[string]core.ObjectType{}
	for _, t := range m.ObjectTypes {
		packTypes[t.Name] = t
		if _, _, err := s.GetObjectTypeVersion(ctx, t.Name, t.Version); err == nil {
			sum.Skipped++
			continue
		}
		if err := store.InsertObjectTypeRow(ctx, s.Pool, t, core.StatusDraft); err != nil {
			return sum, fmt.Errorf("object type %s: %w", t.Name, err)
		}
		members = append(members, core.Member{Kind: core.KindObjectType, Name: t.Name, Version: t.Version})
		sum.Types++
	}
	for _, v := range m.ViewDefs {
		if _, _, err := s.GetViewDefVersion(ctx, v.ID, v.Version); err == nil {
			sum.Skipped++
			continue
		}
		if err := store.InsertViewDefRow(ctx, s.Pool, v, core.StatusDraft); err != nil {
			return sum, fmt.Errorf("view def %s: %w", v.ID, err)
		}
		members = append(members, core.Member{Kind: core.KindViewDef, Name: v.ID, Version: v.Version})
		sum.Views++
	}

	for _, r := range m.Rules {
		upgrade := false
		if have, err := s.GetRule(ctx, r.RuleID); err == nil {
			switch {
			case sameRule(have, r):
				sum.Skipped++ // already present, this version or an earlier one with the same content
				continue
			case !packOwns(have.CreatedBy, m.Pack):
				// evolved locally since: not the pack's to touch
				sum.Conflicts = append(sum.Conflicts, fmt.Sprintf("rule %s v%d (by %s)", have.ID, have.Version, have.CreatedBy))
				continue
			}
			upgrade = true
		}
		var target *core.ObjectType
		if name := r.Spec.Effect.Object.Type; name != "" {
			// The pack's own draft types first: they activate with the rule.
			t, ok := packTypes[name]
			if !ok {
				var err error
				if t, err = s.GetObjectType(ctx, name); err != nil {
					return sum, fmt.Errorf("rule %s: %w", r.RuleID, err)
				}
			}
			target = &t
		}
		if err := r.Spec.Validate(target); err != nil {
			return sum, fmt.Errorf("rule %s: %w", r.RuleID, err)
		}
		// The typed gate: formulas checked against the pack's own draft
		// types first, then the kernel's.
		if err := r.Spec.Check(func(name string) (core.ObjectType, bool) {
			if t, ok := packTypes[name]; ok {
				return t, true
			}
			t, err := s.GetObjectType(ctx, name)
			return t, err == nil
		}); err != nil {
			return sum, fmt.Errorf("rule %s: %w", r.RuleID, err)
		}
		if r.Warrant != nil {
			if err := r.Warrant.ValidateProposed(); err != nil {
				return sum, fmt.Errorf("rule %s warrant: %w", r.RuleID, err)
			}
		}
		rule, err := s.InsertRuleVersion(ctx, core.Rule{
			ID: r.RuleID, Status: core.StatusDraft, Priority: r.Priority,
			EffectiveFrom: r.EffectiveFrom, CreatedBy: "pack:" + m.Pack,
			Description: r.Description, Spec: r.Spec, Warrant: r.Warrant,
		})
		if err != nil {
			return sum, fmt.Errorf("rule %s: %w", r.RuleID, err)
		}
		members = append(members, core.Member{Kind: core.KindRule, Name: rule.ID, Version: rule.Version})
		if upgrade {
			sum.Upgraded++
		} else {
			sum.Rules++
		}
	}

	for _, a := range m.Activities {
		upgrade := false
		if have, err := s.GetActivity(ctx, a.Name); err == nil {
			switch {
			case sameActivity(have, a):
				sum.Skipped++
				continue
			case !packOwns(have.CreatedBy, m.Pack):
				sum.Conflicts = append(sum.Conflicts, fmt.Sprintf("activity %s v%d (by %s)", have.Name, have.Version, have.CreatedBy))
				continue
			}
			upgrade = true
		}
		act, err := s.InsertActivityVersion(ctx, core.Activity{
			Name: a.Name, Status: core.StatusDraft, Domain: a.Domain,
			Description: a.Description, Spec: a.Spec, CreatedBy: "pack:" + m.Pack,
		})
		if err != nil {
			return sum, fmt.Errorf("activity %s: %w", a.Name, err)
		}
		members = append(members, core.Member{Kind: core.KindActivity, Name: act.Name, Version: act.Version})
		if upgrade {
			sum.Upgraded++
		} else {
			sum.Activities++
		}
	}

	if len(members) > 0 {
		warrant := m.Warrant
		if warrant != nil {
			if err := warrant.ValidateProposed(); err != nil {
				return sum, fmt.Errorf("pack %s warrant: %w", m.Pack, err)
			}
		}
		if warrant == nil {
			warrant = &core.Warrant{Basis: "pack", Citations: []string{fmt.Sprintf("pack %s v%d", m.Pack, m.Version)}}
		}
		b, err := s.InsertBundle(ctx, core.Bundle{
			ID: BundleID(m.Pack, m.Version), CreatedBy: "pack:" + m.Pack, Members: members,
			Description: fmt.Sprintf("%s pack %s v%d: %s", map[bool]string{false: "Install", true: "Upgrade to"}[sum.Upgraded > 0], m.Pack, m.Version, m.Description),
			Warrant:     warrant,
		})
		if err != nil {
			return sum, err
		}
		sum.Bundle = b.ID
	}

	for _, ev := range m.Events {
		if ev.DedupKey == "" {
			return sum, fmt.Errorf("pack event %q has no dedup key — idempotent loading is the contract", ev.EventType)
		}
		_, err := s.AppendEvent(ctx, core.Event{
			Kind: core.KindRaw, Type: ev.EventType, OccurredAt: ev.OccurredAt,
			Payload: ev.Payload, DedupKey: ev.DedupKey, Actor: actor,
		})
		switch {
		case err == nil:
			sum.Events++
		case store.IsDuplicate(err):
			sum.Skipped++
		default:
			return sum, fmt.Errorf("event %s (%s): %w", ev.DedupKey, ev.EventType, err)
		}
	}
	return sum, nil
}
