package pl

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"rhea/internal/adapter"
)

// GUS is Statistics Poland's business register (REGON), read through its BIR
// 1.1 service: by NIP, the company's full name, REGON, seat and — from the
// full report of a legal person — its legal form and KRS number. A sole
// trader is a natural person by the register's own type. The service wants a
// key GUS issues free on request (RHEA_GUS_KEY); RHEA_GUS_ENV=test reads its
// test service with the key GUS publishes for it, where the data is real and
// the streets are not. With neither, the register says it is not set up.
// RHEA_GUS_URL names another service of the same shape (a recorded one in a
// rehearsal).
type GUS struct {
	URL  string // the service; default by RHEA_GUS_ENV
	Key  string
	HTTP *http.Client // nil: a 30 s client
	Now  func() time.Time

	mu  sync.Mutex
	sid string
	at  time.Time // when the session was opened: GUS closes one idle for an hour
}

const (
	gusProd    = "https://wyszukiwarkaregon.stat.gov.pl/wsBIR/UslugaBIRzewnPubl.svc"
	gusTest    = "https://wyszukiwarkaregontest.stat.gov.pl/wsBIR/UslugaBIRzewnPubl.svc"
	gusTestKey = "abcde12345abcde12345" // published by GUS for its test service
	gusNS      = "http://CIS/BIR/PUBL/2014/07"
)

func init() { adapter.ProvideCompanyRegister(GUSFromEnv()) }

// GUSFromEnv is the register as the settings name it.
func GUSFromEnv() *GUS {
	g := &GUS{URL: gusProd, Key: os.Getenv("RHEA_GUS_KEY")}
	if os.Getenv("RHEA_GUS_ENV") == "test" {
		g.URL = gusTest
		if g.Key == "" {
			g.Key = gusTestKey
		}
	}
	if u := os.Getenv("RHEA_GUS_URL"); u != "" {
		g.URL = u
	}
	return g
}

// Name is the register's name in the binary.
func (g *GUS) Name() string { return "gus" }

// Look asks the register who nip is.
func (g *GUS) Look(ctx context.Context, nip string) (adapter.RegisterEntry, error) {
	if g.Key == "" {
		return adapter.RegisterEntry{}, errors.New("GUS's register is not set up here: it needs the key GUS issues (RHEA_GUS_KEY)")
	}
	now := time.Now
	if g.Now != nil {
		now = g.Now
	}
	e := adapter.RegisterEntry{Register: g.Name(), TaxID: nip, ReadAt: now().Format(time.RFC3339), Fields: map[string]string{}}
	found, err := g.ask(ctx, "DaneSzukajPodmioty", `<ns:pParametryWyszukiwania><dat:Nip>`+nip+`</dat:Nip></ns:pParametryWyszukiwania>`)
	if err != nil {
		return e, err
	}
	var hits struct {
		Dane []map[string]string
	}
	if err := decodeRows(found, &hits.Dane); err != nil {
		return e, err
	}
	var d map[string]string
	for _, row := range hits.Dane {
		if row["ErrorCode"] == "4" {
			return e, nil // the register knows no such NIP
		}
		if row["ErrorCode"] != "" {
			return e, fmt.Errorf("GUS: %s", row["ErrorMessagePl"])
		}
		if row["Typ"] == "P" || row["Typ"] == "F" { // the entity itself, not one of its local units
			d = row
			break
		}
	}
	if d == nil {
		return e, nil
	}
	e.Found = true
	street := strings.TrimSpace(d["Ulica"] + " " + d["NrNieruchomosci"])
	if d["NrLokalu"] != "" {
		street += "/" + d["NrLokalu"]
	}
	city := d["Miejscowosc"]
	if d["MiejscowoscPoczty"] != "" {
		city = d["MiejscowoscPoczty"]
	}
	put := func(k, v string) {
		if v = strings.TrimSpace(v); v != "" {
			e.Fields[k] = v
		}
	}
	put("name", d["Nazwa"])
	put("regon", d["Regon"])
	put("street", d["Ulica"])
	put("building", d["NrNieruchomosci"])
	put("unit", d["NrLokalu"])
	put("postcode", d["KodPocztowy"])
	put("city", city)
	put("address", strings.TrimLeft(strings.TrimSpace(street)+", "+strings.TrimSpace(d["KodPocztowy"]+" "+city), ", "))
	put("country", "PL")
	put("closed_on", d["DataZakonczeniaDzialalnosci"])
	if d["Typ"] == "F" {
		put("legal_form", "natural_person")
		put("legal_form_name", "osoba fizyczna prowadząca działalność gospodarczą")
		return e, nil
	}
	// a legal person's full report: its legal form, its short name, its KRS number
	report, err := g.ask(ctx, "DanePobierzPelnyRaport", `<ns:pRegon>`+d["Regon"]+`</ns:pRegon><ns:pNazwaRaportu>BIR11OsPrawna</ns:pNazwaRaportu>`)
	if err != nil {
		return e, err
	}
	var full []map[string]string
	if err := decodeRows(report, &full); err != nil || len(full) == 0 {
		return e, err
	}
	f := full[0]
	put("short_name", f["praw_nazwaSkrocona"])
	put("registered_on", f["praw_dataPowstania"])
	if f["praw_rodzajRejestruEwidencji_Symbol"] == "138" { // the National Court Register's register of businesses
		put("krs", f["praw_numerWRejestrzeEwidencji"])
	}
	put("legal_form_name", strings.ToLower(f["praw_szczegolnaFormaPrawna_Nazwa"]))
	put("legal_form_code", f["praw_szczegolnaFormaPrawna_Symbol"])
	if f["praw_podstawowaFormaPrawna_Symbol"] == "1" { // a legal person; a partnership is the person's to say
		put("legal_form", "legal_entity")
	}
	return e, nil
}

// ask calls one operation of the service in a session, opening one when none
// is open or the last has lapsed, and answers its result: the register's own
// XML, unescaped.
func (g *GUS) ask(ctx context.Context, op, params string) (string, error) {
	for try := 0; try < 2; try++ {
		sid, err := g.session(ctx, try > 0)
		if err != nil {
			return "", err
		}
		out, err := g.call(ctx, op, params, sid)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(out) != "" {
			return out, nil
		}
		// an empty answer is a lapsed session: open a new one, once
	}
	return "", fmt.Errorf("GUS answered nothing to %s", op)
}

func (g *GUS) session(ctx context.Context, fresh bool) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !fresh && g.sid != "" && time.Since(g.at) < 50*time.Minute {
		return g.sid, nil
	}
	sid, err := g.call(ctx, "Zaloguj", `<ns:pKluczUzytkownika>`+xmlText(g.Key)+`</ns:pKluczUzytkownika>`, "")
	if err != nil {
		return "", err
	}
	if sid = strings.TrimSpace(sid); sid == "" {
		return "", errors.New("GUS refused the key (RHEA_GUS_KEY)")
	}
	g.sid, g.at = sid, time.Now()
	return sid, nil
}

// call sends one SOAP 1.2 request and reads the operation's result from the
// envelope inside the multipart answer the service sends.
func (g *GUS) call(ctx context.Context, op, params, sid string) (string, error) {
	body := `<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:ns="` + gusNS + `" xmlns:dat="` + gusNS + `/DataContract">` +
		`<soap:Header xmlns:wsa="http://www.w3.org/2005/08/addressing"><wsa:To>` + g.URL + `</wsa:To><wsa:Action>` + gusNS + `/IUslugaBIRzewnPubl/` + op + `</wsa:Action></soap:Header>` +
		`<soap:Body><ns:` + op + `>` + params + `</ns:` + op + `></soap:Body></soap:Envelope>`
	req, err := http.NewRequestWithContext(ctx, "POST", g.URL, strings.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	if sid != "" {
		req.Header.Set("sid", sid)
	}
	client := g.HTTP
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("GUS did not answer: %w", err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if res.StatusCode != 200 {
		return "", fmt.Errorf("GUS answered %d to %s", res.StatusCode, op)
	}
	i, j := bytes.Index(b, []byte("<s:Envelope")), bytes.LastIndex(b, []byte("</s:Envelope>"))
	if i < 0 || j < i {
		return "", fmt.Errorf("GUS's answer to %s holds no envelope", op)
	}
	var env struct {
		Body struct {
			Inner struct {
				Result string `xml:",any"`
			} `xml:",any"`
		} `xml:"Body"`
	}
	if err := xml.Unmarshal(b[i:j+len("</s:Envelope>")], &env); err != nil {
		return "", fmt.Errorf("GUS's answer to %s: %w", op, err)
	}
	return env.Body.Inner.Result, nil
}

// decodeRows reads the register's <root><dane>…</dane></root>: each dane a
// row of named values.
func decodeRows(doc string, rows *[]map[string]string) error {
	dec := xml.NewDecoder(strings.NewReader(doc))
	var row map[string]string
	var name string
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("GUS's data: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "dane" {
				row = map[string]string{}
			} else if row != nil {
				name = t.Name.Local
			}
		case xml.CharData:
			if row != nil && name != "" {
				row[name] += string(t)
			}
		case xml.EndElement:
			if t.Name.Local == "dane" && row != nil {
				*rows = append(*rows, row)
				row = nil
			}
			name = ""
		}
	}
}

func xmlText(s string) string { return html.EscapeString(s) }
