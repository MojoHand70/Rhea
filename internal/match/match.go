// Package match is the deterministic matching engine (DIRECTION 2026-10-11,
// "Matching: a deterministic engine with N sides, as a kernel method"): one
// engine, N sides, a declared structure. A side is a keyed, capped collection
// of items, each carrying a vector of measures and its open portion; a
// strategy is a named standard — a search procedure that proposes a match
// under the structure, deterministic and total; a policy is an ordered list
// of strategies with their parameters, and the first feasible match wins.
//
// The package is pure: it knows nothing of events, rules or the store. The
// kernel builds a Problem from the trigger object and the other sides' keyed
// reads, runs Solve, and bakes the Result into one explained event. Numbers
// are exact rationals (no floats, invariant 6); every search is bounded by
// the side caps, the subset size cap and a node budget (security rule 1,
// scale rule 2); the same Problem always yields the same Result.
package match

import (
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"
)

// The strategies: standards with names. Order in a policy is the company's
// choice; the vocabulary is the kernel's.
const (
	// Reference: the items the trigger's reference names (a shared key);
	// feasible when the named items' open amounts agree with the trigger's,
	// or, with Partial, whenever something is named.
	Reference = "reference"
	// Exact: on every other side, one item whose open vector agrees with
	// the trigger's — the oldest such item.
	Exact = "exact"
	// Subset: on every other side, a subset of at most MaxItems items whose
	// open amounts sum to the trigger's, found by a bounded search that
	// prefers the oldest items.
	Subset = "subset"
	// OldestFirst: the trigger settles the earliest open items as far as it
	// reaches; the last may be partial and the remainder stays open. One
	// allocated measure.
	OldestFirst = "oldest_first"
	// Selected: a person's pairing — the sides hold exactly the items the
	// person chose. One allocated measure allocates them oldest first as
	// far as the trigger reaches; several require the selection to agree
	// with the trigger exactly. Never a policy's step: the verb's.
	Selected = "selected"
)

// Strategies lists the strategy names a policy may declare.
var Strategies = []string{Reference, Exact, Subset, OldestFirst}

// Known reports whether a strategy name is admitted.
func Known(strategy string) bool {
	for _, s := range Strategies {
		if s == strategy {
			return true
		}
	}
	return false
}

const (
	// DefaultMaxItems is the subset size cap when a step declares none;
	// MaxMaxItems is the most a step may declare.
	DefaultMaxItems = 4
	MaxMaxItems     = 8
	// DefaultBudget bounds one subset search in visited nodes; above it the
	// strategy reports "above the search cap" and the policy moves on —
	// the items go to the person, never to a long search.
	DefaultBudget = 100_000
)

// Item is one object on a side as the engine sees it: its open portion per
// allocated measure, its value per compared measure, and the attributes
// strategies read.
type Item struct {
	ID        string
	Open      map[string]*big.Rat // allocated measures → open portion
	Compare   map[string]*big.Rat // compared measures → value (absent when the item has none)
	Currency  string              // of the money measures; empty when none
	Date      string              // YYYY-MM-DD
	Reference string
	Seq       int // log order, the tie-breaker
}

// Side is one collection of candidates, in log order.
type Side struct {
	Name  string
	Items []Item
}

// Measure is one dimension items agree on. An allocated measure is consumed
// from open portions (money, a quantity); a compared one must agree within
// tolerance and consumes nothing (a price). Places renders amounts in
// explanations: 2 for money, 0 for counts, -1 for the shortest exact decimal.
type Measure struct {
	Name      string
	Allocated bool
	Money     bool // currencies must agree
	Places    int
}

// Step is one strategy with its parameters, as the policy declares it.
type Step struct {
	Strategy string
	// Tolerance is the absolute bound per measure within which values
	// agree; TolerancePercent a relative one, of the trigger's value. A
	// difference inside the bound is absorbed and recorded against
	// DifferenceAccount.
	Tolerance         map[string]*big.Rat
	TolerancePercent  map[string]*big.Rat
	DifferenceAccount string
	// WithinDays admits only candidates dated within so many days of the
	// trigger, either way; zero means any date.
	WithinDays int
	// MaxItems caps a subset's size; zero means DefaultMaxItems.
	MaxItems int
	// Partial lets the reference strategy settle what it can when the
	// named items and the trigger disagree.
	Partial bool
}

// Problem is one firing's question: the trigger item that arrived, the
// other sides' candidates, the measures, the policy.
type Problem struct {
	TriggerSide string
	Trigger     Item
	Sides       []Side
	Measures    []Measure
	Policy      []Step
	Budget      int // subset search node budget; zero means DefaultBudget
	// Convert compares money across currencies; nil means every item is
	// in the trigger's currency or excluded.
	Convert *Converter
}

// Converter is the FX clause as the engine sees it: Rate gives the rate
// from one currency to another at a date (1 when they are the same);
// Round rounds an amount to the money's places by the declared method.
// Every candidate's open is read in the trigger's currency at the rate of
// its settlement date — the trigger's date when DateSide is the trigger's
// side, the candidate's own date when it is on DateSide, the trigger's
// otherwise — and every allocation is written back in the item's own
// currency. An item's worth in the functional currency at its own date
// and at its settlement date differ by the rate's move: recorded as an
// "fx" difference, "up" or "down", against the matching account.
type Converter struct {
	Functional      string
	DateSide        string
	Rate            func(from, to, date string) (*big.Rat, error)
	Round           func(*big.Rat) *big.Rat
	RateUpAccount   string
	RateDownAccount string
}

// Allocation is what a match did to one item: the amount taken per
// allocated measure, and the open portion before and after — in the item's
// own currency; Equivalent is the same allocation in the trigger's
// currency when the structure converts.
type Allocation struct {
	ID         string
	Allocated  map[string]*big.Rat
	OpenBefore map[string]*big.Rat
	OpenAfter  map[string]*big.Rat
	Equivalent map[string]*big.Rat
}

// Difference is an amount a match absorbed: a tolerated gap between a side's
// allocation and the trigger's (Item empty), a compared measure's gap on
// one item, or an item's rate difference. Amount is positive; Direction
// says which way: "short" when the trigger brought less than the side
// settles, "over" when more; "up" when an item's rate rose between its own
// date and its settlement, "down" when it fell.
type Difference struct {
	Kind      string // "tolerance" | "fx"
	Measure   string
	Side      string
	Item      string
	Amount    *big.Rat
	Direction string
	Account   string
}

// Attempt is one policy step's outcome, for the explanation: the reason it
// found nothing, or Found.
type Attempt struct {
	Step     int
	Strategy string
	Found    bool
	Reason   string
}

// Result is the match: which step found it, the allocation on every side
// (the trigger side holds exactly the trigger), the differences absorbed,
// and every step tried up to and including the one that found it.
type Result struct {
	Step        int
	Strategy    string
	Sides       map[string][]Allocation
	Differences []Difference
	Tried       []Attempt
}

// Solve runs the policy and returns the first feasible match. ok is false
// when no step found one; Tried then explains every step. An error is a
// malformed problem — never a search that failed.
func Solve(p Problem) (Result, bool, error) {
	if err := p.validate(); err != nil {
		return Result{}, false, err
	}
	original, err := p.convert()
	if err != nil {
		return Result{}, false, err
	}
	var tried []Attempt
	for i, step := range p.Policy {
		r, reason := p.run(step)
		if r == nil {
			tried = append(tried, Attempt{Step: i + 1, Strategy: step.Strategy, Reason: reason})
			continue
		}
		tried = append(tried, Attempt{Step: i + 1, Strategy: step.Strategy, Found: true})
		r.Step, r.Strategy, r.Tried = i+1, step.Strategy, tried
		if err := p.restore(r, original); err != nil {
			return Result{}, false, err
		}
		return *r, true, nil
	}
	return Result{Tried: tried}, false, nil
}

// settleDate is the date whose rate settles an item.
func (p Problem) settleDate(side string, it Item) string {
	if p.Convert.DateSide == p.TriggerSide || side != p.Convert.DateSide {
		return p.Trigger.Date
	}
	return it.Date
}

// convert is the pre-pass: with a converter, every candidate's open money
// is replaced by its equivalent in the trigger's currency, so the
// strategies compare like with like. The originals come back for restore.
func (p *Problem) convert() (map[string]Item, error) {
	if p.Convert == nil {
		return nil, nil
	}
	m := p.allocated()[0].Name
	original := map[string]Item{}
	// The caller's sides stay as they are: the converted view is the
	// engine's own copy.
	sides := make([]Side, len(p.Sides))
	for si, side := range p.Sides {
		sides[si] = Side{Name: side.Name, Items: make([]Item, len(side.Items))}
		for ii, it := range side.Items {
			original[it.ID] = it
			sides[si].Items[ii] = it
			if it.Currency == p.Trigger.Currency || it.Open[m] == nil {
				continue
			}
			rate, err := p.Convert.Rate(it.Currency, p.Trigger.Currency, p.settleDate(side.Name, it))
			if err != nil {
				return nil, err
			}
			sides[si].Items[ii] = Item{ID: it.ID, Open: map[string]*big.Rat{m: new(big.Rat).Mul(it.Open[m], rate)},
				Compare: it.Compare, Currency: p.Trigger.Currency, Date: it.Date, Reference: it.Reference, Seq: it.Seq}
		}
	}
	p.Sides = sides
	return original, nil
}

// restore is the post-pass: allocations found in the trigger's currency
// are written back in each item's own currency — a whole item takes its
// own open exactly, a partial one the rounded equivalent — the equivalent
// kept beside them, and the rate differences are recorded.
func (p Problem) restore(r *Result, original map[string]Item) error {
	if p.Convert == nil {
		return nil
	}
	m := p.allocated()[0].Name
	c := p.Convert
	sideOf := map[string]string{p.Trigger.ID: p.TriggerSide}
	for _, side := range p.Sides {
		for _, it := range side.Items {
			sideOf[it.ID] = side.Name
		}
	}
	latestDateSide := p.Trigger.Date // for the trigger's own settlement when DateSide is elsewhere
	for _, a := range r.Sides[c.DateSide] {
		if o, ok := original[a.ID]; ok && o.Date > latestDateSide {
			latestDateSide = o.Date
		}
	}
	// Side names in a fixed order: the differences list is part of the
	// derived event, and rules fanning out over it mint object ids from
	// their position.
	names := make([]string, 0, len(r.Sides))
	for name := range r.Sides {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		allocs := r.Sides[name]
		for i := range allocs {
			a := &allocs[i]
			o, converted := original[a.ID]
			if !converted {
				o = p.Trigger
			}
			if converted && o.Currency != p.Trigger.Currency {
				equivalent := a.Allocated[m]
				var own *big.Rat
				if a.OpenAfter[m].Sign() == 0 {
					own = new(big.Rat).Set(o.Open[m]) // whole: its own open, exactly
				} else {
					ratio := new(big.Rat).Quo(o.Open[m], a.OpenBefore[m]) // 1 / rate
					own = c.Round(new(big.Rat).Mul(equivalent, ratio))
				}
				a.Equivalent = map[string]*big.Rat{m: equivalent}
				a.Allocated = map[string]*big.Rat{m: own}
				a.OpenBefore = map[string]*big.Rat{m: new(big.Rat).Set(o.Open[m])}
				a.OpenAfter = map[string]*big.Rat{m: new(big.Rat).Sub(o.Open[m], own)}
			}
			// The rate difference: what the allocated amount was worth in
			// the functional currency at the item's own date against its
			// settlement date.
			if o.Currency == c.Functional {
				continue
			}
			settleDate := p.settleDate(name, o)
			if !converted && c.DateSide != p.TriggerSide {
				settleDate = latestDateSide
			}
			var settled *big.Rat
			if p.Trigger.Currency == c.Functional && converted {
				settled = a.Equivalent[m] // what was actually allocated, exactly
			} else {
				rate, err := c.Rate(o.Currency, c.Functional, settleDate)
				if err != nil {
					return err
				}
				settled = c.Round(new(big.Rat).Mul(a.Allocated[m], rate))
			}
			booked, err := c.Rate(o.Currency, c.Functional, o.Date)
			if err != nil {
				return err
			}
			bookedValue := c.Round(new(big.Rat).Mul(a.Allocated[m], booked))
			if d := new(big.Rat).Sub(settled, bookedValue); d.Sign() != 0 {
				dir, account := "up", c.RateUpAccount
				if d.Sign() < 0 {
					dir, account, d = "down", c.RateDownAccount, d.Neg(d)
				}
				r.Differences = append(r.Differences, Difference{Kind: "fx", Measure: m, Side: name, Item: a.ID,
					Amount: d, Direction: dir, Account: account})
			}
		}
		r.Sides[name] = allocs
	}
	return nil
}

func (p Problem) validate() error {
	if p.TriggerSide == "" || p.Trigger.ID == "" {
		return fmt.Errorf("a problem needs a trigger")
	}
	if len(p.Sides) == 0 {
		return fmt.Errorf("a problem needs at least one side besides the trigger's")
	}
	var allocated []string
	for _, m := range p.Measures {
		if m.Allocated {
			allocated = append(allocated, m.Name)
		}
	}
	if len(allocated) == 0 {
		return fmt.Errorf("a problem needs an allocated measure")
	}
	for _, m := range allocated {
		if p.Trigger.Open[m] == nil {
			return fmt.Errorf("the trigger has no open %s", m)
		}
	}
	if c := p.Convert; c != nil {
		if len(allocated) != 1 || !p.measure(allocated[0]).Money {
			return fmt.Errorf("a converter works on one allocated money measure")
		}
		if c.Rate == nil || c.Round == nil {
			return fmt.Errorf("a converter needs a rate source and a rounding")
		}
	}
	for i, s := range p.Policy {
		if !Known(s.Strategy) && s.Strategy != Selected {
			return fmt.Errorf("step %d: unknown strategy %q", i+1, s.Strategy)
		}
		if (s.Strategy == OldestFirst || (s.Strategy == Reference && s.Partial)) && len(allocated) != 1 {
			return fmt.Errorf("step %d: %s allocates one measure, the structure has %d", i+1, s.Strategy, len(allocated))
		}
		if s.MaxItems > MaxMaxItems {
			return fmt.Errorf("step %d: max_items %d is above the cap of %d", i+1, s.MaxItems, MaxMaxItems)
		}
	}
	return nil
}

func (p Problem) allocated() []Measure {
	var out []Measure
	for _, m := range p.Measures {
		if m.Allocated {
			out = append(out, m)
		}
	}
	return out
}

func (p Problem) measure(name string) Measure {
	for _, m := range p.Measures {
		if m.Name == name {
			return m
		}
	}
	return Measure{Name: name, Places: -1}
}

// run tries one step. A nil result carries the reason.
func (p Problem) run(step Step) (*Result, string) {
	for _, m := range p.allocated() {
		if p.Trigger.Open[m.Name].Sign() <= 0 {
			return nil, fmt.Sprintf("the %s has nothing open", p.TriggerSide)
		}
	}
	// selection is what each strategy decides: which items on each side,
	// and for partial strategies how much of the trigger is matched.
	var (
		selected = map[string][]Item{}
		matched  = map[string]*big.Rat{} // per allocated measure, the amount the trigger gives
		partial  = false
	)
	for _, m := range p.allocated() {
		matched[m.Name] = new(big.Rat).Set(p.Trigger.Open[m.Name])
	}
	switch step.Strategy {
	case Reference:
		if p.Trigger.Reference == "" {
			return nil, fmt.Sprintf("the %s names no reference", p.TriggerSide)
		}
		named := map[string][]Item{}
		for _, side := range p.Sides {
			cands, _ := p.candidates(side, step)
			for _, c := range cands {
				if names(p.Trigger.Reference, c.Reference) {
					named[side.Name] = append(named[side.Name], c)
				}
			}
			if len(named[side.Name]) == 0 {
				return nil, fmt.Sprintf("%q names nothing open on %s", p.Trigger.Reference, side.Name)
			}
		}
		for _, side := range p.Sides {
			sums := sumOpen(named[side.Name], p.allocated())
			if p.agree(step, sums, p.Trigger.Open) {
				selected[side.Name] = named[side.Name]
				continue
			}
			if !step.Partial {
				m := p.allocated()[0]
				return nil, fmt.Sprintf("%q names %s on %s, the %s brings %s",
					p.Trigger.Reference, p.show(m, sums[m.Name]), side.Name, p.TriggerSide, p.show(m, p.Trigger.Open[m.Name]))
			}
			partial = true
			m := p.allocated()[0].Name
			if sums[m].Cmp(matched[m]) < 0 {
				matched[m] = sums[m] // the named items settle in full; the trigger keeps the rest
			}
			selected[side.Name] = named[side.Name]
		}
	case Exact:
		for _, side := range p.Sides {
			cands, excluded := p.candidates(side, step)
			var hit *Item
			for i := range cands {
				if p.agree(step, cands[i].Open, p.Trigger.Open) {
					hit = &cands[i]
					break
				}
			}
			if hit == nil {
				return nil, p.nothing(side, step, cands, excluded, "no single item")
			}
			selected[side.Name] = []Item{*hit}
		}
	case Subset:
		budget := p.Budget
		if budget <= 0 {
			budget = DefaultBudget
		}
		maxItems := step.MaxItems
		if maxItems <= 0 {
			maxItems = DefaultMaxItems
		}
		for _, side := range p.Sides {
			cands, excluded := p.candidates(side, step)
			found, ok, exhausted := p.subset(cands, step, maxItems, &budget)
			if exhausted {
				return nil, fmt.Sprintf("the search on %s went above its cap of %d steps — a person decides", side.Name, DefaultBudget)
			}
			if !ok {
				return nil, p.nothing(side, step, cands, excluded, fmt.Sprintf("no subset of up to %d items", maxItems))
			}
			selected[side.Name] = found
		}
	case OldestFirst:
		partial = true
		m := p.allocated()[0].Name
		for _, side := range p.Sides {
			cands, excluded := p.candidates(side, step)
			if len(cands) == 0 {
				return nil, p.nothing(side, step, cands, excluded, "nothing open")
			}
			sums := sumOpen(cands, p.allocated())
			if sums[m].Cmp(matched[m]) < 0 {
				matched[m] = sums[m]
			}
			selected[side.Name] = cands
		}
	case Selected:
		for _, side := range p.Sides {
			cands, excluded := p.candidates(side, step)
			if len(cands) == 0 {
				return nil, p.nothing(side, step, cands, excluded, "nothing open among the chosen")
			}
			sums := sumOpen(cands, p.allocated())
			if len(p.allocated()) == 1 {
				partial = true
				m := p.allocated()[0].Name
				if sums[m].Cmp(matched[m]) < 0 {
					matched[m] = sums[m]
				}
			} else if !p.agree(step, sums, p.Trigger.Open) {
				m := p.allocated()[0]
				return nil, fmt.Sprintf("the chosen items on %s do not add up: %s against the %s's %s",
					side.Name, p.show(m, sums[m.Name]), p.TriggerSide, p.show(m, p.Trigger.Open[m.Name]))
			}
			selected[side.Name] = cands
		}
	}
	if reason := p.compared(step, selected); reason != "" {
		return nil, reason
	}
	return p.allocate(step, selected, matched, partial), ""
}

// candidates filters and orders one side's items for a step: open on every
// allocated measure, in the trigger's currency, within the step's days,
// agreeing on every compared measure the trigger carries; oldest first,
// then log order. The excluded items come back with reasons, for the
// explanation when nothing matches.
func (p Problem) candidates(side Side, step Step) ([]Item, []string) {
	var out []Item
	var excluded []string
	for _, it := range side.Items {
		if reason := p.exclude(it, step); reason != "" {
			excluded = append(excluded, it.ID+": "+reason)
			continue
		}
		out = append(out, it)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Date != out[j].Date {
			return out[i].Date < out[j].Date
		}
		return out[i].Seq < out[j].Seq
	})
	return out, excluded
}

func (p Problem) exclude(it Item, step Step) string {
	for _, m := range p.allocated() {
		v := it.Open[m.Name]
		if v == nil || v.Sign() <= 0 {
			return "nothing open"
		}
		if m.Money && it.Currency != p.Trigger.Currency && p.Convert == nil {
			return fmt.Sprintf("in %s, the %s in %s", it.Currency, p.TriggerSide, p.Trigger.Currency)
		}
	}
	if step.WithinDays > 0 {
		d, err := daysBetween(p.Trigger.Date, it.Date)
		if err != nil {
			return err.Error()
		}
		if d > step.WithinDays {
			return fmt.Sprintf("%d days from the %s, the step allows %d", d, p.TriggerSide, step.WithinDays)
		}
	}
	for _, m := range p.Measures {
		if m.Allocated {
			continue
		}
		tv, iv := p.Trigger.Compare[m.Name], it.Compare[m.Name]
		if tv == nil || iv == nil {
			continue
		}
		if !p.within(step, m, tv, iv) {
			return fmt.Sprintf("%s %s, the %s says %s", m.Name, p.show(m, iv), p.TriggerSide, p.show(m, tv))
		}
	}
	return ""
}

// nothing phrases a side's empty outcome: what was looked for (with the
// amount, when an amount was looked for), how many candidates there were,
// and why the rest were left out.
func (p Problem) nothing(side Side, step Step, cands []Item, excluded []string, what string) string {
	m := p.allocated()[0]
	s := fmt.Sprintf("%s on %s", what, side.Name)
	if what != "nothing open" {
		s += fmt.Sprintf(" opens at %s", p.show(m, p.Trigger.Open[m.Name]))
		if tol := step.Tolerance[m.Name]; tol != nil && tol.Sign() > 0 {
			s += fmt.Sprintf(" ± %s", p.show(m, tol))
		}
	}
	s += fmt.Sprintf(" (%d candidates", len(cands))
	if len(excluded) > 0 {
		s += fmt.Sprintf(", %d left out: %s", len(excluded), strings.Join(excluded, "; "))
	}
	return s + ")"
}

// subset searches one side for items whose open amounts sum to the
// trigger's on every allocated measure, within the step's tolerance: the
// fewest items first (a payment is explained by as few documents as
// possible), and among subsets of one size the one including the oldest
// candidates first — a depth-first search over the candidates in date
// order, including before excluding, pruned where a partial sum already
// exceeds the target, capped in size and in visited nodes. The first subset
// found in that order wins, so the same side always yields the same subset.
func (p Problem) subset(cands []Item, step Step, maxItems int, budget *int) (found []Item, ok, exhausted bool) {
	measures := p.allocated()
	sums := map[string]*big.Rat{}
	for _, m := range measures {
		sums[m.Name] = new(big.Rat)
	}
	bounds := map[string]*big.Rat{}
	for _, m := range measures {
		bounds[m.Name] = new(big.Rat).Add(p.Trigger.Open[m.Name], p.bound(step, m, p.Trigger.Open[m.Name]))
	}
	var chosen []Item
	var walk func(i, size int) bool
	walk = func(i, size int) bool {
		if len(chosen) == size {
			if p.agree(step, sums, p.Trigger.Open) {
				found = append([]Item(nil), chosen...)
				return true
			}
			return false
		}
		if len(cands)-i < size-len(chosen) {
			return false // not enough candidates left to reach the size
		}
		if *budget == 0 {
			exhausted = true
			return false
		}
		*budget--
		// include cands[i], unless it overshoots on some measure
		over := false
		for _, m := range measures {
			if new(big.Rat).Add(sums[m.Name], cands[i].Open[m.Name]).Cmp(bounds[m.Name]) > 0 {
				over = true
			}
		}
		if !over {
			for _, m := range measures {
				sums[m.Name].Add(sums[m.Name], cands[i].Open[m.Name])
			}
			chosen = append(chosen, cands[i])
			if walk(i+1, size) {
				return true
			}
			chosen = chosen[:len(chosen)-1]
			for _, m := range measures {
				sums[m.Name].Sub(sums[m.Name], cands[i].Open[m.Name])
			}
			if exhausted {
				return false
			}
		}
		return walk(i+1, size)
	}
	for size := 1; size <= maxItems && size <= len(cands); size++ {
		if walk(0, size) {
			return found, true, false
		}
		if exhausted {
			return nil, false, true
		}
	}
	return nil, false, false
}

// compared checks the compared measures the trigger does not carry across
// the selected items: every item carrying one agrees with the first that
// does, in side order.
func (p Problem) compared(step Step, selected map[string][]Item) string {
	for _, m := range p.Measures {
		if m.Allocated || p.Trigger.Compare[m.Name] != nil {
			continue
		}
		var ref *big.Rat
		var refItem string
		for _, side := range p.Sides {
			for _, it := range selected[side.Name] {
				v := it.Compare[m.Name]
				if v == nil {
					continue
				}
				if ref == nil {
					ref, refItem = v, it.ID
					continue
				}
				if !p.within(step, m, ref, v) {
					return fmt.Sprintf("%s disagrees: %s says %s, %s says %s", m.Name, refItem, p.show(m, ref), it.ID, p.show(m, v))
				}
			}
		}
	}
	return ""
}

// allocate turns a selection into allocations and differences. Full
// strategies take every selected item whole; partial ones give each side
// the matched amount oldest first, the last item possibly in part.
func (p Problem) allocate(step Step, selected map[string][]Item, matched map[string]*big.Rat, partial bool) *Result {
	r := &Result{Sides: map[string][]Allocation{}}
	measures := p.allocated()
	for _, side := range p.Sides {
		items := selected[side.Name]
		var allocs []Allocation
		if !partial {
			for _, it := range items {
				allocs = append(allocs, full(it, measures))
			}
		} else {
			m := measures[0].Name
			left := new(big.Rat).Set(matched[m])
			for _, it := range items {
				if left.Sign() <= 0 {
					break
				}
				take := new(big.Rat).Set(it.Open[m])
				if take.Cmp(left) > 0 {
					take.Set(left)
				}
				left.Sub(left, take)
				allocs = append(allocs, Allocation{ID: it.ID,
					Allocated:  map[string]*big.Rat{m: take},
					OpenBefore: map[string]*big.Rat{m: new(big.Rat).Set(it.Open[m])},
					OpenAfter:  map[string]*big.Rat{m: new(big.Rat).Sub(it.Open[m], take)}})
			}
		}
		r.Sides[side.Name] = allocs
		// A side's allocation may differ from the trigger's by what the
		// tolerance absorbed; the difference is recorded against the step's
		// account, with its direction.
		for _, m := range measures {
			sum := new(big.Rat)
			for _, a := range allocs {
				sum.Add(sum, a.Allocated[m.Name])
			}
			if d := new(big.Rat).Sub(sum, matched[m.Name]); d.Sign() != 0 {
				dir := "short"
				if d.Sign() < 0 {
					dir, d = "over", d.Neg(d)
				}
				r.Differences = append(r.Differences, Difference{Kind: "tolerance", Measure: m.Name,
					Side: side.Name, Amount: d, Direction: dir, Account: step.DifferenceAccount})
			}
		}
		// Compared measures the trigger carries: each selected item's gap.
		for _, m := range p.Measures {
			if m.Allocated || p.Trigger.Compare[m.Name] == nil {
				continue
			}
			for _, it := range items {
				v := it.Compare[m.Name]
				if v == nil {
					continue
				}
				if d := new(big.Rat).Sub(v, p.Trigger.Compare[m.Name]); d.Sign() != 0 {
					dir := "short"
					if d.Sign() < 0 {
						dir, d = "over", d.Neg(d)
					}
					r.Differences = append(r.Differences, Difference{Kind: "tolerance", Measure: m.Name,
						Side: side.Name, Item: it.ID, Amount: d, Direction: dir, Account: step.DifferenceAccount})
				}
			}
		}
	}
	trig := Allocation{ID: p.Trigger.ID, Allocated: map[string]*big.Rat{},
		OpenBefore: map[string]*big.Rat{}, OpenAfter: map[string]*big.Rat{}}
	for _, m := range measures {
		trig.Allocated[m.Name] = new(big.Rat).Set(matched[m.Name])
		trig.OpenBefore[m.Name] = new(big.Rat).Set(p.Trigger.Open[m.Name])
		trig.OpenAfter[m.Name] = new(big.Rat).Sub(p.Trigger.Open[m.Name], matched[m.Name])
	}
	r.Sides[p.TriggerSide] = []Allocation{trig}
	return r
}

func full(it Item, measures []Measure) Allocation {
	a := Allocation{ID: it.ID, Allocated: map[string]*big.Rat{},
		OpenBefore: map[string]*big.Rat{}, OpenAfter: map[string]*big.Rat{}}
	for _, m := range measures {
		a.Allocated[m.Name] = new(big.Rat).Set(it.Open[m.Name])
		a.OpenBefore[m.Name] = new(big.Rat).Set(it.Open[m.Name])
		a.OpenAfter[m.Name] = new(big.Rat)
	}
	return a
}

// agree reports whether two open vectors agree on every allocated measure
// within the step's tolerance.
func (p Problem) agree(step Step, a, b map[string]*big.Rat) bool {
	for _, m := range p.allocated() {
		if a[m.Name] == nil || b[m.Name] == nil || !p.within(step, m, b[m.Name], a[m.Name]) {
			return false
		}
	}
	return true
}

// within reports whether value agrees with the reference within the step's
// bound for the measure.
func (p Problem) within(step Step, m Measure, reference, value *big.Rat) bool {
	d := new(big.Rat).Sub(value, reference)
	if d.Sign() < 0 {
		d.Neg(d)
	}
	return d.Cmp(p.bound(step, m, reference)) <= 0
}

// bound is the step's tolerance for a measure around a reference value:
// the absolute bound plus the relative one.
func (p Problem) bound(step Step, m Measure, reference *big.Rat) *big.Rat {
	b := new(big.Rat)
	if t := step.Tolerance[m.Name]; t != nil {
		b.Add(b, t)
	}
	if pct := step.TolerancePercent[m.Name]; pct != nil {
		rel := new(big.Rat).Mul(reference, pct)
		rel.Quo(rel, big.NewRat(100, 1))
		if rel.Sign() < 0 {
			rel.Neg(rel)
		}
		b.Add(b, rel)
	}
	return b
}

func sumOpen(items []Item, measures []Measure) map[string]*big.Rat {
	sums := map[string]*big.Rat{}
	for _, m := range measures {
		sums[m.Name] = new(big.Rat)
		for _, it := range items {
			if v := it.Open[m.Name]; v != nil {
				sums[m.Name].Add(sums[m.Name], v)
			}
		}
	}
	return sums
}

// show renders an amount for an explanation, by the measure's places.
func (p Problem) show(m Measure, r *big.Rat) string {
	if r == nil {
		return "nothing"
	}
	if m.Places >= 0 {
		return r.FloatString(m.Places)
	}
	for places := 0; places <= 18; places++ {
		scaled := new(big.Int).Mul(r.Num(), new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(places)), nil))
		if new(big.Int).Mod(scaled, r.Denom()).Sign() == 0 {
			return r.FloatString(places)
		}
	}
	return r.RatString()
}

// names reports whether a reference text names a key: the key appears in
// the text whole — not as part of a longer token, so FV/1 is not named by
// "FV/12" — regardless of case.
func names(text, key string) bool {
	text, key = strings.ToLower(strings.TrimSpace(text)), strings.ToLower(strings.TrimSpace(key))
	if key == "" {
		return false
	}
	for from := 0; from <= len(text)-len(key); {
		at := strings.Index(text[from:], key)
		if at < 0 {
			return false
		}
		start, end := from+at, from+at+len(key)
		if (start == 0 || !alnum(text[start-1])) && (end == len(text) || !alnum(text[end])) {
			return true
		}
		from = start + 1
	}
	return false
}

func alnum(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func daysBetween(a, b string) (int, error) {
	ta, err := time.Parse("2006-01-02", a)
	if err != nil {
		return 0, fmt.Errorf("date %q", a)
	}
	tb, err := time.Parse("2006-01-02", b)
	if err != nil {
		return 0, fmt.Errorf("date %q", b)
	}
	d := int(tb.Sub(ta).Hours() / 24)
	if d < 0 {
		d = -d
	}
	return d, nil
}
