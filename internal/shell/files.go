package shell

import (
	"fmt"
	"net/http"

	"rhea/internal/adapter"
)

// Files out (DIRECTION 2026-10-15, "Month end"): any provided writer renders
// by name — GET /api/files/{writer}?k=v — and is handed over through the
// issue_document door — POST /api/files/{writer}/issue {params, actor,
// occurred_at}. The shell knows no writer; a market provides them, and
// /api/files lists what this binary can render.

func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	if !s.allowSystem(w, r, ViewFiles) {
		return
	}
	writeJSON(w, 200, map[string]any{"writers": adapter.WriterNames()})
}

func (s *Server) handleFile(w http.ResponseWriter, r *http.Request) {
	if !s.allowSystem(w, r, ViewFiles) {
		return
	}
	params := map[string]string{}
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	f, err := adapter.Render(r.Context(), s.Store, r.PathValue("writer"), params)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	w.Header().Set("Content-Type", f.MediaType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", f.Name))
	w.WriteHeader(200)
	w.Write(f.Bytes)
}

func (s *Server) handleIssueFile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Params     map[string]string `json:"params"`
		Actor      string            `json:"actor"`
		OccurredAt string            `json:"occurred_at"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if req.Actor == "" {
		req.Actor = "shell"
	}
	f, t, err := adapter.Issue(r.Context(), s.Store, s.Exec, r.PathValue("writer"), req.Params, req.Actor, req.OccurredAt)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"name": f.Name, "media_type": f.MediaType, "size": len(f.Bytes),
		"event_id": t.EventID, "booked": t.Booked, "errors": errStrings(t.Errors)})
}

func errStrings(errs []error) []string {
	out := make([]string, 0, len(errs))
	for _, e := range errs {
		out = append(out, e.Error())
	}
	return out
}
