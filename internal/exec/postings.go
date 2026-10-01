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
// the field meaning (entry, line, account, side, amount, currency, date).
const PostingObjectType = "posting"

// ExpandPostings evaluates a postings template against one event: one
// currency per entry, each account resolved by code to an existing account
// object, debits equal to credits, and the event's month not locked by a
// period_lock object. Pure given the lookup — the executor passes store
// state, the simulator its in-memory world.
func ExpandPostings(p *core.PostingsTemplate, postingType core.ObjectType, ev core.Event, payload any, lookup core.Lookup) ([]core.MaterializedObject, error) {
	if lookup == nil {
		return nil, fmt.Errorf("postings need object state, none available here")
	}

	// Period lock: no posting into a locked month (business date decides).
	if len(ev.OccurredAt) < 7 {
		return nil, fmt.Errorf("event has no business date")
	}
	month := ev.OccurredAt[:7]
	locks, err := lookup("period_lock", "month", month)
	if err != nil {
		return nil, err
	}
	if len(locks) > 0 {
		return nil, fmt.Errorf("period %s is locked", month)
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
			ObjectID:    fmt.Sprintf("posting-%d-%d", ev.ID, i+1),
			ObjectType:  PostingObjectType,
			TypeVersion: postingType.Version,
			State: map[string]any{
				"entry":    fmt.Sprintf("entry-%d", ev.ID),
				"line":     i + 1,
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
