package store

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Throwaway creates a fresh database on the rhea Postgres, runs the schema and
// returns a Store on it, plus the cleanup that drops it. For tests and for
// runs that must never touch the system of record (rhea eval): a log is
// append-only, so a run that should leave no trace gets its own.
func Throwaway(ctx context.Context, prefix string) (*Store, func(), error) {
	dsn, drop, err := ThrowawayDSN(ctx, prefix)
	if err != nil {
		return nil, nil, err
	}
	s, err := Open(ctx, dsn)
	if err != nil {
		drop()
		return nil, nil, fmt.Errorf("open %s: %w", dsn, err)
	}
	if err := s.Init(ctx); err != nil {
		s.Close()
		drop()
		return nil, nil, fmt.Errorf("init %s: %w", dsn, err)
	}
	return s, func() { s.Close(); drop() }, nil
}

// ThrowawayBare is Throwaway without the seed: the schema in place, no verb
// and no event — the empty system of record a restore lands in.
func ThrowawayBare(ctx context.Context, prefix string) (*Store, func(), error) {
	dsn, drop, err := ThrowawayDSN(ctx, prefix)
	if err != nil {
		return nil, nil, err
	}
	s, err := Open(ctx, dsn)
	if err != nil {
		drop()
		return nil, nil, fmt.Errorf("open %s: %w", dsn, err)
	}
	if err := s.InitSchema(ctx); err != nil {
		s.Close()
		drop()
		return nil, nil, fmt.Errorf("init %s: %w", dsn, err)
	}
	return s, func() { s.Close(); drop() }, nil
}

// ThrowawayDSN creates an empty database on the rhea Postgres and returns its
// connection string and the cleanup that drops it — for stores other than
// the kernel's (the network's).
func ThrowawayDSN(ctx context.Context, prefix string) (string, func(), error) {
	admin, err := Open(ctx, DSN())
	if err != nil {
		return "", nil, err
	}
	defer admin.Close()
	name := fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	if _, err := admin.Pool.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		return "", nil, fmt.Errorf("create %s: %w", name, err)
	}
	drop := func() {
		if a, err := Open(context.Background(), DSN()); err == nil {
			a.Pool.Exec(context.Background(), "DROP DATABASE "+name+" WITH (FORCE)")
			a.Close()
		}
	}
	return withDatabase(DSN(), name), drop, nil
}

// withDatabase names another database in a DSN, keeping its options: a URL
// keeps its query string (sslmode and the like), a keyword DSN its other
// keywords.
func withDatabase(base, name string) string {
	if strings.Contains(base, "://") {
		path, query, hasQuery := strings.Cut(base, "?")
		out := path[:strings.LastIndex(path, "/")+1] + name
		if hasQuery {
			out += "?" + query
		}
		return out
	}
	if dbnameKeyword.MatchString(base) {
		return dbnameKeyword.ReplaceAllString(base, "dbname="+name)
	}
	return base + " dbname=" + name
}

var dbnameKeyword = regexp.MustCompile(`dbname=\S+`)
