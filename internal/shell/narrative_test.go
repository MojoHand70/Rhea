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
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

// The briefing (DIRECTION 2026-10-10, the concierge's first scenario): the
// kernel computes the facts, the host phrases them, the words are logged
// through the host_said door so the next words know the last, and the host
// never computes a number — words carrying one the facts do not are
// refused, and the facts are shown alone with the reason. Without a host,
// the facts speak for themselves.
func TestBriefing(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	var heard string
	words := "Two notes were taken today, worth 1234.56 PLN together."
	var answers []string // when set, the host's next answers in turn
	srv := &shell.Server{Store: s, Exec: x, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb"),
		Agent: &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
			heard = user
			if !strings.Contains(system, "never compute") {
				return "", fmt.Errorf("the host is told its law")
			}
			if len(answers) > 0 {
				a := answers[0]
				answers = answers[1:]
				return a, nil
			}
			return words, nil
		}}}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	get := func(path string, out any) {
		t.Helper()
		res, err := ts.Client().Get(ts.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			var e map[string]string
			json.NewDecoder(res.Body).Decode(&e)
			t.Fatalf("GET %s: %d %s", path, res.StatusCode, e["error"])
		}
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.InsertObjectType(ctx, core.ObjectType{Name: "note", Version: 1, Domain: "work", LabelField: "text",
		Fields: []core.FieldDef{{Name: "text", Type: "string", Required: true}, {Name: "amount", Type: "money", Required: true}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: "note", Status: core.StatusActive, Priority: 100,
		EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: "notes", Spec: core.RuleSpec{
			Match:  core.Match{EventType: "note.taken"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text", "amount": "=$.amount"}}}}}); err != nil {
		t.Fatal(err)
	}
	// The facts: an analysis view the kernel computes.
	facts, _ := json.Marshal(core.AnalysisSpec{
		SQL:     `SELECT COUNT(*) AS notes, SUM(CAST(json_extract(state, '$.amount') AS BIGINT)) AS total FROM objects WHERE object_type = 'note'`,
		Columns: []core.ColumnSpec{{Field: "notes", Label: "Notes", Type: "int"}, {Field: "total", Label: "Total", Type: "money"}}})
	if err := s.InsertViewDef(ctx, core.ViewDef{ID: "notes-today", Version: 1, Notion: "analysis", Title: "Notes today",
		Domain: "work", Function: "analysis", Spec: facts}); err != nil {
		t.Fatal(err)
	}
	narrative, _ := json.Marshal(core.NarrativeSpec{Facts: []core.NarrativeFact{{View: "notes-today", As: "the notes"}},
		Voice: "briefly, in English"})
	if err := s.InsertViewDef(ctx, core.ViewDef{ID: "briefing", Version: 1, Notion: "narrative", Title: "Briefing",
		Domain: "work", Function: "briefing", Spec: narrative}); err != nil {
		t.Fatal(err)
	}
	for i, amount := range []string{"1000.00", "234.56"} {
		if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: "note.taken", OccurredAt: "2026-09-01",
			Payload: []byte(fmt.Sprintf(`{"text":"note %d","amount":"%s"}`, i+1, amount)), DedupKey: fmt.Sprintf("n%d", i)}); err != nil {
			t.Fatal(err)
		}
	}
	if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
		t.Fatal(errs)
	}

	// 1. The host speaks from the facts; the words are logged by the agent.
	// Speaking is the POST (/api/speak); a GET of the view reads the facts
	// and the last words, and never calls the model or appends an event.
	type fact struct {
		As   string `json:"as"`
		Kind string `json:"kind"`
		Data struct {
			Rows [][]struct {
				V string `json:"v"`
			} `json:"rows"`
		} `json:"data"`
	}
	var d struct {
		Said      string
		SaidEvent int64
		Silent    string
		Facts     []fact
	}
	var read struct {
		Said  string `json:"said"`
		Facts []fact `json:"facts"`
	}
	brief := func() {
		t.Helper()
		d.Said, d.SaidEvent, d.Silent, d.Facts = "", 0, "", nil
		res, err := ts.Client().Post(ts.URL+"/api/speak?view=briefing", "application/json", nil)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		var out []struct {
			View, Said, Silent string
			SaidEvent          int64 `json:"said_event"`
		}
		if res.StatusCode != 200 || json.NewDecoder(res.Body).Decode(&out) != nil || len(out) != 1 || out[0].View != "briefing" {
			t.Fatalf("speak: %d %+v", res.StatusCode, out)
		}
		d.Said, d.SaidEvent, d.Silent = out[0].Said, out[0].SaidEvent, out[0].Silent
		get("/api/views/briefing", &read)
		d.Facts = read.Facts
	}
	brief()
	if d.Said != words || d.SaidEvent == 0 || d.Silent != "" {
		t.Fatalf("the briefing: %+v", d)
	}
	if read.Said != words {
		t.Fatalf("a read shows the last words: %q", read.Said)
	}
	if before, _ := srvSaid(ctx, s); before != 1 {
		t.Fatalf("said %d", before)
	} else if get("/api/views/briefing", &read); func() int { n, _ := srvSaid(ctx, s); return n }() != before {
		t.Fatal("a read of the briefing spoke")
	}
	if len(d.Facts) != 1 || d.Facts[0].As != "the notes" || d.Facts[0].Kind != "analysis" || d.Facts[0].Data.Rows[0][0].V != "2" || d.Facts[0].Data.Rows[0][1].V != "1234.56" {
		t.Fatalf("the facts: %+v", d.Facts)
	}
	if !strings.Contains(heard, `"1234.56"`) || !strings.Contains(heard, "briefly, in English") || strings.Contains(heard, "said before") {
		t.Fatalf("the host is handed the facts and the voice, nothing said before: %s", heard)
	}
	ev, err := s.GetEvent(ctx, d.SaidEvent)
	if err != nil || ev.Type != core.EventHostSaid || !strings.HasPrefix(ev.Actor, "agent:") || ev.ActivityName != "host_said" {
		t.Fatalf("the words as a fact: %+v %v", ev, err)
	}
	if left, _ := s.UnmatchedRawEvents(ctx); len(left) != 0 {
		t.Fatalf("the host's words are never residue: %v", left)
	}

	// 2. The next time, the host is told what it said, so it does not repeat itself.
	words = "Nothing new: the 2 notes still stand at 1234.56 PLN."
	brief()
	if !strings.Contains(heard, "What you said before") || !strings.Contains(heard, "worth 1234.56 PLN together") || d.Said != words {
		t.Fatalf("the host does not know what it said: %s", heard)
	}

	// 3. A number the facts do not carry is refused: the host never computes.
	words = "The notes are worth about 1300.00 PLN, roughly 617.28 each."
	brief()
	if d.Said != "" || !strings.Contains(d.Silent, "refused") || !strings.Contains(d.Silent, "1300.00") || !strings.Contains(d.Silent, "617.28") {
		t.Fatalf("computed numbers are refused: %+v", d)
	}
	if said, _ := srvSaid(ctx, s); said != 2 {
		t.Fatalf("refused words are not logged: %d said", said)
	}

	// 3b. A fact passed over is refused too: the host never denies what the
	// facts carry. Words without a number clear the first check and fail
	// the second, and the refusal names the fact and its amounts.
	words = "A quiet day: nothing was taken and nothing waits."
	brief()
	if d.Said != "" || !strings.Contains(d.Silent, "pass over") || !strings.Contains(d.Silent, "the notes (2, 1234.56)") {
		t.Fatalf("a denied fact is refused: %+v", d)
	}
	if said, _ := srvSaid(ctx, s); said != 2 {
		t.Fatalf("refused words are not logged: %d said", said)
	}
	// One amount spoken is enough: the count alone names the fact.
	words = "2 notes today, both booked."
	brief()
	if d.Said != words || d.Silent != "" {
		t.Fatalf("one amount of the fact, spoken, is the fact spoken: %+v", d)
	}

	// 3c. One correction (KK's Helios, 2026-10-18): her earlier words named
	// an invoice today's facts do not carry, she named it again and was
	// refused, so nothing new showed. The sentence is no longer shown to
	// her, and refused words are asked for once more, with the reason.
	if _, err := x.TriggerActivity(ctx, "host_said", map[string]any{"view": "briefing", "text": "Invoice FV7 still waits. 2 notes."}, "agent:test", ""); err != nil {
		t.Fatal(err)
	}
	answers = []string{"Invoice FV7 still waits. 2 notes.", "2 notes stand; nothing else waits."}
	brief()
	if d.Said != "2 notes stand; nothing else waits." || !strings.Contains(heard, "refused: they name what the facts do not (FV7)") ||
		strings.Contains(heard, "- Invoice FV7") || !strings.Contains(heard, "- 2 notes.") {
		t.Fatalf("one correction: %+v\n%s", d, heard)
	}
	// Refused twice, the refusal stands.
	answers = []string{"About 1300.00 PLN.", "About 1300.00 PLN, I said."}
	brief()
	if d.Said != "" || !strings.Contains(d.Silent, "1300.00") || len(answers) != 0 {
		t.Fatalf("refused twice: %+v", d)
	}

	// 4. Without a host, the facts speak for themselves.
	srv.Agent = nil
	brief()
	if d.Said != "" || !strings.Contains(d.Silent, "no host") || len(d.Facts) != 1 {
		t.Fatalf("silent: %+v", d)
	}
}

func srvSaid(ctx context.Context, s interface {
	EventsByKind(ctx context.Context, kind string) ([]core.Event, error)
}) (int, error) {
	evs, err := s.EventsByKind(ctx, core.KindRaw)
	n := 0
	for _, e := range evs {
		if e.Type == core.EventHostSaid {
			n++
		}
	}
	return n, err
}

// Invariant 9 at the shell's doors: two companies through the same server,
// each request scoped by its header; a company's list, worklist and
// analysis view hold that company's facts and the installation's, never
// the other's — the analysis copy is the scope's own, so the pack's SUM is
// one company's sum.
func TestTwoCompaniesAtTheDoors(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	srv := &shell.Server{Store: s, Exec: x, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb")}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	call := func(method, company, path string, body any, out any) int {
		t.Helper()
		var rd io.Reader
		if body != nil {
			b, _ := json.Marshal(body)
			rd = bytes.NewReader(b)
		}
		req, _ := http.NewRequest(method, ts.URL+path, rd)
		req.Header.Set("Content-Type", "application/json")
		// "A,B" is an office's desk (X-Rhea-Scope); "A,B>A" the desk acting
		// for A; a bare key acts for that company.
		if scope, acting, office := strings.Cut(company, ">"); office || strings.Contains(scope, ",") {
			req.Header.Set("X-Rhea-Scope", scope)
			if acting != "" {
				req.Header.Set("X-Rhea-Company", acting)
			}
		} else if company != "" {
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
	if err := s.InsertObjectType(ctx, core.ObjectType{Name: "note", Version: 1, Domain: "work", LabelField: "text",
		Fields: []core.FieldDef{{Name: "text", Type: "string", Required: true}, {Name: "amount", Type: "money", Required: true}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: "note", Status: core.StatusActive, Priority: 100,
		EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: "notes", Spec: core.RuleSpec{
			Match:  core.Match{EventType: "note.taken"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text", "amount": "=$.amount"}}}}}); err != nil {
		t.Fatal(err)
	}
	facts, _ := json.Marshal(core.AnalysisSpec{
		SQL:     `SELECT COUNT(*) AS notes, SUM(CAST(json_extract(state, '$.amount') AS BIGINT)) AS total FROM objects WHERE object_type = 'note'`,
		Columns: []core.ColumnSpec{{Field: "notes", Label: "Notes", Type: "int"}, {Field: "total", Label: "Total", Type: "money"}}})
	if err := s.InsertViewDef(ctx, core.ViewDef{ID: "notes-total", Version: 1, Notion: "analysis", Title: "Notes",
		Domain: "work", Function: "analysis", Spec: facts}); err != nil {
		t.Fatal(err)
	}
	alfa, beta := "PL5250001111", "PL5260001246"
	note := func(company, text, amount string) {
		t.Helper()
		var trig map[string]any
		if code := call("POST", company, "/api/activities/submit_event/trigger", map[string]any{"inputs": map[string]any{
			"event_type": "note.taken", "occurred_at": "2026-09-01", "payload": map[string]any{"text": text, "amount": amount}}}, &trig); code != 200 {
			t.Fatalf("%s's note: %d %v", company, code, trig)
		}
	}
	note(alfa, "alfa 1", "100.00")
	note(alfa, "alfa 2", "25.50")
	note(beta, "beta 1", "999.00")
	var view struct {
		Rows [][]struct {
			V string `json:"v"`
		} `json:"rows"`
	}
	total := func(company string) string {
		t.Helper()
		view.Rows = nil
		if code := call("GET", company, "/api/views/notes-total", nil, &view); code != 200 || len(view.Rows) != 1 {
			t.Fatalf("%q's analysis: %d %+v", company, code, view)
		}
		return view.Rows[0][0].V + " notes, " + view.Rows[0][1].V
	}
	if got := total(alfa); got != "2 notes, 125.50" {
		t.Fatalf("alfa's analysis sums %s", got)
	}
	if got := total(beta); got != "1 notes, 999.00" {
		t.Fatalf("beta's analysis sums %s", got)
	}
	if got := total(alfa + "," + beta); got != "3 notes, 1124.50" {
		t.Fatalf("the office's analysis sums %s", got)
	}
	// A request naming no company is the installation's rows alone, never
	// the owner's whole log: it sums nothing of the companies'.
	if got := total(""); got != "0 notes, " {
		t.Fatalf("the installation's analysis sums %s", got)
	}
	// An office's desk acts for one named client when it writes; for none
	// it may read but not write, and a fact never lands as everyone's.
	var trig map[string]any
	if code := call("POST", alfa+","+beta, "/api/activities/submit_event/trigger", map[string]any{"inputs": map[string]any{
		"event_type": "note.taken", "occurred_at": "2026-09-02", "payload": map[string]any{"text": "office", "amount": "1.00"}}}, &trig); code == 200 {
		t.Fatalf("an office acting for no client wrote a fact: %v", trig)
	}
	if code := call("POST", alfa+","+beta+">"+beta, "/api/activities/submit_event/trigger", map[string]any{"inputs": map[string]any{
		"event_type": "note.taken", "occurred_at": "2026-09-02", "payload": map[string]any{"text": "office for beta", "amount": "1.00"}}}, &trig); code != 200 {
		t.Fatalf("an office acting for beta: %d %v", code, trig)
	}
	if got := total(beta); got != "2 notes, 1000.00" {
		t.Fatalf("beta's analysis after the office wrote for it: %s", got)
	}
	if got := total(alfa); got != "2 notes, 125.50" {
		t.Fatalf("alfa's analysis after the office wrote for beta: %s", got)
	}
	// A malformed header fails closed.
	for _, bad := range []string{alfa + ",", ","} {
		req, _ := http.NewRequest("GET", ts.URL+"/api/objects?type=note", nil)
		req.Header.Set("X-Rhea-Company", bad)
		if res, err := ts.Client().Do(req); err != nil || res.StatusCode != 400 {
			t.Fatalf("header %q: %v %v", bad, err, res)
		}
	}
	// The object list is the scope's too.
	var objs struct {
		Objects []map[string]any `json:"objects"`
	}
	call("GET", beta, "/api/objects?type=note", nil, &objs)
	for _, o := range objs.Objects {
		if st, _ := o["state"].(map[string]any); st["text"] == "alfa 1" || st["text"] == "alfa 2" {
			t.Fatalf("beta's list shows alfa's note: %v", o)
		}
	}
}
