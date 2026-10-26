package pack_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/store/storetest"
)

// The core is shipped by the kernel (DIRECTION 2026-10-23, step 1): on a
// fresh database rhea init drafts it — the common packs and the founding
// vocabulary with Rachunki i karty — and nothing activates until a person
// approves; approved, the Poland pack finds everything it requires and
// installs on it; a second init drafts nothing. Before, the core was
// shipped by nothing: white01 got it from a restore, and a fresh server
// failed at its first pack.
func TestTheKernelShipsItsCore(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}

	sums, err := pack.LoadCore(ctx, s, "init")
	if err != nil {
		t.Fatal(err)
	}
	var waiting []string
	for _, sum := range sums {
		if sum.Bundle == "" {
			t.Fatalf("pack %s drafted no bundle on a fresh database", sum.Pack)
		}
		waiting = append(waiting, sum.Bundle)
	}
	if len(waiting) != 4 || !strings.HasPrefix(waiting[1], "pack-core-v") {
		t.Fatalf("waiting = %v, want Ludzie, the core, then Okresy and Dokumenty", waiting)
	}
	if _, err := s.GetObjectType(ctx, "posting"); err == nil {
		t.Fatal("the core's posting is active before anyone approved it")
	}
	// Okresy's rule writes the core's period lock: approved before the core,
	// it would run over a type nobody approved yet — refused, the order said
	if _, _, _, err := x.ApproveBundle(ctx, waiting[2], "krzysztof", "2026-01-01"); err == nil {
		t.Fatalf("%s was approved before the core it writes into", waiting[2])
	}
	for _, b := range waiting {
		if _, _, errs, err := x.ApproveBundle(ctx, b, "krzysztof", "2026-01-01"); err != nil || len(errs) != 0 {
			t.Fatalf("approve %s: %v %v", b, err, errs)
		}
	}

	pl, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", "pl", "pack.json"), "test")
	if err != nil {
		t.Fatalf("the Poland pack does not load on the core: %v", err)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, pl.Bundle, "krzysztof", "2026-01-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve %s: %v %v", pl.Bundle, err, errs)
	}

	again, err := pack.LoadCore(ctx, s, "init")
	if err != nil {
		t.Fatal(err)
	}
	for _, sum := range again {
		if sum.Bundle != "" || sum.Types+sum.Views+sum.Rules+sum.Activities+sum.Upgraded != 0 {
			t.Fatalf("a second init drafted %+v", sum)
		}
	}
}

// A database built from the founding files, as white01 was (the founding
// vocabulary inserted, the common packs and the Poland pack approved),
// already holds the core as it first shipped: init drafts there only what the
// core has added since — the installed_pack record, its two doors and their
// two rules (core v2), and its thirteen Apps (core v3, 2026-10-26) — and
// nothing of the common packs.
func TestACoreAlreadyThereDraftsNothing(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	for _, file := range []string{"finance.json", "finance_v2.json", "finance_v3.json", "finance_v4.json", "finance_v5.json",
		"finance_v6.json", "finance_v7.json", "finance_v8.json", "finance_v9.json"} {
		b, err := os.ReadFile(filepath.Join("..", "..", "testdata", "seed", file))
		if err != nil {
			t.Fatal(err)
		}
		var defs struct {
			ObjectTypes []core.ObjectType `json:"object_types"`
			ViewDefs    []core.ViewDef    `json:"view_defs"`
		}
		if err := json.Unmarshal(b, &defs); err != nil {
			t.Fatal(err)
		}
		for _, ot := range defs.ObjectTypes {
			if err := s.InsertObjectType(ctx, ot); err != nil {
				t.Fatal(err)
			}
		}
		for _, vd := range defs.ViewDefs {
			if err := s.InsertViewDef(ctx, vd); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, p := range []string{"common/people.json", "common/periods.json", "common/documents.json", "pl/pack.json"} {
		sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", p), "test")
		if err != nil {
			t.Fatal(err)
		}
		if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-01-01"); err != nil || len(errs) != 0 {
			t.Fatalf("approve %s: %v %v", sum.Bundle, err, errs)
		}
	}
	sums, err := pack.LoadCore(ctx, s, "init")
	if err != nil {
		t.Fatal(err)
	}
	for _, sum := range sums {
		added := sum.Pack == "core" && sum.Types == 3 && sum.Views == 2 && sum.Rules == 6 && sum.Activities == 4 && sum.Apps == 18 && sum.Bundle == "pack-core-v5"
		if !added && (sum.Bundle != "" || sum.Types+sum.Views+sum.Rules+sum.Activities+sum.Upgraded != 0) || len(sum.Conflicts) != 0 {
			t.Fatalf("init drafted more than the core added since on a database that holds it: %+v", sum)
		}
	}
}
