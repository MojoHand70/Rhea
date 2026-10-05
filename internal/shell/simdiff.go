package shell

import (
	"context"
	"sort"

	"rhea/internal/core"
	"rhea/internal/exec"
)

// The readable simulation diff (DIRECTION 2026-10-05): the approval gate is
// the trust boundary of the whole design, and what the human reads at that
// moment must be field-level and typed, never a JSON blob. The kernel's
// SimDiff crosses the boundary through the same cell encoding as every view;
// refs resolve against the hypothetical world first, because an added object
// may reference another object that only exists if the rule is approved.
// Cells carry no doors here — half of this world does not exist yet.

type diffItem struct {
	Field string `json:"field"`
	Type  string `json:"type,omitempty"`
	// Added objects fill only After, removed only Before, changed both.
	Before  *cellOut `json:"before,omitempty"`
	After   *cellOut `json:"after,omitempty"`
	Changed bool     `json:"changed,omitempty"`
}

type diffObject struct {
	ObjectID   string `json:"object_id"`
	ObjectType string `json:"object_type"`
	Title      string `json:"title,omitempty"` // label-field value, when the type declares one
	// Detail opens the object as it exists today — so never on an added one.
	Detail string     `json:"detail,omitempty"`
	Fields []diffItem `json:"fields"`
}

type explainedEvent struct {
	EventID    int64  `json:"event_id"`
	EventType  string `json:"event_type"`
	OccurredAt string `json:"occurred_at"`
}

// renderSimDiff shapes the kernel's diff for renderers.
func (s *Server) renderSimDiff(ctx context.Context, diff exec.SimDiff) (map[string]any, error) {
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

	// The hypothetical world: everything the after-run would hold differently.
	simWorld := map[string]core.Object{}
	for _, o := range diff.Added {
		simWorld[o.ID] = o
	}
	for _, c := range diff.Changed {
		simWorld[c.After.ID] = c.After
	}
	resolve := func(id string) (string, string) {
		if o, ok := simWorld[id]; ok {
			label := ""
			if t, ok := typeByName[o.Type]; ok && t.LabelField != "" {
				label, _ = o.State[t.LabelField].(string)
			}
			return label, o.Type
		}
		return s.refInfo(ctx, id)
	}
	noDoors := map[string]string{}
	cell := func(state map[string]any, fd core.FieldDef) *cellOut {
		c := cellWith(state[fd.Name], fd, resolve, noDoors)
		return &c
	}
	// fieldDefs lists the fields to show: the type's declaration, or the
	// state's own keys when the type is unknown.
	fieldDefs := func(o core.Object) []core.FieldDef {
		if t, ok := typeByName[o.Type]; ok {
			return t.Fields
		}
		keys := make([]string, 0, len(o.State))
		for k := range o.State {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		fds := make([]core.FieldDef, len(keys))
		for i, k := range keys {
			fds[i] = core.FieldDef{Name: k, Type: "string"}
		}
		return fds
	}
	title := func(o core.Object) string {
		if t, ok := typeByName[o.Type]; ok && t.LabelField != "" {
			label, _ := o.State[t.LabelField].(string)
			return label
		}
		return ""
	}

	oneSided := func(objs []core.Object, after bool, doors bool) []diffObject {
		out := make([]diffObject, 0, len(objs))
		for _, o := range objs {
			d := diffObject{ObjectID: o.ID, ObjectType: o.Type, Title: title(o), Fields: []diffItem{}}
			if doors {
				d.Detail = detailFor[o.Type]
			}
			for _, fd := range fieldDefs(o) {
				it := diffItem{Field: fd.Name, Type: fd.Type}
				if after {
					it.After = cell(o.State, fd)
				} else {
					it.Before = cell(o.State, fd)
				}
				d.Fields = append(d.Fields, it)
			}
			out = append(out, d)
		}
		return out
	}

	changed := make([]diffObject, 0, len(diff.Changed))
	for _, c := range diff.Changed {
		d := diffObject{ObjectID: c.After.ID, ObjectType: c.After.Type,
			Title: title(c.After), Detail: detailFor[c.After.Type], Fields: []diffItem{}}
		for _, fd := range fieldDefs(c.After) {
			it := diffItem{Field: fd.Name, Type: fd.Type,
				Before: cell(c.Before.State, fd), After: cell(c.After.State, fd)}
			it.Changed = it.Before.V != it.After.V || it.Before.ID != it.After.ID
			d.Fields = append(d.Fields, it)
		}
		changed = append(changed, d)
	}

	// The events this rule would newly explain: unexplained before, not after.
	afterSet := map[int64]bool{}
	for _, id := range diff.UnexplainedAfter {
		afterSet[id] = true
	}
	explains := []explainedEvent{}
	for _, id := range diff.UnexplainedBefore {
		if afterSet[id] {
			continue
		}
		ev, err := s.Store.GetEvent(ctx, id)
		if err != nil {
			return nil, err
		}
		explains = append(explains, explainedEvent{
			EventID: ev.ID, EventType: ev.Type, OccurredAt: ev.OccurredAt})
	}

	errs := diff.Errors
	if errs == nil {
		errs = []string{}
	}
	return map[string]any{
		"rule_id": diff.RuleID, "version": diff.Version,
		"added":   oneSided(diff.Added, true, false),
		"changed": changed,
		"removed": oneSided(diff.Removed, false, true),
		"unexplained_before": len(diff.UnexplainedBefore),
		"unexplained_after":  len(diff.UnexplainedAfter),
		"explains":           explains,
		"errors":             errs,
	}, nil
}
