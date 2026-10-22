package shell_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/shell"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

// grantsWorld is the ground the grant tests stand on (the whole-tree
// review of 2026-10-21, step 3): a kernel behind its shell with the people
// pack approved, notes with amounts through a door of their own, and a
// briefing over them the host speaks — every word she is asked for is
// counted, so a test can say she was never asked.
type grantsWorld struct {
	t     *testing.T
	ctx   context.Context
	s     *store.Store
	x     *exec.Executor
	ts    *httptest.Server
	today string
	asked atomic.Int32
}

func newGrantsWorld(t *testing.T) *grantsWorld {
	t.Helper()
	w := &grantsWorld{t: t, ctx: context.Background(), s: storetest.New(t)}
	w.x = &exec.Executor{Store: w.s}
	srv := &shell.Server{Store: w.s, Exec: w.x, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb"),
		Agent: &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
			w.asked.Add(1)
			return "One note was taken, worth 12.50 PLN.", nil
		}}}
	w.x.Today = srv.Clock.TodayUTC
	w.today = srv.Clock.TodayUTC()
	w.ts = httptest.NewServer(srv.Handler())
	t.Cleanup(w.ts.Close)

	sum, err := pack.Load(w.ctx, w.s, filepath.Join("..", "..", "packs", "common", "people.json"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, errs, err := w.x.ApproveBundle(w.ctx, sum.Bundle, "operator", w.today); err != nil || len(errs) != 0 {
		t.Fatalf("approve: %v %v", err, errs)
	}
	if err := w.s.InsertObjectType(w.ctx, core.ObjectType{Name: "note", Version: 1, Domain: "work", LabelField: "text", Amendable: []string{"text"},
		Fields: []core.FieldDef{{Name: "text", Type: "string", Required: true}, {Name: "amount", Type: "money", Required: true}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := w.s.InsertRuleVersion(w.ctx, core.Rule{ID: "note", Status: core.StatusActive, Priority: 100, EffectiveFrom: "2026-01-01",
		CreatedBy: "test", Description: "notes", Spec: core.RuleSpec{Match: core.Match{EventType: "note.taken"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text", "amount": "=$.amount"}}}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := w.s.InsertActivityVersion(w.ctx, core.Activity{Name: "take_note", Status: core.StatusActive, Domain: "work", Description: "note",
		Spec: core.ActivitySpec{Inputs: []core.FieldDef{{Name: "text", Type: "string", Required: true}, {Name: "amount", Type: "money", Required: true}},
			Emits: "note.taken", Who: []string{"human"}}}); err != nil {
		t.Fatal(err)
	}
	facts, _ := json.Marshal(core.AnalysisSpec{
		SQL:     `SELECT COUNT(*) AS notes, SUM(CAST(json_extract(state, '$.amount') AS BIGINT)) AS total FROM objects WHERE object_type = 'note'`,
		Columns: []core.ColumnSpec{{Field: "notes", Label: "Notes", Type: "int"}, {Field: "total", Label: "Total", Type: "money"}}})
	if err := w.s.InsertViewDef(w.ctx, core.ViewDef{ID: "notes-today", Version: 1, Notion: "analysis", Title: "Notes today",
		Domain: "work", Function: "analysis", Spec: facts}); err != nil {
		t.Fatal(err)
	}
	narrative, _ := json.Marshal(core.NarrativeSpec{Facts: []core.NarrativeFact{{View: "notes-today", As: "the notes"}}, Voice: "briefly"})
	if err := w.s.InsertViewDef(w.ctx, core.ViewDef{ID: "briefing", Version: 1, Notion: "narrative", Title: "Briefing",
		Domain: "work", Function: "briefing", Spec: narrative}); err != nil {
		t.Fatal(err)
	}
	return w
}

// call is one request: the person the face names (none for the operator's
// own tools) and the company they act for.
func (w *grantsWorld) call(person, company, method, path string, body any, out any) int {
	w.t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, w.ts.URL+path, rd)
	req.Header.Set("Content-Type", "application/json")
	if person != "" {
		req.Header.Set("X-Rhea-Person", person)
	}
	if company != "" {
		req.Header.Set("X-Rhea-Company", company)
	}
	res, err := w.ts.Client().Do(req)
	if err != nil {
		w.t.Fatal(err)
	}
	defer res.Body.Close()
	if out != nil {
		json.NewDecoder(res.Body).Decode(out)
	}
	return res.StatusCode
}

func (w *grantsWorld) door(person, company, name string, inputs map[string]any) (int, map[string]any) {
	w.t.Helper()
	var out map[string]any
	code := w.call(person, company, "POST", "/api/activities/"+name+"/trigger", map[string]any{"inputs": inputs}, &out)
	return code, out
}

func (w *grantsWorld) mustDoor(person, company, name string, inputs map[string]any) {
	w.t.Helper()
	if code, out := w.door(person, company, name, inputs); code != 200 || out["booked"] != float64(1) {
		w.t.Fatalf("%s at %s for %s: %d %v", person, name, company, code, out)
	}
}

func (w *grantsWorld) personID(handle string) string {
	w.t.Helper()
	ids, err := w.s.FindObjectIDsByField(w.ctx, "person", "handle", handle)
	if err != nil || len(ids) != 1 {
		w.t.Fatalf("%s: %v %v", handle, ids, err)
	}
	return ids[0]
}

// person registers someone through the operator's tools and, when a
// company is named, grants them a role there from the first of the year.
func (w *grantsWorld) person(name, handle, company, role string) string {
	w.t.Helper()
	w.mustDoor("", "", "register_person", map[string]any{"name": name, "handle": handle, "role": role})
	id := w.personID(handle)
	if company != "" {
		w.mustDoor("", company, "grant_membership", map[string]any{"person": id, "company": company, "role": role, "from": "2026-01-01", "granted_by": "operator"})
	}
	return id
}

// dayFrom is the books' day moved by n days.
func (w *grantsWorld) dayFrom(n int) string {
	d, err := time.Parse("2006-01-02", w.today)
	if err != nil {
		w.t.Fatal(err)
	}
	return d.AddDate(0, 0, n).Format("2006-01-02")
}

// A person whose grants have ended hears nothing and sees no company's
// books (review 2026-10-21, A6): the kernel gave a signed-in person with no
// current grant the installation's own scope, and the host took any request
// in that scope for the morning alarm — she spoke for every company,
// figures and all, at the ex-member's tap, paid a model call per company
// and filed her words in each company's log. The alarm, which names no
// person, still speaks for every company.
func TestAPersonOfNoCompanyHearsNothing(t *testing.T) {
	w := newGrantsWorld(t)
	a, b := "PL5250001111", "PL5260001246"
	w.person("Krzysztof Kowal", "kk", a, "owner")
	w.person("Olga Nowa", "olga", b, "owner")
	w.mustDoor("kk", a, "take_note", map[string]any{"text": "a's note", "amount": "12.50"})
	w.mustDoor("olga", b, "take_note", map[string]any{"text": "b's note", "amount": "12.50"})
	// anna kept a's books until yesterday
	anna := w.person("Anna Nowak", "anna", "", "accountant")
	w.mustDoor("kk", a, "grant_membership", map[string]any{"person": anna, "company": a, "role": "accountant", "from": "2026-01-01", "to": w.dayFrom(-1), "granted_by": "kk"})

	if code := w.call("anna", "", "POST", "/api/speak", nil, nil); code != 403 {
		t.Fatalf("anna asking the host to speak: %d", code)
	}
	if n := w.asked.Load(); n != 0 {
		t.Fatalf("the host was asked %d times for anna", n)
	}
	if said, err := w.s.EventsOfTypes(w.ctx, []string{core.EventHostSaid}); err != nil || len(said) != 0 {
		t.Fatalf("no words filed for anna: %v %d", err, len(said))
	}
	if code := w.call("anna", "", "GET", "/api/objects?type=note", nil, nil); code != 403 {
		t.Fatalf("anna reads no company's notes: %d", code)
	}

	var spoken []struct {
		Company string `json:"company"`
		Said    string `json:"said"`
	}
	if code := w.call("", "", "POST", "/api/speak", nil, &spoken); code != 200 {
		t.Fatalf("the alarm: %d", code)
	}
	heard := map[string]bool{}
	for _, sp := range spoken {
		heard[sp.Company] = sp.Said != ""
	}
	if !heard[a] || !heard[b] {
		t.Fatalf("the alarm speaks for every company: %+v", spoken)
	}
}

// A company someone held is never free to claim (review 2026-10-21, A4):
// a company counted as unclaimed when no membership or mandate held it
// today, and an unclaimed company opened every door to anyone who named
// it — a client whose office's mandate ended yesterday, or whose owner
// starts next month, was anyone's to make themselves owner of. Now a
// company anyone was ever granted into is claimed; while a company is
// truly unclaimed only the claim itself is open; and the key is read once,
// trimmed, so "PL…1111 " is the company it names.
func TestALapsedCompanyIsNoOnesToClaim(t *testing.T) {
	w := newGrantsWorld(t)
	office, lapsed, later, held, fresh, own := "PL5250001111", "PL5260001246", "PL7010001234", "PL8130000001", "PL9990000009", "PL1110000001"
	w.person("Krzysztof Kowal", "kk", office, "owner")
	w.mustDoor("", lapsed, "grant_mandate", map[string]any{"office": office, "client": lapsed, "scope": "type:* view:*", "from": "2026-01-01", "to": w.dayFrom(-1), "granted_by": "operator"})
	w.mustDoor("", lapsed, "take_note", map[string]any{"text": "the client's note", "amount": "12.50"})
	olga := w.person("Olga Nowa", "olga", "", "owner")
	w.mustDoor("", later, "grant_membership", map[string]any{"person": olga, "company": later, "role": "owner", "from": w.dayFrom(30), "granted_by": "operator"})
	w.person("Ewa Owner", "ewa", held, "owner")
	mallory := w.person("Mallory Mal", "mallory", own, "owner")

	claim := func(company string) int {
		t.Helper()
		code, _ := w.door("mallory", company, "grant_membership", map[string]any{"person": mallory, "company": company, "role": "owner", "from": w.today, "granted_by": "mallory"})
		return code
	}
	if code := claim(lapsed); code != 403 {
		t.Fatalf("claiming a client whose mandate ended yesterday: %d", code)
	}
	if code := claim(later); code != 403 {
		t.Fatalf("claiming a company whose owner starts next month: %d", code)
	}
	if code := w.call("mallory", held+" ", "GET", "/api/whoami", nil, nil); code != 403 {
		t.Fatalf("a held company named with a trailing space: %d", code)
	}

	// A company nobody ever held opens to its claim and to nothing else
	// until then.
	if code, out := w.door("mallory", fresh, "take_note", map[string]any{"text": "before the claim", "amount": "1.00"}); code != 403 {
		t.Fatalf("a door before the claim: %d %v", code, out)
	}
	if code := claim(fresh); code != 200 {
		t.Fatalf("the claim: %d", code)
	}
	w.mustDoor("mallory", fresh, "take_note", map[string]any{"text": "after the claim", "amount": "1.00"})
}

// A grant reaches only the company it is made in (review 2026-10-21, A2):
// the door checked only inputs called company and office against the
// company acted for, while grants are placed by what the payload says — a
// stranger acting for her own company granted her office a mandate over
// any client she named, and a substitution naming no company handed her
// every grant its person held, in every company. Now a mandate's client is
// the company acted for, and a door that takes a company and was given
// none acts in that one.
func TestAGrantReachesOnlyItsCompany(t *testing.T) {
	w := newGrantsWorld(t)
	held, own := "PL5250001111", "PL1110000001"
	kk := w.person("Krzysztof Kowal", "kk", held, "owner")
	mallory := w.person("Mallory Mal", "mallory", own, "owner")
	companies := func(handle string) []string {
		t.Helper()
		var who struct {
			Companies []string `json:"companies"`
		}
		if code := w.call(handle, "", "GET", "/api/whoami", nil, &who); code != 200 {
			t.Fatalf("whoami %s: %d", handle, code)
		}
		return who.Companies
	}

	for _, client := range []string{held, held + "," + own} {
		if code, out := w.door("mallory", own, "grant_mandate", map[string]any{"office": own, "client": client, "scope": "*", "from": w.today, "granted_by": "mallory"}); code != 403 {
			t.Fatalf("her office's mandate over %q: %d %v", client, code, out)
		}
	}
	w.mustDoor("mallory", own, "appoint_substitute", map[string]any{"person": kk, "deputy": mallory, "from": w.today, "to": w.dayFrom(30), "granted_by": "mallory"})
	if got := companies("mallory"); len(got) != 1 || got[0] != own {
		t.Fatalf("mallory reaches only her own company: %v", got)
	}

	// The owner's substitution naming no company holds where he made it.
	piotr := w.person("Piotr Zastępca", "piotr", "", "accountant")
	w.mustDoor("kk", held, "appoint_substitute", map[string]any{"person": kk, "deputy": piotr, "from": w.today, "to": w.dayFrom(30), "granted_by": "kk"})
	if got := companies("piotr"); len(got) != 1 || got[0] != held {
		t.Fatalf("piotr stands in for kk at %s: %v", held, got)
	}
}

// A person is changed only by someone who may act for them everywhere they
// are let in (review 2026-10-21, A3): the installation's people doors asked
// for the right in any one company and acted on every person — the owner
// of a company claimed a minute ago marked KK as left, for good, rewrote a
// sole trader's name and birth date, and registered a second "kk" whom
// KK's key then signed in as. Now a door acting on a person needs its
// right in every company the person is let into, a handle is one person's,
// and a duplicate the operator's tools write never shadows the first.
func TestAPersonIsChangedOnlyWhereTheyAreLetIn(t *testing.T) {
	w := newGrantsWorld(t)
	held, own := "PL5250001111", "PL1110000001"
	kk := w.person("Krzysztof Kowal", "kk", held, "owner")
	mallory := w.person("Mallory Mal", "mallory", own, "owner")
	whoami := func(handle string) (int, string) {
		t.Helper()
		var who struct {
			PersonID string `json:"person_id"`
		}
		code := w.call(handle, "", "GET", "/api/whoami", nil, &who)
		return code, who.PersonID
	}

	for _, d := range []struct {
		name   string
		inputs map[string]any
	}{
		{"set_person_status", map[string]any{"person": kk, "status": "left"}},
		{"identify_person", map[string]any{"person": kk, "given_name": "Krzysztof", "family_name": "Mallory", "birth_date": "1970-01-01"}},
		{"appoint_deputy", map[string]any{"person": kk, "deputy": mallory}},
		{"register_person", map[string]any{"name": "Krzysztof Kowal", "handle": "kk", "role": "owner"}},
	} {
		if code, out := w.door("mallory", "", d.name, d.inputs); code != 403 {
			t.Fatalf("mallory at %s: %d %v", d.name, code, out)
		}
	}
	if code, id := whoami("kk"); code != 200 || id != kk {
		t.Fatalf("kk still signs in as himself: %d %q", code, id)
	}
	w.mustDoor("", "", "register_person", map[string]any{"name": "Krzysztof Kopia", "handle": "kk", "role": "owner"})
	if code, id := whoami("kk"); code != 200 || id != kk {
		t.Fatalf("a duplicate the operator wrote shadows kk: %d %q", code, id)
	}

	// The owner changes a person let in only at his company — and not one
	// let in elsewhere as well.
	anna := w.person("Anna Nowak", "anna", held, "accountant")
	w.mustDoor("kk", "", "identify_person", map[string]any{"person": anna, "given_name": "Anna", "family_name": "Nowak", "birth_date": "1990-05-04"})
	w.mustDoor("", own, "grant_membership", map[string]any{"person": anna, "company": own, "role": "accountant", "from": "2026-01-01", "granted_by": "operator"})
	if code, out := w.door("kk", "", "set_person_status", map[string]any{"person": anna, "status": "left"}); code != 403 {
		t.Fatalf("kk marking anna, let in at mallory's too, as left: %d %v", code, out)
	}
}

// What runs on every company's books is approved by someone who may act in
// each (review 2026-10-21, A5): rules, doors and bundles belong to no
// company — one approval activates them for every company's facts — yet
// approving was checked in the company acted for, so the owner of one
// company activated what then booked another's. Now approving, rejecting
// and backfilling need the right in every company anyone is let into; and
// keeping an answer as a report, which drafts a bundle, needs the right to
// ask for one.
func TestWhatRunsEverywhereIsApprovedEverywhere(t *testing.T) {
	w := newGrantsWorld(t)
	a, b := "PL5250001111", "PL5260001246"
	kk := w.person("Krzysztof Kowal", "kk", a, "owner")
	w.person("Olga Nowa", "olga", b, "owner")
	w.person("Ewa Reader", "ewa", a, "reader")
	if _, err := w.s.InsertRuleVersion(w.ctx, core.Rule{ID: "memo", Status: core.StatusDraft, Priority: 100, EffectiveFrom: "2026-01-01",
		CreatedBy: "agent:test", Description: "memos as notes", Spec: core.RuleSpec{Match: core.Match{EventType: "memo.taken"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text", "amount": "=$.amount"}}}}}); err != nil {
		t.Fatal(err)
	}

	if code := w.call("olga", "", "POST", "/api/rules/memo/approve", map[string]any{}, nil); code != 403 {
		t.Fatalf("olga approving, for a's books too: %d", code)
	}
	if code := w.call("kk", "", "POST", "/api/rules/memo/approve", map[string]any{}, nil); code != 403 {
		t.Fatalf("kk approving, for b's books too: %d", code)
	}
	if r, err := w.s.GetRule(w.ctx, "memo"); err != nil || r.Status != core.StatusDraft {
		t.Fatalf("the draft stays a draft: %v %v", err, r.Status)
	}
	w.mustDoor("olga", b, "grant_membership", map[string]any{"person": kk, "company": b, "role": "owner", "from": "2026-01-01", "granted_by": "olga"})
	if code := w.call("kk", a, "POST", "/api/rules/memo/approve", map[string]any{}, nil); code != 200 {
		t.Fatalf("kk, let in at both, approving: %d", code)
	}

	// An answer kept as a report is a draft for the gate: asked for by
	// someone who may ask for one, never by a reader.
	answered, err := w.x.TriggerActivity(store.WithCompany(w.ctx, a), "host_answered", map[string]any{"question": "How many notes?",
		"sql": "SELECT COUNT(*) AS notes FROM objects WHERE object_type = 'note'", "reading": "One note.",
		"visual": `{"kind":"table","title":"Notes"}`, "result": `{"columns":["notes"],"rows":[[1]]}`, "rows": 1, "asked_by": "ewa", "occurred_at": w.today},
		"agent:test", "")
	if err != nil {
		t.Fatal(err)
	}
	keep := fmt.Sprintf("/api/answers/%d/keep", answered.EventID)
	if code := w.call("ewa", a, "POST", keep, nil, nil); code != 403 {
		t.Fatalf("a reader keeping an answer: %d", code)
	}
	if code := w.call("kk", a, "POST", keep, nil, nil); code != 200 {
		t.Fatalf("the owner keeping it: %d", code)
	}
}

// Acting for one client reaches that client alone (review 2026-10-21, B1):
// an office's person acting for client A ran with the whole desk in view —
// every client of the office — and a door or a read naming an object by
// its id reached client B's rows on rights held in A: B's note read, and
// struck, by someone whose mandate from B lets them only read. Now acting
// for a company sees and changes that company and the installation's rows;
// the desk acting for none still reads every client it may.
func TestActingForAClientReachesThatClientAlone(t *testing.T) {
	w := newGrantsWorld(t)
	office, a, b := "PL5250001111", "PL5260001246", "PL7010001234"
	w.person("Krzysztof Kowal", "kk", office, "owner")
	w.person("Alfa Owner", "alfa", a, "owner")
	w.person("Beta Owner", "beta", b, "owner")
	w.person("Anna Nowak", "anna", office, "accountant")
	w.mustDoor("", a, "grant_mandate", map[string]any{"office": office, "client": a, "scope": "door:* view:* type:*", "from": "2026-01-01", "granted_by": "operator"})
	w.mustDoor("", b, "grant_mandate", map[string]any{"office": office, "client": b, "scope": "view:* type:*", "from": "2026-01-01", "granted_by": "operator"})
	if _, err := w.s.InsertActivityVersion(w.ctx, core.Activity{Name: "strike_note", Status: core.StatusActive, Domain: "work", Description: "strike a note",
		Spec: core.ActivitySpec{Inputs: []core.FieldDef{{Name: "note", Type: "ref<note>", Required: true}}, Emits: "note.struck", Who: []string{"human"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := w.s.InsertRuleVersion(w.ctx, core.Rule{ID: "strike-note", Status: core.StatusActive, Priority: 100, EffectiveFrom: "2026-01-01",
		CreatedBy: "test", Description: "a struck note", Spec: core.RuleSpec{Match: core.Match{EventType: "note.struck"},
			Effect: core.Effect{Amend: &core.AmendTemplate{Type: "note", Target: "=$.note", Set: map[string]string{"text": "struck"}}}}}); err != nil {
		t.Fatal(err)
	}
	note := func(company, text string) string {
		t.Helper()
		w.mustDoor("", company, "take_note", map[string]any{"text": text, "amount": "1.00"})
		ids, err := w.s.FindObjectIDsByField(w.ctx, "note", "text", text)
		if err != nil || len(ids) != 1 {
			t.Fatalf("%s: %v %v", text, ids, err)
		}
		return ids[0]
	}
	noteA, noteB := note(a, "alfa's note"), note(b, "beta's note")

	if code := w.call("anna", "", "GET", "/api/objects/"+noteB, nil, nil); code != 200 {
		t.Fatalf("anna's desk reads beta's note under beta's mandate: %d", code)
	}
	if code := w.call("anna", a, "GET", "/api/objects/"+noteB, nil, nil); code != 404 {
		t.Fatalf("anna acting for alfa reads beta's note: %d", code)
	}
	if code, out := w.door("anna", a, "strike_note", map[string]any{"note": noteB}); code == 200 {
		t.Fatalf("anna acting for alfa struck beta's note: %v", out)
	}
	if o, err := w.s.GetObject(w.ctx, noteB); err != nil || o.State["text"] != "beta's note" {
		t.Fatalf("beta's note untouched: %v %v", err, o.State)
	}
	w.mustDoor("anna", a, "strike_note", map[string]any{"note": noteA})
	if o, err := w.s.GetObject(w.ctx, noteA); err != nil || o.State["text"] != "struck" {
		t.Fatalf("alfa's note struck: %v %v", err, o.State)
	}
}

// The analysis copy is the company's it is read for (review 2026-10-21,
// B2): the copy's file and its freshness followed the scope alone while
// its content followed the company acted for — a desk that read client
// alfa's figures and then, with nothing written between, client beta's got
// alfa's again, and the host would have filed them as beta's. Now a desk
// acting for one company reads a copy of that company's own.
func TestTheCopyIsTheCompanyActedFor(t *testing.T) {
	w := newGrantsWorld(t)
	a, b := "PL5250001111", "PL5260001246"
	w.mustDoor("", a, "take_note", map[string]any{"text": "alfa 1", "amount": "12.50"})
	w.mustDoor("", a, "take_note", map[string]any{"text": "alfa 2", "amount": "12.50"})
	w.mustDoor("", b, "take_note", map[string]any{"text": "beta 1", "amount": "99.00"})
	read := func(acting string) string {
		t.Helper()
		req, _ := http.NewRequest("GET", w.ts.URL+"/api/views/notes-today", nil)
		req.Header.Set("X-Rhea-Scope", a+","+b)
		if acting != "" {
			req.Header.Set("X-Rhea-Company", acting)
		}
		res, err := w.ts.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		var view struct {
			Rows [][]struct {
				V string `json:"v"`
			} `json:"rows"`
		}
		if json.NewDecoder(res.Body).Decode(&view); res.StatusCode != 200 || len(view.Rows) != 1 {
			t.Fatalf("the desk acting for %q: %d %+v", acting, res.StatusCode, view)
		}
		return view.Rows[0][0].V + " notes, " + view.Rows[0][1].V
	}
	if got := read(a); got != "2 notes, 25.00" {
		t.Fatalf("alfa's figures: %s", got)
	}
	if got := read(b); got != "1 notes, 99.00" {
		t.Fatalf("beta's figures, read after alfa's: %s", got)
	}
	if got := read(""); got != "3 notes, 124.00" {
		t.Fatalf("the desk's figures: %s", got)
	}
}

// A fact with a door of its own comes through the open door only from
// someone that door admits (review 2026-10-21, C1): the open door refused
// the kernel's own namespaces and nothing else, so an accountant — whose
// door:* reaches it but never the people doors — submitted
// membership.granted and owned another company's books, moved the clock
// every company reads to 2099, and forged a reading whose confirmation
// booked a raw match. Now a type some door or channel declares is checked
// as if through that door, a confirmation turns only a channel's proposal
// into only what the channel declares, and a fact nothing declares still
// comes in.
func TestTheOpenDoorIsNoSideDoor(t *testing.T) {
	w := newGrantsWorld(t)
	a, b := "PL5250001111", "PL5260001246"
	kk := w.person("Krzysztof Kowal", "kk", a, "owner")
	w.person("Olga Nowa", "olga", b, "owner")
	anna := w.person("Anna Nowak", "anna", a, "accountant")
	submit := func(person, typ string, payload map[string]any) (int, map[string]any) {
		t.Helper()
		return w.door(person, a, "submit_event", map[string]any{"event_type": typ, "occurred_at": w.today, "payload": payload})
	}

	if code, out := submit("anna", "membership.granted", map[string]any{"person": anna, "company": b, "role": "owner", "from": w.today, "granted_by": "anna"}); code != 403 {
		t.Fatalf("anna granting herself beta through the open door: %d %v", code, out)
	}
	if code, out := submit("anna", "time.day_opened", map[string]any{"date": "2099-12-31"}); code != 403 {
		t.Fatalf("anna moving the clock: %d %v", code, out)
	}
	code, out := submit("anna", "document.reading_proposed", map[string]any{"channel": "pdf_invoice", "event_type": "objects.matched",
		"payload": map[string]any{"matching": "pl-match-incoming", "items": []string{"sales_invoice-1"}}, "document": "x"})
	if code == 200 {
		proposal, _ := out["event_id"].(float64)
		if code, out := w.door("anna", a, "confirm_reading", map[string]any{"proposal": int64(proposal)}); code == 200 {
			t.Fatalf("a forged reading confirmed: %v", out)
		}
	}
	if matched, _ := w.s.EventsOfTypes(w.ctx, []string{core.EventObjectsMatched}); len(matched) != 0 {
		t.Fatalf("a raw match was booked: %+v", matched)
	}

	// The door still opens: a fact nothing declares, and a grant by the
	// owner whose role opens grant_membership.
	if code, out := submit("anna", "visit.logged", map[string]any{"who": "a supplier"}); code != 200 {
		t.Fatalf("a new fact through the open door: %d %v", code, out)
	}
	if code, out := submit("kk", "membership.granted", map[string]any{"person": kk, "company": a, "role": "reader", "from": w.today, "granted_by": "kk"}); code != 200 {
		t.Fatalf("the owner's grant through the open door: %d %v", code, out)
	}
}

// A key a caller chooses names what the caller did, never a channel's fact
// (review 2026-10-21, C2): a door took the caller's dedup key as it came,
// so a person who sent the key a bank statement's line will carry next year
// pre-claimed it — the real line was then skipped as already logged. Now a
// caller's key lives under its door's name, where no reader keys anything.
func TestACallersKeyClaimsNoChannelsFact(t *testing.T) {
	w := newGrantsWorld(t)
	a := "PL5250001111"
	w.person("Anna Nowak", "anna", a, "accountant")
	key := "bank_mt940/PL61109010140000071219812874/2027/38/1/1"
	var out map[string]any
	if code := w.call("anna", a, "POST", "/api/activities/take_note/trigger", map[string]any{
		"inputs": map[string]any{"text": "a note", "amount": "1.00"}, "dedup_key": key}, &out); code != 200 {
		t.Fatalf("anna's note: %d %v", code, out)
	}
	line, _ := json.Marshal(map[string]any{"date": "2027-01-05", "amount": "99.00"})
	if _, err := w.s.AppendEvent(store.WithCompany(w.ctx, a), core.Event{Kind: core.KindRaw, Type: "bank.statement.line",
		OccurredAt: "2027-01-05", Payload: line, DedupKey: key, Company: a}); err != nil {
		t.Fatalf("the bank line next year: %v", err)
	}
	// The caller's key still makes the caller's act idempotent.
	if code := w.call("anna", a, "POST", "/api/activities/take_note/trigger", map[string]any{
		"inputs": map[string]any{"text": "a note", "amount": "1.00"}, "dedup_key": key}, &out); code == 200 {
		t.Fatalf("the same act twice under one key: %v", out)
	}
}
