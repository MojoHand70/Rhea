package shell

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// handleSimulateBackfill is the dry run of ruled backfill: which past events
// the active rules would explain further, offered forward where a period is
// closed, rendered as the same field-level diff every approval reads.
func (s *Server) handleSimulateBackfill(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	plan, err := s.Exec.PlanBackfill(ctx, time.Now().Format("2006-01-02"))
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	diff, err := s.renderSimDiff(ctx, plan.Diff, nil)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"date": plan.Date, "chains": plan.Chains, "errors": plan.Errors, "diff": diff})
}

// handleApproveBackfill is the deliberate act: approve_backfill, then the
// chains book.
func (s *Server) handleApproveBackfill(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ApprovedBy string `json:"approved_by"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.ApprovedBy == "" {
		req.ApprovedBy = "human"
	}
	n, errs, err := s.Exec.ApproveBackfill(r.Context(), req.ApprovedBy, time.Now().Format("2006-01-02"))
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	resp := map[string]any{"backfilled": n}
	if len(errs) > 0 {
		resp["errors"] = fmt.Sprintf("%v", errs)
	}
	writeJSON(w, 200, resp)
}
