package pl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// The registers a person adding a company reads by its NIP (DIRECTION
// 2026-10-19, "People and rights": onboarding), on the answers GUS's test
// service and the Ministry's VAT list gave on 2026-10-23 (testdata/registers).

func recorded(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "registers", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestGUSSaysWhoANIPIs: a legal person by its search and full report — name,
// REGON, KRS, seat, legal form; a sole trader as a natural person without a
// report; a NIP the register does not know, found false and no error; a
// lapsed session opened again once; no key, and a refused key, said so.
func TestGUSSaysWhoANIPIs(t *testing.T) {
	var logins, reports atomic.Int32
	var lapse atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := new(strings.Builder)
		buf := make([]byte, 1<<16)
		for {
			n, err := r.Body.Read(buf)
			b.Write(buf[:n])
			if err != nil {
				break
			}
		}
		body := b.String()
		w.Header().Set("Content-Type", `multipart/related; type="application/xop+xml"`)
		switch {
		case strings.Contains(body, "/Zaloguj</wsa:Action>"):
			logins.Add(1)
			if strings.Contains(body, ">wrong<") {
				in := recorded(t, "gus-login.txt") // GUS answers a key it refuses with an empty session
				w.Write([]byte(in[:strings.Index(in, "<ZalogujResult>")+len("<ZalogujResult>")] + in[strings.Index(in, "</ZalogujResult>"):]))
				return
			}
			w.Write([]byte(recorded(t, "gus-login.txt")))
		case r.Header.Get("sid") == "":
			t.Errorf("a call without its session: %.80s", body)
		case strings.Contains(body, "/DaneSzukajPodmioty</wsa:Action>"):
			if lapse.CompareAndSwap(true, false) { // GUS answers a lapsed session with an empty result
				none := recorded(t, "gus-search-none.txt")
				w.Write([]byte(none[:strings.Index(none, "<DaneSzukajPodmiotyResult>")+len("<DaneSzukajPodmiotyResult>")] + none[strings.Index(none, "</DaneSzukajPodmiotyResult>"):]))
				return
			}
			switch {
			case strings.Contains(body, "<dat:Nip>7740001454<"):
				w.Write([]byte(recorded(t, "gus-search-p.txt")))
			case strings.Contains(body, "<dat:Nip>5555555555<"): // a sole trader: the same answer as a natural person's
				w.Write([]byte(strings.Replace(recorded(t, "gus-search-p.txt"), "&lt;Typ&gt;P&lt;/Typ&gt;", "&lt;Typ&gt;F&lt;/Typ&gt;", 1)))
			default:
				w.Write([]byte(recorded(t, "gus-search-none.txt")))
			}
		case strings.Contains(body, "/DanePobierzPelnyRaport</wsa:Action>"):
			reports.Add(1)
			w.Write([]byte(recorded(t, "gus-report-p.txt")))
		}
	}))
	defer srv.Close()
	ctx := context.Background()
	g := &GUS{URL: srv.URL, Key: gusTestKey}
	e, err := g.Look(ctx, "7740001454")
	f := e.Fields
	if err != nil || !e.Found || f["name"] != "POLSKI KONCERN NAFTOWY ORLEN SPÓŁKA AKCYJNA" || f["regon"] != "610188201" || f["krs"] != "0000028860" ||
		f["legal_form"] != "legal_entity" || f["legal_form_name"] != "spółki akcyjne" || f["address"] != "ul. Test-Wilcza 7, 09-411 Płock" ||
		f["postcode"] != "09-411" || f["city"] != "Płock" || f["short_name"] != "PKN ORLEN S.A." || e.Register != "gus" || e.TaxID != "7740001454" {
		t.Fatalf("a legal person: %v %+v", err, e)
	}
	if e, err := g.Look(ctx, "5555555555"); err != nil || !e.Found || e.Fields["legal_form"] != "natural_person" || e.Fields["krs"] != "" || reports.Load() != 1 {
		t.Fatalf("a sole trader, a natural person without a full report: %v %+v (%d reports)", err, e, reports.Load())
	}
	if e, err := g.Look(ctx, "1111111111"); err != nil || e.Found {
		t.Fatalf("a NIP the register does not know: %v %+v", err, e)
	}
	if logins.Load() != 1 {
		t.Fatalf("one session serves every look while it lasts: %d logins", logins.Load())
	}
	lapse.Store(true)
	if e, err := g.Look(ctx, "7740001454"); err != nil || !e.Found || logins.Load() != 2 {
		t.Fatalf("a lapsed session is opened again, once: %v %v (%d logins)", err, e.Found, logins.Load())
	}
	if _, err := (&GUS{URL: srv.URL}).Look(ctx, "7740001454"); err == nil || !strings.Contains(err.Error(), "RHEA_GUS_KEY") {
		t.Fatalf("no key: %v", err)
	}
	if _, err := (&GUS{URL: srv.URL, Key: "wrong"}).Look(ctx, "7740001454"); err == nil || !strings.Contains(err.Error(), "refused the key") {
		t.Fatalf("a refused key: %v", err)
	}
}

// TestTheVATListSaysWhoANIPIs: the Ministry's answer — name, VAT status as a
// word of our own, REGON, KRS, address, accounts, its requestId kept as the
// proof of the check — asked for the Polish day, kept for that day so the
// Ministry's daily limit is spent once per NIP; a NIP it does not list found
// false; a refusal in the Ministry's words.
func TestTheVATListSaysWhoANIPIs(t *testing.T) {
	var asked atomic.Int32
	var days []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		asked.Add(1)
		days = append(days, r.URL.Query().Get("date"))
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/search/nip/7740001454":
			w.Write([]byte(recorded(t, "wl-found.json")))
		case "/api/search/nip/123":
			w.WriteHeader(400)
			w.Write([]byte(recorded(t, "wl-bad.json")))
		default:
			w.Write([]byte(recorded(t, "wl-none.json")))
		}
	}))
	defer srv.Close()
	now := time.Date(2026, 10, 22, 22, 30, 0, 0, time.UTC) // already the 23rd in Warsaw
	v := &VATList{Base: srv.URL, Now: func() time.Time { return now }}
	ctx := context.Background()
	e, err := v.Look(ctx, "7740001454")
	f := e.Fields
	if err != nil || !e.Found || f["name"] != "ORLEN SPÓŁKA AKCYJNA" || f["vat_status"] != "active" || f["regon"] != "610188201" || f["krs"] != "0000028860" ||
		f["address"] != "CHEMIKÓW 7, 09-411 PŁOCK" || !strings.HasPrefix(f["accounts"], "30102010260000140203550589,") || e.Ref == "" || e.Register != "wykaz-vat" {
		t.Fatalf("a VAT payer: %v %+v", err, e)
	}
	if days[0] != "2026-10-23" {
		t.Fatalf("asked for the Polish day: %v", days)
	}
	if _, err := v.Look(ctx, "7740001454"); err != nil || asked.Load() != 1 {
		t.Fatalf("the same NIP the same day is the day's answer: %d asked", asked.Load())
	}
	now = now.Add(24 * time.Hour)
	if _, err := v.Look(ctx, "7740001454"); err != nil || asked.Load() != 2 {
		t.Fatalf("a new day asks again: %d asked", asked.Load())
	}
	if e, err := v.Look(ctx, "5260250995"); err != nil || e.Found || e.Ref == "" {
		t.Fatalf("a NIP the list does not hold: %v %+v", err, e)
	}
	if _, err := v.Look(ctx, "123"); err == nil || !strings.Contains(err.Error(), "nieprawidłową długość") {
		t.Fatalf("a refusal in the Ministry's words: %v", err)
	}
}

// TestGUSLive asks GUS's test service for real (RHEA_GUS_LIVE=1): the
// service, its envelope and its data as they are today.
func TestGUSLive(t *testing.T) {
	if os.Getenv("RHEA_GUS_LIVE") == "" {
		t.Skip("RHEA_GUS_LIVE=1 asks GUS's test service")
	}
	e, err := (&GUS{URL: gusTest, Key: gusTestKey}).Look(context.Background(), "7740001454")
	if err != nil || !e.Found || e.Fields["regon"] != "610188201" || e.Fields["legal_form"] != "legal_entity" {
		t.Fatalf("GUS's test service: %v %+v", err, e)
	}
}
