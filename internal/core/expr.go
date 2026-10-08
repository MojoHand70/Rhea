package core

import (
	"fmt"
	"slices"
	"strings"
)

// Templates (SPEC §2, Rule): every field value in an effect is a template.
//
//	literal                  any string not starting with "="
//	=<formula>               a formula (formula.go): a path copy, a lookup,
//	                         or a computation over the payload and the
//	                         objects it links to
//
// Four kinds matter to the validator's laws:
//
//	path     =$.a.b[0].c            a bare copy — for a ref field, an object id
//	                                carried from the event, vouched at expansion
//	ref      =ref(type, field, v)   a resolution — the id of the single object
//	                                of that type whose field equals v; v may
//	                                itself be a ref(), one step through a link
//	formula  anything else          qty × price, round(net × rate, 2, half_up),
//	                                $.state.item.std_cost, date + 14, a fold
//	literal
//
// A ref<type> field takes path or ref, never a literal id and never a
// computed one. Paths: $.seg, seg[N], seg[*]; [*] fans out into a list.

type Template struct {
	raw     string
	kind    string // "literal" | "path" | "ref" | "formula"
	f       *Formula
	path    []pathSeg // path only
	refType string    // ref only: the resolved type
}

// Lookup finds objects in current state: the object_ids of every object of
// the given type whose field equals value (compared as text), in log order.
// The executor supplies it; contexts without state pass nil. ref() demands
// exactly one match; other callers (the period lock) ask only whether any
// exist; objects() takes them all, under the cap.
type Lookup func(objectType, field string, value any) ([]string, error)

// Getter reads one object's state by id — the kernel's state-read surface
// for formulas reading through links and for sub-languages whose parameters
// name objects (the convert clause's fx_rate). ok is false when the object
// does not exist.
type Getter func(objectID string) (state map[string]any, ok bool, err error)

// ResolveRef applies ref() semantics to a lookup result: exactly one match.
func ResolveRef(lookup Lookup, objectType, field string, value any) (string, error) {
	ids, err := lookup(objectType, field, value)
	if err != nil {
		return "", err
	}
	switch len(ids) {
	case 0:
		return "", fmt.Errorf("no %s with %s = %v", objectType, field, value)
	case 1:
		return ids[0], nil
	}
	return "", fmt.Errorf("%d %s objects have %s = %v, ref is ambiguous", len(ids), objectType, field, value)
}

type pathSeg struct {
	key   string
	index int // -1 none, -2 wildcard
	hasIx bool
}

// CarriesID reports whether the template copies a value by path — for a ref
// field, an object id carried from the event (the causing object, or a ref it
// holds), which the kernel vouches for at expansion.
func (t Template) CarriesID() bool { return t.kind == "path" }

// Computes reports a formula proper: a value the walk must explain by its
// inputs, not a copy, a literal or a resolution.
func (t Template) Computes() bool { return t.kind == "formula" }

// Formula is the compiled formula behind an "=" template; nil for literals.
func (t Template) Formula() *Formula { return t.f }

func ParseTemplate(s string) (Template, error) {
	if !strings.HasPrefix(s, "=") {
		return Template{raw: s, kind: "literal"}, nil
	}
	f, err := ParseFormula(strings.TrimSpace(s[1:]))
	if err != nil {
		return Template{}, err
	}
	t := Template{raw: s, kind: "formula", f: f}
	switch {
	case f.prog.copy:
		t.kind, t.path = "path", f.tree.path
	case f.prog.refTyp != "":
		t.kind, t.refType = "ref", f.prog.refTyp
	}
	return t, nil
}

// parsePath parses a bare $.path (conditions, each).
func parsePath(s string) ([]pathSeg, error) {
	toks, err := lex(strings.TrimSpace(s))
	if err != nil {
		return nil, err
	}
	if len(toks) != 2 || toks[0].kind != tPath {
		return nil, fmt.Errorf("path must start with $.")
	}
	return toks[0].path, nil
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

// Eval evaluates the template against a decoded payload with lookups only:
// the structural contexts (each, targets, strings). Formulas that read state
// through links or walk collections need Evaluate with a full Env.
func (t Template) Eval(payload any, fd FieldDef, lookup Lookup) (any, error) {
	v, _, err := t.Evaluate(payload, fd, &Env{Lookup: lookup})
	return v, err
}

// Evaluate evaluates the template against a decoded payload. The field def
// guides coercion and checking: "money" yields int64 minor units, "int" a
// whole number, "decimal" a canonical decimal string, "enum" admits only the
// declared values, ref fields resolve into the target's object_id. A
// computed value comes with its Calc — the formula and the inputs read —
// for the derived event to carry; copies and resolutions return nil.
func (t Template) Evaluate(payload any, fd FieldDef, env *Env) (any, *Calc, error) {
	if t.kind == "literal" {
		v, err := checked(t.raw, fd)
		return v, nil, err
	}
	v, inputs, err := t.f.run(payload, env)
	if err != nil {
		return nil, nil, err
	}
	out, err := toField(v, fd, t.kind == "path")
	if err != nil {
		return nil, nil, err
	}
	if t.kind != "formula" {
		return out, nil, nil
	}
	return out, &Calc{Formula: t.raw, Inputs: inputs}, nil
}

// Check runs the typed static check of a formula template against the field
// it feeds, with a catalog and the typed payload prefixes of its rule.
func (t Template) Check(types Catalog, typed map[string]string, fd FieldDef) error {
	if t.kind == "literal" {
		return nil
	}
	return t.f.Check(types, typed, fd)
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
