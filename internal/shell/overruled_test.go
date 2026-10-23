package shell_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

// The standing test (DIRECTION 2026-10-11, "The host remembers"): the
// overruled rule never returns. Approve a reading; supersede it with a
// different one, with the reason; then: the old reading never fires after
// its time and replay reproduces both as they happened; the author's ask
// names the new reading as the active rule and shows the old one only as
// superseded, with the reason — so the model is told, in so many words,
// never to propose it again.
func TestOverruledRuleNeverReturns(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	srv := &shell.Server{Store: s, Exec: x, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb")}
	ts := serve(t, srv.Handler())
	post := func(path string, body any) {
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
	}
	if err := s.InsertObjectType(ctx, core.ObjectType{Name: "case", Version: 1, Domain: "work", LabelField: "subject",
		Fields: []core.FieldDef{
			{Name: "subject", Type: "string", Required: true},
			{Name: "priority", Type: "enum", Required: true, Values: []string{"normal", "high"}}}}); err != nil {
		t.Fatal(err)
	}
	reading := func(priority string) core.RuleSpec {
		return core.RuleSpec{Match: core.Match{EventType: "complaint.registered"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "case",
				Fields: map[string]string{"subject": "=$.number", "priority": priority}}}}
	}
	draft := func(desc string, spec core.RuleSpec) {
		t.Helper()
		if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: "raise-case", Status: core.StatusDraft, Priority: 100,
			EffectiveFrom: "2026-01-01", CreatedBy: "human", Description: desc, Spec: spec}); err != nil {
			t.Fatal(err)
		}
	}
	submit := func(dedup, number string) {
		t.Helper()
		post("/api/events", map[string]any{"event_type": "complaint.registered", "occurred_at": "2026-09-10",
			"dedup_key": dedup, "payload": map[string]any{"number": number}})
	}

	// The first reading: every complaint is a normal-priority case.
	draft("every complaint opens a normal case", reading("normal"))
	post("/api/rules/raise-case/approve", map[string]any{"approved_by": "krzysztof"})
	submit("r1", "R-1")

	// The correction: the owner changes the reading, and says why.
	draft("every complaint opens a high-priority case", reading("high"))
	post("/api/rules/raise-case/approve", map[string]any{"approved_by": "krzysztof",
		"reason": "complaints are urgent for us; normal was wrong"})
	submit("r2", "R-2")

	// Acting: the old reading explained R-1 and never fires again; the new
	// one explains R-2. Replay reproduces both as they happened.
	cases, _ := s.ObjectsByType(ctx, "case")
	if len(cases) != 2 || cases[0].State["priority"] != "normal" || cases[0].RuleVersion != 2 ||
		cases[1].State["priority"] != "high" || cases[1].RuleVersion != 4 {
		t.Fatalf("cases: %+v", cases)
	}
	active, _ := s.ActiveRules(ctx)
	if len(active) != 1 || active[0].Version != 4 || active[0].Spec.Effect.Object.Fields["priority"] != "high" {
		t.Fatalf("the active rule: %+v", active)
	}
	before, _ := s.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := s.AllObjects(ctx)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("replay does not reproduce the two readings as they happened")
	}

	// Drafting: the ask names the new reading as the active rule and shows
	// the old one only as superseded, with the reason; the model is told.
	ask, err := shell.DraftAsk(ctx, s, nil, "a question about complaints", core.Event{}, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(ask.Rules) != 1 || ask.Rules[0].Version != 4 {
		t.Fatalf("the ask's active rules: %+v", ask.Rules)
	}
	if len(ask.Corrections) != 1 {
		t.Fatalf("the ask's corrections: %+v", ask.Corrections)
	}
	c := ask.Corrections[0]
	if c.RuleID != "raise-case" || c.Version != 2 || c.ReplacedBy != 4 || c.By != "krzysztof" ||
		c.Description != "every complaint opens a normal case" || c.Reason != "complaints are urgent for us; normal was wrong" {
		t.Fatalf("the correction: %+v", c)
	}
	var told string
	a := &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
		told = user
		return "", fmt.Errorf("captured")
	}}
	a.DraftRule(ctx, ask)
	if !strings.Contains(told, "Readings the owner corrected before - superseded, never to be proposed again:") ||
		!strings.Contains(told, `raise-case v2 ("every complaint opens a normal case") was replaced by v4 by krzysztof: complaints are urgent for us; normal was wrong`) {
		t.Fatalf("the author is not told: %s", told)
	}
	if strings.Contains(told, "- raise-case p100") && strings.Count(told, "raise-case") != 2 {
		t.Fatalf("the old reading leaks into the active rules: %s", told)
	}
}
