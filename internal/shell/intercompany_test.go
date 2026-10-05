package shell_test

// E4 of the enterprise-structure ladder (DIRECTION): intercompany. The
// attempt failed on one seam — a ref cannot be re-referenced (DECISIONS
// 2026-10-04, git history) — and the fix is data: sales_invoice v3 echoes
// intragroup and the resolution keys as plain fields, so the mirror rules
// re-resolve by value. Nothing earned; another negative proof. One raw
// event — Alfa invoices Beta in EUR — explains both entities end to end:
// Alfa's document and its PLN entry (E3 converting at the NBP rate), then
// by cascade Beta's purchase document and its EUR entry in the German
// book, all atomic, all provenance-chained to the one root event. That is
// the killer demo: intercompany reconciliation matches BY CONSTRUCTION,
// because both sides of the position share a cause event id — there is
// nothing to reconcile, only to display. Transfer pricing is just the
// rule that prices the derived side; here the mirror books at invoice
// amounts.

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

func TestIntercompany(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	for _, f := range []string{"finance.json", "finance_v2.json", "finance_v3.json",
		"finance_v4.json", "finance_v5.json", "finance_v6.json", "finance_v7.json",
		"finance_v8.json", "finance_v9.json"} {
		seedFromFile(t, s, f)
	}

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
			t.Fatalf("GET %s: %d", path, res.StatusCode)
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
			ID: id, Status: core.StatusDraft, Priority: 150,
			EffectiveFrom: "2026-01-01", CreatedBy: "human", Description: id, Spec: spec,
		}); err != nil {
			t.Fatal(err)
		}
		post("/api/rules/"+id+"/approve", map[string]any{"approved_by": "krzysztof"}, nil)
	}

	for _, p := range []string{"pl", "de"} {
		if _, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", p, "pack.json"), "test"); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{"pl-register-account", "pl-register-vat-rate",
		"pl-book-sales-invoice", "pl-post-sales-invoice",
		"de-register-account", "de-register-vat-rate",
		"de-book-sales-invoice", "de-post-sales-invoice"} {
		post("/api/rules/"+id+"/approve", map[string]any{"approved_by": "krzysztof"}, nil)
	}
	activate("register-company", core.RuleSpec{
		Match: core.Match{EventType: "company.registered"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "company",
			Fields: map[string]string{"name": "=$.name", "kind": "=$.kind",
				"vat_id": "=$.vat_id", "country": "=$.country"}}},
	})
	activate("register-fx-rate", core.RuleSpec{
		Match: core.Match{EventType: "fx.rate.published"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "fx_rate",
			Fields: map[string]string{"code": "=$.code", "base": "=$.base",
				"quote": "=$.quote", "date": "=$.date", "rate": "=$.rate"}}},
	})
	for dedup, c := range map[string]map[string]any{
		"alfa":     {"name": "Alfa Sp. z o.o.", "kind": "self", "vat_id": "5250001111", "country": "PL"},
		"beta":     {"name": "Beta GmbH", "kind": "self", "vat_id": "DE811111111", "country": "DE"},
		"buyer-de": {"name": "Käufer AG", "kind": "customer", "vat_id": "DE123456789", "country": "DE"},
	} {
		post("/api/events", map[string]any{
			"event_type": "company.registered", "occurred_at": "2026-09-01",
			"dedup_key": dedup, "payload": c}, nil)
	}
	post("/api/events", map[string]any{
		"event_type": "fx.rate.published", "occurred_at": "2026-10-05", "dedup_key": "nbp-1005",
		"payload": map[string]any{"code": "EUR/PLN/2026-10-05", "base": "EUR",
			"quote": "PLN", "date": "2026-10-05", "rate": "4.30"}}, nil)

	// The group's mirror rules — what a phase-(a) agent would draft for a
	// subsidiary pair. The purchase re-resolves both companies by the
	// echoed keys, and references the sales invoice it mirrors.
	activate("ic-mirror-purchase", core.RuleSpec{
		Match: core.Match{EventType: core.EventObjectMaterialized, Where: []core.Condition{
			{Path: "$.object_type", Op: "eq", Value: "sales_invoice"},
			{Path: "$.state.intragroup", Op: "eq", Value: "yes"}}},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "purchase_invoice",
			Fields: map[string]string{
				"number": "=$.state.number", "invoice_date": "=$.state.issue_date",
				"supplier":  "=ref(company, vat_id, $.state.seller_nip)",
				"owner":     "=ref(company, vat_id, $.state.buyer_nip)",
				"mirror_of": "=ref(sales_invoice, number, $.state.number)",
				"net":       "=$.state.net", "vat": "=$.state.vat", "gross": "=$.state.gross",
				"currency": "=$.state.currency",
			}}},
	})
	activate("ic-mirror-post", core.RuleSpec{
		Match: core.Match{EventType: core.EventObjectMaterialized, Where: []core.Condition{
			{Path: "$.object_type", Op: "eq", Value: "purchase_invoice"}}},
		Effect: core.Effect{Postings: &core.PostingsTemplate{
			Book: "de-stat", Currency: "=$.state.currency",
			Convert: &core.ConvertTemplate{
				To: "EUR", Date: "=$.state.invoice_date", Rounding: "half_up"},
			Lines: []core.PostingLine{
				{Account: "3400", Debit: "=$.state.net"},
				{Account: "1576", Debit: "=$.state.vat"},
				{Account: "1600", Credit: "=$.state.gross"},
			}}},
	})

	// One raw event: Alfa invoices Beta, in EUR, intragroup declared.
	var resp map[string]any
	post("/api/events", map[string]any{
		"event_type": "sales.invoice.issued", "occurred_at": "2026-10-05", "dedup_key": "fv-ic1",
		"payload": map[string]any{
			"number": "FV IC/1", "issue_date": "2026-10-05", "market": "pl",
			"intragroup": "yes", "seller_nip": "5250001111", "buyer_nip": "DE811111111",
			"net": "1000.00", "vat": "230.00", "gross": "1230.00",
			"vat_rate": "23", "currency": "EUR",
		},
	}, &resp)
	if resp["booked"].(float64) != 1 || resp["errors"] != nil {
		t.Fatalf("ic invoice: %v", resp)
	}

	// Both entities explained from the one event: Alfa's document + PLN
	// entry, Beta's mirror purchase + EUR entry.
	alfa, _ := s.FindObjectIDsByField(ctx, "company", "vat_id", "5250001111")
	beta, _ := s.FindObjectIDsByField(ctx, "company", "vat_id", "DE811111111")
	sales, _ := s.ObjectsByType(ctx, "sales_invoice")
	purchases, _ := s.ObjectsByType(ctx, "purchase_invoice")
	if len(sales) != 1 || len(purchases) != 1 {
		t.Fatalf("documents: %d sales, %d purchases", len(sales), len(purchases))
	}
	p := purchases[0]
	if p.State["supplier"] != alfa[0] || p.State["owner"] != beta[0] ||
		p.State["mirror_of"] != sales[0].ID ||
		int64(p.State["gross"].(float64)) != 123000 || p.State["currency"] != "EUR" {
		t.Fatalf("purchase = %v", p.State)
	}
	wantPostings := map[string][3]any{ // account code → book, side, amount
		"201":  {"pl-stat", "debit", int64(528900)}, // 1230.00 EUR × 4.30 → PLN
		"700":  {"pl-stat", "credit", int64(430000)},
		"222":  {"pl-stat", "credit", int64(98900)},
		"3400": {"de-stat", "debit", int64(100000)}, // Beta's book stays EUR
		"1576": {"de-stat", "debit", int64(23000)},
		"1600": {"de-stat", "credit", int64(123000)},
	}
	postings, _ := s.ObjectsByType(ctx, "posting")
	if len(postings) != 6 {
		t.Fatalf("postings = %d", len(postings))
	}
	for _, po := range postings {
		acc, _ := s.GetObject(ctx, po.State["account"].(string))
		w, ok := wantPostings[acc.State["code"].(string)]
		if !ok {
			t.Fatalf("unexpected posting on %v", acc.State["code"])
		}
		if po.State["book"] != w[0] || po.State["side"] != w[1] ||
			int64(po.State["amount"].(float64)) != w[2].(int64) {
			t.Fatalf("account %s: %v, want %v", acc.State["code"], po.State, w)
		}
	}

	// The killer demo, asserted: every object on both sides walks back to
	// the SAME root raw event. Beta's payable and Alfa's receivable cannot
	// disagree, because they are two explanations of one fact.
	derived, _ := s.EventsByKind(ctx, core.KindDerived)
	evByID := map[int64]core.Event{}
	matEv := map[string]core.Event{} // object id → its materialization event
	for _, ev := range derived {
		evByID[ev.ID] = ev
		var mat core.MaterializedObject
		if json.Unmarshal(ev.Payload, &mat) == nil {
			matEv[mat.ObjectID] = ev
		}
	}
	rootOf := func(objectID string) int64 {
		t.Helper()
		ev, ok := matEv[objectID]
		if !ok {
			t.Fatalf("no materialization event for %s", objectID)
		}
		for {
			cause, ok := evByID[*ev.CauseEventID]
			if !ok {
				return *ev.CauseEventID // a raw event: the root
			}
			ev = cause
		}
	}
	root := rootOf(sales[0].ID)
	if rootOf(p.ID) != root {
		t.Fatalf("purchase chains to %d, sales to %d", rootOf(p.ID), root)
	}
	for _, po := range postings {
		if rootOf(po.ID) != root {
			t.Fatalf("posting %s chains to %d, not the shared root %d", po.ID, rootOf(po.ID), root)
		}
	}
	// And the hop structure is the cascade itself: the purchase's direct
	// cause is the sales invoice's materialization event.
	if *matEv[p.ID].CauseEventID != matEv[sales[0].ID].ID {
		t.Fatal("the purchase is not caused by the sales invoice's materialization")
	}

	// An ordinary (non-intragroup) invoice raises no mirror. Its amounts
	// are chosen to translate cleanly at the closing rate below.
	post("/api/events", map[string]any{
		"event_type": "sales.invoice.issued", "occurred_at": "2026-10-06", "dedup_key": "fv-n1",
		"payload": map[string]any{
			"number": "FV 2/10/2026", "issue_date": "2026-10-06", "market": "pl",
			"intragroup": "no", "seller_nip": "5250001111", "buyer_nip": "DE123456789",
			"net": "410.00", "vat": "94.30", "gross": "504.30",
			"vat_rate": "23", "currency": "PLN",
		},
	}, &resp)
	if resp["booked"].(float64) != 1 {
		t.Fatalf("plain invoice: %v", resp)
	}
	if purchases, _ := s.ObjectsByType(ctx, "purchase_invoice"); len(purchases) != 1 {
		t.Fatalf("a plain invoice raised a mirror: %d purchases", len(purchases))
	}
	// Beta trades with third parties too.
	post("/api/events", map[string]any{
		"event_type": "sales.invoice.issued", "occurred_at": "2026-10-06", "dedup_key": "re-b1",
		"payload": map[string]any{
			"number": "RE 2026-003", "issue_date": "2026-10-06", "market": "de",
			"seller_nip": "DE811111111", "buyer_nip": "DE123456789",
			"net": "200.00", "vat": "38.00", "gross": "238.00",
			"vat_rate": "19", "currency": "EUR",
		},
	}, &resp)
	if resp["booked"].(float64) != 1 {
		t.Fatalf("beta invoice: %v", resp)
	}

	// One trial balance, the group position visible: Alfa's receivable in
	// PLN, Beta's payable in EUR, both from the one event.
	var tb struct {
		Rows [][]cell `json:"rows"`
	}
	get("/api/views/trial-balance", &tb)
	want := map[string][3]string{
		"201":  {"pl-stat", "5793.30", "0.00"}, // 5289.00 + 504.30
		"700":  {"pl-stat", "0.00", "4710.00"},
		"222":  {"pl-stat", "0.00", "1083.30"},
		"3400": {"de-stat", "1000.00", "0.00"},
		"1576": {"de-stat", "230.00", "0.00"},
		"1600": {"de-stat", "0.00", "1230.00"},
		"1400": {"de-stat", "238.00", "0.00"},
		"8400": {"de-stat", "0.00", "200.00"},
		"1776": {"de-stat", "0.00", "38.00"},
	}
	seen := 0
	for _, r := range cellVals(tb.Rows) {
		if w, ok := want[r[1]]; ok {
			if r[0] != w[0] || r[3] != w[1] || r[4] != w[2] {
				t.Fatalf("account %s: %v, want %v", r[1], r, w)
			}
			seen++
		}
	}
	if seen != 9 {
		t.Fatalf("trial balance = %+v", tb.Rows)
	}

	// ——— E5: consolidation. The group is one more explanation, living on
	// the analysis side: eliminations match on provenance, not heuristics.

	// The intercompany-positions view is the killer demo on a screen: both
	// sides of the position, matched by their shared root event, at their
	// original transaction amounts. The difference is zero by construction.
	post("/api/events", map[string]any{
		"event_type": "fx.rate.published", "occurred_at": "2026-10-31", "dedup_key": "nbp-1031",
		"payload": map[string]any{"code": "EUR/PLN/2026-10-31", "base": "EUR",
			"quote": "PLN", "date": "2026-10-31", "rate": "4.10"}}, nil)
	var ic struct {
		Rows [][]cell `json:"rows"`
	}
	get("/api/views/intercompany-positions", &ic)
	if len(ic.Rows) != 1 {
		t.Fatalf("positions = %+v", ic.Rows)
	}
	if r := cellVals(ic.Rows)[0]; r[0] != fmt.Sprint(root) || r[1] != "1230.00" || r[2] != "1230.00" || r[3] != "0.00" {
		t.Fatalf("position row = %v (root %d)", r, root)
	}

	// The group trial balance in EUR: PLN translates at the latest rate
	// (4.10), intercompany positions are eliminated by root — receivable,
	// payable, revenue and cost vanish, each entity's VAT against its tax
	// office survives — and the translation residue (the IC entry booked at
	// 4.30, translated at 4.10) is a visible CTA line, not a hidden leak.
	var gtb struct {
		Rows [][]cell `json:"rows"`
	}
	get("/api/views/group-trial-balance", &gtb)
	wantGroup := [][]string{
		{"1400", "Forderungen aus Lieferungen und Leistungen", "238.00", "0.00", "238.00"},
		{"1576", "Abziehbare Vorsteuer 19%", "230.00", "0.00", "230.00"},
		{"1776", "Umsatzsteuer 19%", "0.00", "38.00", "-38.00"},
		{"201", "Rozrachunki z odbiorcami", "123.00", "0.00", "123.00"},
		{"222", "VAT należny", "0.00", "264.22", "-264.22"},
		{"700", "Sprzedaż produktów", "0.00", "100.00", "-100.00"},
		{"8400", "Erlöse 19% USt", "0.00", "200.00", "-200.00"},
		{"CTA", "Translation difference", "11.22", "0.00", "11.22"},
	}
	if !reflect.DeepEqual(cellVals(gtb.Rows), wantGroup) {
		t.Fatalf("group trial balance:\n got %v\nwant %v", cellVals(gtb.Rows), wantGroup)
	}

	// One replay, two entities, identical state — the intercompany chain
	// reproduces like everything else.
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
