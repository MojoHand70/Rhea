package shell

import (
	"net/http"
	"sort"
	"strings"

	"rhea/internal/rights"
)

// A waitingMove is a record waiting for a move its type gives to a group
// (KK, 2026-10-27, "Tasks, groups and the phone"): the record, the move,
// the door that makes it and the words on its button, and who may make it.
type waitingMove struct {
	Object  string         `json:"object"`
	Type    string         `json:"type"`
	Company string         `json:"company"`
	From    string         `json:"from"`
	To      string         `json:"to"`
	Door    string         `json:"door"`
	Label   string         `json:"label,omitempty"`
	Input   string         `json:"input,omitempty"` // the door's input that names the record
	Group   string         `json:"group"`
	Movers  []string       `json:"movers"` // who may make it, by handle
	State   map[string]any `json:"state"`
}

// handleMoves is GET /api/moves: every record the caller sees that waits for
// a move — for a signed-in person, only the moves they may make (not one
// they proposed while someone else could), so Dziś offers each as a
// decision with its buttons. Read in the caller's scope: another company's
// records are never among them.
func (s *Server) handleMoves(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	p := rights.PrincipalOf(ctx)
	types, err := s.Store.ListObjectTypes(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	check, err := rights.NewMoveCheck(ctx, s.Store, s.Clock.Today())
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	out := []waitingMove{}
	for _, t := range types {
		lc := t.Lifecycle
		if lc == nil || len(lc.Moves) == 0 {
			continue
		}
		if p != nil {
			if _, ok := p.MayAnywhere("type", t.Name, t.Domain); !ok {
				continue
			}
		}
		objs, err := s.Store.ObjectsByType(ctx, t.Name)
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		for _, o := range objs {
			if p != nil && o.Company == "" {
				continue // the installation's own record is its operator's, never on a person's Dziś
			}
			state, _ := o.State[lc.Field].(string)
			for key, mv := range lc.Moves {
				from, to, _ := strings.Cut(key, ">")
				if from != state {
					continue
				}
				proposer := ""
				if mv.NotBy != "" {
					proposer, _ = o.State[mv.NotBy].(string)
				}
				if p != nil && check.May(p.Handle, o.Company, mv.Group, proposer) != "" {
					continue
				}
				wm := waitingMove{Object: o.ID, Type: t.Name, Company: o.Company, From: from, To: to, Door: mv.Door, Label: mv.Label,
					Group: mv.Group, Movers: check.Movers(o.Company, mv.Group), State: o.State}
				if act, err := s.Store.GetActivity(ctx, mv.Door); err == nil {
					for _, f := range act.Spec.Inputs {
						if f.Type == "ref<"+t.Name+">" {
							wm.Input = f.Name
						}
					}
				}
				out = append(out, wm)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Object != out[j].Object {
			return out[i].Object < out[j].Object
		}
		return out[i].To < out[j].To
	})
	writeJSON(w, 200, out)
}
