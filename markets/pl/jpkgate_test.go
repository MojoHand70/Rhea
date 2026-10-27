package pl

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

// jpkTestKey is a stand-in for the Ministry's encryption certificate: the
// tests hold its private half, so they can open what the packer sealed.
func jpkTestKey(t *testing.T) (*rsa.PrivateKey, *x509.Certificate) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(7), Subject: pkix.Name{CommonName: "jpk@mf.gov.pl (fake)"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return key, cert
}

func goldenJPK(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/jpk_v7m_2026-09.xml")
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// openJPKPackage does what the Ministry does with a package: unwraps the key
// with its private half, decrypts the part, unzips it.
func openJPKPackage(t *testing.T, key *rsa.PrivateKey, req jpkInitUpload, part []byte) (name string, doc []byte) {
	t.Helper()
	wrapped, err := base64.StdEncoding.DecodeString(req.EncryptionKey.Value)
	if err != nil {
		t.Fatal(err)
	}
	aesKey, err := rsa.DecryptPKCS1v15(nil, key, wrapped)
	if err != nil {
		t.Fatalf("the Ministry cannot unwrap the key (RSA PKCS#1 v1.5): %v", err)
	}
	if len(aesKey) != 32 {
		t.Fatalf("AES key of %d bytes, want 32", len(aesKey))
	}
	iv, err := base64.StdEncoding.DecodeString(req.Document.FileSignatureList.Encryption.AES.IV.Value)
	if err != nil || len(iv) != 16 {
		t.Fatalf("IV %q: %v", req.Document.FileSignatureList.Encryption.AES.IV.Value, err)
	}
	zipped, err := aesCBCDecrypt(aesKey, iv, part)
	if err != nil {
		t.Fatalf("the part does not decrypt: %v", err)
	}
	zr, err := zip.NewReader(bytes.NewReader(zipped), int64(len(zipped)))
	if err != nil {
		t.Fatalf("the decrypted part is not a ZIP: %v", err)
	}
	if len(zr.File) != 1 {
		t.Fatalf("the ZIP holds %d files, want exactly one document", len(zr.File))
	}
	f := zr.File[0]
	if f.Method != zip.Deflate {
		t.Fatalf("the ZIP entry is stored with method %d, the Ministry wants DEFLATE", f.Method)
	}
	rc, err := f.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	doc, err = io.ReadAll(rc)
	if err != nil {
		t.Fatal(err)
	}
	return f.Name, doc
}

func b64sha256(b []byte) string {
	s := sha256.Sum256(b)
	return base64.StdEncoding.EncodeToString(s[:])
}
func b64md5(b []byte) string { s := md5.Sum(b); return base64.StdEncoding.EncodeToString(s[:]) }

// A JPK packed for the gateway opens with the Ministry's key into exactly
// the file that was packed, and the request says so to the byte: the
// document's SHA-256 and length, the part's MD5 and length, the form code
// from the file's own header — the specification's §1.2 and §2.2.1, and the
// sample InitUpload the Ministry publishes for JPK_V7M(3).
func TestPackJPKOpensWithTheMinistrysKey(t *testing.T) {
	key, cert := jpkTestKey(t)
	doc := goldenJPK(t)
	const name = "JPK_V7M_8058083055_202609.xml"
	p, err := PackJPK(name, doc, cert)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(p.Request, []byte(`<?xml version="1.0" encoding="utf-8"?>`)) {
		t.Errorf("the request must declare utf-8 as the gateway reads it (its code 101): %.60s", p.Request)
	}
	var req jpkInitUpload
	if err := xml.Unmarshal(p.Request, &req); err != nil {
		t.Fatalf("the request is not XML: %v\n%s", err, p.Request)
	}
	if req.XMLName.Space != JPKGatewayNamespace || req.XMLName.Local != "InitUpload" {
		t.Errorf("root %v, want InitUpload in %s", req.XMLName, JPKGatewayNamespace)
	}
	if req.DocumentType != "JPK" || req.Version != "01.02.01.20160617" {
		t.Errorf("DocumentType %q, Version %q", req.DocumentType, req.Version)
	}
	ek := req.EncryptionKey
	if ek.Algorithm != "RSA" || ek.Mode != "ECB" || ek.Padding != "PKCS#1" || ek.Encoding != "Base64" || len(ek.Value) != 344 {
		t.Errorf("EncryptionKey %+v (len %d): the schema fixes RSA/ECB/PKCS#1/Base64 and 344 characters", ek, len(ek.Value))
	}
	d := req.Document
	if d.FormCode.SystemCode != "JPK_V7M (3)" || d.FormCode.SchemaVersion != "1-0E" || d.FormCode.Value != "JPK_VAT" {
		t.Errorf("FormCode %+v, want the file's own KodFormularza", d.FormCode)
	}
	if d.FileName != name || d.ContentLength != int64(len(doc)) {
		t.Errorf("document %q of %d bytes, want %q of %d", d.FileName, d.ContentLength, name, len(doc))
	}
	if d.HashValue.Algorithm != "SHA-256" || d.HashValue.Value != b64sha256(doc) {
		t.Errorf("document hash %+v, want SHA-256 %s", d.HashValue, b64sha256(doc))
	}
	fl := d.FileSignatureList
	if fl.FilesNumber != 1 || len(fl.FileSignature) != 1 || fl.Packaging.SplitZip.Type != "split" || fl.Packaging.SplitZip.Mode != "zip" {
		t.Errorf("FileSignatureList %+v: one part, split zip", fl)
	}
	aes := fl.Encryption.AES
	if aes.Size != 256 || aes.Block != 16 || aes.Mode != "CBC" || aes.Padding != "PKCS#7" || aes.IV.Bytes != 16 || aes.IV.Encoding != "Base64" {
		t.Errorf("AES %+v", aes)
	}
	fs := fl.FileSignature[0]
	pattern := regexp.MustCompile(`^[a-zA-Z0-9_\.\-]{5,55}$`)
	if fs.OrdinalNumber != 1 || fs.FileName != p.PartName || !pattern.MatchString(fs.FileName) || !pattern.MatchString(d.FileName) {
		t.Errorf("part %+v, package part name %q: names the schema admits", fs, p.PartName)
	}
	if fs.ContentLength != len(p.Part) || fs.HashValue.Algorithm != "MD5" || fs.HashValue.Value != b64md5(p.Part) {
		t.Errorf("part %d bytes, hash %+v; the part is %d bytes, MD5 %s", fs.ContentLength, fs.HashValue, len(p.Part), b64md5(p.Part))
	}
	gotName, gotDoc := openJPKPackage(t, key, req, p.Part)
	if gotName != name || !bytes.Equal(gotDoc, doc) {
		t.Errorf("the package opens into %q (%d bytes), want %q exactly as packed (%d bytes)", gotName, len(gotDoc), name, len(doc))
	}

	// Every package has its own key: the same file packed twice seals
	// differently, and each opens.
	p2, err := PackJPK(name, doc, cert)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(p2.Part, p.Part) || bytes.Equal(p2.Request, p.Request) {
		t.Error("two packages of one file share a key: the key must be fresh each time")
	}
	var req2 jpkInitUpload
	if err := xml.Unmarshal(p2.Request, &req2); err != nil {
		t.Fatal(err)
	}
	if _, doc2 := openJPKPackage(t, key, req2, p2.Part); !bytes.Equal(doc2, doc) {
		t.Error("the second package does not open into the file")
	}
}

// What cannot go out is refused before anything is sealed: a file without
// the form code the request must repeat, a name the schema does not admit,
// a key that is not the Ministry's kind.
func TestPackJPKRefuses(t *testing.T) {
	_, cert := jpkTestKey(t)
	doc := goldenJPK(t)
	if _, err := PackJPK("JPK_V7M_202609.xml", []byte("<JPK><Naglowek/></JPK>"), cert); err == nil || !strings.Contains(err.Error(), "KodFormularza") {
		t.Errorf("a file without KodFormularza: %v", err)
	}
	if _, err := PackJPK("jpk v7m.xml", doc, cert); err == nil || !strings.Contains(err.Error(), "name") {
		t.Errorf("a name with a space: %v", err)
	}
	ec, err := SelfSignedSeal("5260250274", "Nie RSA", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := PackJPK("JPK_V7M_202609.xml", doc, ec.Cert); err == nil || !strings.Contains(err.Error(), "RSA") {
		t.Errorf("an EC certificate: %v", err)
	}
}

// The Ministry's two encryption certificates ship with the pack: the test
// gateway's and production's, both RSA 2048, both the Ministry's — a
// package sealed for one environment opens only in that one.
func TestJPKCertificates(t *testing.T) {
	seen := map[string]bool{}
	for _, env := range []string{"test", "prod"} {
		c, err := JPKCertificate(env)
		if err != nil {
			t.Fatalf("%s: %v", env, err)
		}
		pub, ok := c.PublicKey.(*rsa.PublicKey)
		if !ok || pub.N.BitLen() != 2048 {
			t.Errorf("%s: want an RSA 2048 key, got %T", env, c.PublicKey)
		}
		if c.Subject.CommonName != "jpk@mf.gov.pl" {
			t.Errorf("%s: subject %v, want the Ministry's jpk@mf.gov.pl", env, c.Subject)
		}
		if time.Now().After(c.NotAfter) {
			t.Errorf("%s: the certificate expired %v — the Ministry publishes the next one on podatki.gov.pl", env, c.NotAfter)
		}
		seen[string(c.Raw)] = true
	}
	if len(seen) != 2 {
		t.Error("test and production share a certificate")
	}
	if _, err := JPKCertificate("demo"); err == nil {
		t.Error("an environment the gateway does not have")
	}
}

// signedJPKRequest is the person's step done by a test: the request signed
// enveloped with a self-made person's certificate, which the test gateway
// admits.
func signedJPKRequest(t *testing.T, request []byte) []byte {
	t.Helper()
	signer, err := SelfSignedSignerRSA("8058083055", "Jan", "Testowy", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	signed, err := SignJPKRequest(request, signer, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

// The test signer is a person with a NIP, as the gateway wants a signer: on
// the test gateway a seal's VATPL-<NIP> came back 423 ("certificate without
// the required attributes") and a person's TINPL-<NIP> filed with its UPO
// (2026-10-27). RSA, as the gateway signs.
func TestSelfSignedSignerCarriesTheNIP(t *testing.T) {
	s, err := SelfSignedSignerRSA("8058083055", "Jan", "Testowy", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if s.Cert.Subject.SerialNumber != "TINPL-8058083055" {
		t.Errorf("serialNumber %q, want TINPL-8058083055", s.Cert.Subject.SerialNumber)
	}
	for _, a := range s.Cert.Subject.Names {
		if a.Type.Equal(oidOrganizationIdentifier) {
			t.Errorf("a person's certificate carries a seal's organizationIdentifier %v", a.Value)
		}
	}
	if _, ok := s.Key.Public().(*rsa.PublicKey); !ok {
		t.Errorf("key %T, the gateway signs RSA-SHA256", s.Key.Public())
	}
}

// The signed file the person hands back is checked against the request Rhea
// prepared before anything is sent: it must carry a signature, and what it
// signs must be this package — its key, its document, its part. Enveloped
// and enveloping signatures both, as the specification admits (§1.3.1), the
// enveloped object inline or base64.
func TestCheckSignedJPKRequest(t *testing.T) {
	_, cert := jpkTestKey(t)
	doc := goldenJPK(t)
	p, err := PackJPK("JPK_V7M_8058083055_202609.xml", doc, cert)
	if err != nil {
		t.Fatal(err)
	}
	signed := signedJPKRequest(t, p.Request)
	if err := CheckSignedJPKRequest(p.Request, signed); err != nil {
		t.Errorf("the enveloped signature of this request: %v", err)
	}
	if err := CheckSignedJPKRequest(p.Request, p.Request); err == nil || !strings.Contains(err.Error(), "signature") {
		t.Errorf("the request handed back unsigned: %v", err)
	}
	other, err := PackJPK("JPK_V7M_8058083055_202609.xml", doc, cert)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckSignedJPKRequest(p.Request, signedJPKRequest(t, other.Request)); err == nil || !strings.Contains(err.Error(), "another") {
		t.Errorf("a signed request of another package: %v", err)
	}
	if err := CheckSignedJPKRequest(p.Request, []byte("%PDF-1.7")); err == nil {
		t.Error("a file that is no XML")
	}
	body := bytes.TrimPrefix(p.Request, []byte(`<?xml version="1.0" encoding="utf-8"?>`+"\n"))
	enveloping := []byte(`<ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#"><ds:SignedInfo/><ds:SignatureValue>AA==</ds:SignatureValue><ds:Object>` +
		base64.StdEncoding.EncodeToString(body) + `</ds:Object></ds:Signature>`)
	if err := CheckSignedJPKRequest(p.Request, enveloping); err != nil {
		t.Errorf("an enveloping signature with the request in base64: %v", err)
	}
	inline := []byte(`<ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#"><ds:SignedInfo/><ds:SignatureValue>AA==</ds:SignatureValue><ds:Object>` +
		string(body) + `</ds:Object></ds:Signature>`)
	if err := CheckSignedJPKRequest(p.Request, inline); err != nil {
		t.Errorf("an enveloping signature with the request inline: %v", err)
	}
}

// fakeJPKGateway is the Ministry's gateway faked to the specification: the
// signed request in, a session with one upload URL out; the part PUT to the
// storage with the MD5 the request declared; the session finished with the
// blob's name; the status "in verification" once, then 200 with the UPO.
type fakeJPKGateway struct {
	t       *testing.T
	mu      sync.Mutex
	signed  []byte // the request the gateway received
	part    []byte // what reached the storage
	md5     string // the Content-MD5 the PUT carried
	blob    string // the blob named at FinishUpload
	polls   int
	refuse  string // a 400 body InitUploadSigned answers with, when set
	server  *httptest.Server
	foreign string // an upload URL to hand out instead of the storage's
}

const fakeJPKReference = "d4fd41850323d2f6000000b013016327"

func (f *fakeJPKGateway) start() *JPKGateway {
	mux := http.NewServeMux()
	reply := func(w http.ResponseWriter, code int, v any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(v)
	}
	mux.HandleFunc("POST /api/Storage/InitUploadSigned", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/xml") {
			f.t.Errorf("InitUploadSigned with Content-Type %q, want application/xml", ct)
		}
		body, _ := io.ReadAll(r.Body)
		f.mu.Lock()
		f.signed = body
		f.mu.Unlock()
		if f.refuse != "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			io.WriteString(w, f.refuse)
			return
		}
		url := f.server.URL + "/storage/" + fakeJPKReference + "/blob-1?sv=2015-07-08&sig=x"
		if f.foreign != "" {
			url = f.foreign
		}
		reply(w, 200, map[string]any{"ReferenceNumber": fakeJPKReference, "TimeoutInSec": 900,
			"RequestToUploadFileList": []map[string]any{{"BlobName": "blob-1", "FileName": "x.zip.aes", "Url": url, "Method": "PUT",
				"HeaderList": []map[string]string{{"Key": "Content-MD5", "Value": "declared"}, {"Key": "x-ms-blob-type", "Value": "BlockBlob"}}}}})
	})
	mux.HandleFunc("PUT /storage/{ref}/{blob}", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-ms-blob-type") != "BlockBlob" {
			f.t.Errorf("Put Blob without x-ms-blob-type: %v", r.Header)
		}
		body, _ := io.ReadAll(r.Body)
		f.mu.Lock()
		f.part, f.md5 = body, r.Header.Get("Content-MD5")
		f.mu.Unlock()
		w.WriteHeader(201)
	})
	mux.HandleFunc("POST /api/Storage/FinishUpload", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ReferenceNumber   string
			AzureBlobNameList []string
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ReferenceNumber != fakeJPKReference || len(req.AzureBlobNameList) != 1 {
			f.t.Errorf("FinishUpload %+v: %v", req, err)
		}
		f.mu.Lock()
		f.blob = req.AzureBlobNameList[0]
		f.mu.Unlock()
		w.WriteHeader(200)
	})
	mux.HandleFunc("GET /api/Storage/Status/{ref}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("ref") != fakeJPKReference {
			reply(w, 200, map[string]any{"Code": 300, "Description": "Nieprawidłowy numer referencyjny"})
			return
		}
		f.mu.Lock()
		f.polls++
		n := f.polls
		f.mu.Unlock()
		if n == 1 {
			reply(w, 200, map[string]any{"Code": 120, "Description": "Sesja została poprawnie zakończona. Dane zostały poprawnie zapisane. Trwa weryfikacja dokumentu"})
			return
		}
		reply(w, 200, map[string]any{"Code": 200, "Description": "Przetwarzanie dokumentu zakończone poprawnie, pobierz UPO",
			"Details": fakeJPKReference, "Timestamp": "2026-10-25T14:02:11.12+00:00", "Upo": `<?xml version="1.0" encoding="UTF-8"?><Potwierdzenie><NumerReferencyjny>` + fakeJPKReference + `</NumerReferencyjny></Potwierdzenie>`})
	})
	f.server = httptest.NewServer(mux)
	f.t.Cleanup(f.server.Close)
	return &JPKGateway{Env: "test", BaseURL: f.server.URL, Storage: func(string) bool { return true }}
}

// One filing through the fake gateway: the signed request goes as signed,
// byte for byte; the part reaches the storage with its MD5; the session is
// finished with the blob the gateway named; the status is pending, then
// final with the UPO.
func TestJPKGatewayAgainstTheFakeMinistry(t *testing.T) {
	_, cert := jpkTestKey(t)
	p, err := PackJPK("JPK_V7M_8058083055_202609.xml", goldenJPK(t), cert)
	if err != nil {
		t.Fatal(err)
	}
	signed := signedJPKRequest(t, p.Request)
	f := &fakeJPKGateway{t: t}
	g := f.start()
	ctx := context.Background()
	ref, err := g.Send(ctx, signed, p.Part)
	if err != nil {
		t.Fatal(err)
	}
	if ref != fakeJPKReference {
		t.Errorf("reference %q", ref)
	}
	if !bytes.Equal(f.signed, signed) {
		t.Error("the gateway did not receive the signed request byte for byte — any change breaks the signature (its code 130)")
	}
	if !bytes.Equal(f.part, p.Part) || f.md5 != "declared" || f.blob != "blob-1" {
		t.Errorf("storage got %d bytes with Content-MD5 %q, finished %q; want the part (%d bytes) with the header the gateway gave, blob-1", len(f.part), f.md5, f.blob, len(p.Part))
	}
	st, err := g.Status(ctx, ref)
	if err != nil {
		t.Fatal(err)
	}
	if st.Final() || st.Code != 120 {
		t.Errorf("first status %+v: in verification, not final", st)
	}
	st, err = g.Status(ctx, ref)
	if err != nil {
		t.Fatal(err)
	}
	if !st.Final() || !st.Accepted() || !strings.Contains(st.UPO, fakeJPKReference) || st.Timestamp == "" {
		t.Errorf("second status %+v: accepted with the UPO", st)
	}
	st, err = g.Status(ctx, "ffffffffffffffffffffffffffffffff")
	if err != nil {
		t.Fatal(err)
	}
	if !st.Final() || st.Accepted() {
		t.Errorf("an unknown reference %+v is the gateway's final word, not acceptance", st)
	}
}

// The gateway's refusals are its word, kept with their code: a signature it
// does not accept, a duplicate of a document it already processed (with the
// original's reference, so the filing follows that one), and a server error
// that is no refusal at all — the next pass tries again. An upload URL off
// the Ministry's storage gets no bytes.
func TestJPKGatewayRefusals(t *testing.T) {
	_, cert := jpkTestKey(t)
	p, err := PackJPK("JPK_V7M_8058083055_202609.xml", goldenJPK(t), cert)
	if err != nil {
		t.Fatal(err)
	}
	signed := signedJPKRequest(t, p.Request)
	ctx := context.Background()

	f := &fakeJPKGateway{t: t, refuse: `{"Message":"Podpis negatywnie zweryfikowany","Code":120,"RequestId":"172dc3cc-5b97-48de-91dd-6903587cba19"}`}
	_, err = f.start().Send(ctx, signed, p.Part)
	var refusal *JPKRefusal
	if !errors.As(err, &refusal) || refusal.Code != 120 || refusal.Message != "Podpis negatywnie zweryfikowany" {
		t.Fatalf("a refused signature: %v", err)
	}
	if f.part != nil {
		t.Error("a refused session sent the part anyway")
	}

	f = &fakeJPKGateway{t: t, refuse: `{"Message":"Przesłano duplikat przetworzonego dokumentu. Numer referencyjny oryginału: ef7d17780087346e0000004c0c7982ec","Code":170,"RequestId":"x"}`}
	_, err = f.start().Send(ctx, signed, p.Part)
	if !errors.As(err, &refusal) || refusal.Code != 170 || refusal.Original != "ef7d17780087346e0000004c0c7982ec" {
		t.Fatalf("a duplicate: %v (%+v)", err, refusal)
	}

	broken := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		io.WriteString(w, `{"Message":"Wewnętrzny błąd systemu ","RequestId":"x"}`)
	}))
	defer broken.Close()
	_, err = (&JPKGateway{Env: "test", BaseURL: broken.URL}).Send(ctx, signed, p.Part)
	if err == nil || errors.As(err, &refusal) {
		t.Fatalf("a server error is no refusal, the next pass tries again: %v", err)
	}

	f = &fakeJPKGateway{t: t, foreign: "https://example.com/upload"}
	g := f.start()
	g.Storage = nil // the Ministry's storage hosts only
	if _, err := g.Send(ctx, signed, p.Part); err == nil || !strings.Contains(err.Error(), "storage") {
		t.Errorf("an upload URL off the Ministry's storage: %v", err)
	}
	if f.part != nil {
		t.Error("bytes went to a host that is not the Ministry's storage")
	}
	for host, ok := range map[string]bool{
		"taxdocumentstorage09tst.blob.core.windows.net":   true,
		"taxdocumentstorage97.blob.core.windows.net":      true,
		"taxdocumentstorage9.blob.core.windows.net":       false,
		"taxdocumentstorage09.blob.core.windows.net.evil": false,
		"evil.taxdocumentstorage09.blob.core.windows.net": false,
	} {
		if got := jpkStorageHost(host); got != ok {
			t.Errorf("storage host %q: %v, want %v", host, got, ok)
		}
	}
}

// The environment is the operator's choice and nothing is guessed: unset or
// misspelled, the gateway refuses to send (a fake UPO in real books would
// say a return was filed that was not); "fake" must be asked for by name.
func TestJPKGatewayFromEnv(t *testing.T) {
	ctx := context.Background()
	for env, want := range map[string]string{"": "not configured", "production": "not one of"} {
		t.Setenv("RHEA_JPK_ENV", env)
		g := JPKGatewayFromEnv()
		if _, err := g.Send(ctx, nil, nil); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("RHEA_JPK_ENV=%q: send %v, want %q", env, err, want)
		}
		if _, err := g.Certificate(); err == nil {
			t.Errorf("RHEA_JPK_ENV=%q: a certificate to seal with", env)
		}
	}
	for env, base := range map[string]string{"test": "https://test-e-dokumenty.mf.gov.pl", "prod": "https://e-dokumenty.mf.gov.pl"} {
		t.Setenv("RHEA_JPK_ENV", env)
		g, ok := JPKGatewayFromEnv().(*JPKGateway)
		if !ok || g.BaseURL != base || g.Environment() != env {
			t.Errorf("RHEA_JPK_ENV=%s: %+v", env, g)
			continue
		}
		c, err := g.Certificate()
		want, _ := JPKCertificate(env)
		if err != nil || !c.Equal(want) {
			t.Errorf("RHEA_JPK_ENV=%s seals with the wrong certificate: %v", env, err)
		}
	}
	t.Setenv("RHEA_JPK_ENV", "fake")
	if g := JPKGatewayFromEnv(); g.Environment() != "fake" {
		t.Errorf("RHEA_JPK_ENV=fake: %T", g)
	}
}
