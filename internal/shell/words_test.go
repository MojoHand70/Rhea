package shell_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

// TestThePackSpeaks: the face computes no figure and names nothing on its
// own (DIRECTION 2026-10-16, step 3). The Poland pack names every type,
// field and value the kernel holds and the shape each type takes; it lays
// the VAT settlement out as a declaration without computing; and the
// figures the face once summed itself — the morning's, profit, receivables
// by age on the kernel's day — come from the kernel, each row saying what
// it is made of.
func TestThePackSpeaks(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	for _, f := range []string{"finance.json", "finance_v2.json", "finance_v3.json", "finance_v4.json",
		"finance_v5.json", "finance_v6.json", "finance_v7.json", "finance_v8.json"} {
		seedFromFile(t, s, f)
	}
	x := &exec.Executor{Store: s}
	day := "2026-10-16"
	srv := &shell.Server{Store: s, Exec: x, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb"),
		Clock: adapter.Clock{Now: func() time.Time { d, _ := time.Parse("2006-01-02", day); return d.Add(9 * time.Hour) }}}
	ts := serve(t, srv.Handler())
	get := func(path string, out any) {
		t.Helper()
		res, err := ts.Client().Get(ts.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		var b bytes.Buffer
		b.ReadFrom(res.Body)
		if res.StatusCode != 200 {
			t.Fatalf("GET %s: %d %s", path, res.StatusCode, b.String())
		}
		if err := json.Unmarshal(b.Bytes(), out); err != nil {
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
	submit := func(typ, date, key string, payload map[string]any) {
		t.Helper()
		tr, err := x.TriggerActivity(ctx, "submit_event", map[string]any{"event_type": typ, "occurred_at": date, "payload": payload}, "test", key)
		if err != nil || tr.Booked != 1 || len(tr.Errors) > 0 {
			t.Fatalf("%s: %+v %v", key, tr, err)
		}
	}
	submit("company.registered", "2026-09-01", "self", map[string]any{"name": "Sprzedawca Sp. z o.o.", "kind": "self", "vat_id": "5250001111", "country": "PL"})
	submit("company.registered", "2026-09-01", "buyer", map[string]any{"name": "Nabywca S.A.", "kind": "customer", "vat_id": "5260001246", "country": "PL"})
	inv := func(no, date, net, vat, gross string) map[string]any {
		return map[string]any{"number": no, "issue_date": date, "market": "pl", "intragroup": "no", "seller_nip": "5250001111",
			"buyer_nip": "5260001246", "net": net, "vat": vat, "gross": gross, "vat_rate": "23", "currency": "PLN"}
	}
	submit("sales.invoice.issued", "2026-09-21", "fv-1", inv("FV 1/09/2026", "2026-09-21", "1000.00", "230.00", "1230.00"))
	submit("sales.invoice.issued", "2026-09-24", "fv-2", inv("FV 2/09/2026", "2026-09-24", "500.00", "115.50", "615.50"))
	if tr, err := x.TriggerActivity(ctx, "close_period", map[string]any{"period": "2026-09-30", "book": "pl-stat", "occurred_at": "2026-10-05"}, "krzysztof", ""); err != nil || tr.Booked != 1 {
		t.Fatalf("close September: %+v %v", tr, err)
	}

	// The words: every type, every field and every value the kernel holds
	// has a Polish name, and every type a shape.
	var words struct {
		Words core.WordsSpec `json:"words"`
	}
	get("/api/views/slownik", &words)
	w := words.Words
	if w.Lang != "pl" || w.Check() != nil {
		t.Fatalf("words: %s %v", w.Lang, w.Check())
	}
	var types []core.ObjectType
	get("/api/types", &types)
	var missing []string
	for _, ty := range types {
		tw, ok := w.Types[ty.Name]
		if !ok || tw.Shape == "" {
			missing = append(missing, "type "+ty.Name)
		}
		if tw.Kind == "" { // what it is — a document, an entry, master data: the desk files its lists by it (2026-10-22)
			missing = append(missing, "the kind of "+ty.Name)
		}
		for _, f := range ty.Fields {
			if w.Fields[ty.Name+"."+f.Name] == "" && w.Fields[f.Name] == "" {
				missing = append(missing, "field "+ty.Name+"."+f.Name)
			}
			for _, v := range f.Values {
				if w.Values[ty.Name+"."+f.Name][v] == "" && w.Values[f.Name][v] == "" {
					missing = append(missing, "value "+ty.Name+"."+f.Name+"="+v)
				}
			}
		}
	}
	// Every view a person opens is titled in the pack's language: the pack's own,
	// or another pack's (the base seeds, the common packs, all in English) with
	// its Polish title in the words. The kernel's own lists of a type are named
	// by the type's words.
	// the pack's own views, read through its Apps (written as Apps 2026-10-26)
	own, _, err := pack.ReadFile(filepath.Join("..", "..", "packs", "pl", "pack.json"))
	if err != nil {
		t.Fatalf("the Poland pack: %v", err)
	}
	ours := map[string]bool{}
	for _, v := range own.ViewDefs {
		ours[v.ID] = true
	}
	var nav []struct {
		Functions []struct {
			Views []struct {
				ID     string `json:"view_id"`
				Notion string `json:"notion"`
			} `json:"views"`
		} `json:"functions"`
	}
	get("/api/nav", &nav)
	for _, d := range nav {
		for _, f := range d.Functions {
			for _, v := range f.Views {
				if v.Notion == "detail" || ours[v.ID] || strings.HasPrefix(v.ID, "derived:") || w.Views[v.ID] != "" {
					continue
				}
				missing = append(missing, "the title of view "+v.ID)
			}
		}
	}
	// and every App she is served: the pack's own titled in Polish, another
	// pack's (the core's, the common packs', in English) named in the words
	_, ownApps, err := pack.ReadFile(filepath.Join("..", "..", "packs", "pl", "pack.json"))
	if err != nil {
		t.Fatal(err)
	}
	oursApp := map[string]bool{}
	for _, a := range ownApps {
		oursApp[a.ID] = true
	}
	var apps []struct {
		ID string `json:"app"`
	}
	get("/api/apps", &apps)
	if len(apps) == 0 {
		t.Fatal("no App is served")
	}
	for _, a := range apps {
		if !oursApp[a.ID] && w.Apps[a.ID] == "" {
			missing = append(missing, "the title of App "+a.ID)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("the pack leaves unnamed: %s", strings.Join(missing, ", "))
	}

	// The VAT settlement as a declaration: the settlement's own figures,
	// ordered and signed; a carry of zero is not shown.
	var vs struct {
		Objects []core.Object `json:"objects"`
	}
	get("/api/objects?type=vat_settlement", &vs)
	if len(vs.Objects) != 1 {
		t.Fatalf("settlements: %+v", vs.Objects)
	}
	var st struct {
		Sections []struct {
			Title string `json:"title"`
			Rows  []struct {
				Op, Field, Note, V string
			} `json:"rows"`
		} `json:"sections"`
	}
	get("/api/views/deklaracja-vat?object_id="+vs.Objects[0].ID, &st)
	decl := st.Sections[0].Rows
	if len(st.Sections) != 2 || len(decl) != 4 || decl[0].Field != "output_vat" || decl[0].V != "345.50" ||
		decl[1].Op != "-" || decl[3].Op != "=" || decl[3].Field != "vat_due" || decl[3].V != "346.00" || !strings.Contains(decl[3].Note, "pełnych złotych") {
		t.Fatalf("the declaration: %+v", st)
	}

	// The morning's figures, from the kernel: what is owed to us and by
	// whom, the last VAT, and what was done — each row naming its objects.
	type rows struct {
		Columns []struct{ Field string } `json:"columns"`
		Rows    [][]cell                 `json:"rows"`
		Of      [][]string               `json:"row_objects"`
	}
	var rano rows
	get("/api/views/rano", &rano)
	byKey := map[string][]string{}
	ofKey := map[string][]string{}
	for i, r := range cellVals(rano.Rows) {
		byKey[r[0]] = r
		ofKey[r[0]] = rano.Of[i]
	}
	if r := byKey["receivables"]; r == nil || r[2] != "1845.50" || r[3] != "2" || len(ofKey["receivables"]) != 2 {
		t.Fatalf("receivables: %v %v", r, ofKey["receivables"])
	}
	if r := byKey["vat_due"]; r == nil || r[2] != "346.00" || r[4] != "2026-09" || ofKey["vat_due"][0] != vs.Objects[0].ID {
		t.Fatalf("the last VAT: %v", r)
	}
	if r := byKey["closed"]; r == nil || r[2] != "" || r[3] != "1" {
		t.Fatalf("months closed: %v", r)
	}
	if r := byKey["bank"]; r == nil || r[2] != "0.00" {
		t.Fatalf("the bank, before any line: %v", r)
	}

	// Profit since January: revenue less costs, the accounts each row reads.
	var wynik rows
	get("/api/views/wynik", &wynik)
	wv := cellVals(wynik.Rows)
	if len(wv) != 3 || wv[0][1] != "Przychody" || wv[0][2] != "1500.00" || wv[1][0] != "-" || wv[2][0] != "=" || wv[2][2] != "1500.00" || len(wynik.Of[0]) != 1 {
		t.Fatalf("profit: %v %v", wv, wynik.Of)
	}

	// Receivables by age, on the kernel's day: on 16 October both invoices
	// are under 30 days; the parts make the total.
	var age rows
	get("/api/views/naleznosci-wiek", &age)
	av := cellVals(age.Rows)
	if len(av) != 6 || av[0][1] != "do 30 dni" || av[0][2] != "2" || av[0][3] != "1845.50" || av[5][0] != "=" || av[5][3] != "1845.50" {
		t.Fatalf("by age on %s: %v", day, av)
	}
	var why struct {
		Sums *bool `json:"sums"`
	}
	get("/api/explain?view=naleznosci-wiek&row=0&col=open", &why)
	if why.Sums == nil || !*why.Sums {
		t.Fatalf("the youngest bucket does not add up")
	}
	// The same books on New Year's Eve: the kernel's day moved, so did the age.
	day = "2026-12-31"
	get("/api/views/naleznosci-wiek", &age)
	av = cellVals(age.Rows)
	if av[0][2] != "0" || av[3][1] != "91–180 dni" || av[3][2] != "2" {
		t.Fatalf("by age on %s: %v", day, av)
	}

	// The result month by month (KK on the phone, 2026-10-17: "the result
	// for August only"). A sale in August, still open; a supplier's invoice
	// dated in September arriving after September closed is booked in
	// October, the first open month — the result counts it where the books
	// hold it. Each row is made of its month's bookings, and the months
	// add up to the result since January.
	submit("sales.invoice.issued", "2026-08-20", "fv-3", inv("FV 1/08/2026", "2026-08-20", "2000.00", "460.00", "2460.00"))
	submit("purchase.invoice.received", "2026-09-28", "fz-1", map[string]any{"number": "FZ/9/2026", "invoice_date": "2026-09-28", "market": "pl",
		"supplier_nip": "5260001246", "buyer_nip": "5250001111", "supplier": "Nabywca S.A.",
		"net": "400.00", "vat": "92.00", "gross": "492.00", "vat_rate": "23", "currency": "PLN"})
	var months rows
	get("/api/views/wynik-miesiace", &months)
	mv := cellVals(months.Rows)
	want := [][]string{
		{"2026-08", "", "Przychody", "2000.00"}, {"2026-08", "-", "Koszty", "0.00"}, {"2026-08", "=", "Wynik przed opodatkowaniem", "2000.00"},
		{"2026-09", "", "Przychody", "1500.00"}, {"2026-09", "-", "Koszty", "0.00"}, {"2026-09", "=", "Wynik przed opodatkowaniem", "1500.00"},
		{"2026-10", "", "Przychody", "0.00"}, {"2026-10", "-", "Koszty", "400.00"}, {"2026-10", "=", "Wynik przed opodatkowaniem", "-400.00"},
	}
	if len(mv) != len(want) {
		t.Fatalf("month by month: %v", mv)
	}
	for i := range want {
		if strings.Join(mv[i], "|") != strings.Join(want[i], "|") {
			t.Fatalf("month by month, row %d: %v, want %v", i, mv[i], want[i])
		}
	}
	var postings struct {
		Objects []core.Object `json:"objects"`
	}
	get("/api/objects?type=posting", &postings)
	dateOf := map[string]string{}
	for _, p := range postings.Objects {
		dateOf[p.ID], _ = p.State["date"].(string)
	}
	for i, ids := range months.Of {
		if n := map[int]int{0: 1, 1: 0, 2: 1, 3: 2, 4: 0, 5: 2, 6: 0, 7: 1, 8: 1}[i]; len(ids) != n {
			t.Fatalf("row %d %v is made of %v, want %d bookings", i, mv[i], ids, n)
		}
		for _, id := range ids {
			if !strings.HasPrefix(dateOf[id], mv[i][0]) {
				t.Fatalf("row %d (%s) is made of %s, booked %q", i, mv[i][0], id, dateOf[id])
			}
		}
	}
	if d := dateOf[months.Of[7][0]]; d != "2026-10-01" {
		t.Fatalf("the late cost was booked %q, want the first open month", d)
	}
	get("/api/views/wynik", &wynik)
	wv = cellVals(wynik.Rows)
	var sum int64
	for _, r := range mv {
		if r[1] == "=" {
			m, _ := core.ParseMoney(r[3])
			sum += m
		}
	}
	if wv[0][2] != "3500.00" || wv[1][2] != "400.00" || wv[2][2] != core.FormatMoney(sum) || wv[2][2] != "3100.00" {
		t.Fatalf("since January %v, the months add up to %s", wv, core.FormatMoney(sum))
	}

	// The same report as a tally — the shape a person teaches and the
	// author drafts (DIRECTION 2026-10-17, "Teaching her a figure"): the
	// kernel writes the query, and it finds what the pack's own SQL finds,
	// row for row and booking for booking.
	install := func(id string, spec map[string]any) {
		t.Helper()
		b, _ := json.Marshal(map[string]any{"tally": spec})
		if err := s.InsertViewDef(ctx, core.ViewDef{ID: id, Version: 1, Notion: "analysis", Title: id, Domain: "finance", Function: "analysis", Spec: b}); err != nil {
			t.Fatal(err)
		}
	}
	accType := func(v string) []map[string]any {
		return []map[string]any{{"field": "account.type", "op": "eq", "value": v}}
	}
	install("taught-result", map[string]any{"of": "posting", "value": "debit-credit",
		"where": []map[string]any{{"field": "book", "op": "eq", "value": "pl-stat"}, {"field": "account.type", "op": "in", "value": []string{"revenue", "expense"}}},
		"by":    []map[string]any{{"field": "date", "by": "month", "label": "Miesiąc"}},
		"lines": []map[string]any{{"line": "Przychody", "value": "credit-debit", "where": accType("revenue")},
			{"line": "Koszty", "op": "-", "where": accType("expense")}, {"line": "Wynik przed opodatkowaniem", "op": "="}},
		"line_label": "Pozycja", "amount_label": "Kwota"})
	var taught rows
	get("/api/views/taught-result", &taught)
	if a, b := cellVals(taught.Rows), mv; fmt.Sprint(a) != fmt.Sprint(b) || fmt.Sprint(taught.Of) != fmt.Sprint(months.Of) {
		t.Fatalf("the tally finds\n%v %v\nthe pack's SQL\n%v %v", a, taught.Of, b, months.Of)
	}
	if f := taught.Columns; len(f) != 4 || f[0].Field != "month" || f[3].Field != "amount" {
		t.Fatalf("columns: %+v", f)
	}
	// A document tally: sales by buyer and month, its cells checking themselves.
	install("taught-sales", map[string]any{"of": "sales_invoice", "value": "net",
		"by": []map[string]any{{"field": "buyer.name", "label": "Klient"}, {"field": "issue_date", "by": "month", "label": "Miesiąc"}}, "amount_label": "Netto"})
	var sales rows
	get("/api/views/taught-sales", &sales)
	if sv := cellVals(sales.Rows); fmt.Sprint(sv) != "[[Nabywca S.A. 2026-08 2000.00] [Nabywca S.A. 2026-09 1500.00]]" || len(sales.Of[1]) != 2 {
		t.Fatalf("sales by buyer and month: %v %v", sv, sales.Of)
	}
	why.Sums = nil
	get("/api/explain?view=taught-sales&row=1&col=amount", &why)
	if why.Sums == nil || !*why.Sums {
		t.Fatalf("September's sales do not add up on why?")
	}

	// Teaching it: the author drafts a figure, the card shows its rows
	// computed now, the person says what is wrong, and the redraft answers
	// the reason (the conversation is on the record).
	figure := func(id string) string {
		return `{"bundle_id": "` + id + `", "description": "Sprzedaż według klientów i miesięcy.",
		  "warrant": {"basis": "client", "citations": ["the owner asked"], "scope": "this company"},
		  "view_defs": [{"view_id": "` + id + `", "notion": "analysis", "title": "Sprzedaż według klientów", "domain": "sales", "function": "analysis",
		    "spec": {"tally": {"of": "sales_invoice", "value": "net", "by": [{"field": "buyer.name", "label": "Klient"},
		      {"field": "issue_date", "by": "month", "label": "Miesiąc"}], "amount_label": "Netto"}}}]}`
	}
	var asked []string
	answers := []string{figure("sales-by-customer"), figure("sales-by-customer-2")}
	srv.Agent = &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
		asked = append(asked, user)
		a := answers[0]
		answers = answers[1:]
		return a, nil
	}}
	post := func(path string, body, out any) {
		t.Helper()
		b, _ := json.Marshal(body)
		res, err := ts.Client().Post(ts.URL+path, "application/json", bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		var buf bytes.Buffer
		buf.ReadFrom(res.Body)
		if res.StatusCode != 200 {
			t.Fatalf("POST %s: %d %s", path, res.StatusCode, buf.String())
		}
		if out != nil {
			if err := json.Unmarshal(buf.Bytes(), out); err != nil {
				t.Fatal(err)
			}
		}
	}
	var drafted struct {
		ID    string `json:"bundle_id"`
		After string `json:"after"`
	}
	post("/api/bundles/draft", map[string]any{"intent": "Pokaż sprzedaż według klientów, miesiąc po miesiącu"}, &drafted)
	var pv []struct {
		ViewID string   `json:"view_id"`
		Rows   [][]cell `json:"rows"`
		Total  int      `json:"total"`
		Error  string   `json:"error"`
	}
	get("/api/bundles/"+drafted.ID+"/preview", &pv)
	if len(pv) != 1 || pv[0].Error != "" || pv[0].Total != 2 || fmt.Sprint(cellVals(pv[0].Rows)) != fmt.Sprint(cellVals(sales.Rows)) {
		t.Fatalf("the waiting figure, previewed: %+v", pv)
	}
	if _, err := x.TriggerActivity(ctx, "reject_bundle", map[string]any{"bundle_id": drafted.ID, "reason": "chcę też liczbę faktur",
		"rejected_by": "krzysztof", "occurred_at": day}, "krzysztof", ""); err != nil {
		t.Fatal(err)
	}
	post("/api/bundles/draft", map[string]any{"after": drafted.ID}, &drafted)
	if drafted.After != "sales-by-customer" || len(asked) != 2 || !strings.Contains(asked[1], "chcę też liczbę faktur") ||
		!strings.Contains(asked[1], "Pokaż sprzedaż według klientów, miesiąc po miesiącu") {
		t.Fatalf("the redraft did not answer the reason: %+v\n%s", drafted, asked[len(asked)-1])
	}
}
