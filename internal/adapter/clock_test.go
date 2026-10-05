package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

func fixedClock(day string) Clock {
	t, err := time.Parse("2006-01-02", day)
	if err != nil {
		panic(err)
	}
	return Clock{Now: func() time.Time { return t }}
}

// seedClockWorld: the schedule vocabulary as pure data — the clock knows
// the task_schedule convention, the kernel knows nothing. Scheduled work
// raises cases (the M5 story composing with the cases story), and the
// advance is an ordinary amend rule reading the clock's baked next_run.
func seedClockWorld(t *testing.T, x *exec.Executor) {
	t.Helper()
	ctx := context.Background()
	for _, ot := range []core.ObjectType{
		{
			Name: ScheduleType, Version: 1, Domain: "work", LabelField: "name",
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
			Name: "case", Version: 1, Domain: "work", LabelField: "title",
			Fields: []core.FieldDef{
				{Name: "kind", Type: "string", Required: true},
				{Name: "subject_type", Type: "string", Required: true},
				{Name: "subject_id", Type: "string", Required: true},
				{Name: "title", Type: "string", Required: true},
				{Name: "status", Type: "enum", Required: true, Values: []string{"open", "resolved"}},
				{Name: "due_date", Type: "date"},
			},
			Lifecycle: &core.LifecycleDef{Field: "status",
				Transitions: map[string][]string{"open": {"resolved"}}},
		},
	} {
		if err := x.Store.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	rules := map[string]core.RuleSpec{
		"register-schedule": {
			Match: core.Match{EventType: "schedule.created"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: ScheduleType,
				Fields: map[string]string{"name": "=$.name", "cadence": "=$.cadence",
					"next_run": "=$.next_run", "status": "active"}}},
		},
		"raise-scheduled-case": {
			Match: core.Match{EventType: EventScheduleFired},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "case",
				Fields: map[string]string{
					"kind": "scheduled", "subject_type": ScheduleType,
					"subject_id": "=$.schedule", "title": "=$.name",
					"status": "open", "due_date": "=$.occurrence",
				}}},
		},
		"advance-schedule": {
			Match: core.Match{EventType: EventScheduleFired},
			Effect: core.Effect{Amend: &core.AmendTemplate{
				Type: ScheduleType, Target: "=$.schedule",
				Set: map[string]string{"next_run": "=$.next_run"},
			}},
		},
		"pause-schedule": {
			Match: core.Match{EventType: "schedule.paused"},
			Effect: core.Effect{Amend: &core.AmendTemplate{
				Type: ScheduleType, Target: "=$.schedule",
				Set: map[string]string{"status": "paused"},
			}},
		},
		"resume-schedule": {
			Match: core.Match{EventType: "schedule.resumed"},
			Effect: core.Effect{Amend: &core.AmendTemplate{
				Type: ScheduleType, Target: "=$.schedule",
				Set: map[string]string{"status": "active"},
			}},
		},
	}
	prio := 100
	for _, id := range []string{"register-schedule", "raise-scheduled-case",
		"advance-schedule", "pause-schedule", "resume-schedule"} {
		if _, err := x.Store.InsertRuleVersion(ctx, core.Rule{
			ID: id, Status: core.StatusActive, Priority: prio,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: rules[id],
		}); err != nil {
			t.Fatal(err)
		}
		prio += 100
	}
}

func submitClockRaw(t *testing.T, x *exec.Executor, dedup, typ, date, payload string) {
	t.Helper()
	if _, err := x.Store.AppendEvent(context.Background(), core.Event{
		Kind: core.KindRaw, Type: typ, OccurredAt: date,
		Payload: json.RawMessage(payload), DedupKey: dedup,
	}); err != nil {
		t.Fatal(err)
	}
	if _, errs := x.ProcessPending(context.Background()); len(errs) != 0 {
		t.Fatalf("%s: %v", typ, errs)
	}
}

func TestClockSteadyState(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	seedClockWorld(t, x)

	// Genesis: the clock starts counting the day it first runs; the day and
	// its month open together.
	res, err := RunClock(ctx, x.Store, x, fixedClock("2026-10-06"))
	if err != nil || res.Gated || len(res.Opened) != 1 || res.Opened[0] != "2026-10-06" {
		t.Fatalf("genesis: %+v, %v", res, err)
	}
	if _, ok, _ := x.Store.LatestEventOfType(ctx, core.EventMonthOpened); !ok {
		t.Fatal("the genesis month did not open")
	}
	// Same day again: nothing to do, nothing duplicated.
	res, _ = RunClock(ctx, x.Store, x, fixedClock("2026-10-06"))
	if res.Gated || len(res.Opened) != 0 {
		t.Fatalf("idle run: %+v", res)
	}
	// Quiet days stay out of the human worklist.
	if wl, _ := x.Store.UnmatchedRawEvents(ctx); len(wl) != 0 {
		t.Fatalf("time leaked into the worklist: %v", wl)
	}

	// A schedule becomes due: the day opens, the occurrence fires, the case
	// raises with the occurrence as its due date, and the amend rule moves
	// next_run one cadence step — all in one run.
	submitClockRaw(t, x, "s-1", "schedule.created", "2026-10-06",
		`{"name":"weekly VAT check","cadence":"weekly","next_run":"2026-10-07"}`)
	res, err = RunClock(ctx, x.Store, x, fixedClock("2026-10-07"))
	if err != nil || len(res.Opened) != 1 || res.Fired != 1 {
		t.Fatalf("due day: %+v, %v", res, err)
	}
	cases, _ := x.Store.ObjectsByType(ctx, "case")
	if len(cases) != 1 || cases[0].State["due_date"] != "2026-10-07" ||
		cases[0].State["title"] != "weekly VAT check" {
		t.Fatalf("case = %+v", cases)
	}
	scheds, _ := x.Store.ObjectsByType(ctx, ScheduleType)
	if scheds[0].State["next_run"] != "2026-10-14" {
		t.Fatalf("next_run = %v", scheds[0].State["next_run"])
	}
}

func TestClockGateAndCatchUp(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	seedClockWorld(t, x)

	if _, err := RunClock(ctx, x.Store, x, fixedClock("2026-10-06")); err != nil {
		t.Fatal(err)
	}
	submitClockRaw(t, x, "s-1", "schedule.created", "2026-10-06",
		`{"name":"daily backup","cadence":"daily","next_run":"2026-10-07"}`)
	submitClockRaw(t, x, "s-2", "schedule.created", "2026-10-06",
		`{"name":"weekly VAT check","cadence":"weekly","next_run":"2026-10-07"}`)

	// Four days pass unattended. Steady state refuses the burst: nothing
	// opens, nothing fires, the gate names the days.
	events, _ := x.Store.EventsByKind(ctx, core.KindRaw)
	before := len(events)
	res, err := RunClock(ctx, x.Store, x, fixedClock("2026-10-10"))
	if err != nil || !res.Gated || len(res.Pending) != 4 {
		t.Fatalf("gate: %+v, %v", res, err)
	}
	events, _ = x.Store.EventsByKind(ctx, core.KindRaw)
	if len(events) != before {
		t.Fatal("a gated run appended events")
	}

	// The dry run: 4 daily occurrences + 1 weekly = 5 cases would raise,
	// both schedules would advance — and nothing is written.
	diff, days, err := SimulateCatchUp(ctx, x.Store, x, fixedClock("2026-10-10"), "")
	if err != nil || len(days) != 4 {
		t.Fatalf("simulate: %v days, %v", days, err)
	}
	addedCases := 0
	for _, o := range diff.Added {
		if o.Type == "case" {
			addedCases++
		}
	}
	if addedCases != 5 || len(diff.Changed) != 2 {
		t.Fatalf("dry run: %d cases added, %d changed", addedCases, len(diff.Changed))
	}
	if cases, _ := x.Store.ObjectsByType(ctx, "case"); len(cases) != 0 {
		t.Fatal("the dry run wrote")
	}

	// The human act: days open in order, every missed occurrence fires late
	// and in order — two schedules, five occurrences, five cases.
	res, err = CatchUpClock(ctx, x.Store, x, fixedClock("2026-10-10"), "")
	if err != nil || len(res.Opened) != 4 || res.Fired != 5 || len(res.Errors) != 0 {
		t.Fatalf("catch-up: %+v, %v", res, err)
	}
	cases, _ := x.Store.ObjectsByType(ctx, "case")
	if len(cases) != 5 {
		t.Fatalf("cases = %d", len(cases))
	}
	scheds, _ := x.Store.ObjectsByType(ctx, ScheduleType)
	for _, s := range scheds {
		switch s.State["name"] {
		case "daily backup":
			if s.State["next_run"] != "2026-10-11" {
				t.Fatalf("daily next_run = %v", s.State["next_run"])
			}
		case "weekly VAT check":
			if s.State["next_run"] != "2026-10-14" {
				t.Fatalf("weekly next_run = %v", s.State["next_run"])
			}
		}
	}
	// Idempotent: the same catch-up again is a no-op.
	res, _ = CatchUpClock(ctx, x.Store, x, fixedClock("2026-10-10"), "")
	if len(res.Opened) != 0 || res.Fired != 0 {
		t.Fatalf("second catch-up did something: %+v", res)
	}

	// Determinism: time, firings and amendments replay to identical state.
	beforeObjs, _ := x.Store.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	afterObjs, _ := x.Store.AllObjects(ctx)
	bj, _ := json.Marshal(beforeObjs)
	aj, _ := json.Marshal(afterObjs)
	if string(bj) != string(aj) {
		t.Fatal("replay diverged")
	}
}

func TestClockPausedSchedulesDoNotFire(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	seedClockWorld(t, x)

	if _, err := RunClock(ctx, x.Store, x, fixedClock("2026-10-06")); err != nil {
		t.Fatal(err)
	}
	submitClockRaw(t, x, "s-1", "schedule.created", "2026-10-06",
		`{"name":"daily backup","cadence":"daily","next_run":"2026-10-07"}`)
	scheds, _ := x.Store.ObjectsByType(ctx, ScheduleType)
	id := scheds[0].ID

	// Paused by amendment — the switch, per schedule, as data.
	submitClockRaw(t, x, "p-1", "schedule.paused", "2026-10-06",
		fmt.Sprintf(`{"schedule":"%s"}`, id))
	res, err := RunClock(ctx, x.Store, x, fixedClock("2026-10-07"))
	if err != nil || res.Fired != 0 {
		t.Fatalf("paused schedule fired: %+v, %v", res, err)
	}

	// Resumed: the overdue occurrence fires on the next opened day.
	submitClockRaw(t, x, "r-1", "schedule.resumed", "2026-10-07",
		fmt.Sprintf(`{"schedule":"%s"}`, id))
	res, err = RunClock(ctx, x.Store, x, fixedClock("2026-10-08"))
	if err != nil || res.Fired != 1 {
		t.Fatalf("resumed schedule silent: %+v, %v", res, err)
	}
}

func TestCadenceArithmeticClamps(t *testing.T) {
	for _, tc := range [][3]string{
		{"2026-01-31", "monthly", "2026-02-28"},
		{"2026-10-31", "monthly", "2026-11-30"},
		{"2026-10-15", "monthly", "2026-11-15"},
		{"2024-02-29", "yearly", "2025-02-28"},
		{"2026-12-28", "weekly", "2027-01-04"},
	} {
		got, err := nextOccurrence(tc[0], tc[1])
		if err != nil || got != tc[2] {
			t.Errorf("next(%s, %s) = %s, %v; want %s", tc[0], tc[1], got, err, tc[2])
		}
	}
	if _, err := nextOccurrence("2026-01-01", "fortnightly"); err == nil {
		t.Error("unknown cadence accepted")
	}
}
