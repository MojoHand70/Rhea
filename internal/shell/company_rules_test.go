package shell_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"reflect"
	"testing"

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/shell"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

const alfaKey, betaKey = "PL5250001111", "PL5260001246"

// A rule drafted acting for a company is that company's own rule (KK,
// 2026-10-22: company rules and Rhea's global rules): its id says whose it
// is, a redraft of one of Rhea's rules replaces Rhea's for that company
// alone, and its bundle — the question asked and the proposal — is seen only
// where the company is. Drafted acting for no company, both are Rhea's.
func TestADraftActingForACompanyIsItsOwn(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	if err := s.InsertObjectType(ctx, core.ObjectType{Name: "note", Version: 1, Domain: "work", LabelField: "text",
		Fields: []core.FieldDef{{Name: "text", Type: "string", Required: true}}}); err != nil {
		t.Fatal(err)
	}
	spec := core.RuleSpec{Match: core.Match{EventType: "note.taken"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text"}}}}
	if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: "note", Status: core.StatusActive, Priority: 100,
		EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: "Rhea's notes", Spec: spec}); err != nil {
		t.Fatal(err)
	}
	alfa, beta := store.WithCompany(ctx, alfaKey), store.WithCompany(ctx, betaKey)
	draft := func(id string) agent.Draft {
		return agent.Draft{RuleID: id, Description: id + " as drafted", Priority: 100, Spec: spec}
	}
	bundle := func(id string, rules ...agent.Draft) agent.BundleDraft {
		return agent.BundleDraft{BundleID: id, Description: "notes as alfa reads them", Rules: rules,
			Warrant: &core.Warrant{Basis: "client", Citations: []string{"the owner said so"}}}
	}

	b, err := shell.StoreBundleDraft(alfa, s, nil, bundle("notes", draft("note"), draft("memo")), "agent:test", "2026-06-01", "How does alfa read notes?", "")
	if err != nil {
		t.Fatal(err)
	}
	if b.ID != "notes@"+alfaKey || b.Company != alfaKey {
		t.Fatalf("alfa's bundle: %s of %q", b.ID, b.Company)
	}
	var names []string
	for _, m := range b.Members {
		names = append(names, m.Name)
	}
	if !reflect.DeepEqual(names, []string{"note@" + alfaKey, "memo@" + alfaKey}) {
		t.Fatalf("alfa's drafts: %v", names)
	}
	note, err := s.GetRule(alfa, "note@"+alfaKey)
	if err != nil || note.Company != alfaKey || note.Replaces != "note" || note.Status != core.StatusDraft {
		t.Fatalf("alfa's reading of Rhea's rule: %+v %v", note, err)
	}
	if memo, err := s.GetRule(alfa, "memo@"+alfaKey); err != nil || memo.Company != alfaKey || memo.Replaces != "" {
		t.Fatalf("alfa's new rule: %+v %v", memo, err)
	}
	if rhea, _ := s.GetRule(ctx, "note"); rhea.Version != 1 || rhea.Status != core.StatusActive {
		t.Fatalf("Rhea's rule moved: %+v", rhea)
	}

	// Beta sees neither alfa's drafts nor alfa's question.
	if _, err := s.GetRule(beta, "memo@"+alfaKey); err == nil {
		t.Fatal("beta reads alfa's draft")
	}
	if _, err := s.GetBundle(beta, b.ID); err == nil {
		t.Fatal("beta reads alfa's bundle")
	}
	if bs, _ := s.LatestBundles(beta); len(bs) != 0 {
		t.Fatalf("beta lists alfa's bundles: %v", bs)
	}

	// A redraft acting for alfa, naming Rhea's rule again, is the next
	// version of alfa's own reading, in a bundle of alfa's own.
	b2, err := shell.StoreBundleDraft(alfa, s, nil, bundle("notes", draft("note")), "agent:test", "2026-06-01", "Again?", b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if b2.ID != "notes-2@"+alfaKey || b2.Members[0].Name != "note@"+alfaKey || b2.Members[0].Version != 2 {
		t.Fatalf("alfa's redraft: %s %+v", b2.ID, b2.Members)
	}
	if again, _ := s.GetRule(alfa, "note@"+alfaKey); again.Replaces != "note" || again.Company != alfaKey {
		t.Fatalf("the redraft lost what it is: %+v", again)
	}

	// Acting for no company, the proposal is Rhea's.
	b3, err := shell.StoreBundleDraft(store.WithInstallation(ctx), s, nil, bundle("rheas", draft("memo")), "agent:test", "2026-06-01", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if b3.ID != "rheas" || b3.Company != "" || b3.Members[0].Name != "memo" {
		t.Fatalf("Rhea's bundle: %+v", b3)
	}
	if memo, err := s.GetRule(beta, "memo"); err != nil || memo.Company != "" {
		t.Fatalf("Rhea's draft, seen by every company: %+v %v", memo, err)
	}

	// An office reading several clients and acting for none drafts nothing:
	// whose would it be?
	if _, err := shell.StoreBundleDraft(store.WithScope(ctx, []string{alfaKey, betaKey}), s, nil, bundle("whose", draft("x")),
		"agent:test", "2026-06-01", "", ""); err == nil {
		t.Fatal("drafted acting for no one of several companies")
	}
}

// What a company's people taught by hand is the company's own lesson: alfa's
// accountant pairs a part payment the oldest-first way twice, and the host
// drafts alfa's own reading of Rhea's matching rule — replacing Rhea's for
// alfa alone — never a next version of Rhea's that would book every company.
func TestALessonLearnedFromACompanyIsItsOwn(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	alfa := store.WithCompany(ctx, alfaKey)
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
	for i, r := range []struct {
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
	} {
		if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: r.id, Status: core.StatusActive, Priority: 10 * (i + 1),
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: r.id, Spec: r.spec}); err != nil {
			t.Fatal(err)
		}
	}
	submit := func(dedup, evType, date, payload string) {
		t.Helper()
		if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: evType, OccurredAt: date,
			Payload: []byte(payload), DedupKey: dedup, Company: alfaKey}); err != nil {
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
	fv1, line := objectID(t, s, alfa, "sales_invoice", "number", "FV/1"), objectID(t, s, alfa, "bank_line", "status", "open")
	pair := func(line string) exec.Triggered {
		t.Helper()
		res, err := x.TriggerActivity(alfa, "pair_items", map[string]any{
			"pairing": map[string]any{"matching": "match-incoming-payments", "trigger": line, "items": []string{fv1}}}, "test", "")
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	pair(line)
	submit("l2", "bank.statement.line", "2026-09-11", `{"date":"2026-09-11","amount":"30.00","currency":"PLN","reference":"przelew","counterparty_account":"PL60"}`)
	if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
		t.Fatal(errs)
	}
	res := pair(objectID(t, s, alfa, "bank_line", "status", "open")) // the first line is settled whole
	if res.Proposed == nil || res.Proposed.ID != core.LearnedBundleID("incoming-payments", "oldest_first")+"@"+alfaKey || res.Proposed.Company != alfaKey {
		t.Fatalf("the lesson's bundle: %+v", res.Proposed)
	}
	lesson, err := s.GetRule(alfa, "match-incoming-payments@"+alfaKey)
	if err != nil || lesson.Company != alfaKey || lesson.Replaces != "match-incoming-payments" || lesson.Status != core.StatusDraft ||
		len(lesson.Spec.Effect.Matching.Policy) != 2 || res.Proposed.Members[0].Name != lesson.ID {
		t.Fatalf("alfa's lesson: %+v %v", lesson, err)
	}
	if rhea, _ := s.GetRule(ctx, "match-incoming-payments"); rhea.Version != 1 || rhea.Status != core.StatusActive {
		t.Fatalf("Rhea's matching rule moved: %+v", rhea)
	}
}

// objectID is the one object of the type, in the scope, whose field holds the value.
func objectID(t *testing.T, s *store.Store, scope context.Context, typ, field, value string) string {
	t.Helper()
	ids, err := s.FindObjectIDsByField(scope, typ, field, value)
	if err != nil || len(ids) != 1 {
		t.Fatalf("%s with %s %s: %v %v", typ, field, value, ids, err)
	}
	return ids[0]
}

// The one-rule draft door names what it drafts as the bundle door does (KK,
// 2026-10-22): drafted acting for alfa from alfa's waiting fact, the rule is
// alfa's own and replaces Rhea's of the same name for alfa alone; drafted
// acting for no company, it is Rhea's.
func TestOneRuleDraftedForACompanyIsItsOwn(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	if err := s.InsertObjectType(ctx, core.ObjectType{Name: "note", Version: 1, Domain: "work", LabelField: "text",
		Fields: []core.FieldDef{{Name: "text", Type: "string", Required: true}}}); err != nil {
		t.Fatal(err)
	}
	drafted := `{"rule_id": "book-note", "description": "notes", "priority": 100,
		"spec": {"match": {"event_type": "note.taken"}, "effect": {"object": {"type": "note", "fields": {"text": "=$.text"}}}}}`
	if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: "book-note", Status: core.StatusActive, Priority: 100, EffectiveFrom: "2026-01-01",
		CreatedBy: "test", Description: "Rhea's notes", Spec: core.RuleSpec{Match: core.Match{EventType: "note.kept"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text"}}}}}); err != nil {
		t.Fatal(err)
	}
	srv := &shell.Server{Store: s, Exec: &exec.Executor{Store: s}, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb"),
		Agent: &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) { return drafted, nil }}}
	ts := serve(t, srv.Handler())
	post := func(company, path string, body, out any) int {
		t.Helper()
		b, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", ts.URL+path, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		if company != "" {
			req.Header.Set("X-Rhea-Company", company)
		}
		res, err := ts.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		json.NewDecoder(res.Body).Decode(out)
		return res.StatusCode
	}
	waiting := func(company string) int64 {
		t.Helper()
		ev, err := s.AppendEvent(store.WithCompany(ctx, company), core.Event{Kind: core.KindRaw, Type: "note.taken", OccurredAt: "2026-09-01",
			Payload: []byte(`{"text":"a note"}`), DedupKey: "note-" + company, Company: company})
		if err != nil {
			t.Fatal(err)
		}
		return ev
	}

	var rule core.Rule
	if code := post(alfaKey, "/api/rules/draft", map[string]any{"intent": "notes", "sample_event_id": waiting(alfaKey), "object_type": "note"}, &rule); code != 200 {
		t.Fatalf("drafting for alfa: %d %+v", code, rule)
	}
	if rule.ID != "book-note@"+alfaKey || rule.Company != alfaKey || rule.Replaces != "book-note" || rule.Status != core.StatusDraft {
		t.Fatalf("alfa's one-rule draft: %+v", rule)
	}
	if r, _ := s.GetRule(ctx, "book-note"); r.Version != 1 || r.Status != core.StatusActive {
		t.Fatalf("Rhea's rule moved: %+v", r)
	}
	rule = core.Rule{}
	if code := post("", "/api/rules/draft", map[string]any{"intent": "notes", "sample_event_id": waiting(""), "object_type": "note"}, &rule); code != 200 {
		t.Fatalf("drafting for no company: %d %+v", code, rule)
	}
	if rule.ID != "book-note" || rule.Company != "" || rule.Version != 2 {
		t.Fatalf("Rhea's one-rule draft: %+v", rule)
	}
}
