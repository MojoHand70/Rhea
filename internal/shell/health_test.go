package shell_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"rhea/internal/adapter"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

// Failures speak (review 2026-10-21, G3): the clock answered 200 when its
// day did not open — days waiting for a person's catch-up — the morning
// answered 200 when a narrative failed, and nothing told how the books'
// day stood: overnight, nothing reached KK until someone opened Today. Now
// a clock waiting for a person answers 409, a failing morning 500, and
// /api/health says whether the books' day is open — 503 when it lags.
func TestFailuresSpeak(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	day, _ := time.Parse("2006-01-02", "2026-10-10")
	var now atomic.Int64
	now.Store(day.Unix())
	srv := &shell.Server{Store: s, Exec: x, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb"),
		Clock: adapter.Clock{Now: func() time.Time { return time.Unix(now.Load(), 0).UTC() }}}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	call := func(method, path string) (int, map[string]any) {
		t.Helper()
		req, _ := http.NewRequest(method, ts.URL+path, nil)
		req.Header.Set("Content-Type", "application/json")
		res, err := ts.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		var out map[string]any
		json.NewDecoder(res.Body).Decode(&out)
		return res.StatusCode, out
	}

	if code, out := call("POST", "/api/time/advance"); code != 200 {
		t.Fatalf("the first day opens: %d %v", code, out)
	}
	if code, out := call("GET", "/api/health"); code != 200 || out["ok"] != true || out["last_opened"] != "2026-10-10" {
		t.Fatalf("health on an open day: %d %v", code, out)
	}
	// Three days on and nobody caught up: the clock waits for a person.
	now.Store(day.AddDate(0, 0, 3).Unix())
	if code, out := call("POST", "/api/time/advance"); code != 409 || out["gated"] != true {
		t.Fatalf("a clock waiting for a person: %d %v", code, out)
	}
	if code, out := call("GET", "/api/health"); code != 503 || out["ok"] != false {
		t.Fatalf("health with days unopened: %d %v", code, out)
	}
	// A morning whose narrative fails says so.
	spec, _ := json.Marshal(core.NarrativeSpec{Facts: []core.NarrativeFact{{View: "no-such-view", As: "nothing"}}, Voice: "briefly"})
	if err := s.InsertViewDef(ctx, core.ViewDef{ID: "briefing", Version: 1, Notion: "narrative", Title: "Briefing",
		Domain: "work", Function: "briefing", Spec: spec}); err != nil {
		t.Fatal(err)
	}
	if code, out := call("POST", "/api/speak"); code != 500 {
		t.Fatalf("a failing morning: %d %v", code, out)
	}
}
