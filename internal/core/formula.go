package core

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
)

// The formula language (DIRECTION 2026-10-08, "Arithmetic: a formula
// language of our own"): every "=" template is a formula. Ported from the
// sfmt architecture — tokenizer, parser, compiler, a stack VM running the
// compiled program — with Rhea's number core: exact rationals, never floats,
// rounding declared or refused. The language is total by construction: no
// loops, no recursion, no assignment; iteration is a fold over a named,
// bounded collection — the event's own payload, or objects one step through
// a link, keyed, ordered by log sequence and capped.
//
//	formula  := "=" expr
//	expr     := or
//	or       := and ("or" and)*
//	and      := not ("and" not)*
//	not      := "not" not | cmp
//	cmp      := sum (("=" | "<>" | "<" | "<=" | ">" | ">=") sum)?
//	sum      := term (("+" | "-") term)*
//	term     := unary (("*" | "/") unary)*
//	unary    := "-" unary | postfix
//	postfix  := primary ("." ident)*            a field read, through a link when the value is an id
//	primary  := number | string | true | false
//	          | $.path                          the payload: $.a.b[0].c, [*] fans out into a list
//	          | ident                           a fold's binder
//	          | ident "(" args ")"              a function, ref(), objects(), or a fold
//	          | "(" expr ")"
//	fold     := kind "(" ident "in" expr ("where" expr)? (":" expr)? ")"     kind ∈ sum count min max any all
//	          | "fold" "(" ident "in" expr "," ident "=" expr ":" expr ")"   one accumulator
//
// Functions: round(x, places, method) with method ∈ half_up half_even down up;
// if(c, a, b); abs(x); min(a, b); max(a, b); end_of_month(d); add_months(d, n);
// ref(T, field, v) — the id of the single T whose field equals v;
// objects(T, field, v) — the T objects whose field equals v, as a list.
// Division is admitted only under a round(): where precision can be lost,
// the stance is declared, never implied.

// Kind is a value's kind, static and dynamic alike. Money is sticky: money
// combined with a plain number stays money; money × money has no meaning
// and money ÷ money is a rate.
type Kind uint8

const (
	KAny     Kind = iota // static only: not inferable (raw payload reads)
	KInt                 // a whole number: a quantity, days, a count
	KMoney               // an amount in minor units; fields typed money
	KDecimal             // an exact decimal: a rate, a unit cost; raw decimal strings
	KString
	KDate // YYYY-MM-DD
	KBool
	KList
	KObject // a decoded object: a payload element or a linked object's state
	KRef    // an object id, typed <type>-…
	KNull
)

func (k Kind) String() string {
	return [...]string{"any", "int", "money", "decimal", "string", "date", "bool", "list", "object", "ref", "null"}[k]
}

func (k Kind) numeric() bool { return k == KInt || k == KMoney || k == KDecimal }

// RoundingMethods are the admitted ways to lose precision, each a fixed
// integer definition (vm.go). Others join by proof, as convert's did.
var RoundingMethods = []string{"half_up", "half_even", "down", "up"}

// FoldKinds are the bounded iterations: each walks a collection once.
var foldKinds = []string{"sum", "count", "min", "max", "any", "all", "fold"}

// ---- tokens -----------------------------------------------------------------

type tokKind uint8

const (
	tEOF tokKind = iota
	tNum
	tStr
	tIdent
	tPath
	tOp
)

type token struct {
	kind tokKind
	text string
	path []pathSeg
	pos  int
}

func lex(src string) ([]token, error) {
	var toks []token
	i := 0
	for i < len(src) {
		c := src[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == '$':
			start := i
			i++
			var segs []pathSeg
			for i < len(src) && (src[i] == '.' || src[i] == '[') {
				if src[i] == '.' {
					i++
					j := i
					for j < len(src) && isIdentChar(src[j]) {
						j++
					}
					if j == i {
						return nil, fmt.Errorf("at %d: empty path segment", i)
					}
					segs = append(segs, pathSeg{key: src[i:j], index: -1})
					i = j
					continue
				}
				// an index attaches to the previous segment
				if len(segs) == 0 {
					return nil, fmt.Errorf("at %d: index without a key", i)
				}
				j := strings.IndexByte(src[i:], ']')
				if j < 0 {
					return nil, fmt.Errorf("at %d: unterminated index", i)
				}
				ix := src[i+1 : i+j]
				seg := &segs[len(segs)-1]
				if seg.hasIx {
					return nil, fmt.Errorf("at %d: one index per segment", i)
				}
				seg.hasIx = true
				if ix == "*" {
					seg.index = -2
				} else {
					n, err := strconv.Atoi(ix)
					if err != nil || n < 0 {
						return nil, fmt.Errorf("at %d: bad index %q", i, ix)
					}
					seg.index = n
				}
				i += j + 1
			}
			if len(segs) == 0 {
				return nil, fmt.Errorf("at %d: path must start with $.", start)
			}
			toks = append(toks, token{kind: tPath, text: src[start:i], path: segs, pos: start})
		case c >= '0' && c <= '9':
			start := i
			for i < len(src) && (src[i] >= '0' && src[i] <= '9') {
				i++
			}
			if i < len(src) && src[i] == '.' && i+1 < len(src) && src[i+1] >= '0' && src[i+1] <= '9' {
				i++
				for i < len(src) && (src[i] >= '0' && src[i] <= '9') {
					i++
				}
			}
			toks = append(toks, token{kind: tNum, text: src[start:i], pos: start})
		case c == '"' || c == '\'':
			j := strings.IndexByte(src[i+1:], c)
			if j < 0 {
				return nil, fmt.Errorf("at %d: unterminated string", i)
			}
			toks = append(toks, token{kind: tStr, text: src[i+1 : i+1+j], pos: i})
			i += j + 2
		case isIdentStart(c):
			start := i
			for i < len(src) && isIdentChar(src[i]) {
				i++
			}
			toks = append(toks, token{kind: tIdent, text: src[start:i], pos: start})
		default:
			two := ""
			if i+1 < len(src) {
				two = src[i : i+2]
			}
			switch {
			case two == "<>" || two == "<=" || two == ">=":
				toks = append(toks, token{kind: tOp, text: two, pos: i})
				i += 2
			case strings.ContainsRune("+-*/=<>(),:.", rune(c)):
				toks = append(toks, token{kind: tOp, text: string(c), pos: i})
				i++
			default:
				return nil, fmt.Errorf("at %d: unexpected %q", i, string(c))
			}
		}
	}
	toks = append(toks, token{kind: tEOF, pos: len(src)})
	return toks, nil
}

func isIdentStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isIdentChar(c byte) bool { return isIdentStart(c) || (c >= '0' && c <= '9') }

// ---- syntax tree -------------------------------------------------------------

type nodeKind uint8

const (
	nNum nodeKind = iota
	nStr
	nBool
	nPath
	nVar
	nField  // args[0].text
	nUnary  // text "-" | "not"
	nBinary // text op, args[0], args[1]
	nCall   // text name, args
	nRef    // text type, field; args[0] value
	nObjs   // text type, field; args[0] value
	nFold   // text kind; binder; coll; where; body; acc; init
)

type node struct {
	kind   nodeKind
	text   string
	field  string
	path   []pathSeg
	args   []*node
	binder string
	acc    string
	coll   *node
	where  *node
	body   *node
	init   *node
	pos    int
}

type parser struct {
	toks []token
	i    int
}

func (p *parser) peek() token { return p.toks[p.i] }
func (p *parser) next() token { t := p.toks[p.i]; p.i++; return t }
func (p *parser) isOp(s string) bool {
	t := p.peek()
	return t.kind == tOp && t.text == s
}
func (p *parser) isWord(s string) bool {
	t := p.peek()
	return t.kind == tIdent && t.text == s
}
func (p *parser) expectOp(s string) error {
	if !p.isOp(s) {
		return fmt.Errorf("at %d: expected %q", p.peek().pos, s)
	}
	p.i++
	return nil
}

func parseFormula(src string) (*node, error) {
	toks, err := lex(src)
	if err != nil {
		return nil, err
	}
	p := &parser{toks: toks}
	n, err := p.or()
	if err != nil {
		return nil, err
	}
	if p.peek().kind != tEOF {
		return nil, fmt.Errorf("at %d: unexpected %q", p.peek().pos, p.peek().text)
	}
	return n, nil
}

func (p *parser) or() (*node, error) {
	l, err := p.and()
	if err != nil {
		return nil, err
	}
	for p.isWord("or") {
		pos := p.next().pos
		r, err := p.and()
		if err != nil {
			return nil, err
		}
		l = &node{kind: nBinary, text: "or", args: []*node{l, r}, pos: pos}
	}
	return l, nil
}

func (p *parser) and() (*node, error) {
	l, err := p.not()
	if err != nil {
		return nil, err
	}
	for p.isWord("and") {
		pos := p.next().pos
		r, err := p.not()
		if err != nil {
			return nil, err
		}
		l = &node{kind: nBinary, text: "and", args: []*node{l, r}, pos: pos}
	}
	return l, nil
}

func (p *parser) not() (*node, error) {
	if p.isWord("not") {
		pos := p.next().pos
		x, err := p.not()
		if err != nil {
			return nil, err
		}
		return &node{kind: nUnary, text: "not", args: []*node{x}, pos: pos}, nil
	}
	return p.cmp()
}

func (p *parser) cmp() (*node, error) {
	l, err := p.sum()
	if err != nil {
		return nil, err
	}
	t := p.peek()
	if t.kind == tOp && slices.Contains([]string{"=", "<>", "<", "<=", ">", ">="}, t.text) {
		p.next()
		r, err := p.sum()
		if err != nil {
			return nil, err
		}
		return &node{kind: nBinary, text: t.text, args: []*node{l, r}, pos: t.pos}, nil
	}
	return l, nil
}

func (p *parser) sum() (*node, error) {
	l, err := p.term()
	if err != nil {
		return nil, err
	}
	for p.isOp("+") || p.isOp("-") {
		t := p.next()
		r, err := p.term()
		if err != nil {
			return nil, err
		}
		l = &node{kind: nBinary, text: t.text, args: []*node{l, r}, pos: t.pos}
	}
	return l, nil
}

func (p *parser) term() (*node, error) {
	l, err := p.unary()
	if err != nil {
		return nil, err
	}
	for p.isOp("*") || p.isOp("/") {
		t := p.next()
		r, err := p.unary()
		if err != nil {
			return nil, err
		}
		l = &node{kind: nBinary, text: t.text, args: []*node{l, r}, pos: t.pos}
	}
	return l, nil
}

func (p *parser) unary() (*node, error) {
	if p.isOp("-") {
		pos := p.next().pos
		x, err := p.unary()
		if err != nil {
			return nil, err
		}
		return &node{kind: nUnary, text: "-", args: []*node{x}, pos: pos}, nil
	}
	return p.postfix()
}

func (p *parser) postfix() (*node, error) {
	x, err := p.primary()
	if err != nil {
		return nil, err
	}
	for p.isOp(".") {
		pos := p.next().pos
		t := p.next()
		if t.kind != tIdent {
			return nil, fmt.Errorf("at %d: expected a field name after '.'", pos)
		}
		x = &node{kind: nField, field: t.text, args: []*node{x}, pos: pos}
	}
	return x, nil
}

func (p *parser) primary() (*node, error) {
	t := p.next()
	switch t.kind {
	case tNum:
		return &node{kind: nNum, text: t.text, pos: t.pos}, nil
	case tStr:
		return &node{kind: nStr, text: t.text, pos: t.pos}, nil
	case tPath:
		return &node{kind: nPath, text: t.text, path: t.path, pos: t.pos}, nil
	case tOp:
		if t.text == "(" {
			x, err := p.or()
			if err != nil {
				return nil, err
			}
			if err := p.expectOp(")"); err != nil {
				return nil, err
			}
			return x, nil
		}
	case tIdent:
		switch t.text {
		case "true", "false":
			return &node{kind: nBool, text: t.text, pos: t.pos}, nil
		}
		if !p.isOp("(") {
			return &node{kind: nVar, text: t.text, pos: t.pos}, nil
		}
		p.next() // (
		return p.call(t)
	case tEOF:
		return nil, fmt.Errorf("at %d: unexpected end of formula", t.pos)
	}
	return nil, fmt.Errorf("at %d: unexpected %q", t.pos, t.text)
}

// call parses what follows "name(": a fold when the arguments open with
// "binder in", ref()/objects() with their two names, else a plain function.
func (p *parser) call(name token) (*node, error) {
	n := &node{kind: nCall, text: name.text, pos: name.pos}
	isFold := slices.Contains(foldKinds, name.text) &&
		p.peek().kind == tIdent && p.toks[p.i+1].kind == tIdent && p.toks[p.i+1].text == "in"
	switch {
	case isFold:
		return p.fold(n)
	case name.text == "ref" || name.text == "objects":
		typ, field := p.next(), token{}
		if typ.kind != tIdent {
			return nil, fmt.Errorf("at %d: %s() wants (type, field, value)", name.pos, name.text)
		}
		if err := p.expectOp(","); err != nil {
			return nil, err
		}
		field = p.next()
		if field.kind != tIdent {
			return nil, fmt.Errorf("at %d: %s() wants (type, field, value)", name.pos, name.text)
		}
		if err := p.expectOp(","); err != nil {
			return nil, err
		}
		v, err := p.or()
		if err != nil {
			return nil, err
		}
		if err := p.expectOp(")"); err != nil {
			return nil, err
		}
		n.kind = nRef
		if name.text == "objects" {
			n.kind = nObjs
		}
		n.text, n.field, n.args = typ.text, field.text, []*node{v}
		return n, nil
	}
	for !p.isOp(")") {
		// round's method is a bare name, not a variable
		if name.text == "round" && len(n.args) == 2 && p.peek().kind == tIdent {
			n.args = append(n.args, &node{kind: nStr, text: p.next().text, pos: name.pos})
		} else {
			a, err := p.or()
			if err != nil {
				return nil, err
			}
			n.args = append(n.args, a)
		}
		if p.isOp(",") {
			p.next()
			continue
		}
		if !p.isOp(")") {
			return nil, fmt.Errorf("at %d: expected ',' or ')' in %s()", p.peek().pos, name.text)
		}
	}
	p.next() // )
	return n, nil
}

func (p *parser) fold(n *node) (*node, error) {
	n.kind = nFold
	n.binder = p.next().text
	p.next() // in
	coll, err := p.or()
	if err != nil {
		return nil, err
	}
	n.coll = coll
	if n.text == "fold" {
		if !p.isOp(",") {
			return nil, fmt.Errorf("at %d: fold wants (v in coll, acc = init: body)", n.pos)
		}
		p.next()
		acc := p.next()
		if acc.kind != tIdent {
			return nil, fmt.Errorf("at %d: fold wants (v in coll, acc = init: body)", n.pos)
		}
		n.acc = acc.text
		if err := p.expectOp("="); err != nil {
			return nil, err
		}
		if n.init, err = p.or(); err != nil {
			return nil, err
		}
	}
	if p.isWord("where") {
		p.next()
		if n.where, err = p.or(); err != nil {
			return nil, err
		}
	}
	if p.isOp(":") {
		p.next()
		if n.body, err = p.or(); err != nil {
			return nil, err
		}
	}
	if err := p.expectOp(")"); err != nil {
		return nil, err
	}
	switch n.text {
	case "count":
		if n.body != nil {
			return nil, fmt.Errorf("at %d: count takes a where, not a body", n.pos)
		}
	case "fold":
		if n.body == nil || n.where != nil {
			return nil, fmt.Errorf("at %d: fold wants (v in coll, acc = init: body)", n.pos)
		}
	default:
		if n.body == nil {
			return nil, fmt.Errorf("at %d: %s(v in coll: expr) needs its expression", n.pos, n.text)
		}
	}
	return n, nil
}

// ---- static checking --------------------------------------------------------

// Catalog resolves object type names for typing reads through links.
type Catalog func(name string) (ObjectType, bool)

// ty is a static type: a kind and, for refs, objects and lists of objects,
// the object type involved.
type ty struct {
	kind Kind
	typ  string
}

var anyT = ty{kind: KAny}

// checkEnv is what the checker knows at draft time: the catalog, which
// payload prefixes hold typed object state, and the binders in scope.
type checkEnv struct {
	types  Catalog
	typed  map[string]string
	vars   map[string]ty
	reads  []CollectionRead
	copies bool // true when the tree is a single load: no arithmetic
}

// CollectionRead is one objects() read a formula performs: a keyed read of
// state, declared by the formula itself and capped at firing.
type CollectionRead struct {
	Type  string `json:"type"`
	Field string `json:"field"`
}

func (e *checkEnv) fieldType(objType, field string) ty {
	if e.types == nil {
		return anyT
	}
	ot, ok := e.types(objType)
	if !ok {
		return anyT
	}
	fd, ok := ot.Field(field)
	if !ok {
		return ty{kind: KNull, typ: "!" + objType + "." + field} // marks "no such field"
	}
	return fieldTy(fd)
}

func fieldTy(fd FieldDef) ty {
	switch fd.Type {
	case "int":
		return ty{kind: KInt}
	case "money":
		return ty{kind: KMoney}
	case "decimal":
		return ty{kind: KDecimal}
	case "date":
		return ty{kind: KDate}
	case "string", "enum":
		return ty{kind: KString}
	}
	if target, ok := RefTarget(fd.Type); ok {
		return ty{kind: KRef, typ: target}
	}
	return anyT
}

// check infers the static type of a tree, enforcing the laws that hold
// without data: kinds combine per the table, division sits under a round,
// folds walk lists, field reads go through objects or links, names resolve.
func (e *checkEnv) check(n *node, rounded bool) (ty, error) {
	switch n.kind {
	case nNum:
		if strings.Contains(n.text, ".") {
			return ty{kind: KDecimal}, nil
		}
		return ty{kind: KInt}, nil
	case nStr:
		return ty{kind: KString}, nil
	case nBool:
		return ty{kind: KBool}, nil
	case nPath:
		t, err := e.pathTy(n.path)
		if err != nil {
			return anyT, fmt.Errorf("at %d: %w", n.pos, err)
		}
		return t, nil
	case nVar:
		t, ok := e.vars[n.text]
		if !ok {
			return anyT, fmt.Errorf("at %d: unknown name %q — paths start with $., names are fold binders", n.pos, n.text)
		}
		return t, nil
	case nField:
		base, err := e.check(n.args[0], rounded)
		if err != nil {
			return anyT, err
		}
		switch base.kind {
		case KRef, KObject:
			if base.typ == "" {
				return anyT, nil
			}
			t := e.fieldType(base.typ, n.field)
			if t.kind == KNull {
				return anyT, fmt.Errorf("at %d: %s has no field %q", n.pos, base.typ, n.field)
			}
			return t, nil
		case KAny:
			return anyT, nil
		}
		return anyT, fmt.Errorf("at %d: cannot read field %q of a %s — fields are read from objects, or through a link (a ref field, or ref(...))", n.pos, n.field, base.kind)
	case nUnary:
		x, err := e.check(n.args[0], rounded)
		if err != nil {
			return anyT, err
		}
		if n.text == "not" {
			if x.kind != KBool && x.kind != KAny {
				return anyT, fmt.Errorf("at %d: not wants a condition, got %s", n.pos, x.kind)
			}
			return ty{kind: KBool}, nil
		}
		if !x.kind.numeric() && x.kind != KAny {
			return anyT, fmt.Errorf("at %d: cannot negate a %s", n.pos, x.kind)
		}
		return x, nil
	case nBinary:
		l, err := e.check(n.args[0], rounded)
		if err != nil {
			return anyT, err
		}
		r, err := e.check(n.args[1], rounded)
		if err != nil {
			return anyT, err
		}
		switch n.text {
		case "and", "or":
			for _, x := range []ty{l, r} {
				if x.kind != KBool && x.kind != KAny {
					return anyT, fmt.Errorf("at %d: %s wants conditions, got %s", n.pos, n.text, x.kind)
				}
			}
			return ty{kind: KBool}, nil
		case "=", "<>", "<", "<=", ">", ">=":
			if err := comparable(l.kind, r.kind); err != nil {
				return anyT, fmt.Errorf("at %d: %w", n.pos, err)
			}
			return ty{kind: KBool}, nil
		case "/":
			if !rounded {
				return anyT, fmt.Errorf("at %d: division must declare its rounding: round(a / b, places, half_up)", n.pos)
			}
		}
		if l.kind == KString || r.kind == KString {
			return anyT, fmt.Errorf("at %d: cannot compute %s %s %s — a declared string is not a number", n.pos, l.kind, n.text, r.kind)
		}
		k, err := arithKind(n.text, l.kind, r.kind)
		if err != nil {
			return anyT, fmt.Errorf("at %d: %w", n.pos, err)
		}
		return ty{kind: k}, nil
	case nRef, nObjs:
		if e.types != nil {
			if _, ok := e.types(n.text); !ok {
				return anyT, fmt.Errorf("at %d: %s() names unknown type %q", n.pos, callName(n), n.text)
			}
			if t := e.fieldType(n.text, n.field); t.kind == KNull {
				return anyT, fmt.Errorf("at %d: %s has no field %q", n.pos, n.text, n.field)
			}
		}
		if _, err := e.check(n.args[0], rounded); err != nil {
			return anyT, err
		}
		if n.kind == nObjs {
			e.reads = append(e.reads, CollectionRead{Type: n.text, Field: n.field})
			return ty{kind: KList, typ: n.text}, nil
		}
		return ty{kind: KRef, typ: n.text}, nil
	case nFold:
		coll, err := e.check(n.coll, rounded)
		if err != nil {
			return anyT, err
		}
		if coll.kind != KList && coll.kind != KAny {
			return anyT, fmt.Errorf("at %d: %s iterates a collection, got %s", n.pos, n.text, coll.kind)
		}
		if _, taken := e.vars[n.binder]; taken {
			return anyT, fmt.Errorf("at %d: name %q is already bound", n.pos, n.binder)
		}
		elem := anyT
		if coll.typ != "" {
			elem = ty{kind: KObject, typ: coll.typ}
		}
		inner := *e
		inner.vars = cloneVars(e.vars)
		inner.vars[n.binder] = elem
		if n.where != nil {
			w, err := inner.check(n.where, rounded)
			if err != nil {
				return anyT, err
			}
			if w.kind != KBool && w.kind != KAny {
				return anyT, fmt.Errorf("at %d: where wants a condition, got %s", n.pos, w.kind)
			}
		}
		var body ty
		if n.text == "fold" {
			initT, err := e.check(n.init, rounded)
			if err != nil {
				return anyT, err
			}
			if _, taken := inner.vars[n.acc]; taken {
				return anyT, fmt.Errorf("at %d: name %q is already bound", n.pos, n.acc)
			}
			inner.vars[n.acc] = initT
		}
		if n.body != nil {
			if body, err = inner.check(n.body, rounded); err != nil {
				return anyT, err
			}
		}
		e.reads = inner.reads
		switch n.text {
		case "count":
			return ty{kind: KInt}, nil
		case "any", "all":
			if body.kind != KBool && body.kind != KAny {
				return anyT, fmt.Errorf("at %d: %s wants a condition, got %s", n.pos, n.text, body.kind)
			}
			return ty{kind: KBool}, nil
		case "sum":
			if !body.kind.numeric() && body.kind != KAny {
				return anyT, fmt.Errorf("at %d: sum wants numbers, got %s", n.pos, body.kind)
			}
			return body, nil
		case "min", "max":
			if body.kind == KBool || body.kind == KList || body.kind == KObject {
				return anyT, fmt.Errorf("at %d: %s wants comparable values, got %s", n.pos, n.text, body.kind)
			}
			return body, nil
		}
		return body, nil // fold: the accumulator's final value
	case nCall:
		return e.checkCall(n, rounded)
	}
	return anyT, fmt.Errorf("at %d: unknown node", n.pos)
}

func callName(n *node) string {
	if n.kind == nObjs {
		return "objects"
	}
	return "ref"
}

func cloneVars(m map[string]ty) map[string]ty {
	out := make(map[string]ty, len(m)+1)
	for k, v := range m {
		out[k] = v
	}
	return out
}

func (e *checkEnv) checkCall(n *node, rounded bool) (ty, error) {
	argc := func(want int) error {
		if len(n.args) != want {
			return fmt.Errorf("at %d: %s() takes %d argument(s), got %d", n.pos, n.text, want, len(n.args))
		}
		return nil
	}
	switch n.text {
	case "round":
		if err := argc(3); err != nil {
			return anyT, err
		}
		x, err := e.check(n.args[0], true)
		if err != nil {
			return anyT, err
		}
		if !x.kind.numeric() && x.kind != KAny {
			return anyT, fmt.Errorf("at %d: round wants a number, got %s", n.pos, x.kind)
		}
		places := n.args[1]
		if places.kind != nNum || strings.Contains(places.text, ".") {
			return anyT, fmt.Errorf("at %d: round wants its places as a whole number", n.pos)
		}
		if p, _ := strconv.Atoi(places.text); p > 6 {
			return anyT, fmt.Errorf("at %d: round to at most 6 places", n.pos)
		}
		if n.args[2].kind != nStr || !slices.Contains(RoundingMethods, n.args[2].text) {
			return anyT, fmt.Errorf("at %d: rounding method must be one of %v", n.pos, RoundingMethods)
		}
		if x.kind == KInt {
			return ty{kind: KDecimal}, nil
		}
		return x, nil
	case "if":
		if err := argc(3); err != nil {
			return anyT, err
		}
		c, err := e.check(n.args[0], rounded)
		if err != nil {
			return anyT, err
		}
		if c.kind != KBool && c.kind != KAny {
			return anyT, fmt.Errorf("at %d: if wants a condition first, got %s", n.pos, c.kind)
		}
		a, err := e.check(n.args[1], rounded)
		if err != nil {
			return anyT, err
		}
		b, err := e.check(n.args[2], rounded)
		if err != nil {
			return anyT, err
		}
		return join(a, b), nil
	case "abs":
		if err := argc(1); err != nil {
			return anyT, err
		}
		x, err := e.check(n.args[0], rounded)
		if err != nil {
			return anyT, err
		}
		if !x.kind.numeric() && x.kind != KAny {
			return anyT, fmt.Errorf("at %d: abs wants a number, got %s", n.pos, x.kind)
		}
		return x, nil
	case "min", "max":
		if err := argc(2); err != nil {
			return anyT, err
		}
		a, err := e.check(n.args[0], rounded)
		if err != nil {
			return anyT, err
		}
		b, err := e.check(n.args[1], rounded)
		if err != nil {
			return anyT, err
		}
		if err := comparable(a.kind, b.kind); err != nil {
			return anyT, fmt.Errorf("at %d: %w", n.pos, err)
		}
		return join(a, b), nil
	case "end_of_month":
		if err := argc(1); err != nil {
			return anyT, err
		}
		d, err := e.check(n.args[0], rounded)
		if err != nil {
			return anyT, err
		}
		if d.kind != KDate && d.kind != KString && d.kind != KAny {
			return anyT, fmt.Errorf("at %d: end_of_month wants a date, got %s", n.pos, d.kind)
		}
		return ty{kind: KDate}, nil
	case "add_months":
		if err := argc(2); err != nil {
			return anyT, err
		}
		d, err := e.check(n.args[0], rounded)
		if err != nil {
			return anyT, err
		}
		if d.kind != KDate && d.kind != KString && d.kind != KAny {
			return anyT, fmt.Errorf("at %d: add_months wants a date first, got %s", n.pos, d.kind)
		}
		m, err := e.check(n.args[1], rounded)
		if err != nil {
			return anyT, err
		}
		if m.kind != KInt && m.kind != KAny {
			return anyT, fmt.Errorf("at %d: add_months wants whole months, got %s", n.pos, m.kind)
		}
		return ty{kind: KDate}, nil
	case "sum":
		// the legacy form: sum(list) over a fanned-out path
		if err := argc(1); err != nil {
			return anyT, err
		}
		l, err := e.check(n.args[0], rounded)
		if err != nil {
			return anyT, err
		}
		if l.kind != KList && l.kind != KAny {
			return anyT, fmt.Errorf("at %d: sum wants a list ($.lines[*].amount) or a fold (sum(v in $.lines: v.amount))", n.pos)
		}
		return anyT, nil
	}
	return anyT, fmt.Errorf("at %d: unknown function %q", n.pos, n.text)
}

// join is the static type of two branches: equal, or the one that is known.
func join(a, b ty) ty {
	switch {
	case a.kind == KAny:
		return b
	case b.kind == KAny:
		return a
	case a.kind == b.kind:
		return a
	case a.kind.numeric() && b.kind.numeric():
		k, _ := arithKind("+", a.kind, b.kind)
		return ty{kind: k}
	}
	return anyT
}

// pathTy types a payload path: typed prefixes name object state, and from
// there the catalog carries the walk through fields and links — a field the
// type does not declare, or a field read off a number, is refused here.
func (e *checkEnv) pathTy(path []pathSeg) (ty, error) {
	cur := anyT
	keys := "$"
	for _, seg := range path {
		keys += "." + seg.key
		if t, ok := e.typed[keys]; ok {
			cur = ty{kind: KObject, typ: t}
			continue
		}
		switch cur.kind {
		case KObject, KRef:
			if cur.typ == "" {
				cur = anyT
				continue
			}
			t := e.fieldType(cur.typ, seg.key)
			if t.kind == KNull {
				return anyT, fmt.Errorf("%s has no field %q", cur.typ, seg.key)
			}
			cur = t
		case KAny, KList:
			cur = anyT
		default:
			return anyT, fmt.Errorf("cannot read field %q of a %s — fields are read from objects, or through a link", seg.key, cur.kind)
		}
		if seg.hasIx {
			cur = anyT
		}
	}
	return cur, nil
}

func comparable(a, b Kind) error {
	if a == KAny || b == KAny || a == b {
		return nil
	}
	if a.numeric() && b.numeric() {
		return nil
	}
	if (a == KString && (b.numeric() || b == KDate || b == KRef)) || (b == KString && (a.numeric() || a == KDate || a == KRef)) {
		return nil
	}
	return fmt.Errorf("cannot compare %s with %s", a, b)
}

// arithKind is the kinds table: what combining two kinds yields, or why not.
func arithKind(op string, a, b Kind) (Kind, error) {
	// an untyped string is a decimal in waiting (the boundary form)
	if a == KString {
		a = KDecimal
	}
	if b == KString {
		b = KDecimal
	}
	if a == KAny || b == KAny {
		switch {
		case a == KMoney || b == KMoney:
			if op == "/" || op == "*" {
				return KAny, nil
			}
			return KMoney, nil
		}
		return KAny, nil
	}
	if a == KDate || b == KDate {
		switch {
		case op == "+" && a == KDate && b == KInt, op == "-" && a == KDate && b == KInt:
			return KDate, nil
		case op == "+" && a == KInt && b == KDate:
			return KDate, nil
		case op == "-" && a == KDate && b == KDate:
			return KInt, nil
		}
		return KAny, fmt.Errorf("dates take + or - whole days, or date - date: not %s %s %s", a, op, b)
	}
	if !a.numeric() || !b.numeric() {
		return KAny, fmt.Errorf("cannot compute %s %s %s", a, op, b)
	}
	switch {
	case a == KMoney && b == KMoney:
		switch op {
		case "+", "-":
			return KMoney, nil
		case "/":
			return KDecimal, nil
		}
		return KAny, fmt.Errorf("money × money has no meaning")
	case a == KMoney:
		return KMoney, nil
	case b == KMoney:
		if op == "/" {
			return KAny, fmt.Errorf("a number ÷ money has no meaning")
		}
		return KMoney, nil
	case a == KInt && b == KInt:
		if op == "/" {
			return KDecimal, nil
		}
		return KInt, nil
	}
	return KDecimal, nil
}

// ---- compilation ---------------------------------------------------------------

type opcode uint8

const (
	opConst     opcode = iota // push consts[a]
	opPath                    // push the value at paths[a], recorded as an input
	opVar                     // push slots[a]
	opSetVar                  // slots[a] = pop
	opField                   // pop v, push v.names[a], recorded as an input
	opNeg                     //
	opNot                     //
	opBin                     // binary op names[a]
	opJmp                     // pc = a
	opJmpF                    // pop; if false, pc = a
	opJmpFKeep                // if top is false, pc = a (keeping it) else pop
	opJmpTKeep                // if top is true, pc = a (keeping it) else pop
	opCall                    // names[a](argc b)
	opRef                     // pop v; push ref(names[a], names[b], v)
	opObjects                 // pop v; push objects(names[a], names[b], v)
	opIter                    // pop list; iters[a] = it, acc[a] reset for fold names[b]
	opNext                    // if iters[a] done, pc = b; else slots[c] = next
	opAcc                     // pop v; fold it into acc[a]
	opAccResult               // push acc[a]'s result
	opReturn                  //
)

type instr struct {
	op      opcode
	a, b, c int
}

// program is a compiled formula: instructions over pools, plus what the
// checker learned.
type program struct {
	code   []instr
	consts []Value
	paths  [][]pathSeg
	names  []string
	slots  int
	iters  int
	// static facts
	reads  []CollectionRead
	copy   bool   // a single load: a path copy, no computation
	refTyp string // a bare ref(): the type it resolves
	result ty
}

type compiler struct {
	p      *program
	scopes []map[string]int
}

func (c *compiler) name(s string) int {
	for i, n := range c.p.names {
		if n == s {
			return i
		}
	}
	c.p.names = append(c.p.names, s)
	return len(c.p.names) - 1
}

func (c *compiler) emit(op opcode, args ...int) int {
	in := instr{op: op}
	if len(args) > 0 {
		in.a = args[0]
	}
	if len(args) > 1 {
		in.b = args[1]
	}
	if len(args) > 2 {
		in.c = args[2]
	}
	c.p.code = append(c.p.code, in)
	return len(c.p.code) - 1
}

func (c *compiler) lookupVar(s string) (int, bool) {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if slot, ok := c.scopes[i][s]; ok {
			return slot, true
		}
	}
	return 0, false
}

func (c *compiler) bind(s string) int {
	slot := c.p.slots
	c.p.slots++
	c.scopes[len(c.scopes)-1][s] = slot
	return slot
}

func (c *compiler) gen(n *node) {
	switch n.kind {
	case nNum:
		v, _ := parseDecimal(n.text)
		k := KDecimal
		if !strings.Contains(n.text, ".") {
			k = KInt
		}
		c.p.consts = append(c.p.consts, Value{Kind: k, Num: v})
		c.emit(opConst, len(c.p.consts)-1)
	case nStr:
		c.p.consts = append(c.p.consts, Value{Kind: KString, Str: n.text})
		c.emit(opConst, len(c.p.consts)-1)
	case nBool:
		c.p.consts = append(c.p.consts, Value{Kind: KBool, Bool: n.text == "true"})
		c.emit(opConst, len(c.p.consts)-1)
	case nPath:
		c.p.paths = append(c.p.paths, n.path)
		c.emit(opPath, len(c.p.paths)-1)
	case nVar:
		slot, _ := c.lookupVar(n.text)
		c.emit(opVar, slot)
	case nField:
		c.gen(n.args[0])
		c.emit(opField, c.name(n.field))
	case nUnary:
		c.gen(n.args[0])
		if n.text == "not" {
			c.emit(opNot)
		} else {
			c.emit(opNeg)
		}
	case nBinary:
		switch n.text {
		case "and":
			c.gen(n.args[0])
			j := c.emit(opJmpFKeep, 0)
			c.gen(n.args[1])
			c.p.code[j].a = len(c.p.code)
		case "or":
			c.gen(n.args[0])
			j := c.emit(opJmpTKeep, 0)
			c.gen(n.args[1])
			c.p.code[j].a = len(c.p.code)
		default:
			c.gen(n.args[0])
			c.gen(n.args[1])
			c.emit(opBin, c.name(n.text))
		}
	case nRef, nObjs:
		c.gen(n.args[0])
		op := opRef
		if n.kind == nObjs {
			op = opObjects
		}
		c.emit(op, c.name(n.text), c.name(n.field))
	case nCall:
		if n.text == "if" {
			c.gen(n.args[0])
			jElse := c.emit(opJmpF, 0)
			c.gen(n.args[1])
			jEnd := c.emit(opJmp, 0)
			c.p.code[jElse].a = len(c.p.code)
			c.gen(n.args[2])
			c.p.code[jEnd].a = len(c.p.code)
			return
		}
		for _, a := range n.args {
			c.gen(a)
		}
		c.emit(opCall, c.name(n.text), len(n.args))
	case nFold:
		c.scopes = append(c.scopes, map[string]int{})
		var accSlot int
		if n.text == "fold" {
			c.gen(n.init)
			accSlot = c.bind(n.acc)
			c.emit(opSetVar, accSlot)
		}
		c.gen(n.coll)
		it := c.p.iters
		c.p.iters++
		c.emit(opIter, it, c.name(n.text))
		slot := c.bind(n.binder)
		loop := c.emit(opNext, it, 0, slot)
		if n.where != nil {
			c.gen(n.where)
			c.emit(opJmpF, loop)
		}
		switch n.text {
		case "count":
			c.p.consts = append(c.p.consts, Value{Kind: KInt, Num: ratInt(1)})
			c.emit(opConst, len(c.p.consts)-1)
			c.emit(opAcc, it)
		case "fold":
			c.gen(n.body)
			c.emit(opSetVar, accSlot)
		default:
			c.gen(n.body)
			c.emit(opAcc, it)
		}
		c.emit(opJmp, loop)
		c.p.code[loop].b = len(c.p.code)
		if n.text == "fold" {
			c.emit(opVar, accSlot)
		} else {
			c.emit(opAccResult, it)
		}
		c.scopes = c.scopes[:len(c.scopes)-1]
	}
}

// Formula is a parsed, checked and compiled "=" template, immutable and
// cached by text: formulas are data, small and comparable.
type Formula struct {
	Text string
	tree *node
	prog *program
}

var formulaCache sync.Map // text → *Formula

// ParseFormula parses the text after "=" and compiles it against no
// catalog: syntax and the data-free laws only. Typed checks run in Check.
func ParseFormula(text string) (*Formula, error) {
	if f, ok := formulaCache.Load(text); ok {
		return f.(*Formula), nil
	}
	tree, err := parseFormula(text)
	if err != nil {
		return nil, err
	}
	env := &checkEnv{vars: map[string]ty{}}
	result, err := env.check(tree, false)
	if err != nil {
		return nil, err
	}
	c := &compiler{p: &program{}, scopes: []map[string]int{{}}}
	c.gen(tree)
	c.emit(opReturn)
	c.p.reads, c.p.result = env.reads, result
	c.p.copy = tree.kind == nPath
	if tree.kind == nRef {
		c.p.refTyp = tree.text
	}
	f := &Formula{Text: text, tree: tree, prog: c.p}
	formulaCache.Store(text, f)
	return f, nil
}

// Check re-runs the static checker with a catalog and typed prefixes: reads
// through links name real fields, kinds combine lawfully, and the result
// fits the target field. Errors are the draft gate's.
func (f *Formula) Check(types Catalog, typed map[string]string, fd FieldDef) error {
	env := &checkEnv{types: types, typed: typed, vars: map[string]ty{}}
	result, err := env.check(f.tree, false)
	if err != nil {
		return err
	}
	return fitsField(result, fd)
}

// fitsField judges a static result against the field it feeds.
func fitsField(r ty, fd FieldDef) error {
	if r.kind == KAny || fd.Type == "" {
		return nil
	}
	ok := true
	switch fd.Type {
	case "money", "decimal":
		ok = r.kind.numeric() || r.kind == KString
	case "int":
		ok = r.kind == KInt || r.kind == KDecimal || r.kind == KString
	case "date":
		ok = r.kind == KDate || r.kind == KString
	case "string", "enum":
		ok = r.kind != KList && r.kind != KObject && r.kind != KBool
	default:
		if target, isRef := RefTarget(fd.Type); isRef {
			ok = (r.kind == KRef && (r.typ == "" || r.typ == target)) || r.kind == KString
		}
	}
	if !ok {
		return fmt.Errorf("a %s cannot be stored in field %q (%s)", r.kind, fd.Name, fd.Type)
	}
	return nil
}

// Reads lists the objects() collections the formula walks — the keyed state
// reads a rule declares by writing them.
func (f *Formula) Reads() []CollectionRead { return f.prog.reads }

// IsCopy reports a bare path: a copy, not a computation.
func (f *Formula) IsCopy() bool { return f.prog.copy }
