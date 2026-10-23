package store_test

import (
	"context"
	"testing"

	"rhea/internal/core"
	"rhea/internal/store"
)

// A company's request never wipes the object cache (invariant 7; review
// 2026-10-21, C3): the role every scoped request runs as was granted DELETE
// on object, so any path running in a company's scope could delete that
// company's objects, and nothing but convention said only the replay may.
// The replay runs as the owner; a scoped request now cannot delete an
// object, its own company's included. It still amends them: a chain books
// amendments in its company's scope.
func TestACompanysRequestDeletesNoObject(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	const alfa = "PL5250001111"
	ev := rawEvent("cache-1")
	ev.Company = alfa
	id, err := s.AppendEvent(ctx, ev)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.InsertObject(ctx, s.Pool, core.Object{ID: "thing-1", Type: "thing", TypeVersion: 1,
		State: map[string]any{"n": 1}, SourceEventID: id, RuleID: "r", RuleVersion: 1, Company: alfa}); err != nil {
		t.Fatal(err)
	}
	scope := store.WithCompany(ctx, alfa)
	if _, err := s.Pool.Exec(scope, `DELETE FROM object WHERE object_id = 'thing-1'`); err == nil {
		t.Fatal("alfa's request deleted alfa's object")
	}
	if err := store.AmendObject(scope, s.Pool, "thing-1", map[string]any{"n": 2}); err != nil {
		t.Fatalf("alfa's chain amends its object: %v", err)
	}
	if o, err := s.GetObject(scope, "thing-1"); err != nil || o.State["n"] != float64(2) {
		t.Fatalf("the object after: %+v %v", o, err)
	}
}
