// The matching effect's firing path (DIRECTION 2026-10-11, the matching
// brief; DECISIONS 2026-10-11, the attempt). A matching rule fires on the
// materialization of its trigger side's object; the kernel reads the other
// sides' candidates by key through the chain-visible world, builds the
// engine's problem, runs the policy, and bakes the outcome into one derived
// event — objects.matched with every allocation and difference, or
// objects.unmatched with every step tried and why. The engine is pure
// (internal/match); this file is the wiring: fields to measures, amounts to
// exact rationals and back to the boundary encoding, accounts resolved.
// Replay never re-matches: the amendments a match causes are baked events.
package exec

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"rhea/internal/core"
	"rhea/internal/match"
)

// expandMatching runs one matching rule on the materialization payload it
// matched. Exactly one of the two results is non-nil on success.
func (x *Executor) expandMatching(ctx context.Context, r core.Rule, payload any, lookup core.Lookup, get core.Getter) (*core.MatchedObjects, *core.UnmatchedObjects, error) {
	tmpl := r.Spec.Effect.Matching
	pm, _ := payload.(map[string]any)
	objectID, _ := pm["object_id"].(string)
	objectType, _ := pm["object_type"].(string)
	state, _ := pm["state"].(map[string]any)
	triggerSide, _ := tmpl.Side(tmpl.Trigger)
	if objectID == "" || state == nil || objectType != triggerSide.Type {
		return nil, nil, fmt.Errorf("matching %s is triggered by a %s, not by %q", tmpl.Name, triggerSide.Type, objectType)
	}
	env := x.env(ctx, payload, lookup, get)
	kinds, err := tmpl.MeasureKinds(env.Types)
	if err != nil {
		return nil, nil, fmt.Errorf("matching %s: %w", tmpl.Name, err)
	}
	trigger := core.MatchTrigger{Side: tmpl.Trigger, ObjectID: objectID, ObjectType: objectType}
	triggerItem, err := itemOf(triggerSide, objectID, state, kinds, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("matching %s: the %s %s: %w", tmpl.Name, triggerSide.Name, objectID, err)
	}
	sides, inputs, unmatched := x.readSides(tmpl, trigger, state, env, kinds)
	if unmatched != nil {
		return nil, unmatched, nil
	}
	problem := match.Problem{TriggerSide: tmpl.Trigger, Trigger: triggerItem, Sides: sides,
		Measures: measuresOf(tmpl, kinds), Policy: policySteps(tmpl, inputs),
		Convert: converterOf(tmpl, lookup, get, inputs)}
	result, found, err := match.Solve(problem)
	if err != nil {
		return nil, nil, fmt.Errorf("matching %s: %w", tmpl.Name, err)
	}
	if !found {
		return nil, &core.UnmatchedObjects{Matching: tmpl.Name, Trigger: trigger, Tried: attempts(result.Tried), Inputs: inputs}, nil
	}
	m, err := bakeMatch(tmpl, trigger, triggerItem, result, kinds, map[string]map[string]any{objectID: state}, inputs, lookup, get)
	if err != nil {
		return nil, nil, err
	}
	return m, nil, nil
}

// readSides reads every side but the trigger's as the structure declares:
// the keyed read against the trigger's payload, the join, the cap. A side
// that cannot be read or is above its cap ends the firing as unmatched with
// the reason — the trigger stays open for the person, and the log says why;
// its own materialization is never held hostage.
func (x *Executor) readSides(tmpl *core.MatchingTemplate, trigger core.MatchTrigger, state map[string]any, env *core.Env, kinds map[string]string) ([]match.Side, map[string]any, *core.UnmatchedObjects) {
	inputs := map[string]any{}
	payload := map[string]any{"object_id": trigger.ObjectID, "object_type": trigger.ObjectType, "state": state}
	var sides []match.Side
	for _, side := range tmpl.Sides {
		if side.Name == tmpl.Trigger {
			continue
		}
		t, err := core.ParseTemplate(side.Items)
		if err != nil {
			return nil, inputs, &core.UnmatchedObjects{Matching: tmpl.Name, Trigger: trigger, Inputs: inputs,
				Reason: fmt.Sprintf("%s cannot be read: %v", side.Name, err)}
		}
		reads, in, err := t.Objects(payload, env)
		if err != nil {
			return nil, inputs, &core.UnmatchedObjects{Matching: tmpl.Name, Trigger: trigger, Inputs: inputs,
				Reason: fmt.Sprintf("%s cannot be read: %v", side.Name, err)}
		}
		for k, v := range in {
			inputs[k] = v
		}
		s := match.Side{Name: side.Name}
		ids := make([]string, 0, len(reads))
		open := 0
		for i, rd := range reads {
			// The join: a candidate that disagrees with the trigger on a
			// joined field is not on this side at all.
			joined := true
			for cf, tf := range side.Join {
				if core.Text(rd.State[cf]) != core.Text(state[tf]) {
					joined = false
				}
			}
			if !joined {
				continue
			}
			it, err := itemOf(side, rd.ID, rd.State, kinds, i+1)
			if err != nil {
				return nil, inputs, &core.UnmatchedObjects{Matching: tmpl.Name, Trigger: trigger, Inputs: inputs,
					Reason: fmt.Sprintf("the %s %s cannot be read: %v", side.Name, rd.ID, err)}
			}
			s.Items = append(s.Items, it)
			ids = append(ids, rd.ID)
			if isOpen(it) {
				open++
			}
		}
		inputs["candidates."+side.Name] = ids
		for cf, tf := range side.Join {
			inputs[fmt.Sprintf("join.%s.%s", side.Name, cf)] = state[tf]
		}
		// The read returns every object under the key, settled ones
		// included (the key is one field; the VM's collection cap bounds
		// the read itself); the side's cap bounds what the engine may
		// search — the open ones. Above it the items go to the person,
		// never to a long search.
		capacity := side.Cap
		if capacity <= 0 {
			capacity = core.DefaultMatchingCap
		}
		if open > capacity {
			return nil, inputs, &core.UnmatchedObjects{Matching: tmpl.Name, Trigger: trigger, Inputs: inputs,
				Reason: fmt.Sprintf("%s holds %d open candidates, above the cap of %d — a person decides", side.Name, open, capacity)}
		}
		sides = append(sides, s)
	}
	return sides, inputs, nil
}

// learnable lists the standards the host may learn from a person's pairing,
// each as the step it would propose.
var learnable = []core.MatchingStep{
	{Strategy: match.Reference},
	{Strategy: match.Reference, Partial: true},
	{Strategy: match.Exact},
	{Strategy: match.Subset, MaxItems: match.MaxMaxItems},
	{Strategy: match.OldestFirst},
}

// standardsFor names the strategies that, run alone over the sides as the
// policy reads them, would have found exactly the person's pairing: the
// same items on every side.
func (x *Executor) standardsFor(ctx context.Context, tmpl *core.MatchingTemplate, matched *core.MatchedObjects, lookup core.Lookup, get core.Getter) ([]string, error) {
	state, ok, err := get(matched.Trigger.ObjectID)
	if err != nil || !ok {
		return nil, err
	}
	payload := map[string]any{"object_id": matched.Trigger.ObjectID, "object_type": matched.Trigger.ObjectType, "state": state}
	env := x.env(ctx, payload, lookup, get)
	kinds, err := tmpl.MeasureKinds(env.Types)
	if err != nil {
		return nil, err
	}
	triggerSide, _ := tmpl.Side(tmpl.Trigger)
	triggerItem, err := itemOf(triggerSide, matched.Trigger.ObjectID, state, kinds, 0)
	if err != nil {
		return nil, err
	}
	sides, _, unmatched := x.readSides(tmpl, matched.Trigger, state, env, kinds)
	if unmatched != nil {
		return nil, nil // nothing a standard could have read
	}
	chosen := map[string]map[string]bool{}
	for name, allocs := range matched.Sides {
		if name == tmpl.Trigger {
			continue
		}
		chosen[name] = map[string]bool{}
		for _, a := range allocs {
			chosen[name][a.ObjectID] = true
		}
	}
	var out []string
	for _, step := range learnable {
		trial := &core.MatchingTemplate{Name: tmpl.Name, Trigger: tmpl.Trigger, Sides: tmpl.Sides, Policy: []core.MatchingStep{step}, Convert: tmpl.Convert}
		problem := match.Problem{TriggerSide: tmpl.Trigger, Trigger: triggerItem, Sides: sides,
			Measures: measuresOf(tmpl, kinds), Policy: policySteps(trial, map[string]any{}),
			Convert: converterOf(tmpl, lookup, get, map[string]any{})}
		result, found, err := match.Solve(problem)
		if err != nil || !found {
			continue
		}
		same := true
		for name, want := range chosen {
			got := map[string]bool{}
			for _, a := range result.Sides[name] {
				got[a.ID] = true
			}
			if len(got) != len(want) {
				same = false
			}
			for id := range want {
				if !got[id] {
					same = false
				}
			}
		}
		if same {
			out = append(out, standardName(step))
		}
	}
	return out, nil
}

// standardName spells a learnable step the way the policy would.
func standardName(step core.MatchingStep) string {
	if step.Strategy == match.Reference && step.Partial {
		return "reference partial"
	}
	return step.Strategy
}

// hasStandard reports whether a policy already runs a learnable standard.
func hasStandard(policy []core.MatchingStep, standard string) bool {
	for _, st := range policy {
		if standardName(st) == standard {
			return true
		}
	}
	return false
}

// pairSelection runs a matching rule's structure on a person's selection
// (the pair_items verb): the trigger item and exactly the items chosen on
// the other sides, under the Selected strategy — never the policy. The
// event it yields is the policy's, explained the same way.
func (x *Executor) pairSelection(ctx context.Context, r core.Rule, triggerID string, itemIDs []string, lookup core.Lookup, get core.Getter) (*core.MatchedObjects, error) {
	tmpl := r.Spec.Effect.Matching
	if tmpl == nil {
		return nil, fmt.Errorf("rule %s is not a matching rule", r.ID)
	}
	env := x.env(ctx, nil, lookup, get)
	kinds, err := tmpl.MeasureKinds(env.Types)
	if err != nil {
		return nil, fmt.Errorf("matching %s: %w", tmpl.Name, err)
	}
	triggerSide, _ := tmpl.Side(tmpl.Trigger)
	state, ok, err := get(triggerID)
	if err != nil {
		return nil, err
	}
	if !ok || !strings.HasPrefix(triggerID, triggerSide.Type+"-") {
		return nil, fmt.Errorf("matching %s: %q is not a %s — the trigger side is %s", tmpl.Name, triggerID, triggerSide.Type, triggerSide.Name)
	}
	trigger := core.MatchTrigger{Side: tmpl.Trigger, ObjectID: triggerID, ObjectType: triggerSide.Type}
	triggerItem, err := itemOf(triggerSide, triggerID, state, kinds, 0)
	if err != nil {
		return nil, fmt.Errorf("matching %s: the %s %s: %w", tmpl.Name, triggerSide.Name, triggerID, err)
	}
	states := map[string]map[string]any{triggerID: state}
	bySide := map[string]*match.Side{}
	var sides []match.Side
	inputs := map[string]any{"selection": itemIDs}
	for i, id := range itemIDs {
		st, ok, err := get(id)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("matching %s: no object %q among the chosen", tmpl.Name, id)
		}
		var side *core.MatchingSide
		for j := range tmpl.Sides {
			if tmpl.Sides[j].Name != tmpl.Trigger && strings.HasPrefix(id, tmpl.Sides[j].Type+"-") {
				side = &tmpl.Sides[j]
			}
		}
		if side == nil {
			return nil, fmt.Errorf("matching %s: %q is on no side but the trigger's", tmpl.Name, id)
		}
		it, err := itemOf(*side, id, st, kinds, i+1)
		if err != nil {
			return nil, fmt.Errorf("matching %s: the %s %s: %w", tmpl.Name, side.Name, id, err)
		}
		states[id] = st
		if bySide[side.Name] == nil {
			bySide[side.Name] = &match.Side{Name: side.Name}
		}
		bySide[side.Name].Items = append(bySide[side.Name].Items, it)
	}
	for _, side := range tmpl.Sides {
		if side.Name == tmpl.Trigger {
			continue
		}
		s := bySide[side.Name]
		if s == nil {
			return nil, fmt.Errorf("matching %s: nothing chosen on %s", tmpl.Name, side.Name)
		}
		sides = append(sides, *s)
	}
	problem := match.Problem{TriggerSide: tmpl.Trigger, Trigger: triggerItem, Sides: sides,
		Measures: measuresOf(tmpl, kinds), Policy: []match.Step{{Strategy: match.Selected}},
		Convert: converterOf(tmpl, lookup, get, inputs)}
	result, found, err := match.Solve(problem)
	if err != nil {
		return nil, fmt.Errorf("matching %s: %w", tmpl.Name, err)
	}
	if !found {
		return nil, fmt.Errorf("matching %s: the selection pairs nothing — %s", tmpl.Name, result.Tried[0].Reason)
	}
	result.Step = 0 // a person's, not a policy step's
	return bakeMatch(tmpl, trigger, triggerItem, result, kinds, states, inputs, lookup, get)
}

// reverseMatch builds the correcting event for one match: the same
// structure and trigger, every allocation negated with the open amounts as
// they are now before and restored after, every difference with its
// direction flipped. The rules that explained the match explain its
// reversal the same way: open amounts move by the amendment, a settlement
// for the negative amount records the undoing, the flipped difference
// posts the opposite entry.
func reverseMatch(m core.MatchedObjects, eventID int64, reason string, get core.Getter) (*core.MatchedObjects, error) {
	out := &core.MatchedObjects{Matching: m.Matching, Strategy: core.StrategyReversal, Step: 0, Trigger: m.Trigger,
		Kinds: m.Kinds, Sides: map[string][]core.MatchAllocation{}, Differences: []core.MatchDifference{},
		ReversalOf: eventID, Reason: reason, Inputs: map[string]any{"reversal_of": eventID}}
	negate := func(v any, kind string) (any, error) {
		r, err := amountOf(v, kind)
		if err != nil {
			return nil, err
		}
		return encode(r.Neg(r), kind)
	}
	for name, allocs := range m.Sides {
		for _, a := range allocs {
			state, ok, err := get(a.ObjectID)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("%s no longer exists", a.ObjectID)
			}
			r := core.MatchAllocation{ObjectID: a.ObjectID, ObjectType: a.ObjectType, Currency: a.Currency, Date: a.Date, Reference: a.Reference,
				Allocated: map[string]any{}, OpenBefore: map[string]any{}, OpenAfter: map[string]any{}}
			for measure, v := range a.Allocated {
				kind := m.Kinds[measure]
				if r.Allocated[measure], err = negate(v, kind); err != nil {
					return nil, err
				}
				// The open amount as it is now — later matches may have moved
				// it — and what it becomes with this match's share given back.
				var now *big.Rat
				field := openFieldOf(a, m, name, measure)
				if field == "" {
					return nil, fmt.Errorf("%s: no open field for %s", a.ObjectID, measure)
				}
				if state[field] == nil {
					now = new(big.Rat)
				} else if now, err = amountOf(state[field], kind); err != nil {
					return nil, fmt.Errorf("%s: %w", a.ObjectID, err)
				}
				taken, err := amountOf(v, kind)
				if err != nil {
					return nil, err
				}
				if r.OpenBefore[measure], err = encode(now, kind); err != nil {
					return nil, err
				}
				if r.OpenAfter[measure], err = encode(new(big.Rat).Add(now, taken), kind); err != nil {
					return nil, err
				}
			}
			if a.Equivalent != nil {
				r.Equivalent, r.EquivalentCurrency = map[string]any{}, a.EquivalentCurrency
				for measure, v := range a.Equivalent {
					if r.Equivalent[measure], err = negate(v, m.Kinds[measure]); err != nil {
						return nil, err
					}
				}
			}
			out.Sides[name] = append(out.Sides[name], r)
		}
	}
	flip := map[string]string{"short": "over", "over": "short", "up": "down", "down": "up"}
	for _, d := range m.Differences {
		d.Direction = flip[d.Direction]
		out.Differences = append(out.Differences, d)
	}
	return out, nil
}

// openFieldOf names the field an allocation's open amount lives in: the
// match event does not carry the structure, so the caller records it in
// Inputs as "open.<side>.<measure>" when it bakes the match; a match from
// before that convention falls back to the one field named "open".
func openFieldOf(a core.MatchAllocation, m core.MatchedObjects, side, measure string) string {
	if f, ok := m.Inputs["open."+side+"."+measure].(string); ok {
		return f
	}
	return "open"
}

// measuresOf lists the structure's measures for the engine.
func measuresOf(tmpl *core.MatchingTemplate, kinds map[string]string) []match.Measure {
	var measures []match.Measure
	for _, name := range tmpl.AllocatedMeasures() {
		measures = append(measures, match.Measure{Name: name, Allocated: true, Money: kinds[name] == "money", Places: places(kinds[name])})
	}
	for _, name := range tmpl.ComparedMeasures() {
		measures = append(measures, match.Measure{Name: name, Money: kinds[name] == "money", Places: places(kinds[name])})
	}
	return measures
}

// policySteps turns the declared policy into the engine's steps, recording
// every tolerance as an input.
func policySteps(tmpl *core.MatchingTemplate, inputs map[string]any) []match.Step {
	var policy []match.Step
	for i, st := range tmpl.Policy {
		step := match.Step{Strategy: st.Strategy, DifferenceAccount: st.DifferenceAccount,
			WithinDays: st.WithinDays, MaxItems: st.MaxItems, Partial: st.Partial}
		if len(st.Tolerance) > 0 {
			step.Tolerance = map[string]*big.Rat{}
			for m, v := range st.Tolerance {
				step.Tolerance[m], _ = core.ParseDecimal(v)
				inputs[fmt.Sprintf("step %d tolerance.%s", i+1, m)] = v
			}
		}
		if len(st.TolerancePercent) > 0 {
			step.TolerancePercent = map[string]*big.Rat{}
			for m, v := range st.TolerancePercent {
				step.TolerancePercent[m], _ = core.ParseDecimal(v)
				inputs[fmt.Sprintf("step %d tolerance_percent.%s", i+1, m)] = v
			}
		}
		policy = append(policy, step)
	}
	return policy
}

// converterOf builds the engine's converter from the structure's clause:
// rates are master data, as postings' convert reads them — the fx_rate for
// (from, to, date) by its code, or the inverse pair inverted; a missing
// rate refuses the firing into the worklist (unexplainable beats guessed).
// Every rate read is an input. Nil without a clause.
func converterOf(tmpl *core.MatchingTemplate, lookup core.Lookup, get core.Getter, inputs map[string]any) *match.Converter {
	c := tmpl.Convert
	if c == nil {
		return nil
	}
	memo := map[string]*big.Rat{}
	return &match.Converter{Functional: c.Functional, DateSide: c.DateSide,
		RateUpAccount: c.RateUpAccount, RateDownAccount: c.RateDownAccount,
		Round: func(r *big.Rat) *big.Rat { return core.RoundDecimal(r, 2, c.Rounding) },
		Rate: func(from, to, date string) (*big.Rat, error) {
			if from == to {
				return big.NewRat(1, 1), nil
			}
			key := from + "/" + to + "/" + date
			if r, ok := memo[key]; ok {
				return r, nil
			}
			read := func(code string) (*big.Rat, bool, error) {
				ids, err := lookup(FXRateObjectType, "code", code)
				if err != nil || len(ids) != 1 {
					return nil, false, err
				}
				state, ok, err := get(ids[0])
				if err != nil || !ok {
					return nil, false, err
				}
				text, _ := state["rate"].(string)
				r, err := core.ParseDecimal(text)
				if err != nil || r.Sign() <= 0 {
					return nil, false, fmt.Errorf("fx_rate %s: %q is not a positive decimal", code, text)
				}
				inputs["rate "+code] = text
				return r, true, nil
			}
			r, ok, err := read(key)
			if err != nil {
				return nil, err
			}
			if !ok {
				inv, ok, err := read(to + "/" + from + "/" + date)
				if err != nil {
					return nil, err
				}
				if !ok {
					return nil, fmt.Errorf("no fx_rate for %s/%s on %s", from, to, date)
				}
				r = new(big.Rat).Inv(inv)
			}
			memo[key] = r
			return r, nil
		}}
}

// bakeMatch encodes the engine's result as the event the log will hold:
// every allocation in its measure's boundary encoding, every difference with
// its home resolved (a declared account must exist), every input.
func bakeMatch(tmpl *core.MatchingTemplate, trigger core.MatchTrigger, triggerItem match.Item, result match.Result, kinds map[string]string,
	states map[string]map[string]any, inputs map[string]any, lookup core.Lookup, get core.Getter) (*core.MatchedObjects, error) {
	out := &core.MatchedObjects{Matching: tmpl.Name, Strategy: result.Strategy, Step: result.Step,
		Trigger: trigger, Kinds: kinds, Sides: map[string][]core.MatchAllocation{}, Differences: []core.MatchDifference{},
		Tried: attempts(result.Tried), Inputs: inputs}
	typeOf := map[string]string{}
	for _, s := range tmpl.Sides {
		typeOf[s.Name] = s.Type
		for measure, field := range s.Open {
			inputs["open."+s.Name+"."+measure] = field
		}
	}
	sideNames := make([]string, 0, len(result.Sides))
	for name := range result.Sides {
		sideNames = append(sideNames, name)
	}
	sort.Strings(sideNames)
	for _, name := range sideNames {
		side, _ := tmpl.Side(name)
		for _, a := range result.Sides[name] {
			st := states[a.ID]
			if st == nil {
				var err error
				st, _, err = get(a.ID)
				if err != nil {
					return nil, err
				}
			}
			alloc := core.MatchAllocation{ObjectID: a.ID, ObjectType: typeOf[name],
				Allocated: map[string]any{}, OpenBefore: map[string]any{}, OpenAfter: map[string]any{}}
			var err error
			for m := range a.Allocated {
				if alloc.Allocated[m], err = encode(a.Allocated[m], kinds[m]); err != nil {
					return nil, fmt.Errorf("matching %s: %s allocated %s: %w", tmpl.Name, a.ID, m, err)
				}
				if alloc.OpenBefore[m], err = encode(a.OpenBefore[m], kinds[m]); err != nil {
					return nil, err
				}
				if alloc.OpenAfter[m], err = encode(a.OpenAfter[m], kinds[m]); err != nil {
					return nil, err
				}
			}
			if side.Currency != "" {
				alloc.Currency, _ = st[side.Currency].(string)
			}
			if a.Equivalent != nil {
				alloc.Equivalent, alloc.EquivalentCurrency = map[string]any{}, triggerItem.Currency
				for m, v := range a.Equivalent {
					if alloc.Equivalent[m], err = encode(v, kinds[m]); err != nil {
						return nil, err
					}
				}
			}
			alloc.Date, _ = st[side.Date].(string)
			if side.Reference != "" {
				alloc.Reference, _ = st[side.Reference].(string)
			}
			out.Sides[name] = append(out.Sides[name], alloc)
		}
	}
	for _, d := range result.Differences {
		amount, err := encode(d.Amount, kinds[d.Measure])
		if err != nil {
			return nil, fmt.Errorf("matching %s: difference on %s: %w", tmpl.Name, d.Measure, err)
		}
		md := core.MatchDifference{Kind: d.Kind, Measure: d.Measure, Side: d.Side, Item: d.Item,
			Amount: amount, Direction: d.Direction, Account: d.Account}
		if kinds[d.Measure] == "money" {
			md.Currency = triggerItem.Currency
			if d.Kind == "fx" {
				md.Currency = tmpl.Convert.Functional
			}
		}
		if d.Account != "" {
			// The account the difference goes to must exist: a declared
			// home, resolved now so the posting rule that reads the code
			// cannot find it missing later.
			if _, err := core.ResolveRef(lookup, "account", "code", d.Account); err != nil {
				return nil, fmt.Errorf("matching %s: difference_account %s: %w", tmpl.Name, d.Account, err)
			}
		}
		out.Differences = append(out.Differences, md)
	}
	return out, nil
}

// isOpen reports whether an item still has something to allocate on every
// measure — the candidates the engine searches.
func isOpen(it match.Item) bool {
	for _, v := range it.Open {
		if v == nil || v.Sign() <= 0 {
			return false
		}
	}
	return true
}

func attempts(tried []match.Attempt) []core.MatchAttempt {
	out := make([]core.MatchAttempt, 0, len(tried))
	for _, a := range tried {
		out = append(out, core.MatchAttempt{Step: a.Step, Strategy: a.Strategy, Found: a.Found, Reason: a.Reason})
	}
	return out
}

// itemOf reads one object's state into an engine item by the side's
// declared fields and the measures' kinds.
func itemOf(side core.MatchingSide, id string, state map[string]any, kinds map[string]string, seq int) (match.Item, error) {
	it := match.Item{ID: id, Open: map[string]*big.Rat{}, Seq: seq}
	for m, field := range side.Open {
		if state[field] == nil {
			// An object born under an older version of its type, or by a
			// rule that fills no open amount, has nothing open: never a
			// candidate, never an error.
			continue
		}
		v, err := amountOf(state[field], kinds[m])
		if err != nil {
			return it, fmt.Errorf("field %s: %w", field, err)
		}
		it.Open[m] = v
	}
	for m, field := range side.Compare {
		if _, present := state[field]; !present {
			continue
		}
		v, err := amountOf(state[field], kinds[m])
		if err != nil {
			return it, fmt.Errorf("field %s: %w", field, err)
		}
		if it.Compare == nil {
			it.Compare = map[string]*big.Rat{}
		}
		it.Compare[m] = v
	}
	if side.Currency != "" {
		it.Currency, _ = state[side.Currency].(string)
	}
	it.Date, _ = state[side.Date].(string)
	if it.Date == "" {
		return it, fmt.Errorf("field %s: no date", side.Date)
	}
	if side.Reference != "" {
		it.Reference, _ = state[side.Reference].(string)
	}
	return it, nil
}

// amountOf reads a stored amount by its field kind: money as minor units
// (a number in state, a decimal string at the boundary), int as a whole
// number, decimal as its string.
func amountOf(v any, kind string) (*big.Rat, error) {
	switch n := v.(type) {
	case nil:
		return nil, fmt.Errorf("is missing")
	case string:
		if kind == "money" {
			minor, err := core.ParseMoney(n)
			if err != nil {
				return nil, err
			}
			return big.NewRat(minor, 100), nil
		}
		return core.ParseDecimal(n)
	case float64:
		if n != float64(int64(n)) {
			return nil, fmt.Errorf("%v is not a whole number", n)
		}
		return scaled(int64(n), kind), nil
	case int64:
		return scaled(n, kind), nil
	case int:
		return scaled(int64(n), kind), nil
	}
	return nil, fmt.Errorf("%v (%T) is not an amount", v, v)
}

func scaled(n int64, kind string) *big.Rat {
	if kind == "money" {
		return big.NewRat(n, 100)
	}
	return new(big.Rat).SetInt64(n)
}

// encode renders an exact amount in the boundary encoding of its kind.
func encode(r *big.Rat, kind string) (any, error) {
	switch kind {
	case "money":
		minor := new(big.Rat).Mul(r, big.NewRat(100, 1))
		if !minor.IsInt() || !minor.Num().IsInt64() {
			return nil, fmt.Errorf("%s does not fit a money field", r.RatString())
		}
		return minor.Num().Int64(), nil
	case "int":
		if !r.IsInt() || !r.Num().IsInt64() {
			return nil, fmt.Errorf("%s is not a whole number that fits", r.RatString())
		}
		return r.Num().Int64(), nil
	}
	return core.CanonicalDecimal(r)
}

func places(kind string) int {
	switch kind {
	case "money":
		return 2
	case "int":
		return 0
	}
	return -1
}
