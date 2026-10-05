package exec_test

import (
	"context"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

// The M5 attempt, run before the clock adapter existed (the E1 discipline):
// express recurring schedules with rules alone — a time event in the log,
// matched by rules that raise the scheduled work. Kept as the negative
// proof; see DECISIONS 2026-10-05, "M5 attempt".
//
// What the attempt establishes:
//  1. Firing works for exactly one due schedule — by equality resolution
//     (=ref(task_schedule, next_run, $.date)), which is as far as match
//     conditions can see: they read the event payload, never object state.
//  2. Two schedules due the same day make the ref ambiguous and the whole
//     day refuses — the due *set* is inexpressible (fan-out exists only
//     over payload arrays, not over state).
//  3. A day with NOTHING due fails loudly: the ref resolves nothing, the
//     firing is a rule error, and every quiet day lands in the worklist —
//     the encoding cannot even express "usually, nothing happens".
//  4. Nothing can advance next_run: "+1 week" is date arithmetic no
//     template has, so a recurring schedule fires once and dies — the
//     cadence field is a lie the language cannot keep.
//
// The exit (per the adapter contract, SPEC M3): calendars are the clock's
// knowledge. The clock adapter reads due schedules, emits one
// schedule.fired event per occurrence with the advanced next_run baked in
// the payload, and ordinary rules — pure data — raise the work and amend
// the schedule. The kernel learns nothing.
func TestRecurringSchedulesFailAsPureData(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}

	for _, ot := range []core.ObjectType{
		{
			Name: "task_schedule", Version: 1, Domain: "work", LabelField: "name",
			Fields: []core.FieldDef{
				{Name: "name", Type: "string", Required: true},
				{Name: "cadence", Type: "enum", Required: true, Values: []string{"daily", "weekly", "monthly", "yearly"}},
				{Name: "next_run", Type: "date", Required: true},
				{Name: "status", Type: "enum", Required: true, Values: []string{"active", "paused"}},
			},
			Lifecycle: &core.LifecycleDef{Field: "status",
				Transitions: map[string][]string{"active": {"paused"}, "paused": {"active"}}},
		},
		{
			Name: "task", Version: 1, Domain: "work",
			Fields: []core.FieldDef{
				{Name: "schedule", Type: "ref<task_schedule>", Required: true},
				{Name: "date", Type: "date", Required: true},
			},
		},
	} {
		if err := x.Store.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	activateRule(t, x, "register-schedule", 100, core.RuleSpec{
		Match: core.Match{EventType: "schedule.created"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "task_schedule",
			Fields: map[string]string{"name": "=$.name", "cadence": "=$.cadence",
				"next_run": "=$.next_run", "status": "active"}}},
	})
	// The attempt itself: fire whatever schedule is due on the opened day.
	// Equality against the day is the only selection match conditions allow.
	activateRule(t, x, "fire-due-schedule", 200, core.RuleSpec{
		Match: core.Match{EventType: "time.day_opened"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "task",
			Fields: map[string]string{
				"schedule": "=ref(task_schedule, next_run, $.date)",
				"date":     "=$.date",
			}}},
	})

	submitRaw(t, x, "s-1", "schedule.created", "2026-10-01",
		`{"name":"weekly VAT check","cadence":"weekly","next_run":"2026-10-06"}`)
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("schedule: booked %d, errs %v", n, errs)
	}

	// (1) One due schedule fires — this much carries.
	submitRaw(t, x, "d-6", "time.day_opened", "2026-10-06", `{"date":"2026-10-06"}`)
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("day 6: booked %d, errs %v", n, errs)
	}
	tasks, _ := x.Store.ObjectsByType(ctx, "task")
	if len(tasks) != 1 {
		t.Fatalf("tasks = %d", len(tasks))
	}

	// (3)+(4) The week after: next_run is frozen — no template computes
	// "+1 week" — so nothing is due on the 13th, and a day with nothing due
	// is not a quiet day but a rule ERROR: the unresolvable ref refuses the
	// day into the worklist. The weekly schedule fired once and died, and
	// every uneventful day now demands a human.
	submitRaw(t, x, "d-13", "time.day_opened", "2026-10-13", `{"date":"2026-10-13"}`)
	_, errs := x.ProcessPending(ctx)
	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "no task_schedule") {
		t.Fatalf("day 13: want the quiet-day failure, got %v", errs)
	}
	tasks, _ = x.Store.ObjectsByType(ctx, "task")
	if len(tasks) != 1 {
		t.Fatalf("the dead schedule fired anyway: %d tasks", len(tasks))
	}
	schedules, _ := x.Store.ObjectsByType(ctx, "task_schedule")
	if schedules[0].State["next_run"] != "2026-10-06" {
		t.Fatalf("next_run moved without arithmetic: %v", schedules[0].State["next_run"])
	}

	// (2) Two schedules due the same day: the ref is ambiguous and the day
	// refuses whole — the due set cannot fan out from state.
	submitRaw(t, x, "s-2", "schedule.created", "2026-10-01",
		`{"name":"backup","cadence":"daily","next_run":"2026-10-20"}`)
	submitRaw(t, x, "s-3", "schedule.created", "2026-10-01",
		`{"name":"standup","cadence":"daily","next_run":"2026-10-20"}`)
	// The refused quiet day keeps failing honestly on every pass.
	if n, errs := x.ProcessPending(ctx); n != 2 || len(errs) != 1 {
		t.Fatalf("schedules: booked %d, errs %v", n, errs)
	}
	submitRaw(t, x, "d-20", "time.day_opened", "2026-10-20", `{"date":"2026-10-20"}`)
	_, errs = x.ProcessPending(ctx)
	ambiguous := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "ambiguous") {
			ambiguous = true
		}
	}
	if !ambiguous {
		t.Fatalf("two due schedules did not refuse the day: %v", errs)
	}
	tasks, _ = x.Store.ObjectsByType(ctx, "task")
	if len(tasks) != 1 {
		t.Fatalf("an ambiguous day half-fired: %d tasks", len(tasks))
	}
}
