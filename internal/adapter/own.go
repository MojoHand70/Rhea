package adapter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"rhea/internal/exec"
	"rhea/internal/store"
)

// Channels that run on their own (KK, 2026-10-19 and 2026-10-27: a person
// acts in the interface, never through a command or a timer someone keeps on
// a server): a channel whose declaration says how often — poll, "1m" — is run
// by the server that often, for every company. A pass is already a poll over
// what the log still owes the outside world (the adapter contract), so a pass
// that finds nothing owed does nothing; one pass of a channel at a time holds
// across processes (store.LockChannel). A channel that says nothing is run
// as before, by a command.

// Own runs the channels that run on their own. Tick starts each one that is
// due in a goroutine of its own, never a second one while one runs here.
type Own struct {
	Store *store.Store
	Exec  *exec.Executor
	// Log hears what a pass booked and every failure, a failure once until
	// it changes; nil is silent.
	Log func(format string, args ...any)

	mu      sync.Mutex
	last    map[string]time.Time
	running map[string]bool
	said    map[string]string
	wg      sync.WaitGroup
}

// Tick starts the channels due at now and names them. ctx is the scope the
// passes run in — the owner's, so each fans out over every company.
func (o *Own) Tick(ctx context.Context, now time.Time) ([]string, error) {
	acts, err := o.Store.ActiveActivities(ctx)
	if err != nil {
		return nil, err
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.last == nil {
		o.last, o.running, o.said = map[string]time.Time{}, map[string]bool{}, map[string]string{}
	}
	var started []string
	for _, act := range acts {
		ch := act.Spec.Channel
		if ch == nil || ch.Poll == "" {
			continue
		}
		every, err := time.ParseDuration(ch.Poll)
		if err != nil || o.running[ch.Adapter] || now.Sub(o.last[ch.Adapter]) < every {
			continue
		}
		o.last[ch.Adapter] = now
		a, ok := Lookup(ch.Adapter)
		if !ok {
			o.say(ch.Adapter, fmt.Sprintf("channel %s runs on its own, but this binary carries no adapter %s", act.Name, ch.Adapter))
			continue
		}
		o.running[ch.Adapter] = true
		started = append(started, act.Name)
		o.wg.Add(1)
		go func(name string, a Adapter) {
			defer o.wg.Done()
			res, err := Run(ctx, o.Store, o.Exec, a)
			o.mu.Lock()
			defer o.mu.Unlock()
			o.running[name] = false
			switch {
			case err != nil:
				o.say(name, fmt.Sprintf("channel %s: %v", name, err))
			case res.PassErr != nil:
				o.say(name, fmt.Sprintf("channel %s: %v", name, res.PassErr))
			default:
				o.say(name, "")
			}
			if res.Submitted > 0 && o.Log != nil {
				o.Log("channel %s: %d appended, %d booked, %d wait in the worklist", name, res.Submitted, res.Booked, len(res.Errors))
			}
		}(ch.Adapter, a)
	}
	return started, nil
}

// Wait waits for the passes Tick started.
func (o *Own) Wait() { o.wg.Wait() }

// say logs a channel's failure when it is new, and its end; o.mu is held.
func (o *Own) say(name, msg string) {
	if o.said[name] == msg {
		return
	}
	o.said[name] = msg
	if o.Log == nil {
		return
	}
	if msg == "" {
		o.Log("channel %s: runs again", name)
		return
	}
	o.Log("%s", msg)
}
