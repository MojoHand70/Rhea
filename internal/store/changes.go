package store

import (
	"context"
	_ "embed"

	"rhea/internal/core"
)

//go:embed queries/rheas_changes.sql
var rheasChangesSQL string

// RheasChanges lists every version of a rule of Rhea's that became active on
// or after since (YYYY-MM-DD), newest first: what Rhea changed for every
// company, announced to each (KK, 2026-10-22).
func (s *Store) RheasChanges(ctx context.Context, since string) ([]core.Rule, error) {
	rows, err := s.Pool.Query(ctx, rheasChangesSQL, since)
	if err != nil {
		return nil, err
	}
	return scanRules(rows)
}

//go:embed queries/rheas_rules.sql
var rheasRulesSQL string

// RheasRules is the set of Rhea's own rule ids — the rules no company owns,
// which only the house approves.
func (s *Store) RheasRules(ctx context.Context) (map[string]bool, error) {
	rows, err := s.Pool.Query(ctx, rheasRulesSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}
