package shell

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/project"
)

// answerRows is how many rows an answer shows and records; more are
// counted, not read (a listing that long is a report, kept and paged).
const answerRows = 500

// collection is the read side as the author sees it: the tables the catalog
// makes, each link pointing at the table of the type it names.
func (s *Server) collection(ctx context.Context) ([]agent.CollectionTable, error) {
	types, err := s.Store.ListObjectTypes(ctx)
	if err != nil {
		return nil, err
	}
	tables := project.Collection(types)
	where := map[string]string{}
	for _, t := range tables {
		where[t.Type] = t.Schema + "." + t.Name
	}
	out := make([]agent.CollectionTable, 0, len(tables))
	for _, t := range tables {
		ct := agent.CollectionTable{Schema: t.Schema, Name: t.Name, Type: t.Type}
		for _, c := range t.Columns {
			col := agent.CollectionColumn{Name: c.Name, SQL: c.SQL, Kind: c.Kind}
			if target, ok := core.RefTarget(c.Kind); ok {
				col.Links = where[target]
			}
			ct.Columns = append(ct.Columns, col)
		}
		out = append(out, ct)
	}
	return out, nil
}

// handleCollection lists the collection: the business's own databases of
// objects, as a query reads them.
func (s *Server) handleCollection(w http.ResponseWriter, r *http.Request) {
	if !s.allowSystem(w, r, ViewAnswer) {
		return
	}
	c, err := s.collection(r.Context())
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, c)
}

// words is the names of what the kernel holds, as the company's packs give
// them: their words screens merged; nil without one.
func (s *Server) words(ctx context.Context) *core.WordsSpec {
	w, err := s.mergedWords(ctx) // every pack's the company has, merged
	if err != nil {
		return nil
	}
	return w
}

// books are the books the active rules post to (literal books only).
func (s *Server) books(ctx context.Context) []string {
	rules, err := s.Store.ActiveRules(ctx)
	if err != nil {
		return nil
	}
	var out []string
	for _, r := range rules {
		if p := r.Spec.Effect.Postings; p != nil && p.Book != "" && !strings.HasPrefix(p.Book, "=") && !slices.Contains(out, p.Book) {
			out = append(out, p.Book)
		}
	}
	return out
}

type answerColumn struct {
	Name string `json:"name"`
	Type string `json:"type"` // DuckDB's: DECIMAL(18,2), DATE, VARCHAR…
}

// handleAnswer answers a question from the books now (DIRECTION 2026-10-17,
// "Reporting"; KK: answered at once, without approval, when it only reads).
// The author writes one SELECT over the collection and says what it counts;
// the kernel runs it in the locked session, gives DuckDB's refusal back to
// the author once, and records what she told the person through the
// host_answered door. Nothing in the books changes.
func (s *Server) handleAnswer(w http.ResponseWriter, r *http.Request) {
	if !s.allowSystem(w, r, ViewAnswer) {
		return
	}
	ctx := r.Context()
	if s.Agent == nil {
		writeErr(w, 503, fmt.Errorf("answer: no author is connected"))
		return
	}
	var req struct {
		Question string `json:"question"`
		AskedBy  string `json:"asked_by"`
		// Missing is what the face's understanding said she was never taught,
		// when she answers only the nearest thing: the moment is recorded too.
		Missing string `json:"missing"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, 400, err)
		return
	}
	req.Question = strings.TrimSpace(req.Question)
	if req.Question == "" || len(req.Question) > 2000 || req.AskedBy == "" {
		writeErr(w, 400, fmt.Errorf("answer: question (1–2000 bytes) and asked_by are required"))
		return
	}
	tables, err := s.collection(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	reports, kept, err := s.keptReports(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	ask := agent.AnswerAsk{Question: req.Question, Today: s.Clock.Today(), Tables: tables, Words: s.words(ctx), Books: s.books(ctx), Reports: reports}
	var (
		ans agent.Answer
		got componentOut
	)
	// Every moment she could not fully answer is recorded and, when it may
	// be, reaches the house (KK 2026-10-18: a failure belongs in the
	// Learning Centre too). The general words are the face's, or the
	// author's lesson, never the person's question.
	note := func(kind, detail string) *noted {
		missing := cmp.Or(strings.TrimSpace(req.Missing), strings.TrimSpace(ans.Lesson), strings.TrimSpace(ans.Cannot))
		if missing == "" {
			return nil
		}
		n, err := s.noteUntaught(ctx, req.Question, missing, req.AskedBy, kind, detail)
		if err != nil {
			log.Printf("answer: noting %s: %v", kind, err)
			return nil
		}
		return &n
	}
	for try := 0; try < answerTries; try++ {
		if ans, err = s.Agent.Answer(ctx, ask); err != nil {
			writeErr(w, 502, err)
			return
		}
		if ans.Cannot != "" {
			writeJSON(w, 200, map[string]any{"question": req.Question, "cannot": ans.Cannot, "noted": note("cannot", "")})
			return
		}
		if ans.Report != "" {
			s.answerWithReport(ctx, w, req.Question, req.AskedBy, ans, kept[ans.Report])
			return
		}
		if got, err = s.runQuery(ctx, ans.SQL, ans.Visual); err == nil {
			break
		}
		log.Printf("answer: %q: %v", req.Question, err)
		ask.Tried = &agent.AnswerTry{SQL: ans.SQL, Error: err.Error()}
	}
	if err != nil {
		writeJSON(w, 200, map[string]any{"question": req.Question, "failed": true, "error": err.Error(), "noted": note("failed", refusalClass(err))})
		return
	}
	got.Title, got.Reading = ans.Visual.Title, ans.Reading

	vis, _ := json.Marshal(got.Visual)
	told := got.Rows
	if len(told) > 50 {
		told = told[:50]
	}
	names := make([]string, len(got.Columns))
	for i, c := range got.Columns {
		names[i] = c.Name
	}
	result, _ := json.Marshal(map[string]any{"columns": names, "rows": told, "more": got.More || len(got.Rows) > len(told)})
	t, err := s.Exec.TriggerActivity(ctx, "host_answered", map[string]any{"question": req.Question, "sql": ans.SQL, "reading": ans.Reading,
		"visual": string(vis), "result": string(result), "rows": len(got.Rows), "asked_by": req.AskedBy, "occurred_at": s.Clock.Today()}, req.AskedBy, "")
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	var nearest *noted
	if strings.TrimSpace(req.Missing) != "" {
		nearest = note("untaught", "")
	}
	writeJSON(w, 200, map[string]any{"event_id": t.EventID, "question": req.Question, "sql": ans.SQL, "reading": ans.Reading,
		"visual": got.Visual, "columns": got.Columns, "rows": got.Rows, "row_objects": got.Of, "more": got.More, "noted": nearest})
}

// answerTries: the author's query and two fixes after DuckDB's refusal.
const answerTries = 3

// keptReports are the company's kept reports as the author is told of them
// (KK 2026-10-30: a report kept is asked for again, never written again),
// and each by its id.
func (s *Server) keptReports(ctx context.Context) ([]agent.AnswerReport, map[string]core.ViewDef, error) {
	vds, err := s.effectiveViewDefs(ctx)
	if err != nil {
		return nil, nil, err
	}
	var out []agent.AnswerReport
	byID := map[string]core.ViewDef{}
	for _, v := range vds {
		if v.Notion != "report" || !s.mayView(ctx, &v) {
			continue
		}
		var spec core.ReportSpec
		if err := json.Unmarshal(v.Spec, &spec); err != nil {
			continue
		}
		r := agent.AnswerReport{ID: v.ID, Title: v.Title, Period: spec.Period}
		for _, c := range spec.Components {
			r.Parts = append(r.Parts, cmp.Or(c.Reading, c.Title))
		}
		out = append(out, r)
		byID[v.ID] = v
	}
	return out, byID, nil
}

// answerWithReport answers a question with a kept report, computed for the
// period asked (KK 2026-10-30: the RZiS for August is the RZiS kept, the
// period its parameter): no query is written, nothing waits to be kept
// again. What she told is recorded as any answer is, naming the report.
func (s *Server) answerWithReport(ctx context.Context, w http.ResponseWriter, question, askedBy string, ans agent.Answer, vd core.ViewDef) {
	q := url.Values{}
	if ans.From != "" || ans.To != "" {
		q.Set("from", ans.From)
		q.Set("to", ans.To)
	}
	comps, period, code, err := s.reportComponents(ctx, &vd, q)
	if err != nil {
		if code == 400 {
			writeJSON(w, 200, map[string]any{"question": question, "failed": true, "error": err.Error()})
			return
		}
		writeErr(w, code, err)
		return
	}
	first := comps[0]
	reading := strings.TrimSpace(ans.Reading)
	if reading == "" {
		var parts []string
		for _, c := range comps {
			parts = append(parts, c.Reading)
		}
		reading = strings.Join(parts, " ")
	}
	vis, _ := json.Marshal(first.Visual)
	told := first.Rows
	if len(told) > 50 {
		told = told[:50]
	}
	names := make([]string, len(first.Columns))
	for i, c := range first.Columns {
		names[i] = c.Name
	}
	result, _ := json.Marshal(map[string]any{"columns": names, "rows": told, "more": first.More || len(first.Rows) > len(told)})
	inputs := map[string]any{"question": question, "sql": first.SQL, "reading": reading, "visual": string(vis), "result": string(result),
		"rows": len(first.Rows), "asked_by": askedBy, "occurred_at": s.Clock.Today(), "report": vd.ID}
	if period != nil {
		inputs["period_from"], inputs["period_to"] = period.From, period.To
	}
	t, err := s.Exec.TriggerActivity(ctx, "host_answered", inputs, askedBy, "")
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	out := map[string]any{"event_id": t.EventID, "question": question, "report": vd.ID, "title": vd.Title, "components": comps, "reading": reading,
		"visual": first.Visual, "columns": first.Columns, "rows": first.Rows, "row_objects": first.Of, "more": first.More}
	if period != nil {
		out["period"] = period
	}
	writeJSON(w, 200, out)
}

// refusalClass is what the house may learn of a failed query: the class of
// the refusal and, for an unknown function, its name — never the query or
// the words around it, which may carry the client's names.
func refusalClass(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "a float"):
		return "a float column"
	case strings.Contains(msg, "machine's clock"):
		return "the machine's clock"
	case errors.Is(err, project.ErrNotOneSelect):
		return "not one SELECT"
	}
	if m := unknownFunction.FindStringSubmatch(msg); m != nil {
		return "Catalog Error: no function " + m[1]
	}
	if m := errorClass.FindString(msg); m != "" {
		return strings.TrimSuffix(m, ":")
	}
	return "the query could not be run"
}

var (
	unknownFunction = regexp.MustCompile(`Function with name ([A-Za-z_][A-Za-z0-9_]*) does not exist`)
	errorClass      = regexp.MustCompile(`^[A-Z][A-Za-z]* Error:`)
)

// componentOut is one query of hers, run: its rows as a face lays them out.
type componentOut struct {
	Title   string         `json:"title,omitempty"`
	Reading string         `json:"reading,omitempty"`
	SQL     string         `json:"-"` // the query as run, the period in it; the record's, never the face's
	Visual  core.Visual    `json:"visual"`
	Columns []answerColumn `json:"columns"`
	Rows    [][]string     `json:"rows"`
	Of      [][]string     `json:"row_objects,omitempty"`
	More    bool           `json:"more,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// runQuery runs one query of hers — an answer's or a report's component —
// with every guard: not the machine's clock, the locked session, no float
// shown. The lineage column is taken out as each row's objects, and a visual
// that names a column the rows lack falls back to the rows as a table.
func (s *Server) runQuery(ctx context.Context, sql string, visual core.Visual) (componentOut, error) {
	if err := clockless(sql); err != nil {
		return componentOut{}, err
	}
	cols, types, rows, more, err := s.runLocked(ctx, sql)
	if err == nil {
		err = exact(cols, types)
	}
	if err != nil {
		return componentOut{}, err
	}
	out := componentOut{Visual: visual, More: more, Columns: []answerColumn{}, Rows: make([][]string, len(rows))}
	of := slices.Index(cols, core.RowsOfColumn)
	for i, c := range cols {
		if i != of {
			out.Columns = append(out.Columns, answerColumn{Name: c, Type: types[i]})
		}
	}
	if of >= 0 {
		out.Of = make([][]string, len(rows))
	}
	for i, r := range rows {
		out.Rows[i] = []string{}
		for j, v := range r {
			if j == of {
				out.Of[i] = []string{}
				if v != "" {
					out.Of[i] = strings.Split(v, ",")
				}
				continue
			}
			out.Rows[i] = append(out.Rows[i], v)
		}
	}
	names := make([]string, len(out.Columns))
	for i, c := range out.Columns {
		names[i] = c.Name
	}
	if err := visual.Check(names); err != nil {
		out.Visual = core.Visual{Kind: "table", Title: visual.Title}
	}
	return out, nil
}

// serveReport runs a report's components on the copy as it stands, each in
// the locked session; one that fails says why, and the others still show. A
// report that takes a period is computed for the one asked (?from=&to=),
// else the first month the books have not closed, as an analysis view is.
func (s *Server) serveReport(ctx context.Context, w http.ResponseWriter, vd *core.ViewDef, q url.Values) {
	comps, period, code, err := s.reportComponents(ctx, vd, q)
	if err != nil {
		writeErr(w, code, err)
		return
	}
	out := map[string]any{"view": vd, "components": comps}
	if period != nil {
		out["period"] = period // what it was computed for: the face shows it and asks for another
	}
	writeJSON(w, 200, out)
}

// reportComponents runs a report's components, for the period asked where
// the report takes one; the status says whose fault a failure is.
func (s *Server) reportComponents(ctx context.Context, vd *core.ViewDef, q url.Values) ([]componentOut, *core.Period, int, error) {
	var spec core.ReportSpec
	if err := json.Unmarshal(vd.Spec, &spec); err != nil {
		return nil, nil, 500, err
	}
	if err := spec.Check(); err != nil {
		return nil, nil, 500, fmt.Errorf("view %s: %w", vd.ID, err)
	}
	var period *core.Period
	if spec.Period {
		p, code, err := s.periodOf(ctx, core.AnalysisSpec{Period: true}, q)
		if err != nil {
			return nil, nil, code, err
		}
		period = &p
	}
	out := make([]componentOut, 0, len(spec.Components))
	for _, c := range spec.Components {
		sql := c.SQL
		if period != nil {
			sql = period.SQL(sql)
		}
		got, err := s.runQuery(ctx, sql, c.Visual)
		if err != nil {
			got = componentOut{Visual: c.Visual, Error: err.Error(), Columns: []answerColumn{}, Rows: [][]string{}}
		}
		got.Title, got.Reading, got.SQL = c.Title, c.Reading, sql
		out = append(out, got)
	}
	return out, period, 0, nil
}

// The machine's clock is not the books' day: the kernel's clock is (a
// simulation's day differs from the wall's), so a query names the day.
var machineClock = regexp.MustCompile(`(?i)\b(current_date|current_timestamp|current_time|localtimestamp|now|today|get_current_timestamp|get_current_time)\b`)

func clockless(sql string) error {
	if m := machineClock.FindString(sql); m != "" {
		return fmt.Errorf("the query reads the machine's clock (%s); the books' day is given — write it as DATE 'YYYY-MM-DD'", m)
	}
	return nil
}

// exact refuses a float among the columns shown (invariant 6): money, its
// averages and its shares stay decimal.
func exact(cols, types []string) error {
	for i, t := range types {
		if cols[i] == core.RowsOfColumn {
			continue
		}
		switch strings.ToUpper(t) {
		case "DOUBLE", "FLOAT", "REAL":
			return fmt.Errorf("column %q comes back as %s, a float: money and its ratios stay exact — ROUND(..., 2)::DECIMAL(18,2), a percent ROUND(... * 100, 1)::DECIMAL(9,1)", cols[i], t)
		}
	}
	return nil
}

// runLocked runs a query of hers on the copy as it stands: rebuilt when the
// log moved, then read in the locked session.
func (s *Server) runLocked(ctx context.Context, sql string) ([]string, []string, [][]string, bool, error) {
	s.duck.Lock() // the rebuild and the read are one step
	defer s.duck.Unlock()
	path := project.PathOf(ctx, s.DuckPath) // the scope's own copy, the company acted for's within it
	if _, err := project.Rebuild(ctx, s.Store, path); err != nil {
		return nil, nil, nil, false, err
	}
	return project.RunTyped(ctx, path, sql, answerRows)
}

// Where a kept report sits (KK 2026-10-30): the activity of the books and
// its section of reports, by the ids a market's words arrange — a market
// whose words have no such section files it under the rest, never hidden.
const (
	KeptReportActivity = "books"
	KeptReportSection  = "reports"
)

// handleKeepAnswer keeps an answer as a report (KK: what is kept goes through
// the gate): the query she ran, its reading and its visual become a report
// with one component, drafted in a bundle that waits for a person's approval
// with its preview — no second model call, the kept query is the one that
// answered.
func (s *Server) handleKeepAnswer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// keeping drafts a bundle: the right to ask for one, never a reader's
	// (review 2026-10-21, A5)
	if !s.allow(w, r, "door", "draft_bundle", "system") {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("event"), 10, 64)
	if err != nil {
		writeErr(w, 400, fmt.Errorf("keep: %q is not an answer's event", r.PathValue("event")))
		return
	}
	ev, err := s.Store.GetEvent(ctx, id)
	if err != nil || ev.Type != core.EventHostAnswered {
		writeErr(w, 404, fmt.Errorf("keep: event %d is not an answer", id))
		return
	}
	var p struct {
		Question string `json:"question"`
		SQL      string `json:"sql"`
		Reading  string `json:"reading"`
		Visual   string `json:"visual"`
		AskedBy  string `json:"asked_by"`
		Report   string `json:"report"`
	}
	if err := json.Unmarshal(ev.Payload, &p); err != nil {
		writeErr(w, 500, err)
		return
	}
	if p.Report != "" {
		// a kept report answered: there is nothing to keep again (KK 2026-10-30)
		writeJSON(w, 409, map[string]any{"error": "the answer is the kept report " + p.Report + " already", "report": p.Report})
		return
	}
	var visual core.Visual
	if err := json.Unmarshal([]byte(p.Visual), &visual); err != nil {
		writeErr(w, 500, err)
		return
	}
	title := strings.TrimSpace(visual.Title)
	if title == "" {
		title = p.Question
	}
	if rs := []rune(title); len(rs) > 80 {
		title = string(rs[:80])
	}
	// the stretch of time the question named becomes the report's period
	// (KK 2026-10-30: asked for another month, it is the same report)
	sql, takes := core.Periodise(p.SQL)
	spec, _ := json.Marshal(core.ReportSpec{Period: takes, Components: []core.ReportComponent{{Title: title, Reading: p.Reading, SQL: sql, Visual: visual}}})
	viewID := fmt.Sprintf("answer-%d", id)
	bd := agent.BundleDraft{
		BundleID:    "keep-" + viewID,
		Description: fmt.Sprintf("Report kept from the question of %s: %s", ev.OccurredAt, p.Question),
		ViewDefs:    []core.ViewDef{{ID: viewID, Notion: "report", Title: title, Domain: "reports", Function: "report", Spec: spec}},
		// the kept report is an App of one tab, with a place and the faces
		// it shows on (KK 2026-10-30: a report approved must be reachable):
		// it sits with the books' reports, on the desk and the phone alike
		Apps: []core.App{{ID: viewID, Title: title, Layer: "business", Tabs: []core.AppTab{{View: viewID}},
			Place: &core.AppPlace{Activity: KeptReportActivity, Section: KeptReportSection}, Shows: []string{"desk", "phone"}}},
		Warrant: &core.Warrant{Basis: "client", Citations: []string{"asked by " + p.AskedBy + " on " + ev.OccurredAt + ": " + p.Question}, Scope: "this company"},
	}
	// kept already: the one decision, the one bundle
	if b, err := s.Store.GetBundle(ctx, bd.BundleID); err == nil {
		bv, err := s.bundleView(ctx, b)
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		writeJSON(w, 200, bv)
		return
	}
	types, err := s.Store.ListObjectTypes(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	if err := bd.Validate(types); err != nil {
		writeErr(w, 400, err)
		return
	}
	b, err := StoreBundleDraft(ctx, s.Store, s.knowledge(ctx), bd, p.AskedBy, ev.OccurredAt, p.Question, "")
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	bv, err := s.bundleView(ctx, b)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, bv)
}
