package exam

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// scripted answers by block and number; what it does not know it leaves blank.
type scripted struct {
	letters map[string]string
	sources map[string]string
}

func (scripted) Name() string { return "scripted" }
func (s scripted) Answer(_ context.Context, q Question) (Answer, error) {
	k := fmt.Sprintf("%s-%d", q.Block, q.N)
	if s.letters[k] == "error" {
		return Answer{}, fmt.Errorf("the call failed")
	}
	return Answer{Letter: s.letters[k], Source: s.sources[k]}, nil
}

var marking = Scoring{Correct: 2, Wrong: -1, Blank: 0, PassShare: 0.6}

func TestScoringPassesEachBlockOnItsOwn(t *testing.T) {
	paper := Paper{Subject: "S", Year: "2025", Questions: []Question{
		{Block: "I", N: 1, Key: "a"}, {Block: "I", N: 2, Key: "b"}, {Block: "I", N: 3, Key: "c"},
		{Block: "II", N: 1, Key: "a"}, {Block: "II", N: 2, Key: "b"},
		{Block: "II", N: 3}, // the paper did not mark this one
	}}
	a := scripted{
		letters: map[string]string{"I-1": "a", "I-2": "c", "I-3": "", "II-1": "A", "II-2": "b", "II-3": "a"},
		sources: map[string]string{"I-1": SourceRule, "I-2": SourceMemory, "II-1": SourceRule, "II-2": SourceMemory, "II-3": SourceMemory},
	}
	res := Sit(context.Background(), paper, marking, a, 3)
	if len(res.Blocks) != 2 {
		t.Fatalf("blocks: %+v", res.Blocks)
	}
	// block I: right, wrong, blank → 2 − 1 + 0 = 1 of 6: not passed
	if b := res.Blocks[0]; b.Points != 1 || b.Max != 6 || b.Passed {
		t.Errorf("block I: %+v", b)
	}
	// block II: two right of two keyed → 4 of 4: passed; the unkeyed one not scored
	if b := res.Blocks[1]; b.Points != 4 || b.Max != 4 || !b.Passed || b.Questions != 2 {
		t.Errorf("block II: %+v", b)
	}
	if res.Points != 5 || res.Max != 10 || res.Passed || res.Unkeyed != 1 {
		t.Errorf("total: points %d of %d passed %v unkeyed %d", res.Points, res.Max, res.Passed, res.Unkeyed)
	}
	if res.Sources[SourceRule] != 2 || res.Sources[SourceMemory] != 3 || res.Sources[SourceAbstain] != 1 {
		t.Errorf("sources: %v", res.Sources)
	}
	rep := Report{Subject: "S", Year: "2025", Scoring: marking, Results: []Result{res}}
	out := rep.String()
	for _, want := range []string{"5 of 10 points (50%)", "NOT PASSED", "block I ", "not passed", "rule 2 (2 right, 0 wrong)", "memory 3 (1 right, 1 wrong)", "abstain 1"} {
		if !strings.Contains(out, want) {
			t.Errorf("report lacks %q:\n%s", want, out)
		}
	}
}

func TestACallThatFailsScoresAsABlank(t *testing.T) {
	paper := Paper{Questions: []Question{{Block: "I", N: 1, Key: "a"}, {Block: "I", N: 2, Key: "b"}}}
	res := Sit(context.Background(), paper, marking, scripted{letters: map[string]string{"I-1": "error", "I-2": "b"}}, 1)
	if res.Points != 2 || res.Max != 4 || res.Blocks[0].Errors != 1 || res.Answers[0].Verdict != "error" || res.Answers[0].Err == "" {
		t.Errorf("%+v", res)
	}
}

func TestDecodeReplyTakesTheFirstObjectInTheText(t *testing.T) {
	r, err := decodeReply("Sure. ```json\n{\"answer\": \"b\", \"source\": \"rule\", \"rule_id\": \"x\", \"note\": \"n\"}\n``` done")
	if err != nil || r.Answer != "b" || r.Source != "rule" || r.RuleID != "x" {
		t.Fatalf("%+v %v", r, err)
	}
	if _, err := decodeReply("no object here"); err == nil {
		t.Fatal("expected an error")
	}
}

func TestDecodeReplyReadsFieldsWhenTheObjectIsBroken(t *testing.T) {
	// a line break inside the citation breaks the JSON; the fields still read
	broken := "{\"answer\": \"c\", \"source\": \"memory\", \"rule_id\": \"\", \"citation\": \"MSR 8 „Zasady\nrachunkowości” par. 42\", \"note\": \"the pack lacks error correction\"}"
	r, err := decodeReply(broken)
	if err != nil || r.Answer != "c" || r.Source != "memory" || r.Note != "the pack lacks error correction" || !strings.Contains(r.Citation, "par. 42") {
		t.Fatalf("%+v %v", r, err)
	}
	if r, err := decodeReply("{\"answer\": \"\", \"source\": \"abstain\", \"note\": \"x\n\"}"); err != nil || r.Answer != "" || r.Source != "abstain" {
		t.Fatalf("blank answer: %+v %v", r, err)
	}
	if _, err := decodeReply("{\"answer\": \"maybe b\", \"note\": \"x\n\"}"); err == nil {
		t.Fatal("a broken object with a non-letter answer must not pass as an answer")
	}
}
