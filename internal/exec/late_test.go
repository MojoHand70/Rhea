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
// Evolution carries as pure data. Enrichment and backfill fail exactly as
// predicted.
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

	// The contortion, recorded: a lifecycle invented only to buy consent. It
	// works — and consent is all or nothing, so the same type now lets any
	// rule rewrite the account NUMBER, the identity the IBAN was keyed by.
	// Consent spelled as a status machine is too coarse: it must be per field.
	v3 := v2
	v3.Version = 3
	v3.Fields = append(append([]core.FieldDef{}, v2.Fields...),
		core.FieldDef{Name: "state", Type: "enum", Values: []string{"open"}})
	v3.Lifecycle = &core.LifecycleDef{Field: "state", Transitions: map[string][]string{"open": {}}}
	if err := s.InsertObjectType(ctx, v3); err != nil {
		t.Fatal(err)
	}
	if err := (core.RuleSpec{Match: core.Match{EventType: "bank_account.renumbered"},
		Effect: core.Effect{Amend: &core.AmendTemplate{Type: "bank_account",
			Target: "=ref(bank_account, number, $.old)", Set: map[string]string{"number": "=$.new"}}}}).Validate(&v3); err != nil {
		t.Fatalf("contortion should admit any field (that is the finding): %v", err)
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
	// The approval's worklist pass books one event: the waiting IBAN, which
	// the contortion's invented lifecycle has just bought consent for.
	if _, booked, _, err := x.ApproveRule(ctx, "kyc-on-open", "krzysztof", "2026-10-02"); err != nil || booked != 1 {
		t.Fatalf("approve: booked %d, %v", booked, err)
	}
	if pl1, _ := s.ObjectsByType(ctx, "bank_account"); !strings.Contains(stateOf(pl1, "PL-1"), "PL61109010140000071219812874") {
		t.Fatalf("contortion did not carry the IBAN: %v", pl1)
	}
	// The dry run saw three checks; approval can deliver none. The explained
	// events are never revisited — understanding arrived after the facts and
	// has no door in. Predicted: the backfill door.
	if checks, _ := s.ObjectsByType(ctx, "kyc_check"); len(checks) != 0 {
		t.Fatalf("checks = %d — backfill happened silently", len(checks))
	}
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
