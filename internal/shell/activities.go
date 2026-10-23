package shell

import (
	"context"
	"fmt"
	"net/http"

	"rhea/internal/core"
	"rhea/internal/rights"
)

// The second vocabulary over HTTP (DIRECTION: the replaceable interpreter —
// "a shell is anything that can render the notions and offer the
// activities"). What can be done crosses as data the same way what can be
// seen does: declarations with typed inputs; the renderer decides the form.

type activityOut struct {
	core.Activity
	// Offered: this row is the active version a trigger goes through. A
	// draft awaiting approval is listed alongside, unoffered — the gate
	// surface and the verb list are the same list.
	Offered bool `json:"offered"`
	// EventsEmitted counts the events that entered through this door.
	EventsEmitted int `json:"events_emitted"`
	// Allowed says whether the signed-in person may open this door where
	// they act — what a face renders as offered; nil without a person.
	Allowed *bool `json:"allowed,omitempty"`
}

// handleActivities lists the offered verb set plus whatever waits at the
// gate: active versions first, then latest versions that are not the active
// one (drafts in progress, retired verbs).
func (s *Server) handleActivities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	active, err := s.Store.ActiveActivities(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	latest, err := s.Store.LatestActivities(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	counts, err := s.Store.ActivityEventCounts(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	activeVersion := map[string]int{}
	out := []activityOut{}
	p := rights.PrincipalOf(ctx)
	for _, a := range active {
		activeVersion[a.Name] = a.Version
		row := activityOut{Activity: a, Offered: true, EventsEmitted: counts[a.Name]}
		if p != nil {
			ok := a.Spec.Admits(core.ActorHuman)
			if ok && a.Spec.Installation {
				_, ok = p.MayAnywhere("door", a.Name, a.Domain)
			} else if ok {
				ok = s.may(ctx, p, "door", a.Name, a.Domain)
			}
			row.Allowed = &ok
		}
		out = append(out, row)
	}
	for _, a := range latest {
		if a.Version == activeVersion[a.Name] {
			continue
		}
		out = append(out, activityOut{Activity: a, EventsEmitted: counts[a.Name]})
	}
	writeJSON(w, 200, out)
}

// handleTrigger fires a declared verb: POST /api/activities/{name}/trigger
// with {"inputs": {...}}. The kernel's door does all validation; this
// handler only carries the envelope.
func (s *Server) handleTrigger(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Inputs   map[string]any `json:"inputs"`
		DedupKey string         `json:"dedup_key"`
		Actor    string         `json:"actor"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if req.Actor == "" {
		req.Actor = "shell"
	}
	if p := rights.PrincipalOf(r.Context()); p != nil {
		req.Actor = p.Handle // the server says who, whatever the page sent
	}
	t, err := s.Exec.TriggerActivity(r.Context(), r.PathValue("name"), req.Inputs, req.Actor, req.DedupKey)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	resp := map[string]any{"event_id": t.EventID, "booked": t.Booked}
	if t.Proposed != nil {
		resp["proposed"] = t.Proposed // the host's draft, when a pairing taught it something
	}
	if len(t.Errors) > 0 {
		resp["errors"] = fmt.Sprintf("%v", t.Errors)
	}
	writeJSON(w, 200, resp)
}

// handleApproveActivity is the gate applied to the gate: approving a draft
// verb goes through the approve_activity door like rule approval goes
// through approve_rule.
func (s *Server) handleApproveActivity(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ApprovedBy string `json:"approved_by"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if req.ApprovedBy == "" {
		req.ApprovedBy = "human"
	}
	a, booked, procErrs, err := s.Exec.ApproveActivity(r.Context(), r.PathValue("name"),
		req.ApprovedBy, s.Clock.Today())
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	resp := map[string]any{"activity": a, "booked": booked}
	if len(procErrs) > 0 {
		resp["errors"] = fmt.Sprintf("%v", procErrs)
	}
	writeJSON(w, 200, resp)
}

// offeredVerb is a contextual verb on an object: an active activity that
// declares a ref input of the object's type. This is how "the one or two
// verbs that resolve it, inline" renders generically — the surfacing
// convention cases will lean on.
type offeredVerb struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	RefInput    string `json:"ref_input"`
}

// verbsForType returns the active verbs that take a ref to the given type.
func (s *Server) verbsForType(ctx context.Context, typeName string) []offeredVerb {
	acts, err := s.Store.ActiveActivities(ctx)
	if err != nil {
		return nil
	}
	verbs := []offeredVerb{}
	for _, a := range acts {
		for _, in := range a.Spec.Inputs {
			if target, ok := core.RefTarget(in.Type); ok && target == typeName {
				verbs = append(verbs, offeredVerb{Name: a.Name, Description: a.Description, RefInput: in.Name})
				break
			}
		}
	}
	return verbs
}
