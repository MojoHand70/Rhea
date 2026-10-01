package core

import (
	"fmt"
	"strings"
)

// The template expression language, deliberately tiny (SPEC §2, Rule):
//
//	literal          any string not starting with "="
//	=$.a.b[0].c      value at path in the event payload
//	=sum($.x[*].y)   money sum over decimal strings at path, in minor units
//
// Paths: $.seg, seg[N], seg[*]. A [*] fans out into a slice of values.

type Template struct {
	raw  string
	kind string // "literal" | "path" | "sum"
	path []pathSeg
}

type pathSeg struct {
	key   string
	index int  // -1 none, -2 wildcard
	hasIx bool
}

func ParseTemplate(s string) (Template, error) {
	if !strings.HasPrefix(s, "=") {
		return Template{raw: s, kind: "literal"}, nil
	}
	body := strings.TrimSpace(s[1:])
	if strings.HasPrefix(body, "sum(") && strings.HasSuffix(body, ")") {
		inner := strings.TrimSpace(body[4 : len(body)-1])
		p, err := parsePath(inner)
		if err != nil {
			return Template{}, err
		}
		return Template{raw: s, kind: "sum", path: p}, nil
	}
	p, err := parsePath(body)
	if err != nil {
		return Template{}, err
	}
	return Template{raw: s, kind: "path", path: p}, nil
}

func parsePath(s string) ([]pathSeg, error) {
	if !strings.HasPrefix(s, "$.") {
		return nil, fmt.Errorf("path must start with $.")
	}
	var segs []pathSeg
	for _, part := range strings.Split(s[2:], ".") {
		if part == "" {
			return nil, fmt.Errorf("empty path segment in %q", s)
		}
		seg := pathSeg{index: -1}
		if i := strings.IndexByte(part, '['); i >= 0 {
			if !strings.HasSuffix(part, "]") {
				return nil, fmt.Errorf("unterminated index in %q", part)
			}
			ix := part[i+1 : len(part)-1]
			seg.key = part[:i]
			seg.hasIx = true
			if ix == "*" {
				seg.index = -2
			} else {
				n := 0
				if _, err := fmt.Sscanf(ix, "%d", &n); err != nil || n < 0 {
					return nil, fmt.Errorf("bad index %q in %q", ix, part)
				}
				seg.index = n
			}
		} else {
			seg.key = part
		}
		if seg.key == "" {
			return nil, fmt.Errorf("empty key in segment %q", part)
		}
		segs = append(segs, seg)
	}
	return segs, nil
}

// resolve walks payload (decoded JSON) along the path. A [*] segment fans out:
// the result becomes []any of the resolutions of the remaining path.
func resolve(v any, path []pathSeg) (any, error) {
	for i, seg := range path {
		m, ok := v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("segment %q: not an object", seg.key)
		}
		v, ok = m[seg.key]
		if !ok {
			return nil, fmt.Errorf("segment %q: missing", seg.key)
		}
		if seg.hasIx {
			arr, ok := v.([]any)
			if !ok {
				return nil, fmt.Errorf("segment %q: not an array", seg.key)
			}
			if seg.index == -2 { // wildcard: fan out over the rest of the path
				rest := path[i+1:]
				out := make([]any, 0, len(arr))
				for _, el := range arr {
					if len(rest) == 0 {
						out = append(out, el)
						continue
					}
					rv, err := resolve(el, rest)
					if err != nil {
						return nil, err
					}
					out = append(out, rv)
				}
				return out, nil
			}
			if seg.index >= len(arr) {
				return nil, fmt.Errorf("segment %q: index %d out of range", seg.key, seg.index)
			}
			v = arr[seg.index]
		}
	}
	return v, nil
}

// Eval evaluates the template against a decoded payload. The fieldType guides
// coercion: "money" expects decimal strings and yields int64 minor units.
func (t Template) Eval(payload any, fieldType string) (any, error) {
	switch t.kind {
	case "literal":
		return t.raw, nil
	case "path":
		v, err := resolve(payload, t.path)
		if err != nil {
			return nil, err
		}
		if fieldType == "money" {
			s, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("money field wants a decimal string, got %T", v)
			}
			return ParseMoney(s)
		}
		return v, nil
	case "sum":
		v, err := resolve(payload, t.path)
		if err != nil {
			return nil, err
		}
		arr, ok := v.([]any)
		if !ok {
			return nil, fmt.Errorf("sum() needs an array, got %T", v)
		}
		var total int64
		for _, el := range arr {
			s, ok := el.(string)
			if !ok {
				return nil, fmt.Errorf("sum() needs decimal strings, got %T", el)
			}
			m, err := ParseMoney(s)
			if err != nil {
				return nil, err
			}
			total += m
		}
		return total, nil
	}
	return nil, fmt.Errorf("unknown template kind %q", t.kind)
}

// EvalCondition evaluates one match condition against a decoded payload.
func EvalCondition(payload any, c Condition) (bool, error) {
	p, err := parsePath(c.Path)
	if err != nil {
		return false, err
	}
	v, err := resolve(payload, p)
	if c.Op == "exists" {
		return err == nil, nil
	}
	if err != nil {
		return false, nil // missing value matches nothing (except exists)
	}
	switch c.Op {
	case "eq":
		return looseEq(v, c.Value), nil
	case "ne":
		return !looseEq(v, c.Value), nil
	case "gt", "lt":
		a, aok := asFloat(v)
		b, bok := asFloat(c.Value)
		if !aok || !bok {
			return false, fmt.Errorf("op %q needs numbers", c.Op)
		}
		if c.Op == "gt" {
			return a > b, nil
		}
		return a < b, nil
	}
	return false, fmt.Errorf("unknown op %q", c.Op)
}

// looseEq compares across JSON's habit of making every number a float64.
func looseEq(a, b any) bool {
	if af, ok := asFloat(a); ok {
		if bf, ok := asFloat(b); ok {
			return af == bf
		}
		return false
	}
	return a == b
}

func asFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}
