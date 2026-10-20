package shell

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sort"

	"rhea/internal/core"
	"rhea/internal/store"
)

// The queue notion (DIRECTION 2026-10-10, "The host is a concierge": what to
// do today). The objects of one type that match the view's conditions, ranked
// by a formula over each object's state — today in the payload, so "days
// overdue" is arithmetic the language already has — highest first; on every
// item, the verbs a person can take on it: every active activity with an
// input of the item's type, offered inline, prefilled; and, when the view
// asks, the approvals waiting at the gate. No screen knows what a task or
// an invoice is; the view says what is open, how to rank it, what to show.

type queueItem struct {
	ObjectID string             `json:"object_id"`
	Label    string             `json:"label"`
	Detail   string             `json:"detail,omitempty"`
	Rank     string             `json:"rank,omitempty"`
	Problem  string             `json:"problem,omitempty"` // why the formula could not rank it
	Cells    map[string]cellOut `json:"cells"`
	rank     *big.Rat
	seq      int
}

type queueApproval struct {
	BundleID    string `json:"bundle_id"`
	Description string `json:"description"`
	Question    string `json:"question,omitempty"`
}

func (s *Server) serveQueue(ctx context.Context, w http.ResponseWriter, vd *core.ViewDef) {
	var spec core.QueueSpec
	if err := json.Unmarshal(vd.Spec, &spec); err != nil {
		writeErr(w, 500, err)
		return
	}
	ot, err := s.Store.GetObjectType(ctx, spec.ObjectType)
	if err != nil {
		writeErr(w, 400, fmt.Errorf("queue %s: %w", vd.ID, err))
		return
	}
	rank, err := core.ParseTemplate(spec.Rank)
	if err != nil {
		writeErr(w, 400, fmt.Errorf("queue %s rank: %w", vd.ID, err))
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
	detailFor := detailViewByType(vds)
	today := s.Clock.TodayUTC()
	env := s.formulaEnv(ctx, map[string]string{"$.state": spec.ObjectType})
	columns := make([]columnOut, 0, len(spec.Columns))
	for _, f := range spec.Columns {
		fd, ok := ot.Field(f)
		if !ok {
			writeErr(w, 400, fmt.Errorf("queue %s: column %q is not a field of %s", vd.ID, f, spec.ObjectType))
			return
		}
		columns = append(columns, columnOut{Field: f, Label: humanizeField(f), Type: fd.Type})
	}
	var items []queueItem
	for i, o := range objs {
		payload := map[string]any{"object_id": o.ID, "object_type": o.Type, "state": o.State, "today": today}
		open := true
		for _, c := range spec.Where {
			ok, err := core.EvalCondition(payload, c)
			if err != nil {
				writeErr(w, 400, fmt.Errorf("queue %s: %w", vd.ID, err))
				return
			}
			if !ok {
				open = false
			}
		}
		if !open {
			continue
		}
		item := queueItem{ObjectID: o.ID, Label: o.ID, Detail: detailFor[o.Type], Cells: map[string]cellOut{}, seq: i}
		// An item the formula cannot rank — a task with no due date, say — is
		// still something to do: listed last, with the reason where its rank
		// would be, never dropped and never failing the queue.
		if v, _, err := rank.Evaluate(payload, core.FieldDef{Name: "rank", Type: "decimal"}, env); err != nil {
			item.Problem = err.Error()
		} else if text, _ := v.(string); text == "" {
			item.Problem = "the rank is empty"
		} else if r, err := core.ParseDecimal(text); err != nil {
			item.Problem = fmt.Sprintf("the rank is %v, not a number", v)
		} else {
			item.Rank, item.rank = text, r
		}
		if ot.LabelField != "" {
			if l, ok := o.State[ot.LabelField].(string); ok && l != "" {
				item.Label = l
			}
		}
		for _, c := range columns {
			fd, _ := ot.Field(c.Field)
			item.Cells[c.Field] = s.typedCell(ctx, o.State[c.Field], fd, detailFor)
		}
		items = append(items, item)
	}
	// Highest rank first; the unranked last; equal ranks in log order, so
	// the queue is stable.
	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i].rank, items[j].rank
		switch {
		case a == nil && b == nil:
			return items[i].seq < items[j].seq
		case a == nil:
			return false
		case b == nil:
			return true
		}
		if c := a.Cmp(b); c != 0 {
			return c > 0
		}
		return items[i].seq < items[j].seq
	})
	if spec.Limit > 0 && len(items) > spec.Limit {
		items = items[:spec.Limit]
	}
	if items == nil {
		items = []queueItem{}
	}
	// The verbs that take an item: the detail view's contextual verbs, every
	// active activity with a ref input of the item's type.
	out := map[string]any{"view": vd, "today": today, "columns": columns, "items": items, "verbs": s.verbsForType(ctx, spec.ObjectType)}
	if spec.Approvals {
		bundles, err := s.Store.LatestBundles(ctx)
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		approvals := []queueApproval{}
		for _, b := range bundles {
			if b.Status == core.StatusDraft {
				approvals = append(approvals, queueApproval{BundleID: b.ID, Description: b.Description, Question: b.Question})
			}
		}
		out["approvals"] = approvals
	}
	writeJSON(w, 200, out)
}

// formulaEnv is the environment a view's formula runs in: lookups and reads
// against the cache, the catalog, and which payload prefixes hold typed
// state. Views compute nothing into the log; a rank is read-side arithmetic
// over baked state.
func (s *Server) formulaEnv(ctx context.Context, typed map[string]string) *core.Env {
	memo := map[string]*core.ObjectType{}
	return &core.Env{
		Lookup: func(typ, field string, value any) ([]string, error) {
			return s.Store.FindObjectIDsByField(ctx, typ, field, core.Text(value))
		},
		Get: func(id string) (map[string]any, bool, error) {
			o, err := s.Store.GetObject(ctx, id)
			if err != nil {
				return nil, false, nil
			}
			return o.State, true, nil
		},
		Types: func(name string) (core.ObjectType, bool) {
			if t, ok := memo[name]; ok {
				return *t, t != nil
			}
			t, err := s.Store.GetObjectType(ctx, name)
			if err != nil {
				if errors.Is(err, store.ErrNotFound) {
					memo[name] = nil // a type that is not there; a failure is not memoized
				}
				return core.ObjectType{}, false
			}
			memo[name] = &t
			return t, true
		},
		Typed: typed,
	}
}
