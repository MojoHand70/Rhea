package pl

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// KSeF's own certificates (the Poland pack review, KSEF-21 and KSEF-05): a
// login certificate ("Authentication") replaces the token from 2027, and an
// Offline certificate signs the second QR code of an invoice issued while
// KSeF is out of reach. KSeF issues them to whoever logged in with a
// signature — a qualified certificate, a trusted profile, or an earlier
// KSeF certificate; never after a token login (regulation Dz.U. 2025 poz.
// 1815, § 13 ust. 2). The request is made "in one's own name": KSeF says
// which identity the certificate will carry, and Rhea's request must carry
// exactly that. The key is generated here and never leaves: what KSeF
// receives is a signing request with the public half.

// certNamePattern is what KSeF admits as a certificate's name (its answer
// 21405 otherwise).
var certNamePattern = regexp.MustCompile(`^[A-Za-z0-9ąćęłńóśźżĄĆĘŁŃÓŚŹŻ _-]+$`)

// Certificate kinds KSeF issues.
const (
	CertAuthentication = "Authentication"
	CertOffline        = "Offline"
)

// enrollmentData is the identity KSeF reads from the login's signature.
type enrollmentData struct {
	CommonName             string `json:"commonName"`
	CountryName            string `json:"countryName"`
	GivenName              string `json:"givenName"`
	Surname                string `json:"surname"`
	SerialNumber           string `json:"serialNumber"`
	UniqueIdentifier       string `json:"uniqueIdentifier"`
	OrganizationName       string `json:"organizationName"`
	OrganizationIdentifier string `json:"organizationIdentifier"`
}

var (
	oidGivenName        = asn1.ObjectIdentifier{2, 5, 4, 42}
	oidSurname          = asn1.ObjectIdentifier{2, 5, 4, 4}
	oidUniqueIdentifier = asn1.ObjectIdentifier{2, 5, 4, 45}
)

// subject is the distinguished name the request must carry, attribute for
// attribute as KSeF returned it.
func (d enrollmentData) subject() pkix.Name {
	n := pkix.Name{CommonName: d.CommonName, Country: []string{d.CountryName}, SerialNumber: d.SerialNumber}
	if d.OrganizationName != "" {
		n.Organization = []string{d.OrganizationName}
	}
	extra := func(oid asn1.ObjectIdentifier, v string) {
		if v != "" {
			n.ExtraNames = append(n.ExtraNames, pkix.AttributeTypeAndValue{Type: oid, Value: v})
		}
	}
	extra(oidGivenName, d.GivenName)
	extra(oidSurname, d.Surname)
	extra(oidUniqueIdentifier, d.UniqueIdentifier)
	extra(oidOrganizationIdentifier, d.OrganizationIdentifier)
	return n
}

// EnrollCertificate asks KSeF for a certificate of the given kind
// (CertAuthentication or CertOffline) under a name the owner will recognise
// in KSeF's list, with the access token of a signed login. It returns the
// issued certificate with the key generated for it.
func (c *KSeF2) EnrollCertificate(ctx context.Context, access, kind, name string) (*CertSigner, error) {
	ctx, cancel := context.WithTimeout(ctx, ksefDeadline)
	defer cancel()
	if kind != CertAuthentication && kind != CertOffline {
		return nil, fmt.Errorf("ksef certificate: the kind is %s or %s, not %q", CertAuthentication, CertOffline, kind)
	}
	if !certNamePattern.MatchString(name) {
		return nil, fmt.Errorf("ksef certificate: the name %q may hold only letters (Polish ones too), digits, spaces, hyphens and underscores", name)
	}
	var limits struct {
		CanRequest bool `json:"canRequest"`
	}
	if err := c.call(ctx, "GET", "/certificates/limits", access, nil, &limits); err != nil {
		return nil, err
	}
	if !limits.CanRequest {
		return nil, errors.New("ksef certificate: KSeF says no further certificate can be requested now (the limit is reached)")
	}
	var data enrollmentData
	if err := c.call(ctx, "GET", "/certificates/enrollments/data", access, nil, &data); err != nil {
		return nil, fmt.Errorf("%w — a certificate is asked for only after a signed login, never after a token", err)
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	csr, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{
		Subject: data.subject(), SignatureAlgorithm: x509.ECDSAWithSHA256,
	}, key)
	if err != nil {
		return nil, err
	}
	var enrolled struct {
		ReferenceNumber string `json:"referenceNumber"`
	}
	if err := c.call(ctx, "POST", "/certificates/enrollments", access, map[string]any{
		"certificateName": name, "certificateType": kind, "csr": base64.StdEncoding.EncodeToString(csr),
	}, &enrolled); err != nil {
		return nil, err
	}
	c.logf("certificate request %s (%s) sent", enrolled.ReferenceNumber, kind)
	serial := ""
	for wait := pollFirst; serial == ""; wait = min(wait*2, pollMax) {
		var st struct {
			Status                  statusInfo `json:"status"`
			CertificateSerialNumber string     `json:"certificateSerialNumber"`
		}
		if err := c.call(ctx, "GET", "/certificates/enrollments/"+enrolled.ReferenceNumber, access, nil, &st); err != nil {
			return nil, err
		}
		switch {
		case st.Status.Code >= 400:
			return nil, fmt.Errorf("ksef certificate refused: %s", st.Status)
		case st.CertificateSerialNumber != "":
			serial = st.CertificateSerialNumber
			continue
		}
		if err := c.sleep(ctx, wait); err != nil {
			return nil, err
		}
	}
	var got struct {
		Certificates []struct {
			Certificate             string `json:"certificate"`
			CertificateSerialNumber string `json:"certificateSerialNumber"`
			CertificateType         string `json:"certificateType"`
		} `json:"certificates"`
	}
	if err := c.call(ctx, "POST", "/certificates/retrieve", access, map[string]any{
		"certificateSerialNumbers": []string{serial},
	}, &got); err != nil {
		return nil, err
	}
	for _, item := range got.Certificates {
		if !strings.EqualFold(item.CertificateSerialNumber, serial) {
			continue
		}
		der, err := base64.StdEncoding.DecodeString(item.Certificate)
		if err != nil {
			return nil, fmt.Errorf("ksef certificate %s: %w", serial, err)
		}
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			return nil, fmt.Errorf("ksef certificate %s: %w", serial, err)
		}
		if !samePublic(&key.PublicKey, cert.PublicKey) {
			return nil, fmt.Errorf("ksef certificate %s: it is not for the key that asked for it", serial)
		}
		c.logf("certificate %s issued: %s, valid %s – %s", serial, item.CertificateType,
			cert.NotBefore.Format("2006-01-02"), cert.NotAfter.Format("2006-01-02"))
		return &CertSigner{Cert: cert, Key: key}, nil
	}
	return nil, fmt.Errorf("ksef certificate %s: issued but not returned", serial)
}
