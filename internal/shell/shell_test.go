package shell_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/shell"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

// TestM0DemoStory drives the whole vertical slice through the HTTP API:
// seed definitions → submit event → worklist → agent drafts → approve →
// list/detail/analysis views show the booked document.
func TestM0DemoStory(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	seedFromFile(t, s, "finance.json")

	// The agent is the recorded kind: deterministic test, same validation path.
	fixture := &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
		if !strings.Contains(user, "ACME") {
			return "", fmt.Errorf("sample event not in prompt")
		}
		return `{
			"rule_id": "book-pln-invoice",
			"description": "PLN invoices become invoice documents with a summed total.",
			"priority": 100,
			"spec": {
				"match": {"event_type": "invoice.received",
				          "where": [{"path": "$.currency", "op": "eq", "value": "PLN"}]},
				"effect": {"object": {"type": "invoice", "fields": {
					"customer": "=$.customer", "issue_date": "=$.issue_date",
					"currency": "=$.currency", "total": "=sum($.lines[*].amount)"}}}
			}
		}`, nil
	}}

	srv := &shell.Server{
		Store: s, Exec: &exec.Executor{Store: s}, Agent: fixture,
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

	// 1. Navigation: finance domain with documents/analysis + system functions.
	var nav []struct {
		Domain    string `json:"domain"`
		Functions []struct {
			Function string `json:"function"`
		} `json:"functions"`
	}
	get("/api/nav", &nav)
	if len(nav) != 1 || nav[0].Domain != "finance" {
		t.Fatalf("nav = %+v", nav)
	}

	// 2. Submit two PLN invoices and one EUR; nothing books (no rules).
	submitEvent := func(file string) {
		t.Helper()
		b, err := os.ReadFile(filepath.Join("..", "..", "testdata", "events", file))
		if err != nil {
			t.Fatal(err)
		}
		var ev map[string]any
		json.Unmarshal(b, &ev)
		ev["dedup_key"] = file
		post("/api/events", ev, nil)
	}
	submitEvent("invoice_acme.json")
	submitEvent("invoice_beta.json")
	submitEvent("invoice_gamma_eur.json")

	var wl struct {
		Events      []core.Event `json:"events"`
		ObjectTypes []string     `json:"object_types"`
	}
	get("/api/worklist", &wl)
	if len(wl.Events) != 3 || len(wl.ObjectTypes) != 1 {
		t.Fatalf("worklist: %d events, types %v", len(wl.Events), wl.ObjectTypes)
	}

	// 3. Draft a rule from the ACME sample via the agent.
	var rule core.Rule
	post("/api/rules/draft", map[string]any{
		"intent":          "PLN invoices become invoice documents; total is the sum of line amounts.",
		"sample_event_id": wl.Events[0].ID,
		"object_type":     "invoice",
	}, &rule)
	if rule.Status != core.StatusDraft || rule.ID != "book-pln-invoice" {
		t.Fatalf("draft = %+v", rule)
	}

	// 4. Approve: both PLN invoices book, the EUR one stays.
	var approved struct {
		Rule   core.Rule `json:"rule"`
		Booked int       `json:"booked"`
	}
	post("/api/rules/book-pln-invoice/approve", map[string]any{"approved_by": "krzysztof"}, &approved)
	if approved.Rule.Status != core.StatusActive || approved.Booked != 2 {
		t.Fatalf("approve = %+v", approved)
	}
	get("/api/worklist", &wl)
	if len(wl.Events) != 1 { // Gamma EUR remains unexplained
		t.Fatalf("worklist after approve: %d", len(wl.Events))
	}

	// 5. List view shows both documents with formatted money.
	var list struct {
		Columns      []string   `json:"columns"`
		Rows         [][]string `json:"rows"`
		ObjectIDs    []string   `json:"object_ids"`
		DetailViewID string     `json:"detail_view_id"`
	}
	get("/api/views/invoice-list", &list)
	if len(list.Rows) != 2 || list.DetailViewID != "invoice-detail" {
		t.Fatalf("list = %+v", list)
	}
	if list.Rows[0][0] != "ACME Sp. z o.o." || list.Rows[0][3] != "350.50" {
		t.Fatalf("list row 0 = %v", list.Rows[0])
	}

	// 6. Detail view explains the object by its rule.
	var detail struct {
		Provenance string `json:"provenance"`
		Sections   []struct {
			Title  string
			Fields []struct{ Label, Value string }
		} `json:"sections"`
	}
	get("/api/views/invoice-detail?object_id="+list.ObjectIDs[0], &detail)
	if !strings.Contains(detail.Provenance, "book-pln-invoice") {
		t.Fatalf("provenance = %q", detail.Provenance)
	}
	if detail.Sections[1].Fields[0].Value != "350.50" {
		t.Fatalf("detail total = %+v", detail.Sections[1])
	}

	// 7. Analysis view aggregates from the DuckDB read side.
	var analysis struct {
		Columns []string   `json:"columns"`
		Rows    [][]string `json:"rows"`
	}
	get("/api/views/invoice-by-customer", &analysis)
	if len(analysis.Rows) != 2 {
		t.Fatalf("analysis = %+v", analysis)
	}
	if analysis.Rows[0][0] != "Beta Industries GmbH" || analysis.Rows[0][2] != "1200.00" {
		t.Fatalf("analysis row 0 = %v", analysis.Rows[0])
	}

	// 8. Determinism: replay reproduces the object cache.
	x := &exec.Executor{Store: s}
	before, _ := s.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := s.AllObjects(ctx)
	bj, _ := json.Marshal(before)
	aj, _ := json.Marshal(after)
	if string(bj) != string(aj) {
		t.Fatalf("replay diverged")
	}
}

// seedFromFile loads a testdata/seed file the same way `rhea load` does.
func seedFromFile(t *testing.T, s *store.Store, file string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "testdata", "seed", file))
	if err != nil {
		t.Fatal(err)
	}
	var defs struct {
		ObjectTypes []core.ObjectType `json:"object_types"`
		ViewDefs    []core.ViewDef    `json:"view_defs"`
	}
	if err := json.Unmarshal(b, &defs); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for _, ot := range defs.ObjectTypes {
		if err := s.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	for _, vd := range defs.ViewDefs {
		if err := s.InsertViewDef(ctx, vd); err != nil {
			t.Fatal(err)
		}
	}
}

// TestMasterDataRefStory drives the ref<company> slice through the HTTP API:
// companies are master data objects, invoices reference them via =ref(), an
// invoice naming an unknown company waits in the worklist until the company
// is registered, and the fixpoint pass then books it without human help.
func TestMasterDataRefStory(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	seedFromFile(t, s, "finance.json")
	seedFromFile(t, s, "finance_v2.json") // company v1, invoice v2 (ref<company>), joined analysis

	// Recorded agent: one draft per sample event type, both strictly validated.
	fixture := &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
		if strings.Contains(user, `"company.registered"`) {
			return `{
				"rule_id": "register-company",
				"description": "company.registered events become company master data objects.",
				"priority": 50,
				"spec": {
					"match": {"event_type": "company.registered"},
					"effect": {"object": {"type": "company", "fields": {
						"name": "=$.name", "kind": "=$.kind",
						"vat_id": "=$.vat_id", "country": "=$.country"}}}
				}
			}`, nil
		}
		return `{
			"rule_id": "book-pln-invoice",
			"description": "PLN invoices become invoice documents referencing the customer company.",
			"priority": 100,
			"spec": {
				"match": {"event_type": "invoice.received",
				          "where": [{"path": "$.currency", "op": "eq", "value": "PLN"}]},
				"effect": {"object": {"type": "invoice", "fields": {
					"customer": "=ref(company, name, $.customer)", "issue_date": "=$.issue_date",
					"currency": "=$.currency", "total": "=sum($.lines[*].amount)"}}}
			}
		}`, nil
	}}

	srv := &shell.Server{
		Store: s, Exec: &exec.Executor{Store: s}, Agent: fixture,
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
	submitEvent := func(file string) map[string]any {
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

	// 1. An invoice and a company registration arrive; no rules, both pend.
	submitEvent("invoice_acme.json")
	submitEvent("company_acme.json")
	var wl struct {
		Events []core.Event `json:"events"`
	}
	get("/api/worklist", &wl)
	if len(wl.Events) != 2 {
		t.Fatalf("worklist = %d events, want 2", len(wl.Events))
	}
	invoiceEv, companyEv := wl.Events[0], wl.Events[1]

	// 2. Draft + approve the company rule: the company books, the invoice
	// still has no rule and stays.
	var rule core.Rule
	post("/api/rules/draft", map[string]any{
		"intent":          "company.registered events become company master data.",
		"sample_event_id": companyEv.ID,
		"object_type":     "company",
	}, &rule)
	var approved struct {
		Booked int `json:"booked"`
	}
	post("/api/rules/register-company/approve", map[string]any{"approved_by": "krzysztof"}, &approved)
	if approved.Booked != 1 {
		t.Fatalf("company approve booked %d, want 1", approved.Booked)
	}

	// 3. Draft + approve the invoice rule with =ref(): the pending invoice
	// books against the now-existing company.
	post("/api/rules/draft", map[string]any{
		"intent":          "PLN invoices become invoice documents; customer references the company by name.",
		"sample_event_id": invoiceEv.ID,
		"object_type":     "invoice",
	}, &rule)
	post("/api/rules/book-pln-invoice/approve", map[string]any{"approved_by": "krzysztof"}, &approved)
	if approved.Booked != 1 {
		t.Fatalf("invoice approve booked %d, want 1", approved.Booked)
	}
	get("/api/worklist", &wl)
	if len(wl.Events) != 0 {
		t.Fatalf("worklist not empty: %d", len(wl.Events))
	}

	// The stored state holds the company's object id, not its name.
	invoices, err := s.ObjectsByType(ctx, "invoice")
	if err != nil || len(invoices) != 1 {
		t.Fatalf("invoices: %v, %v", invoices, err)
	}
	ref, _ := invoices[0].State["customer"].(string)
	if !strings.HasPrefix(ref, "company-") {
		t.Fatalf("customer = %q, want a company object id", ref)
	}

	// 4. An invoice naming an unknown company waits in the worklist...
	resp := submitEvent("invoice_beta.json")
	if resp["errors"] == nil {
		t.Fatalf("expected an unresolved-ref error, got %v", resp)
	}
	get("/api/worklist", &wl)
	if len(wl.Events) != 1 {
		t.Fatalf("worklist = %d, want the stuck Beta invoice", len(wl.Events))
	}

	// ...and registering the company books both in one submit: the fixpoint
	// pass picks up the invoice the company's materialization unblocked.
	resp = submitEvent("company_beta.json")
	if booked := resp["booked"].(float64); booked != 2 {
		t.Fatalf("fixpoint booked %v, want 2 (company + stuck invoice)", booked)
	}
	get("/api/worklist", &wl)
	if len(wl.Events) != 0 {
		t.Fatalf("worklist not empty after fixpoint: %d", len(wl.Events))
	}

	// 5. Views resolve refs to labels: the list shows the company name.
	var list struct {
		Rows      [][]string `json:"rows"`
		ObjectIDs []string   `json:"object_ids"`
	}
	get("/api/views/invoice-list", &list)
	if len(list.Rows) != 2 || list.Rows[0][0] != "ACME Sp. z o.o." {
		t.Fatalf("list rows = %v", list.Rows)
	}
	var detail struct {
		Sections []struct {
			Title  string
			Fields []struct{ Label, Value string }
		} `json:"sections"`
	}
	get("/api/views/invoice-detail?object_id="+list.ObjectIDs[0], &detail)
	if detail.Sections[0].Fields[0].Value != "ACME Sp. z o.o." {
		t.Fatalf("detail customer = %+v", detail.Sections[0])
	}

	// 6. The analysis view joins invoices to companies on the ref.
	var analysis struct {
		Rows [][]string `json:"rows"`
	}
	get("/api/views/invoice-by-customer", &analysis)
	if len(analysis.Rows) != 2 ||
		analysis.Rows[0][0] != "Beta Industries GmbH" || analysis.Rows[0][2] != "1200.00" ||
		analysis.Rows[1][0] != "ACME Sp. z o.o." || analysis.Rows[1][2] != "350.50" {
		t.Fatalf("analysis rows = %v", analysis.Rows)
	}

	// 7. Determinism: refs were resolved at fire time and baked into derived
	// events, so replay reproduces identical state without re-resolving.
	x := &exec.Executor{Store: s}
	before, _ := s.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := s.AllObjects(ctx)
	bj, _ := json.Marshal(before)
	aj, _ := json.Marshal(after)
	if string(bj) != string(aj) {
		t.Fatalf("replay diverged")
	}
}
