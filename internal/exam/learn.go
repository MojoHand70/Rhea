package exam

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"rhea/internal/agent"
	"rhea/internal/eval"
	"rhea/internal/statute"
	"rhea/internal/store"
)

// The learning-rate test (KK, 2026-10-19 evening: "we will not finish by
// 2030"): does the pack's own knowledge grow at the author's pace, not a
// person's? A prior sitting named every lesson the pack lacks — each answer
// not from a rule, with the provision the host read. Learn hands those
// lessons to the author as intents, lets it draft real definitions on a
// throwaway database (the recorded approver reading each dry run, since
// this is a measurement and never the live pack), then sits the same
// questions again with the rulebook read from that database. The share of
// answers from rules before and after, against the clock, is the result.

// Lesson is one thing the pack was not taught, as the exam named it.
type Lesson struct {
	Block    string `json:"block"`
	N        int    `json:"n"`
	Intent   string `json:"intent"`
	Citation string `json:"citation"`
}

// Lessons lists the lessons of one block of a prior sitting: the answers
// not from a rule, each with the question, the key and the provision.
func Lessons(subject string, paper Paper, prior Result, block string) []Lesson {
	byKey := map[string]Question{}
	for _, q := range paper.Questions {
		byKey[q.Block+"-"+fmt.Sprint(q.N)] = q
	}
	var out []Lesson
	for _, a := range prior.Answers {
		if a.Source == SourceRule || (block != "" && a.Block != block) {
			continue
		}
		q, ok := byKey[a.Block+"-"+fmt.Sprint(a.N)]
		if !ok || q.Key == "" {
			continue
		}
		right := ""
		for _, o := range q.Options {
			if o.Letter == q.Key {
				right = o.Text
			}
		}
		cite := strings.TrimSpace(a.Citation)
		if cite == "" {
			cite = strings.TrimSpace(a.Note)
		}
		intent := fmt.Sprintf("A lesson from the auditor's examination (%s, question %s-%d). The question: %s "+
			"The right answer: %s) %s. The provision it rests on: %s. "+
			"Teach the Poland pack this lesson as the definitions a small company's books need so that they follow this provision: "+
			"a door for the decision a person makes if there is one, the rule that books it on the catalog's accounts (or records the fact as an object when nothing is booked), "+
			"and each rule's warrant quoting the provision. Nothing that leaves the books unchanged; no rule written only to restate the answer.",
			subject, a.Block, a.N, strings.TrimSpace(q.Stem), q.Key, right, cite)
		out = append(out, Lesson{Block: a.Block, N: a.N, Intent: intent, Citation: cite})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Block != out[j].Block {
			return out[i].Block < out[j].Block
		}
		return out[i].N < out[j].N
	})
	return out
}

// LearnOptions says what the test runs on.
type LearnOptions struct {
	Packs    []string // pack files the throwaway starts from, relative to the working directory
	SeedsDir string   // the finance base
	Library  *statute.Library
	Parallel int
	Limit    int // at most this many lessons; 0 = all
	Verbose  io.Writer
	// Keep, when set, is where the author's drafts go as a file for a
	// person to review: one entry per lesson with the bundle as proposed and
	// whether the recorded approver let it through. The throwaway is dropped;
	// this file is what becomes the next pack version after review.
	Keep string
}

// KeptLesson is one lesson with the author's answer to it.
type KeptLesson struct {
	Block    string          `json:"block"`
	N        int             `json:"n"`
	Citation string          `json:"citation"`
	Intent   string          `json:"intent"`
	Bundle   string          `json:"bundle"`
	Approved bool            `json:"approved"`
	Reasks   []string        `json:"reasks,omitempty"`
	Errors   []string        `json:"errors,omitempty"`
	Draft    json.RawMessage `json:"draft,omitempty"`
}

// LearnResult is the measurement.
type LearnResult struct {
	Lessons  int
	Drafted  eval.Report
	Before   Result
	After    Result
	Elapsed  time.Duration
	RulesNow int
}

// Learn runs the test on one block of a paper.
func Learn(ctx context.Context, a *agent.Agent, model string, subject string, paper Paper, prior Result, block string, scoring Scoring, opt LearnOptions, out io.Writer) (LearnResult, error) {
	start := time.Now()
	lessons := Lessons(subject, paper, prior, block)
	if opt.Limit > 0 && len(lessons) > opt.Limit {
		lessons = lessons[:opt.Limit]
	}
	res := LearnResult{Lessons: len(lessons)}
	if len(lessons) == 0 {
		return res, fmt.Errorf("no lessons: every answer of block %q already rests on a rule", block)
	}
	// the block, as a paper of its own
	sub := Paper{Subject: paper.Subject, Year: paper.Year, Source: paper.Source}
	for _, q := range paper.Questions {
		if block == "" || q.Block == block {
			sub.Questions = append(sub.Questions, q)
		}
	}
	var beforeAnswers []Answer
	for _, ans := range prior.Answers {
		if block == "" || ans.Block == block {
			beforeAnswers = append(beforeAnswers, ans)
		}
	}
	res.Before = Score("host, before", beforeAnswers, scoring)

	// the corpus: no events, one bundle to draft per lesson, the pack installed first
	seeds, err := financeSeeds(opt.SeedsDir, ".")
	if err != nil {
		return res, err
	}
	cp := eval.Corpus{Name: "learn-" + Slug(subject) + "-" + strings.ToLower(block), Seeds: seeds, Packs: opt.Packs}
	for _, l := range lessons {
		cp.Tasks = append(cp.Tasks, eval.Task{Intent: l.Intent, Mode: "bundle"})
	}
	st, cleanup, err := store.Throwaway(ctx, "rhea_learn")
	if err != nil {
		return res, err
	}
	defer cleanup()
	fmt.Fprintf(out, "the author drafts %d lesson(s) on a throwaway database…\n", len(lessons))
	rep, err := eval.Run(ctx, st, a, model, cp, ".", nil)
	if err != nil {
		return res, err
	}
	res.Drafted = rep
	if opt.Verbose != nil {
		fmt.Fprint(opt.Verbose, rep.Verbose())
	}
	if opt.Keep != "" {
		var kept []KeptLesson
		for i, row := range rep.Rows {
			k := KeptLesson{Intent: row.Intent, Bundle: row.RuleID, Approved: row.Approved && len(row.ApproveErrs) == 0, Reasks: row.Reasks, Draft: row.Draft}
			if row.Error != "" {
				k.Errors = append(k.Errors, row.Error)
			}
			k.Errors = append(k.Errors, row.ApproveErrs...)
			if i < len(lessons) {
				k.Block, k.N, k.Citation = lessons[i].Block, lessons[i].N, lessons[i].Citation
			}
			kept = append(kept, k)
		}
		b, err := json.MarshalIndent(kept, "", " ")
		if err != nil {
			return res, err
		}
		if err := os.MkdirAll(filepath.Dir(opt.Keep), 0o755); err != nil {
			return res, err
		}
		if err := os.WriteFile(opt.Keep, b, 0o644); err != nil {
			return res, err
		}
		fmt.Fprintf(out, "the author's drafts are kept in %s for review\n", opt.Keep)
	}
	rules, err := st.ActiveRules(ctx)
	if err != nil {
		return res, err
	}
	types, err := st.ListObjectTypes(ctx)
	if err != nil {
		return res, err
	}
	res.RulesNow = len(rules)
	host := NewHost(a, rules, types)
	host.Library = opt.Library
	fmt.Fprintf(out, "the rulebook now holds %d active rules; the host sits block %s again (%d questions)…\n", len(rules), block, len(sub.Questions))
	res.After = Sit(ctx, sub, scoring, host, opt.Parallel)
	res.After.Answerer = "host, after"
	res.Elapsed = time.Since(start)
	return res, nil
}

// financeSeeds lists the finance base's seed files, relative to dir, in
// version order.
func financeSeeds(seedsDir, dir string) ([]string, error) {
	rel, err := filepath.Rel(dir, seedsDir)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(seedsDir)
	if err != nil {
		return nil, err
	}
	var seeds []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "finance") && strings.HasSuffix(e.Name(), ".json") {
			seeds = append(seeds, filepath.Join(rel, e.Name()))
		}
	}
	sort.Slice(seeds, func(i, j int) bool { return seedLess(seeds[i], seeds[j]) })
	return seeds, nil
}

// String is the measurement a person reads.
func (r LearnResult) String() string {
	var sb strings.Builder
	approved, refused := 0, 0
	for _, row := range r.Drafted.Rows {
		if row.Approved && len(row.ApproveErrs) == 0 {
			approved++
		} else {
			refused++
		}
	}
	fmt.Fprintf(&sb, "lessons handed to the author: %d; bundles approved on the throwaway: %d, refused or failed: %d; active rules after: %d; %s elapsed\n",
		r.Lessons, approved, refused, r.RulesNow, r.Elapsed.Round(time.Second))
	for _, res := range []Result{r.Before, r.After} {
		fmt.Fprintf(&sb, "%-13s %3d of %3d points; from rules %d (%d right)", res.Answerer+":", res.Points, res.Max, res.Sources[SourceRule], rightFrom(res, SourceRule))
		fmt.Fprintf(&sb, "; statute %d, memory %d, abstain %d\n", res.Sources[SourceStatute], res.Sources[SourceMemory], res.Sources[SourceAbstain])
	}
	return sb.String()
}

func rightFrom(res Result, source string) int {
	n := 0
	for _, a := range res.Answers {
		if a.Source == source && a.Verdict == "correct" {
			n++
		}
	}
	return n
}
