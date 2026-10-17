package shell

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"rhea/internal/core"
)

// serveWords hands a face the market's words as the pack declared them
// (DIRECTION 2026-10-16, step 3: the pack supplies the words).
func (s *Server) serveWords(w http.ResponseWriter, vd *core.ViewDef) {
	var spec core.WordsSpec
	if err := json.Unmarshal(vd.Spec, &spec); err != nil {
		writeErr(w, 500, err)
		return
	}
	if err := spec.Check(); err != nil {
		writeErr(w, 500, fmt.Errorf("view %s: %w", vd.ID, err))
		return
	}
	writeJSON(w, 200, map[string]any{"view": vd, "words": spec})
}

// serveStatement lays one object's own figures out as the statement says:
// the rows in order, each with its operation and the figure as the kernel
// holds it. Without ?object_id= it answers with the layout alone.
func (s *Server) serveStatement(ctx context.Context, w http.ResponseWriter, vd *core.ViewDef, objectID string) {
	var spec core.StatementSpec
	if err := json.Unmarshal(vd.Spec, &spec); err != nil {
		writeErr(w, 500, err)
		return
	}
	t, err := s.Store.GetObjectType(ctx, spec.ObjectType)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	if err := spec.Check(t); err != nil {
		writeErr(w, 500, fmt.Errorf("view %s: %w", vd.ID, err))
		return
	}
	if objectID == "" {
		writeJSON(w, 200, map[string]any{"view": vd})
		return
	}
	o, err := s.Store.GetObject(ctx, objectID)
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	if o.Type != spec.ObjectType {
		writeErr(w, 400, fmt.Errorf("view %s lays out a %s, not a %s", vd.ID, spec.ObjectType, o.Type))
		return
	}
	type rowOut struct {
		Op    string `json:"op,omitempty"`
		Field string `json:"field"`
		Note  string `json:"note,omitempty"`
		V     string `json:"v"`
	}
	type sectionOut struct {
		Title string   `json:"title,omitempty"`
		Rows  []rowOut `json:"rows"`
	}
	out := make([]sectionOut, 0, len(spec.Sections))
	for _, sec := range spec.Sections {
		so := sectionOut{Title: sec.Title, Rows: []rowOut{}}
		for _, r := range sec.Rows {
			fd, _ := t.Field(r.Field)
			c := cellWith(o.State[r.Field], fd, nil, nil)
			if r.NonZero && (c.V == "" || c.V == "0.00") {
				continue
			}
			so.Rows = append(so.Rows, rowOut{Op: r.Op, Field: r.Field, Note: r.Note, V: c.V})
		}
		out = append(out, so)
	}
	writeJSON(w, 200, map[string]any{"view": vd, "object_id": o.ID, "sections": out})
}
