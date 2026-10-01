package store_test

import (
	"context"
	"encoding/json"
	"testing"

	"rhea/internal/core"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

func testStore(t *testing.T) *store.Store { return storetest.New(t) }

func rawEvent(dedup string) core.Event {
	return core.Event{
		Kind:       core.KindRaw,
		Type:       "invoice.received",
		OccurredAt: "2026-09-15",
		Payload:    json.RawMessage(`{"customer":"ACME","currency":"PLN","lines":[{"amount":"200.00"}]}`),
		DedupKey:   dedup,
	}
}

func TestEventLog(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	id, err := s.AppendEvent(ctx, rawEvent("e1"))
	if err != nil || id == 0 {
		t.Fatalf("append: %d, %v", id, err)
	}

	// Dedup: same key refuses a second insert.
	if _, err := s.AppendEvent(ctx, rawEvent("e1")); err == nil {
		t.Fatal("duplicate dedup_key accepted")
	}

	// Worklist sees the unmatched raw event.
	wl, err := s.UnmatchedRawEvents(ctx)
	if err != nil || len(wl) != 1 || wl[0].ID != id {
		t.Fatalf("worklist = %v, %v; want the one raw event", wl, err)
	}

	// A derived event caused by it removes it from the worklist.
	_, err = s.AppendEvent(ctx, core.Event{
		Kind: core.KindDerived, Type: core.EventObjectMaterialized,
		OccurredAt: "2026-09-15", Payload: json.RawMessage(`{}`),
		CauseEventID: &id, RuleID: "r1", RuleVersion: 1,
	})
	if err != nil {
		t.Fatalf("append derived: %v", err)
	}
	wl, _ = s.UnmatchedRawEvents(ctx)
	if len(wl) != 0 {
		t.Fatalf("worklist should be empty, got %d", len(wl))
	}

	// A derived event without cause/rule violates the CHECK.
	_, err = s.AppendEvent(ctx, core.Event{
		Kind: core.KindDerived, Type: "x", OccurredAt: "2026-09-15",
		Payload: json.RawMessage(`{}`),
	})
	if err == nil {
		t.Fatal("derived event without provenance accepted")
	}
}

func TestAppendOnlyInvariant(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	id, err := s.AppendEvent(ctx, rawEvent("e1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Pool.Exec(ctx, `UPDATE event SET event_type = 'tampered' WHERE event_id = $1`, id); err == nil {
		t.Fatal("UPDATE on event succeeded — invariant 1 broken")
	}
	if _, err := s.Pool.Exec(ctx, `DELETE FROM event WHERE event_id = $1`, id); err == nil {
		t.Fatal("DELETE on event succeeded — invariant 1 broken")
	}
	if _, err := s.InsertRuleVersion(ctx, core.Rule{
		ID: "r1", Status: core.StatusDraft, Priority: 100, EffectiveFrom: "2026-01-01",
		CreatedBy: "human", Description: "d",
		Spec: core.RuleSpec{Match: core.Match{EventType: "x"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "t", Fields: map[string]string{"a": "b"}}}},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Pool.Exec(ctx, `UPDATE rule SET status = 'active'`); err == nil {
		t.Fatal("UPDATE on rule succeeded — invariant 1 broken")
	}
}

func TestRuleVersioning(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	spec := core.RuleSpec{
		Match: core.Match{EventType: "invoice.received"},
		Effect: core.Effect{Object: core.ObjectTemplate{
			Type: "invoice", Fields: map[string]string{"customer": "=$.customer"}}},
	}
	r := core.Rule{ID: "book-invoice", Status: core.StatusDraft, Priority: 100,
		EffectiveFrom: "2026-01-01", CreatedBy: "agent", Description: "draft", Spec: spec}

	v1, err := s.InsertRuleVersion(ctx, r)
	if err != nil || v1.Version != 1 {
		t.Fatalf("v1 = %+v, %v", v1.Version, err)
	}
	r.Status = core.StatusActive
	v2, err := s.InsertRuleVersion(ctx, r)
	if err != nil || v2.Version != 2 {
		t.Fatalf("v2 = %+v, %v", v2.Version, err)
	}

	latest, err := s.LatestRules(ctx)
	if err != nil || len(latest) != 1 || latest[0].Version != 2 || latest[0].Status != core.StatusActive {
		t.Fatalf("latest = %+v, %v", latest, err)
	}
	active, err := s.ActiveRules(ctx)
	if err != nil || len(active) != 1 {
		t.Fatalf("active = %+v, %v", active, err)
	}
	if active[0].Spec.Match.EventType != "invoice.received" {
		t.Fatalf("spec did not round-trip: %+v", active[0].Spec)
	}
}

func TestObjectTypeAndViewDef(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	ot := core.ObjectType{Name: "invoice", Version: 1, Domain: "finance", IsDocument: true,
		Fields: []core.FieldDef{{Name: "customer", Type: "string", Required: true}}}
	if err := s.InsertObjectType(ctx, ot); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetObjectType(ctx, "invoice")
	if err != nil || got.Name != "invoice" || len(got.Fields) != 1 {
		t.Fatalf("object type round-trip: %+v, %v", got, err)
	}

	vd := core.ViewDef{ID: "invoice-list", Version: 1, Notion: "list", Title: "Invoices",
		Domain: "finance", Function: "documents",
		Spec: json.RawMessage(`{"object_type":"invoice","columns":[{"field":"customer","label":"Customer"}]}`)}
	if err := s.InsertViewDef(ctx, vd); err != nil {
		t.Fatal(err)
	}
	vds, err := s.LatestViewDefs(ctx)
	if err != nil || len(vds) != 1 || vds[0].Notion != "list" {
		t.Fatalf("view defs: %+v, %v", vds, err)
	}
}
