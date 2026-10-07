package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"rhea/internal/core"
)

// --- bundles: the scope of one approval -----------------------------------

// InsertBundle records a draft bundle. Every member must be a draft awaiting
// approval and belong to no other bundle: one approval scope per draft, or
// "no partial approval" could be dodged through a second bundle.
func (s *Store) InsertBundle(ctx context.Context, b core.Bundle) (core.Bundle, error) {
	b.Status = core.StatusDraft
	if err := b.Validate(); err != nil {
		return b, err
	}
	if _, err := s.GetBundle(ctx, b.ID); err == nil {
		return b, fmt.Errorf("bundle %s already exists — a redraft is a new bundle", b.ID)
	}
	for _, m := range b.Members {
		state, err := s.MemberStatus(ctx, m)
		if err != nil {
			return b, fmt.Errorf("bundle %s: %w", b.ID, err)
		}
		if state != core.StatusDraft {
			return b, fmt.Errorf("bundle %s: %s %s v%d is %s, only drafts can join a bundle", b.ID, m.Kind, m.Name, m.Version, state)
		}
		if other, _, ok, err := s.BundleOf(ctx, m); err != nil {
			return b, err
		} else if ok {
			return b, fmt.Errorf("bundle %s: %s %s v%d already belongs to bundle %s", b.ID, m.Kind, m.Name, m.Version, other)
		}
	}
	return InsertBundleVersion(ctx, s.Pool, b)
}

// InsertBundleVersion appends the bundle's next version row.
func InsertBundleVersion(ctx context.Context, q Querier, b core.Bundle) (core.Bundle, error) {
	members, err := json.Marshal(b.Members)
	if err != nil {
		return b, err
	}
	err = q.QueryRow(ctx, `
		INSERT INTO bundle (bundle_id, version, status, description, members, created_by)
		VALUES ($1, COALESCE((SELECT MAX(version) FROM bundle WHERE bundle_id = $1), 0) + 1,
		        $2, $3, $4, $5)
		RETURNING version, created_at`,
		b.ID, b.Status, b.Description, members, b.CreatedBy,
	).Scan(&b.Version, &b.CreatedAt)
	return b, err
}

const bundleCols = `bundle_id, version, status, description, members, created_by, created_at`

func scanBundle(row pgx.Row) (core.Bundle, error) {
	var b core.Bundle
	var members []byte
	if err := row.Scan(&b.ID, &b.Version, &b.Status, &b.Description, &members, &b.CreatedBy, &b.CreatedAt); err != nil {
		return b, err
	}
	return b, json.Unmarshal(members, &b.Members)
}

// GetBundle returns a bundle's latest version.
func (s *Store) GetBundle(ctx context.Context, id string) (core.Bundle, error) {
	b, err := scanBundle(s.Pool.QueryRow(ctx,
		`SELECT `+bundleCols+` FROM bundle WHERE bundle_id = $1 ORDER BY version DESC LIMIT 1`, id))
	if err != nil {
		return b, fmt.Errorf("bundle %q: %w", id, err)
	}
	return b, nil
}

// LatestBundles returns the latest version of every bundle, newest first.
func (s *Store) LatestBundles(ctx context.Context) ([]core.Bundle, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT * FROM (SELECT DISTINCT ON (bundle_id) `+bundleCols+`
		               FROM bundle ORDER BY bundle_id, version DESC) b
		ORDER BY created_at DESC, bundle_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []core.Bundle
	for rows.Next() {
		b, err := scanBundle(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// BundleOf reports the bundle a draft version belongs to, if any, with that
// bundle's latest status. A draft answers in exactly one bundle for life:
// the single-definition approval doors refuse it (no partial approval), and
// after a rejection it stays a rejected draft — a redraft is a new version.
func (s *Store) BundleOf(ctx context.Context, m core.Member) (id, status string, ok bool, err error) {
	probe, _ := json.Marshal([]core.Member{m})
	err = s.Pool.QueryRow(ctx, `
		SELECT bundle_id, status FROM (SELECT DISTINCT ON (bundle_id) bundle_id, status, members
		                               FROM bundle ORDER BY bundle_id, version DESC) b
		WHERE members @> $1::jsonb LIMIT 1`, probe).Scan(&id, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", false, nil
	}
	return id, status, err == nil, err
}

// MemberStatus is a member's lifecycle state. Rules and activities mint a
// version per status change, so a member is a draft only while its version is
// the latest and a draft; types and views keep their version and answer by
// their rows.
func (s *Store) MemberStatus(ctx context.Context, m core.Member) (string, error) {
	switch m.Kind {
	case core.KindRule:
		r, err := s.GetRule(ctx, m.Name)
		if err != nil {
			return "", err
		}
		if r.Version != m.Version {
			return fmt.Sprintf("not the latest version (v%d is)", r.Version), nil
		}
		return r.Status, nil
	case core.KindActivity:
		a, err := s.GetActivity(ctx, m.Name)
		if err != nil {
			return "", err
		}
		if a.Version != m.Version {
			return fmt.Sprintf("not the latest version (v%d is)", a.Version), nil
		}
		return a.Status, nil
	case core.KindObjectType:
		_, status, err := s.GetObjectTypeVersion(ctx, m.Name, m.Version)
		return status, err
	case core.KindViewDef:
		_, status, err := s.GetViewDefVersion(ctx, m.Name, m.Version)
		return status, err
	}
	return "", fmt.Errorf("unknown member kind %q", m.Kind)
}

// DraftObjectTypes returns the bundle's draft types, keyed by name: the
// overlay a bundle's dry run validates against.
func (s *Store) DraftObjectTypes(ctx context.Context, b core.Bundle) (map[string]core.ObjectType, error) {
	out := map[string]core.ObjectType{}
	for _, m := range b.Members {
		if m.Kind != core.KindObjectType {
			continue
		}
		t, _, err := s.GetObjectTypeVersion(ctx, m.Name, m.Version)
		if err != nil {
			return nil, err
		}
		out[t.Name] = t
	}
	return out, nil
}
