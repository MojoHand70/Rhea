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
	if !strings.HasPrefix(lines[0], `{"rhea_backup":1,`) || strings.Contains(buf.String(), `"t":"object"`) {
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
