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
	"sort"
	"strings"
	"time"

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

//go:embed queries/last_event.sql
var lastEventSQL string

//go:embed queries/explained.sql
var explainedSQL string

//go:embed queries/latest_rules.sql
var latestRulesSQL string

//go:embed queries/active_rules.sql
var activeRulesSQL string

//go:embed queries/chain.sql
var chainSQL string

//go:embed queries/bundle_rejection.sql
var bundleRejectionSQL string

//go:embed queries/events_of_types.sql
var eventsOfTypesSQL string

//go:embed queries/object_roots.sql
var objectRootsSQL string

//go:embed queries/object_sources.sql
var objectSourcesSQL string

//go:embed queries/dedup_known.sql
var dedupKnownSQL string

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
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	// Invariant 9: a connection handed to a scoped request runs as rhea_app
	// with the scope set, so row-level security decides what it sees and
	// may write; it is reset when the connection returns. A failure to set
	// either discards the connection — the owner's view never leaks into a
	// scoped request.
	cfg.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		scope, scoped := ScopeOf(ctx)
		if !scoped {
			return true // the owner's own work
		}
		if _, err := conn.Exec(ctx, `SET ROLE rhea_app`); err != nil {
			return false
		}
		if _, err := conn.Exec(ctx, `SELECT set_config('rhea.scope', $1, false)`, strings.Join(scope, ",")); err != nil {
			return false
		}
		return true
	}
	cfg.AfterRelease = func(conn *pgx.Conn) bool {
		_, err := conn.Exec(context.Background(), `RESET ROLE; RESET rhea.scope`)
		return err == nil
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres not reachable at %s: %w", RedactDSN(dsn), err)
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

// ErrNotFound is how a read says the object is not there — distinct from a
// store failure, which a caller must not read as "no object".
var ErrNotFound = errors.New("not found")

// IsDuplicate reports a unique-constraint violation — how append-only tables
// say "already there". Idempotent loaders (packs, adapter retries) treat it
// as a skip, not an error.
func IsDuplicate(err error) bool {
	if errors.Is(err, ErrDuplicate) {
		return true
	}
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

// AppendEvent writes one event inside the caller's transaction (or on q
// as it is). A fact rules act on is refused unless the caller holds the
// kernel's lock: a transaction cannot take it half way.
func AppendEvent(ctx context.Context, q Querier, ev core.Event) (int64, error) {
	return appendEvent(ctx, q, ev, false)
}

// ErrDuplicate is how the batch append says "already there" for one event;
// IsDuplicate recognizes it alongside the database's own unique violation.
var ErrDuplicate = errors.New("event already in the log (dedup key)")

// appendEvent is the one write into the log. With skipDuplicate the insert
// does nothing on a known dedup key and reports ErrDuplicate instead of
// aborting the enclosing transaction — the batch path.
func appendEvent(ctx context.Context, q Querier, ev core.Event, skipDuplicate bool) (int64, error) {
	if len(ev.Files) > 0 {
		return 0, fmt.Errorf("%s carries files: an event's files are kept by AppendEvents, which keeps them first", ev.Type)
	}
	// The object.* namespace is the kernel's own: a raw event spelled like a
	// derived one could trigger cascade rules — or claim an amendment — for
	// a fact that never happened.
	if ev.Kind == core.KindRaw && (ev.Type == core.EventObjectMaterialized || ev.Type == core.EventObjectAmended) {
		return 0, fmt.Errorf("event type %q is reserved for the kernel", ev.Type)
	}
	// The system-verb namespaces belong to declared doors: a raw rule.* or
	// activity.* event without an activity stamp would claim a system act
	// (an approval, a request) that no declared verb performed.
	// The one exception: a refusal is the kernel's own word, written by no
	// door — nobody opened one — and signed "kernel".
	if ev.Kind == core.KindRaw && core.ReservedEventType(ev.Type) && ev.ActivityName == "" && !(ev.Type == core.EventAccessRefused && ev.Actor == "kernel") {
		return 0, fmt.Errorf("event type %q is a system verb — it enters only through its declared activity", ev.Type)
	}
	// A fact rules act on is recorded under the kernel's lock, so no chain
	// ever runs beside a half-written one (lock.go).
	if ruleFood(ev) && !Locked(ctx) {
		return 0, fmt.Errorf("a fact rules act on (%s) is recorded under the kernel's lock — through a door, or the store's AppendEvent, which takes it", ev.Type)
	}
	// The file is the fact: a document.received names files the store holds,
	// or it is not appended. Whoever emits it — a door, a channel, a pack —
	// the file comes first.
	if ev.Kind == core.KindRaw && ev.Type == core.EventDocumentReceived {
		hashes := fileHashes(ev.Payload)
		if len(hashes) == 0 {
			return 0, fmt.Errorf("%s names no file — the payload wants {<input>: {hash, name, media_type, size}}", ev.Type)
		}
		for _, h := range hashes {
			ok, err := hasBlob(ctx, q, h)
			if err != nil {
				return 0, err
			}
			if !ok {
				return 0, fmt.Errorf("%s names file %s, which the store does not hold — the file is the fact, so the file comes first", ev.Type, h)
			}
		}
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
	sql := `
		INSERT INTO event (kind, event_type, occurred_at, payload, cause_event_id, rule_id, rule_version, dedup_key, actor, activity_name, activity_version, company)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	if skipDuplicate {
		sql += ` ON CONFLICT (COALESCE(company, ''), dedup_key) WHERE dedup_key IS NOT NULL DO NOTHING`
	}
	company, err := factCompany(ctx, ev)
	if err != nil {
		return 0, err
	}
	var companyKey *string
	if company != "" {
		companyKey = &company
	}
	err = q.QueryRow(ctx, sql+` RETURNING event_id`,
		ev.Kind, ev.Type, ev.OccurredAt, ev.Payload, ev.CauseEventID, ruleID, ruleVersion, dedup, actor, actName, actVersion, companyKey,
	).Scan(&id)
	if skipDuplicate && errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrDuplicate
	}
	return id, err
}

// AppendEvent records one event on its own; a fact rules act on waits for
// the kernel's lock (the clock's day, a pack's master data).
func (s *Store) AppendEvent(ctx context.Context, ev core.Event) (int64, error) {
	if ruleFood(ev) {
		held, release, err := s.Lock(ctx, 0)
		if err != nil {
			return 0, err
		}
		defer release()
		ctx = held
	}
	return AppendEvent(ctx, s.Pool, ev)
}

// AppendEvents appends a batch in one transaction: an adapter pass lands
// whole or not at all (DIRECTION 2026-10-10, the high-volume treatment's
// cheap-now item). Events the log already has, by dedup key, are skipped and
// counted; any other failure rolls the whole batch back.
func (s *Store) AppendEvents(ctx context.Context, evs []core.Event) (appended, duplicates int, err error) {
	for _, ev := range evs {
		if ruleFood(ev) { // the batch waits for the kernel's lock as one fact does
			held, release, err := s.Lock(ctx, 0)
			if err != nil {
				return 0, 0, err
			}
			defer release()
			ctx = held
			break
		}
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback(ctx)
	for _, ev := range evs {
		if len(ev.Files) > 0 {
			kept, err := keepFiles(ctx, tx, ev)
			if err != nil {
				return 0, 0, err
			}
			if !kept {
				duplicates++
				continue
			}
			ev.Files = nil
		}
		switch _, err := appendEvent(ctx, tx, ev, true); {
		case err == nil:
			appended++
		case errors.Is(err, ErrDuplicate):
			duplicates++
		default:
			return 0, 0, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, 0, err
	}
	return appended, duplicates, nil
}

// factCompany is the partition key a fact is appended under: a raw fact's
// own, else the company the call acts for; a derived event always keeps its
// chain's (the root's), never the caller's — a request scoped to one company
// booking installation master data stamped the explanation with that
// company, and replay diverged.
func factCompany(ctx context.Context, ev core.Event) (string, error) {
	company := ev.Company
	if company == "" && ev.Kind == core.KindRaw {
		company = CompanyOf(ctx)
		if set, scoped := ScopeOf(ctx); company == "" && scoped && len(set) > 0 {
			return "", fmt.Errorf("a scope of several companies acts for none — name the company the fact belongs to")
		}
	}
	return company, nil
}

// keepFiles keeps the files a channel's fact carries, in the fact's company,
// before the fact: the file is the fact, so the file comes first. A fact the
// log already has by its dedup key keeps nothing (kept=false) — a package
// sealed again after a crash is not the one the log names.
func keepFiles(ctx context.Context, q Querier, ev core.Event) (kept bool, err error) {
	company, err := factCompany(ctx, ev)
	if err != nil {
		return false, err
	}
	if ev.DedupKey != "" {
		var known bool
		if err := q.QueryRow(ctx, dedupKnownSQL, company, ev.DedupKey).Scan(&known); err != nil {
			return false, err
		}
		if known {
			return false, nil
		}
	}
	for _, f := range ev.Files {
		if _, err := putBlob(ctx, q, company, f.Hash, f.MediaType, f.Bytes); err != nil {
			return false, err
		}
	}
	return true, nil
}

// fileHashes collects the hashes a document.received payload names: every
// top-level value shaped like a file fact, {hash: "sha256:…", …}.
func fileHashes(payload []byte) []string {
	var top map[string]json.RawMessage
	if json.Unmarshal(payload, &top) != nil {
		return nil
	}
	var out []string
	for _, v := range top {
		var f struct {
			Hash string `json:"hash"`
		}
		if json.Unmarshal(v, &f) == nil && strings.HasPrefix(f.Hash, "sha256:") {
			out = append(out, f.Hash)
		}
	}
	sort.Strings(out)
	return out
}

// --- blobs ----------------------------------------------------------------

// Blob is one stored file: the bytes under their content hash.
type Blob struct {
	Hash      string
	MediaType string
	Size      int64
	Bytes     []byte
	CreatedAt time.Time
}

// PutBlob keeps a file under its hash for the context's company, once: a
// second put of the same bytes there is a no-op (stored=false). The hash is
// checked against the bytes — the store vouches the content address,
// whoever computed it. The file is the company's that received it, seen
// only where that company is (invariant 9).
func PutBlob(ctx context.Context, q Querier, hash, mediaType string, data []byte) (stored bool, err error) {
	return putBlob(ctx, q, CompanyOf(ctx), hash, mediaType, data)
}

// putBlob keeps a file for the company named, "" for the installation.
func putBlob(ctx context.Context, q Querier, owner, hash, mediaType string, data []byte) (stored bool, err error) {
	if want := core.HashBytes(data); hash != want {
		return false, fmt.Errorf("blob %s: the bytes hash to %s", hash, want)
	}
	if mediaType == "" {
		mediaType = "application/octet-stream"
	}
	var company *string
	if owner != "" {
		company = &owner
	}
	tag, err := q.Exec(ctx, `
		INSERT INTO blob (hash, media_type, size, bytes, company) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (hash, (COALESCE(company, ''))) DO NOTHING`, hash, mediaType, len(data), data, company)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

func (s *Store) PutBlob(ctx context.Context, hash, mediaType string, data []byte) (bool, error) {
	return PutBlob(ctx, s.Pool, hash, mediaType, data)
}

// GetBlob returns a stored file, bytes included, when the context's scope
// holds it: a file is seen only where the company that received it is.
func (s *Store) GetBlob(ctx context.Context, hash string) (Blob, error) {
	var b Blob
	err := s.Pool.QueryRow(ctx,
		`SELECT hash, media_type, size, bytes, created_at FROM blob WHERE hash = $1 LIMIT 1`, hash,
	).Scan(&b.Hash, &b.MediaType, &b.Size, &b.Bytes, &b.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return b, fmt.Errorf("no file %s in the store", hash)
	}
	return b, err
}

func hasBlob(ctx context.Context, q Querier, hash string) (bool, error) {
	var ok bool
	err := q.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM blob WHERE hash = $1)`, hash).Scan(&ok)
	return ok, err
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
		var company *string
		if err := rows.Scan(&ev.ID, &ev.Kind, &ev.Type, &ev.OccurredAt, &ev.RecordedAt,
			&ev.Payload, &ev.CauseEventID, &ruleID, &ruleVersion, &dedup, &actor,
			&actName, &actVersion, &company); err != nil {
			return nil, err
		}
		if company != nil {
			ev.Company = *company
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
	payload, cause_event_id, rule_id, rule_version, dedup_key, actor, activity_name, activity_version,
       company`

func (s *Store) EventsByKind(ctx context.Context, kind string) ([]core.Event, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT `+eventCols+` FROM event WHERE kind = $1 AND ($2 = '' OR company = $2 OR COALESCE(company, '') = '') ORDER BY event_id`, kind, CompanyOf(ctx))
	if err != nil {
		return nil, err
	}
	return scanEvents(rows)
}

// EventsOfTypes returns the events of the named types in log order, within
// the context's company (empty: the whole log).
func (s *Store) EventsOfTypes(ctx context.Context, types []string) ([]core.Event, error) {
	rows, err := s.Pool.Query(ctx, eventsOfTypesSQL, types, CompanyOf(ctx))
	if err != nil {
		return nil, err
	}
	return scanEvents(rows)
}

// ObjectRoots maps each object of one type to the raw event its chain
// starts from, within the context's company.
func (s *Store) ObjectRoots(ctx context.Context, typ string) (map[string]int64, error) {
	rows, err := s.Pool.Query(ctx, objectRootsSQL, typ, CompanyOf(ctx))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int64{}
	for rows.Next() {
		var id string
		var root int64
		if err := rows.Scan(&id, &root); err != nil {
			return nil, err
		}
		out[id] = root
	}
	return out, rows.Err()
}

// ObjectSources maps each object of one type to the record it was made
// from, within the context's company; an object a raw fact made directly
// is not in the map.
func (s *Store) ObjectSources(ctx context.Context, typ string) (map[string]string, error) {
	rows, err := s.Pool.Query(ctx, objectSourcesSQL, typ, CompanyOf(ctx))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var id, from string
		if err := rows.Scan(&id, &from); err != nil {
			return nil, err
		}
		out[id] = from
	}
	return out, rows.Err()
}

// EarliestFactDate is the business date of the log's first raw fact: where
// an interview's rules begin to apply — they explain the client's history,
// not only its future. Empty when the log holds no facts.
func (s *Store) EarliestFactDate(ctx context.Context) (string, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT event_type, to_char(MIN(occurred_at),'YYYY-MM-DD') FROM event
		WHERE kind = 'raw' GROUP BY event_type`)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	earliest := ""
	for rows.Next() {
		var typ, d string
		if err := rows.Scan(&typ, &d); err != nil {
			return "", err
		}
		// the kernel's own verbs are not the client's facts: one list,
		// core's, decides which
		if core.ReservedEventType(typ) || core.TimeEventType(typ) {
			continue
		}
		if earliest == "" || d < earliest {
			earliest = d
		}
	}
	return earliest, rows.Err()
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

// GetEvent reads one fact by its id: of the company acted for, or the
// installation's, when a company is acted for — as every other read is. A
// scope over several companies acting for one once read another's fact by
// its id, and a door undid that company's match under the company acted for
// (review 2026-10-27).
func (s *Store) GetEvent(ctx context.Context, id int64) (core.Event, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT `+eventCols+` FROM event WHERE event_id = $1 AND ($2 = '' OR company = $2 OR COALESCE(company, '') = '')`, id, CompanyOf(ctx))
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
	rows, err := s.Pool.Query(ctx, unmatchedSQL, CompanyOf(ctx))
	if err != nil {
		return nil, err
	}
	return scanEvents(rows)
}

// PendingEvents is the executor's set: raw events no rule has acted on,
// time events included — the worklist is the human view of the same
// question minus the uneventful days.
func (s *Store) PendingEvents(ctx context.Context) ([]core.Event, error) {
	return s.PendingEventsAfter(ctx, 0, 0)
}

// PendingEventsAfter is the part of that set after one event, up to another
// (0: no bound) — what came since the executor last read the log.
func (s *Store) PendingEventsAfter(ctx context.Context, after, upTo int64) ([]core.Event, error) {
	rows, err := s.Pool.Query(ctx, pendingSQL, CompanyOf(ctx), after, upTo)
	if err != nil {
		return nil, err
	}
	return scanEvents(rows)
}

// LastEventID is the newest event the caller's scope sees (0 in an empty log).
func (s *Store) LastEventID(ctx context.Context) (int64, error) {
	var id int64
	err := s.Pool.QueryRow(ctx, lastEventSQL).Scan(&id)
	return id, err
}

// Explained says which of these events some event names as its cause: a
// rule has acted on them.
func (s *Store) Explained(ctx context.Context, ids []int64) (map[int64]bool, error) {
	out := map[int64]bool{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := s.Pool.Query(ctx, explainedSQL, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
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

// LogMark is where the log stands: its highest id and its count. Ids are
// drawn at insert, not at commit, so a later id can commit first; in an
// append-only log the count moves on every commit, and the two together
// change whenever anything landed.
func (s *Store) LogMark(ctx context.Context) (maxID, count int64, err error) {
	err = s.Pool.QueryRow(ctx, `SELECT COALESCE(MAX(event_id), 0), COUNT(*) FROM event`).Scan(&maxID, &count)
	return maxID, count, err
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
	var warrant []byte
	if r.Warrant != nil {
		if err := r.Warrant.ValidateStored(); err != nil {
			return r, fmt.Errorf("rule %s warrant: %w", r.ID, err)
		}
		if warrant, err = json.Marshal(r.Warrant); err != nil {
			return r, err
		}
	}
	err = q.QueryRow(ctx, `
		INSERT INTO rule (rule_id, version, status, priority, effective_from, created_by, description, spec, warrant, company, replaces)
		VALUES ($1, COALESCE((SELECT MAX(version) FROM rule WHERE rule_id = $1), 0) + 1,
		        $2, $3, $4, $5, $6, $7, $8, NULLIF($9, ''), NULLIF($10, ''))
		RETURNING version, created_at`,
		r.ID, r.Status, r.Priority, r.EffectiveFrom, r.CreatedBy, r.Description, spec, warrant, r.Company, r.Replaces,
	).Scan(&r.Version, &r.CreatedAt)
	if IsDuplicate(err) {
		// MAX(version)+1 is not atomic: two drafts of one rule at the same
		// moment collide on the key, and the later one is told why.
		return r, fmt.Errorf("rule %s moved while this version was drafted — redraft from the latest: %w", r.ID, err)
	}
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
		var spec, warrant []byte
		if err := rows.Scan(&r.ID, &r.Version, &r.Status, &r.Priority, &r.EffectiveFrom,
			&r.CreatedBy, &r.Description, &spec, &warrant, &r.CreatedAt, &r.Company, &r.Replaces); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(spec, &r.Spec); err != nil {
			return nil, fmt.Errorf("rule %s v%d spec: %w", r.ID, r.Version, err)
		}
		if len(warrant) > 0 {
			r.Warrant = &core.Warrant{}
			if err := json.Unmarshal(warrant, r.Warrant); err != nil {
				return nil, fmt.Errorf("rule %s v%d warrant: %w", r.ID, r.Version, err)
			}
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
		       created_by, description, spec, warrant, created_at, COALESCE(company, ''), COALESCE(replaces, '')
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

// RuleNote is what the provenance walk says about one rule version: what
// it does and, when it has one, the source it rests on.
type RuleNote struct {
	Description string
	Warrant     *core.Warrant
}

// RuleNotes returns the note of every rule version ever stored, superseded
// ones included: the provenance walk explains history, and history names
// exact versions.
func (s *Store) RuleNotes(ctx context.Context) (map[RuleKey]RuleNote, error) {
	rows, err := s.Pool.Query(ctx, `SELECT rule_id, version, description, warrant FROM rule`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[RuleKey]RuleNote{}
	for rows.Next() {
		var k RuleKey
		var n RuleNote
		var warrant []byte
		if err := rows.Scan(&k.ID, &k.Version, &n.Description, &warrant); err != nil {
			return nil, err
		}
		if len(warrant) > 0 {
			n.Warrant = &core.Warrant{}
			if err := json.Unmarshal(warrant, n.Warrant); err != nil {
				return nil, fmt.Errorf("rule %s v%d warrant: %w", k.ID, k.Version, err)
			}
		}
		out[k] = n
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
		if errors.Is(err, pgx.ErrNoRows) {
			return core.ObjectType{}, fmt.Errorf("object type %q: %w", name, ErrNotFound)
		}
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
	var company *string
	if o.Company != "" {
		company = &o.Company
	}
	_, err = q.Exec(ctx, `
		INSERT INTO object (object_id, object_type, type_version, state, source_event_id, rule_id, rule_version, company)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		o.ID, o.Type, o.TypeVersion, state, o.SourceEventID, o.RuleID, o.RuleVersion, company)
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

func scanObjects(rows pgx.Rows) ([]core.Object, error) {
	defer rows.Close()
	var out []core.Object
	for rows.Next() {
		var o core.Object
		var state []byte
		var company *string
		if err := rows.Scan(&o.ID, &o.Type, &o.TypeVersion, &state,
			&o.SourceEventID, &o.RuleID, &o.RuleVersion, &company); err != nil {
			return nil, err
		}
		if company != nil {
			o.Company = *company
		}
		if err := json.Unmarshal(state, &o.State); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

const objectCols = `object_id, object_type, type_version, state, source_event_id, rule_id, rule_version, company`

// The partition key as the caller's scope (DIRECTION 2026-10-08, scale rule
// 1 and security rule 2): every store call already carries a context, and
// the context carries the company whose facts the caller is in. An append
// stamps it on the event when the event names none; a read of objects or
// facts sees that company's alone when a scope is set, everything when it is
// not — the latter is today's whole-log behavior, and the enforcement that
// will key on the same column stays out, as the note says.
type companyKey struct{}
type scopeKey struct{}

// scopeVal is a context's scope: the set of companies it sees. A context
// without one is the owner's — the kernel's own work, the whole log.
type scopeVal struct{ set []string }

// WithOwner clears a context's scope: the kernel's own eyes — replay,
// backup, the pack loader, the clock and the adapters fanning out per
// company. A request from a person never runs as the owner.
func WithOwner(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, companyKey{}, "")
	return context.WithValue(ctx, scopeKey{}, nil)
}

// WithInstallation scopes a context to the installation's own rows alone —
// the pack's master data, the kernel's verbs, facts from before companies
// existed — acting for no company. A request naming no company runs so.
func WithInstallation(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, companyKey{}, "")
	return context.WithValue(ctx, scopeKey{}, scopeVal{set: []string{}})
}

// WithCompany makes a context act for one company and see it alone (and
// the installation's rows), whatever it saw before: keeping an office's
// whole desk in view let a chain acting for one client read and amend
// another's rows (review 2026-10-21, B1). An empty company is the owner's
// (WithOwner): the kernel's own work, never a request's.
func WithCompany(ctx context.Context, company string) context.Context {
	if company == "" {
		return WithOwner(ctx)
	}
	ctx = context.WithValue(ctx, companyKey{}, company)
	return context.WithValue(ctx, scopeKey{}, scopeVal{set: []string{company}})
}

// WithScope scopes a context to a set of companies it sees — an accounting
// office and its clients — acting for none until ActFor names one of them.
// Keys are sorted and deduplicated; an empty set is the installation's rows
// alone.
func WithScope(ctx context.Context, companies []string) context.Context {
	set := map[string]bool{}
	for _, c := range companies {
		if c = strings.TrimSpace(c); c != "" {
			set[c] = true
		}
	}
	scope := make([]string, 0, len(set))
	for c := range set {
		scope = append(scope, c)
	}
	sort.Strings(scope)
	ctx = context.WithValue(ctx, companyKey{}, "")
	return context.WithValue(ctx, scopeKey{}, scopeVal{set: scope})
}

// ActFor names the company a scoped context acts for — one of its scope —
// keeping the scope as it is: an office's person opening one client's
// books writes that client's facts and still sees every client. An empty
// company acts for none.
func ActFor(ctx context.Context, company string) context.Context {
	return context.WithValue(ctx, companyKey{}, company)
}

// CompanyOf is the company a context acts for; empty means none (an office
// reading its clients, the installation's own work, or the owner).
func CompanyOf(ctx context.Context) string {
	c, _ := ctx.Value(companyKey{}).(string)
	return c
}

// ScopeOf is the set of companies a context sees and whether it is scoped
// at all: unscoped is the owner's whole log; scoped with an empty set is
// the installation's rows alone.
func ScopeOf(ctx context.Context) ([]string, bool) {
	v, ok := ctx.Value(scopeKey{}).(scopeVal)
	if !ok {
		return nil, false
	}
	return v.set, true
}

// EventByDedupKey is the event a key names within a company (the
// installation's when empty), or ErrNotFound.
func (s *Store) EventByDedupKey(ctx context.Context, company, key string) (core.Event, error) {
	rows, err := s.Pool.Query(ctx, `SELECT `+eventCols+` FROM event WHERE COALESCE(company, '') = $1 AND dedup_key = $2`, company, key)
	if err != nil {
		return core.Event{}, err
	}
	evs, err := scanEvents(rows)
	if err != nil {
		return core.Event{}, err
	}
	if len(evs) == 0 {
		return core.Event{}, fmt.Errorf("event %q: %w", key, ErrNotFound)
	}
	return evs[0], nil
}

// Companies lists the companies the log holds facts for, in key order.
func (s *Store) Companies(ctx context.Context) ([]string, error) {
	rows, err := s.Pool.Query(ctx, `SELECT DISTINCT company FROM event WHERE company IS NOT NULL AND company <> '' ORDER BY 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) ObjectsByType(ctx context.Context, typ string) ([]core.Object, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT `+objectCols+` FROM object WHERE object_type = $1 AND ($2 = '' OR company = $2 OR COALESCE(company, '') = '') ORDER BY source_event_id, object_id`, typ, CompanyOf(ctx))
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
		`SELECT object_type, COUNT(*) FROM object WHERE ($1 = '' OR company = $1 OR COALESCE(company, '') = '') GROUP BY object_type`, CompanyOf(ctx))
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
	rows, err := s.Pool.Query(ctx, `SELECT `+objectCols+` FROM object WHERE ($1 = '' OR company = $1 OR COALESCE(company, '') = '') ORDER BY object_id`, CompanyOf(ctx))
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
		`SELECT object_id FROM object WHERE object_type = $1 AND state->>$2 = $3 AND ($4 = '' OR company = $4 OR COALESCE(company, '') = '') ORDER BY source_event_id, object_id`,
		typ, field, value, CompanyOf(ctx))
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

// AccountBalance is the debits less the credits of one account's postings in
// a book dated from..to inclusive, in minor units — the formula language's
// balance(): one keyed aggregate, never a fold. A posting with no book is in
// the main book.
func (s *Store) AccountBalance(ctx context.Context, book, account, from, to string) (int64, error) {
	var minor int64
	err := s.Pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(CASE WHEN state->>'side' = 'debit' THEN (state->>'amount')::bigint ELSE -(state->>'amount')::bigint END), 0)::bigint
		 FROM object WHERE object_type = 'posting' AND state->>'account' = $1 AND COALESCE(state->>'book', 'main') = $2
		   AND state->>'date' >= $3 AND state->>'date' <= $4 AND ($5 = '' OR company = $5 OR COALESCE(company, '') = '')`,
		account, book, from, to, CompanyOf(ctx)).Scan(&minor)
	return minor, err
}

// GetObject reads one object by its id, the company acted for's or the
// installation's alone when a company is acted for (GetEvent).
func (s *Store) GetObject(ctx context.Context, id string) (core.Object, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT `+objectCols+` FROM object WHERE object_id = $1 AND ($2 = '' OR company = $2 OR COALESCE(company, '') = '')`, id, CompanyOf(ctx))
	if err != nil {
		return core.Object{}, err
	}
	objs, err := scanObjects(rows)
	if err != nil {
		return core.Object{}, err
	}
	if len(objs) == 0 {
		return core.Object{}, fmt.Errorf("object %q: %w", id, ErrNotFound)
	}
	return objs[0], nil
}
