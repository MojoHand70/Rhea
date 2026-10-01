// Package store is the Postgres system of record: the append-only event log,
// the versioned definition stores (rule, object_type, view_def) and the
// rebuildable object projection cache.
package store

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"rhea/internal/core"
)

//go:embed schema.sql
var schemaSQL string

//go:embed queries/unmatched.sql
var unmatchedSQL string

//go:embed queries/latest_rules.sql
var latestRulesSQL string

type Store struct {
	Pool *pgxpool.Pool
}

// DSN returns the connection string, overridable via RHEA_PG_DSN.
func DSN() string {
	if dsn := os.Getenv("RHEA_PG_DSN"); dsn != "" {
		return dsn
	}
	return "postgres://rhea:rhea@127.0.0.1:5432/rhea"
}

func Open(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres not reachable at %s: %w", dsn, err)
	}
	return &Store{Pool: pool}, nil
}

func (s *Store) Close() { s.Pool.Close() }

// Init creates the schema. Idempotent.
func (s *Store) Init(ctx context.Context) error {
	_, err := s.Pool.Exec(ctx, schemaSQL)
	return err
}

// Querier is satisfied by both *pgxpool.Pool and pgx.Tx, so executor code can
// run atomically inside a transaction.
type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// --- events ---------------------------------------------------------------

func AppendEvent(ctx context.Context, q Querier, ev core.Event) (int64, error) {
	var dedup *string
	if ev.DedupKey != "" {
		dedup = &ev.DedupKey
	}
	var ruleID *string
	var ruleVersion *int
	if ev.RuleID != "" {
		ruleID, ruleVersion = &ev.RuleID, &ev.RuleVersion
	}
	var id int64
	err := q.QueryRow(ctx, `
		INSERT INTO event (kind, event_type, occurred_at, payload, cause_event_id, rule_id, rule_version, dedup_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING event_id`,
		ev.Kind, ev.Type, ev.OccurredAt, ev.Payload, ev.CauseEventID, ruleID, ruleVersion, dedup,
	).Scan(&id)
	return id, err
}

func (s *Store) AppendEvent(ctx context.Context, ev core.Event) (int64, error) {
	return AppendEvent(ctx, s.Pool, ev)
}

func scanEvents(rows pgx.Rows) ([]core.Event, error) {
	defer rows.Close()
	var out []core.Event
	for rows.Next() {
		var ev core.Event
		var ruleID *string
		var ruleVersion *int
		var dedup *string
		if err := rows.Scan(&ev.ID, &ev.Kind, &ev.Type, &ev.OccurredAt, &ev.RecordedAt,
			&ev.Payload, &ev.CauseEventID, &ruleID, &ruleVersion, &dedup); err != nil {
			return nil, err
		}
		if ruleID != nil {
			ev.RuleID, ev.RuleVersion = *ruleID, *ruleVersion
		}
		if dedup != nil {
			ev.DedupKey = *dedup
		}
		out = append(out, ev)
	}
	return out, rows.Err()
}

const eventCols = `event_id, kind, event_type, to_char(occurred_at,'YYYY-MM-DD'), recorded_at,
	payload, cause_event_id, rule_id, rule_version, dedup_key`

func (s *Store) EventsByKind(ctx context.Context, kind string) ([]core.Event, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT `+eventCols+` FROM event WHERE kind = $1 ORDER BY event_id`, kind)
	if err != nil {
		return nil, err
	}
	return scanEvents(rows)
}

// UnmatchedRawEvents is the worklist: raw events no rule has acted on yet.
func (s *Store) UnmatchedRawEvents(ctx context.Context) ([]core.Event, error) {
	rows, err := s.Pool.Query(ctx, unmatchedSQL)
	if err != nil {
		return nil, err
	}
	return scanEvents(rows)
}

// --- rules ----------------------------------------------------------------

// InsertRuleVersion appends the next version row for rule.ID and returns it.
// Any change to a rule — including a status transition — goes through here.
func (s *Store) InsertRuleVersion(ctx context.Context, r core.Rule) (core.Rule, error) {
	spec, err := json.Marshal(r.Spec)
	if err != nil {
		return r, err
	}
	err = s.Pool.QueryRow(ctx, `
		INSERT INTO rule (rule_id, version, status, priority, effective_from, created_by, description, spec)
		VALUES ($1, COALESCE((SELECT MAX(version) FROM rule WHERE rule_id = $1), 0) + 1,
		        $2, $3, $4, $5, $6, $7)
		RETURNING version, created_at`,
		r.ID, r.Status, r.Priority, r.EffectiveFrom, r.CreatedBy, r.Description, spec,
	).Scan(&r.Version, &r.CreatedAt)
	return r, err
}

func scanRules(rows pgx.Rows) ([]core.Rule, error) {
	defer rows.Close()
	var out []core.Rule
	for rows.Next() {
		var r core.Rule
		var spec []byte
		if err := rows.Scan(&r.ID, &r.Version, &r.Status, &r.Priority, &r.EffectiveFrom,
			&r.CreatedBy, &r.Description, &spec, &r.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(spec, &r.Spec); err != nil {
			return nil, fmt.Errorf("rule %s v%d spec: %w", r.ID, r.Version, err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// LatestRules returns the newest version of every rule.
func (s *Store) LatestRules(ctx context.Context) ([]core.Rule, error) {
	rows, err := s.Pool.Query(ctx, latestRulesSQL)
	if err != nil {
		return nil, err
	}
	return scanRules(rows)
}

// ActiveRules returns rules whose latest version is active, in firing order
// (priority ascending, rule_id as tiebreak — DECISIONS.md 2026-10-01).
func (s *Store) ActiveRules(ctx context.Context) ([]core.Rule, error) {
	all, err := s.LatestRules(ctx)
	if err != nil {
		return nil, err
	}
	var out []core.Rule
	for _, r := range all {
		if r.Status == core.StatusActive {
			out = append(out, r)
		}
	}
	return out, nil
}

func (s *Store) GetRule(ctx context.Context, id string) (core.Rule, error) {
	rules, err := s.LatestRules(ctx)
	if err != nil {
		return core.Rule{}, err
	}
	for _, r := range rules {
		if r.ID == id {
			return r, nil
		}
	}
	return core.Rule{}, fmt.Errorf("rule %q not found", id)
}

// --- object types ---------------------------------------------------------

func (s *Store) InsertObjectType(ctx context.Context, t core.ObjectType) error {
	spec, err := json.Marshal(t)
	if err != nil {
		return err
	}
	_, err = s.Pool.Exec(ctx,
		`INSERT INTO object_type (name, version, spec) VALUES ($1, $2, $3)`,
		t.Name, t.Version, spec)
	return err
}

func (s *Store) GetObjectType(ctx context.Context, name string) (core.ObjectType, error) {
	var spec []byte
	err := s.Pool.QueryRow(ctx,
		`SELECT spec FROM object_type WHERE name = $1 ORDER BY version DESC LIMIT 1`,
		name).Scan(&spec)
	if err != nil {
		return core.ObjectType{}, fmt.Errorf("object type %q: %w", name, err)
	}
	var t core.ObjectType
	return t, json.Unmarshal(spec, &t)
}

// ListObjectTypes returns the newest version of every object type.
func (s *Store) ListObjectTypes(ctx context.Context) ([]core.ObjectType, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT DISTINCT ON (name) spec FROM object_type ORDER BY name, version DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []core.ObjectType
	for rows.Next() {
		var spec []byte
		if err := rows.Scan(&spec); err != nil {
			return nil, err
		}
		var t core.ObjectType
		if err := json.Unmarshal(spec, &t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// --- view defs ------------------------------------------------------------

func (s *Store) InsertViewDef(ctx context.Context, v core.ViewDef) error {
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO view_def (view_id, version, notion, title, domain, function, spec)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		v.ID, v.Version, v.Notion, v.Title, v.Domain, v.Function, v.Spec)
	return err
}

// LatestViewDefs returns the newest version of every view definition.
func (s *Store) LatestViewDefs(ctx context.Context) ([]core.ViewDef, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT DISTINCT ON (view_id) view_id, version, notion, title, domain, function, spec
		FROM view_def ORDER BY view_id, version DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []core.ViewDef
	for rows.Next() {
		var v core.ViewDef
		if err := rows.Scan(&v.ID, &v.Version, &v.Notion, &v.Title, &v.Domain, &v.Function, &v.Spec); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// --- object projection cache ----------------------------------------------

func InsertObject(ctx context.Context, q Querier, o core.Object) error {
	state, err := json.Marshal(o.State)
	if err != nil {
		return err
	}
	_, err = q.Exec(ctx, `
		INSERT INTO object (object_id, object_type, type_version, state, source_event_id, rule_id, rule_version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		o.ID, o.Type, o.TypeVersion, state, o.SourceEventID, o.RuleID, o.RuleVersion)
	return err
}

func (s *Store) WipeObjects(ctx context.Context) error {
	_, err := s.Pool.Exec(ctx, `DELETE FROM object`)
	return err
}

func scanObjects(rows pgx.Rows) ([]core.Object, error) {
	defer rows.Close()
	var out []core.Object
	for rows.Next() {
		var o core.Object
		var state []byte
		if err := rows.Scan(&o.ID, &o.Type, &o.TypeVersion, &state,
			&o.SourceEventID, &o.RuleID, &o.RuleVersion); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(state, &o.State); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

const objectCols = `object_id, object_type, type_version, state, source_event_id, rule_id, rule_version`

func (s *Store) ObjectsByType(ctx context.Context, typ string) ([]core.Object, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT `+objectCols+` FROM object WHERE object_type = $1 ORDER BY object_id`, typ)
	if err != nil {
		return nil, err
	}
	return scanObjects(rows)
}

func (s *Store) AllObjects(ctx context.Context) ([]core.Object, error) {
	rows, err := s.Pool.Query(ctx, `SELECT `+objectCols+` FROM object ORDER BY object_id`)
	if err != nil {
		return nil, err
	}
	return scanObjects(rows)
}

func (s *Store) GetObject(ctx context.Context, id string) (core.Object, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT `+objectCols+` FROM object WHERE object_id = $1`, id)
	if err != nil {
		return core.Object{}, err
	}
	objs, err := scanObjects(rows)
	if err != nil {
		return core.Object{}, err
	}
	if len(objs) == 0 {
		return core.Object{}, fmt.Errorf("object %q not found", id)
	}
	return objs[0], nil
}
