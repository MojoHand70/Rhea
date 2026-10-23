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
