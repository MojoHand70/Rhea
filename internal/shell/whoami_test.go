package shell_test

import (
	"testing"

	"rhea/internal/core"
)

// TestWhoAmINamesTheCompanies: the desk's switch (DIRECTION 2026-10-22, "The
// desk in order", step 2) lists the companies a person may act for by their
// names, whatever company they act for now — acting for one client the
// person sees that client alone, so the face could name no other. Each name
// is the company's own record in its own books (the base vocabulary's
// company of kind "self"), never a customer's record inside them; and a
// person learns the names of their own companies, nobody else's.
func TestWhoAmINamesTheCompanies(t *testing.T) {
	w := newGrantsWorld(t)
	if err := w.s.InsertObjectType(w.ctx, core.ObjectType{Name: "company", Version: 1, Domain: "work", LabelField: "name",
		Fields: []core.FieldDef{{Name: "name", Type: "string", Required: true}, {Name: "kind", Type: "enum", Values: []string{"customer", "supplier", "self"}, Required: true}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := w.s.InsertRuleVersion(w.ctx, core.Rule{ID: "company", Status: core.StatusActive, Priority: 100, EffectiveFrom: "2026-01-01",
		CreatedBy: "test", Description: "companies", Spec: core.RuleSpec{Match: core.Match{EventType: "company.registered"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "company", Fields: map[string]string{"name": "=$.name", "kind": "=$.kind"}}}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := w.s.InsertActivityVersion(w.ctx, core.Activity{Name: "register_company", Status: core.StatusActive, Domain: "work", Description: "a company",
		Spec: core.ActivitySpec{Inputs: []core.FieldDef{{Name: "name", Type: "string", Required: true}, {Name: "kind", Type: "string", Required: true}},
			Emits: "company.registered", Who: []string{"human"}}}); err != nil {
		t.Fatal(err)
	}
	alfa, beta := "PL5250001111", "PL5260001246"
	w.mustDoor("", alfa, "register_company", map[string]any{"name": "Alfa sp. z o.o.", "kind": "self"})
	w.mustDoor("", alfa, "register_company", map[string]any{"name": "Klient Alfy S.A.", "kind": "customer"})
	w.mustDoor("", beta, "register_company", map[string]any{"name": "Beta S.A.", "kind": "self"})
	w.person("Krzysztof Kowal", "kk", alfa, "owner")
	w.mustDoor("", beta, "grant_membership", map[string]any{"person": w.personID("kk"), "company": beta, "role": "owner", "from": "2026-01-01", "granted_by": "operator"})
	w.person("Anna Nowak", "anna", alfa, "accountant")

	var who struct {
		Companies []string          `json:"companies"`
		Company   string            `json:"company"`
		Names     map[string]string `json:"names"`
	}
	if code := w.call("kk", alfa, "GET", "/api/whoami", nil, &who); code != 200 || who.Company != alfa || len(who.Companies) != 2 {
		t.Fatalf("kk acting for alfa: %d %+v", code, who)
	}
	if who.Names[alfa] != "Alfa sp. z o.o." || who.Names[beta] != "Beta S.A." {
		t.Fatalf("acting for alfa, kk is told both his companies by name: %v", who.Names)
	}
	who.Names = nil
	if code := w.call("anna", alfa, "GET", "/api/whoami", nil, &who); code != 200 || who.Names[alfa] != "Alfa sp. z o.o." {
		t.Fatalf("anna, alfa's accountant: %d %v", code, who.Names)
	}
	if _, told := who.Names[beta]; told || len(who.Names) != 1 {
		t.Fatalf("anna learns no name of a company she does not hold: %v", who.Names)
	}
}
