package shell

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/network"
	"rhea/internal/store"
)

// bundleView is a bundle with its members' definitions inlined: the human
// approves what they can read, all of it on one screen.
type bundleView struct {
	core.Bundle
	Definitions []memberView `json:"definitions"`
	// Rejection is why a superseded bundle was refused, from the log — the
	// start of the conversation its redraft continues.
	Rejection *store.Rejection `json:"rejection,omitempty"`
}

type memberView struct {
	core.Member
	Summary    string `json:"summary"`
	Definition any    `json:"definition"`
	// Learned: how many installations explain this rule's question exactly
	// this way — counted now, from the network; nil when Rhea knows nothing
	// she may say.
	Learned *core.Support `json:"learned,omitempty"`
}

func (s *Server) bundleView(ctx context.Context, b core.Bundle) (bundleView, error) {
	return s.bundleViewWith(ctx, b, s.knowledge(ctx))
}

func (s *Server) bundleViewWith(ctx context.Context, b core.Bundle, k *network.Knowledge) (bundleView, error) {
	out := bundleView{Bundle: b}
	for _, m := range b.Members {
		mv := memberView{Member: m}
		switch m.Kind {
		case core.KindRule:
			r, err := s.Store.GetRuleVersion(ctx, m.Name, m.Version) // the member's, not the latest draft's
			if err != nil {
				return out, err
			}
			mv.Summary, mv.Definition = r.Description, r.Spec
			if k != nil {
				if sup, ok := k.SupportFor(r.Spec); ok {
					mv.Learned = &sup
				}
			}
		case core.KindActivity:
			a, err := s.Store.GetActivityVersion(ctx, m.Name, m.Version)
			if err != nil {
				return out, err
			}
			mv.Summary, mv.Definition = a.Description, a.Spec
		case core.KindObjectType:
			t, _, err := s.Store.GetObjectTypeVersion(ctx, m.Name, m.Version)
			if err != nil {
				return out, err
			}
			mv.Summary = fmt.Sprintf("%d fields, domain %s", len(t.Fields), t.Domain)
			mv.Definition = t
		case core.KindViewDef:
			v, _, err := s.Store.GetViewDefVersion(ctx, m.Name, m.Version)
			if err != nil {
				return out, err
			}
			mv.Summary = fmt.Sprintf("%s · %s", v.Notion, v.Title)
			mv.Definition = json.RawMessage(v.Spec)
		}
		out.Definitions = append(out.Definitions, mv)
	}
	if b.Status == core.StatusSuperseded {
		rj, ok, err := s.Store.BundleRejection(ctx, b.ID)
		if err != nil {
			return out, err
		}
		if ok {
			out.Rejection = &rj
		}
	}
	return out, nil
}

func (s *Server) handleBundles(w http.ResponseWriter, r *http.Request) {
	if !s.allowSystem(w, r, ViewGate) {
		return
	}
	bundles, err := s.Store.LatestBundles(r.Context())
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	out := []bundleView{}
	k := s.knowledge(r.Context()) // learned once per listing
	for _, b := range bundles {
		bv, err := s.bundleViewWith(r.Context(), b, k)
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		out = append(out, bv)
	}
	writeJSON(w, 200, out)
}

// handleSimulateBundle dry-runs the bundle as approval would land it.
func (s *Server) handleSimulateBundle(w http.ResponseWriter, r *http.Request) {
	if !s.allowSystem(w, r, ViewGate) { // as the preview asks (A9)
		return
	}
	ctx := r.Context()
	b, err := s.Store.GetBundle(ctx, r.PathValue("id"))
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	diff, err := s.Exec.SimulateBundle(ctx, b)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	drafts, err := s.Store.DraftObjectTypes(ctx, b)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	out, err := s.renderSimDiff(ctx, diff, drafts)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, out)
}

func (s *Server) handleApproveBundle(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ApprovedBy string `json:"approved_by"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if req.ApprovedBy == "" {
		req.ApprovedBy = "human"
	}
	b, booked, procErrs, err := s.Exec.ApproveBundle(r.Context(), r.PathValue("id"),
		req.ApprovedBy, s.Clock.TodayUTC())
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	resp := map[string]any{"bundle": b, "booked": booked}
	if len(procErrs) > 0 {
		resp["errors"] = fmt.Sprintf("%v", procErrs)
	}
	writeJSON(w, 200, resp)
}

func (s *Server) handleRejectBundle(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Reason     string `json:"reason"`
		RejectedBy string `json:"rejected_by"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if req.RejectedBy == "" {
		req.RejectedBy = "human"
	}
	b, err := s.Exec.RejectBundle(r.Context(), r.PathValue("id"), req.Reason,
		req.RejectedBy, s.Clock.TodayUTC())
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"bundle": b})
}

// handleDraftBundle is the draft_bundle door over HTTP: the ask is in the log
// as bundle.draft_requested before any model answers; the shell then runs the
// agent (its runner, until agent-as-user) and lands the answer as drafts.
// The sample event is optional — the agent always sees the whole residue.
// With `after`, the ask continues a conversation: the named bundle was
// rejected, and the agent sees it and the reason; an empty intent then
// means "the same question again".
func (s *Server) handleDraftBundle(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Intent        string `json:"intent"`
		SampleEventID int64  `json:"sample_event_id"`
		After         string `json:"after"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, 400, err)
		return
	}
	ctx := r.Context()
	var conversation []agent.Rejection
	if req.After != "" {
		var question string
		var err error
		conversation, question, err = Conversation(ctx, s.Store, req.After)
		if err != nil {
			writeErr(w, 400, err)
			return
		}
		if req.Intent == "" {
			req.Intent = question
		}
		if req.Intent == "" {
			writeErr(w, 400, fmt.Errorf("bundle %s has no recorded question: say what you want", req.After))
			return
		}
	}
	var sample core.Event
	if req.SampleEventID != 0 {
		ev, err := s.Store.GetEvent(ctx, req.SampleEventID)
		if err != nil {
			writeErr(w, 404, err)
			return
		}
		sample = ev
	}
	inputs := map[string]any{"intent": req.Intent}
	if req.SampleEventID != 0 {
		inputs["sample_event_id"] = req.SampleEventID
	}
	if req.After != "" {
		inputs["after"] = req.After
	}
	if _, err := s.Exec.TriggerActivity(ctx, "draft_bundle", inputs, "shell", ""); err != nil {
		writeErr(w, 400, err)
		return
	}
	k := s.knowledge(ctx)
	ask, err := DraftAsk(ctx, s.Store, k, req.Intent, sample, "")
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	ask.Rejections = conversation
	// A report's queries read the collection (DIRECTION 2026-10-17, "Reporting").
	if ask.Tables, err = s.collection(ctx); err != nil {
		writeErr(w, 500, err)
		return
	}
	ask.Words, ask.Today = s.words(ctx), s.Clock.TodayUTC()
	bd, err := s.Agent.DraftBundle(ctx, ask)
	if errors.Is(err, agent.ErrAlreadyAnswered) {
		writeJSON(w, 200, map[string]any{"already_answered": true, "description": bd.Description})
		return
	}
	if err != nil {
		writeErr(w, 502, err) // the request event stays — the log is honest about unanswered asks
		return
	}
	// Rules answering an interview explain the history too: without a
	// sample, they apply from the log's first fact (backfill then reaches
	// what the live path already passed).
	effective := sample.OccurredAt
	if effective == "" {
		if effective, err = s.Store.EarliestFactDate(ctx); err != nil {
			writeErr(w, 500, err)
			return
		}
	}
	if effective == "" {
		effective = s.Clock.TodayUTC()
	}
	b, err := StoreBundleDraft(ctx, s.Store, k, bd, "agent:"+agent.Model(), effective, req.Intent, req.After)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	bv, err := s.bundleView(ctx, b)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, bv)
}

// figurePreview is a waiting figure computed on the books as they stand.
type figurePreview struct {
	ViewID  string      `json:"view_id"`
	Title   string      `json:"title"`
	Columns []columnOut `json:"columns,omitempty"`
	Rows    [][]cellOut `json:"rows,omitempty"`
	Total   int         `json:"total"` // rows in all; Rows holds the first few
	Error   string      `json:"error,omitempty"`
}

// handlePreviewBundle shows what a waiting bundle's figures would show —
// its analysis views run now, on the books as they stand — so the person
// approves what they can read, not a description of it (DIRECTION
// 2026-10-17: a drafted figure grouped by an account number and labelled
// "customer" is plain once its rows are on the card).
func (s *Server) handlePreviewBundle(w http.ResponseWriter, r *http.Request) {
	if !s.allowSystem(w, r, ViewGate) {
		return
	}
	ctx := r.Context()
	b, err := s.Store.GetBundle(ctx, r.PathValue("id"))
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	out := []figurePreview{}
	for _, m := range b.Members {
		if m.Kind != core.KindViewDef {
			continue
		}
		v, _, err := s.Store.GetViewDefVersion(ctx, m.Name, m.Version)
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		if v.Notion == "report" {
			comps, err := s.reportComponents(ctx, &v)
			if err != nil {
				out = append(out, figurePreview{ViewID: v.ID, Title: v.Title, Error: err.Error()})
				continue
			}
			for _, c := range comps {
				out = append(out, componentPreview(v, c))
			}
			continue
		}
		if v.Notion != "analysis" {
			continue
		}
		p := figurePreview{ViewID: v.ID, Title: v.Title}
		spec, err := s.analysisSpec(ctx, v.Spec)
		if err == nil {
			p.Columns, p.Rows, _, err = s.analysisData(ctx, spec)
		}
		if err != nil {
			p.Error = err.Error()
		}
		p.Total = len(p.Rows)
		if len(p.Rows) > 8 {
			p.Rows = p.Rows[:8]
		}
		out = append(out, p)
	}
	writeJSON(w, 200, out)
}

// componentPreview is one component of a waiting report, as the card shows it:
// money marked as money, the first rows, the count.
func componentPreview(v core.ViewDef, c componentOut) figurePreview {
	p := figurePreview{ViewID: v.ID, Title: v.Title, Total: len(c.Rows), Error: c.Error}
	if c.Title != "" && c.Title != v.Title {
		p.Title = v.Title + " · " + c.Title
	}
	for _, col := range c.Columns {
		typ := ""
		if strings.HasPrefix(col.Type, "DECIMAL(") && strings.HasSuffix(col.Type, ",2)") {
			typ = "money"
		}
		p.Columns = append(p.Columns, columnOut{Field: col.Name, Label: col.Name, Type: typ})
	}
	for i, r := range c.Rows {
		if i == 8 {
			break
		}
		cells := make([]cellOut, len(r))
		for j, v := range r {
			cells[j] = cellOut{V: v}
		}
		p.Rows = append(p.Rows, cells)
	}
	return p
}
