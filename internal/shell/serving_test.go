package shell_test

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"rhea/internal/exec"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

// What fails inside stays inside, and every request leaves a line (review
// 2026-10-21, G6): a handler's panic answered with the panic's own text,
// and an error the database raised reached the face as Postgres spelled
// it — tables, columns, constraints — while the server logged nothing of
// the requests it served. Now a failure answers in plain words, the detail
// goes to the server's log, and each request is one line there.
func TestFailuresStayInsideAndRequestsLeaveALine(t *testing.T) {
	ctx := context.Background()
	var logged bytes.Buffer
	before := log.Writer()
	log.SetOutput(&logged)
	defer log.SetOutput(before)

	// A panic: a server with no store.
	broken := httptest.NewServer((&shell.Server{}).Handler())
	defer broken.Close()
	res, err := http.Get(broken.URL + "/api/types")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != 500 || strings.Contains(string(body), "runtime error") || strings.Contains(string(body), "nil pointer") {
		t.Fatalf("a panic answered %d %s", res.StatusCode, body)
	}
	if !strings.Contains(logged.String(), "nil pointer") {
		t.Fatalf("the panic's detail is not in the log:\n%s", logged.String())
	}

	// A raw database error: a table the request reads is gone.
	s := storetest.New(t)
	if _, err := s.Pool.Exec(ctx, `DROP TABLE object_type CASCADE`); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer((&shell.Server{Store: s, Exec: &exec.Executor{Store: s}, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb")}).Handler())
	defer srv.Close()
	logged.Reset()
	res, err = http.Get(srv.URL + "/api/types")
	if err != nil {
		t.Fatal(err)
	}
	body, _ = io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != 500 || strings.Contains(string(body), "SQLSTATE") || strings.Contains(string(body), "object_type") {
		t.Fatalf("a database error answered %d %s", res.StatusCode, body)
	}
	if !strings.Contains(logged.String(), "SQLSTATE") {
		t.Fatalf("the database's words are not in the log:\n%s", logged.String())
	}
	if !strings.Contains(logged.String(), "GET /api/types 500") {
		t.Fatalf("the request left no line:\n%s", logged.String())
	}
}
