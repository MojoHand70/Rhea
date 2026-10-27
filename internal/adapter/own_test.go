package adapter

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

// ownProbe counts its passes; while hold is open a pass waits on it.
type ownProbe struct {
	passes *int32
	hold   chan struct{}
}

func (ownProbe) Name() string     { return "ownprobe" }
func (ownProbe) CodeVersion() int { return 1 }
func (p ownProbe) Pass(context.Context, Reads) ([]core.Event, error) {
	atomic.AddInt32(p.passes, 1)
	if p.hold != nil {
		<-p.hold
	}
	return nil, nil
}

// A channel that says how often runs on its own, that often: the server's
// tick starts its pass when it is due and not before, never a second one
// while one runs, and leaves alone a channel that says nothing — the one a
// command runs. A declaration's interval is a duration between ten seconds
// and a day.
func TestAChannelRunsOnItsOwn(t *testing.T) {
	Register(Implemented{Name: "ownprobe", CodeVersion: 1, Format: "test fixture"})
	var passes int32
	hold := make(chan struct{})
	var held atomic.Bool
	Provide("ownprobe", func() Adapter {
		if held.Load() {
			return ownProbe{&passes, hold}
		}
		return ownProbe{&passes, nil}
	})
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	own := probeChannel(core.StatusActive, 1)
	own.Name, own.Spec.Channel.Adapter, own.Spec.Channel.Poll = "ownprobe_channel", "ownprobe", "1m"
	if _, err := s.InsertActivityVersion(ctx, own); err != nil {
		t.Fatal(err)
	}
	if _, err := s.InsertActivityVersion(ctx, probeChannel(core.StatusActive, 2)); err != nil { // says nothing: a command runs it
		t.Fatal(err)
	}
	o := &Own{Store: s, Exec: x}
	owner := store.WithOwner(ctx)
	t0 := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
	tick := func(at time.Duration) []string {
		t.Helper()
		started, err := o.Tick(owner, t0.Add(at))
		if err != nil {
			t.Fatal(err)
		}
		return started
	}
	if started := tick(0); len(started) != 1 || started[0] != "ownprobe_channel" {
		t.Fatalf("the first tick started %v, want the channel that runs on its own", started)
	}
	o.Wait()
	if started := tick(30 * time.Second); len(started) != 0 {
		t.Fatalf("started %v half a minute after a pass that runs every minute", started)
	}
	if started := tick(61 * time.Second); len(started) != 1 {
		t.Fatalf("not started a minute after: %v", started)
	}
	o.Wait()
	if n := atomic.LoadInt32(&passes); n != 2 {
		t.Fatalf("%d passes, want 2", n)
	}

	held.Store(true)
	if started := tick(3 * time.Minute); len(started) != 1 {
		t.Fatalf("not started when due: %v", started)
	}
	if started := tick(5 * time.Minute); len(started) != 0 {
		t.Fatalf("a second pass started while the first runs: %v", started)
	}
	close(hold)
	o.Wait()
	if n := atomic.LoadInt32(&passes); n != 3 {
		t.Fatalf("%d passes, want 3", n)
	}

	for _, poll := range []string{"5s", "soon", "48h"} {
		bad := probeChannel(core.StatusDraft, 1)
		bad.Name, bad.Spec.Channel.Adapter, bad.Spec.Channel.Poll = "badpoll_"+poll, "ownprobe", poll
		if err := bad.Validate(); err == nil {
			t.Errorf("a channel that runs every %q was taken", poll)
		}
	}
}
