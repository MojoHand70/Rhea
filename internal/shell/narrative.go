package shell

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/project"
)

// The narrative notion (DIRECTION 2026-10-10, "The host is a concierge":
// the briefing). The kernel computes the facts — the views the spec names,
// analysis tables and matching grids — and the host phrases them: a few
// sentences in the owner's voice, fresh each time, logged through the
// host_said door so the next words know the last. The host never computes
// a number: every number in its words must be one the facts carry, or the
// words are refused and the facts are shown alone, with the reason. The
// host never denies a fact either (DECISIONS 2026-10-15): a fact with an
// amount in it must be spoken — one of its amounts, exactly — or the words
// are refused the same way. Without an agent the host is silent and the
// facts speak for themselves.

type narrativeFact struct {
	As    string      `json:"as"`
	View  string      `json:"view"`
	Title string      `json:"title"`
	Kind  string      `json:"kind"` // analysis | matching
	Data  interface{} `json:"data"`
}

func (s *Server) serveNarrative(ctx context.Context, w http.ResponseWriter, vd *core.ViewDef) {
	var spec core.NarrativeSpec
	if err := json.Unmarshal(vd.Spec, &spec); err != nil {
		writeErr(w, 500, err)
		return
	}
	vds, err := s.effectiveViewDefs(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	byID := map[string]core.ViewDef{}
	for _, v := range vds {
		byID[v.ID] = v
	}
	var facts []narrativeFact
	for _, f := range spec.Facts {
		fv, ok := byID[f.View]
		if !ok {
			writeErr(w, 400, fmt.Errorf("narrative %s: no view %q to draw on", vd.ID, f.View))
			return
		}
		as := f.As
		if as == "" {
			as = fv.Title
		}
		fact := narrativeFact{As: as, View: fv.ID, Title: fv.Title, Kind: fv.Notion}
		switch fv.Notion {
		case "analysis":
			var as core.AnalysisSpec
			if err := json.Unmarshal(fv.Spec, &as); err != nil {
				writeErr(w, 500, err)
				return
			}
			cols, rows, _, err := s.analysisData(ctx, as)
			if err != nil {
				writeErr(w, 500, err)
				return
			}
			fact.Data = map[string]any{"columns": cols, "rows": rows}
		case "matching":
			data, err := s.matchingData(ctx, &fv)
			if err != nil {
				writeErr(w, 500, err)
				return
			}
			fact.Data = data
		default:
			writeErr(w, 400, fmt.Errorf("narrative %s: a host speaks from analysis views and matching grids, not a %s", vd.ID, fv.Notion))
			return
		}
		facts = append(facts, fact)
	}
	out := map[string]any{"view": vd, "facts": facts, "today": s.Clock.TodayUTC()}
	factsJSON, _ := json.Marshal(facts)
	before, err := s.saidBefore(ctx, vd.ID, 3)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	if s.Agent == nil {
		out["silent"] = "no host to speak: the facts speak for themselves"
		writeJSON(w, 200, out)
		return
	}
	words, err := s.Agent.Narrate(ctx, s.Clock.TodayUTC(), spec.Voice, factsJSON, before)
	if err != nil {
		out["silent"] = "the host could not speak: " + err.Error()
		writeJSON(w, 200, out)
		return
	}
	words = strings.TrimSpace(words)
	if foreign := foreignNumbers(words, factsJSON); len(foreign) > 0 {
		// The host never computes a number: words carrying one the facts
		// do not are refused, and the refusal is shown, not the words.
		out["silent"] = fmt.Sprintf("the host's words were refused: they carry numbers the facts do not (%s)", strings.Join(foreign, ", "))
		writeJSON(w, 200, out)
		return
	}
	if unspoken := unspokenFacts(words, factsJSON); len(unspoken) > 0 {
		// The host never denies a fact: words that pass over a fact with an
		// amount in it are refused, and the fact passed over is named.
		out["silent"] = fmt.Sprintf("the host's words were refused: they pass over %s", strings.Join(unspoken, "; "))
		writeJSON(w, 200, out)
		return
	}
	// The words are a fact: logged through the door, by the agent.
	t, err := s.Exec.TriggerActivity(ctx, "host_said", map[string]any{"view": vd.ID, "text": words}, "agent:"+s.model(), "")
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	out["said"], out["said_event"] = words, t.EventID
	writeJSON(w, 200, out)
}

// saidBefore reads the host's last words on a view, newest first.
func (s *Server) saidBefore(ctx context.Context, viewID string, n int) ([]string, error) {
	raws, err := s.Store.EventsByKind(ctx, core.KindRaw)
	if err != nil {
		return nil, err
	}
	var out []string
	for i := len(raws) - 1; i >= 0 && len(out) < n; i-- {
		if raws[i].Type != core.EventHostSaid {
			continue
		}
		var p struct {
			View string `json:"view"`
			Text string `json:"text"`
		}
		if json.Unmarshal(raws[i].Payload, &p) == nil && p.View == viewID {
			out = append(out, p.Text)
		}
	}
	return out, nil
}

// foreignNumbers lists the numbers in the words that the facts do not
// carry — read as people write them (core.WrittenNumbers) and compared as
// values, so "1 234,56" and "1234.56" are one number, "100 zł" is "100.00"
// (nothing was rounded), and "1234" is not "1234.56" (something was).
func foreignNumbers(words string, facts []byte) []string {
	have := map[string]bool{}
	for _, n := range core.WrittenNumbers(string(facts)) {
		have[n.Value] = true
	}
	var out []string
	seen := map[string]bool{}
	for _, n := range core.WrittenNumbers(words) {
		if have[n.Value] || seen[n.Value] {
			continue
		}
		// A date's parts and small counts the facts imply are not numbers the
		// host computed; a number below 32 with no decimals passes.
		if !strings.Contains(n.Value, ".") && len(n.Value) <= 2 {
			continue
		}
		seen[n.Value] = true
		out = append(out, n.Text)
	}
	return out
}

// unspokenFacts lists the facts the words pass over: a fact with an amount
// in it is spoken when at least one of its amounts appears in the words,
// digit for digit. A fact's amounts are the non-zero cells of its money and
// count columns — an analysis view's typed columns, a matching grid's open
// amounts — or, when no column is typed, every number its cells carry. An
// empty fact (no rows, no open items, only zeros) asks nothing of the host:
// "nothing waits" is then the truth, not a denial. The check is the mirror
// of foreignNumbers: that one refuses a number the facts do not carry, this
// one refuses silence about a fact they do.
func unspokenFacts(words string, facts []byte) []string {
	said := map[string]bool{}
	for _, n := range core.WrittenNumbers(words) {
		said[n.Value] = true
	}
	var shape []struct {
		As   string `json:"as"`
		Kind string `json:"kind"`
		Data struct {
			Columns []struct {
				Type string `json:"type"`
			} `json:"columns"`
			Rows [][]struct {
				V string `json:"v"`
			} `json:"rows"`
			Sides []struct {
				Columns []struct {
					Field string `json:"field"`
					Type  string `json:"type"`
				} `json:"columns"`
				Items []struct {
					Cells map[string]struct {
						V string `json:"v"`
					} `json:"cells"`
				} `json:"items"`
			} `json:"sides"`
		} `json:"data"`
	}
	if json.Unmarshal(facts, &shape) != nil {
		return nil
	}
	var out []string
	for _, f := range shape {
		var amounts []string
		seen := map[string]bool{}
		add := func(cell string) {
			for _, n := range core.WrittenNumbers(cell) {
				if n.Value == "0" || seen[n.Value] {
					continue
				}
				seen[n.Value] = true
				amounts = append(amounts, n.Value)
			}
		}
		switch f.Kind {
		case "analysis":
			typed := false
			for _, c := range f.Data.Columns {
				if amountColumn(c.Type) {
					typed = true
				}
			}
			for _, row := range f.Data.Rows {
				for j, cell := range row {
					if !typed || (j < len(f.Data.Columns) && amountColumn(f.Data.Columns[j].Type)) {
						add(cell.V)
					}
				}
			}
		case "matching":
			for _, side := range f.Data.Sides {
				for _, item := range side.Items {
					for _, col := range side.Columns {
						if amountColumn(col.Type) {
							add(item.Cells[col.Field].V)
						}
					}
				}
			}
		}
		if len(amounts) == 0 {
			continue
		}
		spoken := false
		for _, a := range amounts {
			if said[a] {
				spoken = true
				break
			}
		}
		if spoken {
			continue
		}
		shown := amounts
		if len(shown) > 3 {
			shown = append(append([]string{}, shown[:3]...), "…")
		}
		out = append(out, fmt.Sprintf("%s (%s)", f.As, strings.Join(shown, ", ")))
	}
	return out
}

// amountColumn: the column types whose cells are amounts the host must
// carry — money and counts.
func amountColumn(t string) bool { return t == "money" || t == "int" }

// analysisData runs an analysis view's query on the read side, typed.
// analysisData runs an analysis view. The row-lineage column (core.
// RowsOfColumn), when the view's SQL has one, is taken out of the columns
// and returned beside the rows: what each row is made of, never a figure —
// so the host's facts, which carry the rows, never carry object ids.
func (s *Server) analysisData(ctx context.Context, spec core.AnalysisSpec) ([]columnOut, [][]cellOut, [][]string, error) {
	if _, err := project.Rebuild(ctx, s.Store, s.DuckPath); err != nil {
		return nil, nil, nil, err
	}
	cols, rows, err := project.Query(ctx, s.DuckPath, spec.SQL)
	if err != nil {
		return nil, nil, nil, err
	}
	of := -1
	var outCols []columnOut
	for i, c := range cols {
		if c == core.RowsOfColumn {
			of = i
			continue
		}
		col := columnOut{Field: c, Label: c}
		if k := len(outCols); k < len(spec.Columns) {
			sc := spec.Columns[k]
			if sc.Field != "" {
				col.Field = sc.Field
			}
			if sc.Label != "" {
				col.Label = sc.Label
			}
			col.Type = sc.Type
		}
		outCols = append(outCols, col)
	}
	outRows := make([][]cellOut, len(rows))
	var rowsOf [][]string
	if of >= 0 {
		rowsOf = make([][]string, len(rows))
	}
	for i, r := range rows {
		outRows[i] = make([]cellOut, 0, len(outCols))
		for j, v := range r {
			if j == of {
				if v != "" {
					rowsOf[i] = strings.Split(v, ",")
				} else {
					rowsOf[i] = []string{}
				}
				continue
			}
			k := len(outRows[i])
			if outCols[k].Type == "money" && v != "" {
				if minor, err := parseInt(v); err == nil {
					outRows[i] = append(outRows[i], cellOut{V: core.FormatMoney(minor)})
					continue
				}
			}
			outRows[i] = append(outRows[i], cellOut{V: v})
		}
	}
	return outCols, outRows, rowsOf, nil
}

// model names the host's author on the door: the model the agent runs.
func (s *Server) model() string { return agent.Model() }
