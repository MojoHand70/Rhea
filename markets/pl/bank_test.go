package pl_test

import (
	"context"
	"fmt"
	"testing"

	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

// A bank line says what it is before it is booked (the Poland pack review,
// MON-07, MON-08, MON-18, ZUS-08, OWN-03, OWN-05; pack v21): a fee and
// interest by the bank's code, a transfer between the company's own
// accounts, a payment to the tax micro-account or the ZUS contribution
// account, the owner's money — each on its own account and none waiting for
// an invoice; only the rest is a customer's or supplier's payment. A line
// nothing explains waits until a person says what it is, and then moves off
// the trade receivables. A company's declaration is checked against its
// legal form.
func TestABankLineSaysWhatItIs(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	installPoland(t, s, x)
	selfID := registerParties(t, s, x)
	const (
		self   = "PL83101010230000261395100000"
		tax    = "PL12101010100000000000000001"
		zus    = "PL98600000020260000000000001"
		owner  = "PL61109010140000071219812874"
		other  = "PL27114020040000300201355387"
		credit = "credit"
		debit  = "debit"
	)
	if tr, err := x.TriggerActivity(ctx, "declare_company_accounts", map[string]any{"company": selfID, "tax_account": tax, "zus_account": zus, "owner_account": owner, "occurred_at": "2026-09-01"}, "krzysztof", ""); err != nil || tr.Booked != 1 {
		t.Fatalf("declare_company_accounts: %+v %v", tr, err)
	}
	// A later declaration that names only one account keeps the others.
	if tr, err := x.TriggerActivity(ctx, "declare_company_accounts", map[string]any{"company": selfID, "zus_account": zus, "occurred_at": "2026-09-02"}, "krzysztof", ""); err != nil || tr.Booked != 1 {
		t.Fatalf("the second declaration: %+v %v", tr, err)
	}
	if c, _ := s.GetObject(ctx, selfID); c.State["tax_account"] != tax || c.State["owner_account"] != owner {
		t.Fatalf("the accounts as declared: %v", c.State)
	}

	line := func(n int, code, direction, amount, party string) {
		t.Helper()
		append1(t, s, x, "bank.statement.line", "2026-09-20", fmt.Sprintf("line-%d", n), map[string]any{
			"date": "2026-09-20", "amount": amount, "currency": "PLN", "direction": direction, "reference": fmt.Sprintf("pozycja %d", n),
			"counterparty": "", "counterparty_account": party, "account": self, "code": code, "market": "pl"})
	}
	line(1, "NTRF", debit, "1000.00", tax)    // VAT to the micro-account
	line(2, "NTRF", debit, "1600.00", zus)    // contributions
	line(3, "NTRF", credit, "5000.00", owner) // the owner pays money in
	line(4, "NCHG", debit, "12.00", "")       // the bank's charge
	line(5, "NINT", credit, "3.21", "")       // interest
	line(6, "NTRF", debit, "200.00", self)    // to the company's own account
	line(7, "NTRF", credit, "700.00", other)  // nobody known
	ops, _ := s.ObjectsByType(ctx, "bank_operation")
	kinds := map[string]string{}
	for _, o := range ops {
		kinds[fmt.Sprint(o.State["title"])] = fmt.Sprintf("%v/%v", o.State["kind"], o.State["settlement_status"])
	}
	want := map[string]string{"pozycja 1": "tax/settled", "pozycja 2": "zus/settled", "pozycja 3": "owner/settled", "pozycja 4": "fee/settled",
		"pozycja 5": "interest/settled", "pozycja 6": "own/settled", "pozycja 7": "transfer/unsettled"}
	for title, k := range want {
		if kinds[title] != k {
			t.Fatalf("%s: %s, want %s (all %v)", title, kinds[title], k, kinds)
		}
	}
	d, c := debits(t, s), credits(t, s)
	if d["225"] != 100000 || d["229"] != 160000 || c["249"] != 500000 || d["403"] != 1200 || c["750"] != 321 || d["149"] != 20000 ||
		c["201"] != 70000 || d["202"] != 0 {
		t.Fatalf("by account: debits %v, credits %v", d, c)
	}

	// The stranger's 700 was the owner's from another account: a person
	// says so, and it moves from the receivables to the owner's account.
	var unknown string
	for _, o := range ops {
		if o.State["title"] == "pozycja 7" {
			unknown = o.ID
		}
	}
	if tr, err := x.TriggerActivity(ctx, "classify_bank_operation", map[string]any{"operation": unknown, "kind": "owner", "occurred_at": "2026-09-21"}, "krzysztof", ""); err != nil || tr.Booked != 1 {
		t.Fatalf("classify: %+v %v", tr, err)
	}
	if o, _ := s.GetObject(ctx, unknown); o.State["kind"] != "owner" || o.State["settlement_status"] != "settled" || o.State["open"] != float64(0) {
		t.Fatalf("the classified line: %v", o.State)
	}
	if d, c := debits(t, s), credits(t, s); d["201"] != 70000 || c["249"] != 570000 {
		t.Fatalf("after the classification: debits %v, credits %v", d, c)
	}
	// Said twice, it is already explained: nothing matches.
	if tr, err := x.TriggerActivity(ctx, "classify_bank_operation", map[string]any{"operation": unknown, "kind": "fee", "occurred_at": "2026-09-22"}, "krzysztof", ""); err != nil || tr.Booked != 0 {
		t.Fatalf("a second word on an explained line: %+v %v", tr, err)
	}

	// A legal person pays CIT on full books: the scale and a KPiR are not its
	// to declare, so that declaration waits.
	if tr, err := x.TriggerActivity(ctx, "identify_company", map[string]any{"company": selfID, "email": "biuro@rhea-test.pl", "tax_office": "1471", "legal_form": "legal_entity", "occurred_at": "2026-09-01"}, "krzysztof", ""); err != nil || tr.Booked != 1 {
		t.Fatalf("identify_company: %+v %v", tr, err)
	}
	if tr, err := x.TriggerActivity(ctx, "declare_company_form", map[string]any{"company": selfID, "vat_status": "active", "books": "kpir", "income_tax": "skala", "from": "2026-01-01", "occurred_at": "2026-09-01"}, "krzysztof", ""); err != nil || tr.Booked != 0 {
		t.Fatalf("a company on the scale: %+v %v", tr, err)
	}
	if tr, err := x.TriggerActivity(ctx, "declare_company_form", map[string]any{"company": selfID, "vat_status": "active", "books": "full", "income_tax": "cit", "from": "2026-01-01", "occurred_at": "2026-09-01"}, "krzysztof", ""); err != nil || tr.Booked != 1 {
		t.Fatalf("a company on CIT: %+v %v", tr, err)
	}
}
