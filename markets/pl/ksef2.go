package pl

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// KSeF2 speaks to the national e-invoice API 2.0 (DIRECTION 2026-10-15,
// "Documents out"): a challenge answered either with a certificate — the
// login document signed in XAdES (xades.go), the only way in from 2027 — or
// with the token the owner pasted once, encrypted with the Ministry's
// published key; a short-lived access token; then one interactive session
// per pass — FA(3) declared, every
// invoice encrypted with a fresh AES key wrapped with the Ministry's other
// key, sent, polled to its final status, its UPO fetched, the session
// closed. Standard library only: RSA-OAEP and AES-CBC are in crypto, the
// rest is JSON over HTTPS. The environment picks the base URL
// (KSeFEnvironments); keys and tokens of one environment do not work in
// another, which is the safety.
type KSeF2 struct {
	BaseURL string       // https://api-test.ksef.mf.gov.pl/v2 and its siblings
	NIP     string       // the context: the company whose invoices these are
	Token   string       // the KSeF token the owner generated — a secret, never logged
	Signer  *CertSigner  // a certificate and its key: when set, the login is signed with it, not the token
	HTTP    *http.Client // nil: a 60 s client
	Now     func() time.Time
	// Sleep waits between polls; nil sleeps for real. Tests make it instant.
	Sleep func(ctx context.Context, d time.Duration) error
	// Log, when set, hears each step — the live check prints it.
	Log func(format string, args ...any)
}

// KSeFEnvironments are the Ministry's three systems by their short names.
var KSeFEnvironments = map[string]string{
	"test": "https://api-test.ksef.mf.gov.pl/v2",
	"demo": "https://api-demo.ksef.mf.gov.pl/v2",
	"prod": "https://api.ksef.mf.gov.pl/v2",
}

// Outbound is one invoice to send: the object it stands for, its number
// (for the record), its FA(3) XML.
type Outbound struct {
	ID     string
	Number string
	XML    []byte
	// ContentHash fingerprints the invoice with its creation time fixed:
	// the same content rendered on another day hashes the same, so a
	// refusal is not retried day after day.
	ContentHash string
}

// Submission is what the national system said about one invoice: accepted
// with a KSeF number and its UPO, or rejected with the Ministry's code,
// description and details. Session and Reference locate it in KSeF.
type Submission struct {
	ID, Number  string
	Accepted    bool
	KSeFNumber  string
	UPO         string // the UPO XML, the evidence
	Code        int
	Description string
	Details     []string
	Session     string
	Reference   string
	InvoiceHash string // base64 SHA-256 of the XML sent
	ContentHash string // the outbound's fingerprint with the creation time fixed
	// Final says KSeF's word on this invoice is in hand — a rejection, or
	// an acceptance with its UPO. What is not final was sent and its fate
	// is unknown; the pass books the final ones even when the session
	// failed around them.
	Final bool
}

// Reason is the rejection in one line, for a task's description.
func (s Submission) Reason() string {
	r := fmt.Sprintf("KSeF %d: %s", s.Code, s.Description)
	if len(s.Details) > 0 {
		r += " (" + strings.Join(s.Details, "; ") + ")"
	}
	return r
}

// Polling: the kernel waits for KSeF the way the docs say — status by
// status, with a growing pause, until a final code or the context ends.
const (
	pollFirst = 500 * time.Millisecond
	pollMax   = 5 * time.Second
	// ksefDeadline bounds one call's polling: a status stuck in a
	// processing code would otherwise sleep five seconds forever.
	ksefDeadline = 5 * time.Minute
)

func (c *KSeF2) http() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 60 * time.Second}
}

func (c *KSeF2) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

func (c *KSeF2) sleep(ctx context.Context, d time.Duration) error {
	if c.Sleep != nil {
		return c.Sleep(ctx, d)
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (c *KSeF2) logf(format string, args ...any) {
	if c.Log != nil {
		c.Log(format, args...)
	}
}

// Submit runs one interactive session for the invoices given and reports
// each one's fate. A transport or authentication failure is an error (the
// pass fails, the next pass retries); a rejected invoice is a Submission
// with Accepted false, which is evidence, not failure.
func (c *KSeF2) Submit(ctx context.Context, invoices []Outbound) ([]Submission, error) {
	ctx, cancel := context.WithTimeout(ctx, ksefDeadline)
	defer cancel()
	if len(invoices) == 0 {
		return nil, nil
	}
	if c.BaseURL == "" || c.NIP == "" || (c.Token == "" && c.Signer == nil) {
		return nil, errors.New("ksef: the channel is not configured (environment, NIP, and a certificate or a token)")
	}
	keys, err := c.publicKeys(ctx)
	if err != nil {
		return nil, err
	}
	access, err := c.authenticate(ctx, keys)
	if err != nil {
		return nil, err
	}
	session, aesKey, iv, err := c.openSession(ctx, access, keys)
	if err != nil {
		return nil, err
	}
	c.logf("session %s open", session)
	out := make([]Submission, 0, len(invoices))
	var firstErr error
	for _, inv := range invoices {
		sub := Submission{ID: inv.ID, Number: inv.Number, Session: session, ContentHash: inv.ContentHash}
		plainHash := sha256.Sum256(inv.XML)
		sub.InvoiceHash = base64.StdEncoding.EncodeToString(plainHash[:])
		encrypted, err := aesCBCEncrypt(aesKey, iv, inv.XML)
		if err != nil {
			return out, err
		}
		encHash := sha256.Sum256(encrypted)
		var sent struct {
			ReferenceNumber string `json:"referenceNumber"`
		}
		if err := c.call(ctx, "POST", "/sessions/online/"+session+"/invoices", access, map[string]any{
			"invoiceHash":             sub.InvoiceHash,
			"invoiceSize":             len(inv.XML),
			"encryptedInvoiceHash":    base64.StdEncoding.EncodeToString(encHash[:]),
			"encryptedInvoiceSize":    len(encrypted),
			"encryptedInvoiceContent": base64.StdEncoding.EncodeToString(encrypted),
		}, &sent); err != nil {
			firstErr = fmt.Errorf("send %s: %w", inv.Number, err)
			break // nothing after a failed send is sent; what was sent is polled
		}
		sub.Reference = sent.ReferenceNumber
		c.logf("invoice %s sent as %s", inv.Number, sub.Reference)
		out = append(out, sub)
	}
	// Every invoice sent is polled and, when accepted, its UPO fetched; a
	// failure on one leaves the others' evidence in hand instead of
	// throwing away what KSeF already accepted and numbered.
	for i := range out {
		if err := c.awaitInvoice(ctx, access, &out[i]); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if !out[i].Accepted {
			out[i].Final = true
			continue
		}
		upo, err := c.upo(ctx, access, session, out[i].Reference)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		out[i].UPO, out[i].Final = upo, true
	}
	if firstErr != nil {
		return out, firstErr
	}
	if err := c.call(ctx, "POST", "/sessions/online/"+session+"/close", access, nil, nil); err != nil {
		// The evidence is already in hand; a session that will not close
		// times out on its own in twelve hours.
		c.logf("session %s would not close: %v", session, err)
	} else {
		c.logf("session %s closed", session)
	}
	return out, nil
}

// publicKey is one of the Ministry's published encryption keys.
type publicKey struct {
	Usage []string `json:"usage"`
	ID    string   `json:"publicKeyId"`
	DER   string   `json:"certificate"`
	key   *rsa.PublicKey
}

func (c *KSeF2) publicKeys(ctx context.Context) ([]publicKey, error) {
	var keys []publicKey
	if err := c.call(ctx, "GET", "/security/public-key-certificates", "", nil, &keys); err != nil {
		return nil, err
	}
	for i := range keys {
		der, err := base64.StdEncoding.DecodeString(keys[i].DER)
		if err != nil {
			return nil, fmt.Errorf("ksef public key %s: %w", keys[i].ID, err)
		}
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			return nil, fmt.Errorf("ksef public key %s: %w", keys[i].ID, err)
		}
		pub, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("ksef public key %s: not RSA", keys[i].ID)
		}
		keys[i].key = pub
	}
	return keys, nil
}

func keyFor(keys []publicKey, usage string) (publicKey, error) {
	for _, k := range keys {
		for _, u := range k.Usage {
			if u == usage {
				return k, nil
			}
		}
	}
	return publicKey{}, fmt.Errorf("ksef publishes no key for %s", usage)
}

// authenticate: a challenge; the login document signed with the
// certificate, or the token encrypted with the challenge's timestamp; the
// operation polled to success; the access token redeemed.
func (c *KSeF2) authenticate(ctx context.Context, keys []publicKey) (string, error) {
	ch, err := c.challenge(ctx)
	if err != nil {
		return "", err
	}
	var started authStarted
	if c.Signer != nil {
		// The certificate: the login document for the company's NIP, the
		// identity read from the certificate's subject, signed in XAdES.
		signed, err := c.Signer.SignAuthTokenRequest(ch.Challenge, c.NIP, c.now())
		if err != nil {
			return "", err
		}
		if err := c.callXML(ctx, "/auth/xades-signature", signed, &started); err != nil {
			return "", err
		}
		c.logf("signed with the certificate of %s", c.Signer.Cert.Subject.CommonName)
	} else {
		tokenKey, err := keyFor(keys, "KsefTokenEncryption")
		if err != nil {
			return "", err
		}
		plain := []byte(c.Token + "|" + strconv.FormatInt(ch.TimestampMs, 10))
		encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, tokenKey.key, plain, nil)
		if err != nil {
			return "", err
		}
		if err := c.call(ctx, "POST", "/auth/ksef-token", "", map[string]any{
			"challenge":         ch.Challenge,
			"contextIdentifier": map[string]string{"type": "Nip", "value": c.NIP},
			"encryptedToken":    base64.StdEncoding.EncodeToString(encrypted),
			"publicKeyId":       tokenKey.ID,
		}, &started); err != nil {
			return "", err
		}
	}
	return c.finishLogin(ctx, started)
}

type authChallenge struct {
	Challenge   string `json:"challenge"`
	TimestampMs int64  `json:"timestampMs"`
}

func (c *KSeF2) challenge(ctx context.Context) (authChallenge, error) {
	var ch authChallenge
	err := c.call(ctx, "POST", "/auth/challenge", "", nil, &ch)
	return ch, err
}

type authStarted struct {
	ReferenceNumber     string          `json:"referenceNumber"`
	AuthenticationToken json.RawMessage `json:"authenticationToken"` // {token, validUntil} on the wire
}

// finishLogin polls a started authentication to its end and redeems the
// access token.
func (c *KSeF2) finishLogin(ctx context.Context, started authStarted) (string, error) {
	authToken := tokenString(started.AuthenticationToken)
	if authToken == "" {
		return "", errors.New("ksef: no authentication token started")
	}
	c.logf("authentication %s started", started.ReferenceNumber)
	wait := pollFirst
	for {
		var status struct {
			Status statusInfo `json:"status"`
		}
		if err := c.call(ctx, "GET", "/auth/"+started.ReferenceNumber, authToken, nil, &status); err != nil {
			return "", err
		}
		switch {
		case status.Status.Code == 200:
			var tokens struct {
				AccessToken json.RawMessage `json:"accessToken"`
			}
			if err := c.call(ctx, "POST", "/auth/token/redeem", authToken, nil, &tokens); err != nil {
				return "", err
			}
			access := tokenString(tokens.AccessToken)
			if access == "" {
				return "", errors.New("ksef: no access token redeemed")
			}
			c.logf("authenticated as NIP %s", c.NIP)
			return access, nil
		case status.Status.Code >= 400:
			return "", fmt.Errorf("ksef authentication refused: %s", status.Status)
		}
		if err := c.sleep(ctx, wait); err != nil {
			return "", err
		}
		wait = min(wait*2, pollMax)
	}
}

// Login authenticates as the client is configured — certificate or token —
// and returns the access token.
func (c *KSeF2) Login(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, ksefDeadline)
	defer cancel()
	if c.BaseURL == "" || c.NIP == "" || (c.Token == "" && c.Signer == nil) {
		return "", errors.New("ksef: the channel is not configured (environment, NIP, and a certificate or a token)")
	}
	var keys []publicKey
	if c.Signer == nil {
		var err error
		if keys, err = c.publicKeys(ctx); err != nil {
			return "", err
		}
	}
	return c.authenticate(ctx, keys)
}

// LoginRequest is the login document for the company's NIP with a fresh
// challenge, unsigned — for a person to sign outside Rhea (a trusted profile
// on gov.pl, a qualified signature) within the challenge's ten minutes.
func (c *KSeF2) LoginRequest(ctx context.Context) ([]byte, error) {
	if c.BaseURL == "" || c.NIP == "" {
		return nil, errors.New("ksef: the channel is not configured (environment and NIP)")
	}
	ch, err := c.challenge(ctx)
	if err != nil {
		return nil, err
	}
	return []byte(`<?xml version="1.0" encoding="utf-8"?>` + "\n" + authTokenRequest(ch.Challenge, c.NIP).c14n(nil, nil) + "\n"), nil
}

// LoginSigned completes a login with a document a person signed outside
// Rhea and returns the access token.
func (c *KSeF2) LoginSigned(ctx context.Context, signed []byte) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, ksefDeadline)
	defer cancel()
	var started authStarted
	if err := c.callXML(ctx, "/auth/xades-signature", signed, &started); err != nil {
		return "", err
	}
	return c.finishLogin(ctx, started)
}

// tokenString reads an access token the API returns either as a string
// or as an object with a token field.
func tokenString(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var o struct {
		Token string `json:"token"`
	}
	if json.Unmarshal(raw, &o) == nil {
		return o.Token
	}
	return ""
}

// statusInfo is the API's {code, description, details}.
type statusInfo struct {
	Code        int      `json:"code"`
	Description string   `json:"description"`
	Details     []string `json:"details"`
}

func (s statusInfo) String() string {
	out := fmt.Sprintf("%d %s", s.Code, s.Description)
	if len(s.Details) > 0 {
		out += " (" + strings.Join(s.Details, "; ") + ")"
	}
	return out
}

// openSession declares FA(3) and hands KSeF the session's AES key, wrapped
// with the Ministry's symmetric-key certificate.
func (c *KSeF2) openSession(ctx context.Context, access string, keys []publicKey) (session string, aesKey, iv []byte, err error) {
	symKey, err := keyFor(keys, "SymmetricKeyEncryption")
	if err != nil {
		return "", nil, nil, err
	}
	aesKey, iv = make([]byte, 32), make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return "", nil, nil, err
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", nil, nil, err
	}
	wrapped, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, symKey.key, aesKey, nil)
	if err != nil {
		return "", nil, nil, err
	}
	var opened struct {
		ReferenceNumber string `json:"referenceNumber"`
	}
	if err := c.call(ctx, "POST", "/sessions/online", access, map[string]any{
		"formCode": map[string]string{"systemCode": "FA (3)", "schemaVersion": "1-0E", "value": "FA"},
		"encryption": map[string]string{
			"encryptedSymmetricKey": base64.StdEncoding.EncodeToString(wrapped),
			"initializationVector":  base64.StdEncoding.EncodeToString(iv),
			"publicKeyId":           symKey.ID,
		},
	}, &opened); err != nil {
		return "", nil, nil, err
	}
	return opened.ReferenceNumber, aesKey, iv, nil
}

// awaitInvoice polls one invoice to its final status: 200 is a KSeF
// number, anything from 400 up is a rejection with the Ministry's words.
func (c *KSeF2) awaitInvoice(ctx context.Context, access string, sub *Submission) error {
	wait := pollFirst
	for {
		var st struct {
			KSeFNumber string     `json:"ksefNumber"`
			Status     statusInfo `json:"status"`
		}
		if err := c.call(ctx, "GET", "/sessions/"+sub.Session+"/invoices/"+sub.Reference, access, nil, &st); err != nil {
			return err
		}
		sub.Code, sub.Description, sub.Details = st.Status.Code, st.Status.Description, st.Status.Details
		switch {
		case st.Status.Code == 200:
			sub.Accepted, sub.KSeFNumber = true, st.KSeFNumber
			c.logf("invoice %s accepted: %s", sub.Number, sub.KSeFNumber)
			return nil
		case st.Status.Code >= 400:
			c.logf("invoice %s rejected: %s", sub.Number, st.Status)
			return nil
		}
		if err := c.sleep(ctx, wait); err != nil {
			return err
		}
		wait = min(wait*2, pollMax)
	}
}

// upo fetches the official receipt of one accepted invoice, as XML.
func (c *KSeF2) upo(ctx context.Context, access, session, reference string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/sessions/"+session+"/invoices/"+reference+"/upo", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+access)
	req.Header.Set("Accept", "application/xml")
	res, err := c.http().Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if res.StatusCode != 200 {
		return "", fmt.Errorf("GET upo: %d %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	return string(body), nil
}

// call is one JSON request: the body marshalled, the bearer set when
// given, a 2xx decoded into out, anything else an error carrying the
// API's own words.
func (c *KSeF2) call(ctx context.Context, method, path, bearer string, body any, out any) error {
	var payload io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, payload)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	res, err := c.http().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return fmt.Errorf("%s %s: %d %s", method, path, res.StatusCode, apiWords(raw))
	}
	if out != nil && len(bytes.TrimSpace(raw)) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("%s %s: %w", method, path, err)
		}
	}
	return nil
}

// callXML posts an XML document — the signed login — and decodes the JSON
// answer as call does.
func (c *KSeF2) callXML(ctx context.Context, path string, doc []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+path, bytes.NewReader(doc))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/xml")
	res, err := c.http().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return fmt.Errorf("POST %s: %d %s", path, res.StatusCode, apiWords(raw))
	}
	if out != nil && len(bytes.TrimSpace(raw)) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("POST %s: %w", path, err)
		}
	}
	return nil
}

// apiWords reads the Ministry's error body in either of its shapes — the
// exception list or the problem document — into one line.
func apiWords(raw []byte) string {
	var ex struct {
		Exception struct {
			Details []struct {
				Code        int      `json:"exceptionCode"`
				Description string   `json:"exceptionDescription"`
				Details     []string `json:"details"`
			} `json:"exceptionDetailList"`
		} `json:"exception"`
	}
	if json.Unmarshal(raw, &ex) == nil && len(ex.Exception.Details) > 0 {
		var parts []string
		for _, d := range ex.Exception.Details {
			p := fmt.Sprintf("%d %s", d.Code, d.Description)
			if len(d.Details) > 0 {
				p += " (" + strings.Join(d.Details, "; ") + ")"
			}
			parts = append(parts, p)
		}
		return strings.Join(parts, "; ")
	}
	var problem struct {
		Title  string `json:"title"`
		Detail string `json:"detail"`
		Errors any    `json:"errors"`
	}
	if json.Unmarshal(raw, &problem) == nil && (problem.Title != "" || problem.Detail != "") {
		s := strings.TrimSpace(problem.Title + ": " + problem.Detail)
		if problem.Errors != nil {
			if b, err := json.Marshal(problem.Errors); err == nil {
				s += " " + string(b)
			}
		}
		return s
	}
	s := strings.TrimSpace(string(raw))
	if len(s) > 400 {
		s = s[:400] + "…"
	}
	return s
}

// aesCBCEncrypt is AES-256-CBC with PKCS#7 padding, as the session declares.
func aesCBCEncrypt(key, iv, plain []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	pad := aes.BlockSize - len(plain)%aes.BlockSize
	padded := make([]byte, len(plain)+pad)
	copy(padded, plain)
	for i := len(plain); i < len(padded); i++ {
		padded[i] = byte(pad)
	}
	out := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, padded)
	return out, nil
}

// aesCBCDecrypt undoes aesCBCEncrypt; the fake server and the tests read
// what the client sent.
func aesCBCDecrypt(key, iv, encrypted []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(encrypted) == 0 || len(encrypted)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not whole blocks")
	}
	out := make([]byte, len(encrypted))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(out, encrypted)
	pad := int(out[len(out)-1])
	if pad == 0 || pad > aes.BlockSize || pad > len(out) {
		return nil, errors.New("bad padding")
	}
	return out[:len(out)-pad], nil
}

// KSeFNumberValid checks a KSeF number's shape and CRC-8 (polynomial 0x07,
// initial 0x00) over its first 32 characters, as the Ministry documents it.
func KSeFNumberValid(n string) bool {
	if len(n) != 35 || n[10] != '-' || n[19] != '-' || n[32] != '-' {
		return false
	}
	want, err := strconv.ParseUint(n[33:], 16, 8)
	if err != nil {
		return false
	}
	return crc8([]byte(n[:32])) == byte(want)
}

func crc8(data []byte) byte {
	var crc byte
	for _, b := range data {
		crc ^= b
		for i := 0; i < 8; i++ {
			if crc&0x80 != 0 {
				crc = crc<<1 ^ 0x07
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}
