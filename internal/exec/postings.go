// The double-entry sub-language (SPEC M1), ported from the Sunbeetle
// lessons: balance is checked after expansion and before booking — an
// unbalanced expansion is a rule error and nothing books; the period lock is
// a pre-insert check in the poster; every line carries rule provenance
// through the ordinary derived-event path, so replay and simulation treat
// postings like any other materialization.
package exec

import (
	"fmt"

	"rhea/internal/core"
)

// PostingObjectType is the conventional object type postings materialize as.
// It is seeded data like any other type; the kernel only fixes the name and
// the field meaning (entry, line, book, account, side, amount, currency,
// date, and with conversion tx_amount, tx_currency).
const PostingObjectType = "posting"

// DefaultBook is the book an entry belongs to when its template names none.
const DefaultBook = "main"

// FXRateObjectType is the conventional object type currency rates live in:
// master data materialized from rate events by a plain rule, resolved by the
// composite code "<from>/<to>/<date>" (resolution keys are one global
// namespace — DECISIONS). The rate field is a decimal string; the statutory
// D-1 subtlety is the rate publisher's job, which publishes under
// application dates — calendars stay out of the kernel.
const FXRateObjectType = "fx_rate"

// ExpandPostings evaluates a postings template against one event: one book
// and one currency per entry, each account resolved by code to an existing
// account object, debits equal to credits, and the event's month not locked
// for the entry's book by a period_lock object. With a convert clause
// (DECISIONS 2026-10-04, E3) the lines evaluate and balance in transaction
// currency, then book in the functional currency at the fx_rate for
// (from, to, date) — amounts rounded per the declared method, a broken
// functional balance plugged to the declared rounding account or refused.
// Pure given lookup and get — the executor passes store state, the
// simulator its in-memory world — and conversion runs at firing time, so
// the functional amounts bake into derived events and replay never
// re-converts. Ids and the entry key build on idBase (the root event id, or
// the causing object's id when cascaded) and are rule-qualified, so several
// ledger rules may book the same event (a VAT entry beside a revenue entry)
// and one ledger rule may book each of an event's lines (a valuation entry
// per movement) without colliding.
func ExpandPostings(p *core.PostingsTemplate, postingType core.ObjectType, ev core.Event, idBase, ruleID string, payload any, lookup core.Lookup, get core.Getter) ([]core.MaterializedObject, error) {
	if lookup == nil {
		return nil, fmt.Errorf("postings need object state, none available here")
	}

	if len(ev.OccurredAt) < 7 {
		return nil, fmt.Errorf("event has no business date")
	}
	book := DefaultBook
	if p.Book != "" {
		var err error
		if book, err = evalString(p.Book, payload); err != nil {
			return nil, fmt.Errorf("book: %w", err)
		}
	}

	// Period lock: no posting into a month locked for this entry's book
	// (business date decides). A lock names its (book, month); a lock object
	// from before books existed matches no book and locks nothing from here
	// on (DECISIONS 2026-10-04, E1).
	month := ev.OccurredAt[:7]
	monthLocks, err := lookup("period_lock", "month", month)
	if err != nil {
		return nil, err
	}
	if len(monthLocks) > 0 {
		bookLocks, err := lookup("period_lock", "book", book)
		if err != nil {
			return nil, err
		}
		if intersects(monthLocks, bookLocks) {
			return nil, fmt.Errorf("period %s is locked for book %s", month, book)
		}
	}

	currency, err := evalString(p.Currency, payload)
	if err != nil {
		return nil, fmt.Errorf("currency: %w", err)
	}

	// The lines evaluate and balance in transaction currency first: a
	// template that lies about its own arithmetic is a rule error whether or
	// not conversion follows.
	type txLine struct {
		accountID string
		side      string
		amount    int64
	}
	var debits, credits int64
	lines := make([]txLine, 0, len(p.Lines))
	for i, l := range p.Lines {
		code, err := evalString(l.Account, payload)
		if err != nil {
			return nil, fmt.Errorf("line %d account: %w", i+1, err)
		}
		accountID, err := core.ResolveRef(lookup, "account", "code", code)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		side, tmpl := "debit", l.Debit
		if l.Credit != "" {
			side, tmpl = "credit", l.Credit
		}
		amount, err := evalMoney(tmpl, payload)
		if err != nil {
			return nil, fmt.Errorf("line %d amount: %w", i+1, err)
		}
		if amount <= 0 {
			return nil, fmt.Errorf("line %d: amount must be positive, got %d", i+1, amount)
		}
		if side == "debit" {
			debits += amount
		} else {
			credits += amount
		}
		lines = append(lines, txLine{accountID, side, amount})
	}
	if debits != credits {
		return nil, fmt.Errorf("entry does not balance: debits %s, credits %s %s",
			core.FormatMoney(debits), core.FormatMoney(credits), currency)
	}

	// Conversion (E3): the functional currency and, unless the conversion is
	// the identity, the fx_rate for (from, to, date) — read at firing time,
	// so the functional amounts bake into the derived events.
	bookCurrency, converted := currency, false
	var mant, div int64
	if p.Convert != nil {
		to, err := evalString(p.Convert.To, payload)
		if err != nil {
			return nil, fmt.Errorf("convert.to: %w", err)
		}
		bookCurrency = to
		if to != currency {
			if get == nil {
				return nil, fmt.Errorf("conversion needs object state, none available here")
			}
			date, err := evalString(p.Convert.Date, payload)
			if err != nil {
				return nil, fmt.Errorf("convert.date: %w", err)
			}
			code := currency + "/" + to + "/" + date
			rateID, err := core.ResolveRef(lookup, FXRateObjectType, "code", code)
			if err != nil {
				return nil, err
			}
			state, ok, err := get(rateID)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("fx_rate %s resolved but cannot be read", code)
			}
			rate, _ := state["rate"].(string)
			if mant, div, err = parseRate(rate); err != nil {
				return nil, fmt.Errorf("fx_rate %s: %w", code, err)
			}
			converted = true
		}
	}
	functional := func(tx int64) int64 {
		if !converted {
			return tx
		}
		return divHalfUp(tx*mant, div)
	}

	entry := fmt.Sprintf("entry-%s-%s", idBase, ruleID)
	emit := func(n int, accountID, side string, amount int64, tx *txLine) core.MaterializedObject {
		state := map[string]any{
			"entry":    entry,
			"line":     n,
			"book":     book,
			"account":  accountID,
			"side":     side,
			"amount":   amount,
			"currency": bookCurrency,
			"date":     ev.OccurredAt,
		}
		if p.Convert != nil && tx != nil {
			state["tx_amount"] = tx.amount
			state["tx_currency"] = currency
		}
		return core.MaterializedObject{
			ObjectID:    fmt.Sprintf("posting-%s-%s-%d", idBase, ruleID, n),
			ObjectType:  PostingObjectType,
			TypeVersion: postingType.Version,
			State:       state,
		}
	}
	out := make([]core.MaterializedObject, 0, len(lines)+1)
	var fdebits, fcredits int64
	for i := range lines {
		l := lines[i]
		f := functional(l.amount)
		if l.side == "debit" {
			fdebits += f
		} else {
			fcredits += f
		}
		out = append(out, emit(i+1, l.accountID, l.side, f, &l))
	}

	// Per-line rounding can break the functional balance by minor units;
	// the residue books to the declared rounding account as an explicit
	// plug line, or the entry refuses. Ledgers stay balanced by invariant,
	// and the rounding is visible, never hidden.
	if diff := fdebits - fcredits; diff != 0 {
		if p.Convert == nil || p.Convert.RoundingAccount == "" {
			return nil, fmt.Errorf("rounding broke the balance by %s %s — the rule declares no rounding_account",
				core.FormatMoney(absInt64(diff)), bookCurrency)
		}
		code, err := evalString(p.Convert.RoundingAccount, payload)
		if err != nil {
			return nil, fmt.Errorf("convert.rounding_account: %w", err)
		}
		accountID, err := core.ResolveRef(lookup, "account", "code", code)
		if err != nil {
			return nil, fmt.Errorf("rounding account: %w", err)
		}
		side, amount := "credit", diff
		if diff < 0 {
			side, amount = "debit", -diff
		}
		out = append(out, emit(len(lines)+1, accountID, side, amount, nil))
	}
	return out, nil
}

// parseRate parses a positive decimal string ("4.3215") into mantissa and
// divisor (43215, 10000) for integer conversion arithmetic. No floats, ever.
func parseRate(s string) (mant, div int64, err error) {
	div = 1
	dot := false
	for _, r := range s {
		switch {
		case r == '.' && !dot:
			dot = true
		case r >= '0' && r <= '9':
			mant = mant*10 + int64(r-'0')
			if dot {
				div *= 10
			}
		default:
			return 0, 0, fmt.Errorf("rate %q is not a plain decimal", s)
		}
	}
	if s == "" || mant == 0 {
		return 0, 0, fmt.Errorf("rate %q is not positive", s)
	}
	return mant, div, nil
}

// divHalfUp divides non-negative integers rounding half away from zero —
// the one admitted rounding method (the stance is declared in the rule).
func divHalfUp(n, d int64) int64 {
	q, r := n/d, n%d
	if 2*r >= d {
		q++
	}
	return q
}

func absInt64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// intersects reports whether two id lists share an element — the lock check
// composes two single-field lookups, so the Lookup contract stays one field.
func intersects(a, b []string) bool {
	set := make(map[string]struct{}, len(a))
	for _, id := range a {
		set[id] = struct{}{}
	}
	for _, id := range b {
		if _, ok := set[id]; ok {
			return true
		}
	}
	return false
}

func evalString(tmpl string, payload any) (string, error) {
	t, err := core.ParseTemplate(tmpl)
	if err != nil {
		return "", err
	}
	v, err := t.Eval(payload, core.FieldDef{Type: "string"}, nil)
	if err != nil {
		return "", err
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return "", fmt.Errorf("want a non-empty string, got %v", v)
	}
	return s, nil
}

// evalMoney accepts a path or sum() template over decimal strings, or a
// literal decimal string, and yields minor units.
func evalMoney(tmpl string, payload any) (int64, error) {
	t, err := core.ParseTemplate(tmpl)
	if err != nil {
		return 0, err
	}
	v, err := t.Eval(payload, core.FieldDef{Type: "money"}, nil)
	if err != nil {
		return 0, err
	}
	switch n := v.(type) {
	case int64:
		return n, nil
	case string: // a literal template reaches us unparsed
		return core.ParseMoney(n)
	}
	return 0, fmt.Errorf("want a money amount, got %T", v)
}
