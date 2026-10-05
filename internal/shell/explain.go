package shell

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"rhea/internal/core"
	"rhea/internal/store"
)

// The provenance walk (DIRECTION 2026-10-05): invariant 5 as an interaction.
// GET /api/explain?object=<id> (or ?event=<id>) answers "why does this
// exist?" with the whole story: the chain's root raw event and every derived
// event it caused, each hop naming its exact rule version and the object it
// materialized. Both sides of an intercompany position walk to the same
// root — nothing to reconcile, only to display. Like worklist and rules,
// this is a native surface: it renders the invariant layer, not a ViewDef.

type explainObject struct {
	ObjectID     string `json:"object_id"`
	ObjectType   string `json:"object_type"`
	Label        string `json:"label,omitempty"`
	DetailViewID string `json:"detail_view_id,omitempty"`
}

type explainNode struct {
	EventID    int64  `json:"event_id"`
	Kind       string `json:"kind"`
	EventType  string `json:"event_type"`
	OccurredAt string `json:"occurred_at"`
	Actor      string `json:"actor,omitempty"`
	// The door a raw event came through: the walk starts one hop earlier
	// than the rules — which declared verb created this fact.
	Activity        string `json:"activity,omitempty"`
	ActivityVersion int    `json:"activity_version,omitempty"`
	RuleID          string `json:"rule_id,omitempty"`
	RuleVersion     int    `json:"rule_version,omitempty"`
	RuleDesc        string `json:"rule_description,omitempty"`
	// Object is what this event materialized, when it did.
	Object *explainObject `json:"object,omitempty"`
	// Payload rides only on the raw root: the fact everything explains.
	Payload  json.RawMessage `json:"payload,omitempty"`
	Children []*explainNode  `json:"children"`
}

// chainIndex is the derived log, indexed for walking both directions.
type chainIndex struct {
	byID       map[int64]core.Event
	children   map[int64][]core.Event
	mats       map[int64]core.MaterializedObject // materialization event → payload
	matEventOf map[string]int64                  // object id → its materialization event
}

func (s *Server) indexChains(ctx context.Context) (*chainIndex, error) {
	derived, err := s.Store.EventsByKind(ctx, core.KindDerived)
	if err != nil {
		return nil, err
	}
	ix := &chainIndex{byID: map[int64]core.Event{}, children: map[int64][]core.Event{},
		mats: map[int64]core.MaterializedObject{}, matEventOf: map[string]int64{}}
	for _, ev := range derived {
		ix.byID[ev.ID] = ev
		if ev.CauseEventID != nil {
			ix.children[*ev.CauseEventID] = append(ix.children[*ev.CauseEventID], ev)
		}
		if ev.Type == core.EventObjectMaterialized {
			var mat core.MaterializedObject
			if json.Unmarshal(ev.Payload, &mat) == nil {
				ix.mats[ev.ID] = mat
				ix.matEventOf[mat.ObjectID] = ev.ID
			}
		}
	}
	return ix, nil
}

func (s *Server) handleExplain(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ix, err := s.indexChains(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}

	// What was asked: an object (via its materialization event) or an event.
	resp := map[string]any{}
	var start int64
	switch {
	case r.URL.Query().Get("object") != "":
		objID := r.URL.Query().Get("object")
		o, err := s.Store.GetObject(ctx, objID)
		if err != nil {
			writeErr(w, 404, err)
			return
		}
		evID, ok := ix.matEventOf[objID]
		if !ok {
			writeErr(w, 404, fmt.Errorf("no materialization event for %q", objID))
			return
		}
		start, resp["object"] = evID, o
	case r.URL.Query().Get("event") != "":
		id, err := strconv.ParseInt(r.URL.Query().Get("event"), 10, 64)
		if err != nil {
			writeErr(w, 400, fmt.Errorf("explain: bad event id"))
			return
		}
		start = id
	default:
		writeErr(w, 400, fmt.Errorf("explain wants ?object= or ?event="))
		return
	}

	// Walk up: causes precede effects in the log, so this terminates at the
	// first event that is not derived — the chain's raw root.
	path := []int64{start}
	cur := start
	for {
		ev, ok := ix.byID[cur]
		if !ok || ev.CauseEventID == nil {
			break
		}
		cur = *ev.CauseEventID
		path = append(path, cur)
	}
	root, err := s.Store.GetEvent(ctx, cur)
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	// Root-first, the way the story reads.
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	tree, err := s.buildExplainTree(ctx, ix, root)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	resp["root_event_id"], resp["path"], resp["tree"] = root.ID, path, tree
	writeJSON(w, 200, resp)
}

// buildExplainTree renders the root's whole consequence tree, each node
// annotated with its rule version and materialized object.
func (s *Server) buildExplainTree(ctx context.Context, ix *chainIndex, root core.Event) (*explainNode, error) {
	descs, err := s.Store.RuleDescriptions(ctx)
	if err != nil {
		return nil, err
	}
	types, err := s.Store.ListObjectTypes(ctx)
	if err != nil {
		return nil, err
	}
	typeByName := map[string]core.ObjectType{}
	for _, t := range types {
		typeByName[t.Name] = t
	}
	vds, err := s.effectiveViewDefs(ctx)
	if err != nil {
		return nil, err
	}
	detailFor := detailViewByType(vds)

	var build func(ev core.Event) *explainNode
	build = func(ev core.Event) *explainNode {
		n := &explainNode{EventID: ev.ID, Kind: ev.Kind, EventType: ev.Type,
			OccurredAt: ev.OccurredAt, Actor: ev.Actor,
			Activity: ev.ActivityName, ActivityVersion: ev.ActivityVersion,
			RuleID: ev.RuleID, RuleVersion: ev.RuleVersion,
			RuleDesc: descs[store.RuleKey{ID: ev.RuleID, Version: ev.RuleVersion}],
			Children: []*explainNode{}}
		if ev.Kind == core.KindRaw {
			n.Payload = ev.Payload
		}
		if mat, ok := ix.mats[ev.ID]; ok {
			obj := &explainObject{ObjectID: mat.ObjectID, ObjectType: mat.ObjectType,
				DetailViewID: detailFor[mat.ObjectType]}
			if t, ok := typeByName[mat.ObjectType]; ok && t.LabelField != "" {
				if lbl, ok := mat.State[t.LabelField].(string); ok {
					obj.Label = lbl
				}
			}
			n.Object = obj
		}
		for _, c := range ix.children[ev.ID] {
			n.Children = append(n.Children, build(c))
		}
		return n
	}
	return build(root), nil
}

// detailViewByType maps each object type to the detail view that opens it.
// Stored views win over derived ones because effectiveViewDefs lists them
// first — the same preference serveList applies.
func detailViewByType(vds []core.ViewDef) map[string]string {
	out := map[string]string{}
	for _, v := range vds {
		if v.Notion != "detail" {
			continue
		}
		var sp core.DetailSpec
		if json.Unmarshal(v.Spec, &sp) != nil {
			continue
		}
		if _, seen := out[sp.ObjectType]; !seen {
			out[sp.ObjectType] = v.ID
		}
	}
	return out
}
