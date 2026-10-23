package shell

import (
	"fmt"
	"net/http"
	"strings"

	"rhea/internal/adapter"
)

// The company registers a market reads by tax id (adapter.CompanyRegister):
// GET /api/registers names them; GET /api/registers/{name}?id=<tax id> says
// what one of them holds of a company — what a person adding a company sees
// before confirming it (DIRECTION 2026-10-19, "People and rights":
// onboarding). Public records, read for whoever asks: nothing comes from any
// company's books and nothing is written; what the person confirms reaches
// the log through a door, this answer named in it as the source.
func (s *Server) handleRegisters(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"registers": adapter.CompanyRegisterNames()})
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	name, id := r.PathValue("name"), r.URL.Query().Get("id")
	if strings.TrimSpace(id) == "" {
		writeErr(w, 400, fmt.Errorf("name the tax id (?id=)"))
		return
	}
	known := false
	for _, n := range adapter.CompanyRegisterNames() {
		known = known || n == name
	}
	if !known {
		writeErr(w, 404, fmt.Errorf("no register named %s", name))
		return
	}
	e, err := adapter.LookUpCompany(r.Context(), name, id)
	if err != nil {
		// the register did not answer, or refused the id: not the server's
		// failure, so the person reads it in the register's own words
		writeErr(w, http.StatusFailedDependency, err)
		return
	}
	writeJSON(w, 200, e)
}
