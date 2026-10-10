package adapter

import (
	"context"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

// probe is an adapter that returns whatever the test hands it.
type probe struct{ events []core.Event }

func (probe) Name() string                                        { return "probe" }
func (probe) CodeVersion() int                                    { return 2 }
func (p probe) Pass(context.Context, Reads) ([]core.Event, error) { return p.events, nil }

func probeChannel(status string, codeVersion int, events ...string) core.Activity {
	return core.Activity{
		Name: "probe_channel", Status: status, Domain: "test", CreatedBy: "test",
		Description: "the test's own channel",
		Spec: core.ActivitySpec{
			Emits: "probe.seen", Who: []string{"adapter"},
			Channel: &core.ChannelSpec{
				Adapter: "probe", CodeVersion: codeVersion, Format: "test fixture",
				Identity:   "the probe's sequence number",
				Confidence: core.ConfidenceSpec{Policy: core.ConfidenceExact},
				Events:     events,
			},
		},
	}
}

// The registry: an adapter speaks only through an active channel bound to it
// at its code version; the declaration is the contract for what it may
// append; every event it appends names the channel. The brief (DIRECTION
// 2026-10-10): "every raw event names its channel and version".
func TestChannelRegistry(t *testing.T) {
	Register(Implemented{Name: "probe", CodeVersion: 2, Format: "test fixture"})
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	seen := core.Event{Type: "probe.seen", OccurredAt: "2026-10-11", Payload: []byte(`{"n":1}`), DedupKey: "probe/1"}
	worklist := func() int {
		wl, err := s.UnmatchedRawEvents(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return len(wl)
	}
	refused := func(a Adapter, want string) {
		t.Helper()
		before := worklist()
		if _, err := Run(ctx, s, x, a); err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("want refusal %q, got %v", want, err)
		}
		if worklist() != before {
			t.Fatal("a refused pass appended something")
		}
	}

	// No channel: the adapter is implemented, not connected.
	refused(probe{[]core.Event{seen}}, "no active channel")

	// A draft channel is not a door.
	if _, err := s.InsertActivityVersion(ctx, probeChannel(core.StatusDraft, 2)); err != nil {
		t.Fatal(err)
	}
	refused(probe{[]core.Event{seen}}, "no active channel")

	// An active channel at another code version: the binary does not match
	// what the company connected.
	if _, err := s.InsertActivityVersion(ctx, probeChannel(core.StatusActive, 1)); err != nil {
		t.Fatal(err)
	}
	refused(probe{[]core.Event{seen}}, "code version 1; this binary carries version 2")

	// The matching version: the pass goes through, and the event names the
	// channel the way an event through a door names its activity.
	ch, err := s.InsertActivityVersion(ctx, probeChannel(core.StatusActive, 2, "probe.seen", "probe.lost"))
	if err != nil {
		t.Fatal(err)
	}
	res, err := Run(ctx, s, x, probe{[]core.Event{seen}})
	if err != nil || res.Submitted != 1 || res.Channel != "probe_channel" || res.Version != ch.Version {
		t.Fatalf("run = %+v, %v", res, err)
	}
	ev, ok, err := s.LatestEventOfType(ctx, "probe.seen")
	if err != nil || !ok {
		t.Fatalf("event missing: %v", err)
	}
	if ev.ActivityName != "probe_channel" || ev.ActivityVersion != ch.Version || ev.Actor != "adapter:probe" {
		t.Fatalf("stamp = %s v%d by %s", ev.ActivityName, ev.ActivityVersion, ev.Actor)
	}

	// Retry is free, as before.
	res, err = Run(ctx, s, x, probe{[]core.Event{seen}})
	if err != nil || res.Submitted != 0 || res.Duplicates != 1 {
		t.Fatalf("second pass = %+v, %v", res, err)
	}

	// The declaration is the contract: an undeclared event type, or an
	// event without identity, refuses the whole pass — a bug, not residue.
	stray := core.Event{Type: "probe.stray", OccurredAt: "2026-10-11", Payload: []byte(`{}`), DedupKey: "probe/stray"}
	lost := core.Event{Type: "probe.lost", OccurredAt: "2026-10-11", Payload: []byte(`{}`), DedupKey: "probe/2"}
	refused(probe{[]core.Event{lost, stray}}, `does not declare event type "probe.stray"`)
	refused(probe{[]core.Event{{Type: "probe.lost", OccurredAt: "2026-10-11", Payload: []byte(`{}`)}}}, "without identity")

	// Two active channels binding one adapter is ambiguous, and refused.
	if _, err := s.InsertActivityVersion(ctx, core.Activity{
		Name: "probe_twin", Status: core.StatusActive, Domain: "test", CreatedBy: "test",
		Spec: core.ActivitySpec{Emits: "probe.seen", Channel: &core.ChannelSpec{
			Adapter: "probe", CodeVersion: 2, Format: "f", Identity: "i",
			Confidence: core.ConfidenceSpec{Policy: core.ConfidenceExact}}},
	}); err != nil {
		t.Fatal(err)
	}
	refused(probe{[]core.Event{lost}}, "bound by 2 active channels")

	// The two lists: what is connected, what is implemented.
	reg, err := Registry(ctx, s)
	if err != nil {
		t.Fatal(err)
	}
	states := map[string]string{}
	for _, c := range reg.Channels {
		states[c.Name] = c.State
	}
	if states["probe_channel"] != StateConnected || states["probe_twin"] != StateConnected || states["clock"] != StateConnected {
		t.Fatalf("channel states = %v", states)
	}
	adapters := map[string]AdapterStatus{}
	for _, a := range reg.Adapters {
		adapters[a.Name] = a
	}
	if adapters["ksef"].State != StateNotConnected || adapters["clock"].Channel != "clock" || adapters["probe"].State != StateConnected {
		t.Fatalf("adapter states = %+v", adapters)
	}
	for _, c := range reg.Channels {
		if c.Name == "probe_channel" && (c.EventsEmitted != 1 || c.Implemented != 2 || c.Volume != core.VolumeDocuments) {
			t.Fatalf("probe_channel status = %+v", c)
		}
	}
}

// A channel declared for code the binary does not carry shows as such, and
// its adapter never runs.
func TestChannelNotImplemented(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	a := core.Activity{
		Name: "fax", Status: core.StatusActive, Domain: "test", CreatedBy: "test",
		Spec: core.ActivitySpec{Emits: "fax.received", Channel: &core.ChannelSpec{
			Adapter: "fax", CodeVersion: 1, Format: "paper", Identity: "page hash",
			Confidence: core.ConfidenceSpec{Policy: core.ConfidenceModel, Threshold: "0.9"}}},
	}
	if _, err := s.InsertActivityVersion(ctx, a); err != nil {
		t.Fatal(err)
	}
	reg, err := Registry(ctx, s)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range reg.Channels {
		if c.Name == "fax" {
			if c.State != StateNotImplemented || c.Implemented != 0 {
				t.Fatalf("fax = %+v", c)
			}
			return
		}
	}
	t.Fatal("fax channel not listed")
}
