package pl

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"rhea/internal/adapter"
	"rhea/internal/core"
)

// KSeFClient is the boundary to the national e-invoice system: one
// interactive session for the invoices given, each one's fate reported.
// KSeF2 is the real API; FakeKSeF answers deterministically for demos,
// tests and replays.
type KSeFClient interface {
	Submit(ctx context.Context, invoices []Outbound) ([]Submission, error)
}

// KSeF is the Poland pack's statutory adapter, documents out (DIRECTION
// 2026-10-15): every sales_invoice without an accepted ksef_submission is
// written as FA(3) and sent; the evidence comes back as events keyed by
// the invoice's object id, so a retry after a crash re-submits into a
// dedup no-op. Acceptance is ksef.invoice.submitted with the KSeF number
// and the UPO; a rejection is ksef.invoice.rejected with the Ministry's
// code and words, and the same XML is not sent again — a person changes
// something first.
type KSeF struct {
	Client     KSeFClient
	Today      func() string // business date of submissions; defaults to the wall clock
	SystemInfo string        // what the invoice says wrote it; defaults to "Rhea"
}

// ksefCodeVersion is what a channel declaration binds to; it moves when the
// adapter's reading of the outside world changes (a new API, a new schema).
// 2: KSeF API 2.0, interactive session, FA(3), rejections as evidence.
const ksefCodeVersion = 2

func init() {
	adapter.Register(adapter.Implemented{Name: "ksef", CodeVersion: ksefCodeVersion, Format: "KSeF API 2.0: one interactive session per pass, FA(3), acceptance and rejection as evidence"})
	adapter.Provide("ksef", func() adapter.Adapter { return KSeF{Client: ClientFromEnv()} })
}

// ClientFromEnv is the national API when the channel is configured —
// RHEA_KSEF_ENV (test | demo | prod), RHEA_KSEF_NIP, and the way in: a
// certificate (RHEA_KSEF_CERT and RHEA_KSEF_KEY, PEM files) or, until tokens
// end on 31 December 2026, RHEA_KSEF_TOKEN; the secrets stay outside the log —
// and the deterministic fake otherwise. A certificate wins over a token. A
// misnamed environment, half a certificate or an unreadable one is a client
// that refuses to send, not a guess.
func ClientFromEnv() KSeFClient {
	token := os.Getenv("RHEA_KSEF_TOKEN")
	certPath, keyPath := os.Getenv("RHEA_KSEF_CERT"), os.Getenv("RHEA_KSEF_KEY")
	if token == "" && certPath == "" && keyPath == "" {
		return FakeKSeF{}
	}
	env := os.Getenv("RHEA_KSEF_ENV")
	base, ok := KSeFEnvironments[env]
	if !ok {
		return refusingClient{fmt.Errorf("RHEA_KSEF_ENV must be test, demo or prod (got %q)", env)}
	}
	var signer *CertSigner
	if certPath != "" || keyPath != "" {
		if certPath == "" || keyPath == "" {
			return refusingClient{fmt.Errorf("a KSeF certificate needs both RHEA_KSEF_CERT and RHEA_KSEF_KEY")}
		}
		var err error
		if signer, err = LoadCertSigner(certPath, keyPath); err != nil {
			return refusingClient{err}
		}
		token = ""
	}
	return &KSeF2{BaseURL: base, NIP: os.Getenv("RHEA_KSEF_NIP"), Token: token, Signer: signer,
		Log: func(format string, args ...any) {
			log.Printf("ksef (%s, NIP %s): "+format, append([]any{env, os.Getenv("RHEA_KSEF_NIP")}, args...)...)
		}}
}

type refusingClient struct{ err error }

func (r refusingClient) Submit(context.Context, []Outbound) ([]Submission, error) { return nil, r.err }

func (k KSeF) Name() string     { return "ksef" }
func (k KSeF) CodeVersion() int { return ksefCodeVersion }

// EventKSeFRejected is the evidence of a rejection, beside the pack's
// ksef.invoice.submitted.
const (
	EventKSeFSubmitted = "ksef.invoice.submitted"
	EventKSeFRejected  = "ksef.invoice.rejected"
)

func (k KSeF) Pass(ctx context.Context, reads adapter.Reads) ([]core.Event, error) {
	invoices, err := reads.ObjectsByType(ctx, "sales_invoice")
	if err != nil {
		return nil, err
	}
	subs, err := reads.ObjectsByType(ctx, "ksef_submission")
	if err != nil {
		return nil, err
	}
	companies, err := byID(ctx, reads, "company")
	if err != nil {
		return nil, err
	}
	rates, err := byID(ctx, reads, "vat_rate")
	if err != nil {
		return nil, err
	}
	accepted := map[string]bool{}
	rejected := map[string]string{} // invoice id -> hash of the XML KSeF refused
	for _, s := range subs {
		id, _ := s.State["invoice"].(string)
		status, _ := s.State["status"].(string)
		ref, _ := s.State["ksef_ref"].(string)
		switch {
		case status == "rejected":
			rejected[id], _ = s.State["invoice_hash"].(string)
		case status == "accepted" || (status == "" && ref != ""):
			accepted[id] = true
		}
	}
	today := time.Now().Format("2006-01-02")
	if k.Today != nil {
		today = k.Today()
	}
	var pending []Outbound
	var problems []string
	for _, inv := range invoices {
		if accepted[inv.ID] {
			continue
		}
		fa, err := k.invoiceFor(inv, companies, rates, today)
		if err != nil {
			problems = append(problems, err.Error())
			continue
		}
		body, err := WriteFA3(fa)
		if err != nil {
			problems = append(problems, err.Error())
			continue
		}
		sum := sha256.Sum256(body)
		if rejected[inv.ID] == base64.StdEncoding.EncodeToString(sum[:]) {
			continue // KSeF refused exactly this; sending it again changes nothing
		}
		pending = append(pending, Outbound{ID: inv.ID, Number: fa.Number, XML: body})
	}
	var out []core.Event
	if len(pending) > 0 {
		results, err := k.Client.Submit(ctx, pending)
		if err != nil {
			// Nothing of a failed session is evidence; the next pass retries.
			return nil, fmt.Errorf("ksef: %w", err)
		}
		sent := map[string][]byte{}
		for _, o := range pending {
			sent[o.ID] = o.XML
		}
		for _, r := range results {
			ev, err := k.evidence(r, sent[r.ID], today)
			if err != nil {
				problems = append(problems, err.Error())
				continue
			}
			out = append(out, ev)
		}
	}
	if len(problems) > 0 {
		return out, errors.New("ksef: " + strings.Join(problems, "; "))
	}
	return out, nil
}

// evidence turns one submission into the event the log is owed. An
// acceptance carries the document sent, the KSeF number, the UPO and the
// two times the Ministry stamped — received, and the number assigned; the
// first is the day a structured invoice counts as issued (VAT act, art.
// 106na). The UPO's own fingerprint of the document must be the fingerprint
// of what was sent, or the evidence is refused as inconsistent: nothing is
// booked, and the pass error names it for a person.
func (k KSeF) evidence(r Submission, sent []byte, today string) (core.Event, error) {
	if r.Accepted {
		upo, err := ParseUPO(r.UPO)
		if err != nil {
			return core.Event{}, fmt.Errorf("invoice %s: the UPO cannot be read: %w", r.Number, err)
		}
		if upo.DocumentHash != r.InvoiceHash {
			return core.Event{}, fmt.Errorf("invoice %s: the UPO acknowledges a different document (its fingerprint %s, ours %s) — evidence refused", r.Number, upo.DocumentHash, r.InvoiceHash)
		}
		if upo.KSeFNumber != "" && upo.KSeFNumber != r.KSeFNumber {
			return core.Event{}, fmt.Errorf("invoice %s: the UPO names KSeF number %s, the session said %s — evidence refused", r.Number, upo.KSeFNumber, r.KSeFNumber)
		}
		payload, err := json.Marshal(map[string]any{
			"number": r.Number, "ksef_ref": r.KSeFNumber, "upo": r.Session, "upo_xml": r.UPO,
			"invoice_xml": string(sent), "invoice_hash": r.InvoiceHash,
			"sent_at": upo.ReceivedAt, "accepted_at": upo.NumberedAt, "date": today,
		})
		if err != nil {
			return core.Event{}, err
		}
		return core.Event{Type: EventKSeFSubmitted, OccurredAt: today, Payload: payload,
			DedupKey: "ksef/submit/" + r.ID}, nil
	}
	payload, err := json.Marshal(map[string]any{
		"number": r.Number, "code": r.Code, "reason": r.Reason(), "details": r.Details,
		"date": today, "invoice_hash": r.InvoiceHash,
	})
	if err != nil {
		return core.Event{}, err
	}
	short := r.InvoiceHash
	if len(short) > 16 {
		short = short[:16]
	}
	return core.Event{Type: EventKSeFRejected, OccurredAt: today, Payload: payload,
		DedupKey: "ksef/reject/" + r.ID + "/" + short}, nil
}

// invoiceFor reads one sales invoice and its master data into the FA(3)
// shape: the parties by their company objects, the rate by its object,
// the amounts as state holds them.
func (k KSeF) invoiceFor(inv core.Object, companies, rates map[string]map[string]any, today string) (FA3Invoice, error) {
	number, _ := inv.State["number"].(string)
	if number == "" {
		return FA3Invoice{}, fmt.Errorf("invoice %s has no number", inv.ID)
	}
	net, okN := minorUnits(inv.State["net"])
	vat, okV := minorUnits(inv.State["vat"])
	gross, okG := minorUnits(inv.State["gross"])
	if !okN || !okV || !okG {
		return FA3Invoice{}, fmt.Errorf("invoice %s has no usable amounts", number)
	}
	currency, _ := inv.State["currency"].(string)
	if currency == "" {
		currency = "PLN"
	}
	issueDate, _ := inv.State["issue_date"].(string)
	rateID, _ := inv.State["vat_rate"].(string)
	rate, ok := rates[rateID]
	if !ok {
		return FA3Invoice{}, fmt.Errorf("invoice %s names no VAT rate the pack knows", number)
	}
	code, _ := rate["code"].(string)
	party := func(role, nipField string) FA3Party {
		id, _ := inv.State[role].(string)
		c := companies[id]
		name, _ := c["name"].(string)
		vatID, _ := c["vat_id"].(string)
		if vatID == "" {
			vatID, _ = inv.State[nipField].(string)
		}
		country, _ := c["country"].(string)
		address, _ := c["address"].(string)
		return FA3Party{NIP: NIPDigits(vatID), Name: name, Country: country, Address: address}
	}
	seller, buyer := party("seller", "seller_nip"), party("buyer", "buyer_nip")
	account, _ := companies[sellerID(inv)]["iban"].(string)
	system := k.SystemInfo
	if system == "" {
		system = "Rhea"
	}
	saleDate, _ := inv.State["sale_date"].(string)
	vatPLN, _ := minorUnits(inv.State["vat_pln"])
	basis, _ := companies[sellerID(inv)]["vat_exemption_basis"].(string)
	return FA3Invoice{
		Seller: seller, Buyer: buyer, Currency: currency, IssueDate: issueDate, Number: number,
		SaleDate: saleDate, RateCode: code, ExemptionBasis: basis,
		Net: net, VAT: vat, Gross: gross, VATPLN: vatPLN, Account: account,
		Created: today + "T08:00:00Z", SystemInfo: system,
	}, nil
}

func sellerID(inv core.Object) string {
	id, _ := inv.State["seller"].(string)
	return id
}

// byID indexes one type's objects' state by object id.
func byID(ctx context.Context, reads adapter.Reads, typ string) (map[string]map[string]any, error) {
	objs, err := reads.ObjectsByType(ctx, typ)
	if err != nil {
		return nil, err
	}
	out := make(map[string]map[string]any, len(objs))
	for _, o := range objs {
		out[o.ID] = o.State
	}
	return out, nil
}

// minorUnits reads a money value out of projected object state, where JSONB
// hands numbers back as float64.
func minorUnits(v any) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	}
	return 0, false
}

// UPO is what the official receipt says: the session, the KSeF number, the
// invoice number, the document's fingerprint (base64 SHA-256, the same
// encoding Rhea computes before sending), and the Ministry's two times.
type UPO struct {
	Session      string `xml:"NumerReferencyjnySesji"`
	KSeFNumber   string `xml:"Dokument>NumerKSeFDokumentu"`
	Number       string `xml:"Dokument>NumerFaktury"`
	DocumentHash string `xml:"Dokument>SkrotDokumentu"`
	ReceivedAt   string `xml:"Dokument>DataPrzeslaniaDokumentu"`
	NumberedAt   string `xml:"Dokument>DataNadaniaNumeruKSeF"`
}

// ParseUPO reads the receipt XML as the Ministry returns it (namespace
// http://upo.schematy.mf.gov.pl/KSeF/v4-3, root Potwierdzenie).
func ParseUPO(text string) (UPO, error) {
	var u UPO
	if err := xml.Unmarshal([]byte(text), &u); err != nil {
		return UPO{}, err
	}
	if u.DocumentHash == "" {
		return UPO{}, errors.New("no document fingerprint in the UPO")
	}
	return u, nil
}

// FakeKSeF answers like the national system would, deterministically from
// the invoice — demos and replays stay reproducible. Reject names invoice
// numbers the fake refuses, with the reason, so a test can see a
// rejection without the Ministry.
type FakeKSeF struct {
	Reject map[string]string
}

func (f FakeKSeF) Submit(_ context.Context, invoices []Outbound) ([]Submission, error) {
	out := make([]Submission, 0, len(invoices))
	for _, inv := range invoices {
		var doc struct {
			Seller struct {
				NIP string `xml:"DaneIdentyfikacyjne>NIP"`
			} `xml:"Podmiot1"`
			Fa struct {
				Date   string `xml:"P_1"`
				Number string `xml:"P_2"`
			} `xml:"Fa"`
		}
		if err := xml.Unmarshal(inv.XML, &doc); err != nil {
			return out, fmt.Errorf("fake ksef: %s: %w", inv.Number, err)
		}
		sum := sha256.Sum256(inv.XML)
		sub := Submission{ID: inv.ID, Number: inv.Number, Session: "FAKE-SESSION",
			Reference: "FAKE-" + inv.ID, InvoiceHash: base64.StdEncoding.EncodeToString(sum[:])}
		if reason, refused := f.Reject[inv.Number]; refused {
			sub.Code, sub.Description = 450, reason
		} else {
			sub.Accepted = true
			sub.Code, sub.Description = 200, "Sukces"
			sub.KSeFNumber = KSeFReference(doc.Seller.NIP, doc.Fa.Date, doc.Fa.Number)
			sub.UPO = fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?><Potwierdzenie xmlns="http://upo.schematy.mf.gov.pl/KSeF/v4-3"><NazwaPodmiotuPrzyjmujacego>Rhea (fake KSeF)</NazwaPodmiotuPrzyjmujacego><NumerReferencyjnySesji>%s</NumerReferencyjnySesji><Dokument><NipSprzedawcy>%s</NipSprzedawcy><NumerKSeFDokumentu>%s</NumerKSeFDokumentu><NumerFaktury>%s</NumerFaktury><DataWystawieniaFaktury>%s</DataWystawieniaFaktury><DataPrzeslaniaDokumentu>%sT08:00:00+02:00</DataPrzeslaniaDokumentu><DataNadaniaNumeruKSeF>%sT08:00:01+02:00</DataNadaniaNumeruKSeF><SkrotDokumentu>%s</SkrotDokumentu><TrybWysylki>Online</TrybWysylki></Dokument></Potwierdzenie>`,
				sub.Session, doc.Seller.NIP, sub.KSeFNumber, doc.Fa.Number, doc.Fa.Date, doc.Fa.Date, doc.Fa.Date, sub.InvoiceHash)
		}
		out = append(out, sub)
	}
	return out, nil
}
