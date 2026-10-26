package shell_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"rhea/internal/adapter"
	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
	"rhea/markets/pl"
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
	seedFromFile(t, s, "finance_v5.json") // books: posting v2, per-book trial balance
	seedFromFile(t, s, "finance_v6.json") // sales_invoice v2: the seller is a ref
	seedFromFile(t, s, "finance_v7.json") // posting v3 (tx trace), fx_rate
	seedFromFile(t, s, "finance_v8.json") // sales_invoice v3 (intragroup + nip echoes), purchase_invoice

	srv := &shell.Server{
		Store: s, Exec: &exec.Executor{Store: s},
		Agent:    &agent.Agent{Complete: func(context.Context, string, string) (string, error) { return "", fmt.Errorf("no agent in this test") }},
		DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb"),
	}
	ts := serve(t, srv.Handler())

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

	// People first: the pl pack's task vocabulary assigns work to persons.
	if _, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", "common", "people.json"), "test"); err != nil {
		t.Fatal(err)
	}
	post("/api/bundles/pack-common-people-v4/approve", map[string]any{"approved_by": "krzysztof"}, nil)
	// Periods: the close_period door and the lock rule are common verbs.
	if _, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", "common", "periods.json"), "test"); err != nil {
		t.Fatal(err)
	}
	post("/api/bundles/pack-common-periods-v3/approve", map[string]any{"approved_by": "krzysztof"}, nil)
	// Documents: the issued_document the JPK file becomes.
	if _, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", "common", "documents.json"), "test"); err != nil {
		t.Fatal(err)
	}
	post("/api/bundles/pack-common-documents-v5/approve", map[string]any{"approved_by": "krzysztof"}, nil)

	// Load the pack: 25 types (the invoices with their open amounts and
	// dates, the bank operation, the matching difference, the company with
	// its declared form, the account with its tax treatment, the CIT limits,
	// status and advance, the prepaid cost and its monthly part), 26 views (the two grids, the two mornings — the
	// owner's and the accountant's — and the two queues among them; the
	// words, the VAT declaration as a statement, the morning figures and the
	// four the mornings speak from — the result, money, the tax deadlines,
	// what is open — profit since January and month by month, receivables
	// by age, the CIT advances as a list and a statement, the prepaid
	// costs), 52 draft rules, ten doors (the KSeF channel, identify_company,
	// declare_company_form, set_cost_account, declare_company_accounts,
	// classify_bank_operation, set_tax_treatment, declare_cit_status,
	// recompute_cit_advance, spread_cost), 49 master-data events. (The counts
	// below are the pack as it stands; v44 added enter_company and its two rules,
	// v45 wrote the pack as 24 Apps.)
	packPath := filepath.Join("..", "..", "packs", "pl", "pack.json")
	sum, err := pack.Load(ctx, s, packPath, "test")
	if err != nil {
		t.Fatal(err)
	}
	if sum.Types != 41 || sum.Views != 38 || sum.Rules != 138 || sum.Activities != 39 || sum.Events != 67 || sum.Apps != 24 || sum.Skipped != 0 ||
		sum.Bundle != "pack-pl-v45" {
		t.Fatalf("first load: %+v", sum)
	}
	// Idempotent: a reload changes nothing — its 24 Apps skipped with the rest.
	sum, err = pack.Load(ctx, s, packPath, "test")
	if err != nil {
		t.Fatal(err)
	}
	if sum.Types != 0 || sum.Views != 0 || sum.Rules != 0 || sum.Activities != 0 || sum.Events != 0 || sum.Apps != 0 || sum.Skipped != 347 || sum.Bundle != "" {
		t.Fatalf("reload: %+v", sum)
	}

	// Before approval the KSeF channel is a draft: the adapter is implemented
	// but not connected, and a pass is refused before it reads anything.
	ksefDraft := pl.KSeF{Client: pl.FakeKSeF{}}
	if _, err := adapter.Run(ctx, s, srv.Exec, ksefDraft); err == nil || !strings.Contains(err.Error(), "no active channel") {
		t.Fatalf("a draft channel let the adapter through: %v", err)
	}

	// Nothing is installed yet: the chart of accounts sits in the worklist,
	// because the pack's rules are drafts behind the human gate.
	if wl, _ := s.UnmatchedRawEvents(ctx); len(wl) != 67 {
		t.Fatalf("worklist = %d, want 60 pack events waiting", len(wl))
	}

	// Approving the pack's bundle IS the installation: its types, views and
	// rules activate together, and the chart of accounts and the VAT rates
	// materialize from the waiting events.
	post("/api/bundles/pack-pl-v45/approve", map[string]any{"approved_by": "krzysztof"}, nil)
	if wl, _ := s.UnmatchedRawEvents(ctx); len(wl) != 0 {
		t.Fatalf("worklist not cleared after approvals: %d", len(wl))
	}
	if accs, _ := s.ObjectsByType(ctx, "account"); len(accs) != 55 {
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
				"vat_id": "=$.vat_id", "country": "=$.country", "email": "=$.email", "tax_office": "=$.tax_office", "legal_form": "=$.legal_form"}}},
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
		"payload": map[string]any{"name": "Nabywca S.A.", "kind": "customer", "email": "biuro@nabywca.pl", "tax_office": "1472", "legal_form": "legal_entity",
			"vat_id": "5260001246", "country": "PL", "iban": "PL60102010260000042270201111"},
	}, nil)
	post("/api/events", map[string]any{
		"event_type": "company.registered", "occurred_at": "2026-09-01", "dedup_key": "seller",
		"payload": map[string]any{"name": "Sprzedawca Sp. z o.o.", "kind": "self", "email": "ksiegowosc@sprzedawca.pl", "tax_office": "1471", "legal_form": "legal_entity",
			"vat_id": "5250001111", "country": "PL"},
	}, nil)

	// A KSeF sales invoice arrives: one event becomes a document AND a
	// three-line VAT entry, atomically — and the double-entry balance
	// invariant verifies the invoice's own arithmetic (gross = net + VAT).
	var resp map[string]any
	post("/api/events", map[string]any{
		"event_type": "sales.invoice.issued", "occurred_at": "2026-09-21", "dedup_key": "fv-1",
		"payload": map[string]any{
			"number": "FV 1/09/2026", "issue_date": "2026-09-21", "market": "pl", "intragroup": "no",
			"seller_nip": "5250001111", "buyer_nip": "5260001246",
			"net": "1000.00", "vat": "230.00", "gross": "1230.00",
			"vat_rate": "23", "currency": "PLN",
		},
	}, &resp)
	if resp["booked"].(float64) != 1 || resp["errors"] != nil {
		t.Fatalf("ksef invoice: %v", resp)
	}

	// The document renders with resolved labels; money formats from minor units.
	var list struct {
		Rows [][]cell `json:"rows"`
	}
	get("/api/views/sales-invoice-list", &list)
	if len(list.Rows) != 1 {
		t.Fatalf("invoice list = %+v", list.Rows)
	}
	if r := cellVals(list.Rows)[0]; r[0] != "FV 1/09/2026" || r[2] != "Nabywca S.A." || r[3] != "1000.00" || r[5] != "1230.00" {
		t.Fatalf("invoice row = %v", r)
	}

	// The queue: the open receivable is what there is to chase today, ranked
	// by how long it has been outstanding; the paid one is not in it.
	var queue struct {
		Items []struct {
			Label string `json:"label"`
			Rank  string `json:"rank"`
		} `json:"items"`
		Verbs []struct {
			Name string `json:"name"`
		} `json:"verbs"`
	}
	get("/api/views/naleznosci-do-pilnowania", &queue)
	if len(queue.Items) != 1 || queue.Items[0].Label != "FV 1/09/2026" {
		t.Fatalf("the queue: %+v", queue)
	}

	// The trial balance carries the entry on the Polish accounts.
	var tb struct {
		Rows [][]cell `json:"rows"`
	}
	get("/api/views/trial-balance", &tb)
	want := map[string][2]string{"201": {"1230.00", "0.00"}, "700": {"0.00", "1000.00"}, "222": {"0.00", "230.00"}}
	seen := 0
	for _, r := range cellVals(tb.Rows) {
		if w, ok := want[r[1]]; ok {
			if r[0] != "pl-stat" || r[3] != w[0] || r[4] != w[1] {
				t.Fatalf("account %s: book %s debit %s credit %s, want pl-stat %v", r[1], r[0], r[3], r[4], w)
			}
			seen++
		}
	}
	if seen != 3 {
		t.Fatalf("trial balance missing pack accounts: %+v", tb.Rows)
	}

	// The purchase side (DIRECTION 2026-10-10, step 5): a supplier's invoice
	// as the KSeF reader hands it — the document's own VAT and gross — is a
	// purchase_invoice document and a three-line entry, Wn 403 / Wn 221 /
	// Ma 202 — a service with no delivery note is a cost, not a purchase
	// waiting on 300 for goods — and the purchase VAT register shows it.
	post("/api/events", map[string]any{
		"event_type": "purchase.invoice.received", "occurred_at": "2026-09-22", "dedup_key": "fab-1",
		"payload": map[string]any{
			"number": "FAB/1/2026", "invoice_date": "2026-09-22", "market": "pl",
			"supplier_nip": "5260001246", "buyer_nip": "5250001111", "supplier": "Nabywca S.A.",
			"net": "500.00", "vat": "115.00", "gross": "615.00", "vat_rate": "23", "currency": "PLN",
		},
	}, &resp)
	if resp["booked"].(float64) != 1 || resp["errors"] != nil {
		t.Fatalf("purchase invoice: %v", resp)
	}
	get("/api/views/trial-balance", &tb)
	wantPurchase := map[string][2]string{"403": {"500.00", "0.00"}, "221": {"115.00", "0.00"}, "202": {"0.00", "615.00"}}
	seen = 0
	for _, r := range cellVals(tb.Rows) {
		if w, ok := wantPurchase[r[1]]; ok {
			if r[0] != "pl-stat" || r[3] != w[0] || r[4] != w[1] {
				t.Fatalf("account %s: book %s debit %s credit %s, want pl-stat %v", r[1], r[0], r[3], r[4], w)
			}
			seen++
		}
	}
	if seen != 3 {
		t.Fatalf("trial balance missing the purchase entry: %+v", tb.Rows)
	}
	var purchases struct {
		Rows [][]cell `json:"rows"`
	}
	get("/api/views/rejestr-vat-zakupu", &purchases)
	if len(purchases.Rows) != 1 {
		t.Fatalf("purchase VAT register = %+v", purchases.Rows)
	}
	if r := cellVals(purchases.Rows)[0]; r[0] != "2026-09" || r[1] != "23" || r[2] != "500.00" || r[3] != "115.00" || r[4] != "615.00" {
		t.Fatalf("purchase register row = %v", r)
	}

	// The bank story, the pack's default policy (DIRECTION 2026-10-11, the
	// matching brief): the buyer's account came with its registration (the
	// pack's enrichment copies it); a statement line naming the invoice
	// becomes a bank operation, posts Wn 130 / Ma 201, and is matched by
	// reference — the invoice paid, the operation settled, a settlement
	// recorded. The same party, paid as a supplier, settles the purchase.
	post("/api/events", map[string]any{
		"event_type": "bank.statement.line", "occurred_at": "2026-09-25", "dedup_key": "wb-1",
		"payload": map[string]any{"date": "2026-09-25", "amount": "1230.00", "currency": "PLN", "direction": "credit", "market": "pl",
			"reference": "zapłata FV 1/09/2026", "counterparty": "Nabywca S.A.",
			"counterparty_account": "PL60102010260000042270201111", "account": "PL83101010230000261395100000"},
	}, &resp)
	if resp["booked"].(float64) != 1 || resp["errors"] != nil {
		t.Fatalf("incoming payment: %v", resp)
	}
	post("/api/events", map[string]any{
		"event_type": "bank.statement.line", "occurred_at": "2026-09-26", "dedup_key": "wb-2",
		"payload": map[string]any{"date": "2026-09-26", "amount": "615.00", "currency": "PLN", "direction": "debit", "market": "pl",
			"reference": "FAB/1/2026", "counterparty": "Nabywca S.A.",
			"counterparty_account": "PL60102010260000042270201111", "account": "PL83101010230000261395100000"},
	}, &resp)
	if resp["booked"].(float64) != 1 || resp["errors"] != nil {
		t.Fatalf("outgoing payment: %v", resp)
	}
	for _, typ := range []string{"sales_invoice", "purchase_invoice"} {
		docs, _ := s.ObjectsByType(ctx, typ)
		if len(docs) != 1 || docs[0].State["status"] != "paid" || docs[0].State["open"].(float64) != 0 {
			t.Fatalf("%s after its payment: %v", typ, docs[0].State)
		}
	}
	if ops, _ := s.ObjectsByType(ctx, "bank_operation"); len(ops) != 2 || ops[0].State["settlement_status"] != "settled" || ops[1].State["settlement_status"] != "settled" {
		t.Fatalf("bank operations: %v", ops)
	}
	if stl, _ := s.ObjectsByType(ctx, "settlement"); len(stl) != 2 || stl[0].State["document_type"] != "sales_invoice" || stl[1].State["document_type"] != "purchase_invoice" {
		t.Fatalf("settlements: %v", stl)
	}
	get("/api/views/trial-balance", &tb)
	wantBank := map[string][2]string{"130": {"1230.00", "615.00"}, "201": {"1230.00", "1230.00"}, "202": {"615.00", "615.00"}}
	seen = 0
	for _, r := range cellVals(tb.Rows) {
		if w, ok := wantBank[r[1]]; ok {
			if r[0] != "pl-stat" || r[3] != w[0] || r[4] != w[1] {
				t.Fatalf("account %s: book %s debit %s credit %s, want pl-stat %v", r[1], r[0], r[3], r[4], w)
			}
			seen++
		}
	}
	if seen != 3 {
		t.Fatalf("trial balance missing the bank entries: %+v", tb.Rows)
	}
	// A payment nothing fits stays for the person, on the grid, with why.
	post("/api/events", map[string]any{
		"event_type": "bank.statement.line", "occurred_at": "2026-09-27", "dedup_key": "wb-3",
		"payload": map[string]any{"date": "2026-09-27", "amount": "100.00", "currency": "PLN", "direction": "credit", "market": "pl",
			"reference": "przelew", "counterparty": "Nabywca S.A.",
			"counterparty_account": "PL60102010260000042270201111", "account": "PL83101010230000261395100000"},
	}, &resp)
	if resp["booked"].(float64) != 1 || resp["errors"] != nil {
		t.Fatalf("unmatched payment: %v", resp)
	}
	var grid struct {
		Sides []struct {
			Name  string `json:"name"`
			Items []struct {
				Why []string `json:"why"`
			} `json:"items"`
		} `json:"sides"`
	}
	get("/api/views/wplaty-do-rozliczenia", &grid)
	if len(grid.Sides) != 2 || grid.Sides[0].Name != "payment" || len(grid.Sides[0].Items) != 1 || len(grid.Sides[0].Items[0].Why) != 6 ||
		!strings.Contains(grid.Sides[0].Items[0].Why[5], "oldest_first: nothing open on invoice") {
		t.Fatalf("the grid: %+v", grid)
	}
	// The outgoing grid lists what its rule would fire on — debits — so the
	// open credit is not on it (DECISIONS 2026-10-15, the grid re-diagnosed).
	grid.Sides = nil
	get("/api/views/wyplaty-do-rozliczenia", &grid)
	if len(grid.Sides) != 2 || grid.Sides[0].Name != "payment" || len(grid.Sides[0].Items) != 0 {
		t.Fatalf("a credit on the outgoing grid: %+v", grid)
	}

	// The briefing: the pack's narrative view draws on the result, money,
	// the tax deadlines and what is open (TestTheMorningByRole); with no
	// host in this test the facts speak for themselves.
	var briefing struct {
		Silent string `json:"silent"`
		Facts  []struct {
			View string `json:"view"`
			Kind string `json:"kind"`
		} `json:"facts"`
	}
	get("/api/views/poranek", &briefing)
	if briefing.Silent != "" || len(briefing.Facts) != 4 || briefing.Facts[0].View != "poranek-wynik" || briefing.Facts[2].Kind != "analysis" {
		t.Fatalf("the briefing: %+v", briefing)
	}
	var spoken []struct {
		Silent string `json:"silent"`
	}
	post("/api/speak?view=poranek", nil, &spoken)
	if len(spoken) != 1 || !strings.Contains(spoken[0].Silent, "could not speak") {
		t.Fatalf("speaking without a host: %+v", spoken)
	}

	// The VAT register aggregates by month and rate.
	var reg struct {
		Rows [][]cell `json:"rows"`
	}
	get("/api/views/rejestr-vat-sprzedazy", &reg)
	if len(reg.Rows) != 1 {
		t.Fatalf("rejestr = %+v", reg.Rows)
	}
	if r := cellVals(reg.Rows)[0]; r[0] != "2026-09" || r[1] != "23" || r[2] != "1000.00" || r[3] != "230.00" || r[4] != "1230.00" {
		t.Fatalf("rejestr row = %v", r)
	}

	// The statutory adapter: a user, not a kernel extension. It finds the
	// unsubmitted invoice, "submits" it through its (fake) client, and the
	// evidence comes back as an ordinary raw event that a pack rule books
	// into a ksef_submission object. A second pass finds nothing owed.
	ksef := pl.KSeF{Client: pl.FakeKSeF{}, Today: func() string { return "2026-09-21" }}
	run, err := adapter.Run(ctx, s, srv.Exec, ksef)
	if err != nil || run.PassErr != nil {
		t.Fatalf("adapter run: %v / %v", err, run.PassErr)
	}
	if run.Submitted != 1 || run.Booked != 1 || len(run.Errors) != 0 || run.Channel != "ksef" {
		t.Fatalf("adapter run: %+v", run)
	}
	// The evidence names its channel, the way an event through a door names
	// its activity: provenance on the raw side reaches intake.
	if ev, ok, err := s.LatestEventOfType(ctx, "ksef.invoice.submitted"); err != nil || !ok ||
		ev.ActivityName != "ksef" || ev.ActivityVersion != run.Version || ev.Actor != "adapter:ksef" {
		t.Fatalf("evidence stamp: %+v %v %v", ev, ok, err)
	}
	var registry struct {
		Channels []struct {
			Name, State string
		} `json:"channels"`
		Adapters []struct {
			Name, State, Channel string
		} `json:"adapters"`
	}
	get("/api/channels", &registry)
	connected := map[string]string{}
	for _, c := range registry.Channels {
		connected[c.Name] = c.State
	}
	if connected["ksef"] != "connected" || connected["clock"] != "connected" {
		t.Fatalf("channels = %+v", registry.Channels)
	}
	for _, a := range registry.Adapters {
		if a.Name == "ksef" && (a.State != "connected" || a.Channel != "ksef") {
			t.Fatalf("ksef adapter = %+v", a)
		}
	}
	var subs struct {
		Rows [][]cell `json:"rows"`
	}
	get("/api/views/ksef-submissions", &subs)
	if len(subs.Rows) != 1 {
		t.Fatalf("submissions = %+v", subs.Rows)
	}
	if r := cellVals(subs.Rows)[0]; r[0] != "FV 1/09/2026" || len(r[1]) == 0 || len(r[2]) == 0 {
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
			"number": "FV 2/09/2026", "issue_date": "2026-09-22", "market": "pl", "intragroup": "no",
			"seller_nip": "5250001111", "buyer_nip": "5260001246",
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

	// Month end (DIRECTION 2026-10-15, "Month end"): a person closes
	// September for the statutory book through the common door; the common
	// rule locks the period, the pack's rule settles VAT — every position a
	// fold over the month's invoices, with provenance — and the list view
	// shows it. A posting into the closed month is refused per event.
	closed, err := srv.Exec.TriggerActivity(ctx, "close_period", map[string]any{"period": "2026-09-30", "book": "pl-stat", "occurred_at": "2026-10-05"}, "krzysztof", "")
	if err != nil || closed.Booked != 1 {
		t.Fatalf("close_period: %+v %v", closed, err)
	}
	// The worklist is re-read when the world changes: the unbalanced
	// September invoice still waiting there would now book in October, and
	// still refuses for its own reason — its arithmetic lies.
	if len(closed.Errors) != 1 || !strings.Contains(closed.Errors[0].Error(), "does not balance") {
		t.Fatalf("the unbalanced invoice after the close: %v", closed.Errors)
	}
	settlements, _ := s.ObjectsByType(ctx, "vat_settlement")
	if len(settlements) != 1 {
		t.Fatalf("settlements: %+v", settlements)
	}
	st := settlements[0].State
	expect := map[string]float64{"sales_net_23": 100000, "sales_vat_23": 23000, "sales_net_8": 0, "sales_net_exempt": 0, "output_vat": 23000,
		"purchase_net": 50000, "purchase_vat": 11500, "carried_in": 0, "input_vat": 11500, "vat_due": 11500, "vat_carry": 0}
	for k, v := range expect {
		if got, _ := st[k].(float64); got != v {
			t.Fatalf("settlement %s = %v, want %v (state %v)", k, st[k], v, st)
		}
	}
	if st["period"] != "2026-09" || st["book"] != "pl-stat" || st["closed_on"] != "2026-09-30" {
		t.Fatalf("settlement keys: %v", st)
	}
	if locks, _ := s.ObjectsByType(ctx, "period_lock"); len(locks) != 1 || locks[0].State["month"] != "2026-09" || locks[0].State["book"] != "pl-stat" {
		t.Fatalf("period lock: %+v", locks)
	}
	var vatView struct {
		Rows [][]cell `json:"rows"`
	}
	get("/api/views/rozliczenie-vat", &vatView)
	if len(vatView.Rows) != 1 || cellVals(vatView.Rows)[0][0] != "2026-09" || cellVals(vatView.Rows)[0][5] != "115.00" {
		t.Fatalf("rozliczenie VAT view: %+v", vatView.Rows)
	}
	// Every number on the settlement is explained: the rule, the event, the
	// invoices it folded over.
	var walk walkOut
	get("/api/explain?object="+settlements[0].ID, &walk)
	// One fact, two derivations: the lock (the common rule) and the
	// settlement (the pack's); the settlement's arithmetic is baked with
	// the inputs it read, so the walk never recomputes.
	var settled *walkNode
	for i := range walk.Tree.Children {
		if walk.Tree.Children[i].RuleID == "pl-settle-vat" {
			settled = &walk.Tree.Children[i]
		}
	}
	if len(walk.Tree.Children) != 2 || settled == nil || settled.Object == nil {
		t.Fatalf("the settlement's provenance: %+v", walk.Tree)
	}
	if calc := settled.Object.Calc["vat_due"]; !strings.Contains(calc.Formula, "objects(sales_invoice, vat_period, month($.period))") || len(calc.Inputs) == 0 {
		t.Fatalf("the settlement's arithmetic is baked with its inputs: %+v", calc)
	}
	// The file: the market's writer renders JPK_V7M(3) from the settlement
	// and the month's invoices; the kernel renders it by name and hands it
	// over through the issue_document door — the bytes under their hash, the
	// event the evidence, an issued_document on the list.
	var jpk struct {
		Writers []string `json:"writers"`
	}
	get("/api/files", &jpk)
	if len(jpk.Writers) == 0 || jpk.Writers[0] != "jpk_v7m" {
		t.Fatalf("writers: %v", jpk.Writers)
	}
	res, err := ts.Client().Get(ts.URL + "/api/files/jpk_v7m?period=2026-09&created=2026-10-05T10:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	fileBytes, _ := io.ReadAll(res.Body)
	res.Body.Close()
	file := string(fileBytes)
	if res.StatusCode != 200 || res.Header.Get("Content-Type") != "application/xml" || !strings.Contains(res.Header.Get("Content-Disposition"), "JPK_V7M_5250001111_202609.xml") {
		t.Fatalf("GET the file: %d %v %s", res.StatusCode, res.Header, file[:200])
	}
	for _, want := range []string{
		`<KodFormularza kodSystemowy="JPK_V7M (3)" wersjaSchemy="1-0E">JPK_VAT</KodFormularza>`, `<KodUrzedu>1471</KodUrzedu>`, `<Rok>2026</Rok>`, `<Miesiac>9</Miesiac>`,
		`<NIP>5250001111</NIP>`, `<PelnaNazwa>Sprzedawca Sp. z o.o.</PelnaNazwa>`,
		`<P_19>1000</P_19>`, `<P_20>230</P_20>`, `<P_37>1000</P_37>`, `<P_38>230</P_38>`, `<P_42>500</P_42>`, `<P_43>115</P_43>`, `<P_48>115</P_48>`, `<P_51>115</P_51>`,
		`<DowodSprzedazy>FV 1/09/2026</DowodSprzedazy>`, `<NrKSeF>`, `<K_19>1000.00</K_19>`, `<K_20>230.00</K_20>`, `<PodatekNalezny>230.00</PodatekNalezny>`,
		`<BFK>1</BFK>`, `<K_42>500.00</K_42>`, `<K_43>115.00</K_43>`, `<PodatekNaliczony>115.00</PodatekNaliczony>`,
	} {
		if !strings.Contains(file, want) {
			t.Fatalf("the JPK lacks %s:\n%s", want, file)
		}
	}
	if strings.Contains(file, "<P_53>") || strings.Contains(file, "<P_62>") || strings.Contains(file, "<P_39>") || strings.Contains(file, "<P_10>") {
		t.Fatalf("empty positions are left out:\n%s", file)
	}
	var issued map[string]any
	post("/api/files/jpk_v7m/issue", map[string]any{"params": map[string]string{"period": "2026-09", "created": "2026-10-05T10:00:00Z"}, "actor": "krzysztof", "occurred_at": "2026-10-05"}, &issued)
	if issued["booked"].(float64) != 1 || issued["name"] != "JPK_V7M_5250001111_202609.xml" {
		t.Fatalf("issue: %v", issued)
	}
	docs, _ := s.ObjectsByType(ctx, "issued_document")
	if len(docs) != 1 || docs[0].State["writer"] != "jpk_v7m" || docs[0].State["params"] != "created=2026-10-05T10:00:00Z,period=2026-09" || docs[0].State["status"] != "issued" {
		t.Fatalf("issued document: %+v", docs)
	}
	blob, err := s.GetBlob(ctx, docs[0].State["hash"].(string))
	if err != nil || string(blob.Bytes) != file {
		t.Fatalf("the bytes under the hash are the file rendered: %v", err)
	}

	// After the return is filed, a September invoice turns up late. It is
	// never refused (DECISIONS 2026-10-16, late documents): nothing posts
	// into the closed September — its entries land in October — while its
	// VAT still belongs to September, the month of the sale. September's
	// records and its filed declaration now disagree, so the writer no longer
	// renders the original: what is owed is a correction of the return.
	post("/api/events", map[string]any{
		"event_type": "sales.invoice.issued", "occurred_at": "2026-09-25", "dedup_key": "fv-late",
		"payload": map[string]any{
			"number": "FV 3/09/2026", "issue_date": "2026-09-25", "market": "pl", "intragroup": "no",
			"seller_nip": "5250001111", "buyer_nip": "5260001246",
			"net": "100.00", "vat": "23.00", "gross": "123.00", "vat_rate": "23", "currency": "PLN",
		},
	}, &resp)
	if resp["booked"].(float64) != 1 {
		t.Fatalf("the late invoice: %v", resp)
	}
	for _, p := range func() []core.Object { ps, _ := s.ObjectsByType(ctx, "posting"); return ps }() {
		if p.SourceEventID == int64(resp["event_id"].(float64)) && p.State["date"] != "2026-10-01" {
			t.Fatalf("the late invoice's entry is not in October: %v", p.State)
		}
	}
	for _, inv := range func() []core.Object { is, _ := s.ObjectsByType(ctx, "sales_invoice"); return is }() {
		if inv.State["number"] == "FV 3/09/2026" && inv.State["vat_period"] != "2026-09" {
			t.Fatalf("the late invoice's VAT month: %v", inv.State)
		}
	}
	res, err = ts.Client().Get(ts.URL + "/api/files/jpk_v7m?period=2026-09&created=2026-10-25T10:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	refusal, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode == 200 || !strings.Contains(string(refusal), "correction") {
		t.Fatalf("the original after a late invoice: %d %s", res.StatusCode, refusal)
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
