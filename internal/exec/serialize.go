package exec

import (
	"context"
	"errors"
	"fmt"
)

// executorGate is the in-process half of the lock: one goroutine per
// process holds a pool connection while waiting on Postgres. Without it,
// every waiting caller held a connection in pg_advisory_lock, and once the
// callers outnumbered the pool the holder could not get a connection for
// its own reads and booking — a deadlock the review reproduced with nine
// callers on an eight-connection pool. It is a one-place channel so a wait
// on it gives way to the caller's context: as a mutex, a caller queued
// behind a long run waited whatever became of its request (review
// 2026-10-21, G5).
var executorGate = make(chan struct{}, 1)

// ErrBusy is a wait for the executor that passed its limit: the books are
// being worked on, and the caller may try again.
var ErrBusy = errors.New("the books are busy — another run holds them; try again shortly")

// executorLock is the advisory lock every executor run holds: the kernel
// evaluates one chain at a time, in this process and across processes — the
// CLI beside the server, two handlers in one server. Two evaluations of one
// pending event would otherwise both book: the second reads the first's
// committed state and amends an open amount twice, and replay reproduces the
// wrong state faithfully. Materializations collide on the object key;
// amend-only and match-only chains have nothing else to stop them.
const executorLock = 0x52484541 // "RHEA"

type serializedKey struct{}

// serialize takes the executor's lock on a dedicated connection for the
// run's duration and returns the context that marks it held, so a nested
// run on that context (a trigger's worklist pass) does not wait on itself.
// The lock lives on the session: a dropped connection releases it.
func (x *Executor) serialize(ctx context.Context) (context.Context, func(), error) {
	if ctx.Value(serializedKey{}) != nil {
		return ctx, func() {}, nil
	}
	// The wait is the caller's to end, and the executor's limit bounds it;
	// the run itself keeps the caller's own context.
	wait := ctx
	if x.LockWait > 0 {
		var cancel context.CancelFunc
		wait, cancel = context.WithTimeout(ctx, x.LockWait)
		defer cancel()
	}
	busy := func(err error) error {
		if errors.Is(err, context.DeadlineExceeded) && ctx.Err() == nil {
			return ErrBusy
		}
		return fmt.Errorf("waiting for the executor: %w", err)
	}
	select {
	case executorGate <- struct{}{}:
	case <-wait.Done():
		return ctx, nil, busy(wait.Err())
	}
	conn, err := x.Store.Pool.Acquire(wait)
	if err != nil {
		<-executorGate
		return ctx, nil, busy(err)
	}
	if _, err := conn.Exec(wait, `SELECT pg_advisory_lock($1)`, executorLock); err != nil {
		conn.Release()
		<-executorGate
		if wait.Err() != nil {
			return ctx, nil, busy(wait.Err())
		}
		return ctx, nil, err
	}
	release := func() {
		// Unlock even when the caller's context is gone; a session that
		// dies unlocks by itself.
		_, _ = conn.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, executorLock)
		conn.Release()
		<-executorGate
	}
	return context.WithValue(ctx, serializedKey{}, true), release, nil
}
