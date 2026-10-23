package shell_test

import (
	"testing"

	"rhea/internal/store/storetest"
)

// The stories the shell tells book the same when told twice (invariant 4;
// the whole-tree review of 2026-10-21, H1): the founding story, the Poland
// pack's year, the CIT advance backfilled into the past, two companies
// trading with each other, and the matching grid's pairings.
func TestStoriesBookTheSameTwice(t *testing.T) {
	for _, s := range []struct {
		name  string
		story func(*testing.T)
	}{
		{"M0", TestM0DemoStory},
		{"Poland", TestM3PolandPack},
		{"CITBackfilled", TestCITAdvanceBackfilled},
		{"Intercompany", TestIntercompany},
		{"Matching", TestMatchingGridAndPairing},
	} {
		t.Run(s.name, func(t *testing.T) { storetest.Twice(t, s.story) })
	}
}
