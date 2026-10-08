package shell_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"rhea/internal/adapter"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

// TestTimeSurfaceAndSchedulingNotion drives M5 over HTTP: the burst gate
// (steady state refuses a gap, the dry-run shows what opening it fires,
// catch-up is the deliberate act), scheduled work raising cases, and the
// scheduling notion — the founding table's last entry — rendering the
// cases on their time axis.
func TestTimeSurfaceAndSchedulingNotion(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	day, _ := time.Parse("2006-01-02", "2026-10-10")
	srv := &shell.Server{Store: s, Exec: x, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb"),
		Clock: adapter.Clock{Now: func() time.Time { return day }}}
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

	// The world: schedules and cases as data, one scheduling view over the
	// cases' due dates. The 'scheduling' notion passes the schema's CHECK.
	for _, ot := range []core.ObjectType{
		{
			Name: adapter.ScheduleType, Version: 1, Domain: "work", LabelField: "name",
			Fields: []core.FieldDef{
				{Name: "name", Type: "string", Required: true},
				{Name: "cadence", Type: "enum", Required: true, Values: []string{"daily", "weekly", "monthly", "yearly"}},
				{Name: "next_run", Type: "date", Required: true},
				{Name: "status", Type: "enum", Required: true, Values: []string{"active", "paused"}},
			},
			Lifecycle: &core.LifecycleDef{Field: "status",
				Transitions: map[string][]string{"active": {"paused"}, "paused": {"active"}}},
			Amendable: []string{"next_run"}, // the schedule advances itself
		},
		{
			Name: "case", Version: 1, Domain: "work", LabelField: "title",
			Fields: []core.FieldDef{
				{Name: "kind", Type: "string", Required: true},
				{Name: "subject_id", Type: "string", Required: true},
				{Name: "title", Type: "string", Required: true},
				{Name: "status", Type: "enum", Required: true, Values: []string{"open", "resolved"}},
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
	schedSpec, _ := json.Marshal(core.SchedulingSpec{ObjectType: "case", DateField: "due_date"})
	if err := s.InsertViewDef(ctx, core.ViewDef{
		ID: "case-schedule", Version: 1, Notion: "scheduling", Title: "Work on the axis",
		Domain: "work", Function: "worklist", Spec: schedSpec,
	}); err != nil {
		t.Fatal(err)
	}
	rules := map[string]core.RuleSpec{
		"register-schedule": {
			Match: core.Match{EventType: "schedule.created"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: adapter.ScheduleType,
				Fields: map[string]string{"name": "=$.name", "cadence": "=$.cadence",
					"next_run": "=$.next_run", "status": "active"}}},
		},
		"raise-scheduled-case": {
			Match: core.Match{EventType: adapter.EventScheduleFired},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "case",
				Fields: map[string]string{"kind": "scheduled", "subject_id": "=$.schedule",
					"title": "=$.name", "status": "open", "due_date": "=$.occurrence"}}},
		},
		"advance-schedule": {
			Match: core.Match{EventType: adapter.EventScheduleFired},
			Effect: core.Effect{Amend: &core.AmendTemplate{
				Type: adapter.ScheduleType, Target: "=$.schedule",
				Set: map[string]string{"next_run": "=$.next_run"}}},
		},
	}
	prio := 100
	for _, id := range []string{"register-schedule", "raise-scheduled-case", "advance-schedule"} {
		if _, err := s.InsertRuleVersion(ctx, core.Rule{
			ID: id, Status: core.StatusActive, Priority: prio,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: rules[id],
		}); err != nil {
			t.Fatal(err)
		}
		prio += 100
	}

	// History: the last opened day was the 6th; a daily schedule is due
	// from the 7th. Four days then passed unattended.
	if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: core.EventDayOpened,
		OccurredAt: "2026-10-06", Payload: json.RawMessage(`{"date":"2026-10-06"}`),
		DedupKey: "time/day/2026-10-06", Actor: "adapter:clock"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: "schedule.created",
		OccurredAt: "2026-10-06", DedupKey: "s-1",
		Payload: json.RawMessage(`{"name":"daily backup","cadence":"daily","next_run":"2026-10-07"}`)}); err != nil {
		t.Fatal(err)
	}
	x.ProcessPending(ctx)

	// 1. The status names the gap.
	var st adapter.ClockStatus
	get("/api/time", &st)
	if st.Today != "2026-10-10" || st.LastOpened != "2026-10-06" || len(st.Pending) != 4 {
		t.Fatalf("status = %+v", st)
	}

	// 2. Steady state refuses the burst.
	var res struct {
		Gated  bool     `json:"gated"`
		Opened []string `json:"opened"`
		Fired  int      `json:"fired"`
	}
	post("/api/time/advance", nil, &res)
	if !res.Gated || len(res.Opened) != 0 {
		t.Fatalf("advance on a gap = %+v", res)
	}

	// 3. The dry run reads like any approval: four would-be cases, the
	// schedule changed, nothing written.
	var sim struct {
		Days []string `json:"days"`
		Diff struct {
			Added   []map[string]any `json:"added"`
			Changed []map[string]any `json:"changed"`
		} `json:"diff"`
	}
	post("/api/time/simulate", map[string]any{}, &sim)
	if len(sim.Days) != 4 || len(sim.Diff.Added) != 4 || len(sim.Diff.Changed) != 1 {
		t.Fatalf("dry run: days %d, added %d, changed %d", len(sim.Days), len(sim.Diff.Added), len(sim.Diff.Changed))
	}
	if cases, _ := s.ObjectsByType(ctx, "case"); len(cases) != 0 {
		t.Fatal("the dry run wrote")
	}

	// 4. The deliberate act opens the days; the missed occurrences fire
	// late, in order.
	post("/api/time/catchup", map[string]any{}, &res)
	if len(res.Opened) != 4 || res.Fired != 4 {
		t.Fatalf("catch-up = %+v", res)
	}
	cases, _ := s.ObjectsByType(ctx, "case")
	if len(cases) != 4 {
		t.Fatalf("cases = %d", len(cases))
	}

	// 5. The scheduling notion lays the work on the axis: four days, one
	// case each, today marked for the renderer.
	var sched struct {
		Today string `json:"today"`
		Days  []struct {
			Date  string `json:"date"`
			Items []struct {
				Label  string `json:"label"`
				Status string `json:"status"`
				Detail string `json:"detail"`
			} `json:"items"`
		} `json:"days"`
	}
	get("/api/views/case-schedule", &sched)
	if sched.Today != "2026-10-10" || len(sched.Days) != 4 {
		t.Fatalf("scheduling view = %+v", sched)
	}
	first := sched.Days[0]
	if first.Date != "2026-10-07" || len(first.Items) != 1 ||
		first.Items[0].Label != "daily backup" || first.Items[0].Status != "open" ||
		first.Items[0].Detail == "" {
		t.Fatalf("first day = %+v", first)
	}
}
