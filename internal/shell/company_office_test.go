package shell_test

import (
	"testing"

	"rhea/internal/core"
	"rhea/internal/store"
)

// An office teaches once (KK, 2026-10-22): anna, the office's accountant,
// approves alfa's own lesson for beta in one act, and beta gets it as its own
// rule — seen by beta's owner, written down in beta's books as anna's
// approval, booking beta's facts. A client the office may not approve for
// gets nothing, and then neither does any other: the act is whole or
// nothing. The same act twice is once.
func TestAnOfficeTeachesManyClientsOnce(t *testing.T) {
	w := newGrantsWorld(t)
	office, a, b, c := "PL7010001234", alfaKey, betaKey, "PL8010004321"
	w.person("Anna Nowak", "anna", office, "accountant")
	w.person("Beta Owner", "beta", b, "owner")
	w.mustDoor("", a, "grant_mandate", map[string]any{"office": office, "client": a, "scope": "door:* view:* type:*", "from": "2026-01-01", "granted_by": "operator"})
	w.mustDoor("", b, "grant_mandate", map[string]any{"office": office, "client": b, "scope": "door:* view:* type:*", "from": "2026-01-01", "granted_by": "operator"})
	w.mustDoor("", c, "grant_mandate", map[string]any{"office": office, "client": c, "scope": "view:* type:*", "from": "2026-01-01", "granted_by": "operator"})
	if _, err := w.s.InsertRuleVersion(w.ctx, core.Rule{ID: "memo@" + a, Company: a, Status: core.StatusActive, Priority: 100,
		EffectiveFrom: "2026-01-01", CreatedBy: "agent:test", Description: "a memo is a note",
		Spec: core.RuleSpec{Match: core.Match{EventType: "memo.taken"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text", "amount": "=$.amount"}}}}}); err != nil {
		t.Fatal(err)
	}
	teach := func(clients string) (int, map[string]any) {
		t.Helper()
		return w.door("anna", a, "approve_for_clients", map[string]any{"rule_id": "memo@" + a, "clients": clients,
			"approved_by": "anna", "occurred_at": w.today})
	}
	betas := store.WithCompany(w.ctx, b)

	if code, out := teach(b + "," + c); code != 403 {
		t.Fatalf("teaching beta and a client anna may not approve for: %d %v", code, out)
	}
	if _, err := w.s.GetRule(betas, "memo@"+b); err == nil {
		t.Fatal("beta was taught although the act was refused")
	}
	if code, out := teach(b); code != 200 {
		t.Fatalf("teaching beta: %d %v", code, out)
	}
	r, err := w.s.GetRule(betas, "memo@"+b)
	if err != nil || r.Company != b || r.Status != core.StatusActive || r.Version != 1 {
		t.Fatalf("beta's own copy: %+v %v", r, err)
	}
	if _, err := w.s.GetRule(store.WithCompany(w.ctx, c), "memo@"+b); err == nil {
		t.Fatal("c reads beta's rule")
	}
	approvals, err := w.s.EventsOfTypes(betas, []string{"rule.approved_for_clients"})
	if err != nil || len(approvals) != 1 || approvals[0].Company != b || approvals[0].Actor != "anna" {
		t.Fatalf("beta's books say who approved it for beta: %v %+v", err, approvals)
	}

	// Beta's memo books by beta's own copy.
	w.mustDoor("", b, "submit_event", map[string]any{"event_type": "memo.taken", "occurred_at": w.today,
		"payload": map[string]any{"text": "beta's memo", "amount": "2.00"}})
	if ids, _ := w.s.FindObjectIDsByField(betas, "note", "text", "beta's memo"); len(ids) != 1 {
		t.Fatalf("beta's memo booked: %v", ids)
	}

	// The same act twice is once.
	if code, _ := teach(b); code == 200 {
		t.Fatal("taught beta twice")
	}
	if r, _ := w.s.GetRule(betas, "memo@"+b); r.Version != 1 {
		t.Fatalf("beta's copy moved: v%d", r.Version)
	}
}
