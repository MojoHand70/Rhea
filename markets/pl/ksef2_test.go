package pl

import (
	"context"
	"crypto/ecdsa"
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
	"strings"
	"sync"
	"testing"
	"time"
)

// The national API, faked to the letter of the Ministry's docs: the
// published keys, the challenge, the token encrypted with its timestamp,
// the operation that is "in progress" once before it succeeds, the access
// token redeemed, the session whose AES key arrives wrapped, the invoice
// whose content and hashes must agree, the status that is "processing"
// once before it is final, the UPO, the close. The fake decrypts what the
// client sent and hands the test the plain XML, so the whole cipher path
// is checked end to end without the Ministry.
type fakeKSeFAPI struct {
	t        *testing.T
	tokenKey *rsa.PrivateKey
	symKey   *rsa.PrivateKey
	token    string
	nip      string
	mu       sync.Mutex
	aes, iv  []byte
	polls    map[string]int
	received map[string][]byte // invoice reference -> plain XML
	numbers  map[string]string // invoice reference -> invoice number
	closed   bool
}

func selfSigned(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Ministerstwo Finansów (fake)"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return key, base64.StdEncoding.EncodeToString(der)
}

func (f *fakeKSeFAPI) handler() http.Handler {
	tokenCert, symCert := "", ""
	f.tokenKey, tokenCert = selfSigned(f.t)
	f.symKey, symCert = selfSigned(f.t)
	f.polls, f.received, f.numbers = map[string]int{}, map[string][]byte{}, map[string]string{}
	mux := http.NewServeMux()
	reply := func(w http.ResponseWriter, code int, v any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(v)
	}
	refuse := func(w http.ResponseWriter, code int, words string) {
		reply(w, code, map[string]any{"exception": map[string]any{"exceptionDetailList": []map[string]any{{"exceptionCode": code, "exceptionDescription": words}}}})
	}
	bearer := func(r *http.Request) string { return strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") }
	mux.HandleFunc("GET /security/public-key-certificates", func(w http.ResponseWriter, r *http.Request) {
		reply(w, 200, []map[string]any{
			{"certificate": tokenCert, "publicKeyId": "token-key", "usage": []string{"KsefTokenEncryption"}},
			{"certificate": symCert, "publicKeyId": "sym-key", "usage": []string{"SymmetricKeyEncryption"}},
		})
	})
	mux.HandleFunc("POST /auth/challenge", func(w http.ResponseWriter, r *http.Request) {
		reply(w, 200, map[string]any{"challenge": "20250101-CR-0123456789ABCDEF0123-0123456789-AB", "timestampMs": 1735689600000})
	})
	mux.HandleFunc("POST /auth/ksef-token", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Challenge         string            `json:"challenge"`
			ContextIdentifier map[string]string `json:"contextIdentifier"`
			EncryptedToken    string            `json:"encryptedToken"`
			PublicKeyID       string            `json:"publicKeyId"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		enc, _ := base64.StdEncoding.DecodeString(req.EncryptedToken)
		plain, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, f.tokenKey, enc, nil)
		if err != nil || string(plain) != f.token+"|1735689600000" || req.PublicKeyID != "token-key" ||
			req.ContextIdentifier["type"] != "Nip" || req.ContextIdentifier["value"] != f.nip || !strings.HasPrefix(req.Challenge, "20250101-CR-") {
			refuse(w, 400, "Nieprawidłowe szyfrowanie tokena")
			return
		}
		reply(w, 202, map[string]any{"referenceNumber": "AUTH-1", "authenticationToken": map[string]any{"token": "auth-token", "validUntil": "2099-01-01T00:00:00Z"}})
	})
	mux.HandleFunc("POST /auth/xades-signature", func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		doc := string(raw)
		if r.Header.Get("Content-Type") != "application/xml" || !strings.Contains(doc, "<Challenge>20250101-CR-") ||
			!strings.Contains(doc, "<Nip>"+f.nip+"</Nip>") || !strings.Contains(doc, "<SubjectIdentifierType>certificateSubject</SubjectIdentifierType>") {
			refuse(w, 400, "Nieprawidłowy dokument AuthTokenRequest")
			return
		}
		if err := checkSignedLogin(doc, f.nip); err != nil {
			refuse(w, 400, "Nieprawidłowy podpis: "+err.Error())
			return
		}
		f.mu.Lock()
		f.polls["xades"]++
		f.mu.Unlock()
		reply(w, 202, map[string]any{"referenceNumber": "AUTH-1", "authenticationToken": map[string]any{"token": "auth-token", "validUntil": "2099-01-01T00:00:00Z"}})
	})
	mux.HandleFunc("GET /auth/AUTH-1", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != "auth-token" {
			refuse(w, 401, "unauthorized")
			return
		}
		f.mu.Lock()
		f.polls["auth"]++
		n := f.polls["auth"]
		f.mu.Unlock()
		if n == 1 {
			reply(w, 200, map[string]any{"status": map[string]any{"code": 100, "description": "Uwierzytelnianie w toku"}})
			return
		}
		reply(w, 200, map[string]any{"status": map[string]any{"code": 200, "description": "Uwierzytelnianie zakończone sukcesem"}})
	})
	mux.HandleFunc("POST /auth/token/redeem", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != "auth-token" {
			refuse(w, 401, "unauthorized")
			return
		}
		reply(w, 200, map[string]any{"accessToken": map[string]any{"token": "access-jwt", "validUntil": "2099-01-01T00:00:00Z"}, "refreshToken": map[string]any{"token": "refresh-jwt"}})
	})
	mux.HandleFunc("POST /sessions/online", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != "access-jwt" {
			refuse(w, 401, "unauthorized")
			return
		}
		var req struct {
			FormCode   map[string]string `json:"formCode"`
			Encryption map[string]string `json:"encryption"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		wrapped, _ := base64.StdEncoding.DecodeString(req.Encryption["encryptedSymmetricKey"])
		key, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, f.symKey, wrapped, nil)
		iv, _ := base64.StdEncoding.DecodeString(req.Encryption["initializationVector"])
		if err != nil || len(key) != 32 || len(iv) != 16 || req.Encryption["publicKeyId"] != "sym-key" ||
			req.FormCode["systemCode"] != "FA (3)" || req.FormCode["schemaVersion"] != "1-0E" || req.FormCode["value"] != "FA" {
			refuse(w, 400, "Błąd odszyfrowania dostarczonego klucza")
			return
		}
		f.mu.Lock()
		f.aes, f.iv = key, iv
		f.mu.Unlock()
		reply(w, 201, map[string]any{"referenceNumber": "SESSION-1", "validUntil": "2099-01-01T00:00:00Z"})
	})
	mux.HandleFunc("POST /sessions/online/SESSION-1/invoices", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			InvoiceHash             string `json:"invoiceHash"`
			InvoiceSize             int    `json:"invoiceSize"`
			EncryptedInvoiceHash    string `json:"encryptedInvoiceHash"`
			EncryptedInvoiceSize    int    `json:"encryptedInvoiceSize"`
			EncryptedInvoiceContent string `json:"encryptedInvoiceContent"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		enc, _ := base64.StdEncoding.DecodeString(req.EncryptedInvoiceContent)
		f.mu.Lock()
		defer f.mu.Unlock()
		plain, err := aesCBCDecrypt(f.aes, f.iv, enc)
		if err != nil {
			refuse(w, 400, "Błąd odszyfrowania pliku")
			return
		}
		ph, eh := sha256.Sum256(plain), sha256.Sum256(enc)
		if req.InvoiceHash != base64.StdEncoding.EncodeToString(ph[:]) || req.InvoiceSize != len(plain) ||
			req.EncryptedInvoiceHash != base64.StdEncoding.EncodeToString(eh[:]) || req.EncryptedInvoiceSize != len(enc) {
			refuse(w, 400, "Błąd weryfikacji pliku faktury")
			return
		}
		ref := "INV-" + string(rune('A'+len(f.received)))
		f.received[ref] = plain
		var doc struct {
			Number string `xml:"Fa>P_2"`
		}
		xml.Unmarshal(plain, &doc)
		f.numbers[ref] = doc.Number
		reply(w, 202, map[string]any{"referenceNumber": ref})
	})
	mux.HandleFunc("GET /sessions/SESSION-1/invoices/{ref}", func(w http.ResponseWriter, r *http.Request) {
		ref := r.PathValue("ref")
		f.mu.Lock()
		f.polls[ref]++
		n := f.polls[ref]
		number := f.numbers[ref]
		f.mu.Unlock()
		if n == 1 {
			reply(w, 200, map[string]any{"referenceNumber": ref, "status": map[string]any{"code": 150, "description": "Trwa przetwarzanie"}})
			return
		}
		if strings.HasSuffix(number, "/DUP") {
			reply(w, 200, map[string]any{"referenceNumber": ref, "status": map[string]any{"code": 440, "description": "Duplikat faktury", "details": []string{"originalKsefNumber: 5265877635-20250826-0100001AF629-AF"}}})
			return
		}
		reply(w, 200, map[string]any{"referenceNumber": ref, "ksefNumber": KSeFReference(f.nip, "2026-09-21", number),
			"status": map[string]any{"code": 200, "description": "Sukces"}})
	})
	mux.HandleFunc("GET /sessions/SESSION-1/invoices/{ref}/upo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?><Potwierdzenie><NumerReferencyjnySesji>SESSION-1</NumerReferencyjnySesji><Dokument><NumerFaktury>` + f.numbers[r.PathValue("ref")] + `</NumerFaktury></Dokument></Potwierdzenie>`))
	})
	mux.HandleFunc("POST /sessions/online/SESSION-1/close", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.closed = true
		f.mu.Unlock()
		w.WriteHeader(204)
	})
	return mux
}

func TestKSeF2AgainstTheFakeMinistry(t *testing.T) {
	api := &fakeKSeFAPI{t: t, token: "20250101-TK-SECRET", nip: "8058083055"}
	srv := httptest.NewServer(api.handler())
	defer srv.Close()
	client := &KSeF2{BaseURL: srv.URL, NIP: api.nip, Token: api.token,
		Sleep: func(context.Context, time.Duration) error { return nil },
		Log:   func(format string, args ...any) { t.Logf(format, args...) }}
	invoice := func(number string) []byte {
		body, err := WriteFA3(FA3Invoice{
			Seller:   FA3Party{NIP: api.nip, Name: "Rhea Test sp. z o.o.", Address: "ul. Testowa 1, 00-001 Warszawa"},
			Buyer:    FA3Party{NIP: "5265877635", Name: "Nabywca S.A."},
			Currency: "PLN", IssueDate: "2026-09-21", Number: number, RateCode: "23",
			Net: 100000, VAT: 23000, Gross: 123000, Created: "2026-09-21T08:00:00Z", SystemInfo: "Rhea test",
		})
		if err != nil {
			t.Fatal(err)
		}
		return body
	}
	good, dup := invoice("FV/1/2026"), invoice("FV/2/2026/DUP")
	subs, err := client.Submit(context.Background(), []Outbound{{ID: "sales_invoice-1", Number: "FV/1/2026", XML: good}, {ID: "sales_invoice-2", Number: "FV/2/2026/DUP", XML: dup}})
	if err != nil {
		t.Fatal(err)
	}
	if len(subs) != 2 {
		t.Fatalf("submissions: %+v", subs)
	}
	a, d := subs[0], subs[1]
	if !a.Accepted || a.Code != 200 || !KSeFNumberValid(a.KSeFNumber) || !strings.HasPrefix(a.KSeFNumber, api.nip+"-20260921-") ||
		!strings.Contains(a.UPO, "<NumerFaktury>FV/1/2026</NumerFaktury>") || a.Session != "SESSION-1" || a.Reference != "INV-A" {
		t.Fatalf("the accepted invoice: %+v", a)
	}
	sum := sha256.Sum256(good)
	if a.InvoiceHash != base64.StdEncoding.EncodeToString(sum[:]) {
		t.Fatalf("the hash is of the XML sent: %s", a.InvoiceHash)
	}
	if d.Accepted || d.Code != 440 || d.Description != "Duplikat faktury" || len(d.Details) != 1 || d.UPO != "" ||
		d.Reason() != "KSeF 440: Duplikat faktury (originalKsefNumber: 5265877635-20250826-0100001AF629-AF)" {
		t.Fatalf("the rejected invoice: %+v", d)
	}
	// What the Ministry decrypted is what Rhea wrote, byte for byte.
	if string(api.received["INV-A"]) != string(good) || string(api.received["INV-B"]) != string(dup) {
		t.Fatalf("the Ministry read something else:\n%s", api.received["INV-A"])
	}
	if !api.closed || api.polls["auth"] != 2 || api.polls["INV-A"] != 2 {
		t.Fatalf("session closed %v, auth polled %d, invoice polled %d", api.closed, api.polls["auth"], api.polls["INV-A"])
	}

	// A wrong token is refused in the Ministry's words, and nothing is sent.
	bad := &KSeF2{BaseURL: srv.URL, NIP: api.nip, Token: "wrong", Sleep: client.Sleep}
	if _, err := bad.Submit(context.Background(), []Outbound{{ID: "x", Number: "x", XML: good}}); err == nil || !strings.Contains(err.Error(), "Nieprawidłowe szyfrowanie tokena") {
		t.Fatalf("a wrong token: %v", err)
	}
	// Nothing to send is nothing done, configured or not.
	if subs, err := (&KSeF2{}).Submit(context.Background(), nil); err != nil || subs != nil {
		t.Fatalf("empty: %v %v", subs, err)
	}
	if _, err := (&KSeF2{}).Submit(context.Background(), []Outbound{{}}); err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("unconfigured: %v", err)
	}
}

// checkSignedLogin verifies a signed login the way a verifier would, by
// text surgery rather than by the signer's own canonicalizer: the
// document's digest is the document with its signature cut out, SignedInfo
// canonical is SignedInfo with its namespace declared on itself, and its
// signature verifies with the key of the certificate the document carries,
// a company seal for the context's NIP.
func checkSignedLogin(doc, nip string) error {
	between := func(s, from, to string) (string, int, int) {
		i := strings.Index(s, from)
		j := strings.Index(s, to)
		if i < 0 || j < i {
			return "", -1, -1
		}
		return s[i : j+len(to)], i, j + len(to)
	}
	sig, i, j := between(doc, "<ds:Signature ", "</ds:Signature>")
	if sig == "" {
		return errors.New("no signature")
	}
	inner := func(s, tag string) string {
		v, _, _ := between(s, "<"+tag+">", "</"+tag+">")
		return strings.TrimSuffix(strings.TrimPrefix(v, "<"+tag+">"), "</"+tag+">")
	}
	docSum := sha256.Sum256([]byte(doc[:i] + doc[j:]))
	if !strings.Contains(sig, `<ds:Reference URI="">`) || inner(sig, "ds:DigestValue") != base64.StdEncoding.EncodeToString(docSum[:]) {
		return errors.New("the document's digest")
	}
	der, err := base64.StdEncoding.DecodeString(inner(sig, "ds:X509Certificate"))
	if err != nil {
		return err
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return err
	}
	seal := false
	for _, n := range cert.Subject.Names {
		if n.Type.String() == "2.5.4.97" && n.Value == "VATPL-"+nip {
			seal = true
		}
	}
	if !seal {
		return errors.New("the certificate is not a seal for the context's NIP")
	}
	signedInfo, _, _ := between(sig, "<ds:SignedInfo>", "</ds:SignedInfo>")
	canonical := strings.Replace(signedInfo, "<ds:SignedInfo>", `<ds:SignedInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#">`, 1)
	value, err := base64.StdEncoding.DecodeString(inner(sig, "ds:SignatureValue"))
	if err != nil || len(value) != 64 {
		return errors.New("the signature value")
	}
	pub, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("not an EC key")
	}
	sum := sha256.Sum256([]byte(canonical))
	if !ecdsa.Verify(pub, sum[:], new(big.Int).SetBytes(value[:32]), new(big.Int).SetBytes(value[32:])) {
		return errors.New("SignedInfo does not verify")
	}
	return nil
}

// Certificate login (KSEF-21): the same session as with the token, but the
// way in is the login document signed in XAdES with the company's seal.
// A seal for another NIP is refused, and nothing is sent.
func TestKSeF2CertificateLogin(t *testing.T) {
	api := &fakeKSeFAPI{t: t, token: "unused", nip: "8058083055"}
	srv := httptest.NewServer(api.handler())
	defer srv.Close()
	seal, err := SelfSignedSeal(api.nip, "Rhea Test sp. z o.o.", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	client := &KSeF2{BaseURL: srv.URL, NIP: api.nip, Signer: seal, Sleep: func(context.Context, time.Duration) error { return nil }}
	body, err := WriteFA3(FA3Invoice{
		Seller:   FA3Party{NIP: api.nip, Name: "Rhea Test sp. z o.o.", Address: "ul. Testowa 1, 00-001 Warszawa"},
		Buyer:    FA3Party{NIP: "5265877635", Name: "Nabywca S.A."},
		Currency: "PLN", IssueDate: "2026-10-19", Number: "FV/C/1", RateCode: "23",
		Net: 10000, VAT: 2300, Gross: 12300, Created: "2026-10-19T08:00:00Z", SystemInfo: "Rhea test",
	})
	if err != nil {
		t.Fatal(err)
	}
	subs, err := client.Submit(context.Background(), []Outbound{{ID: "sales_invoice-1", Number: "FV/C/1", XML: body}})
	if err != nil || len(subs) != 1 || !subs[0].Accepted || api.polls["xades"] != 1 {
		t.Fatalf("certificate login: %+v %v (signed logins %d)", subs, err, api.polls["xades"])
	}
	other, _ := SelfSignedSeal("5265877635", "Ktoś Inny S.A.", time.Now())
	stranger := &KSeF2{BaseURL: srv.URL, NIP: api.nip, Signer: other, Sleep: client.Sleep}
	if _, err := stranger.Submit(context.Background(), []Outbound{{ID: "x", Number: "x", XML: body}}); err == nil || !strings.Contains(err.Error(), "seal for the context") {
		t.Fatalf("a seal for another NIP: %v", err)
	}
}

// A certificate and its key travel as PEM and come back the same; a key
// that is not the certificate's, or an encrypted one, is refused.
func TestCertSignerPEM(t *testing.T) {
	seal, err := SelfSignedSeal("8058083055", "Rhea Test sp. z o.o.", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	certPEM, keyPEM, err := seal.PEM()
	if err != nil {
		t.Fatal(err)
	}
	back, err := ParseCertSigner(certPEM, keyPEM)
	if err != nil || !back.Cert.Equal(seal.Cert) {
		t.Fatalf("round trip: %v", err)
	}
	other, _ := SelfSignedSeal("8058083055", "Rhea Test sp. z o.o.", time.Now())
	_, otherKey, _ := other.PEM()
	if _, err := ParseCertSigner(certPEM, otherKey); err == nil || !strings.Contains(err.Error(), "not the certificate's key") {
		t.Fatalf("another key: %v", err)
	}
	if _, err := ParseCertSigner(certPEM, []byte("-----BEGIN ENCRYPTED PRIVATE KEY-----\nAAAA\n-----END ENCRYPTED PRIVATE KEY-----\n")); err == nil || !strings.Contains(err.Error(), "encrypted") {
		t.Fatalf("an encrypted key: %v", err)
	}
}

// The Ministry's own example number passes its CRC-8, and a touched one
// does not; the fake's references are shaped and checksummed the same way.
func TestKSeFNumber(t *testing.T) {
	if !KSeFNumberValid("5265877635-20250826-0100001AF629-AF") {
		t.Fatal("the documented example should validate")
	}
	for _, bad := range []string{"5265877635-20250826-0100001AF629-AE", "5265877635-20250826-0100001AF62-AF", "", "5265877635-20250826-0100001AF629-ZZ"} {
		if KSeFNumberValid(bad) {
			t.Fatalf("%q should not validate", bad)
		}
	}
	ref := KSeFReference("5265877635", "2026-09-21", "FV/1/2026")
	if !KSeFNumberValid(ref) || !strings.HasPrefix(ref, "5265877635-20260921-") || ref != KSeFReference("5265877635", "2026-09-21", "FV/1/2026") {
		t.Fatalf("the fake's reference: %s", ref)
	}
}

// AES-CBC round trip, PKCS#7 at every length.
func TestAESCBC(t *testing.T) {
	key, iv := make([]byte, 32), make([]byte, 16)
	rand.Read(key)
	rand.Read(iv)
	for n := 0; n < 40; n++ {
		plain := []byte(strings.Repeat("x", n))
		enc, err := aesCBCEncrypt(key, iv, plain)
		if err != nil || len(enc)%16 != 0 || len(enc) <= n {
			t.Fatalf("n=%d: %v len %d", n, err, len(enc))
		}
		back, err := aesCBCDecrypt(key, iv, enc)
		if err != nil || string(back) != string(plain) {
			t.Fatalf("n=%d: %v %q", n, err, back)
		}
	}
}

// The channel's configuration: a certificate wins over a token; half a
// certificate, or a misnamed environment, is a client that refuses.
func TestClientFromEnvCertificate(t *testing.T) {
	dir := t.TempDir()
	seal, err := SelfSignedSeal("8058083055", "Rhea Test sp. z o.o.", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	certPEM, keyPEM, _ := seal.PEM()
	certPath, keyPath := dir+"/cert.pem", dir+"/key.pem"
	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RHEA_KSEF_ENV", "test")
	t.Setenv("RHEA_KSEF_NIP", "8058083055")
	t.Setenv("RHEA_KSEF_TOKEN", "a-token")
	t.Setenv("RHEA_KSEF_CERT", certPath)
	t.Setenv("RHEA_KSEF_KEY", keyPath)
	c, ok := ClientFromEnv().(*KSeF2)
	if !ok || c.Signer == nil || c.Token != "" || c.BaseURL != KSeFEnvironments["test"] {
		t.Fatalf("a certificate wins: %+v", c)
	}
	t.Setenv("RHEA_KSEF_KEY", "")
	if _, ok := ClientFromEnv().(refusingClient); !ok {
		t.Fatal("half a certificate is refused")
	}
	t.Setenv("RHEA_KSEF_CERT", "")
	if c, ok := ClientFromEnv().(*KSeF2); !ok || c.Signer != nil || c.Token != "a-token" {
		t.Fatalf("the token, when there is no certificate: %+v", c)
	}
}
