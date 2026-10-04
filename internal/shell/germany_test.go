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

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

// TestM4GermanyPack is the proof that M3's pack boundary was real (SPEC M4):
// a second market arrives as one data file and nothing else — this test and
// packs/de/pack.json are the entire diff; kernel, pack loader and adapter
// contract are byte-for-byte the ones Poland runs on. Germany ships ZERO
// object types: the market-neutral shapes (vat_rate, sales_invoice) live in
// the finance base, and a market is pure rules, master data and views —
// SKR03 instead of the wzorcowy plan kont, 19/7 instead of 23/8/5,
// 1400/8400/1776 instead of 201/700/222.
func TestM4GermanyPack(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	seedFromFile(t, s, "finance.json")
	seedFromFile(t, s, "finance_v2.json") // company master data
	seedFromFile(t, s, "finance_v3.json") // account / posting / trial balance
	seedFromFile(t, s, "finance_v4.json") // market-neutral: vat_rate, sales_invoice
	seedFromFile(t, s, "finance_v5.json") // books: posting v2, per-book trial balance
	seedFromFile(t, s, "finance_v6.json") // sales_invoice v2: the seller is a ref

	srv := &shell.Server{
		Store: s, Exec: &exec.Executor{Store: s},
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

	// A market with no types: 3 views, 4 draft rules, 17 master-data events.
	sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", "de", "pack.json"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if sum.Types != 0 || sum.Views != 3 || sum.Rules != 4 || sum.Events != 17 || sum.Skipped != 0 {
		t.Fatalf("load: %+v", sum)
	}

	// Same gate, same installation: approving the rules books the SKR03
	// chart and the Umsatzsteuer rates out of the worklist.
	for _, id := range []string{"de-register-account", "de-register-vat-rate",
		"de-book-sales-invoice", "de-post-sales-invoice"} {
		post("/api/rules/"+id+"/approve", map[string]any{"approved_by": "krzysztof"}, nil)
	}
	if accs, _ := s.ObjectsByType(ctx, "account"); len(accs) != 14 {
		t.Fatalf("accounts = %d, want the SKR03 chart", len(accs))
	}
	if rates, _ := s.ObjectsByType(ctx, "vat_rate"); len(rates) != 3 {
		t.Fatalf("vat rates = %d", len(rates))
	}

	registerCompany := core.RuleSpec{
		Match: core.Match{EventType: "company.registered"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "company",
			Fields: map[string]string{"name": "=$.name", "kind": "=$.kind",
				"vat_id": "=$.vat_id", "country": "=$.country"}}},
	}
	if _, err := s.InsertRuleVersion(ctx, core.Rule{
		ID: "register-company", Status: core.StatusDraft, Priority: 10,
		EffectiveFrom: "2026-01-01", CreatedBy: "human", Description: "register-company",
		Spec: registerCompany,
	}); err != nil {
		t.Fatal(err)
	}
	post("/api/rules/register-company/approve", map[string]any{"approved_by": "krzysztof"}, nil)
	post("/api/events", map[string]any{
		"event_type": "company.registered", "occurred_at": "2026-09-01", "dedup_key": "kaeufer",
		"payload": map[string]any{"name": "Käufer GmbH", "kind": "customer",
			"vat_id": "DE123456789", "country": "DE"},
	}, nil)
	post("/api/events", map[string]any{
		"event_type": "company.registered", "occurred_at": "2026-09-01", "dedup_key": "verkaeufer",
		"payload": map[string]any{"name": "Verkäufer GmbH", "kind": "self",
			"vat_id": "DE811111111", "country": "DE"},
	}, nil)

	// A German invoice books document + 1400/8400/1776 entry atomically.
	var resp map[string]any
	post("/api/events", map[string]any{
		"event_type": "sales.invoice.issued", "occurred_at": "2026-09-23", "dedup_key": "re-1",
		"payload": map[string]any{
			"number": "RE 2026-001", "issue_date": "2026-09-23", "market": "de",
			"seller_nip": "DE811111111", "buyer_nip": "DE123456789",
			"net": "500.00", "vat": "95.00", "gross": "595.00",
			"vat_rate": "19", "currency": "EUR",
		},
	}, &resp)
	if resp["booked"].(float64) != 1 || resp["errors"] != nil {
		t.Fatalf("rechnung: %v", resp)
	}

	var list struct {
		Rows [][]string `json:"rows"`
	}
	get("/api/views/rechnungen-list", &list)
	if len(list.Rows) != 1 {
		t.Fatalf("rechnungen = %+v", list.Rows)
	}
	if r := list.Rows[0]; r[0] != "RE 2026-001" || r[2] != "Käufer GmbH" || r[5] != "595.00" {
		t.Fatalf("rechnung row = %v", r)
	}

	var tb struct {
		Rows [][]string `json:"rows"`
	}
	get("/api/views/trial-balance", &tb)
	want := map[string][2]string{"1400": {"595.00", "0.00"}, "8400": {"0.00", "500.00"}, "1776": {"0.00", "95.00"}}
	seen := 0
	for _, r := range tb.Rows {
		if w, ok := want[r[1]]; ok {
			if r[0] != "de-stat" || r[3] != w[0] || r[4] != w[1] {
				t.Fatalf("account %s: book %s debit %s credit %s, want de-stat %v", r[1], r[0], r[3], r[4], w)
			}
			seen++
		}
	}
	if seen != 3 {
		t.Fatalf("trial balance missing SKR03 accounts: %+v", tb.Rows)
	}

	var ust struct {
		Rows [][]string `json:"rows"`
	}
	get("/api/views/ust-je-monat", &ust)
	if len(ust.Rows) != 1 {
		t.Fatalf("ust = %+v", ust.Rows)
	}
	if r := ust.Rows[0]; r[0] != "2026-09" || r[1] != "19" || r[2] != "500.00" || r[3] != "95.00" || r[4] != "595.00" {
		t.Fatalf("ust row = %v", r)
	}

	// Exclusivity is explicit in the match predicates: a Polish-market
	// invoice matches no German rule and waits in the worklist — no silent
	// cross-market booking, whatever its currency says.
	post("/api/events", map[string]any{
		"event_type": "sales.invoice.issued", "occurred_at": "2026-09-24", "dedup_key": "fv-pl",
		"payload": map[string]any{
			"number": "FV 3/09/2026", "issue_date": "2026-09-24", "market": "pl",
			"seller_nip": "5250001111", "buyer_nip": "DE123456789",
			"net": "100.00", "vat": "23.00", "gross": "123.00",
			"vat_rate": "19", "currency": "EUR",
		},
	}, &resp)
	if resp["booked"].(float64) != 0 {
		t.Fatalf("PLN invoice booked by the German pack: %v", resp)
	}
	if wl, _ := s.UnmatchedRawEvents(ctx); len(wl) != 1 {
		t.Fatalf("worklist = %d", len(wl))
	}

	// Determinism holds for the second market too.
	x := &exec.Executor{Store: s}
	before, _ := s.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := s.AllObjects(ctx)
	bj, _ := json.Marshal(before)
	aj, _ := json.Marshal(after)
	if !reflect.DeepEqual(bj, aj) {
		t.Fatal("replay diverged")
	}
}
