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
	all     bool
	h       holding
	seen    map[string]bool // packs the company has an install of, open or ended
	running map[string]bool // ... of which one stands on the day
}

// Offered is the offer for a company on a day. The day decides only which
// doors are open; a caller that serves no door may pass none.
func Offered(ctx context.Context, st *store.Store, company, day string) (Offer, error) {
	if company == "" {
		return Offer{all: true}, nil
	}
	h, err := readHolding(ctx, st)
	if err != nil {
		return Offer{}, err
	}
	if !h.perCo {
		return Offer{all: true}, nil // nothing is installed per company: nothing to hold back
	}
	o := Offer{h: h, seen: map[string]bool{}, running: map[string]bool{}}
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

func (o Offer) seenAt(p string) bool    { return o.seen[p] }
func (o Offer) runningAt(p string) bool { return o.running[p] }

// View says whether a screen is offered.
func (o Offer) View(id string) bool { return o.all || o.h.allows(core.KindViewDef, id, o.seenAt) }

// App says whether an App is offered: the company has an install of a pack
// holding it, as for a screen.
func (o Offer) App(id string) bool { return o.all || o.h.allows(core.KindApp, id, o.seenAt) }

// Type says whether a type is offered.
func (o Offer) Type(name string) bool {
	return o.all || o.h.allows(core.KindObjectType, name, o.seenAt)
}

// Rule says whether a rule is shown — the company has an install of a pack holding it.
func (o Offer) Rule(r core.Rule) bool {
	return o.all || r.Company != "" || o.h.allows(core.KindRule, r.ID, o.seenAt)
}

// Runs says whether a rule runs for the company on the day — a pack holding it stands.
func (o Offer) Runs(r core.Rule) bool {
	return o.all || r.Company != "" || o.h.allows(core.KindRule, r.ID, o.runningAt)
}

// Door says whether a door is open to the company on the day.
func (o Offer) Door(a core.Activity) bool {
	return o.all || o.h.allows(core.KindActivity, a.Name, o.runningAt)
}

// DoorPacks names the packs holding a door, for a refusal to say.
func (o Offer) DoorPacks(a core.Activity) []string {
	return o.h.of[[2]string{core.KindActivity, a.Name}]
}

// PackOf names the pack a definition's row was written by ("pack:<name>").
func PackOf(createdBy string) (string, bool) {
	return strings.CutPrefix(createdBy, "pack:")
}
