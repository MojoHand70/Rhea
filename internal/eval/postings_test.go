package eval

import (
	"testing"

	"rhea/internal/core"
)

func TestCheckPostingsAddsUpByAccountAndSide(t *testing.T) {
	accounts := []core.Object{
		{ID: "account-1", State: map[string]any{"code": "751"}},
		{ID: "account-2", State: map[string]any{"code": "145"}},
	}
	postings := []core.Object{
		{State: map[string]any{"account": "account-1", "side": "debit", "amount": float64(600000), "book": "main"}},
		{State: map[string]any{"account": "account-1", "side": "debit", "amount": float64(400000), "book": "main"}},
		{State: map[string]any{"account": "account-2", "side": "credit", "amount": float64(1000000), "book": "main"}},
		{State: map[string]any{"account": "account-2", "side": "credit", "amount": float64(500), "book": "tax"}},
	}
	want := []Posting{
		{Account: "751", Side: "debit", Amount: "6000"}, // two lines of the solution on one account add up
		{Account: "751", Side: "debit", Amount: "4000.00"},
		{Account: "145", Side: "credit", Amount: "10000.00", Book: "main"},
		{Account: "145", Side: "debit", Amount: "10000.00"},
	}
	got := checkPostings(accounts, postings, want)
	if len(got) != 1 || got[0] != "posting 145 debit: 0.00 of 10000.00" {
		t.Fatalf("mismatches: %q", got)
	}
}
