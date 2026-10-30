package agent_test

import (
	"testing"

	"rhea/internal/agent"
)

// Without RHEA_MODEL the kernel runs on Claude Sonnet 5.5 (DECISIONS 2026-10-30).
func TestDefaultModelIsSonnet55(t *testing.T) {
	t.Setenv("RHEA_MODEL", "")
	if got := agent.Model(); got != "claude-sonnet-5-5" {
		t.Fatalf("default model = %q", got)
	}
}
