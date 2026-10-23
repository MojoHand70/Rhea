package shell_test

import (
	"fmt"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/store"
)

// A company's own reading becomes Rhea's when many companies hold it, by
// Rhea's staff alone (KK, 2026-10-22): the same rule held by at least ten
// companies kept by at least three offices or installations, counted by Rhea
// from the rules alone, optionally with an external biegły rewident's
// opinion. Below the floor staff see nothing and the door refuses; an office
// let into four of the companies is not staff. Promoted, the reading books
// for a company that never had it.
func TestARuleManyCompaniesHoldBecomesRheasByStaff(t *testing.T) {
	w := newGrantsWorld(t)
	byWhole := core.RuleSpec{Match: core.Match{EventType: "fuel.bought"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text", "amount": "=$.amount"}}}}
	byVendor := core.RuleSpec{Match: core.Match{EventType: "fuel.bought"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.vendor", "amount": "=$.amount"}}}}
	if _, err := w.s.InsertRuleVersion(w.ctx, core.Rule{ID: "fuel", Status: core.StatusActive, Priority: 100, EffectiveFrom: "2026-01-01",
		CreatedBy: "test", Description: "a fuel bill is a note", Spec: byWhole}); err != nil {
		t.Fatal(err)
	}
	company := func(i int) string { return fmt.Sprintf("PL90000000%02d", i) }
	own := func(i int, spec core.RuleSpec) {
		t.Helper()
		if _, err := w.s.InsertRuleVersion(w.ctx, core.Rule{ID: "fuel@" + company(i), Company: company(i), Replaces: "fuel",
			Status: core.StatusActive, Priority: 100, EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: "fuel by vendor", Spec: spec}); err != nil {
			t.Fatal(err)
		}
	}
	// Two offices keep four companies each; companies 9 and 10 keep their own.
	o1, o2 := "PL7010001234", "PL7020004321"
	for i := 1; i <= 8; i++ {
		office := o1
		if i > 4 {
			office = o2
		}
		w.mustDoor("", company(i), "grant_mandate", map[string]any{"office": office, "client": company(i), "scope": "door:* view:* type:*", "from": "2026-01-01", "granted_by": "operator"})
	}
	w.person("Anna Nowak", "anna", o1, "accountant")
	kk := w.person("Krzysztof Kowal", "kk", "", "owner")
	w.mustDoor("", core.House, "grant_membership", map[string]any{"person": kk, "company": core.House, "role": "owner", "from": "2026-01-01", "granted_by": "operator"})
	for i := 1; i <= 9; i++ {
		own(i, byVendor)
	}
	own(11, byWhole) // a reading of its own that differs — not counted

	type candidate struct {
		Shape     string `json:"shape"`
		Replaces  string `json:"replaces"`
		Companies int    `json:"companies"`
		Keepers   int    `json:"keepers"`
	}
	var listed []candidate
	promote := func(person, acting, shape string) int {
		t.Helper()
		code, _ := w.door(person, acting, "promote_rule", map[string]any{"shape": shape, "opinion": "Opinia biegłego rewidenta 7/2026",
			"approved_by": person, "occurred_at": w.today})
		return code
	}

	// Nine companies: below the floor, nothing to see and nothing to promote.
	if code := w.call("kk", core.House, "GET", "/api/promotions", nil, &listed); code != 200 || len(listed) != 0 {
		t.Fatalf("below the floor: %d %+v", code, listed)
	}
	own(10, byVendor)
	if code := w.call("anna", company(1), "GET", "/api/promotions", nil, &listed); code != 403 {
		t.Fatalf("an office listing what companies hold: %d", code)
	}
	if code := w.call("kk", core.House, "GET", "/api/promotions", nil, &listed); code != 200 || len(listed) != 1 ||
		listed[0].Companies != 10 || listed[0].Keepers != 4 || listed[0].Replaces != "fuel" {
		t.Fatalf("ten companies under four keepers: %d %+v", code, listed)
	}
	shape := listed[0].Shape
	if code := promote("anna", company(1), shape); code != 403 {
		t.Fatalf("an office promoting: %d", code)
	}
	if code := promote("kk", core.House, shape); code != 200 {
		t.Fatalf("staff promoting: %d", code)
	}
	r, err := w.s.GetRule(w.ctx, "fuel")
	if err != nil || r.Company != "" || r.Version != 2 || r.Status != core.StatusActive || r.Spec.Effect.Object.Fields["text"] != "=$.vendor" ||
		r.Warrant == nil || r.Warrant.Basis != "network" || r.Warrant.Support == nil || r.Warrant.Support.Count != 10 ||
		!strings.Contains(strings.Join(r.Warrant.Citations, " "), "Opinia biegłego rewidenta 7/2026") {
		t.Fatalf("Rhea's rule after the promotion: %+v %v", r, err)
	}
	if evs, _ := w.s.EventsOfTypes(w.ctx, []string{"rule.promoted"}); len(evs) != 1 || evs[0].Company != "" {
		t.Fatalf("the promotion is the installation's fact: %+v", evs)
	}
	// A company that never had the reading books by it now.
	w.mustDoor("", company(12), "submit_event", map[string]any{"event_type": "fuel.bought", "occurred_at": w.today,
		"payload": map[string]any{"text": "a bill", "vendor": "Orlen", "amount": "200.00"}})
	if ids, _ := w.s.FindObjectIDsByField(store.WithCompany(w.ctx, company(12)), "note", "text", "Orlen"); len(ids) != 1 {
		t.Fatalf("company 12's fuel bill by Rhea's new reading: %v", ids)
	}
}
