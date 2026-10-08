package core

import (
	"strings"
	"testing"
)

// The formula language: exact arithmetic, declared rounding, reads through
// links, folds over named collections — and every refusal naming its law.

func evalFormula(t *testing.T, text, payload, fieldType string, env *Env) (any, *Calc) {
	t.Helper()
	v, c, err := mustEval(t, text, payload, fieldType, env)
	if err != nil {
		t.Fatalf("%s: %v", text, err)
	}
	return v, c
}

func mustEval(t *testing.T, text, payload, fieldType string, env *Env) (any, *Calc, error) {
	t.Helper()
	tm, err := ParseTemplate(text)
	if err != nil {
		return nil, nil, err
	}
	return tm.Evaluate(decode(t, payload), FieldDef{Name: "f", Type: fieldType}, env)
}

func wantErr(t *testing.T, text, payload, fieldType string, env *Env, fragment string) {
	t.Helper()
	_, _, err := mustEval(t, text, payload, fieldType, env)
	if err == nil || !strings.Contains(err.Error(), fragment) {
		t.Fatalf("%s: err = %v, want %q", text, err, fragment)
	}
}

func TestFormulaExactArithmeticAndRounding(t *testing.T) {
	p := `{"qty": 10, "price": "12.50", "cost": "12.345", "net": "100.10", "rate": 23, "a": 7, "b": 2}`

	// qty × price is exact and fits two places: no rounding needed
	v, calc := evalFormula(t, "=$.qty * $.price", p, "money", nil)
	if v != int64(12500) {
		t.Fatalf("qty × price = %v", v)
	}
	if calc == nil || calc.Formula != "=$.qty * $.price" || calc.Inputs["$.qty"] != int64(10) || calc.Inputs["$.price"] != "12.50" {
		t.Fatalf("calc = %+v", calc)
	}
	// a three-place cost does not fit money: rounding must be declared
	wantErr(t, "=$.qty * $.cost / 1", p, "money", nil, "division must declare its rounding")
	wantErr(t, "=3 * $.cost", p, "money", nil, "declare round")
	// 37.035 by each admitted method
	for method, want := range map[string]int64{"half_up": 3704, "half_even": 3704, "down": 3703, "up": 3704} {
		v, _ := evalFormula(t, "=round(3 * $.cost, 2, "+method+")", p, "money", nil)
		if v != want {
			t.Fatalf("round(37.035, 2, %s) = %v, want %d", method, v, want)
		}
	}
	// 37.045 ties to even: 37.04
	if v, _ := evalFormula(t, "=round(37.045, 2, half_even)", p, "money", nil); v != int64(3704) {
		t.Fatalf("half_even on 37.045 = %v", v)
	}
	// negatives round away from zero on half_up, toward zero on down
	if v, _ := evalFormula(t, "=round(-3 * $.cost, 2, half_up)", p, "money", nil); v != int64(-3704) {
		t.Fatalf("round(-37.035) half_up = %v", v)
	}
	if v, _ := evalFormula(t, "=round(-3 * $.cost, 2, down)", p, "money", nil); v != int64(-3703) {
		t.Fatalf("round(-37.035) down = %v", v)
	}
	// VAT: net × rate / 100, statutory half_up to the grosz
	if v, _ := evalFormula(t, "=round($.net * $.rate / 100, 2, half_up)", p, "money", nil); v != int64(2302) {
		t.Fatalf("vat = %v", v)
	}
	// exact decimals, never floats: 0.1 + 0.2 is 0.3
	if v, _ := evalFormula(t, "=0.1 + 0.2", p, "decimal", nil); v != "0.3" {
		t.Fatalf("0.1 + 0.2 = %v", v)
	}
	// a third has no decimal; a rounded one does
	wantErr(t, "=round(1 / 3, 2, half_up) * 3 + 1 / 3", p, "decimal", nil, "division must declare")
	if v, _ := evalFormula(t, "=round(1 / 3, 4, half_up)", p, "decimal", nil); v != "0.3333" {
		t.Fatalf("1/3 = %v", v)
	}
	// whole numbers
	if v, _ := evalFormula(t, "=$.a + $.b * 2", p, "int", nil); v != int64(11) {
		t.Fatalf("a + b*2 = %v", v)
	}
	if v, _ := evalFormula(t, "=round($.a / $.b, 0, half_up)", p, "int", nil); v != int64(4) {
		t.Fatalf("round(7/2) = %v", v)
	}
	wantErr(t, "=$.a / $.b", p, "int", nil, "division must declare its rounding")
	wantErr(t, "=$.a * 1.5", p, "int", nil, "not a whole number")
	if v, _ := evalFormula(t, "=-$.qty", p, "int", nil); v != int64(-10) {
		t.Fatalf("-qty = %v", v)
	}
	if v, _ := evalFormula(t, "=abs(-$.qty) + min($.a, $.b) + max($.a, $.b)", p, "int", nil); v != int64(19) {
		t.Fatalf("abs/min/max = %v", v)
	}
	// division by zero is a refusal, not a zero
	wantErr(t, "=round($.a / 0, 2, half_up)", p, "decimal", nil, "division by zero")
}

func TestFormulaDatesConditionsStrings(t *testing.T) {
	p := `{"d": "2026-10-05", "e": "2026-09-30", "jan": "2026-01-31", "q": 7}`
	if v, _ := evalFormula(t, "=$.d + 14", p, "date", nil); v != "2026-10-19" {
		t.Fatalf("d + 14 = %v", v)
	}
	if v, _ := evalFormula(t, "=$.d - 5", p, "date", nil); v != "2026-09-30" {
		t.Fatalf("d - 5 = %v", v)
	}
	if v, _ := evalFormula(t, "=$.d - $.e", p, "int", nil); v != int64(5) {
		t.Fatalf("d - e = %v", v)
	}
	if v, _ := evalFormula(t, "=end_of_month($.d)", p, "date", nil); v != "2026-10-31" {
		t.Fatalf("end_of_month = %v", v)
	}
	if v, _ := evalFormula(t, "=add_months($.jan, 1)", p, "date", nil); v != "2026-02-28" {
		t.Fatalf("add_months clamps: %v", v)
	}
	if v, _ := evalFormula(t, "=end_of_month($.d) + 30", p, "date", nil); v != "2026-11-30" {
		t.Fatalf("end of month + 30 = %v", v)
	}
	wantErr(t, "=$.d * 2", p, "date", nil, "dates take")
	wantErr(t, "=$.d + 1.5", p, "date", nil, "whole days")

	if v, _ := evalFormula(t, `=if($.q > 5, "big", "small")`, p, "string", nil); v != "big" {
		t.Fatalf("if = %v", v)
	}
	if v, _ := evalFormula(t, `=if($.q > 5 and not ($.d < $.e), 1, 0)`, p, "int", nil); v != int64(1) {
		t.Fatalf("and/not = %v", v)
	}
	if v, _ := evalFormula(t, `=if($.q = 7 or round($.q / 0, 2, half_up) > 1, "seven", "other")`, p, "string", nil); v != "seven" {
		t.Fatalf("or short-circuits: %v", v)
	}
	wantErr(t, `=if($.q, 1, 0)`, p, "int", nil, "condition")
	if v, _ := evalFormula(t, `="PLN"`, p, "string", nil); v != "PLN" {
		t.Fatalf("string literal = %v", v)
	}
	wantErr(t, `=$.d + "x"`, p, "date", nil, "not a number")
}

func TestFormulaFoldsOverThePayload(t *testing.T) {
	p := `{"lines": [{"qty": 2, "price": "10.00", "kind": "goods"}, {"qty": 1, "price": "0.50", "kind": "service"}], "none": []}`
	if v, calc := evalFormula(t, "=sum(l in $.lines: l.qty * l.price)", p, "money", nil); v != int64(2050) {
		t.Fatalf("sum = %v", v)
	} else if calc.Inputs["$.lines[0].qty"] != int64(2) || calc.Inputs["$.lines[1].price"] != "0.50" {
		t.Fatalf("fold inputs = %v", calc.Inputs)
	}
	if v, _ := evalFormula(t, `=sum(l in $.lines where l.kind = "goods": l.qty * l.price)`, p, "money", nil); v != int64(2000) {
		t.Fatalf("filtered sum = %v", v)
	}
	if v, _ := evalFormula(t, "=count(l in $.lines where l.qty > 1)", p, "int", nil); v != int64(1) {
		t.Fatalf("count = %v", v)
	}
	if v, _ := evalFormula(t, "=count(l in $.lines)", p, "int", nil); v != int64(2) {
		t.Fatalf("count all = %v", v)
	}
	if v, _ := evalFormula(t, "=max(l in $.lines: l.qty) - min(l in $.lines: l.qty)", p, "int", nil); v != int64(1) {
		t.Fatalf("max - min = %v", v)
	}
	if v, _ := evalFormula(t, "=any(l in $.lines: l.qty = 1)", p, "string", nil); v != "true" {
		t.Fatalf("any = %v", v)
	}
	if v, _ := evalFormula(t, "=all(l in $.lines: l.qty > 1)", p, "string", nil); v != "false" {
		t.Fatalf("all = %v", v)
	}
	if v, _ := evalFormula(t, "=fold(l in $.lines, acc = 0: acc + l.qty)", p, "int", nil); v != int64(3) {
		t.Fatalf("fold = %v", v)
	}
	if v, _ := evalFormula(t, "=sum($.lines[*].price)", p, "money", nil); v != int64(1050) {
		t.Fatalf("legacy sum = %v", v)
	}
	if v, _ := evalFormula(t, "=sum(l in $.none: l.qty)", p, "int", nil); v != int64(0) {
		t.Fatalf("empty sum = %v", v)
	}
	wantErr(t, "=min(l in $.none: l.qty)", p, "int", nil, "empty collection")
	wantErr(t, "=sum(l in $.lines: l.qty / l.price)", p, "money", nil, "division must declare")
	if _, err := ParseTemplate("=round(sum(l in $.lines: l.qty / l.price), 2, half_up)"); err != nil {
		t.Fatalf("a round above the fold declares the division's rounding: %v", err)
	}
	wantErr(t, "=count(l in $.lines: l.qty)", p, "int", nil, "count takes a where")
	wantErr(t, "=sum(l in $.lines)", p, "int", nil, "needs its expression")
	wantErr(t, "=sum(l in $.lines: q.qty)", p, "int", nil, "unknown name")
	wantErr(t, "=sum(l in $.lines: sum(l in $.lines: l.qty))", p, "int", nil, "already bound")
}

// a small world: one pz_line linking an item with a four-place standard cost
func linkedWorld() *Env {
	types := map[string]ObjectType{
		"pz_line": {Name: "pz_line", Fields: []FieldDef{
			{Name: "qty", Type: "int"}, {Name: "item", Type: "ref<item>"}, {Name: "value", Type: "money"}}},
		"item": {Name: "item", Fields: []FieldDef{{Name: "sku", Type: "string"}, {Name: "std_cost", Type: "decimal"}}},
		"stock_movement": {Name: "stock_movement", Fields: []FieldDef{
			{Name: "item", Type: "ref<item>"}, {Name: "location", Type: "string"},
			{Name: "direction", Type: "enum", Values: []string{"in", "out"}}, {Name: "qty", Type: "int"}}},
	}
	objects := map[string]map[string]any{
		"item-7":           {"sku": "WID-1", "std_cost": "7.1234"},
		"stock_movement-1": {"item": "item-7", "location": "MAIN", "direction": "in", "qty": float64(10)},
		"stock_movement-2": {"item": "item-7", "location": "MAIN", "direction": "out", "qty": float64(3)},
		"stock_movement-3": {"item": "item-7", "location": "OTHER", "direction": "in", "qty": float64(99)},
	}
	return &Env{
		Types: func(n string) (ObjectType, bool) { t, ok := types[n]; return t, ok },
		Get: func(id string) (map[string]any, bool, error) {
			s, ok := objects[id]
			return s, ok, nil
		},
		Lookup: func(typ, field string, value any) ([]string, error) {
			var ids []string
			for _, id := range []string{"item-7", "stock_movement-1", "stock_movement-2", "stock_movement-3"} {
				if strings.HasPrefix(id, typ+"-") && objects[id][field] == value {
					ids = append(ids, id)
				}
			}
			return ids, nil
		},
		Typed: map[string]string{"$.state": "pz_line"},
	}
}

func TestFormulaReadsThroughLinks(t *testing.T) {
	env := linkedWorld()
	p := `{"object_type": "pz_line", "state": {"qty": 4, "item": "item-7", "value": 1250}, "sku": "WID-1"}`

	// the price of the line's item, four places, rounded on purpose
	v, calc := evalFormula(t, "=round($.state.qty * $.state.item.std_cost, 2, half_up)", p, "money", env)
	if v != int64(2849) {
		t.Fatalf("qty × item.std_cost = %v", v)
	}
	if calc.Inputs["$.state.qty"] != int64(4) || calc.Inputs["$.state.item"] != "item-7" || calc.Inputs["$.state.item.std_cost"] != "7.1234" {
		t.Fatalf("inputs = %v", calc.Inputs)
	}
	wantErr(t, "=$.state.qty * $.state.item.std_cost", p, "money", env, "declare round")
	// typed state: money reads as money, and a bare copy keeps minor units
	if v, _ := evalFormula(t, "=$.state.value * 2", p, "money", env); v != int64(2500) {
		t.Fatalf("value × 2 = %v", v)
	}
	if v, _ := evalFormula(t, "=$.state.value", p, "money", env); v != int64(1250) {
		t.Fatalf("copy = %v", v)
	}
	// after a lookup
	if v, _ := evalFormula(t, "=ref(item, sku, $.sku).std_cost", p, "decimal", env); v != "7.1234" {
		t.Fatalf("ref().std_cost = %v", v)
	}
	// money × money has no meaning; a string has no fields
	wantErr(t, "=$.state.value * $.state.value", p, "money", env, "money × money")
	wantErr(t, "=$.sku.std_cost", p, "decimal", env, "cannot read field")
	wantErr(t, "=$.state.item.nosuch", p, "decimal", env, "missing")
	// the linked object must exist
	bad := `{"object_type": "pz_line", "state": {"qty": 4, "item": "item-99"}}`
	wantErr(t, "=$.state.qty * $.state.item.std_cost", bad, "money", env, `no item "item-99"`)
	// no state at hand: the read says so
	wantErr(t, "=$.state.item.std_cost", p, "decimal", &Env{Types: env.Types, Typed: env.Typed}, "needs object state")
}

func TestFormulaFoldsOverLinkedCollections(t *testing.T) {
	env := linkedWorld()
	p := `{"object_type": "pz_line", "state": {"item": "item-7"}, "location": "MAIN"}`
	book := `=sum(m in objects(stock_movement, item, $.state.item) where m.location = $.location: if(m.direction = "in", m.qty, -m.qty))`
	v, calc := evalFormula(t, book, p, "int", env)
	if v != int64(7) {
		t.Fatalf("book quantity = %v", v)
	}
	ids, _ := calc.Inputs["objects(stock_movement, item, item-7)"].([]string)
	if len(ids) != 3 || calc.Inputs["stock_movement-2.qty"] != int64(3) || calc.Inputs["stock_movement-3.location"] != "OTHER" {
		t.Fatalf("inputs = %v", calc.Inputs)
	}
	// the cap is the kernel's: above it the read is refused, naming itself
	capped := *env
	capped.MaxCollection = 2
	wantErr(t, book, p, "int", &capped, "objects(stock_movement, item, item-7) holds 3 objects, above the cap of 2")
	// the reads a formula performs are declared by writing them
	tm, _ := ParseTemplate(book)
	if r := tm.Formula().Reads(); len(r) != 1 || r[0].Type != "stock_movement" || r[0].Field != "item" {
		t.Fatalf("reads = %v", r)
	}
}

func TestFormulaTypedCheck(t *testing.T) {
	env := linkedWorld()
	fd := FieldDef{Name: "value", Type: "money"}
	for text, fragment := range map[string]string{
		"=$.state.nosuch * 2":                 `pz_line has no field "nosuch"`,
		"=$.state.item.nosuch":                `item has no field "nosuch"`,
		"=$.state.value * $.state.value":      "money × money",
		"=$.state.qty + 1 = 2":                "a bool cannot be stored",
		"=ref(nosuch, sku, $.sku).std_cost":   `unknown type "nosuch"`,
		"=objects(item, nosuch, $.sku)":       `item has no field "nosuch"`,
		"=sum(m in $.state.item: m.qty)":      "iterates a collection, got ref",
		"=$.state.item + 1":                   "cannot compute ref + int",
		"=round($.state.qty, 2, nearest)":     "rounding method must be one of",
		"=round($.state.qty, 1.5, half_up)":   "whole number",
		"=round($.state.qty / 3, 9, half_up)": "at most 6 places",
		"=end_of_month($.state.qty)":          "wants a date",
		"=$.state.qty.x":                      `cannot read field "x" of a int`,
		"=if($.state.qty, 1, 2)":              "wants a condition",
		"=sum(m in objects(stock_movement, item, $.state.item): m.direction)": "sum wants numbers",
	} {
		tm, err := ParseTemplate(text)
		if err == nil {
			err = tm.Check(env.Types, env.Typed, fd)
		}
		if err == nil || !strings.Contains(err.Error(), fragment) {
			t.Errorf("%s: err = %v, want %q", text, err, fragment)
		}
	}
	for _, ok := range []string{
		"=round($.state.qty * $.state.item.std_cost, 2, half_up)",
		"=$.state.value + sum(m in objects(stock_movement, item, $.state.item): m.qty)",
		"=if($.state.qty > 1, $.state.value, 0)",
		"=$.root.lines[0].amount * 2", // raw reads are typed at firing
	} {
		tm, err := ParseTemplate(ok)
		if err == nil {
			err = tm.Check(env.Types, env.Typed, fd)
		}
		if err != nil {
			t.Errorf("%s: %v", ok, err)
		}
	}
	// a date field cannot take an amount, a ref field never a computed id
	tm, _ := ParseTemplate("=$.state.qty + 1")
	if err := tm.Check(env.Types, env.Typed, FieldDef{Name: "due", Type: "date"}); err == nil || !strings.Contains(err.Error(), "cannot be stored") {
		t.Fatalf("int into date: %v", err)
	}
	spec := RuleSpec{Match: Match{EventType: "x"}, Effect: Effect{Object: ObjectTemplate{Type: "pz_line",
		Fields: map[string]string{"item": "=ref(item, sku, $.sku).sku"}}}}
	pz, _ := env.Types("pz_line")
	if err := spec.Validate(&pz); err == nil || !strings.Contains(err.Error(), "never a literal or a computed id") {
		t.Fatalf("computed id: %v", err)
	}
	// the whole-rule gate derives the typed prefixes from the match
	cascade := RuleSpec{
		Match: Match{EventType: EventObjectMaterialized, Where: []Condition{{Path: "$.object_type", Op: "eq", Value: "pz_line"}}},
		Effect: Effect{Object: ObjectTemplate{Type: "pz_line", Fields: map[string]string{
			"qty": "=$.state.qty", "item": "=$.state.item", "value": "=$.state.qty * $.state.item.nosuch"}}}}
	if err := cascade.Check(env.Types); err == nil || !strings.Contains(err.Error(), `item has no field "nosuch"`) {
		t.Fatalf("Check through $.state: %v", err)
	}
	if got := cascade.TypedPrefixes(); got["$.state"] != "pz_line" {
		t.Fatalf("typed prefixes = %v", got)
	}
	cascade.Effect.Object.Each = "=$.root.lines[*]"
	if got := cascade.TypedPrefixes(); got["$.doc.state"] != "pz_line" {
		t.Fatalf("typed prefixes with each = %v", got)
	}
}

func TestFormulaSyntaxErrors(t *testing.T) {
	for text, fragment := range map[string]string{
		"=$.a +":                          "unexpected end",
		"=foo":                            "unknown name",
		"=$.a $.b":                        "unexpected",
		"=(1 + 2":                         `expected ")"`,
		"=ref(item, $.x)":                 "wants (type, field, value)",
		"=round(1, 2)":                    "takes 3 argument",
		"=nosuch(1)":                      "unknown function",
		"=fold(l in $.l: l.q)":            "fold wants",
		"=$.a..b":                         "empty path segment",
		"=1.":                             "expected a field name",
		`="unterminated`:                  "unterminated string",
		"=$.lines[x]":                     "bad index",
		"=sum(l in $.lines where 1: l.q)": "where wants a condition",
	} {
		if _, err := ParseTemplate(text); err == nil || !strings.Contains(err.Error(), fragment) {
			t.Errorf("%s: err = %v, want %q", text, err, fragment)
		}
	}
	// what each kind of template is
	for text, kind := range map[string]string{
		"=$.a.b": "path", "=ref(item, sku, $.x)": "ref", "=ref(a, b, ref(c, d, $.x))": "ref",
		"=$.a + 1": "formula", "=sum($.l[*].x)": "formula", "literal": "literal", "=ref(item, sku, $.x).cost": "formula",
	} {
		tm, err := ParseTemplate(text)
		if err != nil || tm.kind != kind {
			t.Errorf("%s: kind %q, %v; want %q", text, tm.kind, err, kind)
		}
	}
}
