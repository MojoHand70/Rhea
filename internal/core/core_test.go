package core

import (
	"encoding/json"
	"testing"
)

func decode(t *testing.T, s string) any {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return v
}

const samplePayload = `{
	"customer": "ACME Sp. z o.o.",
	"currency": "PLN",
	"issue_date": "2026-09-15",
	"lines": [
		{"desc": "Widget", "qty": 2, "amount": "200.00"},
		{"desc": "Gadget", "qty": 1, "amount": "150.50"}
	]
}`

func TestMoney(t *testing.T) {
	cases := map[string]int64{"0.00": 0, "123.45": 12345, "-7.05": -705, "200.00": 20000}
	for s, want := range cases {
		got, err := ParseMoney(s)
		if err != nil || got != want {
			t.Errorf("ParseMoney(%q) = %d, %v; want %d", s, got, err, want)
		}
		if ParseMoneyMust(s) != want {
			t.Errorf("roundtrip mismatch for %q", s)
		}
	}
	for _, bad := range []string{"1.5", "", "abc", "1.234", "."} {
		if _, err := ParseMoney(bad); err == nil {
			t.Errorf("ParseMoney(%q): want error", bad)
		}
	}
	if FormatMoney(-705) != "-7.05" || FormatMoney(12345) != "123.45" {
		t.Error("FormatMoney wrong")
	}
}

func ParseMoneyMust(s string) int64 {
	m, err := ParseMoney(s)
	if err != nil {
		panic(err)
	}
	return m
}

func TestTemplateEval(t *testing.T) {
	payload := decode(t, samplePayload)

	tm, err := ParseTemplate("=$.customer")
	if err != nil {
		t.Fatal(err)
	}
	v, err := tm.Eval(payload, "string")
	if err != nil || v != "ACME Sp. z o.o." {
		t.Errorf("path eval = %v, %v", v, err)
	}

	tm, _ = ParseTemplate("=sum($.lines[*].amount)")
	v, err = tm.Eval(payload, "money")
	if err != nil || v != int64(35050) {
		t.Errorf("sum eval = %v, %v; want 35050", v, err)
	}

	tm, _ = ParseTemplate("faktura")
	v, _ = tm.Eval(payload, "string")
	if v != "faktura" {
		t.Errorf("literal eval = %v", v)
	}

	tm, _ = ParseTemplate("=$.lines[0].desc")
	v, err = tm.Eval(payload, "string")
	if err != nil || v != "Widget" {
		t.Errorf("indexed eval = %v, %v", v, err)
	}

	if _, err := ParseTemplate("=$.lines[x].desc"); err == nil {
		t.Error("bad index: want parse error")
	}
	if _, err := ParseTemplate("=nonsense"); err == nil {
		t.Error("non-path expression: want parse error")
	}

	tm, _ = ParseTemplate("=$.missing")
	if _, err := tm.Eval(payload, "string"); err == nil {
		t.Error("missing path: want eval error")
	}
}

func TestConditions(t *testing.T) {
	payload := decode(t, samplePayload)
	cases := []struct {
		c    Condition
		want bool
	}{
		{Condition{Path: "$.currency", Op: "eq", Value: "PLN"}, true},
		{Condition{Path: "$.currency", Op: "ne", Value: "EUR"}, true},
		{Condition{Path: "$.lines", Op: "exists"}, true},
		{Condition{Path: "$.nope", Op: "exists"}, false},
		{Condition{Path: "$.lines[0].qty", Op: "gt", Value: 1}, true},
		{Condition{Path: "$.lines[0].qty", Op: "lt", Value: 1}, false},
		{Condition{Path: "$.nope", Op: "eq", Value: "x"}, false},
	}
	for i, tc := range cases {
		got, err := EvalCondition(payload, tc.c)
		if err != nil || got != tc.want {
			t.Errorf("case %d: got %v, %v; want %v", i, got, err, tc.want)
		}
	}
}

func invoiceType() ObjectType {
	return ObjectType{
		Name: "invoice", Version: 1, Domain: "finance", IsDocument: true,
		Fields: []FieldDef{
			{Name: "customer", Type: "string", Required: true},
			{Name: "issue_date", Type: "date", Required: true},
			{Name: "currency", Type: "string", Required: true},
			{Name: "total", Type: "money", Required: true},
		},
	}
}

func validSpec() RuleSpec {
	return RuleSpec{
		Match: Match{EventType: "invoice.received", Where: []Condition{
			{Path: "$.currency", Op: "eq", Value: "PLN"},
		}},
		Effect: Effect{Object: ObjectTemplate{
			Type: "invoice",
			Fields: map[string]string{
				"customer":   "=$.customer",
				"issue_date": "=$.issue_date",
				"currency":   "=$.currency",
				"total":      "=sum($.lines[*].amount)",
			},
		}},
	}
}

func TestRuleSpecValidate(t *testing.T) {
	ot := invoiceType()
	if err := validSpec().Validate(&ot); err != nil {
		t.Fatalf("valid spec rejected: %v", err)
	}

	s := validSpec()
	s.Match.EventType = ""
	if err := s.Validate(&ot); err == nil {
		t.Error("missing event_type accepted")
	}

	s = validSpec()
	delete(s.Effect.Object.Fields, "total")
	if err := s.Validate(&ot); err == nil {
		t.Error("missing required field accepted")
	}

	s = validSpec()
	s.Effect.Object.Fields["bogus"] = "x"
	if err := s.Validate(&ot); err == nil {
		t.Error("unknown field accepted")
	}

	s = validSpec()
	s.Match.Where[0].Op = "regex"
	if err := s.Validate(&ot); err == nil {
		t.Error("unknown op accepted")
	}
}
