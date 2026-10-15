package core

import (
	"regexp"
	"strings"
)

// Numbers as people write them (DECISIONS 2026-10-15). Not the boundary
// format — ParseMoney stays strict, two decimals and a dot — and not a
// channel's fixed layout, which each reader knows for itself (MT940 has
// its comma, the FA(3) XML its dot). This is the prose reader, for the
// host's words today and a mail or a chat line tomorrow: spaces (plain,
// non-breaking, narrow) group thousands; a dot or a comma is the decimal
// mark; when both kinds appear the last one is the decimal mark and the
// others group thousands; one kind appearing more than once groups
// thousands unless the last tail is not three digits long. A lone mark is
// the decimal mark — the continental reading: "1,234" is one and 234
// thousandths, and a thousands group is written with a space or beside a
// decimal mark. A sign is not read: the words say "minus" in their own
// way, and what is compared is the amount.
//
// The value comes back canonical — no grouping, a dot, no leading zeros,
// no trailing zeros after the point — so that "1 234,56", "1.234,56",
// "1,234.56" and "1234.560" are one value and "1234" is another.

// WrittenNumber is one number found in a text: as written, and its value.
type WrittenNumber struct {
	Text  string
	Value string
}

var writtenNumberRe = regexp.MustCompile(`\d+(?:[ \x{00A0}\x{202F}]\d{3})*(?:[.,]\d{3})*(?:[.,]\d+)?`)

// WrittenNumbers finds every number in a text, in order, as written and as
// a canonical value.
func WrittenNumbers(text string) []WrittenNumber {
	var out []WrittenNumber
	for _, m := range writtenNumberRe.FindAllString(text, -1) {
		out = append(out, WrittenNumber{Text: m, Value: ReadWrittenNumber(m)})
	}
	return out
}

// ReadWrittenNumber reads one written number into its canonical value.
func ReadWrittenNumber(written string) string {
	s := strings.NewReplacer(" ", "", " ", "", " ", "").Replace(strings.TrimSpace(written))
	dots, commas := strings.Count(s, "."), strings.Count(s, ",")
	last := strings.LastIndexAny(s, ".,")
	decimal := false
	switch {
	case last < 0:
	case dots > 0 && commas > 0:
		decimal = true // the last mark decides; the other kind groups
	case dots+commas == 1:
		decimal = true // a lone mark is the decimal mark
	default:
		decimal = len(s)-last-1 != 3 // several of one kind: groups, unless the tail says otherwise
	}
	var whole, frac string
	if decimal {
		whole, frac = s[:last], s[last+1:]
	} else {
		whole = s
	}
	whole = strings.NewReplacer(".", "", ",", "").Replace(whole)
	whole = strings.TrimLeft(whole, "0")
	frac = strings.TrimRight(frac, "0")
	if whole == "" {
		whole = "0"
	}
	if frac == "" {
		return whole
	}
	return whole + "." + frac
}
