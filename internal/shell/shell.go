// Package shell is the HTTP face of the kernel: a JSON API plus the embedded
// web shell (activity bar → submenu → tabs). It renders view notions
// generically; it knows nothing about invoices (SPEC §5).
package shell

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"rhea/internal/adapter"
	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/network"
	"rhea/internal/rights"
	"rhea/internal/store"
	"rhea/web"
)

type Server struct {
	// RequirePerson makes the kernel serve signed-in persons only: a request
	// naming no person answers 401, except the machine's own routes. Set
	// where a face is the only door from outside (white01); off on the
	// bench, where the owner's shell and the tools run unscoped.
	RequirePerson bool
	// FaceKey is the key the face proves itself with (RHEA_FACE_KEY): a
	// request naming a person is believed only when it carries it
	// (X-Rhea-Face). A kernel serving persons only must know one; without
	// it, it believes no person.
	FaceKey  string
	Store    *store.Store
	Exec     *exec.Executor
	Agent    *agent.Agent
	DuckPath string
	// Clock serves the Time surface; the zero value reads the wall clock,
	// tests inject their own day.
	Clock adapter.Clock
	// Network is where Rhea learns across installations; nil runs alone.
	Network *network.Network
	// duck serializes the analysis side: each analysis read rebuilds the
	// DuckDB projection and queries it, and two at once collide on the
	// rebuild ("write-write conflict") or read a table mid-rebuild.
	duck sync.Mutex
	// live is the one projection listener, fanning out to every open tab:
	// a pooled connection per tab exhausted the pool at four tabs.
	live struct {
		sync.Mutex
		running bool
		cancel  context.CancelFunc
		subs    map[chan string]struct{}
	}
}

// subscribeLive hands out a channel of projection notices, starting the
// one listener when none runs and stopping it — its connection back in the
// pool — when the last tab leaves; the listener's end closes every
// channel, so tabs reconnect and the next subscriber starts it again.
func (s *Server) subscribeLive() (<-chan string, func(), error) {
	s.live.Lock()
	if s.live.subs == nil {
		s.live.subs = map[chan string]struct{}{}
	}
	start := !s.live.running
	if start {
		s.live.running = true // reserved: the listener is being started
	}
	s.live.Unlock()
	if start {
		// The listener's connection is acquired outside the lock: with the
		// pool exhausted, holding the lock here would block every fan-out
		// and every unsubscribe, so closing tabs could not free connections.
		ctx, cancel := context.WithCancel(context.Background())
		notices, err := s.Store.ProjectionNotices(ctx)
		s.live.Lock()
		if err != nil {
			cancel()
			s.live.running = false
			for ch := range s.live.subs { // tabs that joined meanwhile reconnect
				close(ch)
				delete(s.live.subs, ch)
			}
			s.live.Unlock()
			return nil, nil, err
		}
		s.live.cancel = cancel
		s.live.Unlock()
		go func() {
			for p := range notices {
				s.live.Lock()
				for ch := range s.live.subs {
					select {
					case ch <- p:
					default: // a slow tab drops a notice; a notice only means "look again"
					}
				}
				s.live.Unlock()
			}
			s.live.Lock()
			s.live.running = false
			for ch := range s.live.subs {
				close(ch)
				delete(s.live.subs, ch)
			}
			s.live.Unlock()
		}()
	}
	s.live.Lock()
	defer s.live.Unlock()
	ch := make(chan string, 16)
	s.live.subs[ch] = struct{}{}
	return ch, func() {
		s.live.Lock()
		defer s.live.Unlock()
		if _, ok := s.live.subs[ch]; ok {
			delete(s.live.subs, ch)
			close(ch)
		}
		if len(s.live.subs) == 0 && s.live.cancel != nil {
			s.live.cancel() // nobody listens: the listener ends and frees its connection
			s.live.cancel = nil
		}
	}, nil
}

// knowledge is what Rhea has learned across installations, or nil when no
// network is connected (or it cannot be read — learning is never a reason
// for the shell to fail).
func (s *Server) knowledge(ctx context.Context) *network.Knowledge {
	if s.Network == nil {
		return nil
	}
	k, err := s.Network.Learn(ctx)
	if err != nil {
		log.Printf("network: %v", err)
		return nil
	}
	return &k
}

func (s *Server) Handler() http.Handler {
	return logged(hardened(s.principalHandler(s.mux())))
}

// Drain ends the live streams, so a stopping server waits only for the
// requests that finish — a stream never does (review 2026-10-21, G6).
func (s *Server) Drain() {
	s.live.Lock()
	cancel := s.live.cancel
	s.live.Unlock()
	if cancel != nil {
		cancel()
	}
}

// logged writes one line per request — method, path, status, time taken,
// and the person and company the request named: the server kept no trace
// of what it served (review 2026-10-21, G6).
func logged(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s %d %dms person=%q company=%q", r.Method, r.URL.Path, rec.status,
			time.Since(start).Milliseconds(), r.Header.Get("X-Rhea-Person"), r.Header.Get("X-Rhea-Company"))
	})
}

// statusRecorder notes the status a handler wrote; it flushes, so the
// live stream still streams.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (r *statusRecorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

// maxBody bounds what a door reads: a document through the receive door
// is a few megabytes of base64 at most; a multi-gigabyte body is refused,
// not read into memory.
const maxBody = 32 << 20

// hardened keeps one request from taking the server down: a body is read
// up to maxBody, and a panic in a handler answers 500 with its reason in
// the log instead of ending rhea serve.
func hardened(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxBody)
		}
		// A page on any site could POST text/plain to an approval door and
		// act as the human: a write is JSON from the shell's own origin or
		// from no browser at all.
		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
			if err := sameOrigin(r); err != nil {
				writeErr(w, 403, err)
				return
			}
		}
		defer func() {
			if p := recover(); p != nil {
				// the panic's own words go to the log, never to the client (G6)
				log.Printf("%s %s: panic: %v\n%s", r.Method, r.URL.Path, p, debug.Stack())
				writeErr(w, 500, fmt.Errorf("panic: %v", p))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) mux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/nav", s.handleNav)
	mux.HandleFunc("GET /api/apps", s.handleApps)
	mux.HandleFunc("GET /api/views/{id}", s.handleView)
	mux.HandleFunc("GET /api/worklist", s.handleWorklist)
	mux.HandleFunc("GET /api/types", s.handleTypes)
	mux.HandleFunc("GET /api/explain", s.handleExplain)
	mux.HandleFunc("GET /api/objects", s.handleObjects)
	mux.HandleFunc("GET /api/objects/{id}", s.handleObject)
	mux.HandleFunc("GET /api/said", s.handleSaid)
	mux.HandleFunc("POST /api/speak", s.handleSpeak)
	mux.HandleFunc("GET /api/untaught", s.handleUntaught)
	mux.HandleFunc("POST /api/untaught", s.handleNoteUntaught)
	mux.HandleFunc("GET /api/network/learning", s.handleLearning)
	mux.HandleFunc("GET /api/collection", s.handleCollection)
	mux.HandleFunc("POST /api/answer", s.handleAnswer)
	mux.HandleFunc("POST /api/answers/{event}/keep", s.handleKeepAnswer)
	mux.HandleFunc("POST /api/network/lessons", s.handleNoteLesson)
	mux.HandleFunc("GET /api/files", s.handleFiles)
	mux.HandleFunc("GET /api/files/{writer}", s.handleFile)
	mux.HandleFunc("POST /api/files/{writer}/issue", s.handleIssueFile)
	mux.HandleFunc("GET /api/live", s.handleLive)
	mux.HandleFunc("GET /api/rules", s.handleRules)
	mux.HandleFunc("GET /api/promotions", s.handlePromotions)
	mux.HandleFunc("GET /api/changes", s.handleChanges)
	mux.HandleFunc("GET /api/packs", s.handlePacks)
	mux.HandleFunc("POST /api/rules/draft", s.handleDraft)
	mux.HandleFunc("POST /api/rules/{id}/approve", s.handleApprove)
	mux.HandleFunc("POST /api/rules/{id}/simulate", s.handleSimulate)
	mux.HandleFunc("GET /api/bundles", s.handleBundles)
	mux.HandleFunc("POST /api/bundles/draft", s.handleDraftBundle)
	mux.HandleFunc("POST /api/backfill/simulate", s.handleSimulateBackfill)
	mux.HandleFunc("POST /api/backfill/approve", s.handleApproveBackfill)
	mux.HandleFunc("POST /api/bundles/{id}/simulate", s.handleSimulateBundle)
	mux.HandleFunc("GET /api/bundles/{id}/preview", s.handlePreviewBundle)
	mux.HandleFunc("POST /api/bundles/{id}/approve", s.handleApproveBundle)
	mux.HandleFunc("POST /api/bundles/{id}/reject", s.handleRejectBundle)
	mux.HandleFunc("GET /api/activities", s.handleActivities)
	mux.HandleFunc("POST /api/activities/{name}/trigger", s.handleTrigger)
	mux.HandleFunc("POST /api/activities/{name}/approve", s.handleApproveActivity)
	mux.HandleFunc("GET /api/channels", s.handleChannels)
	mux.HandleFunc("GET /api/rights", s.handleRights)
	mux.HandleFunc("GET /api/whoami", s.handleWhoAmI)
	mux.HandleFunc("GET /api/registers", s.handleRegisters)
	mux.HandleFunc("GET /api/registers/{name}", s.handleRegister)
	mux.HandleFunc("GET /api/blobs/{hash}", s.handleBlob)
	mux.HandleFunc("GET /api/time", s.handleTime)
	mux.HandleFunc("POST /api/time/advance", s.handleTimeAdvance)
	mux.HandleFunc("POST /api/time/catchup", s.handleTimeCatchUp)
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("POST /api/time/simulate", s.handleTimeSimulate)
	mux.HandleFunc("POST /api/events", s.handleSubmit)
	mux.Handle("GET /", http.FileServerFS(web.FS))
	return mux
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// sameOrigin refuses a cross-site write: the body must be JSON, the
// browser's own word on where the request came from (Sec-Fetch-Site) must
// be same-origin or absent, and an Origin header must name this host.
func sameOrigin(r *http.Request) error {
	if ct := r.Header.Get("Content-Type"); ct != "" && !strings.HasPrefix(strings.ToLower(ct), "application/json") {
		return fmt.Errorf("a write is application/json, not %s", ct)
	}
	switch site := r.Header.Get("Sec-Fetch-Site"); site {
	case "", "same-origin", "none":
	default:
		return fmt.Errorf("a %s request cannot act here", site)
	}
	if origin := r.Header.Get("Origin"); origin != "" && origin != "null" {
		u, err := url.Parse(origin)
		if err != nil || !strings.EqualFold(u.Host, r.Host) {
			return fmt.Errorf("origin %s is not this shell", origin)
		}
	}
	return nil
}

// decodeBody reads a JSON request body into v; an empty body is no body.
// A body that does not parse, or one with unknown keys, is an error the
// door answers with 400 — an approver's name mis-keyed must not silently
// become "human", since who approved is part of the record.
func decodeBody(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("request body: %w", err)
	}
	return nil
}

func writeErr(w http.ResponseWriter, status int, err error) {
	// What is not there is 404, whatever door it came through: rejecting an
	// unknown bundle answered 400 with "no rows in result set".
	if errors.Is(err, store.ErrNotFound) || errors.Is(err, pgx.ErrNoRows) {
		status = 404
	}
	// A refusal is 403 whatever door it came through: the person was
	// named, the right was not theirs, and the fact is recorded.
	var refused *rights.ErrRefused
	if errors.As(err, &refused) {
		status = 403
	}
	// Busy books are a wait, not a fault: try again (G5).
	if errors.Is(err, exec.ErrBusy) {
		status = 503
	}
	// The server's own failures, and anything the database said as it said
	// it, stay in the log: the face got Postgres's words — tables, columns,
	// constraints — and a panic's text (review 2026-10-21, G6). A duplicate
	// keeps its words: "done already" is the person's answer.
	var pgErr *pgconn.PgError
	if (status >= 500 && !errors.Is(err, exec.ErrBusy)) || (errors.As(err, &pgErr) && pgErr.Code != "23505") {
		log.Printf("%d: %v", status, err)
		status = max(status, 500)
		err = errors.New("the server failed on this request — what happened is in its log")
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

// --- navigation -----------------------------------------------------------

type navFunction struct {
	Function string         `json:"function"`
	Views    []core.ViewDef `json:"views"`
}

type navDomain struct {
	Domain    string        `json:"domain"`
	Functions []navFunction `json:"functions"`
}

// effectiveViewDefs is what the shell navigates and serves: the stored view
// definitions plus, for every object type missing a list or detail, the view
// derived from the type itself (core.DerivedListView/DerivedDetailView).
// Stored views are the exceptions; derivation is the default.
func (s *Server) effectiveViewDefs(ctx context.Context) ([]core.ViewDef, error) {
	all, err := s.Store.LatestViewDefs(ctx)
	if err != nil {
		return nil, err
	}
	allTypes, err := s.Store.ListObjectTypes(ctx)
	if err != nil {
		return nil, err
	}
	// what the company acted for is offered: a pack installed per company
	// only where it is installed (DIRECTION 2026-10-24, step 2)
	offer, err := s.Exec.Offered(ctx, store.CompanyOf(ctx))
	if err != nil {
		return nil, err
	}
	vds := make([]core.ViewDef, 0, len(all))
	for _, v := range all {
		if offer.View(v.ID) {
			vds = append(vds, v)
		}
	}
	types := make([]core.ObjectType, 0, len(allTypes))
	for _, t := range allTypes {
		if offer.Type(t.Name) {
			types = append(types, t)
		}
	}
	hasList, hasDetail := map[string]bool{}, map[string]bool{}
	for _, v := range vds {
		switch v.Notion {
		case "list":
			var sp core.ListSpec
			if json.Unmarshal(v.Spec, &sp) == nil {
				hasList[sp.ObjectType] = true
			}
		case "detail":
			var sp core.DetailSpec
			if json.Unmarshal(v.Spec, &sp) == nil {
				hasDetail[sp.ObjectType] = true
			}
		}
	}
	for _, t := range types {
		if !hasList[t.Name] {
			vds = append(vds, core.DerivedListView(t))
		}
		if !hasDetail[t.Name] {
			vds = append(vds, core.DerivedDetailView(t))
		}
	}
	return vds, nil
}

// handleNav builds the activity bar and submenus from the effective view
// definitions, then appends the system functions every domain carries.
func (s *Server) handleNav(w http.ResponseWriter, r *http.Request) {
	vds, err := s.effectiveViewDefs(r.Context())
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	type key struct{ domain, function string }
	order := []key{}
	grouped := map[key][]core.ViewDef{}
	for _, v := range vds {
		k := key{v.Domain, v.Function}
		if _, seen := grouped[k]; !seen {
			order = append(order, k)
		}
		grouped[k] = append(grouped[k], v)
	}
	domains := []navDomain{}
	byDomain := map[string]int{}
	for _, k := range order {
		i, ok := byDomain[k.domain]
		if !ok {
			i = len(domains)
			byDomain[k.domain] = i
			domains = append(domains, navDomain{Domain: k.domain})
		}
		domains[i].Functions = append(domains[i].Functions,
			navFunction{Function: k.function, Views: grouped[k]})
	}
	// System functions: rendered by the shell natively, present in every domain.
	for i := range domains {
		domains[i].Functions = append(domains[i].Functions,
			navFunction{Function: "worklist"}, navFunction{Function: "rules"},
			navFunction{Function: "language"}, navFunction{Function: "time"})
	}
	writeJSON(w, 200, domains)
}

// --- views ----------------------------------------------------------------

// handleView serves a view definition together with its data, shaped per
// notion so the client stays a dumb renderer.
func (s *Server) handleView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vds, err := s.effectiveViewDefs(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	var vd *core.ViewDef
	for i := range vds {
		if vds[i].ID == r.PathValue("id") {
			vd = &vds[i]
			break
		}
	}
	if vd == nil {
		writeErr(w, 404, fmt.Errorf("view %q not found", r.PathValue("id")))
		return
	}
	if !s.allowView(w, r, vd) {
		return
	}
	switch vd.Notion {
	case "list":
		s.serveList(ctx, w, vd)
	case "detail":
		s.serveDetail(ctx, w, vd, r.URL.Query().Get("object_id"))
	case "analysis":
		s.serveAnalysis(ctx, w, vd)
	case "scheduling":
		s.serveScheduling(ctx, w, vd)
	case "matching":
		s.serveMatching(ctx, w, vd)
	case "narrative":
		s.serveNarrative(ctx, w, vd)
	case "queue":
		s.serveQueue(ctx, w, vd)
	case "words":
		s.serveWords(ctx, w, vd)
	case "statement":
		s.serveStatement(ctx, w, vd, r.URL.Query().Get("object_id"))
	case "report":
		s.serveReport(ctx, w, vd)
	case "rights":
		s.serveRights(ctx, w, vd, r.URL.Query().Get("company"))
	default:
		writeErr(w, 400, fmt.Errorf("notion %q not renderable", vd.Notion))
	}
}

// The typed view API (DIRECTION 2026-10-05): the contract carries semantics,
// renderers decide formatting. Columns declare a field and its type; cells
// carry the canonical boundary encoding (money as "1234.56" decimal strings,
// invariant 6), and a ref cell carries the referenced object's id next to its
// label so any renderer can both show it and follow it.

type columnOut struct {
	Field string `json:"field"`
	Label string `json:"label"`
	Type  string `json:"type,omitempty"` // "" = undeclared: render as a bare string
}

type cellOut struct {
	V  string `json:"v"`            // canonical value, never locale-formatted
	ID string `json:"id,omitempty"` // refs only: the referenced object id
	// Ref is a narrative's facts only: the reference the host writes this
	// amount by, [F2: −665,83], so the kernel checks it against this cell.
	Ref string `json:"ref,omitempty"`
	// Detail is the view that opens the referenced object — with derived
	// defaults, every type has one, so every ref is a door.
	Detail string `json:"detail,omitempty"`
}

// typedCell encodes one state value. A ref resolves to the referenced
// object's label_field so humans see "ACME Sp. z o.o.", not "company-7"; a
// ref that does not resolve (old type versions hold plain strings) falls
// back to the raw value, and the id rides along either way.
func (s *Server) typedCell(ctx context.Context, v any, fd core.FieldDef, detailFor map[string]string) cellOut {
	return cellWith(v, fd, func(id string) (string, string) { return s.refInfo(ctx, id) }, detailFor)
}

// cellWith is the one encoding path for state values, with a caller-chosen
// ref resolver — the simulation diff resolves refs against the hypothetical
// world first, every live view against the cache.
func cellWith(v any, fd core.FieldDef, resolve func(id string) (label, typ string), detailFor map[string]string) cellOut {
	if v == nil {
		return cellOut{}
	}
	if _, isRef := core.RefTarget(fd.Type); isRef {
		if id, ok := v.(string); ok && id != "" {
			c := cellOut{V: id, ID: id}
			label, typ := resolve(id)
			if label != "" {
				c.V = label
			}
			c.Detail = detailFor[typ]
			return c
		}
	}
	if fd.Type == "money" {
		// Money arrives from state JSON as float64 minor units (exact for
		// int64 magnitudes that fit 2^53) and leaves as a decimal string.
		switch n := v.(type) {
		case float64:
			return cellOut{V: core.FormatMoney(int64(n))}
		case int64:
			return cellOut{V: core.FormatMoney(n)}
		}
	}
	return cellOut{V: core.Text(v)}
}

// refInfo resolves a referenced object to its label and type.
func (s *Server) refInfo(ctx context.Context, objectID string) (label, typ string) {
	o, err := s.Store.GetObject(ctx, objectID)
	if err != nil {
		return "", ""
	}
	ot, err := s.Store.GetObjectType(ctx, o.Type)
	if err != nil || ot.LabelField == "" {
		return "", o.Type
	}
	label, _ = o.State[ot.LabelField].(string)
	return label, o.Type
}

func (s *Server) serveList(ctx context.Context, w http.ResponseWriter, vd *core.ViewDef) {
	var spec core.ListSpec
	if err := json.Unmarshal(vd.Spec, &spec); err != nil {
		writeErr(w, 500, err)
		return
	}
	objType, err := s.Store.GetObjectType(ctx, spec.ObjectType)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	objs, err := s.Store.ObjectsByType(ctx, spec.ObjectType)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	vds, err := s.effectiveViewDefs(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	// With derived defaults in the effective set, every type has a detail:
	// one for this list's rows to open, one behind every ref cell.
	detailFor := detailViewByType(vds)
	cols := make([]columnOut, len(spec.Columns))
	fds := make([]core.FieldDef, len(spec.Columns))
	for i, c := range spec.Columns {
		fd, _ := objType.Field(c.Field)
		fds[i] = fd
		cols[i] = columnOut{Field: c.Field, Label: c.Label, Type: fd.Type}
	}
	rows := make([][]cellOut, 0, len(objs))
	ids := make([]string, 0, len(objs))
	for _, o := range objs {
		row := make([]cellOut, len(spec.Columns))
		for i, c := range spec.Columns {
			row[i] = s.typedCell(ctx, o.State[c.Field], fds[i], detailFor)
		}
		rows = append(rows, row)
		ids = append(ids, o.ID)
	}
	writeJSON(w, 200, map[string]any{"view": vd, "columns": cols, "rows": rows,
		"object_ids": ids, "detail_view_id": detailFor[spec.ObjectType]})
}

func (s *Server) serveDetail(ctx context.Context, w http.ResponseWriter, vd *core.ViewDef, objectID string) {
	var spec core.DetailSpec
	if err := json.Unmarshal(vd.Spec, &spec); err != nil {
		writeErr(w, 500, err)
		return
	}
	if objectID == "" {
		writeErr(w, 400, fmt.Errorf("detail view needs ?object_id="))
		return
	}
	o, err := s.Store.GetObject(ctx, objectID)
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	objType, err := s.Store.GetObjectType(ctx, o.Type)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	type fieldOut struct {
		Field  string `json:"field"`
		Label  string `json:"label"`
		Type   string `json:"type,omitempty"`
		V      string `json:"v"`
		ID     string `json:"id,omitempty"`
		Detail string `json:"detail,omitempty"`
	}
	type sectionOut struct {
		Title  string     `json:"title"`
		Fields []fieldOut `json:"fields"`
	}
	vds, err := s.effectiveViewDefs(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	detailFor := detailViewByType(vds)
	sections := make([]sectionOut, 0, len(spec.Sections))
	for _, sec := range spec.Sections {
		so := sectionOut{Title: sec.Title}
		for _, f := range sec.Fields {
			fd, _ := objType.Field(f)
			c := s.typedCell(ctx, o.State[f], fd, detailFor)
			so.Fields = append(so.Fields, fieldOut{Field: f, Label: f, Type: fd.Type,
				V: c.V, ID: c.ID, Detail: c.Detail})
		}
		sections = append(sections, so)
	}
	// Provenance crosses as data (invariant 5); the renderer phrases it.
	// Activities ride along: the verbs whose ref inputs name this type — a
	// detail view offers what can be done to what it shows. Amendments are
	// the second half of the explanation: the materialization says why the
	// object exists, each amendment why it is what it is now.
	amendEvents, err := s.Store.AmendmentsOf(ctx, objectID)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	amendments := []map[string]any{}
	for _, ev := range amendEvents {
		var am core.AmendedObject
		if json.Unmarshal(ev.Payload, &am) != nil {
			continue
		}
		// each moved value as the typed cell its field renders: money in
		// the boundary form, a ref by its label — never raw minor units
		set := map[string]any{}
		for k, v := range am.Set {
			fd, _ := objType.Field(k)
			set[k] = s.typedCell(ctx, v, fd, detailFor)
		}
		amendments = append(amendments, map[string]any{
			"event_id": ev.ID, "rule_id": ev.RuleID, "rule_version": ev.RuleVersion,
			"occurred_at": ev.OccurredAt, "set": set,
		})
	}
	writeJSON(w, 200, map[string]any{
		"view": vd, "object": o, "sections": sections,
		"provenance": map[string]any{"event_id": o.SourceEventID,
			"rule_id": o.RuleID, "rule_version": o.RuleVersion},
		"activities": s.verbsForType(ctx, o.Type),
		"amendments": amendments,
	})
}

func (s *Server) serveAnalysis(ctx context.Context, w http.ResponseWriter, vd *core.ViewDef) {
	spec, err := s.analysisSpec(ctx, vd.Spec)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	// The read side is derived state: rebuilt from the log so it is always
	// fresh. At M0 scale this is cheap. Columns pair up with the spec's
	// declarations by position; a declared money column arrives from DuckDB
	// as minor units (BIGINT, invariant 6) and crosses as the decimal string.
	cols, rows, rowsOf, err := s.analysisData(ctx, spec)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	out := map[string]any{"view": vd, "columns": cols, "rows": rows}
	if rowsOf != nil {
		out["row_objects"] = rowsOf // what each row is made of: ?view=&row=&col= on /api/explain says how
	}
	writeJSON(w, 200, out)
}

// --- live screens -----------------------------------------------------------

// handleLive streams projection notices as server-sent events: one stream per
// browser tab, one "projection" event per committed write, payload naming the
// touched object types. Nobody refreshes (DIRECTION: Alpha's live screens).
// The stream is a signal, never data — clients re-read through the API.
func (s *Server) handleLive(w http.ResponseWriter, r *http.Request) {
	fl, ok := w.(http.Flusher)
	if !ok {
		writeErr(w, 500, fmt.Errorf("streaming unsupported"))
		return
	}
	notices, unsubscribe, err := s.subscribeLive()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	defer unsubscribe()
	// A booking's notice names its company ("company|types"): the stream
	// passes on what the listener's scope sees — its companies and the
	// installation's — so one company's screen never watches another book
	// (review 2026-10-21, B7). A notice naming none (a replay) is everyone's.
	set, scoped := store.ScopeOf(r.Context())
	sees := func(company string) bool { return !scoped || company == "" || slices.Contains(set, company) }
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(200)
	fl.Flush()
	ping := time.NewTicker(25 * time.Second)
	defer ping.Stop()
	for {
		select {
		case p, ok := <-notices:
			if !ok {
				return
			}
			if company, types, tagged := strings.Cut(p, "|"); tagged {
				if !sees(company) {
					continue
				}
				p = types
			}
			fmt.Fprintf(w, "event: projection\ndata: %s\n\n", p)
			fl.Flush()
		case <-ping.C:
			fmt.Fprint(w, ": ping\n\n") // keeps idle streams alive
			fl.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// --- worklist, rules, actions ---------------------------------------------

func (s *Server) handleWorklist(w http.ResponseWriter, r *http.Request) {
	if !s.allowSystem(w, r, ViewWorklist) {
		return
	}
	events, err := s.Store.UnmatchedRawEvents(r.Context())
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	types, err := s.Store.ListObjectTypes(r.Context())
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	typeNames := make([]string, len(types))
	for i, t := range types {
		typeNames[i] = t.Name
	}
	writeJSON(w, 200, map[string]any{"events": events, "object_types": typeNames})
}

func (s *Server) handleRules(w http.ResponseWriter, r *http.Request) {
	if !s.allowSystem(w, r, ViewGate) {
		return
	}
	rules, err := s.Store.LatestRules(r.Context())
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	approvals, err := s.ruleApprovals(r.Context())
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	out := make([]ruleOut, len(rules))
	for i, ru := range rules {
		out[i] = ruleOut{Rule: ru}
		if a, ok := approvals[approvalKey(ru.ID, ru.Version)]; ok && ru.Status != core.StatusDraft {
			out[i].Approval = &a
		}
	}
	writeJSON(w, 200, out)
}

func (s *Server) handleDraft(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Intent        string `json:"intent"`
		SampleEventID int64  `json:"sample_event_id"`
		ObjectType    string `json:"object_type"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, 400, err)
		return
	}
	ctx := r.Context()
	events, err := s.Store.UnmatchedRawEvents(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	var sample *core.Event
	for i := range events {
		if events[i].ID == req.SampleEventID {
			sample = &events[i]
			break
		}
	}
	if sample == nil {
		writeErr(w, 404, fmt.Errorf("event %d not in worklist", req.SampleEventID))
		return
	}
	if req.ObjectType != "" {
		if _, err := s.Store.GetObjectType(ctx, req.ObjectType); err != nil {
			writeErr(w, 404, err)
			return
		}
	}
	// Named as the bundle door names it: drafted for a company, its own —
	// and a desk acting for none is refused before anything is asked.
	company, known, err := draftingFor(ctx, s.Store)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	// The ask goes through the draft_rule door first: the request is in the
	// log as rule.draft_requested before any model answers. Drafting itself
	// cannot be a kernel reaction — reactions are deterministic, the agent is
	// not — so the shell consumes the request here, as the agent's runner;
	// true agent-as-user (its own worklist) is phase (a).
	if _, err := s.Exec.TriggerActivity(ctx, "draft_rule", map[string]any{
		"intent": req.Intent, "sample_event_id": req.SampleEventID, "object_type": req.ObjectType,
	}, "shell", ""); err != nil {
		writeErr(w, 400, err)
		return
	}
	ask, err := DraftAsk(ctx, s.Store, s.knowledge(ctx), req.Intent, *sample, req.ObjectType)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	draft, err := s.Agent.DraftRule(ctx, ask)
	if err != nil {
		writeErr(w, 502, err) // the request event stays — the log is honest about unanswered asks
		return
	}
	id, replaces := ownRule(known, company, draft.RuleID)
	rule, err := s.Store.InsertRuleVersion(ctx, core.Rule{
		ID: id, Status: core.StatusDraft, Priority: draft.Priority,
		EffectiveFrom: sample.OccurredAt, CreatedBy: "agent:" + agent.Model(),
		Description: draft.Description, Spec: draft.Spec,
		Company: company, Replaces: replaces,
	})
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, rule)
}

func (s *Server) handleApprove(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ApprovedBy string `json:"approved_by"`
		Reason     string `json:"reason"` // why this version replaces the reading before it
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if req.ApprovedBy == "" {
		req.ApprovedBy = "human"
	}
	rule, booked, procErrs, err := s.Exec.ApproveRuleWithReason(r.Context(), r.PathValue("id"),
		req.ApprovedBy, s.Clock.Today(), req.Reason)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	resp := map[string]any{"rule": rule, "booked": booked}
	if len(procErrs) > 0 {
		resp["errors"] = fmt.Sprintf("%v", procErrs)
	}
	writeJSON(w, 200, resp)
}

// handleSimulate dry-runs a rule against the whole log in memory (SPEC M1).
// Read-only: the human sees the diff before deciding to approve — shaped
// field-level and typed (simdiff.go), because the approval gate is the trust
// boundary and what the human reads there must be readable.
func (s *Server) handleSimulate(w http.ResponseWriter, r *http.Request) {
	if !s.allowSystem(w, r, ViewGate) { // the gate's dry run asks what the gate asks (A9)
		return
	}
	ctx := r.Context()
	diff, err := s.Exec.SimulateRule(ctx, r.PathValue("id"))
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	out, err := s.renderSimDiff(ctx, diff, nil)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, out)
}

// handleSubmit is the submit_event door over HTTP: the open door, declared
// as data like every other verb, so the event carries its activity stamp.
func (s *Server) handleSubmit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EventType  string         `json:"event_type"`
		OccurredAt string         `json:"occurred_at"`
		Payload    map[string]any `json:"payload"`
		DedupKey   string         `json:"dedup_key"`
		Actor      string         `json:"actor"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if req.Actor == "" {
		req.Actor = "shell"
	}
	if p := rights.PrincipalOf(r.Context()); p != nil {
		req.Actor = p.Handle // the server says who, whatever the page sent
	}
	t, err := s.Exec.TriggerActivity(r.Context(), "submit_event", map[string]any{
		"event_type": req.EventType, "occurred_at": req.OccurredAt, "payload": req.Payload,
	}, req.Actor, req.DedupKey)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	resp := map[string]any{"event_id": t.EventID, "booked": t.Booked}
	if len(t.Errors) > 0 {
		resp["errors"] = fmt.Sprintf("%v", t.Errors)
	}
	writeJSON(w, 200, resp)
}

func parseInt(v string) (int64, error) { return strconv.ParseInt(v, 10, 64) }
