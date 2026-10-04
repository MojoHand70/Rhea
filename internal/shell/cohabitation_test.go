package shell_test

// E2 of the enterprise-structure ladder (DIRECTION): two operating companies,
// one under the PL pack, one under DE, in one kernel over one log. This file
// is the recorded fail-as-data attempt (DECISIONS 2026-10-04), and the
// failure is worse than a conflict — it is silent, twice. First: packs
// register master data from the same neutral event types with no market
// binding, and approvals process eagerly, so whichever market's
// registration rule is approved first captures the other market's chart of
// accounts too — the German SKR03 materializes with `pl-register-account`
// as its explanation, provenance recording the mis-explanation honestly,
// nothing flagging it. Second: the packs route invoices by currency
// (`$.currency eq PLN`/`EUR`), which keeps their matches disjoint — so the
// same-id conflict that would at least refuse loudly never fires, and a
// Polish company's EUR-denominated invoice books into the German SKR03
// instead. Currency is not a market, and it is certainly not a company.
// The one loud failure: both packs ship a VAT rate coded "0", and ref
// resolution is global, so a 0% invoice cannot resolve its rate at all.

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

	x := &exec.Executor{Store: s}
	srv := &shell.Server{
		Store: s, Exec: x,
		Agent:    &agent.Agent{Complete: func(context.Context, string, string) (string, error) { return "", fmt.Errorf("no agent in this test") }},
		DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb"),
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	post := func(path string, body any) {
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
	}

	// Both markets into one kernel: the loads themselves succeed — packs are
	// data, and data merges.
	for _, p := range []string{"pl", "de"} {
		if _, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", p, "pack.json"), "test"); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{"pl-register-account", "pl-register-vat-rate",
		"pl-book-sales-invoice", "pl-post-sales-invoice", "pl-record-ksef-submission",
		"de-register-account", "de-register-vat-rate",
		"de-book-sales-invoice", "de-post-sales-invoice"} {
		post("/api/rules/"+id+"/approve", map[string]any{"approved_by": "krzysztof"})
	}

	// Buyers are ordinary base master data (as in the single-market tests).
	if _, err := s.InsertRuleVersion(ctx, core.Rule{
		ID: "register-company", Status: core.StatusDraft, Priority: 10,
		EffectiveFrom: "2026-01-01", CreatedBy: "human", Description: "register-company",
		Spec: core.RuleSpec{
			Match: core.Match{EventType: "company.registered"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "company",
				Fields: map[string]string{"name": "=$.name", "kind": "=$.kind",
					"vat_id": "=$.vat_id", "country": "=$.country"}}},
		},
	}); err != nil {
		t.Fatal(err)
	}
	post("/api/rules/register-company/approve", map[string]any{"approved_by": "krzysztof"})
	post("/api/events", map[string]any{
		"event_type": "company.registered", "occurred_at": "2026-09-01", "dedup_key": "buyer-pl",
		"payload": map[string]any{"name": "Nabywca S.A.", "kind": "customer",
			"vat_id": "5260001246", "country": "PL"}})
	post("/api/events", map[string]any{
		"event_type": "company.registered", "occurred_at": "2026-09-01", "dedup_key": "buyer-de",
		"payload": map[string]any{"name": "Käufer AG", "kind": "customer",
			"vat_id": "DE129273398", "country": "DE"}})

	// The silent failure: pl-register-account was approved first, so its
	// eager re-evaluation captured ALL 39 account events — the German SKR03
	// included — before de-register-account existed. No conflict, no error,
	// a clean worklist, and the wrong explanation recorded with perfect
	// provenance.
	if wl, _ := s.UnmatchedRawEvents(ctx); len(wl) != 0 {
		t.Fatalf("worklist = %d, want 0 — the capture is silent", len(wl))
	}
	accs, err := s.ObjectsByType(ctx, "account")
	if err != nil || len(accs) != 39 {
		t.Fatalf("accounts = %d (%v), want all 39 of both markets", len(accs), err)
	}
	captured := 0
	for _, a := range accs {
		if code, _ := a.State["code"].(string); strings.HasPrefix(code, "0") || len(code) == 4 {
			// a German SKR03 account, explained by the Polish rule
			if a.RuleID == "pl-register-account" {
				captured++
			}
		}
	}
	if captured == 0 {
		t.Fatal("expected the Polish rule to have captured German accounts")
	}

	// The second silent failure: a Polish company issues a domestic invoice
	// denominated in EUR (legal and ordinary). Currency-as-market routes it
	// to the German pack: it books — cleanly, provenance and all — as a
	// German invoice, VAT and revenue on the SKR03 accounts.
	post("/api/events", map[string]any{
		"event_type": "sales.invoice.issued", "occurred_at": "2026-09-21", "dedup_key": "fv-eur",
		"payload": map[string]any{
			"number": "FV 3/09/2026", "issue_date": "2026-09-21", "buyer_nip": "5260001246",
			"net": "1000.00", "vat": "230.00", "gross": "1230.00",
			"vat_rate": "23", "currency": "EUR",
		},
	})
	invoices, err := s.ObjectsByType(ctx, "sales_invoice")
	if err != nil || len(invoices) != 1 {
		t.Fatalf("invoices = %d (%v), want the EUR invoice booked", len(invoices), err)
	}
	if invoices[0].RuleID != "de-book-sales-invoice" {
		t.Fatalf("invoice explained by %s — expected the silent German capture", invoices[0].RuleID)
	}
	id1400, err := s.FindObjectIDsByField(ctx, "account", "code", "1400")
	if err != nil || len(id1400) != 1 {
		t.Fatalf("SKR03 1400: %v %v", id1400, err)
	}
	var onSKR03 bool
	postings, _ := s.ObjectsByType(ctx, "posting")
	for _, p := range postings {
		onSKR03 = onSKR03 || p.State["account"] == id1400[0]
	}
	if !onSKR03 {
		t.Fatal("expected the Polish EUR invoice posted to German Forderungen 1400")
	}

	// The loud failure, second form: a 0% PLN invoice matches only Polish
	// rules, but both packs shipped a VAT rate coded "0" and ref resolution
	// is global — the rate is ambiguous and the event waits forever.
	post("/api/events", map[string]any{
		"event_type": "sales.invoice.issued", "occurred_at": "2026-09-22", "dedup_key": "fv-0",
		"payload": map[string]any{
			"number": "FV 2/09/2026", "issue_date": "2026-09-22", "buyer_nip": "5260001246",
			"net": "500.00", "vat": "0.00", "gross": "500.00",
			"vat_rate": "0", "currency": "PLN",
		},
	})
	n, errs := x.ProcessPending(ctx)
	if n != 0 || len(errs) == 0 {
		t.Fatalf("0%% invoice: booked %d, errs %v", n, errs)
	}
	var ambiguous bool
	for _, e := range errs {
		ambiguous = ambiguous || strings.Contains(e.Error(), "ref is ambiguous")
	}
	if !ambiguous {
		t.Fatalf("expected an ambiguous vat_rate ref, got %v", errs)
	}
}
