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
