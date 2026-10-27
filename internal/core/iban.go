package core

import (
	"fmt"
	"math/big"
	"strings"
)

// Formats are the shapes a string field may take (FieldDef.Format).
var Formats = []string{"iban"}

// ibanLength is each country's IBAN length, from SWIFT's IBAN Registry
// (ISO 13616): the countries of the European Economic Area, Switzerland,
// the United Kingdom and the countries round them a European company pays.
// A country not listed is refused — never guessed.
var ibanLength = map[string]int{
	"AD": 24, "AL": 28, "AT": 20, "BA": 20, "BE": 16, "BG": 22, "BY": 28, "CH": 21, "CY": 28, "CZ": 24,
	"DE": 22, "DK": 18, "EE": 20, "ES": 24, "FI": 18, "FO": 18, "FR": 27, "GB": 22, "GE": 22, "GI": 23,
	"GL": 18, "GR": 27, "HR": 21, "HU": 28, "IE": 22, "IL": 23, "IS": 26, "IT": 27, "LI": 21, "LT": 20,
	"LU": 20, "LV": 21, "MC": 27, "MD": 24, "ME": 22, "MK": 19, "MT": 31, "NL": 18, "NO": 15, "PL": 28,
	"PT": 25, "RO": 24, "RS": 22, "SE": 24, "SI": 19, "SK": 24, "SM": 27, "TR": 26, "UA": 29, "VA": 22,
}

// NormalIBAN is an IBAN as it is kept: no spaces, capitals.
func NormalIBAN(s string) string {
	return strings.ToUpper(strings.Join(strings.Fields(s), ""))
}

// CheckIBAN tells why a kept IBAN is not one: its country, its length, its
// characters, its check digits (mod 97 = 1). Nil when it is.
func CheckIBAN(s string) error {
	if len(s) < 4 {
		return fmt.Errorf("%q is too short for an IBAN", s)
	}
	n, ok := ibanLength[s[:2]]
	if !ok {
		return fmt.Errorf("%q: %s is not a country whose IBAN Rhea knows", s, s[:2])
	}
	if len(s) != n {
		return fmt.Errorf("%q: an IBAN of %s has %d characters, this has %d", s, s[:2], n, len(s))
	}
	var digits strings.Builder
	for _, c := range s[4:] + s[:4] {
		switch {
		case c >= '0' && c <= '9':
			digits.WriteRune(c)
		case c >= 'A' && c <= 'Z':
			fmt.Fprintf(&digits, "%d", c-'A'+10)
		default:
			return fmt.Errorf("%q: an IBAN is letters and digits", s)
		}
	}
	v, _ := new(big.Int).SetString(digits.String(), 10)
	if new(big.Int).Mod(v, big.NewInt(97)).Int64() != 1 {
		return fmt.Errorf("%q: its check digits do not match — a digit is wrong or two are swapped", s)
	}
	return nil
}

// CheckFormat tells why a value is not of a field's format.
func CheckFormat(f FieldDef, v any) error {
	switch f.Format {
	case "":
		return nil
	case "iban":
		s, _ := v.(string)
		if err := CheckIBAN(s); err != nil {
			return fmt.Errorf("%s: %w", f.Name, err)
		}
		return nil
	}
	return fmt.Errorf("%s: unknown format %q", f.Name, f.Format)
}
