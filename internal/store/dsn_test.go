package store_test

import (
	"context"
	"strings"
	"testing"

	"rhea/internal/store"
)

// A database's password never reaches a log (review 2026-10-21, G6): a
// failed connection said where it could not connect, password and all.
func TestADatabasesPasswordStaysOutOfErrors(t *testing.T) {
	_, err := store.Open(context.Background(), "postgres://rhea:topsecret@127.0.0.1:1/rhea?connect_timeout=1")
	if err == nil {
		t.Fatal("connected to nothing")
	}
	if strings.Contains(err.Error(), "topsecret") {
		t.Fatalf("the password in an error: %v", err)
	}
	for dsn, want := range map[string]string{
		"postgres://rhea:topsecret@host:5432/db":        "postgres://rhea:xxxxx@host:5432/db",
		"host=h user=rhea password=topsecret dbname=db": "host=h user=rhea password=xxxxx dbname=db",
		"postgres://host/db":                            "postgres://host/db",
	} {
		if got := store.RedactDSN(dsn); got != want {
			t.Fatalf("RedactDSN(%q) = %q, want %q", dsn, got, want)
		}
	}
}
