package pack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"slices"
	"strings"

	"rhea/internal/core"
	"rhea/internal/exec"
)

// AppFile is an App as its pack writes it (KK 2026-10-26, APPS.md "The App
// definition"): one file, apps/<app>.json beside the pack's, holding its
// parts whole, each spelled as a pack spells it today — a kind of record, a
// view, a door, a rule — so the App adds what binds them and no language of
// its own. A query longer than a line may sit beside it in a .sql file, named
// in the view's spec as sql_file.
type AppFile struct {
	App     string            `json:"app"`
	Version int               `json:"version"`
	Title   string            `json:"title"`
	Layer   string            `json:"layer"`
	Data    []core.ObjectType `json:"data,omitempty"`
	Uses    []string          `json:"uses,omitempty"`
	Tabs    []AppFileTab      `json:"tabs,omitempty"`
	Record  *AppFileRecord    `json:"record,omitempty"`
	Actions []Activity        `json:"actions,omitempty"`
	Rules   []Rule            `json:"rules,omitempty"`
}

// AppFileTab is a tab: a view, whole, or a screen the desk draws from its own
// code, named by desk with its title.
type AppFileTab struct {
	core.ViewDef
	Desk string `json:"desk,omitempty"`
}

// AppFileRecord is what one record of the App shows.
type AppFileRecord struct {
	Of   string       `json:"of"`
	Tabs []AppFileTab `json:"tabs"`
}

// the keys an App's file, a tab and a record may carry: a misspelt part
// ("screens") is refused, never loaded as an App without it
var (
	appFileKeys   = []string{"app", "version", "title", "layer", "data", "uses", "tabs", "record", "actions", "rules"}
	appTabKeys    = []string{"view_id", "version", "notion", "title", "domain", "function", "spec", "desk"}
	appRecordKeys = []string{"of", "tabs"}
)

func knownKeys(raw json.RawMessage, keys []string, what string) (map[string]json.RawMessage, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", what, err)
	}
	for k := range m {
		if !slices.Contains(keys, k) {
			return nil, fmt.Errorf("%s: %q is no part of an App (%s)", what, k, strings.Join(keys, ", "))
		}
	}
	return m, nil
}

// readAppFile reads one App's file and checks every key it carries.
func readAppFile(b []byte, name string) (AppFile, error) {
	what := "app " + name
	top, err := knownKeys(b, appFileKeys, what)
	if err != nil {
		return AppFile{}, err
	}
	var tabs []json.RawMessage
	if raw, ok := top["tabs"]; ok {
		if err := json.Unmarshal(raw, &tabs); err != nil {
			return AppFile{}, fmt.Errorf("%s tabs: %w", what, err)
		}
	}
	if raw, ok := top["record"]; ok {
		rec, err := knownKeys(raw, appRecordKeys, what+" record")
		if err != nil {
			return AppFile{}, err
		}
		if raw, ok := rec["tabs"]; ok {
			var rt []json.RawMessage
			if err := json.Unmarshal(raw, &rt); err != nil {
				return AppFile{}, fmt.Errorf("%s record tabs: %w", what, err)
			}
			tabs = append(tabs, rt...)
		}
	}
	for _, t := range tabs {
		if _, err := knownKeys(t, appTabKeys, what+" tab"); err != nil {
			return AppFile{}, err
		}
	}
	var f AppFile
	if err := json.Unmarshal(b, &f); err != nil {
		return AppFile{}, fmt.Errorf("%s: %w", what, err)
	}
	if f.App != name {
		return AppFile{}, fmt.Errorf("the pack lists App %s, and its file apps/%s.json says it is %q", name, name, f.App)
	}
	return f, nil
}

// unpacked is an App and the parts it holds, as the loader loads them.
type unpacked struct {
	app   core.App
	types []core.ObjectType
	views []core.ViewDef
	rules []Rule
	doors []Activity
}

// unpack turns an App's file into the App the kernel keeps and the parts it
// holds. read gives a file beside the App's (its query files).
func unpack(f AppFile, read func(name string) ([]byte, error)) (unpacked, error) {
	u := unpacked{app: core.App{ID: f.App, Version: f.Version, Title: f.Title, Layer: f.Layer, Uses: f.Uses}}
	for _, t := range f.Data {
		u.app.Data = append(u.app.Data, t.Name)
		u.types = append(u.types, t)
	}
	tabs := func(in []AppFileTab) ([]core.AppTab, error) {
		var out []core.AppTab
		for i, t := range in {
			switch {
			case t.Desk != "" && t.ID != "":
				return nil, fmt.Errorf("app %s: tab %d is both view %s and the desk's %s", f.App, i+1, t.ID, t.Desk)
			case t.Desk != "":
				if t.Notion != "" || t.Domain != "" || t.Function != "" || len(t.Spec) > 0 || t.Version != 0 {
					return nil, fmt.Errorf("app %s: tab %d is the desk's %s, and carries a view's parts too", f.App, i+1, t.Desk)
				}
				out = append(out, core.AppTab{Desk: t.Desk, Title: t.Title})
			default:
				v, err := withQueryFile(t.ViewDef, f.App, read)
				if err != nil {
					return nil, err
				}
				u.views = append(u.views, v)
				out = append(out, core.AppTab{View: v.ID})
			}
		}
		return out, nil
	}
	var err error
	if u.app.Tabs, err = tabs(f.Tabs); err != nil {
		return u, err
	}
	if f.Record != nil {
		rt, err := tabs(f.Record.Tabs)
		if err != nil {
			return u, err
		}
		u.app.Record = &core.AppRecord{Of: f.Record.Of, Tabs: rt}
	}
	for _, a := range f.Actions {
		u.app.Actions = append(u.app.Actions, a.Name)
		u.doors = append(u.doors, a)
	}
	for _, r := range f.Rules {
		u.app.Rules = append(u.app.Rules, r.RuleID)
		u.rules = append(u.rules, r)
	}
	if err := u.app.Validate(); err != nil {
		return u, err
	}
	// a rule writes what its App keeps or uses: the kinds an App touches are
	// in its definition, not found by reading its rules
	touches := append(slices.Clone(u.app.Data), u.app.Uses...)
	for _, r := range f.Rules {
		var writes []string
		if t := r.Spec.Effect.Object.Type; t != "" {
			writes = append(writes, t)
		}
		if a := r.Spec.Effect.Amend; a != nil && a.Type != "" {
			writes = append(writes, a.Type)
		}
		if r.Spec.Effect.Postings != nil {
			writes = append(writes, exec.PostingObjectType)
		}
		for _, w := range writes {
			if !slices.Contains(touches, w) {
				return u, fmt.Errorf("app %s: rule %s writes %s, which the App neither keeps nor names in its uses", f.App, r.RuleID, w)
			}
		}
	}
	return u, nil
}

// withQueryFile puts a query file's text where the view's spec names it, so
// the view unpacked is the one a pack writes inline today.
func withQueryFile(v core.ViewDef, app string, read func(string) ([]byte, error)) (core.ViewDef, error) {
	if len(v.Spec) == 0 {
		return v, nil
	}
	var spec map[string]json.RawMessage
	if err := json.Unmarshal(v.Spec, &spec); err != nil {
		return v, fmt.Errorf("app %s: view %s: %w", app, v.ID, err)
	}
	raw, ok := spec["sql_file"]
	if !ok {
		return v, nil
	}
	if _, both := spec["sql"]; both {
		return v, fmt.Errorf("app %s: view %s names its query twice, as sql and as sql_file", app, v.ID)
	}
	var name string
	if err := json.Unmarshal(raw, &name); err != nil || name == "" || path.Base(name) != name || !strings.HasSuffix(name, ".sql") {
		return v, fmt.Errorf("app %s: view %s: sql_file names a .sql file beside the App, not %s", app, v.ID, raw)
	}
	b, err := read(name)
	if err != nil {
		return v, fmt.Errorf("app %s: view %s: its sql_file %s: %w", app, v.ID, name, err)
	}
	sql, _ := json.Marshal(string(bytes.TrimRight(b, "\n")))
	delete(spec, "sql_file")
	spec["sql"] = sql
	if v.Spec, err = json.Marshal(spec); err != nil {
		return v, err
	}
	return v, nil
}

// readApps reads the Apps a pack lists and puts their parts where a pack
// without Apps keeps them, in the order the Apps are listed — so everything
// after this loads a pack of Apps as it loads any pack. Each part is in one
// App; what an App uses is kept by another of the pack's Apps or left for the
// loader to find on the record.
func readApps(m *Manifest, read func(name string) ([]byte, error)) ([]unpacked, error) {
	if len(m.Apps) == 0 {
		return nil, nil
	}
	if read == nil {
		return nil, fmt.Errorf("pack %s is written as Apps: load it from its file, its apps/ folder beside it", m.Pack)
	}
	if len(m.ObjectTypes)+len(m.ViewDefs)+len(m.Rules)+len(m.Activities) > 0 {
		return nil, fmt.Errorf("pack %s is written as Apps and keeps its types, views, rules and doors in its Apps, not beside them", m.Pack)
	}
	owner := map[[2]string]string{}
	own := func(kind, name, app string) error {
		k := [2]string{kind, name}
		if prev, ok := owner[k]; ok {
			return fmt.Errorf("pack %s: %s %s is in App %s and in App %s — a part has one App", m.Pack, kind, name, prev, app)
		}
		owner[k] = app
		return nil
	}
	var out []unpacked
	seen := map[string]bool{}
	for _, name := range m.Apps {
		if seen[name] {
			return nil, fmt.Errorf("pack %s lists App %s twice", m.Pack, name)
		}
		seen[name] = true
		b, err := read("apps/" + name + ".json")
		if err != nil {
			return nil, fmt.Errorf("pack %s lists App %s: %w", m.Pack, name, err)
		}
		f, err := readAppFile(b, name)
		if err != nil {
			return nil, err
		}
		u, err := unpack(f, func(file string) ([]byte, error) { return read("apps/" + file) })
		if err != nil {
			return nil, err
		}
		for _, t := range u.types {
			if err := own("kind of record", t.Name, name); err != nil {
				return nil, err
			}
		}
		for _, v := range u.views {
			if err := own("view", v.ID, name); err != nil {
				return nil, err
			}
		}
		for _, r := range u.rules {
			if err := own("rule", r.RuleID, name); err != nil {
				return nil, err
			}
		}
		for _, a := range u.doors {
			if err := own("door", a.Name, name); err != nil {
				return nil, err
			}
		}
		m.ObjectTypes = append(m.ObjectTypes, u.types...)
		m.ViewDefs = append(m.ViewDefs, u.views...)
		m.Rules = append(m.Rules, u.rules...)
		m.Activities = append(m.Activities, u.doors...)
		out = append(out, u)
	}
	return out, nil
}
