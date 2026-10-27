// Package rights answers one question from the log: may this person, in
// this company, do this kind of thing by this name (DIRECTION 2026-10-19,
// "People and rights"). The face says who; the kernel says whether.
//
// Everything it reads is data a pack declared and people granted through
// doors — persons, memberships, groups, mandates, substitutions — and the
// system's roles, which are the kernel's (core.RoleRights), so replay
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
// with one of the system's roles, a group the person is in there, an office
// grant reaching a client through a mandate, or a substitution for someone
// who holds any of these.
type Grant struct {
	Company string       `json:"company"` // where the rights hold
	Rights  []core.Right `json:"rights"`  // the role's rights
	// Scope narrows a mandate's reach: the rights the client granted the
	// office. Nil for a membership — the role's rights whole.
	Scope  []core.Right `json:"scope,omitempty"`
	Via    string       `json:"via"`    // "membership" | "group" | "mandate" | "substitution"
	Object string       `json:"object"` // the membership, place in a group, mandate or substitution object
	// Through is the grant a mandate or a substitution carries: the office
	// membership behind a mandate, the substituted person's grant behind a
	// substitution.
	Through   string `json:"through,omitempty"`
	Role      string `json:"role,omitempty"`  // a membership's role, as it reads today
	Group     string `json:"group,omitempty"` // a group's name
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

// MayAsMember reports whether the person may do the thing in the company by
// a right of their own there — a membership, or a substitution standing in
// for one — never through a mandate, whose rights an office does not pass on.
func (p Principal) MayAsMember(company, kind, name, domain string) bool {
	for _, g := range p.Grants {
		if g.Company == company && g.Via != "mandate" && g.Scope == nil && g.Covers(kind, name, domain) {
			return true
		}
	}
	return false
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
	groups        map[string]core.Object // "<company>|<name>": the pack's under "", a company's own under its key
	memberships   []core.Object
	members       []core.Object // group_member: a person or a group in a group
	mandates      []core.Object
	substitutions []core.Object
	today         string
}

func load(ctx context.Context, s *store.Store, today string) (*world, error) {
	ctx = store.WithOwner(ctx) // the kernel's own eyes
	w := &world{persons: map[string]core.Object{}, byHandle: map[string]string{}, groups: map[string]core.Object{}, today: today}
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
	// A grant is one Rhea's own rule wrote, in the company it grants into —
	// the company that gave it, at its own door. The company each grant
	// names was believed, whoever wrote it and wherever: a company's own
	// rule, which an accountant may approve, booked her a membership in
	// another company or made her its owner, and a fact of the
	// installation's let the people pack's rule book a grant into a company
	// nobody acting for it gave (review 2026-10-27, critical 2).
	rheas, err := s.RheasRules(ctx)
	if err != nil {
		return nil, err
	}
	grants := func(typ, into string) ([]core.Object, error) {
		all, err := s.ObjectsByType(ctx, typ)
		if err != nil {
			return nil, err
		}
		var out []core.Object
		for _, o := range all {
			if rheas[o.RuleID] && o.Company != "" && o.Company == str(o, into) {
				out = append(out, o)
			}
		}
		return out, nil
	}
	if w.memberships, err = grants("membership", "company"); err != nil {
		return nil, err
	}
	if w.mandates, err = grants("mandate", "client"); err != nil {
		return nil, err
	}
	if w.substitutions, err = grants("substitution", "company"); err != nil {
		return nil, err
	}
	if w.members, err = grants("group_member", "company"); err != nil {
		return nil, err
	}
	// A group says what its members may do, so it is read as a grant is:
	// Rhea's own rule wrote it, the pack's for every company (no company) or
	// a company's own in that company. The first of a name stands — the
	// pack's before any company's, a company's earliest before a later one
	// of the same name (objects come in log order).
	all, err := s.ObjectsByType(ctx, "group")
	if err != nil {
		return nil, err
	}
	for _, g := range all {
		if !rheas[g.RuleID] || g.Company != str(g, "company") {
			continue
		}
		if _, taken := w.groups["|"+str(g, "name")]; taken {
			continue
		}
		if key := g.Company + "|" + str(g, "name"); w.groups[key].ID == "" {
			w.groups[key] = g
		}
	}
	return w, nil
}

// group finds a group by name for a company: the pack's, every company's,
// first; then the company's own.
func (w *world) group(company, name string) (core.Object, bool) {
	if g, ok := w.groups["|"+name]; ok {
		return g, true
	}
	g, ok := w.groups[company+"|"+name]
	return g, ok && company != ""
}

// groupRights is what a group gives; ok is false for a group nobody
// declared, or one whose rights do not read.
func (w *world) groupRights(company, name string) ([]core.Right, bool) {
	g, ok := w.group(company, name)
	if !ok {
		return nil, false
	}
	rights, err := core.ParseRights(str(g, "rights"))
	return rights, err == nil
}

func str(o core.Object, f string) string { v, _ := o.State[f].(string); return v }

// Keepers maps each client to the office keeping its books on the day: the
// office of a current mandate. A company no office keeps keeps its own — the
// count a promotion asks for (KK, 2026-10-22: ten companies kept by three
// offices or installations).
func Keepers(ctx context.Context, s *store.Store, today string) (map[string]string, error) {
	w, err := load(ctx, s, today)
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, md := range w.mandates {
		if w.current(md) {
			out[str(md, "client")] = str(md, "office")
		}
	}
	return out, nil
}

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

// roleRights is what a role gives — the kernel's, never a record's.
func roleRights(role string) []core.Right {
	rights, err := core.ParseRights(core.RoleRights[role])
	if err != nil {
		panic(fmt.Sprintf("role %s: %v", role, err)) // the kernel's own table; a test reads every role
	}
	return rights
}

// A place is how a person is in a group of a company: the group, the record
// that put them (or the group through which they are in it) there, and the
// dates it holds.
type place struct {
	group, object, through, grantedBy, from, to string
}

// groupsOf is every group a person is in, in a company, today: the groups a
// record names them in, the group a former role names, All users when a
// membership lets them in as anything but a guest — and every group one of
// those is in, and so on up, each group once (a circle ends where it began).
// A place is the stay's it was given in: one dated before the person's
// current stay began — they left, and were let in again — gives nothing, or
// a person taken back as a guest would be an accountant again by a place
// nobody remembered.
func (w *world) groupsOf(personID, company, stay string, former []place, allUsers *place) []place {
	seen := map[string]bool{}
	var out []place
	add := func(p place) {
		if !seen[p.group] {
			seen[p.group] = true
			out = append(out, p)
		}
	}
	for _, m := range w.members {
		if str(m, "company") == company && str(m, "person") == personID && str(m, "person") != "" && w.current(m) && str(m, "from") >= stay {
			add(place{group: str(m, "group"), object: m.ID, grantedBy: str(m, "granted_by"), from: str(m, "from"), to: str(m, "to")})
		}
	}
	for _, p := range former {
		add(p)
	}
	if allUsers != nil {
		add(*allUsers)
	}
	for i := 0; i < len(out); i++ {
		for _, m := range w.members {
			if str(m, "company") == company && str(m, "member_group") == out[i].group && str(m, "person") == "" && w.current(m) {
				add(place{group: str(m, "group"), object: m.ID, through: out[i].object, grantedBy: str(m, "granted_by"), from: str(m, "from"), to: str(m, "to")})
			}
		}
	}
	return out
}

// direct is a person's grants from memberships, the groups they are in and,
// through both, mandates — without substitutions, so a substitution never
// chains.
func (w *world) direct(personID string) []Grant {
	var grants []Grant
	former := map[string][]place{}
	allUsers := map[string]*place{}
	stay := map[string]string{} // the day the person's current stay in a company began
	var companies []string
	for _, m := range w.memberships {
		if str(m, "person") != personID || !w.current(m) {
			continue
		}
		company := str(m, "company")
		role, group, ok := core.RoleOf(str(m, "role"))
		if !ok {
			continue // a role the system does not have grants nothing
		}
		grants = append(grants, Grant{Company: company, Rights: roleRights(role), Via: "membership", Object: m.ID,
			Role: role, GrantedBy: str(m, "granted_by"), From: str(m, "from"), To: str(m, "to"), Reason: str(m, "reason")})
		if _, seen := allUsers[company]; !seen {
			companies = append(companies, company)
			allUsers[company] = nil
		}
		if from := str(m, "from"); stay[company] == "" || from < stay[company] {
			stay[company] = from
		}
		at := place{object: m.ID, grantedBy: str(m, "granted_by"), from: str(m, "from"), to: str(m, "to")}
		if group != "" {
			at.group = group
			former[company] = append(former[company], at)
		}
		if role != core.RoleGuest && allUsers[company] == nil {
			at.group = core.GroupAllUsers
			allUsers[company] = &at
		}
	}
	// A group gives its rights only to a person let into its company: a
	// place in a group is no way in.
	for _, company := range companies {
		for _, p := range w.groupsOf(personID, company, stay[company], former[company], allUsers[company]) {
			rights, ok := w.groupRights(company, p.group)
			if !ok {
				continue // a group nobody declared gives nothing
			}
			grants = append(grants, Grant{Company: company, Rights: rights, Via: "group", Object: p.object, Through: p.through,
				Group: p.group, GrantedBy: p.grantedBy, From: p.from, To: p.to})
		}
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
// today's grants made it anyone's to claim (review 2026-10-21, A4). So is a
// company whose books hold a fact: the operator or a channel booked a
// company's documents before anyone was granted a role there, and the first
// signed-in person to name it became its owner and read them (review
// 2026-10-27); its first owner is the operator's to grant. An unclaimed
// company is open to its first owner or office, and only to that: the
// claim is the first grant written into it.
func Claimed(ctx context.Context, s *store.Store, company, today string) (bool, error) {
	if held, err := s.HoldsBooks(store.WithOwner(ctx), company); err != nil || held {
		return held, err
	}
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
			if acting == core.House {
				// Rhea's staff are let in by the installation's owner alone:
				// claimed by the first comer, the house would make anyone the
				// approver of every company's rules.
				return nil, refuse("the house is entered by the installation owner's grant alone")
			}
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
		} else if reason, err := checkPeople(ctx, s, p, act, inputs, acting, today); err != nil {
			return nil, err
		} else if reason != "" {
			return nil, refuse(reason)
		}
		if office, ok := inputs["office"].(string); ok && office != "" && office != acting {
			// Making a company someone's office is that office's own act: a
			// right of the person's own there, never one held through a
			// mandate — what an office holds in a client it does not pass on,
			// or the office's person made the client a third company's office
			// (KK, 2026-10-23; KSeF's default).
			if !p.MayAsMember(office, "door", act.Name, act.Domain) {
				return nil, refuse(fmt.Sprintf("no role of your own in %s opens this door — a right held through a mandate is not passed on", office))
			}
		}
	}
	// What runs on every company's books is Rhea staff's to approve: a right
	// in the house, however many companies the person is let into — an
	// office let into all its clients is still not Rhea (KK, 2026-10-22;
	// before it, a right in every company let in, review 2026-10-21, A5). A
	// door acting on the company's own alone comes here unmarked: the
	// executor says so (exec.companysOwn).
	if act.Spec.Everywhere {
		if _, ok := p.May(core.House, "door", act.Name, act.Domain); !ok {
			return nil, refuse("this changes how every company's books are kept — it is Rhea staff's to approve")
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
