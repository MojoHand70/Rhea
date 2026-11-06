package shell

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"rhea/internal/core"
	"rhea/internal/rights"
)

// The stream of events (DIRECTION 2026-11-05, "The desk as a host"; KK:
// "we say Rhea takes an event, checks for rules and does the job, but we
// never show events" — and the next night: "each event should be an event,
// not just a row"). GET /api/stream answers with cards, newest first: each
// a raw fact that came in, titled in the pack's words, and under it the
// steps the rules took — the chain of derived events, each step a sentence
// of the pack's with the step's own facts in it, the area that handled it,
// and whether it waits for a person. The kernel renders the sentences from
// the words (core.StepWords); it knows no language and no business of its
// own, so a rule without words still shows as its type's name and label.
// Like explain, a native surface: it renders the invariant layer.
//
//	?limit=20            cards per page (at most 100)
//	?before=<event id>   the page before that fact
//	?object=<object id>  a record's history: the chains that name it
//
// Scope: the company acted for sees its own facts (store.RecentRoots); a
// signed-in person sees a card when some step of it made a record of a type
// they may read, or, for a card no rule answered, when they may see the
// worklist. A step the pack marks as waiting for a person waits until a later
// fact names its record (store.NamedAfter), then it stands as done. Assessed
// and left (2026-11-06): one chain query per card, one more per waiting step.

type streamStep struct {
	Says        []core.SayPart `json:"says"`
	Area        string         `json:"area,omitempty"`
	State       string         `json:"state"`
	RuleID      string         `json:"rule_id"`
	RuleVersion int            `json:"rule_version"`
	EventID     int64          `json:"event_id"`
	ObjectID    string         `json:"object_id,omitempty"`
	ObjectType  string         `json:"object_type,omitempty"`
	N           int            `json:"n,omitempty"` // how many facts of one rule this step stands for
}

type streamCard struct {
	EventID    int64          `json:"event_id"`
	OccurredAt string         `json:"occurred_at"`
	EventType  string         `json:"event_type"`
	Actor      string         `json:"actor,omitempty"`
	Activity   string         `json:"activity,omitempty"`
	Title      []core.SayPart `json:"title"`
	Source     string         `json:"source,omitempty"` // the pack's word for where it came from: Bank, KSeF, Sprzedaż
	State      string         `json:"state"`            // ok | wait | no | none
	ObjectID   string         `json:"object_id,omitempty"`
	ObjectType string         `json:"object_type,omitempty"`
	Steps      []streamStep   `json:"steps"`
}

func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	limit := 20
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			writeErr(w, 400, fmt.Errorf("stream: limit %q is not a count", v))
			return
		}
		limit = min(n, 100)
	}
	var before int64
	if v := q.Get("before"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			writeErr(w, 400, fmt.Errorf("stream: before %q is not an event id", v))
			return
		}
		before = n
	}
	var roots []core.Event
	var err error
	if obj := q.Get("object"); obj != "" {
		roots, err = s.Store.RootsNaming(ctx, obj, limit+1)
	} else {
		roots, err = s.Store.RecentRoots(ctx, before, limit+1)
	}
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	more := len(roots) > limit
	if more {
		roots = roots[:limit]
	}
	teller, err := s.newTeller(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	p := rights.PrincipalOf(ctx)
	cards := make([]streamCard, 0, len(roots))
	for _, root := range roots {
		card, err := teller.card(ctx, root)
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		if p != nil && !s.maySeeCard(ctx, p, teller, card) {
			continue
		}
		cards = append(cards, card)
	}
	resp := map[string]any{"cards": cards}
	if more && len(roots) > 0 {
		resp["next_before"] = roots[len(roots)-1].ID
	}
	writeJSON(w, 200, resp)
}

// maySeeCard: a person sees a card when a step of it made or moved a record
// of a type they may read; a card no rule answered is the worklist's.
func (s *Server) maySeeCard(ctx context.Context, p *rights.Principal, t *teller, c streamCard) bool {
	if len(c.Steps) == 0 {
		return s.may(ctx, p, "view", ViewWorklist, "system")
	}
	for _, st := range c.Steps {
		if st.ObjectType == "" {
			continue
		}
		if s.may(ctx, p, "type", st.ObjectType, t.types[st.ObjectType].Domain) {
			return true
		}
	}
	return false
}

// teller turns a chain into a card with the company's words.
type teller struct {
	s     *Server
	words core.WordsSpec
	types map[string]core.ObjectType
}

func (s *Server) newTeller(ctx context.Context) (*teller, error) {
	t := &teller{s: s, types: map[string]core.ObjectType{}}
	if w, err := s.mergedWords(ctx); err != nil {
		return nil, err
	} else if w != nil {
		t.words = *w
	}
	types, err := s.Store.ListObjectTypes(ctx)
	if err != nil {
		return nil, err
	}
	for _, ty := range types {
		t.types[ty.Name] = ty
	}
	return t, nil
}

func (t *teller) typeWord(name string) string {
	if tw, ok := t.words.Types[name]; ok && tw.Label != "" {
		return tw.Label
	}
	return name
}

func (t *teller) labelOf(typ string, state map[string]any) string {
	if ty, ok := t.types[typ]; ok && ty.LabelField != "" {
		if s, ok := state[ty.LabelField].(string); ok {
			return s
		}
	}
	return ""
}

func lookupIn(maps ...map[string]any) func(string) (any, bool) {
	return func(k string) (any, bool) {
		for _, m := range maps {
			if v, ok := m[k]; ok && v != nil {
				return v, true
			}
		}
		return nil, false
	}
}

// card tells one root's story: the title from the raw fact, then the chain
// walked depth-first, siblings of one rule under one cause folded into one
// step that counts them.
func (t *teller) card(ctx context.Context, root core.Event) (streamCard, error) {
	chain, err := t.s.Store.ChainOf(ctx, root.ID)
	if err != nil {
		return streamCard{}, err
	}
	children := map[int64][]core.Event{}
	for _, ev := range chain {
		if ev.CauseEventID != nil {
			children[*ev.CauseEventID] = append(children[*ev.CauseEventID], ev)
		}
	}
	c := streamCard{EventID: root.ID, OccurredAt: root.OccurredAt, EventType: root.Type, Actor: root.Actor, Activity: root.ActivityName, State: "ok", Steps: []streamStep{}}
	var payload map[string]any
	_ = json.Unmarshal(root.Payload, &payload)
	if payload == nil {
		payload = map[string]any{}
	}
	// the first record the fact made names the card: FS 5/2026, HOS/1/2026
	first := map[string]any{}
	for _, ev := range children[root.ID] {
		if ev.Type == core.EventObjectMaterialized {
			var mat core.MaterializedObject
			if json.Unmarshal(ev.Payload, &mat) == nil {
				c.ObjectID, c.ObjectType = mat.ObjectID, mat.ObjectType
				first["label"], first["type"] = t.labelOf(mat.ObjectType, mat.State), t.typeWord(mat.ObjectType)
				break
			}
		}
	}
	if sw, ok := t.words.Steps["event:"+root.Type]; ok {
		c.Title = t.words.Say(sw.Says, lookupIn(payload, first))
		c.Source = sw.Area
	} else {
		name := t.words.Events[root.Type]
		if name == "" {
			name = root.Type
		}
		c.Title = []core.SayPart{{T: name}}
		if lbl, _ := first["label"].(string); lbl != "" {
			c.Title = append(c.Title, core.SayPart{T: " · "}, core.SayPart{F: lbl})
		}
	}
	grave := ""
	var walk func(parent int64)
	walk = func(parent int64) {
		kids := children[parent]
		for i := 0; i < len(kids); {
			// siblings of one rule under one cause: one step, counted
			j := i + 1
			for j < len(kids) && kids[j].RuleID == kids[i].RuleID && kids[j].Type == kids[i].Type {
				j++
			}
			group := kids[i:j]
			step, say := t.step(ctx, group)
			if say {
				c.Steps = append(c.Steps, step)
				switch step.State {
				case "wait":
					grave = "wait"
				case "no":
					if grave != "wait" {
						grave = "no"
					}
				}
			}
			for _, ev := range group {
				walk(ev.ID)
			}
			i = j
		}
	}
	walk(root.ID)
	switch {
	case len(chain) == 0:
		c.State = "none"
	case grave != "":
		c.State = grave
	}
	return c, nil
}

// step words one group of derived facts of one rule: the pack's sentence
// over the step's facts, else the type's name and the record's label.
func (t *teller) step(ctx context.Context, group []core.Event) (streamStep, bool) {
	ev := group[0]
	st := streamStep{RuleID: ev.RuleID, RuleVersion: ev.RuleVersion, EventID: ev.ID, State: "ok", N: len(group)}
	facts := map[string]any{"n": len(group), "strategy": "", "actor": ev.Actor}
	var typ, label string
	switch ev.Type {
	case core.EventObjectMaterialized:
		var mat core.MaterializedObject
		if json.Unmarshal(ev.Payload, &mat) == nil {
			typ, label = mat.ObjectType, t.labelOf(mat.ObjectType, mat.State)
			for k, v := range mat.State {
				facts[k] = v
			}
			st.ObjectID = mat.ObjectID
		}
	case core.EventObjectAmended:
		var am core.AmendedObject
		if json.Unmarshal(ev.Payload, &am) == nil {
			typ = am.ObjectType
			st.ObjectID = am.ObjectID
			// the record as it stands now, under what this step set
			if o, err := t.s.Store.GetObject(ctx, am.ObjectID); err == nil {
				for k, v := range o.State {
					facts[k] = v
				}
				label = t.labelOf(o.Type, o.State)
			}
			for k, v := range am.Set {
				facts[k] = v
			}
		}
	case core.EventObjectsMatched:
		var m core.MatchedObjects
		if json.Unmarshal(ev.Payload, &m) == nil {
			facts["strategy"], facts["matching"] = m.Strategy, m.Matching
			typ, st.ObjectID = m.Trigger.ObjectType, m.Trigger.ObjectID
			// the other side: what the trigger was matched to, by label
			var others []string
			for side, allocs := range m.Sides {
				if side == m.Trigger.Side {
					continue
				}
				for _, a := range allocs {
					if o, err := t.s.Store.GetObject(ctx, a.ObjectID); err == nil {
						if l := t.labelOf(o.Type, o.State); l != "" {
							others = append(others, l)
						}
					}
				}
			}
			facts["other"] = strings.Join(others, ", ")
			facts["others"] = len(others)
		}
	case core.EventObjectsUnmatched:
		var u core.UnmatchedObjects
		if json.Unmarshal(ev.Payload, &u) == nil {
			typ, st.ObjectID = u.Trigger.ObjectType, u.Trigger.ObjectID
			facts["tried"] = len(u.Tried)
		}
	}
	st.ObjectType = typ
	facts["label"], facts["type"] = label, t.typeWord(typ)
	sw, ok := t.words.Steps[ev.RuleID]
	if !ok && typ != "" {
		sw, ok = t.words.Steps["type:"+typ]
	}
	if ok {
		if strings.TrimSpace(sw.Says) == "" {
			return st, false // silent: said by another step
		}
		st.Says, st.Area = t.words.Say(sw.Says, lookupIn(facts)), sw.Area
		if sw.State != "" {
			st.State = sw.State
		}
		// a step that waits for a person waits until a later fact names its record — the approval, the
		// refusal, the decision; then it stands as done, whatever the pack's word for the step says
		if st.State == "wait" && st.ObjectID != "" {
			if named, err := t.s.Store.NamedAfter(ctx, st.ObjectID, ev.ID); err == nil && named {
				st.State = "ok"
			}
		}
		return st, true
	}
	// no words: the kind of record and its label, the event's name for a match
	switch ev.Type {
	case core.EventObjectsMatched, core.EventObjectsUnmatched:
		name := t.words.Events[ev.Type]
		if name == "" {
			name = ev.Type
		}
		st.Says = []core.SayPart{{T: name}}
		if other, _ := facts["other"].(string); other != "" {
			st.Says = append(st.Says, core.SayPart{T: ": "}, core.SayPart{F: other})
		}
	default:
		st.Says = []core.SayPart{{T: t.typeWord(typ)}}
		if label != "" {
			st.Says = append(st.Says, core.SayPart{T: " "}, core.SayPart{F: label})
		}
		if len(group) > 1 {
			st.Says = append(st.Says, core.SayPart{T: " × "}, core.SayPart{F: strconv.Itoa(len(group))})
		}
	}
	return st, true
}
