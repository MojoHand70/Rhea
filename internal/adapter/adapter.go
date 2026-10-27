// Package adapter holds statutory adapters (SPEC M3) and declares their
// contract — the answer to SPEC §7's parked question:
//
//   - An adapter is a user, not a kernel extension. It reads projected
//     objects and acts only by returning raw events; the runner appends them
//     under actor "adapter:<name>". The kernel never calls out, and replay
//     never re-runs an adapter — its events are already in the log.
//   - Async by construction: a pass is a poll over the projected world for
//     work the outside world still owes the log. Pending statutory state is
//     ordinary objects, materialized by ordinary (pack) rules.
//   - Retry is free: returned events carry deterministic dedup keys, so a
//     crashed or repeated pass re-submits as a no-op.
//   - Evidence lives in the event payload (a KSeF reference, an UPO) — the
//     append-only log is the evidence store.
//
// An adapter speaks through a channel (DIRECTION 2026-10-10, "Documents in"):
// an activity that code triggers instead of a person, declared as data,
// versioned and approved like a rule. The binary's registry says what Rhea
// can read; the active channel declarations say what this company has
// connected; Bind joins the two before every pass, and every event a channel
// appends is stamped with the channel's name and version.
package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store"
)

// Reads is the slice of the read side an adapter may see: projected objects
// and the stored files they name. *store.Store satisfies it; adapters get
// no writes.
type Reads interface {
	ObjectsByType(ctx context.Context, typ string) ([]core.Object, error)
	GetBlob(ctx context.Context, hash string) (store.Blob, error)
}

// readers are the adapters that need no client to run — the readers of the
// document store — by name, so a corpus or a CLI verb can name one.
var readers = map[string]func() Adapter{}

// Provide registers an adapter the binary can run by name — a reader of the
// document store, or a channel out whose client is built from the
// environment — so a corpus, the CLI or the shell can name it without
// knowing the market it belongs to. Markets call it from their own package
// at init (markets/<cc>); the kernel never names one.
func Provide(name string, mk func() Adapter) { readers[name] = mk }

// Lookup returns a provided adapter by its name.
func Lookup(name string) (Adapter, bool) {
	mk, ok := readers[name]
	if !ok {
		return nil, false
	}
	return mk(), true
}

// Adapter is the declared contract: inspect the world, talk to the outside
// through your own client, return the raw events the log is now owed. A pass
// must be idempotent — returning the same (dedup-keyed) events twice is fine.
// Name and CodeVersion are what a channel declaration binds to.
type Adapter interface {
	Name() string
	CodeVersion() int
	Pass(ctx context.Context, reads Reads) ([]core.Event, error)
}

// Acting is an adapter that says which types it acts on — the invoices it
// sends, the submissions it reads back. In a company's pass it sees of
// those the company's own objects alone, so the installation's own are
// never acted on in a company's name; every other type it only consults,
// and sees as the store shows the company: its rows and the installation's
// (the pack's VAT rates). An adapter that says nothing sees every type as
// its own — a document read from every company's pass would be booked for
// each — and a master record it needs is then missing, visibly.
type Acting interface {
	ActsOn() []string
}

// Shared is an adapter whose facts are the installation's whoever runs it:
// published master data every company reads, a central bank's rates. Its
// pass runs once, for the installation — fanned out per company, a
// company's pass stamped the rates its own and a company's lookup of a
// day's rate could find two (review 2026-10-21, F2).
type Shared interface {
	Installation() bool
}

// Implemented is one adapter the binary carries: the "implemented" side of
// the channel registry. The "connected" side is the active channel
// declarations in the activity store (core.ChannelSpec).
type Implemented struct {
	Name        string `json:"name"`
	CodeVersion int    `json:"code_version"`
	Format      string `json:"format"`
}

var implemented = map[string]Implemented{}

// Register declares an adapter the binary carries, from the adapter's own
// file at init. Registering the same adapter twice is harmless; registering
// a different version under a known name is a programming error.
func Register(a Implemented) {
	if prev, ok := implemented[a.Name]; ok && prev != a {
		panic(fmt.Sprintf("adapter %s registered twice with different versions: %+v and %+v", a.Name, prev, a))
	}
	implemented[a.Name] = a
}

// ImplementedAdapters lists what Rhea can read, by name.
func ImplementedAdapters() []Implemented {
	out := make([]Implemented, 0, len(implemented))
	for _, a := range implemented {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Bind finds the channel a running adapter may speak through: the active
// channel declaration bound to this adapter name, at this code version. No
// declaration, a draft, two declarations, or a version mismatch refuse the
// pass before anything is read — the registry's check, run every time code
// wants to append.
func Bind(ctx context.Context, s *store.Store, name string, codeVersion int) (core.Activity, error) {
	acts, err := s.ActiveActivities(ctx)
	if err != nil {
		return core.Activity{}, err
	}
	var bound []core.Activity
	for _, a := range acts {
		if a.Spec.Channel != nil && a.Spec.Channel.Adapter == name {
			bound = append(bound, a)
		}
	}
	switch len(bound) {
	case 0:
		return core.Activity{}, fmt.Errorf("adapter %s has no active channel — declare one (a pack ships it as an activity) and approve it", name)
	case 1:
	default:
		names := make([]string, len(bound))
		for i, a := range bound {
			names[i] = a.Name
		}
		return core.Activity{}, fmt.Errorf("adapter %s is bound by %d active channels (%v) — supersede all but one", name, len(bound), names)
	}
	ch := bound[0]
	if want := ch.Spec.Channel.CodeVersion; want != codeVersion {
		return core.Activity{}, fmt.Errorf("channel %s v%d binds adapter %s at code version %d; this binary carries version %d — supersede the channel or run the matching binary",
			ch.Name, ch.Version, name, want, codeVersion)
	}
	return ch, nil
}

// Result says what one adapter run did. PassErr carries the adapter's own
// failure, if any — events gathered before it are still booked, which is
// safe precisely because passes are idempotent.
type Result struct {
	Channel    string       `json:"channel"`         // the channel the pass spoke through
	Version    int          `json:"channel_version"` // and its version
	Events     []core.Event `json:"-"`               // what the pass returned, stamped
	Submitted  int          `json:"submitted"`       // events newly appended
	Duplicates int          `json:"duplicates"`      // retries the log already had
	Booked     int          `json:"booked"`          // objects materialized from them
	Errors     []error      `json:"-"`               // worklist conditions from booking
	PassErr    error        `json:"-"`
}

// Run executes one adapter pass through its channel: bind, let the adapter
// read and return, check every event against the declaration, append what
// the log does not have yet (stamped with the channel) in one transaction,
// then process pending so the evidence materializes. An event of a type the
// channel never declared or one without identity is a bug in the adapter,
// not residue: the whole pass is refused and nothing is appended — and the
// same holds for any failure inside the batch.
func Run(ctx context.Context, s *store.Store, x *exec.Executor, a Adapter) (Result, error) {
	// One pass of a channel at a time, here and in every other process: two
	// passes of a channel that sends would both send what the log owes.
	release, err := s.LockChannel(ctx, a.Name())
	if err != nil {
		return Result{}, err
	}
	defer release()
	// A pass is one company's. Run as the owner it fans out over every
	// company the log holds and then the installation's own documents;
	// run for an office it fans out over the office's clients; each pass
	// sees and stamps one company — rhea pass without a company stored
	// company A's bank lines as rows every company saw.
	if sh, ok := a.(Shared); ok && sh.Installation() {
		return runScoped(store.WithInstallation(ctx), s, x, a)
	}
	set, scoped := store.ScopeOf(ctx)
	var scopes []context.Context
	switch {
	case !scoped:
		companies, err := s.Companies(ctx)
		if err != nil {
			return Result{}, err
		}
		for _, c := range companies {
			scopes = append(scopes, store.WithCompany(ctx, c))
		}
		scopes = append(scopes, store.WithInstallation(ctx))
	case len(set) > 1 && store.CompanyOf(ctx) == "":
		for _, c := range set {
			scopes = append(scopes, store.ActFor(ctx, c))
		}
	default:
		return runScoped(ctx, s, x, a)
	}
	var total Result
	for _, sctx := range scopes {
		res, err := runScoped(sctx, s, x, a)
		total.Channel, total.Version = res.Channel, res.Version
		total.Events = append(total.Events, res.Events...)
		total.Submitted += res.Submitted
		total.Duplicates += res.Duplicates
		total.Booked += res.Booked
		total.Errors = append(total.Errors, res.Errors...)
		if err != nil {
			return total, err
		}
		if res.PassErr != nil {
			total.PassErr = errors.Join(total.PassErr, res.PassErr)
		}
	}
	return total, nil
}

// scopedReads shows an adapter one company's objects: the installation's
// rows are every scope's to read, but a document is one company's, and a
// reader seeing it from every company would book it for each. Of the types
// an adapter only consults (Acting), the installation's rows stay: hiding
// them hid the pack's VAT rates from every company's KSeF pass, which sent
// nothing (review 2026-10-21, F1).
type scopedReads struct {
	Reads
	company string
	acts    map[string]bool // nil: every type is acted on
}

// EventsOfTypes is the store's, within the scope, for readers that look at
// what they proposed before.
func (r scopedReads) EventsOfTypes(ctx context.Context, types []string) ([]core.Event, error) {
	er, ok := r.Reads.(interface {
		EventsOfTypes(ctx context.Context, types []string) ([]core.Event, error)
	})
	if !ok {
		return nil, nil
	}
	return er.EventsOfTypes(ctx, types)
}

func (r scopedReads) ObjectsByType(ctx context.Context, typ string) ([]core.Object, error) {
	objs, err := r.Reads.ObjectsByType(ctx, typ)
	if err != nil {
		return nil, err
	}
	if r.acts != nil && !r.acts[typ] {
		return objs, nil
	}
	out := objs[:0]
	for _, o := range objs {
		if o.Company == r.company {
			out = append(out, o)
		}
	}
	return out, nil
}

// runScoped is one pass for the company the context acts for (the
// installation when it acts for none), every event stamped with it.
func runScoped(ctx context.Context, s *store.Store, x *exec.Executor, a Adapter) (Result, error) {
	ch, err := Bind(ctx, s, a.Name(), a.CodeVersion())
	if err != nil {
		return Result{}, err
	}
	res := Result{Channel: ch.Name, Version: ch.Version}
	// A reader that wants its channel's policy (a confidence threshold)
	// gets the declaration before it reads.
	if b, ok := a.(interface{ Bound(core.Activity) }); ok {
		b.Bound(ch)
	}
	company := store.CompanyOf(ctx)
	reads := scopedReads{Reads: s, company: company}
	if acting, ok := a.(Acting); ok {
		reads.acts = map[string]bool{}
		for _, t := range acting.ActsOn() {
			reads.acts[t] = true
		}
	}
	events, passErr := a.Pass(ctx, reads)
	res.PassErr = passErr
	for _, ev := range events {
		if !ch.Spec.Allows(ev.Type) {
			return res, fmt.Errorf("channel %s v%d does not declare event type %q — nothing of this pass is appended", ch.Name, ch.Version, ev.Type)
		}
		if ev.DedupKey == "" {
			return res, fmt.Errorf("channel %s v%d: a %s event without identity (no dedup key) — nothing of this pass is appended", ch.Name, ch.Version, ev.Type)
		}
	}
	for i := range events {
		events[i].Kind = core.KindRaw
		events[i].Company = company
		events[i].Actor = "adapter:" + a.Name()
		events[i].ActivityName, events[i].ActivityVersion = ch.Name, ch.Version
		if ch.Spec.Channel.Market != "" {
			stamped, err := withMarket(events[i].Payload, ch.Spec.Channel.Market)
			if err != nil {
				return res, fmt.Errorf("channel %s v%d: %s event: %w", ch.Name, ch.Version, events[i].Type, err)
			}
			events[i].Payload = stamped
		}
		if len(events[i].Files) > 0 {
			named, err := withFiles(events[i].Payload, events[i].Files)
			if err != nil {
				return res, fmt.Errorf("channel %s v%d: %s event: %w — nothing of this pass is appended", ch.Name, ch.Version, events[i].Type, err)
			}
			events[i].Payload = named
		}
	}
	res.Events = events
	if res.Submitted, res.Duplicates, err = s.AppendEvents(ctx, events); err != nil {
		return res, err
	}
	if res.Submitted > 0 {
		res.Booked, res.Errors = x.ProcessPending(ctx)
	}
	return res, nil
}

// withFiles names a fact's files in its payload as a door names its file —
// {<input>: {hash, name, media_type, size}} — the hash computed here, never
// taken from the adapter; the store keeps the bytes before the fact. A file
// with no input or no name, one whose input the payload already holds, or
// one past the store's cap refuses the pass.
func withFiles(payload json.RawMessage, files []core.FileInput) (json.RawMessage, error) {
	var m map[string]any
	if err := json.Unmarshal(payload, &m); err != nil {
		return nil, err
	}
	for i := range files {
		f := &files[i]
		switch _, taken := m[f.Input]; {
		case f.Input == "" || f.Name == "":
			return nil, fmt.Errorf("a file needs its input and its name (%q, %q)", f.Input, f.Name)
		case taken:
			return nil, fmt.Errorf("file %s: the payload already holds %q", f.Name, f.Input)
		case len(f.Bytes) > core.MaxFileBytes:
			return nil, fmt.Errorf("file %s is %d bytes; the store keeps a file of at most %d MiB", f.Name, len(f.Bytes), core.MaxFileBytes>>20)
		}
		if f.MediaType == "" {
			f.MediaType = "application/octet-stream"
		}
		f.Hash = core.HashBytes(f.Bytes)
		m[f.Input] = map[string]any{"hash": f.Hash, "name": f.Name, "media_type": f.MediaType, "size": len(f.Bytes)}
	}
	return json.Marshal(m)
}

// withMarket stamps the channel's market into a payload that names none:
// the market is provenance of the raw side, declared on the channel by the
// pack that declares it, never a reader's guess.
func withMarket(payload json.RawMessage, market string) (json.RawMessage, error) {
	var m map[string]any
	if err := json.Unmarshal(payload, &m); err != nil {
		return nil, err
	}
	if _, has := m["market"]; has {
		return payload, nil
	}
	m["market"] = market
	return json.Marshal(m)
}
