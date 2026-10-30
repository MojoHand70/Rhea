package shell_test

import (
	"fmt"
	"reflect"
	"testing"

	"rhea/internal/pack"
)

// TestClientsAndSuppliersAreRolesOfOneFirm (KK 2026-10-27, option A): a firm
// is one record; our client and our supplier are records of their own
// pointing at it, each with its own terms. A firm's first sales invoice
// makes it a client from that invoice's date and its later invoices leave
// the client as it is; its first purchase invoice makes it a supplier. A
// firm that both buys from us and sells to us is one firm with both roles.
// The terms are set through a door each, what a setting leaves out staying
// as it was; replay reproduces it all.
func TestClientsAndSuppliersAreRolesOfOneFirm(t *testing.T) {
	w := newCITWorld(t)
	ctx, s, x := w.ctx, w.s, w.x
	sums, err := pack.LoadCore(ctx, s, "init")
	if err != nil {
		t.Fatal(err)
	}
	for _, sum := range sums {
		if sum.Bundle == "" {
			continue
		}
		if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-09-01"); err != nil || len(errs) != 0 {
			t.Fatalf("approve %s: %v %v", sum.Bundle, err, errs)
		}
	}
	buyer, restaurant := w.id["Nabywca S.A."], w.id["Restauracja Pod Lipą"]
	roles := func(typ string) map[string]map[string]any {
		t.Helper()
		objs, err := s.ObjectsByType(ctx, typ)
		if err != nil {
			t.Fatal(err)
		}
		out := map[string]map[string]any{}
		for _, o := range objs {
			firm := fmt.Sprint(o.State["company"])
			if _, twice := out[firm]; twice {
				t.Fatalf("two %s records for one firm %s", typ, firm)
			}
			st := map[string]any{"id": o.ID}
			for k, v := range o.State {
				st[k] = v
			}
			out[firm] = st
		}
		return out
	}

	w.sale("fv-1", "FV 1/09/2026", "2026-09-21", "1000.00", "230.00", "1230.00")
	w.sale("fv-2", "FV 2/09/2026", "2026-09-24", "500.00", "115.00", "615.00")
	clients := roles("client")
	if len(clients) != 1 || clients[buyer] == nil || clients[buyer]["since"] != "2026-09-21" {
		t.Fatalf("two invoices to one firm make one client, from the first: %v", clients)
	}
	if len(roles("supplier")) != 0 {
		t.Fatalf("selling makes no supplier: %v", roles("supplier"))
	}

	w.buy("fz-1", "FZ 7/09/2026", "2026-09-25", "5260001246", "200.00", "46.00", "246.00")
	w.buy("fz-2", "R 12/2026", "2026-09-26", "5270003330", "100.00", "8.00", "108.00")
	suppliers := roles("supplier")
	if len(suppliers) != 2 || suppliers[buyer] == nil || suppliers[buyer]["since"] != "2026-09-25" || suppliers[restaurant] == nil {
		t.Fatalf("each firm we buy from is a supplier, from its first invoice: %v", suppliers)
	}
	if clients = roles("client"); len(clients) != 1 || clients[buyer]["since"] != "2026-09-21" {
		t.Fatalf("the firm that also sells to us stays one client: %v", clients)
	}

	cl, sp := clients[buyer]["id"], suppliers[buyer]["id"]
	w.door("set_client_terms", map[string]any{"client": cl, "payment_term_days": 14, "discount_percent": "5", "occurred_at": "2026-09-27"})
	w.door("set_client_terms", map[string]any{"client": cl, "credit_limit": "20000.00", "occurred_at": "2026-09-28"})
	w.door("set_supplier_terms", map[string]any{"supplier": sp, "payment_term_days": 30, "needs_approval": "yes", "occurred_at": "2026-09-28"})
	c, sup := roles("client")[buyer], roles("supplier")[buyer]
	if fmt.Sprint(c["payment_term_days"]) != "14" || fmt.Sprint(c["discount_percent"]) != "5" || c["credit_limit"] == nil {
		t.Fatalf("the client's terms, a later setting leaving the earlier as they were: %v", c)
	}
	if fmt.Sprint(sup["payment_term_days"]) != "30" || sup["needs_approval"] != "yes" || sup["discount_percent"] != nil {
		t.Fatalf("the same firm as our supplier keeps terms of its own: %v", sup)
	}

	before, _ := s.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := s.AllObjects(ctx)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("replay does not reproduce the clients and suppliers")
	}
}

// TestTheLookBackFindsClientsAndSuppliersInPastInvoices (KK 2026-10-27): a
// company already has its invoices when Klienci and Dostawcy arrive; asked to
// look back, Rhea says what she would add — one client, two suppliers, no
// entries in the books — and once a person approves exactly that, the roles
// are there, dated by each firm's first invoice. Looking back again finds
// nothing; replay reproduces it.
func TestTheLookBackFindsClientsAndSuppliersInPastInvoices(t *testing.T) {
	w := newCITWorld(t)
	ctx, s, x := w.ctx, w.s, w.x
	w.sale("fv-1", "FV 1/09/2026", "2026-09-21", "1000.00", "230.00", "1230.00")
	w.sale("fv-2", "FV 2/09/2026", "2026-09-24", "500.00", "115.00", "615.00")
	w.buy("fz-1", "FZ 7/09/2026", "2026-09-25", "5260001246", "200.00", "46.00", "246.00")
	w.buy("fz-2", "R 12/2026", "2026-09-26", "5270003330", "100.00", "8.00", "108.00")

	// Klienci and Dostawcy arrive with the core's version: nothing yet.
	sums, err := pack.LoadCore(ctx, s, "init")
	if err != nil {
		t.Fatal(err)
	}
	for _, sum := range sums {
		if sum.Bundle == "" {
			continue
		}
		if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-10-17"); err != nil || len(errs) != 0 {
			t.Fatalf("approve %s: %v %v", sum.Bundle, err, errs)
		}
	}
	if clients, _ := s.ObjectsByType(ctx, "client"); len(clients) != 0 {
		t.Fatalf("clients before anyone looked back: %d", len(clients))
	}

	var plan struct {
		Plan    string `json:"plan"`
		Summary struct {
			Records []map[string]any `json:"records"`
			Entries []map[string]any `json:"entries"`
		} `json:"summary"`
	}
	w.post("/api/backfill/simulate", nil, &plan)
	if got := fmt.Sprint(plan.Summary.Records); got != "[map[added:1 type:client] map[added:2 type:supplier]]" || len(plan.Summary.Entries) != 0 {
		t.Fatalf("what the look-back would do: %s, entries %v", got, plan.Summary.Entries)
	}
	tr, err := x.TriggerActivity(ctx, "approve_backfill", map[string]any{"plan": plan.Plan, "approved_by": "krzysztof", "occurred_at": "2026-10-17"}, "krzysztof", "")
	if err != nil || tr.Backfilled != 3 {
		t.Fatalf("approve the look-back: %+v %v", tr, err)
	}
	buyer, restaurant := w.id["Nabywca S.A."], w.id["Restauracja Pod Lipą"]
	since := func(typ string) map[string]string {
		objs, _ := s.ObjectsByType(ctx, typ)
		out := map[string]string{}
		for _, o := range objs {
			out[fmt.Sprint(o.State["company"])] = fmt.Sprint(o.State["since"])
		}
		return out
	}
	if c, sp := since("client"), since("supplier"); len(c) != 1 || c[buyer] != "2026-09-21" || len(sp) != 2 || sp[buyer] != "2026-09-25" || sp[restaurant] != "2026-09-26" {
		t.Fatalf("the roles, dated by each firm's first invoice: clients %v, suppliers %v", c, sp)
	}

	w.post("/api/backfill/simulate", nil, &plan)
	if len(plan.Summary.Records) != 0 || len(plan.Summary.Entries) != 0 {
		t.Fatalf("looked back twice: %+v", plan.Summary)
	}
	before, _ := s.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := s.AllObjects(ctx)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("replay does not reproduce what the look-back filled")
	}
}

// TestAClientIsAddedByHand (KK 2026-10-30: "we need to be able to add
// companies by NIP"): a firm becomes our client by hand before any invoice —
// a contract signed, its terms agreed — from the day the person names, with
// the terms given then; its first invoice later finds the client and adds
// none. Added again, the firm stays one client and the terms given update
// it. A supplier likewise.
func TestAClientIsAddedByHand(t *testing.T) {
	w := newCITWorld(t)
	ctx, s, x := w.ctx, w.s, w.x
	sums, err := pack.LoadCore(ctx, s, "init")
	if err != nil {
		t.Fatal(err)
	}
	for _, sum := range sums {
		if sum.Bundle == "" {
			continue
		}
		if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-09-01"); err != nil || len(errs) != 0 {
			t.Fatalf("approve %s: %v %v", sum.Bundle, err, errs)
		}
	}
	buyer, restaurant := w.id["Nabywca S.A."], w.id["Restauracja Pod Lipą"]
	role := func(typ, firm string) []map[string]any {
		t.Helper()
		objs, err := s.ObjectsByType(ctx, typ)
		if err != nil {
			t.Fatal(err)
		}
		var out []map[string]any
		for _, o := range objs {
			if o.State["company"] == firm {
				out = append(out, o.State)
			}
		}
		return out
	}
	w.door("add_client", map[string]any{"company": buyer, "since": "2026-09-01", "payment_term_days": 14, "occurred_at": "2026-09-01"})
	if c := role("client", buyer); len(c) != 1 || c[0]["since"] != "2026-09-01" || fmt.Sprint(c[0]["payment_term_days"]) != "14" {
		t.Fatalf("a client added by hand, from the day named, with its terms: %v", c)
	}
	w.sale("fv-1", "FV 1/09/2026", "2026-09-21", "1000.00", "230.00", "1230.00")
	if c := role("client", buyer); len(c) != 1 || c[0]["since"] != "2026-09-01" {
		t.Fatalf("its first invoice finds the client and adds none: %v", c)
	}
	w.door("add_client", map[string]any{"company": buyer, "since": "2026-10-01", "credit_limit": "5000.00", "occurred_at": "2026-10-01"})
	if c := role("client", buyer); len(c) != 1 || c[0]["since"] != "2026-09-01" || fmt.Sprint(c[0]["payment_term_days"]) != "14" || c[0]["credit_limit"] == nil {
		t.Fatalf("added again, one client, the terms given update it: %v", c)
	}
	w.door("add_supplier", map[string]any{"company": restaurant, "since": "2026-09-15", "needs_approval": "yes", "occurred_at": "2026-09-15"})
	w.buy("fz-1", "R 12/2026", "2026-09-26", "5270003330", "100.00", "8.00", "108.00")
	if sp := role("supplier", restaurant); len(sp) != 1 || sp[0]["since"] != "2026-09-15" || sp[0]["needs_approval"] != "yes" {
		t.Fatalf("a supplier added by hand: %v", sp)
	}

	before, _ := s.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := s.AllObjects(ctx)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("replay does not reproduce the clients added by hand")
	}
}
