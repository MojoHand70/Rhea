package shell_test

// E4 of the enterprise-structure ladder (DIRECTION): intercompany — entity
// A's sale raises entity B's purchase by cascade, provenance crossing the
// boundary. This file is the recorded fail-as-data attempt (DECISIONS
// 2026-10-04). Two jaws of one gap: a ref cannot be re-referenced. The
// invoice's state carries seller/buyer as resolved object ids, and (1) a
// direct-id template into the purchase's supplier ref is rejected by design
// — referential integrity only through resolution — while (2) resolving by
// value fails, because `=ref(company, vat_id, $.state.seller)` hands
// vat_id an object id. And the chain is atomic, so the broken mirror
// refuses the whole root event: Alfa's own document does not book either.
// A side observation with teeth: the invoice object cannot even say it is
// intragroup — nothing carries the flag — so the mirror here is keyed to a
// literal invoice number, an absurdity that only sharpens the verdict.

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

func TestIntercompany(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	for _, f := range []string{"finance.json", "finance_v2.json", "finance_v3.json",
		"finance_v4.json", "finance_v5.json", "finance_v6.json", "finance_v7.json"} {
		seedFromFile(t, s, f)
	}
	if err := s.InsertObjectType(ctx, core.ObjectType{
		Name: "purchase_invoice", Version: 1, Domain: "finance", IsDocument: true,
		LabelField: "number",
		Fields: []core.FieldDef{
			{Name: "number", Type: "string", Required: true},
			{Name: "invoice_date", Type: "date", Required: true},
			{Name: "supplier", Type: "ref<company>", Required: true},
			{Name: "owner", Type: "ref<company>", Required: true},
			{Name: "mirror_of", Type: "ref<sales_invoice>", Required: true},
			{Name: "net", Type: "money", Required: true},
			{Name: "vat", Type: "money", Required: true},
			{Name: "gross", Type: "money", Required: true},
			{Name: "currency", Type: "string", Required: true},
		},
	}); err != nil {
		t.Fatal(err)
	}

	x := &exec.Executor{Store: s}
	srv := &shell.Server{
		Store: s, Exec: x,
		Agent:    &agent.Agent{Complete: func(context.Context, string, string) (string, error) { return "", fmt.Errorf("no agent in this test") }},
		DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb"),
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

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
		"de-register-account", "de-register-vat-rate"} {
		post("/api/rules/"+id+"/approve", map[string]any{"approved_by": "krzysztof"}, nil)
	}
	activate("register-company", core.RuleSpec{
		Match: core.Match{EventType: "company.registered"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "company",
			Fields: map[string]string{"name": "=$.name", "kind": "=$.kind",
				"vat_id": "=$.vat_id", "country": "=$.country"}}},
	})
	for dedup, c := range map[string]map[string]any{
		"alfa": {"name": "Alfa Sp. z o.o.", "kind": "self", "vat_id": "5250001111", "country": "PL"},
		"beta": {"name": "Beta GmbH", "kind": "self", "vat_id": "DE811111111", "country": "DE"},
	} {
		post("/api/events", map[string]any{
			"event_type": "company.registered", "occurred_at": "2026-09-01",
			"dedup_key": dedup, "payload": c}, nil)
	}

	// Jaw one: the direct-id template is rejected at validation, by design.
	purchaseType, err := s.GetObjectType(ctx, "purchase_invoice")
	if err != nil {
		t.Fatal(err)
	}
	direct := core.RuleSpec{
		Match: core.Match{EventType: core.EventObjectMaterialized, Where: []core.Condition{
			{Path: "$.object_type", Op: "eq", Value: "sales_invoice"}}},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "purchase_invoice",
			Fields: map[string]string{
				"number": "=$.state.number", "invoice_date": "=$.state.issue_date",
				"supplier": "=$.state.seller", "owner": "=$.state.buyer",
				"mirror_of": "=ref(sales_invoice, number, $.state.number)",
				"net":       "=$.state.net", "vat": "=$.state.vat", "gross": "=$.state.gross",
				"currency": "=$.state.currency",
			}}},
	}
	if err := direct.Validate(&purchaseType); err == nil ||
		!strings.Contains(err.Error(), "must use =ref(company") {
		t.Fatalf("direct-id template: %v", err)
	}

	// Jaw two: resolving by value hands vat_id an object id. The mirror is
	// keyed to a literal invoice number because nothing on the invoice can
	// say "intragroup".
	mirror := direct
	mirror.Match.Where = append(mirror.Match.Where,
		core.Condition{Path: "$.state.number", Op: "eq", Value: "FV IC/1"})
	mirror.Effect.Object.Fields["supplier"] = "=ref(company, vat_id, $.state.seller)"
	mirror.Effect.Object.Fields["owner"] = "=ref(company, vat_id, $.state.buyer)"
	activate("ic-mirror-purchase", mirror)

	var resp map[string]any
	post("/api/events", map[string]any{
		"event_type": "sales.invoice.issued", "occurred_at": "2026-09-21", "dedup_key": "fv-ic1",
		"payload": map[string]any{
			"number": "FV IC/1", "issue_date": "2026-09-21", "market": "pl",
			"seller_nip": "5250001111", "buyer_nip": "DE811111111",
			"net": "1000.00", "vat": "230.00", "gross": "1230.00",
			"vat_rate": "23", "currency": "PLN",
		},
	}, &resp)
	if resp["booked"].(float64) != 0 ||
		!strings.Contains(fmt.Sprint(resp["errors"]), "no company with vat_id = company-") {
		t.Fatalf("expected the ref re-reference to fail: %v", resp)
	}
	// Atomicity makes the gap total: the broken mirror refuses the whole
	// chain, so Alfa's own document does not book either. Never
	// half-explained across entities — but intercompany stays inexpressible.
	if docs, _ := s.ObjectsByType(ctx, "sales_invoice"); len(docs) != 0 {
		t.Fatalf("sales invoice booked despite the broken mirror: %d", len(docs))
	}
}
