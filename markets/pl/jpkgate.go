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
	"embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// The Ministry's gateway for JPK files, e-dokumenty.mf.gov.pl (the interface
// specification v5.6.0 of 10 September 2026; KK, 2026-10-27: the JPK goes out
// from Rhea, signed by a person). A file goes out as one ZIP (DEFLATE),
// encrypted AES-256-CBC with a fresh key; the key is wrapped with the
// Ministry's public certificate (RSA, PKCS#1 v1.5) and named in the request,
// InitUpload, beside the document's SHA-256 and the encrypted part's MD5. The
// request is signed by a person — a qualified signature, or a trusted
// profile on gov.pl — with their own tool, outside Rhea: Rhea prepares the
// package and the request, checks that the signed file it is handed back is
// that request, and sends. It never holds anyone's key. Then four calls:
// InitUploadSigned opens a session and names where the part goes (the
// Ministry's storage), Put Blob uploads it, FinishUpload closes the session,
// Status says what became of the document, with the UPO when it is filed.
// Standard library only.

// JPKGatewayNamespace is the request's namespace (InitUpload.xsd).
const JPKGatewayNamespace = "http://e-dokumenty.mf.gov.pl"

// jpkAPIVersion is the REST version every request names; the schema fixes it.
const jpkAPIVersion = "01.02.01.20160617"

// jpkMaxPart is the largest part the storage takes (60 MB, §1.2.1).
const jpkMaxPart = 60 << 20

// JPKEnvironments are the gateway's two systems. A package is sealed with
// one environment's certificate and opens only there, which is the safety.
var JPKEnvironments = map[string]string{
	"test": "https://test-e-dokumenty.mf.gov.pl",
	"prod": "https://e-dokumenty.mf.gov.pl",
}

// The Ministry's encryption certificates, as published on podatki.gov.pl
// (JPK_VAT z deklaracją, pliki do pobrania): production from 18 July 2025,
// test from 9 July 2025, both valid to 7 July 2027. A new one is announced
// there; replacing the file is a fix to this pack.
//
//go:embed certs/jpk-test.pem certs/jpk-prod.pem
var jpkCertFiles embed.FS

// JPKCertificate is the Ministry's encryption certificate for an environment.
func JPKCertificate(env string) (*x509.Certificate, error) {
	if _, ok := JPKEnvironments[env]; !ok {
		return nil, fmt.Errorf("jpk gateway: no environment %q (test | prod)", env)
	}
	b, err := jpkCertFiles.ReadFile("certs/jpk-" + env + ".pem")
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("jpk gateway: the %s certificate is no PEM CERTIFICATE", env)
	}
	return x509.ParseCertificate(block.Bytes)
}

// jpkInitUpload is the request, as InitUpload.xsd lays it out, for one
// document in one part.
type jpkInitUpload struct {
	XMLName       xml.Name          `xml:"http://e-dokumenty.mf.gov.pl InitUpload"`
	DocumentType  string            `xml:"DocumentType"`
	Version       string            `xml:"Version"`
	EncryptionKey jpkEncryptionKey  `xml:"EncryptionKey"`
	Document      jpkUploadDocument `xml:"DocumentList>Document"`
}

type jpkEncryptionKey struct {
	Algorithm string `xml:"algorithm,attr"`
	Mode      string `xml:"mode,attr"`
	Padding   string `xml:"padding,attr"`
	Encoding  string `xml:"encoding,attr"`
	Value     string `xml:",chardata"`
}

type jpkUploadDocument struct {
	FormCode          jpkUploadFormCode    `xml:"FormCode"`
	FileName          string               `xml:"FileName"`
	ContentLength     int64                `xml:"ContentLength"`
	HashValue         jpkHash              `xml:"HashValue"`
	FileSignatureList jpkFileSignatureList `xml:"FileSignatureList"`
}

type jpkUploadFormCode struct {
	SystemCode    string `xml:"systemCode,attr"`
	SchemaVersion string `xml:"schemaVersion,attr"`
	Value         string `xml:",chardata"`
}

type jpkHash struct {
	Algorithm string `xml:"algorithm,attr"`
	Encoding  string `xml:"encoding,attr"`
	Value     string `xml:",chardata"`
}

type jpkFileSignatureList struct {
	FilesNumber int `xml:"filesNumber,attr"`
	Packaging   struct {
		SplitZip struct {
			Type string `xml:"type,attr"`
			Mode string `xml:"mode,attr"`
		} `xml:"SplitZip"`
	} `xml:"Packaging"`
	Encryption struct {
		AES jpkAES `xml:"AES"`
	} `xml:"Encryption"`
	FileSignature []jpkFileSignature `xml:"FileSignature"`
}

type jpkAES struct {
	Size    int    `xml:"size,attr"`
	Block   int    `xml:"block,attr"`
	Mode    string `xml:"mode,attr"`
	Padding string `xml:"padding,attr"`
	IV      struct {
		Bytes    int    `xml:"bytes,attr"`
		Encoding string `xml:"encoding,attr"`
		Value    string `xml:",chardata"`
	} `xml:"IV"`
}

type jpkFileSignature struct {
	OrdinalNumber int     `xml:"OrdinalNumber"`
	FileName      string  `xml:"FileName"`
	ContentLength int     `xml:"ContentLength"`
	HashValue     jpkHash `xml:"HashValue"`
}

// tree is the request as the signer's element tree, in the schema's order.
func (r jpkInitUpload) tree() *xel {
	d := r.Document
	list := el("", "FileSignatureList").attr("filesNumber", strconv.Itoa(d.FileSignatureList.FilesNumber))
	aes := d.FileSignatureList.Encryption.AES
	list.add(
		el("", "Packaging").add(el("", "SplitZip").attr("type", "split").attr("mode", "zip")),
		el("", "Encryption").add(el("", "AES").attr("size", strconv.Itoa(aes.Size)).attr("block", strconv.Itoa(aes.Block)).
			attr("mode", aes.Mode).attr("padding", aes.Padding).add(
			el("", "IV").attr("bytes", strconv.Itoa(aes.IV.Bytes)).attr("encoding", aes.IV.Encoding).txt(aes.IV.Value))),
	)
	for _, fs := range d.FileSignatureList.FileSignature {
		list.add(el("", "FileSignature").add(
			el("", "OrdinalNumber").txt(strconv.Itoa(fs.OrdinalNumber)),
			el("", "FileName").txt(fs.FileName),
			el("", "ContentLength").txt(strconv.Itoa(fs.ContentLength)),
			el("", "HashValue").attr("algorithm", fs.HashValue.Algorithm).attr("encoding", fs.HashValue.Encoding).txt(fs.HashValue.Value),
		))
	}
	ek := r.EncryptionKey
	return el("", "InitUpload").declare("", JPKGatewayNamespace).add(
		el("", "DocumentType").txt(r.DocumentType),
		el("", "Version").txt(r.Version),
		el("", "EncryptionKey").attr("algorithm", ek.Algorithm).attr("mode", ek.Mode).attr("padding", ek.Padding).
			attr("encoding", ek.Encoding).txt(ek.Value),
		el("", "DocumentList").add(el("", "Document").add(
			el("", "FormCode").attr("systemCode", d.FormCode.SystemCode).attr("schemaVersion", d.FormCode.SchemaVersion).txt(d.FormCode.Value),
			el("", "FileName").txt(d.FileName),
			el("", "ContentLength").txt(strconv.FormatInt(d.ContentLength, 10)),
			el("", "HashValue").attr("algorithm", d.HashValue.Algorithm).attr("encoding", d.HashValue.Encoding).txt(d.HashValue.Value),
			list,
		)),
	)
}

// jpkDeclaration opens every request Rhea writes: the gateway reads utf-8
// declared so, and refuses another declaration (its code 101).
const jpkDeclaration = `<?xml version="1.0" encoding="utf-8"?>` + "\n"

// jpkFileName is what the schema admits as a file's name.
var jpkFileName = regexp.MustCompile(`^[a-zA-Z0-9_\.\-]{5,55}$`)

// JPKPackage is a JPK file made ready for the gateway: the encrypted part
// that goes to the Ministry's storage, its name, and the request a person
// signs before it is sent.
type JPKPackage struct {
	Part     []byte
	PartName string
	Request  []byte
}

// PackJPK seals a JPK file for the environment whose certificate is given:
// zipped, encrypted with a fresh key, the key wrapped for the Ministry, the
// request naming the document by its hash and its form code as the file's
// own header states it.
func PackJPK(name string, doc []byte, cert *x509.Certificate) (JPKPackage, error) {
	if !jpkFileName.MatchString(name) {
		return JPKPackage{}, fmt.Errorf("jpk gateway: the file name %q is not one the gateway admits (5–55 of letters, digits, _ . -)", name)
	}
	pub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return JPKPackage{}, fmt.Errorf("jpk gateway: the Ministry's key is RSA; this certificate carries %T", cert.PublicKey)
	}
	form, err := jpkFormCodeOf(doc)
	if err != nil {
		return JPKPackage{}, err
	}
	var zipped bytes.Buffer
	zw := zip.NewWriter(&zipped)
	w, err := zw.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Deflate})
	if err != nil {
		return JPKPackage{}, err
	}
	if _, err := w.Write(doc); err != nil {
		return JPKPackage{}, err
	}
	if err := zw.Close(); err != nil {
		return JPKPackage{}, err
	}
	if zipped.Len() > jpkMaxPart {
		return JPKPackage{}, fmt.Errorf("jpk gateway: %s is %d bytes zipped; over 60 MB the gateway takes a file in parts, which Rhea does not split", name, zipped.Len())
	}
	key, iv := make([]byte, 32), make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		return JPKPackage{}, err
	}
	if _, err := rand.Read(iv); err != nil {
		return JPKPackage{}, err
	}
	part, err := aesCBCEncrypt(key, iv, zipped.Bytes())
	if err != nil {
		return JPKPackage{}, err
	}
	wrapped, err := rsa.EncryptPKCS1v15(rand.Reader, pub, key)
	if err != nil {
		return JPKPackage{}, err
	}
	partName := strings.TrimSuffix(name, ".xml")
	if len(partName) > 45 {
		partName = partName[:45]
	}
	partName += ".zip.aes"
	docSum, partSum := sha256.Sum256(doc), md5.Sum(part)
	req := jpkInitUpload{
		DocumentType: "JPK", Version: jpkAPIVersion,
		EncryptionKey: jpkEncryptionKey{Algorithm: "RSA", Mode: "ECB", Padding: "PKCS#1", Encoding: "Base64",
			Value: base64.StdEncoding.EncodeToString(wrapped)},
		Document: jpkUploadDocument{
			FormCode: form, FileName: name, ContentLength: int64(len(doc)),
			HashValue: jpkHash{Algorithm: "SHA-256", Encoding: "Base64", Value: base64.StdEncoding.EncodeToString(docSum[:])},
		},
	}
	fl := &req.Document.FileSignatureList
	fl.FilesNumber = 1
	fl.Packaging.SplitZip.Type, fl.Packaging.SplitZip.Mode = "split", "zip"
	fl.Encryption.AES = jpkAES{Size: 256, Block: 16, Mode: "CBC", Padding: "PKCS#7"}
	fl.Encryption.AES.IV.Bytes, fl.Encryption.AES.IV.Encoding = 16, "Base64"
	fl.Encryption.AES.IV.Value = base64.StdEncoding.EncodeToString(iv)
	fl.FileSignature = []jpkFileSignature{{OrdinalNumber: 1, FileName: partName, ContentLength: len(part),
		HashValue: jpkHash{Algorithm: "MD5", Encoding: "Base64", Value: base64.StdEncoding.EncodeToString(partSum[:])}}}
	return JPKPackage{Part: part, PartName: partName, Request: []byte(jpkDeclaration + req.tree().c14n(nil, nil))}, nil
}

// jpkFormCodeOf reads KodFormularza from a JPK file's header: the request
// repeats it (systemCode, schemaVersion, the code itself).
func jpkFormCodeOf(doc []byte) (jpkUploadFormCode, error) {
	d := xml.NewDecoder(bytes.NewReader(doc))
	for {
		tok, err := d.Token()
		if err == io.EOF {
			return jpkUploadFormCode{}, errors.New("jpk gateway: the file has no KodFormularza in its header — it is no JPK")
		}
		if err != nil {
			return jpkUploadFormCode{}, fmt.Errorf("jpk gateway: the file is no XML: %w", err)
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "KodFormularza" {
			continue
		}
		var code struct {
			System  string `xml:"kodSystemowy,attr"`
			Version string `xml:"wersjaSchemy,attr"`
			Value   string `xml:",chardata"`
		}
		if err := d.DecodeElement(&code, &se); err != nil {
			return jpkUploadFormCode{}, err
		}
		if code.System == "" || code.Version == "" || strings.TrimSpace(code.Value) == "" {
			return jpkUploadFormCode{}, errors.New("jpk gateway: KodFormularza without kodSystemowy, wersjaSchemy or the code")
		}
		return jpkUploadFormCode{SystemCode: code.System, SchemaVersion: code.Version, Value: strings.TrimSpace(code.Value)}, nil
	}
}

// SignJPKRequest signs a request enveloped (XAdES-BES, RSA-SHA256): what a
// person's tool does, here for the test gateway, which takes a self-signed
// certificate unless it is asked to check (enableValidateQualifiedSignature).
func SignJPKRequest(request []byte, s *CertSigner, at time.Time) ([]byte, error) {
	var req jpkInitUpload
	if err := xml.Unmarshal(request, &req); err != nil {
		return nil, fmt.Errorf("jpk gateway: the request is no InitUpload: %w", err)
	}
	signed, err := s.signEnveloped(req.tree(), at)
	if err != nil {
		return nil, err
	}
	return append([]byte(jpkDeclaration), signed...), nil
}

// SelfSignedSignerRSA is a person's signing certificate made by Rhea, RSA
// 2048 as the gateway signs (RSA-SHA256): the person named, their NIP in
// serialNumber as Polish qualified certificates carry it (TINPL-<NIP>). The
// gateway wants the signer's NIP or PESEL, to check who may sign for the
// taxpayer: a company seal's VATPL-<NIP> was refused on the test gateway
// (status 423, 2026-10-27), TINPL-<NIP> filed. For the test gateway only;
// production takes a qualified signature or a trusted profile.
func SelfSignedSignerRSA(nip, givenName, surname string, now time.Time) (*CertSigner, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 63))
	if err != nil {
		return nil, err
	}
	name := pkix.Name{
		CommonName: givenName + " " + surname, Country: []string{"PL"}, SerialNumber: "TINPL-" + nip,
		ExtraNames: []pkix.AttributeTypeAndValue{{Type: oidGivenName, Value: givenName}, {Type: oidSurname, Value: surname}},
	}
	tmpl := &x509.Certificate{
		SerialNumber: serial, Subject: name, Issuer: name,
		NotBefore: now.Add(-time.Hour), NotAfter: now.AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageContentCommitment,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, err
	}
	return &CertSigner{Cert: cert, Key: key}, nil
}

// What can be wrong with a signed file handed back, beside a file that is
// no XML at all: no signature, no request in it, another package's request.
var (
	ErrJPKUnsigned     = errors.New("jpk gateway: the file carries no signature — sign the request with a qualified signature or a trusted profile and hand back the signed file")
	ErrJPKNoRequest    = errors.New("jpk gateway: the signed file holds no InitUpload request")
	ErrJPKOtherRequest = errors.New("jpk gateway: the signed file is another request than the one prepared for this filing — sign the request downloaded for it")
)

// CheckSignedJPKRequest says whether a signed file is the prepared request,
// signed: it carries a signature, and the request in it names this package
// — its key, its document, its part. A file signed for another package
// would open a session whose key does not fit the part. Whether the
// signature is valid, and whether its signer may sign for the taxpayer, is
// the gateway's to say; its answer comes back as the filing's outcome.
func CheckSignedJPKRequest(prepared, signed []byte) error {
	var want jpkInitUpload
	if err := xml.Unmarshal(prepared, &want); err != nil {
		return fmt.Errorf("jpk gateway: the prepared request is no InitUpload: %w", err)
	}
	if !jpkHasSignature(signed) {
		return ErrJPKUnsigned
	}
	got, ok := jpkFindRequest(signed)
	if !ok {
		return ErrJPKNoRequest
	}
	wd, gd := want.Document, got.Document
	same := got.EncryptionKey.Value == want.EncryptionKey.Value &&
		gd.FileName == wd.FileName && gd.ContentLength == wd.ContentLength && gd.HashValue.Value == wd.HashValue.Value &&
		gd.FormCode == wd.FormCode &&
		gd.FileSignatureList.Encryption.AES.IV.Value == wd.FileSignatureList.Encryption.AES.IV.Value &&
		len(gd.FileSignatureList.FileSignature) == len(wd.FileSignatureList.FileSignature)
	for i := 0; same && i < len(wd.FileSignatureList.FileSignature); i++ {
		g, w := gd.FileSignatureList.FileSignature[i], wd.FileSignatureList.FileSignature[i]
		same = g.FileName == w.FileName && g.ContentLength == w.ContentLength && g.HashValue.Value == w.HashValue.Value
	}
	if !same {
		return ErrJPKOtherRequest
	}
	return nil
}

// jpkHasSignature reports whether an XML file carries a ds:Signature.
func jpkHasSignature(b []byte) bool {
	d := xml.NewDecoder(bytes.NewReader(b))
	for {
		tok, err := d.Token()
		if err != nil {
			return false
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Space == nsDS && se.Name.Local == "Signature" {
			return true
		}
	}
}

// jpkFindRequest finds the InitUpload in a signed file: the root of an
// enveloped signature, or inside the ds:Object of an enveloping one —
// inline, or base64 when the signer used the base64 transform.
func jpkFindRequest(b []byte) (jpkInitUpload, bool) {
	d := xml.NewDecoder(bytes.NewReader(b))
	inObject := 0
	var text strings.Builder
	for {
		tok, err := d.Token()
		if err != nil {
			return jpkInitUpload{}, false
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Space == JPKGatewayNamespace && t.Name.Local == "InitUpload" {
				var req jpkInitUpload
				if d.DecodeElement(&req, &t) != nil {
					return jpkInitUpload{}, false
				}
				return req, true
			}
			if t.Name.Space == nsDS && t.Name.Local == "Object" {
				inObject++
				text.Reset()
			}
		case xml.CharData:
			if inObject > 0 {
				text.Write(t)
			}
		case xml.EndElement:
			if t.Name.Space == nsDS && t.Name.Local == "Object" && inObject > 0 {
				inObject--
				raw, err := base64.StdEncoding.DecodeString(strings.Join(strings.Fields(text.String()), ""))
				if err == nil {
					if req, ok := jpkFindRequest(raw); ok {
						return req, true
					}
				}
			}
		}
	}
}

// JPKGatewayClient is the boundary to the gateway: which environment a
// package is sealed for, the certificate to seal with, sending a signed
// request with its part, the status of what was sent. JPKGateway is the
// Ministry's; FakeJPKGateway answers for simulations, named explicitly.
type JPKGatewayClient interface {
	Environment() string
	Certificate() (*x509.Certificate, error)
	Send(ctx context.Context, signedRequest, part []byte) (reference string, err error)
	Status(ctx context.Context, reference string) (JPKStatus, error)
}

// JPKStatus is the gateway's word on a document sent: 1xx the session, 120
// the document being verified, 200 filed with its UPO, 300 a reference the
// gateway does not know, 4xx refused (§2.2.4).
type JPKStatus struct {
	Code        int
	Description string
	Details     string
	UPO         string // the UPO XML, the evidence of filing
	Timestamp   string
}

// Final says the gateway has decided: filed, or not.
func (s JPKStatus) Final() bool { return s.Code == 200 || s.Code >= 300 }

// Accepted says the document is filed.
func (s JPKStatus) Accepted() bool { return s.Code == 200 }

// JPKRefusal is the gateway refusing a request (400 at InitUploadSigned or
// FinishUpload): its code, its words, and for a duplicate of a document it
// already processed (170) the original's reference.
type JPKRefusal struct {
	Code      int
	Message   string
	Errors    []string
	RequestID string
	Original  string
}

func (r *JPKRefusal) Error() string {
	s := fmt.Sprintf("jpk gateway refused (%d): %s", r.Code, r.Message)
	if len(r.Errors) > 0 {
		s += " (" + strings.Join(r.Errors, "; ") + ")"
	}
	return s
}

// jpkCode reads the gateway's code, written as a number or as a string.
type jpkCode int

func (c *jpkCode) UnmarshalJSON(b []byte) error {
	s := strings.Trim(strings.TrimSpace(string(b)), `"`)
	if s == "" || s == "null" {
		*c = 0
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("jpk gateway: code %s is no number", b)
	}
	*c = jpkCode(n)
	return nil
}

// jpkStorage is the Ministry's storage, where parts go (§2.2):
// taxdocumentstorageNN.blob.core.windows.net, NNtst on the test gateway.
var jpkStorage = regexp.MustCompile(`^taxdocumentstorage[0-9]{2}(tst)?\.blob\.core\.windows\.net$`)

func jpkStorageHost(host string) bool { return jpkStorage.MatchString(host) }

// JPKGateway speaks to the Ministry's gateway.
type JPKGateway struct {
	Env     string // "test" | "prod": the certificate a package is sealed with
	BaseURL string
	HTTP    *http.Client // nil: a five-minute client, a part may be large
	// Storage says where a part may be uploaded; nil admits the Ministry's
	// storage over HTTPS only — the company's file goes nowhere else.
	Storage func(host string) bool
	// Log, when set, hears each step.
	Log func(format string, args ...any)
}

func (g *JPKGateway) Environment() string { return g.Env }

func (g *JPKGateway) Certificate() (*x509.Certificate, error) { return JPKCertificate(g.Env) }

func (g *JPKGateway) http() *http.Client {
	if g.HTTP != nil {
		return g.HTTP
	}
	return &http.Client{Timeout: 5 * time.Minute}
}

func (g *JPKGateway) logf(format string, args ...any) {
	if g.Log != nil {
		g.Log(format, args...)
	}
}

// jpkOriginal is the original's reference in a duplicate's words.
var jpkOriginal = regexp.MustCompile(`[0-9a-f]{32}`)

// refusal reads a 400's body; anything else is an error the next pass
// retries.
func jpkRefusalOf(step string, status int, body []byte) error {
	if status == 400 {
		var r struct {
			Message   string
			Code      jpkCode
			Errors    json.RawMessage
			RequestId string
		}
		if err := json.Unmarshal(body, &r); err == nil && r.Message != "" {
			out := &JPKRefusal{Code: int(r.Code), Message: strings.TrimSpace(r.Message), RequestID: r.RequestId}
			var list []string
			if json.Unmarshal(r.Errors, &list) == nil {
				out.Errors = list
			} else if len(r.Errors) > 0 && string(r.Errors) != "null" {
				out.Errors = []string{strings.Trim(string(r.Errors), `"`)}
			}
			if out.Code == 170 {
				out.Original = jpkOriginal.FindString(out.Message)
			}
			return out
		}
	}
	snippet := string(body)
	if len(snippet) > 300 {
		snippet = snippet[:300]
	}
	return fmt.Errorf("jpk gateway: %s answered HTTP %d: %s", step, status, strings.TrimSpace(snippet))
}

func (g *JPKGateway) do(ctx context.Context, method, target, contentType string, body []byte, headers map[string]string) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, target, bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := g.http().Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	return resp.StatusCode, b, err
}

// Send opens a session with the signed request, uploads the part where the
// gateway says, and finishes the session; the reference locates the
// document from then on. The signed request goes byte for byte — a changed
// byte breaks the signature.
func (g *JPKGateway) Send(ctx context.Context, signedRequest, part []byte) (string, error) {
	status, body, err := g.do(ctx, "POST", g.BaseURL+"/api/Storage/InitUploadSigned", "application/xml", signedRequest, nil)
	if err != nil {
		return "", fmt.Errorf("jpk gateway: InitUploadSigned: %w", err)
	}
	if status != 200 {
		return "", jpkRefusalOf("InitUploadSigned", status, body)
	}
	var session struct {
		ReferenceNumber         string
		TimeoutInSec            int
		RequestToUploadFileList []struct {
			BlobName, FileName, Url, Method string
			HeaderList                      []struct{ Key, Value string }
		}
	}
	if err := json.Unmarshal(body, &session); err != nil {
		return "", fmt.Errorf("jpk gateway: InitUploadSigned answered no session: %w", err)
	}
	ref := strings.TrimSpace(session.ReferenceNumber)
	if ref == "" || len(session.RequestToUploadFileList) != 1 {
		return "", fmt.Errorf("jpk gateway: InitUploadSigned opened %q with %d uploads; the request declares one part", ref, len(session.RequestToUploadFileList))
	}
	up := session.RequestToUploadFileList[0]
	u, err := url.Parse(up.Url)
	if err != nil {
		return "", fmt.Errorf("jpk gateway: the upload URL: %w", err)
	}
	allowed := g.Storage
	if allowed == nil {
		allowed = func(host string) bool { return u.Scheme == "https" && jpkStorageHost(host) }
	}
	if !allowed(u.Hostname()) {
		return "", fmt.Errorf("jpk gateway: the session %s names an upload to %s, which is not the Ministry's storage — nothing is uploaded", ref, u.Host)
	}
	if up.Method != "" && up.Method != "PUT" {
		return "", fmt.Errorf("jpk gateway: the session %s asks for an upload by %s, the storage takes PUT", ref, up.Method)
	}
	headers := map[string]string{}
	for _, h := range up.HeaderList {
		headers[h.Key] = h.Value
	}
	g.logf("session %s: uploading %d bytes", ref, len(part))
	status, body, err = g.do(ctx, "PUT", up.Url, "", part, headers)
	if err != nil {
		return "", fmt.Errorf("jpk gateway: Put Blob for %s: %w", ref, err)
	}
	if status != 201 {
		return "", jpkRefusalOf("Put Blob", status, body)
	}
	finish, _ := json.Marshal(map[string]any{"ReferenceNumber": ref, "AzureBlobNameList": []string{up.BlobName}})
	status, body, err = g.do(ctx, "POST", g.BaseURL+"/api/Storage/FinishUpload", "application/json", finish, nil)
	if err != nil {
		return "", fmt.Errorf("jpk gateway: FinishUpload for %s: %w", ref, err)
	}
	if status != 200 {
		return "", jpkRefusalOf("FinishUpload", status, body)
	}
	g.logf("session %s: finished", ref)
	return ref, nil
}

// Status asks what became of a document sent.
func (g *JPKGateway) Status(ctx context.Context, reference string) (JPKStatus, error) {
	status, body, err := g.do(ctx, "GET", g.BaseURL+"/api/Storage/Status/"+url.PathEscape(reference), "", nil, nil)
	if err != nil {
		return JPKStatus{}, fmt.Errorf("jpk gateway: Status %s: %w", reference, err)
	}
	if status != 200 {
		return JPKStatus{}, jpkRefusalOf("Status", status, body)
	}
	var s struct {
		Code                                 jpkCode
		Description, Details, Upo, Timestamp string
	}
	if err := json.Unmarshal(body, &s); err != nil {
		return JPKStatus{}, fmt.Errorf("jpk gateway: Status %s answered no status: %w", reference, err)
	}
	return JPKStatus{Code: int(s.Code), Description: s.Description, Details: s.Details, UPO: s.Upo, Timestamp: s.Timestamp}, nil
}

// FakeJPKGateway answers like the gateway without it, for simulations and
// demos: every document filed at once, its UPO saying it is fake. Only
// RHEA_JPK_ENV=fake gives it — never the absence of a setting.
type FakeJPKGateway struct{}

func (FakeJPKGateway) Environment() string { return "fake" }

func (FakeJPKGateway) Certificate() (*x509.Certificate, error) { return JPKCertificate("test") }

func (FakeJPKGateway) Send(_ context.Context, signedRequest, _ []byte) (string, error) {
	sum := sha256.Sum256(signedRequest)
	return hex.EncodeToString(sum[:16]), nil
}

func (FakeJPKGateway) Status(_ context.Context, reference string) (JPKStatus, error) {
	return JPKStatus{Code: 200, Description: "Przetwarzanie dokumentu zakończone poprawnie, pobierz UPO", Details: reference,
		UPO: `<?xml version="1.0" encoding="UTF-8"?><UPO fake="true"><NumerReferencyjny>` + reference +
			`</NumerReferencyjny><Uwaga>Bramka udawana (RHEA_JPK_ENV=fake): dokument nie został złożony.</Uwaga></UPO>`}, nil
}

type refusingJPKGateway struct{ err error }

func (r refusingJPKGateway) Environment() string                     { return "" }
func (r refusingJPKGateway) Certificate() (*x509.Certificate, error) { return nil, r.err }
func (r refusingJPKGateway) Send(context.Context, []byte, []byte) (string, error) {
	return "", r.err
}
func (r refusingJPKGateway) Status(context.Context, string) (JPKStatus, error) {
	return JPKStatus{}, r.err
}

// JPKGatewayFromEnv is the gateway the operator named in RHEA_JPK_ENV (test
// | prod | fake). Unset or misspelled, it refuses to send: a fake UPO in
// real books would say a return was filed that was not.
func JPKGatewayFromEnv() JPKGatewayClient {
	env := os.Getenv("RHEA_JPK_ENV")
	switch base, known := JPKEnvironments[env]; {
	case env == "":
		return refusingJPKGateway{errors.New("jpk gateway not configured: RHEA_JPK_ENV is not set (test | prod) — nothing is sent")}
	case env == "fake":
		return FakeJPKGateway{}
	case known:
		return &JPKGateway{Env: env, BaseURL: base, Log: func(format string, args ...any) {
			log.Printf("jpk gateway (%s): "+format, append([]any{env}, args...)...)
		}}
	default:
		return refusingJPKGateway{fmt.Errorf("jpk gateway: RHEA_JPK_ENV=%q is not one of test, prod, fake — nothing is sent", env)}
	}
}
