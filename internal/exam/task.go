package exam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/eval"
	"rhea/internal/store"
)

// A Case is a situational task read into Rhea's shape — an implementation
// in miniature: the company's chart of accounts, the facts as events, the
// instructions as the intents an accountant would give the author, and
// what the examiner's model solution expects of the books, entry by entry.
// The model reads it from the examiner's text; a person checks the file
// before it is trusted, and the file then stands as the case's corpus.
// What the solution answers beyond the entries — balance sheet lines,
// verdicts with their justification — is kept as questions for the host,
// not sat through the kernel yet.
type Case struct {
	Paper          string         `json:"paper"`
	Task           string         `json:"task"`
	Company        string         `json:"company"`
	Accounts       []CaseAccount  `json:"accounts"`
	Events         []CaseEvent    `json:"events"`
	Intents        []string       `json:"intents"`
	ExpectPostings []eval.Posting `json:"expect_postings"`
	Answers        []CaseAnswer   `json:"answers,omitempty"`
	Notes          string         `json:"notes,omitempty"`
}

// CaseAccount is one account of the task's chart, its code as the task
// gives it.
type CaseAccount struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Type string `json:"type"` // asset | liability | equity | revenue | expense
}

// CaseEvent is one business fact of the task as a raw event.
type CaseEvent struct {
	EventType  string         `json:"event_type"`
	OccurredAt string         `json:"occurred_at"`
	Payload    map[string]any `json:"payload"`
}

// CaseAnswer is one item of the model solution that is not an entry.
type CaseAnswer struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

const compileSystem = `You read one situational task of a professional accounting examination into a
bookkeeping system's test case. The task gives a company, its facts and its instructions; the
examiners' model solution gives the entries and answers they expect. Reply with exactly one
JSON object and nothing else:

{"company": "", "accounts": [{"code": "", "name": "", "type": ""}], "events": [{"event_type": "",
"occurred_at": "YYYY-MM-DD", "payload": {}}], "intents": [""], "expect_postings": [{"account": "",
"side": "debit|credit", "amount": ""}], "answers": [{"question": "", "answer": ""}], "notes": ""}

- accounts: every account the task or the solution names, the code exactly as given, the name
  as given, the type one of asset, liability, equity, revenue, expense (from the chart's
  conventions: 0-3 assets, 1 cash and financial assets, 2 settlements by nature, 4-5 costs,
  6 products and accruals, 7 revenues and costs of operations, 8 equity, reserves and result).
- events: one event per business fact whose entry the solution lists, in the order it
  happened — nothing else: a fact the solution books no entry for (a purchase whose entry
  is not asked) is folded into the payload of the event it bears on, or left to notes.
  event_type in lower snake case with a dot, e.g. "shares.purchased", "shares.valued".
  occurred_at as a date: the task's dates; a year written 202X or 20XX is 2024; a valuation at
  the balance sheet date is 2024-12-31. The payload carries every number an entry needs, as the
  task gives them: amounts as decimal strings in złoty ("10000.00" — convert thousands, "tys.
  zł", to złoty), quantities and unit prices as decimal strings, "currency": "PLN" unless the
  task says otherwise. Give the entry's amount itself in the payload when the task's figures
  determine it (for instance "write_down": "10000.00" for 1000 shares bought at 54 and worth 44).
- intents: one per entry or group of entries the solution expects, in the words an accountant
  uses to tell the system what to book: the event type, the payload fields, the accounts by
  code and side, e.g. "When shares.valued arrives, book the write_down: debit 751, credit 145."
  The author drafts rules from these; say everything the rule needs and nothing the task did
  not say.
- expect_postings: every line of every entry the solution gives: account code, side, amount
  as a decimal string in złoty. Where the solution gives a formula, give its result.
- answers: every other item the solution answers — a balance sheet line, a correct/incorrect
  verdict, a justification — as a question and the solution's answer, in the task's language.
- notes: in one or two lines, anything you had to assume.
Never invent facts the task does not state.`

// CompileCase has the model read one task and its model solution into a
// Case. It is the model reading, as it reads a document into a fact: the
// result is data for a person to check, never a rule.
func CompileCase(ctx context.Context, a *agent.Agent, paper Paper, t Task) (Case, error) {
	user := fmt.Sprintf("Paper: %s, %s. Task %s-%d.\n\nTHE TASK:\n%s\n\nTHE MODEL SOLUTION:\n%s\n",
		paper.Subject, paper.Year, t.Block, t.N, t.Text, t.Solution)
	raw, err := a.Complete(ctx, compileSystem, user)
	if err != nil {
		return Case{}, err
	}
	var c Case
	if err := decodeFirst(raw, &c); err != nil {
		return Case{}, err
	}
	c.Paper = strings.TrimSpace(paper.Subject + " " + paper.Year)
	c.Task = fmt.Sprintf("%s-%d", t.Block, t.N)
	return c, c.Check()
}

// Check is what a case needs before it can be sat.
func (c Case) Check() error {
	var errs []string
	if len(c.Accounts) == 0 {
		errs = append(errs, "no accounts")
	}
	for _, a := range c.Accounts {
		switch a.Type {
		case "asset", "liability", "equity", "revenue", "expense":
		default:
			errs = append(errs, fmt.Sprintf("account %s: type %q", a.Code, a.Type))
		}
	}
	if len(c.Events) == 0 {
		errs = append(errs, "no events")
	}
	for _, e := range c.Events {
		if e.EventType == "" || e.OccurredAt == "" {
			errs = append(errs, fmt.Sprintf("event %+v: needs event_type and occurred_at", e))
		}
	}
	if len(c.Intents) == 0 {
		errs = append(errs, "no intents")
	}
	for _, e := range c.Events {
		named := false
		for _, in := range c.Intents {
			if strings.Contains(in, e.EventType) {
				named = true
			}
		}
		if !named {
			errs = append(errs, fmt.Sprintf("event %s: no intent names it — fold it into another event's payload or drop it", e.EventType))
		}
	}
	if len(c.ExpectPostings) == 0 {
		errs = append(errs, "no expected postings")
	}
	for _, p := range c.ExpectPostings {
		if p.Side != "debit" && p.Side != "credit" {
			errs = append(errs, fmt.Sprintf("posting %s: side %q", p.Account, p.Side))
		}
		if _, err := core.ParseMoney(p.Amount); err != nil {
			errs = append(errs, fmt.Sprintf("posting %s: amount %q", p.Account, p.Amount))
		}
	}
	if len(errs) > 0 {
		return errors.New("the case is not whole: " + strings.Join(errs, "; "))
	}
	return nil
}

// ReadCase and WriteCase keep a case as a JSON file a person edits.
func ReadCase(path string) (Case, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Case{}, err
	}
	var c Case
	if err := json.Unmarshal(b, &c); err != nil {
		return Case{}, fmt.Errorf("%s: %w", path, err)
	}
	return c, nil
}

func WriteCase(path string, c Case) error {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

// chartIntent and chartReference are the first task of every case: the
// chart of accounts as master data, worded as the easy corpus words it.
const chartIntent = "Accounts from our chart of accounts become account master data."

var chartReference = json.RawMessage(`{"bundle_id": "chart", "description": "The chart of accounts as master data.", "rules": [{"rule_id": "register-account", "description": "accounts", "priority": 10, "spec": {"match": {"event_type": "account.created"}, "effect": {"object": {"type": "account", "fields": {"code": "=$.code", "name": "=$.name", "type": "=$.type"}}}}}]}`)

// CaseCorpus is the case as eval runs it: the finance base seeded, the
// chart and the facts as intake, one bundle to draft per intent, and the
// books read against the solution. Seeds are relative to the case's file.
func CaseCorpus(c Case, seeds []string) (eval.Corpus, error) {
	cp := eval.Corpus{
		Name:           "exam-" + Slug(c.Paper) + "-" + strings.ToLower(c.Task),
		Seeds:          seeds,
		Expect:         map[string]int{"account": len(c.Accounts)},
		ExpectPostings: c.ExpectPostings,
	}
	// the chart is in place on the day of the first fact
	opening := ""
	for _, e := range c.Events {
		if opening == "" || e.OccurredAt < opening {
			opening = e.OccurredAt
		}
	}
	if opening == "" {
		opening = "2024-01-01"
	}
	for _, a := range c.Accounts {
		b, err := json.Marshal(map[string]any{"event_type": "account.created", "occurred_at": opening,
			"payload": map[string]any{"code": a.Code, "name": a.Name, "type": a.Type}})
		if err != nil {
			return cp, err
		}
		cp.Events = append(cp.Events, b)
	}
	for _, e := range c.Events {
		b, err := json.Marshal(e)
		if err != nil {
			return cp, err
		}
		cp.Events = append(cp.Events, b)
	}
	cp.Tasks = append(cp.Tasks, eval.Task{Intent: chartIntent, Mode: "bundle", Reference: chartReference})
	for _, in := range c.Intents {
		cp.Tasks = append(cp.Tasks, eval.Task{Intent: in, Mode: "bundle"})
	}
	return cp, nil
}

// SitCase sits one case: a fresh throwaway log, the finance base from
// seedsDir, the chart and the facts as intake, the author drafting a bundle
// per intent with the recorded approver reading each dry run, then the
// books read against the model solution. Never the system of record.
func SitCase(ctx context.Context, a *agent.Agent, model, casePath, seedsDir string) (eval.Report, error) {
	c, err := ReadCase(casePath)
	if err != nil {
		return eval.Report{}, err
	}
	if err := c.Check(); err != nil {
		return eval.Report{}, err
	}
	dir := filepath.Dir(casePath)
	rel, err := filepath.Rel(dir, seedsDir)
	if err != nil {
		return eval.Report{}, err
	}
	entries, err := os.ReadDir(seedsDir)
	if err != nil {
		return eval.Report{}, err
	}
	var seeds []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "finance") && strings.HasSuffix(e.Name(), ".json") {
			seeds = append(seeds, filepath.Join(rel, e.Name()))
		}
	}
	sort.Slice(seeds, func(i, j int) bool { return seedLess(seeds[i], seeds[j]) })
	cp, err := CaseCorpus(c, seeds)
	if err != nil {
		return eval.Report{}, err
	}
	st, cleanup, err := store.Throwaway(ctx, "rhea_exam")
	if err != nil {
		return eval.Report{}, err
	}
	defer cleanup()
	return eval.Run(ctx, st, a, model, cp, dir, nil)
}

// seedLess orders finance.json, finance_v2.json … finance_v10.json by
// version, as they must load.
func seedLess(a, b string) bool {
	return seedVersion(a) < seedVersion(b)
}

func seedVersion(path string) int {
	base := strings.TrimSuffix(filepath.Base(path), ".json")
	i := strings.LastIndex(base, "_v")
	if i < 0 {
		return 1
	}
	n := 0
	fmt.Sscanf(base[i+2:], "%d", &n)
	return n
}

// Slug names a paper or a case for a file: lower case, ASCII, dashes.
func Slug(s string) string {
	var sb strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		if t, ok := polish[r]; ok {
			r = t
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
			dash = false
			continue
		}
		if !dash && sb.Len() > 0 {
			sb.WriteByte('-')
			dash = true
		}
	}
	return strings.TrimRight(sb.String(), "-")
}

var polish = map[rune]rune{'ą': 'a', 'ć': 'c', 'ę': 'e', 'ł': 'l', 'ń': 'n', 'ó': 'o', 'ś': 's', 'ź': 'z', 'ż': 'z'}

// decodeFirst decodes the first complete JSON object in a model's text into v.
func decodeFirst(s string, v any) error {
	for i := 0; i < len(s); i++ {
		if s[i] != '{' {
			continue
		}
		if err := json.NewDecoder(strings.NewReader(s[i:])).Decode(v); err == nil {
			return nil
		}
	}
	return fmt.Errorf("no JSON object in the reply: %.400q", s)
}
