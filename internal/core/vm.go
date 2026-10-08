package core

import (
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"
	"time"
)

// The formula VM. Numbers are exact rationals (math/big), so every
// intermediate is exact and precision is lost only where a formula says so:
// round(x, places, method), or storing into a field whose scale the value
// already fits. No floats, ever (invariant 6). Evaluation is total: the only
// backward jump is a fold's, bounded by its collection, and a step budget
// guards the mechanism itself. Every leaf read is recorded as an input, so
// the walk can show which values and which formula produced a number
// (invariant 5 extended to computed values).

// Value is a runtime value.
type Value struct {
	Kind Kind
	Num  *big.Rat // Int, Money (major units), Decimal
	Str  string   // String, Date, Ref (the id)
	Bool bool
	Raw  any     // Object: its map; List: the raw slice
	List []Value // List: the elements
	Type string  // Ref/Object: the object type; List: the element type, when known
	// Origin is the concrete path the value was read at — $.lines[2].qty,
	// $.state.item.std_cost, stock_movement-12.qty — for the inputs record.
	Origin string
	// untyped marks a number read from an untyped payload: in a bare copy
	// into a money field it keeps the boundary convention (minor units).
	untyped bool
}

// Env is what a formula may reach beyond its payload: lookups and state
// reads (nil where there is no state), the catalog that types object fields,
// which payload prefixes hold typed object state ("$.state": "pz_line"),
// and the collection cap.
type Env struct {
	Lookup        Lookup
	Get           Getter
	Types         Catalog
	Typed         map[string]string
	MaxCollection int
}

// DefaultMaxCollection caps a linked collection read (objects()): the
// cascade depth cap's sibling. Above it the firing is refused into the
// worklist, naming the read — scale rule 2: reads are keyed and bounded,
// never scans.
const DefaultMaxCollection = 1000

// maxSteps bounds one evaluation mechanically. Structural totality (no
// loops but folds over capped collections) makes this unreachable; it stays
// as the proof's belt and braces.
const maxSteps = 2_000_000

// Calc is the explanation of one computed value, baked into the derived
// event beside the value: the formula as written and every input it read,
// by concrete path. Replay never recomputes; the walk shows the arithmetic.
type Calc struct {
	Formula string         `json:"formula"`
	Inputs  map[string]any `json:"inputs"`
}

type iterator struct {
	elems []Value
	pos   int
	kind  string
	acc   Value
	has   bool
}

type vm struct {
	env    *Env
	root   any
	stack  []Value
	slots  []Value
	iters  []iterator
	inputs map[string]any
	steps  int
}

// run executes a program against a payload and returns the result and the
// inputs it read.
func (f *Formula) run(payload any, env *Env) (Value, map[string]any, error) {
	if env == nil {
		env = &Env{}
	}
	m := &vm{env: env, root: payload, slots: make([]Value, f.prog.slots),
		iters: make([]iterator, f.prog.iters), inputs: map[string]any{}}
	v, err := m.exec(f.prog)
	if err != nil {
		return Value{}, nil, err
	}
	return v, m.inputs, nil
}

func (m *vm) push(v Value) { m.stack = append(m.stack, v) }
func (m *vm) pop() Value {
	v := m.stack[len(m.stack)-1]
	m.stack = m.stack[:len(m.stack)-1]
	return v
}

func (m *vm) exec(p *program) (Value, error) {
	pc := 0
	for {
		m.steps++
		if m.steps > maxSteps {
			return Value{}, fmt.Errorf("formula exceeded %d steps", maxSteps)
		}
		in := p.code[pc]
		pc++
		switch in.op {
		case opConst:
			m.push(p.consts[in.a])
		case opPath:
			v, err := m.load(p.paths[in.a])
			if err != nil {
				return Value{}, err
			}
			m.push(v)
		case opVar:
			m.push(m.slots[in.a])
		case opSetVar:
			m.slots[in.a] = m.pop()
		case opField:
			v, err := m.member(m.pop(), p.names[in.a])
			if err != nil {
				return Value{}, err
			}
			m.record(v)
			m.push(v)
		case opNeg:
			v := m.pop()
			r, k, err := toNum(v)
			if err != nil {
				return Value{}, err
			}
			m.push(Value{Kind: k, Num: new(big.Rat).Neg(r)})
		case opNot:
			v := m.pop()
			if v.Kind != KBool {
				return Value{}, fmt.Errorf("not wants a condition, got %s", v.Kind)
			}
			m.push(Value{Kind: KBool, Bool: !v.Bool})
		case opBin:
			b, a := m.pop(), m.pop()
			v, err := binop(p.names[in.a], a, b)
			if err != nil {
				return Value{}, err
			}
			m.push(v)
		case opJmp:
			pc = in.a
		case opJmpF:
			v := m.pop()
			if v.Kind != KBool {
				return Value{}, fmt.Errorf("condition wants true or false, got %s", v.Kind)
			}
			if !v.Bool {
				pc = in.a
			}
		case opJmpFKeep, opJmpTKeep:
			v := m.stack[len(m.stack)-1]
			if v.Kind != KBool {
				return Value{}, fmt.Errorf("and/or want conditions, got %s", v.Kind)
			}
			if v.Bool == (in.op == opJmpTKeep) {
				pc = in.a
			} else {
				m.pop()
			}
		case opCall:
			args := make([]Value, in.b)
			for i := in.b - 1; i >= 0; i-- {
				args[i] = m.pop()
			}
			v, err := call(p.names[in.a], args)
			if err != nil {
				return Value{}, err
			}
			m.push(v)
		case opRef:
			v, err := m.ref(p.names[in.a], p.names[in.b], m.pop())
			if err != nil {
				return Value{}, err
			}
			m.push(v)
		case opObjects:
			v, err := m.objects(p.names[in.a], p.names[in.b], m.pop())
			if err != nil {
				return Value{}, err
			}
			m.push(v)
		case opIter:
			coll := m.pop()
			if coll.Kind != KList {
				return Value{}, fmt.Errorf("%s iterates a collection, got %s", p.names[in.b], coll.Kind)
			}
			m.iters[in.a] = iterator{elems: coll.List, kind: p.names[in.b]}
		case opNext:
			it := &m.iters[in.a]
			if it.pos >= len(it.elems) {
				pc = in.b
			} else {
				m.slots[in.c] = it.elems[it.pos]
				it.pos++
			}
		case opAcc:
			if err := m.iters[in.a].accumulate(m.pop()); err != nil {
				return Value{}, err
			}
		case opAccResult:
			v, err := m.iters[in.a].result()
			if err != nil {
				return Value{}, err
			}
			m.push(v)
		case opReturn:
			if len(m.stack) != 1 {
				return Value{}, fmt.Errorf("formula left %d values on the stack", len(m.stack))
			}
			return m.pop(), nil
		}
	}
}

// record notes a leaf read as an input, by its concrete path.
func (m *vm) record(v Value) {
	if v.Kind == KList || v.Kind == KObject || v.Origin == "" {
		return
	}
	m.inputs[v.Origin] = jsonOf(v)
}

// load walks the payload along a path. Typed prefixes turn a map into
// object state typed by the catalog; from there field kinds come from the
// type, and a ref field's id is followed through the link when the path
// continues past it. A [*] fans out into a list.
func (m *vm) load(path []pathSeg) (Value, error) {
	root := Value{Kind: KObject, Raw: m.root, Origin: "$", Type: m.env.Typed["$"]}
	v, err := m.walk(root, path, "$")
	if err != nil {
		return Value{}, err
	}
	m.record(v)
	return v, nil
}

func (m *vm) walk(cur Value, path []pathSeg, keys string) (Value, error) {
	for i, seg := range path {
		keys += "." + seg.key
		next, err := m.member(cur, seg.key)
		if err != nil {
			return Value{}, err
		}
		if t, ok := m.env.Typed[keys]; ok && next.Kind == KObject {
			next.Type = t
		}
		if seg.hasIx {
			if next.Kind != KList {
				return Value{}, fmt.Errorf("%s is not a list", next.Origin)
			}
			if seg.index == -2 {
				rest := path[i+1:]
				out := Value{Kind: KList, Origin: next.Origin + "[*]", List: make([]Value, 0, len(next.List))}
				raws := make([]any, 0, len(next.List))
				for _, el := range next.List {
					rv := el
					if len(rest) > 0 {
						if rv, err = m.walk(el, rest, keys); err != nil {
							return Value{}, err
						}
					}
					m.record(rv)
					out.List = append(out.List, rv)
					raws = append(raws, rawOf(rv))
				}
				out.Raw = raws
				return out, nil
			}
			if seg.index >= len(next.List) {
				return Value{}, fmt.Errorf("%s: index %d out of range", next.Origin, seg.index)
			}
			next = next.List[seg.index]
		}
		cur = next
	}
	return cur, nil
}

// member reads one field: of an object (typed by the catalog when the
// object's type is known), or of the object behind a ref — the read through
// a link, which needs state and is itself recorded as an input.
func (m *vm) member(v Value, key string) (Value, error) {
	switch v.Kind {
	case KObject:
		mp, ok := v.Raw.(map[string]any)
		if !ok {
			return Value{}, fmt.Errorf("%s is not an object", v.Origin)
		}
		raw, ok := mp[key]
		if !ok {
			return Value{}, fmt.Errorf("%s.%s: missing", v.Origin, key)
		}
		origin := v.Origin + "." + key
		if v.Type != "" && m.env.Types != nil {
			if ot, ok := m.env.Types(v.Type); ok {
				if fd, ok := ot.Field(key); ok {
					return fromTyped(raw, fd, origin)
				}
			}
		}
		return fromJSON(raw, origin, ""), nil
	case KRef:
		if m.env.Get == nil {
			return Value{}, fmt.Errorf("%s.%s: reading through a link needs object state, none available here", v.Origin, key)
		}
		state, ok, err := m.env.Get(v.Str)
		if err != nil {
			return Value{}, err
		}
		if !ok {
			return Value{}, fmt.Errorf("%s: no %s %q", v.Origin, v.Type, v.Str)
		}
		m.record(v) // the link itself: which object was read
		obj := Value{Kind: KObject, Raw: state, Type: v.Type, Origin: v.Origin}
		return m.member(obj, key)
	case KNull:
		return Value{}, fmt.Errorf("%s.%s: %s is missing", v.Origin, key, v.Origin)
	}
	return Value{}, fmt.Errorf("cannot read field %q of %s (%s) — fields are read from objects, or through a link", key, v.Origin, v.Kind)
}

func (m *vm) ref(typ, field string, v Value) (Value, error) {
	if m.env.Lookup == nil {
		return Value{}, fmt.Errorf("ref() needs object state, none available here")
	}
	key, err := keyText(v)
	if err != nil {
		return Value{}, fmt.Errorf("ref(%s, %s, …): %w", typ, field, err)
	}
	id, err := ResolveRef(m.env.Lookup, typ, field, key)
	if err != nil {
		return Value{}, err
	}
	origin := fmt.Sprintf("ref(%s, %s, %s)", typ, field, key)
	m.inputs[origin] = id
	return Value{Kind: KRef, Str: id, Type: typ, Origin: id}, nil
}

// objects reads a keyed collection of state: every typ whose field equals
// the value, in log order, refused above the cap.
func (m *vm) objects(typ, field string, v Value) (Value, error) {
	if m.env.Lookup == nil || m.env.Get == nil {
		return Value{}, fmt.Errorf("objects() needs object state, none available here")
	}
	key, err := keyText(v)
	if err != nil {
		return Value{}, fmt.Errorf("objects(%s, %s, …): %w", typ, field, err)
	}
	ids, err := m.env.Lookup(typ, field, key)
	if err != nil {
		return Value{}, err
	}
	capacity := m.env.MaxCollection
	if capacity <= 0 {
		capacity = DefaultMaxCollection
	}
	origin := fmt.Sprintf("objects(%s, %s, %s)", typ, field, key)
	if len(ids) > capacity {
		return Value{}, fmt.Errorf("%s holds %d objects, above the cap of %d — a formula walks a bounded collection", origin, len(ids), capacity)
	}
	out := Value{Kind: KList, Type: typ, Origin: origin, List: make([]Value, 0, len(ids))}
	raws := make([]any, 0, len(ids))
	for _, id := range ids {
		state, ok, err := m.env.Get(id)
		if err != nil {
			return Value{}, err
		}
		if !ok {
			return Value{}, fmt.Errorf("%s: %s resolved but cannot be read", origin, id)
		}
		out.List = append(out.List, Value{Kind: KObject, Raw: state, Type: typ, Origin: id})
		raws = append(raws, state)
	}
	out.Raw = raws
	m.inputs[origin] = ids
	return out, nil
}

// keyText is a lookup key as the store compares it: the stored text —
// money as minor units, numbers as their canonical decimal.
func keyText(v Value) (string, error) {
	switch v.Kind {
	case KString, KDate, KRef:
		return v.Str, nil
	case KMoney:
		if !fits(v.Num, 2) {
			return "", fmt.Errorf("%s does not fit a money field (two places)", describe(v.Num))
		}
		minor := new(big.Int).Quo(new(big.Int).Mul(v.Num.Num(), big.NewInt(100)), v.Num.Denom())
		return minor.String(), nil
	case KInt, KDecimal:
		return canonical(v.Num)
	case KBool:
		return fmt.Sprint(v.Bool), nil
	}
	return "", fmt.Errorf("a %s cannot be a lookup key", v.Kind)
}

// ---- reading JSON into values ------------------------------------------------

// fromTyped reads a field of a typed object by its declared type: money is
// minor units in state, decimals and dates are strings, refs are ids.
func fromTyped(raw any, fd FieldDef, origin string) (Value, error) {
	bad := func() (Value, error) {
		return Value{}, fmt.Errorf("%s is declared %s but holds %v", origin, fd.Type, raw)
	}
	switch fd.Type {
	case "int":
		n, ok := integral(raw)
		if !ok {
			return bad()
		}
		return Value{Kind: KInt, Num: ratInt(n), Origin: origin}, nil
	case "money":
		switch x := raw.(type) {
		case string:
			minor, err := ParseMoney(x)
			if err != nil {
				return bad()
			}
			return Value{Kind: KMoney, Num: new(big.Rat).SetFrac64(minor, 100), Origin: origin}, nil
		}
		n, ok := integral(raw)
		if !ok {
			return bad()
		}
		return Value{Kind: KMoney, Num: new(big.Rat).SetFrac64(n, 100), Origin: origin}, nil
	case "decimal":
		switch x := raw.(type) {
		case string:
			r, err := parseDecimal(x)
			if err != nil {
				return bad()
			}
			return Value{Kind: KDecimal, Num: r, Origin: origin}, nil
		}
		n, ok := integral(raw)
		if !ok {
			return bad()
		}
		return Value{Kind: KDecimal, Num: ratInt(n), Origin: origin}, nil
	case "date":
		s, ok := raw.(string)
		if !ok {
			return bad()
		}
		return Value{Kind: KDate, Str: s, Origin: origin}, nil
	case "string", "enum":
		switch x := raw.(type) {
		case string:
			return Value{Kind: KString, Str: x, Origin: origin}, nil
		case nil:
			return Value{Kind: KNull, Origin: origin}, nil
		}
		return Value{Kind: KString, Str: fmt.Sprint(raw), Origin: origin}, nil
	}
	if target, ok := RefTarget(fd.Type); ok {
		s, ok := raw.(string)
		if !ok {
			return bad()
		}
		return Value{Kind: KRef, Str: s, Type: target, Origin: origin}, nil
	}
	return fromJSON(raw, origin, ""), nil
}

// fromJSON reads an untyped value by its JSON shape. A number is what it
// says; a non-integral JSON number is refused (decimals travel as strings).
func fromJSON(raw any, origin, typ string) Value {
	switch x := raw.(type) {
	case nil:
		return Value{Kind: KNull, Origin: origin}
	case bool:
		return Value{Kind: KBool, Bool: x, Origin: origin}
	case string:
		return Value{Kind: KString, Str: x, Origin: origin}
	case map[string]any:
		return Value{Kind: KObject, Raw: x, Origin: origin, Type: typ}
	case []any:
		out := Value{Kind: KList, Raw: x, Origin: origin, List: make([]Value, 0, len(x))}
		for i, el := range x {
			out.List = append(out.List, fromJSON(el, fmt.Sprintf("%s[%d]", origin, i), ""))
		}
		return out
	}
	if n, ok := integral(raw); ok {
		return Value{Kind: KInt, Num: ratInt(n), Origin: origin, untyped: true}
	}
	// a float with a fraction: carried as a string so the error names it
	return Value{Kind: KString, Str: fmt.Sprintf("%v", raw), Origin: origin, untyped: true}
}

func integral(raw any) (int64, bool) {
	switch n := raw.(type) {
	case int:
		return int64(n), true
	case int64:
		return n, true
	case float64:
		if n == math.Trunc(n) && math.Abs(n) < 1<<53 {
			return int64(n), true
		}
	}
	return 0, false
}

func rawOf(v Value) any {
	switch v.Kind {
	case KObject, KList:
		return v.Raw
	}
	return jsonOf(v)
}

// ---- numbers --------------------------------------------------------------------

func ratInt(n int64) *big.Rat { return new(big.Rat).SetInt64(n) }

// parseDecimal accepts a plain decimal: optional sign, digits, optional
// fraction. Nothing else — no exponents, no fractions, no floats.
func parseDecimal(s string) (*big.Rat, error) {
	t := strings.TrimSpace(s)
	body := strings.TrimPrefix(t, "-")
	whole, frac, _ := strings.Cut(body, ".")
	if whole == "" || !digits(whole) || (frac != "" && !digits(frac)) || strings.Count(body, ".") > 1 {
		return nil, fmt.Errorf("%q is not a decimal number", s)
	}
	r, ok := new(big.Rat).SetString(t)
	if !ok {
		return nil, fmt.Errorf("%q is not a decimal number", s)
	}
	return r, nil
}

func digits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// toNum reads a value as a number: numeric kinds as they are, a string as
// the decimal it spells (the boundary form).
func toNum(v Value) (*big.Rat, Kind, error) {
	switch v.Kind {
	case KInt, KMoney, KDecimal:
		return v.Num, v.Kind, nil
	case KString:
		r, err := parseDecimal(v.Str)
		if err != nil {
			return nil, KAny, fmt.Errorf("%s: %w", orDefault(v.Origin, "value"), err)
		}
		return r, KDecimal, nil
	case KNull:
		return nil, KAny, fmt.Errorf("%s is missing", orDefault(v.Origin, "a value"))
	}
	return nil, KAny, fmt.Errorf("%s is a %s, not a number", orDefault(v.Origin, "value"), v.Kind)
}

func orDefault(s, d string) string {
	if s == "" {
		return d
	}
	return s
}

// fits reports whether r is exactly representable with the given decimal
// places.
func fits(r *big.Rat, places int) bool {
	scaled := new(big.Int).Mul(r.Num(), pow10(places))
	return new(big.Int).Mod(scaled, r.Denom()).Sign() == 0
}

func pow10(n int) *big.Int { return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil) }

// canonical renders a rational as the shortest exact decimal, or fails when
// none exists (a third has no decimal: declare rounding).
func canonical(r *big.Rat) (string, error) {
	for p := 0; p <= 18; p++ {
		if fits(r, p) {
			return r.FloatString(p), nil
		}
	}
	return "", fmt.Errorf("%s is not a terminating decimal — declare round(…, places, method)", r.RatString())
}

// roundRat rounds to places by the named method. Methods are integer
// definitions over the scaled quotient and remainder: half_up is half away
// from zero (convert's method), half_even ties to even, down is toward zero,
// up is away from zero.
func roundRat(r *big.Rat, places int, method string) *big.Rat {
	scale := pow10(places)
	num := new(big.Int).Mul(r.Num(), scale)
	den := r.Denom()
	q, rem := new(big.Int).QuoRem(num, den, new(big.Int))
	if rem.Sign() != 0 {
		twice := new(big.Int).Mul(new(big.Int).Abs(rem), big.NewInt(2))
		half := twice.Cmp(den) // <0 below half, 0 at half, >0 above
		away := func() {
			if num.Sign() < 0 {
				q.Sub(q, big.NewInt(1))
			} else {
				q.Add(q, big.NewInt(1))
			}
		}
		switch method {
		case "half_up":
			if half >= 0 {
				away()
			}
		case "half_even":
			if half > 0 || (half == 0 && q.Bit(0) == 1) {
				away()
			}
		case "up":
			away()
		case "down":
		}
	}
	return new(big.Rat).SetFrac(q, scale)
}

// ---- operators ----------------------------------------------------------------------

func binop(op string, a, b Value) (Value, error) {
	switch op {
	case "=", "<>", "<", "<=", ">", ">=":
		return compare(op, a, b)
	}
	if a.Kind == KDate || b.Kind == KDate || looksDate(a) || looksDate(b) {
		return dateArith(op, a, b)
	}
	x, ka, err := toNum(a)
	if err != nil {
		return Value{}, err
	}
	y, kb, err := toNum(b)
	if err != nil {
		return Value{}, err
	}
	k, err := arithKind(op, ka, kb)
	if err != nil {
		return Value{}, err
	}
	r := new(big.Rat)
	switch op {
	case "+":
		r.Add(x, y)
	case "-":
		r.Sub(x, y)
	case "*":
		r.Mul(x, y)
	case "/":
		if y.Sign() == 0 {
			return Value{}, fmt.Errorf("division by zero (%s)", orDefault(b.Origin, "divisor"))
		}
		r.Quo(x, y)
	default:
		return Value{}, fmt.Errorf("unknown operator %q", op)
	}
	return Value{Kind: k, Num: r}, nil
}

func looksDate(v Value) bool {
	if v.Kind != KString || len(v.Str) != 10 {
		return false
	}
	_, err := time.Parse("2006-01-02", v.Str)
	return err == nil
}

func toDate(v Value) (time.Time, error) {
	if v.Kind != KDate && v.Kind != KString {
		return time.Time{}, fmt.Errorf("%s is a %s, not a date", orDefault(v.Origin, "value"), v.Kind)
	}
	t, err := time.Parse("2006-01-02", v.Str)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s: %q is not a date (YYYY-MM-DD)", orDefault(v.Origin, "value"), v.Str)
	}
	return t, nil
}

func toInt(v Value) (int64, error) {
	r, _, err := toNum(v)
	if err != nil {
		return 0, err
	}
	if !r.IsInt() || !r.Num().IsInt64() {
		return 0, fmt.Errorf("%s: %s is not a whole number", orDefault(v.Origin, "value"), r.RatString())
	}
	return r.Num().Int64(), nil
}

func dateArith(op string, a, b Value) (Value, error) {
	if op == "*" || op == "/" {
		return Value{}, fmt.Errorf("dates take + or - whole days, or date - date: not %s", op)
	}
	d, err := toDate(a)
	if err != nil {
		if op == "+" && (b.Kind == KDate || looksDate(b)) { // int + date
			return dateArith(op, b, a)
		}
		return Value{}, err
	}
	switch {
	case op == "-" && (b.Kind == KDate || looksDate(b)):
		e, err := toDate(b)
		if err != nil {
			return Value{}, err
		}
		days := (d.Unix() - e.Unix()) / 86400
		return Value{Kind: KInt, Num: ratInt(days)}, nil
	case op == "+" || op == "-":
		n, err := toInt(b)
		if err != nil {
			return Value{}, fmt.Errorf("a date takes whole days: %w", err)
		}
		if op == "-" {
			n = -n
		}
		return Value{Kind: KDate, Str: d.AddDate(0, 0, int(n)).Format("2006-01-02")}, nil
	}
	return Value{}, fmt.Errorf("dates take + or - whole days, or date - date: not %s", op)
}

func compare(op string, a, b Value) (Value, error) {
	var c int
	switch {
	case a.Kind == KNull || b.Kind == KNull:
		eq := a.Kind == KNull && b.Kind == KNull
		switch op {
		case "=":
			return Value{Kind: KBool, Bool: eq}, nil
		case "<>":
			return Value{Kind: KBool, Bool: !eq}, nil
		}
		return Value{}, fmt.Errorf("%s is missing", orDefault(a.Origin, "a value"))
	case a.Kind == KBool || b.Kind == KBool:
		if a.Kind != b.Kind || (op != "=" && op != "<>") {
			return Value{}, fmt.Errorf("cannot compare %s with %s", a.Kind, b.Kind)
		}
		c = 1
		if a.Bool == b.Bool {
			c = 0
		}
	case a.Kind.numeric() || b.Kind.numeric():
		x, _, err := toNum(a)
		if err != nil {
			return Value{}, err
		}
		y, _, err := toNum(b)
		if err != nil {
			return Value{}, err
		}
		c = x.Cmp(y)
	case a.Kind == KList || a.Kind == KObject || b.Kind == KList || b.Kind == KObject:
		return Value{}, fmt.Errorf("cannot compare %s with %s", a.Kind, b.Kind)
	default: // strings, dates, ids: text order (ISO dates order lexically)
		c = strings.Compare(a.Str, b.Str)
	}
	var out bool
	switch op {
	case "=":
		out = c == 0
	case "<>":
		out = c != 0
	case "<":
		out = c < 0
	case "<=":
		out = c <= 0
	case ">":
		out = c > 0
	case ">=":
		out = c >= 0
	}
	return Value{Kind: KBool, Bool: out}, nil
}

// ---- functions -------------------------------------------------------------------------

func call(name string, args []Value) (Value, error) {
	switch name {
	case "round":
		r, k, err := toNum(args[0])
		if err != nil {
			return Value{}, err
		}
		places, _ := toInt(args[1])
		if k == KInt {
			k = KDecimal
		}
		return Value{Kind: k, Num: roundRat(r, int(places), args[2].Str)}, nil
	case "abs":
		r, k, err := toNum(args[0])
		if err != nil {
			return Value{}, err
		}
		return Value{Kind: k, Num: new(big.Rat).Abs(r)}, nil
	case "min", "max":
		less, err := compare("<", args[0], args[1])
		if err != nil {
			return Value{}, err
		}
		if less.Bool == (name == "min") {
			return args[0], nil
		}
		return args[1], nil
	case "end_of_month":
		d, err := toDate(args[0])
		if err != nil {
			return Value{}, err
		}
		first := time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, time.UTC)
		return Value{Kind: KDate, Str: first.AddDate(0, 1, -1).Format("2006-01-02")}, nil
	case "add_months":
		d, err := toDate(args[0])
		if err != nil {
			return Value{}, err
		}
		n, err := toInt(args[1])
		if err != nil {
			return Value{}, err
		}
		// the day clamps to the target month's length: Jan 31 + 1 month is Feb 28
		first := time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, int(n), 0)
		last := first.AddDate(0, 1, -1).Day()
		day := d.Day()
		if day > last {
			day = last
		}
		return Value{Kind: KDate, Str: time.Date(first.Year(), first.Month(), day, 0, 0, 0, 0, time.UTC).Format("2006-01-02")}, nil
	case "sum": // the legacy form over a fanned-out list
		if args[0].Kind != KList {
			return Value{}, fmt.Errorf("sum() needs a list, got %s", args[0].Kind)
		}
		it := iterator{kind: "sum"}
		for _, el := range args[0].List {
			if err := it.accumulate(el); err != nil {
				return Value{}, err
			}
		}
		return it.result()
	}
	return Value{}, fmt.Errorf("unknown function %q", name)
}

// ---- folds -------------------------------------------------------------------------------

func (it *iterator) accumulate(v Value) error {
	switch it.kind {
	case "sum", "count":
		r, k, err := toNum(v)
		if err != nil {
			return err
		}
		if !it.has {
			it.acc = Value{Kind: k, Num: new(big.Rat).Set(r)}
			it.has = true
			return nil
		}
		kind, err := arithKind("+", it.acc.Kind, k)
		if err != nil {
			return err
		}
		it.acc = Value{Kind: kind, Num: new(big.Rat).Add(it.acc.Num, r)}
	case "min", "max":
		if !it.has {
			it.acc, it.has = v, true
			return nil
		}
		better, err := compare("<", v, it.acc)
		if err != nil {
			return err
		}
		if better.Bool == (it.kind == "min") {
			it.acc = v
		}
	case "any", "all":
		if v.Kind != KBool {
			return fmt.Errorf("%s wants a condition, got %s", it.kind, v.Kind)
		}
		if !it.has {
			it.acc, it.has = v, true
			return nil
		}
		if it.kind == "any" {
			it.acc.Bool = it.acc.Bool || v.Bool
		} else {
			it.acc.Bool = it.acc.Bool && v.Bool
		}
	}
	return nil
}

func (it *iterator) result() (Value, error) {
	if it.has {
		return it.acc, nil
	}
	switch it.kind {
	case "sum", "count":
		return Value{Kind: KInt, Num: ratInt(0)}, nil
	case "any":
		return Value{Kind: KBool, Bool: false}, nil
	case "all":
		return Value{Kind: KBool, Bool: true}, nil
	}
	return Value{}, fmt.Errorf("%s over an empty collection has no value", it.kind)
}

// ---- out: values into fields and into the record -----------------------------------------

// jsonOf renders a value for the inputs record: money and decimals as
// decimal strings, whole numbers as numbers, the rest as they are.
func jsonOf(v Value) any {
	switch v.Kind {
	case KInt:
		if v.Num.IsInt() && v.Num.Num().IsInt64() {
			return v.Num.Num().Int64()
		}
		s, _ := canonical(v.Num)
		return s
	case KMoney:
		if fits(v.Num, 2) {
			return v.Num.FloatString(2)
		}
		return v.Num.RatString()
	case KDecimal:
		s, err := canonical(v.Num)
		if err != nil {
			return v.Num.RatString()
		}
		return s
	case KString, KDate, KRef:
		return v.Str
	case KBool:
		return v.Bool
	case KList:
		return v.Raw
	case KObject:
		return v.Raw
	}
	return nil
}

// toField converts a result into the value a field of the given type
// stores, refusing what does not fit: a money field holds two places, an
// int a whole number — any finer value must have been rounded on purpose.
// copy marks a bare path: untyped numbers then keep the boundary convention
// (minor units), exactly as copies always worked.
func toField(v Value, fd FieldDef, copy bool) (any, error) {
	switch fd.Type {
	case "":
		return natural(v), nil
	case "money":
		if v.Kind == KNull {
			return nil, fmt.Errorf("%s is missing", orDefault(v.Origin, "the value"))
		}
		if copy && v.Kind == KInt && v.untyped {
			return v.Num.Num().Int64(), nil // minor units, the derived-payload form
		}
		if copy && v.Kind == KString {
			return ParseMoney(v.Str) // the boundary form, strictly
		}
		r, _, err := toNum(v)
		if err != nil {
			return nil, err
		}
		if !fits(r, 2) {
			return nil, fmt.Errorf("%s does not fit a money field (two places) — declare round(…, 2, half_up)", describe(r))
		}
		minor := new(big.Int).Quo(new(big.Int).Mul(r.Num(), big.NewInt(100)), r.Denom())
		if !minor.IsInt64() {
			return nil, fmt.Errorf("%s overflows money", describe(r))
		}
		return minor.Int64(), nil
	case "int":
		if v.Kind == KString && copy {
			return v.Str, nil // copies keep what they copy
		}
		n, err := toInt(v)
		if err != nil {
			return nil, err
		}
		return n, nil
	case "decimal":
		if v.Kind == KString && copy {
			if _, err := parseDecimal(v.Str); err != nil {
				return nil, err
			}
			return v.Str, nil
		}
		r, _, err := toNum(v)
		if err != nil {
			return nil, err
		}
		return canonical(r)
	case "date":
		if _, err := toDate(v); err != nil {
			return nil, err
		}
		return v.Str, nil
	case "string", "enum":
		switch v.Kind {
		case KString, KDate, KRef:
			return checked(v.Str, fd)
		case KInt, KDecimal, KMoney:
			s, err := canonical(v.Num)
			if err != nil {
				return nil, err
			}
			return checked(s, fd)
		case KBool:
			return checked(fmt.Sprint(v.Bool), fd)
		case KNull:
			if copy {
				return nil, nil
			}
		}
		return nil, fmt.Errorf("a %s cannot be stored in a %s field", v.Kind, fd.Type)
	}
	if target, ok := RefTarget(fd.Type); ok {
		switch v.Kind {
		case KRef:
			if v.Type != "" && v.Type != target {
				return nil, fmt.Errorf("%s is a %s, the field wants a %s", v.Str, v.Type, target)
			}
			return v.Str, nil
		case KString:
			return v.Str, nil // vouched by the caller
		}
		return nil, fmt.Errorf("a %s cannot fill a %s field", v.Kind, fd.Type)
	}
	return natural(v), nil
}

func describe(r *big.Rat) string {
	if s, err := canonical(r); err == nil {
		return s
	}
	return r.RatString()
}

// natural is a value as plain Go: what structural callers (each, targets)
// expect.
func natural(v Value) any {
	switch v.Kind {
	case KInt:
		if v.Num.IsInt() && v.Num.Num().IsInt64() {
			return v.Num.Num().Int64()
		}
	case KMoney:
		if fits(v.Num, 2) {
			minor := new(big.Int).Quo(new(big.Int).Mul(v.Num.Num(), big.NewInt(100)), v.Num.Denom())
			return minor.Int64()
		}
	case KList, KObject:
		return v.Raw
	}
	return jsonOf(v)
}

// SortedInputs renders a calc's inputs in a stable order, for display.
func (c Calc) SortedInputs() []string {
	keys := make([]string, 0, len(c.Inputs))
	for k := range c.Inputs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
