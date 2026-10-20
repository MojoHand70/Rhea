package shell

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"rhea/internal/core"
)

// The parties a business deals with, whose names never leave it: the
// companies (the partition key's own type) and the people its actors are.
var namedTypes = []string{"company", "person"}

// noted is what became of an untaught moment: recorded always, shared with
// the house when it may be, and why not when it may not.
type noted struct {
	EventID  int64  `json:"event_id"`
	Kind     string `json:"kind"`
	Shared   bool   `json:"shared"`
	KeptHome string `json:"kept_home_because,omitempty"`
}

// noteUntaught records what the host was asked and could not fully answer
// (DIRECTION 2026-10-17): never taught it, the books do not hold it, or her
// query failed. The person's words stay here; her general description
// reaches the network when it carries no number, no address and no name the
// installation holds, and when a network is connected — and the fact says
// which, so the owner sees everything she reported. A failure's detail is
// only the class of the refusal.
func (s *Server) noteUntaught(ctx context.Context, asked, missing, askedBy, kind, detail string) (noted, error) {
	asked, missing = strings.TrimSpace(asked), strings.Join(strings.Fields(missing), " ")
	if kind == "" {
		kind = "untaught"
	}
	if asked == "" || len(asked) > 2000 || missing == "" || len(missing) > 400 || askedBy == "" {
		return noted{}, fmt.Errorf("untaught: asked (1–2000 bytes), missing (1–400) and asked_by are required")
	}
	var names []string
	for _, t := range namedTypes {
		objs, err := s.Store.ObjectsByType(ctx, t) // the caller's scope, never the owner's
		if err != nil {
			return noted{}, err
		}
		for _, o := range objs {
			for _, f := range []string{"name", "handle"} {
				if v, ok := o.State[f].(string); ok && v != "" {
					names = append(names, v)
				}
			}
		}
	}
	why := core.KeptHome(missing, names)
	if why == "" && s.Network == nil {
		why = "no network is connected"
	}
	if why == "" {
		inst, err := s.Store.Installation(ctx)
		if err == nil {
			err = s.Network.NoteUntaught(ctx, inst, missing, kind, detail)
		}
		if err != nil {
			log.Printf("network: untaught: %v", err)
			why = "the network could not be reached"
		}
	}
	shared := "yes"
	if why != "" {
		shared = "no"
	}
	inputs := map[string]any{"asked": asked, "missing": missing, "asked_by": askedBy, "shared": shared, "kind": kind,
		"occurred_at": s.Clock.TodayUTC()}
	if detail != "" {
		inputs["detail"] = detail
	}
	if why != "" {
		inputs["kept_home_because"] = why
	}
	t, err := s.Exec.TriggerActivity(ctx, "note_untaught", inputs, askedBy, "")
	if err != nil {
		return noted{}, err
	}
	return noted{EventID: t.EventID, Kind: kind, Shared: shared == "yes", KeptHome: why}, nil
}

// handleNoteUntaught records an untaught moment the face saw.
func (s *Server) handleNoteUntaught(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Asked   string `json:"asked"`
		Missing string `json:"missing"`
		AskedBy string `json:"asked_by"`
		Kind    string `json:"kind"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, 400, err)
		return
	}
	n, err := s.noteUntaught(r.Context(), req.Asked, req.Missing, req.AskedBy, req.Kind, "")
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, n)
}

// handleUntaught lists what the host was asked and had not been taught,
// newest first: what the person asked, what she said was missing, and
// whether it reached the house.
func (s *Server) handleUntaught(w http.ResponseWriter, r *http.Request) {
	if !s.allowSystem(w, r, ViewUntaught) {
		return
	}
	n := 20
	if v := r.URL.Query().Get("n"); v != "" {
		k, err := strconv.Atoi(v)
		if err != nil || k < 1 {
			writeErr(w, 400, fmt.Errorf("untaught: n must be a positive number"))
			return
		}
		n = k
	}
	evs, err := s.Store.EventsOfTypes(r.Context(), []string{core.EventHostUntaught})
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	type untaught struct {
		EventID    int64  `json:"event_id"`
		Asked      string `json:"asked"`
		Missing    string `json:"missing"`
		AskedBy    string `json:"asked_by"`
		Shared     bool   `json:"shared"`
		KeptHome   string `json:"kept_home_because,omitempty"`
		Kind       string `json:"kind"`
		Detail     string `json:"detail,omitempty"`
		OccurredAt string `json:"occurred_at"`
	}
	out := []untaught{}
	for i := len(evs) - 1; i >= 0 && len(out) < n; i-- {
		var p struct {
			Asked    string `json:"asked"`
			Missing  string `json:"missing"`
			AskedBy  string `json:"asked_by"`
			Shared   string `json:"shared"`
			KeptHome string `json:"kept_home_because"`
			Kind     string `json:"kind"`
			Detail   string `json:"detail"`
		}
		if json.Unmarshal(evs[i].Payload, &p) != nil {
			continue
		}
		out = append(out, untaught{EventID: evs[i].ID, Asked: p.Asked, Missing: p.Missing, AskedBy: p.AskedBy,
			Shared: p.Shared == "yes", KeptHome: p.KeptHome, Kind: cmp.Or(p.Kind, "untaught"), Detail: p.Detail, OccurredAt: evs[i].OccurredAt})
	}
	writeJSON(w, 200, out)
}
