package core

import "testing"

func TestADoorTakesADecimal(t *testing.T) {
	f := FieldDef{Name: "rate", Type: "decimal"}
	for in, want := range map[any]string{"20": "20", "2.5": "2.5", 2.5: "2.5", 14: "14"} {
		got, err := coerceInput(f, in)
		if err != nil || got != want {
			t.Errorf("coerceInput(%v) = %v, %v; want %q", in, got, err, want)
		}
	}
	if _, err := coerceInput(f, "twenty"); err == nil {
		t.Error("a word is not a decimal")
	}
	a := Activity{Name: "accept", Spec: ActivitySpec{Emits: "x.accepted", Inputs: []FieldDef{{Name: "rate", Type: "decimal", Required: true}}}}
	if err := a.Validate(); err != nil {
		t.Errorf("a decimal input must validate: %v", err)
	}
}
