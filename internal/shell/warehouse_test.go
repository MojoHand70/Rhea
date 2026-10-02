package shell_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

// TestM2WarehouseWithoutKernelChanges is the falsifiability test (SPEC M2):
// a genuinely distant domain — locations, items, goods receipts and issues,
// stock — expressed purely as data. Object types, views and rules are loaded
// at runtime; the kernel is exactly the one finance runs on. If this test
// had required kernel changes, the language would have failed.
func TestM2WarehouseWithoutKernelChanges(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	seedFromFile(t, s, "finance.json")    // invoice v1 (unused here, realistic base)
	seedFromFile(t, s, "finance_v2.json") // company master data
	seedFromFile(t, s, "warehouse.json")  // the new domain, as data

	// No drafting in this test: rules are authored by hand and go through
	// the same validation and approval path (SPEC §2: the kernel cannot
	// tell the difference).
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

	// The warehouse rule set, hand-authored. Approval is the same human
	// activity as everywhere, through the API.
	obj := func(typ string, fields map[string]string) core.Effect {
		return core.Effect{Object: core.ObjectTemplate{Type: typ, Fields: fields}}
	}
	rules := []struct {
		id        string
		eventType string
		effect    core.Effect
	}{
		{"register-company", "company.registered", obj("company", map[string]string{
			"name": "=$.name", "kind": "=$.kind", "vat_id": "=$.vat_id", "country": "=$.country"})},
		{"register-location", "location.created", obj("location", map[string]string{
			"code": "=$.code", "name": "=$.name"})},
		{"register-item", "item.created", obj("item", map[string]string{
			"sku": "=$.sku", "name": "=$.name", "unit": "=$.unit"})},
		{"book-goods-receipt", "goods.received", obj("goods_receipt", map[string]string{
			"supplier": "=ref(company, name, $.supplier)", "item": "=ref(item, sku, $.item)",
			"location": "=ref(location, code, $.location)", "qty": "=$.qty", "date": "=$.date"})},
		{"move-stock-in", "goods.received", obj("stock_movement", map[string]string{
			"item": "=ref(item, sku, $.item)", "location": "=ref(location, code, $.location)",
			"direction": "in", "qty": "=$.qty", "date": "=$.date"})},
		{"book-goods-issue", "goods.issued", obj("goods_issue", map[string]string{
			"item": "=ref(item, sku, $.item)", "location": "=ref(location, code, $.location)",
			"qty": "=$.qty", "date": "=$.date", "reason": "=$.reason"})},
		{"move-stock-out", "goods.issued", obj("stock_movement", map[string]string{
			"item": "=ref(item, sku, $.item)", "location": "=ref(location, code, $.location)",
			"direction": "out", "qty": "=$.qty", "date": "=$.date"})},
	}
	for _, r := range rules {
		if _, err := s.InsertRuleVersion(ctx, core.Rule{
			ID: r.id, Status: core.StatusDraft, Priority: 100,
			EffectiveFrom: "2026-09-01", CreatedBy: "human", Description: r.id,
			Spec: core.RuleSpec{Match: core.Match{EventType: r.eventType}, Effect: r.effect},
		}); err != nil {
			t.Fatalf("rule %s: %v", r.id, err)
		}
		post("/api/rules/"+r.id+"/approve", map[string]any{"approved_by": "krzysztof"}, nil)
	}

	// The navigation now carries a second domain, rendered with zero code.
	var nav []struct {
		Domain string `json:"domain"`
	}
	get("/api/nav", &nav)
	if len(nav) != 2 || nav[1].Domain != "warehouse" && nav[0].Domain != "warehouse" {
		t.Fatalf("nav = %+v, want finance and warehouse", nav)
	}

	// Master data, then goods events: a receipt fires two rules atomically.
	submit := func(file string) map[string]any {
		t.Helper()
		b, err := os.ReadFile(filepath.Join("..", "..", "testdata", "events", file))
		if err != nil {
			t.Fatal(err)
		}
		var ev map[string]any
		json.Unmarshal(b, &ev)
		ev["dedup_key"] = file
		var resp map[string]any
		post("/api/events", ev, &resp)
		return resp
	}
	for _, f := range []string{"company_omikron.json", "location_main.json", "item_widget.json"} {
		if resp := submit(f); resp["booked"].(float64) != 1 {
			t.Fatalf("%s: %v", f, resp)
		}
	}
	if resp := submit("goods_received_widgets.json"); resp["booked"].(float64) != 1 || resp["errors"] != nil {
		t.Fatalf("receipt: %v", resp)
	}
	if resp := submit("goods_issued_widgets.json"); resp["booked"].(float64) != 1 || resp["errors"] != nil {
		t.Fatalf("issue: %v", resp)
	}

	// Multi-fire provenance: the one receipt event produced a document and a
	// movement, each explained by its own rule.
	receipts, _ := s.ObjectsByType(ctx, "goods_receipt")
	movements, _ := s.ObjectsByType(ctx, "stock_movement")
	if len(receipts) != 1 || len(movements) != 2 {
		t.Fatalf("receipts %d, movements %d", len(receipts), len(movements))
	}
	if receipts[0].RuleID != "book-goods-receipt" {
		t.Fatalf("receipt provenance: %s", receipts[0].RuleID)
	}
	if receipts[0].SourceEventID != movements[0].SourceEventID {
		t.Fatalf("document and movement from different events")
	}

	// Views resolve cross-domain refs: the receipt names its supplier from
	// the finance domain and goods from the warehouse domain, as labels.
	var list struct {
		Rows [][]string `json:"rows"`
	}
	get("/api/views/goods-receipt-list", &list)
	if len(list.Rows) != 1 {
		t.Fatalf("receipt list = %+v", list.Rows)
	}
	row := list.Rows[0]
	if row[1] != "Omikron GmbH" || row[2] != "Widget" || row[3] != "Magazyn główny" || row[4] != "25" {
		t.Fatalf("receipt row = %v", row)
	}

	// The analysis side aggregates stock on hand: 25 in − 8 out = 17.
	var stock struct {
		Rows [][]string `json:"rows"`
	}
	get("/api/views/stock-on-hand", &stock)
	if len(stock.Rows) != 1 {
		t.Fatalf("stock = %+v", stock.Rows)
	}
	if r := stock.Rows[0]; r[0] != "WID-1" || r[2] != "MAIN" || r[3] != "17" {
		t.Fatalf("stock row = %v", r)
	}

	// Determinism holds across domains.
	x := &exec.Executor{Store: s}
	before, _ := s.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := s.AllObjects(ctx)
	bj, _ := json.Marshal(before)
	aj, _ := json.Marshal(after)
	if string(bj) != string(aj) {
		t.Fatal("replay diverged")
	}
}
