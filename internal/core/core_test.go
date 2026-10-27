package core

import (
	"encoding/json"
	"math"
	"strings"
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
	// Garbage is refused, never read up to the first stray character, and a
	// magnitude past int64 minor units is refused, never wrapped.
	for _, bad := range []string{"1.5", "", "abc", "1.234", ".", "12abc.00", "5. 1", "1_0.00", "1e3.00", "+1.00",
		"92233720368547758.08", "99999999999999999.00"} {
		if _, err := ParseMoney(bad); err == nil {
			t.Errorf("ParseMoney(%q): want error", bad)
		}
	}
	if got, err := ParseMoney("92233720368547758.07"); err != nil || got != math.MaxInt64 {
		t.Errorf("ParseMoney(max) = %d, %v", got, err)
	}
	if FormatMoney(-705) != "-7.05" || FormatMoney(12345) != "123.45" {
		t.Error("FormatMoney wrong")
	}
	if got := FormatMoney(math.MinInt64); got != "-92233720368547758.08" {
		t.Errorf("FormatMoney(MinInt64) = %q", got)
	}
}

// Text spells a JSON-decoded value as Postgres spells it: a number from a
// million upward positionally, never 1e+06, so lookups and joins hold.
func TestText(t *testing.T) {
	for v, want := range map[any]string{
		float64(1234567890): "1234567890", float64(1000000): "1000000", 0.5: "0.5", float64(-3): "-3",
		"abc": "abc", int64(7): "7", true: "true", nil: "",
	} {
		if got := Text(v); got != want {
			t.Errorf("Text(%v) = %q, want %q", v, got, want)
		}
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

// Rule JSON cannot take the kernel down: a condition value that is not a
// scalar is refused at validation and compares false at evaluation, never a
// comparison of uncomparable types; a formula nested past the cap or longer
// than the cap is refused by the parser with its reason, never a fatal
// stack overflow.
func TestRuleJSONCannotPanic(t *testing.T) {
	array := Condition{Path: "$.a", Op: "eq", Value: []any{"a", "b"}}
	spec := RuleSpec{Match: Match{EventType: "x", Where: []Condition{array}},
		Effect: Effect{Object: ObjectTemplate{Type: "t", Fields: map[string]string{"f": "=$.a"}}}}
	if err := spec.Validate(nil); err == nil || !strings.Contains(err.Error(), "the value must be") {
		t.Fatalf("array condition value validated: %v", err)
	}
	payload := map[string]any{"a": []any{"a", "b"}, "m": map[string]any{"k": 1.0}}
	for _, c := range []Condition{array, {Path: "$.m", Op: "eq", Value: map[string]any{"k": 1.0}},
		{Path: "$.a", Op: "ne", Value: []any{"a", "b"}}} {
		ok, err := EvalCondition(payload, c)
		if err != nil || (c.Op == "eq" && ok) || (c.Op == "ne" && !ok) {
			t.Fatalf("%s %s on a non-scalar: %v, %v", c.Path, c.Op, ok, err)
		}
	}
	deep := strings.Repeat("(", 20000) + "1" + strings.Repeat(")", 20000)
	if _, err := ParseFormula(deep); err == nil || !strings.Contains(err.Error(), "nests deeper") {
		t.Fatalf("deep formula: %v", err)
	}
	long := "1" + strings.Repeat(" + 1", 20000)
	if _, err := ParseFormula(long); err == nil || !strings.Contains(err.Error(), "the most is") {
		t.Fatalf("long formula: %v", err)
	}
	if _, err := ParseFormula(strings.Repeat("(", 60) + "1" + strings.Repeat(")", 60)); err != nil {
		t.Fatalf("sixty levels refused: %v", err)
	}
	for _, deep := range []string{strings.Repeat("-", 20000) + "1", strings.Repeat("not ", 10000) + "true"} {
		if _, err := ParseFormula(deep); err == nil || !strings.Contains(err.Error(), "nests deeper") {
			t.Fatalf("twenty thousand unary operators: %v", err)
		}
	}
}

// Identity is checked by prefix, so a type name takes no hyphen; a field
// declared twice would answer by whichever came last.
func TestObjectTypeNamesAndFields(t *testing.T) {
	f := []FieldDef{{Name: "a", Type: "string"}}
	if err := (ObjectType{Name: "a-b", Version: 1, Domain: "x", Fields: f}).Validate(); err == nil || !strings.Contains(err.Error(), "hyphen") {
		t.Fatalf("hyphenated type validated: %v", err)
	}
	twice := []FieldDef{{Name: "a", Type: "string"}, {Name: "a", Type: "int"}}
	if err := (ObjectType{Name: "ab", Version: 1, Domain: "x", Fields: twice}).Validate(); err == nil || !strings.Contains(err.Error(), "twice") {
		t.Fatalf("duplicate field validated: %v", err)
	}
	if err := (ObjectType{Name: "a_b", Version: 1, Domain: "x", Fields: f}).Validate(); err != nil {
		t.Fatal(err)
	}
}

// The language's boundaries: a bare number copied into money is refused
// (it was minor units on a copy and major units in any computation), a
// string copied into an int is read or refused, a literal on a numeric
// field is a number and nothing else, path conditions compare exactly and
// read money text, fx() checks its kinds, and a date takes a sane number
// of days.
func TestLanguageBoundaries(t *testing.T) {
	payload := decode(t, `{"net": 12345, "text": "abc", "n": "12", "amount": "70.00", "big": 1000000, "d": "2026-01-31"}`)
	eval := func(tmpl, typ string) (any, error) {
		t.Helper()
		tm, err := ParseTemplate(tmpl)
		if err != nil {
			t.Fatal(err)
		}
		v, _, err := tm.Evaluate(payload, FieldDef{Name: "f", Type: typ}, &Env{Raw: true})
		return v, err
	}
	if _, err := eval("=$.net", "money"); err == nil || !strings.Contains(err.Error(), "bare number") {
		t.Fatalf("bare number into money from a raw fact: %v", err)
	}
	if tm, _ := ParseTemplate("=$.net"); tm.raw != "" {
		if v, _, err := tm.Evaluate(payload, FieldDef{Name: "f", Type: "money"}, &Env{}); err != nil || v != int64(12345) {
			t.Fatalf("a derived payload's bare number is minor units: %v, %v", v, err)
		}
	}
	if v, err := eval("=$.amount", "money"); err != nil || v != int64(7000) {
		t.Fatalf("money text copied: %v, %v", v, err)
	}
	if _, err := eval("=$.text", "int"); err == nil {
		t.Fatal("abc stored in an int field")
	}
	if v, err := eval("=$.n", "int"); err != nil || v != int64(12) {
		t.Fatalf("string into int: %v, %v", v, err)
	}
	if _, err := eval("$.net * 2", "money"); err == nil || !strings.Contains(err.Error(), "starts with") {
		t.Fatalf("a computation without = on money: %v", err)
	}
	if v, err := eval("12.50", "money"); err != nil || v != int64(1250) {
		t.Fatalf("money literal: %v, %v", v, err)
	}
	if v, err := eval("-3", "int"); err != nil || v != int64(-3) {
		t.Fatalf("int literal: %v, %v", v, err)
	}
	if _, err := eval("=fx($.net, $.big, $.text)", "decimal"); err == nil || !strings.Contains(err.Error(), "currency codes") {
		t.Fatalf("fx on numbers: %v", err)
	}
	if _, err := eval("=$.d + 9223372036854775807", "date"); err == nil || !strings.Contains(err.Error(), "at most") {
		t.Fatalf("date plus a wild int: %v", err)
	}
	if v, err := eval("=$.d + 1", "date"); err != nil || v != "2026-02-01" {
		t.Fatalf("date plus a day: %v, %v", v, err)
	}

	holds := func(c Condition) bool {
		t.Helper()
		ok, err := EvalCondition(payload, c)
		if err != nil {
			t.Fatalf("%+v: %v", c, err)
		}
		return ok
	}
	if !holds(Condition{Path: "$.amount", Op: "gt", Value: "50.00"}) || holds(Condition{Path: "$.amount", Op: "lt", Value: "50.00"}) {
		t.Fatal("money text does not compare as a number")
	}
	if !holds(Condition{Path: "$.amount", Op: "gt", Value: 9.0}) || !holds(Condition{Path: "$.big", Op: "eq", Value: "1000000"}) {
		t.Fatal("number against text")
	}
	if holds(Condition{Path: "$.n", Op: "eq", Value: "012"}) || !holds(Condition{Path: "$.n", Op: "eq", Value: 12.0}) {
		t.Fatal("two texts compare as text, a number against text as numbers")
	}
	if _, err := EvalCondition(payload, Condition{Path: "$.text", Op: "gt", Value: 1.0}); err == nil {
		t.Fatal("gt on text passed")
	}
}

// A typed read whose type the catalog cannot read is an error, never a raw
// read: a money field's minor units read raw as major units, a hundredfold.
func TestTypedReadNeverFallsBackToRaw(t *testing.T) {
	payload := decode(t, `{"state": {"net": 1250}}`)
	tm, err := ParseTemplate("=$.state.net * 2")
	if err != nil {
		t.Fatal(err)
	}
	known := &Env{Typed: map[string]string{"$.state": "invoice"}, Types: func(string) (ObjectType, bool) {
		return ObjectType{Name: "invoice", Version: 1, Fields: []FieldDef{{Name: "net", Type: "money"}}}, true
	}}
	if v, _, err := tm.Evaluate(payload, FieldDef{Name: "f", Type: "money"}, known); err != nil || v != int64(2500) {
		t.Fatalf("typed: %v %v", v, err)
	}
	unknown := &Env{Typed: map[string]string{"$.state": "invoice"}, Types: func(string) (ObjectType, bool) { return ObjectType{}, false }}
	if _, _, err := tm.Evaluate(payload, FieldDef{Name: "f", Type: "money"}, unknown); err == nil || !strings.Contains(err.Error(), "cannot be read from the catalog") {
		t.Fatalf("an unreadable type read raw: %v", err)
	}
}

// A computation over a bare number in a derived fact is refused in money
// (review 2026-10-27, high: the hundredfold slip). A derived payload's bare
// number is minor units when copied — =$.amount over {"amount": 1234} is
// 12.34 — but abs, a minus, a product or a sum over it read major units and
// stored 1234.00: a match's allocation or difference turned a hundred times
// larger the moment a rule computed with it. Now a value computed from a bare
// number has no unit and a money field refuses it; a bare number a formula
// only picks (coalesce, if, min, max) is read as the copy it is. A raw
// fact's bare numbers are counts, as before.
func TestABareNumberKeepsItsUnitThroughAFormula(t *testing.T) {
	derived := decode(t, `{"amount": 1234, "items": [{"amount": 1000}, {"amount": 234}]}`)
	eval := func(tmpl string, env *Env) (any, error) {
		t.Helper()
		tm, err := ParseTemplate(tmpl)
		if err != nil {
			t.Fatal(err)
		}
		v, _, err := tm.Evaluate(derived, FieldDef{Name: "f", Type: "money"}, env)
		return v, err
	}
	for _, f := range []string{"=$.amount", "=coalesce($.amount, 0)", "=if(true, $.amount, 0)", "=max($.amount, 0)"} {
		if v, err := eval(f, &Env{}); err != nil || v != int64(1234) {
			t.Fatalf("%s over a derived fact: %v, %v (want 12.34)", f, v, err)
		}
	}
	for _, f := range []string{"=abs($.amount)", "=-$.amount", "=$.amount * 1", "=$.amount + 0", "=round($.amount, 2, half_up)", "=sum(i in $.items : i.amount)"} {
		if v, err := eval(f, &Env{}); err == nil {
			t.Fatalf("%s over a derived fact stored %v", f, v)
		}
	}
	// a raw fact's bare number is a count: a computation with it still runs
	if v, err := eval("=$.amount * 1.50", &Env{Raw: true}); err != nil || v != int64(185100) {
		t.Fatalf("a count times a price in a raw fact: %v, %v", v, err)
	}
}

// A number longer than any the books can hold is refused before it is read
// (review 2026-10-27, high): a door took {"amount": "777…7.5"} and a rule's
// condition read it — a megabyte of digits took seconds, four took a minute,
// all under the executor's one lock, and the fact stayed in the log to be
// read again on every pass and every replay. A number past the bits a value
// may hold (maxBits) is refused at the text, in time that does not grow.
func TestALongNumberIsRefusedAtItsText(t *testing.T) {
	long := strings.Repeat("7", 100_000) + ".5"
	if _, err := ParseDecimal(long); err == nil {
		t.Fatal("a hundred thousand digits read as a number")
	}
	if _, err := ParseMoney(long); err == nil {
		t.Fatal("a hundred thousand digits read as money")
	}
	// the longest number a value may hold still reads
	if _, err := ParseDecimal(strings.Repeat("9", 1200) + ".25"); err != nil {
		t.Fatalf("1200 digits: %v", err)
	}
}
