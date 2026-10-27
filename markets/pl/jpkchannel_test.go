package pl

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/store"
)

// jpkReads is the read side a JPK pass sees: objects by type, files by hash.
type jpkReads struct {
	objects map[string][]core.Object
	blobs   map[string][]byte
}

func (r jpkReads) ObjectsByType(_ context.Context, typ string) ([]core.Object, error) {
	return r.objects[typ], nil
}

func (r jpkReads) GetBlob(_ context.Context, hash string) (store.Blob, error) {
	b, ok := r.blobs[hash]
	if !ok {
		return store.Blob{}, fmt.Errorf("no file %s in the store", hash)
	}
	return store.Blob{Hash: hash, Bytes: b}, nil
}

func (r jpkReads) keep(b []byte) string {
	h := core.HashBytes(b)
	r.blobs[h] = b
	return h
}

// scriptedGateway answers as the test says.
type scriptedGateway struct {
	env    string
	cert   *x509.Certificate
	send   func(signed, part []byte) (string, error)
	status func(ref string) (JPKStatus, error)
	sent   int
}

func (g *scriptedGateway) Environment() string { return g.env }
func (g *scriptedGateway) Certificate() (*x509.Certificate, error) {
	if g.cert == nil {
		return nil, errors.New("jpk gateway not configured")
	}
	return g.cert, nil
}
func (g *scriptedGateway) Send(_ context.Context, signed, part []byte) (string, error) {
	g.sent++
	return g.send(signed, part)
}
func (g *scriptedGateway) Status(_ context.Context, ref string) (JPKStatus, error) {
	return g.status(ref)
}

func payloadOf(t *testing.T, ev core.Event) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(ev.Payload, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

// A handed-over JPK is sealed once for the gateway the server sends to,
// keyed by the file's hash; a file of another writer is not a JPK; with no
// gateway named nothing is sealed and nothing refused — the JPK waits.
func TestJPKChannelPrepares(t *testing.T) {
	_, cert := jpkTestKey(t)
	r := jpkReads{objects: map[string][]core.Object{}, blobs: map[string][]byte{}}
	doc := goldenJPK(t)
	hash := r.keep(doc)
	r.objects["issued_document"] = []core.Object{
		{ID: "d1", Type: "issued_document", State: map[string]any{"writer": "jpk_v7m", "name": "JPK_V7M_8058083055_202609.xml", "hash": hash, "params": "created=2026-10-05T10:00:00Z,period=2026-09"}},
		{ID: "d2", Type: "issued_document", State: map[string]any{"writer": "statement", "name": "x.pdf", "hash": "sha256:00"}},
	}
	ctx := context.Background()
	today := func() string { return "2026-10-05" }

	evs, err := JPKChannel{Gateway: &scriptedGateway{env: "test"}, Today: today}.Pass(ctx, r)
	if len(evs) != 0 || err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("no gateway named: %d facts, %v", len(evs), err)
	}

	evs, err = JPKChannel{Gateway: &scriptedGateway{env: "test", cert: cert}, Today: today}.Pass(ctx, r)
	if err != nil || len(evs) != 1 {
		t.Fatalf("prepared: %d facts, %v", len(evs), err)
	}
	ev := evs[0]
	p := payloadOf(t, ev)
	if ev.Type != EventJPKPrepared || ev.DedupKey != "jpk_gateway/prepared/"+hash || p["document"] != hash || p["period"] != "2026-09" ||
		p["gateway"] != "test" || p["date"] != "2026-10-05" || len(ev.Files) != 2 || ev.Files[0].Input != "package" || ev.Files[1].Input != "request" {
		t.Fatalf("the prepared fact: %s %s %v files %d", ev.Type, ev.DedupKey, p, len(ev.Files))
	}
	if err := CheckSignedJPKRequest(ev.Files[1].Bytes, signedJPKRequest(t, ev.Files[1].Bytes)); err != nil {
		t.Errorf("the request handed out is the one a signature makes sendable: %v", err)
	}

	r.objects["jpk_filing"] = []core.Object{{ID: "f1", Type: "jpk_filing", State: map[string]any{"document": hash, "status": "to_sign"}}}
	if evs, err := (JPKChannel{Gateway: &scriptedGateway{env: "test", cert: cert}, Today: today}).Pass(ctx, r); err != nil || len(evs) != 0 {
		t.Fatalf("a file already filed is sealed again: %d facts, %v", len(evs), err)
	}
}

// A signed filing is checked before it is sent, and every answer becomes
// its fact: sent with the reference, a duplicate following the original,
// the gateway's refusal and Rhea's own (an unsigned file, a file prepared
// for the other gateway) as a rejection with the reason in Polish, keyed by
// the signed file — a new signature is a new try. A server error is no
// refusal: nothing is booked, the next pass sends again.
func TestJPKChannelSends(t *testing.T) {
	_, cert := jpkTestKey(t)
	p, err := PackJPK("JPK_V7M_8058083055_202609.xml", goldenJPK(t), cert)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	filingWith := func(signed []byte, gateway string) (jpkReads, string) {
		r := jpkReads{objects: map[string][]core.Object{}, blobs: map[string][]byte{}}
		signedHash := r.keep(signed)
		r.objects["jpk_filing"] = []core.Object{{ID: "f1", Type: "jpk_filing", State: map[string]any{
			"status": "signed", "gateway": gateway, "request": r.keep(p.Request), "package": r.keep(p.Part), "signed": signedHash}}}
		return r, signedHash
	}
	signed := signedJPKRequest(t, p.Request)
	pass := func(r jpkReads, g *scriptedGateway) ([]core.Event, error) {
		return JPKChannel{Gateway: g, Today: func() string { return "2026-10-06" }}.Pass(ctx, r)
	}

	r, signedHash := filingWith(signed, "test")
	g := &scriptedGateway{env: "test", cert: cert, send: func(got, part []byte) (string, error) {
		if string(got) != string(signed) || string(part) != string(p.Part) {
			t.Error("the gateway did not get the signed file and the part as kept")
		}
		return fakeJPKReference, nil
	}}
	evs, err := pass(r, g)
	if err != nil || len(evs) != 1 || evs[0].Type != EventJPKSent || evs[0].DedupKey != "jpk_gateway/sent/f1/"+signedHash || payloadOf(t, evs[0])["reference"] != fakeJPKReference {
		t.Fatalf("sent: %v %v", evs, err)
	}

	g.send = func([]byte, []byte) (string, error) {
		return "", &JPKRefusal{Code: 170, Message: "duplikat", Original: "ef7d17780087346e0000004c0c7982ec"}
	}
	if evs, err := pass(r, g); err != nil || len(evs) != 1 || evs[0].Type != EventJPKSent || payloadOf(t, evs[0])["reference"] != "ef7d17780087346e0000004c0c7982ec" {
		t.Fatalf("a duplicate follows the original: %v %v", evs, err)
	}

	g.send = func([]byte, []byte) (string, error) {
		return "", &JPKRefusal{Code: 120, Message: "Podpis negatywnie zweryfikowany"}
	}
	evs, err = pass(r, g)
	if err != nil || len(evs) != 1 || evs[0].Type != EventJPKRejected || evs[0].DedupKey != "jpk_gateway/rejected/f1/"+signedHash {
		t.Fatalf("a refusal: %v %v", evs, err)
	}
	if m := payloadOf(t, evs[0]); m["code"] != float64(120) || !strings.Contains(m["reason"].(string), "Podpis negatywnie zweryfikowany") {
		t.Fatalf("the refusal's code and words: %v", m)
	}

	g.send = func([]byte, []byte) (string, error) {
		return "", errors.New("jpk gateway: InitUploadSigned answered HTTP 503")
	}
	if evs, err := pass(r, g); err == nil || len(evs) != 0 {
		t.Fatalf("a server error: %d facts, %v", len(evs), err)
	}

	sent := g.sent
	r, _ = filingWith(p.Request, "test") // handed back unsigned
	evs, err = pass(r, g)
	if err != nil || len(evs) != 1 || evs[0].Type != EventJPKRejected || !strings.Contains(payloadOf(t, evs[0])["reason"].(string), "nie ma podpisu") || g.sent != sent {
		t.Fatalf("an unsigned file: %v %v, sent %d times more", evs, err, g.sent-sent)
	}
	r, _ = filingWith(signed, "prod")
	evs, err = pass(r, g)
	if err != nil || len(evs) != 1 || evs[0].Type != EventJPKRejected || !strings.Contains(payloadOf(t, evs[0])["reason"].(string), "bramki prod") || g.sent != sent {
		t.Fatalf("a filing prepared for the other gateway: %v %v", evs, err)
	}
	r, _ = filingWith(signed, "test")
	if evs, err := pass(r, &scriptedGateway{env: "test"}); err == nil || len(evs) != 0 {
		t.Fatalf("no gateway named: a signed filing waits, it is not refused: %v %v", evs, err)
	}
}

// A filing sent is asked for its status until the gateway decides: filed
// with the UPO kept as a file, or refused with the gateway's code and words.
func TestJPKChannelHearsTheStatus(t *testing.T) {
	ctx := context.Background()
	r := jpkReads{objects: map[string][]core.Object{"jpk_filing": {{ID: "f1", Type: "jpk_filing",
		State: map[string]any{"status": "sent", "gateway": "test", "reference": fakeJPKReference}}}}, blobs: map[string][]byte{}}
	answer := JPKStatus{Code: 120, Description: "Trwa weryfikacja dokumentu"}
	g := &scriptedGateway{env: "test", status: func(ref string) (JPKStatus, error) {
		if ref != fakeJPKReference {
			t.Errorf("status asked for %q", ref)
		}
		return answer, nil
	}}
	pass := func() ([]core.Event, error) {
		return JPKChannel{Gateway: g, Today: func() string { return "2026-10-06" }}.Pass(ctx, r)
	}
	if evs, err := pass(); err != nil || len(evs) != 0 {
		t.Fatalf("in verification: %v %v", evs, err)
	}
	answer = JPKStatus{Code: 200, Description: "Przetwarzanie dokumentu zakończone poprawnie", Timestamp: "2026-10-06T10:00:00Z",
		UPO: "<Potwierdzenie><NumerReferencyjny>" + fakeJPKReference + "</NumerReferencyjny></Potwierdzenie>"}
	evs, err := pass()
	if err != nil || len(evs) != 1 || evs[0].Type != EventJPKAccepted || evs[0].DedupKey != "jpk_gateway/accepted/f1/"+fakeJPKReference ||
		len(evs[0].Files) != 1 || evs[0].Files[0].Input != "upo" || !strings.Contains(string(evs[0].Files[0].Bytes), fakeJPKReference) {
		t.Fatalf("filed: %v %v", evs, err)
	}
	answer = JPKStatus{Code: 420, Description: "Brak aktualnego pełnomocnictwa/upoważnienia do podpisywania dokumentu"}
	evs, err = pass()
	if err != nil || len(evs) != 1 || evs[0].Type != EventJPKRejected || evs[0].DedupKey != "jpk_gateway/rejected/f1/"+fakeJPKReference {
		t.Fatalf("refused: %v %v", evs, err)
	}
	if m := payloadOf(t, evs[0]); m["code"] != float64(420) || !strings.Contains(m["reason"].(string), "pełnomocnictwa") {
		t.Fatalf("the refusal's code and words: %v", m)
	}
}
