package shell_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"rhea/internal/adapter"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/network"
	"rhea/internal/shell"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

// TestWhatSheWasNeverTaught (DIRECTION 2026-10-17, KK: Rhea learns from all
// its clients what she was never taught, with people in the loop): each
// untaught moment is a fact in the installation's log — the person's words
// stay there — and her general description reaches the house, counted
// across installations, unless it carries a number or a name the
// installation holds; the fact says which.
func TestWhatSheWasNeverTaught(t *testing.T) {
	ctx := context.Background()
	dsn, drop, err := store.ThrowawayDSN(ctx, "rhea_network_untaught")
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	defer drop()
	nw, err := network.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer nw.Close()

	installation := func(withNetwork bool) func(path string, body, out any) {
		s := storetest.New(t)
		for _, f := range []string{"finance.json", "finance_v2.json", "finance_v3.json", "finance_v4.json",
			"finance_v5.json", "finance_v6.json", "finance_v7.json", "finance_v8.json"} {
			seedFromFile(t, s, f)
		}
		x := &exec.Executor{Store: s}
		srv := &shell.Server{Store: s, Exec: x, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb"),
			Clock: adapter.Clock{Now: func() time.Time { return time.Date(2026, 10, 17, 9, 0, 0, 0, time.UTC) }}}
		if withNetwork {
			srv.Network = nw
		}
		ts := httptest.NewServer(srv.Handler())
		t.Cleanup(ts.Close)
		if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: "register-company", Status: core.StatusDraft, Priority: 10, EffectiveFrom: "2026-01-01",
			CreatedBy: "human", Description: "register-company",
			Spec: core.RuleSpec{Match: core.Match{EventType: "company.registered"}, Effect: core.Effect{Object: core.ObjectTemplate{Type: "company",
				Fields: map[string]string{"name": "=$.name", "kind": "=$.kind", "vat_id": "=$.vat_id", "country": "=$.country"}}}},
		}); err != nil {
			t.Fatal(err)
		}
		if _, _, _, err := x.ApproveRule(ctx, "register-company", "krzysztof", "2026-09-01"); err != nil {
			t.Fatal(err)
		}
		if tr, err := x.TriggerActivity(ctx, "submit_event", map[string]any{"event_type": "company.registered", "occurred_at": "2026-09-01",
			"payload": map[string]any{"name": "Agencja Reklamowa Nova sp. z o.o.", "kind": "customer", "vat_id": "5260001246", "country": "PL"}}, "test", "nova"); err != nil || tr.Booked != 1 {
			t.Fatalf("company: %+v %v", tr, err)
		}
		return func(path string, body, out any) {
			t.Helper()
			method, rd := "GET", (*bytes.Reader)(nil)
			if body != nil {
				b, _ := json.Marshal(body)
				method, rd = "POST", bytes.NewReader(b)
			}
			req := httptest.NewRequest(method, ts.URL+path, nil)
			if rd != nil {
				req = httptest.NewRequest(method, ts.URL+path, rd)
			}
			req.RequestURI = ""
			res, err := ts.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			var b bytes.Buffer
			b.ReadFrom(res.Body)
			if res.StatusCode != 200 {
				t.Fatalf("%s %s: %d %s", method, path, res.StatusCode, b.String())
			}
			if err := json.Unmarshal(b.Bytes(), out); err != nil {
				t.Fatal(err)
			}
		}
	}
	type noted struct {
		Shared   bool   `json:"shared"`
		KeptHome string `json:"kept_home_because"`
	}
	note := func(call func(string, any, any), asked, missing string) noted {
		var n noted
		call("/api/untaught", map[string]any{"asked": asked, "missing": missing, "asked_by": "krzysztof"}, &n)
		return n
	}

	one := installation(true)
	if n := note(one, "Możesz pokazać sprzedaż dla Novej, miesiąc po miesiącu?", "sprzedaż według klientów miesiąc po miesiącu"); !n.Shared {
		t.Fatalf("a general description stayed home: %+v", n)
	}
	if n := note(one, "faktury Novej", "faktury Agencji Novej z sierpnia"); n.Shared || !strings.Contains(n.KeptHome, "Nova") {
		t.Fatalf("a name left: %+v", n)
	}
	if n := note(one, "wynik", "wynik za 2026-08"); n.Shared || n.KeptHome != "it carries a number" {
		t.Fatalf("a number left: %+v", n)
	}
	var log []struct {
		Asked, Missing string
		Shared         bool
		KeptHome       string `json:"kept_home_because"`
	}
	one("/api/untaught", nil, &log)
	if len(log) != 3 || log[0].Missing != "wynik za 2026-08" || log[2].Asked != "Możesz pokazać sprzedaż dla Novej, miesiąc po miesiącu?" || !log[2].Shared {
		t.Fatalf("the installation's own record, newest first: %+v", log)
	}

	// A second installation was never taught the same, in other letters: the
	// house counts two installations, and holds none of their words.
	two := installation(true)
	note(two, "a sprzedaż według klientów?", "Sprzedaż według klientów miesiąc po miesiącu ")
	us, err := nw.Untaught(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(us) != 1 || us[0].Installations != 2 || us[0].Times != 2 {
		t.Fatalf("the house reads %+v", us)
	}
	var leaked int
	if err := nw.Pool.QueryRow(ctx, `SELECT count(*) FROM untaught WHERE missing ILIKE '%nov%' OR missing ILIKE '%pokaż%' OR missing ~ '[0-9]'`).Scan(&leaked); err != nil || leaked != 0 {
		t.Fatalf("words, names or numbers reached the house: %d %v", leaked, err)
	}
	if _, err := nw.Pool.Exec(ctx, `DELETE FROM untaught`); err == nil {
		t.Fatalf("the house's record can be deleted")
	}

	// The Learning Centre reads the house: the lesson counted across both
	// installations, open until the team says where it stands; the team's
	// latest note is the lesson's, and every note stays on the record.
	type centre struct {
		Network bool `json:"network"`
		Lessons []struct {
			Missing, Lesson, Status, Note string
			NotedBy                       string   `json:"noted_by"`
			Installations                 int      `json:"installations"`
			Names                         []string `json:"names"`
		} `json:"lessons"`
	}
	var lc centre
	one("/api/network/learning", nil, &lc)
	if !lc.Network || len(lc.Lessons) != 1 || lc.Lessons[0].Installations != 2 || len(lc.Lessons[0].Names) != 2 || lc.Lessons[0].Status != "open" {
		t.Fatalf("the Learning Centre reads %+v", lc)
	}
	lesson := lc.Lessons[0].Lesson
	var filed map[string]any
	one("/api/network/lessons", map[string]any{"lesson": lesson, "status": "planned", "note": "Poland pack v27: sprzedaż według klientów", "by": "krzysztof"}, &filed)
	one("/api/network/lessons", map[string]any{"lesson": lesson, "status": "done", "note": "shipped in v27", "by": "krzysztof"}, &filed)
	one("/api/network/learning", nil, &lc)
	if l := lc.Lessons[0]; l.Status != "done" || l.Note != "shipped in v27" || l.NotedBy != "krzysztof" {
		t.Fatalf("the latest note is not the lesson's: %+v", l)
	}
	if _, err := nw.NoteLesson(ctx, lesson, "maybe", "", "krzysztof"); err == nil {
		t.Fatalf("a status the team does not use was filed")
	}
	if _, err := nw.Pool.Exec(ctx, `DELETE FROM lesson_note`); err == nil {
		t.Fatalf("the team's notes can be deleted")
	}

	// Without a network it is still a fact, kept home and saying why.
	alone := installation(false)
	if n := note(alone, "x", "wynik za jeden miesiąc"); n.Shared || n.KeptHome != "no network is connected" {
		t.Fatalf("alone: %+v", n)
	}
	alone("/api/network/learning", nil, &lc)
	if lc.Network {
		t.Fatalf("a kernel without a network reads one")
	}
}
