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
	"strings"
	"testing"

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

// The sign-in step at the shell (DIRECTION 2026-10-19, "People and rights"):
// the face names the person, the kernel decides. A person's grants are
// their scope; a door they hold no right to is refused and the refusal is
// a fact; an office reaches a client through a mandate and nothing else;
// a company nobody holds is claimed by its first grant; the kernel signs
// every "…_by" input with the person; and the who-can-do-what view opens
// every right to its grant. Replay keeps the refusals.
func TestPersonsAtTheDoors(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	srv := &shell.Server{Store: s, Exec: x, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb")}
	x.Today = srv.Clock.TodayUTC
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	today := srv.Clock.TodayUTC()

	call := func(person, company, method, path string, body any, out any) int {
		t.Helper()
		var rd io.Reader
		if body != nil {
			b, _ := json.Marshal(body)
			rd = bytes.NewReader(b)
		}
		req, _ := http.NewRequest(method, ts.URL+path, rd)
		req.Header.Set("Content-Type", "application/json")
		if person != "" {
			req.Header.Set("X-Rhea-Person", person)
		}
		if company != "" {
			req.Header.Set("X-Rhea-Company", company)
		}
		res, err := ts.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if out != nil {
			json.NewDecoder(res.Body).Decode(out)
		}
		return res.StatusCode
	}
	door := func(person, company, name string, inputs map[string]any) (int, map[string]any) {
		t.Helper()
		var out map[string]any
		code := call(person, company, "POST", "/api/activities/"+name+"/trigger", map[string]any{"inputs": inputs}, &out)
		return code, out
	}
	mustDoor := func(person, company, name string, inputs map[string]any) map[string]any {
		t.Helper()
		code, out := door(person, company, name, inputs)
		if code != 200 || out["booked"] != float64(1) {
			t.Fatalf("%s at %s for %s: %d %v", person, name, company, code, out)
		}
		return out
	}

	// The pack and the first owner come from the operator's unscoped tools.
	sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", "common", "people.json"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "operator", today); err != nil || len(errs) != 0 {
		t.Fatalf("approve: %v %v", err, errs)
	}
	if err := s.InsertObjectType(ctx, core.ObjectType{Name: "note", Version: 1, Domain: "work", LabelField: "text",
		Fields: []core.FieldDef{{Name: "text", Type: "string", Required: true}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: "note", Status: core.StatusActive, Priority: 100, EffectiveFrom: "2026-01-01",
		CreatedBy: "test", Description: "notes", Spec: core.RuleSpec{Match: core.Match{EventType: "note.taken"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text"}}}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.InsertActivityVersion(ctx, core.Activity{Name: "take_note", Status: core.StatusActive, Domain: "work", Description: "note",
		Spec: core.ActivitySpec{Inputs: []core.FieldDef{{Name: "text", Type: "string", Required: true}}, Emits: "note.taken", Who: []string{"human"}}}); err != nil {
		t.Fatal(err)
	}
	office, client, stranger := "PL5250001111", "PL5260001246", "PL7010001234"
	mustDoor("", "", "register_person", map[string]any{"name": "Krzysztof Kowal", "handle": "kk", "role": "owner"})
	personID := func(handle string) string {
		ids, err := s.FindObjectIDsByField(ctx, "person", "handle", handle)
		if err != nil || len(ids) != 1 {
			t.Fatalf("%s: %v %v", handle, ids, err)
		}
		return ids[0]
	}
	mustDoor("", office, "grant_membership", map[string]any{"person": personID("kk"), "company": office, "role": "owner", "from": "2026-01-01", "granted_by": "operator"})

	// A stranger is no principal; a person's scope is their grants.
	if code := call("nobody", "", "GET", "/api/whoami", nil, nil); code != 403 {
		t.Fatalf("a stranger: %d", code)
	}
	var who struct {
		Person    string   `json:"person"`
		Companies []string `json:"companies"`
		Company   string   `json:"company"`
	}
	if code := call("kk", "", "GET", "/api/whoami", nil, &who); code != 200 || who.Person != "kk" || len(who.Companies) != 1 || who.Company != office {
		t.Fatalf("kk: %d %+v", code, who)
	}

	// The owner invites an accountant: the kernel signs the grant with the
	// owner, whatever the page sent; the accountant's seat is visible at once.
	mustDoor("kk", "", "register_person", map[string]any{"name": "Anna Nowak", "handle": "anna", "role": "accountant"})
	grant := mustDoor("kk", office, "grant_membership", map[string]any{"person": personID("anna"), "company": office, "role": "accountant", "from": today, "granted_by": "mallory"})
	ev, err := s.GetEvent(ctx, int64(grant["event_id"].(float64)))
	var signed struct {
		GrantedBy string `json:"granted_by"`
	}
	json.Unmarshal(ev.Payload, &signed)
	if err != nil || ev.Actor != "kk" || signed.GrantedBy != "kk" || ev.Company != office {
		t.Fatalf("the grant is signed by the owner: %v actor %q company %q %s", err, ev.Actor, ev.Company, ev.Payload)
	}
	if ev, _ := s.GetEvent(ctx, int64(mustDoor("kk", "", "identify_person", map[string]any{"person": personID("anna"), "given_name": "Anna", "family_name": "Nowak", "birth_date": "1990-05-04"})["event_id"].(float64))); ev.Company != "" {
		t.Fatalf("a person's own facts are the installation's: %q", ev.Company)
	}

	// The accountant keeps the books and is refused at the people doors;
	// the refusal is a fact under the office, signed by her.
	mustDoor("anna", "", "take_note", map[string]any{"text": "the office's note"})
	code, out := door("anna", "", "grant_membership", map[string]any{"person": personID("kk"), "company": office, "role": "reader", "from": today, "granted_by": "anna"})
	if code != 403 || !strings.Contains(fmt.Sprint(out["error"]), "no role of yours opens this door") {
		t.Fatalf("anna granting: %d %v", code, out)
	}
	refusals, err := s.EventsOfTypes(ctx, []string{core.EventAccessRefused})
	if err != nil || len(refusals) != 1 || refusals[0].Actor != "kernel" || refusals[0].Company != office ||
		!strings.Contains(string(refusals[0].Payload), "grant_membership") || !strings.Contains(string(refusals[0].Payload), `"anna"`) {
		t.Fatalf("the refusal recorded: %v %+v", err, refusals)
	}
	// The machine's doors are nobody's: the clock is not a person's to move.
	if code, out := door("kk", "", "clock", map[string]any{}); code != 403 || !strings.Contains(fmt.Sprint(out["error"]), "not a person's") {
		t.Fatalf("kk at the clock: %d %v", code, out)
	}
	// Who may open which door, said on the list.
	var acts []struct {
		Name    string `json:"name"`
		Allowed *bool  `json:"allowed"`
	}
	call("anna", "", "GET", "/api/activities", nil, &acts)
	allowed := map[string]bool{}
	for _, a := range acts {
		if a.Allowed != nil {
			allowed[a.Name] = *a.Allowed
		}
	}
	if !allowed["take_note"] || allowed["grant_membership"] || allowed["clock"] {
		t.Fatalf("anna's doors: %v", allowed)
	}

	// Reads: the office's notes are hers; a company she holds nothing in is
	// refused, and the refusal of the company is a fact too.
	var objs struct {
		Objects []map[string]any `json:"objects"`
	}
	if code := call("anna", "", "GET", "/api/objects?type=note", nil, &objs); code != 200 || len(objs.Objects) != 1 {
		t.Fatalf("anna's notes: %d %+v", code, objs)
	}
	mustDoor("", client, "grant_membership", map[string]any{"person": personID("kk"), "company": client, "role": "owner", "from": "2026-01-01", "granted_by": "operator"})
	if code := call("anna", client, "GET", "/api/objects?type=note", nil, nil); code != 403 {
		t.Fatalf("anna at the client: %d", code)
	}
	// Three refusals so far: anna at the people door, kk at the clock, and
	// anna at the client's company.
	if refusals, _ = s.EventsOfTypes(ctx, []string{core.EventAccessRefused}); len(refusals) != 3 || refusals[2].Company != client || !strings.Contains(string(refusals[2].Payload), "reaches this company") {
		t.Fatalf("the company refusal recorded: %d %s", len(refusals), refusals[len(refusals)-1].Payload)
	}
	// A reader sees and does nothing.
	mustDoor("", "", "register_person", map[string]any{"name": "Ewa Reader", "handle": "ewa", "role": "reader"})
	mustDoor("kk", office, "grant_membership", map[string]any{"person": personID("ewa"), "company": office, "role": "reader", "from": today, "granted_by": "kk"})
	if code := call("ewa", "", "GET", "/api/objects?type=note", nil, &objs); code != 200 || len(objs.Objects) != 1 {
		t.Fatalf("ewa reads: %d", code)
	}
	if code, _ := door("ewa", "", "take_note", map[string]any{"text": "ewa's note"}); code != 403 {
		t.Fatalf("ewa writes: %d", code)
	}
	if code := call("ewa", "", "GET", "/api/worklist", nil, nil); code != 200 {
		t.Fatalf("ewa's worklist: %d", code)
	}

	// The office keeps the client's books under a mandate: kk, the client's
	// owner here, grants it; anna reaches the client's notes through it
	// and no further than its scope.
	mustDoor("kk", client, "grant_mandate", map[string]any{"office": office, "client": client, "scope": "type:* view:*", "from": today, "granted_by": "kk"})
	mustDoor("kk", client, "take_note", map[string]any{"text": "the client's note"})
	if code := call("anna", client, "GET", "/api/objects?type=note", nil, &objs); code != 200 || len(objs.Objects) != 1 || objs.Objects[0]["state"].(map[string]any)["text"] != "the client's note" {
		t.Fatalf("anna reads the client under the mandate: %d %+v", code, objs)
	}
	if code, _ := door("anna", client, "take_note", map[string]any{"text": "anna's note at the client"}); code != 403 {
		t.Fatalf("the mandate's scope stops at reading: %d", code)
	}
	call("anna", "", "GET", "/api/whoami", nil, &who)
	if len(who.Companies) != 2 || who.Company != "" {
		t.Fatalf("anna's desk: %+v", who)
	}
	// On the desk open on both, she sees notes where she may in every
	// company: both here. The acting company is required to write.
	if code := call("anna", "", "GET", "/api/objects?type=note", nil, &objs); code != 200 || len(objs.Objects) != 2 {
		t.Fatalf("anna's desk shows both: %d %d", code, len(objs.Objects))
	}
	if code, out := door("anna", "", "take_note", map[string]any{"text": "where?"}); code != 403 || !strings.Contains(fmt.Sprint(out["error"]), "name the company") {
		t.Fatalf("a write names its company: %d %v", code, out)
	}

	// A company nobody holds is claimed by its first grant: a new owner
	// grants herself; the second claimant is refused.
	mustDoor("", "", "register_person", map[string]any{"name": "Olga Nowa", "handle": "olga", "role": "owner"})
	mustDoor("olga", stranger, "grant_membership", map[string]any{"person": personID("olga"), "company": stranger, "role": "owner", "from": today, "granted_by": "olga"})
	if code, _ := door("anna", stranger, "grant_membership", map[string]any{"person": personID("anna"), "company": stranger, "role": "owner", "from": today, "granted_by": "anna"}); code != 403 {
		t.Fatalf("a claimed company is not claimed twice: %d", code)
	}
	// A grant naming one company while acting for another is refused.
	if code, out := door("kk", office, "grant_membership", map[string]any{"person": personID("ewa"), "company": client, "role": "reader", "from": today, "granted_by": "kk"}); code != 403 || !strings.Contains(fmt.Sprint(out["error"]), "act for") {
		t.Fatalf("naming another company: %d %v", code, out)
	}

	// Who can do what, every cell a grant with its reason.
	var table struct {
		Columns []struct{ Kind, Name string } `json:"columns"`
		Rows    []struct {
			Handle string                    `json:"handle"`
			Cells  map[string]map[string]any `json:"cells"`
		} `json:"rows"`
	}
	if code := call("ewa", office, "GET", "/api/views/who-can-do-what", nil, &table); code != 200 {
		t.Fatalf("ewa reads who can do what: %d", code)
	}
	if len(table.Rows) != 3 || table.Rows[0].Handle != "anna" || table.Rows[1].Handle != "ewa" || table.Rows[2].Handle != "kk" {
		t.Fatalf("rows: %+v", table.Rows)
	}
	cell := table.Rows[0].Cells["door:take_note"]
	if cell["via"] != "membership" || cell["granted_by"] != "kk" || cell["role"] != "accountant" || cell["object"] != "membership-"+fmt.Sprint(int64(grant["event_id"].(float64))) {
		t.Fatalf("anna's right to take_note opens to kk's grant: %v", cell)
	}
	if _, ok := table.Rows[0].Cells["door:grant_membership"]; ok {
		t.Fatal("anna has no cell at the people doors")
	}
	if _, ok := table.Rows[1].Cells["door:take_note"]; ok {
		t.Fatal("ewa reads only")
	}
	call("anna", client, "GET", "/api/rights?company="+client, nil, &table)
	found := false
	for _, r := range table.Rows {
		if r.Handle == "anna" {
			found = cell["via"] != "" && r.Cells["type:note"]["via"] == "mandate"
		}
	}
	if !found {
		t.Fatalf("anna's reach into the client opens to the mandate: %+v", table.Rows)
	}

	// Replay keeps who-may-do-what and the refusals (invariant 4).
	before, _ := s.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := s.AllObjects(ctx)
	if diffs := exec.StateDiff(before, after); len(diffs) != 0 {
		t.Fatalf("replay differs: %v", diffs)
	}
	if refusals, _ = s.EventsOfTypes(ctx, []string{core.EventAccessRefused}); len(refusals) < 3 {
		t.Fatal("the refusals are in the log")
	}
	if code := call("anna", client, "GET", "/api/objects?type=note", nil, &objs); code != 200 || len(objs.Objects) != 1 {
		t.Fatalf("after replay anna still reads the client: %d", code)
	}
}

// A signed-in person is always the one checked and the one who signs
// (review 2026-10-21, A1): whether the kernel asked about rights at all
// turned on the actor string, and the caller wrote it — a reader posting
// {"approved_by":"kernel"} activated a rule for every company, and a
// person registered under the handle "kernel" passed every door unchecked.
// The machine's names are no person's; the host still speaks, as the
// agent, while a person asks.
func TestAPersonIsAlwaysChecked(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	srv := &shell.Server{Store: s, Exec: x, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb"),
		Agent: &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
			return "One note was taken, worth 12.50 PLN.", nil
		}}}
	x.Today = srv.Clock.TodayUTC
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	today := srv.Clock.TodayUTC()

	call := func(person, company, method, path string, body any, out any) int {
		t.Helper()
		var rd io.Reader
		if body != nil {
			b, _ := json.Marshal(body)
			rd = bytes.NewReader(b)
		}
		req, _ := http.NewRequest(method, ts.URL+path, rd)
		req.Header.Set("Content-Type", "application/json")
		if person != "" {
			req.Header.Set("X-Rhea-Person", person)
		}
		if company != "" {
			req.Header.Set("X-Rhea-Company", company)
		}
		res, err := ts.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if out != nil {
			json.NewDecoder(res.Body).Decode(out)
		}
		return res.StatusCode
	}
	mustDoor := func(person, company, name string, inputs map[string]any) {
		t.Helper()
		var out map[string]any
		if code := call(person, company, "POST", "/api/activities/"+name+"/trigger", map[string]any{"inputs": inputs}, &out); code != 200 || out["booked"] != float64(1) {
			t.Fatalf("%s at %s for %s: %d %v", person, name, company, code, out)
		}
	}
	personID := func(handle string) string {
		ids, err := s.FindObjectIDsByField(ctx, "person", "handle", handle)
		if err != nil || len(ids) != 1 {
			t.Fatalf("%s: %v %v", handle, ids, err)
		}
		return ids[0]
	}

	sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", "common", "people.json"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "operator", today); err != nil || len(errs) != 0 {
		t.Fatalf("approve: %v %v", err, errs)
	}
	if err := s.InsertObjectType(ctx, core.ObjectType{Name: "note", Version: 1, Domain: "work", LabelField: "text",
		Fields: []core.FieldDef{{Name: "text", Type: "string", Required: true}, {Name: "amount", Type: "money", Required: true}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: "note", Status: core.StatusActive, Priority: 100, EffectiveFrom: "2026-01-01",
		CreatedBy: "test", Description: "notes", Spec: core.RuleSpec{Match: core.Match{EventType: "note.taken"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text", "amount": "=$.amount"}}}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: "memo", Status: core.StatusDraft, Priority: 100, EffectiveFrom: "2026-01-01",
		CreatedBy: "agent:test", Description: "memos as notes", Spec: core.RuleSpec{Match: core.Match{EventType: "memo.taken"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text", "amount": "=$.amount"}}}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.InsertActivityVersion(ctx, core.Activity{Name: "take_note", Status: core.StatusActive, Domain: "work", Description: "note",
		Spec: core.ActivitySpec{Inputs: []core.FieldDef{{Name: "text", Type: "string", Required: true}, {Name: "amount", Type: "money", Required: true}},
			Emits: "note.taken", Who: []string{"human"}}}); err != nil {
		t.Fatal(err)
	}
	facts, _ := json.Marshal(core.AnalysisSpec{
		SQL:     `SELECT COUNT(*) AS notes, SUM(CAST(json_extract(state, '$.amount') AS BIGINT)) AS total FROM objects WHERE object_type = 'note'`,
		Columns: []core.ColumnSpec{{Field: "notes", Label: "Notes", Type: "int"}, {Field: "total", Label: "Total", Type: "money"}}})
	if err := s.InsertViewDef(ctx, core.ViewDef{ID: "notes-today", Version: 1, Notion: "analysis", Title: "Notes today",
		Domain: "work", Function: "analysis", Spec: facts}); err != nil {
		t.Fatal(err)
	}
	narrative, _ := json.Marshal(core.NarrativeSpec{Facts: []core.NarrativeFact{{View: "notes-today", As: "the notes"}}, Voice: "briefly"})
	if err := s.InsertViewDef(ctx, core.ViewDef{ID: "briefing", Version: 1, Notion: "narrative", Title: "Briefing",
		Domain: "work", Function: "briefing", Spec: narrative}); err != nil {
		t.Fatal(err)
	}

	office := "PL5250001111"
	mustDoor("", "", "register_person", map[string]any{"name": "Krzysztof Kowal", "handle": "kk", "role": "owner"})
	mustDoor("", office, "grant_membership", map[string]any{"person": personID("kk"), "company": office, "role": "owner", "from": "2026-01-01", "granted_by": "operator"})
	mustDoor("", "", "register_person", map[string]any{"name": "Ewa Reader", "handle": "ewa", "role": "reader"})
	mustDoor("kk", office, "grant_membership", map[string]any{"person": personID("ewa"), "company": office, "role": "reader", "from": today, "granted_by": "kk"})
	mustDoor("kk", office, "take_note", map[string]any{"text": "the office's note", "amount": "12.50"})

	// A reader cannot approve by naming the machine, or the agent, as the
	// approver: the door is checked against her, and refused.
	for _, by := range []string{"kernel", "agent:host", "adapter:face"} {
		var out map[string]any
		if code := call("ewa", "", "POST", "/api/rules/memo/approve", map[string]any{"approved_by": by}, &out); code != 403 {
			t.Fatalf("ewa approving as %q: %d %v", by, code, out)
		}
	}
	if r, err := s.GetRule(ctx, "memo"); err != nil || r.Status != core.StatusDraft {
		t.Fatalf("the draft stays a draft: %v %+v", err, r.Status)
	}
	refusals, err := s.EventsOfTypes(ctx, []string{core.EventAccessRefused})
	if err != nil || len(refusals) != 3 || !strings.Contains(string(refusals[0].Payload), `"ewa"`) || !strings.Contains(string(refusals[0].Payload), "approve_rule") {
		t.Fatalf("each refusal names ewa at approve_rule: %v %d", err, len(refusals))
	}

	// The owner may approve, and the approval is hers whatever the body
	// named: the approver is the person signed in.
	if code := call("kk", "", "POST", "/api/rules/memo/approve", map[string]any{"approved_by": "kernel"}, nil); code != 200 {
		t.Fatalf("kk approving: %d", code)
	}
	approvals, err := s.EventsOfTypes(ctx, []string{"rule.approved"})
	var approval struct {
		ApprovedBy string `json:"approved_by"`
	}
	if err == nil && len(approvals) == 1 {
		err = json.Unmarshal(approvals[0].Payload, &approval)
	}
	if err != nil || len(approvals) != 1 || approvals[0].Actor != "kk" || approval.ApprovedBy != "kk" {
		t.Fatalf("the approval is kk's: %v %d %+v", err, len(approvals), approval)
	}

	// The machine's names are no person's: registered under one, nobody
	// signs in with it.
	for _, handle := range []string{"kernel", "agent:host", "adapter:face", "channel:ksef"} {
		mustDoor("", "", "register_person", map[string]any{"name": "Not A Person", "handle": handle, "role": "owner"})
		if code := call(handle, "", "GET", "/api/whoami", nil, nil); code != 403 {
			t.Fatalf("signing in as %q: %d", handle, code)
		}
	}

	// The host still speaks while a person asks: the words are the agent's,
	// under the person's company.
	var spoken []struct {
		Said      string `json:"said"`
		SaidEvent int64  `json:"said_event"`
		Error     string `json:"error"`
		Silent    string `json:"silent"`
	}
	if code := call("kk", "", "POST", "/api/speak?view=briefing", nil, &spoken); code != 200 || len(spoken) != 1 || spoken[0].SaidEvent == 0 {
		t.Fatalf("the host speaks for kk: %d %+v", code, spoken)
	}
	said, err := s.GetEvent(ctx, spoken[0].SaidEvent)
	if err != nil || !strings.HasPrefix(said.Actor, "agent:") || said.Company != office {
		t.Fatalf("the words are the agent's, under the office: %v %q %q", err, said.Actor, said.Company)
	}
}

// A kernel that serves persons only answers 401 to a request naming none —
// except the machine's own routes: the clock, the morning, the face's
// report of a sign-in.
func TestPersonsOnly(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	srv := &shell.Server{Store: s, Exec: x, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb"), RequirePerson: true}
	x.Today = srv.Clock.TodayUTC
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	get := func(path string) int {
		res, err := ts.Client().Get(ts.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		return res.StatusCode
	}
	if code := get("/api/worklist"); code != 401 {
		t.Fatalf("nobody at the worklist: %d", code)
	}
	if code := get("/api/time"); code != 200 {
		t.Fatalf("the clock is read by the machine: %d", code)
	}
	body, _ := json.Marshal(map[string]any{"inputs": map[string]any{"person": "kk", "outcome": "failed", "method": "passkey", "detail": "no key for this device"}, "actor": "adapter:face"})
	res, err := ts.Client().Post(ts.URL+"/api/activities/record_access/trigger", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("the face records a failed sign-in: %d", res.StatusCode)
	}
	evs, err := s.EventsOfTypes(ctx, []string{core.EventAccessRecorded})
	if err != nil || len(evs) != 1 || evs[0].Actor != "adapter:face" || evs[0].Company != "" {
		t.Fatalf("the sign-in fact: %v %+v", err, evs)
	}
	// Reserved for the kernel: the open door does not speak it, and the
	// worklist never asks about it.
	body, _ = json.Marshal(map[string]any{"event_type": core.EventAccessRecorded, "occurred_at": "2026-10-20", "payload": map[string]any{}})
	res, _ = ts.Client().Post(ts.URL+"/api/events", "application/json", bytes.NewReader(body))
	res.Body.Close()
	if res.StatusCode != 401 {
		t.Fatalf("the open door without a person: %d", res.StatusCode)
	}
	if pending, _ := s.PendingEvents(ctx); len(pending) != 0 {
		t.Fatalf("access facts are never a question: %+v", pending)
	}
}
