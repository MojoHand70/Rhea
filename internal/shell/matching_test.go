package shell_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

// The grid (DIRECTION 2026-10-11, "the remainder is the grid"): the matching
// notion shows every side's open items, each open trigger item with why the
// policy left it; the pairing verb runs the same engine on a person's
// selection and emits the same event, explained the same way — the walk
// names the person as the actor and the structure as the rule. Replay
// reproduces the person's match as it reproduces the policy's.
func TestMatchingGridAndPairing(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	srv := &shell.Server{Store: s, Exec: x, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb")}
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
	post := func(path string, body any, out any) (int, string) {
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
			return res.StatusCode, e["error"]
		}
		if out != nil {
			json.NewDecoder(res.Body).Decode(out)
		}
		return 200, ""
	}

	for _, ot := range []core.ObjectType{
		{Name: "company", Version: 1, Domain: "finance", LabelField: "name",
			Fields: []core.FieldDef{{Name: "name", Type: "string", Required: true}, {Name: "iban", Type: "string", Required: true}}},
		{Name: "sales_invoice", Version: 1, Domain: "finance", IsDocument: true, LabelField: "number",
			Fields: []core.FieldDef{
				{Name: "number", Type: "string", Required: true},
				{Name: "issue_date", Type: "date", Required: true},
				{Name: "buyer", Type: "ref<company>", Required: true},
				{Name: "gross", Type: "money", Required: true},
				{Name: "currency", Type: "string", Required: true},
				{Name: "open", Type: "money", Required: true},
				{Name: "status", Type: "enum", Required: true, Values: []string{"open", "paid"}}},
			Lifecycle: &core.LifecycleDef{Field: "status", Transitions: map[string][]string{"open": {"open", "paid"}}},
			Amendable: []string{"open"}},
		{Name: "bank_line", Version: 1, Domain: "finance", LabelField: "reference",
			Fields: []core.FieldDef{
				{Name: "date", Type: "date", Required: true},
				{Name: "amount", Type: "money", Required: true},
				{Name: "currency", Type: "string", Required: true},
				{Name: "reference", Type: "string", Required: true},
				{Name: "counterparty_account", Type: "string", Required: true},
				{Name: "open", Type: "money", Required: true},
				{Name: "status", Type: "enum", Required: true, Values: []string{"open", "settled"}}},
			Lifecycle: &core.LifecycleDef{Field: "status", Transitions: map[string][]string{"open": {"open", "settled"}}},
			Amendable: []string{"open"}},
	} {
		if err := s.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	onMatch := core.Match{EventType: core.EventObjectsMatched, Where: []core.Condition{{Path: "$.matching", Op: "eq", Value: "incoming-payments"}}}
	rules := []struct {
		id   string
		spec core.RuleSpec
	}{
		{"register-company", core.RuleSpec{Match: core.Match{EventType: "company.registered"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "company", Fields: map[string]string{"name": "=$.name", "iban": "=$.iban"}}}}},
		{"book-sales-invoice", core.RuleSpec{Match: core.Match{EventType: "sales.invoice.issued"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "sales_invoice", Fields: map[string]string{"number": "=$.number",
				"issue_date": "=$.issue_date", "buyer": "=ref(company, iban, $.buyer_iban)", "gross": "=$.gross", "currency": "=$.currency",
				"open": "=$.gross", "status": "open"}}}}},
		{"bank-line", core.RuleSpec{Match: core.Match{EventType: "bank.statement.line"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "bank_line", Fields: map[string]string{"date": "=$.date", "amount": "=$.amount",
				"currency": "=$.currency", "reference": "=$.reference", "counterparty_account": "=$.counterparty_account",
				"open": "=$.amount", "status": "open"}}}}},
		// A strict policy: exact only, so a part payment stays for the person.
		{"match-incoming-payments", core.RuleSpec{
			Match: core.Match{EventType: core.EventObjectMaterialized, Where: []core.Condition{{Path: "$.object_type", Op: "eq", Value: "bank_line"}}},
			Effect: core.Effect{Matching: &core.MatchingTemplate{Name: "incoming-payments", Trigger: "payment",
				Sides: []core.MatchingSide{
					{Name: "payment", Type: "bank_line", Open: map[string]string{"money": "open"}, Currency: "currency", Date: "date", Reference: "reference"},
					{Name: "invoice", Type: "sales_invoice", Items: "=objects(sales_invoice, buyer, ref(company, iban, $.state.counterparty_account))",
						Open: map[string]string{"money": "open"}, Currency: "currency", Date: "issue_date", Reference: "number"}},
				Policy: []core.MatchingStep{{Strategy: "exact"}}}}}},
		{"settle-invoice", core.RuleSpec{Match: onMatch, Effect: core.Effect{Amend: &core.AmendTemplate{Type: "sales_invoice",
			Each: "=$.sides.invoice[*]", Target: "=$.line.object_id", Set: map[string]string{"open": "=$.line.open_after.money",
				"status": `=if($.line.open_after.money = 0, "paid", "open")`}}}}},
		{"settle-line", core.RuleSpec{Match: onMatch, Effect: core.Effect{Amend: &core.AmendTemplate{Type: "bank_line",
			Each: "=$.sides.payment[*]", Target: "=$.line.object_id", Set: map[string]string{"open": "=$.line.open_after.money",
				"status": `=if($.line.open_after.money = 0, "settled", "open")`}}}}},
	}
	for i, r := range rules {
		if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: r.id, Status: core.StatusActive, Priority: 10 * (i + 1),
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: r.id, Spec: r.spec}); err != nil {
			t.Fatal(err)
		}
	}
	spec, _ := json.Marshal(core.MatchingSpec{Rule: "match-incoming-payments"})
	if err := s.InsertViewDef(ctx, core.ViewDef{ID: "bank-grid", Version: 1, Notion: "matching", Title: "Bank matching",
		Domain: "finance", Function: "matching", Spec: spec}); err != nil {
		t.Fatal(err)
	}
	submit := func(dedup, evType, date, payload string) {
		t.Helper()
		if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: evType, OccurredAt: date,
			Payload: []byte(payload), DedupKey: dedup}); err != nil {
			t.Fatal(err)
		}
	}
	submit("nova", "company.registered", "2026-09-01", `{"name":"Nova","iban":"PL60"}`)
	submit("fv1", "sales.invoice.issued", "2026-09-02", `{"number":"FV/1","issue_date":"2026-09-02","buyer_iban":"PL60","gross":"100.00","currency":"PLN"}`)
	submit("fv2", "sales.invoice.issued", "2026-09-03", `{"number":"FV/2","issue_date":"2026-09-03","buyer_iban":"PL60","gross":"200.00","currency":"PLN"}`)
	submit("l1", "bank.statement.line", "2026-09-10", `{"date":"2026-09-10","amount":"50.00","currency":"PLN","reference":"przelew","counterparty_account":"PL60"}`)
	if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
		t.Fatal(errs)
	}

	// 1. The grid: the open line on the trigger side with why it stayed,
	// both invoices on the other.
	type grid struct {
		Rule    string `json:"rule"`
		Trigger string `json:"trigger"`
		Verb    string `json:"verb"`
		Sides   []struct {
			Name    string `json:"name"`
			Trigger bool   `json:"trigger"`
			Columns []struct {
				Field string `json:"field"`
				Type  string `json:"type"`
			} `json:"columns"`
			Items []struct {
				ObjectID string                       `json:"object_id"`
				Label    string                       `json:"label"`
				Cells    map[string]map[string]string `json:"cells"`
				Why      []string                     `json:"why"`
			} `json:"items"`
		} `json:"sides"`
	}
	var g grid
	get("/api/views/bank-grid", &g)
	if g.Rule != "match-incoming-payments" || g.Trigger != "payment" || g.Verb != "pair_items" || len(g.Sides) != 2 {
		t.Fatalf("grid = %+v", g)
	}
	payment, invoice := g.Sides[0], g.Sides[1]
	if !payment.Trigger || len(payment.Items) != 1 || payment.Items[0].Cells["open"]["v"] != "50.00" ||
		len(payment.Items[0].Why) != 1 || !strings.Contains(payment.Items[0].Why[0], "step 1 exact: no single item on invoice opens at 50.00") {
		t.Fatalf("the trigger side: %+v", payment)
	}
	if invoice.Trigger || len(invoice.Items) != 2 || invoice.Items[0].Label != "FV/1" || invoice.Items[0].Cells["open"]["v"] != "100.00" {
		t.Fatalf("the invoice side: %+v", invoice)
	}
	if len(invoice.Columns) != 4 || invoice.Columns[0].Field != "number" || invoice.Columns[1].Field != "issue_date" || invoice.Columns[2].Type != "money" {
		t.Fatalf("the columns the structure names: %+v", invoice.Columns)
	}
	lineID, fv1 := payment.Items[0].ObjectID, invoice.Items[0].ObjectID

	// 2. The pairing: the person settles FV/1 with the 50.00 — the same
	// engine, the Selected strategy, a part of the invoice.
	var res struct {
		EventID int64 `json:"event_id"`
		Booked  int   `json:"booked"`
	}
	if code, msg := post("/api/activities/pair_items/trigger", map[string]any{"inputs": map[string]any{
		"pairing": map[string]any{"matching": "match-incoming-payments", "trigger": lineID, "items": []string{fv1}}}}, &res); code != 200 {
		t.Fatalf("pairing: %d %s", code, msg)
	}
	get("/api/views/bank-grid", &g)
	if len(g.Sides[0].Items) != 0 {
		t.Fatalf("the line is settled: %+v", g.Sides[0].Items)
	}
	if len(g.Sides[1].Items) != 2 || g.Sides[1].Items[0].Cells["open"]["v"] != "50.00" {
		t.Fatalf("FV/1 keeps 50.00 open: %+v", g.Sides[1].Items)
	}

	// 3. The walk: the person's pairing is the root, the match its child
	// with the structure as its rule and "selected" as the strategy, the
	// amendments the match's children.
	var ex struct {
		Tree struct {
			EventType string `json:"event_type"`
			Actor     string `json:"actor"`
			Children  []struct {
				EventType string `json:"event_type"`
				RuleID    string `json:"rule_id"`
				Matched   *struct {
					Strategy string `json:"strategy"`
					Step     int    `json:"step"`
				} `json:"matched"`
				Children []struct {
					EventType string `json:"event_type"`
				} `json:"children"`
			} `json:"children"`
		} `json:"tree"`
	}
	get("/api/explain?event="+strconvItoa(res.EventID), &ex)
	root := ex.Tree
	if root.EventType != core.EventMatchingPaired || !strings.HasPrefix(root.Actor, "shell") || len(root.Children) != 1 {
		t.Fatalf("the root: %+v", root)
	}
	m := root.Children[0]
	if m.EventType != core.EventObjectsMatched || m.RuleID != "match-incoming-payments" || m.Matched == nil || m.Matched.Strategy != "selected" || m.Matched.Step != 0 || len(m.Children) != 2 {
		t.Fatalf("the match: %+v", m)
	}

	// 4. A pairing that names nothing on a side, or a trigger of the wrong
	// kind, is refused before anything is appended.
	if code, msg := post("/api/activities/pair_items/trigger", map[string]any{"inputs": map[string]any{
		"pairing": map[string]any{"matching": "match-incoming-payments", "trigger": fv1, "items": []string{fv1}}}}, nil); code == 200 || !strings.Contains(msg, "is not a bank_line") {
		t.Fatalf("wrong trigger: %d %s", code, msg)
	}
	left, _ := s.UnmatchedRawEvents(ctx)
	if len(left) != 0 {
		t.Fatalf("a pairing is never residue: %v", left)
	}

	// 5. The host learns. The pairing was what oldest-first would have done,
	// and the event says so; a second pairing the same way makes the host
	// propose oldest-first as a step — a draft bundle with a client warrant
	// naming both pairings, behind the gate.
	var m1 core.MatchedObjects
	for _, ev := range mustDerived(t, s) {
		if ev.Type == core.EventObjectsMatched {
			json.Unmarshal(ev.Payload, &m1)
		}
	}
	if !reflect.DeepEqual(m1.Standards, []string{"oldest_first"}) {
		t.Fatalf("the standards that would have found the pairing: %v", m1.Standards)
	}
	if _, err := s.GetBundle(ctx, core.LearnedBundleID("incoming-payments", "oldest_first")); err == nil {
		t.Fatal("one pairing proposes nothing yet")
	}
	submit("l2", "bank.statement.line", "2026-09-11", `{"date":"2026-09-11","amount":"30.00","currency":"PLN","reference":"przelew","counterparty_account":"PL60"}`)
	if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
		t.Fatal(errs)
	}
	get("/api/views/bank-grid", &g)
	line2 := g.Sides[0].Items[0].ObjectID
	var res2 struct {
		EventID  int64 `json:"event_id"`
		Proposed *struct {
			BundleID string `json:"bundle_id"`
		} `json:"proposed"`
	}
	if code, msg := post("/api/activities/pair_items/trigger", map[string]any{"inputs": map[string]any{
		"pairing": map[string]any{"matching": "match-incoming-payments", "trigger": line2, "items": []string{fv1}}}}, &res2); code != 200 {
		t.Fatalf("second pairing: %d %s", code, msg)
	}
	if res2.Proposed == nil || res2.Proposed.BundleID != "learned-incoming-payments-oldest_first" {
		t.Fatalf("the second pairing the same way proposes: %+v", res2)
	}
	b, err := s.GetBundle(ctx, "learned-incoming-payments-oldest_first")
	if err != nil || b.Status != core.StatusDraft || b.Warrant == nil || b.Warrant.Basis != "client" || len(b.Warrant.Citations) != 2 ||
		!strings.Contains(b.Warrant.Citations[0], "#"+strconvItoa(res.EventID)) || !strings.Contains(b.Warrant.Citations[1], "#"+strconvItoa(res2.EventID)) {
		t.Fatalf("the proposal: %+v %v", b, err)
	}
	draft, err := s.GetRule(ctx, "match-incoming-payments")
	if err != nil || draft.Status != core.StatusDraft || draft.Version != 2 || len(draft.Spec.Effect.Matching.Policy) != 2 ||
		draft.Spec.Effect.Matching.Policy[1].Strategy != "oldest_first" || draft.CreatedBy != "kernel" {
		t.Fatalf("the drafted next version: %+v %v", draft.Spec.Effect.Matching, err)
	}
	// The owner approves; the third part payment (10.00 against the 20.00
	// still open) matches by itself, by the learned step.
	if _, _, procErrs, err := x.ApproveBundle(ctx, b.ID, "krzysztof", "2026-09-30"); err != nil || len(procErrs) > 0 {
		t.Fatalf("approve: %v %v", err, procErrs)
	}
	submit("l3", "bank.statement.line", "2026-09-12", `{"date":"2026-09-12","amount":"10.00","currency":"PLN","reference":"przelew","counterparty_account":"PL60"}`)
	if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
		t.Fatal(errs)
	}
	get("/api/views/bank-grid", &g)
	if len(g.Sides[0].Items) != 0 || g.Sides[1].Items[0].Label != "FV/1" || g.Sides[1].Items[0].Cells["open"]["v"] != "10.00" {
		t.Fatalf("after the learned step: %+v", g.Sides)
	}
	var last core.MatchedObjects
	for _, ev := range mustDerived(t, s) {
		if ev.Type == core.EventObjectsMatched {
			json.Unmarshal(ev.Payload, &last)
		}
	}
	if last.Strategy != "oldest_first" || last.Step != 2 {
		t.Fatalf("the third payment, matched by the learned step: %+v", last)
	}
	// Pairing the same way again proposes nothing: the standard is a step now.
	// And an overruled proposal never returns — the bundle id is one per
	// structure and standard.

	// 6. Replay reproduces the person's match as it does the policy's.
	before, _ := s.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := s.AllObjects(ctx)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("replay does not reproduce the pairing")
	}
}

func strconvItoa(n int64) string { return json.Number(fmtInt(n)).String() }

func fmtInt(n int64) string {
	b, _ := json.Marshal(n)
	return string(b)
}

func mustDerived(t *testing.T, s interface {
	EventsByKind(ctx context.Context, kind string) ([]core.Event, error)
}) []core.Event {
	t.Helper()
	evs, err := s.EventsByKind(context.Background(), core.KindDerived)
	if err != nil {
		t.Fatal(err)
	}
	return evs
}
