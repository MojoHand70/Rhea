package shell

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"

	"rhea/internal/adapter"
	"rhea/internal/core"
)

// The Time surface: where the system stands against the calendar, and the
// burst gate (DIRECTION 2026-10-05 — steady state is automatic, bursts
// need a human). One pending day opens on a click; a gap shows its days,
// dry-runs through the same simulation renderer the rule gate uses, and
// opens only on an explicit catch-up.

func (s *Server) handleTime(w http.ResponseWriter, r *http.Request) {
	st, err := adapter.ClockState(r.Context(), s.Store, s.Clock)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, st)
}

func clockResult(res adapter.ClockResult) map[string]any {
	out := map[string]any{"opened": res.Opened, "fired": res.Fired,
		"booked": res.Booked, "gated": res.Gated, "pending": res.Pending}
	if len(res.Errors) > 0 {
		errs := make([]string, len(res.Errors))
		for i, e := range res.Errors {
			errs[i] = e.Error()
		}
		out["errors"] = errs
	}
	return out
}

// handleTimeAdvance is the steady state over HTTP: opens at most one day;
// a burst comes back gated, untouched.
func (s *Server) handleTimeAdvance(w http.ResponseWriter, r *http.Request) {
	res, err := adapter.RunClock(r.Context(), s.Store, s.Exec, s.Clock)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, clockResult(res))
}

// handleTimeCatchUp is the human act: open the pending days, deliberately.
func (s *Server) handleTimeCatchUp(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Through string `json:"through"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	res, err := adapter.CatchUpClock(r.Context(), s.Store, s.Exec, s.Clock, req.Through)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, clockResult(res))
}

// handleTimeSimulate is the gate's dry run, shaped by the same renderer as
// a rule approval's diff — the burst reads like any other decision.
func (s *Server) handleTimeSimulate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Through string `json:"through"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	ctx := r.Context()
	diff, days, err := adapter.SimulateCatchUp(ctx, s.Store, s.Exec, s.Clock, req.Through)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	out, err := s.renderSimDiff(ctx, diff)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"days": days, "diff": out})
}

// serveScheduling renders the last notion of the founding table (SPEC §2):
// time-axis placement of objects. The spec names a type and a date field;
// label and status derive from the ObjectType, like derived views.
func (s *Server) serveScheduling(ctx context.Context, w http.ResponseWriter, vd *core.ViewDef) {
	var spec core.SchedulingSpec
	if err := json.Unmarshal(vd.Spec, &spec); err != nil {
		writeErr(w, 500, err)
		return
	}
	objType, err := s.Store.GetObjectType(ctx, spec.ObjectType)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	objs, err := s.Store.ObjectsByType(ctx, spec.ObjectType)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	vds, err := s.effectiveViewDefs(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	detail := detailViewByType(vds)[spec.ObjectType]
	statusField := ""
	if objType.Lifecycle != nil {
		statusField = objType.Lifecycle.Field
	}

	type itemOut struct {
		ObjectID string `json:"object_id"`
		Label    string `json:"label"`
		Status   string `json:"status,omitempty"`
		Detail   string `json:"detail,omitempty"`
	}
	byDate := map[string][]itemOut{}
	for _, o := range objs {
		date, _ := o.State[spec.DateField].(string)
		item := itemOut{ObjectID: o.ID, Label: o.ID, Detail: detail}
		if objType.LabelField != "" {
			if l, ok := o.State[objType.LabelField].(string); ok && l != "" {
				item.Label = l
			}
		}
		if statusField != "" {
			item.Status, _ = o.State[statusField].(string)
		}
		byDate[date] = append(byDate[date], item)
	}
	dates := make([]string, 0, len(byDate))
	for d := range byDate {
		if d != "" {
			dates = append(dates, d)
		}
	}
	sort.Strings(dates)
	type dayOut struct {
		Date  string    `json:"date"`
		Items []itemOut `json:"items"`
	}
	days := make([]dayOut, 0, len(dates))
	for _, d := range dates {
		days = append(days, dayOut{Date: d, Items: byDate[d]})
	}
	// Today is presentation: the renderer marks past and present. The
	// kernel's time is the log's; an interpreter may know what day it is.
	writeJSON(w, 200, map[string]any{
		"view": vd, "date_field": spec.DateField,
		"today": s.Clock.TodayUTC(),
		"days":  days, "undated": byDate[""],
	})
}
