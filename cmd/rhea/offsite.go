package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"rhea/internal/store"
)

// The night's backup leaves the machine (the whole-tree review of
// 2026-10-21, G1: the only backup was the deploy's, on the server's own
// disk; KK chose OVH Object Storage, 2026-10-22). rhea backup -upload FILE
// verifies the file and puts it into an S3-style bucket, signed with AWS
// Signature Version 4 on the standard library alone: a server's curl may
// sign an upload wrongly or not at all, and the secret never stands on a
// command line where another user's ps reads it.

// offsite is the bucket a backup goes to.
type offsite struct {
	base        *url.URL // the bucket and its prefix: https://s3.waw.io.cloud.ovh.net/rhea-backups/white01
	region      string   // the bucket's region: waw, gra, de …
	key, secret string
	client      *http.Client
	now         func() time.Time
}

// offsiteFromEnv reads the bucket from the settings (RHEA_BACKUP_S3_URL,
// RHEA_BACKUP_S3_REGION, RHEA_BACKUP_S3_KEY, RHEA_BACKUP_S3_SECRET).
func offsiteFromEnv() (offsite, error) {
	o := offsite{region: os.Getenv("RHEA_BACKUP_S3_REGION"), key: os.Getenv("RHEA_BACKUP_S3_KEY"),
		secret: os.Getenv("RHEA_BACKUP_S3_SECRET"), client: &http.Client{Timeout: 10 * time.Minute}, now: time.Now}
	raw := os.Getenv("RHEA_BACKUP_S3_URL")
	if raw == "" || o.region == "" || o.key == "" || o.secret == "" {
		return o, fmt.Errorf("the bucket is not set: RHEA_BACKUP_S3_URL, RHEA_BACKUP_S3_REGION, RHEA_BACKUP_S3_KEY and RHEA_BACKUP_S3_SECRET")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return o, fmt.Errorf("RHEA_BACKUP_S3_URL %q: want https://host/bucket[/prefix]", raw)
	}
	o.base = u
	return o, nil
}

// uploadBackup puts a backup file into the bucket under its own name, only
// once it reads whole: a cut file never leaves as a backup.
func uploadBackup(ctx context.Context, o offsite, path string) (store.BackupCounts, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	counts, err := store.VerifyBackup(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%s is not uploaded: %w", path, err)
	}
	return counts, o.put(ctx, filepath.Base(path), body)
}

// put uploads one object, its payload's hash signed and checked by the
// bucket.
func (o offsite) put(ctx context.Context, name string, body []byte) error {
	if name == "" || strings.Trim(name, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._-") != "" || strings.HasPrefix(name, ".") {
		return fmt.Errorf("an object name is letters, digits, dots, dashes and underscores: %q", name)
	}
	u := *o.base
	u.Path = strings.TrimSuffix(u.Path, "/") + "/" + name
	sum := sha256.Sum256(body)
	payload := hex.EncodeToString(sum[:])
	at := o.now().UTC()
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("x-amz-content-sha256", payload)
	req.Header.Set("x-amz-date", at.Format("20060102T150405Z"))
	req.Header.Set("Authorization", o.authorization(http.MethodPut, u.Host, u.EscapedPath(), payload, at))
	res, err := o.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		said, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return fmt.Errorf("the bucket answered %s: %s", res.Status, strings.TrimSpace(string(said)))
	}
	return nil
}

// authorization is the request's AWS Signature Version 4 for S3 over the
// host, the payload's hash and the date — the three headers it signs.
func (o offsite) authorization(method, host, path, payload string, at time.Time) string {
	day, stamp := at.Format("20060102"), at.Format("20060102T150405Z")
	signed := "host;x-amz-content-sha256;x-amz-date"
	canonical := method + "\n" + path + "\n\n" +
		"host:" + host + "\n" + "x-amz-content-sha256:" + payload + "\n" + "x-amz-date:" + stamp + "\n\n" +
		signed + "\n" + payload
	scope := day + "/" + o.region + "/s3/aws4_request"
	h := sha256.Sum256([]byte(canonical))
	toSign := "AWS4-HMAC-SHA256\n" + stamp + "\n" + scope + "\n" + hex.EncodeToString(h[:])
	k := hmacSHA256([]byte("AWS4"+o.secret), day)
	for _, part := range []string{o.region, "s3", "aws4_request"} {
		k = hmacSHA256(k, part)
	}
	return "AWS4-HMAC-SHA256 Credential=" + o.key + "/" + scope + ", SignedHeaders=" + signed +
		", Signature=" + hex.EncodeToString(hmacSHA256(k, toSign))
}

func hmacSHA256(key []byte, data string) []byte {
	m := hmac.New(sha256.New, key)
	m.Write([]byte(data))
	return m.Sum(nil)
}
