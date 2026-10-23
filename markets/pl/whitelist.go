package pl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"rhea/internal/adapter"
)

// VATList is the Ministry of Finance's list of VAT payers (Wykaz podatników
// VAT, the "white list", VAT act art. 96b): by NIP, the company's name, its
// VAT status, REGON and KRS, its address and the bank accounts it declared.
// It is open and wants no key; the Ministry lets a caller search a hundred
// times a day, so an answer is kept for the day it was given. The Ministry's
// requestId is its own mark of the answer — what an accountant keeps as
// proof of a check — and travels as the entry's Ref. A natural person's
// PESEL, which the list shows for some sole traders, is never passed on.
type VATList struct {
	Base string       // default https://wl-api.mf.gov.pl, or RHEA_VATLIST_URL
	HTTP *http.Client // nil: a 30 s client
	Now  func() time.Time

	mu   sync.Mutex
	kept map[string]adapter.RegisterEntry // nip|day → the day's answer
}

func init() {
	adapter.ProvideCompanyRegister(&VATList{Base: os.Getenv("RHEA_VATLIST_URL")})
}

// Name is the register's name in the binary.
func (v *VATList) Name() string { return "wykaz-vat" }

var vatStatus = map[string]string{"Czynny": "active", "Zwolniony": "exempt", "Niezarejestrowany": "unregistered"}

// Look asks the list what it says of nip today.
func (v *VATList) Look(ctx context.Context, nip string) (adapter.RegisterEntry, error) {
	now := time.Now
	if v.Now != nil {
		now = v.Now
	}
	at := now().In(warsaw())
	day, key := at.Format("2006-01-02"), nip+"|"+at.Format("2006-01-02")
	v.mu.Lock()
	if e, ok := v.kept[key]; ok {
		v.mu.Unlock()
		return e, nil
	}
	v.mu.Unlock()
	base := v.Base
	if base == "" {
		base = "https://wl-api.mf.gov.pl"
	}
	client := v.HTTP
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, "GET", strings.TrimRight(base, "/")+"/api/search/nip/"+url.PathEscape(nip)+"?date="+day, nil)
	if err != nil {
		return adapter.RegisterEntry{}, err
	}
	res, err := client.Do(req)
	if err != nil {
		return adapter.RegisterEntry{}, fmt.Errorf("the VAT list did not answer: %w", err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	var out struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Result  struct {
			Subject *struct {
				Name                  string   `json:"name"`
				NIP                   string   `json:"nip"`
				StatusVat             string   `json:"statusVat"`
				Regon                 string   `json:"regon"`
				KRS                   string   `json:"krs"`
				ResidenceAddress      string   `json:"residenceAddress"`
				WorkingAddress        string   `json:"workingAddress"`
				RegistrationLegalDate string   `json:"registrationLegalDate"`
				RemovalDate           string   `json:"removalDate"`
				AccountNumbers        []string `json:"accountNumbers"`
			} `json:"subject"`
			RequestID string `json:"requestId"`
		} `json:"result"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return adapter.RegisterEntry{}, fmt.Errorf("the VAT list answered %d, not as documented: %.200s", res.StatusCode, b)
	}
	if res.StatusCode != 200 {
		return adapter.RegisterEntry{}, fmt.Errorf("the VAT list refused: %s (%s)", out.Message, out.Code)
	}
	e := adapter.RegisterEntry{Register: v.Name(), TaxID: nip, Ref: out.Result.RequestID, ReadAt: at.Format(time.RFC3339), Fields: map[string]string{}}
	if s := out.Result.Subject; s != nil {
		e.Found = true
		address := s.WorkingAddress
		if address == "" {
			address = s.ResidenceAddress
		}
		status := vatStatus[s.StatusVat]
		if status == "" {
			status = s.StatusVat
		}
		for k, val := range map[string]string{"name": s.Name, "vat_status": status, "regon": s.Regon, "krs": s.KRS, "address": address,
			"accounts": strings.Join(s.AccountNumbers, ","), "registered_on": s.RegistrationLegalDate, "closed_on": s.RemovalDate, "country": "PL"} {
			if val != "" {
				e.Fields[k] = val
			}
		}
	}
	v.mu.Lock()
	if v.kept == nil {
		v.kept = map[string]adapter.RegisterEntry{}
	}
	v.kept[key] = e
	v.mu.Unlock()
	return e, nil
}

// warsaw is the day the Ministry keeps; a machine without the zone data falls
// back to the Polish offset of the season.
func warsaw() *time.Location {
	if l, err := time.LoadLocation("Europe/Warsaw"); err == nil {
		return l
	}
	return time.FixedZone("CET", 3600)
}
