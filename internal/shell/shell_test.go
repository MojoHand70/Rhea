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

	// 5. List view shows both documents; money crosses typed and canonical.
	var list struct {
		Columns      []column `json:"columns"`
		Rows         [][]cell `json:"rows"`
		ObjectIDs    []string `json:"object_ids"`
		DetailViewID string   `json:"detail_view_id"`
	}
	get("/api/views/invoice-list", &list)
	if len(list.Rows) != 2 || list.DetailViewID != "invoice-detail" {
		t.Fatalf("list = %+v", list)
	}
	if list.Columns[3].Type != "money" {
		t.Fatalf("total column not typed money: %+v", list.Columns)
	}
	rows := cellVals(list.Rows)
	if rows[0][0] != "ACME Sp. z o.o." || rows[0][3] != "350.50" {
		t.Fatalf("list row 0 = %v", rows[0])
	}

	// 6. Detail view explains the object by its rule, provenance as data.
	var detail struct {
		Provenance struct {
			EventID     int64  `json:"event_id"`
			RuleID      string `json:"rule_id"`
			RuleVersion int    `json:"rule_version"`
		} `json:"provenance"`
		Sections []struct {
			Title  string `json:"title"`
			Fields []struct {
				Label string `json:"label"`
				Type  string `json:"type"`
				V     string `json:"v"`
			} `json:"fields"`
		} `json:"sections"`
	}
	get("/api/views/invoice-detail?object_id="+list.ObjectIDs[0], &detail)
	if detail.Provenance.RuleID != "book-pln-invoice" || detail.Provenance.EventID == 0 {
		t.Fatalf("provenance = %+v", detail.Provenance)
	}
	if detail.Sections[1].Fields[0].V != "350.50" {
		t.Fatalf("detail total = %+v", detail.Sections[1])
	}

	// 6b. The provenance walk: the object's whole story from the raw fact —
	// the root carries the payload, the hop names the rule, the object is a
	// door back to its detail.
	var walk walkOut
	get("/api/explain?object="+list.ObjectIDs[0], &walk)
	if len(walk.Path) != 2 || walk.Path[0] != walk.RootEventID {
		t.Fatalf("walk path = %v, root %d", walk.Path, walk.RootEventID)
	}
	if walk.Tree.Kind != "raw" || len(walk.Tree.Payload) == 0 || len(walk.Tree.Children) != 1 {
		t.Fatalf("walk tree = %+v", walk.Tree)
	}
	if c := walk.Tree.Children[0]; c.RuleID != "book-pln-invoice" || c.Object == nil ||
		c.Object.ObjectID != list.ObjectIDs[0] || c.Object.DetailViewID != "invoice-detail" {
		t.Fatalf("walk child = %+v", walk.Tree.Children[0])
	}
	// A computed value explains itself: the formula and the inputs it read,
	// as baked into the event — the walk never recomputes.
	if calc := walk.Tree.Children[0].Object.Calc["total"]; calc.Formula != "=sum($.lines[*].amount)" || calc.Inputs["$.lines[0].amount"] == nil {
		t.Fatalf("walk calc = %+v", walk.Tree.Children[0].Object.Calc)
	}

	// 7. Analysis view aggregates from the DuckDB read side.
	var analysis struct {
		Rows [][]cell `json:"rows"`
	}
	get("/api/views/invoice-by-customer", &analysis)
	if len(analysis.Rows) != 2 {
		t.Fatalf("analysis = %+v", analysis)
	}
	if arows := cellVals(analysis.Rows); arows[0][0] != "Beta Industries GmbH" || arows[0][2] != "1200.00" {
		t.Fatalf("analysis row 0 = %v", arows[0])
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

// TestDerivedDefaultViews: an ObjectType without stored views still has a
// serviceable list and detail, derived from the type itself (DIRECTION
// 2026-10-05). Stored ViewDefs are the exceptions, never a prerequisite for
// seeing objects; period_lock has never had a view and never needed one.
func TestDerivedDefaultViews(t *testing.T) {
	s := storetest.New(t)
	seedFromFile(t, s, "finance.json")
	seedFromFile(t, s, "finance_v3.json") // period_lock arrives with no views

	srv := &shell.Server{Store: s, Exec: &exec.Executor{Store: s},
		DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb")}
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

	// The derived list is navigable, grouped as master data, and no derived
	// view shadows a type that has stored views.
	var nav []struct {
		Domain    string `json:"domain"`
		Functions []struct {
			Function string `json:"function"`
			Views    []struct {
				ViewID string `json:"view_id"`
				Title  string `json:"title"`
			} `json:"views"`
		} `json:"functions"`
	}
	get("/api/nav", &nav)
	found := ""
	for _, d := range nav {
		for _, fn := range d.Functions {
			for _, v := range fn.Views {
				switch v.ViewID {
				case "derived:list:period_lock":
					found = fn.Function
					if v.Title != "Period lock" {
						t.Fatalf("derived title = %q", v.Title)
					}
				case "derived:list:invoice", "derived:list:account":
					t.Fatalf("derived view shadows a stored one: %s", v.ViewID)
				}
			}
		}
	}
	if found != "master data" {
		t.Fatalf("derived period_lock list grouped as %q, want master data", found)
	}

	// It serves like any stored view: version 0, typed columns from the
	// ObjectType, and a derived detail to click through to.
	var list struct {
		View struct {
			Version int `json:"version"`
		} `json:"view"`
		Columns      []column `json:"columns"`
		Rows         [][]cell `json:"rows"`
		DetailViewID string   `json:"detail_view_id"`
	}
	get("/api/views/derived:list:period_lock", &list)
	if list.View.Version != 0 || list.DetailViewID != "derived:detail:period_lock" {
		t.Fatalf("derived list = %+v", list)
	}
	if len(list.Columns) != 1 || list.Columns[0].Field != "month" || list.Columns[0].Type != "string" {
		t.Fatalf("derived columns = %+v", list.Columns)
	}
}

// column and cell mirror the typed view API: columns declare semantics
// ({field, label, type}), cells carry canonical values ({v}) plus the
// referenced object id ({id}) on refs. Formatting is the renderer's job.
type column struct {
	Field string `json:"field"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

type cell struct {
	V      string `json:"v"`
	ID     string `json:"id"`
	Detail string `json:"detail"`
}

// walkNode mirrors the provenance walk's tree: events annotated with the
// rule version that fired and the object it materialized.
type walkNode struct {
	EventID int64           `json:"event_id"`
	Kind    string          `json:"kind"`
	RuleID  string          `json:"rule_id"`
	Payload json.RawMessage `json:"payload"`
	Object  *struct {
		ObjectID     string               `json:"object_id"`
		ObjectType   string               `json:"object_type"`
		DetailViewID string               `json:"detail_view_id"`
		Calc         map[string]core.Calc `json:"calc"`
	} `json:"object"`
	Children []walkNode `json:"children"`
}

type walkOut struct {
	RootEventID int64    `json:"root_event_id"`
	Path        []int64  `json:"path"`
	Tree        walkNode `json:"tree"`
}

// cellVals flattens typed rows to their canonical values, which is what most
// assertions care about.
func cellVals(rows [][]cell) [][]string {
	out := make([][]string, len(rows))
	for i, r := range rows {
		out[i] = make([]string, len(r))
		for j, c := range r {
			out[i][j] = c.V
		}
	}
	return out
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
		if strings.Contains(user, `Sample event (type "company.registered"`) {
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

	// Dry run before approving (SPEC M1): the diff is field-level and typed —
	// refs resolved against the simulated world, money canonical, and the
	// events the rule would newly explain named. Added objects carry no door:
	// they do not exist yet.
	var sim struct {
		Added []struct {
			ObjectType string `json:"object_type"`
			Detail     string `json:"detail"`
			Fields     []struct {
				Field string `json:"field"`
				Type  string `json:"type"`
				After *cell  `json:"after"`
			} `json:"fields"`
		} `json:"added"`
		Explains []struct {
			EventID   int64  `json:"event_id"`
			EventType string `json:"event_type"`
		} `json:"explains"`
	}
	post("/api/rules/book-pln-invoice/simulate", nil, &sim)
	if len(sim.Added) != 1 || sim.Added[0].ObjectType != "invoice" || sim.Added[0].Detail != "" {
		t.Fatalf("simulate added = %+v", sim.Added)
	}
	for _, f := range sim.Added[0].Fields {
		switch f.Field {
		case "customer":
			if f.After.V != "ACME Sp. z o.o." || !strings.HasPrefix(f.After.ID, "company-") {
				t.Fatalf("diff customer = %+v", f.After)
			}
		case "total":
			if f.Type != "money" || f.After.V != "350.50" {
				t.Fatalf("diff total = %+v", f)
			}
		}
	}
	if len(sim.Explains) != 1 || sim.Explains[0].EventType != "invoice.received" {
		t.Fatalf("simulate explains = %+v", sim.Explains)
	}

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

	// 5. Views resolve refs to labels and carry the referenced id alongside:
	// the list shows the company name, the cell knows the company object.
	var list struct {
		Rows      [][]cell `json:"rows"`
		ObjectIDs []string `json:"object_ids"`
	}
	get("/api/views/invoice-list", &list)
	// Rows order by object_id as text, which shifts with event ids — find
	// the ACME invoice instead of assuming a position.
	acme := -1
	for i, row := range list.Rows {
		if row[0].V == "ACME Sp. z o.o." {
			acme = i
		}
	}
	if len(list.Rows) != 2 || acme < 0 {
		t.Fatalf("list rows = %v", list.Rows)
	}
	if !strings.HasPrefix(list.Rows[acme][0].ID, "company-") {
		t.Fatalf("ref cell id = %q, want the company object id", list.Rows[acme][0].ID)
	}
	if list.Rows[acme][0].Detail != "company-detail" {
		t.Fatalf("ref cell detail = %q, want the company detail view", list.Rows[acme][0].Detail)
	}
	var detail struct {
		Sections []struct {
			Title  string `json:"title"`
			Fields []struct {
				Label string `json:"label"`
				V     string `json:"v"`
				ID    string `json:"id"`
			} `json:"fields"`
		} `json:"sections"`
	}
	get("/api/views/invoice-detail?object_id="+list.ObjectIDs[acme], &detail)
	if f := detail.Sections[0].Fields[0]; f.V != "ACME Sp. z o.o." || !strings.HasPrefix(f.ID, "company-") {
		t.Fatalf("detail customer = %+v", detail.Sections[0])
	}

	// 6. The analysis view joins invoices to companies on the ref.
	var analysis struct {
		Rows [][]cell `json:"rows"`
	}
	get("/api/views/invoice-by-customer", &analysis)
	arows := cellVals(analysis.Rows)
	if len(arows) != 2 ||
		arows[0][0] != "Beta Industries GmbH" || arows[0][2] != "1200.00" ||
		arows[1][0] != "ACME Sp. z o.o." || arows[1][2] != "350.50" {
		t.Fatalf("analysis rows = %v", arows)
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
