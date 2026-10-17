package shell

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"rhea/internal/core"
	"rhea/internal/store"
)

// The parties a business deals with, whose names never leave it: the
// companies (the partition key's own type) and the people its actors are.
var namedTypes = []string{"company", "person"}

// handleNoteUntaught records what the host was asked and had not been taught
// (DIRECTION 2026-10-17). The person's words stay here; her general
// description reaches the network when it carries no number, no address and
// no name the installation holds, and when a network is connected — and the
// fact says which, so the owner sees everything she reported.
func (s *Server) handleNoteUntaught(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		Asked   string `json:"asked"`
		Missing string `json:"missing"`
		AskedBy string `json:"asked_by"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 16<<10)).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	req.Asked, req.Missing = strings.TrimSpace(req.Asked), strings.Join(strings.Fields(req.Missing), " ")
	if req.Asked == "" || len(req.Asked) > 2000 || req.Missing == "" || len(req.Missing) > 400 || req.AskedBy == "" {
		writeErr(w, 400, fmt.Errorf("untaught: asked (1–2000 bytes), missing (1–400) and asked_by are required"))
		return
	}
	var names []string
	for _, t := range namedTypes {
		objs, err := s.Store.ObjectsByType(store.WithCompany(ctx, ""), t)
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		for _, o := range objs {
			for _, f := range []string{"name", "handle"} {
				if v, ok := o.State[f].(string); ok && v != "" {
					names = append(names, v)
				}
			}
		}
	}
	why := core.KeptHome(req.Missing, names)
	if why == "" && s.Network == nil {
		why = "no network is connected"
	}
	if why == "" {
		inst, err := s.Store.Installation(ctx)
		if err == nil {
			err = s.Network.NoteUntaught(ctx, inst, req.Missing)
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
	inputs := map[string]any{"asked": req.Asked, "missing": req.Missing, "asked_by": req.AskedBy, "shared": shared,
		"occurred_at": s.Clock.TodayUTC()}
	if why != "" {
		inputs["kept_home_because"] = why
	}
	t, err := s.Exec.TriggerActivity(ctx, "note_untaught", inputs, req.AskedBy, "")
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"event_id": t.EventID, "shared": shared == "yes", "kept_home_because": why})
}

// handleUntaught lists what the host was asked and had not been taught,
// newest first: what the person asked, what she said was missing, and
// whether it reached the house.
func (s *Server) handleUntaught(w http.ResponseWriter, r *http.Request) {
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
		}
		if json.Unmarshal(evs[i].Payload, &p) != nil {
			continue
		}
		out = append(out, untaught{EventID: evs[i].ID, Asked: p.Asked, Missing: p.Missing, AskedBy: p.AskedBy,
			Shared: p.Shared == "yes", KeptHome: p.KeptHome, OccurredAt: evs[i].OccurredAt})
	}
	writeJSON(w, 200, out)
}
