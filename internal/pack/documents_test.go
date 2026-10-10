package pack_test

import (
	"context"
	"path/filepath"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/store/storetest"
)

// The documents pack (DIRECTION 2026-10-10, "the file is the fact"): a file
// through the receive_document door becomes a received_document, waiting
// for a reader — pure data over the kernel's blob store, no code knows the
// type.
func TestDocumentsPack(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}

	sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", "common", "documents.json"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if sum.Types != 1 || sum.Rules != 1 || sum.Bundle != "pack-common-documents-v1" {
		t.Fatalf("load: %+v", sum)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-10-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve: %v %v", err, errs)
	}

	data := []byte(":20:STMT-2026-10\n:25:PL61109010140000071219812874\n:60F:C261001PLN1000,00\n")
	hash := core.HashBytes(data)
	tr, err := x.TriggerActivity(ctx, "receive_document", map[string]any{
		"file":   map[string]any{"name": "2026-10.sta", "media_type": "text/plain", "bytes": data},
		"source": "the bank's site", "occurred_at": "2026-10-11",
	}, "shell:kk", "")
	if err != nil {
		t.Fatal(err)
	}
	if tr.Booked != 1 || len(tr.Errors) != 0 {
		t.Fatalf("receive: booked %d, errors %v", tr.Booked, tr.Errors)
	}
	ids, err := s.FindObjectIDsByField(ctx, "received_document", "hash", hash)
	if err != nil || len(ids) != 1 {
		t.Fatalf("received_document by hash: %v %v", ids, err)
	}
	o, err := s.GetObject(ctx, ids[0])
	if err != nil {
		t.Fatal(err)
	}
	if o.State["name"] != "2026-10.sta" || o.State["status"] != "received" || o.State["source"] != "the bank's site" ||
		o.State["media_type"] != "text/plain" || o.State["size"] != float64(len(data)) {
		t.Fatalf("state = %v", o.State)
	}
	if o.SourceEventID != tr.EventID || o.RuleID != "register-received-document" {
		t.Fatalf("provenance = event %d rule %s", o.SourceEventID, o.RuleID)
	}
	// The bytes are there, whole, under the hash the object names.
	if b, err := s.GetBlob(ctx, o.State["hash"].(string)); err != nil || string(b.Bytes) != string(data) {
		t.Fatalf("blob: %v", err)
	}
	// Nothing waits: the file is explained as received.
	if wl, _ := s.UnmatchedRawEvents(ctx); len(wl) != 0 {
		t.Fatalf("worklist = %d", len(wl))
	}

	// A file without a stated source is refused at the door, so the rule
	// never meets a received document that cannot say where it came from.
	data2 := []byte("second file")
	if _, err := x.TriggerActivity(ctx, "receive_document", map[string]any{
		"file": map[string]any{"name": "note.txt", "media_type": "text/plain", "bytes": data2},
	}, "shell:kk", ""); err == nil {
		t.Fatal("a document without a source was received")
	}
	if tr, err := x.TriggerActivity(ctx, "receive_document", map[string]any{
		"file":   map[string]any{"name": "note.txt", "media_type": "text/plain", "bytes": data2},
		"source": "scan",
	}, "shell:kk", ""); err != nil || tr.Booked != 1 || len(tr.Errors) != 0 {
		t.Fatalf("second receive: %+v %v", tr, err)
	}
	if objs, _ := s.ObjectsByType(ctx, "received_document"); len(objs) != 2 {
		t.Fatalf("received documents = %d, want 2", len(objs))
	}
}
