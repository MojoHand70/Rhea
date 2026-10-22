package store_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

// Backup and restore (DIRECTION 2026-10-15): the event log plus the
// definitions are the whole system. A world is built — types in two
// versions, a rule that fires, a draft at the gate in a bundle, a file
// through the door, events through verbs with their actors — backed up as
// JSON lines, restored into an empty database, and replayed: every object
// comes back identical with its provenance, every row of every append-only
// table is the same byte for byte, the file is under its hash, and the log
// continues at the next id. The object table never travels.
func TestBackupRestore(t *testing.T) {
	ctx := context.Background()
	a := storetest.New(t)
	xa := &exec.Executor{Store: a}
	if err := a.InsertObjectType(ctx, core.ObjectType{Name: "note", Version: 1, Domain: "work", LabelField: "text",
		Fields: []core.FieldDef{{Name: "text", Type: "string", Required: true}, {Name: "amount", Type: "money", Required: true}}}); err != nil {
		t.Fatal(err)
	}
	if err := a.InsertObjectType(ctx, core.ObjectType{Name: "note", Version: 2, Domain: "work", LabelField: "text", Amendable: []string{"amount"},
		Fields: []core.FieldDef{{Name: "text", Type: "string", Required: true}, {Name: "amount", Type: "money", Required: true}, {Name: "tag", Type: "string"}}}); err != nil {
		t.Fatal(err)
	}
	view, _ := json.Marshal(map[string]any{"object_type": "note", "columns": []map[string]string{{"field": "text", "label": "Text"}}})
	if err := a.InsertViewDef(ctx, core.ViewDef{ID: "notes", Version: 1, Notion: "list", Title: "Notes", Domain: "work", Function: "notes", Spec: view}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.InsertRuleVersion(ctx, core.Rule{ID: "note", Status: core.StatusActive, Priority: 100, EffectiveFrom: "2026-01-01", CreatedBy: "krzysztof",
		Description: "notes", Spec: core.RuleSpec{Match: core.Match{EventType: "note.taken"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text", "amount": "=$.amount"}}}}}); err != nil {
		t.Fatal(err)
	}
	draft, err := a.InsertRuleVersion(ctx, core.Rule{ID: "note", Status: core.StatusDraft, Priority: 100, EffectiveFrom: "2026-01-01", CreatedBy: "agent:test",
		Description: "notes, tagged", Spec: core.RuleSpec{Match: core.Match{EventType: "note.taken"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text", "amount": "=$.amount", "tag": "=$.tag"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.InsertBundle(ctx, core.Bundle{ID: "notes-tagged", CreatedBy: "agent:test", Description: "tag the notes",
		Members: []core.Member{{Kind: core.KindRule, Name: "note", Version: draft.Version}},
		Warrant: &core.Warrant{Basis: "pack", Citations: []string{"a test"}}, Question: "can notes carry a tag?"}); err != nil {
		t.Fatal(err)
	}
	for i, amount := range []string{"1000.00", "234.56", "5.00"} {
		if _, err := a.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: "note.taken", OccurredAt: fmt.Sprintf("2026-09-0%d", i+1), Actor: "cli:test",
			Payload: []byte(fmt.Sprintf(`{"text":"note %d","amount":"%s"}`, i+1, amount)), DedupKey: fmt.Sprintf("n%d", i)}); err != nil {
			t.Fatal(err)
		}
	}
	if _, errs := xa.ProcessPending(ctx); len(errs) > 0 {
		t.Fatal(errs)
	}
	// A file through the door: the blob under its hash, the event naming it.
	tr, err := xa.TriggerActivity(ctx, "receive_document", map[string]any{
		"file": map[string]any{"name": "scan.pdf", "media_type": "application/pdf", "bytes": []byte("%PDF-1.4 fake")}, "source": "scan", "occurred_at": "2026-09-04",
	}, "shell:krzysztof", "")
	if err != nil {
		t.Fatal(err)
	}
	var received struct {
		File struct {
			Hash string `json:"hash"`
		} `json:"file"`
	}
	ev, _ := a.GetEvent(ctx, tr.EventID)
	json.Unmarshal(ev.Payload, &received)

	// The backup: a header and one line per row, objects excluded.
	var buf bytes.Buffer
	counts, err := a.Backup(ctx, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if counts["object_type"] != 2 || counts["rule"] != 2 || counts["view_def"] != 1 || counts["bundle"] != 1 || counts["blob"] != 1 || counts["event"] < 5 || counts["object"] != 0 {
		t.Fatalf("backup counts: %v", counts)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if !strings.HasPrefix(lines[0], `{"rhea_backup":2,`) || strings.Contains(buf.String(), `"t":"object"`) {
		t.Fatalf("the backup's shape: %s", lines[0])
	}

	// The restore, into an empty database; then the replay.
	b := storetest.NewBare(t)
	if _, err := b.Restore(ctx, strings.NewReader("{\"rhea_backup\":1}\n{\"t\":\"object\",\"r\":{}}\n")); err == nil || !strings.Contains(err.Error(), "does not take") {
		t.Fatalf("a backup naming the object table: %v", err)
	}
	restored, err := b.Restore(ctx, bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(restored, counts) {
		t.Fatalf("restored %v, backed up %v", restored, counts)
	}
	xb := &exec.Executor{Store: b}
	if _, err := xb.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	// Identical objects with identical provenance.
	oa, _ := a.AllObjects(ctx)
	ob, _ := b.AllObjects(ctx)
	ja, _ := json.Marshal(oa)
	jb, _ := json.Marshal(ob)
	if len(oa) != 3 || !bytes.Equal(ja, jb) {
		t.Fatalf("objects differ after restore:\n%s\n%s", ja, jb)
	}
	// Identical rows: a backup of the restore is the backup, header aside.
	var again bytes.Buffer
	if _, err := b.Backup(ctx, &again); err != nil {
		t.Fatal(err)
	}
	if la, lb := lines[1:], strings.Split(strings.TrimSpace(again.String()), "\n")[1:]; !reflect.DeepEqual(la, lb) {
		t.Fatalf("the restored system of record differs from the backup:\n%d vs %d lines", len(la), len(lb))
	}
	// The file is under its hash; the verbs came along; the draft waits.
	if blob, err := b.GetBlob(ctx, received.File.Hash); err != nil || string(blob.Bytes) != "%PDF-1.4 fake" {
		t.Fatalf("the file after restore: %v %q", err, blob.Bytes)
	}
	if _, err := b.GetActiveActivity(ctx, "receive_document"); err != nil {
		t.Fatalf("the doors come with the backup: %v", err)
	}
	if latest, _ := b.GetRule(ctx, "note"); latest.Version != draft.Version || latest.Status != core.StatusDraft || latest.CreatedBy != "agent:test" {
		t.Fatalf("the draft at the gate after restore: %+v", latest)
	}
	// The log continues at the next id, and the restored world keeps working.
	last, _ := a.EventsByKind(ctx, core.KindRaw)
	id, err := b.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: "note.taken", OccurredAt: "2026-09-05", Actor: "cli:test",
		Payload: []byte(`{"text":"note 4","amount":"1.00"}`), DedupKey: "n4"})
	if err != nil {
		t.Fatal(err)
	}
	maxID := int64(0)
	for _, e := range last {
		maxID = max(maxID, e.ID)
	}
	derived, _ := a.EventsByKind(ctx, core.KindDerived)
	for _, e := range derived {
		maxID = max(maxID, e.ID)
	}
	if id <= maxID {
		t.Fatalf("the sequence did not move past the restored log: %d after %d", id, maxID)
	}
	if booked, errs := xb.ProcessPending(ctx); booked != 1 || len(errs) != 0 {
		t.Fatalf("the restored world books: %d %v", booked, errs)
	}
	// A restore refuses a system of record that is not empty.
	if _, err := b.Restore(ctx, bytes.NewReader(buf.Bytes())); err == nil || !strings.Contains(err.Error(), "empty system of record") {
		t.Fatalf("restore over a live log: %v", err)
	}
	_ = store.BackupCounts{}
}

// A log from before the company key (DECISIONS 2026-10-21): every fact was
// written acting for no company, so under isolation the books are the
// installation's rows — a signed-in owner sees them, but a door acting for
// the company that amends one of them is refused by the database, and the
// next company would see them. Restored for the company, the backup comes
// back as if every fact had been written acting for it: a door's fact and
// everything derived from it take the key; the pack's master data (written
// through no door), the kernel's own facts and an installation-wide door's
// stay the installation's; a fact that already names a company keeps it.
// Then the company's amendment books, the same fact twice is still one, and
// replay rebuilds the same objects under the key.
func TestRestoreForCompany(t *testing.T) {
	ctx := context.Background()
	const alfa, beta = "PL5250001111", "PL5260001246"
	a := storetest.New(t)
	xa := &exec.Executor{Store: a}
	for _, ot := range []core.ObjectType{
		{Name: "note", Version: 1, Domain: "work", LabelField: "text", Amendable: []string{"amount"},
			Fields: []core.FieldDef{{Name: "text", Type: "string", Required: true}, {Name: "amount", Type: "money", Required: true}}},
		{Name: "account", Version: 1, Domain: "finance", LabelField: "number",
			Fields: []core.FieldDef{{Name: "number", Type: "string", Required: true}}},
	} {
		if err := a.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	for _, r := range []core.Rule{
		{ID: "note", Spec: core.RuleSpec{Match: core.Match{EventType: "note.taken"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text", "amount": "=$.amount"}}}}},
		{ID: "note-corrected", Spec: core.RuleSpec{Match: core.Match{EventType: "note.corrected"},
			Effect: core.Effect{Amend: &core.AmendTemplate{Type: "note", Target: "=ref(note, text, $.text)", Set: map[string]string{"amount": "=$.amount"}}}}},
		{ID: "account", Spec: core.RuleSpec{Match: core.Match{EventType: "account.created"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "account", Fields: map[string]string{"number": "=$.number"}}}}},
	} {
		r.Status, r.Priority, r.EffectiveFrom, r.CreatedBy, r.Description = core.StatusActive, 100, "2026-01-01", "test", r.ID
		if _, err := a.InsertRuleVersion(ctx, r); err != nil {
			t.Fatal(err)
		}
	}
	door := func(x *exec.Executor, scope context.Context, typ, dedup string, payload map[string]any) exec.Triggered {
		t.Helper()
		tr, err := x.TriggerActivity(scope, "submit_event", map[string]any{
			"event_type": typ, "occurred_at": "2026-09-01", "payload": payload}, "krzysztof", dedup)
		if err != nil {
			t.Fatal(err)
		}
		return tr
	}
	// The old world: the books written acting for no company, the pack's
	// chart as the loader writes it, a sign-in the face reported, and a
	// second company already under its key.
	installation := store.WithInstallation(ctx)
	door(xa, installation, "note.taken", "rent-1", map[string]any{"text": "rent", "amount": "1000.00"})
	if _, err := a.AppendEvent(store.WithOwner(ctx), core.Event{Kind: core.KindRaw, Type: "account.created", OccurredAt: "2026-01-01",
		Payload: []byte(`{"number":"201"}`), DedupKey: "pack/test/account/201", Actor: "cli:test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := xa.TriggerActivity(store.WithCompany(ctx, alfa), "record_access", map[string]any{
		"person": "krzysztof", "outcome": "signed_in", "method": "passkey"}, "adapter:face", ""); err != nil {
		t.Fatal(err)
	}
	door(xa, store.WithCompany(ctx, beta), "note.taken", "", map[string]any{"text": "beta's", "amount": "5.00"})
	if _, errs := xa.ProcessPending(ctx); len(errs) > 0 {
		t.Fatal(errs)
	}
	var buf bytes.Buffer
	if _, err := a.Backup(ctx, &buf); err != nil {
		t.Fatal(err)
	}
	correct := func(x *exec.Executor) exec.Triggered {
		t.Helper()
		return door(x, store.WithCompany(ctx, alfa), "note.corrected", "", map[string]any{"text": "rent", "amount": "1200.00"})
	}

	// Restored as it is, the company's correction is refused by the database.
	plain := storetest.NewBare(t)
	if _, err := plain.Restore(ctx, bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatal(err)
	}
	xp := &exec.Executor{Store: plain}
	if _, err := xp.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	if tr := correct(xp); tr.Booked != 0 || len(tr.Errors) == 0 || !strings.Contains(fmt.Sprint(tr.Errors), "row-level security") {
		t.Fatalf("the unstamped books under the company's scope: booked %d, %v", tr.Booked, tr.Errors)
	}

	// Restored for the company: whose each fact is.
	if _, _, err := storetest.NewBare(t).RestoreFor(ctx, bytes.NewReader(buf.Bytes()), alfa+","+beta); err == nil {
		t.Fatal("a restore for a set of companies: the key is one company's")
	}
	b := storetest.NewBare(t)
	counts, stamped, err := b.RestoreFor(ctx, bytes.NewReader(buf.Bytes()), alfa)
	if err != nil {
		t.Fatal(err)
	}
	if counts["event"] == 0 || stamped.Company == 0 || stamped.Installation == 0 {
		t.Fatalf("restored %v, stamped %+v", counts, stamped)
	}
	xb := &exec.Executor{Store: b}
	if _, err := xb.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	raws, _ := b.EventsByKind(ctx, core.KindRaw)
	derived, _ := b.EventsByKind(ctx, core.KindDerived)
	byID := map[int64]core.Event{}
	for _, ev := range append(raws, derived...) {
		byID[ev.ID] = ev
	}
	rootOf := func(ev core.Event) core.Event {
		for ev.CauseEventID != nil {
			ev = byID[*ev.CauseEventID]
		}
		return ev
	}
	took, kept := 0, 0
	for _, ev := range append(raws, derived...) {
		root := rootOf(ev)
		var want string
		switch {
		case root.Type == "note.taken" && strings.Contains(string(root.Payload), "beta's"):
			want = beta
		case root.Type == "note.taken":
			want = alfa
		case root.Type == "account.created", root.Type == core.EventAccessRecorded, root.Actor == "kernel":
			want = ""
		default:
			t.Fatalf("a fact the test did not write: %s %s", root.Type, root.Actor)
		}
		if ev.Company != want {
			t.Fatalf("event %d %s (root %s): company %q, want %q", ev.ID, ev.Type, root.Type, ev.Company, want)
		}
		if want == alfa {
			took++
		} else if want == "" {
			kept++
		}
	}
	if took != stamped.Company || kept != stamped.Installation {
		t.Fatalf("the count said %+v; the log holds %d stamped, %d kept", stamped, took, kept)
	}
	objs, _ := b.AllObjects(ctx)
	for _, o := range objs {
		want := map[string]string{"rent": alfa, "beta's": beta}[fmt.Sprint(o.State["text"])]
		if o.Type == "account" {
			want = ""
		}
		if o.Company != want {
			t.Fatalf("%s %v: company %q, want %q", o.ID, o.State, o.Company, want)
		}
	}
	// The same objects as before the restore, the key aside.
	oa, _ := a.AllObjects(ctx)
	for i := range oa {
		oa[i].Company = ""
	}
	for i := range objs {
		objs[i].Company = ""
	}
	if ja, jb := mustJSON(t, oa), mustJSON(t, objs); ja != jb {
		t.Fatalf("objects differ after the restore for the company:\n%s\n%s", ja, jb)
	}

	// The company's correction books now; the same note through the same
	// door twice is still one fact; the other company's scope sees none of it.
	if tr := correct(xb); tr.Booked == 0 || len(tr.Errors) > 0 {
		t.Fatalf("the company's correction after the restore: booked %d, %v", tr.Booked, tr.Errors)
	}
	if _, err := xb.TriggerActivity(store.WithCompany(ctx, alfa), "submit_event", map[string]any{
		"event_type": "note.taken", "occurred_at": "2026-09-01", "payload": map[string]any{"text": "rent", "amount": "1000.00"}}, "krzysztof", "rent-1"); err == nil || !store.IsDuplicate(err) {
		t.Fatalf("the same fact twice after the restore: %v", err)
	}
	if ids, _ := b.FindObjectIDsByField(store.WithCompany(ctx, beta), "note", "text", "rent"); len(ids) != 0 {
		t.Fatalf("the other company sees the stamped books: %v", ids)
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	j, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(j)
}

// A backup that is not whole is refused (review 2026-10-21, G2): a file cut
// at a line boundary restored as a shorter log, a row dropped from its
// middle restored as a log with a hole, and a row carrying a column this
// database lacks — a backup written by a newer Rhea restored by an older
// one, the company key among them — was restored without it; each said it
// succeeded. Now a backup ends with what it carried, a restore refuses one
// that does not say so or says otherwise, and every field of a row must be
// a column here. A backup from before the end line still restores.
func TestARestoreRefusesABackupThatIsNotWhole(t *testing.T) {
	ctx := context.Background()
	a := storetest.New(t)
	xa := &exec.Executor{Store: a}
	if err := a.InsertObjectType(ctx, core.ObjectType{Name: "note", Version: 1, Domain: "work", LabelField: "text",
		Fields: []core.FieldDef{{Name: "text", Type: "string", Required: true}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.InsertRuleVersion(ctx, core.Rule{ID: "note", Status: core.StatusActive, Priority: 100, EffectiveFrom: "2026-01-01", CreatedBy: "krzysztof",
		Description: "notes", Spec: core.RuleSpec{Match: core.Match{EventType: "note.taken"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text"}}}}}); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 3; i++ {
		if _, err := a.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: "note.taken", OccurredAt: "2026-09-01", Actor: "cli:test",
			Payload: []byte(fmt.Sprintf(`{"text":"note %d"}`, i)), DedupKey: fmt.Sprintf("n%d", i), Company: "PL5250001111"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, errs := xa.ProcessPending(ctx); len(errs) > 0 {
		t.Fatal(errs)
	}
	var buf bytes.Buffer
	if _, err := a.Backup(ctx, &buf); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	join := func(ls []string) string { return strings.Join(ls, "\n") + "\n" }
	b := storetest.NewBare(t) // a refused restore leaves it empty for the next
	restore := func(text string) error {
		t.Helper()
		_, err := b.Restore(ctx, strings.NewReader(text))
		return err
	}

	if err := restore(join(lines[:len(lines)-1])); err == nil {
		t.Fatal("a backup cut at a line boundary restored")
	}
	var holed []string
	dropped := false
	for _, l := range lines {
		if !dropped && strings.Contains(l, `"t":"event"`) {
			dropped = true
			continue
		}
		holed = append(holed, l)
	}
	if err := restore(join(holed)); err == nil {
		t.Fatal("a backup with an event dropped from its middle restored")
	}
	if err := restore(strings.ReplaceAll(buf.String(), `{"event_id":`, `{"branch":"north","event_id":`)); err == nil || !strings.Contains(err.Error(), "branch") {
		t.Fatalf("a backup with a column this database lacks: %v", err)
	}
	if err := restore(buf.String()); err != nil {
		t.Fatalf("the whole backup: %v", err)
	}

	// The same check reads a backup without a database (rhea backup
	// -verify), and a backup from before the end line still restores.
	if _, err := store.VerifyBackup(strings.NewReader(join(lines[:len(lines)-1]))); err == nil {
		t.Fatal("a cut backup verified")
	}
	if counts, err := store.VerifyBackup(bytes.NewReader(buf.Bytes())); err != nil || counts["event"] == 0 {
		t.Fatalf("the whole backup verifies: %v %v", counts, err)
	}
	old := strings.Replace(join(lines[:len(lines)-1]), `{"rhea_backup":2,`, `{"rhea_backup":1,`, 1)
	if _, err := storetest.NewBare(t).Restore(ctx, strings.NewReader(old)); err != nil {
		t.Fatalf("a format-1 backup: %v", err)
	}
}
