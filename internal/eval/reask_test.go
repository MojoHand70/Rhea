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

// The recorded approver gives the second ask a person gives: a refused
// draft goes back with its reason, once, and the next draft answers it. The
// verdict counts first drafts and re-asks rather than failing the run on a
// refusal that booked nothing wrong.
func TestApproverReasksOnce(t *testing.T) {
	c, dir, err := eval.Load("../../testdata/eval/operations.json")
	if err != nil {
		t.Fatal(err)
	}
	c.Tasks = c.Tasks[:1] // master data, as a bundle
	s := storetest.New(t)
	asks := 0
	a := &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
		asks++
		if !strings.Contains(user, "The conversation so far") {
			// first draft: a bundle the gate refuses — a view on a field no type has
			return `{"bundle_id":"wrong","description":"a view on nothing","warrant":{"basis":"client"},
			  "view_defs":[{"view_id":"v","notion":"list","title":"T","domain":"d","function":"f",
			  "spec":{"object_type":"account","columns":[{"field":"no_such_field","label":"x"}]}}]}`, nil
		}
		if !strings.Contains(user, "Refused by the approver: bundle failed validation") || !strings.Contains(user, `"bundle_id":"wrong"`) {
			return "", fmt.Errorf("the re-ask must carry the refusal and the refused proposal:\n%s", user)
		}
		for _, task := range c.Tasks {
			if strings.HasSuffix(strings.TrimSpace(user), "User intent: "+task.Intent) {
				return string(task.Reference), nil
			}
		}
		return "", fmt.Errorf("no reference")
	}}
	rep, err := eval.Run(context.Background(), s, a, "reasking", c, dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	row := rep.Rows[0]
	if asks != 2 || len(row.Reasks) != 1 || !row.Approved || !row.Fixed() || row.FirstDraft() {
		t.Fatalf("asks=%d row=%+v", asks, row)
	}
	if !strings.Contains(rep.String(), "re-asked: bundle failed validation") {
		t.Fatalf("report does not show the re-ask:\n%s", rep)
	}
	v := eval.Verdict{Corpus: c.Name, Outcomes: []eval.Outcome{{Voice: 0, Run: 1, Report: rep}}}
	if cnt := v.Counts(); cnt != (eval.Counts{Questions: 1, FirstDraft: 0, Reasked: 1, Fixed: 1}) {
		t.Fatalf("counts = %+v", cnt)
	}
	if !strings.Contains(v.String(), "re-asked 1") || !strings.Contains(v.String(), "first draft right on 0 of 1 questions; 1 re-asked, 1 of those fixed") {
		t.Fatalf("verdict = %s", v)
	}
}

// Two wrong drafts in a row: the approver asks again once, not forever.
func TestApproverGivesUpAfterOneReask(t *testing.T) {
	c, dir, err := eval.Load("../../testdata/eval/operations.json")
	if err != nil {
		t.Fatal(err)
	}
	c.Tasks = c.Tasks[:1]
	s := storetest.New(t)
	asks := 0
	a := &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
		asks++
		return `not json at all`, nil
	}}
	rep, err := eval.Run(context.Background(), s, a, "stubborn", c, dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if asks != 2 || len(rep.Rows[0].Reasks) != 1 || rep.Rows[0].Approved || rep.Done() {
		t.Fatalf("asks=%d row=%+v done=%v", asks, rep.Rows[0], rep.Done())
	}
}
