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
	inputs := map[string]any{}
	var sides []match.Side
	for _, side := range tmpl.Sides {
		if side.Name == tmpl.Trigger {
			continue
		}
		t, err := core.ParseTemplate(side.Items)
		if err != nil {
			return nil, nil, fmt.Errorf("matching %s side %s: %w", tmpl.Name, side.Name, err)
		}
		reads, in, err := t.Objects(payload, env)
		if err != nil {
			// A side that cannot be read — a payer no company carries the
			// account of, a key above the collection cap — has no
			// candidates: the trigger stays open for the person, and the
			// log says why. Its own materialization is never held hostage.
			return nil, &core.UnmatchedObjects{Matching: tmpl.Name, Trigger: trigger, Inputs: inputs,
				Reason: fmt.Sprintf("%s cannot be read: %v", side.Name, err)}, nil
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
				if fmt.Sprint(rd.State[cf]) != fmt.Sprint(state[tf]) {
					joined = false
				}
			}
			if !joined {
				continue
			}
			it, err := itemOf(side, rd.ID, rd.State, kinds, i+1)
			if err != nil {
				return nil, nil, fmt.Errorf("matching %s: the %s %s: %w", tmpl.Name, side.Name, rd.ID, err)
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
		// never to a long search: the trigger stays open, the log says why.
		capacity := side.Cap
		if capacity <= 0 {
			capacity = core.DefaultMatchingCap
		}
		if open > capacity {
			return nil, &core.UnmatchedObjects{Matching: tmpl.Name, Trigger: trigger, Inputs: inputs,
				Reason: fmt.Sprintf("%s holds %d open candidates, above the cap of %d — a person decides", side.Name, open, capacity)}, nil
		}
		sides = append(sides, s)
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
		if !minor.IsInt() {
			return nil, fmt.Errorf("%s does not fit a money field", r.RatString())
		}
		return minor.Num().Int64(), nil
	case "int":
		if !r.IsInt() {
			return nil, fmt.Errorf("%s is not a whole number", r.RatString())
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
