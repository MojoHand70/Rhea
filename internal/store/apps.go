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
	var company *string
	if a.Company != "" {
		company = &a.Company
	}
	_, err = q.Exec(ctx, `INSERT INTO app (app_id, version, status, spec, company) VALUES ($1, $2, $3, $4, $5)`,
		a.ID, a.Version, status, spec, company)
	return err
}

// NextAppVersion is the version a new draft of an App takes: one past every
// row of any status, as for a view.
func (s *Store) NextAppVersion(ctx context.Context, id string) (int, error) {
	var v int
	err := s.Pool.QueryRow(ctx, `SELECT COALESCE(MAX(version), 0) + 1 FROM app WHERE app_id = $1`, id).Scan(&v)
	return v, err
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

// LatestApp returns an App's newest row, of any status, with that status:
// what a removal writes above, and whether the App shows — the active row
// of a version answers before its draft.
func (s *Store) LatestApp(ctx context.Context, id string) (core.App, string, error) {
	var a core.App
	var spec []byte
	var status string
	if err := s.Pool.QueryRow(ctx, latestAppSQL, id).Scan(&spec, &status); err != nil {
		return a, "", fmt.Errorf("app %s: %w", id, err)
	}
	return a, status, json.Unmarshal(spec, &a)
}

// LatestApps returns the newest active version of every App, by id — of the
// Apps whose very latest version is not superseded: a draft above the
// running version does not suspend it; a removal does (2026-10-30).
func (s *Store) LatestApps(ctx context.Context) ([]core.App, error) {
	rows, err := s.Pool.Query(ctx, latestAppsSQL)
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

// HasViewDef says whether a view is on the record, in any version and state:
// an App of one pack may name a view another pack has drafted or approved.
func (s *Store) HasViewDef(ctx context.Context, id string) (bool, error) {
	var ok bool
	err := s.Pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM view_def WHERE view_id = $1)`, id).Scan(&ok)
	return ok, err
}
