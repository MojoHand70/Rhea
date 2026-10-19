package pl

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strings"
	"time"
)

// Certificate login (the Poland pack review, KSEF-21): from 1 January 2027
// a KSeF token no longer logs in (regulation Dz.U. 2025 poz. 1815, § 13); a
// certificate does. The login document, AuthTokenRequest, is signed in
// XAdES — an enveloped XAdES-BES signature: the document's own digest and the
// signed properties' digest in SignedInfo, SignedInfo signed with the
// certificate's key, the certificate in KeyInfo, its digest among the signed
// properties. Standard library only: the documents are ours, small and
// fixed in shape, so the signer builds them as a tree that prints its own
// exclusive canonical form (http://www.w3.org/2001/10/xml-exc-c14n#) — what
// the verifier recomputes is what was digested and signed.

// CertSigner is a certificate and the private key that signs with it — a
// secret like the token, kept outside the log.
type CertSigner struct {
	Cert *x509.Certificate
	Key  crypto.Signer
}

// LoadCertSigner reads a certificate and its private key from PEM files: the
// certificate (CERTIFICATE), the key as PKCS#8, EC or RSA. An encrypted key
// is refused — decrypt it before handing it over.
func LoadCertSigner(certPath, keyPath string) (*CertSigner, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("ksef certificate: %w", err)
	}
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("ksef key: %w", err)
	}
	return ParseCertSigner(certPEM, keyPEM)
}

// ParseCertSigner is LoadCertSigner over bytes.
func ParseCertSigner(certPEM, keyPEM []byte) (*CertSigner, error) {
	cb, _ := pem.Decode(certPEM)
	if cb == nil || cb.Type != "CERTIFICATE" {
		return nil, errors.New("ksef certificate: no PEM CERTIFICATE block")
	}
	cert, err := x509.ParseCertificate(cb.Bytes)
	if err != nil {
		return nil, fmt.Errorf("ksef certificate: %w", err)
	}
	kb, _ := pem.Decode(keyPEM)
	if kb == nil {
		return nil, errors.New("ksef key: no PEM block")
	}
	if strings.Contains(kb.Type, "ENCRYPTED") {
		return nil, errors.New("ksef key: encrypted — decrypt it before handing it over")
	}
	var key any
	switch kb.Type {
	case "PRIVATE KEY":
		key, err = x509.ParsePKCS8PrivateKey(kb.Bytes)
	case "EC PRIVATE KEY":
		key, err = x509.ParseECPrivateKey(kb.Bytes)
	case "RSA PRIVATE KEY":
		key, err = x509.ParsePKCS1PrivateKey(kb.Bytes)
	default:
		return nil, fmt.Errorf("ksef key: unknown PEM block %q", kb.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("ksef key: %w", err)
	}
	signer, ok := key.(crypto.Signer)
	if !ok {
		return nil, errors.New("ksef key: not a signing key")
	}
	if !samePublic(signer.Public(), cert.PublicKey) {
		return nil, errors.New("ksef key: it is not the certificate's key")
	}
	return &CertSigner{Cert: cert, Key: signer}, nil
}

func samePublic(a, b crypto.PublicKey) bool {
	type equaler interface{ Equal(crypto.PublicKey) bool }
	e, ok := a.(equaler)
	return ok && e.Equal(b)
}

// PEM writes the certificate and its key as PEM (PKCS#8) — what
// LoadCertSigner reads back.
func (s *CertSigner) PEM() (certPEM, keyPEM []byte, err error) {
	der, err := x509.MarshalPKCS8PrivateKey(s.Key)
	if err != nil {
		return nil, nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: s.Cert.Raw}),
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), nil
}

// oidOrganizationIdentifier is X.520 organizationIdentifier, where a
// company seal carries its NIP as VATPL-<NIP>.
var oidOrganizationIdentifier = asn1.ObjectIdentifier{2, 5, 4, 97}

// SelfSignedSeal makes a company seal for a NIP, signed by itself — accepted
// for login on the Ministry's TEST environment only, where any integrator may
// act for any made-up NIP; DEMO and production take only a qualified or a
// KSeF-issued certificate. EC P-256, as the Ministry recommends.
func SelfSignedSeal(nip, organization string, now time.Time) (*CertSigner, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 63))
	if err != nil {
		return nil, err
	}
	name := pkix.Name{
		CommonName: organization, Organization: []string{organization}, Country: []string{"PL"},
		ExtraNames: []pkix.AttributeTypeAndValue{{Type: oidOrganizationIdentifier, Value: "VATPL-" + nip}},
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

const (
	nsAuthToken = "http://ksef.mf.gov.pl/auth/token/2.0"
	nsDS        = "http://www.w3.org/2000/09/xmldsig#"
	nsXAdES     = "http://uri.etsi.org/01903/v1.3.2#"
	algExcC14N  = "http://www.w3.org/2001/10/xml-exc-c14n#"
	algEnvelope = "http://www.w3.org/2000/09/xmldsig#enveloped-signature"
	algSHA256   = "http://www.w3.org/2001/04/xmlenc#sha256"
	algECDSA256 = "http://www.w3.org/2001/04/xmldsig-more#ecdsa-sha256"
	algRSA256   = "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"
)

// SignAuthTokenRequest builds the login document for a company (context NIP)
// with the identity read from the certificate's subject, and signs it. The
// result is what POST /auth/xades-signature takes.
func (s *CertSigner) SignAuthTokenRequest(challenge, nip string, signingTime time.Time) ([]byte, error) {
	return s.signEnveloped(authTokenRequest(challenge, nip), signingTime)
}

// authTokenRequest is the login document (schema auth 2.0): the challenge,
// the company it logs in for, the identity read from the certificate's subject.
func authTokenRequest(challenge, nip string) *xel {
	return el("", "AuthTokenRequest").declare("", nsAuthToken).add(
		el("", "Challenge").txt(challenge),
		el("", "ContextIdentifier").add(el("", "Nip").txt(nip)),
		el("", "SubjectIdentifierType").txt("certificateSubject"),
	)
}

// signEnveloped appends an enveloped XAdES-BES signature to root and returns
// the signed document.
func (s *CertSigner) signEnveloped(root *xel, signingTime time.Time) ([]byte, error) {
	sigAlg := ""
	switch s.Key.Public().(type) {
	case *ecdsa.PublicKey:
		sigAlg = algECDSA256
	case *rsa.PublicKey:
		sigAlg = algRSA256
	default:
		return nil, errors.New("ksef signature: the key is neither EC nor RSA")
	}
	certDigest := sha256.Sum256(s.Cert.Raw)

	// The document's digest: the document as it stands, before the
	// signature is in it — what the enveloped-signature transform leaves.
	docDigest := sha256.Sum256([]byte(root.c14n(nil, nil)))

	signedProps := el("xades", "SignedProperties").attr("Id", "SignedProperties-1").add(
		el("xades", "SignedSignatureProperties").add(
			el("xades", "SigningTime").txt(signingTime.UTC().Format("2006-01-02T15:04:05Z")),
			el("xades", "SigningCertificate").add(
				el("xades", "Cert").add(
					el("xades", "CertDigest").add(
						el("ds", "DigestMethod").attr("Algorithm", algSHA256),
						el("ds", "DigestValue").txt(base64.StdEncoding.EncodeToString(certDigest[:])),
					),
					el("xades", "IssuerSerial").add(
						el("ds", "X509IssuerName").txt(s.Cert.Issuer.String()),
						el("ds", "X509SerialNumber").txt(s.Cert.SerialNumber.String()),
					),
				),
			),
		),
	)
	qualifying := el("xades", "QualifyingProperties").declare("xades", nsXAdES).
		attr("Target", "#Signature-1").add(signedProps)

	reference := func(uri, typ string, digest []byte, transforms ...string) *xel {
		r := el("ds", "Reference").attr("URI", uri)
		if typ != "" {
			r.attr("Type", typ)
		}
		ts := el("ds", "Transforms")
		for _, t := range transforms {
			ts.add(el("ds", "Transform").attr("Algorithm", t))
		}
		return r.add(ts, el("ds", "DigestMethod").attr("Algorithm", algSHA256),
			el("ds", "DigestValue").txt(base64.StdEncoding.EncodeToString(digest)))
	}
	signedInfo := el("ds", "SignedInfo").add(
		el("ds", "CanonicalizationMethod").attr("Algorithm", algExcC14N),
		el("ds", "SignatureMethod").attr("Algorithm", sigAlg),
		reference("", "", docDigest[:], algEnvelope, algExcC14N),
	)
	signature := el("ds", "Signature").declare("ds", nsDS).attr("Id", "Signature-1")
	signatureValue := el("ds", "SignatureValue")
	signature.add(signedInfo, signatureValue,
		el("ds", "KeyInfo").add(el("ds", "X509Data").add(
			el("ds", "X509Certificate").txt(base64.StdEncoding.EncodeToString(s.Cert.Raw)))),
		el("ds", "Object").add(qualifying))
	root.add(signature)

	// The signed properties' digest, canonical in their place in the tree.
	propsDigest := sha256.Sum256([]byte(signedProps.c14n(root.scopeOf(signedProps), nil)))
	signedInfo.add(reference("#SignedProperties-1", "http://uri.etsi.org/01903#SignedProperties", propsDigest[:], algExcC14N))

	// SignedInfo, canonical in its place, signed.
	sum := sha256.Sum256([]byte(signedInfo.c14n(root.scopeOf(signedInfo), nil)))
	var value []byte
	switch key := s.Key.(type) {
	case *ecdsa.PrivateKey:
		r, ss, err := ecdsa.Sign(rand.Reader, key, sum[:])
		if err != nil {
			return nil, err
		}
		// XMLDSig 1.1: R || S, each the curve's size.
		size := (key.Curve.Params().BitSize + 7) / 8
		value = make([]byte, 2*size)
		r.FillBytes(value[:size])
		ss.FillBytes(value[size:])
	default:
		var err error
		if value, err = s.Key.Sign(rand.Reader, sum[:], crypto.SHA256); err != nil {
			return nil, err
		}
	}
	signatureValue.txt(base64.StdEncoding.EncodeToString(value))
	return []byte(root.c14n(nil, nil)), nil
}

// xel is one element of a document the signer builds: its prefix ("" for
// the default namespace), name, the namespaces declared on it, its
// attributes (none namespaced), its text or its children.
type xel struct {
	prefix, name string
	ns           [][2]string
	attrs        [][2]string
	text         string
	kids         []*xel
}

func el(prefix, name string) *xel              { return &xel{prefix: prefix, name: name} }
func (e *xel) declare(prefix, uri string) *xel { e.ns = append(e.ns, [2]string{prefix, uri}); return e }
func (e *xel) attr(name, value string) *xel {
	e.attrs = append(e.attrs, [2]string{name, value})
	return e
}
func (e *xel) txt(t string) *xel     { e.text = t; return e }
func (e *xel) add(kids ...*xel) *xel { e.kids = append(e.kids, kids...); return e }
func (e *xel) qname() string {
	if e.prefix == "" {
		return e.name
	}
	return e.prefix + ":" + e.name
}

// scopeOf is the namespaces in scope at target: every declaration on the
// path from e down to target, target's own included.
func (e *xel) scopeOf(target *xel) map[string]string {
	var walk func(n *xel, scope map[string]string) map[string]string
	walk = func(n *xel, scope map[string]string) map[string]string {
		inner := map[string]string{}
		for k, v := range scope {
			inner[k] = v
		}
		for _, d := range n.ns {
			inner[d[0]] = d[1]
		}
		if n == target {
			return scope // the parent's: target's own declarations apply when it is printed
		}
		for _, k := range n.kids {
			if found := walk(k, inner); found != nil {
				return found
			}
		}
		return nil
	}
	return walk(e, map[string]string{})
}

// c14n prints the element in exclusive canonical form: given the namespaces
// in scope above it and those already printed by output ancestors, it
// prints a namespace where it is first visibly used — the element's own
// prefix — sorted, then its attributes sorted, start and end tags always,
// text and attribute values escaped as the recommendation says.
func (e *xel) c14n(inScope, rendered map[string]string) string {
	scope := map[string]string{}
	for k, v := range inScope {
		scope[k] = v
	}
	for _, d := range e.ns {
		scope[d[0]] = d[1]
	}
	out := map[string]string{}
	for k, v := range rendered {
		out[k] = v
	}
	var b strings.Builder
	b.WriteString("<" + e.qname())
	if uri, used := scope[e.prefix]; used && out[e.prefix] != uri {
		if e.prefix == "" {
			b.WriteString(` xmlns="` + escapeAttr(uri) + `"`)
		} else {
			b.WriteString(` xmlns:` + e.prefix + `="` + escapeAttr(uri) + `"`)
		}
		out[e.prefix] = uri
	}
	attrs := append([][2]string(nil), e.attrs...)
	sort.Slice(attrs, func(i, j int) bool { return attrs[i][0] < attrs[j][0] })
	for _, a := range attrs {
		b.WriteString(" " + a[0] + `="` + escapeAttr(a[1]) + `"`)
	}
	b.WriteString(">")
	b.WriteString(escapeText(e.text))
	for _, k := range e.kids {
		b.WriteString(k.c14n(scope, out))
	}
	b.WriteString("</" + e.qname() + ">")
	return b.String()
}

func escapeText(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\r", "&#xD;").Replace(s)
}

func escapeAttr(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", `"`, "&quot;", "\t", "&#x9;", "\n", "&#xA;", "\r", "&#xD;").Replace(s)
}
