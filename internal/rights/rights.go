// Package rights answers one question from the log: may this person, in
// this company, do this kind of thing by this name (DIRECTION 2026-10-19,
// "People and rights"). The face says who; the kernel says whether.
//
// Everything it reads is data a pack declared and people granted through
// doors — persons, roles, memberships, mandates, substitutions — so replay
// rebuilds who-may-do-what identically and every right opens to the grant
// that gives it. The kernel itself computes with its own eyes (unscoped):
// a person's grants decide their scope, so they cannot be read under it.
package rights

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"rhea/internal/core"
	"rhea/internal/store"
)

// A Grant is one reason a person holds rights in a company: a membership
// with a role, an office membership reaching a client through a mandate, or
// a substitution for someone who holds either.
type Grant struct {
	Company string       `json:"company"` // where the rights hold
	Rights  []core.Right `json:"rights"`  // the role's rights
	// Scope narrows a mandate's reach: the rights the client granted the
	// office. Nil for a membership — the role's rights whole.
	Scope  []core.Right `json:"scope,omitempty"`
	Via    string       `json:"via"`    // "membership" | "mandate" | "substitution"
	Object string       `json:"object"` // the membership, mandate or substitution object
	// Through is the grant a mandate or a substitution carries: the office
	// membership behind a mandate, the substituted person's grant behind a
	// substitution.
	Through   string `json:"through,omitempty"`
	Role      string `json:"role"`
	Group     string `json:"group,omitempty"`
	GrantedBy string `json:"granted_by"`
	From      string `json:"from"`
	To        string `json:"to,omitempty"`
	// Reason is the grant's own words — a day of support, a guest's purpose.
	Reason string `json:"reason,omitempty"`
}

// Covers reports whether the grant gives the thing: the role's rights, and
// for a mandate the client's scope as well.
func (g Grant) Covers(kind, name, domain string) bool {
	if !core.CoveredBy(g.Rights, kind, name, domain) {
		return false
	}
	return g.Scope == nil || core.CoveredBy(g.Scope, kind, name, domain)
}

// A Principal is a signed-in person with every grant they hold today.
type Principal struct {
	Handle   string  `json:"handle"`
	PersonID string  `json:"person_id"`
	Name     string  `json:"name"`
	Grants   []Grant `json:"grants"`
}

// Companies is the set the person may see: every company a grant names,
// sorted. Empty for a person nobody has let in anywhere.
func (p Principal) Companies() []string {
	set := map[string]bool{}
	for _, g := range p.Grants {
		set[g.Company] = true
	}
	out := make([]string, 0, len(set))
	for c := range set {
		out = append(out, c)
	}
	sort.Strings(out)
	return out
}

// May reports whether the person may do the thing in the company, and the
// first grant that says so.
func (p Principal) May(company, kind, name, domain string) (Grant, bool) {
	for _, g := range p.Grants {
		if g.Company == company && g.Covers(kind, name, domain) {
			return g, true
		}
	}
	return Grant{}, false
}

// MayAnywhere reports whether the person may do the thing in any company —
// the question an installation-wide door asks.
func (p Principal) MayAnywhere(kind, name, domain string) (Grant, bool) {
	for _, g := range p.Grants {
		if g.Covers(kind, name, domain) {
			return g, true
		}
	}
	return Grant{}, false
}

// MayEverywhere reports whether the person may do the thing in every one of
// the companies — what an office desk open on all its clients at once asks.
func (p Principal) MayEverywhere(companies []string, kind, name, domain string) bool {
	if len(companies) == 0 {
		return false
	}
	for _, c := range companies {
		if _, ok := p.May(c, kind, name, domain); !ok {
			return false
		}
	}
	return true
}

type principalKey struct{}

// WithPrincipal puts the signed-in person on a context; the executor reads
// it at every door, the shell at every read.
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, &p)
}

// PrincipalOf is the signed-in person a context carries, or nil for the
// kernel's own work and the owner's unscoped shell.
func PrincipalOf(ctx context.Context) *Principal {
	p, _ := ctx.Value(principalKey{}).(*Principal)
	return p
}

// WithoutPrincipal is the kernel speaking for itself while a person's
// request runs — the host's words are the agent's, under the person's
// scope. Only kernel code sets the person aside; nothing a caller sends
// does.
func WithoutPrincipal(ctx context.Context) context.Context {
	return context.WithValue(ctx, principalKey{}, (*Principal)(nil))
}

// ErrRefused marks a refusal the kernel has recorded: the caller answers
// 403 and says why, and the fact is already in the log.
type ErrRefused struct {
	Person, Door, Company, Reason string
}

func (e *ErrRefused) Error() string {
	where := ""
	if e.Company != "" {
		where = " in " + e.Company
	}
	what := ""
	if e.Door != "" {
		what = " at " + e.Door
	}
	return fmt.Sprintf("refused: %s%s%s — %s", e.Person, what, where, e.Reason)
}

// NoCompany is the scope of a person nobody has let in anywhere: a key no
// company carries, so row-level security shows the installation's rows and
// no company's. An empty scope would be the whole log.
const NoCompany = "-"

// world is the log's people data read once per resolution.
type world struct {
	persons       map[string]core.Object // by object id
	byHandle      map[string]string      // handle → object id
	roles         map[string][]core.Object
	memberships   []core.Object
	mandates      []core.Object
	substitutions []core.Object
	today         string
}

func load(ctx context.Context, s *store.Store, today string) (*world, error) {
	ctx = store.WithOwner(ctx) // the kernel's own eyes
	w := &world{persons: map[string]core.Object{}, byHandle: map[string]string{}, roles: map[string][]core.Object{}, today: today}
	persons, err := s.ObjectsByType(ctx, "person")
	if err != nil {
		return nil, err
	}
	for _, p := range persons {
		w.persons[p.ID] = p
		// a handle is the first person's: persons come in log order, and a
		// later one registered under it never shadows them (review
		// 2026-10-21, A3)
		if h, _ := p.State["handle"].(string); h != "" {
			if _, taken := w.byHandle[h]; !taken {
				w.byHandle[h] = p.ID
			}
		}
	}
	roles, err := s.ObjectsByType(ctx, "role")
	if err != nil {
		return nil, err
	}
	for _, r := range roles {
		name, _ := r.State["name"].(string)
		w.roles[name] = append(w.roles[name], r)
	}
	if w.memberships, err = s.ObjectsByType(ctx, "membership"); err != nil {
		return nil, err
	}
	if w.mandates, err = s.ObjectsByType(ctx, "mandate"); err != nil {
		return nil, err
	}
	if w.substitutions, err = s.ObjectsByType(ctx, "substitution"); err != nil {
		return nil, err
	}
	return w, nil
}

func str(o core.Object, f string) string { v, _ := o.State[f].(string); return v }

// current reports whether a dated grant holds on the books' day: begun and
// not past its last day. The dates decide, not the status — an ended
// membership still answers who could do what on the days it held, and to
// revoke someone now the owner ends it with yesterday as its last day.
func (w *world) current(o core.Object) bool {
	if from := str(o, "from"); from == "" || from > w.today {
		return false
	}
	if to := str(o, "to"); to != "" && to < w.today {
		return false
	}
	return true
}

// roleRights finds a role by name for a company: the company's own role of
// that name first, then the installation's (the pack's).
func (w *world) roleRights(company, name string) ([]core.Right, bool) {
	var installation *core.Object
	for i := range w.roles[name] {
		r := &w.roles[name][i]
		if r.Company == company && company != "" {
			rights, err := core.ParseRights(str(*r, "rights"))
			return rights, err == nil
		}
		if r.Company == "" {
			installation = r
		}
	}
	if installation == nil {
		return nil, false
	}
	rights, err := core.ParseRights(str(*installation, "rights"))
	return rights, err == nil
}

// direct is a person's grants from memberships and, through them, mandates
// — without substitutions, so a substitution never chains.
func (w *world) direct(personID string) []Grant {
	var grants []Grant
	for _, m := range w.memberships {
		if str(m, "person") != personID || !w.current(m) {
			continue
		}
		company := str(m, "company")
		rights, ok := w.roleRights(company, str(m, "role"))
		if !ok {
			continue // a role nobody declared grants nothing
		}
		grants = append(grants, Grant{Company: company, Rights: rights, Via: "membership", Object: m.ID,
			Role: str(m, "role"), Group: str(m, "group"), GrantedBy: str(m, "granted_by"),
			From: str(m, "from"), To: str(m, "to"), Reason: str(m, "reason")})
	}
	// An office's person reaches a client through the mandate: the office
	// role's rights, narrowed to what the client granted.
	var reached []Grant
	for _, md := range w.mandates {
		if !w.current(md) {
			continue
		}
		scope, err := core.ParseRights(str(md, "scope"))
		if err != nil {
			continue
		}
		for _, g := range grants {
			if g.Company != str(md, "office") {
				continue
			}
			reached = append(reached, Grant{Company: str(md, "client"), Rights: g.Rights, Scope: scope, Via: "mandate",
				Object: md.ID, Through: g.Object, Role: g.Role, Group: g.Group, GrantedBy: str(md, "granted_by"),
				From: str(md, "from"), To: str(md, "to")})
		}
	}
	return append(grants, reached...)
}

func (w *world) resolve(personID string) []Grant {
	grants := w.direct(personID)
	// A deputy holds the substituted person's own grants, in the named
	// company or in all of theirs, while the substitution lasts.
	for _, sub := range w.substitutions {
		if str(sub, "deputy") != personID || !w.current(sub) {
			continue
		}
		only := str(sub, "company")
		for _, g := range w.direct(str(sub, "person")) {
			if only != "" && g.Company != only {
				continue
			}
			grants = append(grants, Grant{Company: g.Company, Rights: g.Rights, Scope: g.Scope, Via: "substitution",
				Object: sub.ID, Through: g.Object, Role: g.Role, Group: g.Group, GrantedBy: str(sub, "granted_by"),
				From: str(sub, "from"), To: str(sub, "to")})
		}
	}
	return grants
}

// Resolve computes a person's grants on the books' day from the log's
// objects. A handle no person carries, or a person who has left, is not a
// principal.
func Resolve(ctx context.Context, s *store.Store, handle, today string) (Principal, error) {
	// The machine's names are no person's: a person registered as "kernel"
	// or "agent:…" would sign facts the log reads as the machine's.
	if core.ActorKind(handle) != core.ActorHuman {
		return Principal{}, fmt.Errorf("%q is a name the machine signs with, never a person's", handle)
	}
	w, err := load(ctx, s, today)
	if err != nil {
		return Principal{}, err
	}
	id, ok := w.byHandle[handle]
	if !ok {
		return Principal{}, fmt.Errorf("no person signs as %q", handle)
	}
	p := w.persons[id]
	if str(p, "status") == "left" {
		return Principal{}, fmt.Errorf("%s has left", handle)
	}
	return Principal{Handle: handle, PersonID: id, Name: str(p, "name"), Grants: w.resolve(id)}, nil
}

// Claimed reports whether anyone was ever let into the company: a
// membership or mandate names it, whatever its dates. A company whose
// grants have lapsed, or not yet begun, is still someone's — counting only
// today's grants made it anyone's to claim (review 2026-10-21, A4). An
// unclaimed company is open to its first owner or office, and only to
// that: the claim is the first grant written into it.
func Claimed(ctx context.Context, s *store.Store, company, today string) (bool, error) {
	w, err := load(ctx, s, today)
	if err != nil {
		return false, err
	}
	for _, m := range w.memberships {
		if str(m, "company") == company {
			return true, nil
		}
	}
	for _, md := range w.mandates {
		if str(md, "client") == company {
			return true, nil
		}
	}
	return false, nil
}

// letIn is every company someone is let into today: a current membership's
// company, a current mandate's client.
func (w *world) letIn() []string {
	set := map[string]bool{}
	for _, m := range w.memberships {
		if w.current(m) {
			set[str(m, "company")] = true
		}
	}
	for _, md := range w.mandates {
		if w.current(md) {
			set[str(md, "client")] = true
		}
	}
	out := make([]string, 0, len(set))
	for c := range set {
		out = append(out, c)
	}
	sort.Strings(out)
	return out
}

// LetIn is every company someone is let into today — where an act on every
// company's books needs its right.
func LetIn(ctx context.Context, s *store.Store, today string) ([]string, error) {
	w, err := load(ctx, s, today)
	if err != nil {
		return nil, err
	}
	return w.letIn(), nil
}

// claims reports whether a door writes a grant into a company — the only
// doors open in a company nobody was ever let into.
func claims(act core.Activity) bool {
	return act.Spec.Emits == "membership.granted" || act.Spec.Emits == "mandate.granted"
}

// Table is who can do what in a company: every person with a grant there,
// each with the grants that reach it — the rows of the view.
func Table(ctx context.Context, s *store.Store, company, today string) ([]Principal, error) {
	w, err := load(ctx, s, today)
	if err != nil {
		return nil, err
	}
	var out []Principal
	for id, p := range w.persons {
		var here []Grant
		for _, g := range w.resolve(id) {
			if g.Company == company {
				here = append(here, g)
			}
		}
		if len(here) == 0 {
			continue
		}
		out = append(out, Principal{Handle: str(p, "handle"), PersonID: id, Name: str(p, "name"), Grants: here})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Handle < out[j].Handle })
	return out, nil
}

// Authorize decides a door for a signed-in person: the door must admit a
// person, and the person must hold its right — for an installation door in
// every company a person it acts on is let into (and a handle it registers
// is nobody's yet), in the company they act for otherwise, where a
// company nobody was ever let into is open to its first claim alone. A
// door that names an office (grant_mandate) is the office's act too. The
// inputs come back signed: every "…_by" input carries the person, whatever
// was sent, and a company input left empty names the company acted for.
func Authorize(ctx context.Context, s *store.Store, p Principal, act core.Activity, inputs map[string]any, acting, today string) (map[string]any, error) {
	refuse := func(reason string) error {
		return &ErrRefused{Person: p.Handle, Door: act.Name, Company: acting, Reason: reason}
	}
	if !act.Spec.Admits(core.ActorHuman) {
		return nil, refuse(fmt.Sprintf("the door is not a person's (who: %s)", strings.Join(act.Spec.Who, ", ")))
	}
	if act.Spec.Installation {
		if _, ok := p.MayAnywhere("door", act.Name, act.Domain); !ok {
			return nil, refuse("no role of yours opens this door")
		}
		// An installation door acts on a person wherever they are let in, so
		// its right is needed in every company the person holds a grant in:
		// a right in any one company let an owner anywhere mark KK as left
		// for good (review 2026-10-21, A3). And a handle is one person's.
		w, err := load(ctx, s, today)
		if err != nil {
			return nil, err
		}
		for _, f := range act.Spec.Inputs {
			id, _ := inputs[f.Name].(string)
			if f.Type != "ref<person>" || id == "" {
				continue
			}
			var theirs []string
			for _, g := range w.resolve(id) {
				theirs = append(theirs, g.Company)
			}
			if !p.MayEverywhere(theirs, "door", act.Name, act.Domain) {
				return nil, refuse("no role of yours opens this door everywhere this person is let in")
			}
		}
		if h, _ := inputs["handle"].(string); h != "" {
			if _, taken := w.byHandle[h]; taken {
				return nil, refuse(fmt.Sprintf("%s is another person's handle", h))
			}
		}
	} else {
		if acting == "" {
			return nil, refuse("name the company you act for")
		}
		// Every input that names the company a grant lands in names the one
		// acted for — a mandate's client too: unchecked, it put a stranger's
		// office into any company she named (review 2026-10-21, A2).
		for _, f := range []string{"company", "client"} {
			if named, ok := inputs[f].(string); ok && named != "" && named != acting {
				return nil, refuse(fmt.Sprintf("this names %s — act for %s to do it there", named, named))
			}
		}
		if _, ok := p.May(acting, "door", act.Name, act.Domain); !ok {
			claimed, err := Claimed(ctx, s, acting, today)
			if err != nil {
				return nil, err
			}
			if claimed {
				return nil, refuse("no role of yours opens this door here")
			}
			if !claims(act) {
				return nil, refuse("nobody holds this company yet — the first grant into it is its claim")
			}
		}
		if office, ok := inputs["office"].(string); ok && office != "" && office != acting {
			if _, ok := p.May(office, "door", act.Name, act.Domain); !ok {
				return nil, refuse(fmt.Sprintf("no role of yours opens this door for %s", office))
			}
		}
	}
	// What runs on every company's books is approved by someone who may act
	// in each: approved in one company, a rule booked every other's facts
	// (review 2026-10-21, A5).
	if act.Spec.Everywhere {
		w, err := load(ctx, s, today)
		if err != nil {
			return nil, err
		}
		if !p.MayEverywhere(w.letIn(), "door", act.Name, act.Domain) {
			return nil, refuse("this changes how every company's books are kept — it needs your right in every company let in here")
		}
	}
	signed := make(map[string]any, len(inputs)+1)
	for k, v := range inputs {
		signed[k] = v
	}
	for _, f := range act.Spec.Inputs {
		if core.SignedInput(f) {
			signed[f.Name] = p.Handle
		}
		// A company door that takes a company and was given none acts in
		// the one acted for: a substitution naming none handed the deputy
		// every grant its person held, in every company (A2).
		if f.Name == "company" && f.Type == "string" && !act.Spec.Installation {
			if named, _ := signed["company"].(string); named == "" {
				signed["company"] = acting
			}
		}
	}
	return signed, nil
}
