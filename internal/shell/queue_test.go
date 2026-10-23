package shell_test

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"rhea/internal/adapter"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

// The queue (DIRECTION 2026-10-10, the concierge's second scenario): what to
// do today. Open invoices ranked by how overdue they are — a formula over
// the invoice with today in it, never code — the paid one absent, the verbs
// that take an invoice offered on each, and the approvals waiting at the
// gate beside them.
func TestQueue(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	day, _ := time.Parse("2006-01-02", "2026-10-14")
	srv := &shell.Server{Store: s, Exec: x, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb"),
		Clock: adapter.Clock{Now: func() time.Time { return day }}}
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
	if err := s.InsertObjectType(ctx, core.ObjectType{Name: "sales_invoice", Version: 1, Domain: "finance", IsDocument: true, LabelField: "number",
		Fields: []core.FieldDef{
			{Name: "number", Type: "string", Required: true},
			{Name: "due_date", Type: "date", Required: true},
			{Name: "open", Type: "money", Required: true},
			{Name: "status", Type: "enum", Required: true, Values: []string{"open", "paid"}}},
		Lifecycle: &core.LifecycleDef{Field: "status", Transitions: map[string][]string{"open": {"paid"}}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: "book", Status: core.StatusActive, Priority: 100,
		EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: "book", Spec: core.RuleSpec{
			Match: core.Match{EventType: "sales.invoice.issued"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "sales_invoice", Fields: map[string]string{
				"number": "=$.number", "due_date": "=$.due_date", "open": "=$.gross", "status": "=$.status"}}}}}); err != nil {
		t.Fatal(err)
	}
	// A verb that takes an invoice, and one that does not.
	for _, a := range []core.Activity{
		{Name: "send_reminder", Domain: "finance", Description: "remind", Spec: core.ActivitySpec{
			Inputs: []core.FieldDef{{Name: "invoice", Type: "ref<sales_invoice>", Required: true}}, Emits: "reminder.sent", Who: []string{"human"}}},
		{Name: "note_something", Domain: "work", Description: "note", Spec: core.ActivitySpec{
			Inputs: []core.FieldDef{{Name: "text", Type: "string", Required: true}}, Emits: "note.taken", Who: []string{"human"}}},
	} {
		a.Status = core.StatusActive
		a.CreatedBy = "test"
		if _, err := s.InsertActivityVersion(ctx, a); err != nil {
			t.Fatal(err)
		}
	}
	spec, _ := json.Marshal(core.QueueSpec{ObjectType: "sales_invoice",
		Where:   []core.Condition{{Path: "$.state.status", Op: "eq", Value: "open"}},
		Rank:    "=$.today - $.state.due_date",
		Columns: []string{"number", "due_date", "open"}, Approvals: true})
	if err := s.InsertViewDef(ctx, core.ViewDef{ID: "chase", Version: 1, Notion: "queue", Title: "To chase today",
		Domain: "finance", Function: "briefing", Spec: spec}); err != nil {
		t.Fatal(err)
	}
	for i, inv := range []struct{ number, due, gross, status string }{
		{"FV/1", "2026-10-01", "100.00", "open"}, // 13 days overdue
		{"FV/2", "2026-10-10", "250.00", "open"}, // 4 days overdue
		{"FV/3", "2026-09-20", "999.00", "paid"}, // paid: not in the queue
		{"FV/4", "2026-10-20", "300.00", "open"}, // due in 6 days
	} {
		if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: "sales.invoice.issued", OccurredAt: "2026-09-20",
			Payload:  []byte(fmt.Sprintf(`{"number":"%s","due_date":"%s","gross":"%s","status":"%s"}`, inv.number, inv.due, inv.gross, inv.status)),
			DedupKey: fmt.Sprintf("fv%d", i)}); err != nil {
			t.Fatal(err)
		}
	}
	if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
		t.Fatal(errs)
	}
	// A bundle waiting at the gate.
	if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: "later", Status: core.StatusDraft, Priority: 100,
		EffectiveFrom: "2026-01-01", CreatedBy: "agent", Description: "later", Spec: core.RuleSpec{
			Match:  core.Match{EventType: "x"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "sales_invoice", Fields: map[string]string{"number": "=$.n"}}}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.InsertBundle(ctx, core.Bundle{ID: "chasing-policy", Status: core.StatusDraft, Description: "how we chase",
		Members: []core.Member{{Kind: core.KindRule, Name: "later", Version: 1}}, CreatedBy: "agent"}); err != nil {
		t.Fatal(err)
	}

	var q struct {
		Today string `json:"today"`
		Items []struct {
			Label string                       `json:"label"`
			Rank  string                       `json:"rank"`
			Cells map[string]map[string]string `json:"cells"`
		} `json:"items"`
		Verbs []struct {
			Name  string `json:"name"`
			Input string `json:"ref_input"`
		} `json:"verbs"`
		Approvals []struct {
			BundleID string `json:"bundle_id"`
		} `json:"approvals"`
	}
	get("/api/views/chase", &q)
	if q.Today != "2026-10-14" || len(q.Items) != 3 {
		t.Fatalf("the queue: %+v", q)
	}
	if q.Items[0].Label != "FV/1" || q.Items[0].Rank != "13" || q.Items[1].Label != "FV/2" || q.Items[1].Rank != "4" ||
		q.Items[2].Label != "FV/4" || q.Items[2].Rank != "-6" {
		t.Fatalf("ranked most overdue first: %+v", q.Items)
	}
	if q.Items[0].Cells["open"]["v"] != "100.00" || q.Items[0].Cells["due_date"]["v"] != "2026-10-01" {
		t.Fatalf("the columns: %+v", q.Items[0].Cells)
	}
	if len(q.Verbs) != 1 || q.Verbs[0].Name != "send_reminder" || q.Verbs[0].Input != "invoice" {
		t.Fatalf("the verbs that take an invoice: %+v", q.Verbs)
	}
	if len(q.Approvals) != 1 || q.Approvals[0].BundleID != "chasing-policy" {
		t.Fatalf("the approvals waiting: %+v", q.Approvals)
	}
}
