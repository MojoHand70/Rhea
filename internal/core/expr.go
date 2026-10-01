package core

import (
	"fmt"
	"slices"
	"strings"
)

// The template expression language, deliberately tiny (SPEC §2, Rule):
//
//	literal                  any string not starting with "="
//	=$.a.b[0].c              value at path in the event payload
//	=sum($.x[*].y)           money sum over decimal strings at path, in minor units
//	=ref(type, field, $.p)   object_id of the single object of that type whose
//	                         field equals the value at path (ref<type> fields)
//
// Paths: $.seg, seg[N], seg[*]. A [*] fans out into a slice of values.

type Template struct {
	raw      string
	kind     string // "literal" | "path" | "sum" | "ref"
	path     []pathSeg
	refType  string // ref only: target object type
	refField string // ref only: field matched against the path's value
}

// Lookup resolves a ref() against current object state: the object_id of the
// single object of the given type whose field equals value, or an error (none
// or several). The executor supplies it; contexts without state pass nil.
type Lookup func(objectType, field string, value any) (string, error)

type pathSeg struct {
	key   string
	index int // -1 none, -2 wildcard
	hasIx bool
}

func ParseTemplate(s string) (Template, error) {
	if !strings.HasPrefix(s, "=") {
		return Template{raw: s, kind: "literal"}, nil
	}
	body := strings.TrimSpace(s[1:])
	if strings.HasPrefix(body, "ref(") && strings.HasSuffix(body, ")") {
		parts := strings.SplitN(body[4:len(body)-1], ",", 3)
		if len(parts) != 3 {
			return Template{}, fmt.Errorf("ref() wants (type, field, $.path)")
		}
		typ, field := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if typ == "" || field == "" {
			return Template{}, fmt.Errorf("ref() wants (type, field, $.path)")
		}
		p, err := parsePath(strings.TrimSpace(parts[2]))
		if err != nil {
			return Template{}, err
		}
		return Template{raw: s, kind: "ref", refType: typ, refField: field, path: p}, nil
	}
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

// Eval evaluates the template against a decoded payload. The field def guides
// coercion and checking: "money" expects decimal strings and yields int64
// minor units, "enum" admits only the declared values, ref fields resolve
// through lookup into the target's object_id.
func (t Template) Eval(payload any, fd FieldDef, lookup Lookup) (any, error) {
	switch t.kind {
	case "literal":
		return checked(t.raw, fd)
	case "ref":
		if lookup == nil {
			return nil, fmt.Errorf("ref() needs object state, none available here")
		}
		v, err := resolve(payload, t.path)
		if err != nil {
			return nil, err
		}
		return lookup(t.refType, t.refField, v)
	case "path":
		v, err := resolve(payload, t.path)
		if err != nil {
			return nil, err
		}
		if fd.Type == "money" {
			s, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("money field wants a decimal string, got %T", v)
			}
			return ParseMoney(s)
		}
		return checked(v, fd)
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

// checked enforces per-type value constraints that need the field definition.
func checked(v any, fd FieldDef) (any, error) {
	if fd.Type == "enum" {
		s, ok := v.(string)
		if !ok || !slices.Contains(fd.Values, s) {
			return nil, fmt.Errorf("value %v is not one of %v", v, fd.Values)
		}
	}
	return v, nil
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
