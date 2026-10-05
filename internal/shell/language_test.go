package shell_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

// The language explains itself: /api/types serves every object type with its
// explanation — producing rules (with the event type they consume), views
// (derived ones included), ref edges in both directions, instance counts.
func TestLanguageExplainsItself(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	seedFromFile(t, s, "finance_v2.json") // company v1, invoice v2 (customer ref<company>)

	if _, err := s.InsertRuleVersion(ctx, core.Rule{
		ID: "book-invoice", Status: core.StatusActive, Priority: 100,
		EffectiveFrom: "2026-01-01", CreatedBy: "test",
		Description: "received invoices become invoice documents",
		Spec: core.RuleSpec{
			Match: core.Match{EventType: "invoice.received"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "invoice", Fields: map[string]string{
				"customer": "=$.customer", "issue_date": "=$.issue_date",
				"currency": "=$.currency", "total": "=$.total",
			}}},
		},
	}); err != nil {
		t.Fatal(err)
	}

	srv := &shell.Server{Store: s, Exec: &exec.Executor{Store: s},
		DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb")}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	res, err := ts.Client().Get(ts.URL + "/api/types")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("GET /api/types: %d", res.StatusCode)
	}
	var types []struct {
		core.ObjectType
		ProducedBy []struct {
			RuleID    string `json:"rule_id"`
			Status    string `json:"status"`
			EventType string `json:"event_type"`
		} `json:"produced_by"`
		Views []struct {
			ViewID  string `json:"view_id"`
			Notion  string `json:"notion"`
			Derived bool   `json:"derived"`
		} `json:"views"`
		References []struct {
			FromType string `json:"from_type"`
			Field    string `json:"field"`
			ToType   string `json:"to_type"`
		} `json:"references"`
		ReferencedBy []struct {
			FromType string `json:"from_type"`
			Field    string `json:"field"`
			ToType   string `json:"to_type"`
		} `json:"referenced_by"`
		Instances    int                                        `json:"instances"`
	}
	if err := json.NewDecoder(res.Body).Decode(&types); err != nil {
		t.Fatal(err)
	}

	byName := map[string]int{}
	for i, x := range types {
		byName[x.Name] = i
	}
	inv, ok := byName["invoice"]
	if !ok {
		t.Fatalf("invoice missing from /api/types: %v", byName)
	}
	com, ok := byName["company"]
	if !ok {
		t.Fatalf("company missing from /api/types: %v", byName)
	}

	// The producing rule, with the event type it consumes.
	if len(types[inv].ProducedBy) != 1 ||
		types[inv].ProducedBy[0].RuleID != "book-invoice" ||
		types[inv].ProducedBy[0].EventType != "invoice.received" {
		t.Fatalf("invoice produced_by wrong: %+v", types[inv].ProducedBy)
	}
	// Vocabulary without rules says so, honestly.
	if len(types[com].ProducedBy) != 0 {
		t.Fatalf("company should have no producing rule, got %+v", types[com].ProducedBy)
	}

	// Views include the derived list/detail even where none are stored.
	hasDerivedList := false
	for _, v := range types[inv].Views {
		if v.Notion == "list" && v.Derived {
			hasDerivedList = true
		}
	}
	if !hasDerivedList {
		t.Fatalf("invoice views lack the derived list: %+v", types[inv].Views)
	}

	// The ref edge shows from both ends.
	if len(types[inv].References) != 1 || types[inv].References[0].ToType != "company" {
		t.Fatalf("invoice references wrong: %+v", types[inv].References)
	}
	if len(types[com].ReferencedBy) != 1 || types[com].ReferencedBy[0].FromType != "invoice" ||
		types[com].ReferencedBy[0].Field != "customer" {
		t.Fatalf("company referenced_by wrong: %+v", types[com].ReferencedBy)
	}
}
