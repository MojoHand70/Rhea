package core

import (
	"strings"
	"unicode"
)

// KeptHome says why a text may not leave the installation, or "" when it
// may (DIRECTION 2026-10-17: what she was never taught reaches the house,
// her general description only). It refuses a digit (an amount, a date, a
// tax number, an account), an address (@), and any word of a name the
// installation holds — a company's, a person's — or a word sharing that
// word's stem, so an inflected name is caught too. It errs on the side of
// keeping a text home: a description kept home is still in the log.
func KeptHome(text string, names []string) string {
	for _, r := range text {
		if unicode.IsDigit(r) {
			return "it carries a number"
		}
		if r == '@' {
			return "it carries an address"
		}
	}
	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool { return !unicode.IsLetter(r) })
	for _, name := range names {
		for _, nw := range strings.FieldsFunc(strings.ToLower(name), func(r rune) bool { return !unicode.IsLetter(r) }) {
			rs := []rune(nw)
			if len(rs) < 4 {
				continue // too short to name anyone: "sp", "z", "o", "sa"
			}
			stem := string(rs[:min(len(rs)-1, 5)])
			for _, w := range words {
				if w == nw || strings.HasPrefix(w, stem) {
					return "it names " + name
				}
			}
		}
	}
	return ""
}
