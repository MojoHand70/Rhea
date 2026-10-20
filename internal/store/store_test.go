package store_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

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
	if _, err := s.Pool.Exec(ctx, `TRUNCATE event CASCADE`); err == nil {
		t.Fatal("TRUNCATE on event succeeded — invariant 1 broken")
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
	// The activity table joins the append-only family: a status transition is
	// a new version row, never an edit. Seeded builtins give us rows to poke.
	if _, err := s.Pool.Exec(ctx, `UPDATE activity SET status = 'superseded'`); err == nil {
		t.Fatal("UPDATE on activity succeeded — invariant 1 broken")
	}
	if _, err := s.Pool.Exec(ctx, `TRUNCATE activity`); err == nil {
		t.Fatal("TRUNCATE on activity succeeded — invariant 1 broken")
	}
	if _, err := s.Pool.Exec(ctx, `DELETE FROM activity`); err == nil {
		t.Fatal("DELETE on activity succeeded — invariant 1 broken")
	}
	// Bundles and the gated definitions too: approval appends the active row,
	// it never flips a draft in place.
	if _, err := s.Pool.Exec(ctx, `INSERT INTO bundle (bundle_id, version, status, description, members, created_by)
		VALUES ('b', 1, 'draft', 'd', '[]', 'test')`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Pool.Exec(ctx, `UPDATE bundle SET status = 'active'`); err == nil {
		t.Fatal("UPDATE on bundle succeeded — invariant 1 broken")
	}
	if _, err := s.Pool.Exec(ctx, `INSERT INTO object_type (name, version, status, spec)
		VALUES ('t', 1, 'draft', '{}')`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Pool.Exec(ctx, `UPDATE object_type SET status = 'active'`); err == nil {
		t.Fatal("UPDATE on object_type succeeded — invariant 1 broken")
	}
	// The blob store joins the family (DIRECTION 2026-10-10): a file is kept
	// whole under its hash, and never changed or removed.
	data := []byte("a scanned page")
	if _, err := s.PutBlob(ctx, core.HashBytes(data), "image/png", data); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Pool.Exec(ctx, `UPDATE blob SET bytes = 'x'`); err == nil {
		t.Fatal("UPDATE on blob succeeded — invariant 1 broken")
	}
	if _, err := s.Pool.Exec(ctx, `DELETE FROM blob`); err == nil {
		t.Fatal("DELETE on blob succeeded — invariant 1 broken")
	}
}

// The file is the fact: bytes are kept once under their content hash, the
// hash is checked against the bytes, and a document.received event is
// refused until the file it names is in the store.
func TestBlobStoreAndTheFileFirst(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	data := []byte("<Faktura>…</Faktura>")
	hash := core.HashBytes(data)

	received := func(h string) core.Event {
		return core.Event{Kind: core.KindRaw, Type: core.EventDocumentReceived, OccurredAt: "2026-10-11",
			Payload:  json.RawMessage(`{"file":{"hash":"` + h + `","name":"fv.xml","media_type":"application/xml","size":21},"source":"email"}`),
			DedupKey: "doc/" + h}
	}
	if _, err := s.AppendEvent(ctx, received(hash)); err == nil || !strings.Contains(err.Error(), "does not hold") {
		t.Fatalf("document.received without its file was appended: %v", err)
	}
	if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: core.EventDocumentReceived,
		OccurredAt: "2026-10-11", Payload: json.RawMessage(`{"note":"no file here"}`)}); err == nil ||
		!strings.Contains(err.Error(), "names no file") {
		t.Fatalf("document.received naming no file was appended: %v", err)
	}

	if _, err := s.PutBlob(ctx, "sha256:0000", "application/xml", data); err == nil {
		t.Fatal("a wrong hash was accepted")
	}
	stored, err := s.PutBlob(ctx, hash, "application/xml", data)
	if err != nil || !stored {
		t.Fatalf("first put: %v %v", stored, err)
	}
	stored, err = s.PutBlob(ctx, hash, "application/xml", data)
	if err != nil || stored {
		t.Fatalf("second put should be a no-op: %v %v", stored, err)
	}
	b, err := s.GetBlob(ctx, hash)
	if err != nil || string(b.Bytes) != string(data) || b.MediaType != "application/xml" || b.Size != int64(len(data)) {
		t.Fatalf("get: %+v %v", b, err)
	}
	if _, err := s.GetBlob(ctx, "sha256:nothing"); err == nil {
		t.Fatal("an unknown hash was found")
	}

	if _, err := s.AppendEvent(ctx, received(hash)); err != nil {
		t.Fatalf("with the file in the store: %v", err)
	}
}

// A batch lands whole or not at all; known events are skipped and counted.
func TestAppendEventsIsOneTransaction(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	if _, err := s.AppendEvent(ctx, rawEvent("e1")); err != nil {
		t.Fatal(err)
	}
	count := func() int {
		evs, err := s.EventsByKind(ctx, core.KindRaw)
		if err != nil {
			t.Fatal(err)
		}
		n := 0
		for _, ev := range evs {
			if ev.Type == "invoice.received" {
				n++
			}
		}
		return n
	}
	bad := rawEvent("e3")
	bad.OccurredAt = "not a date"
	if _, _, err := s.AppendEvents(ctx, []core.Event{rawEvent("e2"), bad}); err == nil {
		t.Fatal("a batch with a bad event was accepted")
	}
	if count() != 1 {
		t.Fatalf("a failed batch left %d events, want 1", count())
	}
	appended, duplicates, err := s.AppendEvents(ctx, []core.Event{rawEvent("e1"), rawEvent("e2"), rawEvent("e3")})
	if err != nil || appended != 2 || duplicates != 1 {
		t.Fatalf("batch = %d appended, %d duplicates, %v", appended, duplicates, err)
	}
	if count() != 3 {
		t.Fatalf("after the batch: %d events, want 3", count())
	}
	if _, err := s.AppendEvent(ctx, rawEvent("e1")); !store.IsDuplicate(err) {
		t.Fatalf("the single append still reports a duplicate: %v", err)
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

// TestDraftDoesNotSuspendActive: the executing rule set is the newest ACTIVE
// version per rule. Work in progress (a draft) must not stop live booking;
// a latest version of superseded retires the rule.
func TestDraftDoesNotSuspendActive(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	spec := core.RuleSpec{
		Match: core.Match{EventType: "x"},
		Effect: core.Effect{Object: core.ObjectTemplate{
			Type: "t", Fields: map[string]string{"f": "v"}}},
	}
	ins := func(status string) {
		t.Helper()
		if _, err := s.InsertRuleVersion(ctx, core.Rule{
			ID: "r1", Status: status, Priority: 100, EffectiveFrom: "2026-01-01",
			CreatedBy: "test", Description: "r1", Spec: spec,
		}); err != nil {
			t.Fatal(err)
		}
	}
	ins(core.StatusDraft)  // v1
	ins(core.StatusActive) // v2
	active, err := s.ActiveRules(ctx)
	if err != nil || len(active) != 1 || active[0].Version != 2 {
		t.Fatalf("active = %+v, %v; want v2", active, err)
	}

	ins(core.StatusDraft) // v3 in progress
	active, _ = s.ActiveRules(ctx)
	if len(active) != 1 || active[0].Version != 2 {
		t.Fatalf("draft suspended the rule: active = %+v", active)
	}

	ins(core.StatusSuperseded) // v4 retires it
	active, _ = s.ActiveRules(ctx)
	if len(active) != 0 {
		t.Fatalf("superseded rule still active: %+v", active)
	}
}

// TestProjectionNotices: the announce channel honors transaction boundaries —
// a notice sent inside a transaction reaches listeners only on commit.
func TestProjectionNotices(t *testing.T) {
	s := testStore(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	notices, err := s.ProjectionNotices(ctx)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if err := store.NotifyProjection(ctx, tx, "invoice,posting"); err != nil {
		t.Fatal(err)
	}
	select {
	case p := <-notices:
		t.Fatalf("notice %q arrived before commit", p)
	case <-time.After(150 * time.Millisecond):
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case p := <-notices:
		if p != "invoice,posting" {
			t.Fatalf("notice = %q", p)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no notice after commit")
	}
}
