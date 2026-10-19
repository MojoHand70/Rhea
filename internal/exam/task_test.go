package exam

import (
	"encoding/json"
	"strings"
	"testing"

	"rhea/internal/eval"
)

func TestCaseCorpusIsTheChartTheFactsAndOneBundlePerIntent(t *testing.T) {
	c := Case{
		Paper:   "Rachunkowość finansowa 2025",
		Task:    "I-1",
		Company: "MANITU S.A.",
		Accounts: []CaseAccount{
			{Code: "141", Name: "Krótkoterminowe aktywa finansowe", Type: "asset"},
			{Code: "145", Name: "Odpisy aktualizujące", Type: "asset"},
			{Code: "751", Name: "Koszty finansowe", Type: "expense"},
		},
		Events: []CaseEvent{
			{EventType: "shares.valued", OccurredAt: "2024-12-31", Payload: map[string]any{"quantity": "1000", "price": "54.00", "market_price": "44.00", "write_down": "10000.00"}},
		},
		Intents: []string{"When shares.valued arrives, book the write_down: debit 751, credit 145."},
		ExpectPostings: []eval.Posting{
			{Account: "751", Side: "debit", Amount: "10000.00"},
			{Account: "145", Side: "credit", Amount: "10000.00"},
		},
	}
	if err := c.Check(); err != nil {
		t.Fatal(err)
	}
	cp, err := CaseCorpus(c, []string{"../../seed/finance.json"})
	if err != nil {
		t.Fatal(err)
	}
	if cp.Name != "exam-rachunkowosc-finansowa-2025-i-1" {
		t.Errorf("name %q", cp.Name)
	}
	if len(cp.Events) != 4 || len(cp.Tasks) != 2 || cp.Tasks[0].Intent != chartIntent || cp.Tasks[1].Mode != "bundle" {
		t.Errorf("events %d tasks %+v", len(cp.Events), cp.Tasks)
	}
	var first map[string]any
	json.Unmarshal(cp.Events[0], &first)
	if first["event_type"] != "account.created" || first["occurred_at"] != "2024-12-31" {
		t.Errorf("the chart opens with the first fact: %v", first)
	}
	if cp.Expect["account"] != 3 || len(cp.ExpectPostings) != 2 {
		t.Errorf("expectations: %v %v", cp.Expect, cp.ExpectPostings)
	}
}

func TestCaseCheckNamesWhatIsMissing(t *testing.T) {
	err := Case{Accounts: []CaseAccount{{Code: "1", Type: "cash"}}, Events: []CaseEvent{{EventType: "x.y", OccurredAt: "2024-01-01"}}, Intents: []string{"book z.w"}}.Check()
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, want := range []string{"type \"cash\"", "event x.y: no intent names it", "no expected postings"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error lacks %q: %v", want, err)
		}
	}
}

func TestSlug(t *testing.T) {
	for in, want := range map[string]string{
		"Rachunkowość finansowa 2025":                "rachunkowosc-finansowa-2025",
		"Prawo – prawo cywilne, pracy… 2025":         "prawo-prawo-cywilne-pracy-2025",
		"Sprawozdania finansowe i ich analiza, 2025": "sprawozdania-finansowe-i-ich-analiza-2025",
	} {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSeedsLoadInVersionOrder(t *testing.T) {
	seeds := []string{"x/finance_v10.json", "x/finance_v2.json", "x/finance.json", "x/finance_v9.json"}
	if !seedLess(seeds[2], seeds[1]) || !seedLess(seeds[3], seeds[0]) || seedLess(seeds[0], seeds[3]) {
		t.Errorf("order: %v", seeds)
	}
}
