package exec_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/store/storetest"
)

// A month is closed once (review 2026-10-21, D3): close_period had no
// identity and nothing downstream asked whether the month was closed, so a
// second close of September — by another person, or the same one twice —
// booked a second lock and, through the Poland pack, a second VAT
// settlement and a second CIT advance: account 225 overstated, October's
// carried-in VAT summing both, the JPK filing what was never declared. Now
// the door says what makes a close the same close — the month and the book
// — and a second one is refused at the door, so nothing downstream fires.
func TestAMonthIsClosedOnce(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	if err := s.InsertObjectType(ctx, core.ObjectType{Name: "period_lock", Version: 2, Domain: "finance",
		Fields: []core.FieldDef{{Name: "month", Type: "string", Required: true}, {Name: "book", Type: "string", Required: true}}}); err != nil {
		t.Fatal(err)
	}
	sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", "common", "periods.json"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-10-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve: %v %v", err, errs)
	}
	closeMonth := func(day, book string) error {
		t.Helper()
		_, err := x.TriggerActivity(ctx, "close_period", map[string]any{"period": day, "book": book, "occurred_at": "2026-10-05"}, "krzysztof", "")
		return err
	}
	if err := closeMonth("2026-09-30", "pl-stat"); err != nil {
		t.Fatal(err)
	}
	if err := closeMonth("2026-09-05", "pl-stat"); err == nil || !strings.Contains(err.Error(), "already") {
		t.Fatalf("September closed twice: %v", err)
	}
	if err := closeMonth("2026-10-31", "pl-stat"); err != nil {
		t.Fatalf("October: %v", err)
	}
	if err := closeMonth("2026-09-30", "main"); err != nil {
		t.Fatalf("September in another book: %v", err)
	}
	locks, err := s.ObjectsByType(ctx, "period_lock")
	if err != nil || len(locks) != 3 {
		t.Fatalf("three locks, one per month and book: %v %d", err, len(locks))
	}
}
