package rights

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"rhea/internal/core"
	"rhea/internal/store"
)

// checkPeople is what a door that says who is in what asks beyond its right
// (KK, 2026-10-27: the roles are the system's, the business is groups): a
// role must be one of the system's; a group must exist, may not be inside
// itself, and All users is Rhea's to keep; a person is let into the company
// before a group takes them. And nobody raises their own rights by their own
// hand (Claude, 2026-10-27): a role given to oneself, or a group one is put
// into — oneself, or a group one is in — gives nothing the person does not
// already hold in the company. A Tenant Owner says who is in a company and
// sees no books by that role alone; adding herself to Accountant would undo
// it. Its answer is the refusal's reason, empty when the act may go on.
func checkPeople(ctx context.Context, s *store.Store, p Principal, act core.Activity, inputs map[string]any, company, today string) (string, error) {
	in := func(name string) string { v, _ := inputs[name].(string); return v }
	switch act.Spec.Emits {
	case "membership.granted":
		role := in("role")
		if !core.IsRole(role) {
			return "", nil // the door's own list refuses it, naming the roles
		}
		if in("person") != p.PersonID {
			return "", nil
		}
		if r, ok := p.lacks(ctx, s, company, roleRights(role)); !ok {
			return fmt.Sprintf("you cannot give yourself the role %s: it gives %s, which you do not hold here", role, r), nil
		}
		return "", nil
	case "group.declared":
		w, err := load(ctx, s, today)
		if err != nil {
			return "", err
		}
		name := in("name")
		if _, ok := w.groups["|"+name]; ok {
			return fmt.Sprintf("%s is one of the groups every company has", name), nil
		}
		if _, ok := w.groups[company+"|"+name]; ok {
			return fmt.Sprintf("the company has a group %s already", name), nil
		}
		if _, err := core.ParseRights(in("rights")); err != nil {
			return fmt.Sprintf("what the group gives does not read: %v", err), nil
		}
		return "", nil
	case "group.joined":
		w, err := load(ctx, s, today)
		if err != nil {
			return "", err
		}
		group, person, member := in("group"), in("person"), in("member_group")
		switch {
		case (person == "") == (member == ""):
			return fmt.Sprintf("name a person or a group to put in %s — one of them", group), nil
		case group == core.GroupAllUsers:
			return "Rhea keeps All users: everyone let into the company is in it", nil
		}
		if _, ok := w.group(company, group); !ok {
			return fmt.Sprintf("the company has no group %s", group), nil
		}
		if person != "" {
			if !w.letInto(person, company, in("from")) {
				return "the person is not let into the company on the place's first day — invite them first", nil
			}
			if m, ok := w.placeOf(company, group, "person", person); ok {
				return fmt.Sprintf("the person is in %s already (from %s)", group, str(m, "from")), nil
			}
		}
		if member != "" {
			if _, ok := w.group(company, member); !ok {
				return fmt.Sprintf("the company has no group %s", member), nil
			}
			if member == group || w.within(company, group, member) {
				return fmt.Sprintf("%s is inside %s already: a group cannot be inside itself", group, member), nil
			}
			if m, ok := w.placeOf(company, group, "member_group", member); ok {
				return fmt.Sprintf("%s is in %s already (from %s)", member, group, str(m, "from")), nil
			}
		}
		// what the person acting would gain: the group and every group it is in
		mine := person == p.PersonID
		if member != "" && !mine {
			for _, g := range p.Grants {
				if g.Company == company && g.Via == "group" && g.Group == member {
					mine = true
				}
			}
		}
		if !mine {
			return "", nil
		}
		var gained []core.Right
		for _, g := range append([]string{group}, w.above(company, group)...) {
			rights, _ := w.groupRights(company, g)
			gained = append(gained, rights...)
		}
		if r, ok := p.lacks(ctx, s, company, gained); !ok {
			return fmt.Sprintf("you cannot put yourself in %s: it gives %s, which you do not hold here", group, r), nil
		}
		return "", nil
	}
	return "", nil
}

// letInto reports whether a membership lets the person into the company on
// a day — a place in a group is given within a stay, whatever day it is
// recorded on.
func (w *world) letInto(personID, company, day string) bool {
	for _, m := range w.memberships {
		if str(m, "person") != personID || str(m, "company") != company {
			continue
		}
		if from, to := str(m, "from"), str(m, "to"); from != "" && from <= day && (to == "" || to >= day) {
			return true
		}
	}
	return false
}

// placeOf finds a place in a group that holds today or later — a person or
// a group already put there.
func (w *world) placeOf(company, group, field, value string) (core.Object, bool) {
	for _, m := range w.members {
		if str(m, "company") != company || str(m, "group") != group || str(m, field) != value {
			continue
		}
		if to := str(m, "to"); to == "" || to >= w.today {
			return m, true
		}
	}
	return core.Object{}, false
}

// above is every group a group is in, and so on up, on any day — a circle
// is refused whenever its edges hold.
func (w *world) above(company, group string) []string {
	seen := map[string]bool{group: true}
	var out []string
	for i, at := -1, group; i < len(out); i++ {
		if i >= 0 {
			at = out[i]
		}
		for _, m := range w.members {
			if str(m, "company") == company && str(m, "member_group") == at && !seen[str(m, "group")] {
				seen[str(m, "group")] = true
				out = append(out, str(m, "group"))
			}
		}
	}
	return out
}

// within reports whether group is, through any chain, inside outer.
func (w *world) within(company, group, outer string) bool {
	for _, g := range w.above(company, group) {
		if g == outer {
			return true
		}
	}
	return false
}

// lacks finds the first of the rights the person does not already hold in
// the company; ok when they hold them all.
func (p Principal) lacks(ctx context.Context, s *store.Store, company string, rights []core.Right) (core.Right, bool) {
	for _, r := range rights {
		domain := domainOf(ctx, s, r)
		held := false
		for _, g := range p.Grants {
			if g.Company == company && core.Implies(g.Rights, r, domain) && (g.Scope == nil || core.Implies(g.Scope, r, domain)) {
				held = true
				break
			}
		}
		if !held {
			return r, false
		}
	}
	return core.Right{}, true
}

// domainOf is the domain a right names a thing in: a door's, a view's or a
// kind of record's; empty for a wildcard or a name nothing defines.
func domainOf(ctx context.Context, s *store.Store, r core.Right) string {
	if r.Name == "*" {
		return ""
	}
	switch r.Kind {
	case "door":
		if a, err := s.GetActivity(ctx, r.Name); err == nil {
			return a.Domain
		}
	case "type":
		if t, err := s.GetObjectType(ctx, r.Name); err == nil {
			return t.Domain
		}
	case "view":
		if vds, err := s.LatestViewDefs(ctx); err == nil {
			for _, v := range vds {
				if v.ID == r.Name {
					return v.Domain
				}
			}
		}
	}
	return ""
}

// A Member is someone in a group of a company: a person or a group, by the
// record that put them there — or, with no record, how Rhea counts them.
type Member struct {
	Person    string `json:"person,omitempty"` // the person's id
	Handle    string `json:"handle,omitempty"`
	Name      string `json:"name,omitempty"`
	Group     string `json:"group,omitempty"` // a group inside the group
	Object    string `json:"object,omitempty"`
	GrantedBy string `json:"granted_by,omitempty"`
	From      string `json:"from,omitempty"`
	To        string `json:"to,omitempty"`
	// Counted says how a member with no record of their own is in it: "all
	// users" for everyone let in, "former role" for a grant made under the
	// roles of v3, "company manager" for the manager of a company whose
	// starting group nobody was put in yet (KK, 2026-10-27: until handed
	// over, so a sole trader does everything on day one).
	Counted string `json:"counted,omitempty"`
}

// A GroupView is one group of a company, what it gives and who is in it today.
type GroupView struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Rights      string   `json:"rights"`
	Everyones   bool     `json:"everyones"` // the pack's, every company's
	Kept        bool     `json:"kept"`      // All users: Rhea keeps it
	Object      string   `json:"object"`
	Members     []Member `json:"members"`
}

// Groups is every group of a company today — the pack's first, then the
// company's own — each with who is in it: the people and groups put there,
// everyone let in for All users, a grant of a former role, and the company's
// managers in a starting group nobody was put into yet.
func Groups(ctx context.Context, s *store.Store, company, today string) ([]GroupView, error) {
	w, err := load(ctx, s, today)
	if err != nil {
		return nil, err
	}
	person := func(id string) (string, string) {
		p := w.persons[id]
		return str(p, "handle"), str(p, "name")
	}
	var out []GroupView
	for key, g := range w.groups {
		owner, name, _ := strings.Cut(key, "|")
		if owner != "" && owner != company {
			continue
		}
		v := GroupView{Name: name, Description: str(g, "description"), Rights: str(g, "rights"), Everyones: owner == "",
			Kept: name == core.GroupAllUsers, Object: g.ID, Members: []Member{}}
		for _, m := range w.members {
			if str(m, "company") != company || str(m, "group") != name || !w.current(m) {
				continue
			}
			mb := Member{Group: str(m, "member_group"), Object: m.ID, GrantedBy: str(m, "granted_by"), From: str(m, "from"), To: str(m, "to")}
			if id := str(m, "person"); id != "" {
				mb.Person = id
				mb.Handle, mb.Name = person(id)
			}
			v.Members = append(v.Members, mb)
		}
		var managers []Member
		for _, m := range w.memberships {
			if str(m, "company") != company || !w.current(m) {
				continue
			}
			role, former, ok := core.RoleOf(str(m, "role"))
			if !ok {
				continue
			}
			mb := Member{Person: str(m, "person"), Object: m.ID, GrantedBy: str(m, "granted_by"), From: str(m, "from"), To: str(m, "to")}
			mb.Handle, mb.Name = person(mb.Person)
			switch {
			case v.Kept && role != core.RoleGuest:
				mb.Counted = "all users"
				v.Members = append(v.Members, mb)
			case former == name:
				mb.Counted = "former role"
				v.Members = append(v.Members, mb)
			case role == core.RoleCompanyManager && v.Everyones && !v.Kept:
				mb.Counted = "company manager"
				managers = append(managers, mb)
			}
		}
		nobody := true
		for _, mb := range v.Members {
			if mb.Person != "" && mb.Counted != "company manager" {
				nobody = false
			}
		}
		if nobody {
			v.Members = append(v.Members, managers...)
		}
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Everyones != out[j].Everyones {
			return out[i].Everyones
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// MayMove answers whether a person may make a move a kind of record gives
// to a group, in a company, on the books' day (KK, 2026-10-27): the move is
// a person's, never the machine's; the person is in the group there —
// their own place, a group in it, an office's group of that name reaching
// a client through its mandate, standing in for someone who is; and when
// the record names who proposed it, that person makes the move only when
// nobody else could. A group every company has that nobody was put in yet
// is its manager's, until handed over. The answer is the refusal's reason,
// empty when the person may.
func MayMove(ctx context.Context, s *store.Store, handle, company, group, proposer, today string) (string, error) {
	m, err := NewMoveCheck(ctx, s, today)
	if err != nil {
		return "", err
	}
	return m.May(handle, company, group, proposer), nil
}

// A MoveCheck answers MayMove for many records with the log read once — the
// list of what waits for a person (GET /api/moves).
type MoveCheck struct{ w *world }

// NewMoveCheck reads who is in what on the books' day.
func NewMoveCheck(ctx context.Context, s *store.Store, today string) (*MoveCheck, error) {
	w, err := load(ctx, s, today)
	if err != nil {
		return nil, err
	}
	return &MoveCheck{w: w}, nil
}

// May is MayMove's answer: the refusal's reason, empty when the person may.
func (m *MoveCheck) May(handle, company, group, proposer string) string {
	w := m.w
	if core.ActorKind(handle) != core.ActorHuman {
		return fmt.Sprintf("a person in %s makes this move, never the machine", group)
	}
	id, ok := w.byHandle[handle]
	if !ok || str(w.persons[id], "status") == "left" {
		return fmt.Sprintf("%s is no person let into Rhea — a person in %s makes this move", handle, group)
	}
	movers := w.movers(company, group)
	if len(movers) == 0 {
		return fmt.Sprintf("nobody is in %s here — put someone in it first", group)
	}
	if !movers[id] {
		return fmt.Sprintf("a person in %s makes this move, and you are not in it", group)
	}
	if proposer != "" && proposer == handle && len(movers) > 1 {
		return fmt.Sprintf("you proposed it — another person in %s makes this move", group)
	}
	return ""
}

// Movers is everyone in a company who may make a move a group is given.
func (m *MoveCheck) Movers(company, group string) []string {
	var out []string
	for id := range m.w.movers(company, group) {
		out = append(out, str(m.w.persons[id], "handle"))
	}
	sort.Strings(out)
	return out
}

// Movers is everyone who may make a move a group is given, in a company,
// today — the people a task waiting for the group is for.
func Movers(ctx context.Context, s *store.Store, company, group, today string) ([]string, error) {
	m, err := NewMoveCheck(ctx, s, today)
	if err != nil {
		return nil, err
	}
	return m.Movers(company, group), nil
}

// movers is the people of a group in a company today, by person id.
func (w *world) movers(company, group string) map[string]bool {
	out := map[string]bool{}
	for id, p := range w.persons {
		if str(p, "status") == "left" {
			continue
		}
		for _, g := range w.resolve(id) {
			if g.Company == company && g.Group == group {
				out[id] = true
				break
			}
		}
	}
	if _, everyones := w.groups["|"+group]; len(out) == 0 && everyones && group != core.GroupAllUsers {
		for _, m := range w.memberships {
			if role, _, ok := core.RoleOf(str(m, "role")); ok && role == core.RoleCompanyManager && str(m, "company") == company && w.current(m) {
				out[str(m, "person")] = true
			}
		}
	}
	return out
}
