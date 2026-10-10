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

// Lookup returns a reader by its adapter name.
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
	ch, err := Bind(ctx, s, a.Name(), a.CodeVersion())
	if err != nil {
		return Result{}, err
	}
	res := Result{Channel: ch.Name, Version: ch.Version}
	events, passErr := a.Pass(ctx, s)
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
		events[i].Actor = "adapter:" + a.Name()
		events[i].ActivityName, events[i].ActivityVersion = ch.Name, ch.Version
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
