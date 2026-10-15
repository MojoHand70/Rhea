package core_test

import (
	"encoding/base64"
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
		// A builtin channel speaks the outside world's events — the clock's
		// time.* is deliberately not reserved (M5: forgery stays visible via
		// the actor); the kernel's own verbs stay in their namespaces.
		if a.Spec.Channel != nil {
			continue
		}
		// The three open doors — facts as JSON, facts as files in, files
		// out — speak the business namespace by design; the kernel's verbs
		// stay in theirs.
		if a.Name == "submit_event" || a.Name == "receive_document" || a.Name == "issue_document" {
			continue
		}
		if !strings.HasPrefix(a.Spec.Emits, "=") && !core.ReservedEventType(a.Spec.Emits) {
			t.Errorf("builtin %s emits %q outside the kernel namespaces", a.Name, a.Spec.Emits)
		}
	}
}

// A channel is an activity triggered by code (DIRECTION 2026-10-10): the
// declaration must say what it binds to, what it reads, what identity is,
// how sure the reader is, and it declares no inputs.
func TestChannelValidate(t *testing.T) {
	sound := func() core.Activity {
		return core.Activity{Name: "bank_mt940", Domain: "finance", Spec: core.ActivitySpec{
			Emits: "bank.statement.received",
			Channel: &core.ChannelSpec{
				Adapter: "mt940", CodeVersion: 1, Format: "MT940", Identity: "statement number and line",
				Confidence: core.ConfidenceSpec{Policy: core.ConfidenceExact},
				Events:     []string{"bank.statement.received", "bank.line.received"},
			},
		}}
	}
	if err := sound().Validate(); err != nil {
		t.Fatalf("sound channel refused: %v", err)
	}
	model := sound()
	model.Spec.Channel.Confidence = core.ConfidenceSpec{Policy: core.ConfidenceModel, Threshold: "0.95"}
	if err := model.Validate(); err != nil {
		t.Fatalf("model reader refused: %v", err)
	}
	if sp := sound().Spec; !sp.Allows("bank.line.received") || !sp.Allows("bank.statement.received") || sp.Allows("bank.fee") {
		t.Fatal("Allows does not follow the declaration")
	}
	if sp := (core.ActivitySpec{Emits: "x"}); !sp.Allows("x") || sp.Allows("y") {
		t.Fatal("a plain verb allows only what it emits")
	}

	bad := []struct {
		name   string
		mutate func(a *core.Activity)
		want   string
	}{
		{"no adapter", func(a *core.Activity) { a.Spec.Channel.Adapter = "" }, "adapter is required"},
		{"code version 0", func(a *core.Activity) { a.Spec.Channel.CodeVersion = 0 }, "code_version"},
		{"no format", func(a *core.Activity) { a.Spec.Channel.Format = "" }, "format is required"},
		{"no identity", func(a *core.Activity) { a.Spec.Channel.Identity = "" }, "identity is required"},
		{"unknown policy", func(a *core.Activity) { a.Spec.Channel.Confidence.Policy = "guess" }, "confidence policy"},
		{"exact with threshold", func(a *core.Activity) { a.Spec.Channel.Confidence.Threshold = "0.5" }, "no threshold"},
		{"model without threshold", func(a *core.Activity) {
			a.Spec.Channel.Confidence = core.ConfidenceSpec{Policy: core.ConfidenceModel}
		}, "threshold in (0, 1]"},
		{"model threshold above one", func(a *core.Activity) {
			a.Spec.Channel.Confidence = core.ConfidenceSpec{Policy: core.ConfidenceModel, Threshold: "1.5"}
		}, "threshold in (0, 1]"},
		{"bad volume", func(a *core.Activity) { a.Spec.Channel.Volume = "firehose" }, "volume must be"},
		{"inputs on a channel", func(a *core.Activity) {
			a.Spec.Inputs = []core.FieldDef{{Name: "file", Type: "string"}}
		}, "declares no inputs"},
		{"emits outside events", func(a *core.Activity) { a.Spec.Emits = "bank.fee" }, "must be among the channel's events"},
		{"event listed twice", func(a *core.Activity) {
			a.Spec.Channel.Events = append(a.Spec.Channel.Events, "bank.line.received")
		}, "listed twice"},
		{"the open door as a channel", func(a *core.Activity) {
			a.Spec.Inputs = []core.FieldDef{{Name: "event_type", Type: "string", Required: true}}
			a.Spec.Emits = "=event_type"
			a.Spec.Channel.Events = nil
		}, "declares no inputs"},
	}
	for _, tc := range bad {
		a := sound()
		tc.mutate(&a)
		err := a.Validate()
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: want %q, got %v", tc.name, tc.want, err)
		}
	}

	// The clock is the first channel, and a builtin: its declaration is sound
	// and names the three event types time arrives as.
	for _, a := range core.BuiltinActivities() {
		if a.Name == "clock" {
			if err := a.Validate(); err != nil {
				t.Fatalf("clock builtin: %v", err)
			}
			if !a.Spec.Allows(core.EventMonthOpened) || !a.Spec.Allows("schedule.fired") {
				t.Fatal("the clock channel does not declare what the clock emits")
			}
			return
		}
	}
	t.Fatal("clock is not a builtin")
}

// A file through a door: the bytes become a blob under their hash, the
// payload carries the facts about the file, and the hash is the trigger's
// default identity (DIRECTION 2026-10-10, "the file is the fact").
func TestFileInput(t *testing.T) {
	var door core.Activity
	for _, a := range core.BuiltinActivities() {
		if a.Name == "receive_document" {
			door = a
		}
	}
	if door.Name == "" {
		t.Fatal("receive_document is not a builtin")
	}
	if err := door.Validate(); err != nil {
		t.Fatal(err)
	}
	data := []byte("%PDF-1.4 not really")
	built, err := door.BuildEvent(map[string]any{
		"file":   map[string]any{"name": "fv-12.pdf", "media_type": "application/pdf", "bytes": data},
		"source": "email",
	}, "2026-10-11")
	if err != nil {
		t.Fatal(err)
	}
	want := core.HashBytes(data)
	if built.Type != core.EventDocumentReceived || built.Identity != want || len(built.Files) != 1 ||
		built.Files[0].Input != "file" || string(built.Files[0].Bytes) != string(data) || built.Files[0].Hash != want {
		t.Fatalf("built = %+v", built)
	}
	var payload map[string]any
	if err := json.Unmarshal(built.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	file, _ := payload["file"].(map[string]any)
	if file["hash"] != want || file["name"] != "fv-12.pdf" || file["media_type"] != "application/pdf" ||
		file["size"] != float64(len(data)) || payload["source"] != "email" {
		t.Fatalf("payload = %v", payload)
	}
	if strings.Contains(string(built.Payload), "not really") {
		t.Fatal("the bytes leaked into the payload")
	}

	// Over the wire the content is base64; the hash is computed, never trusted.
	enc := base64.StdEncoding.EncodeToString(data)
	built2, err := door.BuildEvent(map[string]any{
		"file":   map[string]any{"name": "fv-12.pdf", "content_base64": enc, "hash": "sha256:forged"},
		"source": "email",
	}, "2026-10-11")
	if err != nil || built2.Identity != want || built2.Files[0].MediaType != "application/octet-stream" {
		t.Fatalf("base64 file: %+v %v", built2, err)
	}

	for name, in := range map[string]any{
		"not an object": "fv-12.pdf",
		"no name":       map[string]any{"content_base64": enc},
		"no content":    map[string]any{"name": "x.pdf"},
		"empty":         map[string]any{"name": "x.pdf", "bytes": []byte{}},
		"bad base64":    map[string]any{"name": "x.pdf", "content_base64": "***"},
		"too big":       map[string]any{"name": "x.pdf", "bytes": make([]byte, core.MaxFileBytes+1)},
	} {
		if _, err := door.BuildEvent(map[string]any{"file": in, "source": "scan"}, "2026-10-11"); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	// A document without a stated source is not received: the person says
	// where it came from.
	if _, err := door.BuildEvent(map[string]any{"file": map[string]any{"name": "x.pdf", "bytes": data}}, "2026-10-11"); err == nil ||
		!strings.Contains(err.Error(), `"source" is required`) {
		t.Fatalf("no source: %v", err)
	}

	// Two files, no default identity: the caller names the fact.
	two := core.Activity{Name: "attach_pair", Spec: core.ActivitySpec{Emits: "pair.attached", Inputs: []core.FieldDef{
		{Name: "front", Type: "file", Required: true}, {Name: "back", Type: "file", Required: true}}}}
	built3, err := two.BuildEvent(map[string]any{
		"front": map[string]any{"name": "a", "bytes": []byte("a")},
		"back":  map[string]any{"name": "b", "bytes": []byte("b")},
	}, "2026-10-11")
	if err != nil || len(built3.Files) != 2 || built3.Identity != "" {
		t.Fatalf("two files: %+v %v", built3, err)
	}
}
