package eval_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"rhea/internal/agent"
	"rhea/internal/eval"
	"rhea/internal/sim"
	"rhea/internal/store/storetest"
)

// The corpus is solvable: its reference answers, played back as the agent
// through the same draft → validate → simulate → approve path, explain the
// whole intake and land the expected state. A corpus whose own references
// fail would measure nothing.
func TestCorpusReferenceIsDone(t *testing.T) {
	for _, corpus := range []string{"corpus.json", "operations.json"} {
		t.Run(corpus, func(t *testing.T) { referenceIsDone(t, corpus) })
	}
}

func referenceIsDone(t *testing.T, corpus string) {
	c, dir, err := eval.Load("../../testdata/eval/" + corpus)
	if err != nil {
		t.Fatal(err)
	}
	s := storetest.New(t)
	recorded := &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
		for _, task := range c.Tasks {
			if strings.HasSuffix(strings.TrimSpace(user), "User intent: "+task.Intent) {
				return string(task.Reference), nil
			}
		}
		return "", fmt.Errorf("no reference for this ask")
	}}
	rep, err := eval.Run(context.Background(), s, recorded, "reference", c, dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("\n" + rep.String())
	if !rep.Done() {
		t.Fatalf("reference run not done:\n%s", rep)
	}
}

// A wrong answer shows up as residue, not as a crash: the run keeps going and
// the verdict says NOT DONE.
func TestRefusedDraftIsResidue(t *testing.T) {
	c, dir, err := eval.Load("../../testdata/eval/corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	c.Tasks = c.Tasks[:1]
	s := storetest.New(t)
	wrong := &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
		return `{"rule_id":"x","description":"d","spec":{"match":{"event_type":"account.created"},"effect":{"object":{"type":"ledger_account","fields":{"a":"=$.code"}}}}}`, nil
	}}
	rep, err := eval.Run(context.Background(), s, wrong, "wrong", c, dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Done() || rep.Rows[0].Error == "" || len(rep.Residue) != rep.Intake {
		t.Fatalf("wrong answer not reported as residue:\n%s", rep)
	}
}

// A simulated customer is solvable the same way: its scripted interview
// answers, played back as the agent, explain the whole generated intake and
// land the state the generator computed independently of any rule — counts,
// lifecycle moves, and the sums of computed values.
func TestPersonaReferenceIsDone(t *testing.T) {
	for _, name := range []string{"nordwind", "helios"} {
		t.Run(name, func(t *testing.T) {
			p, dir, err := sim.Load("../../testdata/customers/" + name + ".json")
			if err != nil {
				t.Fatal(err)
			}
			p.Months = 2 // the generator is uniform per month; two keep the test quick
			c, facts, err := sim.Compile(p, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			t.Log("\n" + sim.Summary(p, facts))
			s := storetest.New(t)
			recorded := &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
				for _, task := range c.Tasks {
					if strings.HasSuffix(strings.TrimSpace(user), "User intent: "+task.Intent) {
						return string(task.Reference), nil
					}
				}
				return "", fmt.Errorf("no reference for this ask")
			}}
			rep, err := eval.Run(context.Background(), s, recorded, "reference", c, dir, nil)
			if err != nil {
				t.Fatal(err)
			}
			t.Log("\n" + rep.String())
			if !rep.Done() {
				t.Fatalf("persona %s not done:\n%s", name, rep)
			}
		})
	}
}

// The verdict over several runs is all or nothing: an author that passes
// most of the time is not done.
func TestVerdictIsAllOrNothing(t *testing.T) {
	done := eval.Report{Intake: 3}
	notDone := eval.Report{Intake: 3, Residue: []string{"x #1"}}
	v := eval.Verdict{Corpus: "sim:x", Outcomes: []eval.Outcome{{Voice: 0, Run: 1, Report: done}, {Voice: 1, Run: 1, Report: done}}}
	if !v.Done() {
		t.Fatal("every run done, verdict not done")
	}
	v.Outcomes = append(v.Outcomes, eval.Outcome{Voice: 1, Run: 2, Report: notDone})
	if v.Done() || !strings.Contains(v.String(), "NOT DONE: 2 of 3 runs done") {
		t.Fatalf("verdict = %s", v)
	}
	if (eval.Verdict{}).Done() {
		t.Fatal("no runs is not done")
	}
}

// A refused draft stays readable: the gate's verdict comes with the
// evidence, so an authoring failure can be read without rerunning.
func TestRefusedDraftIsKept(t *testing.T) {
	c, dir, err := eval.Load("../../testdata/eval/corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	c.Tasks = c.Tasks[:1]
	s := storetest.New(t)
	wrong := &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
		return `{"rule_id":"x","description":"d","spec":{"match":{"event_type":"account.created"},"effect":{"object":{"type":"ledger_account","fields":{"a":"=$.code"}}}}}`, nil
	}}
	rep, err := eval.Run(context.Background(), s, wrong, "wrong", c, dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Rows[0].Error == "" || !strings.Contains(string(rep.Rows[0].Draft), "ledger_account") || !strings.Contains(rep.Verbose(), "ledger_account") {
		t.Fatalf("refused draft not kept: %+v", rep.Rows[0])
	}
}
