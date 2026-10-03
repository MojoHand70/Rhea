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
package adapter

import (
	"context"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store"
)

// Reads is the slice of the read side an adapter may see. *store.Store
// satisfies it; adapters get no writes.
type Reads interface {
	ObjectsByType(ctx context.Context, typ string) ([]core.Object, error)
}

// Adapter is the declared contract: inspect the world, talk to the outside
// through your own client, return the raw events the log is now owed. A pass
// must be idempotent — returning the same (dedup-keyed) events twice is fine.
type Adapter interface {
	Name() string
	Pass(ctx context.Context, reads Reads) ([]core.Event, error)
}

// Result says what one adapter run did. PassErr carries the adapter's own
// failure, if any — events gathered before it are still booked, which is
// safe precisely because passes are idempotent.
type Result struct {
	Submitted  int     `json:"submitted"`  // events newly appended
	Duplicates int     `json:"duplicates"` // retries the log already had
	Booked     int     `json:"booked"`     // objects materialized from them
	Errors     []error `json:"-"`          // worklist conditions from booking
	PassErr    error   `json:"-"`
}

// Run executes one adapter pass: append what it returns (skipping events the
// log already has), then process pending so the evidence materializes.
func Run(ctx context.Context, s *store.Store, x *exec.Executor, a Adapter) (Result, error) {
	var res Result
	events, passErr := a.Pass(ctx, s)
	res.PassErr = passErr
	for _, ev := range events {
		ev.Kind = core.KindRaw
		ev.Actor = "adapter:" + a.Name()
		switch _, err := s.AppendEvent(ctx, ev); {
		case err == nil:
			res.Submitted++
		case store.IsDuplicate(err):
			res.Duplicates++
		default:
			return res, err
		}
	}
	if res.Submitted > 0 {
		res.Booked, res.Errors = x.ProcessPending(ctx)
	}
	return res, nil
}
