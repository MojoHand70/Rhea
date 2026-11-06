package shell_test

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

// TestStreamTellsWhatHappened: the stream of events (DIRECTION 2026-11-05;
// KK: "we never show events", then "each event should be an event, not just
// a row"). A face asks GET /api/stream and gets cards, newest first: the fact
// that came in, and under it the steps the rules took — the chain of derived
// facts, grouped by rule, each with its rule version and the record it made
// — so a person reads what Rhea did with a payment without opening anything.
// A record's history is the chains that name it. Nothing here writes.
func TestStreamTellsWhatHappened(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	for _, f := range []string{"finance.json", "finance_v2.json", "finance_v3.json", "finance_v4.json",
		"finance_v5.json", "finance_v6.json", "finance_v7.json", "finance_v8.json"} {
		seedFromFile(t, s, f)
	}
	x := &exec.Executor{Store: s}
	srv := &shell.Server{Store: s, Exec: x, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb")}
	ts := serve(t, srv.Handler())
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
			Fields: map[string]string{"name": "=$.name", "kind": "=$.kind", "vat_id": "=$.vat_id", "country": "=$.country", "iban": "=$.iban"}}}},
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
	submit("company.registered", "2026-09-01", "buyer", map[string]any{"name": "Nabywca S.A.", "kind": "customer", "vat_id": "5260001246", "country": "PL", "iban": "PL60102010260000042270201111"})
	inv := func(no, date, net, vat, gross string) map[string]any {
		return map[string]any{"number": no, "issue_date": date, "market": "pl", "intragroup": "no", "seller_nip": "5250001111",
			"buyer_nip": "5260001246", "net": net, "vat": vat, "gross": gross, "vat_rate": "23", "currency": "PLN"}
	}
	submit("sales.invoice.issued", "2026-09-21", "fv-1", inv("FV 1/09/2026", "2026-09-21", "1000.00", "230.00", "1230.00"))
	submit("sales.invoice.issued", "2026-09-24", "fv-2", inv("FV 2/09/2026", "2026-09-24", "500.00", "115.00", "615.00"))
	// the company's own account first (Poland pack v54): a line of an account nobody approved waits
	if _, err := x.TriggerActivity(ctx, "propose_bank_account", map[string]any{"name": "Rachunek bieżący", "number": "PL83101010230000261395100000", "currency": "PLN", "proposed_by": "krzysztof"}, "krzysztof", "acct"); err != nil {
		t.Fatal(err)
	}
	accounts, _ := s.ObjectsByType(ctx, "bank_account")
	if len(accounts) != 1 {
		t.Fatalf("the account proposed: %+v", accounts)
	}
	if _, err := x.TriggerActivity(ctx, "approve_bank_account", map[string]any{"bank_account": accounts[0].ID, "approved_by": "krzysztof"}, "krzysztof", "acct-ok"); err != nil {
		t.Fatal(err)
	}
	submit("bank.statement.line", "2026-09-25", "wb-1", map[string]any{"date": "2026-09-25", "amount": "1230.00", "currency": "PLN", "direction": "credit", "market": "pl",
		"reference": "zapłata FV 1/09/2026", "counterparty": "Nabywca S.A.", "counterparty_account": "PL60102010260000042270201111", "account": "PL83101010230000261395100000"})

	type step struct {
		Says        []core.SayPart `json:"says"`
		Area        string         `json:"area"`
		State       string         `json:"state"`
		RuleID      string         `json:"rule_id"`
		RuleVersion int            `json:"rule_version"`
		EventID     int64          `json:"event_id"`
		ObjectID    string         `json:"object_id"`
		ObjectType  string         `json:"object_type"`
		N           int            `json:"n"`
	}
	type card struct {
		EventID    int64          `json:"event_id"`
		OccurredAt string         `json:"occurred_at"`
		EventType  string         `json:"event_type"`
		Title      []core.SayPart `json:"title"`
		Source     string         `json:"source"`
		State      string         `json:"state"`
		ObjectID   string         `json:"object_id"`
		ObjectType string         `json:"object_type"`
		Steps      []step         `json:"steps"`
	}
	flat := func(parts []core.SayPart) string {
		var b strings.Builder
		for _, p := range parts {
			b.WriteString(p.T + p.F + p.M + p.D)
		}
		return b.String()
	}
	rules := func(c card) string {
		var out []string
		for _, st := range c.Steps {
			out = append(out, st.RuleID)
		}
		return strings.Join(out, ",")
	}

	var page struct {
		Cards      []card `json:"cards"`
		NextBefore int64  `json:"next_before"`
	}
	get("/api/stream", 200, &page)
	// unscoped, the whole log: the doors' facts and the pack's master data too, newest first; the payment first of all
	find := func(cards []card, typ string) card {
		for _, c := range cards {
			if c.EventType == typ {
				return c
			}
		}
		t.Fatalf("no card of %s among %d", typ, len(cards))
		return card{}
	}
	if len(page.Cards) != 20 || page.NextBefore == 0 || page.Cards[0].EventType != "bank.statement.line" {
		t.Fatalf("the first page: %d cards, next %d, first %s", len(page.Cards), page.NextBefore, page.Cards[0].EventType)
	}
	pay := page.Cards[0]
	// the payment's card: what came in, then the chain — recognised, posted (two entries as one step),
	// matched to the invoice, the operation and the invoice settled, the settlement recorded
	if pay.State != "ok" || pay.ObjectType != "bank_operation" || pay.ObjectID == "" {
		t.Fatalf("the payment's card: %+v", pay)
	}
	// the settlement record is silent: the settled invoice's step says it (the pack's words)
	if got := rules(pay); got != "pl-bank-operation,pl-post-incoming-payment,pl-match-incoming,pl-settle-incoming-operation,pl-settle-sales-invoice" {
		t.Fatalf("the payment's steps: %s", got)
	}
	if pay.Steps[1].N != 2 || pay.Steps[1].ObjectType != "posting" || pay.Steps[0].RuleVersion == 0 || pay.Steps[0].ObjectID != pay.ObjectID {
		t.Fatalf("the posting step folds the rule's two entries: %+v", pay.Steps[1])
	}
	// the steps read in the pack's Polish, each with the area that handled it; money travels as minor
	// units with its currency, a month as a month, for the face to format
	for i, want := range []struct{ says, area string }{
		{"rozpoznana jako zapłata kontrahenta", "Bank"}, {"zaksięgowana: 2 zapisy", "Księgi"},
		{"dopasowana do FV 1/09/2026 po numerze", "Rozrachunki"}, {"wpłata rozliczona", "Rozrachunki"},
		{"faktura FV 1/09/2026 zapłacona", "Sprzedaż"}} {
		if got := flat(pay.Steps[i].Says); got != want.says || pay.Steps[i].Area != want.area {
			t.Fatalf("step %d reads %q (%s), want %q (%s)", i, got, pay.Steps[i].Area, want.says, want.area)
		}
	}
	if got := flat(pay.Title); got != "wpływ 123000: Nabywca S.A., tytułem zapłata FV 1/09/2026" || pay.Title[1].C != "PLN" || pay.Source != "Bank" {
		t.Fatalf("the payment's title: %q, money in %q, from %q", got, pay.Title[1].C, pay.Source)
	}
	// every step names its rule version and the fact it is
	for _, st := range pay.Steps {
		if st.RuleVersion == 0 || st.EventID == 0 || st.State != "ok" || len(st.Says) == 0 {
			t.Fatalf("a step without its provenance: %+v", st)
		}
	}
	// the invoice's card: booked, posted in three entries
	fv2 := find(page.Cards, "sales.invoice.issued")
	if got := rules(fv2); got != "pl-book-sales-invoice,pl-post-sales-invoice" || fv2.Steps[1].N != 3 || fv2.ObjectType != "sales_invoice" {
		t.Fatalf("the invoice's card: %s %+v", got, fv2)
	}
	if got := flat(fv2.Steps[0].Says); got != "w księgach jako FV 2/09/2026, VAT za 2026-09" || fv2.Steps[0].Says[3].As != "month" || flat(fv2.Steps[1].Says) != "zaksięgowana: 3 zapisy" {
		t.Fatalf("the invoice's steps: %q, %q", got, flat(fv2.Steps[1].Says))
	}
	if got := flat(fv2.Title); got != "wystawiono fakturę FV 2/09/2026 na 61500" {
		t.Fatalf("the invoice's title: %q", got)
	}

	// the page before: the next facts, none repeated
	var page2 struct {
		Cards []card `json:"cards"`
	}
	get("/api/stream?before="+strconv.FormatInt(page.NextBefore, 10)+"&limit=5", 200, &page2)
	if len(page2.Cards) != 5 || page2.Cards[0].EventID >= page.NextBefore {
		t.Fatalf("the page before %d: %+v", page.NextBefore, page2.Cards)
	}

	// a record's history: the invoice's own fact and the payment that settled it, newest first
	var hist struct {
		Cards []card `json:"cards"`
	}
	invoiceID := ""
	for _, st := range pay.Steps {
		if st.RuleID == "pl-settle-sales-invoice" {
			invoiceID = st.ObjectID
		}
	}
	get("/api/stream?object="+invoiceID, 200, &hist)
	if len(hist.Cards) != 2 || hist.Cards[0].EventType != "bank.statement.line" || hist.Cards[1].EventType != "sales.invoice.issued" || hist.Cards[1].ObjectID != invoiceID {
		t.Fatalf("the invoice's history: %+v", hist.Cards)
	}
	// a step the pack marks as waiting for a person waits only until someone acts: the account proposed and
	// approved above reads as done, a second one nobody has decided on waits, and its card with it
	if _, err := x.TriggerActivity(ctx, "propose_bank_account", map[string]any{"name": "Rachunek pomocniczy", "number": "PL27114020040000300201355387", "currency": "PLN", "proposed_by": "krzysztof"}, "krzysztof", "acct-2"); err != nil {
		t.Fatal(err)
	}
	get("/api/stream?limit=40", 200, &page)
	var proposals []card
	for _, c := range page.Cards {
		if c.EventType == "bank_account.proposed" {
			proposals = append(proposals, c)
		}
	}
	if len(proposals) != 2 || proposals[0].State != "wait" || proposals[0].Steps[0].State != "wait" || proposals[1].State != "ok" || proposals[1].Steps[0].State != "ok" {
		t.Fatalf("the proposals' states, newest first: %+v", proposals)
	}
	if got := flat(proposals[0].Steps[0].Says); got != "rachunek PL27114020040000300201355387 zgłoszony, czeka na zatwierdzenie" {
		t.Fatalf("the waiting step: %q", got)
	}

	// a fact no rule answered is a card too, standing in the state "none": a question, not a story
	tr, err := x.TriggerActivity(ctx, "submit_event", map[string]any{"event_type": "parcel.lost", "occurred_at": "2026-09-26", "payload": map[string]any{"what": "a parcel"}}, "test", "lost")
	if err != nil || tr.Booked != 0 {
		t.Fatalf("an unexplained fact: %+v %v", tr, err)
	}
	get("/api/stream?limit=1", 200, &page)
	if len(page.Cards) != 1 || page.Cards[0].EventType != "parcel.lost" || page.Cards[0].State != "none" || len(page.Cards[0].Steps) != 0 {
		t.Fatalf("the unexplained fact's card: %+v", page.Cards)
	}
	// a fact the words do not know still has a title: its own name — the kernel speaks, in no language
	if got := flat(page.Cards[0].Title); got != "parcel.lost" {
		t.Fatalf("a title without words: %q", got)
	}
	get("/api/stream?limit=x", 400, nil)
}
