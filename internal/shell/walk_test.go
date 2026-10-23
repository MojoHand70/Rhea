package shell_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/store"
)

// Every read route, walked by a person of one company over another's books
// (invariant 9; the whole-tree review of 2026-10-21, H2): the two-company
// walk reached the doors, the analysis views, the object list and the
// rights, and never the routes several findings lived on — one object by
// its id, an explanation, what the host said, the files, the rules and the
// bundles, the dry runs, an answer. Now company B keeps one of every kind
// of thing a company keeps, each carrying a mark, and A's owner reads every
// route, naming B's ids where a route takes one: no answer carries the mark
// or one of B's ids. The routes are read from the server's own table, so a
// read route added later fails here until it is walked too.
func TestEveryReadIsTheCompanysOwn(t *testing.T) {
	w := newGrantsWorld(t)
	a, b := "PL5250001111", "PL5260001246"
	const mark = "betasecret"
	w.person("Krzysztof Kowal", "kk", a, "owner")
	w.person("Olga Nowa", "olga", b, "owner")
	// the author answers every question with the company's notes
	w.reply = func(system, user string) string {
		if strings.Contains(system, "reporting author") {
			return `{"sql": "SELECT id, text, amount FROM work.note", "reading": "the notes", "visual": {"kind": "table", "title": "the notes"}}`
		}
		return "One note was taken, worth 12.50 PLN."
	}
	w.mustDoor("kk", a, "take_note", map[string]any{"text": "alfa's note", "amount": "12.50"})

	// B's books: a note, a file received, a fact no rule explains, the
	// host's words, a question she could not answer, an answer kept as a
	// report, and a rule of B's own waiting for approval
	w.mustDoor("olga", b, "take_note", map[string]any{"text": mark + " note", "amount": "12.50"})
	file := []byte("%PDF-1.7 " + mark)
	hash := core.HashBytes(file)
	if code, out := w.door("olga", b, "receive_document", map[string]any{
		"file":   map[string]any{"name": mark + ".pdf", "media_type": "application/pdf", "content_base64": base64.StdEncoding.EncodeToString(file)},
		"source": mark + " mail"}); code != 200 {
		t.Fatalf("b receives a file: %d %v", code, out)
	}
	var raw map[string]any
	if code := w.call("olga", b, "POST", "/api/events", map[string]any{"event_type": mark + ".happened", "occurred_at": w.today,
		"payload": map[string]any{"what": mark}}, &raw); code != 200 {
		t.Fatalf("b's unexplained fact: %d %v", code, raw)
	}
	said, err := w.x.TriggerActivity(store.WithCompany(w.ctx, b), "host_said", map[string]any{"view": "briefing", "text": mark + " words"}, "agent:test", "")
	if err != nil {
		t.Fatal(err)
	}
	if code := w.call("olga", b, "POST", "/api/untaught", map[string]any{"asked": mark + " asked", "missing": mark + " missing",
		"asked_by": "olga", "kind": "cannot"}, nil); code != 200 {
		t.Fatalf("b's untaught question: %d", code)
	}
	var answered struct {
		EventID int64 `json:"event_id"`
	}
	if code := w.call("olga", b, "POST", "/api/answer", map[string]any{"question": mark + " question", "asked_by": "olga"}, &answered); code != 200 || answered.EventID == 0 {
		t.Fatalf("b's answer: %d %+v", code, answered)
	}
	var kept struct {
		ID string `json:"bundle_id"`
	}
	if code := w.call("olga", b, "POST", fmt.Sprintf("/api/answers/%d/keep", answered.EventID), nil, &kept); code != 200 || kept.ID == "" {
		t.Fatalf("b keeps the answer: %d %+v", code, kept)
	}
	rule := core.OwnedID(mark, b)
	if _, err := w.s.InsertRuleVersion(w.ctx, core.Rule{ID: rule, Company: b, Status: core.StatusDraft, Priority: 100, EffectiveFrom: "2026-01-01",
		CreatedBy: "olga", Description: mark + " rule", Spec: core.RuleSpec{Match: core.Match{EventType: "note.taken"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text", "amount": "=$.amount"}}}}}); err != nil {
		t.Fatal(err)
	}

	// B's own ids: an answer naming one is an answer about B
	var objects []string
	note := ""
	for _, typ := range []string{"note", "membership"} {
		objs, err := w.s.ObjectsByType(store.WithCompany(w.ctx, b), typ)
		if err != nil {
			t.Fatal(err)
		}
		for _, o := range objs {
			if o.Company == b {
				objects = append(objects, o.ID)
				if typ == "note" {
					note = o.ID
				}
			}
		}
	}
	if note == "" || len(objects) < 2 {
		t.Fatalf("b's objects: %v", objects)
	}
	secret := append([]string{rule, kept.ID, hash}, objects...)

	var nav []struct {
		Functions []struct {
			Views []struct {
				ID string `json:"view_id"`
			} `json:"views"`
		} `json:"functions"`
	}
	w.call("kk", a, "GET", "/api/nav", nil, &nav)
	var views []string
	for _, d := range nav {
		for _, f := range d.Functions {
			for _, v := range f.Views {
				views = append(views, v.ID)
			}
		}
	}
	if len(views) == 0 {
		t.Fatal("a's nav lists no view to walk")
	}

	type step struct {
		route  string // the server's pattern this step walks
		method string
		path   string
		body   any
	}
	steps := []step{
		{"GET /api/nav", "GET", "/api/nav", nil},
		{"GET /api/worklist", "GET", "/api/worklist", nil},
		{"GET /api/types", "GET", "/api/types", nil},
		{"GET /api/explain", "GET", "/api/explain?object=" + note, nil},
		{"GET /api/explain", "GET", fmt.Sprintf("/api/explain?event=%d", said.EventID), nil},
		{"GET /api/objects", "GET", "/api/objects?type=note", nil},
		{"GET /api/objects", "GET", "/api/objects?type=membership", nil},
		{"GET /api/said", "GET", "/api/said?view=briefing", nil},
		{"GET /api/said", "GET", "/api/said?person=olga", nil},
		{"GET /api/untaught", "GET", "/api/untaught", nil},
		{"GET /api/network/learning", "GET", "/api/network/learning", nil},
		{"GET /api/collection", "GET", "/api/collection", nil},
		{"GET /api/files", "GET", "/api/files", nil},
		{"GET /api/rules", "GET", "/api/rules", nil},
		{"GET /api/promotions", "GET", "/api/promotions", nil},
		{"GET /api/changes", "GET", "/api/changes?since=2026-01-01", nil},
		{"GET /api/packs", "GET", "/api/packs", nil},
		{"GET /api/bundles", "GET", "/api/bundles", nil},
		{"GET /api/bundles/{id}/preview", "GET", "/api/bundles/" + kept.ID + "/preview", nil},
		{"GET /api/activities", "GET", "/api/activities", nil},
		{"GET /api/channels", "GET", "/api/channels", nil},
		{"GET /api/rights", "GET", "/api/rights", nil},
		{"GET /api/rights", "GET", "/api/rights?company=" + b, nil},
		{"GET /api/whoami", "GET", "/api/whoami", nil},
		{"GET /api/registers", "GET", "/api/registers", nil},
		{"GET /api/registers/{name}", "GET", "/api/registers/no-such?id=5260250995", nil}, // public records, and no register reached from a test
		{"GET /api/blobs/{hash}", "GET", "/api/blobs/" + hash, nil},
		{"GET /api/time", "GET", "/api/time", nil},
		{"GET /api/health", "GET", "/api/health", nil},
		{"POST /api/rules/{id}/simulate", "POST", "/api/rules/" + rule + "/simulate", nil},
		{"POST /api/backfill/simulate", "POST", "/api/backfill/simulate", nil},
		{"POST /api/bundles/{id}/simulate", "POST", "/api/bundles/" + kept.ID + "/simulate", nil},
		{"POST /api/time/simulate", "POST", "/api/time/simulate", map[string]any{"through": w.dayFrom(1)}},
		{"POST /api/answer", "POST", "/api/answer", map[string]any{"question": "what notes", "asked_by": "kk"}},
	}
	for _, id := range objects {
		steps = append(steps, step{"GET /api/objects/{id}", "GET", "/api/objects/" + id, nil})
	}
	for _, v := range views {
		steps = append(steps, step{"GET /api/views/{id}", "GET", "/api/views/" + v, nil},
			step{"GET /api/views/{id}", "GET", "/api/views/" + v + "?object_id=" + note, nil},
			step{"GET /api/views/{id}", "GET", "/api/views/" + v + "?company=" + b, nil})
	}
	var writers struct {
		Writers []string `json:"writers"`
	}
	w.call("kk", a, "GET", "/api/files", nil, &writers)
	for _, wr := range writers.Writers {
		steps = append(steps, step{"GET /api/files/{writer}", "GET", "/api/files/" + wr + "?period=" + w.today[:7] + "&company=" + b, nil})
	}
	// the live stream never ends on its own: TestTheLiveStreamIsTheListenersOwn walks it
	walked := map[string]bool{"GET /api/live": true}
	if len(writers.Writers) == 0 {
		walked["GET /api/files/{writer}"] = true // no writer is registered in this binary: nothing to render
	}

	for _, s := range steps {
		walked[s.route] = true
		code, body := walkCall(t, w, "kk", a, s.method, s.path, s.body)
		if code >= 500 {
			t.Errorf("%s %s: %d %s", s.method, s.path, code, body)
		}
		// a refusal may say back what it was asked for, and nothing else
		if code >= 400 {
			for _, id := range secret {
				if strings.Contains(s.path, id) {
					body = strings.ReplaceAll(body, id, "")
				}
			}
		}
		if strings.Contains(body, mark) {
			t.Errorf("%s %s answered a's owner with b's books: %s", s.method, s.path, body)
		}
		for _, id := range secret {
			if strings.Contains(body, id) {
				t.Errorf("%s %s answered a's owner with b's %s: %s", s.method, s.path, id, body)
			}
		}
	}

	// B's owner reads the same routes and finds B's books: what A was not
	// shown is there to be shown
	for _, path := range []string{"/api/objects?type=note", "/api/objects/" + note, "/api/explain?object=" + note, "/api/worklist",
		"/api/said?view=briefing", "/api/untaught", "/api/rules", "/api/bundles", "/api/bundles/" + kept.ID + "/preview", "/api/blobs/" + hash} {
		if code, body := walkCall(t, w, "olga", b, "GET", path, nil); code != 200 || !strings.Contains(body, mark) {
			t.Errorf("b's owner at %s: %d %s", path, code, body)
		}
	}

	// every read route of the server is walked
	src, err := os.ReadFile("shell.go")
	if err != nil {
		t.Fatal(err)
	}
	routes := regexp.MustCompile(`mux\.HandleFunc\("(GET [^"]+)"`).FindAllStringSubmatch(string(src), -1)
	if len(routes) < 20 {
		t.Fatalf("the server's route table read as %d read routes", len(routes))
	}
	for _, m := range routes {
		if !walked[m[1]] {
			t.Errorf("%s is never walked by a person of another company", m[1])
		}
	}
}

// walkCall is one request answered as its bytes, whatever they are.
func walkCall(t *testing.T, w *grantsWorld, person, company, method, path string, body any) (int, string) {
	t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, w.ts.URL+path, rd)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Rhea-Person", person)
	req.Header.Set("X-Rhea-Company", company)
	res, err := w.ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	got, _ := io.ReadAll(res.Body)
	return res.StatusCode, string(got)
}
