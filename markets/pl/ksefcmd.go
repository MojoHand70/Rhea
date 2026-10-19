package pl

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"rhea/internal/adapter"
)

// The certificate tool (rhea run ksef-certificate …): what the owner does
// once, and again before a certificate ends, to keep Rhea's way into KSeF.
// It writes files the owner keeps — a certificate and its key, never
// overwriting one — and nothing to the log.
func init() {
	adapter.ProvideCommand(adapter.Command{
		Name:    "ksef-certificate",
		Summary: "KSeF certificates: test-seal | login-file | request (rhea run ksef-certificate for help)",
		Run:     runKSeFCertificate,
	})
}

const ksefCertificateHelp = `rhea run ksef-certificate SUBCOMMAND

  test-seal  -nip NIP -name NAME -out DIR
      a company seal made by Rhea for a made-up NIP: the Ministry's TEST
      environment accepts it for login, DEMO and production do not

  login-file -out FILE
      the login document for RHEA_KSEF_NIP with a fresh challenge, unsigned:
      sign it within ten minutes (a trusted profile on gov.pl, a qualified
      signature or seal) and hand it to request -signed

  request    -kind login|offline -name NAME -out DIR [-signed FILE]
      asks KSeF (RHEA_KSEF_ENV) for its own certificate for RHEA_KSEF_NIP: a
      login certificate, which replaces the token, or an Offline one, which
      signs the second QR code of an invoice issued while KSeF is down. The
      login that asks must be signed: the signed file from login-file, or
      the certificate in RHEA_KSEF_CERT and RHEA_KSEF_KEY (on TEST, a
      test-seal). Writes the certificate and the key Rhea generated for it.`

func runKSeFCertificate(ctx context.Context, args []string, out io.Writer) error {
	if len(args) == 0 {
		fmt.Fprintln(out, ksefCertificateHelp)
		return nil
	}
	fs := flag.NewFlagSet("ksef-certificate "+args[0], flag.ContinueOnError)
	fs.SetOutput(out)
	switch args[0] {
	case "test-seal":
		nip := fs.String("nip", "", "the made-up NIP")
		name := fs.String("name", "Rhea Test sp. z o.o.", "the organisation's name")
		dir := fs.String("out", ".", "where to write the seal")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *nip == "" {
			return errors.New("test-seal needs -nip")
		}
		seal, err := SelfSignedSeal(*nip, *name, time.Now())
		if err != nil {
			return err
		}
		certPath, keyPath, err := writeCertSigner(*dir, "test-seal-"+*nip, seal)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "a seal for NIP %s, for the TEST environment only:\n  RHEA_KSEF_CERT=%s\n  RHEA_KSEF_KEY=%s\n", *nip, certPath, keyPath)
		return nil

	case "login-file":
		path := fs.String("out", "", "where to write the unsigned login document")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *path == "" {
			return errors.New("login-file needs -out")
		}
		client, err := ksef2FromEnv()
		if err != nil {
			return err
		}
		doc, err := client.LoginRequest(ctx)
		if err != nil {
			return err
		}
		if err := writeNew(*path, doc, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(out, "%s: the login document for NIP %s. Sign it now — the challenge inside it lasts ten minutes — and run\n  rhea run ksef-certificate request -signed SIGNED.xml -kind login -name NAME -out DIR\n", *path, client.NIP)
		return nil

	case "request":
		kind := fs.String("kind", "login", "login or offline")
		name := fs.String("name", "", "the certificate's name in KSeF's list (letters, digits, spaces, - and _)")
		dir := fs.String("out", ".", "where to write the certificate and its key")
		signedPath := fs.String("signed", "", "a login document signed outside Rhea (from login-file)")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		certKind := map[string]string{"login": CertAuthentication, "offline": CertOffline}[*kind]
		if certKind == "" {
			return fmt.Errorf("-kind is login or offline, not %q", *kind)
		}
		if *name == "" {
			return errors.New("request needs -name: how the certificate will be called in KSeF's list")
		}
		client, err := ksef2FromEnv()
		if err != nil {
			return err
		}
		client.Log = func(format string, a ...any) { fmt.Fprintf(out, "  "+format+"\n", a...) }
		var access string
		if *signedPath != "" {
			signed, err := os.ReadFile(*signedPath)
			if err != nil {
				return err
			}
			access, err = client.LoginSigned(ctx, signed)
			if err != nil {
				return err
			}
		} else {
			if client.Signer == nil {
				return errors.New("a certificate is asked for after a signed login: give -signed FILE, or RHEA_KSEF_CERT and RHEA_KSEF_KEY — a token cannot ask")
			}
			if access, err = client.Login(ctx); err != nil {
				return err
			}
		}
		issued, err := client.EnrollCertificate(ctx, access, certKind, *name)
		if err != nil {
			return err
		}
		stem := fmt.Sprintf("ksef-%s-%s-%016X", *kind, client.NIP, issued.Cert.SerialNumber)
		certPath, keyPath, err := writeCertSigner(*dir, stem, issued)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "KSeF issued a %s certificate for NIP %s, valid until %s:\n  certificate %s\n  key         %s (keep it secret; Rhea made it and KSeF never saw it)\n",
			*kind, client.NIP, issued.Cert.NotAfter.Format("2006-01-02"), certPath, keyPath)
		if certKind == CertAuthentication {
			fmt.Fprintf(out, "To log in with it:\n  RHEA_KSEF_CERT=%s\n  RHEA_KSEF_KEY=%s\n", certPath, keyPath)
		}
		return nil
	}
	return fmt.Errorf("unknown subcommand %q\n%s", args[0], ksefCertificateHelp)
}

// writeCertSigner writes <stem>.cert.pem and <stem>.key.pem (the key readable
// by its owner only), refusing to overwrite either.
func writeCertSigner(dir, stem string, s *CertSigner) (certPath, keyPath string, err error) {
	certPEM, keyPEM, err := s.PEM()
	if err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", "", err
	}
	certPath, keyPath = filepath.Join(dir, stem+".cert.pem"), filepath.Join(dir, stem+".key.pem")
	if err := writeNew(keyPath, keyPEM, 0o600); err != nil {
		return "", "", err
	}
	if err := writeNew(certPath, certPEM, 0o644); err != nil {
		return "", "", err
	}
	return certPath, keyPath, nil
}

// writeNew writes a file that must not exist yet.
func writeNew(path string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("%s exists — Rhea does not overwrite a certificate or a key", path)
		}
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// ksef2FromEnv is the national API as the environment configures it — no
// fake, no fallback: RHEA_KSEF_ENV and RHEA_KSEF_NIP, and a certificate
// (RHEA_KSEF_CERT, RHEA_KSEF_KEY), a token (RHEA_KSEF_TOKEN), or neither
// (enough to fetch a challenge or finish a login signed elsewhere).
func ksef2FromEnv() (*KSeF2, error) {
	env := os.Getenv("RHEA_KSEF_ENV")
	base, ok := KSeFEnvironments[env]
	if !ok {
		return nil, fmt.Errorf("RHEA_KSEF_ENV must be test, demo or prod (got %q)", env)
	}
	nip := strings.TrimSpace(os.Getenv("RHEA_KSEF_NIP"))
	if nip == "" {
		return nil, errors.New("RHEA_KSEF_NIP names the company")
	}
	c := &KSeF2{BaseURL: base, NIP: nip, Token: os.Getenv("RHEA_KSEF_TOKEN")}
	certPath, keyPath := os.Getenv("RHEA_KSEF_CERT"), os.Getenv("RHEA_KSEF_KEY")
	if certPath != "" || keyPath != "" {
		if certPath == "" || keyPath == "" {
			return nil, errors.New("a KSeF certificate needs both RHEA_KSEF_CERT and RHEA_KSEF_KEY")
		}
		signer, err := LoadCertSigner(certPath, keyPath)
		if err != nil {
			return nil, err
		}
		c.Signer, c.Token = signer, ""
	}
	return c, nil
}
