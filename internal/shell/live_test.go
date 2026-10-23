package shell_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net"
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
	ts := serve(t, srv.Handler())

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

// serve starts a test server for h, closed when the test ends.
func serve(t *testing.T, h http.Handler) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(h)
	t.Cleanup(func() { closeServer(ts) })
	return ts
}

// closeServer ends a test server whatever its clients hold open: httptest's
// Close waits for every request, and a live stream never ends by itself. The
// listener goes first, so nothing new arrives while the open ones are dropped.
func closeServer(ts *httptest.Server) {
	ts.Listener.Close()
	ts.CloseClientConnections()
	ts.Close()
}

// TestAStrayStreamDoesNotHoldATestServer (2026-10-23): a browser left
// running on the machine — the face's walk left 28 — redials its dead
// server's port for the live stream; a test server handed that port serves
// it the stream, which never ends by itself, and closing the server waited
// for it until the suite's 20-minute limit. Closing a test server must not
// wait on a stream nobody in the test opened.
func TestAStrayStreamDoesNotHoldATestServer(t *testing.T) {
	s := storetest.New(t)
	srv := &shell.Server{Store: s, Exec: &exec.Executor{Store: s}, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb")}
	ts := httptest.NewServer(srv.Handler())
	// the stranger: a bare connection that asks for the stream and stays
	stranger, err := net.Dial("tcp", ts.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer stranger.Close()
	if _, err := stranger.Write([]byte("GET /api/live HTTP/1.1\r\nHost: localhost\r\n\r\n")); err != nil {
		t.Fatal(err)
	}
	if line, err := bufio.NewReader(stranger).ReadString('\n'); err != nil || !strings.HasPrefix(line, "HTTP/1.1 200") {
		t.Fatalf("the stream did not open: %q %v", line, err)
	}
	closed := make(chan struct{})
	go func() { closeServer(ts); close(closed) }()
	select {
	case <-closed:
	case <-time.After(5 * time.Second):
		t.Fatal("closing the test server waits on a stranger's live stream")
	}
}
