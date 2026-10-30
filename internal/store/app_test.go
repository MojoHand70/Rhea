package store_test

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"rhea/internal/core"
	"rhea/internal/store"
)

// An App is kept as views are (KK 2026-10-26): a draft row, then the active
// row of the same version when its bundle is approved; served only once
// active, its newest active version; carried by the backup.
func TestAppRows(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	a := core.App{ID: "notatki", Version: 1, Title: "Notatki", Layer: "business",
		Data: []string{"note"}, Tabs: []core.AppTab{{View: "note-list"}, {Desk: "flow", Title: "Przepływ"}},
		Record: &core.AppRecord{Of: "note", Tabs: []core.AppTab{{View: "note-detail"}}}, Rules: []string{"take-note"}}
	if err := store.InsertAppRow(ctx, s.Pool, a, core.StatusDraft); err != nil {
		t.Fatal(err)
	}
	got, status, err := s.GetAppVersion(ctx, "notatki", 1)
	if err != nil || status != core.StatusDraft || !reflect.DeepEqual(got, a) {
		t.Fatalf("the draft: %+v %s %v", got, status, err)
	}
	if st, err := s.MemberStatus(ctx, core.Member{Kind: core.KindApp, Name: "notatki", Version: 1}); err != nil || st != core.StatusDraft {
		t.Fatalf("member status: %s %v", st, err)
	}
	if apps, err := s.LatestApps(ctx); err != nil || len(apps) != 0 {
		t.Fatalf("a draft is served to no one: %+v %v", apps, err)
	}
	if err := store.InsertAppRow(ctx, s.Pool, a, core.StatusActive); err != nil {
		t.Fatal(err)
	}
	a2 := a
	a2.Version, a2.Title = 2, "Notatki firmy"
	if err := store.InsertAppRow(ctx, s.Pool, a2, core.StatusDraft); err != nil {
		t.Fatal(err)
	}
	if _, status, err := s.GetAppVersion(ctx, "notatki", 1); err != nil || status != core.StatusActive {
		t.Fatalf("approved: %s %v", status, err)
	}
	apps, err := s.LatestApps(ctx)
	if err != nil || len(apps) != 1 || !reflect.DeepEqual(apps[0], a) {
		t.Fatalf("the newest active version alone, the draft above it not: %+v %v", apps, err)
	}
	if err := store.InsertAppRow(ctx, s.Pool, a, core.StatusActive); err == nil {
		t.Fatal("the same version approved twice")
	}

	// a company's own App (a kept report) keeps its company, through every
	// version, and never changes hands
	own := core.App{ID: "answer-7@company-95", Version: 1, Title: "RZiS", Layer: "business", Tabs: []core.AppTab{{View: "answer-7@company-95"}}, Company: "company-95"}
	if err := store.InsertAppRow(ctx, s.Pool, own, core.StatusActive); err != nil {
		t.Fatal(err)
	}
	if v, err := s.NextAppVersion(ctx, own.ID); err != nil || v != 2 {
		t.Fatalf("the next version of the company's App: %d %v", v, err)
	}
	if apps, err := s.LatestApps(ctx); err != nil || len(apps) != 2 || !reflect.DeepEqual(apps[0], own) {
		t.Fatalf("the company's App, with its company: %+v %v", apps, err)
	}
	other := own
	other.Version, other.Company = 2, "company-96"
	if err := store.InsertAppRow(ctx, s.Pool, other, core.StatusDraft); err == nil {
		t.Fatal("an App's company changed under its id")
	}

	var buf bytes.Buffer
	counts, err := s.Backup(ctx, &buf)
	if err != nil || counts["app"] != 4 {
		t.Fatalf("the backup carries every App row: %v %v", counts, err)
	}
}
