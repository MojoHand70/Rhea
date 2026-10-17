package core

import "testing"

// TestKeptHome: her general description leaves; a number, an address or a
// name the installation holds — inflected, too — keeps it home.
func TestKeptHome(t *testing.T) {
	names := []string{"Agencja Reklamowa Nova sp. z o.o.", "Krzysztof Kowal", "kk"}
	for text, home := range map[string]bool{
		"sprzedaż netto według klientów, miesiąc po miesiącu": false,
		"wynik za jeden miesiąc":                              false,
		"wynik za 2026-08":                                    true,
		"faktury do biuro@nova.pl":                            true,
		"faktury Agencji Novej":                               true,
		"co kupił Kowalski":                                   true,
		"NOVA i jej faktury":                                  true,
	} {
		if got := KeptHome(text, names) != ""; got != home {
			t.Errorf("%q: kept home %v, want %v (%s)", text, got, home, KeptHome(text, names))
		}
	}
}
