// Package eval measures the agent as an author: a fixed corpus of intake
// events and plain-language intents runs through draft → simulate → approve,
// and the result is the number phase (a) is built around — worklist residue
// (DIRECTION, "Implementation by interview"). An implementation in miniature,
// on a throwaway database: the system of record is append-only, so a run that
// must leave no trace gets its own log.
package eval

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/network"
	"rhea/internal/shell"
	"rhea/internal/store"
)

// Corpus is an implementation script as data: the definitions in place, the
// client's events as intake, the questions a human answers, the state a
// correct implementation ends in.
type Corpus struct {
	Name string `json:"name"`
	// Seeds are definition files ({object_types, view_defs}), relative to the
	// corpus file — the vocabulary the agent drafts against.
	Seeds []string `json:"seeds"`
	// Events are the intake: a string is an event file relative to the corpus
	// file, an object is an inline event {event_type, occurred_at, payload}.
	Events []json.RawMessage `json:"events"`
	Tasks  []Task            `json:"tasks"`
	// Expect is the object count per type after every task — what a correct
	// implementation of this corpus materializes.
	Expect map[string]int `json:"expect"`
	// ExpectWhere counts objects by one field's value: a lifecycle that
	// moved, a book that was posted to — state a bare count cannot see.
	ExpectWhere []Where `json:"expect_where,omitempty"`
}

// Where expects Count objects of Type whose Field equals Value.
type Where struct {
	Type  string `json:"type"`
	Field string `json:"field"`
	Value string `json:"value"`
	Count int    `json:"count"`
}

// Task is one answer a human gives in the interview. The sample is the first
// raw event of SampleType in the log (optional for bundles: the agent sees
// the whole residue). Mode "bundle" asks for a bundle — everything the
// answer needs, approved whole — instead of one rule. Reference is a
// known-good answer of the same mode: it proves the corpus solvable and
// answers for the recorded agent in tests; the live agent never sees it.
type Task struct {
	Intent     string          `json:"intent"`
	Mode       string          `json:"mode,omitempty"` // "" (one rule) | "bundle"
	SampleType string          `json:"sample_type,omitempty"`
	Hint       string          `json:"hint,omitempty"`
	Reference  json.RawMessage `json:"reference"`
}

type rawEvent struct {
	EventType  string         `json:"event_type"`
	OccurredAt string         `json:"occurred_at"`
	Payload    map[string]any `json:"payload"`
}

// Row is one task's outcome.
type Row struct {
	Intent      string
	RuleID      string // the rule, or the bundle with what it proposes
	Error       string // drafting or validation failed; nothing stored
	SimErrors   []string
	Explains    int // raw events the draft would explain
	Added       int // objects the dry run would add
	Approved    bool
	Booked      int // events booked by approval
	ApproveErrs []string
	Draft       json.RawMessage // what the agent proposed, for the verbose report
	Backfilled  int             // past events the approval's backfill explained further
}

// Report is the run's verdict.
type Report struct {
	Corpus   string
	Model    string
	Intake   int
	Residue  []string // raw events left unexplained, "type #id"
	Rows     []Row
	Objects  map[string]int
	Expect   map[string]int
	Mismatch []string // per type: "posting: 10 of 12"
}

// Done is the implementation's definition of done: nothing unexplained, and
// the state a correct implementation would hold.
func (r Report) Done() bool { return len(r.Residue) == 0 && len(r.Mismatch) == 0 }

// Load reads a corpus file.
func Load(path string) (Corpus, string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Corpus{}, "", err
	}
	var c Corpus
	if err := json.Unmarshal(b, &c); err != nil {
		return Corpus{}, "", fmt.Errorf("%s: %w", path, err)
	}
	return c, filepath.Dir(path), nil
}

// Run plays the corpus against a fresh store: seeds, intake through the open
// door, then per task ask → validate → store draft → dry run → approve when
// the dry run shows no errors (the human's role, played by a policy; actor
// "eval"). The agent sees only what the shell would show it.
func Run(ctx context.Context, st *store.Store, a *agent.Agent, model string, c Corpus, dir string, k *network.Knowledge) (Report, error) {
	x := &exec.Executor{Store: st}
	rep := Report{Corpus: c.Name, Model: model, Expect: c.Expect}
	for _, seed := range c.Seeds {
		if err := loadSeed(ctx, st, filepath.Join(dir, seed)); err != nil {
			return rep, err
		}
	}
	for i, raw := range c.Events {
		ev, err := readEvent(raw, dir)
		if err != nil {
			return rep, fmt.Errorf("event %d: %w", i, err)
		}
		if _, err := x.TriggerActivity(ctx, "submit_event", map[string]any{
			"event_type": ev.EventType, "occurred_at": ev.OccurredAt, "payload": ev.Payload,
		}, "eval:intake", ""); err != nil {
			return rep, fmt.Errorf("intake %d (%s): %w", i, ev.EventType, err)
		}
		rep.Intake++
	}
	raws, err := st.EventsByKind(ctx, core.KindRaw)
	if err != nil {
		return rep, err
	}
	for _, task := range c.Tasks {
		row := Row{Intent: task.Intent}
		var sample core.Event
		if task.SampleType != "" {
			i := slices.IndexFunc(raws, func(e core.Event) bool { return e.Type == task.SampleType })
			if i < 0 {
				return rep, fmt.Errorf("task %q: no %s event in the intake", task.Intent, task.SampleType)
			}
			sample = raws[i]
		}
		ask, err := shell.DraftAsk(ctx, st, k, task.Intent, sample, task.Hint)
		if err != nil {
			return rep, err
		}
		if task.Mode == "bundle" {
			if err := runBundle(ctx, st, x, a, model, ask, k, &row); err != nil {
				return rep, err
			}
			rep.Rows = append(rep.Rows, row)
			continue
		}
		d, err := a.DraftRule(ctx, ask)
		if err != nil {
			row.Error = err.Error()
			rep.Rows = append(rep.Rows, row)
			continue
		}
		row.RuleID = d.RuleID
		row.Draft, _ = json.MarshalIndent(d, "       ", "  ")
		if _, err := st.InsertRuleVersion(ctx, core.Rule{
			ID: d.RuleID, Status: core.StatusDraft, Priority: d.Priority,
			EffectiveFrom: sample.OccurredAt, CreatedBy: "agent:" + model,
			Description: d.Description, Spec: d.Spec,
		}); err != nil {
			row.Error = "store draft: " + err.Error()
			rep.Rows = append(rep.Rows, row)
			continue
		}
		diff, err := x.SimulateRule(ctx, d.RuleID)
		if err != nil {
			return rep, err
		}
		row.SimErrors = diff.Errors
		row.Explains = len(diff.UnexplainedBefore) - len(diff.UnexplainedAfter)
		row.Added = len(diff.Added)
		if len(diff.Errors) == 0 {
			_, booked, procErrs, err := x.ApproveRule(ctx, d.RuleID, "eval", sample.OccurredAt)
			if err != nil {
				row.ApproveErrs = append(row.ApproveErrs, err.Error())
			} else {
				row.Approved, row.Booked = true, booked
			}
			for _, e := range procErrs {
				row.ApproveErrs = append(row.ApproveErrs, e.Error())
			}
		}
		rep.Rows = append(rep.Rows, row)
	}
	left, err := st.UnmatchedRawEvents(ctx)
	if err != nil {
		return rep, err
	}
	for _, e := range left {
		if e.ActivityName == "submit_event" { // the intake; door events are the run's own
			rep.Residue = append(rep.Residue, fmt.Sprintf("%s #%d", e.Type, e.ID))
		}
	}
	if rep.Objects, err = st.CountObjectsByType(ctx); err != nil {
		return rep, err
	}
	for _, w := range c.ExpectWhere {
		objs, err := st.ObjectsByType(ctx, w.Type)
		if err != nil {
			return rep, err
		}
		n := 0
		for _, o := range objs {
			if fmt.Sprint(o.State[w.Field]) == w.Value {
				n++
			}
		}
		if n != w.Count {
			rep.Mismatch = append(rep.Mismatch, fmt.Sprintf("%s with %s=%s: %d of %d", w.Type, w.Field, w.Value, n, w.Count))
		}
	}
	types := make([]string, 0, len(c.Expect))
	for t := range c.Expect {
		types = append(types, t)
	}
	slices.Sort(types)
	for _, t := range types {
		if got := rep.Objects[t]; got != c.Expect[t] {
			rep.Mismatch = append(rep.Mismatch, fmt.Sprintf("%s: %d of %d", t, got, c.Expect[t]))
		}
	}
	return rep, nil
}

// runBundle plays one bundle task: draft, land as drafts, dry-run the whole
// bundle, approve it whole when the dry run is clean.
func runBundle(ctx context.Context, st *store.Store, x *exec.Executor, a *agent.Agent, model string, ask agent.Ask, k *network.Knowledge, row *Row) error {
	bd, err := a.DraftBundle(ctx, ask)
	if errors.Is(err, agent.ErrAlreadyAnswered) {
		// Nothing to approve; whether that was right, the state check says.
		row.RuleID, row.Approved = "already answered: "+bd.Description, true
		return nil
	}
	if err != nil {
		row.Error = err.Error()
		return nil
	}
	date := ask.Sample.OccurredAt
	if date == "" {
		date = "2026-01-01"
	}
	b, err := shell.StoreBundleDraft(ctx, st, k, bd, "agent:"+model, date)
	if err != nil {
		row.Error = "store bundle: " + err.Error()
		return nil
	}
	row.RuleID = fmt.Sprintf("%s [%s]", b.ID, bd.Summary())
	row.Draft, _ = json.MarshalIndent(bd, "       ", "  ")
	diff, err := x.SimulateBundle(ctx, b)
	if err != nil {
		return err
	}
	row.SimErrors = diff.Errors
	row.Explains = len(diff.UnexplainedBefore) - len(diff.UnexplainedAfter)
	row.Added = len(diff.Added)
	if len(diff.Errors) > 0 {
		return nil
	}
	_, booked, procErrs, err := x.ApproveBundle(ctx, b.ID, "eval", date)
	if err != nil {
		row.ApproveErrs = append(row.ApproveErrs, err.Error())
		return nil
	}
	row.Approved, row.Booked = true, booked
	for _, e := range procErrs {
		row.ApproveErrs = append(row.ApproveErrs, e.Error())
	}
	// The human's second, deliberate act: when the approved bundle would
	// explain the past further, the dry run is read and the backfill approved.
	plan, err := x.PlanBackfill(ctx, date)
	if err != nil {
		return err
	}
	if len(plan.Chains) > 0 && len(plan.Errors) == 0 {
		n, errs, err := x.ApproveBackfill(ctx, "eval", date)
		if err != nil {
			row.ApproveErrs = append(row.ApproveErrs, "backfill: "+err.Error())
		}
		row.Backfilled = n
		for _, e := range errs {
			row.ApproveErrs = append(row.ApproveErrs, "backfill: "+e.Error())
		}
	}
	return nil
}

func loadSeed(ctx context.Context, st *store.Store, path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var defs struct {
		ObjectTypes []core.ObjectType `json:"object_types"`
		ViewDefs    []core.ViewDef    `json:"view_defs"`
	}
	if err := json.Unmarshal(b, &defs); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	for _, t := range defs.ObjectTypes {
		if err := st.InsertObjectType(ctx, t); err != nil {
			return fmt.Errorf("%s: type %s: %w", path, t.Name, err)
		}
	}
	for _, v := range defs.ViewDefs {
		if err := st.InsertViewDef(ctx, v); err != nil {
			return fmt.Errorf("%s: view %s: %w", path, v.ID, err)
		}
	}
	return nil
}

func readEvent(raw json.RawMessage, dir string) (rawEvent, error) {
	var file string
	if json.Unmarshal(raw, &file) == nil {
		b, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			return rawEvent{}, err
		}
		raw = b
	}
	var ev rawEvent
	if err := json.Unmarshal(raw, &ev); err != nil {
		return rawEvent{}, err
	}
	if ev.EventType == "" || ev.OccurredAt == "" {
		return rawEvent{}, fmt.Errorf("event needs event_type and occurred_at")
	}
	return ev, nil
}

// String renders the report for a terminal: one line per task, then the
// verdict — residue first, it is the number.
func (r Report) String() string { return r.render(false) }

// Verbose renders the report with every draft the agent proposed.
func (r Report) Verbose() string { return r.render(true) }

func (r Report) render(verbose bool) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "corpus %s, model %s: %d intake events, %d tasks\n\n", r.Corpus, r.Model, r.Intake, len(r.Rows))
	for i, row := range r.Rows {
		mark := "✗"
		if row.Approved && len(row.ApproveErrs) == 0 {
			mark = "✓"
		}
		fmt.Fprintf(&sb, "%s %2d. %s\n", mark, i+1, row.Intent)
		switch {
		case row.Error != "":
			fmt.Fprintf(&sb, "       refused: %s\n", row.Error)
		default:
			fmt.Fprintf(&sb, "       %s: explains %d, adds %d", row.RuleID, row.Explains, row.Added)
			if row.Approved {
				fmt.Fprintf(&sb, ", approved, booked %d", row.Booked)
			}
			if row.Backfilled > 0 {
				fmt.Fprintf(&sb, ", backfilled %d", row.Backfilled)
			}
			sb.WriteByte('\n')
			for _, e := range row.SimErrors {
				fmt.Fprintf(&sb, "       dry run: %s\n", e)
			}
			for _, e := range row.ApproveErrs {
				fmt.Fprintf(&sb, "       worklist: %s\n", e)
			}
			if verbose && len(row.Draft) > 0 {
				fmt.Fprintf(&sb, "       %s\n", row.Draft)
			}
		}
	}
	fmt.Fprintf(&sb, "\nresidue: %d of %d intake events unexplained", len(r.Residue), r.Intake)
	if len(r.Residue) > 0 {
		fmt.Fprintf(&sb, " (%s)", strings.Join(r.Residue, ", "))
	}
	sb.WriteByte('\n')
	if len(r.Mismatch) > 0 {
		fmt.Fprintf(&sb, "state: %s\n", strings.Join(r.Mismatch, "; "))
	} else {
		sb.WriteString("state: every expected object count matches\n")
	}
	if r.Done() {
		sb.WriteString("DONE: the corpus is explained\n")
	} else {
		sb.WriteString("NOT DONE\n")
	}
	return sb.String()
}
