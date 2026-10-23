// Package storetest provides a throwaway-database Store for tests.
package storetest

import (
	"cmp"
	"context"
	"encoding/json"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"rhea/internal/core"
	"rhea/internal/store"
)

// New returns a Store on a throwaway database (schema run, dropped at test
// cleanup). Skips the test when Postgres is down.
func New(t *testing.T) *store.Store {
	t.Helper()
	ctx := context.Background()
	admin, err := store.Open(ctx, store.DSN())
	if err != nil {
		// A box without the container reports most of the kernel untested:
		// a skip on a laptop, a failure where RHEA_CI says the run counts.
		if os.Getenv("RHEA_CI") != "" {
			t.Fatalf("postgres unavailable and RHEA_CI is set: %v", err)
		}
		t.Skipf("postgres unavailable: %v", err)
	}
	admin.Close()
	s, cleanup, err := store.Throwaway(ctx, "rhea_test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	watched(t, s)
	return s
}

// NewBare returns a Store on a throwaway database with the schema and
// nothing else — no seeded verbs, no events — the target a restore wants.
func NewBare(t *testing.T) *store.Store {
	t.Helper()
	ctx := context.Background()
	admin, err := store.Open(ctx, store.DSN())
	if err != nil {
		if os.Getenv("RHEA_CI") != "" {
			t.Fatalf("postgres unavailable and RHEA_CI is set: %v", err)
		}
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

// Twice tells a story — a test's whole body — twice, each telling on
// throwaway databases of its own, and fails when the two logs differ in any
// fact but the moment it was recorded (invariant 4; the whole-tree review of
// 2026-10-21, H1). Replay rebuilds objects from the derived facts the log
// keeps and never runs a rule again, so it proves the projection, not the
// rules: told twice, the same raw facts go through the same rule versions a
// second time, every approval's reaction and chain with them, and must book
// the same derived facts — a rule reading the machine's clock or a map's
// order would book something else the second time.
func Twice(t *testing.T, story func(*testing.T)) {
	t.Helper()
	var logs [2][][]string
	for i, telling := range []string{"first", "second"} {
		name := t.Name() + "/" + telling
		watchMu.Lock()
		watchers[name] = func(t *testing.T, s *store.Store) { logs[i] = append(logs[i], logOf(t, s)) }
		watchMu.Unlock()
		ok := t.Run(telling, story)
		watchMu.Lock()
		delete(watchers, name)
		watchMu.Unlock()
		if !ok {
			return
		}
	}
	if len(logs[0]) == 0 || len(logs[0]) != len(logs[1]) {
		t.Fatalf("the tellings opened %d and %d databases", len(logs[0]), len(logs[1]))
	}
	for d := range logs[0] {
		a, b := logs[0][d], logs[1][d]
		for i := range max(len(a), len(b)) {
			if i >= len(a) || i >= len(b) || a[i] != b[i] {
				t.Fatalf("database %d, fact %d of %d and %d: the first telling logged\n%s\nthe second\n%s", d+1, i+1, len(a), len(b), line(a, i), line(b, i))
			}
		}
	}
}

var (
	watchMu  sync.Mutex
	watchers = map[string]func(*testing.T, *store.Store){} // by the telling's test name
)

// watched hands a store opened inside a telling to Twice, which reads its
// log when the telling ends — before the database is dropped, as cleanups
// run last first.
func watched(t *testing.T, s *store.Store) {
	watchMu.Lock()
	defer watchMu.Unlock()
	for name, keep := range watchers {
		if t.Name() == name || strings.HasPrefix(t.Name(), name+"/") {
			t.Cleanup(func() { keep(t, s) })
		}
	}
}

// logOf is a store's whole log in id order, one line per fact, the moment
// it was recorded left out.
func logOf(t *testing.T, s *store.Store) []string {
	ctx := store.WithOwner(context.Background())
	var evs []core.Event
	for _, kind := range []string{core.KindRaw, core.KindDerived} {
		got, err := s.EventsByKind(ctx, kind)
		if err != nil {
			t.Fatal(err)
		}
		evs = append(evs, got...)
	}
	slices.SortFunc(evs, func(a, b core.Event) int { return cmp.Compare(a.ID, b.ID) })
	out := make([]string, len(evs))
	for i, ev := range evs {
		ev.RecordedAt = time.Time{}
		b, err := json.Marshal(ev)
		if err != nil {
			t.Fatal(err)
		}
		out[i] = string(b)
	}
	return out
}

func line(log []string, i int) string {
	if i < len(log) {
		return log[i]
	}
	return "(nothing)"
}
