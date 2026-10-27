package shell

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"rhea/internal/agent"
)

// The live check of the host's words (DIRECTION 2026-10-14, the model
// measurement): the model named by RHEA_MODEL is handed one recorded
// morning — the facts the kernel computed in the Poland story — and asked
// to speak, several times in a row, each time told what it said before.
// The kernel's own law is applied to every answer: an amount unlike its
// reference or written without one refuses the words (KK 2026-10-27, her
// figures are the books'), so does a number the facts do not carry, a fact
// passed over, and repeating itself. The verdict
// is printed, not asserted: this is a measurement, run by hand with
// RHEA_LIVE_HOST=1 (RHEA_LIVE_HOST_RUNS picks how many times, default 5).
func TestHostWordsLive(t *testing.T) {
	if os.Getenv("RHEA_LIVE_HOST") == "" {
		t.Skip("set RHEA_LIVE_HOST=1 to ask the live model for the host's words")
	}
	raw, err := os.ReadFile(filepath.Join("..", "..", "testdata", "eval", "briefing.json"))
	if err != nil {
		t.Fatal(err)
	}
	var morning struct {
		Today string          `json:"today"`
		Voice string          `json:"voice"`
		Facts json.RawMessage `json:"facts"`
	}
	if err := json.Unmarshal(raw, &morning); err != nil {
		t.Fatal(err)
	}
	runs := 5
	if n, err := strconv.Atoi(os.Getenv("RHEA_LIVE_HOST_RUNS")); err == nil && n > 0 {
		runs = n
	}
	facts := recordedFacts(t, morning.Facts)
	plain, _ := json.Marshal(facts)
	refs := referFigures(facts)
	withRefs, _ := json.Marshal(facts)
	a := agent.New()
	ctx := context.Background()
	var before []string
	accepted := 0
	for i := 1; i <= runs; i++ {
		start := time.Now()
		words, err := a.Narrate(ctx, morning.Today, morning.Voice, withRefs, before, "")
		took := time.Since(start).Round(time.Second)
		if err != nil {
			t.Logf("run %d (%s): the host could not speak: %v", i, took, err)
			continue
		}
		words = strings.TrimSpace(words)
		verdict := "accepted"
		clean, figs, why := herFigures(words, refs)
		if why != "" {
			verdict = "refused: " + why
		} else if foreign := foreignNumbers(clean, plain); len(foreign) > 0 {
			verdict = "refused: numbers the facts do not carry (" + strings.Join(foreign, ", ") + ")"
		} else if unspoken := unspokenFacts(clean, plain); len(unspoken) > 0 {
			verdict = "refused: passes over " + strings.Join(unspoken, "; ")
		} else if slices.Contains(before, words) {
			verdict = "refused: repeated itself"
		} else {
			accepted++
		}
		t.Logf("run %d (%s, %d chars, %d figures): %s\n  -> %s", i, took, len(words), len(figs), words, verdict)
		if why == "" {
			words = clean
		}
		before = append([]string{words}, before...)
		if len(before) > 3 {
			before = before[:3]
		}
	}
	t.Logf("%s: %d/%d accepted", agent.Model(), accepted, runs)
}

// recordedFacts reads recorded facts back into the shape the kernel hands
// the host, so their amounts take their references as they would live.
func recordedFacts(t *testing.T, raw json.RawMessage) []narrativeFact {
	t.Helper()
	var recorded []struct {
		narrativeFact
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &recorded); err != nil {
		t.Fatal(err)
	}
	out := make([]narrativeFact, 0, len(recorded))
	for _, r := range recorded {
		f := r.narrativeFact
		var data map[string]json.RawMessage
		if err := json.Unmarshal(r.Data, &data); err != nil {
			t.Fatal(err)
		}
		d := map[string]any{}
		for k, v := range data {
			d[k] = v
		}
		switch f.Kind {
		case "analysis":
			var cols []columnOut
			var rows [][]cellOut
			if json.Unmarshal(data["columns"], &cols) != nil || json.Unmarshal(data["rows"], &rows) != nil {
				t.Fatalf("recorded fact %s", f.View)
			}
			d["columns"], d["rows"] = cols, rows
		case "matching":
			var sides []matchingSide
			if json.Unmarshal(data["sides"], &sides) != nil {
				t.Fatalf("recorded fact %s", f.View)
			}
			d["sides"] = sides
		}
		f.Data = d
		out = append(out, f)
	}
	return out
}
