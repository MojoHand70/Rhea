package shell_test

import (
	"encoding/base64"
	"testing"

	"rhea/internal/core"
)

// A file is the company's that received it (review 2026-10-21, B4): files
// were kept by their hash alone, with no company, and served to any scope
// in which some document named the hash. A person of company A who wrote a
// fact of an issued document naming B's file — the hash is all it takes —
// was then served B's file. Now a file is kept for the company whose door
// received it, and the database shows it only where that company is: the
// same bytes received by both are each company's own, and a fact naming
// another company's file opens nothing.
func TestAFileIsTheCompanysThatReceivedIt(t *testing.T) {
	w := newGrantsWorld(t)
	a, b := "PL5250001111", "PL5260001246"
	w.person("Krzysztof Kowal", "kk", a, "owner")
	w.person("Olga Nowa", "olga", b, "owner")
	receive := func(person, company string, data []byte) string {
		t.Helper()
		if code, out := w.door(person, company, "receive_document", map[string]any{
			"file":   map[string]any{"name": "scan.pdf", "media_type": "application/pdf", "content_base64": base64.StdEncoding.EncodeToString(data)},
			"source": "mail"}); code != 200 {
			t.Fatalf("%s receives a file for %s: %d %v", person, company, code, out)
		}
		return core.HashBytes(data)
	}
	read := func(person, company, hash string) (int, string) {
		t.Helper()
		code, body := walkCall(t, w, person, company, "GET", "/api/blobs/"+hash, nil)
		return code, body
	}

	beta := []byte("%PDF-1.7 beta's contract")
	hash := receive("olga", b, beta)
	if code, body := read("olga", b, hash); code != 200 || body != string(beta) {
		t.Fatalf("beta reads its own file: %d %q", code, body)
	}

	// alfa writes the fact of an issued document naming beta's file
	var out map[string]any
	if code := w.call("kk", a, "POST", "/api/events", map[string]any{"event_type": core.EventDocumentIssued, "occurred_at": w.today,
		"payload": map[string]any{"writer": "jpk_v7m", "params": "{}",
			"file": map[string]any{"hash": hash, "name": "scan.pdf", "media_type": "application/pdf", "size": len(beta)}}}, &out); code != 200 {
		t.Logf("alfa's fact naming beta's file was refused: %d %v", code, out)
	}
	if code, body := read("kk", a, hash); code != 404 || body == string(beta) {
		t.Fatalf("alfa read beta's file through a fact naming it: %d %q", code, body)
	}

	// the same bytes received by both are each company's own
	shared := []byte("%PDF-1.7 the price list everyone gets")
	h := receive("kk", a, shared)
	receive("olga", b, shared)
	for _, c := range []struct{ person, company string }{{"kk", a}, {"olga", b}} {
		if code, body := read(c.person, c.company, h); code != 200 || body != string(shared) {
			t.Fatalf("%s reads the file it received: %d %q", c.person, code, body)
		}
	}
}
