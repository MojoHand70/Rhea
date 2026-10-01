// Package storetest provides a throwaway-database Store for tests.
package storetest

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"rhea/internal/store"
)

// New connects to the rhea Postgres, creates a throwaway database, runs the
// schema and returns a Store on it. Skips the test when Postgres is down.
func New(t *testing.T) *store.Store {
	t.Helper()
	ctx := context.Background()
	admin, err := store.Open(ctx, store.DSN())
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	defer admin.Close()
	name := fmt.Sprintf("rhea_test_%d", time.Now().UnixNano())
	if _, err := admin.Pool.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		t.Fatalf("create test db: %v", err)
	}
	base := store.DSN()
	dsn := base[:strings.LastIndex(base, "/")+1] + name
	s, err := store.Open(ctx, dsn)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := s.Init(ctx); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	t.Cleanup(func() {
		s.Close()
		a, err := store.Open(context.Background(), store.DSN())
		if err == nil {
			a.Pool.Exec(context.Background(), "DROP DATABASE "+name+" WITH (FORCE)")
			a.Close()
		}
	})
	return s
}
