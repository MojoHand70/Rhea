package shell

import (
	"fmt"
	"net/http"

	"rhea/internal/exec"
)

// handleSimulateBackfill is the dry run of ruled backfill: which past events
// the active rules would explain further, offered forward where a period is
// closed, rendered as the same field-level diff every approval reads.
func (s *Server) handleSimulateBackfill(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	plan, err := s.Exec.PlanBackfill(ctx, s.Clock.TodayUTC())
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	diff, err := s.renderSimDiff(ctx, plan.Diff, nil)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	chains, errs := plan.Chains, plan.Errors
	if chains == nil {
		chains = []exec.BackfillChain{} // never null: the client counts them
	}
	if errs == nil {
		errs = []string{}
	}
	writeJSON(w, 200, map[string]any{"date": plan.Date, "chains": chains, "errors": errs, "diff": diff})
}

// handleApproveBackfill is the deliberate act: approve_backfill, then the
// chains book.
func (s *Server) handleApproveBackfill(w http.ResponseWriter, r *http.Request) {
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
	n, errs, err := s.Exec.ApproveBackfill(r.Context(), req.ApprovedBy, s.Clock.TodayUTC())
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
