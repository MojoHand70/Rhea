// Package shell is the HTTP face of the kernel: a JSON API plus the embedded
// web shell (activity bar → submenu → tabs). It renders view notions
// generically; it knows nothing about invoices (SPEC §5).
package shell

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
	mux.HandleFunc("GET /api/rules", s.handleRules)
	mux.HandleFunc("POST /api/rules/draft", s.handleDraft)
	mux.HandleFunc("POST /api/rules/{id}/approve", s.handleApprove)
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

// handleNav builds the activity bar and submenus from the stored view
// definitions, then appends the system functions every domain carries.
func (s *Server) handleNav(w http.ResponseWriter, r *http.Request) {
	vds, err := s.Store.LatestViewDefs(r.Context())
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
			navFunction{Function: "worklist"}, navFunction{Function: "rules"})
	}
	writeJSON(w, 200, domains)
}

// --- views ----------------------------------------------------------------

// handleView serves a view definition together with its data, shaped per
// notion so the client stays a dumb renderer.
func (s *Server) handleView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vds, err := s.Store.LatestViewDefs(ctx)
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

// displayValue formats one state value; a ref field renders as the referenced
// object's label_field so humans see "ACME Sp. z o.o.", not "company-7". A
// ref that does not resolve (old type versions hold plain strings) falls back
// to the raw value.
func (s *Server) displayValue(ctx context.Context, v any, fd core.FieldDef) string {
	if _, isRef := core.RefTarget(fd.Type); isRef {
		if id, ok := v.(string); ok && id != "" {
			if label := s.refLabel(ctx, id); label != "" {
				return label
			}
		}
	}
	return display(v, fd.Type)
}

func (s *Server) refLabel(ctx context.Context, objectID string) string {
	o, err := s.Store.GetObject(ctx, objectID)
	if err != nil {
		return ""
	}
	ot, err := s.Store.GetObjectType(ctx, o.Type)
	if err != nil || ot.LabelField == "" {
		return ""
	}
	label, _ := o.State[ot.LabelField].(string)
	return label
}

// display formats a state value for humans, by field type. Money arrives from
// JSON as float64 minor units (exact for int64 magnitudes that fit 2^53).
func display(v any, fieldType string) string {
	if v == nil {
		return ""
	}
	if fieldType == "money" {
		switch n := v.(type) {
		case float64:
			return core.FormatMoney(int64(n))
		case int64:
			return core.FormatMoney(n)
		}
	}
	return fmt.Sprintf("%v", v)
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
	cols := make([]string, len(spec.Columns))
	for i, c := range spec.Columns {
		cols[i] = c.Label
	}
	rows := make([][]string, 0, len(objs))
	ids := make([]string, 0, len(objs))
	for _, o := range objs {
		row := make([]string, len(spec.Columns))
		for i, c := range spec.Columns {
			fd, _ := objType.Field(c.Field)
			row[i] = s.displayValue(ctx, o.State[c.Field], fd)
		}
		rows = append(rows, row)
		ids = append(ids, o.ID)
	}
	// Row click opens the detail view for the same object type, if one exists.
	detailViewID := ""
	if vds, err := s.Store.LatestViewDefs(ctx); err == nil {
		for _, v := range vds {
			if v.Notion != "detail" {
				continue
			}
			var ds core.DetailSpec
			if json.Unmarshal(v.Spec, &ds) == nil && ds.ObjectType == spec.ObjectType {
				detailViewID = v.ID
				break
			}
		}
	}
	writeJSON(w, 200, map[string]any{"view": vd, "columns": cols, "rows": rows,
		"object_ids": ids, "detail_view_id": detailViewID})
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
	type fieldOut struct{ Label, Value string }
	type sectionOut struct {
		Title  string
		Fields []fieldOut
	}
	sections := make([]sectionOut, 0, len(spec.Sections))
	for _, sec := range spec.Sections {
		so := sectionOut{Title: sec.Title}
		for _, f := range sec.Fields {
			fd, _ := objType.Field(f)
			so.Fields = append(so.Fields, fieldOut{Label: f, Value: s.displayValue(ctx, o.State[f], fd)})
		}
		sections = append(sections, so)
	}
	writeJSON(w, 200, map[string]any{
		"view": vd, "object": o, "sections": sections,
		"provenance": fmt.Sprintf("event %d via rule %s v%d", o.SourceEventID, o.RuleID, o.RuleVersion),
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
	writeJSON(w, 200, map[string]any{"view": vd, "columns": cols, "rows": rows})
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
	rule, booked, err := s.Exec.ApproveRule(r.Context(), r.PathValue("id"),
		req.ApprovedBy, time.Now().Format("2006-01-02"))
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"rule": rule, "booked": booked})
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
