package store_test

import (
	"context"
	"encoding/json"
	"fmt"
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
	// Every append-only table, both triggers, enabled always (a superuser's
	// session_replication_role cannot switch them off), and TRUNCATE refused.
	for _, table := range []string{"event", "rule", "object_type", "view_def", "activity", "bundle", "blob"} {
		for _, trig := range []string{table + "_append_only", table + "_no_truncate"} {
			var enabled string
			if err := s.Pool.QueryRow(ctx, `SELECT tgenabled FROM pg_trigger WHERE tgname = $1`, trig).Scan(&enabled); err != nil || enabled != "A" {
				t.Fatalf("trigger %s: %q %v — not there or not enabled always", trig, enabled, err)
			}
		}
		if _, err := s.Pool.Exec(ctx, `TRUNCATE `+table+` CASCADE`); err == nil {
			t.Fatalf("TRUNCATE on %s succeeded — invariant 1 broken", table)
		}
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

// Invariant 9, enforced in the database: a scoped connection runs as
// rhea_app under row-level security and sees its scope's rows and the
// installation's — never another company's — and cannot write outside its
// scope, whatever SQL it runs. An office's scope is a set; the kernel's
// unscoped work sees the whole log.
func TestIsolationInTheDatabase(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	for i, c := range []string{"PL5250001111", "PL5260001246", ""} {
		ev := rawEvent(fmt.Sprintf("iso-%d", i))
		ev.Company = c
		id, err := s.AppendEvent(ctx, ev)
		if err != nil {
			t.Fatal(err)
		}
		if err := store.InsertObject(ctx, s.Pool, core.Object{ID: fmt.Sprintf("thing-%d", i), Type: "thing", TypeVersion: 1,
			State: map[string]any{"n": i}, SourceEventID: id, RuleID: "r", RuleVersion: 1, Company: c}); err != nil {
			t.Fatal(err)
		}
	}
	count := func(ctx context.Context, table string) int {
		t.Helper()
		var n int
		if err := s.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM "+table).Scan(&n); err != nil {
			t.Fatalf("%s: %v", table, err)
		}
		return n
	}
	alfa := store.WithCompany(ctx, "PL5250001111")
	office := store.WithScope(ctx, []string{"PL5260001246", "PL5250001111"})
	if count(ctx, "object") != 3 || count(ctx, "event") < 3 {
		t.Fatal("the unscoped kernel sees the whole log")
	}
	if count(alfa, "object") != 2 { // alfa's own and the installation's
		t.Fatalf("alfa sees %d objects through raw SQL", count(alfa, "object"))
	}
	var seen []string
	rows, _ := s.Pool.Query(alfa, `SELECT COALESCE(company, '') FROM object ORDER BY object_id`)
	for rows.Next() {
		var c string
		rows.Scan(&c)
		seen = append(seen, c)
	}
	rows.Close()
	if fmt.Sprint(seen) != "[PL5250001111 ]" {
		t.Fatalf("alfa's rows: %v", seen)
	}
	if count(office, "object") != 3 {
		t.Fatalf("the office sees %d objects of its two clients and the installation", count(office, "object"))
	}
	// A scoped connection cannot write another company's fact, through the
	// store or through raw SQL.
	beta := rawEvent("iso-beta-by-alfa")
	beta.Company = "PL5260001246"
	if _, err := s.AppendEvent(alfa, beta); err == nil {
		t.Fatal("alfa appended beta's fact")
	}
	if _, err := s.Pool.Exec(alfa, `INSERT INTO object (object_id, object_type, type_version, state, source_event_id, rule_id, rule_version, company)
		VALUES ('thing-x', 'thing', 1, '{}', 1, 'r', 1, 'PL5260001246')`); err == nil {
		t.Fatal("alfa wrote beta's object through raw SQL")
	}
	// The connection comes back to the pool as the owner: the next unscoped
	// use sees everything again.
	if count(ctx, "object") != 3 {
		t.Fatal("the scope leaked into the pool")
	}
	if s, _ := store.ScopeOf(store.WithScope(ctx, []string{"b", " a ", "b", ""})); fmt.Sprint(s) != "[a b]" {
		t.Fatalf("scope = %v", s)
	}
}

// A row's company comes from the data, never the caller: an office acting
// for no client cannot write (the store and the policy both refuse), a
// derived event keeps its chain's company under any caller, a company's
// dedup key is its own, and the installation's scope sees and writes the
// installation's rows alone.
func TestCompanyComesFromTheData(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	office := store.WithScope(ctx, []string{"PL5250001111", "PL5260001246"})
	ev := rawEvent("office-1")
	if _, err := s.AppendEvent(office, ev); err == nil || !strings.Contains(err.Error(), "name the company") {
		t.Fatalf("an office acting for none wrote a fact: %v", err)
	}
	// the data decides within the scope: the desk acting for alfa may
	// write a fact that says beta (a client of the office), never gamma's
	ev.Company = "PL5260001246"
	if _, err := s.AppendEvent(store.ActFor(office, "PL5250001111"), ev); err != nil {
		t.Fatalf("the office acting for alfa could not write beta's fact: %v", err)
	}
	gamma := rawEvent("office-gamma")
	gamma.Company = "PL7010001111"
	if _, err := s.AppendEvent(store.ActFor(office, "PL5250001111"), gamma); err == nil {
		t.Fatal("the office wrote a stranger's fact")
	}
	id, err := s.AppendEvent(store.ActFor(office, "PL5260001246"), rawEvent("office-2"))
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := s.GetEvent(ctx, id); got.Company != "PL5260001246" {
		t.Fatalf("the office's fact for beta carries %q", got.Company)
	}
	// A derived event with no company stays the installation's whatever
	// scope booked it: its explanation must be visible wherever its object is.
	alfa := store.WithCompany(ctx, "PL5250001111")
	root, _ := s.AppendEvent(store.WithOwner(ctx), rawEvent("inst-root"))
	derived := core.Event{Kind: core.KindDerived, Type: core.EventObjectMaterialized, OccurredAt: "2026-09-01",
		Payload: []byte(`{"object_id":"x","object_type":"x","type_version":1,"state":{}}`), CauseEventID: &root, RuleID: "r", RuleVersion: 1}
	if _, err := s.AppendEvent(alfa, derived); err == nil {
		t.Fatal("alfa's scope wrote an installation row")
	}
	did, err := s.AppendEvent(store.WithInstallation(ctx), derived)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := s.GetEvent(ctx, did); got.Company != "" {
		t.Fatalf("the derived event took a caller's company: %q", got.Company)
	}
	// The same key in two companies is two facts; the same key in one is one.
	beta := store.WithCompany(ctx, "PL5260001246")
	if _, err := s.AppendEvent(alfa, rawEvent("shared-key")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AppendEvent(beta, rawEvent("shared-key")); err != nil {
		t.Fatalf("beta's fact with alfa's key: %v", err)
	}
	if _, err := s.AppendEvent(beta, rawEvent("shared-key")); !store.IsDuplicate(err) {
		t.Fatalf("beta's own key twice: %v", err)
	}
	// The installation's scope sees its own rows and none of the companies'.
	var n int
	if err := s.Pool.QueryRow(store.WithInstallation(ctx), `SELECT COUNT(*) FROM event WHERE company IS NOT NULL AND company <> ''`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("the installation's scope sees %d company rows: %v", n, err)
	}
	if cs, _ := s.Companies(ctx); fmt.Sprint(cs) != "[PL5250001111 PL5260001246]" {
		t.Fatalf("companies: %v", cs)
	}
}
