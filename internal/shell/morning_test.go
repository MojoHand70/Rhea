package shell_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"rhea/internal/adapter"
	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

// TestTheMorningByRole (KK 2026-10-17: "the CEO wants something different
// than the chief accountant"; "VAT is a necessary evil, not the centre").
// The Poland pack's morning leads with the result — this month so far, last
// month, since January, before tax — then money, the tax deadlines in one
// line and what waits; the accountant's morning leads with what is open.
// The deadlines move off weekends and holidays the way the Ordynacja moves
// them, and an amount the books do not compute is said to be missing, never
// guessed. A person hears the morning for their role; anyone else the one
// that names no role.
func TestTheMorningByRole(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	for _, f := range []string{"finance.json", "finance_v2.json", "finance_v3.json", "finance_v4.json",
		"finance_v5.json", "finance_v6.json", "finance_v7.json", "finance_v8.json"} {
		seedFromFile(t, s, f)
	}
	x := &exec.Executor{Store: s}
	day := "2026-10-17"
	srv := &shell.Server{Store: s, Exec: x, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb"),
		Clock: adapter.Clock{Now: func() time.Time { d, _ := time.Parse("2006-01-02", day); return d.Add(9 * time.Hour) }}}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	fetch := func(path string) (int, []byte) {
		t.Helper()
		res, err := ts.Client().Get(ts.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		var b bytes.Buffer
		b.ReadFrom(res.Body)
		return res.StatusCode, b.Bytes()
	}
	get := func(path string, out any) {
		t.Helper()
		code, b := fetch(path)
		if code != 200 {
			t.Fatalf("GET %s: %d %s", path, code, b)
		}
		if err := json.Unmarshal(b, out); err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
	}
	for _, p := range []string{"common/people.json", "common/periods.json", "common/documents.json", "pl/pack.json"} {
		sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", p), "test")
		if err != nil {
			t.Fatal(err)
		}
		if _, _, _, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-09-01"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.InsertRuleVersion(ctx, core.Rule{
		ID: "register-company", Status: core.StatusDraft, Priority: 10, EffectiveFrom: "2026-01-01", CreatedBy: "human",
		Description: "register-company",
		Spec: core.RuleSpec{Match: core.Match{EventType: "company.registered"}, Effect: core.Effect{Object: core.ObjectTemplate{Type: "company",
			Fields: map[string]string{"name": "=$.name", "kind": "=$.kind", "vat_id": "=$.vat_id", "country": "=$.country"}}}},
	}); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := x.ApproveRule(ctx, "register-company", "krzysztof", "2026-09-01"); err != nil {
		t.Fatal(err)
	}
	door := func(name string, in map[string]any) {
		t.Helper()
		tr, err := x.TriggerActivity(ctx, name, in, "krzysztof", "")
		if err != nil || tr.Booked != 1 || len(tr.Errors) > 0 {
			t.Fatalf("%s: %+v %v", name, tr, err)
		}
	}
	submit := func(typ, date, key string, payload map[string]any) {
		t.Helper()
		tr, err := x.TriggerActivity(ctx, "submit_event", map[string]any{"event_type": typ, "occurred_at": date, "payload": payload}, "test", key)
		if err != nil || tr.Booked != 1 || len(tr.Errors) > 0 {
			t.Fatalf("%s: %+v %v", key, tr, err)
		}
	}
	submit("company.registered", "2026-09-01", "self", map[string]any{"name": "Sprzedawca Sp. z o.o.", "kind": "self", "vat_id": "5250001111", "country": "PL"})
	submit("company.registered", "2026-09-01", "buyer", map[string]any{"name": "Nabywca S.A.", "kind": "customer", "vat_id": "5260001246", "country": "PL"})
	companies, err := s.ObjectsByType(ctx, "company")
	if err != nil {
		t.Fatal(err)
	}
	selfID := ""
	for _, c := range companies {
		if c.State["kind"] == "self" {
			selfID = c.ID
		}
	}
	door("declare_company_form", map[string]any{"company": selfID, "vat_status": "active", "books": "full", "income_tax": "cit",
		"from": "2026-01-01", "occurred_at": "2026-09-01"})
	inv := func(no, date, net, vat, gross string) map[string]any {
		return map[string]any{"number": no, "issue_date": date, "market": "pl", "intragroup": "no", "seller_nip": "5250001111",
			"buyer_nip": "5260001246", "net": net, "vat": vat, "gross": gross, "vat_rate": "23", "currency": "PLN"}
	}
	submit("sales.invoice.issued", "2026-09-21", "fv-1", inv("FV 1/09/2026", "2026-09-21", "1000.00", "230.00", "1230.00"))
	submit("sales.invoice.issued", "2026-09-24", "fv-2", inv("FV 2/09/2026", "2026-09-24", "500.00", "115.50", "615.50"))
	door("close_period", map[string]any{"period": "2026-09-30", "book": "pl-stat", "occurred_at": "2026-10-05"})
	submit("purchase.invoice.received", "2026-10-02", "fz-1", map[string]any{"number": "FZ/10/2026", "invoice_date": "2026-10-02", "market": "pl",
		"supplier_nip": "5260001246", "buyer_nip": "5250001111", "supplier": "Nabywca S.A.",
		"net": "400.00", "vat": "92.00", "gross": "492.00", "vat_rate": "23", "currency": "PLN"})
	submit("sales.invoice.issued", "2026-10-08", "fv-3", inv("FV 1/10/2026", "2026-10-08", "200.00", "46.00", "246.00"))
	submit("bank.statement.line", "2026-10-10", "wb-1", map[string]any{"date": "2026-10-10", "amount": "100.00", "currency": "PLN", "direction": "credit",
		"market": "pl", "reference": "przelew", "counterparty": "Ktoś Nieznany", "counterparty_account": "PL27114020040000300201355387",
		"account": "PL83101010230000261395100000"})
	for _, p := range [][3]string{{"Jan Kowalski", "jan", "owner"}, {"Anna Nowak", "anna", "accountant"}, {"Ola Wiśniewska", "ola", "sales"}} {
		door("register_person", map[string]any{"name": p[0], "handle": p[1], "role": p[2], "occurred_at": "2026-09-01"})
	}

	type rows struct {
		Rows [][]cell `json:"rows"`
	}
	table := func(view string) string {
		t.Helper()
		var r rows
		get("/api/views/"+view, &r)
		return fmt.Sprint(cellVals(r.Rows))
	}
	// The result first: October so far against September, and since
	// January — before tax, because the books do not compute CIT yet.
	if got, want := table("poranek-wynik"), "[[2026-10 do dziś, miesiąc trwa 200.00 400.00 -200.00] [2026-09 1500.00 0.00 1500.00] [2026 od stycznia 1700.00 400.00 1300.00]]"; got != want {
		t.Fatalf("the result:\n got %s\nwant %s", got, want)
	}
	// Money: what is owed to us as a total and its oldest invoice, not a list.
	var money rows
	get("/api/views/poranek-pieniadze", &money)
	m := map[string][]string{}
	for _, r := range cellVals(money.Rows) {
		m[r[0]] = r
	}
	if r := m["receivables"]; r == nil || r[2] != "2091.50" || r[3] != "3" {
		t.Fatalf("receivables: %v", money.Rows)
	}
	if r := m["oldest_receivable"]; r == nil || r[2] != "1230.00" || r[4] != "26" || r[5] != "Nabywca S.A." {
		t.Fatalf("the oldest receivable: %v", r)
	}
	if r := m["payables"]; r == nil || r[2] != "492.00" || r[3] != "1" {
		t.Fatalf("payables: %v", r)
	}
	// The deadlines: the CIT advance for September on the 20th, computed at
	// the close (no status declared for the year: 19% of 1500.00); VAT for
	// September on the 25th — a Sunday this year, so Monday the 26th — with
	// the closed month's amount.
	cal := func() string { return table("kalendarz-podatkowy") }
	if got, want := cal(), "[[2026-10-20 Zaliczka na CIT 2026-09 285.00 ] [2026-10-26 VAT i JPK_V7M 2026-09 346.00 ]]"; got != want {
		t.Fatalf("the deadlines on %s:\n got %s\nwant %s", day, got, want)
	}
	// What is open: the payment nothing explains; no month waits to close.
	if got, want := table("poranek-otwarte"), "[[unclosed Miesiące do zamknięcia 0  ] [credits Wpłaty bez rozliczenia 1 100.00 ] [debits Wypłaty bez rozliczenia 0 0.00 ] [tasks Zadania otwarte 0  ]]"; got != want {
		t.Fatalf("what is open:\n got %s\nwant %s", got, want)
	}
	// The morning's figures lead with the result too.
	var rano rows
	get("/api/views/rano", &rano)
	if rv := cellVals(rano.Rows); len(rv) < 3 || rv[0][0] != "result_month" || rv[0][2] != "-200.00" || rv[1][0] != "result_year" || rv[1][2] != "1300.00" || rv[2][0] != "bank" {
		t.Fatalf("the morning's figures: %v", rv)
	}

	// December: the 20th is a Sunday, and the 25th, 26th and 27th are
	// Christmas, a Saturday and a Sunday; November is not closed, so neither
	// has an amount yet, and October waits to be closed.
	day = "2026-12-10"
	if got, want := cal(), "[[2026-12-21 Zaliczka na CIT 2026-11  miesiąc jeszcze niezamknięty] [2026-12-28 VAT i JPK_V7M 2026-11  miesiąc jeszcze niezamknięty]]"; got != want {
		t.Fatalf("the deadlines on %s:\n got %s\nwant %s", day, got, want)
	}
	if got := table("poranek-otwarte"); !strings.HasPrefix(got, "[[unclosed Miesiące do zamknięcia 1  2026-10]") {
		t.Fatalf("what is open on %s: %s", day, got)
	}
	// June 2030: Corpus Christi falls on the 20th (Easter 21 April).
	day = "2030-06-01"
	if got := cal(); !strings.HasPrefix(got, "[[2030-06-21 Zaliczka na CIT 2030-05") {
		t.Fatalf("the deadlines on %s: %s", day, got)
	}
	day = "2026-10-17"

	// The two mornings: the owner's names no role, so it is anyone's who has
	// none of their own; the accountant's names hers. Each speaks from the
	// figures above, and the host's laws hold on both.
	srv.Agent = &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
		if strings.Contains(user, "biura rachunkowego") {
			return "Jedna wpłata, 100.00 zł, czeka na wyjaśnienie. VAT za wrzesień 346.00 zł do 26 października; zaliczki na CIT księgi jeszcze nie liczą. Wynik od stycznia 1300.00 zł.", nil
		}
		return "Październik do dziś na minusie, wrzesień dał 1500.00 zł, od stycznia 1300.00 zł przed opodatkowaniem. Klienci są winni 2091.50 zł. VAT 346.00 zł do 26 października. Wpłata 100.00 zł czeka na Twoje słowo.", nil
	}}
	var spoken struct {
		Silent string `json:"silent"`
		Said   string `json:"said"`
		Facts  []struct {
			View string `json:"view"`
		} `json:"facts"`
	}
	get("/api/views/poranek", &spoken)
	if spoken.Said == "" || len(spoken.Facts) != 4 || spoken.Facts[0].View != "poranek-wynik" {
		t.Fatalf("the owner's morning: %+v", spoken)
	}
	spoken.Said, spoken.Facts = "", nil
	get("/api/views/poranek-ksiegowosc", &spoken)
	if spoken.Said == "" || len(spoken.Facts) != 3 || spoken.Facts[0].View != "poranek-otwarte" {
		t.Fatalf("the accountant's morning: %+v", spoken)
	}
	heard := func(handle string) string {
		t.Helper()
		var said []struct {
			View string `json:"view"`
			Text string `json:"text"`
		}
		get("/api/said?n=1&person="+handle, &said)
		if len(said) != 1 {
			t.Fatalf("%s heard %+v", handle, said)
		}
		return said[0].View
	}
	if v := heard("jan"); v != "poranek" {
		t.Fatalf("the owner heard %s", v)
	}
	if v := heard("anna"); v != "poranek-ksiegowosc" {
		t.Fatalf("the accountant heard %s", v)
	}
	if v := heard("ola"); v != "poranek" {
		t.Fatalf("a role with no morning of its own heard %s", v)
	}
	if code, _ := fetch("/api/said?person=nikt"); code != 404 {
		t.Fatalf("no such person: %d", code)
	}
	if code, _ := fetch("/api/said?person=jan&view=poranek"); code != 400 {
		t.Fatalf("a view and a person: %d", code)
	}

	// Live (RHEA_LIVE_HOST=1): the model named by RHEA_MODEL speaks both
	// mornings from these figures, through the kernel's laws; the words and
	// any refusal are printed, not asserted — a measurement, as in
	// TestHostWordsLive.
	if os.Getenv("RHEA_LIVE_HOST") != "" {
		srv.Agent = agent.New()
		for _, v := range []string{"poranek", "poranek-ksiegowosc"} {
			var out struct{ Said, Silent string }
			get("/api/views/"+v, &out)
			t.Logf("%s\n  said: %s\n  refused: %s", v, out.Said, out.Silent)
		}
	}
}
