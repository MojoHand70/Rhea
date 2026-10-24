package exec

import (
	"context"

	"rhea/internal/core"
	"rhea/internal/store"
)

// holding says which packs each definition belongs to (DIRECTION 2026-10-24,
// step 3): the packs whose newest approved version declares it — not
// whoever wrote its row, as a rule carried unchanged into a new pack still
// names its old one. A pack another approved pack succeeds holds only what
// no successor carries.
type holding struct {
	of      map[[2]string][]string      // (kind, name) → the packs that hold it
	headers map[string]*core.PackHeader // what each approved pack says it is
	perCo   bool                        // some pack is installed per company
}

func readHolding(ctx context.Context, st *store.Store) (holding, error) {
	headers, err := st.ApprovedPacks(ctx)
	if err != nil {
		return holding{}, err
	}
	h := holding{headers: headers, of: map[[2]string][]string{}}
	for _, hd := range headers {
		if hd != nil && hd.Install == core.InstallCompany {
			h.perCo = true
		}
	}
	if !h.perCo {
		return h, nil // every pack runs for every company: nothing to read
	}
	decl, err := st.PackDeclarations(ctx)
	if err != nil {
		return holding{}, err
	}
	successors := map[string][]string{} // a succeeded pack → the approved packs succeeding it
	for p, hd := range headers {
		if hd != nil {
			for _, q := range hd.Succeeds {
				successors[q] = append(successors[q], p)
			}
		}
	}
	held := map[string]map[[2]string]bool{}
	for p, ms := range decl {
		held[p] = map[[2]string]bool{}
		for _, m := range ms {
			held[p][[2]string{m.Kind, m.Name}] = true
		}
	}
	for p, ms := range decl {
	member:
		for _, m := range ms {
			k := [2]string{m.Kind, m.Name}
			for _, q := range successors[p] {
				if held[q][k] {
					continue member // carried by a successor: the successor's
				}
			}
			h.of[k] = append(h.of[k], p)
		}
	}
	return h, nil
}

// everyCompany says whether a pack runs for every company: approved for the
// whole installation, or before packs said how they are installed.
func (h holding) everyCompany(pack string) bool {
	hd := h.headers[pack]
	return hd == nil || hd.Install != core.InstallCompany
}

// allows says whether a definition is there for a company: it belongs to no
// pack, or to one that runs for every company, or to one the test passes.
func (h holding) allows(kind, name string, installed func(pack string) bool) bool {
	packs := h.of[[2]string{kind, name}]
	if len(packs) == 0 {
		return true
	}
	for _, p := range packs {
		if h.everyCompany(p) || installed(p) {
			return true
		}
	}
	return false
}

// kindRank orders a company's packs as their words are laid over each other.
var kindRank = map[string]int{"core": 0, "market": 1, "regime": 2, "module": 3, "business": 4}

// ViewRanks says, for the screens a pack holds, where the pack stands in
// that order (DIRECTION 2026-10-24, step 3): its kind's place, a pack that
// does not say its kind counted as a market's; a screen of no pack is last.
func ViewRanks(ctx context.Context, st *store.Store) (func(viewID string) int, error) {
	headers, err := st.ApprovedPacks(ctx)
	if err != nil {
		return nil, err
	}
	decl, err := st.PackDeclarations(ctx)
	if err != nil {
		return nil, err
	}
	rank := map[string]int{}
	for p, ms := range decl {
		r := kindRank["market"]
		if h := headers[p]; h != nil {
			if k, ok := kindRank[h.Kind]; ok {
				r = k
			}
		}
		for _, m := range ms {
			if m.Kind != core.KindViewDef {
				continue
			}
			if have, ok := rank[m.Name]; !ok || r < have {
				rank[m.Name] = r
			}
		}
	}
	return func(id string) int {
		if r, ok := rank[id]; ok {
			return r
		}
		return len(kindRank)
	}, nil
}
