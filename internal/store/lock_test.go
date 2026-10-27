package store_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"rhea/internal/core"
	"rhea/internal/store"
)

// TestAFactWaitsForTheChainThatRuns (2026-10-27, the 100,000-invoice
// measurement): a fact rules act on is recorded only under the kernel's
// lock, so while a chain runs — here in another process, which holds the
// lock on a connection of its own — nothing such is half-written, and the
// kernel may remember which facts still wait instead of reading the whole log
// after every one. The clock's days, a pack's master data and a reader's
// batch were written beside a running chain. Each now waits for it; inside
// a transaction that does not hold the lock it is refused; a fact no rule
// reads (a refusal the kernel records) needs no lock.
func TestAFactWaitsForTheChainThatRuns(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	other, err := s.Pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Release()
	if _, err := other.Exec(ctx, `SELECT pg_advisory_lock(1380468033)`); err != nil { // the kernel's lock, "RHEA"
		t.Fatal(err)
	}
	waits := func(what string, write func() error) chan error {
		t.Helper()
		done := make(chan error, 1)
		go func() { done <- write() }()
		select {
		case err := <-done:
			t.Fatalf("%s recorded while another process's chain runs: %v", what, err)
		case <-time.After(300 * time.Millisecond):
		}
		return done
	}
	one := waits("a fact", func() error { _, err := s.AppendEvent(ctx, rawEvent("while-a-chain-runs")); return err })
	batch := waits("a reader's batch", func() error {
		_, _, err := s.AppendEvents(ctx, []core.Event{rawEvent("batch-1"), rawEvent("batch-2")})
		return err
	})
	if _, err := other.Exec(ctx, `SELECT pg_advisory_unlock(1380468033)`); err != nil {
		t.Fatal(err)
	}
	for _, done := range []chan error{one, batch} {
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(10 * time.Second):
			t.Fatal("still waiting after the chain ended")
		}
	}

	// inside a transaction of its own, without the lock: refused
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err := store.AppendEvent(ctx, tx, rawEvent("in-a-transaction")); err == nil || !strings.Contains(err.Error(), "lock") {
		t.Fatalf("a fact rules act on, in a transaction without the lock: %v", err)
	}
	tx.Rollback(ctx)
	// under it: recorded
	held, release, err := s.Lock(ctx, 0)
	if err != nil {
		t.Fatal(err)
	}
	tx, err = s.Pool.Begin(held)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.AppendEvent(held, tx, rawEvent("under-the-lock")); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(held); err != nil {
		t.Fatal(err)
	}
	release()
	// a fact no rule reads needs none
	if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: core.EventAccessRefused, OccurredAt: "2026-10-27",
		Payload: json.RawMessage(`{"person":"x","door":"y"}`), Actor: "kernel"}); err != nil {
		t.Fatal(err)
	}
}
