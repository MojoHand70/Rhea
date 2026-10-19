package pl

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"rhea/internal/adapter"
	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/exam"
	"rhea/internal/pack"
	"rhea/internal/statute"
)

// The Commission's sample papers (Komisja Egzaminacyjna at the Polska Izba
// Biegłych Rewidentów, pibr.org.pl/pl/egzaminy-pytania): one PDF per
// subject per year, the single-choice questions with the right option set
// in bold, the situational tasks with the Commission's model solution
// ("rozwiązanie wzorcowe"). The actual exam items are not published (the
// item bank is reused); these are the Commission's own illustrations, with
// its own key — the public yardstick for the Poland pack (DIRECTION
// 2026-10-19, "The auditor's exam").

// PIBRScoring is the marking rule of the qualification regulation
// (rozporządzenie Ministra Finansów w sprawie postępowania kwalifikacyjnego
// na biegłych rewidentów): a right test answer +2, a wrong one −1, a blank
// 0; situational tasks carry no negative points; a block passes at 60% of
// its maximum, each block on its own.
var PIBRScoring = exam.Scoring{Correct: 2, Wrong: -1, Blank: 0, PassShare: 0.6}

// A line of the paper as pdftohtml -xml lays it out: every text element
// on one baseline joined, bold if any of them was — the Commission marks
// the right option in bold.
type paperLine struct {
	page, top, left int
	bold            bool
	text            string
}

var (
	reElement  = regexp.MustCompile(`(?s)<page number="(\d+)"|<text top="(\d+)" left="(\d+)"[^>]*>(.*?)</text>`)
	reTag      = regexp.MustCompile(`<[^>]+>`)
	reSpaces   = regexp.MustCompile(`\s+`)
	reQuestion = regexp.MustCompile(`^(\d{1,3})\.(?:\s+(.*))?$`)
	reOption   = regexp.MustCompile(`^([a-c])\)(?:\s+(.*))?$`)
	reTask     = regexp.MustCompile(`^Zadanie\s+(\d+)\b`)
	reBlock    = regexp.MustCompile(`^BLOK\s+TEMATYCZNY\s+([IVX]+)`)
	reYear     = regexp.MustCompile(`\b(20\d\d)\b`)
	reFooter   = regexp.MustCompile(`(?i)^\W*(str\.?\s*\d*|strona\s*\d*|\d{1,3}|Komisja Egzaminacyjna)\W*$`) // ". str. 3", "Strona 4", "strona" + "13" apart
)

// paperLines groups the XML's text elements into baselines.
func paperLines(xml string) []paperLine {
	var lines []paperLine
	page := 0
	for _, m := range reElement.FindAllStringSubmatch(xml, -1) {
		if m[1] != "" {
			page, _ = strconv.Atoi(m[1])
			continue
		}
		top, _ := strconv.Atoi(m[2])
		left, _ := strconv.Atoi(m[3])
		bold := strings.Contains(m[4], "<b>")
		text := strings.TrimSpace(html.UnescapeString(reTag.ReplaceAllString(m[4], "")))
		if text == "" {
			continue
		}
		if n := len(lines); n > 0 && lines[n-1].page == page && abs(lines[n-1].top-top) <= 5 { // a footer's word and number sit 4 apart
			lines[n-1].text += "  " + text
			lines[n-1].bold = lines[n-1].bold || bold
			continue
		}
		lines = append(lines, paperLine{page: page, top: top, left: left, bold: bold, text: text})
	}
	for i := range lines {
		lines[i].text = reSpaces.ReplaceAllString(unspace(lines[i].text), " ")
	}
	return lines
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// unspace undoes letter-spacing: the paper's headings come out of the PDF
// as "B L O K   T E M A T Y C Z N Y   I" — single letters a space apart,
// words three apart. A line that is mostly single characters is collapsed.
func unspace(s string) string {
	fields := strings.Fields(s)
	if len(fields) < 4 {
		return s
	}
	singles := 0
	for _, f := range fields {
		if utf8.RuneCountInString(f) == 1 {
			singles++
		}
	}
	if singles*10 < len(fields)*8 {
		return s
	}
	var sb strings.Builder
	spaces := 0
	for _, r := range s {
		if r == ' ' {
			spaces++
			continue
		}
		if spaces >= 2 && sb.Len() > 0 {
			sb.WriteByte(' ')
		}
		spaces = 0
		sb.WriteRune(r)
	}
	return sb.String()
}

func squash(s string) string { return strings.ToUpper(strings.Join(strings.Fields(s), "")) }

// ParsePIBRPaper reads one sample paper from pdftohtml's XML
// (pdftohtml -xml -i -stdout FILE.pdf): the subject and year from the page
// header, the blocks, each block's test questions with the option in bold
// as the key, and the situational tasks with their model solutions.
func ParsePIBRPaper(xml []byte, source string) (exam.Paper, error) {
	p := exam.Paper{Source: source}
	const (
		modeNone = iota
		modeQuestions
		modeTask
		modeSolution
	)
	mode := modeNone
	block := ""
	var q *exam.Question
	var opt *exam.Option
	var optBold []bool // per option of q: bold anywhere on its lines
	var task *exam.Task
	closeQuestion := func() {
		if q == nil {
			return
		}
		var marked []string
		for i, b := range optBold {
			if b {
				marked = append(marked, q.Options[i].Letter)
			}
		}
		switch len(marked) {
		case 1:
			q.Key = marked[0]
		case 0:
			q.KeyNote = "no option marked in bold"
		default: // the paper's marking is ambiguous; a person decides
			q.KeyNote = "options marked in bold: " + strings.Join(marked, ", ")
		}
		p.Questions = append(p.Questions, *q)
		q, opt, optBold = nil, nil, nil
	}
	closeTask := func() {
		if task != nil {
			task.Text = strings.TrimSpace(task.Text)
			task.Solution = strings.TrimSpace(task.Solution)
			p.Tasks = append(p.Tasks, *task)
			task = nil
		}
	}
	nextN := 1
	headerTop := -1 // the running header's baseline: the first line of page 1, repeated on every page
	for _, ln := range paperLines(string(xml)) {
		text := ln.text
		if headerTop < 0 {
			headerTop = ln.top
		}
		if ln.top <= headerTop+8 { // the page header: subject and year
			if ln.page == 1 && p.Subject == "" {
				if m := reYear.FindString(text); m != "" {
					p.Year = m
					text = strings.Replace(text, m, "", 1)
				}
				text = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(text), "Rok"))
				p.Subject = text
			}
			continue
		}
		if reFooter.MatchString(text) {
			continue
		}
		sq := squash(text)
		switch {
		case strings.HasPrefix(sq, "ROZWIĄZANIAWZORCOWE"):
			continue
		case reBlock.MatchString(strings.ToUpper(text)):
			closeQuestion()
			closeTask()
			block = reBlock.FindStringSubmatch(strings.ToUpper(text))[1]
			mode, nextN = modeQuestions, 1
			continue
		case sq == "PYTANIATESTOWE":
			closeQuestion()
			closeTask()
			mode, nextN = modeQuestions, 1
			continue
		case sq == "ZADANIASYTUACYJNE":
			closeQuestion()
			closeTask()
			mode = modeNone
			continue
		case reTask.MatchString(text):
			closeQuestion()
			closeTask()
			n, _ := strconv.Atoi(reTask.FindStringSubmatch(text)[1])
			task = &exam.Task{Block: block, N: n}
			mode = modeTask
			continue
		case (strings.HasPrefix(sq, "ROZWIĄZANIEWZORCOWE") || sq == "ROZWIĄZANIE" || sq == "ROZWIĄZANIE:") && task != nil:
			// the model solution's heading; one 2025 task heads it plain "Rozwiązanie"
			mode = modeSolution
			continue
		}
		switch mode {
		case modeQuestions:
			if block == "" {
				block = "I"
			}
			if m := reQuestion.FindStringSubmatch(text); m != nil {
				if n, _ := strconv.Atoi(m[1]); n == nextN {
					closeQuestion()
					q = &exam.Question{Block: block, N: n, Stem: strings.TrimSpace(m[2])}
					nextN++
					continue
				}
			}
			if q == nil {
				continue
			}
			if m := reOption.FindStringSubmatch(text); m != nil {
				q.Options = append(q.Options, exam.Option{Letter: m[1], Text: strings.TrimSpace(m[2])})
				optBold = append(optBold, ln.bold)
				opt = &q.Options[len(q.Options)-1]
				continue
			}
			if opt != nil {
				opt.Text = strings.TrimSpace(opt.Text + " " + text)
				optBold[len(optBold)-1] = optBold[len(optBold)-1] || ln.bold
			} else {
				q.Stem = strings.TrimSpace(q.Stem + " " + text)
			}
		case modeTask:
			task.Text += text + "\n"
		case modeSolution:
			task.Solution += text + "\n"
		}
	}
	closeQuestion()
	closeTask()
	if len(p.Questions) == 0 {
		return p, fmt.Errorf("no test questions found in %s — is it pdftohtml -xml output of a Commission paper?", source)
	}
	return p, nil
}

// paperXML returns pdftohtml's XML for a paper: the file itself when it is
// XML, or pdftohtml run on the PDF (poppler-utils).
func paperXML(path string) ([]byte, error) {
	if strings.EqualFold(filepath.Ext(path), ".pdf") {
		out, err := exec.Command("pdftohtml", "-xml", "-i", "-stdout", path).Output()
		if err != nil {
			return nil, fmt.Errorf("pdftohtml -xml %s: %w (poppler-utils needed)", path, err)
		}
		return out, nil
	}
	return os.ReadFile(path)
}

// rulebook reads pack files as the rules and object types they ship —
// the pack as published, no database: what the host is taught.
func rulebook(paths []string) ([]core.Rule, []core.ObjectType, error) {
	var rules []core.Rule
	var types []core.ObjectType
	for _, path := range paths {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, err
		}
		var m pack.Manifest
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, nil, fmt.Errorf("%s: %w", path, err)
		}
		for _, r := range m.Rules {
			rules = append(rules, core.Rule{ID: r.RuleID, Priority: r.Priority, EffectiveFrom: r.EffectiveFrom, Description: r.Description, Spec: r.Spec, Warrant: r.Warrant})
		}
		types = append(types, m.ObjectTypes...)
	}
	return rules, types, nil
}

// The exam tool (rhea run pibr-exam …): sits one of the Commission's sample
// papers with the bare model and with the host, prints the report and keeps
// every answer in a JSON file. No log, no store.
func init() {
	adapter.ProvideCommand(adapter.Command{
		Name:    "pibr-exam",
		Summary: "sit one of the Commission's sample papers: the bare model and the host, marked as the regulation marks (rhea run pibr-exam for help)",
		Run:     runPIBRExam,
	})
}

const pibrExamHelp = `rhea run pibr-exam -paper FILE [-packs a.json,b.json] [-who bare,host] [-parallel N] [-only N] [-out FILE] [-dry]
rhea run pibr-exam -paper FILE -task I-1 [-cases DIR] [-seeds DIR] [-recompile] [-dry]
rhea run pibr-exam -paper FILE -learn SITTING.json -block I [-limit N]

  -paper    a Commission sample paper: the PDF (pdftohtml runs on it) or its pdftohtml -xml output
  -task     sit one situational task through the kernel instead of the test questions: the
            model reads the task and its model solution into a case file under -cases
            (testdata/exam/<paper>/<task>.json — review it; it is kept and reused unless
            -recompile), then the case runs as an implementation in miniature on a throwaway
            log: the finance base from -seeds (testdata/seed), the chart and the facts as
            intake, the author drafting a bundle per intent, the approver reading each dry
            run, and the books read against the solution's entries
  -packs    the pack files the host is taught (default: the Poland pack and the common packs)
  -who      who sits it: bare (the model alone), host (with the rulebook open), or both
  -statute  the acts and standards the host may consult (rhea run pl-acts fetch, rhea run
            pl-standards fetch; default data/acts/pl,data/standards/pl, missing folders skipped;
            "" for none): before answering, the host names the provisions it would open, is handed
            their text, and may answer "statute" only citing one of them
  -learn    the learning-rate test: the answers of a prior sitting (its -out file) that did not
            rest on a rule become lessons for the author, which drafts them as definitions on a
            throwaway database; the block is sat again with that rulebook; before and after are
            reported against the clock. -block names the block (I), -limit caps the lessons.
  -parallel calls in flight at once (default 6)
  -only     sit only the first N questions (a smoke run)
  -out      write the report with every answer as JSON here
  -dry      parse the paper and print what was found; call no model

The test questions only, marked +2 / -1 / 0 with 60% of each block to pass; the situational
tasks are parsed and counted, not yet sat. Needs ANTHROPIC_API_KEY (RHEA_MODEL optional).`

var defaultPacks = []string{
	"packs/pl/pack.json", "packs/pl/warehouse.json", "packs/pl/nbp.json",
	"packs/common/documents.json", "packs/common/people.json", "packs/common/periods.json",
}

func runPIBRExam(ctx context.Context, args []string, out io.Writer) error {
	if len(args) == 0 {
		fmt.Fprintln(out, pibrExamHelp)
		return nil
	}
	fs := flag.NewFlagSet("pibr-exam", flag.ContinueOnError)
	fs.SetOutput(out)
	paperPath := fs.String("paper", "", "the paper: PDF or pdftohtml -xml output")
	packs := fs.String("packs", strings.Join(defaultPacks, ","), "pack files the host is taught")
	who := fs.String("who", "bare,host", "bare, host, or both")
	parallel := fs.Int("parallel", 6, "calls in flight")
	only := fs.Int("only", 0, "sit only the first N questions")
	outPath := fs.String("out", "", "write the JSON report here")
	dry := fs.Bool("dry", false, "parse only")
	taskRef := fs.String("task", "", "sit this situational task through the kernel, e.g. I-1")
	casesDir := fs.String("cases", "testdata/exam", "where case files live")
	seedsDir := fs.String("seeds", "testdata/seed", "the finance base a case runs on")
	recompile := fs.Bool("recompile", false, "read the task again even if its case file exists")
	statuteDir := fs.String("statute", "data/acts/pl,data/standards/pl", "the acts and standards the host may consult; \"\" for none")
	learnFrom := fs.String("learn", "", "a prior sitting's -out file whose non-rule answers become the author's lessons")
	block := fs.String("block", "I", "the block the learning-rate test runs on")
	limit := fs.Int("limit", 0, "at most this many lessons (0: all)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *paperPath == "" {
		return errors.New("pibr-exam needs -paper")
	}
	xml, err := paperXML(*paperPath)
	if err != nil {
		return err
	}
	paper, err := ParsePIBRPaper(xml, filepath.Base(*paperPath))
	if err != nil {
		return err
	}
	if *taskRef != "" {
		return sitTask(ctx, out, paper, *taskRef, *casesDir, *seedsDir, *recompile, *dry)
	}
	if *learnFrom != "" {
		return learnTest(ctx, out, paper, *learnFrom, *block, *limit, strings.Split(*packs, ","), *seedsDir, *statuteDir, *parallel)
	}
	keyed, blocks := 0, map[string]int{}
	for _, q := range paper.Questions {
		blocks[q.Block]++
		if q.Key != "" {
			keyed++
		}
	}
	fmt.Fprintf(out, "%s, %s: %d test questions in %d block(s), %d keyed, %d unkeyed; %d situational task(s)\n",
		paper.Subject, paper.Year, len(paper.Questions), len(blocks), keyed, len(paper.Questions)-keyed, len(paper.Tasks))
	if *dry {
		for _, q := range paper.Questions {
			fmt.Fprintf(out, "  %s-%d key=%s options=%d: %.80s", q.Block, q.N, q.Key, len(q.Options), q.Stem)
			if q.KeyNote != "" {
				fmt.Fprintf(out, "  [%s]", q.KeyNote)
			}
			fmt.Fprintln(out)
		}
		for _, t := range paper.Tasks {
			fmt.Fprintf(out, "  task %s-%d: %d chars, solution %d chars\n", t.Block, t.N, len(t.Text), len(t.Solution))
		}
		return nil
	}
	if *only > 0 && *only < len(paper.Questions) {
		paper.Questions = paper.Questions[:*only]
	}
	a := agent.New()
	var sitters []exam.Answerer
	for _, w := range strings.Split(*who, ",") {
		switch strings.TrimSpace(w) {
		case "bare":
			sitters = append(sitters, exam.Bare{Agent: a})
		case "host":
			rules, types, err := rulebook(strings.Split(*packs, ","))
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "the host is taught %d rules and %d object types\n", len(rules), len(types))
			host := exam.NewHost(a, rules, types)
			if *statuteDir != "" {
				if lib, err := statute.Load(*statuteDir); err == nil && len(lib.Acts()) > 0 {
					host.Library = lib
					var names []string
					for _, act := range lib.Acts() {
						names = append(names, act.Key)
					}
					fmt.Fprintf(out, "the host may consult %d acts: %s\n", len(names), strings.Join(names, ", "))
				} else if err != nil && !errors.Is(err, os.ErrNotExist) {
					return err
				} else {
					fmt.Fprintln(out, "no acts to consult (rhea run pl-acts fetch)")
				}
			}
			sitters = append(sitters, host)
		default:
			return fmt.Errorf("-who: %q is neither bare nor host", w)
		}
	}
	rep := exam.Report{Subject: paper.Subject, Year: paper.Year, Source: paper.Source, Model: agent.Model(), Scoring: PIBRScoring, Tasks: len(paper.Tasks)}
	for _, s := range sitters {
		fmt.Fprintf(out, "%s sits %d questions…\n", s.Name(), len(paper.Questions))
		rep.Results = append(rep.Results, exam.Sit(ctx, paper, PIBRScoring, s, *parallel))
	}
	fmt.Fprint(out, "\n"+rep.String())
	if *outPath != "" {
		b, err := json.MarshalIndent(rep, "", " ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(*outPath, b, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(out, "every answer kept in %s\n", *outPath)
	}
	return nil
}

// sitTask reads one situational task into a case (or takes the case file a
// person reviewed) and sits it through the kernel.
func sitTask(ctx context.Context, out io.Writer, paper exam.Paper, ref, casesDir, seedsDir string, recompile, dry bool) error {
	var task *exam.Task
	for i := range paper.Tasks {
		if fmt.Sprintf("%s-%d", paper.Tasks[i].Block, paper.Tasks[i].N) == strings.ToUpper(ref) {
			task = &paper.Tasks[i]
		}
	}
	if task == nil {
		return fmt.Errorf("no task %s on this paper (%d tasks)", ref, len(paper.Tasks))
	}
	casePath := filepath.Join(casesDir, exam.Slug(paper.Subject+" "+paper.Year), strings.ToLower(ref)+".json")
	c, err := exam.ReadCase(casePath)
	switch {
	case err == nil && !recompile:
		fmt.Fprintf(out, "case %s as reviewed\n", casePath)
	case err == nil || errors.Is(err, os.ErrNotExist):
		fmt.Fprintf(out, "the model reads task %s (%d chars, solution %d chars)…\n", ref, len(task.Text), len(task.Solution))
		c, err = exam.CompileCase(ctx, agent.New(), paper, *task)
		if err != nil {
			return err
		}
		if err := exam.WriteCase(casePath, c); err != nil {
			return err
		}
		fmt.Fprintf(out, "case written to %s — review it before trusting the result\n", casePath)
	default:
		return err
	}
	fmt.Fprintf(out, "%s, task %s: %s\n  %d accounts, %d events, %d intents, %d expected postings, %d answers for the host\n",
		c.Paper, c.Task, c.Company, len(c.Accounts), len(c.Events), len(c.Intents), len(c.ExpectPostings), len(c.Answers))
	if c.ForTheHost() {
		fmt.Fprintln(out, "a task without entries — the host's to answer in words, not the kernel's to sit; its answers are kept in the case")
		return nil
	}
	for _, in := range c.Intents {
		fmt.Fprintf(out, "  intent: %s\n", in)
	}
	for _, p := range c.ExpectPostings {
		fmt.Fprintf(out, "  expects: %s %s %s\n", p.Account, p.Side, p.Amount)
	}
	if c.Notes != "" {
		fmt.Fprintf(out, "  notes: %s\n", c.Notes)
	}
	if dry {
		return nil
	}
	rep, err := exam.SitCase(ctx, agent.New(), agent.Model(), casePath, seedsDir)
	if err != nil {
		return err
	}
	fmt.Fprint(out, "\n"+rep.Verbose()) // every draft and dry run: what the author did with the case
	if rep.Done() {
		fmt.Fprintln(out, "DONE: the books hold every entry the solution expects, nothing left unexplained")
	} else {
		fmt.Fprintln(out, "NOT DONE")
	}
	return nil
}

// learnTest runs the learning-rate test on one block of a paper.
func learnTest(ctx context.Context, out io.Writer, paper exam.Paper, sittingPath, block string, limit int, packs []string, seedsDir, statuteDir string, parallel int) error {
	b, err := os.ReadFile(sittingPath)
	if err != nil {
		return err
	}
	var prior exam.Report
	if err := json.Unmarshal(b, &prior); err != nil {
		return fmt.Errorf("%s: %w", sittingPath, err)
	}
	var host *exam.Result
	for i := range prior.Results {
		if prior.Results[i].Answerer == "host" {
			host = &prior.Results[i]
		}
	}
	if host == nil {
		return fmt.Errorf("%s holds no host sitting", sittingPath)
	}
	opt := exam.LearnOptions{Packs: packs, SeedsDir: seedsDir, Parallel: parallel, Limit: limit, Verbose: out}
	if statuteDir != "" {
		if lib, err := statute.Load(statuteDir); err == nil && len(lib.Acts()) > 0 {
			opt.Library = lib
		}
	}
	res, err := exam.Learn(ctx, agent.New(), agent.Model(), paper.Subject+" "+paper.Year, paper, *host, block, PIBRScoring, opt, out)
	if err != nil {
		return err
	}
	fmt.Fprint(out, "\n"+res.String())
	return nil
}
