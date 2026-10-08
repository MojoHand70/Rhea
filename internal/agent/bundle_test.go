package agent_test

import (
	"context"
	"strings"
	"testing"

	"rhea/internal/agent"
)

// A bundle may bring its own types: the complaint and its case, with the
// case's lifecycle, the rules that raise both, a view and the verb that
// records a complaint by hand.
const complaintBundle = `{
  "bundle_id": "complaints",
  "description": "Complaints are registered and each opens a follow-up case.",
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
