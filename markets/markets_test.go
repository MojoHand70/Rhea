package markets

import (
	"os/exec"
	"strings"
	"testing"
)

// Invariant 8 (CLAUDE.md; DIRECTION 2026-10-15, "Local packs, clean
// kernel"): a market's code lives under markets/<cc> behind the adapter
// contract, and the kernel never imports a market. The dependency points
// one way — a market imports the contract, the kernel knows no market by
// name — so a fix to the Polish pack is a fix to the Polish pack, never to
// the kernel, and Germany or France arrive without a kernel change.
//
// Two named exceptions, both outside the kernel: the binary (cmd/rhea)
// links each market by one blank import so the market registers itself;
// the simulator (internal/sim) writes a Polish persona's documents through
// the Polish writer — a bench tool's debt, to move into the market when a
// second market's personas exist.
func TestKernelNeverImportsAMarket(t *testing.T) {
	out, err := exec.Command("go", "list", "-f", "{{.ImportPath}} {{join .Imports \" \"}}", "rhea/...").Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	allowed := map[string]bool{"rhea/cmd/rhea": true, "rhea/internal/sim": true}
	var markets []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		pkg, imports := fields[0], fields[1:]
		if strings.HasPrefix(pkg, "rhea/markets/") {
			markets = append(markets, pkg)
			for _, imp := range imports {
				if strings.HasPrefix(imp, "rhea/markets/") && imp != pkg {
					t.Errorf("%s imports another market %s: markets share through the kernel's contract, not each other", pkg, imp)
				}
			}
			continue
		}
		for _, imp := range imports {
			if strings.HasPrefix(imp, "rhea/markets/") && !allowed[pkg] {
				t.Errorf("%s imports %s: the kernel never imports a market", pkg, imp)
			}
		}
	}
	if len(markets) == 0 {
		t.Fatal("no market packages found under rhea/markets")
	}
}
