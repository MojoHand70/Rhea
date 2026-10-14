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
// The kernel's own law is applied to every answer: a number the facts do
// not carry refuses the words, and so does repeating itself. The verdict
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
	a := agent.New()
	ctx := context.Background()
	var before []string
	accepted := 0
	for i := 1; i <= runs; i++ {
		start := time.Now()
		words, err := a.Narrate(ctx, morning.Today, morning.Voice, morning.Facts, before)
		took := time.Since(start).Round(time.Second)
		if err != nil {
			t.Logf("run %d (%s): the host could not speak: %v", i, took, err)
			continue
		}
		words = strings.TrimSpace(words)
		verdict := "accepted"
		if foreign := foreignNumbers(words, morning.Facts); len(foreign) > 0 {
			verdict = "refused: numbers the facts do not carry (" + strings.Join(foreign, ", ") + ")"
		} else if slices.Contains(before, words) {
			verdict = "refused: repeated itself"
		} else {
			accepted++
		}
		t.Logf("run %d (%s, %d chars): %s\n  -> %s", i, took, len(words), words, verdict)
		before = append([]string{words}, before...)
		if len(before) > 3 {
			before = before[:3]
		}
	}
	t.Logf("%s: %d/%d accepted", agent.Model(), accepted, runs)
}
