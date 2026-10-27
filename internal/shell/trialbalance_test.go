package shell_test

import (
	"reflect"
	"testing"

	"rhea/internal/core"
)

// The trial balance is the Zestawienie obrotów i sald the law describes (UoR
// art. 18 ust. 1; KK, 2026-10-27), for the period a person picks: each
// account's balance on the day the financial year opened, the period's
// turnover, the turnover since the year opened, and the balance at the
// period's end, each on its side — and each row names the account it is,
// so a face opens that account's ledger and "why?" on a cell is answered.
// It summed every entry since the books began, period or none.
//
// Alfa's year begins in April. In September 2026: 201 opens at 1 000,00 Wn
// (March's invoice, before the year), turns 250,00 Wn in September and
// 250,00 Wn / 600,00 Ma since April, and closes at 650,00 Wn; 700 opens at
// 1 000,00 Ma and closes at 1 250,00 Ma; 130 holds May's 600,00. October's
// payment is after the period and counts nowhere. Beta sees none of it.
func TestTheTrialBalanceIsTheStatutoryOneForAPeriod(t *testing.T) {
	w, a, b := newBooksWorld(t)
	for _, r := range []core.Rule{
		{ID: "test-posting", Description: "a line of the books", Spec: core.RuleSpec{Match: core.Match{EventType: "test.posted"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "posting", Fields: map[string]string{"entry": "=$.entry", "line": "=$.line", "book": "main",
				"account": "=$.account", "side": "=$.side", "amount": "=$.amount", "currency": "PLN", "date": "=$.date"}}}}},
	} {
		r.Status, r.Priority, r.EffectiveFrom, r.CreatedBy = core.StatusActive, 100, "2026-01-01", "test"
		if _, err := w.s.InsertRuleVersion(w.ctx, r); err != nil {
			t.Fatal(err)
		}
	}
	w.mustDoor("kk", a, "declare_financial_year", map[string]any{"financial_year_start": "04"})
	event := func(typ, date, key string, payload map[string]any) {
		t.Helper()
		var out map[string]any
		if code := w.call("kk", a, "POST", "/api/events", map[string]any{"event_type": typ, "occurred_at": date, "dedup_key": key, "payload": payload}, &out); code != 200 || out["booked"] != float64(1) {
			t.Fatalf("%s: %d %v", key, code, out)
		}
	}
	accounts, names := map[string]string{}, map[string]string{} // the Poland pack's chart of accounts
	for _, code := range []string{"130", "201", "700"} {
		ids, err := w.s.FindObjectIDsByField(w.ctx, "account", "code", code)
		if err != nil || len(ids) != 1 {
			t.Fatalf("account %s: %v %v", code, ids, err)
		}
		o, err := w.s.GetObject(w.ctx, ids[0])
		if err != nil {
			t.Fatal(err)
		}
		accounts[code], names[code] = ids[0], o.State["name"].(string)
	}
	entry := func(no, date, debit, credit, amount string) {
		t.Helper()
		event("test.posted", date, no+"-1", map[string]any{"entry": no, "line": 1, "account": accounts[debit], "side": "debit", "amount": amount, "date": date})
		event("test.posted", date, no+"-2", map[string]any{"entry": no, "line": 2, "account": accounts[credit], "side": "credit", "amount": amount, "date": date})
	}
	entry("FV/1", "2026-03-15", "201", "700", "1000.00")
	entry("WB/1", "2026-05-10", "130", "201", "600.00")
	entry("FV/2", "2026-09-05", "201", "700", "250.00")
	entry("WB/2", "2026-10-02", "130", "201", "100.00")

	var tb struct {
		Columns []struct {
			Field string `json:"field"`
			Type  string `json:"type"`
		} `json:"columns"`
		Rows   [][]struct{ V string } `json:"rows"`
		Of     [][]string             `json:"row_objects"`
		Period core.Period            `json:"period"`
	}
	september := "?from=2026-09-01&to=2026-09-30"
	if code := w.call("kk", a, "GET", "/api/views/trial-balance"+september, nil, &tb); code != 200 {
		t.Fatalf("the trial balance: %d", code)
	}
	var fields []string
	for _, c := range tb.Columns {
		fields = append(fields, c.Field)
	}
	if want := []string{"book", "code", "account", "opening_debit", "opening_credit", "debit", "credit", "ytd_debit", "ytd_credit", "closing_debit", "closing_credit"}; !reflect.DeepEqual(fields, want) {
		t.Fatalf("columns %v, want %v", fields, want)
	}
	var got [][]string
	for _, r := range tb.Rows {
		var row []string
		for _, c := range r {
			row = append(row, c.V)
		}
		got = append(got, row)
	}
	want := [][]string{
		{"main", "130", names["130"], "0.00", "0.00", "0.00", "0.00", "600.00", "0.00", "600.00", "0.00"},
		{"main", "201", names["201"], "1000.00", "0.00", "250.00", "0.00", "250.00", "600.00", "650.00", "0.00"},
		{"main", "700", names["700"], "0.00", "1000.00", "0.00", "250.00", "0.00", "250.00", "0.00", "1250.00"},
	}
	if !reflect.DeepEqual(got, want) || tb.Period != (core.Period{From: "2026-09-01", To: "2026-09-30", YearFrom: "2026-04-01"}) {
		t.Fatalf("Alfa's September:\n got %v %+v\nwant %v", got, tb.Period, want)
	}
	if !reflect.DeepEqual(tb.Of, [][]string{{accounts["130"]}, {accounts["201"]}, {accounts["700"]}}) {
		t.Fatalf("each row names its account: %v, want %v", tb.Of, accounts)
	}
	var why struct {
		Objects []struct {
			ID string `json:"object_id"`
		} `json:"objects"`
	}
	if code := w.call("kk", a, "GET", "/api/explain?view=trial-balance&row=1&col=closing_debit&from=2026-09-01&to=2026-09-30", nil, &why); code != 200 ||
		len(why.Objects) != 1 || why.Objects[0].ID != accounts["201"] {
		t.Fatalf("why on 201's closing balance: %d %+v", code, why)
	}
	if code := w.call("olga", b, "GET", "/api/views/trial-balance"+september, nil, &tb); code != 200 || len(tb.Rows) != 0 || tb.Period.YearFrom != "2026-01-01" {
		t.Fatalf("Beta's trial balance: %d %v %+v", code, tb.Rows, tb.Period)
	}
}

// The group trial balance in EUR is for a period too (KK, 2026-10-27): the
// same columns as a company's, every company's lines translated at the EUR
// rate of the period's last day — the latest published on or before it —
// and the translation difference (CTA) balancing each pair of columns. It
// summed every line since the books began at the latest rate there was.
//
// Alfa keeps PLN, Beta EUR; the rate is 4.25 at September's end, 4.10 at
// October's. In September: 201 opens at 200,00 Wn (December's 850,00 PLN),
// turns 140,00 Wn in the month (Alfa's 170,00 PLN and Beta's 100,00 EUR)
// and 140,00 / 100,00 since January, and closes at 240,00 Wn; 700 closes at
// 340,00 Ma; 130 holds May's 425,00 PLN and Beta's 10,00 EUR, a line with no
// other side, which the CTA line balances in the month, the year and the
// close. October's line counts nowhere.
func TestTheGroupTrialBalanceIsForAPeriodToo(t *testing.T) {
	w, a, b := newBooksWorld(t)
	if _, err := w.s.InsertRuleVersion(w.ctx, core.Rule{ID: "test-posting", Status: core.StatusActive, Priority: 100, EffectiveFrom: "2025-01-01", CreatedBy: "test",
		Description: "a line of the books", Spec: core.RuleSpec{Match: core.Match{EventType: "test.posted"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "posting", Fields: map[string]string{"entry": "=$.entry", "line": "=$.line", "book": "main",
				"account": "=$.account", "side": "=$.side", "amount": "=$.amount", "currency": "=$.currency", "date": "=$.date"}}}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := w.s.InsertRuleVersion(w.ctx, core.Rule{ID: "test-fx-rate", Status: core.StatusActive, Priority: 100, EffectiveFrom: "2025-01-01", CreatedBy: "test",
		Description: "a published rate, as the NBP channel brings it", Spec: core.RuleSpec{Match: core.Match{EventType: "fx.rate.published"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "fx_rate", Fields: map[string]string{"code": "=$.code", "base": "=$.base", "quote": "=$.quote", "date": "=$.date", "rate": "=$.rate"}}}}}); err != nil {
		t.Fatal(err)
	}
	for _, r := range [][2]string{{"2026-09-30", "4.25"}, {"2026-10-31", "4.10"}} {
		if code := w.call("", "", "POST", "/api/events", map[string]any{"event_type": "fx.rate.published", "occurred_at": r[0], "dedup_key": "nbp-" + r[0],
			"payload": map[string]any{"code": "EUR/PLN/" + r[0], "base": "EUR", "quote": "PLN", "date": r[0], "rate": r[1]}}, nil); code != 200 {
			t.Fatalf("rate %s: %d", r[0], code)
		}
	}
	accounts := map[string]string{}
	for _, code := range []string{"130", "201", "700"} {
		ids, err := w.s.FindObjectIDsByField(w.ctx, "account", "code", code)
		if err != nil || len(ids) != 1 {
			t.Fatalf("account %s: %v %v", code, ids, err)
		}
		accounts[code] = ids[0]
	}
	line := func(person, company, key, date, account, side, amount, currency string) {
		t.Helper()
		var out map[string]any
		if code := w.call(person, company, "POST", "/api/events", map[string]any{"event_type": "test.posted", "occurred_at": date, "dedup_key": key,
			"payload": map[string]any{"entry": key, "line": 1, "account": accounts[account], "side": side, "amount": amount, "currency": currency, "date": date}}, &out); code != 200 || out["booked"] != float64(1) {
			t.Fatalf("%s: %d %v", key, code, out)
		}
	}
	entry := func(person, company, no, date, debit, credit, amount, currency string) {
		t.Helper()
		line(person, company, no+"-1", date, debit, "debit", amount, currency)
		line(person, company, no+"-2", date, credit, "credit", amount, currency)
	}
	entry("kk", a, "FV/12", "2025-12-10", "201", "700", "850.00", "PLN")
	entry("kk", a, "WB/5", "2026-05-10", "130", "201", "425.00", "PLN")
	entry("kk", a, "FV/9", "2026-09-05", "201", "700", "170.00", "PLN")
	entry("kk", a, "WB/10", "2026-10-02", "130", "201", "42.50", "PLN")
	entry("olga", b, "INV/9", "2026-09-20", "201", "700", "100.00", "EUR")
	line("olga", b, "ODD/9", "2026-09-25", "130", "debit", "10.00", "EUR")

	// read by one who sees the group: kk let into Beta as well, acting for neither
	w.mustDoor("", b, "grant_membership", map[string]any{"person": w.personID("kk"), "company": b, "role": "owner", "from": "2026-01-01", "granted_by": "operator"})
	var gtb struct {
		Columns []struct {
			Field string `json:"field"`
		} `json:"columns"`
		Rows   [][]struct{ V string } `json:"rows"`
		Of     [][]string             `json:"row_objects"`
		Period core.Period            `json:"period"`
	}
	if code := w.call("kk", "", "GET", "/api/views/group-trial-balance?from=2026-09-01&to=2026-09-30", nil, &gtb); code != 200 {
		t.Fatalf("the group trial balance: %d", code)
	}
	var fields []string
	for _, c := range gtb.Columns {
		fields = append(fields, c.Field)
	}
	if want := []string{"code", "account", "opening_debit", "opening_credit", "debit", "credit", "ytd_debit", "ytd_credit", "closing_debit", "closing_credit"}; !reflect.DeepEqual(fields, want) {
		t.Fatalf("columns %v, want %v", fields, want)
	}
	var got [][]string
	for _, r := range gtb.Rows {
		var row []string
		for i, c := range r {
			if i != 1 { // the account's name is the chart's
				row = append(row, c.V)
			}
		}
		got = append(got, row)
	}
	want := [][]string{
		{"130", "0.00", "0.00", "10.00", "0.00", "110.00", "0.00", "110.00", "0.00"},
		{"201", "200.00", "0.00", "140.00", "0.00", "140.00", "100.00", "240.00", "0.00"},
		{"700", "0.00", "200.00", "0.00", "140.00", "0.00", "140.00", "0.00", "340.00"},
		{"CTA", "0.00", "0.00", "0.00", "10.00", "0.00", "10.00", "0.00", "10.00"},
	}
	if !reflect.DeepEqual(got, want) || gtb.Period != (core.Period{From: "2026-09-01", To: "2026-09-30", YearFrom: "2026-01-01"}) {
		t.Fatalf("the group's September in EUR:\n got %v %+v\nwant %v", got, gtb.Period, want)
	}
	if !reflect.DeepEqual(gtb.Of, [][]string{{accounts["130"]}, {accounts["201"]}, {accounts["700"]}, {}}) {
		t.Fatalf("each row names its account, the CTA none: %v", gtb.Of)
	}
	// October at October's rate: December's 850,00 PLN opens at 207,32
	if code := w.call("kk", "", "GET", "/api/views/group-trial-balance?from=2026-10-01&to=2026-10-31", nil, &gtb); code != 200 || len(gtb.Rows) < 2 || gtb.Rows[1][0].V != "201" || gtb.Rows[1][2].V != "207.32" {
		t.Fatalf("October's rate: %d %v", code, gtb.Rows)
	}
}
