package shell_test

import (
	"fmt"
	"testing"

	"rhea/internal/core"
)

// Dziś asks each person only what they may decide (KK, 2026-10-27): the
// records waiting for a move their type gives to a group, listed for a
// person in that group — the door and its button with each — and for
// nobody else: not the accountant outside it, not the one who proposed it
// while another could, not a person of another company.
func TestAPersonIsAskedWhatTheyMayDecide(t *testing.T) {
	w := newGrantsWorld(t)
	a, b := "PL5250001111", "PL5260001246"
	w.person("Krzysztof Kowal", "kk", a, "company_manager")
	w.person("Olga Nowa", "olga", b, "company_manager")
	anna := w.person("Anna Nowak", "anna", a, "accountant")
	ewa := w.person("Ewa Bank", "ewa", a, "accountant")
	w.mustDoor("kk", a, "declare_group", map[string]any{"name": "Bankowość", "company": a})
	for _, id := range []string{anna, ewa} {
		w.mustDoor("kk", a, "join_group", map[string]any{"group": "Bankowość", "person": id, "company": a, "from": "2026-01-01"})
	}
	if err := w.s.InsertObjectType(w.ctx, core.ObjectType{Name: "account", Version: 1, Domain: "work", LabelField: "number", Key: "number",
		Fields: []core.FieldDef{{Name: "number", Type: "string", Required: true, Format: "iban"}, {Name: "status", Type: "enum", Required: true, Values: []string{"proposed", "active", "rejected"}},
			{Name: "proposed_by", Type: "string"}},
		Lifecycle: &core.LifecycleDef{Field: "status", Initial: "proposed", Transitions: map[string][]string{"proposed": {"active", "rejected"}},
			Moves: map[string]core.Move{"proposed>active": {Group: "Bankowość", NotBy: "proposed_by", Door: "approve_account", Label: "Zatwierdź"},
				"proposed>rejected": {Group: "Bankowość", Door: "reject_account", Label: "Odrzuć"}}}}); err != nil {
		t.Fatal(err)
	}
	for _, d := range []core.Activity{
		{Name: "propose_account", Spec: core.ActivitySpec{Emits: "account.proposed", Inputs: []core.FieldDef{{Name: "number", Type: "string", Required: true, Format: "iban"}, {Name: "proposed_by", Type: "string", Required: true}}}},
		{Name: "approve_account", Spec: core.ActivitySpec{Emits: "account.approved", Inputs: []core.FieldDef{{Name: "account", Type: "ref<account>", Required: true}}}},
		{Name: "reject_account", Spec: core.ActivitySpec{Emits: "account.rejected", Inputs: []core.FieldDef{{Name: "account", Type: "ref<account>", Required: true}}}},
	} {
		d.Status, d.Domain, d.Description, d.Spec.Who = core.StatusActive, "work", d.Name, []string{"human"}
		if _, err := w.s.InsertActivityVersion(w.ctx, d); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := w.s.InsertRuleVersion(w.ctx, core.Rule{ID: "propose-account", Status: core.StatusActive, Priority: 100, EffectiveFrom: "2026-01-01", CreatedBy: "test",
		Description: "an account proposed", Spec: core.RuleSpec{Match: core.Match{EventType: "account.proposed"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "account", Fields: map[string]string{"number": "=$.number", "status": "proposed", "proposed_by": "=$.proposed_by"}}}}}); err != nil {
		t.Fatal(err)
	}
	w.mustDoor("anna", a, "propose_account", map[string]any{"number": "PL61109010140000071219812874"})
	asked := func(person, company string) string {
		t.Helper()
		var out []struct {
			Object, Door, Label, Input, To string
			Movers                         []string
		}
		if code := w.call(person, company, "GET", "/api/moves", nil, &out); code != 200 {
			t.Fatalf("%s: %d", person, code)
		}
		s := ""
		for _, m := range out {
			s += fmt.Sprintf("%s %s>%s %s(%s) %v; ", m.Object[:8], "proposed", m.To, m.Door, m.Label, m.Movers)
			if m.Input != "account" {
				t.Fatalf("the door's input naming the record: %q", m.Input)
			}
		}
		return s
	}
	if got := asked("ewa", a); got != "account- proposed>active approve_account(Zatwierdź) [anna ewa]; account- proposed>rejected reject_account(Odrzuć) [anna ewa]; " {
		t.Fatalf("ewa is asked: %s", got)
	}
	if got := asked("anna", a); got != "account- proposed>rejected reject_account(Odrzuć) [anna ewa]; " {
		t.Fatalf("anna, who proposed it, is asked only to reject: %s", got)
	}
	for _, who := range [][2]string{{"kk", a}, {"olga", b}, {"olga", ""}} {
		if got := asked(who[0], who[1]); got != "" {
			t.Fatalf("%s at %q is asked: %s", who[0], who[1], got)
		}
	}
}
