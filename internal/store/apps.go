package store

import (
	"context"
	"encoding/json"
	"fmt"

	"rhea/internal/core"
)

// --- apps -----------------------------------------------------------------

// InsertAppRow appends one (app_id, version, status) row: a draft from the
// loader, the active row of the same version from its bundle's approval.
func InsertAppRow(ctx context.Context, q Querier, a core.App, status string) error {
	spec, err := json.Marshal(a)
	if err != nil {
		return err
	}
	_, err = q.Exec(ctx, `INSERT INTO app (app_id, version, status, spec) VALUES ($1, $2, $3, $4)`,
		a.ID, a.Version, status, spec)
	return err
}

// GetAppVersion returns one version of an App and its lifecycle status,
// resolved like GetViewDefVersion: an approved version answers active.
func (s *Store) GetAppVersion(ctx context.Context, id string, version int) (core.App, string, error) {
	var a core.App
	var spec []byte
	var status string
	err := s.Pool.QueryRow(ctx, `
		SELECT spec, status FROM app WHERE app_id = $1 AND version = $2
		ORDER BY CASE status WHEN 'draft' THEN 1 ELSE 0 END LIMIT 1`,
		id, version).Scan(&spec, &status)
	if err != nil {
		return a, "", fmt.Errorf("app %s v%d: %w", id, version, err)
	}
	if err := json.Unmarshal(spec, &a); err != nil {
		return a, "", fmt.Errorf("app %s v%d: %w", id, version, err)
	}
	return a, status, nil
}

// LatestApps returns the newest active version of every App, by id.
func (s *Store) LatestApps(ctx context.Context) ([]core.App, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT DISTINCT ON (app_id) spec FROM app
		WHERE status = 'active' ORDER BY app_id, version DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []core.App
	for rows.Next() {
		var spec []byte
		if err := rows.Scan(&spec); err != nil {
			return nil, err
		}
		var a core.App
		if err := json.Unmarshal(spec, &a); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
