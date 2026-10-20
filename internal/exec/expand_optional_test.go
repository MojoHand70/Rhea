package exec_test

import (
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
)

// TestExpandLeavesAnOptionalFieldUnsetWhenItsInputIsAbsent: a door's optional
// input the person did not give (a loan's term, its lender, its rate) is
// copied into an optional field by naming the input; when nothing came, the
// field stays unset — a date, a ref or a decimal has nothing that could stand
// in for it. A required field from an absent input still fails, and so does
// a computation over one, whatever the field.
func TestExpandLeavesAnOptionalFieldUnsetWhenItsInputIsAbsent(t *testing.T) {
	typ := core.ObjectType{Name: "loan", Version: 1, Domain: "finance", LabelField: "title", Fields: []core.FieldDef{
		{Name: "title", Type: "string", Required: true},
		{Name: "due_date", Type: "date"},
		{Name: "lender", Type: "ref<company>"},
		{Name: "rate", Type: "decimal"},
		{Name: "note", Type: "string"},
	}}
	payload := map[string]any{"title": "Pożyczka od wspólnika"}
	copies := core.ObjectTemplate{Type: "loan", Fields: map[string]string{
		"title": "=$.title", "due_date": "=$.due_date", "lender": "=$.lender", "rate": "=$.rate", "note": "=$.note"}}
	state, _, err := exec.Expand(copies, typ, payload, &core.Env{})
	if err != nil {
		t.Fatalf("optional inputs absent: %v", err)
	}
	if state["title"] != "Pożyczka od wspólnika" {
		t.Fatalf("the given input is copied: %v", state)
	}
	for _, f := range []string{"due_date", "lender", "rate", "note"} {
		if v, set := state[f]; set {
			t.Fatalf("%s was set from nothing: %v", f, v)
		}
	}
	// Given, the same inputs land.
	full := map[string]any{"title": "Pożyczka", "due_date": "2027-03-01", "rate": "6", "note": "roczna"}
	state, _, err = exec.Expand(copies, typ, full, &core.Env{})
	if err != nil || state["due_date"] != "2027-03-01" || state["note"] != "roczna" {
		t.Fatalf("given inputs: %v %v", state, err)
	}
	required := core.ObjectTemplate{Type: "loan", Fields: map[string]string{"title": "=$.name"}}
	if _, _, err := exec.Expand(required, typ, payload, &core.Env{}); err == nil {
		t.Fatal("a required field from an absent input must fail")
	}
	computed := core.ObjectTemplate{Type: "loan", Fields: map[string]string{"title": "=$.title", "rate": "=$.rate * 2"}}
	if _, _, err := exec.Expand(computed, typ, payload, &core.Env{}); err == nil {
		t.Fatal("a computation over an absent input must fail, optional field or not")
	}
}
