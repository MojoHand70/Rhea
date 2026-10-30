package agent_test

import (
	"context"
	"strings"
	"testing"

	"rhea/internal/agent"
	"rhea/internal/core"
)

// TestAnswer: the author sees the collection in the business's own names and
// answers with a query, what it counts and how to show it — or why not.
func TestAnswer(t *testing.T) {
	var seen string
	reply := "Here it is:\n" + `{"sql": "SELECT 1 AS \"Liczba\"", "reading": "jedna liczba", "visual": {"kind": "kpi", "value": "Liczba"}}` + "\nDone."
	a := &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) { seen = user; return reply, nil }}
	ask := agent.AnswerAsk{Question: "Sprzedaż według klientów?", Today: "2026-10-18", Books: []string{"pl-stat"},
		Tables: []agent.CollectionTable{{Schema: "finance", Name: "sales_invoice", Type: "sales_invoice", Columns: []agent.CollectionColumn{
			{Name: "id", SQL: "VARCHAR"}, {Name: "net", SQL: "DECIMAL(18,2)", Kind: "money"}, {Name: "buyer", SQL: "VARCHAR", Kind: "ref<company>", Links: "finance.company"},
			{Name: "_state", SQL: "VARCHAR"}}}},
		Words: &core.WordsSpec{Lang: "pl", Types: map[string]core.TypeWords{"sales_invoice": {Label: "Faktura sprzedaży"}}, Fields: map[string]string{"net": "Netto", "sales_invoice.buyer": "Nabywca"},
			Notes: []string{"A ZUS payment is a bank operation of kind zus going out."}},
		Tried: &agent.AnswerTry{SQL: "SELECT nope", Error: "Binder Error: nope"}}
	ans, err := a.Answer(context.Background(), ask)
	if err != nil || ans.SQL != `SELECT 1 AS "Liczba"` || ans.Visual.Kind != "kpi" || ans.Reading != "jedna liczba" {
		t.Fatalf("answer: %+v %v", ans, err)
	}
	for _, want := range []string{"finance.sales_invoice — Faktura sprzedaży", "net DECIMAL(18,2) (money) — Netto", "buyer VARCHAR (ref<company>) → finance.company.id — Nabywca",
		"Books the active rules post to: pl-stat", "Today: 2026-10-18", "How this market's books are read", "- A ZUS payment is a bank operation of kind zus going out.", "DuckDB said: Binder Error: nope", "Question: Sprzedaż według klientów?"} {
		if !strings.Contains(seen, want) {
			t.Errorf("the ask lacks %q:\n%s", want, seen)
		}
	}
	if strings.Contains(seen, "_state VARCHAR") {
		t.Errorf("the provenance is listed per table")
	}
	reply = `{"cannot": "w księgach nie ma magazynu"}`
	if ans, err := a.Answer(context.Background(), ask); err != nil || ans.Cannot == "" {
		t.Fatalf("cannot: %+v %v", ans, err)
	}
	reply = `{"sql": "SELECT 1"}`
	if _, err := a.Answer(context.Background(), ask); err == nil {
		t.Fatalf("a query that says nothing of what it counts was taken")
	}
	// A kept report asked for again, for another month, is named, not
	// written again (KK 2026-10-30); one the company has not is refused.
	ask.Reports = []agent.AnswerReport{{ID: "answer-12@company-95", Title: "RZiS", Parts: []string{"przychody, koszty i wynik miesiąca"}, Period: true}}
	reply = `{"report": "answer-12@company-95", "from": "2026-08-01", "to": "2026-08-31", "reading": "RZiS za sierpień"}`
	ans, err = a.Answer(context.Background(), ask)
	if err != nil || ans.Report != "answer-12@company-95" || ans.From != "2026-08-01" || ans.To != "2026-08-31" || ans.SQL != "" {
		t.Fatalf("a kept report named: %+v %v", ans, err)
	}
	if !strings.Contains(seen, "Kept reports") || !strings.Contains(seen, "answer-12@company-95 „RZiS” - przychody, koszty i wynik miesiąca - takes a period") {
		t.Errorf("the ask lacks the kept reports:\n%s", seen)
	}
	reply = `{"report": "answer-99", "reading": "x"}`
	if _, err := a.Answer(context.Background(), ask); err == nil {
		t.Fatal("a report the company has not was taken")
	}
}
