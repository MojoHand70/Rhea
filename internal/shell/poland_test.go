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

	"rhea/internal/adapter"
	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

// TestM3PolandPack is the market-pack milestone (SPEC M3): Poland arrives as
// one data file — wzorcowy plan kont, VAT rates, the KSeF invoice schema,
// booking rules and the VAT register — layered on the finance base with zero
// kernel changes. Loading installs nothing: the pack's rules are drafts, its
// master data waits in the worklist, and approving the rules in the shell is
// what brings the market to life. Loading twice is a no-op. The one piece of
// code a market brings — the statutory adapter — stays behind the declared
// contract: reads objects, returns evidence events, never writes.
func TestM3PolandPack(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	seedFromFile(t, s, "finance.json")    // invoice v1 base
	seedFromFile(t, s, "finance_v2.json") // company master data
	seedFromFile(t, s, "finance_v3.json") // account / posting / trial balance
	seedFromFile(t, s, "finance_v4.json") // market-neutral: vat_rate, sales_invoice

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

	// Load the pack: 1 type, 5 views, 5 draft rules, 30 master-data events.
	packPath := filepath.Join("..", "..", "packs", "pl", "pack.json")
	sum, err := pack.Load(ctx, s, packPath, "test")
	if err != nil {
		t.Fatal(err)
	}
	if sum.Types != 1 || sum.Views != 5 || sum.Rules != 5 || sum.Events != 30 || sum.Skipped != 0 {
		t.Fatalf("first load: %+v", sum)
	}
	// Idempotent: a reload changes nothing.
	sum, err = pack.Load(ctx, s, packPath, "test")
	if err != nil {
		t.Fatal(err)
	}
	if sum.Types != 0 || sum.Views != 0 || sum.Rules != 0 || sum.Events != 0 || sum.Skipped != 41 {
		t.Fatalf("reload: %+v", sum)
	}

	// Nothing is installed yet: the chart of accounts sits in the worklist,
	// because the pack's rules are drafts behind the human gate.
	if wl, _ := s.UnmatchedRawEvents(ctx); len(wl) != 30 {
		t.Fatalf("worklist = %d, want 30 pack events waiting", len(wl))
	}

	// Approving the pack's rules IS the installation: the chart of accounts
	// and the VAT rates materialize from the waiting events.
	for _, id := range []string{"pl-register-account", "pl-register-vat-rate",
		"pl-book-sales-invoice", "pl-post-sales-invoice", "pl-record-ksef-submission"} {
		post("/api/rules/"+id+"/approve", map[string]any{"approved_by": "krzysztof"}, nil)
	}
	if wl, _ := s.UnmatchedRawEvents(ctx); len(wl) != 0 {
		t.Fatalf("worklist not cleared after approvals: %d", len(wl))
	}
	if accs, _ := s.ObjectsByType(ctx, "account"); len(accs) != 25 {
		t.Fatalf("accounts = %d, want the chart of accounts", len(accs))
	}
	if rates, _ := s.ObjectsByType(ctx, "vat_rate"); len(rates) != 5 {
		t.Fatalf("vat rates = %d", len(rates))
	}

	// The buyer is ordinary company master data (the finance base at work).
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
		"event_type": "company.registered", "occurred_at": "2026-09-01", "dedup_key": "buyer",
		"payload": map[string]any{"name": "Nabywca S.A.", "kind": "customer",
			"vat_id": "5260001246", "country": "PL"},
	}, nil)

	// A KSeF sales invoice arrives: one event becomes a document AND a
	// three-line VAT entry, atomically — and the double-entry balance
	// invariant verifies the invoice's own arithmetic (gross = net + VAT).
	var resp map[string]any
	post("/api/events", map[string]any{
		"event_type": "sales.invoice.issued", "occurred_at": "2026-09-21", "dedup_key": "fv-1",
		"payload": map[string]any{
			"number": "FV 1/09/2026", "issue_date": "2026-09-21", "buyer_nip": "5260001246",
			"net": "1000.00", "vat": "230.00", "gross": "1230.00",
			"vat_rate": "23", "currency": "PLN",
		},
	}, &resp)
	if resp["booked"].(float64) != 1 || resp["errors"] != nil {
		t.Fatalf("ksef invoice: %v", resp)
	}

	// The document renders with resolved labels; money formats from minor units.
	var list struct {
		Rows [][]string `json:"rows"`
	}
	get("/api/views/sales-invoice-list", &list)
	if len(list.Rows) != 1 {
		t.Fatalf("invoice list = %+v", list.Rows)
	}
	if r := list.Rows[0]; r[0] != "FV 1/09/2026" || r[2] != "Nabywca S.A." || r[3] != "1000.00" || r[5] != "1230.00" {
		t.Fatalf("invoice row = %v", r)
	}

	// The trial balance carries the entry on the Polish accounts.
	var tb struct {
		Rows [][]string `json:"rows"`
	}
	get("/api/views/trial-balance", &tb)
	want := map[string][2]string{"201": {"1230.00", "0.00"}, "700": {"0.00", "1000.00"}, "222": {"0.00", "230.00"}}
	seen := 0
	for _, r := range tb.Rows {
		if w, ok := want[r[0]]; ok {
			if r[2] != w[0] || r[3] != w[1] {
				t.Fatalf("account %s: debit %s credit %s, want %v", r[0], r[2], r[3], w)
			}
			seen++
		}
	}
	if seen != 3 {
		t.Fatalf("trial balance missing pack accounts: %+v", tb.Rows)
	}

	// The VAT register aggregates by month and rate.
	var reg struct {
		Rows [][]string `json:"rows"`
	}
	get("/api/views/rejestr-vat-sprzedazy", &reg)
	if len(reg.Rows) != 1 {
		t.Fatalf("rejestr = %+v", reg.Rows)
	}
	if r := reg.Rows[0]; r[0] != "2026-09" || r[1] != "23" || r[2] != "1000.00" || r[3] != "230.00" || r[4] != "1230.00" {
		t.Fatalf("rejestr row = %v", r)
	}

	// The statutory adapter: a user, not a kernel extension. It finds the
	// unsubmitted invoice, "submits" it through its (fake) client, and the
	// evidence comes back as an ordinary raw event that a pack rule books
	// into a ksef_submission object. A second pass finds nothing owed.
	ksef := adapter.KSeF{Client: adapter.FakeKSeF{}, Today: func() string { return "2026-09-21" }}
	run, err := adapter.Run(ctx, s, srv.Exec, ksef)
	if err != nil || run.PassErr != nil {
		t.Fatalf("adapter run: %v / %v", err, run.PassErr)
	}
	if run.Submitted != 1 || run.Booked != 1 || len(run.Errors) != 0 {
		t.Fatalf("adapter run: %+v", run)
	}
	var subs struct {
		Rows [][]string `json:"rows"`
	}
	get("/api/views/ksef-submissions", &subs)
	if len(subs.Rows) != 1 {
		t.Fatalf("submissions = %+v", subs.Rows)
	}
	if r := subs.Rows[0]; r[0] != "FV 1/09/2026" || len(r[1]) == 0 || len(r[2]) == 0 {
		t.Fatalf("submission row = %v", r)
	}
	run, err = adapter.Run(ctx, s, srv.Exec, ksef)
	if err != nil || run.Submitted != 0 || run.Duplicates != 0 {
		t.Fatalf("second pass should owe nothing: %+v, %v", run, err)
	}

	// An invoice whose arithmetic lies (gross ≠ net + VAT) is refused whole
	// by the balance invariant: no document, no postings, into the worklist.
	post("/api/events", map[string]any{
		"event_type": "sales.invoice.issued", "occurred_at": "2026-09-22", "dedup_key": "fv-bad",
		"payload": map[string]any{
			"number": "FV 2/09/2026", "issue_date": "2026-09-22", "buyer_nip": "5260001246",
			"net": "1000.00", "vat": "230.00", "gross": "1200.00",
			"vat_rate": "23", "currency": "PLN",
		},
	}, &resp)
	if resp["booked"].(float64) != 0 || resp["errors"] == nil {
		t.Fatalf("unbalanced invoice: %v", resp)
	}
	if docs, _ := s.ObjectsByType(ctx, "sales_invoice"); len(docs) != 1 {
		t.Fatalf("unbalanced invoice half-booked: %d docs", len(docs))
	}

	// Determinism holds for a whole market loaded at runtime.
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
