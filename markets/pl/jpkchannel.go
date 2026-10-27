package pl

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"rhea/internal/adapter"
	"rhea/internal/core"
)

// The JPK channel (KK, 2026-10-27: the JPK goes out from Rhea, signed by a
// person): the Poland pack's statutory adapter for the Ministry's gateway.
// Each pass, per company, does what the log still owes the gateway:
//
//   - every JPK_V7M handed over (an issued_document of the writer jpk_v7m)
//     without a filing naming its hash is sealed for the gateway this server sends to — the
//     package and the request a person signs, kept as the filing's files
//     (jpk.filing.prepared);
//   - every filing a person has signed (sign_jpk_filing, the signed file
//     uploaded) is checked against its request and sent
//     (jpk.filing.sent, or jpk.filing.rejected with the reason);
//   - every filing sent is asked for its status, and a final one is booked:
//     filed with its UPO (jpk.filing.accepted), or the gateway's code and
//     words (jpk.filing.rejected).
//
// The signature is the person's, made outside Rhea; Rhea never holds a key.
// Every fact is keyed by what it is — the issued file, the filing and the
// signed file sent, the filing and the reference asked — so a pass run twice
// appends nothing twice, and one pass of the channel at a time holds across
// processes. The gateway itself refuses a second copy of a processed
// document by its hash (170) and names the original, which the filing then
// follows.

const jpkChannelVersion = 1

// The channel's facts.
const (
	EventJPKPrepared = "jpk.filing.prepared"
	EventJPKSigned   = "jpk.filing.signed"
	EventJPKSent     = "jpk.filing.sent"
	EventJPKAccepted = "jpk.filing.accepted"
	EventJPKRejected = "jpk.filing.rejected"
)

func init() {
	adapter.Register(adapter.Implemented{Name: "jpk_gateway", CodeVersion: jpkChannelVersion,
		Format: "the Ministry's gateway e-dokumenty.mf.gov.pl (interface v5.6.0): a JPK sealed, the request signed by a person, sent, its UPO"})
	adapter.Provide("jpk_gateway", func() adapter.Adapter { return JPKChannel{Gateway: JPKGatewayFromEnv()} })
}

// JPKChannel is the adapter: the gateway it speaks to and the business day.
type JPKChannel struct {
	Gateway JPKGatewayClient
	Today   func() string // defaults to the wall clock's day
}

func (JPKChannel) Name() string     { return "jpk_gateway" }
func (JPKChannel) CodeVersion() int { return jpkChannelVersion }

// ActsOn: the files a company handed over and its filings are its own.
func (JPKChannel) ActsOn() []string { return []string{"issued_document", "jpk_filing"} }

func (c JPKChannel) Pass(ctx context.Context, reads adapter.Reads) ([]core.Event, error) {
	issued, err := reads.ObjectsByType(ctx, "issued_document")
	if err != nil {
		return nil, err
	}
	filings, err := reads.ObjectsByType(ctx, "jpk_filing")
	if err != nil {
		return nil, err
	}
	today := core.Today()
	if c.Today != nil {
		today = c.Today()
	}
	// A filing names the handed-over file by its hash: the content it seals.
	filed := map[string]bool{}
	for _, f := range filings {
		filed[text(f.State["document"])] = true
	}
	var events []core.Event
	var problems []error
	for _, d := range issued {
		if text(d.State["writer"]) != "jpk_v7m" || filed[text(d.State["hash"])] {
			continue
		}
		ev, err := c.prepare(ctx, reads, d, today)
		if err != nil {
			problems = append(problems, err)
			continue
		}
		events = append(events, *ev)
	}
	for _, f := range filings {
		var ev *core.Event
		var err error
		switch text(f.State["status"]) {
		case "signed":
			ev, err = c.send(ctx, reads, f, today)
		case "sent":
			ev, err = c.status(ctx, f, today)
		}
		if err != nil {
			problems = append(problems, err)
		}
		if ev != nil {
			events = append(events, *ev)
		}
	}
	return events, errors.Join(problems...)
}

// prepare seals one handed-over JPK for the gateway this server sends to.
func (c JPKChannel) prepare(ctx context.Context, reads adapter.Reads, d core.Object, today string) (*core.Event, error) {
	cert, err := c.Gateway.Certificate()
	if err != nil {
		return nil, err
	}
	name := text(d.State["name"])
	b, err := reads.GetBlob(ctx, text(d.State["hash"]))
	if err != nil {
		return nil, fmt.Errorf("jpk gateway: %s: %w", name, err)
	}
	p, err := PackJPK(name, b.Bytes, cert)
	if err != nil {
		return nil, err
	}
	hash := text(d.State["hash"])
	return jpkFact(EventJPKPrepared, today, "prepared/"+hash,
		map[string]any{"document": hash, "name": name, "period": paramOf(text(d.State["params"]), "period"),
			"gateway": c.Gateway.Environment(), "date": today},
		core.FileInput{Input: "package", Name: p.PartName, MediaType: "application/octet-stream", Bytes: p.Part},
		core.FileInput{Input: "request", Name: strings.TrimSuffix(name, ".xml") + "_do_podpisu.xml", MediaType: "application/xml", Bytes: p.Request})
}

// send checks a signed filing and sends it.
func (c JPKChannel) send(ctx context.Context, reads adapter.Reads, f core.Object, today string) (*core.Event, error) {
	if _, err := c.Gateway.Certificate(); err != nil {
		return nil, err // no gateway named: the filing waits, it is not refused
	}
	signedHash := text(f.State["signed"])
	key := f.ID + "/" + signedHash
	reject := func(code int, reason string) (*core.Event, error) {
		return jpkFact(EventJPKRejected, today, "rejected/"+key,
			map[string]any{"filing": f.ID, "code": code, "reason": reason, "signed": signedHash, "date": today})
	}
	if env := text(f.State["gateway"]); env != c.Gateway.Environment() {
		return reject(0, fmt.Sprintf("Wysyłka przygotowana dla bramki %s, a Rhea wysyła teraz do bramki %s — przekaż plik JPK ponownie.", env, c.Gateway.Environment()))
	}
	blob := func(field string) ([]byte, error) {
		b, err := reads.GetBlob(ctx, text(f.State[field]))
		if err != nil {
			return nil, fmt.Errorf("jpk gateway: filing %s, %s: %w", f.ID, field, err)
		}
		return b.Bytes, nil
	}
	request, err := blob("request")
	if err != nil {
		return nil, err
	}
	signed, err := blob("signed")
	if err != nil {
		return nil, err
	}
	part, err := blob("package")
	if err != nil {
		return nil, err
	}
	if err := CheckSignedJPKRequest(request, signed); err != nil {
		return reject(0, signedReason(err))
	}
	ref, err := c.Gateway.Send(ctx, signed, part)
	var refusal *JPKRefusal
	switch {
	case errors.As(err, &refusal) && refusal.Original != "":
		ref = refusal.Original // the document was filed before: follow that one
	case errors.As(err, &refusal):
		return reject(refusal.Code, fmt.Sprintf("Bramka odmówiła (%d): %s", refusal.Code, refusal.Message))
	case err != nil:
		return nil, err // no refusal: the next pass sends again
	}
	return jpkFact(EventJPKSent, today, "sent/"+key,
		map[string]any{"filing": f.ID, "reference": ref, "signed": signedHash, "date": today})
}

// status asks what became of a filing sent, and books a final answer.
func (c JPKChannel) status(ctx context.Context, f core.Object, today string) (*core.Event, error) {
	ref := text(f.State["reference"])
	st, err := c.Gateway.Status(ctx, ref)
	if err != nil {
		return nil, err
	}
	if !st.Final() {
		return nil, nil
	}
	key := f.ID + "/" + ref
	if !st.Accepted() {
		reason := fmt.Sprintf("Bramka (%d): %s", st.Code, st.Description)
		if st.Details != "" && st.Details != ref {
			reason += " — " + st.Details
		}
		return jpkFact(EventJPKRejected, today, "rejected/"+key,
			map[string]any{"filing": f.ID, "code": st.Code, "reason": reason, "reference": ref, "date": today})
	}
	return jpkFact(EventJPKAccepted, today, "accepted/"+key,
		map[string]any{"filing": f.ID, "reference": ref, "timestamp": st.Timestamp, "date": today},
		core.FileInput{Input: "upo", Name: "UPO_" + ref + ".xml", MediaType: "application/xml", Bytes: upoBytes(st.UPO)})
}

// jpkFact is one of the channel's facts, keyed by what it is.
func jpkFact(typ, today, key string, payload map[string]any, files ...core.FileInput) (*core.Event, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &core.Event{Type: typ, OccurredAt: today, Payload: b, DedupKey: "jpk_gateway/" + key, Files: files}, nil
}

// signedReason says in the pack's words why a signed file is not sent.
func signedReason(err error) string {
	switch {
	case errors.Is(err, ErrJPKUnsigned):
		return "Plik nie ma podpisu — podpisz plik do podpisu podpisem kwalifikowanym albo profilem zaufanym i wgraj podpisany plik."
	case errors.Is(err, ErrJPKOtherRequest):
		return "Podpisany plik to inne zgłoszenie niż przygotowane dla tej wysyłki — podpisz plik do podpisu pobrany z tej wysyłki."
	default:
		return "W pliku nie ma zgłoszenia do podpisu (InitUpload) — wgraj plik do podpisu tej wysyłki, podpisany."
	}
}

// upoBytes is the UPO as the gateway gave it: XML, or base64 of it.
func upoBytes(upo string) []byte {
	if s := strings.TrimSpace(upo); strings.HasPrefix(s, "<") {
		return []byte(s)
	}
	if b, err := base64.StdEncoding.DecodeString(strings.Join(strings.Fields(upo), "")); err == nil {
		return b
	}
	return []byte(upo)
}

// paramOf reads one parameter from an issued document's "k=v,k=v" text.
func paramOf(params, key string) string {
	for _, kv := range strings.Split(params, ",") {
		if k, v, ok := strings.Cut(kv, "="); ok && k == key {
			return v
		}
	}
	return ""
}
