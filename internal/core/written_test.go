package core

import (
	"reflect"
	"testing"
)

// The prose reader of numbers (DECISIONS 2026-10-15): every spelling a
// Polish, German or English sentence uses for one value reads as that
// value, and the ambiguous "1,234" takes the continental reading.
func TestWrittenNumbers(t *testing.T) {
	for written, value := range map[string]string{
		"1234.56":    "1234.56",
		"1234,56":    "1234.56",
		"1 234,56":   "1234.56",
		"1 234,56":   "1234.56",
		"1 234.56":   "1234.56",
		"1.234,56":   "1234.56",
		"1,234.56":   "1234.56",
		"1234.560":   "1234.56",
		"1234.00":    "1234",
		"1234":       "1234",
		"1 234":      "1234",
		"1.234.567":  "1234567",
		"12,345,678": "12345678",
		"1.234.56":   "1234.56",
		"1,234":      "1.234",
		"0,234":      "0.234",
		"0.50":       "0.5",
		"000":        "0",
		"0.00":       "0",
		"007":        "7",
	} {
		if got := ReadWrittenNumber(written); got != value {
			t.Errorf("%q reads as %q, want %q", written, got, value)
		}
	}
	got := WrittenNumbers("Sprzedaż 1 230,00 zł (VAT 230.00) z dnia 2026-09-27, nr 142/2026; zakupy 615,00, 23%.")
	want := []WrittenNumber{
		{"1 230,00", "1230"}, {"230.00", "230"}, {"2026", "2026"}, {"09", "9"}, {"27", "27"},
		{"142", "142"}, {"2026", "2026"}, {"615,00", "615"}, {"23", "23"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("the numbers in a sentence:\n got %v\nwant %v", got, want)
	}
	// Two amounts in one breath stay two: a comma and a space between them
	// is punctuation, not a thousands group.
	if got := WrittenNumbers("netto 1000.00, VAT 230.00"); len(got) != 2 || got[0].Value != "1000" || got[1].Value != "230" {
		t.Fatalf("punctuation between amounts: %v", got)
	}
}
