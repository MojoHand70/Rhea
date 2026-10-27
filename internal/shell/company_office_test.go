package shell_test

import (
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/store"
)

// officeScene is an office (anna, its accountant) keeping three clients:
// alfa and beta under a mandate that opens every door, c under one that only
// reads; alfa has a rule of its own, that a memo is a note; beta has an owner.
type officeScene struct {
	w               *grantsWorld
	office, a, b, c string
}

func newOfficeScene(t *testing.T) *officeScene {
	w := newGrantsWorld(t)
	o := &officeScene{w: w, office: "PL7010001234", a: alfaKey, b: betaKey, c: "PL8010004321"}
	w.person("Anna Nowak", "anna", o.office, "accountant")
	w.person("Beta Owner", "beta", o.b, "company_manager")
	for client, scope := range map[string]string{o.a: "door:* view:* type:*", o.b: "door:* view:* type:*", o.c: "view:* type:*"} {
		w.mustDoor("", client, "grant_mandate", map[string]any{"office": o.office, "client": client, "scope": scope, "from": "2026-01-01", "granted_by": "operator"})
	}
	if _, err := w.s.InsertRuleVersion(w.ctx, core.Rule{ID: "memo@" + o.a, Company: o.a, Status: core.StatusActive, Priority: 100,
		EffectiveFrom: "2026-01-01", CreatedBy: "agent:test", Description: "a memo is a note",
		Spec: core.RuleSpec{Match: core.Match{EventType: "memo.taken"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text", "amount": "=$.amount"}}}}}); err != nil {
		t.Fatal(err)
	}
	return o
}

// teach is anna approving alfa's rule for the clients named.
func (o *officeScene) teach(clients string) (int, map[string]any) {
	o.w.t.Helper()
	return o.w.door("anna", o.a, "approve_for_clients", map[string]any{"rule_id": "memo@" + o.a, "clients": clients,
		"approved_by": "anna", "occurred_at": o.w.today})
}

// An office teaches once (KK, 2026-10-22): anna, the office's accountant,
// approves alfa's own lesson for beta in one act, and beta gets it as its own
// rule — seen by beta's owner, written down in beta's books as anna's
// approval, booking beta's facts. A client the office may not approve for
// gets nothing, and then neither does any other: the act is whole or
// nothing. The same act twice is once.
func TestAnOfficeTeachesManyClientsOnce(t *testing.T) {
	o := newOfficeScene(t)
	w, b, c, teach := o.w, o.b, o.c, o.teach
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

// A client switches a lesson off for itself (KK, 2026-10-22: the client's
// veto): beta's owner retires the copy the office taught, with a reason. It
// books beta's facts no more, the bookings it made stay, the reason is in
// beta's books for the office to read, and the office cannot teach it to beta
// again. A reader switches nothing off, and no company switches off another
// company's rule or Rhea's.
func TestAClientSwitchesALessonOff(t *testing.T) {
	o := newOfficeScene(t)
	w, b := o.w, o.b
	betas := store.WithCompany(w.ctx, b)
	w.person("Ben Reader", "ben", b, "reader")
	if code, out := o.teach(b); code != 200 {
		t.Fatalf("teaching beta: %d %v", code, out)
	}
	w.mustDoor("", b, "submit_event", map[string]any{"event_type": "memo.taken", "occurred_at": w.today,
		"payload": map[string]any{"text": "beta's first memo", "amount": "2.00"}})
	retire := func(person, rule string) (int, map[string]any) {
		t.Helper()
		return w.door(person, b, "retire_rule", map[string]any{"rule_id": rule, "reason": "we book memos by hand",
			"retired_by": person, "occurred_at": w.today})
	}
	if code, out := retire("ben", "memo@"+b); code != 403 {
		t.Fatalf("beta's reader switching a rule off: %d %v", code, out)
	}
	if code, _ := retire("beta", "note"); code == 200 {
		t.Fatal("beta switched off Rhea's rule")
	}
	if code, _ := retire("beta", "memo@"+o.a); code == 200 {
		t.Fatal("beta switched off alfa's rule")
	}
	if code, out := retire("beta", "memo@"+b); code != 200 {
		t.Fatalf("beta's owner switching beta's lesson off: %d %v", code, out)
	}
	if r, _ := w.s.GetRule(betas, "memo@"+b); r.Status != core.StatusSuperseded || r.Version != 2 {
		t.Fatalf("beta's lesson after the veto: %+v", r)
	}
	if r, _ := w.s.GetRule(store.WithCompany(w.ctx, o.a), "memo@"+o.a); r.Status != core.StatusActive {
		t.Fatalf("alfa's own rule went with beta's copy: %+v", r)
	}

	// It books beta's facts no more; what it booked stays.
	if code, out := w.door("", b, "submit_event", map[string]any{"event_type": "memo.taken", "occurred_at": w.today,
		"payload": map[string]any{"text": "beta's second memo", "amount": "3.00"}}); code != 200 || out["booked"] != float64(0) {
		t.Fatalf("beta's second memo: %d %v", code, out)
	}
	if ids, _ := w.s.FindObjectIDsByField(betas, "note", "text", "beta's first memo"); len(ids) != 1 {
		t.Fatalf("what the lesson booked before stays: %v", ids)
	}

	// The office reads why in beta's books, and cannot teach it again.
	retired, err := w.s.EventsOfTypes(betas, []string{"rule.retired"})
	if err != nil || len(retired) != 1 || retired[0].Actor != "beta" || !strings.Contains(string(retired[0].Payload), "we book memos by hand") {
		t.Fatalf("the veto in beta's books: %v %+v", err, retired)
	}
	if code, _ := o.teach(b); code == 200 {
		t.Fatal("the office taught beta again what beta switched off")
	}
}

// Only the office's own people make a company someone's office (KK,
// 2026-10-23): olek owns the office, which keeps client d under a full
// mandate. Through that mandate he may still manage d's people, but he may not
// make d the office of a new company — a right held through a mandate is not
// passed on (KSeF's default). Making his own office the new company's office
// is his to do.
func TestAMandateIsNotPassedOn(t *testing.T) {
	o := newOfficeScene(t)
	w, d, z := o.w, "PL8020005678", "PL8030009012"
	w.person("Olek Owner", "olek", o.office, "company_manager")
	w.mustDoor("", d, "grant_mandate", map[string]any{"office": o.office, "client": d, "scope": "*", "from": "2026-01-01", "granted_by": "operator"})
	mandate := func(office string) (int, map[string]any) {
		t.Helper()
		return w.door("olek", z, "grant_mandate", map[string]any{"office": office, "client": z, "scope": "door:* view:* type:*",
			"from": w.today, "granted_by": "olek"})
	}
	if code, out := mandate(d); code != 403 {
		t.Fatalf("olek making his client d the office of a new company: %d %v", code, out)
	}
	if code, out := mandate(o.office); code != 200 {
		t.Fatalf("olek making his own office the new company's office: %d %v", code, out)
	}
	// Under the full mandate he still manages d's people.
	w.mustDoor("", "", "register_person", map[string]any{"name": "Dorota Pracownik", "handle": "dorota", "role": "reader"})
	if code, out := w.door("olek", d, "grant_membership", map[string]any{"person": w.personID("dorota"), "company": d, "role": "user",
		"from": w.today, "granted_by": "olek"}); code != 200 {
		t.Fatalf("olek onboarding d's employee under the mandate: %d %v", code, out)
	}
}
