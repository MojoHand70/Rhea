package exec

import (
	"context"
	"strings"

	"rhea/internal/core"
	"rhea/internal/store"
)

// Offer is what a company is served of the installation's definitions
// (DIRECTION 2026-10-24, step 2). A pack installed per company is offered
// to the companies that installed it: its screens, types and rules once the
// company has an install of it — still after it is switched off, as what it
// booked stays readable — and its doors only while it is installed. Every
// other definition — the core's, a pack's for the whole installation, a
// company's own — is offered as before; acting for no company (the
// operator's own tools) is offered everything.
type Offer struct {
	all        bool
	perCompany map[string]bool // the packs installed per company
	seen       map[string]bool // ... of which the company has an install, open or ended
	running    map[string]bool // ... of which the company has one standing on the day
	held       map[[2]string][]string
}

// Offered is the offer for a company on a day. The day decides only which
// doors are open; a caller that serves no door may pass none.
func Offered(ctx context.Context, st *store.Store, company, day string) (Offer, error) {
	if company == "" {
		return Offer{all: true}, nil
	}
	packs, err := st.ApprovedPacks(ctx)
	if err != nil {
		return Offer{}, err
	}
	o := Offer{perCompany: map[string]bool{}, seen: map[string]bool{}, running: map[string]bool{}, held: map[[2]string][]string{}}
	for p, h := range packs {
		if h != nil && h.Install == core.InstallCompany {
			o.perCompany[p] = true
		}
	}
	if len(o.perCompany) == 0 {
		o.all = true // nothing is installed per company: nothing to hold back
		return o, nil
	}
	members, err := st.PackMembers(ctx)
	if err != nil {
		return Offer{}, err
	}
	other := map[[2]string]bool{} // held by a pack every company has
	for p, ms := range members {
		for _, m := range ms {
			k := [2]string{m.Kind, m.Name}
			if o.perCompany[p] {
				o.held[k] = append(o.held[k], p)
			} else {
				other[k] = true
			}
		}
	}
	for k := range other {
		delete(o.held, k)
	}
	objs, err := st.ObjectsByType(store.WithCompany(ctx, company), "installed_pack")
	if err != nil {
		return Offer{}, err
	}
	for _, ob := range objs {
		p, _ := ob.State["pack"].(string)
		from, _ := ob.State["from"].(string)
		to, _ := ob.State["to"].(string)
		o.seen[p] = true
		if day != "" && from != "" && from <= day && (to == "" || day <= to) {
			o.running[p] = true
		}
	}
	return o, nil
}

// Offered is the offer for a company today.
func (x *Executor) Offered(ctx context.Context, company string) (Offer, error) {
	return Offered(ctx, x.Store, company, x.today())
}

func (o Offer) def(kind, name string) bool {
	if o.all {
		return true
	}
	packs, ok := o.held[[2]string{kind, name}]
	if !ok {
		return true
	}
	for _, p := range packs {
		if o.seen[p] {
			return true
		}
	}
	return false
}

// View says whether a screen is offered.
func (o Offer) View(id string) bool { return o.def(core.KindViewDef, id) }

// Type says whether a type is offered.
func (o Offer) Type(name string) bool { return o.def(core.KindObjectType, name) }

// Rule says whether a rule is shown — the company has an install of its pack.
func (o Offer) Rule(r core.Rule) bool {
	p, ok := PackOf(r.CreatedBy)
	return o.all || !ok || r.Company != "" || !o.perCompany[p] || o.seen[p]
}

// Runs says whether a rule runs for the company on the day — its pack stands.
func (o Offer) Runs(r core.Rule) bool {
	p, ok := PackOf(r.CreatedBy)
	return o.all || !ok || r.Company != "" || !o.perCompany[p] || o.running[p]
}

// Door says whether a door is open to the company on the day.
func (o Offer) Door(a core.Activity) bool {
	p, ok := PackOf(a.CreatedBy)
	return o.all || !ok || !o.perCompany[p] || o.running[p]
}

// PackOf names the pack a definition came from ("pack:<name>").
func PackOf(createdBy string) (string, bool) {
	return strings.CutPrefix(createdBy, "pack:")
}
