package pack_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"rhea/internal/pack"
)

// TestEveryPackSaysWhatItIs (KK 2026-10-18: "every pack must contain a
// disclaimer that it was prepared by AI and that it is an initial version";
// "users should see that packs develop"): every pack in the repository
// names its stage, carries its disclaimer in its own language, and keeps a
// history whose newest entry is the pack's version and whose entries run
// down to version 1 without a gap.
func TestEveryPackSaysWhatItIs(t *testing.T) {
	var files []string
	filepath.WalkDir(filepath.Join("..", "..", "packs"), func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".json") {
			files = append(files, p)
		}
		return nil
	})
	if len(files) < 10 {
		t.Fatalf("found %d pack files", len(files))
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var m pack.Manifest
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		if m.Pack == "" {
			continue // not a pack manifest
		}
		if !slices.Contains(pack.Stages, m.Stage) {
			t.Errorf("%s: stage %q is not one of %v", f, m.Stage, pack.Stages)
		}
		if strings.TrimSpace(m.Disclaimer) == "" {
			t.Errorf("%s: no disclaimer", f)
		}
		if len(m.History) == 0 {
			t.Errorf("%s: no history", f)
			continue
		}
		next := m.Version // the version the next entry must reach down from
		for _, h := range m.History {
			lo, hi, err := span(h.Versions)
			if err != nil || hi != next || lo > hi || strings.TrimSpace(h.Changes) == "" {
				t.Errorf("%s: history entry %q (%v) does not continue from version %d", f, h.Versions, err, next)
				break
			}
			next = lo - 1
		}
		if next != 0 {
			t.Errorf("%s: history stops at version %d, not 1", f, next+1)
		}
	}
}

// span reads "7" or "1–18" (en dash or hyphen).
func span(s string) (int, int, error) {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '–' || r == '-' })
	if len(parts) == 1 {
		n, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		return n, n, err
	}
	lo, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, err
	}
	hi, err := strconv.Atoi(strings.TrimSpace(parts[len(parts)-1]))
	return lo, hi, err
}
