package shell_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"rhea/internal/adapter"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/shell"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

// citWorld is the Poland pack installed, the seller, its buyer and a
// restaurant registered, with the doors and documents the CIT tests use.
type citWorld struct {
	ctx    context.Context
	s      *store.Store
	x      *exec.Executor
	srv    *shell.Server
	day    *string
	get    func(path string, out any)
	door   func(name string, in map[string]any)
	submit func(typ, date, key string, payload map[string]any)
	id     map[string]string
	self   string

	sale       func(key, no, date, net, vat, gross string)
	buy        func(key, no, date, nip, net, vat, gross string)
	closeMonth func(end, on string)
	advance    func(period string) map[string]any
	want       func(period string, fields map[string]string)
}

func newCITWorld(t *testing.T) *citWorld {
	ctx := context.Background()
	s := storetest.New(t)
	for _, f := range []string{"finance.json", "finance_v2.json", "finance_v3.json", "finance_v4.json",
		"finance_v5.json", "finance_v6.json", "finance_v7.json", "finance_v8.json"} {
		seedFromFile(t, s, f)
	}
	x := &exec.Executor{Store: s}
	day := "2026-10-17"
	srv := &shell.Server{Store: s, Exec: x, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb"),
		Clock: adapter.Clock{Now: func() time.Time { d, _ := time.Parse("2006-01-02", day); return d.Add(9 * time.Hour) }}}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	get := func(path string, out any) {
		t.Helper()
		res, err := ts.Client().Get(ts.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		var b bytes.Buffer
		b.ReadFrom(res.Body)
		if res.StatusCode != 200 {
			t.Fatalf("GET %s: %d %s", path, res.StatusCode, b.String())
		}
		if err := json.Unmarshal(b.Bytes(), out); err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
	}
	for _, p := range []string{"common/people.json", "common/periods.json", "common/documents.json", "pl/pack.json"} {
		sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", p), "test")
		if err != nil {
			t.Fatal(err)
		}
		if _, _, _, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-09-01"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.InsertRuleVersion(ctx, core.Rule{
		ID: "register-company", Status: core.StatusDraft, Priority: 10, EffectiveFrom: "2026-01-01", CreatedBy: "human",
		Description: "register-company",
		Spec: core.RuleSpec{Match: core.Match{EventType: "company.registered"}, Effect: core.Effect{Object: core.ObjectTemplate{Type: "company",
			Fields: map[string]string{"name": "=$.name", "kind": "=$.kind", "vat_id": "=$.vat_id", "country": "=$.country"}}}},
	}); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := x.ApproveRule(ctx, "register-company", "krzysztof", "2026-09-01"); err != nil {
		t.Fatal(err)
	}
	door := func(name string, in map[string]any) {
		t.Helper()
		tr, err := x.TriggerActivity(ctx, name, in, "krzysztof", "")
		if err != nil || tr.Booked != 1 || len(tr.Errors) > 0 {
			t.Fatalf("%s: %+v %v", name, tr, err)
		}
	}
	submit := func(typ, date, key string, payload map[string]any) {
		t.Helper()
		tr, err := x.TriggerActivity(ctx, "submit_event", map[string]any{"event_type": typ, "occurred_at": date, "payload": payload}, "test", key)
		if err != nil || tr.Booked != 1 || len(tr.Errors) > 0 {
			t.Fatalf("%s: %+v %v", key, tr, err)
		}
	}
	submit("company.registered", "2026-03-01", "self", map[string]any{"name": "Sprzedawca Sp. z o.o.", "kind": "self", "vat_id": "5250001111", "country": "PL"})
	submit("company.registered", "2026-03-01", "buyer", map[string]any{"name": "Nabywca S.A.", "kind": "customer", "vat_id": "5260001246", "country": "PL"})
	submit("company.registered", "2026-03-01", "restaurant", map[string]any{"name": "Restauracja Pod Lipą", "kind": "supplier", "vat_id": "5270003330", "country": "PL"})
	companies, _ := s.ObjectsByType(ctx, "company")
	id := map[string]string{}
	for _, c := range companies {
		id[c.State["name"].(string)] = c.ID
	}
	self := id["Sprzedawca Sp. z o.o."]
	sale := func(key, no, date, net, vat, gross string) {
		t.Helper()
		submit("sales.invoice.issued", date, key, map[string]any{"number": no, "issue_date": date, "market": "pl", "intragroup": "no",
			"seller_nip": "5250001111", "buyer_nip": "5260001246", "net": net, "vat": vat, "gross": gross, "vat_rate": "23", "currency": "PLN"})
	}
	buy := func(key, no, date, nip, net, vat, gross string) {
		t.Helper()
		submit("purchase.invoice.received", date, key, map[string]any{"number": no, "invoice_date": date, "market": "pl",
			"supplier_nip": nip, "buyer_nip": "5250001111", "net": net, "vat": vat, "gross": gross, "vat_rate": "23", "currency": "PLN"})
	}
	closeMonth := func(end, on string) {
		t.Helper()
		door("close_period", map[string]any{"period": end, "book": "pl-stat", "occurred_at": on})
	}
	advance := func(period string) map[string]any {
		t.Helper()
		all, _ := s.ObjectsByType(ctx, "cit_advance")
		for _, a := range all {
			if a.State["period"] == period {
				if a.RuleID != "pl-cit-advance" {
					t.Fatalf("%s explained by %s", period, a.RuleID)
				}
				return a.State
			}
		}
		t.Fatalf("no advance for %s among %d", period, len(all))
		return nil
	}
	want := func(period string, fields map[string]string) {
		t.Helper()
		a := advance(period)
		for f, v := range fields {
			got := fmt.Sprint(a[f])
			if n, ok := a[f].(float64); ok {
				got = fmt.Sprint(int64(n))
			}
			if typ := map[string]bool{"revenue": true, "not_taxable": true, "costs": true, "non_deductible": true, "income": true,
				"base": true, "tax": true, "advances_before": true, "advance": true}; typ[f] {
				m, err := core.ParseMoney(v)
				if err != nil {
					t.Fatal(err)
				}
				v = fmt.Sprint(m)
			}
			if got != v {
				t.Fatalf("%s %s = %s, want %s (%v)", period, f, got, v, a)
			}
		}
	}

	return &citWorld{ctx: ctx, s: s, x: x, srv: srv, day: &day, get: get, door: door, submit: submit, id: id, self: self,
		sale: sale, buy: buy, closeMonth: closeMonth, advance: advance, want: want}
}

// TestCITAdvance (KK 2026-10-17: "CIT, KUP and NKUP are very important"):
// the Poland pack computes the CIT advance when a month closes, from
// January to the month's end — revenue less tax-deductible costs, the
// accounts saying what they are for income tax (entertainment is not a tax
// cost); the rate 9% while the year's status allows it and the year's
// revenue stays within the limit, else 19%; the advance is the year's tax
// less the advances before it, never below zero, booked Wn 870 / Ma 225.
// The result before tax does not see the charge, and replay reproduces it.
func TestCITAdvance(t *testing.T) {
	w := newCITWorld(t)
	ctx, s, x, get, door, id, self := w.ctx, w.s, w.x, w.get, w.door, w.id, w.self
	sale, buy, closeMonth, advance, want := w.sale, w.buy, w.closeMonth, w.advance, w.want
	day := w.day
	door("declare_company_form", map[string]any{"company": self, "vat_status": "active", "books": "full", "income_tax": "cit",
		"from": "2026-03-01", "occurred_at": "2026-03-01"})
	// The restaurant's invoices are entertainment: not a tax cost.
	door("set_cost_account", map[string]any{"company": id["Restauracja Pod Lipą"], "account": "409-1", "occurred_at": "2026-03-01"})

	// The year's status: first said wrong (a transformation — no 9% this
	// year or next), then corrected; the newer declaration stands.
	door("declare_cit_status", map[string]any{"company": self, "year_start": "2026-01-01", "started_on": "2026-03-01",
		"formed_by": "transformation", "occurred_at": "2026-03-01"})
	door("declare_cit_status", map[string]any{"company": self, "year_start": "2026-01-01", "started_on": "2026-03-01",
		"formed_by": "new", "occurred_at": "2026-03-02"})
	statuses, _ := s.ObjectsByType(ctx, "cit_status")
	if len(statuses) != 2 || statuses[0].State["reduced"] != "no" || statuses[1].State["reduced"] != "yes" {
		t.Fatalf("statuses: %+v", statuses)
	}

	// September: 1500.00 sold; 400.00 of services and 200.00 of
	// entertainment bought. Income 1500 − (600 − 200) = 1100; 9% = 99.
	sale("fv-1", "FV 1/09/2026", "2026-09-21", "1000.00", "230.00", "1230.00")
	sale("fv-2", "FV 2/09/2026", "2026-09-24", "500.00", "115.00", "615.00")
	buy("fz-1", "FZ/9/2026", "2026-09-25", "5260001246", "400.00", "92.00", "492.00")
	buy("fz-2", "R/9/2026", "2026-09-26", "5270003330", "200.00", "46.00", "246.00")
	closeMonth("2026-09-30", "2026-10-05")
	want("2026-09", map[string]string{"revenue": "1500.00", "costs": "600.00", "non_deductible": "200.00", "income": "1100.00",
		"base": "1100.00", "rate": "9", "tax": "99.00", "advances_before": "0.00", "advance": "99.00"})

	// October: 2000.00 more sold. The year's tax 3100 × 9% = 279, less 99.
	sale("fv-3", "FV 1/10/2026", "2026-10-08", "2000.00", "460.00", "2460.00")
	closeMonth("2026-10-31", "2026-11-04")
	want("2026-10", map[string]string{"revenue": "3500.00", "costs": "600.00", "income": "3100.00", "rate": "9",
		"tax": "279.00", "advances_before": "99.00", "advance": "180.00"})

	// The deadline for October's advance, on the morning of 17 November.
	*day = "2026-11-17"
	type rows struct {
		Rows [][]cell `json:"rows"`
	}
	var cal rows
	get("/api/views/kalendarz-podatkowy", &cal)
	if r := cellVals(cal.Rows); len(r) != 2 || r[0][0] != "2026-11-20" || r[0][1] != "Zaliczka na CIT" || r[0][3] != "180.00" {
		t.Fatalf("the deadlines on %s: %v", *day, r)
	}

	// November: a large sale takes the year's revenue past 8 431 000 zł —
	// from this month the whole year is taxed at 19%.
	sale("fv-4", "FV 1/11/2026", "2026-11-10", "8500000.00", "1955000.00", "10455000.00")
	closeMonth("2026-11-30", "2026-12-03")
	want("2026-11", map[string]string{"revenue": "8503500.00", "income": "8503100.00", "rate": "19",
		"tax": "1615589.00", "advances_before": "279.00", "advance": "1615310.00"})
	if b := advance("2026-11")["rate_basis"].(string); b[:4] != "19%:" {
		t.Fatalf("the rate's basis: %s", b)
	}

	// December: a cost larger than the year's income — a loss so far; the
	// advance is nothing, not a refund, and nothing is booked.
	buy("fz-3", "FZ/12/2026", "2026-12-10", "5260001246", "8600000.00", "1978000.00", "10578000.00")
	closeMonth("2026-12-31", "2027-01-05")
	want("2026-12", map[string]string{"income": "-96900.00", "base": "0.00", "tax": "0.00", "advances_before": "1615589.00", "advance": "0.00"})

	// The charge is booked Wn 870 / Ma 225, three times — the year's
	// advances; the result before tax never sees it, and it counts the
	// entertainment the tax adds back: -96 900 for tax, -97 100 in the books.
	postings, _ := s.ObjectsByType(ctx, "posting")
	accounts, _ := s.ObjectsByType(ctx, "account")
	code := map[string]string{}
	for _, a := range accounts {
		code[a.ID] = a.State["code"].(string)
	}
	var on870, on225 int64
	n870 := 0
	for _, p := range postings {
		amt, _ := core.ParseMoney(fmt.Sprint(p.State["amount"]))
		if m, ok := p.State["amount"].(float64); ok {
			amt = int64(m)
		}
		switch code[fmt.Sprint(p.State["account"])] {
		case "870":
			n870++
			if p.State["side"] == "debit" {
				on870 += amt
			}
		case "225":
			if p.State["side"] == "credit" {
				on225 += amt
			}
		}
	}
	if n870 != 3 || on870 != 161558900 || on225 != 161558900 {
		t.Fatalf("870: %d postings, %d; 225: %d", n870, on870, on225)
	}
	var wynik rows
	get("/api/views/wynik", &wynik)
	if wv := cellVals(wynik.Rows); wv[2][2] != "-97100.00" {
		t.Fatalf("the result before tax: %v", wv)
	}

	// The statement that opens an advance: income from January, the tax, the
	// advances before, the advance.
	all, _ := s.ObjectsByType(ctx, "cit_advance")
	var st struct {
		Sections []struct {
			Rows []struct{ Op, Field, V string } `json:"rows"`
		} `json:"sections"`
	}
	get("/api/views/zaliczka-cit?object_id="+all[0].ID, &st)
	if len(st.Sections) != 2 {
		t.Fatalf("the statement: %+v", st)
	}
	var lines []string
	for _, sec := range st.Sections {
		for _, r := range sec.Rows {
			lines = append(lines, r.Op+r.Field+" "+r.V)
		}
	}
	if fmt.Sprint(lines) != "[revenue 1500.00 -costs 600.00 +non_deductible 200.00 =income 1100.00 base 1100.00 tax 99.00 -advances_before 0.00 =advance 99.00]" {
		t.Fatalf("the statement: %v", lines)
	}

	// An account says what it is for income tax through the door; a cost
	// cannot be called a revenue — the declaration waits.
	door("set_tax_treatment", map[string]any{"account": "403", "tax_treatment": "non_deductible", "occurred_at": "2027-01-06"})
	if tr, err := x.TriggerActivity(ctx, "set_tax_treatment", map[string]any{"account": "402", "tax_treatment": "taxable", "occurred_at": "2027-01-06"}, "krzysztof", ""); err != nil || tr.Booked != 0 {
		t.Fatalf("a cost called a revenue was booked: %+v %v", tr, err)
	}
	accounts, _ = s.ObjectsByType(ctx, "account")
	for _, a := range accounts {
		if c := a.State["code"]; (c == "403" && a.State["tax_treatment"] != "non_deductible") || (c == "402" && a.State["tax_treatment"] != nil) ||
			(c == "409-1" && a.State["tax_treatment"] != "non_deductible") {
			t.Fatalf("account %s: %v", c, a.State)
		}
	}

	before, _ := s.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := s.AllObjects(ctx)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("replay does not reproduce the advances")
	}
}

// TestCITAdvanceBackfilled (KK's Helios, 2026-10-18): September was closed
// before the company said it pays CIT, October after — October's advance
// took the whole year's tax, there being no advance before it. Ruled
// backfill gives September its advance after the fact (planned in log
// order, each chain reading what the chains before it leave), and
// recomputing October against the advances now in the books leaves it its
// own part, the difference booked back on the day: Wn 225 / Ma 870.
func TestCITAdvanceBackfilled(t *testing.T) {
	w := newCITWorld(t)
	ctx, x, door := w.ctx, w.x, w.door
	w.sale("fv-1", "FV 1/09/2026", "2026-09-21", "1000.00", "230.00", "1230.00")
	w.sale("fv-2", "FV 2/09/2026", "2026-09-24", "500.00", "115.00", "615.00")
	w.buy("fz-1", "FZ/9/2026", "2026-09-25", "5260001246", "400.00", "92.00", "492.00")
	w.closeMonth("2026-09-30", "2026-10-05")
	if all, _ := w.s.ObjectsByType(ctx, "cit_advance"); len(all) != 0 {
		t.Fatalf("advances before the company pays CIT: %d", len(all))
	}
	door("declare_company_form", map[string]any{"company": w.self, "vat_status": "active", "books": "full", "income_tax": "cit",
		"from": "2026-01-01", "occurred_at": "2026-10-06"})
	door("declare_cit_status", map[string]any{"company": w.self, "year_start": "2026-01-01", "started_on": "2026-03-01",
		"formed_by": "new", "occurred_at": "2026-10-06"})
	w.sale("fv-3", "FV 1/10/2026", "2026-10-08", "2000.00", "460.00", "2460.00")
	w.closeMonth("2026-10-31", "2026-11-04")
	w.want("2026-10", map[string]string{"income": "3100.00", "tax": "279.00", "advances_before": "0.00", "advance": "279.00"})

	plan, err := x.PlanBackfill(ctx, "2026-11-10")
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Chains) != 1 || len(plan.Errors) != 0 || plan.Chains[0].EventType != "period.closed" ||
		fmt.Sprint(plan.Chains[0].Rules) != "[pl-cit-advance pl-post-cit-advance]" || plan.Chains[0].Forwarded == "" || plan.Chains[0].BookedAt != "2026-11-10" {
		t.Fatalf("the plan: %+v", plan)
	}
	if n, errs, err := x.ApproveBackfill(ctx, "krzysztof", "2026-11-10"); err != nil || len(errs) != 0 || n != 1 {
		t.Fatalf("backfill: %d %v %v", n, errs, err)
	}
	w.want("2026-09", map[string]string{"income": "1100.00", "rate": "9", "tax": "99.00", "advances_before": "0.00", "advance": "99.00"})

	door("recompute_cit_advance", map[string]any{"period": "2026-10", "occurred_at": "2026-11-10"})
	w.want("2026-10", map[string]string{"tax": "279.00", "advances_before": "99.00", "advance": "180.00"})
	// Recomputed again, nothing moves: no difference, nothing booked.
	door("recompute_cit_advance", map[string]any{"period": "2026-10", "occurred_at": "2026-11-11"})
	w.want("2026-10", map[string]string{"advances_before": "99.00", "advance": "180.00"})

	// On 870: 279 at October's close, 99 for September after the fact, 99
	// back on the recompute — the year's 279, as the advances now say.
	postings, _ := w.s.ObjectsByType(ctx, "posting")
	accounts, _ := w.s.ObjectsByType(ctx, "account")
	code := map[string]string{}
	for _, a := range accounts {
		code[a.ID] = a.State["code"].(string)
	}
	var on870 int64
	lines := map[string]int{}
	for _, p := range postings {
		if code[fmt.Sprint(p.State["account"])] != "870" {
			continue
		}
		amt := int64(p.State["amount"].(float64))
		if p.State["side"] == "credit" {
			amt = -amt
		}
		on870 += amt
		lines[p.RuleID]++
		if d, _ := p.State["date"].(string); d < "2026-11-01" {
			t.Fatalf("a charge for a closed month dated %v", d)
		}
	}
	if on870 != 27900 || lines["pl-post-cit-advance"] != 2 || lines["pl-post-cit-advance-correction"] != 1 {
		t.Fatalf("870: %d %v", on870, lines)
	}
}
