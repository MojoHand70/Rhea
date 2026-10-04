package shell_test

// The read side of E1 (DIRECTION, enterprise-structure ladder): parallel
// books are visible as parallel trial balances through the same generic
// analysis notion — one view, a book column, no book-specific screen. The
// statutory and group explanations of one sale land on their own charts of
// accounts, and the trial balance splits per book by data alone.

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
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

func TestParallelBooksTrialBalance(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	seedFromFile(t, s, "finance.json")
	seedFromFile(t, s, "finance_v2.json")
	seedFromFile(t, s, "finance_v3.json") // account / posting v1 / trial balance v1
	seedFromFile(t, s, "finance_v4.json")
	seedFromFile(t, s, "finance_v5.json") // posting v2 + book, per-book trial balance

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
	activate := func(id string, spec core.RuleSpec) {
		t.Helper()
		if _, err := s.InsertRuleVersion(ctx, core.Rule{
			ID: id, Status: core.StatusDraft, Priority: 100,
			EffectiveFrom: "2026-01-01", CreatedBy: "human", Description: id, Spec: spec,
		}); err != nil {
			t.Fatal(err)
		}
		post("/api/rules/"+id+"/approve", map[string]any{"approved_by": "krzysztof"}, nil)
	}

	activate("register-account", core.RuleSpec{
		Match: core.Match{EventType: "account.created"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "account",
			Fields: map[string]string{"code": "=$.code", "name": "=$.name", "type": "=$.type"}}},
	})
	activate("pl-stat-post", core.RuleSpec{
		Match: core.Match{EventType: "sale.recorded"},
		Effect: core.Effect{Postings: &core.PostingsTemplate{
			Book: "pl-stat", Currency: "=$.currency",
			Lines: []core.PostingLine{
				{Account: "201", Debit: "=$.gross"},
				{Account: "700", Credit: "=$.gross"},
			}}},
	})
	activate("group-post", core.RuleSpec{
		Match: core.Match{EventType: "sale.recorded"},
		Effect: core.Effect{Postings: &core.PostingsTemplate{
			Book: "group", Currency: "=$.currency",
			Lines: []core.PostingLine{
				{Account: "G1200", Debit: "=$.gross"},
				{Account: "G4000", Credit: "=$.gross"},
			}}},
	})

	for dedup, acc := range map[string]string{
		"acc-201":   `{"code":"201","name":"Rozrachunki z odbiorcami","type":"debtor"}`,
		"acc-700":   `{"code":"700","name":"Sprzedaż produktów","type":"revenue"}`,
		"acc-G1200": `{"code":"G1200","name":"Trade receivables","type":"debtor"}`,
		"acc-G4000": `{"code":"G4000","name":"Revenue","type":"revenue"}`,
	} {
		var payload map[string]any
		if err := json.Unmarshal([]byte(acc), &payload); err != nil {
			t.Fatal(err)
		}
		post("/api/events", map[string]any{
			"event_type": "account.created", "occurred_at": "2026-09-01",
			"dedup_key": dedup, "payload": payload,
		}, nil)
	}

	var resp map[string]any
	post("/api/events", map[string]any{
		"event_type": "sale.recorded", "occurred_at": "2026-09-15", "dedup_key": "sale-1",
		"payload": map[string]any{"gross": "100.00", "currency": "PLN"},
	}, &resp)
	if resp["booked"].(float64) != 1 || resp["errors"] != nil {
		t.Fatalf("sale: %v", resp)
	}

	// One generic analysis view, two ledgers: the trial balance splits per
	// book, each book balanced on its own chart of accounts.
	var tb struct {
		Rows [][]string `json:"rows"`
	}
	get("/api/views/trial-balance", &tb)
	want := [][]string{
		{"group", "G1200", "Trade receivables", "100.00", "0.00", "100.00"},
		{"group", "G4000", "Revenue", "0.00", "100.00", "-100.00"},
		{"pl-stat", "201", "Rozrachunki z odbiorcami", "100.00", "0.00", "100.00"},
		{"pl-stat", "700", "Sprzedaż produktów", "0.00", "100.00", "-100.00"},
	}
	if !reflect.DeepEqual(tb.Rows, want) {
		t.Fatalf("trial balance:\n got %v\nwant %v", tb.Rows, want)
	}
}
