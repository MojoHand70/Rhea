package store

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Throwaway creates a fresh database on the rhea Postgres, runs the schema and
// returns a Store on it, plus the cleanup that drops it. For tests and for
// runs that must never touch the system of record (rhea eval): a log is
// append-only, so a run that should leave no trace gets its own.
func Throwaway(ctx context.Context, prefix string) (*Store, func(), error) {
	admin, err := Open(ctx, DSN())
	if err != nil {
		return nil, nil, err
	}
	defer admin.Close()
	name := fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	if _, err := admin.Pool.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		return nil, nil, fmt.Errorf("create %s: %w", name, err)
	}
	drop := func() {
		if a, err := Open(context.Background(), DSN()); err == nil {
			a.Pool.Exec(context.Background(), "DROP DATABASE "+name+" WITH (FORCE)")
			a.Close()
		}
	}
	base := DSN()
	s, err := Open(ctx, base[:strings.LastIndex(base, "/")+1]+name)
	if err != nil {
		drop()
		return nil, nil, fmt.Errorf("open %s: %w", name, err)
	}
	if err := s.Init(ctx); err != nil {
		s.Close()
		drop()
		return nil, nil, fmt.Errorf("init %s: %w", name, err)
	}
	return s, func() { s.Close(); drop() }, nil
}
