//go:build measure

package shell_test

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strconv"
	"testing"
	"time"

	"rhea/internal/exec"
	"rhea/internal/store"
)

// TestMeasureOneInvoice says what one invoice's booking costs on a large
// company's books — a database a measurement kept (RHEA_MEASURE_DB, e.g.
// rhea_measure_38k from TestMeasureLookBack) — run only when asked for:
//
//	RHEA_MEASURE_DB=rhea_measure_38k go test -tags measure ./internal/shell \
//	  -run TestMeasureOneInvoice -v -cpuprofile cpu.out
//
// RHEA_MEASURE_K sales invoices (default 20) are booked one by one through
// the door the server books them through, each timed. The books keep them.
func TestMeasureOneInvoice(t *testing.T) {
	name := os.Getenv("RHEA_MEASURE_DB")
	if name == "" {
		t.Fatal("RHEA_MEASURE_DB names the database to measure on")
	}
	k := 20
	if v, err := strconv.Atoi(os.Getenv("RHEA_MEASURE_K")); err == nil && v > 0 {
		k = v
	}
	ctx := context.Background()
	s, err := store.Open(ctx, "postgres://rhea:rhea@127.0.0.1:5432/"+name)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	x := &exec.Executor{Store: s}
	run := time.Now().UnixNano()
	var took []time.Duration
	for i := 0; i < k; i++ {
		date := "2026-10-16"
		t0 := time.Now()
		tr, err := x.TriggerActivity(ctx, "submit_event", map[string]any{"event_type": "sales.invoice.issued", "occurred_at": date, "payload": map[string]any{
			"number": fmt.Sprintf("FV M%d/%d", run, i), "issue_date": date, "market": "pl", "intragroup": "no", "seller_nip": "5250001111",
			"buyer_nip": fmt.Sprintf("7%09d", i), "net": "100.00", "vat": "23.00", "gross": "123.00", "vat_rate": "23", "currency": "PLN"}},
			"test", fmt.Sprintf("measure-%d-%d", run, i))
		if err != nil || tr.Booked != 1 || len(tr.Errors) != 0 {
			t.Fatalf("invoice %d: %+v %v", i, tr, err)
		}
		took = append(took, time.Since(t0))
	}
	sorted := slices.Clone(took)
	slices.Sort(sorted)
	fmt.Printf("%d invoices on %s: median %s, fastest %s, slowest %s\n", k, name,
		sorted[len(sorted)/2].Round(time.Millisecond), sorted[0].Round(time.Millisecond), sorted[len(sorted)-1].Round(time.Millisecond))
}
