package core

import (
	"strings"
	"testing"
)

// An IBAN is checked as ISO 13616 says, wherever it comes from: its
// country's length and its check digits (2026-10-27). The examples are the
// registry's and the banks' own published ones.
func TestAnIBANIsCheckedAsTheStandardSays(t *testing.T) {
	for _, ok := range []string{"PL61109010140000071219812874", "DE89370400440532013000", "GB82WEST12345698765432", "FR1420041010050500013M02606",
		"NL91ABNA0417164300", "CH9300762011623852957", "NO9386011117947", "MT84MALT011000012345MTLCAST001S"} {
		if err := CheckIBAN(ok); err != nil {
			t.Errorf("%s: %v", ok, err)
		}
	}
	for bad, why := range map[string]string{
		"PL61109010140000071219812875":  "check digits",      // the last digit mistyped
		"PL16109010140000071219812874":  "check digits",      // the check digits swapped
		"PL6110901014000007121981287":   "has 28 characters", // one digit short
		"XX61109010140000071219812874":  "not a country",
		"PL61-109010140000071219812874": "has 28 characters",
		"PL":                            "too short",
	} {
		if err := CheckIBAN(bad); err == nil || !strings.Contains(err.Error(), why) {
			t.Errorf("%s: want %q, got %v", bad, why, err)
		}
	}
	if got := NormalIBAN(" pl61 1090 1014 0000 0712 1981 2874 "); got != "PL61109010140000071219812874" {
		t.Errorf("as a person types it: %q", got)
	}
}

// A type's moves name declared transitions and a group; its key and its
// format name its own fields.
func TestATypeDeclaresItsMovesWhole(t *testing.T) {
	base := func() ObjectType {
		return ObjectType{Name: "account", Version: 1, Fields: []FieldDef{{Name: "number", Type: "string", Format: "iban"},
			{Name: "status", Type: "enum", Values: []string{"proposed", "active"}}, {Name: "proposed_by", Type: "string"}},
			Key: "number", Lifecycle: &LifecycleDef{Field: "status", Initial: "proposed", Transitions: map[string][]string{"proposed": {"active"}},
				Moves: map[string]Move{"proposed>active": {Group: "Bankowość", NotBy: "proposed_by"}}}}
	}
	if err := base().Validate(); err != nil {
		t.Fatal(err)
	}
	for why, spoil := range map[string]func(*ObjectType){
		"not a declared transition": func(t *ObjectType) { t.Lifecycle.Moves = map[string]Move{"active>proposed": {Group: "x"}} },
		"names no group":            func(t *ObjectType) { t.Lifecycle.Moves = map[string]Move{"proposed>active": {}} },
		"not_by": func(t *ObjectType) {
			t.Lifecycle.Moves = map[string]Move{"proposed>active": {Group: "x", NotBy: "nobody"}}
		},
		"begins in": func(t *ObjectType) { t.Lifecycle.Initial = "draft" },
		"key":       func(t *ObjectType) { t.Key = "iban" },
		"format":    func(t *ObjectType) { t.Fields[1].Format = "iban" },
	} {
		ty := base()
		spoil(&ty)
		if err := ty.Validate(); err == nil || !strings.Contains(err.Error(), why) {
			t.Errorf("%s: %v", why, err)
		}
	}
}
