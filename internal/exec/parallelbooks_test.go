package exec_test

// E1 of the enterprise-structure ladder (DIRECTION): can two rule-books — PL
// statutory and group GAAP — explain the same events into two independent
// ledgers with nothing but the existing vocabulary? This file is the recorded
// fail-as-data attempt (DECISIONS 2026-10-04). What survives: multi-rule
// firing books both entries from one event, by convention — book identity
// smuggled into account-code prefixes, first-class nowhere. What fails:
// independent period closes. The period lock is month-only and global, so
// closing the statutory book locks the group book too, and atomic firing
// makes it total: one locked book refuses the event for every book.

import (
	"context"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

func TestParallelBooksAsPureData(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	for _, ot := range []core.ObjectType{
		{Name: "account", Version: 1, Domain: "finance", LabelField: "name",
			Fields: []core.FieldDef{
				{Name: "code", Type: "string", Required: true},
				{Name: "name", Type: "string", Required: true},
				{Name: "type", Type: "enum", Required: true,
					Values: []string{"asset", "liability", "equity", "revenue", "expense", "debtor", "creditor", "tax", "bank", "clearing"}},
			}},
		{Name: "posting", Version: 1, Domain: "finance",
			Fields: []core.FieldDef{
				{Name: "entry", Type: "string", Required: true},
				{Name: "line", Type: "int", Required: true},
				{Name: "account", Type: "ref<account>", Required: true},
				{Name: "side", Type: "enum", Values: []string{"debit", "credit"}, Required: true},
				{Name: "amount", Type: "money", Required: true},
				{Name: "currency", Type: "string", Required: true},
				{Name: "date", Type: "date", Required: true},
			}},
		{Name: "period_lock", Version: 1, Domain: "finance",
			Fields: []core.FieldDef{{Name: "month", Type: "string", Required: true}}},
	} {
		if err := x.Store.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}

	submit := func(dedup, evType, date, payload string) {
		t.Helper()
		if _, err := x.Store.AppendEvent(ctx, core.Event{
			Kind: core.KindRaw, Type: evType, OccurredAt: date,
			Payload: []byte(payload), DedupKey: dedup,
		}); err != nil {
			t.Fatal(err)
		}
	}
	activate := func(id string, spec core.RuleSpec) {
		t.Helper()
		if _, err := x.Store.InsertRuleVersion(ctx, core.Rule{
			ID: id, Status: core.StatusDraft, Priority: 100,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: spec,
		}); err != nil {
			t.Fatal(err)
		}
		if _, _, procErrs, err := x.ApproveRule(ctx, id, "krzysztof", "2026-08-31"); err != nil {
			t.Fatal(err)
		} else if len(procErrs) > 0 {
			t.Fatalf("approve %s: %v", id, procErrs)
		}
	}

	activate("register-account", core.RuleSpec{
		Match: core.Match{EventType: "account.created"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "account",
			Fields: map[string]string{"code": "=$.code", "name": "=$.name", "type": "=$.type"}}},
	})
	activate("lock-period", core.RuleSpec{
		Match: core.Match{EventType: "period.locked"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "period_lock",
			Fields: map[string]string{"month": "=$.month"}}},
	})

	// Two charts of accounts in one population: the group CoA can only exist
	// beside the statutory one by code-prefix convention ("G" + its own codes).
	submit("acc-201", "account.created", "2026-09-01", `{"code":"201","name":"Rozrachunki z odbiorcami","type":"debtor"}`)
	submit("acc-700", "account.created", "2026-09-01", `{"code":"700","name":"Sprzedaż produktów","type":"revenue"}`)
	submit("acc-G1200", "account.created", "2026-09-01", `{"code":"G1200","name":"Trade receivables","type":"debtor"}`)
	submit("acc-G4000", "account.created", "2026-09-01", `{"code":"G4000","name":"Revenue","type":"revenue"}`)
	if n, errs := x.ProcessPending(ctx); n != 4 || len(errs) != 0 {
		t.Fatalf("accounts: booked %d, errs %v", n, errs)
	}

	// Two rule-books over the same event: the statutory explanation and the
	// group-GAAP explanation of one sale.
	activate("pl-stat-post", core.RuleSpec{
		Match: core.Match{EventType: "sale.recorded"},
		Effect: core.Effect{Postings: &core.PostingsTemplate{
			Currency: "=$.currency",
			Lines: []core.PostingLine{
				{Account: "201", Debit: "=$.gross"},
				{Account: "700", Credit: "=$.gross"},
			},
		}},
	})
	activate("group-post", core.RuleSpec{
		Match: core.Match{EventType: "sale.recorded"},
		Effect: core.Effect{Postings: &core.PostingsTemplate{
			Currency: "=$.currency",
			Lines: []core.PostingLine{
				{Account: "G1200", Debit: "=$.gross"},
				{Account: "G4000", Credit: "=$.gross"},
			},
		}},
	})

	// One sale, two explanations: both entries book from the one event.
	submit("sale-1", "sale.recorded", "2026-09-15", `{"gross":"100.00","currency":"PLN"}`)
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("sale: fired %d, errs %v", n, errs)
	}
	postings, err := x.Store.ObjectsByType(ctx, "posting")
	if err != nil {
		t.Fatal(err)
	}
	if len(postings) != 4 {
		t.Fatalf("postings = %d, want 4 (two balanced entries)", len(postings))
	}
	// The contortion, stated as an assertion: no posting says which book it
	// belongs to. The only way to split the ledger is to resolve each line's
	// account and inspect its code prefix.
	for _, p := range postings {
		if _, has := p.State["book"]; has {
			t.Fatal("posting carries a book — the vocabulary grew; rewrite this attempt")
		}
	}

	// The fail: PL law closes the statutory September while the group book
	// stays open until group reporting. Inexpressible — the lock has no idea
	// books exist, and the kernel check is global per month...
	submit("lock-09", "period.locked", "2026-09-30", `{"month":"2026-09"}`)
	if n, _ := x.ProcessPending(ctx); n != 1 {
		t.Fatal("lock did not book")
	}
	// ...so a September event that the still-open group book must explain is
	// refused outright: the one lock locks every book, and atomic multi-rule
	// firing refuses the event whole.
	submit("sale-2", "sale.recorded", "2026-09-20", `{"gross":"50.00","currency":"PLN"}`)
	n, errs := x.ProcessPending(ctx)
	if n != 0 || len(errs) == 0 || !strings.Contains(errs[0].Error(), "locked") {
		t.Fatalf("expected the global lock to refuse the event for both books: fired %d, errs %v", n, errs)
	}
	if after, _ := x.Store.ObjectsByType(ctx, "posting"); len(after) != 4 {
		t.Fatalf("postings = %d, want still 4", len(after))
	}
}
