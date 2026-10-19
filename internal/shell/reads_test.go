package shell_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

// TestFaceReads: a face built apart reads only the API (DIRECTION
// 2026-10-16, the face's first build), so the API answers what the face
// shows — the objects with their provenance, who approved each rule, what
// the host said — and "why?" on a cell of a view: the invoices a VAT
// register row is made of, each one's part, and whether the parts make the
// cell. Nothing here writes.
func TestFaceReads(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	for _, f := range []string{"finance.json", "finance_v2.json", "finance_v3.json", "finance_v4.json",
		"finance_v5.json", "finance_v6.json", "finance_v7.json", "finance_v8.json"} {
		seedFromFile(t, s, f)
	}
	x := &exec.Executor{Store: s}
	srv := &shell.Server{Store: s, Exec: x, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb")}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	get := func(path string, want int, out any) {
		t.Helper()
		res, err := ts.Client().Get(ts.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		var b bytes.Buffer
		b.ReadFrom(res.Body)
		if res.StatusCode != want {
			t.Fatalf("GET %s: %d, want %d: %s", path, res.StatusCode, want, b.String())
		}
		if out != nil {
			if err := json.Unmarshal(b.Bytes(), out); err != nil {
				t.Fatalf("GET %s: %v", path, err)
			}
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
	if _, _, _, err := x.ApproveRule(ctx, "register-company", "anna", "2026-09-01"); err != nil {
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
	pay := func(no, date, net, vat, gross string) map[string]any {
		return map[string]any{"number": no, "issue_date": date, "market": "pl", "intragroup": "no", "seller_nip": "5250001111",
			"buyer_nip": "5260001246", "net": net, "vat": vat, "gross": gross, "vat_rate": "23", "currency": "PLN"}
	}
	submit("sales.invoice.issued", "2026-09-21", "fv-1", pay("FV 1/09/2026", "2026-09-21", "1000.00", "230.00", "1230.00"))
	submit("sales.invoice.issued", "2026-09-24", "fv-2", pay("FV 2/09/2026", "2026-09-24", "500.00", "115.00", "615.00"))

	// The objects, with their provenance.
	var objs struct {
		Objects []core.Object `json:"objects"`
	}
	get("/api/objects?type=sales_invoice", 200, &objs)
	if len(objs.Objects) != 2 || objs.Objects[0].RuleID != "pl-book-sales-invoice" || objs.Objects[0].SourceEventID == 0 ||
		objs.Objects[0].State["gross"] != float64(123000) {
		t.Fatalf("objects: %+v", objs.Objects)
	}
	first := objs.Objects[0].ID
	var one struct {
		core.Object
		Root int64 `json:"root_event_id"`
	}
	get("/api/objects/"+first, 200, &one)
	if one.ID != first || one.State["number"] != "FV 1/09/2026" {
		t.Fatalf("object: %+v", one)
	}
	// Every object names the fact its chain starts from: the invoice and the
	// entry it booked share the sales.invoice.issued event.
	invoice, err := s.GetEvent(ctx, one.Root)
	if err != nil || invoice.Kind != core.KindRaw || invoice.Type != "sales.invoice.issued" || invoice.DedupKey != "fv-1" {
		t.Fatalf("the invoice's root %d: %+v %v", one.Root, invoice, err)
	}
	var posts struct {
		Objects []struct {
			ID   string `json:"object_id"`
			Root int64  `json:"root_event_id"`
		} `json:"objects"`
	}
	get("/api/objects?type=posting", 200, &posts)
	booked := 0
	for _, p := range posts.Objects {
		if p.Root == one.Root {
			booked++
		}
	}
	if booked != 3 {
		t.Fatalf("postings sharing the invoice's root: %d of %+v", booked, posts.Objects)
	}
	get("/api/objects", 400, nil)
	get("/api/objects/no-such-thing", 404, nil)

	// Who approved each active rule: through a bundle, or through its own door.
	var rules []struct {
		ID       string `json:"rule_id"`
		Version  int    `json:"version"`
		Approval *struct {
			By     string `json:"approved_by"`
			At     string `json:"approved_at"`
			Bundle string `json:"bundle"`
		} `json:"approval"`
	}
	get("/api/rules", 200, &rules)
	seen := 0
	for _, r := range rules {
		switch r.ID {
		case "pl-book-sales-invoice":
			if r.Approval == nil || r.Approval.By != "krzysztof" || r.Approval.Bundle != "pack-pl-v32" || r.Approval.At != "2026-09-01" {
				t.Fatalf("pack rule approval: %+v", r.Approval)
			}
			seen++
		case "register-company":
			if r.Approval == nil || r.Approval.By != "anna" || r.Approval.Bundle != "" {
				t.Fatalf("rule door approval: %+v", r.Approval)
			}
			seen++
		}
	}
	if seen != 2 {
		t.Fatalf("rules seen %d", seen)
	}

	// A VAT register row says what it is made of; the column is not shown.
	var reg struct {
		Columns    []struct{ Field string } `json:"columns"`
		Rows       [][]cell                 `json:"rows"`
		RowObjects [][]string               `json:"row_objects"`
	}
	get("/api/views/rejestr-vat-sprzedazy", 200, &reg)
	if len(reg.Columns) != 5 || len(reg.Rows) != 1 || len(reg.Rows[0]) != 5 || len(reg.RowObjects) != 1 || len(reg.RowObjects[0]) != 2 {
		t.Fatalf("register: %+v", reg)
	}
	// Faces ask for several reports at once (and people open two tabs): the
	// analysis reads take turns instead of colliding on the rebuild.
	errs := make(chan string, 8)
	for i := 0; i < 8; i++ {
		go func(i int) {
			v := []string{"rejestr-vat-sprzedazy", "trial-balance", "rejestr-vat-zakupu", "rozliczenie-vat"}[i%4]
			res, err := ts.Client().Get(ts.URL + "/api/views/" + v)
			if err != nil {
				errs <- err.Error()
				return
			}
			res.Body.Close()
			if res.StatusCode != 200 {
				errs <- fmt.Sprintf("%s: %d", v, res.StatusCode)
				return
			}
			errs <- ""
		}(i)
	}
	for i := 0; i < 8; i++ {
		if e := <-errs; e != "" {
			t.Fatalf("reports at once: %s", e)
		}
	}

	// "Why?" on the cell: the invoices, each one's part, and the check.
	var why struct {
		V       string `json:"v"`
		Total   string `json:"total"`
		Sums    *bool  `json:"sums"`
		Objects []struct {
			ID    string `json:"object_id"`
			Label string `json:"label"`
			Field string `json:"field"`
			V     string `json:"v"`
		} `json:"objects"`
	}
	get("/api/explain?view=rejestr-vat-sprzedazy&row=0&col=vat", 200, &why)
	if why.V != "345.00" || why.Total != "345.00" || why.Sums == nil || !*why.Sums || len(why.Objects) != 2 ||
		why.Objects[0].Label != "FV 1/09/2026" || why.Objects[0].Field != "vat_pln" || why.Objects[0].V != "230.00" || why.Objects[1].V != "115.00" {
		t.Fatalf("why on the VAT cell: %+v", why)
	}
	why.Sums, why.Objects, why.Total = nil, nil, ""
	get("/api/explain?view=rejestr-vat-sprzedazy&row=0&col=month", 200, &why)
	if why.Sums != nil || len(why.Objects) != 2 || why.Objects[0].V != "" {
		t.Fatalf("why on a column that sums nothing: %+v", why)
	}
	get("/api/explain?view=trial-balance&row=0&col=debit", 400, nil) // says nothing of its rows: refused, never guessed
	get("/api/explain?view=rejestr-vat-sprzedazy&row=1&col=vat", 400, nil)
	get("/api/explain?view=rejestr-vat-sprzedazy&row=0&col=nope", 400, nil)

	// The host's facts carry a report's rows, never what they are made
	// of: her words are checked against the facts' numbers, and an id is
	// not a number she may speak.
	var brief struct {
		Facts []struct {
			Kind string          `json:"kind"`
			View string          `json:"view"`
			Data json.RawMessage `json:"data"`
		} `json:"facts"`
	}
	get("/api/views/poranek", 200, &brief)
	reports := 0
	for _, f := range brief.Facts {
		if f.Kind != "analysis" {
			continue
		}
		reports++
		if bytes.Contains(f.Data, []byte("_invoice-")) || bytes.Contains(f.Data, []byte("row_objects")) || bytes.Contains(f.Data, []byte(core.RowsOfColumn)) {
			t.Fatalf("a report's lineage reached the host's facts (%s): %s", f.View, f.Data)
		}
	}
	if reports != 4 {
		t.Fatalf("reports among the host's facts: %d", reports)
	}

	// What the host said, newest first, without asking her to speak again.
	var said []struct {
		View string `json:"view"`
		Text string `json:"text"`
	}
	get("/api/said?view=poranek", 200, &said)
	if len(said) != 0 {
		t.Fatalf("said before speaking: %+v", said)
	}
	for _, words := range []string{"Wczoraj spokojnie.", "Dwie faktury czekają."} {
		if _, err := x.TriggerActivity(ctx, "host_said", map[string]any{"view": "poranek", "text": words}, "agent:test", ""); err != nil {
			t.Fatal(err)
		}
	}
	get("/api/said?view=poranek", 200, &said)
	if len(said) != 2 || said[0].Text != "Dwie faktury czekają." {
		t.Fatalf("said: %+v", said)
	}
	get("/api/said?view=poranek&n=1", 200, &said)
	if len(said) != 1 || !strings.HasPrefix(said[0].Text, "Dwie") {
		t.Fatalf("said, n=1: %+v", said)
	}
	get("/api/said?view=elsewhere", 200, &said)
	if len(said) != 0 {
		t.Fatalf("said in another view: %+v", said)
	}
}
