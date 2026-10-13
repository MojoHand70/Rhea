package shell

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
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
// words are refused and the facts are shown alone, with the reason. Without
// an agent the host is silent and the facts speak for themselves.

type narrativeFact struct {
	As    string      `json:"as"`
	View  string      `json:"view"`
	Title string      `json:"title"`
	Kind  string      `json:"kind"` // analysis | matching
	Data  interface{} `json:"data"`
}

// numberRe finds the numbers in the host's words, with a dot or a comma
// as the decimal separator and spaces as thousand separators dropped.
var numberRe = regexp.MustCompile(`\d[\d ]*(?:[.,]\d+)?`)

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
			cols, rows, err := s.analysisData(ctx, as)
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
// carry — compared digit for digit, separators dropped, so "1 234,56" and
// "1234.56" are one number and "1234" is not "1234.56".
func foreignNumbers(words string, facts []byte) []string {
	have := map[string]bool{}
	for _, m := range numberRe.FindAllString(string(facts), -1) {
		have[canonicalNumber(m)] = true
	}
	var out []string
	seen := map[string]bool{}
	for _, m := range numberRe.FindAllString(words, -1) {
		c := canonicalNumber(m)
		if c == "" || have[c] || seen[c] {
			continue
		}
		// A date's parts and small counts the facts imply are not numbers the
		// host computed; a number below 32 with no decimals passes.
		if !strings.Contains(c, ".") && len(c) <= 2 {
			continue
		}
		seen[c] = true
		out = append(out, strings.TrimSpace(m))
	}
	return out
}

func canonicalNumber(m string) string {
	m = strings.ReplaceAll(strings.TrimSpace(m), " ", "")
	m = strings.ReplaceAll(m, ",", ".")
	m = strings.TrimLeft(m, "0")
	if strings.HasPrefix(m, ".") {
		m = "0" + m
	}
	if m == "" {
		return "0"
	}
	return m
}

// analysisData runs an analysis view's query on the read side, typed.
func (s *Server) analysisData(ctx context.Context, spec core.AnalysisSpec) ([]columnOut, [][]cellOut, error) {
	if _, err := project.Rebuild(ctx, s.Store, s.DuckPath); err != nil {
		return nil, nil, err
	}
	cols, rows, err := project.Query(ctx, s.DuckPath, spec.SQL)
	if err != nil {
		return nil, nil, err
	}
	outCols := make([]columnOut, len(cols))
	for i, c := range cols {
		outCols[i] = columnOut{Field: c, Label: c}
		if i < len(spec.Columns) {
			sc := spec.Columns[i]
			if sc.Field != "" {
				outCols[i].Field = sc.Field
			}
			if sc.Label != "" {
				outCols[i].Label = sc.Label
			}
			outCols[i].Type = sc.Type
		}
	}
	outRows := make([][]cellOut, len(rows))
	for i, r := range rows {
		outRows[i] = make([]cellOut, len(r))
		for j, v := range r {
			if outCols[j].Type == "money" && v != "" {
				if minor, err := parseInt(v); err == nil {
					outRows[i][j] = cellOut{V: core.FormatMoney(minor)}
					continue
				}
			}
			outRows[i][j] = cellOut{V: v}
		}
	}
	return outCols, outRows, nil
}

// model names the host's author on the door: the model the agent runs.
func (s *Server) model() string { return agent.Model() }
