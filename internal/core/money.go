package core

import (
	"fmt"
	"strings"
)

// Money crosses boundaries as decimal strings ("123.45") and lives in Go as
// int64 minor units (12345). Two decimal places, always; no floats, ever
// (invariant 6).

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
	var w, f int64
	if _, err := fmt.Sscanf(whole, "%d", &w); err != nil || w < 0 {
		return 0, fmt.Errorf("money %q: bad whole part", s)
	}
	if _, err := fmt.Sscanf(frac, "%d", &f); err != nil || f < 0 {
		return 0, fmt.Errorf("money %q: bad fraction", s)
	}
	minor := w*100 + f
	if neg {
		minor = -minor
	}
	return minor, nil
}

func FormatMoney(minor int64) string {
	sign := ""
	if minor < 0 {
		sign = "-"
		minor = -minor
	}
	return fmt.Sprintf("%s%d.%02d", sign, minor/100, minor%100)
}
