package shell_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

// TestCasesArePureData is the falsifiability test passing: Alpha's case
// model — kind, subject, status, deadline, resolving verbs, raised by the
// system from events, never typed in, closing itself when the world moves
// on — expressed entirely as object types, rules and activities. The only
// kernel change it needed was the amendment, which earned admission by the
// recorded fail-as-data proof (DECISIONS 2026-10-05). The shell renders all
// of it generically: no case-specific screen exists anywhere.
func TestCasesArePureData(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
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
	post := func(path string, body any, out any) {
		t.Helper()
		b, _ := json.Marshal(body)
		res, err := ts.Client().Post(ts.URL+path, "application/json", bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			var e map[string]string
			json.NewDecoder(res.Body).Decode(&e)
			t.Fatalf("POST %s: %d %s", path, res.StatusCode, e["error"])
		}
		if out != nil {
			json.NewDecoder(res.Body).Decode(out)
		}
	}

	// --- the vocabulary: all data, no kernel, no screens -------------------
	for _, ot := range []core.ObjectType{
		{
			Name: "complaint", Version: 1, Domain: "crm", IsDocument: true,
			LabelField: "customer",
			Fields: []core.FieldDef{
				{Name: "customer", Type: "string", Required: true},
				{Name: "details", Type: "string"},
				{Name: "registered_on", Type: "date", Required: true},
			},
		},
		{
			Name: "case", Version: 1, Domain: "work", LabelField: "title",
			Fields: []core.FieldDef{
				{Name: "kind", Type: "string", Required: true},
				{Name: "subject_type", Type: "string", Required: true},
				{Name: "subject_id", Type: "string", Required: true},
				{Name: "title", Type: "string", Required: true},
				{Name: "status", Type: "enum", Required: true, Values: []string{"open", "resolved"}},
				{Name: "resolution", Type: "string"},
				{Name: "due_date", Type: "date"},
			},
			Lifecycle: &core.LifecycleDef{Field: "status",
				Transitions: map[string][]string{"open": {"resolved"}}},
		},
	} {
		if err := s.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	openCases, _ := json.Marshal(core.AnalysisSpec{
		SQL: `SELECT json_extract_string(state,'$.title') AS title,
		             json_extract_string(state,'$.kind') AS kind,
		             json_extract_string(state,'$.due_date') AS due
		      FROM objects
		      WHERE object_type = 'case' AND json_extract_string(state,'$.status') = 'open'
		      ORDER BY due, title`,
		Columns: []core.ColumnSpec{{Field: "title", Label: "case"},
			{Field: "kind", Label: "kind"}, {Field: "due", Label: "due", Type: "date"}},
	})
	if err := s.InsertViewDef(ctx, core.ViewDef{
		ID: "open-cases", Version: 1, Notion: "analysis", Title: "Open cases",
		Domain: "work", Function: "worklist", Spec: openCases,
	}); err != nil {
		t.Fatal(err)
	}
	for _, a := range []core.Activity{
		{
			Name: "register_complaint", Status: core.StatusActive, Domain: "crm",
			Description: "register a customer complaint", CreatedBy: "test",
			Spec: core.ActivitySpec{
				Inputs: []core.FieldDef{
					{Name: "customer", Type: "string", Required: true},
					{Name: "details", Type: "string"},
					{Name: "registered_on", Type: "date", Required: true},
				},
				Emits: "complaint.registered", Who: []string{"human"},
			},
		},
		{
			Name: "resolve_case", Status: core.StatusActive, Domain: "work",
			Description: "resolve a case", CreatedBy: "test",
			Spec: core.ActivitySpec{
				Inputs: []core.FieldDef{
					{Name: "case", Type: "ref<case>", Required: true},
					{Name: "resolution", Type: "string", Required: true},
				},
				Emits: "case.resolved", Who: []string{"human"},
			},
		},
	} {
		if _, err := s.InsertActivityVersion(ctx, a); err != nil {
			t.Fatal(err)
		}
	}
	activate := func(id string, priority int, spec core.RuleSpec) {
		t.Helper()
		if _, err := s.InsertRuleVersion(ctx, core.Rule{
			ID: id, Status: core.StatusActive, Priority: priority,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: spec,
		}); err != nil {
			t.Fatal(err)
		}
	}
	activate("book-complaint", 100, core.RuleSpec{
		Match: core.Match{EventType: "complaint.registered"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "complaint",
			Fields: map[string]string{"customer": "=$.customer", "details": "=$.details",
				"registered_on": "=$.registered_on"}}},
	})
	activate("raise-complaint-case", 200, core.RuleSpec{
		Match: core.Match{EventType: core.EventObjectMaterialized, Where: []core.Condition{
			{Path: "$.object_type", Op: "eq", Value: "complaint"},
		}},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "case",
			Fields: map[string]string{
				"kind": "complaint_followup", "subject_type": "complaint",
				"subject_id": "=$.object_id", "title": "=$.state.customer",
				"status": "open", "due_date": "=$.state.registered_on",
			}}},
	})
	activate("resolve-case", 300, core.RuleSpec{
		Match: core.Match{EventType: "case.resolved"},
		Effect: core.Effect{Amend: &core.AmendTemplate{
			Type: "case", Target: "=$.case",
			Set:  map[string]string{"status": "resolved", "resolution": "=$.resolution"},
		}},
	})
	activate("close-on-withdrawal", 400, core.RuleSpec{
		Match: core.Match{EventType: "complaint.withdrawn"},
		Effect: core.Effect{Amend: &core.AmendTemplate{
			Type: "case", Target: "=ref(case, subject_id, $.complaint)",
			Set:  map[string]string{"status": "resolved", "resolution": "withdrawn by customer"},
		}},
	})

	// --- raised by the system, never typed in ------------------------------
	var trig struct {
		EventID int64 `json:"event_id"`
		Booked  int   `json:"booked"`
	}
	post("/api/activities/register_complaint/trigger", map[string]any{
		"inputs": map[string]any{"customer": "ACME Sp. z o.o.",
			"details": "wrong amount on FV 12/2026", "registered_on": "2026-10-05"},
	}, &trig)
	if trig.Booked != 1 {
		t.Fatalf("booked = %d", trig.Booked) // one explained event: complaint + its case, one chain
	}
	cases, _ := s.ObjectsByType(ctx, "case")
	if len(cases) != 1 || cases[0].State["status"] != "open" {
		t.Fatalf("cases = %+v", cases)
	}
	caseID := cases[0].ID

	// --- the resolving verb, offered inline on the generic detail ----------
	var detail struct {
		Activities []struct {
			Name     string `json:"name"`
			RefInput string `json:"ref_input"`
		} `json:"activities"`
		Amendments []struct {
			RuleID string         `json:"rule_id"`
			Set    map[string]any `json:"set"`
		} `json:"amendments"`
	}
	get("/api/views/derived:detail:case?object_id="+caseID, &detail)
	if len(detail.Activities) != 1 || detail.Activities[0].Name != "resolve_case" ||
		detail.Activities[0].RefInput != "case" {
		t.Fatalf("contextual verbs = %+v", detail.Activities)
	}

	// --- "what do I do now": the open-cases surface -------------------------
	var analysis struct {
		Rows [][]struct {
			V string `json:"v"`
		} `json:"rows"`
	}
	get("/api/views/open-cases", &analysis)
	if len(analysis.Rows) != 1 || analysis.Rows[0][0].V != "ACME Sp. z o.o." {
		t.Fatalf("open cases = %+v", analysis.Rows)
	}

	// --- resolved by the human verb -----------------------------------------
	post("/api/activities/resolve_case/trigger", map[string]any{
		"inputs": map[string]any{"case": caseID, "resolution": "credited the difference"},
	}, &trig)
	cases, _ = s.ObjectsByType(ctx, "case")
	if cases[0].State["status"] != "resolved" || cases[0].State["resolution"] != "credited the difference" {
		t.Fatalf("case after verb = %+v", cases[0].State)
	}
	get("/api/views/derived:detail:case?object_id="+caseID, &detail)
	if len(detail.Amendments) != 1 || detail.Amendments[0].RuleID != "resolve-case" ||
		detail.Amendments[0].Set["status"] != "resolved" {
		t.Fatalf("amendments = %+v", detail.Amendments)
	}

	// --- closes itself when the world moves on ------------------------------
	post("/api/activities/register_complaint/trigger", map[string]any{
		"inputs": map[string]any{"customer": "Beta GmbH",
			"details": "late delivery", "registered_on": "2026-10-05"},
	}, &trig)
	complaints, _ := s.ObjectsByType(ctx, "complaint")
	var beta string
	for _, c := range complaints {
		if c.State["customer"] == "Beta GmbH" {
			beta = c.ID
		}
	}
	var withdrawal struct {
		EventID int64 `json:"event_id"`
	}
	post("/api/events", map[string]any{
		"event_type": "complaint.withdrawn", "occurred_at": "2026-10-06",
		"payload":   map[string]any{"complaint": beta},
		"dedup_key": "w-1",
	}, &withdrawal)
	cases, _ = s.ObjectsByType(ctx, "case")
	for _, c := range cases {
		if c.State["status"] != "resolved" {
			t.Fatalf("a case survived the world moving on: %+v", c.State)
		}
	}

	// An empty list means the day's work is done.
	get("/api/views/open-cases", &analysis)
	if len(analysis.Rows) != 0 {
		t.Fatalf("open cases = %+v", analysis.Rows)
	}

	// --- the walk explains the close ----------------------------------------
	var walk struct {
		Tree struct {
			Children []struct {
				EventType string `json:"event_type"`
				RuleID    string `json:"rule_id"`
				Amend     *struct {
					ObjectID string         `json:"object_id"`
					Set      map[string]any `json:"set"`
				} `json:"amend"`
			} `json:"children"`
		} `json:"tree"`
	}
	get(fmt.Sprintf("/api/explain?event=%d", withdrawal.EventID), &walk)
	if len(walk.Tree.Children) != 1 || walk.Tree.Children[0].EventType != core.EventObjectAmended ||
		walk.Tree.Children[0].RuleID != "close-on-withdrawal" ||
		walk.Tree.Children[0].Amend == nil ||
		walk.Tree.Children[0].Amend.Set["status"] != "resolved" {
		t.Fatalf("walk = %+v", walk.Tree)
	}

	// --- determinism: the whole story replays --------------------------------
	before, _ := s.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := s.AllObjects(ctx)
	bj, _ := json.Marshal(before)
	aj, _ := json.Marshal(after)
	if string(bj) != string(aj) {
		t.Fatal("replay diverged")
	}
}
