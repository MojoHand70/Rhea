package store

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"sync"
	"time"

	"rhea/internal/core"
)

// kernelGate is the in-process half of the kernel's lock: one goroutine per
// process holds a pool connection while waiting on Postgres. Without it,
// every waiting caller held a connection in pg_advisory_lock, and once the
// callers outnumbered the pool the holder could not get a connection for
// its own reads and booking — a deadlock the review reproduced with nine
// callers on an eight-connection pool. It is a one-place channel so a wait
// on it gives way to the caller's context: as a mutex, a caller queued
// behind a long run waited whatever became of its request (review
// 2026-10-21, G5).
var kernelGate = make(chan struct{}, 1)

// ErrBusy is a wait for the kernel's lock that passed its limit: the books
// are being worked on, and the caller may try again.
var ErrBusy = errors.New("the books are busy — another run holds them; try again shortly")

// kernelLock is the advisory lock every executor run holds: the kernel
// evaluates one chain at a time, in this process and across processes — the
// CLI beside the server, two handlers in one server. Two evaluations of one
// pending event would otherwise both book: the second reads the first's
// committed state and amends an open amount twice, and replay reproduces the
// wrong state faithfully. Materializations collide on the object key;
// amend-only and match-only chains have nothing else to stop them.
//
// A fact rules act on is recorded under it too (2026-10-27): then, while a
// chain runs, no such fact is half-written, and what the log holds below
// the newest event is final — the executor remembers which facts still wait
// instead of reading the whole log after every one. The lock lives in the
// store, where every writer passes, so the clock, the pack loader and a
// reader's batch take it as a door does.
const kernelLock = 0x52484541 // "RHEA"

type lockedKey struct{}

// Locked reports whether ctx runs under the kernel's lock.
func Locked(ctx context.Context) bool { return ctx.Value(lockedKey{}) != nil }

// Lock takes the kernel's lock on a dedicated connection for the run's
// duration and returns the context that marks it held, so a nested run on
// that context (a trigger's worklist pass, a fact recorded inside a chain)
// does not wait on itself. wait bounds the wait — past it the caller is told
// the books are busy — and zero waits as long as ctx lets it. The lock lives
// on the session: a dropped connection releases it.
func (s *Store) Lock(ctx context.Context, wait time.Duration) (context.Context, func(), error) {
	if Locked(ctx) {
		return ctx, func() {}, nil
	}
	// The wait is the caller's to end, and the limit bounds it; the run
	// itself keeps the caller's own context.
	waiting := ctx
	if wait > 0 {
		var cancel context.CancelFunc
		waiting, cancel = context.WithTimeout(ctx, wait)
		defer cancel()
	}
	busy := func(err error) error {
		if errors.Is(err, context.DeadlineExceeded) && ctx.Err() == nil {
			return ErrBusy
		}
		return fmt.Errorf("waiting for the kernel's lock: %w", err)
	}
	select {
	case kernelGate <- struct{}{}:
	case <-waiting.Done():
		return ctx, nil, busy(waiting.Err())
	}
	conn, err := s.Pool.Acquire(waiting)
	if err != nil {
		<-kernelGate
		return ctx, nil, busy(err)
	}
	if _, err := conn.Exec(waiting, `SELECT pg_advisory_lock($1)`, kernelLock); err != nil {
		conn.Release()
		<-kernelGate
		if waiting.Err() != nil {
			return ctx, nil, busy(waiting.Err())
		}
		return ctx, nil, err
	}
	release := func() {
		// Unlock even when the caller's context is gone; a session that
		// dies unlocks by itself.
		_, _ = conn.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, kernelLock)
		conn.Release()
		<-kernelGate
	}
	return context.WithValue(ctx, lockedKey{}, true), release, nil
}

// ruleFood is a fact rules may act on: a raw event outside the system
// namespaces — what the executor's waiting set holds (queries/pending.sql).
func ruleFood(ev core.Event) bool {
	return ev.Kind == core.KindRaw && !core.ReservedEventType(ev.Type)
}

// One pass of a channel at a time (invariant 10): two passes of a channel
// that sends — KSeF, the JPK gateway — would both send what the log still
// owes, the CLI's pass beside the server's, two servers side by side. The
// lock is an advisory lock keyed by the channel's name (the class "CHAN"),
// held on a dedicated connection for the pass; in process, a gate per
// channel queues the waiters first, so they do not hold the pool's
// connections while the pass needs them (the kernel lock's lesson, G5).

//go:embed queries/channel_lock.sql
var channelLockSQL string

//go:embed queries/channel_unlock.sql
var channelUnlockSQL string

var (
	channelGatesMu sync.Mutex
	channelGates   = map[string]chan struct{}{}
)

func channelGate(name string) chan struct{} {
	channelGatesMu.Lock()
	defer channelGatesMu.Unlock()
	g, ok := channelGates[name]
	if !ok {
		g = make(chan struct{}, 1)
		channelGates[name] = g
	}
	return g
}

// LockChannel takes a channel's lock for one pass, waiting as long as ctx
// lets it; release lets it go. A dropped connection releases it.
func (s *Store) LockChannel(ctx context.Context, name string) (release func(), err error) {
	gate := channelGate(name)
	select {
	case gate <- struct{}{}:
	case <-ctx.Done():
		return nil, fmt.Errorf("waiting for channel %s: %w", name, ctx.Err())
	}
	conn, err := s.Pool.Acquire(ctx)
	if err != nil {
		<-gate
		return nil, fmt.Errorf("waiting for channel %s: %w", name, err)
	}
	if _, err := conn.Exec(ctx, channelLockSQL, name); err != nil {
		conn.Release()
		<-gate
		return nil, fmt.Errorf("waiting for channel %s: %w", name, err)
	}
	return func() {
		_, _ = conn.Exec(context.Background(), channelUnlockSQL, name)
		conn.Release()
		<-gate
	}, nil
}
