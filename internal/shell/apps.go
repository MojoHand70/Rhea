package shell

import (
	"net/http"

	"rhea/internal/core"
	"rhea/internal/store"
)

// appServed is an App as a person is served it (KK 2026-10-26, APPS.md "The
// App definition"): the tabs they may see, each named — a view by its own
// title, a screen of the desk by the App's — and where the App sits in the
// drawer, read from its first tab: the domain and function its view is filed
// under.
type appServed struct {
	core.App
	Domain   string `json:"domain"`
	Function string `json:"function"`
}

// handleApps serves the Apps of the company acted for: each App of a pack the
// company has, as the person may see it. An App opens on its first tab, so a
// person who may not see that tab is not served the App; within it, a view's
// tab they may not see is left out, and a screen of the desk stays — it draws
// only what the kernel serves it. An App with no tab opens on the list of its
// first kind of record, kept or else used, and is served where that list may
// be seen; one whose tabs are all the desk's opens on its first, and sits where
// that list would.
func (s *Server) handleApps(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	apps, err := s.Store.LatestApps(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	offer, err := s.Exec.Offered(ctx, store.CompanyOf(ctx))
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	vds, err := s.effectiveViewDefs(ctx) // what the company is offered, derived lists included
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	byID := map[string]*core.ViewDef{}
	listOf := map[string]*core.ViewDef{} // a kind of record → its list, a written one before the derived
	for i := range vds {
		v := &vds[i]
		byID[v.ID] = v
		if typ := viewType(v); v.Notion == "list" && typ != "" {
			if have, ok := listOf[typ]; !ok || have.ID == core.DerivedListID(typ) {
				listOf[typ] = v
			}
		}
	}
	seen := func(id string) *core.ViewDef {
		if v := byID[id]; v != nil && s.mayView(ctx, v) {
			return v
		}
		return nil
	}
	visible := func(ts []core.AppTab) []core.AppTab {
		var keep []core.AppTab
		for _, t := range ts {
			if t.Desk != "" {
				keep = append(keep, t)
			} else if v := seen(t.View); v != nil {
				keep = append(keep, core.AppTab{View: v.ID, Title: v.Title})
			}
		}
		return keep
	}
	out := []appServed{}
	for _, a := range apps {
		if !offer.App(a.ID) {
			continue
		}
		first := ""
		for _, t := range a.Tabs {
			if t.View != "" {
				first = t.View
				break
			}
		}
		sv := appServed{App: a}
		sv.Tabs = visible(a.Tabs)
		var kindList *core.ViewDef // the list of its first kind of record, kept or else used
		if kinds := append(append([]string(nil), a.Data...), a.Uses...); len(kinds) > 0 {
			kindList = listOf[kinds[0]]
		}
		switch {
		case first != "":
			v := seen(first)
			if v == nil {
				continue
			}
			sv.Domain, sv.Function = v.Domain, v.Function
		case len(a.Tabs) == 0 && kindList != nil:
			if !s.mayView(ctx, kindList) {
				continue
			}
			sv.Domain, sv.Function = kindList.Domain, kindList.Function
			sv.Tabs = []core.AppTab{{View: kindList.ID, Title: kindList.Title}}
		case kindList != nil:
			sv.Domain, sv.Function = kindList.Domain, kindList.Function
		}
		if a.Record != nil {
			rec := core.AppRecord{Of: a.Record.Of, Tabs: visible(a.Record.Tabs)}
			sv.Record = nil
			if len(rec.Tabs) > 0 {
				sv.Record = &rec
			}
		}
		out = append(out, sv)
	}
	writeJSON(w, 200, out)
}
