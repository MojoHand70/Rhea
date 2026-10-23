package store_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

// A file's bytes are what its name says, whoever writes the row (review
// 2026-10-21, B4): the store checked the hash when a door kept a file, but
// a restore wrote file rows as the backup carried them, so a backup whose
// bytes were changed came back with the changed bytes under the old name —
// the evidence a fact names, silently other. Now the database checks every
// file row against its hash, and the restore is refused.
func TestARestoredFileIsWhatItsNameSays(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	data := []byte("%PDF-1.7 the invoice as received")
	if _, err := s.PutBlob(store.WithOwner(ctx), core.HashBytes(data), "application/pdf", data); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := s.Backup(ctx, &buf); err != nil {
		t.Fatal(err)
	}
	want, other := hex.EncodeToString(data), hex.EncodeToString([]byte("%PDF-1.7 the invoice as altered!"))
	if !strings.Contains(buf.String(), want) {
		t.Fatal("the backup does not carry the file's bytes as hex")
	}
	altered := strings.Replace(buf.String(), want, other, 1)
	if _, err := storetest.NewBare(t).Restore(ctx, strings.NewReader(altered)); err == nil {
		t.Fatal("a file whose bytes are not its hash's was restored")
	}
	if _, err := storetest.NewBare(t).Restore(ctx, strings.NewReader(buf.String())); err != nil {
		t.Fatalf("the backup as written: %v", err)
	}
}
