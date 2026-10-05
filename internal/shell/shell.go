// Package shell is the HTTP face of the kernel: a JSON API plus the embedded
// web shell (activity bar → submenu → tabs). It renders view notions
// generically; it knows nothing about invoices (SPEC §5).
package shell

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/project"
	"rhea/internal/store"
	"rhea/web"
)

type Server struct {
	Store    *store.Store
	Exec     *exec.Executor
	Agent    *agent.Agent
	DuckPath string
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/nav", s.handleNav)
	mux.HandleFunc("GET /api/views/{id}", s.handleView)
	mux.HandleFunc("GET /api/worklist", s.handleWorklist)
	mux.HandleFunc("GET /api/types", s.handleTypes)
	mux.HandleFunc("GET /api/explain", s.handleExplain)
	mux.HandleFunc("GET /api/live", s.handleLive)
	mux.HandleFunc("GET /api/rules", s.handleRules)
	mux.HandleFunc("POST /api/rules/draft", s.handleDraft)
	mux.HandleFunc("POST /api/rules/{id}/approve", s.handleApprove)
	mux.HandleFunc("POST /api/rules/{id}/simulate", s.handleSimulate)
	mux.HandleFunc("POST /api/events", s.handleSubmit)
	mux.Handle("GET /", http.FileServerFS(web.FS))
	return mux
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, err error) {
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
	vds, err := s.Store.LatestViewDefs(ctx)
	if err != nil {
		return nil, err
	}
	types, err := s.Store.ListObjectTypes(ctx)
	if err != nil {
		return nil, err
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
			navFunction{Function: "language"})
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
	switch vd.Notion {
	case "list":
		s.serveList(ctx, w, vd)
	case "detail":
		s.serveDetail(ctx, w, vd, r.URL.Query().Get("object_id"))
	case "analysis":
		s.serveAnalysis(ctx, w, vd)
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
	// Detail is the view that opens the referenced object — with derived
	// defaults, every type has one, so every ref is a door.
	Detail string `json:"detail,omitempty"`
}

// typedCell encodes one state value. A ref resolves to the referenced
// object's label_field so humans see "ACME Sp. z o.o.", not "company-7"; a
// ref that does not resolve (old type versions hold plain strings) falls
// back to the raw value, and the id rides along either way.
func (s *Server) typedCell(ctx context.Context, v any, fd core.FieldDef, detailFor map[string]string) cellOut {
	if v == nil {
		return cellOut{}
	}
	if _, isRef := core.RefTarget(fd.Type); isRef {
		if id, ok := v.(string); ok && id != "" {
			c := cellOut{V: id, ID: id}
			label, typ := s.refInfo(ctx, id)
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
	return cellOut{V: fmt.Sprintf("%v", v)}
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
	writeJSON(w, 200, map[string]any{
		"view": vd, "object": o, "sections": sections,
		"provenance": map[string]any{"event_id": o.SourceEventID,
			"rule_id": o.RuleID, "rule_version": o.RuleVersion},
	})
}

func (s *Server) serveAnalysis(ctx context.Context, w http.ResponseWriter, vd *core.ViewDef) {
	var spec core.AnalysisSpec
	if err := json.Unmarshal(vd.Spec, &spec); err != nil {
		writeErr(w, 500, err)
		return
	}
	// The read side is derived state: rebuild from the log so it is always
	// fresh. At M0 scale this is cheap.
	if _, err := project.Rebuild(ctx, s.Store, s.DuckPath); err != nil {
		writeErr(w, 500, err)
		return
	}
	cols, rows, err := project.Query(ctx, s.DuckPath, spec.SQL)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	// Columns pair up with the spec's declarations by position; the SQL name
	// is the fallback. A declared money column arrives from DuckDB as minor
	// units (BIGINT, invariant 6) and crosses as the decimal string.
	outCols := make([]columnOut, len(cols))
	for i, c := range cols {
		outCols[i] = columnOut{Field: c, Label: c}
		if i < len(spec.Columns) {
			sc := spec.Columns[i]
			if sc.Field != "" {
				outCols[i].Field = sc.Field
			}
			if sc.Label != "" {
				outCols[i].Label = sc.Label
			}
			outCols[i].Type = sc.Type
		}
	}
	outRows := make([][]cellOut, len(rows))
	for i, r := range rows {
		outRows[i] = make([]cellOut, len(r))
		for j, v := range r {
			if outCols[j].Type == "money" && v != "" {
				if minor, err := strconv.ParseInt(v, 10, 64); err == nil {
					outRows[i][j] = cellOut{V: core.FormatMoney(minor)}
					continue
				}
			}
			outRows[i][j] = cellOut{V: v}
		}
	}
	writeJSON(w, 200, map[string]any{"view": vd, "columns": outCols, "rows": outRows})
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
	notices, err := s.Store.ProjectionNotices(r.Context())
	if err != nil {
		writeErr(w, 500, err)
		return
	}
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
	rules, err := s.Store.LatestRules(r.Context())
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	if rules == nil {
		rules = []core.Rule{}
	}
	writeJSON(w, 200, rules)
}

func (s *Server) handleDraft(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Intent        string `json:"intent"`
		SampleEventID int64  `json:"sample_event_id"`
		ObjectType    string `json:"object_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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
	objType, err := s.Store.GetObjectType(ctx, req.ObjectType)
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	draft, err := s.Agent.DraftRule(ctx, req.Intent, *sample, objType)
	if err != nil {
		writeErr(w, 502, err)
		return
	}
	rule, err := s.Store.InsertRuleVersion(ctx, core.Rule{
		ID: draft.RuleID, Status: core.StatusDraft, Priority: draft.Priority,
		EffectiveFrom: sample.OccurredAt, CreatedBy: "agent:" + agent.Model(),
		Description: draft.Description, Spec: draft.Spec,
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
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.ApprovedBy == "" {
		req.ApprovedBy = "human"
	}
	rule, booked, procErrs, err := s.Exec.ApproveRule(r.Context(), r.PathValue("id"),
		req.ApprovedBy, time.Now().Format("2006-01-02"))
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
// Read-only: the human sees the diff before deciding to approve.
func (s *Server) handleSimulate(w http.ResponseWriter, r *http.Request) {
	diff, err := s.Exec.SimulateRule(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, diff)
}

func (s *Server) handleSubmit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EventType  string          `json:"event_type"`
		OccurredAt string          `json:"occurred_at"`
		Payload    json.RawMessage `json:"payload"`
		DedupKey   string          `json:"dedup_key"`
		Actor      string          `json:"actor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if req.Actor == "" {
		req.Actor = "shell" // single-user experiment; identity arrives with auth, never
	}
	id, err := s.Store.AppendEvent(r.Context(), core.Event{
		Kind: core.KindRaw, Type: req.EventType, OccurredAt: req.OccurredAt,
		Payload: req.Payload, DedupKey: req.DedupKey, Actor: req.Actor,
	})
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	booked, errs := s.Exec.ProcessPending(r.Context())
	resp := map[string]any{"event_id": id, "booked": booked}
	if len(errs) > 0 {
		resp["errors"] = fmt.Sprintf("%v", errs)
	}
	writeJSON(w, 200, resp)
}
