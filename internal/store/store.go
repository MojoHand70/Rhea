// Package store is the Postgres system of record: the append-only event log,
// the versioned definition stores (rule, object_type, view_def) and the
// rebuildable object projection cache.
package store

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
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

//go:embed queries/pending.sql
var pendingSQL string

//go:embed queries/latest_rules.sql
var latestRulesSQL string

//go:embed queries/active_rules.sql
var activeRulesSQL string

//go:embed queries/chain.sql
var chainSQL string

//go:embed queries/bundle_rejection.sql
var bundleRejectionSQL string

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

// Init creates the schema and seeds the builtin activities, each activation
// recorded as an event — the gate's own birth is in the log. Idempotent.
func (s *Store) Init(ctx context.Context) error {
	if _, err := s.Pool.Exec(ctx, schemaSQL); err != nil {
		return err
	}
	return s.seedBuiltinActivities(ctx)
}

// IsDuplicate reports a unique-constraint violation — how append-only tables
// say "already there". Idempotent loaders (packs, adapter retries) treat it
// as a skip, not an error.
func IsDuplicate(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
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
	// The object.* namespace is the kernel's own: a raw event spelled like a
	// derived one could trigger cascade rules — or claim an amendment — for
	// a fact that never happened.
	if ev.Kind == core.KindRaw && (ev.Type == core.EventObjectMaterialized || ev.Type == core.EventObjectAmended) {
		return 0, fmt.Errorf("event type %q is reserved for the kernel", ev.Type)
	}
	// The system-verb namespaces belong to declared doors: a raw rule.* or
	// activity.* event without an activity stamp would claim a system act
	// (an approval, a request) that no declared verb performed.
	if ev.Kind == core.KindRaw && core.ReservedEventType(ev.Type) && ev.ActivityName == "" {
		return 0, fmt.Errorf("event type %q is a system verb — it enters only through its declared activity", ev.Type)
	}
	var dedup *string
	if ev.DedupKey != "" {
		dedup = &ev.DedupKey
	}
	var ruleID *string
	var ruleVersion *int
	if ev.RuleID != "" {
		ruleID, ruleVersion = &ev.RuleID, &ev.RuleVersion
	}
	var actor *string
	if ev.Actor != "" {
		actor = &ev.Actor
	}
	var actName *string
	var actVersion *int
	if ev.ActivityName != "" {
		actName, actVersion = &ev.ActivityName, &ev.ActivityVersion
	}
	var id int64
	err := q.QueryRow(ctx, `
		INSERT INTO event (kind, event_type, occurred_at, payload, cause_event_id, rule_id, rule_version, dedup_key, actor, activity_name, activity_version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING event_id`,
		ev.Kind, ev.Type, ev.OccurredAt, ev.Payload, ev.CauseEventID, ruleID, ruleVersion, dedup, actor, actName, actVersion,
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
		var actor *string
		var actName *string
		var actVersion *int
		if err := rows.Scan(&ev.ID, &ev.Kind, &ev.Type, &ev.OccurredAt, &ev.RecordedAt,
			&ev.Payload, &ev.CauseEventID, &ruleID, &ruleVersion, &dedup, &actor,
			&actName, &actVersion); err != nil {
			return nil, err
		}
		if ruleID != nil {
			ev.RuleID, ev.RuleVersion = *ruleID, *ruleVersion
		}
		if dedup != nil {
			ev.DedupKey = *dedup
		}
		if actor != nil {
			ev.Actor = *actor
		}
		if actName != nil {
			ev.ActivityName = *actName
			if actVersion != nil {
				ev.ActivityVersion = *actVersion
			}
		}
		out = append(out, ev)
	}
	return out, rows.Err()
}

const eventCols = `event_id, kind, event_type, to_char(occurred_at,'YYYY-MM-DD'), recorded_at,
	payload, cause_event_id, rule_id, rule_version, dedup_key, actor, activity_name, activity_version`

func (s *Store) EventsByKind(ctx context.Context, kind string) ([]core.Event, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT `+eventCols+` FROM event WHERE kind = $1 ORDER BY event_id`, kind)
	if err != nil {
		return nil, err
	}
	return scanEvents(rows)
}

// EarliestFactDate is the business date of the log's first raw fact: where
// an interview's rules begin to apply — they explain the client's history,
// not only its future. Empty when the log holds no facts.
func (s *Store) EarliestFactDate(ctx context.Context) (string, error) {
	var d *string
	err := s.Pool.QueryRow(ctx, `
		SELECT to_char(MIN(occurred_at),'YYYY-MM-DD') FROM event
		WHERE kind = 'raw' AND event_type NOT LIKE 'rule.%' AND event_type NOT LIKE 'activity.%'
		  AND event_type NOT LIKE 'bundle.%' AND event_type NOT LIKE 'backfill.%'
		  AND event_type NOT LIKE 'time.%'`).Scan(&d)
	if err != nil || d == nil {
		return "", err
	}
	return *d, nil
}

// Installation names this kernel in the network: RHEA_INSTALLATION, or the
// database's own name — one database, one installation, in this experiment.
func (s *Store) Installation(ctx context.Context) (string, error) {
	if name := os.Getenv("RHEA_INSTALLATION"); name != "" {
		return name, nil
	}
	var name string
	err := s.Pool.QueryRow(ctx, `SELECT current_database()`).Scan(&name)
	return name, err
}

// ChainOf returns every derived event rooted at a raw event, in log order.
func (s *Store) ChainOf(ctx context.Context, rootID int64) ([]core.Event, error) {
	rows, err := s.Pool.Query(ctx, chainSQL, rootID)
	if err != nil {
		return nil, err
	}
	return scanEvents(rows)
}

// GetEvent returns one event by id.
func (s *Store) GetEvent(ctx context.Context, id int64) (core.Event, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT `+eventCols+` FROM event WHERE event_id = $1`, id)
	if err != nil {
		return core.Event{}, err
	}
	evs, err := scanEvents(rows)
	if err != nil {
		return core.Event{}, err
	}
	if len(evs) == 0 {
		return core.Event{}, fmt.Errorf("event %d not found", id)
	}
	return evs[0], nil
}

// UnmatchedRawEvents is the worklist: raw events no rule has acted on yet.
func (s *Store) UnmatchedRawEvents(ctx context.Context) ([]core.Event, error) {
	rows, err := s.Pool.Query(ctx, unmatchedSQL)
	if err != nil {
		return nil, err
	}
	return scanEvents(rows)
}

// PendingEvents is the executor's set: raw events no rule has acted on,
// time events included — the worklist is the human view of the same
// question minus the uneventful days.
func (s *Store) PendingEvents(ctx context.Context) ([]core.Event, error) {
	rows, err := s.Pool.Query(ctx, pendingSQL)
	if err != nil {
		return nil, err
	}
	return scanEvents(rows)
}

// LatestEventOfType returns the newest event of one type, or ok=false when
// none exists — how the clock finds the last opened day.
func (s *Store) LatestEventOfType(ctx context.Context, eventType string) (core.Event, bool, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT `+eventCols+` FROM event WHERE event_type = $1 ORDER BY event_id DESC LIMIT 1`,
		eventType)
	if err != nil {
		return core.Event{}, false, err
	}
	evs, err := scanEvents(rows)
	if err != nil || len(evs) == 0 {
		return core.Event{}, false, err
	}
	return evs[0], true, nil
}

// MaxEventID returns the highest event id — the base for the synthetic ids
// a dry run gives hypothetical events.
func (s *Store) MaxEventID(ctx context.Context) (int64, error) {
	var id int64
	err := s.Pool.QueryRow(ctx, `SELECT COALESCE(MAX(event_id), 0) FROM event`).Scan(&id)
	return id, err
}

// AmendmentsOf returns the object.amended events that moved one object, in
// log order — the second half of its explanation (invariant 5): the
// materialization says why it exists, the amendments say why it is what it
// is now.
func (s *Store) AmendmentsOf(ctx context.Context, objectID string) ([]core.Event, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT `+eventCols+` FROM event
		 WHERE event_type = $1 AND payload->>'object_id' = $2 ORDER BY event_id`,
		core.EventObjectAmended, objectID)
	if err != nil {
		return nil, err
	}
	return scanEvents(rows)
}

// --- live notices -----------------------------------------------------------

// The executor is the single writer of the object cache, so it is the single
// announcer (DIRECTION: Alpha's live screens). A notice is a signal, never
// data: its payload names the touched object types, and listeners re-read
// through the ordinary API.

const projectionChannel = "rhea_projection"

// NotifyProjection announces a projection write from inside its transaction;
// Postgres delivers it only if the transaction commits, so a rolled-back
// write never announces itself.
func NotifyProjection(ctx context.Context, q Querier, payload string) error {
	_, err := q.Exec(ctx, `SELECT pg_notify($1, $2)`, projectionChannel, payload)
	return err
}

// ProjectionNotices listens for projection announcements, delivering each
// payload on the returned channel until ctx ends. It holds one pooled
// connection for the duration. A slow consumer drops notices rather than
// blocking the listener — a notice only ever means "look again".
func (s *Store) ProjectionNotices(ctx context.Context) (<-chan string, error) {
	conn, err := s.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := conn.Exec(ctx, `LISTEN `+projectionChannel); err != nil {
		conn.Release()
		return nil, err
	}
	ch := make(chan string, 16)
	go func() {
		defer conn.Release()
		defer close(ch)
		for {
			n, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				return // ctx ended or the connection broke; SSE clients reconnect
			}
			select {
			case ch <- n.Payload:
			default:
			}
		}
	}()
	return ch, nil
}

// --- rules ----------------------------------------------------------------

// InsertRuleVersion appends the next version row for rule.ID and returns it.
// Any change to a rule — including a status transition — goes through here.
// Package-level so a door's reaction can write the approval event and the
// version flip in one transaction.
func InsertRuleVersion(ctx context.Context, q Querier, r core.Rule) (core.Rule, error) {
	spec, err := json.Marshal(r.Spec)
	if err != nil {
		return r, err
	}
	err = q.QueryRow(ctx, `
		INSERT INTO rule (rule_id, version, status, priority, effective_from, created_by, description, spec)
		VALUES ($1, COALESCE((SELECT MAX(version) FROM rule WHERE rule_id = $1), 0) + 1,
		        $2, $3, $4, $5, $6, $7)
		RETURNING version, created_at`,
		r.ID, r.Status, r.Priority, r.EffectiveFrom, r.CreatedBy, r.Description, spec,
	).Scan(&r.Version, &r.CreatedAt)
	return r, err
}

func (s *Store) InsertRuleVersion(ctx context.Context, r core.Rule) (core.Rule, error) {
	return InsertRuleVersion(ctx, s.Pool, r)
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

// ActiveRules returns the executing rule set in firing order (priority
// ascending, rule_id as tiebreak — DECISIONS.md 2026-10-01): the newest
// active version of each rule. A draft in progress does not suspend the
// running version; only a latest version of superseded retires the rule.
func (s *Store) ActiveRules(ctx context.Context) ([]core.Rule, error) {
	rows, err := s.Pool.Query(ctx, activeRulesSQL)
	if err != nil {
		return nil, err
	}
	return scanRules(rows)
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

// GetRuleVersion returns one exact rule version, whatever its status: the
// draft a rejected bundle carried is read back as it was proposed.
func (s *Store) GetRuleVersion(ctx context.Context, id string, version int) (core.Rule, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT rule_id, version, status, priority, to_char(effective_from,'YYYY-MM-DD'),
		       created_by, description, spec, created_at
		FROM rule WHERE rule_id = $1 AND version = $2`, id, version)
	if err != nil {
		return core.Rule{}, err
	}
	rules, err := scanRules(rows)
	if err != nil {
		return core.Rule{}, err
	}
	if len(rules) == 0 {
		return core.Rule{}, fmt.Errorf("rule %s v%d not found", id, version)
	}
	return rules[0], nil
}

// RuleKey names one exact rule version — the unit provenance speaks in.
type RuleKey struct {
	ID      string
	Version int
}

// RuleDescriptions returns the description of every rule version ever stored,
// superseded ones included: the provenance walk explains history, and history
// names exact versions.
func (s *Store) RuleDescriptions(ctx context.Context) (map[RuleKey]string, error) {
	rows, err := s.Pool.Query(ctx, `SELECT rule_id, version, description FROM rule`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[RuleKey]string{}
	for rows.Next() {
		var k RuleKey
		var desc string
		if err := rows.Scan(&k.ID, &k.Version, &desc); err != nil {
			return nil, err
		}
		out[k] = desc
	}
	return out, rows.Err()
}

// --- object types ---------------------------------------------------------

// InsertObjectType writes an object type straight in as active: the
// bootstrap and test path, the way tests insert active rules. Every door a
// person or an agent uses lands types as drafts inside a bundle.
func (s *Store) InsertObjectType(ctx context.Context, t core.ObjectType) error {
	return InsertObjectTypeRow(ctx, s.Pool, t, core.StatusActive)
}

// InsertObjectTypeRow appends one (name, version, status) row. Approval and
// rejection append rows of an existing version; the version itself is the
// schema's identity and never moves with the lifecycle.
func InsertObjectTypeRow(ctx context.Context, q Querier, t core.ObjectType, status string) error {
	if err := t.Validate(); err != nil {
		return err
	}
	spec, err := json.Marshal(t)
	if err != nil {
		return err
	}
	_, err = q.Exec(ctx,
		`INSERT INTO object_type (name, version, status, spec) VALUES ($1, $2, $3, $4)`,
		t.Name, t.Version, status, spec)
	return err
}

// GetObjectType returns the newest ACTIVE version of a type: what the kernel
// validates against and the shell renders. A draft version waits its bundle.
func (s *Store) GetObjectType(ctx context.Context, name string) (core.ObjectType, error) {
	var spec []byte
	err := s.Pool.QueryRow(ctx, `
		SELECT spec FROM object_type WHERE name = $1 AND status = 'active'
		ORDER BY version DESC LIMIT 1`,
		name).Scan(&spec)
	if err != nil {
		return core.ObjectType{}, fmt.Errorf("object type %q: %w", name, err)
	}
	var t core.ObjectType
	return t, json.Unmarshal(spec, &t)
}

// GetObjectTypeVersion returns one version of a type and its lifecycle
// status: active or superseded once either row exists, draft otherwise.
func (s *Store) GetObjectTypeVersion(ctx context.Context, name string, version int) (core.ObjectType, string, error) {
	var spec []byte
	var status string
	err := s.Pool.QueryRow(ctx, `
		SELECT spec, status FROM object_type WHERE name = $1 AND version = $2
		ORDER BY CASE status WHEN 'draft' THEN 1 ELSE 0 END LIMIT 1`,
		name, version).Scan(&spec, &status)
	if err != nil {
		return core.ObjectType{}, "", fmt.Errorf("object type %s v%d: %w", name, version, err)
	}
	var t core.ObjectType
	return t, status, json.Unmarshal(spec, &t)
}

// ListObjectTypes returns the newest active version of every object type.
func (s *Store) ListObjectTypes(ctx context.Context) ([]core.ObjectType, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT DISTINCT ON (name) spec FROM object_type WHERE status = 'active'
		ORDER BY name, version DESC`)
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

// InsertViewDef writes a view straight in as active: the bootstrap and test
// path, like InsertObjectType.
func (s *Store) InsertViewDef(ctx context.Context, v core.ViewDef) error {
	return InsertViewDefRow(ctx, s.Pool, v, core.StatusActive)
}

// InsertViewDefRow appends one (view_id, version, status) row.
func InsertViewDefRow(ctx context.Context, q Querier, v core.ViewDef, status string) error {
	_, err := q.Exec(ctx, `
		INSERT INTO view_def (view_id, version, status, notion, title, domain, function, spec)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		v.ID, v.Version, status, v.Notion, v.Title, v.Domain, v.Function, v.Spec)
	return err
}

// GetViewDefVersion returns one version of a view and its lifecycle status,
// resolved like GetObjectTypeVersion.
func (s *Store) GetViewDefVersion(ctx context.Context, id string, version int) (core.ViewDef, string, error) {
	var v core.ViewDef
	var status string
	err := s.Pool.QueryRow(ctx, `
		SELECT view_id, version, notion, title, domain, function, spec, status
		FROM view_def WHERE view_id = $1 AND version = $2
		ORDER BY CASE status WHEN 'draft' THEN 1 ELSE 0 END LIMIT 1`,
		id, version).Scan(&v.ID, &v.Version, &v.Notion, &v.Title, &v.Domain, &v.Function, &v.Spec, &status)
	if err != nil {
		return v, "", fmt.Errorf("view %s v%d: %w", id, version, err)
	}
	return v, status, nil
}

// LatestViewDefs returns the newest active version of every view definition.
func (s *Store) LatestViewDefs(ctx context.Context) ([]core.ViewDef, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT DISTINCT ON (view_id) view_id, version, notion, title, domain, function, spec
		FROM view_def WHERE status = 'active' ORDER BY view_id, version DESC`)
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

// AmendObject merges an amendment's baked delta into the cached state. The
// object table is the kernel's rebuildable cache (invariant 7) — this is the
// executor/replayer applying an object.amended event, never anyone editing.
func AmendObject(ctx context.Context, q Querier, objectID string, set map[string]any) error {
	delta, err := json.Marshal(set)
	if err != nil {
		return err
	}
	tag, err := q.Exec(ctx,
		`UPDATE object SET state = state || $2::jsonb WHERE object_id = $1`, objectID, delta)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("amendment of %q touched %d objects, want exactly 1", objectID, tag.RowsAffected())
	}
	return nil
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
		`SELECT `+objectCols+` FROM object WHERE object_type = $1 ORDER BY source_event_id, object_id`, typ)
	if err != nil {
		return nil, err
	}
	return scanObjects(rows)
}

// CountObjectsByType returns instance counts for every type present in the
// cache, in one query — the Language catalog asks about all types at once,
// and counting must not mean loading every object.
func (s *Store) CountObjectsByType(ctx context.Context) (map[string]int, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT object_type, COUNT(*) FROM object GROUP BY object_type`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var typ string
		var n int
		if err := rows.Scan(&typ, &n); err != nil {
			return nil, err
		}
		out[typ] = n
	}
	return out, rows.Err()
}

func (s *Store) AllObjects(ctx context.Context) ([]core.Object, error) {
	rows, err := s.Pool.Query(ctx, `SELECT `+objectCols+` FROM object ORDER BY object_id`)
	if err != nil {
		return nil, err
	}
	return scanObjects(rows)
}

// FindObjectIDsByField backs the kernel's Lookup: every object of the given
// type whose state field equals value, compared as text. ref() demands one
// match; the period lock asks only whether any exist.
func (s *Store) FindObjectIDsByField(ctx context.Context, typ, field, value string) ([]string, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT object_id FROM object WHERE object_type = $1 AND state->>$2 = $3 ORDER BY source_event_id, object_id`,
		typ, field, value)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
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
