package shell

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"rhea/internal/core"
	"rhea/internal/rights"
	"rhea/internal/store"
)

// The sign-in step (DIRECTION 2026-10-19, "People and rights"): the face
// authenticates a person and names them on every request (X-Rhea-Person);
// the kernel resolves their grants from the log, makes those grants the
// request's scope (invariant 9: a person sees the companies they hold
// rights in and nothing else), and checks every door and every read
// against them. On white01 the kernel is bound to the machine and the
// face's header is enough; a signed token replaces it when other faces
// exist on the open API.

// machineRoutes are the machine's own hands, open without a person even
// when the kernel serves persons only: the clock the morning timer moves,
// the host speaking on the timer's call, the face reporting a sign-in, and
// the two reads a face needs before anyone is signed in.
func machineRoute(r *http.Request) bool {
	switch r.Method + " " + r.URL.Path {
	case "POST /api/time/advance", "POST /api/time/catchup", "POST /api/time/simulate",
		"POST /api/speak", "POST /api/activities/record_access/trigger",
		"GET /api/time", "GET /api/live":
		return true
	}
	return false
}

// principalHandler puts the request's scope and person on the context.
// Without a person: the caller's company from the header (one key acts for
// a company, a comma list is an office's desk), or the whole log — the
// owner's own shell and the kernel's tools — unless the server serves
// persons only. With a person: their grants are the scope, the named
// company must be one of them (or unclaimed: the first grant into a
// company is its claim), and the refusal of a company is a recorded fact.
func (s *Server) principalHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handle := r.Header.Get("X-Rhea-Person")
		company := r.Header.Get("X-Rhea-Company")
		if company == "" {
			company = r.URL.Query().Get("company")
		}
		if handle == "" {
			if s.RequirePerson && !machineRoute(r) {
				writeErr(w, 401, fmt.Errorf("sign in: this kernel serves signed-in persons only"))
				return
			}
			if keys := strings.Split(company, ","); len(keys) > 1 {
				r = r.WithContext(store.WithScope(r.Context(), keys))
			} else if company != "" {
				r = r.WithContext(store.WithCompany(r.Context(), company))
			}
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		today := s.Clock.TodayUTC()
		p, err := rights.Resolve(ctx, s.Store, handle, today)
		if err != nil {
			writeErr(w, 403, fmt.Errorf("refused: %w", err))
			return
		}
		if strings.Contains(company, ",") {
			writeErr(w, 400, fmt.Errorf("a person acts for one company, or for none to see all of theirs"))
			return
		}
		scope := p.Companies()
		switch {
		case company == "" && len(scope) == 1:
			company = scope[0]
		case company != "" && !contains(scope, company):
			claimed, err := rights.Claimed(ctx, s.Store, company, today)
			if err != nil {
				writeErr(w, 500, err)
				return
			}
			if claimed {
				refused := &rights.ErrRefused{Person: handle, Company: company, Reason: "no role of yours reaches this company"}
				if err := s.Exec.RecordRefusal(ctx, refused); err != nil {
					writeErr(w, 500, err)
					return
				}
				writeErr(w, 403, refused)
				return
			}
			scope = append(scope, company) // a claim in progress: nobody holds it yet
		}
		if len(scope) == 0 {
			scope = []string{rights.NoCompany}
		}
		ctx = store.WithScope(ctx, scope)
		if company != "" {
			ctx = store.ActFor(ctx, company)
		}
		ctx = rights.WithPrincipal(ctx, p)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func contains(set []string, x string) bool {
	for _, s := range set {
		if s == x {
			return true
		}
	}
	return false
}

// allow answers whether the signed-in person may read a thing — in the
// company they act for, or in every company of theirs when they act for
// none (an office's desk open on all its clients shows what its person
// may see in each). Without a person every read is the owner's. A
// refusal answers 403 and says why.
func (s *Server) allow(w http.ResponseWriter, r *http.Request, kind, name, domain string) bool {
	p := rights.PrincipalOf(r.Context())
	if p == nil {
		return true
	}
	if s.may(r.Context(), p, kind, name, domain) {
		return true
	}
	where := store.CompanyOf(r.Context())
	if where == "" {
		where = "every company of yours"
	}
	writeErr(w, 403, fmt.Errorf("refused: no role of %s reaches %s %s in %s", p.Handle, kind, name, where))
	return false
}

func (s *Server) may(ctx context.Context, p *rights.Principal, kind, name, domain string) bool {
	if acting := store.CompanyOf(ctx); acting != "" {
		_, ok := p.May(acting, kind, name, domain)
		return ok
	}
	return p.MayEverywhere(p.Companies(), kind, name, domain)
}

// allowView gates a view: by its own name, or — for a view that shows one
// type's objects — by the type.
func (s *Server) allowView(w http.ResponseWriter, r *http.Request, vd *core.ViewDef) bool {
	p := rights.PrincipalOf(r.Context())
	if p == nil || s.may(r.Context(), p, "view", vd.ID, vd.Domain) {
		return true
	}
	if typ := viewType(vd); typ != "" {
		if t, err := s.Store.GetObjectType(r.Context(), typ); err == nil && s.may(r.Context(), p, "type", typ, t.Domain) {
			return true
		}
	}
	return s.allow(w, r, "view", vd.ID, vd.Domain)
}

// viewType is the object type a list, detail or statement view shows.
func viewType(vd *core.ViewDef) string {
	switch vd.Notion {
	case "list", "detail", "statement":
		var spec struct {
			ObjectType string `json:"object_type"`
		}
		if json.Unmarshal(vd.Spec, &spec) == nil {
			return spec.ObjectType
		}
	}
	return ""
}

// allowType gates a read of one object type by the type's domain.
func (s *Server) allowType(w http.ResponseWriter, r *http.Request, typ string) bool {
	if rights.PrincipalOf(r.Context()) == nil {
		return true
	}
	t, err := s.Store.GetObjectType(r.Context(), typ)
	if err != nil {
		writeErr(w, 404, err)
		return false
	}
	return s.allow(w, r, "type", typ, t.Domain)
}

// System views, named so a role can grant them one by one: view:* and
// domain:system grant them all.
const (
	ViewWorklist = "worklist"
	ViewGate     = "gate" // rules, bundles, their previews
	ViewFiles    = "files"
	ViewChannels = "channels"
	ViewSaid     = "said"
	ViewAnswer   = "answer" // the host querying the books
	ViewLearning = "learning"
	ViewUntaught = "untaught"
)

func (s *Server) allowSystem(w http.ResponseWriter, r *http.Request, view string) bool {
	return s.allow(w, r, "view", view, "system")
}

// --- who can do what ------------------------------------------------------

// A rights column is one thing a right can name: a door, a view or a type.
type rightsColumn struct {
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Label  string `json:"label"`
}

type rightsRow struct {
	Handle string                  `json:"handle"`
	Name   string                  `json:"name"`
	Grants []rights.Grant          `json:"grants"`
	Cells  map[string]rights.Grant `json:"cells"` // "<kind>:<name>" → the first grant that opens it
}

// serveRights renders the rights notion (the common people pack's
// "who-can-do-what"): rows the persons with rights in the company, columns
// every door, view and type the installation offers, each cell the grant
// — membership, mandate or substitution — that opens it, with who granted
// it and from when. The number-opens move applied to rights.
func (s *Server) serveRights(ctx context.Context, w http.ResponseWriter, vd *core.ViewDef, company string) {
	if company == "" {
		company = store.CompanyOf(ctx)
	}
	if company == "" {
		writeErr(w, 400, fmt.Errorf("who can do what: name the company (?company=)"))
		return
	}
	if p := rights.PrincipalOf(ctx); p != nil {
		if _, ok := p.May(company, "view", "who-can-do-what", core.PeopleDomain); !ok {
			writeErr(w, 403, fmt.Errorf("refused: no role of %s sees who can do what in %s", p.Handle, company))
			return
		}
	}
	today := s.Clock.TodayUTC()
	table, err := rights.Table(ctx, s.Store, company, today)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	var cols []rightsColumn
	acts, err := s.Store.ActiveActivities(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	for _, a := range acts {
		if a.Spec.Channel != nil || !a.Spec.Admits(core.ActorHuman) {
			continue // a channel or the machine's own door is nobody's right
		}
		cols = append(cols, rightsColumn{Kind: "door", Name: a.Name, Domain: a.Domain, Label: a.Description})
	}
	vds, err := s.effectiveViewDefs(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	for _, v := range vds {
		if v.Notion == "detail" {
			continue
		}
		cols = append(cols, rightsColumn{Kind: "view", Name: v.ID, Domain: v.Domain, Label: v.Title})
	}
	types, err := s.Store.ListObjectTypes(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	for _, t := range types {
		cols = append(cols, rightsColumn{Kind: "type", Name: t.Name, Domain: t.Domain, Label: t.Name})
	}
	rows := make([]rightsRow, 0, len(table))
	for _, p := range table {
		row := rightsRow{Handle: p.Handle, Name: p.Name, Grants: p.Grants, Cells: map[string]rights.Grant{}}
		for _, c := range cols {
			if g, ok := p.May(company, c.Kind, c.Name, c.Domain); ok {
				row.Cells[c.Kind+":"+c.Name] = g
			}
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Handle < rows[j].Handle })
	out := map[string]any{"company": company, "today": today, "columns": cols, "rows": rows}
	if vd != nil {
		out["view"] = vd
	}
	writeJSON(w, 200, out)
}

// handleRights is the same table without a view definition: GET
// /api/rights?company=<key> — what a face asks before it renders a People
// screen, and what a person asks of their own entity.
func (s *Server) handleRights(w http.ResponseWriter, r *http.Request) {
	s.serveRights(r.Context(), w, nil, r.URL.Query().Get("company"))
}

// handleWhoAmI tells a face who the kernel takes the request for: the
// person, their companies and the one they act for — the desk's header.
func (s *Server) handleWhoAmI(w http.ResponseWriter, r *http.Request) {
	p := rights.PrincipalOf(r.Context())
	if p == nil {
		writeJSON(w, 200, map[string]any{"person": nil, "scope": store.ScopeOf(r.Context()), "company": store.CompanyOf(r.Context())})
		return
	}
	writeJSON(w, 200, map[string]any{"person": p.Handle, "name": p.Name, "person_id": p.PersonID,
		"companies": p.Companies(), "company": store.CompanyOf(r.Context()), "grants": p.Grants})
}
