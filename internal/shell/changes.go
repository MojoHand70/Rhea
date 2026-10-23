package shell

import (
	"fmt"
	"net/http"
	"time"

	"rhea/internal/core"
)

// handleChanges tells a company what Rhea changed since a day (?since=
// YYYY-MM-DD): the versions of Rhea's own rules that became active, with
// their words, their date and their source — a change of Rhea's reaches
// every company at once, announced to each (KK, 2026-10-22). A company's
// own rules are not in it.
func (s *Server) handleChanges(w http.ResponseWriter, r *http.Request) {
	since := r.URL.Query().Get("since")
	if _, err := time.Parse("2006-01-02", since); err != nil {
		writeErr(w, 400, fmt.Errorf("since: a day, YYYY-MM-DD"))
		return
	}
	changes, err := s.Store.RheasChanges(r.Context(), since)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	if changes == nil {
		changes = []core.Rule{}
	}
	writeJSON(w, 200, changes)
}
