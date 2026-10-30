package core

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// App is a Rhea App as the kernel keeps it (DIRECTION 2026-10-23, "Rhea Apps
// and business packs"; the definition KK agreed 2026-10-26, APPS.md "The App
// definition"): a small program a pack brings, one definition holding its
// parts — the kinds of record it keeps, its tabs, what one of its records
// shows, its doors and its rules. The App's file holds the parts whole; the
// loader unpacks them into the definitions the kernel already runs, and the
// App is kept as what binds them, each part by name: a part's versions are
// its own rows, so a new version of a rule or a view moves no App. Variables,
// workflow, triggers, autonomy and connectors join as their steps land.
type App struct {
	ID      string `json:"app"`
	Version int    `json:"version"`
	Title   string `json:"title"` // in the pack's language; another pack's words may rename it
	Layer   string `json:"layer"`
	// Data: the kinds of record it keeps; Uses: other Apps' kinds it reads
	// or its rules write (a booking rule writes the journal's posting).
	Data []string `json:"data,omitempty"`
	Uses []string `json:"uses,omitempty"`
	// Tabs, in order: the first opens it. An App with none opens on the
	// list of its first kind of record, kept or else used.
	Tabs []AppTab `json:"tabs,omitempty"`
	// Record: what one of its records shows, in a desk tab of its own — a
	// kind it keeps, or one it uses (a market's App over the core's kind).
	Record  *AppRecord `json:"record,omitempty"`
	Actions []string   `json:"actions,omitempty"` // its doors
	Rules   []string   `json:"rules,omitempty"`
	// Place: where the App sits on a face (KK 2026-10-30: every App says
	// its location, the activity and the folder) — the Activity Bar's
	// activity and the Nav Bar's section within it, by the ids a market's
	// words arrange. A market's words may still file it elsewhere; with
	// no place and no arrangement naming it, a face files it under the rest.
	Place *AppPlace `json:"place,omitempty"`
	// Shows: the faces it is visible on, "desk" and "phone" (KK 2026-10-30);
	// none named means both.
	Shows []string `json:"shows,omitempty"`
	// Company: a company's own App — a kept report — seen only where the
	// company is (invariant 9), as its views and rules are; empty is Rhea's.
	Company string `json:"company,omitempty"`
}

// AppPlace is where an App sits: its activity, and the section of that
// activity's Nav Bar (none: the activity's own items).
type AppPlace struct {
	Activity string `json:"activity"`
	Section  string `json:"section,omitempty"`
}

// AppFaces are the faces an App may show on.
var AppFaces = []string{"desk", "phone"}

// ShownOn says whether the App is visible on a face: every face unless it
// names the ones it shows on.
func (a App) ShownOn(face string) bool {
	return len(a.Shows) == 0 || slices.Contains(a.Shows, face)
}

// AppTab is one tab: a view, or a screen the desk still draws from its own
// code ("flow", "document-sheet"), named until it has a definition. Tabs go
// one level deep (KK 2026-10-26): an App's inside its desk tab, a record's
// inside its own, never a tab inside a tab.
type AppTab struct {
	View  string `json:"view,omitempty"`
	Desk  string `json:"desk,omitempty"`
	Title string `json:"title,omitempty"` // a desk screen's; a view carries its own
}

// AppRecord is what one record of the App shows: tabs of its own.
type AppRecord struct {
	Of   string   `json:"of"`
	Tabs []AppTab `json:"tabs"`
}

// AppLayers: what never changes, what the law imposes, what the business
// chooses, and the platform that carries them (APPS.md).
var AppLayers = []string{"core", "law", "business", "platform"}

// an App's id is a name; a company's own App carries its owner after "@"
// (OwnedID), as a company's rule or view does
var (
	appName    = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*(@[a-z0-9]+(-[a-z0-9]+)*)?$`)
	appSegment = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
)

// Validate checks an App's shape. Whether its parts exist, and whether each
// is in no other App, is the loader's to judge with the pack in hand.
func (a App) Validate() error {
	if !appName.MatchString(a.ID) {
		return fmt.Errorf("app %q: its id is a name — lowercase letters, digits and hyphens", a.ID)
	}
	if a.Version < 1 {
		return fmt.Errorf("app %s: needs a version", a.ID)
	}
	if strings.TrimSpace(a.Title) == "" {
		return fmt.Errorf("app %s: needs a title", a.ID)
	}
	if !slices.Contains(AppLayers, a.Layer) {
		return fmt.Errorf("app %s: layer %q is not one of %v", a.ID, a.Layer, AppLayers)
	}
	if len(a.Tabs) == 0 && a.Record == nil && len(a.Data) == 0 && len(a.Uses) == 0 {
		return fmt.Errorf("app %s: shows nothing — give it a tab or a kind of record", a.ID)
	}
	if p := a.Place; p != nil {
		if !appSegment.MatchString(p.Activity) {
			return fmt.Errorf("app %s: its place names an activity by its id — lowercase letters, digits and hyphens — not %q", a.ID, p.Activity)
		}
		if p.Section != "" && !appSegment.MatchString(p.Section) {
			return fmt.Errorf("app %s: its place names a section by its id — lowercase letters, digits and hyphens — not %q", a.ID, p.Section)
		}
	}
	for i, f := range a.Shows {
		if !slices.Contains(AppFaces, f) {
			return fmt.Errorf("app %s: it shows on %q, which is not one of %v", a.ID, f, AppFaces)
		}
		if slices.Contains(a.Shows[:i], f) {
			return fmt.Errorf("app %s: shows on %s twice", a.ID, f)
		}
	}
	for what, names := range map[string][]string{"kind of record": append(slices.Clone(a.Data), a.Uses...), "door": a.Actions, "rule": a.Rules} {
		seen := map[string]bool{}
		for _, n := range names {
			if n == "" {
				return fmt.Errorf("app %s: a %s with no name", a.ID, what)
			}
			if seen[n] {
				return fmt.Errorf("app %s: %s %s named twice (a kind it keeps is not also one it uses)", a.ID, what, n)
			}
			seen[n] = true
		}
	}
	views := map[string]bool{}
	tabs := func(where string, ts []AppTab) error {
		for i, t := range ts {
			switch {
			case t.View != "" && t.Desk != "":
				return fmt.Errorf("app %s: %s tab %d is both view %s and the desk's %s", a.ID, where, i+1, t.View, t.Desk)
			case t.View != "":
				if views[t.View] {
					return fmt.Errorf("app %s: view %s is in two tabs", a.ID, t.View)
				}
				views[t.View] = true
			case t.Desk != "":
				if !appSegment.MatchString(t.Desk) {
					return fmt.Errorf("app %s: %s tab %d: the desk's screen %q is a name — lowercase letters, digits and hyphens", a.ID, where, i+1, t.Desk)
				}
				if strings.TrimSpace(t.Title) == "" {
					return fmt.Errorf("app %s: %s tab %d (the desk's %s) needs its title", a.ID, where, i+1, t.Desk)
				}
			default:
				return fmt.Errorf("app %s: %s tab %d names no view and no screen of the desk", a.ID, where, i+1)
			}
		}
		return nil
	}
	if err := tabs("its", a.Tabs); err != nil {
		return err
	}
	if r := a.Record; r != nil {
		if !slices.Contains(a.Data, r.Of) && !slices.Contains(a.Uses, r.Of) {
			return fmt.Errorf("app %s: its record is of %q, which it neither keeps nor uses", a.ID, r.Of)
		}
		if len(r.Tabs) == 0 {
			return fmt.Errorf("app %s: its record shows no tab", a.ID)
		}
		if err := tabs("the record's", r.Tabs); err != nil {
			return err
		}
	}
	return nil
}

// Views names the views an App holds, its tabs' first, then its record's.
func (a App) Views() []string {
	var out []string
	for _, t := range a.Tabs {
		if t.View != "" {
			out = append(out, t.View)
		}
	}
	if a.Record != nil {
		for _, t := range a.Record.Tabs {
			if t.View != "" {
				out = append(out, t.View)
			}
		}
	}
	return out
}
