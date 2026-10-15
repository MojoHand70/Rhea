package store_test

import (
	"context"
	"os"
	"testing"

	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

// TestRestoreLiveBackup restores a real backup file (RHEA_LIVE_BACKUP=path,
// written by rhea backup) into a throwaway database and replays it — the
// claim checked on a live log, by hand, whenever a backup is taken.
func TestRestoreLiveBackup(t *testing.T) {
	path := os.Getenv("RHEA_LIVE_BACKUP")
	if path == "" {
		t.Skip("set RHEA_LIVE_BACKUP=<file from rhea backup> to restore it into a throwaway database")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ctx := context.Background()
	b := storetest.NewBare(t)
	counts, err := b.Restore(ctx, f)
	if err != nil {
		t.Fatal(err)
	}
	objects, err := (&exec.Executor{Store: b}).Replay(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("restored %v; replay rebuilt %d objects", counts, len(objects))
}
