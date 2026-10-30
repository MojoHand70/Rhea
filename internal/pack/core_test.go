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
	"rhea/internal/store"
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
// two rules (core v2), and its thirteen Apps (core v3, 2026-10-26), the
// trial balance and the group's by period (core v6, v7, 2026-10-27) — and
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
		// the kinds the core added since (v8: Nowy klient, Nowy dostawca —
		// two doors, four rules), and bank_account v1, which the
		// Poland pack no longer carries since v54 brings v2 (2026-10-27)
		added := sum.Pack == "core" && sum.Types == 4 && sum.Views == 4 && sum.Rules == 10 && sum.Activities == 6 && sum.Apps == 18 && sum.Bundle == "pack-core-v8"
		if !added && (sum.Bundle != "" || sum.Types+sum.Views+sum.Rules+sum.Activities+sum.Upgraded != 0) || len(sum.Conflicts) != 0 {
			t.Fatalf("init drafted more than the core added since on a database that holds it: %+v", sum)
		}
	}
}

// A core version loads where a market pack has extended the core's Apps and
// kinds of record before (white01, 2026-10-27): an App's versions and a
// kind's are one sequence across packs, and the Poland pack had taken the
// "parties" App's v2 and v3 and the company kind's v2 to v7. The core v6
// first written gave its "parties" App and its company a v2 of their own,
// and every deploy stopped at rhea init — "app parties v2 differs from the
// one stored" — leaving the new programs installed and the old ones running.
// A core never takes a number a market has used for the same App or kind.
func TestTheCoreLoadsWhereAMarketExtendedItsAppsBefore(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	for v := 2; v <= 3; v++ {
		a := core.App{ID: "parties", Version: v, Title: "Kontrahenci", Layer: "law", Data: []string{"company"}}
		if err := store.InsertAppRow(ctx, s.Pool, a, core.StatusActive); err != nil {
			t.Fatal(err)
		}
	}
	for v := 2; v <= 7; v++ {
		ot := core.ObjectType{Name: "company", Version: v, Domain: "finance", LabelField: "name",
			Fields: []core.FieldDef{{Name: "name", Type: "string", Required: true}, {Name: "kind", Type: "string", Required: true}, {Name: "nip", Type: "string"}}}
		if err := store.InsertObjectTypeRow(ctx, s.Pool, ot, core.StatusActive); err != nil {
			t.Fatal(err)
		}
	}
	sums, err := pack.LoadCore(ctx, s, "init")
	if err != nil {
		t.Fatalf("init on the Poland pack's history: %v", err)
	}
	for _, sum := range sums {
		if len(sum.Conflicts) != 0 {
			t.Fatalf("%s: %v", sum.Pack, sum.Conflicts)
		}
	}
}
