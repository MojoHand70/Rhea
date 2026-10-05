package core_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"rhea/internal/core"
)

func submitEvent() core.Activity {
	for _, a := range core.BuiltinActivities() {
		if a.Name == "submit_event" {
			a.Version = 1
			return a
		}
	}
	panic("submit_event not builtin")
}

func TestActivityValidate(t *testing.T) {
	base := core.Activity{Name: "register_complaint", Domain: "crm", Spec: core.ActivitySpec{
		Inputs: []core.FieldDef{
			{Name: "customer", Type: "string", Required: true},
			{Name: "details", Type: "string"},
		},
		Emits: "complaint.registered",
	}}
	if err := base.Validate(); err != nil {
		t.Fatalf("sound activity refused: %v", err)
	}

	bad := []struct {
		name   string
		mutate func(a *core.Activity)
	}{
		{"no emits", func(a *core.Activity) { a.Spec.Emits = "" }},
		{"duplicate input", func(a *core.Activity) {
			a.Spec.Inputs = append(a.Spec.Inputs, core.FieldDef{Name: "customer", Type: "string"})
		}},
		{"unknown input type", func(a *core.Activity) { a.Spec.Inputs[0].Type = "blob" }},
		{"enum without values", func(a *core.Activity) { a.Spec.Inputs[0].Type = "enum" }},
		{"emits source not declared", func(a *core.Activity) { a.Spec.Emits = "=missing" }},
		{"emits source not required string", func(a *core.Activity) {
			a.Spec.Emits = "=details" // declared, but optional
		}},
		{"occurred_at not a date", func(a *core.Activity) {
			a.Spec.Inputs = append(a.Spec.Inputs, core.FieldDef{Name: "occurred_at", Type: "string"})
		}},
		{"two json inputs", func(a *core.Activity) {
			a.Spec.Inputs = []core.FieldDef{{Name: "a", Type: "json"}, {Name: "b", Type: "json"}}
		}},
		{"passthrough with payload chatter", func(a *core.Activity) {
			// A passthrough activity has nothing else to say about the payload.
			a.Spec.Inputs = []core.FieldDef{
				{Name: "payload", Type: "json", Required: true},
				{Name: "note", Type: "string"},
			}
		}},
	}
	for _, tc := range bad {
		a := base
		a.Spec.Inputs = append([]core.FieldDef{}, base.Spec.Inputs...)
		tc.mutate(&a)
		if err := a.Validate(); err == nil {
			t.Errorf("%s: accepted", tc.name)
		}
	}
}

// TestPassthroughIndistinguishable: an event through the open door reads
// exactly like the same fact submitted free-form — the payload is the json
// input's value itself, never nested, so rules match either identically.
func TestPassthroughIndistinguishable(t *testing.T) {
	raw := map[string]any{"customer": "ACME", "currency": "PLN",
		"lines": []any{map[string]any{"amount": "200.00"}}}
	built, err := submitEvent().BuildEvent(map[string]any{
		"event_type": "invoice.received", "occurred_at": "2026-09-15", "payload": raw,
	}, "2026-10-05")
	if err != nil {
		t.Fatal(err)
	}
	if built.Type != "invoice.received" || built.OccurredAt != "2026-09-15" {
		t.Fatalf("envelope = %q %q", built.Type, built.OccurredAt)
	}
	var got map[string]any
	if err := json.Unmarshal(built.Payload, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, raw) {
		t.Fatalf("payload transformed on the way through the door:\n%v\n%v", got, raw)
	}
}

func TestBuildEventDoorCourtesy(t *testing.T) {
	door := submitEvent()
	refuse := func(name string, in map[string]any) {
		t.Helper()
		if _, err := door.BuildEvent(in, "2026-10-05"); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	payload := map[string]any{"x": 1}
	// The open door does not speak kernel namespaces.
	refuse("forged approval", map[string]any{
		"event_type": "rule.approved", "occurred_at": "2026-10-05", "payload": payload})
	refuse("forged activation", map[string]any{
		"event_type": "activity.approved", "occurred_at": "2026-10-05", "payload": payload})
	refuse("forged materialization", map[string]any{
		"event_type": core.EventObjectMaterialized, "occurred_at": "2026-10-05", "payload": payload})
	// A typed verb is strict or it is not typed.
	refuse("undeclared input", map[string]any{
		"event_type": "x.y", "occurred_at": "2026-10-05", "payload": payload, "extra": 1})
	refuse("missing required", map[string]any{
		"event_type": "x.y", "occurred_at": "2026-10-05"})
	refuse("bad date", map[string]any{
		"event_type": "x.y", "occurred_at": "soon", "payload": payload})
	refuse("payload not an object", map[string]any{
		"event_type": "x.y", "occurred_at": "2026-10-05", "payload": "scalar"})
}

func TestInputCoercions(t *testing.T) {
	verb := core.Activity{Name: "v", Spec: core.ActivitySpec{
		Inputs: []core.FieldDef{
			{Name: "amount", Type: "money", Required: true},
			{Name: "qty", Type: "int", Required: true},
			{Name: "kind", Type: "enum", Values: []string{"a", "b"}},
		},
		Emits: "v.done",
	}}
	built, err := verb.BuildEvent(map[string]any{
		"amount": "123.45", "qty": float64(3), "kind": "a",
	}, "2026-10-05")
	if err != nil {
		t.Fatal(err)
	}
	var p map[string]any
	json.Unmarshal(built.Payload, &p)
	// Money crosses as the canonical decimal string (invariant 6) — the
	// payload is boundary data, same as a free-form submit would carry.
	if p["amount"] != "123.45" || p["qty"] != float64(3) || p["kind"] != "a" {
		t.Fatalf("payload = %v", p)
	}
	if built.OccurredAt != "2026-10-05" {
		t.Fatalf("business date not stamped: %q", built.OccurredAt)
	}

	for name, in := range map[string]map[string]any{
		"money one decimal": {"amount": "123.4", "qty": float64(3)},
		"money as number":   {"amount": 123.45, "qty": float64(3)},
		"fractional int":    {"amount": "123.45", "qty": 3.5},
		"enum off the list": {"amount": "123.45", "qty": float64(3), "kind": "c"},
	} {
		if _, err := verb.BuildEvent(in, "2026-10-05"); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestDeclaredActivitiesKeepOutOfKernelNamespaces(t *testing.T) {
	for _, reserved := range []string{"rule.approved", "activity.approved", "rule.draft_requested"} {
		if !core.ReservedEventType(reserved) {
			t.Errorf("%s not reserved", reserved)
		}
	}
	if core.ReservedEventType("invoice.received") {
		t.Error("business namespace reserved")
	}
	// The builtins themselves validate — the retrofit's first claim: the
	// shell's own verbs are expressible as the data every future verb is.
	for _, a := range core.BuiltinActivities() {
		if err := a.Validate(); err != nil {
			t.Errorf("builtin %s: %v", a.Name, err)
		}
		if !strings.HasPrefix(a.Spec.Emits, "=") && a.Name != "submit_event" && !core.ReservedEventType(a.Spec.Emits) {
			t.Errorf("builtin %s emits %q outside the kernel namespaces", a.Name, a.Spec.Emits)
		}
	}
}
