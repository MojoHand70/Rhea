// Package pack loads market packs (SPEC M3): bundles of object types, views,
// rules and master-data events that teach a running kernel a market — a data
// file, not a code fork. A pack installs no behavior: its rules arrive as
// drafts behind the same human approval gate as everything else, and its
// events wait in the worklist until those rules are approved — approving the
// pack's rules in the shell IS the installation. Loading is idempotent:
// types, views and events already present are skipped, and a rule id the
// system already knows is left alone.
package pack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgconn"

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
	Events      []Event           `json:"events,omitempty"`
}

// Rule is a pack-shipped rule. It always lands as a draft; status is not a
// pack's to set.
type Rule struct {
	RuleID        string        `json:"rule_id"`
	Priority      int           `json:"priority"`
	EffectiveFrom string        `json:"effective_from"`
	Description   string        `json:"description"`
	Spec          core.RuleSpec `json:"spec"`
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
type Summary struct {
	Pack    string `json:"pack"`
	Types   int    `json:"types"`
	Views   int    `json:"views"`
	Rules   int    `json:"rules"`
	Events  int    `json:"events"`
	Skipped int    `json:"skipped"`
}

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

	for _, t := range m.ObjectTypes {
		switch err := s.InsertObjectType(ctx, t); {
		case err == nil:
			sum.Types++
		case isDuplicate(err):
			sum.Skipped++
		default:
			return sum, fmt.Errorf("object type %s: %w", t.Name, err)
		}
	}
	for _, v := range m.ViewDefs {
		switch err := s.InsertViewDef(ctx, v); {
		case err == nil:
			sum.Views++
		case isDuplicate(err):
			sum.Skipped++
		default:
			return sum, fmt.Errorf("view def %s: %w", v.ID, err)
		}
	}

	for _, r := range m.Rules {
		if _, err := s.GetRule(ctx, r.RuleID); err == nil {
			sum.Skipped++ // the id is known; its versions are not a pack's to touch
			continue
		}
		var target *core.ObjectType
		if r.Spec.Effect.Object.Type != "" {
			t, err := s.GetObjectType(ctx, r.Spec.Effect.Object.Type)
			if err != nil {
				return sum, fmt.Errorf("rule %s: %w", r.RuleID, err)
			}
			target = &t
		}
		if err := r.Spec.Validate(target); err != nil {
			return sum, fmt.Errorf("rule %s: %w", r.RuleID, err)
		}
		if _, err := s.InsertRuleVersion(ctx, core.Rule{
			ID: r.RuleID, Status: core.StatusDraft, Priority: r.Priority,
			EffectiveFrom: r.EffectiveFrom, CreatedBy: "pack:" + m.Pack,
			Description: r.Description, Spec: r.Spec,
		}); err != nil {
			return sum, fmt.Errorf("rule %s: %w", r.RuleID, err)
		}
		sum.Rules++
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
		case isDuplicate(err):
			sum.Skipped++
		default:
			return sum, fmt.Errorf("event %s (%s): %w", ev.DedupKey, ev.EventType, err)
		}
	}
	return sum, nil
}

func isDuplicate(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
