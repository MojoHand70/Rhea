package shell

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"rhea/internal/adapter"
	"rhea/internal/core"
	"rhea/internal/store"
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
		out["error"] = fmt.Sprintf("the day's consequences failed and wait in the worklist: %s", strings.Join(errs, "; "))
	} else if res.Gated {
		out["error"] = fmt.Sprintf("%d days wait for a person's catch-up", len(res.Pending))
	}
	return out
}

// clockStatus is how a clock run answers: 409 when days wait for a
// person's catch-up, 500 when the day's consequences failed, 200 otherwise —
// both answered 200, and an overnight failure reached nobody (review
// 2026-10-21, G3). The body is the run's either way, and a failure is
// logged.
func clockStatus(res adapter.ClockResult) int {
	switch {
	case len(res.Errors) > 0:
		for _, e := range res.Errors {
			log.Printf("clock: %v", e)
		}
		return 500
	case res.Gated:
		log.Printf("clock: %d days wait for a person's catch-up", len(res.Pending))
		return 409
	}
	return 200
}

// handleHealth says whether the books' day is open — the machine's question
// before anyone opens Today (G3): the books' day, the last opened, and how
// many days are not; more than today's waits is 503.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	st, err := adapter.ClockState(store.WithOwner(r.Context()), s.Store, s.Clock)
	if err != nil {
		log.Printf("health: %v", err)
		writeJSON(w, 503, map[string]any{"ok": false, "error": "the books cannot be read"})
		return
	}
	ok := len(st.Pending) <= 1
	status := 200
	if !ok {
		status = 503
	}
	writeJSON(w, status, map[string]any{"ok": ok, "today": st.Today, "last_opened": st.LastOpened, "days_unopened": len(st.Pending)})
}

// handleTimeAdvance is the steady state over HTTP: opens at most one day;
// a burst comes back gated, untouched.
func (s *Server) handleTimeAdvance(w http.ResponseWriter, r *http.Request) {
	// Opening a day is the installation's act over every company: the
	// clock runs as the owner and fans out per company, deliberately.
	res, err := adapter.RunClock(store.WithOwner(r.Context()), s.Store, s.Exec, s.Clock)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, clockStatus(res), clockResult(res))
}

// handleTimeCatchUp is the human act: open the pending days, deliberately.
func (s *Server) handleTimeCatchUp(w http.ResponseWriter, r *http.Request) {
	if !s.allowEverywhere(w, r, "view", ViewGate, "system") {
		return
	}
	var req struct {
		Through string `json:"through"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, 400, err)
		return
	}
	res, err := adapter.CatchUpClock(store.WithOwner(r.Context()), s.Store, s.Exec, s.Clock, req.Through)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, clockStatus(res), clockResult(res))
}

// handleTimeSimulate is the gate's dry run, shaped by the same renderer as
// a rule approval's diff — the burst reads like any other decision.
func (s *Server) handleTimeSimulate(w http.ResponseWriter, r *http.Request) {
	if !s.allowEverywhere(w, r, "view", ViewGate, "system") {
		return
	}
	var req struct {
		Through string `json:"through"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, 400, err)
		return
	}
	ctx := store.WithOwner(r.Context())
	diff, days, err := adapter.SimulateCatchUp(ctx, s.Store, s.Exec, s.Clock, req.Through)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	out, err := s.renderSimDiff(ctx, diff, nil)
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
