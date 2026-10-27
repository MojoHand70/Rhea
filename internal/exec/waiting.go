package exec

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"rhea/internal/core"
	"rhea/internal/store"
)

// waitingMemory is what the executor remembers of the facts still waiting
// for a rule, per scope (2026-10-27, the 100,000-invoice measurement): at
// 39,000 invoices every door read the whole log about four times to find
// them — 92% of an invoice's time in the database, growing with everything
// the installation ever recorded.
type waitingMemory struct {
	mu   sync.Mutex
	sets map[string]waitingSet
}

// waitingSet is one scope's: the log read up to an event, and the facts that
// still waited then, in log order.
type waitingSet struct {
	read   int64
	events []core.Event
}

// memoryInit guards the memory's first use; the memory itself is shared by
// an executor and the copies a dry run makes of it, all over one store.
var memoryInit sync.Mutex

func (x *Executor) memory() *waitingMemory {
	memoryInit.Lock()
	defer memoryInit.Unlock()
	if x.waiting == nil {
		x.waiting = &waitingMemory{sets: map[string]waitingSet{}}
	}
	return x.waiting
}

// waitingFacts is the worklist pass's set: the raw events no rule has acted
// on that the caller's scope sees. Under the kernel's lock no fact rules act
// on is half-written (the store records every one under it), so what the log
// holds up to its newest event is final: an explained fact stays explained,
// and a fact recorded later has a later id. The executor therefore reads
// only the facts after the last it read, and asks of those it remembers
// waiting only whether a rule has acted on them since — another process
// included, which records and books under the same lock. Without the lock,
// the whole log, as before.
func (x *Executor) waitingFacts(ctx context.Context) ([]core.Event, error) {
	if !store.Locked(ctx) {
		return x.Store.PendingEvents(ctx)
	}
	m := x.memory()
	set, scoped := store.ScopeOf(ctx)
	key := fmt.Sprintf("%t %v %s", scoped, set, store.CompanyOf(ctx))
	m.mu.Lock()
	w := m.sets[key]
	m.mu.Unlock()
	last, err := x.Store.LastEventID(ctx)
	if err != nil {
		return nil, err
	}
	if last == w.read {
		return w.events, nil // nothing recorded since, nothing booked
	}
	still := make([]core.Event, 0, len(w.events))
	if len(w.events) > 0 {
		ids := make([]int64, len(w.events))
		for i, ev := range w.events {
			ids[i] = ev.ID
		}
		explained, err := x.Store.Explained(ctx, ids)
		if err != nil {
			return nil, err
		}
		for _, ev := range w.events {
			if !explained[ev.ID] {
				still = append(still, ev)
			}
		}
	}
	fresh, err := x.Store.PendingEventsAfter(ctx, w.read, last)
	if err != nil {
		return nil, err
	}
	out := append(still, fresh...)
	m.mu.Lock()
	m.sets[key] = waitingSet{read: last, events: slices.Clip(out)}
	m.mu.Unlock()
	return out, nil
}
