package shell_test

import (
	"bytes"
	"context"
	"encoding/json"
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
	if err := w.s.InsertObjectType(w.ctx, core.ObjectType{Name: "note", Version: 1, Domain: "work", LabelField: "text",
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
