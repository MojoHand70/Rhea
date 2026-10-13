package shell_test

import (
	"context"
	"encoding/json"
	"fmt"
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
	srv := &shell.Server{Store: s, Exec: x, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb"),
		Agent: &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
			heard = user
			if !strings.Contains(system, "never compute") {
				return "", fmt.Errorf("the host is told its law")
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
	var d struct {
		Said      string `json:"said"`
		SaidEvent int64  `json:"said_event"`
		Silent    string `json:"silent"`
		Facts     []struct {
			As   string `json:"as"`
			Kind string `json:"kind"`
			Data struct {
				Rows [][]struct {
					V string `json:"v"`
				} `json:"rows"`
			} `json:"data"`
		} `json:"facts"`
	}
	get("/api/views/briefing", &d)
	if d.Said != words || d.SaidEvent == 0 || d.Silent != "" {
		t.Fatalf("the briefing: %+v", d)
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

	fresh := func() { d.Said, d.SaidEvent, d.Silent, d.Facts = "", 0, "", nil }
	// 2. The next time, the host is told what it said, so it does not repeat itself.
	words = "Nothing new since the two notes."
	fresh()
	get("/api/views/briefing", &d)
	if !strings.Contains(heard, "What you said before") || !strings.Contains(heard, "worth 1234.56 PLN together") || d.Said != words {
		t.Fatalf("the host does not know what it said: %s", heard)
	}

	// 3. A number the facts do not carry is refused: the host never computes.
	words = "The notes are worth about 1300.00 PLN, roughly 617.28 each."
	fresh()
	get("/api/views/briefing", &d)
	if d.Said != "" || !strings.Contains(d.Silent, "refused") || !strings.Contains(d.Silent, "1300.00") || !strings.Contains(d.Silent, "617.28") {
		t.Fatalf("computed numbers are refused: %+v", d)
	}
	if said, _ := srvSaid(ctx, s); said != 2 {
		t.Fatalf("refused words are not logged: %d said", said)
	}

	// 4. Without a host, the facts speak for themselves.
	srv.Agent = nil
	fresh()
	get("/api/views/briefing", &d)
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
