package shell

import (
	"net/http"
	"slices"
	"strings"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store"
)

// catalogEntry is one pack of the installation's catalog as Ustawienia →
// Pakiety shows it to the company acted for (DIRECTION 2026-10-24, step 2).
type catalogEntry struct {
	Pack        string   `json:"pack"`
	Version     int      `json:"version"`
	Bundle      string   `json:"bundle"`
	Kind        string   `json:"kind"`
	Market      string   `json:"market,omitempty"`
	Install     string   `json:"install"`
	Needs       []string `json:"needs"`
	Description string   `json:"description"`
	// EveryCompany: the pack runs for every company of the installation —
	// nothing to install, nothing to switch off.
	EveryCompany bool `json:"every_company"`
	// Brings counts what the pack's approved version declares.
	Brings struct {
		Types int `json:"types"`
		Views int `json:"views"`
		Rules int `json:"rules"`
		Doors int `json:"doors"`
		Apps  int `json:"apps"`
	} `json:"brings"`
	// Installed is the company's install of the pack — the standing one,
	// else the last it switched off; nil when it has none.
	Installed *catalogInstall `json:"installed"`
}

type catalogInstall struct {
	Object string `json:"object"`
	From   string `json:"from"`
	To     string `json:"to,omitempty"`
	Status string `json:"status"`
}

var kindOrder = map[string]int{"core": 0, "market": 1, "regime": 2, "module": 3, "business": 4}

// handlePacks is the catalog: every approved pack with what it says it is
// and what it brings, and, for the company acted for, whether it is
// installed there.
func (s *Server) handlePacks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	versions, err := s.Store.ApprovedPackVersions(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	members, err := s.Store.PackDeclarations(ctx) // what each version declares, carried definitions included
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	installs := map[string]*catalogInstall{}
	if company := store.CompanyOf(ctx); company != "" {
		objs, err := s.Store.ObjectsByType(ctx, "installed_pack")
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		for _, o := range objs {
			st := func(k string) string { v, _ := o.State[k].(string); return v }
			in := &catalogInstall{Object: o.ID, From: st("from"), To: st("to"), Status: st("status")}
			have := installs[st("pack")]
			if have == nil || have.Status != "installed" && (in.Status == "installed" || in.From > have.From) {
				installs[st("pack")] = in
			}
		}
	}
	out := make([]catalogEntry, 0, len(versions))
	for _, b := range versions {
		name, _ := exec.PackOf(b.CreatedBy)
		e := catalogEntry{Pack: name, Bundle: b.ID, Install: core.InstallInstallation, Needs: []string{},
			Description: b.Description}
		if h := b.Pack; h != nil {
			e.Version, e.Kind, e.Market, e.Install, e.Description = h.Version, h.Kind, h.Market, h.Install, h.Description
			if h.Needs != nil {
				e.Needs = h.Needs
			}
		}
		e.EveryCompany = e.Install != core.InstallCompany
		seen := map[[2]string]bool{}
		for _, m := range members[name] {
			k := [2]string{m.Kind, m.Name}
			if seen[k] {
				continue
			}
			seen[k] = true
			switch m.Kind {
			case core.KindObjectType:
				e.Brings.Types++
			case core.KindViewDef:
				e.Brings.Views++
			case core.KindRule:
				e.Brings.Rules++
			case core.KindActivity:
				e.Brings.Doors++
			case core.KindApp:
				e.Brings.Apps++
			}
		}
		if !e.EveryCompany {
			e.Installed = installs[name]
		}
		out = append(out, e)
	}
	slices.SortFunc(out, func(a, b catalogEntry) int {
		if d := kindOrder[a.Kind] - kindOrder[b.Kind]; d != 0 {
			return d
		}
		return strings.Compare(a.Pack, b.Pack)
	})
	writeJSON(w, 200, out)
}
