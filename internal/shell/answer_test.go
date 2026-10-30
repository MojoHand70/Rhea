package shell_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"rhea/internal/adapter"
	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/shell"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

// TestAnswerFromTheBooks (DIRECTION 2026-10-17, "Reporting"; KK: answered at
// once when it only reads): the author writes one SELECT over the
// collection, the kernel runs it locked — giving DuckDB's refusal back once —
// and records what she told the person; a query that would write is refused
// and the books stay as they were.
func TestAnswerFromTheBooks(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	for _, f := range []string{"finance.json", "finance_v2.json", "finance_v3.json", "finance_v4.json",
		"finance_v5.json", "finance_v6.json", "finance_v7.json", "finance_v8.json"} {
		seedFromFile(t, s, f)
	}
	x := &exec.Executor{Store: s}
	var asked []string
	var replies []string
	srv := &shell.Server{Store: s, Exec: x, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb"),
		Clock: adapter.Clock{Now: func() time.Time { return time.Date(2026, 10, 18, 9, 0, 0, 0, time.UTC) }},
		Agent: &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
			asked = append(asked, user)
			if len(replies) == 0 {
				return "", fmt.Errorf("nothing queued for the author")
			}
			r := replies[0]
			replies = replies[1:]
			return r, nil
		}}}
	ts := serve(t, srv.Handler())
	call := func(method, path string, body, out any) int {
		t.Helper()
		var rd *bytes.Reader
		if body != nil {
			b, _ := json.Marshal(body)
			rd = bytes.NewReader(b)
		} else {
			rd = bytes.NewReader(nil)
		}
		req := httptest.NewRequest(method, ts.URL+path, rd)
		req.RequestURI = ""
		res, err := ts.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if out != nil {
			json.NewDecoder(res.Body).Decode(out)
		}
		return res.StatusCode
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
	if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: "register-company", Status: core.StatusDraft, Priority: 10, EffectiveFrom: "2026-01-01", CreatedBy: "human",
		Description: "register-company", Spec: core.RuleSpec{Match: core.Match{EventType: "company.registered"}, Effect: core.Effect{Object: core.ObjectTemplate{Type: "company",
			Fields: map[string]string{"name": "=$.name", "kind": "=$.kind", "vat_id": "=$.vat_id", "country": "=$.country"}}}}}); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := x.ApproveRule(ctx, "register-company", "krzysztof", "2026-09-01"); err != nil {
		t.Fatal(err)
	}
	submit := func(typ, date, key string, payload map[string]any) {
		t.Helper()
		if tr, err := x.TriggerActivity(ctx, "submit_event", map[string]any{"event_type": typ, "occurred_at": date, "payload": payload}, "test", key); err != nil || tr.Booked != 1 {
			t.Fatalf("%s: %+v %v", key, tr, err)
		}
	}
	submit("company.registered", "2026-09-01", "self", map[string]any{"name": "Sprzedawca Sp. z o.o.", "kind": "self", "vat_id": "5250001111", "country": "PL"})
	submit("company.registered", "2026-09-01", "buyer", map[string]any{"name": "Nabywca S.A.", "kind": "customer", "vat_id": "5260001246", "country": "PL"})
	for i, d := range []struct{ date, net, vat, gross string }{{"2026-08-20", "2000.00", "460.00", "2460.00"}, {"2026-09-21", "1000.00", "230.00", "1230.00"}, {"2026-09-24", "500.00", "115.00", "615.00"}} {
		submit("sales.invoice.issued", d.date, "fv-"+string(rune('a'+i)), map[string]any{"number": "FV " + string(rune('1'+i)), "issue_date": d.date, "market": "pl", "intragroup": "no",
			"seller_nip": "5250001111", "buyer_nip": "5260001246", "net": d.net, "vat": d.vat, "gross": d.gross, "vat_rate": "23", "currency": "PLN"})
	}

	// The collection as the author sees it: the invoice in its domain, its
	// buyer a link to the company table, its money exact.
	var coll []agent.CollectionTable
	call("GET", "/api/collection", nil, &coll)
	var inv *agent.CollectionTable
	for i := range coll {
		if coll[i].Type == "sales_invoice" {
			inv = &coll[i]
		}
	}
	if inv == nil {
		t.Fatalf("no sales_invoice in %+v", coll)
	}
	cols := map[string]agent.CollectionColumn{}
	for _, c := range inv.Columns {
		cols[c.Name] = c
	}
	if c := cols["buyer"]; c.Links != inv.Schema+".company" {
		t.Fatalf("the buyer links to %q", c.Links)
	}
	if cols["net"].SQL != "DECIMAL(18,2)" {
		t.Fatalf("net is %q", cols["net"].SQL)
	}
	table := inv.Schema + ".sales_invoice"

	// A broken query is given back with DuckDB's words; the fix is run.
	good := `SELECT c.name AS "Klient", sum(i.net) AS "Netto", string_agg(i.id, ',' ORDER BY i.id) AS _of FROM ` + table + ` i JOIN ` + inv.Schema + `.company c ON c.id = i.buyer GROUP BY c.name ORDER BY 2 DESC`
	replies = []string{
		`{"sql": "SELECT nope FROM ` + table + `", "reading": "x", "visual": {"kind": "table"}}`,
		mustJSON(map[string]any{"sql": good, "reading": "faktury sprzedaży, suma netto według nabywcy", "visual": map[string]any{"kind": "bar", "x": "Klient", "y": []string{"Netto"}, "title": "Sprzedaż według klientów"}}),
	}
	var ans struct {
		EventID int64       `json:"event_id"`
		Reading string      `json:"reading"`
		Visual  core.Visual `json:"visual"`
		Columns []struct{ Name, Type string }
		Rows    [][]string `json:"rows"`
		Of      [][]string `json:"row_objects"`
		Cannot  string     `json:"cannot"`
	}
	if code := call("POST", "/api/answer", map[string]any{"question": "Sprzedaż netto według klientów", "asked_by": "krzysztof"}, &ans); code != 200 {
		t.Fatalf("answer: %d", code)
	}
	// The market's notes travel with the collection: the author reads how Polish books are read.
	if !strings.Contains(asked[0], "How this market's books are read") || !strings.Contains(asked[0], "Wpłaty do ZUS") || !strings.Contains(asked[0], "Wn 229") {
		t.Fatalf("the author was not shown the pack's notes:\n%s", asked[0])
	}
	if len(asked) != 2 || !strings.Contains(asked[1], "DuckDB said") || !strings.Contains(asked[1], "SELECT nope") {
		t.Fatalf("the retry did not carry DuckDB's refusal:\n%s", asked[len(asked)-1])
	}
	if len(ans.Rows) != 1 || ans.Rows[0][0] != "Nabywca S.A." || ans.Rows[0][1] != "3500.00" || len(ans.Of[0]) != 3 || ans.Columns[1].Type != "DECIMAL(38,2)" ||
		ans.Visual.Kind != "bar" || len(ans.Columns) != 2 {
		t.Fatalf("the answer: %+v", ans)
	}
	firstID := ans.EventID
	evs, err := s.EventsOfTypes(ctx, []string{core.EventHostAnswered})
	if err != nil || len(evs) != 1 || !strings.Contains(string(evs[0].Payload), "Nabywca S.A.") || !strings.Contains(string(evs[0].Payload), "suma netto według nabywcy") {
		t.Fatalf("what she told was not recorded: %v %v", evs, err)
	}

	// A query that would write is refused, three times, and the books stay;
	// the failure is a moment of its own, recorded with only its class.
	replies = []string{`{"sql": "DELETE FROM ` + table + `", "reading": "x", "visual": {"kind": "table"}, "lesson": "usuwanie faktur"}`,
		`{"sql": "DROP TABLE ` + table + `", "reading": "x", "visual": {"kind": "table"}, "lesson": "usuwanie faktur"}`,
		`{"sql": "TRUNCATE ` + table + `", "reading": "x", "visual": {"kind": "table"}, "lesson": "usuwanie faktur"}`}
	var failed struct {
		Failed bool   `json:"failed"`
		Error  string `json:"error"`
		Noted  *struct {
			Kind     string `json:"kind"`
			Shared   bool   `json:"shared"`
			KeptHome string `json:"kept_home_because"`
		} `json:"noted"`
	}
	if code := call("POST", "/api/answer", map[string]any{"question": "usuń faktury Nabywcy", "asked_by": "krzysztof"}, &failed); code != 200 || !failed.Failed || failed.Noted == nil || failed.Noted.Kind != "failed" {
		t.Fatalf("a writing query: %d %+v", code, failed)
	}
	moments, _ := s.EventsOfTypes(ctx, []string{core.EventHostUntaught})
	if len(moments) != 1 || !strings.Contains(string(moments[0].Payload), `"detail": "not one SELECT"`) && !strings.Contains(string(moments[0].Payload), `"detail":"not one SELECT"`) ||
		!strings.Contains(string(moments[0].Payload), `usuwanie faktur`) {
		t.Fatalf("the failure was not recorded as such: %v", moments)
	}
	replies = []string{`{"sql": "SELECT count(*) AS n FROM ` + table + `", "reading": "liczba faktur", "visual": {"kind": "kpi", "value": "missing"}}`}
	call("POST", "/api/answer", map[string]any{"question": "ile faktur?", "asked_by": "krzysztof"}, &ans)
	if len(ans.Rows) != 1 || ans.Rows[0][0] != "3" || ans.Visual.Kind != "table" {
		t.Fatalf("after a refused delete, the books: %+v (a misnamed visual shows the rows)", ans)
	}
	// A float and the machine's clock go back to the author with the reason.
	asked = nil
	replies = []string{`{"sql": "SELECT avg(net) AS a FROM ` + table + `", "reading": "średnia", "visual": {"kind": "kpi", "value": "a"}}`,
		`{"sql": "SELECT ROUND(avg(net), 2)::DECIMAL(18,2) AS a FROM ` + table + `", "reading": "średnia", "visual": {"kind": "kpi", "value": "a"}}`}
	call("POST", "/api/answer", map[string]any{"question": "średnia faktura?", "asked_by": "krzysztof"}, &ans)
	if len(asked) != 2 || !strings.Contains(asked[1], "a float") || ans.Rows[0][0] != "1166.67" {
		t.Fatalf("a float was shown: %+v", ans.Rows)
	}
	asked = nil
	replies = []string{`{"sql": "SELECT count(*) AS n FROM ` + table + ` WHERE issue_date > CURRENT_DATE - 30", "reading": "x", "visual": {"kind": "kpi", "value": "n"}}`,
		`{"sql": "SELECT count(*) AS n FROM ` + table + ` WHERE issue_date > DATE '2026-10-18' - 30", "reading": "x", "visual": {"kind": "kpi", "value": "n"}}`}
	call("POST", "/api/answer", map[string]any{"question": "ile faktur w ostatnich 30 dniach?", "asked_by": "krzysztof"}, &ans)
	if len(asked) != 2 || !strings.Contains(asked[1], "machine's clock") || ans.Rows[0][0] != "2" {
		t.Fatalf("the machine's clock was read: %+v", ans.Rows)
	}
	replies = []string{`{"cannot": "w księgach nie ma magazynu", "lesson": "stany magazynowe"}`}
	call("POST", "/api/answer", map[string]any{"question": "ile mam w magazynie?", "asked_by": "krzysztof"}, &ans)
	if ans.Cannot == "" {
		t.Fatalf("cannot: %+v", ans)
	}
	// The nearest thing shown for what she was never taught: the answer, and the moment.
	replies = []string{`{"sql": "SELECT sum(net) AS n FROM ` + table + `", "reading": "sprzedaż do dziś", "visual": {"kind": "kpi", "value": "n"}, "lesson": "prognoza"}`}
	call("POST", "/api/answer", map[string]any{"question": "jaki będę miał dochód w przyszłych miesiącach?", "asked_by": "krzysztof", "missing": "prognoza dochodu na kolejne miesiące"}, &ans)
	moments, _ = s.EventsOfTypes(ctx, []string{core.EventHostUntaught})
	kinds := map[string]bool{}
	for _, m := range moments {
		var p struct{ Kind, Missing string }
		json.Unmarshal(m.Payload, &p)
		kinds[p.Kind+":"+p.Missing] = true
	}
	if len(ans.Rows) != 1 || !kinds["cannot:stany magazynowe"] || !kinds["untaught:prognoza dochodu na kolejne miesiące"] || !kinds["failed:usuwanie faktur"] {
		t.Fatalf("the moments: %v (rows %v)", kinds, ans.Rows)
	}

	// Kept (KK: what is kept goes through the gate): the answer becomes a
	// report waiting for approval, its rows on the card; approved, it shows
	// them to everyone, "why?" and all.
	var kept struct {
		ID       string `json:"bundle_id"`
		Status   string `json:"status"`
		Question string `json:"question"`
	}
	if code := call("POST", fmt.Sprintf("/api/answers/%d/keep", firstID), nil, &kept); code != 200 || kept.Status != "draft" || kept.Question != "Sprzedaż netto według klientów" {
		t.Fatalf("keep: %d %+v", code, kept)
	}
	type preview []struct {
		Title string
		Rows  [][]struct {
			V string `json:"v"`
		}
		Total int
		Error string
	}
	var pv preview
	call("GET", "/api/bundles/"+kept.ID+"/preview", nil, &pv)
	if len(pv) != 1 || pv[0].Error != "" || pv[0].Total != 1 || pv[0].Rows[0][0].V != "Nabywca S.A." {
		t.Fatalf("the kept report on the card: %+v", pv)
	}
	if _, _, _, err := x.ApproveBundle(ctx, kept.ID, "krzysztof", "2026-10-18"); err != nil {
		t.Fatal(err)
	}
	type report struct {
		Components []struct {
			Title, Reading, Error string
			Visual                core.Visual
			Rows                  [][]string
			Of                    [][]string `json:"row_objects"`
		}
	}
	var rep report
	call("GET", fmt.Sprintf("/api/views/answer-%d", firstID), nil, &rep)
	if len(rep.Components) != 1 || rep.Components[0].Rows[0][1] != "3500.00" || rep.Components[0].Visual.Kind != "bar" || len(rep.Components[0].Of[0]) != 3 {
		t.Fatalf("the kept report: %+v", rep)
	}
	// Approved, the kept report is an App a face places and shows (KK
	// 2026-10-30: a report approved must be reachable): one tab, its view;
	// with the books' reports; on the desk and the phone.
	type servedApp struct {
		App   string `json:"app"`
		Title string `json:"title"`
		Layer string `json:"layer"`
		Tabs  []struct {
			View string `json:"view"`
		} `json:"tabs"`
		Place *struct {
			Activity string `json:"activity"`
			Section  string `json:"section"`
		} `json:"place"`
		Shows    []string `json:"shows"`
		Domain   string   `json:"domain"`
		Function string   `json:"function"`
	}
	var apps []servedApp
	call("GET", "/api/apps", nil, &apps)
	keptID := fmt.Sprintf("answer-%d", firstID)
	i := slices.IndexFunc(apps, func(a servedApp) bool { return a.App == keptID })
	if i < 0 {
		t.Fatalf("the kept report is no App: %+v", apps)
	}
	if a := apps[i]; a.Title != "Sprzedaż według klientów" || a.Layer != "business" || len(a.Tabs) != 1 || a.Tabs[0].View != keptID ||
		a.Place == nil || a.Place.Activity != "books" || a.Place.Section != "reports" || strings.Join(a.Shows, ",") != "desk,phone" || a.Domain != "reports" || a.Function != "report" {
		t.Fatalf("the kept report's App: %+v", a)
	}

	// A report kept from a question about one month takes the month as its
	// parameter (KK 2026-10-30: "it's only the accounting period which
	// changes"): asked for again, for August, it is the same report computed
	// for August — the author names it, writes no query, and nothing waits
	// to be kept a second time.
	asked = nil
	replies = []string{`{"sql": "SELECT sum(i.net) AS \"Netto\", string_agg(i.id, ',' ORDER BY i.id) AS _of FROM ` + table + ` i WHERE strftime(i.issue_date, '%Y-%m') = '2026-09'", "reading": "sprzedaż netto we wrześniu", "visual": {"kind": "kpi", "value": "Netto", "title": "Sprzedaż we wrześniu"}}`}
	call("POST", "/api/answer", map[string]any{"question": "sprzedaż netto we wrześniu", "asked_by": "krzysztof"}, &ans)
	if len(ans.Rows) != 1 || ans.Rows[0][0] != "1500.00" {
		t.Fatalf("September's sales: %+v", ans.Rows)
	}
	if !strings.Contains(asked[0], "Kept reports") || !strings.Contains(asked[0], keptID+" „Sprzedaż według klientów”") {
		t.Fatalf("the author is told of the kept reports:\n%s", asked[0])
	}
	septID := ans.EventID
	if code := call("POST", fmt.Sprintf("/api/answers/%d/keep", septID), nil, &kept); code != 200 {
		t.Fatalf("keep September's: %d", code)
	}
	if _, _, _, err := x.ApproveBundle(ctx, kept.ID, "krzysztof", "2026-10-18"); err != nil {
		t.Fatal(err)
	}
	monthly := fmt.Sprintf("answer-%d", septID)
	var page struct {
		Components []struct {
			Rows [][]string
		}
		Period *core.Period `json:"period"`
	}
	call("GET", "/api/views/"+monthly+"?from=2026-08-01&to=2026-08-31", nil, &page)
	if page.Period == nil || page.Period.From != "2026-08-01" || len(page.Components) != 1 || len(page.Components[0].Rows) != 1 || page.Components[0].Rows[0][0] != "2000.00" {
		t.Fatalf("the kept report for August: %+v", page)
	}
	call("GET", "/api/views/"+monthly, nil, &page)
	if page.Period == nil || page.Period.From != "2026-10-01" || page.Period.To != "2026-10-31" {
		t.Fatalf("asked for no period, the first month not closed: %+v", page.Period)
	}
	if code := call("GET", "/api/views/"+monthly+"?from=2026-08-31&to=2026-08-01", nil, nil); code != 400 {
		t.Fatalf("a period whose days are out of order: %d", code)
	}
	asked = nil
	replies = []string{`{"report": "` + monthly + `", "from": "2026-08-01", "to": "2026-08-31", "reading": "sprzedaż netto za sierpień, z zachowanego raportu"}`}
	var again struct {
		EventID int64        `json:"event_id"`
		Report  string       `json:"report"`
		Title   string       `json:"title"`
		Period  *core.Period `json:"period"`
		Reading string       `json:"reading"`
		Rows    [][]string   `json:"rows"`
		Visual  core.Visual  `json:"visual"`
	}
	if code := call("POST", "/api/answer", map[string]any{"question": "a za sierpień?", "asked_by": "krzysztof"}, &again); code != 200 {
		t.Fatalf("answer by the kept report: %d", code)
	}
	if again.Report != monthly || again.Title != "Sprzedaż we wrześniu" || again.Period == nil || again.Period.To != "2026-08-31" || len(again.Rows) != 1 || again.Rows[0][0] != "2000.00" || again.Visual.Kind != "kpi" {
		t.Fatalf("the kept report, for August: %+v", again)
	}
	if !strings.Contains(asked[0], monthly+" „Sprzedaż we wrześniu” - sprzedaż netto we wrześniu - parameters: period (period, default open-month)") {
		t.Fatalf("the author is told the report's parameters:\n%s", asked[0])
	}
	evs, _ = s.EventsOfTypes(ctx, []string{core.EventHostAnswered})
	last := evs[len(evs)-1]
	if last.ID != again.EventID || !strings.Contains(string(last.Payload), `"report":"`+monthly+`"`) && !strings.Contains(string(last.Payload), `"report": "`+monthly+`"`) ||
		!strings.Contains(string(last.Payload), "2026-08-31") {
		t.Fatalf("what she told names the report and the period: %s", last.Payload)
	}
	var twice struct {
		Error  string `json:"error"`
		Report string `json:"report"`
	}
	if code := call("POST", fmt.Sprintf("/api/answers/%d/keep", again.EventID), nil, &twice); code != 409 || twice.Report != monthly {
		t.Fatalf("keeping what the kept report answered: %d %+v", code, twice)
	}

	// A report's parameters (KK 2026-10-30, the context selector): a firm, a
	// period, a choice, each put in by the kernel after checking it, each
	// served with the value the report was computed for; a value the books
	// do not hold is refused, never put near a query.
	companies, _ := s.ObjectsByType(ctx, "company")
	var buyerID, sellerID string
	for _, c := range companies {
		switch c.State["vat_id"] {
		case "5260001246":
			buyerID = c.ID
		case "5250001111":
			sellerID = c.ID
		}
	}
	paramSpec, _ := json.Marshal(core.ReportSpec{
		Components: []core.ReportComponent{{Title: "Sprzedaż", Reading: "sprzedaż netto do firmy w okresie, w walucie",
			SQL:    `SELECT COALESCE(sum(i.net), 0) AS "Netto", string_agg(i.id, ',' ORDER BY i.id) AS _of FROM ` + table + ` i WHERE i.buyer = {{firma}} AND i.issue_date BETWEEN {{period_from}} AND {{period_to}} AND i.currency = {{waluta}}`,
			Visual: core.Visual{Kind: "kpi", Value: "Netto", Title: "Sprzedaż do firmy"}}},
		Parameters: []core.Parameter{{Name: "firma", Kind: "firm", Label: "Klient", Default: buyerID}, {Name: "period", Kind: "period", Label: "Okres", Default: "2026-09"},
			{Name: "waluta", Kind: "choice", Label: "Waluta", Values: []string{"PLN", "EUR"}, Default: "PLN"}}})
	if err := store.InsertViewDefRow(ctx, s.Pool, core.ViewDef{ID: "sprzedaz-do-firmy", Version: 1, Notion: "report", Title: "Sprzedaż do firmy", Domain: "reports", Function: "report", Spec: paramSpec}, core.StatusActive); err != nil {
		t.Fatal(err)
	}
	var served struct {
		Components []struct{ Rows [][]string }
		Parameters []struct {
			Name  string `json:"name"`
			Kind  string `json:"kind"`
			Value any    `json:"value"`
		} `json:"parameters"`
		Period *core.Period `json:"period"`
	}
	call("GET", "/api/views/sprzedaz-do-firmy", nil, &served)
	if len(served.Components) != 1 || served.Components[0].Rows[0][0] != "1500.00" || len(served.Parameters) != 3 || served.Parameters[0].Value != buyerID ||
		served.Period == nil || served.Period.From != "2026-09-01" || served.Parameters[2].Value != "PLN" {
		t.Fatalf("the report on its defaults: %+v", served)
	}
	call("GET", "/api/views/sprzedaz-do-firmy?period=2026-08", nil, &served)
	if served.Components[0].Rows[0][0] != "2000.00" || served.Period.To != "2026-08-31" {
		t.Fatalf("for August: %+v", served)
	}
	call("GET", "/api/views/sprzedaz-do-firmy?firma="+sellerID+"&period=2026-01-01..2026-12-31", nil, &served)
	if served.Components[0].Rows[0][0] != "0.00" || served.Parameters[0].Value != sellerID || served.Period.From != "2026-01-01" {
		t.Fatalf("for another firm, the year: %+v", served)
	}
	for _, bad := range []string{"?firma=company-nope", "?firma=" + url.QueryEscape("x' OR '1'='1"), "?waluta=USD", "?period=wrzesień", "?period=2026-09-30..2026-09-01"} {
		if code := call("GET", "/api/views/sprzedaz-do-firmy"+bad, nil, nil); code != 400 {
			t.Fatalf("%s: %d, a value the books do not hold is refused", bad, code)
		}
	}

	// Rhea proposes the parameters when she creates a report (KK 2026-10-30):
	// asked "sales to Nabywca in September", kept, the author makes the firm
	// and the period its parameters, the firm named as the question wrote it
	// and found in the books; the card shows them; approved, the report
	// takes them, and asked "and for August?" the author names them.
	replies = []string{`{"sql": "SELECT COALESCE(sum(i.net), 0) AS \"Netto\", string_agg(i.id, ',' ORDER BY i.id) AS _of FROM ` + table + ` i JOIN ` + inv.Schema + `.company c ON c.id = i.buyer WHERE c.name = 'Nabywca S.A.' AND strftime(i.issue_date, '%Y-%m') = '2026-09'", "reading": "sprzedaż netto do Nabywcy we wrześniu", "visual": {"kind": "kpi", "value": "Netto", "title": "Sprzedaż do Nabywcy"}}`,
		`{"sql": "SELECT COALESCE(sum(i.net), 0) AS \"Netto\", string_agg(i.id, ',' ORDER BY i.id) AS _of FROM ` + table + ` i WHERE i.buyer = {{firma}} AND i.issue_date BETWEEN {{period_from}} AND {{period_to}}", "parameters": [{"name": "firma", "kind": "firm", "label": "Klient", "default": "Nabywca S.A."}, {"name": "period", "kind": "period", "label": "Okres", "default": "2026-09"}]}`}
	call("POST", "/api/answer", map[string]any{"question": "sprzedaż do Nabywcy we wrześniu", "asked_by": "krzysztof"}, &ans)
	if len(ans.Rows) != 1 || ans.Rows[0][0] != "1500.00" {
		t.Fatalf("sales to Nabywca in September: %+v", ans.Rows)
	}
	asked = nil
	if code := call("POST", fmt.Sprintf("/api/answers/%d/keep", ans.EventID), nil, &kept); code != 200 {
		t.Fatalf("keep: %d", code)
	}
	if len(asked) != 1 || !strings.Contains(asked[0], "Question: sprzedaż do Nabywcy we wrześniu") || !strings.Contains(asked[0], "c.name = 'Nabywca S.A.'") {
		t.Fatalf("the author was asked for the parameters once, with the question and the query:\n%v", asked)
	}
	var cards []struct {
		Title      string
		Parameters []struct {
			Name, Kind, Label, Default string
			Value                      any
		}
		Rows [][]struct{ V string }
	}
	call("GET", "/api/bundles/"+kept.ID+"/preview", nil, &cards)
	if len(cards) != 1 || len(cards[0].Parameters) != 2 || cards[0].Parameters[0].Name != "firma" || cards[0].Parameters[0].Default != buyerID || cards[0].Parameters[0].Label != "Klient" ||
		cards[0].Parameters[1].Name != "period" || cards[0].Rows[0][0].V != "1500.00" {
		t.Fatalf("the card shows the parameters, the firm found by its name: %+v", cards)
	}
	if _, _, _, err := x.ApproveBundle(ctx, kept.ID, "krzysztof", "2026-10-18"); err != nil {
		t.Fatal(err)
	}
	byNabywca := fmt.Sprintf("answer-%d", ans.EventID)
	call("GET", "/api/views/"+byNabywca+"?period=2026-08", nil, &served)
	if served.Components[0].Rows[0][0] != "2000.00" || served.Parameters[0].Value != buyerID {
		t.Fatalf("the kept report for August, the firm its default: %+v", served)
	}
	asked = nil
	replies = []string{`{"report": "` + byNabywca + `", "parameters": {"period": "2026-08", "firma": "Nabywca S.A."}, "reading": "sprzedaż do Nabywcy za sierpień"}`}
	if code := call("POST", "/api/answer", map[string]any{"question": "a za sierpień?", "asked_by": "krzysztof"}, &again); code != 200 || again.Report != byNabywca || len(again.Rows) != 1 || again.Rows[0][0] != "2000.00" {
		t.Fatalf("the kept report answers for August by its parameters: %d %+v", code, again)
	}
	if !strings.Contains(asked[0], byNabywca+" „Sprzedaż do Nabywcy” - sprzedaż netto do Nabywcy we wrześniu - parameters: firma (firm, Klient, default "+buyerID+"); period (period, Okres, default 2026-09)") {
		t.Fatalf("the author is told the parameters:\n%s", asked[0])
	}
	// a proposal that does not hold — a firm the books have not — falls back to the period alone
	replies = []string{`{"sql": "SELECT COALESCE(sum(i.net), 0) AS \"Netto\", string_agg(i.id, ',' ORDER BY i.id) AS _of FROM ` + table + ` i WHERE strftime(i.issue_date, '%Y-%m') = '2026-08'", "reading": "sprzedaż netto w sierpniu", "visual": {"kind": "kpi", "value": "Netto"}}`,
		`{"sql": "SELECT 1 FROM x WHERE f = {{firma}}", "parameters": [{"name": "firma", "kind": "firm", "default": "Nikt Taki"}]}`}
	call("POST", "/api/answer", map[string]any{"question": "sprzedaż w sierpniu", "asked_by": "krzysztof"}, &ans)
	call("POST", fmt.Sprintf("/api/answers/%d/keep", ans.EventID), nil, &kept)
	call("GET", "/api/bundles/"+kept.ID+"/preview", nil, &cards)
	if len(cards) != 1 || len(cards[0].Parameters) != 1 || cards[0].Parameters[0].Name != "period" || cards[0].Parameters[0].Kind != "period" {
		t.Fatalf("a proposal that does not hold leaves the period alone: %+v", cards)
	}

	// A dashboard drafted in words: the author sees the collection; on the
	// page, a component whose query fails says so and the others still show.
	asked = nil
	replies = []string{mustJSON(map[string]any{"bundle_id": "sales-page", "description": "Sprzedaż w skrócie.",
		"warrant": map[string]any{"basis": "client", "citations": []string{"the owner asked"}, "scope": "this company"},
		"view_defs": []any{map[string]any{"view_id": "sales-page", "notion": "report", "title": "Sprzedaż w skrócie", "domain": "reports", "function": "report",
			"spec": map[string]any{"components": []any{
				map[string]any{"title": "Netto razem", "reading": "netto wszystkich faktur sprzedaży", "sql": `SELECT sum(net) AS "Netto" FROM ` + table, "visual": map[string]any{"kind": "kpi", "value": "Netto"}},
				map[string]any{"title": "Zepsute", "reading": "x", "sql": "SELECT nope FROM " + table, "visual": map[string]any{"kind": "table"}}}}}}})}
	var drafted struct {
		ID string `json:"bundle_id"`
	}
	if code := call("POST", "/api/bundles/draft", map[string]any{"intent": "Zrób mi pulpit sprzedaży"}, &drafted); code != 200 || drafted.ID != "sales-page" {
		t.Fatalf("the dashboard draft: %d %+v", code, drafted)
	}
	if len(asked) != 1 || !strings.Contains(asked[0], table) || !strings.Contains(asked[0], "Today: 2026-10-18") {
		t.Fatalf("the author did not see the collection:\n%s", asked[0])
	}
	if _, _, _, err := x.ApproveBundle(ctx, drafted.ID, "krzysztof", "2026-10-18"); err != nil {
		t.Fatal(err)
	}
	rep = report{}
	call("GET", "/api/views/sales-page", nil, &rep)
	if len(rep.Components) != 2 || rep.Components[0].Rows[0][0] != "3500.00" || !strings.Contains(rep.Components[1].Error, "nope") {
		t.Fatalf("the dashboard: %+v", rep)
	}
}

func mustJSON(v any) string { b, _ := json.Marshal(v); return string(b) }
