package exec

import (
	"context"

	"rhea/internal/store"
)

// ErrBusy is a wait for the executor that passed its limit: the books are
// being worked on, and the caller may try again.
var ErrBusy = store.ErrBusy

// serialize takes the kernel's lock (store.Lock: one chain at a time, in
// this process and across processes) for the run's duration, waiting at most
// the executor's LockWait, and returns the context that marks it held, so a
// nested run on that context does not wait on itself.
func (x *Executor) serialize(ctx context.Context) (context.Context, func(), error) {
	return x.Store.Lock(ctx, x.LockWait)
}
