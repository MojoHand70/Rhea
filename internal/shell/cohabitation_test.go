package shell_test

// E2 of the enterprise-structure ladder (DIRECTION): two operating companies
// — Alfa under the PL pack, Beta under the DE pack — in one kernel over one
// log. The pure-data attempt failed three ways (DECISIONS 2026-10-04, git
// history): silent master-data capture, silent currency-as-market
// misrouting, and a globally ambiguous vat_rate "0". The fix is also pure
// data — pack v2s: shared-type rules carry a market binding, invoices carry
// their seller as a ref, posting rules name their book, resolution keys stay
// globally unique. No kernel change anywhere — E2 is an M2-style negative
// proof. Enterprise structure is refs, not tenancy: each market's rules
// explain only its company's events, each company's ledger closes on its
// own (book, month), the statutory registers split by the seller's country,
// an event naming no market waits for a human instead of guessing, and one
// replay reproduces both companies exactly.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

func TestCohabitation(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	seedFromFile(t, s, "finance.json")
	seedFromFile(t, s, "finance_v2.json")
	seedFromFile(t, s, "finance_v3.json")
	seedFromFile(t, s, "finance_v4.json")
	seedFromFile(t, s, "finance_v5.json") // books: posting v2, per-book trial balance
	seedFromFile(t, s, "finance_v6.json") // sales_invoice v2: the seller is a ref
	seedFromFile(t, s, "finance_v7.json") // posting v3 (tx trace), fx_rate

	x := &exec.Executor{Store: s}
	srv := &shell.Server{
		Store: s, Exec: x,
		Agent:    &agent.Agent{Complete: func(context.Context, string, string) (string, error) { return "", fmt.Errorf("no agent in this test") }},
		DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb"),
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	get := func(path string, out any) {
		t.Helper()
		res, err := ts.Client().Get(ts.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			var e map[string]string
			json.NewDecoder(res.Body).Decode(&e)
			t.Fatalf("GET %s: %d %s", path, res.StatusCode, e["error"])
		}
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			t.Fatal(err)
		}
	}
	post := func(path string, body any, out any) {
		t.Helper()
		b, _ := json.Marshal(body)
		res, err := ts.Client().Post(ts.URL+path, "application/json", bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			var e map[string]string
			json.NewDecoder(res.Body).Decode(&e)
			t.Fatalf("POST %s: %d %s", path, res.StatusCode, e["error"])
		}
		if out != nil {
			if err := json.NewDecoder(res.Body).Decode(out); err != nil {
				t.Fatal(err)
			}
		}
	}
	activate := func(id string, spec core.RuleSpec) {
		t.Helper()
		if _, err := s.InsertRuleVersion(ctx, core.Rule{
			ID: id, Status: core.StatusDraft, Priority: 10,
			EffectiveFrom: "2026-01-01", CreatedBy: "human", Description: id, Spec: spec,
		}); err != nil {
			t.Fatal(err)
		}
		post("/api/rules/"+id+"/approve", map[string]any{"approved_by": "krzysztof"}, nil)
	}
	invoice := func(dedup, date, market, seller, number, net, vat, gross, rate, currency string) map[string]any {
		var resp map[string]any
		post("/api/events", map[string]any{
			"event_type": "sales.invoice.issued", "occurred_at": date, "dedup_key": dedup,
			"payload": map[string]any{
				"number": number, "issue_date": date, "market": market,
				"seller_nip": seller, "buyer_nip": map[string]string{"pl": "5260001246", "de": "DE123456789"}[market],
				"net": net, "vat": vat, "gross": gross,
				"vat_rate": rate, "currency": currency,
			},
		}, &resp)
		return resp
	}

	// Both markets into one kernel; approving each pack's rules installs
	// only its own master data now — the market binding keeps the other
	// pack's events out of reach.
	for _, p := range []string{"pl", "de"} {
		if _, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", p, "pack.json"), "test"); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{"pl-register-account", "pl-register-vat-rate",
		"pl-book-sales-invoice", "pl-post-sales-invoice", "pl-record-ksef-submission",
		"de-register-account", "de-register-vat-rate",
		"de-book-sales-invoice", "de-post-sales-invoice"} {
		post("/api/rules/"+id+"/approve", map[string]any{"approved_by": "krzysztof"}, nil)
	}
	if wl, _ := s.UnmatchedRawEvents(ctx); len(wl) != 0 {
		t.Fatalf("worklist = %d, want 0 — both markets' master data installed", len(wl))
	}
	accs, _ := s.ObjectsByType(ctx, "account")
	if len(accs) != 40 {
		t.Fatalf("accounts = %d, want both charts", len(accs))
	}
	// The capture is gone: the SKR03 is explained by its own market's rule.
	for _, a := range accs {
		if code, _ := a.State["code"].(string); len(code) == 4 && a.RuleID != "de-register-account" {
			t.Fatalf("SKR03 account %s explained by %s", code, a.RuleID)
		}
	}
	// The namespace holds: both zero rates exist under distinct codes.
	for _, code := range []string{"0", "0-de"} {
		if ids, _ := s.FindObjectIDsByField(ctx, "vat_rate", "code", code); len(ids) != 1 {
			t.Fatalf("vat_rate %q = %d objects", code, len(ids))
		}
	}

	// Two operating companies and their buyers — all ordinary master data.
	activate("register-company", core.RuleSpec{
		Match: core.Match{EventType: "company.registered"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "company",
			Fields: map[string]string{"name": "=$.name", "kind": "=$.kind",
				"vat_id": "=$.vat_id", "country": "=$.country"}}},
	})
	for dedup, c := range map[string]map[string]any{
		"alfa":     {"name": "Alfa Sp. z o.o.", "kind": "self", "vat_id": "5250001111", "country": "PL"},
		"beta":     {"name": "Beta GmbH", "kind": "self", "vat_id": "DE811111111", "country": "DE"},
		"buyer-pl": {"name": "Nabywca S.A.", "kind": "customer", "vat_id": "5260001246", "country": "PL"},
		"buyer-de": {"name": "Käufer AG", "kind": "customer", "vat_id": "DE123456789", "country": "DE"},
	} {
		post("/api/events", map[string]any{
			"event_type": "company.registered", "occurred_at": "2026-09-01",
			"dedup_key": dedup, "payload": c}, nil)
	}

	// Each company's invoice is explained by its own market only.
	if r := invoice("fv-a1", "2026-09-21", "pl", "5250001111", "FV 1/09/2026", "1000.00", "230.00", "1230.00", "23", "PLN"); r["booked"].(float64) != 1 {
		t.Fatalf("alfa invoice: %v", r)
	}
	if r := invoice("re-b1", "2026-09-23", "de", "DE811111111", "RE 2026-001", "500.00", "95.00", "595.00", "19", "EUR"); r["booked"].(float64) != 1 {
		t.Fatalf("beta invoice: %v", r)
	}
	// Currency no longer routes — it converts (E3). Alfa's EUR-denominated
	// domestic invoice stays Polish, and its entry books functional PLN at
	// the NBP rate the kernel reads from an fx_rate object: 246.00 EUR at
	// 4.25 is 1045.50 PLN, transaction amounts carried on every line.
	activate("register-fx-rate", core.RuleSpec{
		Match: core.Match{EventType: "fx.rate.published"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "fx_rate",
			Fields: map[string]string{"code": "=$.code", "base": "=$.base",
				"quote": "=$.quote", "date": "=$.date", "rate": "=$.rate"}}},
	})
	post("/api/events", map[string]any{
		"event_type": "fx.rate.published", "occurred_at": "2026-10-02", "dedup_key": "nbp-1002",
		"payload": map[string]any{"code": "EUR/PLN/2026-10-02", "base": "EUR",
			"quote": "PLN", "date": "2026-10-02", "rate": "4.25"}}, nil)
	if r := invoice("fv-a2", "2026-10-02", "pl", "5250001111", "FV 1/10/2026", "200.00", "46.00", "246.00", "23", "EUR"); r["booked"].(float64) != 1 {
		t.Fatalf("alfa EUR invoice: %v", r)
	}
	alfa, _ := s.FindObjectIDsByField(ctx, "company", "vat_id", "5250001111")
	invoices, _ := s.ObjectsByType(ctx, "sales_invoice")
	if len(invoices) != 3 {
		t.Fatalf("invoices = %d", len(invoices))
	}
	for _, inv := range invoices {
		if inv.State["currency"] == "EUR" && inv.State["number"] == "FV 1/10/2026" {
			if inv.RuleID != "pl-book-sales-invoice" || inv.State["seller"] != alfa[0] {
				t.Fatalf("EUR invoice: rule %s seller %v — the misrouting is back", inv.RuleID, inv.State["seller"])
			}
		}
	}

	// The statutory registers split by the seller's country: each market's
	// register sees only its company's invoices.
	var reg, ust struct {
		Rows [][]string `json:"rows"`
	}
	get("/api/views/rejestr-vat-sprzedazy", &reg)
	if len(reg.Rows) != 2 ||
		reg.Rows[0][0] != "2026-09" || reg.Rows[0][1] != "23" || reg.Rows[0][4] != "1230.00" ||
		reg.Rows[1][0] != "2026-10" || reg.Rows[1][4] != "246.00" {
		t.Fatalf("rejestr = %+v", reg.Rows)
	}
	get("/api/views/ust-je-monat", &ust)
	if len(ust.Rows) != 1 || ust.Rows[0][0] != "2026-09" || ust.Rows[0][1] != "19" || ust.Rows[0][4] != "595.00" {
		t.Fatalf("ust = %+v", ust.Rows)
	}

	// One trial balance notion, two companies' ledgers — and the Polish one
	// is currency-coherent now: PLN plus converted-PLN, never EUR mixed in.
	var tb struct {
		Rows [][]string `json:"rows"`
	}
	get("/api/views/trial-balance", &tb)
	want := map[string][3]string{
		"201":  {"pl-stat", "2275.50", "0.00"},    // 1230.00 + 246.00×4.25
		"700":  {"pl-stat", "0.00", "1850.00"},    // 1000.00 + 200.00×4.25
		"222":  {"pl-stat", "0.00", "425.50"},     // 230.00 + 46.00×4.25
		"1400": {"de-stat", "595.00", "0.00"},
		"8400": {"de-stat", "0.00", "500.00"},
		"1776": {"de-stat", "0.00", "95.00"},
	}
	seen := 0
	for _, r := range tb.Rows {
		if w, ok := want[r[1]]; ok {
			if r[0] != w[0] || r[3] != w[1] || r[4] != w[2] {
				t.Fatalf("account %s: %v, want %v", r[1], r, w)
			}
			seen++
		}
	}
	if seen != 6 {
		t.Fatalf("trial balance = %+v", tb.Rows)
	}

	// Alfa closes September; Beta's September stays open — the E1 book
	// machinery is the per-company close.
	activate("lock-book-period", core.RuleSpec{
		Match: core.Match{EventType: "period.locked"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "period_lock",
			Fields: map[string]string{"month": "=$.month", "book": "=$.book"}}},
	})
	post("/api/events", map[string]any{
		"event_type": "period.locked", "occurred_at": "2026-09-30", "dedup_key": "lock-a-09",
		"payload": map[string]any{"month": "2026-09", "book": "pl-stat"}}, nil)
	if r := invoice("fv-a3", "2026-09-28", "pl", "5250001111", "FV 2/09/2026", "10.00", "2.30", "12.30", "23", "PLN"); r["booked"].(float64) != 0 ||
		!strings.Contains(fmt.Sprint(r["errors"]), "locked for book pl-stat") {
		t.Fatalf("alfa after close: %v", r)
	}
	if r := invoice("re-b2", "2026-09-29", "de", "DE811111111", "RE 2026-002", "100.00", "19.00", "119.00", "19", "EUR"); r["booked"].(float64) != 1 {
		t.Fatalf("beta after alfa's close: %v", r)
	}

	// An event naming no market matches nothing and waits for a human —
	// the silent captures of the attempt are structurally gone.
	var resp map[string]any
	post("/api/events", map[string]any{
		"event_type": "sales.invoice.issued", "occurred_at": "2026-09-25", "dedup_key": "fv-x",
		"payload": map[string]any{
			"number": "FV ???", "issue_date": "2026-09-25", "seller_nip": "5250001111",
			"buyer_nip": "5260001246", "net": "1.00", "vat": "0.23", "gross": "1.23",
			"vat_rate": "23", "currency": "PLN",
		},
	}, &resp)
	if resp["booked"].(float64) != 0 {
		t.Fatalf("marketless invoice booked: %v", resp)
	}
	if wl, _ := s.UnmatchedRawEvents(ctx); len(wl) != 2 { // fv-a3 (locked) + fv-x (unroutable)
		t.Fatalf("worklist = %d", len(wl))
	}

	// One replay, two companies, identical state.
	before, _ := s.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := s.AllObjects(ctx)
	bj, _ := json.Marshal(before)
	aj, _ := json.Marshal(after)
	if !bytes.Equal(bj, aj) {
		t.Fatal("replay diverged")
	}
}
