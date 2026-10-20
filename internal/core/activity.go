package core

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
)

// Activity is a declared verb (SPEC §2): a named action a human or agent can
// take — typed inputs, the raw event it emits, who may trigger it. The schema
// belongs to the verb, not the event (DIRECTION 2026-10-05): inputs validate
// at trigger time, before the event exists; the emitted event is a raw event
// like any other — schemaless in the log, still needing a rule to explain it —
// and activities are never consulted during replay. Activities have the rule
// lifecycle, where `active` means *offered and triggerable*, never *executes
// in replay*: superseding an activity never touches the events it emitted.
type Activity struct {
	Name        string       `json:"name"`
	Version     int          `json:"version"`
	Status      string       `json:"status"` // draft | approved | active | superseded
	Domain      string       `json:"domain"`
	Description string       `json:"description"`
	Spec        ActivitySpec `json:"spec"`
	CreatedBy   string       `json:"created_by"`
	CreatedAt   time.Time    `json:"created_at"`
}

type ActivitySpec struct {
	Inputs []FieldDef `json:"inputs"`
	// Emits is the raw event type the activity appends: a literal, or the
	// passthrough form "=<input>" taking the type from a string input — the
	// open door declared as data (submit_event).
	Emits string `json:"emits"`
	// Who may trigger: actor-pattern strings ("human", "agent"), declared and
	// displayed, never enforced — the seed of authz-in-the-language, which
	// stays a falsifiability test of its own (DIRECTION). Attribution is the
	// actor column; this field is vocabulary.
	Who []string `json:"who,omitempty"`
	// Channel makes the activity a channel: a door code triggers instead of
	// a person. Nil for every ordinary verb.
	Channel *ChannelSpec `json:"channel,omitempty"`
}

// ChannelSpec is the registry entry of one channel (DIRECTION 2026-10-10,
// "Documents in"): the registry of channels is the activity store itself —
// versioned, approved like a rule — and no seventh concept. A channel
// declares what it binds to (an adapter the kernel has code for, at a code
// version), what it reads, what makes a document the same one twice, how
// sure its reader is, and which event types it may append. The runner checks
// the binding before every pass and stamps every event the channel appends
// with the channel's name and version: invariant 5 reaches the raw side.
type ChannelSpec struct {
	// Adapter names the code the channel binds to; CodeVersion is the code
	// version the kernel must carry — a mismatch refuses the channel.
	Adapter     string `json:"adapter"`
	CodeVersion int    `json:"code_version"`
	// Format is what the reader reads, in the outside world's words:
	// "FA(3)", "MT940", "PDF", "calendar".
	Format string `json:"format"`
	// Identity says what makes a document the same one twice; the reader
	// turns it into the dedup key on every event it returns, and the runner
	// refuses an event that has none.
	Identity string `json:"identity"`
	// Confidence is how sure the reader is: exact (a parser, the outside
	// system's own numbers) or a model's reading with a threshold below
	// which the reading waits for a person.
	Confidence ConfidenceSpec `json:"confidence"`
	// Events lists every event type the channel may append; Emits is one of
	// them. Empty means Emits alone.
	Events []string `json:"events,omitempty"`
	// Volume is "documents" (the default) or "stream": a stream channel is
	// batched and its residue counted by shape — the high-volume treatment,
	// declared now, honored as the treatment lands.
	Volume string `json:"volume,omitempty"`
	// Market is the market whose facts the channel delivers — the pack's,
	// since the pack declares the channel. The runner stamps it into every
	// payload the channel appends that carries none, so a pack's rules can
	// guard on it the way they guard on a document's own market (the E2
	// lesson): provenance of the raw side, declared as data, never a
	// reader's guess. Empty: the channel's facts say their market themselves.
	Market string `json:"market,omitempty"`
}

// ConfidenceSpec says how sure a channel's reader is.
type ConfidenceSpec struct {
	Policy    string `json:"policy"`              // ConfidenceExact | ConfidenceModel
	Threshold string `json:"threshold,omitempty"` // model only: a decimal in (0, 1], as a string
}

const (
	ConfidenceExact = "exact"
	ConfidenceModel = "model"

	VolumeDocuments = "documents"
	VolumeStream    = "stream"
)

// Allows reports whether a channel may append an event of this type: the
// declared emits or one of the declared events. The declaration is the
// contract; an adapter speaking outside it has a bug, not residue.
func (sp ActivitySpec) Allows(eventType string) bool {
	if sp.Channel == nil {
		return eventType == sp.Emits
	}
	if eventType == sp.Emits {
		return true
	}
	for _, t := range sp.Channel.Events {
		if t == eventType {
			return true
		}
	}
	return false
}

// validate checks a channel declaration against the activity it sits on.
func (c ChannelSpec) validate(a Activity) error {
	sp := a.Spec
	if c.Adapter == "" {
		return fmt.Errorf("channel %q: adapter is required — the code it binds to", a.Name)
	}
	if c.CodeVersion < 1 {
		return fmt.Errorf("channel %q: code_version must be at least 1", a.Name)
	}
	if c.Format == "" {
		return fmt.Errorf("channel %q: format is required — what the reader reads", a.Name)
	}
	if c.Identity == "" {
		return fmt.Errorf("channel %q: identity is required — what makes a document the same one twice", a.Name)
	}
	switch c.Confidence.Policy {
	case ConfidenceExact:
		if c.Confidence.Threshold != "" {
			return fmt.Errorf("channel %q: an exact reader has no threshold", a.Name)
		}
	case ConfidenceModel:
		t, ok := new(big.Rat).SetString(c.Confidence.Threshold)
		if !ok || t.Sign() <= 0 || t.Cmp(big.NewRat(1, 1)) > 0 {
			return fmt.Errorf("channel %q: a model reader wants a threshold in (0, 1], got %q", a.Name, c.Confidence.Threshold)
		}
	default:
		return fmt.Errorf("channel %q: confidence policy must be %q or %q", a.Name, ConfidenceExact, ConfidenceModel)
	}
	switch c.Volume {
	case "", VolumeDocuments, VolumeStream:
	default:
		return fmt.Errorf("channel %q: volume must be %q or %q", a.Name, VolumeDocuments, VolumeStream)
	}
	if len(sp.Inputs) > 0 {
		return fmt.Errorf("channel %q: a channel is triggered by code and declares no inputs", a.Name)
	}
	if _, passthrough := sp.emitsSource(); passthrough {
		return fmt.Errorf("channel %q: a channel names the event types it appends; the open door is not one", a.Name)
	}
	seen := map[string]bool{}
	for _, t := range c.Events {
		if t == "" {
			return fmt.Errorf("channel %q: an empty event type", a.Name)
		}
		if seen[t] {
			return fmt.Errorf("channel %q: event type %q listed twice", a.Name, t)
		}
		seen[t] = true
	}
	if len(c.Events) > 0 && !seen[sp.Emits] {
		return fmt.Errorf("channel %q: emits %q must be among the channel's events", a.Name, sp.Emits)
	}
	return nil
}

// ReservedEventType reports whether an event type belongs to the kernel: the
// materialization and amendment namespaces, and the rule.* / activity.*
// / bundle.* / backfill.* system-verb namespaces that only declared system activities may
// speak in.
func ReservedEventType(t string) bool {
	return t == EventObjectMaterialized || t == EventObjectAmended ||
		t == EventObjectsMatched || t == EventObjectsUnmatched ||
		strings.HasPrefix(t, "matching.") || strings.HasPrefix(t, "reading.") || strings.HasPrefix(t, "host.") ||
		strings.HasPrefix(t, "rule.") || strings.HasPrefix(t, "activity.") ||
		strings.HasPrefix(t, "bundle.") || strings.HasPrefix(t, "backfill.")
}

// emitsSource returns the input name a passthrough emits template reads:
// "=event_type" → ("event_type", true).
func (sp ActivitySpec) emitsSource() (string, bool) {
	if strings.HasPrefix(sp.Emits, "=") {
		return sp.Emits[1:], true
	}
	return "", false
}

// envelopeInput reports whether an input is consumed by the event envelope
// rather than the payload: the emits source and the business date.
func (sp ActivitySpec) envelopeInput(name string) bool {
	if src, ok := sp.emitsSource(); ok && name == src {
		return true
	}
	return name == "occurred_at"
}

// Validate checks an activity declaration. Beyond the ObjectType field
// vocabulary, inputs admit "json": the schemaless passthrough. A passthrough
// activity has nothing else to say about the payload — the json input must be
// the only payload-contributing one — so declared verbs stay typed and the
// escape hatch stays a hatch, not a loophole.
func (a Activity) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("activity needs a name")
	}
	switch a.Status {
	case StatusDraft, StatusActive, StatusSuperseded:
	case "":
	default:
		return fmt.Errorf("activity %q: unknown status %q", a.Name, a.Status)
	}
	sp := a.Spec
	if sp.Emits == "" {
		return fmt.Errorf("activity %q: emits is required", a.Name)
	}
	jsonInputs := 0
	byName := map[string]FieldDef{}
	for _, f := range sp.Inputs {
		if f.Name == "" {
			return fmt.Errorf("activity %q: input needs a name", a.Name)
		}
		if _, dup := byName[f.Name]; dup {
			return fmt.Errorf("activity %q: duplicate input %q", a.Name, f.Name)
		}
		byName[f.Name] = f
		switch f.Type {
		case "string", "int", "decimal", "date", "money", "file":
		case "enum":
			if len(f.Values) == 0 {
				return fmt.Errorf("activity %q input %q: enum needs values", a.Name, f.Name)
			}
		case "json":
			jsonInputs++
		default:
			if _, ok := RefTarget(f.Type); !ok {
				return fmt.Errorf("activity %q input %q: unknown type %q", a.Name, f.Name, f.Type)
			}
		}
		if f.Type != "enum" && len(f.Values) > 0 {
			return fmt.Errorf("activity %q input %q: values only belong on enum inputs", a.Name, f.Name)
		}
	}
	if src, ok := sp.emitsSource(); ok {
		fd, declared := byName[src]
		if !declared || fd.Type != "string" || !fd.Required {
			return fmt.Errorf("activity %q: emits %q wants a required string input %q", a.Name, sp.Emits, src)
		}
	}
	if fd, ok := byName["occurred_at"]; ok && fd.Type != "date" {
		return fmt.Errorf("activity %q: occurred_at is the business date and must be a date input", a.Name)
	}
	if jsonInputs > 1 {
		return fmt.Errorf("activity %q: at most one json input — one event, one payload", a.Name)
	}
	if jsonInputs == 1 {
		for _, f := range sp.Inputs {
			if f.Type != "json" && !sp.envelopeInput(f.Name) {
				return fmt.Errorf("activity %q: a passthrough activity has nothing else to say about the payload; input %q must go", a.Name, f.Name)
			}
		}
	}
	if sp.Channel != nil {
		return sp.Channel.validate(a)
	}
	return nil
}

// FileInput is a file handed to a door (DIRECTION 2026-10-10, "the file is
// the fact"): its bytes, which the kernel keeps in the blob store under their
// hash in the same transaction as the event, and the facts about it that
// enter the payload under the input's name — {hash, name, media_type, size}.
// The bytes never enter the payload; the hash is the file's identity.
type FileInput struct {
	Input     string
	Name      string
	MediaType string
	Bytes     []byte
	Hash      string // "sha256:" + hex of the bytes
}

// MaxFileBytes caps one file through a door — the experiment's blob store is
// a Postgres table, not an object store.
const MaxFileBytes = 32 << 20

// EventDocumentReadingConfirmed is what the confirm_reading door declares it
// emits; the kernel's reaction makes the appended event the proposed typed
// event itself, with the person as its author and the proposal as its cause.
const EventDocumentReadingConfirmed = "reading.confirmed"

// EventDocumentReceived is the fact that a file arrived: the first event of
// every document that is not born as JSON. The store refuses one whose file
// it does not hold, so the claim "a document was received" is never made
// without the document.
const EventDocumentReceived = "document.received"

// EventDocumentIssued: a file Rhea rendered was handed over (issue_document).
const EventDocumentIssued = "document.issued"

// HashBytes is the content address every file is kept and named by.
func HashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// RefInput is a declared ref input and the id a trigger handed it; existence
// and type are the caller's to verify against state — core stays pure.
type RefInput struct {
	Input      string
	TargetType string
	ObjectID   string
}

// BuiltEvent is what a validated trigger becomes: the envelope of the raw
// event the door will append. The payload carries the canonical boundary
// encoding — money as decimal strings (invariant 6), refs as object ids —
// indistinguishable from the same fact submitted free-form.
type BuiltEvent struct {
	Type       string
	OccurredAt string
	Payload    json.RawMessage
	Refs       []RefInput
	// Files are the bytes the door must keep before the event exists.
	Files []FileInput
	// Identity is the door's default dedup key when the caller gives none:
	// the hash of the one file a trigger carried. The same file through the
	// same door is the same fact.
	Identity string
}

// BuildEvent validates trigger inputs against the declaration and shapes the
// event: the door's courtesy, spent entirely before the event exists. Unknown
// inputs are refused — a typed verb is strict or it is not typed. `today` is
// the business date when the activity declares no occurred_at input.
func (a Activity) BuildEvent(in map[string]any, today string) (BuiltEvent, error) {
	sp := a.Spec
	declared := map[string]bool{}
	for _, f := range sp.Inputs {
		declared[f.Name] = true
	}
	for name := range in {
		if !declared[name] {
			return BuiltEvent{}, fmt.Errorf("activity %q has no input %q", a.Name, name)
		}
	}

	out := BuiltEvent{Type: sp.Emits, OccurredAt: today}
	payload := map[string]any{}
	for _, f := range sp.Inputs {
		v, given := in[f.Name]
		if !given || v == nil || v == "" {
			if f.Required {
				return BuiltEvent{}, fmt.Errorf("input %q is required", f.Name)
			}
			continue
		}
		cv, err := coerceInput(f, v)
		if err != nil {
			return BuiltEvent{}, err
		}
		if target, isRef := RefTarget(f.Type); isRef {
			out.Refs = append(out.Refs, RefInput{Input: f.Name, TargetType: target, ObjectID: cv.(string)})
		}
		if f.Name == "occurred_at" {
			out.OccurredAt = cv.(string)
			continue
		}
		if src, ok := sp.emitsSource(); ok && f.Name == src {
			out.Type = cv.(string)
			continue
		}
		if f.Type == "json" {
			// The passthrough: this value IS the payload, so an event through
			// the open door reads exactly like one a rule will be asked to
			// explain — never nested under the input's name.
			b, err := json.Marshal(cv)
			if err != nil {
				return BuiltEvent{}, err
			}
			out.Payload = b
			continue
		}
		if fi, ok := cv.(FileInput); ok {
			fi.Input = f.Name
			out.Files = append(out.Files, fi)
			payload[f.Name] = map[string]any{
				"hash": fi.Hash, "name": fi.Name, "media_type": fi.MediaType, "size": len(fi.Bytes),
			}
			continue
		}
		payload[f.Name] = cv
	}
	if len(out.Files) == 1 {
		out.Identity = out.Files[0].Hash
	}

	if _, ok := sp.emitsSource(); ok && ReservedEventType(out.Type) {
		return BuiltEvent{}, fmt.Errorf("event type %q is a kernel namespace — the open door does not speak it", out.Type)
	}
	if out.OccurredAt == "" {
		return BuiltEvent{}, fmt.Errorf("activity %q: no business date — declare an occurred_at input or let the door stamp today", a.Name)
	}
	if _, err := time.Parse("2006-01-02", out.OccurredAt); err != nil {
		return BuiltEvent{}, fmt.Errorf("occurred_at %q: want YYYY-MM-DD", out.OccurredAt)
	}
	if out.Payload == nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return BuiltEvent{}, err
		}
		out.Payload = b
	}
	return out, nil
}

// coerceInput checks one value against its declared type and returns it in
// the payload's canonical encoding.
func coerceInput(f FieldDef, v any) (any, error) {
	switch f.Type {
	case "string":
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("input %q wants a string, got %T", f.Name, v)
		}
		return s, nil
	case "int":
		switch n := v.(type) {
		case int:
			return int64(n), nil
		case int64:
			return n, nil
		case float64:
			if n != float64(int64(n)) {
				return nil, fmt.Errorf("input %q wants an integer, got %v", f.Name, n)
			}
			return int64(n), nil
		}
		return nil, fmt.Errorf("input %q wants an integer, got %T", f.Name, v)
	case "date":
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("input %q wants a YYYY-MM-DD date, got %T", f.Name, v)
		}
		if _, err := time.Parse("2006-01-02", s); err != nil {
			return nil, fmt.Errorf("input %q: %q is not a YYYY-MM-DD date", f.Name, s)
		}
		return s, nil
	case "money":
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("input %q wants money as a decimal string, got %T", f.Name, v)
		}
		if _, err := ParseMoney(s); err != nil {
			return nil, fmt.Errorf("input %q: %w", f.Name, err)
		}
		return s, nil // canonical boundary encoding: the string crosses, int64 lives in Go
	case "decimal":
		// a rate, a coefficient: exact, as a string across the boundary like
		// money; a JSON number is accepted and written out exactly
		switch n := v.(type) {
		case string:
			if _, err := parseDecimal(n); err != nil {
				return nil, fmt.Errorf("input %q: %q is not a decimal", f.Name, n)
			}
			return n, nil
		case float64:
			return strconv.FormatFloat(n, 'f', -1, 64), nil
		case int:
			return strconv.Itoa(n), nil
		case int64:
			return strconv.FormatInt(n, 10), nil
		}
		return nil, fmt.Errorf("input %q wants a decimal, got %T", f.Name, v)
	case "enum":
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("input %q wants one of %v, got %T", f.Name, f.Values, v)
		}
		for _, allowed := range f.Values {
			if s == allowed {
				return s, nil
			}
		}
		return nil, fmt.Errorf("input %q: %q is not one of %v", f.Name, s, f.Values)
	case "json":
		m, ok := v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("input %q wants a JSON object, got %T", f.Name, v)
		}
		return m, nil
	case "file":
		return coerceFile(f, v)
	default: // ref<T>, vetted by Validate
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("input %q wants an object id, got %T", f.Name, v)
		}
		return s, nil
	}
}

// coerceFile reads a file input: {name, media_type, content_base64} over the
// wire, or {name, media_type, bytes} from a caller in the same process. The
// hash is computed here, never trusted from the caller.
func coerceFile(f FieldDef, v any) (any, error) {
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("input %q wants a file {name, media_type, content_base64}, got %T", f.Name, v)
	}
	name, _ := m["name"].(string)
	if name == "" {
		return nil, fmt.Errorf("input %q: a file needs a name", f.Name)
	}
	media, _ := m["media_type"].(string)
	if media == "" {
		media = "application/octet-stream"
	}
	var data []byte
	switch c := m["bytes"].(type) {
	case []byte:
		data = c
	case nil:
		enc, _ := m["content_base64"].(string)
		if enc == "" {
			return nil, fmt.Errorf("input %q: a file needs content", f.Name)
		}
		var err error
		if data, err = base64.StdEncoding.DecodeString(enc); err != nil {
			return nil, fmt.Errorf("input %q: content_base64: %w", f.Name, err)
		}
	default:
		return nil, fmt.Errorf("input %q: bytes must be []byte, got %T", f.Name, c)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("input %q: an empty file is not a document", f.Name)
	}
	if len(data) > MaxFileBytes {
		return nil, fmt.Errorf("input %q: %d bytes is over the %d-byte cap", f.Name, len(data), MaxFileBytes)
	}
	return FileInput{Name: name, MediaType: media, Bytes: data, Hash: HashBytes(data)}, nil
}

// BuiltinActivities are the kernel's own verbs, seeded active at init with
// their activation events in the log — the gate's own birth is recorded. The
// retrofit test (DIRECTION 2026-10-05): the shell's hardcoded verbs expressed
// as the same declared data every future verb will be.
func BuiltinActivities() []Activity {
	return []Activity{
		{
			Name: "submit_event", Domain: "system",
			Description: "The open door: submit any raw event. Fail-as-data depends on it — what no rule explains yet waits in the worklist, and the residue is the completeness gauge.",
			Spec: ActivitySpec{
				Inputs: []FieldDef{
					{Name: "event_type", Type: "string", Required: true},
					{Name: "occurred_at", Type: "date", Required: true},
					{Name: "payload", Type: "json", Required: true},
				},
				Emits: "=event_type",
				Who:   []string{"human", "agent"},
			},
		},
		{
			Name: "receive_document", Domain: "system",
			Description: "The door for documents: the file itself is the fact. Its bytes are kept under their hash, the event names them, and a reader — a parser or the model — turns the file into the typed event later. The same file through this door twice is the same fact.",
			Spec: ActivitySpec{
				Inputs: []FieldDef{
					{Name: "file", Type: "file", Required: true},
					{Name: "source", Type: "string", Required: true}, // where it came from, in the person's words: email, scan, a bank's site
					{Name: "occurred_at", Type: "date"},
				},
				Emits: EventDocumentReceived,
				Who:   []string{"human", "agent"},
			},
		},
		{
			Name: "issue_document", Domain: "system",
			Description: "The door for documents going out: a file Rhea rendered from its objects — a filing, a statement — handed to a person or an office. The writer and its parameters say what was rendered, the bytes are kept under their hash, and the event is the evidence of what was handed over, sealed by the hash. A correction is a new file, a new fact.",
			Spec: ActivitySpec{
				Inputs: []FieldDef{
					{Name: "writer", Type: "string", Required: true},
					{Name: "params", Type: "string", Required: true},
					{Name: "file", Type: "file", Required: true},
					{Name: "occurred_at", Type: "date"},
				},
				Emits: EventDocumentIssued,
				Who:   []string{"human", "agent"},
			},
		},
		{
			Name: "approve_rule", Domain: "system",
			Description: "Activate a draft rule. The approval is recorded as rule.approved (invariant 2); the kernel reacts by writing the new active version row.",
			Spec: ActivitySpec{
				Inputs: []FieldDef{
					{Name: "rule_id", Type: "string", Required: true}, // a definition ref; FieldDef refs point at ObjectTypes (recorded contortion, exit at ref-to-definition)
					{Name: "approved_by", Type: "string", Required: true},
					{Name: "occurred_at", Type: "date", Required: true},
					// Why this version replaces the reading before it — the
					// correction's reason, on the record (the host remembers).
					{Name: "reason", Type: "string"},
				},
				Emits: "rule.approved",
				Who:   []string{"human"},
			},
		},
		{
			Name: "approve_activity", Domain: "system",
			Description: "Activate a draft activity — the gate applied to the gate; nothing in Rhea needs a second governance mechanism.",
			Spec: ActivitySpec{
				Inputs: []FieldDef{
					{Name: "activity", Type: "string", Required: true},
					{Name: "approved_by", Type: "string", Required: true},
					{Name: "occurred_at", Type: "date", Required: true},
				},
				Emits: "activity.approved",
				Who:   []string{"human"},
			},
		},
		{
			Name: "approve_bundle", Domain: "system",
			Description: "Activate a bundle: every member draft — rules, activities, object types, views — becomes active in one transaction, or none does. Recorded as bundle.approved; there is no partial approval (reject and redraft instead).",
			Spec: ActivitySpec{
				Inputs: []FieldDef{
					{Name: "bundle_id", Type: "string", Required: true},
					{Name: "approved_by", Type: "string", Required: true},
					{Name: "occurred_at", Type: "date", Required: true},
				},
				Emits: "bundle.approved",
				Who:   []string{"human"},
			},
		},
		{
			Name: "reject_bundle", Domain: "system",
			Description: "Reject a bundle with a reason: it and its member drafts are superseded, never activated. The reason stays on the record — a human correcting the agent is the richest signal the network gets.",
			Spec: ActivitySpec{
				Inputs: []FieldDef{
					{Name: "bundle_id", Type: "string", Required: true},
					{Name: "reason", Type: "string", Required: true},
					{Name: "rejected_by", Type: "string", Required: true},
					{Name: "occurred_at", Type: "date", Required: true},
				},
				Emits: "bundle.rejected",
				Who:   []string{"human"},
			},
		},
		{
			Name: "approve_backfill", Domain: "system",
			Description: "Let the active rules explain the past further: every past event gets the consequences approved rules would have given it, and nothing already explained is rewritten. A consequence bound for a closed period is booked on this date instead, linked to its original event. Recorded as backfill.approved, naming every chain.",
			Spec: ActivitySpec{
				Inputs: []FieldDef{
					{Name: "approved_by", Type: "string", Required: true},
					{Name: "occurred_at", Type: "date", Required: true},
				},
				Emits: "backfill.approved",
				Who:   []string{"human"},
			},
		},
		{
			Name: "draft_bundle", Domain: "system",
			Description: "Ask for a bundle in plain language: the types, rules, verbs and views one answer needs. Recorded as bundle.draft_requested; the agent answers as an actor outside the kernel, and its bundle waits for approval.",
			Spec: ActivitySpec{
				Inputs: []FieldDef{
					{Name: "intent", Type: "string", Required: true},
					{Name: "sample_event_id", Type: "int"},
					{Name: "after", Type: "string"}, // the rejected bundle this ask redrafts
				},
				Emits: "bundle.draft_requested",
				Who:   []string{"human"},
			},
		},
		{
			Name: "clock", Domain: "system",
			Description: "Time passing as facts: the clock opens days and months and fires due schedules. The first channel — code triggers it, never a person — with the burst gate as its switch (M5).",
			Spec: ActivitySpec{
				Emits: EventDayOpened,
				Who:   []string{"adapter"},
				Channel: &ChannelSpec{
					Adapter: "clock", CodeVersion: 1, Format: "calendar",
					Identity:   "one event per opened day, per opened month and per schedule occurrence",
					Confidence: ConfidenceSpec{Policy: ConfidenceExact},
					Events:     []string{EventDayOpened, EventMonthOpened, "schedule.fired"},
				},
			},
		},
		{
			Name: "pair_items", Domain: "system",
			Description: "Pair by hand what no strategy settled: one item of the trigger side and the items it settles on the other sides, under a matching rule's structure. The kernel runs the same engine on the selection and emits the same objects.matched event the policy would — explained the same way, whoever made it. Recorded as matching.paired.",
			Spec: ActivitySpec{
				Inputs: []FieldDef{
					{Name: "pairing", Type: "json", Required: true}, // {"matching": <rule id>, "trigger": <object id>, "items": [<object id>, …]}
					{Name: "occurred_at", Type: "date"},
				},
				Emits: EventMatchingPaired,
				Who:   []string{"human"},
			},
		},
		{
			Name: "unpair_items", Domain: "system",
			Description: "Undo one match: the open amounts it took are restored and every difference it absorbed is reversed, by the same rules that explained the match — a correcting event in the log, never a deletion. Recorded as matching.unpaired with the reason.",
			Spec: ActivitySpec{
				Inputs: []FieldDef{
					{Name: "match_event", Type: "int", Required: true}, // the objects.matched event to undo
					{Name: "reason", Type: "string", Required: true},
					{Name: "occurred_at", Type: "date"},
				},
				Emits: EventMatchingUnpaired,
				Who:   []string{"human"},
			},
		},
		{
			Name: "confirm_reading", Domain: "system",
			Description: "Confirm a document reading the channel was not sure enough of: the proposed event is appended as the person's own reading, caused by the proposal, and the document is read. Recorded as the typed event itself, through this door.",
			Spec: ActivitySpec{
				Inputs: []FieldDef{
					{Name: "proposal", Type: "int", Required: true}, // the document.reading_proposed event
					{Name: "occurred_at", Type: "date"},
				},
				Emits: EventDocumentReadingConfirmed,
				Who:   []string{"human"},
			},
		},
		{
			Name: "host_said", Domain: "system",
			Description: "What the host said: the words it phrased for a narrative view, logged as a fact so it never repeats itself and so what was said can be read back. The agent speaks through this door; the words it says never carry a number the facts did not, and never pass over a fact that has an amount in it.",
			Spec: ActivitySpec{
				Inputs: []FieldDef{
					{Name: "view", Type: "string", Required: true},
					{Name: "text", Type: "string", Required: true},
					{Name: "occurred_at", Type: "date"},
				},
				Emits: EventHostSaid,
				Who:   []string{"agent"},
			},
		},
		{
			Name: "host_answered", Domain: "system",
			Description: "A question answered from the books now: what was asked, the read-only query the author wrote, what it counts in plain words, how it was shown and the first rows as they were told. Nothing in the books changed; this is the record of what she said.",
			Spec: ActivitySpec{
				Inputs: []FieldDef{
					{Name: "question", Type: "string", Required: true},
					{Name: "sql", Type: "string", Required: true},
					{Name: "reading", Type: "string", Required: true},
					{Name: "visual", Type: "string", Required: true},
					{Name: "result", Type: "string", Required: true},
					{Name: "rows", Type: "int", Required: true},
					{Name: "asked_by", Type: "string", Required: true},
					{Name: "occurred_at", Type: "date", Required: true},
				},
				Emits: EventHostAnswered,
				Who:   []string{"human"},
			},
		},
		{
			Name: "note_untaught", Domain: "system",
			Description: "What the host was asked and could not fully answer: never taught it (she showed the nearest thing or nothing), the books do not hold it, or the query she wrote could not be run. The person's own words stay in this installation; her general description of the missing skill reaches the Rhea team when it carries no number and no name the installation holds — whether it did, and why not, is part of the fact. Rhea learns from every client what she cannot yet do, with people in the loop.",
			Spec: ActivitySpec{
				Inputs: []FieldDef{
					{Name: "asked", Type: "string", Required: true},
					{Name: "missing", Type: "string", Required: true},
					{Name: "asked_by", Type: "string", Required: true},
					{Name: "kind", Type: "enum", Values: []string{"untaught", "cannot", "failed"}},
					{Name: "detail", Type: "string"},
					{Name: "shared", Type: "enum", Values: []string{"yes", "no"}, Required: true},
					{Name: "kept_home_because", Type: "string"},
					{Name: "occurred_at", Type: "date", Required: true},
				},
				Emits: EventHostUntaught,
				Who:   []string{"human"},
			},
		},
		{
			Name: "draft_rule", Domain: "system",
			Description: "Ask for a rule in plain language. Records the request as rule.draft_requested; the agent answers as an actor outside the kernel — a reaction must be deterministic, and drafting is not.",
			Spec: ActivitySpec{
				Inputs: []FieldDef{
					{Name: "intent", Type: "string", Required: true},
					{Name: "sample_event_id", Type: "int", Required: true},
					{Name: "object_type", Type: "string", Required: true},
				},
				Emits: "rule.draft_requested",
				Who:   []string{"human"},
			},
		},
	}
}
