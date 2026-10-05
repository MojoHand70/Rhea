package shell_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

// TestVerbSurface drives the second vocabulary over HTTP: the offered verb
// set, the gate (a pack-style draft approved in the shell), the trigger
// door, and the contextual verbs a detail view offers — what can be done,
// served as data next to what can be seen.
func TestVerbSurface(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	seedFromFile(t, s, "finance.json")
	x := &exec.Executor{Store: s}
	srv := &shell.Server{Store: s, Exec: x, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb")}
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
	post := func(path string, body any, out any) int {
		t.Helper()
		b, _ := json.Marshal(body)
		res, err := ts.Client().Post(ts.URL+path, "application/json", bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if out != nil && res.StatusCode == 200 {
			json.NewDecoder(res.Body).Decode(out)
		}
		return res.StatusCode
	}

	type verb struct {
		Name          string            `json:"name"`
		Version       int               `json:"version"`
		Status        string            `json:"status"`
		Spec          core.ActivitySpec `json:"spec"`
		Offered       bool              `json:"offered"`
		EventsEmitted int               `json:"events_emitted"`
	}

	// 1. The builtins are offered; the seed stamped approve_activity with
	// one activation event per builtin.
	var verbs []verb
	get("/api/activities", &verbs)
	byName := map[string]verb{}
	for _, v := range verbs {
		byName[v.Name] = v
	}
	for _, name := range []string{"submit_event", "approve_rule", "approve_activity", "draft_rule"} {
		if !byName[name].Offered {
			t.Fatalf("builtin %s not offered: %+v", name, verbs)
		}
	}
	if byName["approve_activity"].EventsEmitted != len(core.BuiltinActivities()) {
		t.Fatalf("approve_activity emitted %d, want %d",
			byName["approve_activity"].EventsEmitted, len(core.BuiltinActivities()))
	}

	// 2. A pack-style draft verb: listed unoffered, refuses triggers, and the
	// shell's approve endpoint — the gate applied to the gate — activates it.
	if _, err := s.InsertActivityVersion(ctx, core.Activity{
		Name: "dispute_invoice", Status: core.StatusDraft, Domain: "finance",
		Description: "dispute an invoice", CreatedBy: "pack:test",
		Spec: core.ActivitySpec{
			Inputs: []core.FieldDef{
				{Name: "invoice", Type: "ref<invoice>", Required: true},
				{Name: "reason", Type: "string", Required: true},
			},
			Emits: "invoice.disputed", Who: []string{"human"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	get("/api/activities", &verbs)
	var draft *verb
	for i := range verbs {
		if verbs[i].Name == "dispute_invoice" {
			draft = &verbs[i]
		}
	}
	if draft == nil || draft.Offered || draft.Status != core.StatusDraft {
		t.Fatalf("draft verb listing = %+v", draft)
	}
	if code := post("/api/activities/dispute_invoice/trigger",
		map[string]any{"inputs": map[string]any{"invoice": "x", "reason": "y"}}, nil); code == 200 {
		t.Fatal("a draft verb was triggerable")
	}
	var approved struct {
		Activity core.Activity `json:"activity"`
	}
	if code := post("/api/activities/dispute_invoice/approve",
		map[string]any{"approved_by": "krzysztof"}, &approved); code != 200 {
		t.Fatalf("approve: %d", code)
	}
	if approved.Activity.Version != 2 || approved.Activity.Status != core.StatusActive {
		t.Fatalf("approved = %+v", approved.Activity)
	}

	// 3. Materialize an invoice so the ref verb has something to act on.
	if _, err := s.InsertRuleVersion(ctx, core.Rule{
		ID: "book-pln-invoice", Status: core.StatusDraft, Priority: 100,
		EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: "book PLN invoices",
		Spec: core.RuleSpec{
			Match: core.Match{EventType: "invoice.received"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "invoice",
				Fields: map[string]string{
					"customer": "=$.customer", "issue_date": "=$.issue_date",
					"currency": "=$.currency", "total": "=sum($.lines[*].amount)",
				}}},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if code := post("/api/rules/book-pln-invoice/approve",
		map[string]any{"approved_by": "krzysztof"}, nil); code != 200 {
		t.Fatal("rule approval failed")
	}
	var submitted struct {
		EventID int64 `json:"event_id"`
		Booked  int   `json:"booked"`
	}
	if code := post("/api/events", map[string]any{
		"event_type": "invoice.received", "occurred_at": "2026-09-15",
		"payload": map[string]any{
			"customer": "ACME Sp. z o.o.", "currency": "PLN", "issue_date": "2026-09-15",
			"lines": []map[string]any{{"desc": "Widget", "qty": 1, "amount": "100.00"}},
		},
	}, &submitted); code != 200 || submitted.Booked != 1 {
		t.Fatalf("submit: %d booked=%d", code, submitted.Booked)
	}
	objs, _ := s.ObjectsByType(ctx, "invoice")
	if len(objs) != 1 {
		t.Fatalf("objects = %v", objs)
	}
	// The submitted event wears the submit_event stamp — even over HTTP,
	// the free-form endpoint is the declared open door.
	ev, _ := s.GetEvent(ctx, submitted.EventID)
	if ev.ActivityName != "submit_event" {
		t.Fatalf("event stamp = %q", ev.ActivityName)
	}

	// 4. The detail view offers the contextual verb: dispute_invoice takes a
	// ref<invoice>, so every invoice detail carries it.
	var detail struct {
		Activities []struct {
			Name     string `json:"name"`
			RefInput string `json:"ref_input"`
		} `json:"activities"`
	}
	get("/api/views/invoice-detail?object_id="+objs[0].ID, &detail)
	if len(detail.Activities) != 1 || detail.Activities[0].Name != "dispute_invoice" ||
		detail.Activities[0].RefInput != "invoice" {
		t.Fatalf("contextual verbs = %+v", detail.Activities)
	}

	// 5. Triggering it emits the stamped event into the worklist.
	var triggered struct {
		EventID int64 `json:"event_id"`
	}
	if code := post("/api/activities/dispute_invoice/trigger", map[string]any{
		"inputs": map[string]any{"invoice": objs[0].ID, "reason": "wrong amount"},
	}, &triggered); code != 200 {
		t.Fatal("trigger failed")
	}
	ev, _ = s.GetEvent(ctx, triggered.EventID)
	if ev.Type != "invoice.disputed" || ev.ActivityName != "dispute_invoice" || ev.ActivityVersion != 2 {
		t.Fatalf("triggered event = %+v", ev)
	}
	wl, _ := s.UnmatchedRawEvents(ctx)
	if len(wl) != 1 || wl[0].ID != triggered.EventID {
		t.Fatalf("worklist = %v", wl)
	}
}
