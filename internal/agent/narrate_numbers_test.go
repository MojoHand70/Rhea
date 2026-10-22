package agent_test

import (
	"strings"
	"testing"

	"rhea/internal/agent"
)

// TestNarrateAsksForTheOwnersWayOfWritingNumbers: the host is told to say an
// amount with the facts' value but written the way the owner's language
// writes numbers. Told to copy it "exactly as given", she copied the facts'
// 665.83 into Polish words on white01 (2026-10-22, KK's screenshot) beside
// 44 900,00 written the Polish way — one morning, two ways of writing money.
// The kernel's laws compare values, so the language's way passes them; the
// instruction names no language, the voice does.
func TestNarrateAsksForTheOwnersWayOfWritingNumbers(t *testing.T) {
	said := strings.Join(strings.Fields(agent.NarrateSystem), " ") // the prompt's lines wrap anywhere
	if strings.Contains(said, "exactly as given") {
		t.Fatal("the host is still told to copy amounts exactly as given, decimal point and all")
	}
	if !strings.Contains(said, "the way the owner's language writes") {
		t.Fatal("the host is not told to write amounts the way the owner's language writes them")
	}
}
