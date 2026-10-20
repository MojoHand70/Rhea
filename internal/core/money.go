package core

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Money crosses boundaries as decimal strings ("123.45") and lives in Go as
// int64 minor units (12345). Two decimal places, always; no floats, ever
// (invariant 6).

// ParseMoney reads the boundary form strictly: an optional sign, digits, and
// exactly two digits after the point. Anything else is refused, and so is a
// magnitude that does not fit int64 minor units — a parser that stopped at
// the first stray character or wrapped silently would book a wrong amount
// with err == nil.
func ParseMoney(s string) (int64, error) {
	s = strings.TrimSpace(s)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	whole, frac, found := strings.Cut(s, ".")
	if !found {
		frac = "00"
	}
	if len(frac) != 2 {
		return 0, fmt.Errorf("money %q: need exactly 2 decimal places", s)
	}
	if whole == "" {
		return 0, fmt.Errorf("money %q: missing whole part", s)
	}
	if !digits(whole) {
		return 0, fmt.Errorf("money %q: bad whole part", s)
	}
	if !digits(frac) {
		return 0, fmt.Errorf("money %q: bad fraction", s)
	}
	w, err := strconv.ParseInt(whole, 10, 64)
	f, _ := strconv.ParseInt(frac, 10, 64)
	if err != nil || w > math.MaxInt64/100 || (w == math.MaxInt64/100 && f > math.MaxInt64%100) {
		return 0, fmt.Errorf("money %q: does not fit", s)
	}
	minor := w*100 + f
	if neg {
		minor = -minor
	}
	return minor, nil
}

// FormatMoney writes minor units in the boundary form; every int64 prints,
// the most negative one included.
func FormatMoney(minor int64) string {
	sign := ""
	mag := uint64(minor)
	if minor < 0 {
		sign = "-"
		mag = uint64(-(minor + 1)) + 1
	}
	return fmt.Sprintf("%s%d.%02d", sign, mag/100, mag%100)
}

// Abs is the magnitude of an amount in minor units.
func Abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// Text is the one spelling of a payload or state value for a lookup, a join
// or a cell: the text Postgres gives for the same JSON value. A number that
// came through JSON as float64 prints positionally — 1234567890, 1000000 —
// never in exponent form, so a NIP or an amount from a million upward still
// finds the object whose stored text it equals.
func Text(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case json.Number:
		return x.String()
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case bool:
		return strconv.FormatBool(x)
	}
	return fmt.Sprintf("%v", v)
}
