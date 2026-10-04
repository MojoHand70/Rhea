package exec_test

// E1 of the enterprise-structure ladder (DIRECTION): parallel accounting as
// parallel rule-books over one event log. The pure-data attempt failed on
// independent period closes and is preserved in DECISIONS 2026-10-04 and git
// history; this is the earned form. A GAAP is a rule-book: the statutory and
// group explanations of the same sale post to their own books, each book
// closes on its own (book, month) locks, a mixed event with one closed book
// is refused whole — never half-explained — and replay reproduces both
// ledgers exactly.

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

func TestParallelBooks(t *testing.T) {
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
		{Name: "posting", Version: 2, Domain: "finance",
			Fields: []core.FieldDef{
				{Name: "entry", Type: "string", Required: true},
				{Name: "line", Type: "int", Required: true},
				{Name: "book", Type: "string", Required: true},
				{Name: "account", Type: "ref<account>", Required: true},
				{Name: "side", Type: "enum", Values: []string{"debit", "credit"}, Required: true},
				{Name: "amount", Type: "money", Required: true},
				{Name: "currency", Type: "string", Required: true},
				{Name: "date", Type: "date", Required: true},
			}},
		{Name: "period_lock", Version: 2, Domain: "finance",
			Fields: []core.FieldDef{
				{Name: "month", Type: "string", Required: true},
				{Name: "book", Type: "string", Required: true}}},
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
	entry := func(book string, lines []core.PostingLine) core.Effect {
		return core.Effect{Postings: &core.PostingsTemplate{
			Book: book, Currency: "=$.currency", Lines: lines}}
	}

	activate("register-account", core.RuleSpec{
		Match: core.Match{EventType: "account.created"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "account",
			Fields: map[string]string{"code": "=$.code", "name": "=$.name", "type": "=$.type"}}},
	})
	// Closing a period is per book: the lock activity says which book it means.
	activate("lock-book-period", core.RuleSpec{
		Match: core.Match{EventType: "period.locked"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "period_lock",
			Fields: map[string]string{"month": "=$.month", "book": "=$.book"}}},
	})

	// Two charts of accounts; codes stay globally unique (DECISIONS: book-
	// scoped resolution waits for a real collision to demand it).
	submit("acc-201", "account.created", "2026-09-01", `{"code":"201","name":"Rozrachunki z odbiorcami","type":"debtor"}`)
	submit("acc-700", "account.created", "2026-09-01", `{"code":"700","name":"Sprzedaż produktów","type":"revenue"}`)
	submit("acc-G1200", "account.created", "2026-09-01", `{"code":"G1200","name":"Trade receivables","type":"debtor"}`)
	submit("acc-G4000", "account.created", "2026-09-01", `{"code":"G4000","name":"Revenue","type":"revenue"}`)
	if n, errs := x.ProcessPending(ctx); n != 4 || len(errs) != 0 {
		t.Fatalf("accounts: booked %d, errs %v", n, errs)
	}

	// The two explanations of one sale: statutory and group rule-books.
	activate("pl-stat-post", core.RuleSpec{
		Match: core.Match{EventType: "sale.recorded"},
		Effect: entry("pl-stat", []core.PostingLine{
			{Account: "201", Debit: "=$.gross"},
			{Account: "700", Credit: "=$.gross"},
		}),
	})
	activate("group-post", core.RuleSpec{
		Match: core.Match{EventType: "sale.recorded"},
		Effect: entry("group", []core.PostingLine{
			{Account: "G1200", Debit: "=$.gross"},
			{Account: "G4000", Credit: "=$.gross"},
		}),
	})
	// Book-local events: a statutory correction and a group adjustment.
	activate("pl-stat-correct", core.RuleSpec{
		Match: core.Match{EventType: "stat.correction.recorded"},
		Effect: entry("pl-stat", []core.PostingLine{
			{Account: "201", Debit: "=$.amount"},
			{Account: "700", Credit: "=$.amount"},
		}),
	})
	activate("group-adjust", core.RuleSpec{
		Match: core.Match{EventType: "group.adjustment.recorded"},
		Effect: entry("group", []core.PostingLine{
			{Account: "G1200", Debit: "=$.amount"},
			{Account: "G4000", Credit: "=$.amount"},
		}),
	})

	// One sale, two books: both entries book from the one event, and every
	// posting names its book first-class.
	submit("sale-1", "sale.recorded", "2026-09-15", `{"gross":"100.00","currency":"PLN"}`)
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("sale: fired %d, errs %v", n, errs)
	}
	balance := func(want map[string][2]int64) {
		t.Helper()
		postings, err := x.Store.ObjectsByType(ctx, "posting")
		if err != nil {
			t.Fatal(err)
		}
		got := map[string][2]int64{}
		for _, p := range postings {
			book, _ := p.State["book"].(string)
			if book == "" {
				t.Fatalf("posting %s has no book", p.ID)
			}
			amt := int64(p.State["amount"].(float64))
			dc := got[book]
			if p.State["side"] == "debit" {
				dc[0] += amt
			} else {
				dc[1] += amt
			}
			got[book] = dc
		}
		if len(got) != len(want) {
			t.Fatalf("books = %v, want %v", got, want)
		}
		for book, dc := range want {
			if got[book] != dc {
				t.Fatalf("book %s: D/C = %v, want %v", book, got[book], dc)
			}
		}
	}
	balance(map[string][2]int64{"pl-stat": {10000, 10000}, "group": {10000, 10000}})

	// Statutory September closes; the group book stays open.
	submit("lock-09-stat", "period.locked", "2026-09-30", `{"month":"2026-09","book":"pl-stat"}`)
	if n, _ := x.ProcessPending(ctx); n != 1 {
		t.Fatal("lock did not book")
	}

	// Independence, both directions: a group adjustment still books into
	// September, a statutory correction is refused — and the refusal names
	// the book.
	submit("adj-1", "group.adjustment.recorded", "2026-09-25", `{"amount":"7.00","currency":"PLN"}`)
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("open group book: fired %d, errs %v", n, errs)
	}
	submit("corr-1", "stat.correction.recorded", "2026-09-26", `{"amount":"3.00","currency":"PLN"}`)
	if n, errs := x.ProcessPending(ctx); n != 0 || len(errs) == 0 ||
		!strings.Contains(errs[0].Error(), "locked for book pl-stat") {
		t.Fatalf("closed statutory book: fired %d, errs %v", n, errs)
	}

	// A mixed event — both books match, one is closed — refuses whole: an
	// event is never half-explained. The open-book half of the story is a
	// different event in an open period (the korekta, DIRECTION's backfill
	// stance), decided by a human from the worklist.
	submit("sale-2", "sale.recorded", "2026-09-27", `{"gross":"50.00","currency":"PLN"}`)
	if n, errs := x.ProcessPending(ctx); n != 0 || len(errs) == 0 ||
		!strings.Contains(errs[0].Error(), "locked for book pl-stat") {
		t.Fatalf("mixed event in a half-closed month: fired %d, errs %v", n, errs)
	}
	balance(map[string][2]int64{"pl-stat": {10000, 10000}, "group": {10700, 10700}})

	// October is open for both books; the lock was per (book, month).
	submit("sale-3", "sale.recorded", "2026-10-02", `{"gross":"20.00","currency":"PLN"}`)
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) == 0 {
		// sale-2 keeps failing in the same pass — that error stays
		t.Fatalf("october: fired %d, errs %v", n, errs)
	}
	balance(map[string][2]int64{"pl-stat": {12000, 12000}, "group": {12700, 12700}})

	// Determinism: replay reproduces both ledgers exactly.
	before, _ := x.Store.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := x.Store.AllObjects(ctx)
	if len(before) == 0 || !reflect.DeepEqual(normalize(t, before), normalize(t, after)) {
		t.Fatal("replay diverged")
	}
}
