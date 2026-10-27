package shell

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"rhea/internal/core"
	"rhea/internal/exec"
)

// serveWords hands a face the company's words (DIRECTION 2026-10-16, step 3:
// the pack supplies the words): every words screen it is served, merged in
// the order of their packs (DIRECTION 2026-10-24, step 3) — whichever of them
// is read answers with the whole.
func (s *Server) serveWords(ctx context.Context, w http.ResponseWriter, vd *core.ViewDef) {
	spec, err := s.mergedWords(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	if spec == nil {
		writeErr(w, 500, fmt.Errorf("view %s: no words", vd.ID))
		return
	}
	writeJSON(w, 200, map[string]any{"view": vd, "words": spec})
}

// mergedWords is the company's words: the words screens it is served, each
// checked, laid over each other in the order of their packs; nil with none.
func (s *Server) mergedWords(ctx context.Context) (*core.WordsSpec, error) {
	vds, err := s.effectiveViewDefs(ctx)
	if err != nil {
		return nil, err
	}
	var words []core.ViewDef
	for _, v := range vds {
		if v.Notion == "words" {
			words = append(words, v)
		}
	}
	if len(words) == 0 {
		return nil, nil
	}
	if len(words) > 1 {
		rank, err := exec.ViewRanks(ctx, s.Store)
		if err != nil {
			return nil, err
		}
		slices.SortStableFunc(words, func(a, b core.ViewDef) int {
			if d := rank(a.ID) - rank(b.ID); d != 0 {
				return d
			}
			return strings.Compare(a.ID, b.ID)
		})
	}
	var out core.WordsSpec
	for _, v := range words {
		var spec core.WordsSpec
		if err := json.Unmarshal(v.Spec, &spec); err != nil {
			return nil, fmt.Errorf("view %s: %w", v.ID, err)
		}
		if err := spec.Check(); err != nil {
			return nil, fmt.Errorf("view %s: %w", v.ID, err)
		}
		out.Merge(spec)
	}
	return &out, nil
}

// marketHeaders names, in the company's words, the columns of a view whose
// packs bring no words of their own (KK's screenshot, 2026-10-27: Wszystkie
// firmy headed NAME, KIND, COUNTRY). The core and the common packs label
// their lists and reports in English for any market; a market's words name
// them where they name the column — a list's as its kind's field (kind.field,
// else the field), a report's as the view's (view.column, else the column).
// A pack with words of its own labels its lists in its language on purpose,
// often shorter than its words ("Nr inw." for "Numer"), and keeps them; so
// does a view no pack declares (a company's saved report).
func (s *Server) marketHeaders(ctx context.Context, vd *core.ViewDef, kind string, cols []columnOut) error {
	words, err := s.mergedWords(ctx)
	if err != nil || words == nil {
		return err
	}
	decl, err := s.Store.PackDeclarations(ctx)
	if err != nil {
		return err
	}
	vds, err := s.effectiveViewDefs(ctx)
	if err != nil {
		return err
	}
	wordsViews := map[string]bool{}
	for _, v := range vds {
		if v.Notion == "words" {
			wordsViews[v.ID] = true
		}
	}
	declared := false
	for _, ms := range decl {
		if !slices.ContainsFunc(ms, func(m core.Member) bool { return m.Kind == core.KindViewDef && m.Name == vd.ID }) {
			continue
		}
		declared = true
		if slices.ContainsFunc(ms, func(m core.Member) bool { return m.Kind == core.KindViewDef && wordsViews[m.Name] }) {
			return nil // its pack speaks for itself
		}
	}
	if !declared {
		return nil
	}
	of := cmp.Or(kind, vd.ID)
	for i, c := range cols {
		if name := cmp.Or(words.Fields[of+"."+c.Field], words.Fields[c.Field]); name != "" {
			cols[i].Label = name
		}
	}
	return nil
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
