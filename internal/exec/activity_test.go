package exec_test

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

// TestShellVerbsAreDeclaredDoors is the retrofit falsifiability test
// (DIRECTION 2026-10-05): the shell's own verbs — submit event, approve
// rule, draft rule — expressed as the same declared activity data every
// future verb will be. If the shape could not carry them, it would be wrong.
func TestShellVerbsAreDeclaredDoors(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}

	// The gate's birth is in the log: init seeded the builtins active, one
	// activation event each, stamped with the door that approves activities —
	// the bootstrap fixed point, recorded honestly.
	acts, err := x.Store.ActiveActivities(ctx)
	if err != nil {
		t.Fatal(err)
	}
	offered := map[string]core.Activity{}
	for _, a := range acts {
		offered[a.Name] = a
	}
	for _, name := range []string{"submit_event", "approve_rule", "approve_activity", "draft_rule"} {
		a, ok := offered[name]
		if !ok || a.Version != 1 || a.Status != core.StatusActive {
			t.Fatalf("builtin %s not offered: %+v", name, a)
		}
	}
	raws, _ := x.Store.EventsByKind(ctx, core.KindRaw)
	births := 0
	for _, ev := range raws {
		if ev.Type == "activity.approved" {
			births++
			if ev.ActivityName != "approve_activity" || ev.ActivityVersion != 1 {
				t.Fatalf("activation event %d names door %q v%d", ev.ID, ev.ActivityName, ev.ActivityVersion)
			}
		}
	}
	if births != len(core.BuiltinActivities()) {
		t.Fatalf("activation events = %d, want %d", births, len(core.BuiltinActivities()))
	}
	// System verbs are not business events: the worklist stays empty.
	if wl, _ := x.Store.UnmatchedRawEvents(ctx); len(wl) != 0 {
		t.Fatalf("worklist holds system events: %d", len(wl))
	}

	// submit_event is a door, never the door: the same fact through the open
	// door and appended free-form (an adapter's path) differ only in the
	// stamp — the payloads are indistinguishable and rules fire on both.
	seed(t, x)
	var payload map[string]any
	json.Unmarshal([]byte(invoicePayload), &payload)
	doored, err := x.TriggerActivity(ctx, "submit_event", map[string]any{
		"event_type": "invoice.received", "occurred_at": "2026-09-15", "payload": payload,
	}, "krzysztof", "door-1")
	if err != nil {
		t.Fatal(err)
	}
	freeID, err := x.Store.AppendEvent(ctx, core.Event{
		Kind: core.KindRaw, Type: "invoice.received", OccurredAt: "2026-09-15",
		Payload: json.RawMessage(invoicePayload), DedupKey: "free-1", Actor: "adapter:test",
	})
	if err != nil {
		t.Fatal(err)
	}
	dooredEv, _ := x.Store.GetEvent(ctx, doored.EventID)
	freeEv, _ := x.Store.GetEvent(ctx, freeID)
	if dooredEv.ActivityName != "submit_event" || dooredEv.ActivityVersion != 1 {
		t.Fatalf("door stamp missing: %+v", dooredEv)
	}
	if freeEv.ActivityName != "" {
		t.Fatalf("free append grew a stamp: %+v", freeEv)
	}
	var a, b map[string]any
	json.Unmarshal(dooredEv.Payload, &a)
	json.Unmarshal(freeEv.Payload, &b)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("the door transformed the fact:\n%v\n%v", a, b)
	}

	// approve_rule through its door: the approval event lands first, stamped,
	// and the kernel reaction writes the new active version row in the same
	// transaction; the enlarged rule set books both pending invoices.
	draftRule(t, x)
	r, booked, procErrs, err := x.ApproveRule(ctx, "book-pln-invoice", "krzysztof", "2026-09-15")
	if err != nil || len(procErrs) > 0 {
		t.Fatalf("approve: %v %v", err, procErrs)
	}
	if r.Version != 2 || r.Status != core.StatusActive || booked != 2 {
		t.Fatalf("approve: v%d %s booked=%d", r.Version, r.Status, booked)
	}
	raws, _ = x.Store.EventsByKind(ctx, core.KindRaw)
	var approval core.Event
	for _, ev := range raws {
		if ev.Type == "rule.approved" {
			approval = ev
		}
	}
	if approval.ActivityName != "approve_rule" || approval.ActivityVersion != 1 {
		t.Fatalf("approval names door %q v%d", approval.ActivityName, approval.ActivityVersion)
	}
	var ap struct {
		RuleID          string `json:"rule_id"`
		ApprovedVersion int    `json:"approved_version"`
		ApprovedBy      string `json:"approved_by"`
	}
	json.Unmarshal(approval.Payload, &ap)
	if ap.RuleID != "book-pln-invoice" || ap.ApprovedVersion != 2 || ap.ApprovedBy != "krzysztof" {
		t.Fatalf("approval payload = %+v", ap)
	}
	// Only drafts approve; the door refuses before any event lands.
	before := len(raws)
	if _, _, _, err := x.ApproveRule(ctx, "book-pln-invoice", "krzysztof", "2026-09-15"); err == nil {
		t.Fatal("re-approval accepted")
	}
	raws, _ = x.Store.EventsByKind(ctx, core.KindRaw)
	if len(raws) != before {
		t.Fatal("a refused approval still appended an event")
	}

	// draft_rule records the ask and nothing more: the request is in the log,
	// stamped, out of the human worklist — and no reaction drafts anything,
	// because reactions are deterministic kernel code and the agent is not.
	rulesBefore, _ := x.Store.LatestRules(ctx)
	req, err := x.TriggerActivity(ctx, "draft_rule", map[string]any{
		"intent": "PLN invoices become invoice documents", "sample_event_id": doored.EventID,
		"object_type": "invoice",
	}, "krzysztof", "")
	if err != nil {
		t.Fatal(err)
	}
	reqEv, _ := x.Store.GetEvent(ctx, req.EventID)
	if reqEv.Type != "rule.draft_requested" || reqEv.ActivityName != "draft_rule" {
		t.Fatalf("request event = %+v", reqEv)
	}
	if wl, _ := x.Store.UnmatchedRawEvents(ctx); len(wl) != 0 {
		t.Fatalf("the ask leaked into the worklist")
	}
	if rulesAfter, _ := x.Store.LatestRules(ctx); len(rulesAfter) != len(rulesBefore) {
		t.Fatal("a kernel reaction drafted a rule — drafting must stay outside the kernel")
	}

	// The forgery guards hold at both layers: the open door refuses kernel
	// namespaces, and a direct unstamped append into them is refused too.
	if _, err := x.TriggerActivity(ctx, "submit_event", map[string]any{
		"event_type": "rule.approved", "occurred_at": "2026-09-15", "payload": payload,
	}, "krzysztof", ""); err == nil {
		t.Fatal("the open door spoke a kernel namespace")
	}
	if _, err := x.Store.AppendEvent(ctx, core.Event{
		Kind: core.KindRaw, Type: "rule.approved", OccurredAt: "2026-09-15",
		Payload: json.RawMessage(`{}`),
	}); err == nil {
		t.Fatal("an unstamped system verb was appended")
	}

	// Replay with stamped events in the log reproduces identical state:
	// activities are not in the determinism path (invariant 4 untouched).
	objsBefore, _ := x.Store.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	objsAfter, _ := x.Store.AllObjects(ctx)
	bj, _ := json.Marshal(objsBefore)
	aj, _ := json.Marshal(objsAfter)
	if string(bj) != string(aj) {
		t.Fatal("replay diverged with activity stamps in the log")
	}
}

// TestActivityLifecycleGate: a declared verb arrives as a draft (the pack
// path), is not offered until a human approves it through approve_activity —
// the gate applied to the gate — and then emits stamped events like any
// builtin. Superseding is a new version row, never an edit.
func TestActivityLifecycleGate(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}

	draft := core.Activity{
		Name: "register_complaint", Status: core.StatusDraft, Domain: "crm",
		Description: "register a customer complaint", CreatedBy: "pack:test",
		Spec: core.ActivitySpec{
			Inputs: []core.FieldDef{
				{Name: "customer", Type: "string", Required: true},
				{Name: "severity", Type: "enum", Values: []string{"low", "high"}},
				{Name: "claimed", Type: "money"},
			},
			Emits: "complaint.registered",
			Who:   []string{"human"},
		},
	}
	if _, err := x.Store.InsertActivityVersion(ctx, draft); err != nil {
		t.Fatal(err)
	}
	// A draft is not a door.
	if _, err := x.TriggerActivity(ctx, "register_complaint", map[string]any{
		"customer": "ACME",
	}, "krzysztof", ""); err == nil {
		t.Fatal("a draft activity was offered")
	}
	// Declared activities do not speak kernel namespaces.
	evil := draft
	evil.Name = "fake_gate"
	evil.Spec.Emits = "rule.approved"
	if _, err := x.Store.InsertActivityVersion(ctx, evil); err == nil {
		t.Fatal("a declared activity claimed a kernel namespace")
	}

	a, _, _, err := x.ApproveActivity(ctx, "register_complaint", "krzysztof", "2026-10-05")
	if err != nil {
		t.Fatal(err)
	}
	if a.Version != 2 || a.Status != core.StatusActive {
		t.Fatalf("activation: v%d %s", a.Version, a.Status)
	}

	tr, err := x.TriggerActivity(ctx, "register_complaint", map[string]any{
		"customer": "ACME", "severity": "high", "claimed": "250.00",
	}, "krzysztof", "")
	if err != nil {
		t.Fatal(err)
	}
	ev, _ := x.Store.GetEvent(ctx, tr.EventID)
	if ev.Type != "complaint.registered" || ev.ActivityName != "register_complaint" || ev.ActivityVersion != 2 {
		t.Fatalf("event = %+v", ev)
	}
	var p map[string]any
	json.Unmarshal(ev.Payload, &p)
	if p["customer"] != "ACME" || p["severity"] != "high" || p["claimed"] != "250.00" {
		t.Fatalf("payload = %v", p)
	}
	// Unexplained like any business event: the verb exists, its explanation
	// does not yet — fail-as-data, one layer up.
	wl, _ := x.Store.UnmatchedRawEvents(ctx)
	if len(wl) != 1 || wl[0].ID != tr.EventID {
		t.Fatalf("worklist = %v", wl)
	}
}

// TestRefInputsVouchForObjects: a ref input is the door's courtesy — the
// named object must exist and be what the declaration says. This is the
// shape cases will lean on: a case's resolving verbs take a ref to the case.
func TestRefInputsVouchForObjects(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	seed(t, x)
	if _, err := x.Store.AppendEvent(ctx, core.Event{
		Kind: core.KindRaw, Type: "invoice.received", OccurredAt: "2026-09-15",
		Payload: json.RawMessage(invoicePayload), DedupKey: "inv-1",
	}); err != nil {
		t.Fatal(err)
	}
	draftRule(t, x)
	if _, _, _, err := x.ApproveRule(ctx, "book-pln-invoice", "krzysztof", "2026-09-15"); err != nil {
		t.Fatal(err)
	}
	objs, _ := x.Store.ObjectsByType(ctx, "invoice")
	if len(objs) != 1 {
		t.Fatalf("objects = %v", objs)
	}

	verb := core.Activity{
		Name: "dispute_invoice", Status: core.StatusDraft, Domain: "finance",
		Description: "dispute an invoice", CreatedBy: "test",
		Spec: core.ActivitySpec{
			Inputs: []core.FieldDef{
				{Name: "invoice", Type: "ref<invoice>", Required: true},
				{Name: "reason", Type: "string", Required: true},
			},
			Emits: "invoice.disputed",
		},
	}
	if _, err := x.Store.InsertActivityVersion(ctx, verb); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := x.ApproveActivity(ctx, "dispute_invoice", "krzysztof", "2026-10-05"); err != nil {
		t.Fatal(err)
	}

	if _, err := x.TriggerActivity(ctx, "dispute_invoice", map[string]any{
		"invoice": "invoice-999", "reason": "no such thing",
	}, "krzysztof", ""); err == nil || !strings.Contains(err.Error(), "no object") {
		t.Fatalf("dangling ref accepted: %v", err)
	}
	tr, err := x.TriggerActivity(ctx, "dispute_invoice", map[string]any{
		"invoice": objs[0].ID, "reason": "wrong amount",
	}, "krzysztof", "")
	if err != nil {
		t.Fatal(err)
	}
	ev, _ := x.Store.GetEvent(ctx, tr.EventID)
	var p map[string]any
	json.Unmarshal(ev.Payload, &p)
	if p["invoice"] != objs[0].ID {
		t.Fatalf("payload = %v", p)
	}
}
