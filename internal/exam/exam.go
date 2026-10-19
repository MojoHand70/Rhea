// Package exam sits a published examination paper: the examining body's
// questions with its own answer key, answered by the bare model and by the
// host with the pack's rulebook open, and scored the way the body's
// regulation scores. A bench tool like eval — it reads definitions, calls
// the model and prints a report; nothing it does touches the log.
//
// The paper's syntax and its marking rule are a market's (invariant 8): the
// market parses its body's papers into a Paper and provides the Scoring;
// this package knows no examining body by name.
package exam

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// Option is one answer a single-choice question offers.
type Option struct {
	Letter string `json:"letter"`
	Text   string `json:"text"`
}

// Question is one single-choice question: its block on the paper, its
// number within the block, the stem, the options, and Key — the letter
// the examining body marked as right, empty when the paper did not mark it.
type Question struct {
	Block   string   `json:"block"`
	N       int      `json:"n"`
	Stem    string   `json:"stem"`
	Options []Option `json:"options"`
	Key     string   `json:"key,omitempty"`
	// KeyNote says why Key is empty: the paper marked no option, or more
	// than one (the parser's reading of the body's marking, for a person
	// to check against the page).
	KeyNote string `json:"key_note,omitempty"`
}

// Prompt renders the question as the candidate sees it.
func (q Question) Prompt() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d. %s\n", q.N, q.Stem)
	for _, o := range q.Options {
		fmt.Fprintf(&sb, "%s) %s\n", o.Letter, o.Text)
	}
	return sb.String()
}

// Task is one open, situational task: the facts and instructions, and the
// body's model solution. Parsed and counted; not sat by this package yet.
type Task struct {
	Block    string `json:"block"`
	N        int    `json:"n"`
	Text     string `json:"text"`
	Solution string `json:"solution"`
}

// Paper is one published paper.
type Paper struct {
	Subject   string     `json:"subject"`
	Year      string     `json:"year"`
	Source    string     `json:"source"`
	Questions []Question `json:"questions"`
	Tasks     []Task     `json:"tasks"`
}

// Scoring is the marking rule, as the market's regulation sets it: points
// for a right, a wrong and a blank answer, and the share of the maximum a
// block needs to pass.
type Scoring struct {
	Correct   int     `json:"correct"`
	Wrong     int     `json:"wrong"`
	Blank     int     `json:"blank"`
	PassShare float64 `json:"pass_share"`
}

// Where an answer came from. The host says which; the bare model's is
// always its own.
const (
	SourceRule    = "rule"    // a rule of the pack, named, with the citation relied on
	SourceMemory  = "memory"  // the model's own knowledge of the statute — the pack was never taught it
	SourceModel   = "model"   // the bare model, no rulebook
	SourceAbstain = "abstain" // left blank: never taught, not sure
)

// Answer is one answer to one question and how it scored.
type Answer struct {
	Block    string `json:"block"`
	N        int    `json:"n"`
	Letter   string `json:"letter"`
	Source   string `json:"source"`
	RuleID   string `json:"rule_id,omitempty"`
	Citation string `json:"citation,omitempty"`
	Note     string `json:"note,omitempty"`
	Key      string `json:"key,omitempty"`
	Verdict  string `json:"verdict"` // correct | wrong | blank | unkeyed | error
	Points   int    `json:"points"`
	Err      string `json:"error,omitempty"`
}

// An Answerer sits the paper: the bare model, or the host.
type Answerer interface {
	Name() string
	Answer(ctx context.Context, q Question) (Answer, error)
}

// BlockScore is one block's tally: the regulation passes each block on
// its own.
type BlockScore struct {
	Block     string `json:"block"`
	Questions int    `json:"questions"`
	Correct   int    `json:"correct"`
	Wrong     int    `json:"wrong"`
	Blank     int    `json:"blank"`
	Errors    int    `json:"errors"`
	Points    int    `json:"points"`
	Max       int    `json:"max"`
	Passed    bool   `json:"passed"`
}

// Result is one answerer's whole sitting.
type Result struct {
	Answerer string         `json:"answerer"`
	Answers  []Answer       `json:"answers"`
	Blocks   []BlockScore   `json:"blocks"`
	Points   int            `json:"points"`
	Max      int            `json:"max"`
	Passed   bool           `json:"passed"`
	Sources  map[string]int `json:"sources"`
	Unkeyed  int            `json:"unkeyed"`
}

// Report is the sitting of one paper by every answerer.
type Report struct {
	Subject string   `json:"subject"`
	Year    string   `json:"year"`
	Source  string   `json:"source"`
	Model   string   `json:"model"`
	Scoring Scoring  `json:"scoring"`
	Tasks   int      `json:"tasks"`
	Results []Result `json:"results"`
}

// Sit has one answerer answer every question of the paper, parallel calls
// at a time, and scores the sitting. A call that fails is recorded as an
// error on its question and scores as a blank: our failure, never a point.
func Sit(ctx context.Context, paper Paper, scoring Scoring, a Answerer, parallel int) Result {
	if parallel < 1 {
		parallel = 1
	}
	answers := make([]Answer, len(paper.Questions))
	var wg sync.WaitGroup
	sem := make(chan struct{}, parallel)
	for i, q := range paper.Questions {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, q Question) {
			defer wg.Done()
			defer func() { <-sem }()
			ans, err := a.Answer(ctx, q)
			if err != nil {
				ans.Err = err.Error()
				ans.Letter = ""
			}
			answers[i] = ans
		}(i, q)
	}
	wg.Wait()
	for i := range answers {
		answers[i].Block, answers[i].N, answers[i].Key = paper.Questions[i].Block, paper.Questions[i].N, paper.Questions[i].Key
	}
	return Score(a.Name(), answers, scoring)
}

// Score tallies answers against their keys under the marking rule. A
// question the paper did not key is reported but not scored.
func Score(name string, answers []Answer, s Scoring) Result {
	res := Result{Answerer: name, Sources: map[string]int{}}
	blocks := map[string]*BlockScore{}
	var order []string
	for i := range answers {
		a := &answers[i]
		b := blocks[a.Block]
		if b == nil {
			b = &BlockScore{Block: a.Block}
			blocks[a.Block] = b
			order = append(order, a.Block)
		}
		a.Letter = strings.ToLower(strings.TrimSpace(a.Letter))
		if a.Source == "" {
			a.Source = SourceModel
			if a.Letter == "" {
				a.Source = SourceAbstain
			}
		}
		res.Sources[a.Source]++
		if a.Key == "" {
			a.Verdict = "unkeyed"
			a.Points = 0
			res.Unkeyed++
			continue
		}
		b.Questions++
		b.Max += s.Correct
		switch {
		case a.Err != "":
			a.Verdict, a.Points = "error", 0
			b.Errors++
		case a.Letter == "":
			a.Verdict, a.Points = "blank", s.Blank
			b.Blank++
		case a.Letter == a.Key:
			a.Verdict, a.Points = "correct", s.Correct
			b.Correct++
		default:
			a.Verdict, a.Points = "wrong", s.Wrong
			b.Wrong++
		}
		b.Points += a.Points
	}
	res.Passed = len(order) > 0
	for _, k := range order {
		b := blocks[k]
		b.Passed = b.Max > 0 && float64(b.Points) >= s.PassShare*float64(b.Max)
		if !b.Passed {
			res.Passed = false
		}
		res.Blocks = append(res.Blocks, *b)
		res.Points += b.Points
		res.Max += b.Max
	}
	res.Answers = answers
	return res
}

func pct(points, max int) string {
	if max <= 0 {
		return "–"
	}
	return fmt.Sprintf("%d%%", (points*100)/max)
}

// String is the report a person reads: the paper, the rule, and for every
// answerer its blocks, its total, its verdict and where its answers came
// from — with, per source, how many were right.
func (r Report) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s, %s (%s) — test questions only; %d situational task(s) not sat\n", r.Subject, r.Year, r.Source, r.Tasks)
	fmt.Fprintf(&sb, "model %s; marking %+d right / %+d wrong / %+d blank; a block passes at %.0f%% of its maximum\n",
		r.Model, r.Scoring.Correct, r.Scoring.Wrong, r.Scoring.Blank, r.Scoring.PassShare*100)
	for _, res := range r.Results {
		verdict := "NOT PASSED"
		if res.Passed {
			verdict = "PASSED"
		}
		var n, c, w, bl, e int
		for _, b := range res.Blocks {
			n, c, w, bl, e = n+b.Questions, c+b.Correct, w+b.Wrong, bl+b.Blank, e+b.Errors
		}
		fmt.Fprintf(&sb, "\n%s: %d questions — %d right, %d wrong, %d blank", res.Answerer, n, c, w, bl)
		if e > 0 {
			fmt.Fprintf(&sb, ", %d call error(s)", e)
		}
		if res.Unkeyed > 0 {
			fmt.Fprintf(&sb, ", %d unkeyed (not scored)", res.Unkeyed)
		}
		fmt.Fprintf(&sb, " → %d of %d points (%s) — %s\n", res.Points, res.Max, pct(res.Points, res.Max), verdict)
		for _, b := range res.Blocks {
			v := "passed"
			if !b.Passed {
				v = "not passed"
			}
			fmt.Fprintf(&sb, "  block %-4s %3d questions: %3d right %3d wrong %3d blank → %4d of %3d (%s) %s\n",
				b.Block, b.Questions, b.Correct, b.Wrong, b.Blank, b.Points, b.Max, pct(b.Points, b.Max), v)
		}
		// per source: how many, and how many of those were right
		type tally struct{ n, right, wrong int }
		by := map[string]*tally{}
		for _, a := range res.Answers {
			t := by[a.Source]
			if t == nil {
				t = &tally{}
				by[a.Source] = t
			}
			t.n++
			switch a.Verdict {
			case "correct":
				t.right++
			case "wrong":
				t.wrong++
			}
		}
		keys := make([]string, 0, len(by))
		for k := range by {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		sb.WriteString("  where the answers came from:")
		for _, k := range keys {
			t := by[k]
			if k == SourceAbstain {
				fmt.Fprintf(&sb, " %s %d;", k, t.n)
				continue
			}
			fmt.Fprintf(&sb, " %s %d (%d right, %d wrong);", k, t.n, t.right, t.wrong)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// reply is what both answerers ask the model for: one JSON object.
type reply struct {
	Answer   string `json:"answer"`
	Source   string `json:"source"`
	RuleID   string `json:"rule_id"`
	Citation string `json:"citation"`
	Note     string `json:"note"`
}

// decodeReply tolerates a model that wraps its JSON in prose or a fence: it
// decodes the first complete JSON object in the text. When no object
// decodes — a citation with a raw line break or a bare quote inside it —
// it reads the fields one by one, the answer letter strictly; a reply
// without an "answer" field is an error that keeps the raw text.
func decodeReply(s string) (reply, error) {
	for i := 0; i < len(s); i++ {
		if s[i] != '{' {
			continue
		}
		var r reply
		if err := json.NewDecoder(strings.NewReader(s[i:])).Decode(&r); err == nil {
			return r, nil
		}
	}
	if m := reAnswerField.FindStringSubmatch(s); m != nil {
		return reply{Answer: m[1], Source: field(s, "source"), RuleID: field(s, "rule_id"), Citation: field(s, "citation"), Note: field(s, "note")}, nil
	}
	return reply{}, fmt.Errorf("no JSON object in the reply: %.400q", s)
}

var (
	reAnswerField = regexp.MustCompile(`"answer"\s*:\s*"([a-cA-C]?)"`)
	reFields      = map[string]*regexp.Regexp{}
)

func field(s, name string) string {
	re := reFields[name]
	if re == nil {
		re = regexp.MustCompile(`"` + name + `"\s*:\s*"((?:[^"\\]|\\.)*)"`)
		reFields[name] = re
	}
	if m := re.FindStringSubmatch(s); m != nil {
		return strings.TrimSpace(strings.ReplaceAll(m[1], "\n", " "))
	}
	return ""
}
