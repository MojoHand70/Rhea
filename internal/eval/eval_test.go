package eval_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"rhea/internal/agent"
	"rhea/internal/eval"
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
	rep, err := eval.Run(context.Background(), s, recorded, "reference", c, dir)
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
	rep, err := eval.Run(context.Background(), s, wrong, "wrong", c, dir)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Done() || rep.Rows[0].Error == "" || len(rep.Residue) != rep.Intake {
		t.Fatalf("wrong answer not reported as residue:\n%s", rep)
	}
}
