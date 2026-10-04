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
// date).
const PostingObjectType = "posting"

// DefaultBook is the book an entry belongs to when its template names none.
const DefaultBook = "main"

// ExpandPostings evaluates a postings template against one event: one book
// and one currency per entry, each account resolved by code to an existing
// account object, debits equal to credits, and the event's month not locked
// for the entry's book by a period_lock object. Pure given the lookup — the
// executor passes store state, the simulator its in-memory world. Ids and
// the entry key build on idBase (the root event id, or the causing object's
// id when cascaded) and are rule-qualified, so several ledger rules may book
// the same event (a VAT entry beside a revenue entry) and one ledger rule
// may book each of an event's lines (a valuation entry per movement) without
// colliding.
func ExpandPostings(p *core.PostingsTemplate, postingType core.ObjectType, ev core.Event, idBase, ruleID string, payload any, lookup core.Lookup) ([]core.MaterializedObject, error) {
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

	var debits, credits int64
	out := make([]core.MaterializedObject, 0, len(p.Lines))
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
		out = append(out, core.MaterializedObject{
			ObjectID:    fmt.Sprintf("posting-%s-%s-%d", idBase, ruleID, i+1),
			ObjectType:  PostingObjectType,
			TypeVersion: postingType.Version,
			State: map[string]any{
				"entry":    fmt.Sprintf("entry-%s-%s", idBase, ruleID),
				"line":     i + 1,
				"book":     book,
				"account":  accountID,
				"side":     side,
				"amount":   amount,
				"currency": currency,
				"date":     ev.OccurredAt,
			},
		})
	}
	if debits != credits {
		return nil, fmt.Errorf("entry does not balance: debits %s, credits %s %s",
			core.FormatMoney(debits), core.FormatMoney(credits), currency)
	}
	return out, nil
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
