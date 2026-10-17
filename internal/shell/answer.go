package shell

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
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
	c, err := s.collection(r.Context())
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, c)
}

// words is the market's names for what the kernel holds, from the active
// words view; nil without one.
func (s *Server) words(ctx context.Context) *core.WordsSpec {
	vds, err := s.effectiveViewDefs(ctx)
	if err != nil {
		return nil
	}
	for _, v := range vds {
		if v.Notion == "words" {
			var w core.WordsSpec
			if json.Unmarshal(v.Spec, &w) == nil {
				return &w
			}
		}
	}
	return nil
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
	ctx := r.Context()
	if s.Agent == nil {
		writeErr(w, 503, fmt.Errorf("answer: no author is connected"))
		return
	}
	var req struct {
		Question string `json:"question"`
		AskedBy  string `json:"asked_by"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 16<<10)).Decode(&req); err != nil {
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
	ask := agent.AnswerAsk{Question: req.Question, Today: s.Clock.TodayUTC(), Tables: tables, Words: s.words(ctx), Books: s.books(ctx)}
	var (
		ans   agent.Answer
		cols  []string
		types []string
		rows  [][]string
		more  bool
	)
	for try := 0; try < 2; try++ {
		if ans, err = s.Agent.Answer(ctx, ask); err != nil {
			writeErr(w, 502, err)
			return
		}
		if ans.Cannot != "" {
			writeJSON(w, 200, map[string]any{"question": req.Question, "cannot": ans.Cannot})
			return
		}
		cols, types, rows, more, err = s.runLocked(ctx, ans.SQL)
		if err == nil {
			break
		}
		log.Printf("answer: %q: %v", req.Question, err)
		ask.Tried = &agent.AnswerTry{SQL: ans.SQL, Error: err.Error()}
	}
	if err != nil {
		writeErr(w, 422, fmt.Errorf("answer: the query could not be run: %w", err))
		return
	}

	// The lineage column names each row's objects, for "why?"; it is never shown.
	of := slices.Index(cols, core.RowsOfColumn)
	var rowsOf [][]string
	shown := make([]answerColumn, 0, len(cols))
	for i, c := range cols {
		if i != of {
			shown = append(shown, answerColumn{Name: c, Type: types[i]})
		}
	}
	out := make([][]string, len(rows))
	if of >= 0 {
		rowsOf = make([][]string, len(rows))
	}
	for i, r := range rows {
		for j, v := range r {
			if j == of {
				if v != "" {
					rowsOf[i] = strings.Split(v, ",")
				} else {
					rowsOf[i] = []string{}
				}
				continue
			}
			out[i] = append(out[i], v)
		}
	}
	names := make([]string, len(shown))
	for i, c := range shown {
		names[i] = c.Name
	}
	visual := ans.Visual
	if err := visual.Check(names); err != nil {
		// a visual that misnames a column falls back to the rows themselves
		visual = core.Visual{Kind: "table", Title: visual.Title}
	}

	vis, _ := json.Marshal(visual)
	told := out
	if len(told) > 50 {
		told = told[:50]
	}
	result, _ := json.Marshal(map[string]any{"columns": names, "rows": told, "more": more || len(out) > len(told)})
	t, err := s.Exec.TriggerActivity(ctx, "host_answered", map[string]any{"question": req.Question, "sql": ans.SQL, "reading": ans.Reading,
		"visual": string(vis), "result": string(result), "rows": len(out), "asked_by": req.AskedBy, "occurred_at": s.Clock.TodayUTC()}, req.AskedBy, "")
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"event_id": t.EventID, "question": req.Question, "sql": ans.SQL, "reading": ans.Reading,
		"visual": visual, "columns": shown, "rows": out, "row_objects": rowsOf, "more": more})
}

// runLocked runs a query of hers on the copy as it stands: rebuilt when the
// log moved, then read in the locked session.
func (s *Server) runLocked(ctx context.Context, sql string) ([]string, []string, [][]string, bool, error) {
	s.duck.Lock() // the rebuild and the read are one step
	defer s.duck.Unlock()
	if _, err := project.Rebuild(ctx, s.Store, s.DuckPath); err != nil {
		return nil, nil, nil, false, err
	}
	return project.RunTyped(ctx, s.DuckPath, sql, answerRows)
}
