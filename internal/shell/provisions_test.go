package shell_test

import (
	"fmt"
	"testing"
)

// TestProvisionsAndWriteDowns (the auditor's exam, 2026-10-19; EXAM.md
// 6.1.2 d, 6.1.3 d): a provision for a liability charges other operating
// costs against 830 when created, shrinks when the liability arrives (the
// invoice's cost comes off, it was in the provision), and what is left goes
// to other operating revenues when the risk ends; an accrual for warranty
// repairs sits on 641 and, released, reduces the cost it charged. A
// receivable written down for a debtor's insolvency is a tax cost on 761,
// one written down for being overdue is not (761-3); a write-down ends
// either with the debtor paying (back to 760) or with the receivable
// leaving the books against the write-down.
func TestProvisionsAndWriteDowns(t *testing.T) {
	w := newCITWorld(t)
	ctx, s, door := w.ctx, w.s, w.door
	door("declare_company_form", map[string]any{"company": w.self, "vat_status": "active", "books": "full", "income_tax": "cit",
		"from": "2026-01-01", "occurred_at": "2026-08-01"})
	w.sale("fv-1", "FV 1/09/2026", "2026-09-21", "1000.00", "230.00", "1230.00")
	w.sale("fv-2", "FV 2/09/2026", "2026-09-24", "500.00", "115.00", "615.00")
	w.buy("srv", "SRV/2026", "2026-11-20", "5260001246", "6000.00", "1380.00", "7380.00")
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
	byTitle := func() map[string]map[string]any {
		t.Helper()
		all, _ := s.ObjectsByType(ctx, "provision")
		out := map[string]map[string]any{}
		for _, p := range all {
			out[fmt.Sprint(p.State["title"])] = p.State
			out[fmt.Sprint(p.State["title"])]["_id"] = p.ID
		}
		return out
	}

	// A lawsuit: 10 000 provided for on 15 November.
	door("create_provision", map[string]any{"kind": "liability", "title": "Sprawa sądowa", "amount": "10000.00", "date": "2026-11-15", "occurred_at": "2026-11-15"})
	// Warranty repairs: 5 000 accrued at year end against 403.
	door("create_provision", map[string]any{"kind": "accrual", "title": "Naprawy gwarancyjne", "amount": "5000.00", "date": "2026-12-31", "cost_account": "403", "occurred_at": "2026-12-31"})
	after := sums()
	if after["761 debit"] != 1000000 || after["830 credit"] != 1000000 || after["641 credit"] != 500000 {
		t.Fatalf("provisions created: %v", after)
	}
	pv := byTitle()
	if pv["Sprawa sądowa"]["cost_account"] != "761" || pv["Sprawa sądowa"]["status"] != "open" || pv["Naprawy gwarancyjne"]["cost_account"] != "403" {
		t.Fatalf("provisions: %v", pv)
	}

	// The judgment: 6 000 of the lawsuit's provision is used; the rest, 4 000, is released when the appeal is dropped.
	door("use_provision", map[string]any{"provision": pv["Sprawa sądowa"]["_id"], "amount": "6000.00", "date": "2027-02-10", "occurred_at": "2027-02-10"})
	pv = byTitle()
	if fmt.Sprint(pv["Sprawa sądowa"]["used"]) != "600000" || pv["Sprawa sądowa"]["status"] != "open" {
		t.Fatalf("after use: %v", pv["Sprawa sądowa"])
	}
	door("release_provision", map[string]any{"provision": pv["Sprawa sądowa"]["_id"], "amount": "4000.00", "date": "2027-03-01", "reason": "apelacja wycofana", "occurred_at": "2027-03-01"})
	pv = byTitle()
	if fmt.Sprint(pv["Sprawa sądowa"]["released"]) != "400000" || pv["Sprawa sądowa"]["status"] != "closed" {
		t.Fatalf("after release: %v", pv["Sprawa sądowa"])
	}
	after = sums()
	if after["830 debit"] != 1000000 || after["761 credit"] != 600000 || after["760 credit"] != 400000 {
		t.Fatalf("use and release: %v", after)
	}
	// The warranty accrual: 1 500 of repairs invoiced, the rest released against the cost.
	door("use_provision", map[string]any{"provision": pv["Naprawy gwarancyjne"]["_id"], "amount": "1500.00", "date": "2027-01-20", "occurred_at": "2027-01-20"})
	door("release_provision", map[string]any{"provision": pv["Naprawy gwarancyjne"]["_id"], "amount": "3500.00", "date": "2027-06-30", "occurred_at": "2027-06-30"})
	after = sums()
	if after["641 debit"] != 500000 || after["403 credit"] != 500000 || byTitle()["Naprawy gwarancyjne"]["status"] != "closed" {
		t.Fatalf("accrual used and released: %v", after)
	}

	// Write-downs: the first customer went bankrupt (a tax cost), the second is merely late (not).
	invoice := map[string]string{}
	docs, _ := s.ObjectsByType(ctx, "sales_invoice")
	for _, d := range docs {
		invoice[d.State["number"].(string)] = d.ID
	}
	door("write_down_receivable", map[string]any{"invoice": invoice["FV 1/09/2026"], "amount": "1230.00", "date": "2026-12-31", "reason": "insolvency", "occurred_at": "2026-12-31"})
	door("write_down_receivable", map[string]any{"invoice": invoice["FV 2/09/2026"], "amount": "615.00", "date": "2026-12-31", "reason": "overdue", "occurred_at": "2026-12-31"})
	after = sums()
	if after["761 debit"] != 1000000+123000 || after["761-3 debit"] != 61500 || after["280 credit"] != 184500 {
		t.Fatalf("write-downs: %v", after)
	}
	wds, _ := s.ObjectsByType(ctx, "receivable_write_down")
	byInvoice := map[string]string{}
	for _, x := range wds {
		byInvoice[fmt.Sprint(x.State["invoice"])] = x.ID
		if x.State["status"] != "open" {
			t.Fatalf("a new write-down is open: %v", x.State)
		}
	}
	// The bankrupt's receivable leaves the books; the late one pays and the write-down returns to revenue.
	door("settle_write_down", map[string]any{"write_down": byInvoice[invoice["FV 1/09/2026"]], "outcome": "written_off", "date": "2027-04-15", "occurred_at": "2027-04-15"})
	door("settle_write_down", map[string]any{"write_down": byInvoice[invoice["FV 2/09/2026"]], "outcome": "reversed", "date": "2027-02-01", "occurred_at": "2027-02-01"})
	after = sums()
	if after["280 debit"] != 184500 || after["201 credit"] != 123000 || after["760 credit"] != 400000+61500 {
		t.Fatalf("write-downs settled: %v", after)
	}
	wds, _ = s.ObjectsByType(ctx, "receivable_write_down")
	for _, x := range wds {
		if x.State["status"] == "open" || x.State["settled_on"] == nil {
			t.Fatalf("a settled write-down: %v", x.State)
		}
	}
}
