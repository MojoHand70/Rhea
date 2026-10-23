package shell

import (
	"net/http"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/rights"
)

// handlePromotions lists the readings companies hold that have reached the
// floor for promotion (KK, 2026-10-22): Rhea's staff's to see — a right in
// the house to promote — because it is counted across every company. It
// names no company, and a reading below the floor is never listed.
func (s *Server) handlePromotions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if p := rights.PrincipalOf(ctx); p != nil {
		if _, ok := p.May(core.House, "door", "promote_rule", "system"); !ok {
			writeErr(w, 403, &rights.ErrRefused{Person: p.Handle, Door: "promote_rule", Company: core.House,
				Reason: "what companies hold is counted for Rhea's staff alone"})
			return
		}
	}
	cands, err := s.Exec.Promotable(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	if cands == nil {
		cands = []exec.Candidate{}
	}
	writeJSON(w, 200, cands)
}
