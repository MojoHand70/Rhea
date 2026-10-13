package core

import (
	"fmt"
	"math/big"
	"slices"
	"sort"
	"strings"

	"rhea/internal/match"
)

// The matching effect (DIRECTION 2026-10-11, "Matching: a deterministic
// engine with N sides, as a kernel method"; DECISIONS 2026-10-11, the
// attempt): a kernel method parameterized by rule data. The template is the
// structure — N sides, each a keyed read of one object type with the fields
// that carry its measures, its open portion, its date and its reference —
// and the policy, an ordered list of named strategies with their parameters.
// The engine (internal/match) is kernel code; the policy is data the agent
// may compose and the gate approves; no strategy is ever a loop in data.
//
// A matching rule fires on the materialization of the trigger side's object
// and reads the other sides' candidates by key at that moment. Its outcome
// is one event, baked: objects.matched naming every item on every side with
// the amount allocated, the open portion before and after, every difference
// absorbed with its account, and the strategy that found it — or
// objects.unmatched naming every step tried and why it found nothing.
// Ordinary rules explain the match: the items' open amounts move by
// amendment, the settlement object materializes, the differences post.
// Replay never re-matches.

const (
	// EventObjectsMatched is the derived event a matching rule emits when a
	// strategy finds a match; its payload is a MatchedObjects.
	EventObjectsMatched = "objects.matched"
	// EventObjectsUnmatched is the derived event a matching rule emits when
	// no step of its policy finds a match; its payload is an
	// UnmatchedObjects. The remainder is the person's grid; this event is
	// its explanation.
	EventObjectsUnmatched = "objects.unmatched"

	// DefaultMatchingCap is how many candidates a side may hold before the
	// items go to the person instead of the engine.
	DefaultMatchingCap = 200
)

// MatchingTemplate is the structure and the policy, as rule data.
type MatchingTemplate struct {
	// Name is what the match event carries as "matching", so the rules
	// that explain a match (amendments, settlements, differences) key on
	// it — and two rules with different triggers over one structure share
	// their consequences.
	Name string `json:"name"`
	// Trigger names the side whose object's materialization fires the
	// rule; the rule's match must name that side's type.
	Trigger string         `json:"trigger"`
	Sides   []MatchingSide `json:"sides"`
	Policy  []MatchingStep `json:"policy"`
	// Convert lets the sides carry different currencies: the FX precedent
	// of postings' convert, as a clause of the structure.
	Convert *MatchingConvert `json:"convert,omitempty"`
}

// MatchingConvert compares money across currencies. Every candidate's open
// amount is read in the trigger's currency at the fx_rate of the settlement
// date — the date of the item on DateSide: the trigger's own when it is
// that side (a payment against invoices), each candidate's own otherwise
// (an invoice against the payments waiting for it) — and every allocation
// is written back in the item's own currency, rounded as declared. The
// difference between what an item was worth in the functional currency at
// its own date and at its settlement date (różnice kursowe) is recorded
// against RateUpAccount when the rate rose and RateDownAccount when it
// fell; whether that is a gain or a loss depends on the side — the rules
// that post it know.
type MatchingConvert struct {
	Functional      string `json:"functional"`
	DateSide        string `json:"date_side"`
	Rounding        string `json:"rounding"`
	RateUpAccount   string `json:"rate_up_account"`
	RateDownAccount string `json:"rate_down_account"`
}

// MatchingSide declares one side: the object type, how its candidates are
// read when another side triggers (a keyed objects(...) read written against
// the trigger's payload), and which fields carry its measures.
type MatchingSide struct {
	Name string `json:"name"`
	Type string `json:"type"`
	// Items is the keyed read of this side's candidates, absent on the
	// trigger side (its one item is the object that fired the rule).
	Items string `json:"items,omitempty"`
	// Join narrows the candidates to those agreeing with the trigger on
	// these fields — candidate field → trigger field (the order's lines of
	// the same item, the group's documents of the same seller): the keys
	// that join candidates, beyond the one the read is keyed by.
	Join map[string]string `json:"join,omitempty"`
	// Open maps each allocated measure to the field holding the item's open
	// portion — money or a quantity, declared amendable, because the match's
	// consequence is an amendment of it. Every side declares the same
	// measures.
	Open map[string]string `json:"open"`
	// Compare maps each compared measure (a price) to its field: it must
	// agree within tolerance and is never consumed. A side may lack one.
	Compare map[string]string `json:"compare,omitempty"`
	// Currency names the field carrying the currency of the money
	// measures; required when a measure is money.
	Currency string `json:"currency,omitempty"`
	// Date names the field strategies order by and the time window reads.
	Date string `json:"date"`
	// Reference names the field the reference strategy matches: the
	// trigger's reference text must name it.
	Reference string `json:"reference,omitempty"`
	// Cap bounds the candidates; above it the items go to the person. Zero
	// means DefaultMatchingCap.
	Cap int `json:"cap,omitempty"`
}

// MatchingStep is one strategy with its parameters.
type MatchingStep struct {
	Strategy string `json:"strategy"`
	// Tolerance is an absolute bound per measure, as a decimal string in
	// the measure's unit; TolerancePercent a relative one, of the trigger's
	// value. A difference inside the bound is absorbed and recorded
	// against DifferenceAccount, an account code.
	Tolerance         map[string]string `json:"tolerance,omitempty"`
	TolerancePercent  map[string]string `json:"tolerance_percent,omitempty"`
	DifferenceAccount string            `json:"difference_account,omitempty"`
	// WithinDays admits candidates dated within so many days of the
	// trigger; zero means any.
	WithinDays int `json:"within_days,omitempty"`
	// MaxItems caps a subset's size (subset only); zero means the engine's
	// default.
	MaxItems int `json:"max_items,omitempty"`
	// Partial lets the reference strategy settle what it can when the
	// named items and the trigger disagree (reference only).
	Partial bool `json:"partial,omitempty"`
}

// MatchedObjects is the payload of an objects.matched event. Amounts are in
// the boundary encoding of their measure's field: money as int64 minor
// units, int as integers, decimal as strings.
type MatchedObjects struct {
	Matching string       `json:"matching"`
	Strategy string       `json:"strategy"`
	Step     int          `json:"step"`
	Trigger  MatchTrigger `json:"trigger"`
	// Kinds says how each measure's amounts are encoded (money, int,
	// decimal), so a reader of the event renders them without the catalog.
	Kinds       map[string]string            `json:"kinds"`
	Sides       map[string][]MatchAllocation `json:"sides"`
	Differences []MatchDifference            `json:"differences"`
	Tried       []MatchAttempt               `json:"tried"`
	// Inputs is what the engine read to decide: the candidates per side,
	// the tolerances applied — the calc of a match, for the walk.
	Inputs map[string]any `json:"inputs,omitempty"`
	// Standards, on a person's pairing, names the strategies that would have
	// found the same pairing over the sides as the policy reads them — what
	// the host learns from: after a person has paired the same way twice,
	// the host proposes that standard as a step, behind the gate.
	Standards []string `json:"standards,omitempty"`
	// ReversalOf names the match this one undoes (the unpairing verb): the
	// allocations negated, the open amounts restored, every difference's
	// direction flipped — a correcting event the same rules explain, never
	// a deletion. Reason is the person's.
	ReversalOf int64  `json:"reversal_of,omitempty"`
	Reason     string `json:"reason,omitempty"`
	// Seq counts the matches this structure made on this trigger before
	// this one — the policy's, pairings, reversals — so each match of a
	// trigger's life is its own chain point and what it causes never
	// collides with an earlier match's.
	Seq int `json:"seq,omitempty"`
}

// StrategyReversal is the strategy a reversal carries: no search found it,
// a person undid a match.
const StrategyReversal = "reversal"

// LearnedBundleID names the proposal the host makes when a person has paired
// one structure the same way twice: one bundle per structure and standard,
// so a rejected proposal never returns (the overruled rule never returns).
func LearnedBundleID(matching, standard string) string { return "learned-" + matching + "-" + standard }

// UnmatchedObjects is the payload of an objects.unmatched event: the trigger
// and every step tried, with why it found nothing — or, when no step could
// run (a side above its cap), the one reason.
type UnmatchedObjects struct {
	Matching string         `json:"matching"`
	Trigger  MatchTrigger   `json:"trigger"`
	Tried    []MatchAttempt `json:"tried,omitempty"`
	Reason   string         `json:"reason,omitempty"`
	Inputs   map[string]any `json:"inputs,omitempty"`
}

// MatchTrigger names the item whose arrival fired the rule.
type MatchTrigger struct {
	Side       string `json:"side"`
	ObjectID   string `json:"object_id"`
	ObjectType string `json:"object_type"`
}

// MatchAllocation is what a match did to one item.
type MatchAllocation struct {
	ObjectID   string         `json:"object_id"`
	ObjectType string         `json:"object_type"`
	Allocated  map[string]any `json:"allocated"`
	OpenBefore map[string]any `json:"open_before"`
	OpenAfter  map[string]any `json:"open_after"`
	Currency   string         `json:"currency,omitempty"`
	Date       string         `json:"date,omitempty"`
	Reference  string         `json:"reference,omitempty"`
	// Equivalent is the allocation read in the trigger's currency, when the
	// structure converts: what the match compared.
	Equivalent         map[string]any `json:"equivalent,omitempty"`
	EquivalentCurrency string         `json:"equivalent_currency,omitempty"`
}

// MatchDifference is an amount a match absorbed, positive, with its
// direction: "short" when the trigger brought less than the side settles,
// "over" when more. Item names the one item a compared measure's gap sits on.
type MatchDifference struct {
	Kind      string `json:"kind"`
	Measure   string `json:"measure"`
	Side      string `json:"side"`
	Item      string `json:"item,omitempty"`
	Amount    any    `json:"amount"`
	Direction string `json:"direction"`
	Currency  string `json:"currency,omitempty"`
	Account   string `json:"account,omitempty"`
}

// MatchAttempt is one policy step's outcome.
type MatchAttempt struct {
	Step     int    `json:"step"`
	Strategy string `json:"strategy"`
	Found    bool   `json:"found,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

// MatchingSpec: spec of a `matching` notion (the grid, DIRECTION 2026-10-11):
// the N-sided view of what no strategy settled — every side's open items,
// side by side — with the pairing verb. Rule names the matching rule whose
// structure says what the sides are; the view adds no opinion of its own.
type MatchingSpec struct {
	Rule string `json:"rule"`
}

// EventMatchingPaired is the raw event the pairing verb appends: a person's
// selection of one trigger item and the items it settles on the other
// sides. The kernel reacts by running the same engine on the selection and
// emitting the same objects.matched event the policy would — one fact,
// explained the same way, whoever made it.
const EventMatchingPaired = "matching.paired"

// EventMatchingUnpaired is the raw event the unpairing verb appends: a
// person undoing one match, with a reason. The kernel reacts with the
// reversal — an objects.matched that restores what the match allocated.
const EventMatchingUnpaired = "matching.unpaired"

// Side finds a declared side by name.
func (m *MatchingTemplate) Side(name string) (MatchingSide, bool) {
	for _, s := range m.Sides {
		if s.Name == name {
			return s, true
		}
	}
	return MatchingSide{}, false
}

// AllocatedMeasures lists the allocated measures, by name, sorted.
func (m *MatchingTemplate) AllocatedMeasures() []string {
	t, ok := m.Side(m.Trigger)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(t.Open))
	for name := range t.Open {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// ComparedMeasures lists the compared measures declared on any side.
func (m *MatchingTemplate) ComparedMeasures() []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range m.Sides {
		for name := range s.Compare {
			if !seen[name] {
				seen[name] = true
				out = append(out, name)
			}
		}
	}
	sort.Strings(out)
	return out
}

// validate checks the template without a catalog: the structure's shape,
// the policy's parameters, and that the rule fires on the trigger side's
// materialization.
func (m *MatchingTemplate) validate(rule Match) error {
	if m.Name == "" {
		return fmt.Errorf("matching.name is required — the match event carries it")
	}
	if len(m.Sides) < 2 {
		return fmt.Errorf("matching needs at least two sides")
	}
	seen := map[string]bool{}
	for i, s := range m.Sides {
		if s.Name == "" || s.Type == "" {
			return fmt.Errorf("matching side %d needs a name and a type", i+1)
		}
		if seen[s.Name] {
			return fmt.Errorf("matching side %q is declared twice", s.Name)
		}
		seen[s.Name] = true
		if len(s.Open) == 0 {
			return fmt.Errorf("matching side %q declares no open measure", s.Name)
		}
		if s.Date == "" {
			return fmt.Errorf("matching side %q needs a date field", s.Name)
		}
		if s.Cap < 0 {
			return fmt.Errorf("matching side %q: cap must not be negative", s.Name)
		}
	}
	trigger, ok := m.Side(m.Trigger)
	if !ok {
		return fmt.Errorf("matching.trigger %q is not a declared side", m.Trigger)
	}
	if trigger.Items != "" {
		return fmt.Errorf("the trigger side %q reads no items — its one item is the object that fired the rule", trigger.Name)
	}
	if len(trigger.Join) > 0 {
		return fmt.Errorf("the trigger side %q joins nothing — joins narrow the other sides to it", trigger.Name)
	}
	allocated := m.AllocatedMeasures()
	for _, s := range m.Sides {
		if s.Name == m.Trigger {
			continue
		}
		if s.Items == "" {
			return fmt.Errorf("matching side %q needs an items read — the keyed objects(...) of its candidates", s.Name)
		}
		for cf, tf := range s.Join {
			if cf == "" || tf == "" {
				return fmt.Errorf("matching side %q: a join names a candidate field and a trigger field", s.Name)
			}
		}
		t, err := ParseTemplate(s.Items)
		if err != nil {
			return fmt.Errorf("matching side %q items: %w", s.Name, err)
		}
		if t.kind != "formula" {
			return fmt.Errorf("matching side %q items wants an objects(...) read, got %q", s.Name, s.Items)
		}
		var theirs []string
		for name := range s.Open {
			theirs = append(theirs, name)
		}
		sort.Strings(theirs)
		if !slices.Equal(theirs, allocated) {
			return fmt.Errorf("matching side %q opens %v, the trigger side opens %v — every side declares the same measures", s.Name, theirs, allocated)
		}
	}
	// The rule fires on the trigger's materialization: object.materialized
	// where object_type is the trigger side's type.
	if rule.EventType != EventObjectMaterialized {
		return fmt.Errorf("a matching rule matches %s of its trigger side (%s), not %q", EventObjectMaterialized, trigger.Type, rule.EventType)
	}
	named := false
	for _, c := range rule.Where {
		if c.Path == "$.object_type" && c.Op == "eq" {
			if v, _ := c.Value.(string); v == trigger.Type {
				named = true
			} else {
				return fmt.Errorf("a matching rule triggered by %q fires on a %s, but its match names %v", trigger.Name, trigger.Type, c.Value)
			}
		}
	}
	if !named {
		return fmt.Errorf("a matching rule's match must say where $.object_type eq %q — the trigger side's type", trigger.Type)
	}
	if len(m.Policy) == 0 {
		return fmt.Errorf("matching.policy is empty — at least one strategy")
	}
	if c := m.Convert; c != nil {
		if c.Functional == "" {
			return fmt.Errorf("matching.convert needs the functional currency")
		}
		if _, ok := m.Side(c.DateSide); !ok {
			return fmt.Errorf("matching.convert.date_side %q is not a declared side", c.DateSide)
		}
		if !IsRoundingMethod(c.Rounding) {
			return fmt.Errorf("matching.convert.rounding must be declared; %q is not an admitted method", c.Rounding)
		}
		if c.RateUpAccount == "" || c.RateDownAccount == "" {
			return fmt.Errorf("matching.convert needs rate_up_account and rate_down_account — a rate difference has a home")
		}
		if len(allocated) != 1 {
			return fmt.Errorf("matching.convert works on one allocated money measure, the structure has %d", len(allocated))
		}
	}
	compared := m.ComparedMeasures()
	for i, st := range m.Policy {
		at := fmt.Sprintf("matching.policy step %d", i+1)
		if !match.Known(st.Strategy) {
			return fmt.Errorf("%s: unknown strategy %q; the strategies are %s", at, st.Strategy, strings.Join(match.Strategies, ", "))
		}
		for name, v := range st.Tolerance {
			if !slices.Contains(allocated, name) && !slices.Contains(compared, name) {
				return fmt.Errorf("%s: tolerance names %q, not a measure", at, name)
			}
			r, err := ParseDecimal(v)
			if err != nil || r.Sign() < 0 {
				return fmt.Errorf("%s: tolerance %q for %s is not a non-negative decimal", at, v, name)
			}
		}
		for name, v := range st.TolerancePercent {
			if !slices.Contains(allocated, name) && !slices.Contains(compared, name) {
				return fmt.Errorf("%s: tolerance_percent names %q, not a measure", at, name)
			}
			r, err := ParseDecimal(v)
			if err != nil || r.Sign() < 0 {
				return fmt.Errorf("%s: tolerance_percent %q for %s is not a non-negative decimal", at, v, name)
			}
		}
		if (len(st.Tolerance) > 0 || len(st.TolerancePercent) > 0) && st.DifferenceAccount == "" {
			return fmt.Errorf("%s: a tolerance absorbs a difference, so difference_account must name where it goes", at)
		}
		if st.WithinDays < 0 {
			return fmt.Errorf("%s: within_days must not be negative", at)
		}
		if st.MaxItems != 0 && st.Strategy != match.Subset {
			return fmt.Errorf("%s: max_items belongs to the subset strategy", at)
		}
		if st.MaxItems < 0 || st.MaxItems > match.MaxMaxItems {
			return fmt.Errorf("%s: max_items must be between 1 and %d", at, match.MaxMaxItems)
		}
		if st.Partial && st.Strategy != match.Reference {
			return fmt.Errorf("%s: partial belongs to the reference strategy", at)
		}
		if (st.Strategy == match.OldestFirst || (st.Strategy == match.Reference && st.Partial)) && len(allocated) != 1 {
			return fmt.Errorf("%s: %s allocates one measure, the structure has %d", at, st.Strategy, len(allocated))
		}
		if st.Strategy == match.Reference {
			for _, s := range m.Sides {
				if s.Reference == "" {
					return fmt.Errorf("%s: the reference strategy needs a reference field on every side; %q declares none", at, s.Name)
				}
			}
		}
	}
	return nil
}

// Check is the typed gate: every side's type is declared, the measure
// fields exist with the same numeric type across sides and are amendable
// (the match's consequence is an amendment), currency, date and reference
// fields exist with their types, and each items read is a collection of
// the side's type, typed against the trigger's state.
func (m *MatchingTemplate) Check(types Catalog) error {
	trigger, _ := m.Side(m.Trigger)
	kinds := map[string]string{} // measure → field type
	for _, s := range m.Sides {
		ot, ok := types(s.Type)
		if !ok {
			return fmt.Errorf("matching side %q: %q is not a declared object type", s.Name, s.Type)
		}
		money := false
		field := func(role, name string) (FieldDef, error) {
			fd, ok := ot.Field(name)
			if !ok {
				return fd, fmt.Errorf("matching side %q: %s field %q is not a field of %s", s.Name, role, name, s.Type)
			}
			return fd, nil
		}
		for measure, name := range s.Open {
			fd, err := field("open", name)
			if err != nil {
				return err
			}
			if fd.Type != "money" && fd.Type != "int" && fd.Type != "decimal" {
				return fmt.Errorf("matching side %q: open field %q is %s, not an amount", s.Name, name, fd.Type)
			}
			if prev, seen := kinds[measure]; seen && prev != fd.Type {
				return fmt.Errorf("matching: measure %q is %s on one side and %s on %q", measure, prev, fd.Type, s.Name)
			}
			kinds[measure] = fd.Type
			if !ot.CanAmend(name) {
				return fmt.Errorf("matching side %q: open field %q of %s must be amendable — a match moves it", s.Name, name, s.Type)
			}
			money = money || fd.Type == "money"
		}
		for measure, name := range s.Compare {
			fd, err := field("compare", name)
			if err != nil {
				return err
			}
			if fd.Type != "money" && fd.Type != "int" && fd.Type != "decimal" {
				return fmt.Errorf("matching side %q: compare field %q is %s, not an amount", s.Name, name, fd.Type)
			}
			if prev, seen := kinds[measure]; seen && prev != fd.Type {
				return fmt.Errorf("matching: measure %q is %s on one side and %s on %q", measure, prev, fd.Type, s.Name)
			}
			kinds[measure] = fd.Type
			money = money || fd.Type == "money"
		}
		if money && s.Currency == "" {
			return fmt.Errorf("matching side %q carries money, so it needs a currency field", s.Name)
		}
		if s.Currency != "" {
			if fd, err := field("currency", s.Currency); err != nil {
				return err
			} else if fd.Type != "string" {
				return fmt.Errorf("matching side %q: currency field %q is %s, not a string", s.Name, s.Currency, fd.Type)
			}
		}
		if fd, err := field("date", s.Date); err != nil {
			return err
		} else if fd.Type != "date" {
			return fmt.Errorf("matching side %q: date field %q is %s, not a date", s.Name, s.Date, fd.Type)
		}
		if s.Reference != "" {
			if fd, err := field("reference", s.Reference); err != nil {
				return err
			} else if fd.Type != "string" {
				return fmt.Errorf("matching side %q: reference field %q is %s, not a string", s.Name, s.Reference, fd.Type)
			}
		}
		if s.Name == m.Trigger {
			continue
		}
		t, _ := ParseTemplate(s.Items)
		if err := t.f.CheckList(types, map[string]string{"$.state": trigger.Type}, s.Type); err != nil {
			return fmt.Errorf("matching side %q items: %w", s.Name, err)
		}
		tt, _ := types(trigger.Type)
		for cf, tf := range s.Join {
			cfd, err := field("join", cf)
			if err != nil {
				return err
			}
			tfd, ok := tt.Field(tf)
			if !ok {
				return fmt.Errorf("matching side %q: join field %q is not a field of the trigger's %s", s.Name, tf, trigger.Type)
			}
			if cfd.Type != tfd.Type {
				return fmt.Errorf("matching side %q: join %s (%s) with the trigger's %s (%s) — the fields differ in type", s.Name, cf, cfd.Type, tf, tfd.Type)
			}
		}
	}
	if m.Convert != nil {
		if kinds[m.AllocatedMeasures()[0]] != "money" {
			return fmt.Errorf("matching.convert works on money; %q is %s", m.AllocatedMeasures()[0], kinds[m.AllocatedMeasures()[0]])
		}
	}
	return nil
}

// MeasureKinds reports each measure's field type (money, int, decimal) from
// the catalog — the encoding of its amounts in the match event.
func (m *MatchingTemplate) MeasureKinds(types Catalog) (map[string]string, error) {
	kinds := map[string]string{}
	for _, s := range m.Sides {
		ot, ok := types(s.Type)
		if !ok {
			return nil, fmt.Errorf("%q is not a declared object type", s.Type)
		}
		for measure, name := range s.Open {
			fd, _ := ot.Field(name)
			kinds[measure] = fd.Type
		}
		for measure, name := range s.Compare {
			fd, _ := ot.Field(name)
			kinds[measure] = fd.Type
		}
	}
	return kinds, nil
}

// ParseDecimal reads a plain decimal string as an exact rational: optional
// sign, digits, optional fraction — the boundary form of rates, tolerances
// and decimal fields. Never a float.
func ParseDecimal(s string) (*big.Rat, error) { return parseDecimal(s) }

// CanonicalDecimal renders an exact rational as the shortest decimal that
// represents it, or fails when none does.
func CanonicalDecimal(r *big.Rat) (string, error) { return canonical(r) }

// RoundDecimal rounds an exact rational to the given places by an admitted
// method — the one place a kernel method may lose precision, declared.
func RoundDecimal(r *big.Rat, places int, method string) *big.Rat { return roundRat(r, places, method) }

// IsRoundingMethod reports whether a method is admitted.
func IsRoundingMethod(m string) bool { return slices.Contains(RoundingMethods, m) }
