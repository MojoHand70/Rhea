package exec_test

import (
	"context"
	"errors"
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

// A kind of record guards itself (KK, 2026-10-27: "such important metadata
// like bank accounts should be approved by someone responsible for bank
// accounts"; the approval guards the table, not one App's screen): an
// account begins proposed whichever rule makes it, one per number in a
// company, its number an IBAN whose check digits hold; the accountants may
// press every door, but only a person in Bankowość approves — never the one
// who proposed it while someone else in Bankowość could — and a door of
// another App that tries the same move meets the same law. A refusal is
// recorded and leaves nothing waiting. Replay gives the same records.
func TestARecordGuardsItsMoves(t *testing.T) {
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
	co, other := "PL5250001111", "PL5260001246"
	at := func(c string) context.Context { return store.WithCompany(ctx, c) }
	door := func(ctx context.Context, name string, inputs map[string]any) (exec.Triggered, error) {
		return x.TriggerActivity(ctx, name, inputs, "shell:kk", "")
	}
	must := func(ctx context.Context, name string, inputs map[string]any) string {
		t.Helper()
		tr, err := door(ctx, name, inputs)
		if err != nil || tr.Booked != 1 || len(tr.Errors) != 0 {
			t.Fatalf("%s: %+v %v", name, tr, err)
		}
		return fmt.Sprint(tr.EventID)
	}
	id := func(handle string) string {
		ids, _ := s.FindObjectIDsByField(ctx, "person", "handle", handle)
		return ids[0]
	}
	for _, h := range []string{"kk", "anna", "ewa", "piotr"} {
		must(ctx, "register_person", map[string]any{"name": h, "handle": h, "role": "user"})
	}
	for _, c := range []string{co, other} {
		must(at(c), "grant_membership", map[string]any{"person": id("kk"), "company": c, "role": "company_manager", "from": "2026-01-01", "granted_by": "kk"})
	}
	must(at(co), "declare_group", map[string]any{"name": "Bankowość", "company": co, "rights": ""})
	for _, h := range []string{"anna", "ewa", "piotr"} {
		must(at(co), "grant_membership", map[string]any{"person": id(h), "company": co, "role": "user", "from": "2026-01-01", "granted_by": "kk"})
		must(at(co), "join_group", map[string]any{"group": "accountant", "person": id(h), "company": co, "from": "2026-01-01", "granted_by": "kk"})
	}
	must(at(co), "join_group", map[string]any{"group": "Bankowość", "person": id("anna"), "company": co, "from": "2026-01-01", "granted_by": "kk"})

	account := core.ObjectType{Name: "account", Version: 1, Domain: "finance", LabelField: "number", Key: "number",
		Fields: []core.FieldDef{{Name: "number", Type: "string", Required: true, Format: "iban"},
			{Name: "status", Type: "enum", Required: true, Values: []string{"proposed", "active", "rejected"}},
			{Name: "proposed_by", Type: "string", Required: true}, {Name: "approved_by", Type: "string"}},
		Lifecycle: &core.LifecycleDef{Field: "status", Initial: "proposed",
			Transitions: map[string][]string{"proposed": {"active", "rejected"}},
			Moves: map[string]core.Move{"proposed>active": {Group: "Bankowość", NotBy: "proposed_by", Door: "approve_account", Label: "Zatwierdź"},
				"proposed>rejected": {Group: "Bankowość", Door: "reject_account", Label: "Odrzuć"}}},
		Amendable: []string{"approved_by"}}
	if err := s.InsertObjectType(ctx, account); err != nil {
		t.Fatal(err)
	}
	rule := func(id, event string, eff core.Effect) {
		if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: id, Status: core.StatusActive, Priority: 100, EffectiveFrom: "2026-01-01",
			CreatedBy: "test", Description: id, Spec: core.RuleSpec{Match: core.Match{EventType: event}, Effect: eff}}); err != nil {
			t.Fatal(err)
		}
	}
	activity := func(name, emits string, inputs ...core.FieldDef) {
		if _, err := s.InsertActivityVersion(ctx, core.Activity{Name: name, Status: core.StatusActive, Domain: "finance", Description: name,
			Spec: core.ActivitySpec{Inputs: inputs, Emits: emits, Who: []string{"human"}}}); err != nil {
			t.Fatal(err)
		}
	}
	activity("propose_account", "account.proposed", core.FieldDef{Name: "number", Type: "string", Required: true, Format: "iban"}, core.FieldDef{Name: "proposed_by", Type: "string", Required: true})
	activity("approve_account", "account.approved", core.FieldDef{Name: "account", Type: "ref<account>", Required: true}, core.FieldDef{Name: "approved_by", Type: "string", Required: true})
	activity("reject_account", "account.rejected", core.FieldDef{Name: "account", Type: "ref<account>", Required: true})
	activity("bless_account", "account.blessed", core.FieldDef{Name: "account", Type: "ref<account>", Required: true}) // another App's door
	activity("import_account", "account.imported", core.FieldDef{Name: "number", Type: "string", Required: true})
	rule("propose", "account.proposed", core.Effect{Object: core.ObjectTemplate{Type: "account", Fields: map[string]string{"number": "=$.number", "status": "proposed", "proposed_by": "=$.proposed_by"}}})
	rule("approve", "account.approved", core.Effect{Amend: &core.AmendTemplate{Type: "account", Target: "=$.account", Set: map[string]string{"status": "active", "approved_by": "=$.approved_by"}}})
	rule("reject", "account.rejected", core.Effect{Amend: &core.AmendTemplate{Type: "account", Target: "=$.account", Set: map[string]string{"status": "rejected"}}})
	rule("bless", "account.blessed", core.Effect{Amend: &core.AmendTemplate{Type: "account", Target: "=$.account", Set: map[string]string{"status": "active"}}})
	rule("import", "account.imported", core.Effect{Object: core.ObjectTemplate{Type: "account", Fields: map[string]string{"number": "=$.number", "status": "active", "proposed_by": "import"}}})

	as := func(handle string) context.Context {
		p, err := rights.Resolve(ctx, s, handle, "2026-10-27")
		if err != nil {
			t.Fatal(err)
		}
		return rights.WithPrincipal(at(co), p)
	}
	refusals := func() int {
		evs, _ := s.EventsOfTypes(ctx, []string{core.EventAccessRefused})
		return len(evs)
	}
	refused := func(err error, why string) {
		t.Helper()
		if err == nil || !strings.Contains(err.Error(), why) {
			t.Fatalf("want refused for %q, got %v", why, err)
		}
	}

	// the number as a person types it, kept as the bank prints it; a wrong digit refused
	if _, err := door(as("ewa"), "propose_account", map[string]any{"number": "PL61 1090 1014 0000 0712 1981 2875", "proposed_by": "x"}); err == nil || !strings.Contains(err.Error(), "check digits") {
		t.Fatalf("a mistyped IBAN: %v", err)
	}
	ewaProposes, err := door(as("ewa"), "propose_account", map[string]any{"number": "pl61 1090 1014 0000 0712 1981 2874", "proposed_by": "x"})
	if err != nil || ewaProposes.Booked != 1 {
		t.Fatalf("ewa proposes: %+v %v", ewaProposes, err)
	}
	acc := fmt.Sprintf("account-%d", ewaProposes.EventID)
	if o, _ := s.GetObject(ctx, acc); o.State["number"] != "PL61109010140000071219812874" || o.State["status"] != "proposed" || o.State["proposed_by"] != "ewa" {
		t.Fatalf("the proposed account: %v", o.State)
	}
	// one account per number, refused at the door with nothing written
	n := refusals()
	if _, err := door(as("piotr"), "propose_account", map[string]any{"number": "PL61109010140000071219812874", "proposed_by": "x"}); err == nil || !strings.Contains(err.Error(), "registered already") {
		t.Fatalf("a second account of the number: %v", err)
	}
	// born approved by another App's rule: refused, whatever the door
	if _, err := door(as("piotr"), "import_account", map[string]any{"number": "PL27114020040000300201355387"}); err == nil || !strings.Contains(err.Error(), "begins proposed") {
		t.Fatalf("an account born active: %v", err)
	}
	// piotr, an accountant who may press every door, is not in Bankowość
	_, err = door(as("piotr"), "approve_account", map[string]any{"account": acc, "approved_by": "x"})
	refused(err, "a person in Bankowość makes this move, and you are not in it")
	_, err = door(as("piotr"), "bless_account", map[string]any{"account": acc})
	refused(err, "you are not in it")
	if refusals() != n+2 {
		t.Fatalf("each refusal is recorded: %d", refusals()-n)
	}
	if o, _ := s.GetObject(ctx, acc); o.State["status"] != "proposed" {
		t.Fatalf("still proposed: %v", o.State)
	}
	// the machine never approves
	_, err = x.TriggerActivity(at(co), "approve_account", map[string]any{"account": acc, "approved_by": "kernel"}, "kernel", "")
	refused(err, "never the machine")

	// ewa joins Bankowość: she proposed it and anna could approve, so she may not
	must(at(co), "join_group", map[string]any{"group": "Bankowość", "person": id("ewa"), "company": co, "from": "2026-01-01", "granted_by": "kk"})
	_, err = door(as("ewa"), "approve_account", map[string]any{"account": acc, "approved_by": "x"})
	refused(err, "you proposed it — another person in Bankowość makes this move")
	if tr, err := door(as("anna"), "approve_account", map[string]any{"account": acc, "approved_by": "x"}); err != nil || tr.Booked != 1 {
		t.Fatalf("anna approves: %+v %v", tr, err)
	}
	if o, _ := s.GetObject(ctx, acc); o.State["status"] != "active" || o.State["approved_by"] != "anna" {
		t.Fatalf("approved by anna: %v", o.State)
	}
	waiting, err := s.UnmatchedRawEvents(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, ev := range waiting {
		if strings.HasPrefix(ev.Type, "account.") {
			t.Fatalf("a refused move waits in the worklist: %s %s", ev.Type, ev.Payload)
		}
	}

	// alone in Bankowość, the one who proposed approves too, and the record shows it
	alone, err := door(as("anna"), "propose_account", map[string]any{"number": "PL27114020040000300201355387", "proposed_by": "x"})
	if err != nil {
		t.Fatal(err)
	}
	must(at(co), "leave_group", map[string]any{"group_member": groupPlace(t, s, co, "Bankowość", id("ewa")), "to": "2026-10-26", "ended_by": "kk"})
	if tr, err := door(as("anna"), "approve_account", map[string]any{"account": fmt.Sprintf("account-%d", alone.EventID), "approved_by": "x"}); err != nil || tr.Booked != 1 {
		t.Fatalf("anna, alone in Bankowość, approves what she proposed: %+v %v", tr, err)
	}

	// another company: a's Bankowość is a's — kk, who manages b, finds
	// nobody in a Bankowość there, and approves nothing until he puts
	// someone in one
	kk, err := rights.Resolve(ctx, s, "kk", "2026-10-27")
	if err != nil {
		t.Fatal(err)
	}
	inB := rights.WithPrincipal(at(other), kk)
	there, err := door(inB, "propose_account", map[string]any{"number": "DE89370400440532013000", "proposed_by": "x"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = door(inB, "approve_account", map[string]any{"account": fmt.Sprintf("account-%d", there.EventID), "approved_by": "x"})
	refused(err, "nobody is in Bankowość here")
	// a record of no company — the installation's own, which no company's
	// door writes — has no group's people: its operator decides, never the
	// machine
	inst, err := door(ctx, "propose_account", map[string]any{"number": "GB82WEST12345698765432", "proposed_by": "operator"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.TriggerActivity(ctx, "approve_account", map[string]any{"account": fmt.Sprintf("account-%d", inst.EventID), "approved_by": "kernel"}, "kernel", "")
	refused(err, "never the machine")
	if tr, err := door(ctx, "approve_account", map[string]any{"account": fmt.Sprintf("account-%d", inst.EventID), "approved_by": "operator"}); err != nil || tr.Booked != 1 {
		t.Fatalf("the operator approves the installation's own: %+v %v", tr, err)
	}
	assertReplayIdentical(t, x)
}

// groupPlace finds a person's place in a group.
func groupPlace(t *testing.T, s *store.Store, company, group, person string) string {
	t.Helper()
	objs, err := s.ObjectsByType(store.WithCompany(context.Background(), company), "group_member")
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range objs {
		if o.State["group"] == group && o.State["person"] == person {
			return o.ID
		}
	}
	t.Fatalf("no place of %s in %s", person, group)
	return ""
}

var _ = errors.New
