package agent_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"rhea/internal/agent"
	"rhea/internal/core"
)

// A bundle may bring its own types: the complaint and its case, with the
// case's lifecycle, the rules that raise both, a view and the verb that
// records a complaint by hand.
const complaintBundle = `{
  "bundle_id": "complaints",
  "description": "Complaints are registered and each opens a follow-up case.",
  "warrant": {"basis": "practice", "citations": ["complaint handling as a case with a lifecycle"], "scope": "any market"},
  "object_types": [
    {"name": "complaint", "domain": "sales", "is_document": true, "label_field": "number",
     "fields": [{"name": "number", "type": "string", "required": true},
                {"name": "customer", "type": "string", "required": true}]},
    {"name": "followup", "domain": "work", "label_field": "subject",
     "fields": [{"name": "subject", "type": "string", "required": true},
                {"name": "status", "type": "enum", "values": ["open", "resolved"], "required": true}],
     "lifecycle": {"field": "status", "transitions": {"open": ["resolved"]}}}
  ],
  "rules": [
    {"rule_id": "book-complaint", "description": "d", "priority": 100,
     "spec": {"match": {"event_type": "complaint.registered"},
              "effect": {"object": {"type": "complaint", "fields": {"number": "=$.number", "customer": "=$.customer"}}}}},
    {"rule_id": "raise-followup", "description": "d", "priority": 200,
     "spec": {"match": {"event_type": "object.materialized", "where": [{"path": "$.object_type", "op": "eq", "value": "complaint"}]},
              "effect": {"object": {"type": "followup", "fields": {"subject": "=$.state.number", "status": "open"}}}}},
    {"rule_id": "close-followup", "description": "d", "priority": 300,
     "spec": {"match": {"event_type": "complaint.withdrawn"},
              "effect": {"amend": {"type": "followup", "target": "=ref(followup, subject, $.number)", "set": {"status": "resolved"}}}}}
  ],
  "view_defs": [
    {"view_id": "complaints", "notion": "list", "title": "Complaints", "domain": "sales", "function": "complaints",
     "spec": {"object_type": "complaint", "columns": [{"field": "number", "label": "No."}]}}
  ],
  "activities": [
    {"name": "register_complaint", "domain": "sales", "description": "d",
     "spec": {"inputs": [{"name": "number", "type": "string", "required": true},
                         {"name": "customer", "type": "string", "required": true},
                         {"name": "occurred_at", "type": "date", "required": true}],
              "emits": "complaint.registered", "who": ["human"]}}
  ]
}`

func TestDraftBundleFixture(t *testing.T) {
	b, err := fixtureAgent(complaintBundle).DraftBundle(context.Background(), ask("complaints open cases"))
	if err != nil {
		t.Fatal(err)
	}
	if b.Summary() != "2 type(s), 3 rule(s), 1 view(s), 1 verb(s)" {
		t.Fatalf("summary = %s", b.Summary())
	}
}

func TestDraftBundleRefusals(t *testing.T) {
	cases := map[string][2]string{
		"rule into an undeclared type": {`"type": "complaint", "fields": {"number"`, `"type": "grievance", "fields": {"number"`},
		"view of a missing field":      {`{"field": "number", "label": "No."}`, `{"field": "severity", "label": "S"}`},
		"verb speaking a system verb":  {`"emits": "complaint.registered"`, `"emits": "bundle.approved"`},
		"rule proposed twice":          {`"rule_id": "close-followup"`, `"rule_id": "book-complaint"`},
		"ref to an undeclared type":    {`{"name": "customer", "type": "string", "required": true}]},`, `{"name": "customer", "type": "ref<customer>", "required": true}]},`},
		"amending without consent":     {`"lifecycle": {"field": "status", "transitions": {"open": ["resolved"]}}`, `"label_field": "status"`},
		"no warrant":                   {`"warrant": {"basis": "practice", "citations": ["complaint handling as a case with a lifecycle"], "scope": "any market"},`, ``},
		"support figures claimed":      {`"scope": "any market"}`, `"scope": "any market", "support": {"count": 9, "of": 10, "population": "PL"}}`},
		"standard without citation":    {`"basis": "practice", "citations": ["complaint handling as a case with a lifecycle"]`, `"basis": "standard"`},
	}
	for name, edit := range cases {
		answer := strings.Replace(complaintBundle, edit[0], edit[1], 1)
		if answer == complaintBundle {
			t.Fatalf("%s: edit did not apply", name)
		}
		if _, err := fixtureAgent(answer).DraftBundle(context.Background(), ask("x")); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

// The bundle ask shows the residue: every unexplained event shape.
func TestBundleAskShowsResidue(t *testing.T) {
	var seen string
	a := &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
		seen = user
		return complaintBundle, nil
	}}
	in := ask("complaints")
	in.Residue = []agent.Cluster{{EventType: "complaint.withdrawn", Count: 3, Sample: []byte(`{"number":"C-1"}`)}}
	if _, err := a.DraftBundle(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(seen, `complaint.withdrawn ×3: {"number":"C-1"}`) {
		t.Fatalf("residue missing from the ask:\n%s", seen)
	}
}

// "Already answered" is an answer, not a failure: the agent took up a
// learned rule earlier, and the question finds nothing left to add.
func TestDraftBundleAlreadyAnswered(t *testing.T) {
	b, err := fixtureAgent(`{"bundle_id":"withdrawals","description":"already answered by resolve-case-on-withdrawal"}`).
		DraftBundle(context.Background(), ask("withdrawals resolve cases"))
	if !errors.Is(err, agent.ErrAlreadyAnswered) || !strings.Contains(b.Description, "resolve-case-on-withdrawal") {
		t.Fatalf("got %v, %+v", err, b)
	}
}

// The conversation so far reaches the agent before the question: every
// refusal with its reason and the refused proposal, oldest first, and the
// request still ends with the intent — the line the test bench matches on.
func TestAskCarriesTheConversation(t *testing.T) {
	var got string
	a := &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
		got = user
		return complaintBundle, nil
	}}
	ask := ask("complaints open cases")
	ask.Rejections = []agent.Rejection{
		{Proposal: []byte(`{"bundle_id":"one"}`), Reason: "too many fields", By: "krzysztof"},
		{Reason: "model returned unparseable JSON", By: "the approver"},
	}
	if _, err := a.DraftBundle(context.Background(), ask); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"The conversation so far",
		"1. Refused by krzysztof: too many fields",
		`The refused proposal was: {"bundle_id":"one"}`,
		"2. Refused by the approver: model returned unparseable JSON",
		"Answer the latest refusal",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("ask lacks %q:\n%s", want, got)
		}
	}
	if !strings.HasSuffix(strings.TrimSpace(got), "User intent: complaints open cases") {
		t.Fatalf("ask must end with the intent:\n%s", got)
	}
	if strings.Index(got, "The conversation so far") > strings.Index(got, "User intent:") {
		t.Fatal("the conversation must come before the intent")
	}
}

// A figure is a tally: the author says what is added up and the kernel
// writes the query (DIRECTION 2026-10-17, "Teaching her a figure").
const figureBundle = `{
  "bundle_id": "result-by-month",
  "description": "The result for each month of the books.",
  "warrant": {"basis": "client", "citations": ["the owner asked for the result of one month"], "scope": "this company"},
  "view_defs": [
    {"view_id": "result-by-month", "notion": "analysis", "title": "Result month by month", "domain": "finance", "function": "analysis",
     "spec": {"tally": {"of": "posting", "value": "debit-credit",
       "where": [{"field": "book", "op": "eq", "value": "main"}],
       "by": [{"field": "date", "by": "month", "label": "Month"}],
       "lines": [{"line": "Revenue", "value": "credit-debit", "where": [{"field": "account.type", "op": "eq", "value": "revenue"}]},
                 {"line": "Costs", "op": "-", "where": [{"field": "account.type", "op": "eq", "value": "expense"}]},
                 {"line": "Result before tax", "op": "="}]}}}
  ]
}`

func figureAsk() agent.Ask {
	in := ask("show me the result for one month")
	in.Types = append(in.Types,
		core.ObjectType{Name: "account", Fields: []core.FieldDef{{Name: "code", Type: "string"}, {Name: "type", Type: "enum", Values: []string{"revenue", "expense"}}}},
		core.ObjectType{Name: "posting", Fields: []core.FieldDef{{Name: "account", Type: "ref<account>"}, {Name: "side", Type: "enum", Values: []string{"debit", "credit"}},
			{Name: "amount", Type: "money"}, {Name: "date", Type: "date"}, {Name: "book", Type: "string"}}})
	in.Rules = []core.Rule{{ID: "book-sale", Spec: core.RuleSpec{Match: core.Match{EventType: "sales.invoice.issued"},
		Effect: core.Effect{Postings: &core.PostingsTemplate{Book: "main"}}}}}
	return in
}

func TestDraftFigure(t *testing.T) {
	var seen string
	a := &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
		seen = system + user
		return figureBundle, nil
	}}
	b, err := a.DraftBundle(context.Background(), figureAsk())
	if err != nil {
		t.Fatal(err)
	}
	if b.Summary() != "1 view(s)" {
		t.Fatalf("summary = %s", b.Summary())
	}
	if !strings.Contains(seen, "REPORTS") || !strings.Contains(seen, "Books the active rules post to: main") {
		t.Fatalf("the author was not shown the reports or the book")
	}
	for name, edit := range map[string][2]string{
		"SQL instead of a tally":    {`"spec": {"tally": {"of": "posting"`, `"spec": {"sql": "SELECT 1", "tally": {"of": "posting"`},
		"no tally at all":           {`"spec": {"tally": {"of": "posting"`, `"spec": {"talley": {"of": "posting"`},
		"a field the type lacks":    {`"field": "book"`, `"field": "ledger"`},
		"through something no link": {`"field": "account.type", "op": "eq", "value": "revenue"`, `"field": "book.type", "op": "eq", "value": "revenue"`},
		"the month of a string":     {`{"field": "date", "by": "month"`, `{"field": "book", "by": "month"`},
	} {
		answer := strings.Replace(figureBundle, edit[0], edit[1], 1)
		if answer == figureBundle {
			t.Fatalf("%s: edit did not apply", name)
		}
		if _, err := fixtureAgent(answer).DraftBundle(context.Background(), figureAsk()); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

// A dashboard is a report: components, each a query and a visual; the
// author sees the collection to write them (DIRECTION 2026-10-17, "Reporting").
const reportBundle = `{
  "bundle_id": "sales-dashboard",
  "description": "Sprzedaż w skrócie.",
  "warrant": {"basis": "client", "citations": ["the owner asked for a sales page"], "scope": "this company"},
  "view_defs": [{"view_id": "sales-dashboard", "notion": "report", "title": "Sprzedaż w skrócie", "domain": "reports", "function": "report",
    "spec": {"components": [
      {"title": "W tym miesiącu", "reading": "netto faktur sprzedaży z tego miesiąca", "sql": "SELECT sum(net) AS \"Netto\" FROM finance.sales_invoice", "visual": {"kind": "kpi", "value": "Netto"}},
      {"title": "Według klientów", "reading": "netto według nabywcy", "sql": "SELECT 1 AS \"Klient\", 2 AS \"Netto\"", "visual": {"kind": "bar", "x": "Klient", "y": ["Netto"]}}]}}]
}`

func TestDraftReport(t *testing.T) {
	var seen string
	a := &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) { seen = user; return reportBundle, nil }}
	in := figureAsk()
	in.Today = "2026-10-18"
	in.Tables = []agent.CollectionTable{{Schema: "finance", Name: "sales_invoice", Type: "sales_invoice", Columns: []agent.CollectionColumn{{Name: "net", SQL: "DECIMAL(18,2)", Kind: "money"}}}}
	b, err := a.DraftBundle(context.Background(), in)
	if err != nil || b.Summary() != "1 view(s)" {
		t.Fatalf("a report: %v %v", b.Summary(), err)
	}
	if !strings.Contains(seen, "finance.sales_invoice") || !strings.Contains(seen, "net DECIMAL(18,2) (money)") || !strings.Contains(seen, "Today: 2026-10-18") {
		t.Fatalf("the author was not shown the collection:\n%s", seen)
	}
	for name, edit := range map[string][2]string{
		"no components":           {`"spec": {"components": [`, `"spec": {"parts": [`},
		"a component no query":    {`"sql": "SELECT 1 AS \"Klient\", 2 AS \"Netto\""`, `"sql": ""`},
		"a reading of nothing":    {`"reading": "netto według nabywcy"`, `"reading": " "`},
		"a visual it cannot draw": {`"visual": {"kind": "bar"`, `"visual": {"kind": "gauge"`},
	} {
		answer := strings.Replace(reportBundle, edit[0], edit[1], 1)
		if answer == reportBundle {
			t.Fatalf("%s: edit did not apply", name)
		}
		if _, err := fixtureAgent(answer).DraftBundle(context.Background(), in); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}
