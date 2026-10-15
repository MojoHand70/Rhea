package adapter_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/store"
)

// installs the documents pack and the KSeF inbound channel, approved.
func installIntake(t *testing.T, s *store.Store, x *exec.Executor) {
	t.Helper()
	ctx := context.Background()
	for _, p := range []string{filepath.Join("..", "..", "packs", "common", "documents.json"),
		filepath.Join("..", "..", "packs", "pl", "ksef-inbound.json")} {
		sum, err := pack.Load(ctx, s, p, "test")
		if err != nil {
			t.Fatal(err)
		}
		if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-01-01"); err != nil || len(errs) != 0 {
			t.Fatalf("approve %s: %v %v", sum.Bundle, err, errs)
		}
	}
}

func receive(t *testing.T, x *exec.Executor, name, media, source, date string, data []byte) {
	t.Helper()
	tr, err := x.TriggerActivity(context.Background(), "receive_document", map[string]any{
		"file": map[string]any{"name": name, "media_type": media, "bytes": data}, "source": source, "occurred_at": date,
	}, "test", "")
	if err != nil || tr.Booked != 1 || len(tr.Errors) != 0 {
		t.Fatalf("receive %s: %+v %v", name, tr, err)
	}
}

func jsonUnmarshal(b []byte, v any) error { return json.Unmarshal(b, v) }
