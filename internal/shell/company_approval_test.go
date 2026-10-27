package shell_test

import (
	"testing"

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/shell"
	"rhea/internal/store"
)

// A company's own rule is approved inside the company; anything of Rhea's —
// a rule for every company, a new version of a type every company books — by
// Rhea's staff alone (KK, 2026-10-22). Staff are the house's people: a
// membership in the house, granted by the installation's owner and never
// claimed by a signed-in person. A right in every company is no longer
// enough: an office let into all its clients is still not Rhea.
func TestCompanyRulesAreTheCompanysAndRheasAreStaffs(t *testing.T) {
	w := newGrantsWorld(t)
	a, b := alfaKey, betaKey
	kk := w.person("Krzysztof Kowal", "kk", a, "owner")
	w.person("Olga Nowa", "olga", b, "owner")
	w.person("Ewa Reader", "ewa", a, "reader")
	draft := func(id, company string) {
		t.Helper()
		if _, err := w.s.InsertRuleVersion(w.ctx, core.Rule{ID: id, Company: company, Status: core.StatusDraft, Priority: 100,
			EffectiveFrom: "2026-01-01", CreatedBy: "agent:test", Description: id, Spec: core.RuleSpec{Match: core.Match{EventType: "memo.taken"},
				Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text", "amount": "=$.amount"}}}}}); err != nil {
			t.Fatal(err)
		}
	}
	approve := func(person, company, id string) int {
		t.Helper()
		return w.call(person, company, "POST", "/api/rules/"+id+"/approve", map[string]any{}, nil)
	}

	// a's own rule: a's owner approves it acting for a; a's reader does not,
	// nor b's owner, from b or from a.
	draft("memo@"+a, a)
	if code := approve("ewa", a, "memo@"+a); code != 403 {
		t.Fatalf("a's reader approving a's rule: %d", code)
	}
	if code := approve("olga", a, "memo@"+a); code != 403 {
		t.Fatalf("b's owner approving a's rule, acting for a: %d", code)
	}
	if code := approve("olga", b, "memo@"+a); code == 200 {
		t.Fatalf("b's owner approving a's rule, acting for b: %d", code)
	}
	if code := approve("kk", a, "memo@"+a); code != 200 {
		t.Fatalf("a's owner approving a's rule: %d", code)
	}
	if r, err := w.s.GetRule(store.WithCompany(w.ctx, a), "memo@"+a); err != nil || r.Status != core.StatusActive {
		t.Fatalf("a's rule after a's approval: %+v %v", r, err)
	}

	// Rhea's rule: not a's owner, and not one let in at both companies.
	draft("memo", "")
	if code := approve("kk", a, "memo"); code != 403 {
		t.Fatalf("a's owner approving Rhea's rule: %d", code)
	}
	w.mustDoor("olga", b, "grant_membership", map[string]any{"person": kk, "company": b, "role": "owner", "from": "2026-01-01", "granted_by": "olga"})
	if code := approve("kk", a, "memo"); code != 403 {
		t.Fatalf("an owner of both companies approving Rhea's rule: %d", code)
	}
	// No signed-in person claims the house, even while nobody holds it.
	if code, out := w.door("olga", core.House, "grant_membership", map[string]any{"person": w.personID("olga"), "company": core.House,
		"role": "owner", "from": "2026-01-01", "granted_by": "olga"}); code != 403 {
		t.Fatalf("olga claiming the house: %d %v", code, out)
	}
	// The installation's owner makes kk staff; kk approves Rhea's rule acting for the house.
	w.mustDoor("", core.House, "grant_membership", map[string]any{"person": kk, "company": core.House, "role": "owner", "from": "2026-01-01", "granted_by": "operator"})
	if code := approve("kk", core.House, "memo"); code != 200 {
		t.Fatalf("staff approving Rhea's rule: %d", code)
	}
	if r, _ := w.s.GetRule(w.ctx, "memo"); r.Status != core.StatusActive {
		t.Fatalf("Rhea's rule after staff's approval: %+v", r)
	}

	// A bundle drafted for a that changes a type every company books is
	// staff's, drafted for a or not; one of a's rules and a word of a's own
	// (a type nobody had) is a's. Adam owns a and is not staff.
	w.person("Adam Nowak", "adam", a, "owner")
	spec := core.RuleSpec{Match: core.Match{EventType: "fuel.bought"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text", "amount": "=$.amount"}}}}
	warrant := &core.Warrant{Basis: "client", Citations: []string{"the owner said so"}}
	widened := core.ObjectType{Name: "note", Version: 1, Domain: "work", LabelField: "text", Amendable: []string{"text", "amount"},
		Fields: []core.FieldDef{{Name: "text", Type: "string", Required: true}, {Name: "amount", Type: "money", Required: true}, {Name: "car", Type: "string"}}}
	changes, err := shell.StoreBundleDraft(store.WithCompany(w.ctx, a), w.s, nil, agent.BundleDraft{BundleID: "fuel", Description: "fuel",
		ObjectTypes: []core.ObjectType{widened}, Rules: []agent.Draft{{RuleID: "fuel", Description: "fuel", Priority: 100, Spec: spec}}, Warrant: warrant},
		"agent:test", "2026-01-01", "Fuel?", "")
	if err != nil {
		t.Fatal(err)
	}
	if code := w.call("adam", a, "POST", "/api/bundles/"+changes.ID+"/approve", map[string]any{}, nil); code != 403 {
		t.Fatalf("a's owner changing a type every company books: %d", code)
	}
	ownWord := core.ObjectType{Name: "fuel_note", Version: 1, Domain: "work", LabelField: "text",
		Fields: []core.FieldDef{{Name: "text", Type: "string", Required: true}}}
	spec.Match.EventType, spec.Effect.Object = "fuel.noted", core.ObjectTemplate{Type: "fuel_note", Fields: map[string]string{"text": "=$.text"}}
	own, err := shell.StoreBundleDraft(store.WithCompany(w.ctx, a), w.s, nil, agent.BundleDraft{BundleID: "fuel-notes", Description: "fuel notes",
		ObjectTypes: []core.ObjectType{ownWord}, Rules: []agent.Draft{{RuleID: "fuel-note", Description: "fuel notes", Priority: 100, Spec: spec}}, Warrant: warrant},
		"agent:test", "2026-01-01", "Fuel notes?", "")
	if err != nil {
		t.Fatal(err)
	}
	if code := w.call("adam", a, "POST", "/api/bundles/"+own.ID+"/approve", map[string]any{}, nil); code != 200 {
		t.Fatalf("a's owner approving a's own rule and word: %d", code)
	}
}

// A company's own new door acts in the company alone (review 2026-10-27,
// critical 3): a new door was the company's word whatever it did, so Anna,
// an accountant — whose right approves the company's own bundles — approved
// one that wrote the installation's facts and booked herself a membership
// in another company through Rhea's own rule, live for every company. A
// door that writes the installation's facts, reaches every company's books,
// reads through a channel, or sends a fact another door already declares
// is no company's word: Rhea's staff approve it.
func TestACompanysNewDoorActsInTheCompanyAlone(t *testing.T) {
	w := newGrantsWorld(t)
	a, b := alfaKey, betaKey
	w.person("Krzysztof Kowal", "kk", a, "owner")
	w.person("Olga Nowa", "olga", b, "owner")
	w.person("Anna Nowak", "anna", a, "accountant")
	warrant := &core.Warrant{Basis: "client", Citations: []string{"the accountant said so"}}
	inputs := []core.FieldDef{{Name: "person", Type: "ref<person>", Required: true}, {Name: "company", Type: "string", Required: true},
		{Name: "role", Type: "string", Required: true}, {Name: "from", Type: "date", Required: true}, {Name: "granted_by", Type: "string", Required: true}}
	approve := func(name string, spec core.ActivitySpec) int {
		t.Helper()
		bundle, err := shell.StoreBundleDraft(store.WithCompany(w.ctx, a), w.s, nil, agent.BundleDraft{BundleID: name, Description: name,
			Activities: []agent.ActivityDraft{{Name: name, Domain: "work", Description: name, Spec: spec}}, Warrant: warrant},
			"agent:test", "2026-01-01", name+"?", "")
		if err != nil {
			t.Fatal(err)
		}
		return w.call("anna", a, "POST", "/api/bundles/"+bundle.ID+"/approve", map[string]any{}, nil)
	}
	for name, spec := range map[string]core.ActivitySpec{
		"claim_anywhere": {Inputs: inputs, Emits: "membership.granted", Who: []string{"human"}, Installation: true},
		"claim_here":     {Inputs: inputs, Emits: "membership.granted", Who: []string{"human"}},
		"note_for_all":   {Inputs: []core.FieldDef{{Name: "text", Type: "string", Required: true}}, Emits: "memo.left", Who: []string{"human"}, Everywhere: true},
	} {
		if code := approve(name, spec); code != 403 {
			t.Fatalf("anna approving %s: %d", name, code)
		}
		if _, err := w.s.GetActiveActivity(w.ctx, name); err == nil {
			t.Fatalf("%s went live", name)
		}
	}
	// a door of a's own, sending a fact of its own, is still a's word
	if code := approve("log_visit", core.ActivitySpec{Inputs: []core.FieldDef{{Name: "who", Type: "string", Required: true}},
		Emits: "visit.logged", Who: []string{"human"}}); code != 200 {
		t.Fatalf("anna approving a's own door: %d", code)
	}
}
