package pl_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"rhea/internal/adapter"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
	"rhea/markets/pl"
)

// The NBP channel (the Poland pack review, MON-16, FB-13, VAT-18): a
// document dated a day converts at the average rate of the last table
// published before that day, so a pass publishes, for every day of its
// window, the rate a document of that day uses — a weekend falls back to
// Friday's table, Monday to Friday's, Tuesday to Monday's — and the pack's
// rule makes each an fx_rate under its code. A second pass adds nothing.
func TestNBPRatesForEveryDay(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	installPoland(t, s, x)
	sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", "pl", "nbp.json"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-09-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve: %v %v", err, errs)
	}
	table := func(no, date, eur string) pl.NBPTable {
		return pl.NBPTable{No: no, EffectiveDate: date, Rates: []pl.NBPRate{{Code: "EUR", Mid: jsonNumber(eur)}, {Code: "JPY", Mid: "0.026"}}}
	}
	nbp := pl.NBP{
		Client: pl.FakeNBP{Tables_: []pl.NBPTable{
			table("187/A/NBP/2026", "2026-09-25", "4.2600"), // Friday
			table("188/A/NBP/2026", "2026-09-28", "4.2650"), // Monday
			table("189/A/NBP/2026", "2026-09-29", "4.2700"), // Tuesday
		}},
		Today: func() string { return "2026-09-30" }, From: "2026-09-26", Currencies: []string{"EUR"},
	}
	res, err := adapter.Run(ctx, s, x, nbp)
	if err != nil || res.PassErr != nil || res.Submitted != 5 || res.Booked != 5 {
		t.Fatalf("pass: %+v %v", res, err)
	}
	rates, _ := s.ObjectsByType(ctx, "fx_rate")
	got := map[string]string{}
	for _, r := range rates {
		got[r.State["code"].(string)] = r.State["rate"].(string)
	}
	want := map[string]string{
		"EUR/PLN/2026-09-26": "4.2600", "EUR/PLN/2026-09-27": "4.2600", "EUR/PLN/2026-09-28": "4.2600",
		"EUR/PLN/2026-09-29": "4.2650", "EUR/PLN/2026-09-30": "4.2700",
	}
	if len(got) != len(want) {
		t.Fatalf("rates: %v", got)
	}
	for code, rate := range want {
		if got[code] != rate {
			t.Fatalf("%s = %q, want %q (all %v)", code, got[code], rate, got)
		}
	}
	if res, err := adapter.Run(ctx, s, x, nbp); err != nil || res.Submitted != 0 || res.Duplicates != 5 {
		t.Fatalf("second pass: %+v %v", res, err)
	}
}

// TestNBPLive reads the real table A. Run by hand: RHEA_LIVE_NBP=1.
func TestNBPLive(t *testing.T) {
	if os.Getenv("RHEA_LIVE_NBP") == "" {
		t.Skip("set RHEA_LIVE_NBP=1 to read the NBP's API")
	}
	tables, err := pl.NBPWeb{}.Tables(context.Background(), "2026-09-21", "2026-09-30")
	if err != nil || len(tables) == 0 {
		t.Fatalf("tables: %d %v", len(tables), err)
	}
	for _, tb := range tables {
		for _, r := range tb.Rates {
			if r.Code == "EUR" {
				if !strings.Contains(r.Mid.String(), ".") {
					t.Fatalf("EUR on %s: %q", tb.EffectiveDate, r.Mid)
				}
				t.Logf("table %s of %s: EUR %s", tb.No, tb.EffectiveDate, r.Mid)
			}
		}
	}
}

func jsonNumber(s string) json.Number { return json.Number(s) }

// A rate is the installation's, once, whoever runs the pass (review
// 2026-10-21, F2): the owner's pass fanned out per company, so every company
// held its own copy of each day's rate beside the installation's, and a
// company's lookup of a day's rate found two — every EUR invoice, posting
// and match then waited in the worklist as ambiguous. Now the NBP's facts
// are the installation's: the pass runs once, for the installation, run by
// the owner or for a company alike.
func TestNBPRatesAreTheInstallations(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	installPoland(t, s, x)
	sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", "pl", "nbp.json"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-09-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve: %v %v", err, errs)
	}
	company := "company-95"
	raw, _ := json.Marshal(map[string]any{"name": "Helios sp. z o.o.", "kind": "self", "vat_id": "PL8058083055", "country": "PL"})
	if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: "company.registered", OccurredAt: "2026-09-01", Payload: raw, DedupKey: "self", Company: company}); err != nil {
		t.Fatal(err)
	}
	if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
		t.Fatal(errs)
	}
	nbp := pl.NBP{
		Client: pl.FakeNBP{Tables_: []pl.NBPTable{{No: "189/A/NBP/2026", EffectiveDate: "2026-09-29", Rates: []pl.NBPRate{{Code: "EUR", Mid: "4.2700"}}}}},
		Today:  func() string { return "2026-09-30" }, From: "2026-09-29", Currencies: []string{"EUR"},
	}
	// The window's one rate (the 30th converts at the 29th's table), once.
	if res, err := adapter.Run(ctx, s, x, nbp); err != nil || res.PassErr != nil || res.Submitted != 1 {
		t.Fatalf("the owner's pass, the rate once: %+v %v", res, err)
	}
	if res, err := adapter.Run(store.WithCompany(ctx, company), s, x, nbp); err != nil || res.Submitted != 0 {
		t.Fatalf("a pass for company-95 adds nothing: %+v %v", res, err)
	}
	rates, err := s.ObjectsByType(ctx, "fx_rate")
	if err != nil || len(rates) != 1 {
		t.Fatalf("one rate in all: %v %d", err, len(rates))
	}
	for _, r := range rates {
		if r.Company != "" {
			t.Fatalf("a rate of %q's", r.Company)
		}
	}
	if ids, err := s.FindObjectIDsByField(store.WithCompany(ctx, company), "fx_rate", "code", "EUR/PLN/2026-09-30"); err != nil || len(ids) != 1 {
		t.Fatalf("company-95 finds one rate for the day: %v %v", ids, err)
	}
}
