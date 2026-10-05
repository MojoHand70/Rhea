package shell_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

// TestLiveScreens: the executor is the single writer of the object cache, so
// it is the single announcer. A browser tab holds one SSE stream; when an
// event books, the stream carries a projection notice naming the touched
// types — the signal live view tabs re-render on. Nobody refreshes.
func TestLiveScreens(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	seedFromFile(t, s, "finance.json")

	x := &exec.Executor{Store: s}
	srv := &shell.Server{Store: s, Exec: x,
		DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb")}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	// An active rule, approved the ordinary way.
	if _, err := s.InsertRuleVersion(ctx, core.Rule{
		ID: "book-pln-invoice", Status: core.StatusDraft, Priority: 100,
		EffectiveFrom: "2026-01-01", CreatedBy: "human", Description: "book invoices",
		Spec: core.RuleSpec{
			Match: core.Match{EventType: "invoice.received"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "invoice",
				Fields: map[string]string{
					"customer": "=$.customer", "issue_date": "=$.issue_date",
					"currency": "=$.currency", "total": "=sum($.lines[*].amount)"}}},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := x.ApproveRule(ctx, "book-pln-invoice", "test", "2026-01-01"); err != nil {
		t.Fatal(err)
	}

	// Open the stream first: once the headers arrive, the LISTEN is in place.
	streamCtx, stop := context.WithCancel(ctx)
	defer stop()
	req, err := http.NewRequestWithContext(streamCtx, "GET", ts.URL+"/api/live", nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if ct := res.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content type = %q", ct)
	}
	lines := make(chan string, 32)
	go func() {
		sc := bufio.NewScanner(res.Body)
		for sc.Scan() {
			lines <- sc.Text()
		}
		close(lines)
	}()

	// A submitted event books — and announces itself.
	body, _ := json.Marshal(map[string]any{
		"event_type": "invoice.received", "occurred_at": "2026-09-15", "dedup_key": "live-1",
		"payload": map[string]any{"customer": "ACME", "issue_date": "2026-09-15",
			"currency": "PLN", "lines": []map[string]any{{"amount": "100.00"}}},
	})
	if _, err := ts.Client().Post(ts.URL+"/api/events", "application/json", bytes.NewReader(body)); err != nil {
		t.Fatal(err)
	}

	deadline := time.After(5 * time.Second)
	sawEvent, sawData := false, false
	for !(sawEvent && sawData) {
		select {
		case l, ok := <-lines:
			if !ok {
				t.Fatal("stream closed early")
			}
			if l == "event: projection" {
				sawEvent = true
			}
			if strings.HasPrefix(l, "data: ") && strings.Contains(l, "invoice") {
				sawData = true
			}
		case <-deadline:
			t.Fatal("no projection notice on the stream")
		}
	}
}
