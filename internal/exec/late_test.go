package exec_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

// The late-understanding attempt (DIRECTION, "Late understanding", adopted
// 2026-10-08), kept as executable evidence. Bank accounts exist before anyone
// knew to ask for their IBANs; the type grows to v2; the IBANs arrive later as
// facts; a check nobody thought of is approved after the accounts were opened.
// Evolution carried as pure data; enrichment failed as predicted and earned
// per-field consent (amendable); backfill fails as predicted.
func TestLateUnderstandingAttempt(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}

	v1 := core.ObjectType{Name: "bank_account", Version: 1, Domain: "finance", LabelField: "number",
		Fields: []core.FieldDef{
			{Name: "number", Type: "string", Required: true},
			{Name: "bank", Type: "string", Required: true}}}
	if err := s.InsertObjectType(ctx, v1); err != nil {
		t.Fatal(err)
	}
	activate := func(id string, priority int, spec core.RuleSpec) core.Rule {
		t.Helper()
		r, err := s.InsertRuleVersion(ctx, core.Rule{ID: id, Status: core.StatusDraft, Priority: priority,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: spec})
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	submit := func(typ, date, payload string) {
		t.Helper()
		var p map[string]any
		json.Unmarshal([]byte(payload), &p)
		if _, err := x.TriggerActivity(ctx, "submit_event", map[string]any{
			"event_type": typ, "occurred_at": date, "payload": p}, "test", ""); err != nil {
			t.Fatal(err)
		}
	}
	activate("open-bank-account", 10, core.RuleSpec{Match: core.Match{EventType: "bank_account.opened"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "bank_account",
			Fields: map[string]string{"number": "=$.number", "bank": "=$.bank"}}}})
	if _, _, _, err := x.ApproveRule(ctx, "open-bank-account", "krzysztof", "2026-09-01"); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{
		`{"number":"PL-1","bank":"mBank"}`, `{"number":"PL-2","bank":"PKO BP"}`, `{"number":"DE-1","bank":"Commerzbank"}`,
	} {
		submit("bank_account.opened", "2026-09-02", p)
	}

	// --- 1. evolution carries: v2 adds optional fields, old objects keep v1 --
	v2 := v1
	v2.Version = 2
	v2.Fields = append(append([]core.FieldDef{}, v1.Fields...),
		core.FieldDef{Name: "iban", Type: "string"}, core.FieldDef{Name: "bank_country", Type: "string"})
	if err := s.InsertObjectType(ctx, v2); err != nil {
		t.Fatal(err)
	}
	accounts, _ := s.ObjectsByType(ctx, "bank_account")
	if len(accounts) != 3 {
		t.Fatalf("accounts = %d", len(accounts))
	}
	for _, a := range accounts {
		if a.TypeVersion != 1 {
			t.Fatalf("%s rewritten to v%d — a version bump must never touch old objects", a.ID, a.TypeVersion)
		}
	}

	// --- 2. enrichment fails: the IBANs are new facts with nowhere to go -----
	activate("enrich-bank-account", 20, core.RuleSpec{Match: core.Match{EventType: "bank_account.details_provided"},
		Effect: core.Effect{Amend: &core.AmendTemplate{Type: "bank_account",
			Target: "=ref(bank_account, number, $.number)",
			Set:    map[string]string{"iban": "=$.iban", "bank_country": "=$.country"}}}})
	_, _, procErrs, err := x.ApproveRule(ctx, "enrich-bank-account", "krzysztof", "2026-10-01")
	if err != nil {
		t.Fatal(err)
	}
	submit("bank_account.details_provided", "2026-10-01", `{"number":"PL-1","iban":"PL61109010140000071219812874","country":"PL"}`)
	if wl, _ := s.UnmatchedRawEvents(ctx); len(wl) != 1 {
		t.Fatalf("worklist = %d, want the IBAN waiting: %v", len(wl), procErrs)
	}
	if _, errs := x.ProcessPending(ctx); len(errs) != 1 || !strings.Contains(errs[0].Error(), "declares no lifecycle") {
		t.Fatalf("predicted refusal (no lifecycle, no consent), got %v", errs)
	}

	// The contortion the attempt recorded (DECISIONS 2026-10-08): a lifecycle
	// invented only to buy consent used to open EVERY field — the account
	// number included. Consent is per field now: the invented lifecycle buys
	// its own status and nothing else.
	contorted := v2
	contorted.Version = 3
	contorted.Fields = append(append([]core.FieldDef{}, v2.Fields...),
		core.FieldDef{Name: "state", Type: "enum", Values: []string{"open"}})
	contorted.Lifecycle = &core.LifecycleDef{Field: "state", Transitions: map[string][]string{"open": {}}}
	if err := (core.RuleSpec{Match: core.Match{EventType: "bank_account.renumbered"},
		Effect: core.Effect{Amend: &core.AmendTemplate{Type: "bank_account",
			Target: "=ref(bank_account, number, $.old)", Set: map[string]string{"number": "=$.new"}}}}).Validate(&contorted); err == nil ||
		!strings.Contains(err.Error(), "fixed at birth") {
		t.Fatalf("the account number must be fixed at birth: %v", err)
	}

	// The exit: v3 declares which fields legitimately arrive later. The
	// waiting IBAN books on the next pass, and nothing else became mutable.
	v3 := v2
	v3.Version = 3
	v3.Amendable = []string{"iban", "bank_country"}
	if err := s.InsertObjectType(ctx, v3); err != nil {
		t.Fatal(err)
	}
	if booked, errs := x.ProcessPending(ctx); booked != 1 || len(errs) != 0 {
		t.Fatalf("enrichment: booked %d, %v", booked, errs)
	}
	accounts, _ = s.ObjectsByType(ctx, "bank_account")
	if !strings.Contains(stateOf(accounts, "PL-1"), `"iban":"PL61109010140000071219812874"`) ||
		!strings.Contains(stateOf(accounts, "PL-1"), `"bank_country":"PL"`) {
		t.Fatalf("PL-1 not enriched: %v", stateOf(accounts, "PL-1"))
	}
	// Enrichment is a fact with provenance, never a silent edit: the IBAN's
	// delta is an object.amended event caused by the details event.
	for _, a := range accounts {
		if a.State["number"] != "PL-1" {
			continue
		}
		amends, _ := s.AmendmentsOf(ctx, a.ID)
		if len(amends) != 1 || amends[0].RuleID != "enrich-bank-account" {
			t.Fatalf("IBAN provenance = %+v", amends)
		}
		if a.TypeVersion != 1 {
			t.Fatalf("enrichment rewrote the birth version: v%d", a.TypeVersion)
		}
	}

	// --- 3. backfill fails: a check approved after the accounts were opened --
	if err := s.InsertObjectType(ctx, core.ObjectType{Name: "kyc_check", Version: 1, Domain: "finance",
		Fields: []core.FieldDef{{Name: "account", Type: "string", Required: true}}}); err != nil {
		t.Fatal(err)
	}
	activate("kyc-on-open", 30, core.RuleSpec{Match: core.Match{EventType: "bank_account.opened"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "kyc_check",
			Fields: map[string]string{"account": "=$.number"}}}})
	diff, err := x.SimulateRule(ctx, "kyc-on-open")
	if err != nil {
		t.Fatal(err)
	}
	if len(diff.Added) != 3 {
		t.Fatalf("dry run should see the three past accounts: added %d", len(diff.Added))
	}
	if _, booked, _, err := x.ApproveRule(ctx, "kyc-on-open", "krzysztof", "2026-10-02"); err != nil || booked != 0 {
		t.Fatalf("approve: booked %d, %v", booked, err)
	}
	// The dry run saw three checks; approval can deliver none. The explained
	// events are never revisited by the live path — understanding arrived
	// after the facts. Nothing happens silently:
	if checks, _ := s.ObjectsByType(ctx, "kyc_check"); len(checks) != 0 {
		t.Fatalf("checks = %d — backfill happened silently", len(checks))
	}

	// --- the exit: ruled backfill, a deliberate act with a dry run -----------
	plan, err := x.PlanBackfill(ctx, "2026-10-02")
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Chains) != 3 || len(plan.Diff.Added) != 3 || len(plan.Errors) != 0 {
		t.Fatalf("plan: %d chains, %d added, errors %v", len(plan.Chains), len(plan.Diff.Added), plan.Errors)
	}
	for _, c := range plan.Chains {
		if c.BookedAt != "2026-09-02" || c.Forwarded != "" || len(c.Rules) != 1 || c.Rules[0] != "kyc-on-open" {
			t.Fatalf("chain = %+v — dated by the fact, only the missing rule", c)
		}
	}
	before, _ := s.AllObjects(ctx)
	n, errs, err := x.ApproveBackfill(ctx, "krzysztof", "2026-10-02")
	if err != nil || n != 3 || len(errs) != 0 {
		t.Fatalf("backfill: %d, %v, %v", n, errs, err)
	}
	checks, _ := s.ObjectsByType(ctx, "kyc_check")
	if len(checks) != 3 {
		t.Fatalf("checks = %d", len(checks))
	}
	// Additive only: nothing that existed changed, and each check is caused
	// by its original opening event, dated by it.
	after, _ := s.AllObjects(ctx)
	if len(after) != len(before)+3 {
		t.Fatalf("objects %d → %d", len(before), len(after))
	}
	for _, c := range checks {
		ev, _ := s.GetEvent(ctx, c.SourceEventID)
		if ev.Type != "bank_account.opened" || ev.OccurredAt != "2026-09-02" {
			t.Fatalf("check caused by %s on %s", ev.Type, ev.OccurredAt)
		}
	}
	approval, ok, _ := s.LatestEventOfType(ctx, "backfill.approved")
	if !ok || approval.ActivityName != "approve_backfill" || !strings.Contains(string(approval.Payload), `"kyc-on-open"`) {
		t.Fatalf("approval = %+v", approval)
	}
	// Idempotent by construction: the past is now explained as far as the
	// rules go, and a second backfill has nothing to promote.
	if _, _, err := x.ApproveBackfill(ctx, "krzysztof", "2026-10-03"); err == nil || !strings.Contains(err.Error(), "nothing to backfill") {
		t.Fatalf("second backfill: %v", err)
	}
	assertReplayIdentical(t, x)
}

func assertReplayIdentical(t *testing.T, x *exec.Executor) {
	t.Helper()
	ctx := context.Background()
	before, _ := x.Store.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := x.Store.AllObjects(ctx)
	b, _ := json.Marshal(byID(before))
	a, _ := json.Marshal(byID(after))
	if string(a) != string(b) {
		t.Fatalf("replay diverged:\n%s\n%s", b, a)
	}
}

func byID(objs []core.Object) map[string]map[string]any {
	m := map[string]map[string]any{}
	for _, o := range objs {
		m[o.ID] = o.State
	}
	return m
}

func stateOf(objs []core.Object, number string) string {
	for _, o := range objs {
		if o.State["number"] == number {
			b, _ := json.Marshal(o.State)
			return string(b)
		}
	}
	return ""
}
