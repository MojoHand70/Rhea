package shell

import (
	"net/http"

	"rhea/internal/adapter"
)

// The channel registry over HTTP (DIRECTION 2026-10-10, "Documents in"):
// two lists, what this company has connected and what Rhea can read. Both
// are generic rows the shell renders; neither is a screen for a particular
// channel.
func (s *Server) handleChannels(w http.ResponseWriter, r *http.Request) {
	reg, err := adapter.Registry(r.Context(), s.Store)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, reg)
}
