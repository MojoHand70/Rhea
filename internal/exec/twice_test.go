package exec_test

import (
	"testing"

	"rhea/internal/store/storetest"
)

// The matching stories book the same when told twice (invariant 4; the
// whole-tree review of 2026-10-21, H1): a bank statement paired with its
// invoices, a payment in another currency, netting, the three-way match,
// two companies settling with each other, and a pairing undone.
func TestMatchingBooksTheSameTwice(t *testing.T) {
	for _, s := range []struct {
		name  string
		story func(*testing.T)
	}{
		{"Bank", TestMatchingBankStory},
		{"FX", TestMatchingFX},
		{"Netting", TestMatchingNetting},
		{"ThreeWay", TestMatchingThreeWay},
		{"Intercompany", TestMatchingIntercompany},
		{"Unpairing", TestMatchingUnpairing},
	} {
		t.Run(s.name, func(t *testing.T) { storetest.Twice(t, s.story) })
	}
}
