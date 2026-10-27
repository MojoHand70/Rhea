package rights_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/rights"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

// Who may do what is computed from the log alone (DIRECTION 2026-10-19,
// "People and rights"; 2026-10-27, "Tasks, groups and the phone"):
// memberships give the system's roles, groups give what a business decides,
// an office reaches a client through a mandate within the client's scope, a
// deputy holds the substituted person's grants between two dates, and an
// ended grant gives nothing after its last day. Every grant names the
// object and the person that gave it. (Rewritten 2026-10-27, the law
// changed: the roles are the system's seven, and an accountant is a user in
// the group Accountant.)
func TestGrantsFromTheLog(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s, Today: func() string { return "2026-10-20" }}
	sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", "common", "people.json"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-09-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve: %v %v", err, errs)
	}
	// door fires a verb and returns the id of the object its event made:
	// ids are "<type>-<root event id>", reproducible by construction.
	door := func(ctx context.Context, name string, inputs map[string]any) string {
		t.Helper()
		tr, err := x.TriggerActivity(ctx, name, inputs, "shell:kk", "")
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if tr.Booked != 1 || len(tr.Errors) != 0 {
			t.Fatalf("%s: booked %d, errors %v", name, tr.Booked, tr.Errors)
		}
		typ := map[string]string{"grant_membership": "membership", "grant_mandate": "mandate", "appoint_substitute": "substitution", "declare_role": "role", "join_group": "group_member"}[name]
		return fmt.Sprintf("%s-%d", typ, tr.EventID)
	}
	office, client, other := "PL5250001111", "PL5260001246", "PL7010001234"
	door(ctx, "register_person", map[string]any{"name": "Krzysztof Kowal", "handle": "kk", "role": "owner"})
	door(ctx, "register_person", map[string]any{"name": "Anna Nowak", "handle": "anna", "role": "accountant"})
	door(ctx, "register_person", map[string]any{"name": "Piotr Lis", "handle": "piotr", "role": "clerk"})
	id := func(handle string) string {
		ids, err := s.FindObjectIDsByField(ctx, "person", "handle", handle)
		if err != nil || len(ids) != 1 {
			t.Fatalf("person %s: %v %v", handle, ids, err)
		}
		return ids[0]
	}
	// Persons are the installation's: one human, many workplaces.
	if o, _ := s.GetObject(ctx, id("anna")); o.Company != "" {
		t.Fatalf("a person is nobody's row: %q", o.Company)
	}

	// kk owns the office; anna keeps its books; the grant is the office's fact.
	at := func(company string) context.Context { return store.WithCompany(ctx, company) }
	door(at(office), "grant_membership", map[string]any{"person": id("kk"), "company": office, "role": "company_manager", "from": "2026-01-01", "granted_by": "kk"})
	annaMembership := door(at(office), "grant_membership", map[string]any{"person": id("anna"), "company": office, "role": "user", "from": "2026-10-01", "granted_by": "kk"})
	annaAccountant := door(at(office), "join_group", map[string]any{"group": "accountant", "person": id("anna"), "company": office, "from": "2026-10-01", "granted_by": "kk"})

	anna, err := rights.Resolve(ctx, s, "anna", "2026-10-20")
	if err != nil {
		t.Fatal(err)
	}
	if got := anna.Companies(); len(got) != 1 || got[0] != office {
		t.Fatalf("anna's companies: %v", got)
	}
	if g, ok := anna.May(office, "door", "close_period", "periods"); !ok || g.Via != "group" || g.Object != annaAccountant || g.GrantedBy != "kk" || g.Group != "accountant" {
		t.Fatalf("anna closes periods at the office as an accountant, by kk's grant: %+v %v", g, ok)
	}
	if _, ok := anna.May(office, "door", "grant_membership", "people"); ok {
		t.Fatal("an accountant never says who may do what")
	}
	kk, _ := rights.Resolve(ctx, s, "kk", "2026-10-20")
	if _, ok := kk.May(office, "door", "grant_membership", "people"); !ok {
		t.Fatal("the owner does")
	}
	// Before her first day she holds nothing.
	early, _ := rights.Resolve(ctx, s, "anna", "2026-09-30")
	if len(early.Grants) != 0 {
		t.Fatalf("anna before her first day: %+v", early.Grants)
	}

	// The office keeps the client's books within the scope the client
	// granted: anna reaches the client through the mandate, within it.
	mandate := door(at(client), "grant_mandate", map[string]any{"office": office, "client": client, "scope": "domain:finance view:*", "from": "2026-10-10", "granted_by": "kk"})
	anna, _ = rights.Resolve(ctx, s, "anna", "2026-10-20")
	if got := anna.Companies(); len(got) != 2 || got[1] != client {
		t.Fatalf("anna's companies with the mandate: %v", got)
	}
	if g, ok := anna.May(client, "door", "book_invoice", "finance"); !ok || g.Via != "mandate" || g.Object != mandate || g.Through != annaAccountant {
		t.Fatalf("anna books the client's invoices through the mandate: %+v %v", g, ok)
	}
	if _, ok := anna.May(client, "door", "close_period", "periods"); ok {
		t.Fatal("the mandate's scope stops at finance")
	}
	if _, ok := anna.May(client, "view", "balance", "finance"); !ok {
		t.Fatal("the scope grants every view")
	}
	if claimed, _ := rights.Claimed(ctx, s, client, "2026-10-20"); !claimed {
		t.Fatal("the mandate claims the client")
	}
	if claimed, _ := rights.Claimed(ctx, s, other, "2026-10-20"); claimed {
		t.Fatal("nobody holds the other company")
	}

	// A deputy holds anna's grants between two dates, in one company only.
	door(at(office), "appoint_substitute", map[string]any{"person": id("anna"), "deputy": id("piotr"), "company": office, "from": "2026-10-19", "to": "2026-10-21", "granted_by": "kk"})
	piotr, _ := rights.Resolve(ctx, s, "piotr", "2026-10-20")
	if g, ok := piotr.May(office, "door", "close_period", "periods"); !ok || g.Via != "substitution" || g.Through != annaAccountant {
		t.Fatalf("piotr stands in for anna at the office: %+v %v", g, ok)
	}
	if _, ok := piotr.May(client, "door", "book_invoice", "finance"); ok {
		t.Fatal("the substitution names the office, not the client")
	}
	if after, _ := rights.Resolve(ctx, s, "piotr", "2026-10-22"); len(after.Grants) != 0 {
		t.Fatalf("nothing after the substitution's last day: %+v", after.Grants)
	}

	// A membership granted until a day expires by itself after it, and the
	// groups it let him into with it.
	door(at(client), "grant_membership", map[string]any{"person": id("piotr"), "company": client, "role": "guest", "from": "2026-10-01", "to": "2026-10-19", "reason": "the audit", "granted_by": "kk"})
	door(at(client), "join_group", map[string]any{"group": "reader", "person": id("piotr"), "company": client, "from": "2026-10-01", "granted_by": "kk"})
	if p, _ := rights.Resolve(ctx, s, "piotr", "2026-10-19"); p.Grants[0].Reason != "the audit" || p.Grants[0].Role != "guest" {
		t.Fatalf("piotr on the audit's last day: %+v", p.Grants)
	} else if _, ok := p.May(client, "view", "balance", "finance"); !ok {
		t.Fatalf("piotr reads the client's books in the group Reader: %+v", p.Grants)
	}
	if p, _ := rights.Resolve(ctx, s, "piotr", "2026-10-20"); len(p.Companies()) != 1 || p.Companies()[0] != office {
		t.Fatalf("piotr after the audit — a group is no way in: %+v", p.Grants)
	}

	// Ending anna's membership ends everything that rested on it after its
	// last day; the day itself is still hers, and the days before keep
	// their answer. To revoke now, an owner names yesterday as the last day.
	door(at(office), "end_membership", map[string]any{"membership": annaMembership, "to": "2026-10-20", "ended_by": "kk"})
	if last, _ := rights.Resolve(ctx, s, "anna", "2026-10-20"); len(last.Grants) == 0 {
		t.Fatal("her last day is still hers")
	}
	gone, _ := rights.Resolve(ctx, s, "anna", "2026-10-21")
	if len(gone.Grants) != 0 {
		t.Fatalf("anna after her last day: %+v", gone.Grants)
	}
	if o, _ := s.GetObject(ctx, annaMembership); o.State["status"] != "ended" || o.State["to"] != "2026-10-20" {
		t.Fatalf("the ended membership keeps its history: %v", o.State)
	}
	// The table of the office: kk and anna (on her last day), each grant
	// opening to its object.
	table, err := rights.Table(ctx, s, office, "2026-10-21")
	if err != nil || len(table) != 1 || table[0].Handle != "kk" {
		t.Fatalf("the office's table after anna's last day: %+v %v", table, err)
	}
	if earlier, _ := rights.Table(ctx, s, office, "2026-10-20"); len(earlier) != 3 || earlier[0].Handle != "anna" || earlier[2].Handle != "piotr" || earlier[2].Grants[0].Via != "substitution" {
		t.Fatalf("the office's table on her last day: %+v", earlier)
	}

	// A person who left is no principal; a stranger neither.
	door(ctx, "set_person_status", map[string]any{"person": id("piotr"), "status": "left"})
	if _, err := rights.Resolve(ctx, s, "piotr", "2026-10-20"); err == nil || !strings.Contains(err.Error(), "left") {
		t.Fatalf("piotr left: %v", err)
	}
	if _, err := rights.Resolve(ctx, s, "nobody", "2026-10-20"); err == nil {
		t.Fatal("a stranger is no principal")
	}

	// The roles are the system's: a role a company declares, of a role's
	// name or another, changes nothing (until 2026-10-27 the company's own
	// role won over the pack's of the same name).
	door(at(office), "declare_role", map[string]any{"name": "user", "rights": "*", "description": "here every user is the owner"})
	door(at(office), "grant_membership", map[string]any{"person": id("anna"), "company": office, "role": "user", "from": "2026-10-21", "granted_by": "kk"})
	anna, _ = rights.Resolve(ctx, s, "anna", "2026-10-21")
	if _, ok := anna.May(office, "door", "close_period", "periods"); ok {
		t.Fatalf("a user holds nothing of her own — her groups do: %+v", anna.Grants)
	}
	if _, ok := anna.May(office, "view", "balance", "finance"); ok {
		t.Fatalf("…nor reads: %+v", anna.Grants)
	}
	_ = core.ActorHuman
}

// A grant is the people pack's, written in the company it grants into
// (review 2026-10-27, critical 2): the rights read every membership, mandate
// and substitution with the owner's eyes and believed the company each
// named, whichever rule wrote it and wherever. Anna, an accountant at the
// office, approved a rule of the office's own that books a membership — her
// right to approve the office's own rules is enough — and named another
// company, or herself its owner; or a door of the installation's let
// Rhea's own rule book a grant into a company nobody acting for it gave.
// Each made her a member where nobody let her in.
func TestAGrantIsThePeoplePacksInItsOwnCompany(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s, Today: func() string { return "2026-10-20" }}
	sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", "common", "people.json"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-09-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve: %v %v", err, errs)
	}
	trigger := func(ctx context.Context, name string, inputs map[string]any) {
		t.Helper()
		tr, err := x.TriggerActivity(ctx, name, inputs, "shell:kk", "")
		if err != nil || tr.Booked != 1 || len(tr.Errors) != 0 {
			t.Fatalf("%s: %+v %v", name, tr, err)
		}
	}
	submit := func(ctx context.Context, typ string, payload map[string]any) {
		t.Helper()
		trigger(ctx, "submit_event", map[string]any{"event_type": typ, "occurred_at": "2026-10-01", "payload": payload})
	}
	office, other := "PL5250001111", "PL5260001246"
	at := func(company string) context.Context { return store.WithCompany(ctx, company) }
	trigger(ctx, "register_person", map[string]any{"name": "Krzysztof Kowal", "handle": "kk", "role": "owner"})
	trigger(ctx, "register_person", map[string]any{"name": "Anna Nowak", "handle": "anna", "role": "accountant"})
	id := func(handle string) string {
		ids, err := s.FindObjectIDsByField(ctx, "person", "handle", handle)
		if err != nil || len(ids) != 1 {
			t.Fatalf("person %s: %v %v", handle, ids, err)
		}
		return ids[0]
	}
	kk, anna := id("kk"), id("anna")
	trigger(at(office), "grant_membership", map[string]any{"person": anna, "company": office, "role": "user", "from": "2026-01-01", "granted_by": "kk"})
	trigger(at(other), "grant_membership", map[string]any{"person": kk, "company": other, "role": "company_manager", "from": "2026-01-01", "granted_by": "kk"})

	// the office's own rule, booking a membership wherever its fact says
	claim := core.Rule{ID: "claim@" + office, Company: office, Status: core.StatusActive, Priority: 100, EffectiveFrom: "2026-01-01",
		CreatedBy: "agent", Description: "a claim", Spec: core.RuleSpec{Match: core.Match{EventType: "claim.made"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "membership", Fields: map[string]string{
				"person": "=$.person", "company": "=$.company", "role": "=$.role", "from": "=$.from", "status": "active", "granted_by": "=$.granted_by"}}}}}
	if _, err := s.InsertRuleVersion(ctx, claim); err != nil {
		t.Fatal(err)
	}
	submit(at(office), "claim.made", map[string]any{"person": anna, "company": other, "role": "owner", "from": "2026-01-01", "granted_by": "kk"})
	submit(at(office), "claim.made", map[string]any{"person": anna, "company": office, "role": "owner", "from": "2026-01-01", "granted_by": "kk"})
	// the installation's facts, booked by the people pack's own rules into
	// a company no fact of its own named
	installation := store.WithInstallation(ctx)
	submit(installation, "membership.granted", map[string]any{"person": anna, "company": other, "role": "owner", "from": "2026-01-01", "granted_by": "kk"})
	submit(installation, "mandate.granted", map[string]any{"office": office, "client": other, "scope": "*", "from": "2026-01-01", "granted_by": "kk"})
	submit(installation, "substitution.appointed", map[string]any{"person": kk, "deputy": anna, "company": other, "from": "2026-01-01", "to": "2026-12-31", "granted_by": "kk"})
	submit(at(office), "substitution.appointed", map[string]any{"person": kk, "deputy": anna, "company": other, "from": "2026-01-01", "to": "2026-12-31", "granted_by": "kk"})

	p, err := rights.Resolve(ctx, s, "anna", "2026-10-20")
	if err != nil {
		t.Fatal(err)
	}
	if got := p.Companies(); len(got) != 1 || got[0] != office {
		t.Fatalf("anna reaches %v: %+v", got, p.Grants)
	}
	if _, ok := p.May(office, "door", "grant_membership", "people"); ok {
		t.Fatalf("anna made herself the office's owner: %+v", p.Grants)
	}
	// the grants given at their own company's door still hold
	if k, _ := rights.Resolve(ctx, s, "kk", "2026-10-20"); len(k.Companies()) != 1 || k.Grants[0].Company != other || k.Grants[0].Role != "company_manager" {
		t.Fatalf("kk's own grant: %+v", k.Grants)
	}
}
