package shell

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/network"
)

// bundleView is a bundle with its members' definitions inlined: the human
// approves what they can read, all of it on one screen.
type bundleView struct {
	core.Bundle
	Definitions []memberView `json:"definitions"`
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
			r, err := s.Store.GetRule(ctx, m.Name)
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
			a, err := s.Store.GetActivity(ctx, m.Name)
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
	return out, nil
}

func (s *Server) handleBundles(w http.ResponseWriter, r *http.Request) {
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
	json.NewDecoder(r.Body).Decode(&req)
	if req.ApprovedBy == "" {
		req.ApprovedBy = "human"
	}
	b, booked, procErrs, err := s.Exec.ApproveBundle(r.Context(), r.PathValue("id"),
		req.ApprovedBy, time.Now().Format("2006-01-02"))
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if req.RejectedBy == "" {
		req.RejectedBy = "human"
	}
	b, err := s.Exec.RejectBundle(r.Context(), r.PathValue("id"), req.Reason,
		req.RejectedBy, time.Now().Format("2006-01-02"))
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
func (s *Server) handleDraftBundle(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Intent        string `json:"intent"`
		SampleEventID int64  `json:"sample_event_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	ctx := r.Context()
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
		effective = time.Now().Format("2006-01-02")
	}
	b, err := StoreBundleDraft(ctx, s.Store, k, bd, "agent:"+agent.Model(), effective)
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
