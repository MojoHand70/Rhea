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
	v, err := tm.Eval(payload, FieldDef{Type: "string"}, nil)
	if err != nil || v != "ACME Sp. z o.o." {
		t.Errorf("path eval = %v, %v", v, err)
	}

	tm, _ = ParseTemplate("=sum($.lines[*].amount)")
	v, err = tm.Eval(payload, FieldDef{Type: "money"}, nil)
	if err != nil || v != int64(35050) {
		t.Errorf("sum eval = %v, %v; want 35050", v, err)
	}

	tm, _ = ParseTemplate("faktura")
	v, _ = tm.Eval(payload, FieldDef{Type: "string"}, nil)
	if v != "faktura" {
		t.Errorf("literal eval = %v", v)
	}

	tm, _ = ParseTemplate("=$.lines[0].desc")
	v, err = tm.Eval(payload, FieldDef{Type: "string"}, nil)
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
	if _, err := tm.Eval(payload, FieldDef{Type: "string"}, nil); err == nil {
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

func TestRefTemplate(t *testing.T) {
	payload := decode(t, samplePayload)

	tm, err := ParseTemplate("=ref(company, name, $.customer)")
	if err != nil {
		t.Fatal(err)
	}
	lookup := func(objectType, field string, value any) ([]string, error) {
		if objectType != "company" || field != "name" || value != "ACME Sp. z o.o." {
			t.Fatalf("lookup got (%s, %s, %v)", objectType, field, value)
		}
		return []string{"company-7"}, nil
	}
	v, err := tm.Eval(payload, FieldDef{Type: "ref<company>"}, lookup)
	if err != nil || v != "company-7" {
		t.Errorf("ref eval = %v, %v; want company-7", v, err)
	}

	// No object state available: ref() must refuse, not guess.
	if _, err := tm.Eval(payload, FieldDef{Type: "ref<company>"}, nil); err == nil {
		t.Error("ref eval without lookup succeeded")
	}

	for _, bad := range []string{"=ref(company, name)", "=ref(, name, $.x)", "=ref(company, , $.x)", "=ref(company, name, customer)"} {
		if _, err := ParseTemplate(bad); err == nil {
			t.Errorf("ParseTemplate(%q): want error", bad)
		}
	}
}

func TestEnumEval(t *testing.T) {
	fd := FieldDef{Name: "kind", Type: "enum", Values: []string{"customer", "supplier", "self"}}
	payload := decode(t, `{"kind": "customer"}`)

	tm, _ := ParseTemplate("=$.kind")
	if v, err := tm.Eval(payload, fd, nil); err != nil || v != "customer" {
		t.Errorf("enum path eval = %v, %v", v, err)
	}
	tm, _ = ParseTemplate("supplier")
	if v, err := tm.Eval(payload, fd, nil); err != nil || v != "supplier" {
		t.Errorf("enum literal eval = %v, %v", v, err)
	}
	tm, _ = ParseTemplate("partner")
	if _, err := tm.Eval(payload, fd, nil); err == nil {
		t.Error("enum accepted a value outside its set")
	}
}

func companyType() ObjectType {
	return ObjectType{
		Name: "company", Version: 1, Domain: "finance", LabelField: "name",
		Fields: []FieldDef{
			{Name: "name", Type: "string", Required: true},
			{Name: "kind", Type: "enum", Values: []string{"customer", "supplier", "self"}, Required: true},
		},
	}
}

func TestObjectTypeValidate(t *testing.T) {
	if err := companyType().Validate(); err != nil {
		t.Fatalf("valid type rejected: %v", err)
	}
	ot := companyType()
	ot.Fields[0].Type = "varchar"
	if err := ot.Validate(); err == nil {
		t.Error("unknown field type accepted")
	}
	ot = companyType()
	ot.Fields[1].Values = nil
	if err := ot.Validate(); err == nil {
		t.Error("enum without values accepted")
	}
	ot = companyType()
	ot.LabelField = "nope"
	if err := ot.Validate(); err == nil {
		t.Error("label_field pointing nowhere accepted")
	}
	ot = companyType()
	ot.Fields[0].Type = "ref<>"
	if err := ot.Validate(); err == nil {
		t.Error("empty ref target accepted")
	}
}

func TestRuleSpecValidateRefs(t *testing.T) {
	ot := invoiceType()
	ot.Version = 2
	ot.Fields[0] = FieldDef{Name: "customer", Type: "ref<company>", Required: true}

	s := validSpec()
	s.Effect.Object.Fields["customer"] = "=ref(company, name, $.customer)"
	if err := s.Validate(&ot); err != nil {
		t.Fatalf("valid ref spec rejected: %v", err)
	}

	// A ref field may carry an id by path (DECISIONS 2026-10-08, carried
	// links): integrity moves to expansion, where the kernel vouches the id's
	// kind and existence. Structurally the path is admitted...
	s = validSpec()
	if err := s.Validate(&ot); err != nil {
		t.Errorf("ref field with a carried id rejected: %v", err)
	}
	// ...a literal never is: a rule cannot hard-code what it points at.
	s = validSpec()
	s.Effect.Object.Fields["customer"] = "company-12"
	if err := s.Validate(&ot); err == nil {
		t.Error("ref field with a literal id accepted")
	}

	s = validSpec()
	s.Effect.Object.Fields["customer"] = "=ref(account, name, $.customer)"
	if err := s.Validate(&ot); err == nil {
		t.Error("ref target mismatch accepted")
	}

	// And the other way round: ref() into a string field.
	plain := invoiceType()
	s = validSpec()
	s.Effect.Object.Fields["customer"] = "=ref(company, name, $.customer)"
	if err := s.Validate(&plain); err == nil {
		t.Error("ref template on string field accepted")
	}
}

func postingsSpec() RuleSpec {
	return RuleSpec{
		Match: Match{EventType: "invoice.received"},
		Effect: Effect{Postings: &PostingsTemplate{
			Currency: "=$.currency",
			Lines: []PostingLine{
				{Account: "201", Debit: "=sum($.lines[*].amount)"},
				{Account: "702", Credit: "=sum($.lines[*].amount)"},
			},
		}},
	}
}

func TestPostingsValidate(t *testing.T) {
	if err := postingsSpec().Validate(nil); err != nil {
		t.Fatalf("valid postings rejected: %v", err)
	}

	s := postingsSpec()
	s.Effect.Object = ObjectTemplate{Type: "invoice", Fields: map[string]string{"x": "y"}}
	if err := s.Validate(nil); err == nil {
		t.Error("object and postings together accepted")
	}

	s = postingsSpec()
	s.Effect.Postings.Lines = s.Effect.Postings.Lines[:1]
	if err := s.Validate(nil); err == nil {
		t.Error("one-line entry accepted")
	}

	s = postingsSpec()
	s.Effect.Postings.Lines[0].Credit = "=$.x" // both sides set
	if err := s.Validate(nil); err == nil {
		t.Error("line with debit and credit accepted")
	}

	s = postingsSpec()
	s.Effect.Postings.Lines[1].Credit = "" // no side
	if err := s.Validate(nil); err == nil {
		t.Error("line with neither side accepted")
	}

	s = postingsSpec()
	s.Effect.Postings.Currency = ""
	if err := s.Validate(nil); err == nil {
		t.Error("missing currency accepted")
	}
}
