package core

import (
	"fmt"
	"strings"
)

// The rights vocabulary (DIRECTION 2026-10-19, "People and rights"): a right
// is one word a role is made of, and the kernel asks the same question at
// every door and every read — may this person, in this company, do this
// kind of thing by this name. The words:
//
//	"*"                everything, including who may do what
//	domain:<d>         every door, view and type of domain d
//	door:<name>        one door;  door:* every door outside the people domain
//	view:<id>          one view;  view:* every view
//	type:<name>        one object type;  type:* every type
//	state:<type>/<s>   a smart document's state (vocabulary now; enforced
//	                   when smart documents land)
//	sees:<group>       sees the work of a group — the supervisor's first right
//	approves:<group>   approves for a group — the supervisor's second right
//
// door:* stops at the people domain the way KSeF's ladder does: the right
// to do the work never implies the right to say who may do it. Managing
// who-may-do-what is granted by name — domain:people, a door's name, or *.
type Right struct {
	Kind string // "*", "domain", "door", "view", "type", "state", "sees", "approves"
	Name string // "*" or the thing's name; empty for Kind "*"
}

// PeopleDomain is the domain of the doors that say who may do what; a
// wildcard over doors never reaches it.
const PeopleDomain = "people"

// RightKinds is the closed list of kinds a right may name.
var RightKinds = []string{"domain", "door", "view", "type", "state", "sees", "approves"}

// ParseRight reads one word of the vocabulary.
func ParseRight(s string) (Right, error) {
	s = strings.TrimSpace(s)
	if s == "*" {
		return Right{Kind: "*"}, nil
	}
	kind, name, ok := strings.Cut(s, ":")
	if !ok || name == "" {
		return Right{}, fmt.Errorf("right %q: want kind:name (%s) or *", s, strings.Join(RightKinds, ", "))
	}
	known := false
	for _, k := range RightKinds {
		if k == kind {
			known = true
		}
	}
	if !known {
		return Right{}, fmt.Errorf("right %q: unknown kind %q (%s)", s, kind, strings.Join(RightKinds, ", "))
	}
	if name == "*" && (kind == "domain" || kind == "sees" || kind == "approves" || kind == "state") {
		return Right{}, fmt.Errorf("right %q: %s takes a name, not *", s, kind)
	}
	if kind == "state" && !strings.Contains(name, "/") {
		return Right{}, fmt.Errorf("right %q: a state is <type>/<state>", s)
	}
	if strings.ContainsAny(name, " \t\n") {
		return Right{}, fmt.Errorf("right %q: a name has no spaces", s)
	}
	return Right{Kind: kind, Name: name}, nil
}

// ParseRights reads a role's rights: words separated by spaces.
func ParseRights(s string) ([]Right, error) {
	var out []Right
	for _, w := range strings.Fields(s) {
		r, err := ParseRight(w)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func (r Right) String() string {
	if r.Kind == "*" {
		return "*"
	}
	return r.Kind + ":" + r.Name
}

// Covers reports whether the right grants a thing: its kind (door, view,
// type, state, sees, approves), its name, and the domain it belongs to.
func (r Right) Covers(kind, name, domain string) bool {
	switch r.Kind {
	case "*":
		return true
	case "domain":
		return r.Name == domain && (kind == "door" || kind == "view" || kind == "type")
	}
	if r.Kind != kind {
		return false
	}
	if r.Name == "*" {
		return kind != "door" || domain != PeopleDomain
	}
	return r.Name == name
}

// CoveredBy reports whether any of the rights covers the thing.
func CoveredBy(rights []Right, kind, name, domain string) bool {
	for _, r := range rights {
		if r.Covers(kind, name, domain) {
			return true
		}
	}
	return false
}

// Implies reports whether the rights already give everything one right
// gives: "*" only by "*", a domain by the same domain, a wildcard by the same
// wildcard, a named thing by whatever covers it in its domain. A right is
// never held by a narrower one: door:* is not held by domain:finance.
func Implies(have []Right, r Right, domain string) bool {
	for _, h := range have {
		switch {
		case h.Kind == "*":
			return true
		case r.Kind == "*":
		case r.Kind == "domain":
			if h.Kind == "domain" && h.Name == r.Name {
				return true
			}
		case r.Name == "*":
			if h.Kind == r.Kind && h.Name == "*" {
				return true
			}
		default:
			if h.Covers(r.Kind, r.Name, domain) {
				return true
			}
		}
	}
	return false
}

// Actor kinds: the vocabulary Activity.Who is written in. A person signs
// with a bare handle (the face) or "shell:<user>" / "cli:<user>"; the kernel
// and its adapters and the model have their own prefixes.
const (
	ActorHuman   = "human"
	ActorAgent   = "agent"
	ActorAdapter = "adapter"
)

// ActorKind tells what kind of actor signed: "agent:<model>" is the model,
// "adapter:<name>", "channel:<name>" and "kernel" are the machine's own
// hands, everything else is a person.
func ActorKind(actor string) string {
	switch {
	case strings.HasPrefix(actor, "agent:"):
		return ActorAgent
	case actor == "kernel", strings.HasPrefix(actor, "adapter:"), strings.HasPrefix(actor, "channel:"):
		return ActorAdapter
	}
	return ActorHuman
}

// Admits reports whether a door's Who admits an actor kind; a door that
// declares nobody admits everybody — Who is a restriction, never a lock.
func (sp ActivitySpec) Admits(kind string) bool {
	if len(sp.Who) == 0 {
		return true
	}
	for _, w := range sp.Who {
		if w == kind {
			return true
		}
	}
	return false
}

// SignedInput reports whether an input carries who did the thing — "by",
// "approved_by", "granted_by" — the inputs a signed-in person's door writes
// the person into, whatever the caller sent: the server says who decided.
func SignedInput(f FieldDef) bool {
	return f.Type == "string" && (f.Name == "by" || strings.HasSuffix(f.Name, "_by"))
}

// The access namespace: the kernel's own record of who was let in and who
// was refused — raw facts no rule explains, kept out of the worklist like
// the other system verbs.
const (
	// EventAccessRefused: a person was refused at a door or a company —
	// payload {person, door, company, reason}, under the company refused.
	EventAccessRefused = "access.refused"
	// EventAccessRecorded: a sign-in, a failed sign-in, a sign-out, an
	// enrolment — the face's report through record_access, payload
	// {person, outcome, method, detail}.
	EventAccessRecorded = "access.recorded"
)
