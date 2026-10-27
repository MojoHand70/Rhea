package shell_test

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"rhea/internal/core"
)

// TestPrepaidCosts (KK 2026-10-17: "I have clients where RMK is very
// important"): a purchase invoice that covers several months is spread —
// its cost moves to 640 and comes back to the cost account in equal parts,
// each part when the next month opens, dated on the last day of the month
// that ended, so a month's close (and its CIT advance) already sees it. The
// release is cumulative: a schedule spread late gives back the months that
// passed at once; the last part completes the amount, and a schedule fully
// released stops. The CIT charge lands in the month it closes.
func TestPrepaidCosts(t *testing.T) {
	w := newCITWorld(t)
	ctx, s, x, door := w.ctx, w.s, w.x, w.door
	door("declare_company_form", map[string]any{"company": w.self, "vat_status": "active", "books": "full", "income_tax": "cit",
		"from": "2026-01-01", "occurred_at": "2026-08-01"})
	door("declare_cit_status", map[string]any{"company": w.self, "year_start": "2026-01-01", "started_on": "2026-03-01",
		"formed_by": "new", "occurred_at": "2026-08-01"})
	w.buy("lic", "LIC/2026", "2026-09-15", "5260001246", "1200.00", "276.00", "1476.00")
	w.buy("ubz", "UBZ/2026", "2026-08-10", "5260001246", "600.00", "138.00", "738.00")
	w.buy("srv", "SRV/2026", "2026-09-16", "5260001246", "50.01", "11.50", "61.51")
	w.sale("fv-1", "FV 1/09/2026", "2026-09-21", "1000.00", "230.00", "1230.00")
	w.sale("fv-2", "FV 2/09/2026", "2026-09-24", "500.00", "115.00", "615.00")
	invoice := map[string]string{}
	docs, _ := s.ObjectsByType(ctx, "purchase_invoice")
	for _, d := range docs {
		invoice[d.State["number"].(string)] = d.ID
	}
	spread := func(no, start string, months int, on string) {
		t.Helper()
		door("spread_cost", map[string]any{"invoice": invoice[no], "start": start, "months": months, "occurred_at": on})
	}
	monthOpens := func(date string) {
		t.Helper()
		if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: core.EventMonthOpened, OccurredAt: date,
			Payload: []byte(fmt.Sprintf(`{"month":%q,"date":%q}`, date[:7], date)), DedupKey: "time/month/" + date}); err != nil {
			t.Fatal(err)
		}
		if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
			t.Fatalf("month opens %s: %v", date, errs)
		}
	}
	releases := func(period string) map[string]string {
		t.Helper()
		all, _ := s.ObjectsByType(ctx, "prepaid_release")
		pcs, _ := s.ObjectsByType(ctx, "prepaid_cost")
		no := map[string]string{}
		for _, pc := range pcs {
			for n, id := range invoice {
				if pc.State["invoice"] == id {
					no[pc.ID] = n
				}
			}
		}
		out := map[string]string{}
		for _, r := range all {
			if r.State["period"] == period {
				out[no[r.State["prepaid"].(string)]] = core.FormatMoney(int64(r.State["amount"].(float64)))
			}
		}
		return out
	}
	balance := func(code string) (int64, map[string]bool) {
		t.Helper()
		accounts, _ := s.ObjectsByType(ctx, "account")
		id := ""
		for _, a := range accounts {
			if a.State["code"] == code {
				id = a.ID
			}
		}
		postings, _ := s.ObjectsByType(ctx, "posting")
		var n int64
		dates := map[string]bool{}
		for _, p := range postings {
			if p.State["account"] != id {
				continue
			}
			amt := int64(p.State["amount"].(float64))
			if p.State["side"] == "credit" {
				amt = -amt
			}
			n += amt
			dates[p.State["date"].(string)] = true
		}
		return n, dates
	}

	// A year's licence from September and a two-month service: spread.
	spread("LIC/2026", "2026-09-01", 12, "2026-09-20")
	spread("SRV/2026", "2026-09-01", 2, "2026-09-20")
	// Twice, or for one month: not a spread — it waits.
	for _, in := range []map[string]any{{"invoice": invoice["LIC/2026"], "start": "2026-09-01", "months": 6},
		{"invoice": invoice["UBZ/2026"], "start": "2026-08-01", "months": 1}} {
		in["occurred_at"] = "2026-09-21"
		if tr, err := x.TriggerActivity(ctx, "spread_cost", in, "krzysztof", ""); err != nil || tr.Booked != 0 {
			t.Fatalf("a spread that should wait: %+v %v", tr, err)
		}
	}

	// October opens: September's parts, dated 30 September.
	monthOpens("2026-10-01")
	if got := releases("2026-09"); fmt.Sprint(got) != "map[LIC/2026:100.00 SRV/2026:25.01]" {
		t.Fatalf("September's parts: %v", got)
	}
	// September closes: its advance sees the parts — costs 100.00 + 25.01 of
	// the spread invoices, 600.00 of insurance not yet spread; 1500.00 sold.
	w.closeMonth("2026-09-30", "2026-10-05")
	w.want("2026-09", map[string]string{"costs": "725.01", "income": "774.99", "base": "775.00", "tax": "70.00", "advance": "70.00"})
	postings, _ := s.ObjectsByType(ctx, "posting")
	for _, p := range postings {
		if p.RuleID == "pl-post-cit-advance" && p.State["date"] != "2026-09-30" {
			t.Fatalf("September's CIT charge dated %v", p.State["date"])
		}
	}

	// The insurance, spread late from August: November's opening gives back
	// August, September and October at once — in October, September being closed.
	spread("UBZ/2026", "2026-08-01", 6, "2026-10-06")
	monthOpens("2026-11-01")
	if got := releases("2026-10"); fmt.Sprint(got) != "map[LIC/2026:100.00 SRV/2026:25.00 UBZ/2026:300.00]" {
		t.Fatalf("October's parts: %v", got)
	}
	pcs, _ := s.ObjectsByType(ctx, "prepaid_cost")
	done := 0
	for _, pc := range pcs {
		if pc.State["status"] == "done" {
			done++
			if pc.State["invoice"] != invoice["SRV/2026"] {
				t.Fatalf("finished: %v", pc.State)
			}
		}
	}
	if done != 1 {
		t.Fatalf("the two-month service is not finished: %d", done)
	}
	monthOpens("2026-12-01")
	if got := releases("2026-11"); fmt.Sprint(got) != "map[LIC/2026:100.00 UBZ/2026:100.00]" {
		t.Fatalf("November's parts: %v", got)
	}

	// What is left on 640: 1850.01 spread, 750.01 back in costs.
	if n, dates := balance("640"); n != 110000 || !dates["2026-09-30"] || !dates["2026-10-31"] || !dates["2026-11-30"] {
		t.Fatalf("640: %d, dated %v", n, dates)
	}
	type rows struct {
		Rows [][]cell `json:"rows"`
	}
	var rm rows
	w.get("/api/views/rmk", &rm)
	if got := fmt.Sprint(cellVals(rm.Rows)); got != "[[UBZ/2026 Nabywca S.A. 403 2026-08-01 2027-01-31 600.00 400.00 200.00] [LIC/2026 Nabywca S.A. 403 2026-09-01 2027-08-31 1200.00 300.00 900.00] [SRV/2026 Nabywca S.A. 403 2026-09-01 2026-10-31 50.01 50.01 0.00]]" {
		t.Fatalf("the schedules: %s", got)
	}
	// A line of the books names the record it was made from (KK 2026-10-27:
	// an RMK entry opened onto "Tego nie ma w księgach" — it shares no
	// invoice's fact): the spread's entry its schedule, a month's share its
	// release, the list and the one object alike.
	var made struct {
		Objects []struct {
			ID       string         `json:"object_id"`
			Rule     string         `json:"rule_id"`
			MadeFrom string         `json:"made_from"`
			State    map[string]any `json:"state"`
		} `json:"objects"`
	}
	w.get("/api/objects?type=posting", &made)
	kinds := map[string]int{}
	for _, p := range made.Objects {
		want := map[string]string{"pl-post-prepaid-cost": "prepaid_cost-", "pl-post-prepaid-release": "prepaid_release-"}[p.Rule]
		if want == "" {
			continue
		}
		if len(p.MadeFrom) < len(want) || p.MadeFrom[:len(want)] != want {
			t.Fatalf("%s made from %q, want a %s…", p.ID, p.MadeFrom, want)
		}
		var one struct {
			MadeFrom string `json:"made_from"`
		}
		w.get("/api/objects/"+p.ID, &one)
		if one.MadeFrom != p.MadeFrom {
			t.Fatalf("%s alone made from %q, in the list %q", p.ID, one.MadeFrom, p.MadeFrom)
		}
		kinds[want]++
	}
	if kinds["prepaid_cost-"] != 6 || kinds["prepaid_release-"] == 0 {
		t.Fatalf("RMK entries seen: %v", kinds)
	}
	var rano rows
	w.get("/api/views/rano", &rano)
	found := false
	for _, r := range cellVals(rano.Rows) {
		if r[0] == "prepaid" {
			found = r[2] == "1100.00" && r[3] == "2"
		}
	}
	if !found {
		t.Fatalf("the morning's figures: %v", cellVals(rano.Rows))
	}

	before, _ := s.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := s.AllObjects(ctx)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("replay does not reproduce the schedules")
	}
}

// TestPrepaidCostsInTheirMonths (KK's screenshot, 2026-10-27): on white01
// BIU/17/2026 (799,00, 14 September) was spread over twelve months from 1
// September on 18 October, September already closed, and the clock opened
// October only on the 19th, its first day running. The release took the
// day before the clock ran for the month's end, counted October as over
// and gave back September's and October's shares, 133,17, with October
// twelve days from its end; and the move to 640 was dated the day it was
// entered, whatever month the invoice's cost stood in. Now the release
// counts the months up to the one that opened, whatever day the clock
// runs, and the move is dated on the invoice's date — a closed month
// sending it to the day it was entered, as before.
func TestPrepaidCostsInTheirMonths(t *testing.T) {
	w := newCITWorld(t)
	ctx, s, x, door := w.ctx, w.s, w.x, w.door
	door("declare_company_form", map[string]any{"company": w.self, "vat_status": "active", "books": "full", "income_tax": "cit",
		"from": "2026-01-01", "occurred_at": "2026-08-01"})
	w.buy("biu", "BIU/17/2026", "2026-09-14", "5260001246", "799.00", "183.77", "982.77")
	w.buy("ins", "INS/2026", "2026-10-02", "5260001246", "1200.00", "276.00", "1476.00")
	invoice := map[string]string{}
	docs, _ := s.ObjectsByType(ctx, "purchase_invoice")
	for _, d := range docs {
		invoice[d.ID] = d.State["number"].(string)
		invoice[d.State["number"].(string)] = d.ID
	}
	monthOpens := func(date string) {
		t.Helper()
		if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: core.EventMonthOpened, OccurredAt: date,
			Payload: []byte(fmt.Sprintf(`{"month":%q,"date":%q}`, date[:7], date)), DedupKey: "time/month/" + date[:7]}); err != nil {
			t.Fatal(err)
		}
		if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
			t.Fatalf("month opens %s: %v", date, errs)
		}
	}
	// What the RMK did to the books: each entry by the invoice it spreads,
	// the period a share is for, the day it is dated and its amount.
	booked := func() []string {
		t.Helper()
		pcs, _ := s.ObjectsByType(ctx, "prepaid_cost")
		rels, _ := s.ObjectsByType(ctx, "prepaid_release")
		of := map[string]string{}
		for _, pc := range pcs {
			of[pc.ID] = invoice[pc.State["invoice"].(string)] + " na 640"
		}
		for _, r := range rels {
			of[r.ID] = fmt.Sprintf("%s za %v", invoice[r.State["invoice"].(string)], r.State["period"])
		}
		var lines struct {
			Objects []struct {
				MadeFrom string         `json:"made_from"`
				State    map[string]any `json:"state"`
			} `json:"objects"`
		}
		w.get("/api/objects?type=posting", &lines)
		var out []string
		for _, p := range lines.Objects {
			if what, ok := of[p.MadeFrom]; ok && p.State["side"] == "debit" {
				out = append(out, fmt.Sprintf("%s %v %s", what, p.State["date"], core.FormatMoney(int64(p.State["amount"].(float64)))))
			}
		}
		sort.Strings(out)
		return out
	}

	w.closeMonth("2026-09-30", "2026-10-05")
	door("spread_cost", map[string]any{"invoice": invoice["BIU/17/2026"], "start": "2026-09-01", "months": 12, "occurred_at": "2026-10-18"})
	monthOpens("2026-10-19")
	// An October invoice spread later in October, October open: the move
	// stands on the invoice's own day.
	door("spread_cost", map[string]any{"invoice": invoice["INS/2026"], "start": "2026-10-01", "months": 12, "occurred_at": "2026-10-20"})
	if got := fmt.Sprint(booked()); got != "[BIU/17/2026 na 640 2026-10-18 799.00 BIU/17/2026 za 2026-09 2026-10-19 66.58 INS/2026 na 640 2026-10-02 1200.00]" {
		t.Fatalf("October to date: %s", got)
	}
	// November opens on its first day: October's shares, dated its last.
	monthOpens("2026-11-01")
	if got := fmt.Sprint(booked()); got != "[BIU/17/2026 na 640 2026-10-18 799.00 BIU/17/2026 za 2026-09 2026-10-19 66.58 BIU/17/2026 za 2026-10 2026-10-31 66.59 INS/2026 na 640 2026-10-02 1200.00 INS/2026 za 2026-10 2026-10-31 100.00]" {
		t.Fatalf("October whole: %s", got)
	}
	before, _ := s.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := s.AllObjects(ctx)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("replay does not reproduce the RMK")
	}
}
