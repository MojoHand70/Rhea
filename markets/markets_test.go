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
// second market's personas exist. Test binaries count too: the kernel's
// own tests may exercise a market (the shell's and the eval's Poland pack
// tests), named here so the list cannot grow unseen.
//
// The other direction is bounded as well: a market reaches the kernel
// through a named list of packages — the contract, the language, the
// statute library, the pack loader, the agent, and the exam harness the
// Polish exam command drives. The exam harness pulls in the eval and the
// shell behind it; moving its Polish parts under markets/pl is the debt
// that shortens this list (review 2026-10-20).
func TestKernelNeverImportsAMarket(t *testing.T) {
	out, err := exec.Command("go", "list", "-f",
		"{{.ImportPath}}|{{join .Imports \" \"}}|{{join .TestImports \" \"}} {{join .XTestImports \" \"}}", "rhea/...").Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	allowed := map[string]bool{"rhea/cmd/rhea": true, "rhea/internal/sim": true}
	allowedInTests := map[string]bool{"rhea/internal/eval": true, "rhea/internal/shell": true}
	marketMayImport := map[string]bool{
		"rhea/internal/adapter": true, "rhea/internal/core": true, "rhea/internal/statute": true,
		"rhea/internal/pack": true, "rhea/internal/agent": true, "rhea/internal/exam": true,
	}
	var markets []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			continue
		}
		pkg, imports, testImports := parts[0], strings.Fields(parts[1]), strings.Fields(parts[2])
		if strings.HasPrefix(pkg, "rhea/markets/") {
			markets = append(markets, pkg)
			for _, imp := range append(imports, testImports...) {
				if strings.HasPrefix(imp, "rhea/markets/") && imp != pkg {
					t.Errorf("%s imports another market %s: markets share through the kernel's contract, not each other", pkg, imp)
				}
			}
			// A market's own tests may drive the whole kernel (the pack end
			// to end); its code reaches it through the named packages only.
			for _, imp := range imports {
				if strings.HasPrefix(imp, "rhea/") && !strings.HasPrefix(imp, "rhea/markets") && !marketMayImport[imp] {
					t.Errorf("%s imports %s: a market reaches the kernel through the named packages only", pkg, imp)
				}
			}
			continue
		}
		for _, imp := range imports {
			if strings.HasPrefix(imp, "rhea/markets/") && !allowed[pkg] {
				t.Errorf("%s imports %s: the kernel never imports a market", pkg, imp)
			}
		}
		for _, imp := range testImports {
			if strings.HasPrefix(imp, "rhea/markets/") && !allowed[pkg] && !allowedInTests[pkg] {
				t.Errorf("%s's tests import %s: the kernel's tests exercise a market only where named", pkg, imp)
			}
		}
	}
	if len(markets) == 0 {
		t.Fatal("no market packages found under rhea/markets")
	}
}
