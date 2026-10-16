package pl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"rhea/internal/adapter"
	"rhea/internal/core"
)

// NBP is the channel of the National Bank of Poland's average exchange rates
// (table A). Polish law converts a document in a foreign currency at the
// average rate the NBP published on the last working day before the
// document's date — VAT art. 31a ust. 1, the accounting act art. 30 ust. 2
// pkt 2, PIT art. 11a ust. 1, CIT art. 12 ust. 2 — so a pass publishes, for
// every calendar day of its window and every currency it follows, the rate a
// document dated that day converts at: fx.rate.published, code CUR/PLN/day,
// the table it came from named on the event. A day falls back to the last
// table before it, so weekends and holidays need no rule of their own. The
// rate is the NBP's own decimal text, never a float.
type NBP struct {
	Client     NBPClient
	Today      func() string // the last day of the window; defaults to the wall clock
	From       string        // the first day; defaults to 31 days before Today
	Currencies []string      // defaults to EUR, USD, GBP, CHF
}

// NBPTable is one published table A.
type NBPTable struct {
	No            string    `json:"no"`
	EffectiveDate string    `json:"effectiveDate"`
	Rates         []NBPRate `json:"rates"`
}

// NBPRate is one currency's average rate, as published.
type NBPRate struct {
	Code string      `json:"code"`
	Mid  json.Number `json:"mid"`
}

// NBPClient fetches the tables published between two dates, inclusive.
type NBPClient interface {
	Tables(ctx context.Context, start, end string) ([]NBPTable, error)
}

const nbpCodeVersion = 1

func init() {
	adapter.Register(adapter.Implemented{Name: "nbp", CodeVersion: nbpCodeVersion,
		Format: "NBP Web API, table A (average rates), JSON"})
	adapter.Provide("nbp", func() adapter.Adapter {
		return NBP{Client: NBPWeb{}, From: os.Getenv("RHEA_NBP_FROM")}
	})
}

func (NBP) Name() string     { return "nbp" }
func (NBP) CodeVersion() int { return nbpCodeVersion }

// EventFXRatePublished is the fact a rate is: master data, append-only.
const EventFXRatePublished = "fx.rate.published"

func (n NBP) Pass(ctx context.Context, _ adapter.Reads) ([]core.Event, error) {
	today := time.Now().UTC().Format("2006-01-02")
	if n.Today != nil {
		today = n.Today()
	}
	last, err := time.Parse("2006-01-02", today)
	if err != nil {
		return nil, fmt.Errorf("nbp: today %q: %w", today, err)
	}
	first := last.AddDate(0, 0, -31)
	if n.From != "" {
		if first, err = time.Parse("2006-01-02", n.From); err != nil {
			return nil, fmt.Errorf("nbp: from %q: %w", n.From, err)
		}
	}
	if first.After(last) {
		return nil, fmt.Errorf("nbp: the window starts after it ends (%s > %s)", first.Format("2006-01-02"), today)
	}
	currencies := n.Currencies
	if len(currencies) == 0 {
		currencies = []string{"EUR", "USD", "GBP", "CHF"}
	}
	// The tables that can stand for a day of the window: from ten days
	// before its first day (a long holiday) to the day before its last.
	var tables []NBPTable
	for start := first.AddDate(0, 0, -10); start.Before(last); start = start.AddDate(0, 0, 93) {
		end := start.AddDate(0, 0, 92)
		if !end.Before(last) {
			end = last.AddDate(0, 0, -1)
		}
		got, err := n.Client.Tables(ctx, start.Format("2006-01-02"), end.Format("2006-01-02"))
		if err != nil {
			return nil, err
		}
		tables = append(tables, got...)
	}
	sort.Slice(tables, func(i, j int) bool { return tables[i].EffectiveDate < tables[j].EffectiveDate })

	var out []core.Event
	for day := first; !day.After(last); day = day.AddDate(0, 0, 1) {
		d := day.Format("2006-01-02")
		// The last table published before the day.
		var table *NBPTable
		for i := range tables {
			if tables[i].EffectiveDate < d {
				table = &tables[i]
			}
		}
		if table == nil {
			continue // no table before it in reach: the day stays without a rate, and a document dated it waits
		}
		for _, cur := range currencies {
			var mid string
			for _, r := range table.Rates {
				if r.Code == cur {
					mid = r.Mid.String()
				}
			}
			if mid == "" {
				continue
			}
			if _, err := core.ParseDecimal(mid); err != nil {
				return nil, fmt.Errorf("nbp: table %s %s: %q is not a decimal", table.No, cur, mid)
			}
			code := cur + "/PLN/" + d
			payload, err := json.Marshal(map[string]any{
				"code": code, "base": cur, "quote": "PLN", "date": d, "rate": mid,
				"table": table.No, "table_date": table.EffectiveDate, "market": "pl",
			})
			if err != nil {
				return nil, err
			}
			out = append(out, core.Event{Type: EventFXRatePublished, OccurredAt: d, Payload: payload,
				DedupKey: "nbp/" + code})
		}
	}
	return out, nil
}

// NBPWeb is the NBP's public Web API: no key, JSON, at most 93 days a call.
type NBPWeb struct {
	BaseURL string // defaults to https://api.nbp.pl
	HTTP    *http.Client
}

func (w NBPWeb) Tables(ctx context.Context, start, end string) ([]NBPTable, error) {
	base := w.BaseURL
	if base == "" {
		base = "https://api.nbp.pl"
	}
	client := w.HTTP
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/api/exchangerates/tables/A/%s/%s/?format=json", base, start, end), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	// The API refuses Go's default agent ("request forbidden by
	// administrative rules"): a client says who it is.
	req.Header.Set("User-Agent", "Rhea/1 (exchange rates for bookkeeping)")
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("nbp: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("nbp: %w", err)
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, nil // "Brak danych": no table published in the range
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nbp: %s %s–%s: %s %s", "table A", start, end, res.Status, strings.TrimSpace(string(body)))
	}
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.UseNumber() // a rate is the bank's decimal text, never a float
	var tables []NBPTable
	if err := dec.Decode(&tables); err != nil {
		return nil, fmt.Errorf("nbp: %w", err)
	}
	return tables, nil
}

// FakeNBP serves fixed tables: the deterministic client of tests and demos.
type FakeNBP struct{ Tables_ []NBPTable }

func (f FakeNBP) Tables(_ context.Context, start, end string) ([]NBPTable, error) {
	var out []NBPTable
	for _, t := range f.Tables_ {
		if t.EffectiveDate >= start && t.EffectiveDate <= end {
			out = append(out, t)
		}
	}
	return out, nil
}
