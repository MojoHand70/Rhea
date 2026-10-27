package rights_test

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/rights"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

// people is an installation with the people pack approved and a few persons
// registered: kk, anna, piotr, ewa.
type people struct {
	ctx context.Context
	s   *store.Store
	x   *exec.Executor
	t   *testing.T
}

func newPeople(t *testing.T) *people {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s, Today: func() string { return "2026-10-27" }}
	sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", "common", "people.json"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-09-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve: %v %v", err, errs)
	}
	w := &people{ctx: ctx, s: s, x: x, t: t}
	for _, p := range [][2]string{{"Krzysztof Kowal", "kk"}, {"Anna Nowak", "anna"}, {"Piotr Lis", "piotr"}, {"Ewa Wiśniewska", "ewa"}} {
		w.door(ctx, "register_person", map[string]any{"name": p[0], "handle": p[1], "role": "user"})
	}
	return w
}

func (w *people) at(company string) context.Context { return store.WithCompany(w.ctx, company) }

// door fires a door and answers the object its fact made.
func (w *people) door(ctx context.Context, name string, inputs map[string]any) string {
	w.t.Helper()
	tr, err := w.x.TriggerActivity(ctx, name, inputs, "shell:kk", "")
	if err != nil {
		w.t.Fatalf("%s: %v", name, err)
	}
	if tr.Booked != 1 || len(tr.Errors) != 0 {
		w.t.Fatalf("%s: booked %d, errors %v", name, tr.Booked, tr.Errors)
	}
	typ := map[string]string{"grant_membership": "membership", "join_group": "group_member", "declare_group": "group"}[name]
	return fmt.Sprintf("%s-%d", typ, tr.EventID)
}

func (w *people) id(handle string) string {
	w.t.Helper()
	ids, err := w.s.FindObjectIDsByField(w.ctx, "person", "handle", handle)
	if err != nil || len(ids) != 1 {
		w.t.Fatalf("person %s: %v %v", handle, ids, err)
	}
	return ids[0]
}

func (w *people) member(company, handle, role, from string) string {
	return w.door(w.at(company), "grant_membership", map[string]any{"person": w.id(handle), "company": company, "role": role, "from": from, "granted_by": "kk"})
}

func (w *people) join(company, group string, who map[string]any, from string) string {
	in := map[string]any{"group": group, "company": company, "from": from, "granted_by": "kk"}
	for k, v := range who {
		in[k] = v
	}
	return w.door(w.at(company), "join_group", in)
}

func (w *people) resolve(handle, day string) rights.Principal {
	w.t.Helper()
	p, err := rights.Resolve(w.ctx, w.s, handle, day)
	if err != nil {
		w.t.Fatal(err)
	}
	return p
}

// The grants made under the roles of 2026-10-19 stay in the log as they were
// written and read as what they became (KK, 2026-10-27): owner as Company
// Manager, support as Rhea Support, chief accountant, accountant and reader
// as a User in the group of that name — so the people let in before the
// change keep exactly what they could do, and a door takes the seven roles
// alone from now on.
func TestFormerRolesReadAsTheyBecame(t *testing.T) {
	w := newPeople(t)
	co := "PL5250001111"
	// as white01's log holds them: facts the people pack's door wrote
	// under its v3 roles, the people pack's own rule booking each
	for _, g := range [][2]string{{"kk", "owner"}, {"anna", "accountant"}, {"piotr", "reader"}, {"ewa", "support"}} {
		w.door(w.at(co), "submit_event", map[string]any{"event_type": "membership.granted", "occurred_at": "2026-10-01",
			"payload": map[string]any{"person": w.id(g[0]), "company": co, "role": g[1], "from": "2026-10-01", "granted_by": "kk"}})
	}
	kk := w.resolve("kk", "2026-10-27")
	if g, ok := kk.May(co, "door", "grant_membership", "people"); !ok || g.Role != core.RoleCompanyManager {
		t.Fatalf("the owner is the Company Manager: %+v %v", g, ok)
	}
	anna := w.resolve("anna", "2026-10-27")
	if g, ok := anna.May(co, "door", "close_period", "periods"); !ok || g.Via != "group" || g.Group != "accountant" || !strings.HasPrefix(g.Object, "membership-") {
		t.Fatalf("the accountant keeps the books, in the group Accountant her old grant names: %+v %v", g, ok)
	}
	if _, ok := anna.May(co, "door", "grant_membership", "people"); ok {
		t.Fatal("an accountant never says who may do what")
	}
	if anna.Grants[0].Role != core.RoleUser {
		t.Fatalf("an accountant is a User: %+v", anna.Grants[0])
	}
	piotr := w.resolve("piotr", "2026-10-27")
	if _, ok := piotr.May(co, "view", "balance", "finance"); !ok {
		t.Fatal("the reader reads")
	}
	if _, ok := piotr.May(co, "door", "close_period", "periods"); ok {
		t.Fatal("…and changes nothing")
	}
	ewa := w.resolve("ewa", "2026-10-27")
	if g, ok := ewa.May(co, "door", "close_period", "periods"); !ok || g.Role != core.RoleRheaSupport {
		t.Fatalf("support is Rhea Support, the books: %+v %v", g, ok)
	}
	if _, ok := ewa.May(co, "door", "grant_membership", "people"); ok {
		t.Fatal("…never who may do what")
	}
	// the door takes the system's roles alone
	_, err := w.x.TriggerActivity(w.at(co), "grant_membership", map[string]any{"person": w.id("anna"), "company": co, "role": "accountant", "from": "2026-10-27", "granted_by": "kk"}, "shell:kk", "")
	if err == nil || !strings.Contains(err.Error(), "company_manager") {
		t.Fatalf("a door granting a former role: %v", err)
	}
	// the groups list counts each of them where their old grant puts them
	groups, err := rights.Groups(w.ctx, w.s, co, "2026-10-27")
	if err != nil {
		t.Fatal(err)
	}
	in := map[string][]string{}
	for _, g := range groups {
		for _, m := range g.Members {
			in[g.Name] = append(in[g.Name], m.Handle+"/"+m.Counted)
		}
	}
	if fmt.Sprint(in["accountant"]) != "[anna/former role]" || fmt.Sprint(in["reader"]) != "[piotr/former role]" || fmt.Sprint(in["chief_accountant"]) != "[kk/company manager]" {
		t.Fatalf("who is in which group: %v", in)
	}
	if len(in["all_users"]) != 4 {
		t.Fatalf("all users: %v", in["all_users"])
	}
}

// A group says what its members may do (KK, 2026-10-27: "The rest we can
// manage by Groups"): a person in it, or a group in it, holds what it gives
// between the place's two dates, in its own company alone, and only while
// let into that company; a group inside a group passes the outer group's
// rights down; All users is everyone let in, a guest apart. A company's own
// group never stands in for one the pack gives every company, and another
// company sees nothing of it. Replay gives the same answers.
func TestGroupsGiveWhatTheyHold(t *testing.T) {
	w := newPeople(t)
	a, b := "PL5250001111", "PL5260001246"
	w.member(a, "kk", "company_manager", "2026-01-01")
	w.member(b, "kk", "company_manager", "2026-01-01")
	w.member(a, "anna", "user", "2026-10-01")
	w.member(a, "piotr", "user", "2026-10-01")
	w.member(a, "ewa", "guest", "2026-10-01")
	w.member(b, "anna", "user", "2026-10-01")

	// a company's own group, and one inside another
	w.door(w.at(a), "declare_group", map[string]any{"name": "Bankowość", "company": a, "rights": "door:approve_bank_account", "description": "odpowiedzialni za banki"})
	w.door(w.at(a), "declare_group", map[string]any{"name": "Skarb", "company": a, "rights": "door:pay_supplier"})
	w.door(w.at(b), "declare_group", map[string]any{"name": "Bankowość", "company": b, "rights": "*"})
	w.join(a, "Bankowość", map[string]any{"person": w.id("anna")}, "2026-10-20")
	w.join(a, "Skarb", map[string]any{"member_group": "Bankowość"}, "2026-10-20")
	place := w.join(a, "accountant", map[string]any{"person": w.id("piotr")}, "2026-10-01")

	anna := w.resolve("anna", "2026-10-27")
	if g, ok := anna.May(a, "door", "approve_bank_account", "finance"); !ok || g.Group != "Bankowość" {
		t.Fatalf("anna approves bank accounts in a, in Bankowość: %+v %v", g, ok)
	}
	if g, ok := anna.May(a, "door", "pay_supplier", "finance"); !ok || g.Group != "Skarb" || g.Through == "" {
		t.Fatalf("…and pays, Bankowość being in Skarb: %+v %v", g, ok)
	}
	if _, ok := anna.May(b, "door", "approve_bank_account", "finance"); ok {
		t.Fatalf("b's Bankowość is b's, and anna is in a's: %+v", anna.Grants)
	}
	if early := w.resolve("anna", "2026-10-19"); func() bool { _, ok := early.May(a, "door", "approve_bank_account", "finance"); return ok }() {
		t.Fatal("before her place's first day she holds nothing of the group")
	}
	// the pack's groups are every company's: a company's own of the same
	// name is refused at the door and never stands in
	if _, err := w.x.TriggerActivity(w.at(b), "submit_event", map[string]any{"event_type": "group.declared", "occurred_at": "2026-10-27",
		"payload": map[string]any{"name": "reader", "company": b, "rights": "*"}}, "shell:kk", ""); err != nil {
		t.Fatal(err)
	}
	w.door(w.at(b), "join_group", map[string]any{"group": "reader", "person": w.id("anna"), "company": b, "from": "2026-10-01", "granted_by": "kk"})
	if p := w.resolve("anna", "2026-10-27"); func() bool { _, ok := p.May(b, "door", "close_period", "periods"); return ok }() {
		t.Fatalf("b's own 'reader' with every right stood in for the pack's: %+v", p.Grants)
	}
	// All users: everyone let in, a guest apart; put in a group, it gives
	// everyone that group
	w.join(a, "reader", map[string]any{"member_group": core.GroupAllUsers}, "2026-10-01")
	for handle, reads := range map[string]bool{"anna": true, "piotr": true, "ewa": false} {
		p := w.resolve(handle, "2026-10-27")
		if _, ok := p.May(a, "view", "balance", "finance"); ok != reads {
			t.Fatalf("%s reads a's books through All users: %v, want %v (%+v)", handle, ok, reads, p.Grants)
		}
	}
	// a place in a group is no way in, and gives nothing once its day ends
	w.door(w.at(a), "leave_group", map[string]any{"group_member": place, "to": "2026-10-26", "ended_by": "kk"})
	if p := w.resolve("piotr", "2026-10-27"); func() bool { _, ok := p.May(a, "door", "close_period", "periods"); return ok }() {
		t.Fatalf("piotr's place in Accountant ended yesterday: %+v", p.Grants)
	}
	if o, _ := w.s.GetObject(w.ctx, place); o.State["status"] != "ended" || o.State["to"] != "2026-10-26" {
		t.Fatalf("an ended place keeps its history: %v", o.State)
	}
	if _, err := w.x.TriggerActivity(w.at(a), "submit_event", map[string]any{"event_type": "group.joined", "occurred_at": "2026-10-27",
		"payload": map[string]any{"group": "Bankowość", "person": w.id("ewa"), "company": b, "from": "2026-10-01", "granted_by": "kk"}}, "shell:kk", ""); err != nil {
		t.Fatal(err)
	}
	if p := w.resolve("ewa", "2026-10-27"); len(p.Companies()) != 1 || p.Companies()[0] != a {
		t.Fatalf("a place a's fact wrote into b's group let ewa into b: %+v", p.Grants)
	}

	before := map[string]rights.Principal{}
	for _, h := range []string{"kk", "anna", "piotr", "ewa"} {
		before[h] = w.resolve(h, "2026-10-27")
	}
	if _, err := w.x.Replay(w.ctx); err != nil {
		t.Fatal(err)
	}
	for h, p := range before {
		if after := w.resolve(h, "2026-10-27"); !reflect.DeepEqual(p, after) {
			t.Fatalf("replay changed %s's grants:\n%+v\n%+v", h, p.Grants, after.Grants)
		}
	}
}

// Who is in which group is a role's to say (KK, 2026-10-27: the Rhea
// Designer, the Tenant Owner, the Office Manager, the Company Manager), and
// nobody raises their own rights by their own hand: a Tenant Owner, who sees
// no books by that role alone, puts the accountant into Accountant but not
// herself, nor a group she is in, nor gives herself a role above hers; a
// group is never inside itself; All users is Rhea's; a person is let in
// before a group takes them. The Company Manager, who holds everything, may.
func TestNobodyRaisesTheirOwnRights(t *testing.T) {
	w := newPeople(t)
	co := "PL5250001111"
	w.member(co, "kk", "company_manager", "2026-01-01")
	w.member(co, "ewa", "tenant_owner", "2026-01-01")
	w.member(co, "anna", "user", "2026-10-01")
	w.door(w.at(co), "declare_group", map[string]any{"name": "Zarząd", "company": co, "rights": "view:*"})
	w.door(w.at(co), "declare_group", map[string]any{"name": "Bankowość", "company": co, "rights": "door:approve_bank_account"})
	w.join(co, "Bankowość", map[string]any{"member_group": "Zarząd"}, "2026-10-01")
	w.join(co, "Zarząd", map[string]any{"person": w.id("ewa")}, "2026-10-01")

	act := func(name string) core.Activity {
		a, err := w.s.GetActivity(w.ctx, name)
		if err != nil {
			t.Fatal(err)
		}
		return a
	}
	try := func(handle, door string, inputs map[string]any) error {
		t.Helper()
		_, err := rights.Authorize(w.ctx, w.s, w.resolve(handle, "2026-10-27"), act(door), inputs, co, "2026-10-27")
		return err
	}
	join := func(who map[string]any) map[string]any {
		in := map[string]any{"group": "accountant", "company": co, "from": "2026-10-27"}
		for k, v := range who {
			in[k] = v
		}
		return in
	}
	refused := func(err error, why string) {
		t.Helper()
		if err == nil || !strings.Contains(err.Error(), why) {
			t.Fatalf("want refused for %q, got %v", why, err)
		}
	}
	if err := try("ewa", "join_group", join(map[string]any{"person": w.id("anna")})); err != nil {
		t.Fatalf("the tenant owner puts the accountant into Accountant: %v", err)
	}
	refused(try("ewa", "join_group", join(map[string]any{"person": w.id("ewa")})), "you cannot put yourself in accountant: it gives door:*")
	refused(try("ewa", "join_group", join(map[string]any{"member_group": "Zarząd"})), "you cannot put yourself in accountant")
	refused(try("ewa", "grant_membership", map[string]any{"person": w.id("ewa"), "company": co, "role": "company_manager", "from": "2026-10-27"}), "you cannot give yourself the role company_manager")
	if err := try("ewa", "grant_membership", map[string]any{"person": w.id("piotr"), "company": co, "role": "user", "from": "2026-10-27"}); err != nil {
		t.Fatalf("the tenant owner lets a user in: %v", err)
	}
	refused(try("ewa", "join_group", join(map[string]any{"person": w.id("piotr")})), "not let into the company")
	refused(try("ewa", "join_group", map[string]any{"group": core.GroupAllUsers, "company": co, "from": "2026-10-27", "person": w.id("anna")}), "Rhea keeps All users")
	refused(try("ewa", "join_group", map[string]any{"group": "Zarząd", "company": co, "from": "2026-10-27", "member_group": "Bankowość"}), "a group cannot be inside itself")
	refused(try("ewa", "join_group", map[string]any{"group": "Zarząd", "company": co, "from": "2026-10-27", "person": w.id("ewa")}), "in Zarząd already")
	refused(try("ewa", "join_group", map[string]any{"group": "Kasa", "company": co, "from": "2026-10-27", "person": w.id("anna")}), "no group Kasa")
	refused(try("ewa", "join_group", join(nil)), "name a person or a group")
	refused(try("ewa", "declare_group", map[string]any{"name": "accountant", "company": co}), "every company has")
	refused(try("ewa", "declare_group", map[string]any{"name": "Zarząd", "company": co}), "a group Zarząd already")
	refused(try("ewa", "declare_group", map[string]any{"name": "Kasa", "company": co, "rights": "door:"}), "does not read")
	// anna, a user in no group, decides nothing about anyone
	refused(try("anna", "join_group", join(map[string]any{"person": w.id("anna")})), "no role of yours opens this door")
	// the Company Manager holds everything, so putting himself anywhere
	// raises nothing
	if err := try("kk", "join_group", join(map[string]any{"person": w.id("kk")})); err != nil {
		t.Fatalf("the company manager puts himself in Accountant: %v", err)
	}
}

// A starting group nobody was put in yet holds the company's manager, so a
// sole trader does everything on day one; once someone is put in, it is
// theirs (KK, 2026-10-27: "until handed over").
func TestAStartingGroupIsTheManagersUntilHandedOver(t *testing.T) {
	w := newPeople(t)
	co := "PL5250001111"
	w.member(co, "kk", "company_manager", "2026-01-01")
	w.member(co, "anna", "user", "2026-10-01")
	who := func() map[string]string {
		groups, err := rights.Groups(w.ctx, w.s, co, "2026-10-27")
		if err != nil {
			t.Fatal(err)
		}
		out := map[string]string{}
		for _, g := range groups {
			var hs []string
			for _, m := range g.Members {
				hs = append(hs, m.Handle)
			}
			out[g.Name] = strings.Join(hs, ",")
		}
		return out
	}
	if got := who(); got["accountant"] != "kk" || got["chief_accountant"] != "kk" || got["reader"] != "kk" || got["all_users"] != "kk,anna" {
		t.Fatalf("day one: %v", got)
	}
	w.join(co, "accountant", map[string]any{"person": w.id("anna")}, "2026-10-27")
	if got := who(); got["accountant"] != "anna" || got["chief_accountant"] != "kk" {
		t.Fatalf("handed over: %v", got)
	}
}
