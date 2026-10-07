package shell

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"rhea/internal/core"
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
}

func (s *Server) bundleView(ctx context.Context, b core.Bundle) (bundleView, error) {
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
	for _, b := range bundles {
		bv, err := s.bundleView(r.Context(), b)
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
