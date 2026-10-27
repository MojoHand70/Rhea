//go:build measure

package shell_test

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
)

// TestMeasureLookBack measures the look-back at a large company's size (KK
// 2026-10-27: "What if Client has 100 thousand invoices?") — run only when
// asked for, never by the suite:
//
//	RHEA_CI=1 RHEA_MEASURE_N=100000 go test -tags measure ./internal/shell \
//	  -run TestMeasureLookBack -v -timeout 6h
//
// N invoices (default 2,000), four in five sales to one buyer per twenty,
// one in five purchases from one supplier per sixty, January to October, are
// booked one by one through the door the server books them through; then
// Klienci and Dostawcy arrive (the core's version), and the look-back is
// asked, approved — with a door knocking while it runs, the server's 30 s
// wait for the books — and asked again. Last, a rule that adds a record to
// every sales invoice: the heavy case. Each step says how long it took, how
// much memory the dry run held, and how large the approval's record is.
func TestMeasureLookBack(t *testing.T) {
	n := 2000
	if v, err := strconv.Atoi(os.Getenv("RHEA_MEASURE_N")); err == nil && v > 0 {
		n = v
	}
	w := newCITWorld(t)
	ctx, s, x := w.ctx, w.s, w.x
	x.LockWait = 30 * time.Second // as rhea serve waits for the books
	report := func(format string, a ...any) { t.Logf(format, a...); fmt.Printf(format+"\n", a...) }

	// peak heap over a step, sampled
	measure := func(step string, run func()) time.Duration {
		var peak atomic.Uint64
		done := make(chan struct{})
		go func() {
			var m runtime.MemStats
			for {
				runtime.ReadMemStats(&m)
				if m.HeapAlloc > peak.Load() {
					peak.Store(m.HeapAlloc)
				}
				select {
				case <-done:
					return
				case <-time.After(50 * time.Millisecond):
				}
			}
		}()
		runtime.GC()
		t0 := time.Now()
		run()
		d := time.Since(t0)
		close(done)
		report("%-58s %10s   peak heap %5d MB", step, d.Round(time.Millisecond), peak.Load()>>20)
		return d
	}

	sales, purchases := n*4/5, n-n*4/5
	buyers, suppliers := max(1, sales/20), max(1, purchases/60)
	nip := func(prefix, i int) string { return fmt.Sprintf("%d%09d", prefix, i) }
	date := func(i int) string {
		d, _ := time.Parse("2006-01-02", "2026-01-02")
		return d.AddDate(0, 0, i%285).Format("2006-01-02")
	}
	amounts := func(i int) (string, string, string) {
		net := int64(10000 + (i%900)*100)
		vat := net * 23 / 100
		return core.FormatMoney(net), core.FormatMoney(vat), core.FormatMoney(net + vat)
	}
	report("look-back at %d invoices: %d sales to %d buyers, %d purchases from %d suppliers", n, sales, buyers, purchases, suppliers)

	measure(fmt.Sprintf("register %d firms", buyers+suppliers), func() {
		for i := 0; i < buyers; i++ {
			w.submit("company.registered", "2026-01-01", "buyer-"+strconv.Itoa(i), map[string]any{"name": fmt.Sprintf("Nabywca %05d sp. z o.o.", i), "kind": "customer", "vat_id": nip(7, i), "country": "PL"})
		}
		for i := 0; i < suppliers; i++ {
			w.submit("company.registered", "2026-01-01", "supplier-"+strconv.Itoa(i), map[string]any{"name": fmt.Sprintf("Dostawca %05d sp. z o.o.", i), "kind": "supplier", "vat_id": nip(8, i), "country": "PL"})
		}
	})
	t0 := time.Now()
	measure(fmt.Sprintf("book %d invoices, one by one", n), func() {
		sold, bought := 0, 0
		for i := 0; i < n; i++ {
			net, vat, gross := amounts(i)
			if i%5 != 4 {
				w.submit("sales.invoice.issued", date(i), "fv-"+strconv.Itoa(i), map[string]any{"number": fmt.Sprintf("FV %d/2026", i), "issue_date": date(i), "market": "pl", "intragroup": "no",
					"seller_nip": "5250001111", "buyer_nip": nip(7, sold%buyers), "net": net, "vat": vat, "gross": gross, "vat_rate": "23", "currency": "PLN"})
				sold++
			} else {
				w.submit("purchase.invoice.received", date(i), "fz-"+strconv.Itoa(i), map[string]any{"number": fmt.Sprintf("FZ %d/2026", i), "invoice_date": date(i), "market": "pl",
					"supplier_nip": nip(8, bought%suppliers), "buyer_nip": "5250001111", "net": net, "vat": vat, "gross": gross, "vat_rate": "23", "currency": "PLN"})
				bought++
			}
			if (i+1)%max(1, n/10) == 0 {
				report("  %7d booked, %6.1f invoices a second", i+1, float64(i+1)/time.Since(t0).Seconds())
			}
		}
	})

	measure("Klienci and Dostawcy arrive (the core's version approved)", func() {
		sums, err := pack.LoadCore(ctx, s, "init")
		if err != nil {
			t.Fatal(err)
		}
		for _, sum := range sums {
			if sum.Bundle == "" {
				continue
			}
			if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", *w.day); err != nil || len(errs) != 0 {
				t.Fatalf("approve %s: %v %v", sum.Bundle, err, errs)
			}
		}
	})

	lookBack := func(what string, records, entries int) {
		var plan exec.BackfillPlan
		measure("the dry run: "+what, func() {
			var err error
			if plan, err = x.PlanBackfill(ctx, *w.day); err != nil {
				t.Fatal(err)
			}
		})
		added := 0
		for _, r := range plan.Summary.Records {
			added += r.Added
		}
		report("  %d documents to look at again, %d records to add (%v), %d months of entries, %d refused", len(plan.Chains), added, plan.Summary.Records, len(plan.Summary.Entries), len(plan.Errors))
		if added != records || len(plan.Summary.Entries) != entries {
			t.Fatalf("%s: %d records, %d months of entries; want %d, %d", what, added, len(plan.Summary.Entries), records, entries)
		}
		// the approval, and a door knocking two seconds in
		var tr exec.Triggered
		knock := make(chan string, 1)
		go func() {
			time.Sleep(2 * time.Second)
			t1 := time.Now()
			_, err := x.TriggerActivity(ctx, "submit_event", map[string]any{"event_type": "company.registered", "occurred_at": *w.day,
				"payload": map[string]any{"name": "Pukający " + what, "kind": "customer", "vat_id": nip(9, len(what)), "country": "PL"}}, "test", "knock-"+what)
			knock <- fmt.Sprintf("waited %s: %v", time.Since(t1).Round(time.Millisecond), err)
		}()
		measure("the approval: "+what, func() {
			var err error
			if tr, err = x.TriggerActivity(ctx, "approve_backfill", map[string]any{"plan": plan.Fingerprint, "approved_by": "krzysztof", "occurred_at": *w.day}, "krzysztof", ""); err != nil {
				t.Fatal(err)
			}
		})
		ev, err := s.GetEvent(ctx, tr.EventID)
		if err != nil {
			t.Fatal(err)
		}
		report("  %d documents explained further; the approval's record %d KB", tr.Backfilled, len(ev.Payload)>>10)
		report("  a door knocking 2 s in: %s", <-knock)
		measure("the dry run again, nothing left (Dziś on every load)", func() {
			again, err := x.PlanBackfill(ctx, *w.day)
			if err != nil || len(again.Chains) != 0 {
				t.Fatalf("again: %d chains, %v", len(again.Chains), err)
			}
		})
	}
	lookBack("Klienci and Dostawcy", buyers+suppliers, 0)

	// the heavy case: a rule that adds a record to every sales invoice
	if err := s.InsertObjectType(ctx, core.ObjectType{Name: "measure_mark", Version: 1, Domain: "sales",
		Fields: []core.FieldDef{{Name: "number", Type: "string", Required: true}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: "measure-mark", Status: core.StatusActive, Priority: 900, EffectiveFrom: "2026-01-01",
		CreatedBy: "test", Description: "a mark on every sales invoice", Spec: core.RuleSpec{
			Match:  core.Match{EventType: core.EventObjectMaterialized, Where: []core.Condition{{Path: "$.object_type", Op: "eq", Value: "sales_invoice"}}},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "measure_mark", Fields: map[string]string{"number": "=$.state.number"}}}}}); err != nil {
		t.Fatal(err)
	}
	lookBack("a record on every sales invoice", sales, 0)
}
