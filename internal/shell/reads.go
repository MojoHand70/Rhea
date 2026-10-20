package shell

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"rhea/internal/core"
	"rhea/internal/store"
)

// The reads a face needs that the views do not carry (DIRECTION 2026-10-16,
// the face's first build): the objects themselves with their provenance,
// who approved each active rule, what the host said, and what a view's cell
// is made of. A face built apart reads only this API; nothing here writes.

// objectOut is an object with the raw event its chain starts from: the fact
// it ultimately explains, shared by everything that fact became.
type objectOut struct {
	core.Object
	RootEventID int64 `json:"root_event_id"`
}

// handleObjects lists one type's objects, each with its state as the
// kernel holds it (money in minor units, refs as ids) and its provenance:
// the event, rule and rule version that made it, and the chain's root.
func (s *Server) handleObjects(w http.ResponseWriter, r *http.Request) {
	typ := r.URL.Query().Get("type")
	if typ == "" {
		writeErr(w, 400, fmt.Errorf("objects wants ?type= — /api/types lists them"))
		return
	}
	if !s.allowType(w, r, typ) {
		return
	}
	objs, err := s.Store.ObjectsByType(r.Context(), typ)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	roots, err := s.Store.ObjectRoots(r.Context(), typ)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	out := make([]objectOut, len(objs))
	for i, o := range objs {
		out[i] = objectOut{Object: o, RootEventID: roots[o.ID]}
	}
	writeJSON(w, 200, map[string]any{"type": typ, "objects": out})
}

func (s *Server) handleObject(w http.ResponseWriter, r *http.Request) {
	o, err := s.Store.GetObject(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	if !s.allowType(w, r, o.Type) {
		return
	}
	roots, err := s.Store.ObjectRoots(r.Context(), o.Type)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, objectOut{Object: o, RootEventID: roots[o.ID]})
}

// handleSaid is what the host said in a narrative view, newest first: the
// words as the log keeps them, so a face shows this morning's briefing
// without asking the host to speak again (reading the view speaks anew).
// With person=<handle> the view is that person's: the narrative for their
// role (narrativeFor).
func (s *Server) handleSaid(w http.ResponseWriter, r *http.Request) {
	if !s.allowSystem(w, r, ViewSaid) {
		return
	}
	view := r.URL.Query().Get("view")
	if h := r.URL.Query().Get("person"); h != "" {
		if view != "" {
			writeErr(w, 400, fmt.Errorf("said: a view or a person, not both"))
			return
		}
		v, err := s.narrativeFor(r.Context(), h)
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		if v == "" {
			writeJSON(w, 200, []any{})
			return
		}
		view = v
	}
	n := 10
	if v := r.URL.Query().Get("n"); v != "" {
		k, err := strconv.Atoi(v)
		if err != nil || k < 1 {
			writeErr(w, 400, fmt.Errorf("said: n must be a positive number"))
			return
		}
		n = k
	}
	evs, err := s.Store.EventsOfTypes(r.Context(), []string{core.EventHostSaid})
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	type said struct {
		EventID    int64  `json:"event_id"`
		View       string `json:"view"`
		Text       string `json:"text"`
		OccurredAt string `json:"occurred_at"`
		RecordedAt string `json:"recorded_at"`
		Actor      string `json:"actor"`
	}
	out := []said{}
	for i := len(evs) - 1; i >= 0 && len(out) < n; i-- {
		var p struct {
			View string `json:"view"`
			Text string `json:"text"`
		}
		if json.Unmarshal(evs[i].Payload, &p) != nil || (view != "" && p.View != view) {
			continue
		}
		out = append(out, said{EventID: evs[i].ID, View: p.View, Text: p.Text, OccurredAt: evs[i].OccurredAt,
			RecordedAt: evs[i].RecordedAt.Format("2006-01-02T15:04:05Z07:00"), Actor: evs[i].Actor})
	}
	writeJSON(w, 200, out)
}

// approvalOut is the approval that made a rule version active.
type approvalOut struct {
	EventID int64  `json:"event_id"`
	By      string `json:"approved_by"`
	At      string `json:"approved_at"`      // the business date the approval carries
	Bundle  string `json:"bundle,omitempty"` // when it was approved as a bundle member
	Reason  string `json:"reason,omitempty"` // a correction's reason
}

type ruleOut struct {
	core.Rule
	Approval *approvalOut `json:"approval,omitempty"`
}

func approvalKey(ruleID string, version int) string { return fmt.Sprintf("%s/%d", ruleID, version) }

// ruleApprovals maps each approved rule version to the event that approved
// it: its own door (rule.approved names the version it made) or a bundle's
// (bundle.approved names the drafts it activated; each became the next
// version). Definitions are installation-wide, so the whole log is read.
func (s *Server) ruleApprovals(ctx context.Context) (map[string]approvalOut, error) {
	evs, err := s.Store.EventsOfTypes(store.WithCompany(ctx, ""), []string{"rule.approved", "bundle.approved"})
	if err != nil {
		return nil, err
	}
	out := map[string]approvalOut{}
	for _, ev := range evs {
		var p struct {
			RuleID          string        `json:"rule_id"`
			ApprovedVersion int           `json:"approved_version"`
			ApprovedBy      string        `json:"approved_by"`
			Reason          string        `json:"reason"`
			BundleID        string        `json:"bundle_id"`
			Members         []core.Member `json:"members"`
		}
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return nil, err
		}
		a := approvalOut{EventID: ev.ID, By: p.ApprovedBy, At: ev.OccurredAt, Reason: p.Reason}
		if ev.Type == "rule.approved" {
			out[approvalKey(p.RuleID, p.ApprovedVersion)] = a
			continue
		}
		a.Bundle = p.BundleID
		for _, m := range p.Members {
			if m.Kind == core.KindRule {
				out[approvalKey(m.Name, m.Version+1)] = a
			}
		}
	}
	return out, nil
}

// explainCell answers "why?" on one cell of an analysis view: the objects
// its row is made of (the view's row-lineage column), each object's part
// when the column declares what it sums, and whether the parts make the
// cell. A view that does not say what its rows are made of is refused —
// the face must not guess.
func (s *Server) explainCell(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	vds, err := s.effectiveViewDefs(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	var vd *core.ViewDef
	for i := range vds {
		if vds[i].ID == q.Get("view") {
			vd = &vds[i]
		}
	}
	if vd == nil || vd.Notion != "analysis" {
		writeErr(w, 404, fmt.Errorf("explain: no analysis view %q", q.Get("view")))
		return
	}
	if !s.allowView(w, r, vd) {
		return
	}
	spec, err := s.analysisSpec(ctx, vd.Spec)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	cols, rows, rowsOf, err := s.analysisData(ctx, spec)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	if rowsOf == nil {
		writeErr(w, 400, fmt.Errorf("explain: view %s does not say what its rows are made of (no %s column)", vd.ID, core.RowsOfColumn))
		return
	}
	row, err := strconv.Atoi(q.Get("row"))
	if err != nil || row < 0 || row >= len(rows) {
		writeErr(w, 400, fmt.Errorf("explain: view %s has %d rows; row %q is not one", vd.ID, len(rows), q.Get("row")))
		return
	}
	col := -1
	for i, c := range cols {
		if c.Field == q.Get("col") {
			col = i
		}
	}
	if col < 0 {
		writeErr(w, 400, fmt.Errorf("explain: view %s has no column %q", vd.ID, q.Get("col")))
		return
	}
	var sumOf []string
	for i, sc := range spec.Columns {
		if i == col {
			sumOf = sc.SumOf
		}
	}
	type part struct {
		ObjectID   string `json:"object_id"`
		ObjectType string `json:"object_type"`
		Label      string `json:"label"`
		Field      string `json:"field,omitempty"` // which of SumOf the object carried
		V          string `json:"v,omitempty"`
	}
	resp := map[string]any{"view": vd.ID, "row": rows[row], "column": cols[col], "v": rows[row][col].V}
	parts := []part{}
	var total int64
	summed := len(sumOf) > 0 && cols[col].Type == "money"
	objs := make([]core.Object, 0, len(rowsOf[row]))
	for _, id := range rowsOf[row] {
		o, err := s.Store.GetObject(ctx, id)
		if err != nil {
			writeErr(w, 500, fmt.Errorf("explain: row %d of %s names %s: %w", row, vd.ID, id, err))
			return
		}
		objs = append(objs, o)
	}
	// The parts in the order the log made them, whatever order the SQL kept.
	sort.SliceStable(objs, func(i, j int) bool { return objs[i].SourceEventID < objs[j].SourceEventID })
	for _, o := range objs {
		p := part{ObjectID: o.ID, ObjectType: o.Type}
		p.Label, _ = s.refInfo(ctx, o.ID)
		if summed {
			for _, f := range sumOf {
				if v, ok := o.State[f]; ok && v != nil {
					m, ok := moneyMinor(v)
					if !ok {
						writeErr(w, 500, fmt.Errorf("explain: %s.%s is not money: %v", o.ID, f, v))
						return
					}
					p.Field, p.V = f, core.FormatMoney(m)
					total += m
					break
				}
			}
		}
		parts = append(parts, p)
	}
	resp["objects"] = parts
	if summed {
		cell, err := core.ParseMoney(rows[row][col].V)
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		resp["total"], resp["sums"] = core.FormatMoney(total), total == cell
	}
	writeJSON(w, 200, resp)
}

// moneyMinor reads a money field as the store hands it back from JSON:
// minor units as a JSON number (exact for int64 magnitudes that fit 2^53,
// as typedCell reads them); anything with a fraction is not money.
func moneyMinor(v any) (int64, bool) {
	switch n := v.(type) {
	case float64:
		if n != float64(int64(n)) {
			return 0, false
		}
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	case json.Number:
		k, err := n.Int64()
		return k, err == nil
	}
	return 0, false
}
