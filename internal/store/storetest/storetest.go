// Package storetest provides a throwaway-database Store for tests.
package storetest

import (
	"context"
	"testing"

	"rhea/internal/store"
)

// New returns a Store on a throwaway database (schema run, dropped at test
// cleanup). Skips the test when Postgres is down.
func New(t *testing.T) *store.Store {
	t.Helper()
	ctx := context.Background()
	admin, err := store.Open(ctx, store.DSN())
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	admin.Close()
	s, cleanup, err := store.Throwaway(ctx, "rhea_test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	return s
}

// NewBare returns a Store on a throwaway database with the schema and
// nothing else — no seeded verbs, no events — the target a restore wants.
func NewBare(t *testing.T) *store.Store {
	t.Helper()
	ctx := context.Background()
	admin, err := store.Open(ctx, store.DSN())
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	admin.Close()
	s, cleanup, err := store.ThrowawayBare(ctx, "rhea_bare")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	return s
}
