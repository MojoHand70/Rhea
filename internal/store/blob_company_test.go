package store_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

// Files kept before they carried a company become their company's, never
// every company's (review 2026-10-21, B4): a database whose files have no
// company — white01's, the books restored for one company — would show its
// company none of them once a file is seen only where its company is, and
// before that showed every company all of them. At init, and at the end of
// a restore of a backup written before, a file with no company becomes the
// company's whose documents name it; the row with no company stays, seen
// in the installation's own scope alone.
func TestFilesFromBeforeTakeTheirCompany(t *testing.T) {
	ctx := context.Background()
	const alfa, beta = "PL5250001111", "PL5260001246"
	s := storetest.New(t)
	data := []byte("%PDF-1.7 alfa's scan")
	hash := core.HashBytes(data)
	// the old world: the file kept with no company, the fact naming it alfa's
	if _, err := s.PutBlob(store.WithOwner(ctx), hash, "application/pdf", data); err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(map[string]any{"file": map[string]any{"hash": hash, "name": "scan.pdf", "media_type": "application/pdf", "size": len(data)},
		"source": "mail", "occurred_at": "2026-09-01"})
	if _, err := s.AppendEvent(store.WithOwner(ctx), core.Event{Kind: core.KindRaw, Type: core.EventDocumentReceived, OccurredAt: "2026-09-01",
		Payload: payload, Actor: "krzysztof", ActivityName: "receive_document", ActivityVersion: 1, Company: alfa}); err != nil {
		t.Fatal(err)
	}
	var old bytes.Buffer
	if _, err := s.Backup(ctx, &old); err != nil {
		t.Fatal(err)
	}
	seen := func(st *store.Store, scope context.Context) bool {
		t.Helper()
		b, err := st.GetBlob(scope, hash)
		return err == nil && string(b.Bytes) == string(data)
	}
	if seen(s, store.WithCompany(ctx, alfa)) || seen(s, store.WithCompany(ctx, beta)) {
		t.Fatal("a file with no company is seen in a company's scope")
	}

	if err := s.InitSchema(ctx); err != nil {
		t.Fatal(err)
	}
	if !seen(s, store.WithCompany(ctx, alfa)) || seen(s, store.WithCompany(ctx, beta)) || !seen(s, store.WithInstallation(ctx)) {
		t.Fatalf("after init: alfa %v, beta %v, the installation %v", seen(s, store.WithCompany(ctx, alfa)), seen(s, store.WithCompany(ctx, beta)), seen(s, store.WithInstallation(ctx)))
	}
	if err := s.InitSchema(ctx); err != nil {
		t.Fatalf("init again: %v", err)
	}

	// a backup from before files carried a company: no company in its rows
	var lines []string
	stripped := 0
	for _, l := range strings.Split(strings.TrimSpace(old.String()), "\n") {
		var line map[string]json.RawMessage
		if err := json.Unmarshal([]byte(l), &line); err != nil {
			t.Fatal(err)
		}
		if string(line["t"]) == `"blob"` {
			var row map[string]json.RawMessage
			json.Unmarshal(line["r"], &row)
			if _, ok := row["company"]; !ok {
				t.Fatalf("a file's row carries no company column: %s", line["r"])
			}
			delete(row, "company")
			line["r"], _ = json.Marshal(row)
			stripped++
		}
		b, _ := json.Marshal(line)
		lines = append(lines, string(b))
	}
	if stripped != 1 {
		t.Fatalf("%d files in the backup", stripped)
	}
	r := storetest.NewBare(t)
	if _, err := r.Restore(ctx, strings.NewReader(strings.Join(lines, "\n")+"\n")); err != nil {
		t.Fatal(err)
	}
	if !seen(r, store.WithCompany(ctx, alfa)) || seen(r, store.WithCompany(ctx, beta)) {
		t.Fatalf("after the restore: alfa %v, beta %v", seen(r, store.WithCompany(ctx, alfa)), seen(r, store.WithCompany(ctx, beta)))
	}
}
