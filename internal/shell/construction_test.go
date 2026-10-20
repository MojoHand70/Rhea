package shell_test

import (
	"context"
	"fmt"
	"testing"

	"rhea/internal/core"
	"rhea/internal/store"
)

// TestConstructionAndAdvances (the auditor's exam, 2026-10-20, lessons I-3
// and I-14 of financial accounting 2025; EXAM.md 6.1.2 a, 6.1.1 c): a cost
// that is not an outlay on the asset — a fire during construction — leaves
// 080 for other operating costs; an advance paid for a fixed asset in a
// foreign currency is booked in the złoty that left the account that day and
// moves to the construction at that same amount when the delivery comes,
// with no exchange difference, because an advance is not a monetary item.
func TestConstructionAndAdvances(t *testing.T) {
	w := newCITWorld(t)
	ctx, s, door := w.ctx, w.s, w.door
	door("declare_company_form", map[string]any{"company": w.self, "vat_status": "active", "books": "full", "income_tax": "cit",
		"from": "2026-01-01", "occurred_at": "2026-08-01"})
	sums := func() map[string]int64 {
		t.Helper()
		accounts, _ := s.ObjectsByType(ctx, "account")
		code := map[string]string{}
		for _, a := range accounts {
			code[a.ID] = fmt.Sprint(a.State["code"])
		}
		ps, _ := s.ObjectsByType(ctx, "posting")
		out := map[string]int64{}
		for _, p := range ps {
			n, _ := p.State["amount"].(float64)
			out[code[fmt.Sprint(p.State["account"])]+" "+fmt.Sprint(p.State["side"])] += int64(n)
		}
		return out
	}
	// The builder's invoices go to the construction account once the person says so.
	w.buy("b1", "BUD/1/2026", "2026-09-10", "5260001246", "50000.00", "11500.00", "61500.00")
	supplier := ""
	for _, c := range mustObjects(t, s, ctx, "company") {
		if fmt.Sprint(c.State["vat_id"]) == "5260001246" {
			supplier = c.ID
		}
	}
	door("set_cost_account", map[string]any{"company": supplier, "account": "080", "occurred_at": "2026-09-01"})
	w.buy("b2", "BUD/2/2026", "2026-09-20", "5260001246", "20000.00", "4600.00", "24600.00")
	after := sums()
	if after["080 debit"] != 2000000 {
		t.Fatalf("the second invoice goes to the construction: %v", after)
	}
	// A fire: 3 000 of the invoiced work was clearing the damage, not building.
	door("move_from_construction", map[string]any{"title": "Usuwanie skutków pożaru", "amount": "3000.00", "date": "2026-09-25", "reason": "casualty", "occurred_at": "2026-09-25"})
	after = sums()
	if after["761 debit"] != 300000 || after["080 credit"] != 300000 {
		t.Fatalf("the casualty cost leaves the construction: %v", after)
	}

	// An advance of 10 000 EUR for the machine: 42 569 złoty left the account that day.
	door("pay_asset_advance", map[string]any{"title": "Zaliczka na maszynę", "supplier": supplier, "currency": "EUR", "amount_foreign": "10000.00", "amount_pln": "42569.00",
		"date": "2026-11-15", "occurred_at": "2026-11-15"})
	advances := mustObjects(t, s, ctx, "asset_advance")
	if len(advances) != 1 || advances[0].State["status"] != "open" {
		t.Fatalf("advance: %+v", advances)
	}
	if pln, _ := advances[0].State["amount_pln"].(float64); int64(pln) != 4256900 {
		t.Fatalf("advance keeps the złoty of the day it was paid: %v", advances[0].State)
	}
	after = sums()
	if after["249 debit"] != 4256900 || after["130 credit"] != 4256900 {
		t.Fatalf("advance postings: %v", after)
	}
	// The machine arrives in February at a different rate: the advance moves at its złoty, no difference.
	door("settle_asset_advance", map[string]any{"advance": advances[0].ID, "date": "2027-02-10", "occurred_at": "2027-02-10"})
	after = sums()
	if after["080 debit"] != 2000000+4256900 || after["249 credit"] != 4256900 || after["750 credit"] != 0 || after["751 debit"] != 0 {
		t.Fatalf("settlement: %v", after)
	}
	if a := mustObjects(t, s, ctx, "asset_advance")[0]; a.State["status"] != "settled" || a.State["settled_on"] != "2027-02-10" {
		t.Fatalf("settled: %v", a.State)
	}
}

func mustObjects(t *testing.T, s *store.Store, ctx context.Context, typ string) []core.Object {
	t.Helper()
	objs, err := s.ObjectsByType(ctx, typ)
	if err != nil {
		t.Fatal(err)
	}
	return objs
}
