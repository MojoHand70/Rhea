package shell

import (
	"context"
	"crypto/subtle"
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
// against them. The header is believed only from the face: the face proves
// itself with a key the two share (X-Rhea-Face, RHEA_FACE_KEY), because on
// white01 every process on the machine reaches the kernel's port (review
// 2026-10-21, A7); a signed token replaces it when other faces exist on the
// open API.

// machineRoutes are the machine's own hands, open without a person even
// when the kernel serves persons only: the clock the morning timer moves (a
// catch-up over many days and its simulation are a person's decision, never
// a machine's — review 2026-10-21, A9),
// the host speaking on the timer's call, the face reporting a sign-in, and
// the two reads a face needs before anyone is signed in.
func machineRoute(r *http.Request) bool {
	switch r.Method + " " + r.URL.Path {
	case "POST /api/time/advance",
		"POST /api/speak", "POST /api/activities/record_access/trigger",
		"GET /api/time", "GET /api/live", "GET /api/health":
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
		if handle != "" && (s.RequirePerson || s.FaceKey != "") && !s.fromFace(r) {
			writeErr(w, 401, fmt.Errorf("a person is named by the face only"))
			return
		}
		company := r.Header.Get("X-Rhea-Company")
		if company == "" {
			company = r.URL.Query().Get("company")
		}
		// one spelling of the key from here on: the claim check compared
		// "PL…1111 " untrimmed while the scope trimmed it (review
		// 2026-10-21, A4)
		company = strings.TrimSpace(company)
		if handle == "" {
			if s.RequirePerson && !machineRoute(r) {
				writeErr(w, 401, fmt.Errorf("sign in: this kernel serves signed-in persons only"))
				return
			}
			// Without a person: X-Rhea-Scope lists the companies seen (an
			// office's desk), X-Rhea-Company the one acted for, which must
			// be among them; neither names the installation's rows alone.
			// A request never runs as the owner.
			if strings.Contains(company, ",") || (company == "" && r.Header.Get("X-Rhea-Company") != "") {
				writeErr(w, 400, fmt.Errorf("X-Rhea-Company names one company; X-Rhea-Scope lists the ones seen"))
				return
			}
			ctx := store.WithInstallation(r.Context())
			if sc := r.Header.Get("X-Rhea-Scope"); sc != "" {
				keys := strings.Split(sc, ",")
				for _, k := range keys {
					if strings.TrimSpace(k) == "" {
						writeErr(w, 400, fmt.Errorf("X-Rhea-Scope: an empty company key"))
						return
					}
				}
				ctx = store.WithScope(ctx, keys)
				if company != "" {
					if set, _ := store.ScopeOf(ctx); !contains(set, company) {
						writeErr(w, 400, fmt.Errorf("X-Rhea-Company %s is not in X-Rhea-Scope", company))
						return
					}
					ctx = store.ActFor(ctx, company)
				}
			} else if company != "" {
				ctx = store.WithCompany(ctx, company)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		ctx := r.Context()
		today := s.Clock.Today()
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
			scope = append(scope, company) // a claim in progress: nobody was ever let in
		}
		switch {
		case company != "":
			// Acting for a company is seeing that company alone (and the
			// installation's rows): with the whole desk in view, a door or a
			// read naming an object by its id reached another client's rows
			// on rights held in this one (review 2026-10-21, B1).
			ctx = store.WithCompany(ctx, company)
		case len(scope) == 0:
			// A person nobody lets in anywhere today: a key no company
			// carries, so the installation's rows are read and none written.
			// The installation's own scope here once made an ex-member's
			// tap the morning alarm (review 2026-10-21, A6).
			ctx = store.WithScope(ctx, []string{rights.NoCompany})
		default:
			ctx = store.WithScope(ctx, scope) // the desk: every company of theirs, acting for none
		}
		ctx = rights.WithPrincipal(ctx, p)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// fromFace reports whether a request carries the face's key; a kernel that
// knows none believes no face.
func (s *Server) fromFace(r *http.Request) bool {
	return s.FaceKey != "" && subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Rhea-Face")), []byte(s.FaceKey)) == 1
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
	if s.mayView(r.Context(), vd) {
		return true
	}
	return s.allow(w, r, "view", vd.ID, vd.Domain)
}

// mayView answers allowView's question without answering the request.
func (s *Server) mayView(ctx context.Context, vd *core.ViewDef) bool {
	p := rights.PrincipalOf(ctx)
	if p == nil || s.may(ctx, p, "view", vd.ID, vd.Domain) {
		return true
	}
	if typ := viewType(vd); typ != "" {
		if t, err := s.Store.GetObjectType(ctx, typ); err == nil && s.may(ctx, p, "type", typ, t.Domain) {
			return true
		}
	}
	return false
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

// allowEverywhere answers for an act on every company's books — the clock's
// catch-up and its simulation: a signed-in person needs the right in every
// company anyone is let into; the simulation ran as the owner over every
// company for anyone who asked (review 2026-10-21, A9). Without a person,
// the operator's own shell, it is the owner's.
func (s *Server) allowEverywhere(w http.ResponseWriter, r *http.Request, kind, name, domain string) bool {
	p := rights.PrincipalOf(r.Context())
	if p == nil {
		return true
	}
	companies, err := rights.LetIn(r.Context(), s.Store, s.Clock.Today())
	if err != nil {
		writeErr(w, 500, err)
		return false
	}
	if p.MayEverywhere(companies, kind, name, domain) {
		return true
	}
	writeErr(w, 403, fmt.Errorf("refused: this reaches every company's books — %s needs %s %s in each company let in here", p.Handle, kind, name))
	return false
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
	today := s.Clock.Today()
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
		scope, _ := store.ScopeOf(r.Context())
		writeJSON(w, 200, map[string]any{"person": nil, "scope": scope, "company": store.CompanyOf(r.Context())})
		return
	}
	writeJSON(w, 200, map[string]any{"person": p.Handle, "name": p.Name, "person_id": p.PersonID,
		"companies": p.Companies(), "names": s.companyNames(r.Context(), p.Companies()),
		"company": store.CompanyOf(r.Context()), "grants": p.Grants})
}

// companyNames tells each company a person holds by its own name: the
// company record of kind "self" in that company's own books (the base
// vocabulary's), never a customer's record inside them. It reads in the
// person's own scope, whatever company the request acts for, so a person
// learns the names of their companies and of no other — what the desk's
// switch lists (DIRECTION 2026-10-22, "The desk in order"). A company with
// no such record is left out, and the face shows its key.
func (s *Server) companyNames(ctx context.Context, companies []string) map[string]string {
	names := map[string]string{}
	if len(companies) == 0 {
		return names
	}
	objs, err := s.Store.ObjectsByType(store.WithScope(ctx, companies), "company")
	if err != nil {
		return names
	}
	for _, o := range objs {
		if n, _ := o.State["name"].(string); o.Company != "" && o.State["kind"] == "self" && n != "" {
			names[o.Company] = n
		}
	}
	return names
}
