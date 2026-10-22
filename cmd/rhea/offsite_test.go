package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The signature is the one an independent signer gives: curl 8.18's
// --aws-sigv4 "aws:amz:waw:s3" signed this upload (captured against a
// local listener, 2026-10-22), and the kernel's signer must agree to the
// last hex digit.
func TestOffsiteSignsAsCurlDoes(t *testing.T) {
	o := offsite{region: "waw", key: "AKIDEXAMPLE", secret: "wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY"}
	at := time.Date(2026, 10, 22, 19, 47, 26, 0, time.UTC)
	got := o.authorization("PUT", "127.0.0.1:18765", "/rhea-backups/white01/rhea-20261022-023000.jsonl",
		"80407763b00bda5f1d24ab81a8e3cb00b5a8e1002644d070aed9bc212fce8921", at)
	want := "AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20261022/waw/s3/aws4_request, SignedHeaders=host;x-amz-content-sha256;x-amz-date, Signature=6a9fa03ccbda44aafdbf318e2b82419a548f5fcb8fb67c28cd28a1a4229f1b5e"
	if got != want {
		t.Fatalf("signature:\n got %s\nwant %s", got, want)
	}
}

// The night's backup leaves the machine whole or not at all (the
// whole-tree review of 2026-10-21, G1): a whole backup goes into the bucket
// under its own name with its hash signed; a cut one is refused before any
// request; a refusal by the bucket is an error, never a quiet success.
func TestOffsiteUploadsOnlyAWholeBackup(t *testing.T) {
	var got struct {
		path, hash, auth string
		body             []byte
	}
	status := http.StatusOK
	bucket := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.path, got.hash, got.auth = r.URL.Path, r.Header.Get("x-amz-content-sha256"), r.Header.Get("Authorization")
		got.body, _ = io.ReadAll(r.Body)
		w.WriteHeader(status)
	}))
	defer bucket.Close()
	base, _ := url.Parse(bucket.URL + "/rhea-backups/white01/")
	o := offsite{base: base, region: "waw", key: "AKIDEXAMPLE", secret: "s", client: bucket.Client(),
		now: func() time.Time { return time.Date(2026, 10, 22, 2, 30, 0, 0, time.UTC) }}
	dir := t.TempDir()
	whole := "{\"rhea_backup\":2}\n{\"t\":\"event\",\"r\":{\"event_id\":7}}\n{\"rhea_backup_end\":{\"counts\":{\"event\":1},\"max_event_id\":7}}\n"
	write := func(name, text string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}

	if _, err := uploadBackup(context.Background(), o, write("rhea-20261022-023000.jsonl", whole)); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(whole))
	if got.path != "/rhea-backups/white01/rhea-20261022-023000.jsonl" || got.hash != hex.EncodeToString(sum[:]) || string(got.body) != whole ||
		!strings.HasPrefix(got.auth, "AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20261022/waw/s3/aws4_request, ") {
		t.Fatalf("the upload: %+v", got)
	}

	got.path = ""
	cut := strings.Join(strings.SplitAfter(whole, "\n")[:2], "")
	if _, err := uploadBackup(context.Background(), o, write("rhea-cut.jsonl", cut)); err == nil || got.path != "" {
		t.Fatalf("a cut backup left the machine: %v %q", err, got.path)
	}
	status = http.StatusForbidden
	if _, err := uploadBackup(context.Background(), o, write("rhea-refused.jsonl", whole)); err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("the bucket's refusal: %v", err)
	}
	if err := o.put(context.Background(), "../rhea.jsonl", []byte(whole)); err == nil {
		t.Fatal("an object name climbing out of the prefix")
	}
}
