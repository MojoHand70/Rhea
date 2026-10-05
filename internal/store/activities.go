package store

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"rhea/internal/core"
)

// todayUTC is the business date of acts the kernel performs on its own
// clock, like seeding its verbs.
func todayUTC() string { return time.Now().UTC().Format("2006-01-02") }

//go:embed queries/active_activities.sql
var activeActivitiesSQL string

// InsertActivityVersion appends the next version row for the activity and
// returns it — the same append-only discipline as rules; a status transition
// is a new row. Package-level so approve_activity's reaction can flip the
// version inside the trigger's transaction. Declared activities may not emit
// into the kernel namespaces: rule.* and activity.* are the system's own
// verbs, spoken only by the builtins seeded at init.
func InsertActivityVersion(ctx context.Context, q Querier, a core.Activity) (core.Activity, error) {
	if err := a.Validate(); err != nil {
		return a, err
	}
	if !strings.HasPrefix(a.Spec.Emits, "=") && core.ReservedEventType(a.Spec.Emits) {
		return a, fmt.Errorf("activity %q: %q is a kernel namespace — declared activities do not speak it", a.Name, a.Spec.Emits)
	}
	return insertActivityVersion(ctx, q, a)
}

func (s *Store) InsertActivityVersion(ctx context.Context, a core.Activity) (core.Activity, error) {
	return InsertActivityVersion(ctx, s.Pool, a)
}

// insertActivityVersion is the write itself, shared with the seeder — which
// is kernel code declaring the kernel's own doors, reserved namespace and all.
func insertActivityVersion(ctx context.Context, q Querier, a core.Activity) (core.Activity, error) {
	spec, err := json.Marshal(a.Spec)
	if err != nil {
		return a, err
	}
	err = q.QueryRow(ctx, `
		INSERT INTO activity (name, version, status, domain, description, spec, created_by)
		VALUES ($1, COALESCE((SELECT MAX(version) FROM activity WHERE name = $1), 0) + 1,
		        $2, $3, $4, $5, $6)
		RETURNING version, created_at`,
		a.Name, a.Status, a.Domain, a.Description, spec, a.CreatedBy,
	).Scan(&a.Version, &a.CreatedAt)
	return a, err
}

func scanActivities(rows pgx.Rows) ([]core.Activity, error) {
	defer rows.Close()
	var out []core.Activity
	for rows.Next() {
		var a core.Activity
		var spec []byte
		if err := rows.Scan(&a.Name, &a.Version, &a.Status, &a.Domain,
			&a.Description, &spec, &a.CreatedBy, &a.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(spec, &a.Spec); err != nil {
			return nil, fmt.Errorf("activity %s v%d spec: %w", a.Name, a.Version, err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

const activityCols = `name, version, status, domain, description, spec, created_by, created_at`

// LatestActivities returns the newest version of every activity.
func (s *Store) LatestActivities(ctx context.Context) ([]core.Activity, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT DISTINCT ON (name) `+activityCols+`
		FROM activity ORDER BY name, version DESC`)
	if err != nil {
		return nil, err
	}
	return scanActivities(rows)
}

// ActiveActivities returns the offered verb set: the newest active version
// of every activity.
func (s *Store) ActiveActivities(ctx context.Context) ([]core.Activity, error) {
	rows, err := s.Pool.Query(ctx, activeActivitiesSQL)
	if err != nil {
		return nil, err
	}
	return scanActivities(rows)
}

// GetActivity returns the newest version of one activity, any status.
func (s *Store) GetActivity(ctx context.Context, name string) (core.Activity, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT `+activityCols+` FROM activity WHERE name = $1 ORDER BY version DESC LIMIT 1`, name)
	if err != nil {
		return core.Activity{}, err
	}
	acts, err := scanActivities(rows)
	if err != nil {
		return core.Activity{}, err
	}
	if len(acts) == 0 {
		return core.Activity{}, fmt.Errorf("activity %q not found", name)
	}
	return acts[0], nil
}

// GetActiveActivity returns the offered version of one activity — the door a
// trigger goes through. A draft or superseded verb is not a door.
func (s *Store) GetActiveActivity(ctx context.Context, name string) (core.Activity, error) {
	acts, err := s.ActiveActivities(ctx)
	if err != nil {
		return core.Activity{}, err
	}
	for _, a := range acts {
		if a.Name == name {
			return a, nil
		}
	}
	return core.Activity{}, fmt.Errorf("activity %q is not offered — no active version", name)
}

// ActivityEventCounts returns how many events each door has emitted — the
// verb's usage, the way CountObjectsByType is a type's.
func (s *Store) ActivityEventCounts(ctx context.Context) (map[string]int, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT activity_name, COUNT(*) FROM event
		WHERE activity_name IS NOT NULL GROUP BY activity_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var name string
		var n int
		if err := rows.Scan(&name, &n); err != nil {
			return nil, err
		}
		out[name] = n
	}
	return out, rows.Err()
}

// seedBuiltinActivities declares the kernel's own verbs: version 1, active,
// each activation recorded as an activity.approved event stamped with the
// door that approves activities — the gate applied to itself, the bootstrap
// fixed point recorded honestly. Idempotent: existing names are left alone,
// and the events' dedup keys make a crashed seed resumable.
func (s *Store) seedBuiltinActivities(ctx context.Context) error {
	for _, a := range core.BuiltinActivities() {
		var exists bool
		if err := s.Pool.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM activity WHERE name = $1)`, a.Name).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}
		approval, err := json.Marshal(map[string]any{
			"activity": a.Name, "approved_version": 1, "approved_by": "init",
		})
		if err != nil {
			return err
		}
		if _, err := s.AppendEvent(ctx, core.Event{
			Kind: core.KindRaw, Type: "activity.approved", OccurredAt: todayUTC(),
			Payload: approval, DedupKey: fmt.Sprintf("activity-approve/%s/1", a.Name),
			Actor: "kernel", ActivityName: "approve_activity", ActivityVersion: 1,
		}); err != nil && !IsDuplicate(err) {
			return fmt.Errorf("seed %s: %w", a.Name, err)
		}
		a.Status = core.StatusActive
		a.CreatedBy = "kernel"
		if _, err := insertActivityVersion(ctx, s.Pool, a); err != nil {
			return fmt.Errorf("seed %s: %w", a.Name, err)
		}
	}
	return nil
}
