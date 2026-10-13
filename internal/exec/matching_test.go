package exec_test

// The matching engine's admission tests (DIRECTION 2026-10-11, the matching
// brief; DECISIONS 2026-10-11, the attempt): the bank story, first customer.
// One structure — a bank line against the open sales invoices of the payer —
// and one policy, the order a company chooses: by reference, exact, a
// subset that sums, exact within two grosz to a declared account, oldest
// first for a part payment. Every match is one baked event; ordinary rules
// explain it — the open amounts move by amendment over the match's items,
// the settlement objects materialize, the difference posts. What no step
// settles is unmatched, with every step's reason in the log; a payer nobody
// knows leaves the line open, never unexplained. Replay reproduces the
// state without re-matching; the simulator shares the path.

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

func bankWorld(t *testing.T) (*exec.Executor, core.Catalog) {
	t.Helper()
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	types := []core.ObjectType{
		{Name: "company", Version: 1, Domain: "finance", LabelField: "name",
			Fields: []core.FieldDef{
				{Name: "name", Type: "string", Required: true},
				{Name: "iban", Type: "string", Required: true}}},
		{Name: "account", Version: 1, Domain: "finance", LabelField: "name",
			Fields: []core.FieldDef{
				{Name: "code", Type: "string", Required: true},
				{Name: "name", Type: "string", Required: true}}},
		{Name: "posting", Version: 3, Domain: "finance",
			Fields: []core.FieldDef{
				{Name: "entry", Type: "string", Required: true},
				{Name: "line", Type: "int", Required: true},
				{Name: "book", Type: "string", Required: true},
				{Name: "account", Type: "ref<account>", Required: true},
				{Name: "side", Type: "enum", Values: []string{"debit", "credit"}, Required: true},
				{Name: "amount", Type: "money", Required: true},
				{Name: "currency", Type: "string", Required: true},
				{Name: "date", Type: "date", Required: true}}},
		{Name: "sales_invoice", Version: 1, Domain: "finance", IsDocument: true, LabelField: "number",
			Fields: []core.FieldDef{
				{Name: "number", Type: "string", Required: true},
				{Name: "issue_date", Type: "date", Required: true},
				{Name: "buyer", Type: "ref<company>", Required: true},
				{Name: "gross", Type: "money", Required: true},
				{Name: "currency", Type: "string", Required: true},
				{Name: "open", Type: "money", Required: true},
				{Name: "status", Type: "enum", Required: true, Values: []string{"open", "paid"}}},
			// open → open is the type's consent to a partial payment.
			Lifecycle: &core.LifecycleDef{Field: "status", Transitions: map[string][]string{"open": {"open", "paid"}}},
			Amendable: []string{"open"}},
		{Name: "bank_line", Version: 1, Domain: "finance", LabelField: "reference",
			Fields: []core.FieldDef{
				{Name: "date", Type: "date", Required: true},
				{Name: "amount", Type: "money", Required: true},
				{Name: "currency", Type: "string", Required: true},
				{Name: "reference", Type: "string", Required: true},
				{Name: "counterparty_account", Type: "string", Required: true},
				{Name: "open", Type: "money", Required: true},
				{Name: "status", Type: "enum", Required: true, Values: []string{"open", "settled"}}},
			Lifecycle: &core.LifecycleDef{Field: "status", Transitions: map[string][]string{"open": {"open", "settled"}}},
			Amendable: []string{"open"}},
		{Name: "settlement", Version: 1, Domain: "finance",
			Fields: []core.FieldDef{
				{Name: "payment", Type: "ref<bank_line>", Required: true},
				{Name: "invoice", Type: "ref<sales_invoice>", Required: true},
				{Name: "amount", Type: "money", Required: true},
				{Name: "date", Type: "date", Required: true}}},
		{Name: "matching_difference", Version: 1, Domain: "finance",
			Fields: []core.FieldDef{
				{Name: "matching", Type: "string", Required: true},
				{Name: "kind", Type: "string", Required: true},
				{Name: "amount", Type: "money", Required: true},
				{Name: "currency", Type: "string", Required: true},
				{Name: "direction", Type: "enum", Required: true, Values: []string{"short", "over", "up", "down"}},
				{Name: "account", Type: "string", Required: true},
				{Name: "date", Type: "date", Required: true}}},
	}
	byName := map[string]core.ObjectType{}
	for _, ot := range types {
		if err := x.Store.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
		byName[ot.Name] = ot
	}
	return x, func(name string) (core.ObjectType, bool) { ot, ok := byName[name]; return ot, ok }
}

// incomingPayments is the structure and the policy of the first customer.
func incomingPayments() core.RuleSpec {
	return core.RuleSpec{
		Match: core.Match{EventType: core.EventObjectMaterialized,
			Where: []core.Condition{{Path: "$.object_type", Op: "eq", Value: "bank_line"}}},
		Effect: core.Effect{Matching: &core.MatchingTemplate{
			Name: "incoming-payments", Trigger: "payment",
			Sides: []core.MatchingSide{
				{Name: "payment", Type: "bank_line", Open: map[string]string{"money": "open"},
					Currency: "currency", Date: "date", Reference: "reference"},
				{Name: "invoice", Type: "sales_invoice",
					Items:    "=objects(sales_invoice, buyer, ref(company, iban, $.state.counterparty_account))",
					Open:     map[string]string{"money": "open"},
					Currency: "currency", Date: "issue_date", Reference: "number"},
			},
			Policy: []core.MatchingStep{
				{Strategy: "reference"},
				{Strategy: "exact"},
				{Strategy: "subset", MaxItems: 4},
				{Strategy: "exact", Tolerance: map[string]string{"money": "0.02"}, DifferenceAccount: "755"},
				{Strategy: "oldest_first"},
			},
		}},
	}
}

func TestMatchingGate(t *testing.T) {
	_, catalog := bankWorld(t)
	good := incomingPayments()
	if err := good.Check(catalog); err != nil {
		t.Fatalf("the first customer's rule must pass the gate: %v", err)
	}
	refuse := func(what string, mutate func(s *core.RuleSpec), want string) {
		t.Helper()
		s := incomingPayments()
		mutate(&s)
		err := s.Check(catalog)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("%s: want %q, got %v", what, want, err)
		}
	}
	refuse("a rule not on the trigger's type", func(s *core.RuleSpec) { s.Match.Where[0].Value = "sales_invoice" },
		`triggered by "payment" fires on a bank_line, but its match names sales_invoice`)
	refuse("a tolerance without a home", func(s *core.RuleSpec) { s.Effect.Matching.Policy[3].DifferenceAccount = "" },
		"difference_account must name where it goes")
	refuse("an unknown strategy", func(s *core.RuleSpec) { s.Effect.Matching.Policy[0].Strategy = "guess" },
		`unknown strategy "guess"; the strategies are reference, exact, subset, oldest_first`)
	refuse("an open field that is not amendable", func(s *core.RuleSpec) { s.Effect.Matching.Sides[1].Open["money"] = "gross" },
		`open field "gross" of sales_invoice must be amendable`)
	refuse("a read of the wrong type", func(s *core.RuleSpec) {
		s.Effect.Matching.Sides[1].Items = "=objects(bank_line, counterparty_account, $.state.counterparty_account)"
	}, "want a collection of sales_invoice objects")
	refuse("a side with a different measure", func(s *core.RuleSpec) { s.Effect.Matching.Sides[1].Open = map[string]string{"qty": "open"} },
		"every side declares the same measures")
	refuse("two effects", func(s *core.RuleSpec) {
		s.Effect.Amend = &core.AmendTemplate{Type: "x", Target: "=$.a", Set: map[string]string{"a": "b"}}
	},
		"more than one of object/postings/amend/matching")
}

func TestMatchingBankStory(t *testing.T) {
	ctx := context.Background()
	x, _ := bankWorld(t)
	submit := func(dedup, evType, date, payload string) {
		t.Helper()
		if _, err := x.Store.AppendEvent(ctx, core.Event{
			Kind: core.KindRaw, Type: evType, OccurredAt: date,
			Payload: []byte(payload), DedupKey: dedup,
		}); err != nil {
			t.Fatal(err)
		}
	}
	activate := func(id string, priority int, spec core.RuleSpec) {
		t.Helper()
		if _, err := x.Store.InsertRuleVersion(ctx, core.Rule{
			ID: id, Status: core.StatusDraft, Priority: priority,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: spec,
		}); err != nil {
			t.Fatal(err)
		}
		if _, _, procErrs, err := x.ApproveRule(ctx, id, "krzysztof", "2026-09-30"); err != nil {
			t.Fatal(err)
		} else if len(procErrs) > 0 {
			t.Fatalf("approve %s: %v", id, procErrs)
		}
	}
	process := func(what string) {
		t.Helper()
		if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
			t.Fatalf("%s: %v", what, errs)
		}
	}
	invoice := func(number string) core.Object {
		t.Helper()
		ids, err := x.Store.FindObjectIDsByField(ctx, "sales_invoice", "number", number)
		if err != nil || len(ids) != 1 {
			t.Fatalf("invoice %s: %v %v", number, ids, err)
		}
		o, err := x.Store.GetObject(ctx, ids[0])
		if err != nil {
			t.Fatal(err)
		}
		return o
	}
	expect := func(number, status string, open int64) {
		t.Helper()
		o := invoice(number)
		if o.State["status"] != status || int64(o.State["open"].(float64)) != open {
			t.Fatalf("%s: want %s open %d, got %v open %v", number, status, open, o.State["status"], o.State["open"])
		}
	}
	line := func(dedup, date, amount, reference, iban string) {
		submit(dedup, "bank.statement.line", date, `{"date":"`+date+`","amount":"`+amount+
			`","currency":"PLN","reference":"`+reference+`","counterparty_account":"`+iban+`"}`)
	}
	lastLine := func() core.Object {
		t.Helper()
		lines, err := x.Store.ObjectsByType(ctx, "bank_line")
		if err != nil || len(lines) == 0 {
			t.Fatal(err)
		}
		return lines[len(lines)-1]
	}
	matches := func(typ string) []json.RawMessage {
		t.Helper()
		derived, err := x.Store.EventsByKind(ctx, core.KindDerived)
		if err != nil {
			t.Fatal(err)
		}
		var out []json.RawMessage
		for _, ev := range derived {
			if ev.Type == typ {
				out = append(out, ev.Payload)
			}
		}
		return out
	}

	activate("register-company", 10, core.RuleSpec{Match: core.Match{EventType: "company.registered"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "company",
			Fields: map[string]string{"name": "=$.name", "iban": "=$.iban"}}}})
	activate("register-account", 10, core.RuleSpec{Match: core.Match{EventType: "account.created"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "account",
			Fields: map[string]string{"code": "=$.code", "name": "=$.name"}}}})
	activate("book-sales-invoice", 20, core.RuleSpec{Match: core.Match{EventType: "sales.invoice.issued"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "sales_invoice",
			Fields: map[string]string{"number": "=$.number", "issue_date": "=$.issue_date",
				"buyer": "=ref(company, iban, $.buyer_iban)", "gross": "=$.gross", "currency": "=$.currency",
				"open": "=$.gross", "status": "open"}}}})
	activate("bank-line", 20, core.RuleSpec{Match: core.Match{EventType: "bank.statement.line"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "bank_line",
			Fields: map[string]string{"date": "=$.date", "amount": "=$.amount", "currency": "=$.currency",
				"reference": "=$.reference", "counterparty_account": "=$.counterparty_account",
				"open": "=$.amount", "status": "open"}}}})
	// The structure and the policy — the company's choice, as data.
	activate("match-incoming-payments", 30, incomingPayments())
	// What a match means, in ordinary rules: the open amounts move over the
	// match's items, the settlement materializes, a difference becomes an
	// object that posts.
	onMatch := core.Match{EventType: core.EventObjectsMatched,
		Where: []core.Condition{{Path: "$.matching", Op: "eq", Value: "incoming-payments"}}}
	activate("settle-invoice", 40, core.RuleSpec{Match: onMatch,
		Effect: core.Effect{Amend: &core.AmendTemplate{Type: "sales_invoice", Each: "=$.sides.invoice[*]",
			Target: "=$.line.object_id",
			Set: map[string]string{"open": "=$.line.open_after.money",
				"status": `=if($.line.open_after.money = 0, "paid", "open")`}}}})
	activate("settle-line", 40, core.RuleSpec{Match: onMatch,
		Effect: core.Effect{Amend: &core.AmendTemplate{Type: "bank_line", Each: "=$.sides.payment[*]",
			Target: "=$.line.object_id",
			Set: map[string]string{"open": "=$.line.open_after.money",
				"status": `=if($.line.open_after.money = 0, "settled", "open")`}}}})
	activate("record-settlement", 40, core.RuleSpec{Match: onMatch,
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "settlement", Each: "=$.sides.invoice[*]",
			Fields: map[string]string{"payment": "=$.doc.trigger.object_id", "invoice": "=$.line.object_id",
				"amount": "=$.line.allocated.money", "date": "=$.root.date"}}}})
	activate("record-difference", 40, core.RuleSpec{
		Match: core.Match{EventType: core.EventObjectsMatched, Where: []core.Condition{
			{Path: "$.matching", Op: "eq", Value: "incoming-payments"},
			{Path: "$.differences[0]", Op: "exists"}}},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "matching_difference", Each: "=$.differences[*]",
			Fields: map[string]string{"matching": "=$.doc.matching", "kind": "=$.line.kind", "amount": "=$.line.amount",
				"currency": "=$.line.currency", "direction": "=$.line.direction", "account": "=$.line.account",
				"date": "=$.root.date"}}}})
	activate("post-short-difference", 50, core.RuleSpec{
		Match: core.Match{EventType: core.EventObjectMaterialized, Where: []core.Condition{
			{Path: "$.object_type", Op: "eq", Value: "matching_difference"},
			{Path: "$.state.direction", Op: "eq", Value: "short"}}},
		Effect: core.Effect{Postings: &core.PostingsTemplate{Currency: "=$.state.currency", Lines: []core.PostingLine{
			{Account: "=$.state.account", Debit: "=$.state.amount"},
			{Account: "201", Credit: "=$.state.amount"}}}}})

	submit("nova", "company.registered", "2026-09-01", `{"name":"Nova","iban":"PL60"}`)
	submit("orion", "company.registered", "2026-09-01", `{"name":"Orion","iban":"PL61"}`)
	submit("vega", "company.registered", "2026-09-01", `{"name":"Vega","iban":"PL62"}`)
	for _, a := range [][2]string{{"201", "Należności"}, {"755", "Koszty finansowe"}} {
		submit("acc-"+a[0], "account.created", "2026-09-01", `{"code":"`+a[0]+`","name":"`+a[1]+`"}`)
	}
	issue := func(dedup, number, date, iban, gross string) {
		submit(dedup, "sales.invoice.issued", date, `{"number":"`+number+`","issue_date":"`+date+
			`","buyer_iban":"`+iban+`","gross":"`+gross+`","currency":"PLN"}`)
	}
	issue("fv1", "FV/1", "2026-09-02", "PL60", "123.00")
	issue("fv2", "FV/2", "2026-09-03", "PL60", "200.00")
	issue("fv3", "FV/3", "2026-09-04", "PL60", "300.00")
	issue("fv4", "FV/4", "2026-09-05", "PL61", "123.00") // Orion owes the same 123.00
	issue("fv5", "FV/5", "2026-09-06", "PL60", "110.00")
	issue("fv6", "FV/6", "2026-09-07", "PL60", "260.00")
	issue("fv7", "FV/7", "2026-09-08", "PL60", "380.00")
	issue("fv8", "FV/8", "2026-09-09", "PL60", "99.99")
	issue("fv9", "FV/9", "2026-09-10", "PL60", "500.00")
	issue("fv10", "FV/10", "2026-09-11", "PL60", "600.00")
	process("world")

	// By reference: the rule of today, now one step of a policy.
	line("l1", "2026-09-12", "200.00", "zapłata FV/2", "PL60")
	process("by reference")
	expect("FV/2", "paid", 0)
	if l := lastLine(); l.State["status"] != "settled" || l.State["open"].(float64) != 0 {
		t.Fatalf("the line settles too: %v", l.State)
	}
	// Exact, keyed by the payer: Orion's identical 123.00 is no candidate.
	line("l2", "2026-09-13", "300.00", "przelew", "PL60")
	line("l3", "2026-09-14", "123.00", "przelew", "PL60")
	process("exact")
	expect("FV/3", "paid", 0)
	expect("FV/1", "paid", 0)
	expect("FV/4", "open", 12300)
	// A subset that sums: three invoices at once, the fewest that explain it.
	line("l4", "2026-09-15", "750.00", "przelew", "PL60")
	process("subset")
	expect("FV/5", "paid", 0)
	expect("FV/6", "paid", 0)
	expect("FV/7", "paid", 0)
	// Two grosz short, within the tolerance, the difference to 755.
	line("l5", "2026-09-16", "99.97", "przelew", "PL60")
	process("tolerance")
	expect("FV/8", "paid", 0)
	diffs, _ := x.Store.ObjectsByType(ctx, "matching_difference")
	if len(diffs) != 1 || diffs[0].State["amount"].(float64) != 2 || diffs[0].State["direction"] != "short" || diffs[0].State["account"] != "755" {
		t.Fatalf("the difference as an object: %v", diffs)
	}
	postings, _ := x.Store.ObjectsByType(ctx, "posting")
	if len(postings) != 2 || postings[0].State["amount"].(float64) != 2 || postings[0].State["side"] != "debit" {
		t.Fatalf("the difference posts Wn 755 / Ma 201 0.02: %v", postings)
	}
	// A part payment: the oldest open invoice whole, the next in part; the
	// invoice stays open by the transition its type declared.
	line("l6", "2026-09-17", "700.00", "przelew", "PL60")
	process("oldest first")
	expect("FV/9", "paid", 0)
	expect("FV/10", "open", 40000)
	if l := lastLine(); l.State["status"] != "settled" {
		t.Fatalf("the line is allocated whole: %v", l.State)
	}
	// The rest of it, later, exactly.
	line("l7", "2026-09-18", "400.00", "przelew", "PL60")
	process("the remainder")
	expect("FV/10", "paid", 0)
	// Nothing fits — Vega owes nothing: every step says why, the line stays
	// open for the person.
	line("l8", "2026-09-19", "50.00", "przelew", "PL62")
	process("unmatched")
	if l := lastLine(); l.State["status"] != "open" || l.State["open"].(float64) != 5000 {
		t.Fatalf("an unmatched line stays open: %v", l.State)
	}
	unmatched := matches(core.EventObjectsUnmatched)
	if len(unmatched) != 1 {
		t.Fatalf("want one unmatched event, got %d", len(unmatched))
	}
	var u core.UnmatchedObjects
	json.Unmarshal(unmatched[0], &u)
	if len(u.Tried) != 5 || !strings.Contains(u.Tried[1].Reason, "no single item on invoice opens at 50.00 (0 candidates)") || u.Tried[4].Reason != "nothing open on invoice (0 candidates)" {
		t.Fatalf("every step explains itself: %+v", u.Tried)
	}
	// A payer no company carries the account of: the line materializes and
	// waits; its own chain is never held hostage by the match.
	line("l9", "2026-09-20", "10.00", "przelew", "PL99")
	process("unknown payer")
	if l := lastLine(); l.State["counterparty_account"] != "PL99" || l.State["status"] != "open" {
		t.Fatalf("the unknown payer's line: %v", l.State)
	}
	json.Unmarshal(matches(core.EventObjectsUnmatched)[1], &u)
	if !strings.Contains(u.Reason, "invoice cannot be read: no company with iban = PL99") {
		t.Fatalf("the reason names the read: %q", u.Reason)
	}

	// The match events say what they did, baked.
	matched := matches(core.EventObjectsMatched)
	if len(matched) != 7 {
		t.Fatalf("want seven matches, got %d", len(matched))
	}
	var m core.MatchedObjects
	json.Unmarshal(matched[3], &m)
	if m.Strategy != "subset" || m.Step != 3 || len(m.Sides["invoice"]) != 3 || m.Sides["payment"][0].OpenAfter["money"].(float64) != 0 {
		t.Fatalf("the subset match: %+v", m)
	}
	if ids, _ := m.Inputs["candidates.invoice"].([]any); len(ids) != 9 {
		t.Fatalf("the match records what it read — every invoice under the key: %v", m.Inputs)
	}
	json.Unmarshal(matched[4], &m)
	if m.Strategy != "exact" || m.Step != 4 || len(m.Differences) != 1 || m.Differences[0].Amount.(float64) != 2 || m.Differences[0].Currency != "PLN" {
		t.Fatalf("the tolerated match: %+v", m)
	}
	settlements, _ := x.Store.ObjectsByType(ctx, "settlement")
	if len(settlements) != 10 {
		t.Fatalf("want ten settlements (1+1+1+3+1+2+1), got %d", len(settlements))
	}

	// Determinism: replay re-applies the baked amendments and materializations
	// and never re-matches.
	before, err := x.Store.AllObjects(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, err := x.Store.AllObjects(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("replay does not reproduce the matched state")
	}

	// The simulator shares the path: a would-be payment of FV/4 shows the
	// invoice paid and the settlement added, nothing booked.
	diff, err := x.SimulateEvents(ctx, []core.Event{{ID: 1 << 40, Kind: core.KindRaw, Type: "bank.statement.line",
		OccurredAt: "2026-09-21", Payload: []byte(`{"date":"2026-09-21","amount":"123.00","currency":"PLN","reference":"FV/4","counterparty_account":"PL61"}`)}})
	if err != nil {
		t.Fatal(err)
	}
	paid := false
	for _, c := range diff.Changed {
		if c.After.ID == invoice("FV/4").ID && c.After.State["status"] == "paid" {
			paid = true
		}
	}
	if !paid || len(diff.Added) != 2 {
		t.Fatalf("the dry run: changed %d, added %d, FV/4 paid %v", len(diff.Changed), len(diff.Added), paid)
	}
	expect("FV/4", "open", 12300)
}

// The second customer: an EUR invoice paid in PLN. The structure converts
// at the rate of the payment's date (convert's rate source), the invoice
// settles in EUR, the payment in PLN, and the rate's move since the invoice
// was booked becomes an fx difference that posts — a gain on a receivable
// when the rate went up. Replay reproduces it without a rate in sight.
func TestMatchingFX(t *testing.T) {
	ctx := context.Background()
	x, _ := bankWorld(t)
	if err := x.Store.InsertObjectType(ctx, core.ObjectType{Name: "fx_rate", Version: 1, Domain: "finance", LabelField: "code",
		Fields: []core.FieldDef{
			{Name: "code", Type: "string", Required: true},
			{Name: "base", Type: "string", Required: true},
			{Name: "quote", Type: "string", Required: true},
			{Name: "date", Type: "date", Required: true},
			{Name: "rate", Type: "string", Required: true}}}); err != nil {
		t.Fatal(err)
	}
	submit := func(dedup, evType, date, payload string) {
		t.Helper()
		if _, err := x.Store.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: evType, OccurredAt: date,
			Payload: []byte(payload), DedupKey: dedup}); err != nil {
			t.Fatal(err)
		}
	}
	activate := func(id string, priority int, spec core.RuleSpec) {
		t.Helper()
		if _, err := x.Store.InsertRuleVersion(ctx, core.Rule{ID: id, Status: core.StatusDraft, Priority: priority,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: spec}); err != nil {
			t.Fatal(err)
		}
		if _, _, procErrs, err := x.ApproveRule(ctx, id, "krzysztof", "2026-09-30"); err != nil {
			t.Fatal(err)
		} else if len(procErrs) > 0 {
			t.Fatalf("approve %s: %v", id, procErrs)
		}
	}
	activate("register-company", 10, core.RuleSpec{Match: core.Match{EventType: "company.registered"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "company", Fields: map[string]string{"name": "=$.name", "iban": "=$.iban"}}}})
	activate("register-account", 10, core.RuleSpec{Match: core.Match{EventType: "account.created"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "account", Fields: map[string]string{"code": "=$.code", "name": "=$.name"}}}})
	activate("register-fx-rate", 10, core.RuleSpec{Match: core.Match{EventType: "fx.rate.published"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "fx_rate", Fields: map[string]string{"code": "=$.code", "base": "=$.base",
			"quote": "=$.quote", "date": "=$.date", "rate": "=$.rate"}}}})
	activate("book-sales-invoice", 20, core.RuleSpec{Match: core.Match{EventType: "sales.invoice.issued"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "sales_invoice", Fields: map[string]string{"number": "=$.number",
			"issue_date": "=$.issue_date", "buyer": "=ref(company, iban, $.buyer_iban)", "gross": "=$.gross", "currency": "=$.currency",
			"open": "=$.gross", "status": "open"}}}})
	activate("bank-line", 20, core.RuleSpec{Match: core.Match{EventType: "bank.statement.line"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "bank_line", Fields: map[string]string{"date": "=$.date", "amount": "=$.amount",
			"currency": "=$.currency", "reference": "=$.reference", "counterparty_account": "=$.counterparty_account",
			"open": "=$.amount", "status": "open"}}}})
	spec := incomingPayments()
	spec.Effect.Matching.Convert = &core.MatchingConvert{Functional: "PLN", DateSide: "payment", Rounding: "half_up",
		RateUpAccount: "750", RateDownAccount: "755"}
	activate("match-incoming-payments", 30, spec)
	onMatch := core.Match{EventType: core.EventObjectsMatched, Where: []core.Condition{{Path: "$.matching", Op: "eq", Value: "incoming-payments"}}}
	activate("settle-invoice", 40, core.RuleSpec{Match: onMatch, Effect: core.Effect{Amend: &core.AmendTemplate{Type: "sales_invoice",
		Each: "=$.sides.invoice[*]", Target: "=$.line.object_id", Set: map[string]string{"open": "=$.line.open_after.money",
			"status": `=if($.line.open_after.money = 0, "paid", "open")`}}}})
	activate("settle-line", 40, core.RuleSpec{Match: onMatch, Effect: core.Effect{Amend: &core.AmendTemplate{Type: "bank_line",
		Each: "=$.sides.payment[*]", Target: "=$.line.object_id", Set: map[string]string{"open": "=$.line.open_after.money",
			"status": `=if($.line.open_after.money = 0, "settled", "open")`}}}})
	activate("record-difference", 40, core.RuleSpec{Match: core.Match{EventType: core.EventObjectsMatched, Where: []core.Condition{
		{Path: "$.matching", Op: "eq", Value: "incoming-payments"}, {Path: "$.differences[0]", Op: "exists"}}},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "matching_difference", Each: "=$.differences[*]", Fields: map[string]string{
			"matching": "=$.doc.matching", "kind": "=$.line.kind", "amount": "=$.line.amount", "currency": "=$.line.currency",
			"direction": "=$.line.direction", "account": "=$.line.account", "date": "=$.root.date"}}}})
	// A receivable settled at a higher rate: the gain credits 750 and the
	// receivable's PLN value grows to meet the cash — Wn 201 / Ma 750.
	activate("post-rate-up", 50, core.RuleSpec{Match: core.Match{EventType: core.EventObjectMaterialized, Where: []core.Condition{
		{Path: "$.object_type", Op: "eq", Value: "matching_difference"}, {Path: "$.state.direction", Op: "eq", Value: "up"}}},
		Effect: core.Effect{Postings: &core.PostingsTemplate{Currency: "=$.state.currency", Lines: []core.PostingLine{
			{Account: "201", Debit: "=$.state.amount"}, {Account: "=$.state.account", Credit: "=$.state.amount"}}}}})

	submit("nova", "company.registered", "2026-09-01", `{"name":"Nova","iban":"PL60"}`)
	for _, a := range [][2]string{{"201", "Należności"}, {"750", "Przychody finansowe"}, {"755", "Koszty finansowe"}} {
		submit("acc-"+a[0], "account.created", "2026-09-01", `{"code":"`+a[0]+`","name":"`+a[1]+`"}`)
	}
	submit("nbp-0902", "fx.rate.published", "2026-09-02", `{"code":"EUR/PLN/2026-09-02","base":"EUR","quote":"PLN","date":"2026-09-02","rate":"4.30"}`)
	submit("nbp-0921", "fx.rate.published", "2026-09-21", `{"code":"EUR/PLN/2026-09-21","base":"EUR","quote":"PLN","date":"2026-09-21","rate":"4.3215"}`)
	submit("fve1", "sales.invoice.issued", "2026-09-02", `{"number":"FV/E1","issue_date":"2026-09-02","buyer_iban":"PL60","gross":"1000.00","currency":"EUR"}`)
	if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
		t.Fatal(errs)
	}
	// 1000 EUR at the payment date's 4.3215 is exactly 4321.50 PLN.
	submit("l1", "bank.statement.line", "2026-09-21", `{"date":"2026-09-21","amount":"4321.50","currency":"PLN","reference":"przelew","counterparty_account":"PL60"}`)
	if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
		t.Fatal(errs)
	}
	ids, _ := x.Store.FindObjectIDsByField(ctx, "sales_invoice", "number", "FV/E1")
	inv, _ := x.Store.GetObject(ctx, ids[0])
	if inv.State["status"] != "paid" || inv.State["open"].(float64) != 0 {
		t.Fatalf("the EUR invoice: %v", inv.State)
	}
	diffs, _ := x.Store.ObjectsByType(ctx, "matching_difference")
	if len(diffs) != 1 || diffs[0].State["kind"] != "fx" || diffs[0].State["amount"].(float64) != 2150 ||
		diffs[0].State["direction"] != "up" || diffs[0].State["currency"] != "PLN" || diffs[0].State["account"] != "750" {
		t.Fatalf("the rate difference as an object: %v", diffs)
	}
	postings, _ := x.Store.ObjectsByType(ctx, "posting")
	if len(postings) != 2 || postings[0].State["amount"].(float64) != 2150 || postings[0].State["currency"] != "PLN" {
		t.Fatalf("Wn 201 / Ma 750 21.50: %v", postings)
	}
	derived, _ := x.Store.EventsByKind(ctx, core.KindDerived)
	var m core.MatchedObjects
	for _, ev := range derived {
		if ev.Type == core.EventObjectsMatched {
			json.Unmarshal(ev.Payload, &m)
		}
	}
	a := m.Sides["invoice"][0]
	if a.Currency != "EUR" || a.Allocated["money"].(float64) != 100000 || a.Equivalent["money"].(float64) != 432150 || a.EquivalentCurrency != "PLN" {
		t.Fatalf("the invoice's allocation in EUR with its PLN equivalent: %+v", a)
	}
	if m.Inputs["rate EUR/PLN/2026-09-21"] != "4.3215" || m.Inputs["rate EUR/PLN/2026-09-02"] != "4.30" {
		t.Fatalf("the rates read are inputs: %v", m.Inputs)
	}
	// No rate for the day: the line waits in the worklist, nothing guessed.
	submit("l2", "bank.statement.line", "2026-09-22", `{"date":"2026-09-22","amount":"10.00","currency":"PLN","reference":"przelew","counterparty_account":"PL60"}`)
	submit("fve2", "sales.invoice.issued", "2026-09-02", `{"number":"FV/E2","issue_date":"2026-09-02","buyer_iban":"PL60","gross":"10.00","currency":"EUR"}`)
	_, errs := x.ProcessPending(ctx)
	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "no fx_rate for EUR/PLN on 2026-09-22") {
		t.Fatalf("missing rate: %v", errs)
	}
	before, _ := x.Store.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := x.Store.AllObjects(ctx)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("replay does not reproduce the converted match")
	}
}

// The third customer: netting per saldo (kompensata) for a party who is
// both customer and supplier. Nothing new in the engine — a structure whose
// trigger is our purchase invoice and whose other side is the same party's
// open sales invoices; the match is the netting document and posts
// Wn 202 / Ma 201.
func TestMatchingNetting(t *testing.T) {
	ctx := context.Background()
	x, _ := bankWorld(t)
	for _, ot := range []core.ObjectType{
		{Name: "purchase_invoice", Version: 1, Domain: "finance", IsDocument: true, LabelField: "number",
			Fields: []core.FieldDef{
				{Name: "number", Type: "string", Required: true},
				{Name: "invoice_date", Type: "date", Required: true},
				{Name: "supplier", Type: "ref<company>", Required: true},
				{Name: "gross", Type: "money", Required: true},
				{Name: "currency", Type: "string", Required: true},
				{Name: "open", Type: "money", Required: true},
				{Name: "status", Type: "enum", Required: true, Values: []string{"open", "paid"}}},
			Lifecycle: &core.LifecycleDef{Field: "status", Transitions: map[string][]string{"open": {"open", "paid"}}},
			Amendable: []string{"open"}},
		{Name: "netting", Version: 1, Domain: "finance", IsDocument: true,
			Fields: []core.FieldDef{
				{Name: "payable", Type: "ref<purchase_invoice>", Required: true},
				{Name: "receivable", Type: "ref<sales_invoice>", Required: true},
				{Name: "amount", Type: "money", Required: true},
				{Name: "date", Type: "date", Required: true}}},
	} {
		if err := x.Store.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	submit := func(dedup, evType, date, payload string) {
		t.Helper()
		if _, err := x.Store.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: evType, OccurredAt: date,
			Payload: []byte(payload), DedupKey: dedup}); err != nil {
			t.Fatal(err)
		}
	}
	activate := func(id string, priority int, spec core.RuleSpec) {
		t.Helper()
		if _, err := x.Store.InsertRuleVersion(ctx, core.Rule{ID: id, Status: core.StatusDraft, Priority: priority,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: spec}); err != nil {
			t.Fatal(err)
		}
		if _, _, procErrs, err := x.ApproveRule(ctx, id, "krzysztof", "2026-09-30"); err != nil {
			t.Fatal(err)
		} else if len(procErrs) > 0 {
			t.Fatalf("approve %s: %v", id, procErrs)
		}
	}
	activate("register-company", 10, core.RuleSpec{Match: core.Match{EventType: "company.registered"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "company", Fields: map[string]string{"name": "=$.name", "iban": "=$.iban"}}}})
	activate("register-account", 10, core.RuleSpec{Match: core.Match{EventType: "account.created"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "account", Fields: map[string]string{"code": "=$.code", "name": "=$.name"}}}})
	activate("book-sales-invoice", 20, core.RuleSpec{Match: core.Match{EventType: "sales.invoice.issued"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "sales_invoice", Fields: map[string]string{"number": "=$.number",
			"issue_date": "=$.issue_date", "buyer": "=ref(company, iban, $.buyer_iban)", "gross": "=$.gross", "currency": "=$.currency",
			"open": "=$.gross", "status": "open"}}}})
	activate("book-purchase-invoice", 20, core.RuleSpec{Match: core.Match{EventType: "purchase.invoice.received"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "purchase_invoice", Fields: map[string]string{"number": "=$.number",
			"invoice_date": "=$.invoice_date", "supplier": "=ref(company, iban, $.supplier_iban)", "gross": "=$.gross", "currency": "=$.currency",
			"open": "=$.gross", "status": "open"}}}})
	// The structure: our payable against the same party's open receivables.
	activate("net-with-receivables", 30, core.RuleSpec{
		Match: core.Match{EventType: core.EventObjectMaterialized, Where: []core.Condition{{Path: "$.object_type", Op: "eq", Value: "purchase_invoice"}}},
		Effect: core.Effect{Matching: &core.MatchingTemplate{Name: "netting", Trigger: "payable",
			Sides: []core.MatchingSide{
				{Name: "payable", Type: "purchase_invoice", Open: map[string]string{"money": "open"}, Currency: "currency", Date: "invoice_date", Reference: "number"},
				{Name: "receivable", Type: "sales_invoice", Items: "=objects(sales_invoice, buyer, $.state.supplier)",
					Open: map[string]string{"money": "open"}, Currency: "currency", Date: "issue_date", Reference: "number"}},
			Policy: []core.MatchingStep{{Strategy: "exact"}, {Strategy: "subset"}, {Strategy: "oldest_first"}}}}})
	onNetting := core.Match{EventType: core.EventObjectsMatched, Where: []core.Condition{{Path: "$.matching", Op: "eq", Value: "netting"}}}
	activate("settle-receivable", 40, core.RuleSpec{Match: onNetting, Effect: core.Effect{Amend: &core.AmendTemplate{Type: "sales_invoice",
		Each: "=$.sides.receivable[*]", Target: "=$.line.object_id", Set: map[string]string{"open": "=$.line.open_after.money",
			"status": `=if($.line.open_after.money = 0, "paid", "open")`}}}})
	activate("settle-payable", 40, core.RuleSpec{Match: onNetting, Effect: core.Effect{Amend: &core.AmendTemplate{Type: "purchase_invoice",
		Each: "=$.sides.payable[*]", Target: "=$.line.object_id", Set: map[string]string{"open": "=$.line.open_after.money",
			"status": `=if($.line.open_after.money = 0, "paid", "open")`}}}})
	activate("record-netting", 40, core.RuleSpec{Match: onNetting, Effect: core.Effect{Object: core.ObjectTemplate{Type: "netting",
		Each: "=$.sides.receivable[*]", Fields: map[string]string{"payable": "=$.doc.trigger.object_id", "receivable": "=$.line.object_id",
			"amount": "=$.line.allocated.money", "date": "=$.root.invoice_date"}}}})
	activate("post-netting", 50, core.RuleSpec{Match: core.Match{EventType: core.EventObjectMaterialized,
		Where: []core.Condition{{Path: "$.object_type", Op: "eq", Value: "netting"}}},
		Effect: core.Effect{Postings: &core.PostingsTemplate{Currency: "PLN", Lines: []core.PostingLine{
			{Account: "202", Debit: "=$.state.amount"}, {Account: "201", Credit: "=$.state.amount"}}}}})

	submit("delta", "company.registered", "2026-09-01", `{"name":"Delta","iban":"PL70"}`)
	for _, a := range [][2]string{{"201", "Należności"}, {"202", "Zobowiązania"}} {
		submit("acc-"+a[0], "account.created", "2026-09-01", `{"code":"`+a[0]+`","name":"`+a[1]+`"}`)
	}
	submit("fv1", "sales.invoice.issued", "2026-09-02", `{"number":"FV/1","issue_date":"2026-09-02","buyer_iban":"PL70","gross":"300.00","currency":"PLN"}`)
	submit("fv2", "sales.invoice.issued", "2026-09-03", `{"number":"FV/2","issue_date":"2026-09-03","buyer_iban":"PL70","gross":"200.00","currency":"PLN"}`)
	// Their invoice for 500: nets against both of ours.
	submit("p1", "purchase.invoice.received", "2026-09-10", `{"number":"D/1","invoice_date":"2026-09-10","supplier_iban":"PL70","gross":"500.00","currency":"PLN"}`)
	if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
		t.Fatal(errs)
	}
	status := func(typ, number string) (string, float64) {
		t.Helper()
		ids, _ := x.Store.FindObjectIDsByField(ctx, typ, "number", number)
		o, err := x.Store.GetObject(ctx, ids[0])
		if err != nil {
			t.Fatal(err)
		}
		return o.State["status"].(string), o.State["open"].(float64)
	}
	for _, n := range []string{"FV/1", "FV/2"} {
		if st, open := status("sales_invoice", n); st != "paid" || open != 0 {
			t.Fatalf("%s after the netting: %s %v", n, st, open)
		}
	}
	if st, open := status("purchase_invoice", "D/1"); st != "paid" || open != 0 {
		t.Fatalf("D/1 after the netting: %s %v", st, open)
	}
	// A bigger receivable, a smaller payable: the payable nets whole, the
	// receivable stays open for the rest.
	submit("fv3", "sales.invoice.issued", "2026-09-12", `{"number":"FV/3","issue_date":"2026-09-12","buyer_iban":"PL70","gross":"1000.00","currency":"PLN"}`)
	submit("p2", "purchase.invoice.received", "2026-09-15", `{"number":"D/2","invoice_date":"2026-09-15","supplier_iban":"PL70","gross":"400.00","currency":"PLN"}`)
	if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
		t.Fatal(errs)
	}
	if st, open := status("sales_invoice", "FV/3"); st != "open" || open != 60000 {
		t.Fatalf("FV/3 after the second netting: %s %v", st, open)
	}
	if st, _ := status("purchase_invoice", "D/2"); st != "paid" {
		t.Fatalf("D/2: %s", st)
	}
	nettings, _ := x.Store.ObjectsByType(ctx, "netting")
	postings, _ := x.Store.ObjectsByType(ctx, "posting")
	if len(nettings) != 3 || len(postings) != 6 {
		t.Fatalf("three nettings (300, 200, 400) each posting Wn 202 / Ma 201: %d, %d", len(nettings), len(postings))
	}
}

// The fourth customer: three-way — the purchase order, the receipts (PZ)
// and the supplier's invoice, quantities allocated and prices compared
// within a tolerance. The invoice line triggers; the order's lines and the
// receipts' lines are keyed by the order and joined on the item; two
// receipts sum to one invoice line; the price gap is recorded on the order
// line and becomes a variance that posts. The engine knows N; the
// structure is data.
func TestMatchingThreeWay(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	line := func(name string, extra ...core.FieldDef) core.ObjectType {
		fields := append([]core.FieldDef{
			{Name: "order", Type: "ref<purchase_order>", Required: true},
			{Name: "item", Type: "ref<item>", Required: true},
			{Name: "qty", Type: "int", Required: true},
			{Name: "open_qty", Type: "int", Required: true},
			{Name: "date", Type: "date", Required: true},
			{Name: "status", Type: "enum", Required: true, Values: []string{"open", "closed"}}}, extra...)
		return core.ObjectType{Name: name, Version: 1, Domain: "purchasing", Fields: fields,
			Lifecycle: &core.LifecycleDef{Field: "status", Transitions: map[string][]string{"open": {"open", "closed"}}},
			Amendable: []string{"open_qty"}}
	}
	header := func(name, numberField string) core.ObjectType {
		return core.ObjectType{Name: name, Version: 1, Domain: "purchasing", IsDocument: true, LabelField: numberField,
			Fields: []core.FieldDef{
				{Name: numberField, Type: "string", Required: true},
				{Name: "order", Type: "ref<purchase_order>"},
				{Name: "date", Type: "date", Required: true}}}
	}
	for _, ot := range []core.ObjectType{
		{Name: "item", Version: 1, Domain: "warehouse", LabelField: "sku", Fields: []core.FieldDef{{Name: "sku", Type: "string", Required: true}}},
		{Name: "account", Version: 1, Domain: "finance", LabelField: "code", Fields: []core.FieldDef{{Name: "code", Type: "string", Required: true}}},
		{Name: "posting", Version: 3, Domain: "finance", Fields: []core.FieldDef{
			{Name: "entry", Type: "string", Required: true}, {Name: "line", Type: "int", Required: true},
			{Name: "book", Type: "string", Required: true}, {Name: "account", Type: "ref<account>", Required: true},
			{Name: "side", Type: "enum", Values: []string{"debit", "credit"}, Required: true},
			{Name: "amount", Type: "money", Required: true}, {Name: "currency", Type: "string", Required: true},
			{Name: "date", Type: "date", Required: true}}},
		{Name: "purchase_order", Version: 1, Domain: "purchasing", IsDocument: true, LabelField: "number",
			Fields: []core.FieldDef{{Name: "number", Type: "string", Required: true}, {Name: "date", Type: "date", Required: true}}},
		line("po_line", core.FieldDef{Name: "unit_price", Type: "money", Required: true}),
		header("pz", "supplier_document"),
		line("pz_line", core.FieldDef{Name: "document", Type: "ref<pz>", Required: true}),
		header("purchase_invoice", "number"),
		line("invoice_line", core.FieldDef{Name: "document", Type: "ref<purchase_invoice>", Required: true},
			core.FieldDef{Name: "unit_price", Type: "money", Required: true}),
		{Name: "three_way_match", Version: 1, Domain: "purchasing", Fields: []core.FieldDef{
			{Name: "invoice_line", Type: "ref<invoice_line>", Required: true},
			{Name: "order_line", Type: "ref<po_line>", Required: true},
			{Name: "qty", Type: "int", Required: true},
			{Name: "date", Type: "date", Required: true}}},
		{Name: "price_variance", Version: 1, Domain: "purchasing", Fields: []core.FieldDef{
			{Name: "order_line", Type: "ref<po_line>", Required: true},
			{Name: "per_unit", Type: "money", Required: true},
			{Name: "amount", Type: "money", Required: true},
			{Name: "direction", Type: "enum", Required: true, Values: []string{"short", "over"}},
			{Name: "date", Type: "date", Required: true}}},
	} {
		if err := x.Store.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	submit := func(dedup, evType, date, payload string) {
		t.Helper()
		if _, err := x.Store.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: evType, OccurredAt: date,
			Payload: []byte(payload), DedupKey: dedup}); err != nil {
			t.Fatal(err)
		}
	}
	activate := func(id string, priority int, spec core.RuleSpec) {
		t.Helper()
		if _, err := x.Store.InsertRuleVersion(ctx, core.Rule{ID: id, Status: core.StatusDraft, Priority: priority,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: spec}); err != nil {
			t.Fatal(err)
		}
		if _, _, procErrs, err := x.ApproveRule(ctx, id, "krzysztof", "2026-09-30"); err != nil {
			t.Fatal(err)
		} else if len(procErrs) > 0 {
			t.Fatalf("approve %s: %v", id, procErrs)
		}
	}
	cascade := func(typ string) core.Match {
		return core.Match{EventType: core.EventObjectMaterialized, Where: []core.Condition{{Path: "$.object_type", Op: "eq", Value: typ}}}
	}
	activate("register-item", 10, core.RuleSpec{Match: core.Match{EventType: "item.created"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "item", Fields: map[string]string{"sku": "=$.sku"}}}})
	activate("register-account", 10, core.RuleSpec{Match: core.Match{EventType: "account.created"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "account", Fields: map[string]string{"code": "=$.code"}}}})
	// Documents with lines: a header from the fact, its lines by cascade.
	activate("book-order", 20, core.RuleSpec{Match: core.Match{EventType: "purchase.order.placed"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "purchase_order", Fields: map[string]string{"number": "=$.number", "date": "=$.date"}}}})
	activate("order-lines", 21, core.RuleSpec{Match: cascade("purchase_order"),
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "po_line", Each: "=$.root.lines[*]", Fields: map[string]string{
			"order": "=$.doc.object_id", "item": "=ref(item, sku, $.line.item)", "qty": "=$.line.qty", "open_qty": "=$.line.qty",
			"unit_price": "=$.line.price", "date": "=$.root.date", "status": "open"}}}})
	activate("book-pz", 20, core.RuleSpec{Match: core.Match{EventType: "delivery.received"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "pz", Fields: map[string]string{"supplier_document": "=$.document",
			"order": "=ref(purchase_order, number, $.order)", "date": "=$.date"}}}})
	activate("pz-lines", 21, core.RuleSpec{Match: cascade("pz"),
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "pz_line", Each: "=$.root.lines[*]", Fields: map[string]string{
			"document": "=$.doc.object_id", "order": "=$.doc.state.order", "item": "=ref(item, sku, $.line.item)",
			"qty": "=$.line.qty", "open_qty": "=$.line.qty", "date": "=$.root.date", "status": "open"}}}})
	activate("book-invoice", 20, core.RuleSpec{Match: core.Match{EventType: "purchase.invoice.received"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "purchase_invoice", Fields: map[string]string{"number": "=$.number",
			"order": "=ref(purchase_order, number, $.order)", "date": "=$.date"}}}})
	activate("invoice-lines", 21, core.RuleSpec{Match: cascade("purchase_invoice"),
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "invoice_line", Each: "=$.root.lines[*]", Fields: map[string]string{
			"document": "=$.doc.object_id", "order": "=$.doc.state.order", "item": "=ref(item, sku, $.line.item)",
			"qty": "=$.line.qty", "open_qty": "=$.line.qty", "unit_price": "=$.line.price", "date": "=$.root.date", "status": "open"}}}})
	// The structure: three sides, the quantity allocated, the price compared
	// within two percent, the difference to the variance account.
	threeWay := core.RuleSpec{Match: cascade("invoice_line"),
		Effect: core.Effect{Matching: &core.MatchingTemplate{Name: "three-way", Trigger: "invoice",
			Sides: []core.MatchingSide{
				{Name: "invoice", Type: "invoice_line", Open: map[string]string{"qty": "open_qty"},
					Compare: map[string]string{"price": "unit_price"}, Date: "date"},
				{Name: "order", Type: "po_line", Items: "=objects(po_line, order, $.state.order)", Join: map[string]string{"item": "item"},
					Open: map[string]string{"qty": "open_qty"}, Compare: map[string]string{"price": "unit_price"}, Date: "date"},
				{Name: "receipt", Type: "pz_line", Items: "=objects(pz_line, order, $.state.order)", Join: map[string]string{"item": "item"},
					Open: map[string]string{"qty": "open_qty"}, Date: "date"}},
			Policy: []core.MatchingStep{
				{Strategy: "exact", TolerancePercent: map[string]string{"price": "2"}, DifferenceAccount: "341"},
				{Strategy: "subset", MaxItems: 4, TolerancePercent: map[string]string{"price": "2"}, DifferenceAccount: "341"}}}}}
	activate("three-way-match", 30, threeWay)
	onMatch := core.Match{EventType: core.EventObjectsMatched, Where: []core.Condition{{Path: "$.matching", Op: "eq", Value: "three-way"}}}
	for _, side := range []struct{ name, typ string }{{"invoice", "invoice_line"}, {"order", "po_line"}, {"receipt", "pz_line"}} {
		activate("close-"+side.name, 40, core.RuleSpec{Match: onMatch, Effect: core.Effect{Amend: &core.AmendTemplate{Type: side.typ,
			Each: "=$.sides." + side.name + "[*]", Target: "=$.line.object_id", Set: map[string]string{"open_qty": "=$.line.open_after.qty",
				"status": `=if($.line.open_after.qty = 0, "closed", "open")`}}}})
	}
	activate("record-three-way", 40, core.RuleSpec{Match: onMatch, Effect: core.Effect{Object: core.ObjectTemplate{Type: "three_way_match",
		Fields: map[string]string{"invoice_line": "=$.trigger.object_id", "order_line": "=$.sides.order[0].object_id",
			"qty": "=$.sides.invoice[0].allocated.qty", "date": "=$.root.date"}}}})
	activate("record-variance", 40, core.RuleSpec{Match: core.Match{EventType: core.EventObjectsMatched, Where: []core.Condition{
		{Path: "$.matching", Op: "eq", Value: "three-way"}, {Path: "$.differences[0]", Op: "exists"}}},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "price_variance", Each: "=$.differences[*]", Fields: map[string]string{
			"order_line": "=$.line.item", "per_unit": "=$.line.amount",
			"amount":    "=round($.line.amount * $.doc.sides.invoice[0].allocated.qty / 100, 2, half_up)",
			"direction": "=$.line.direction", "date": "=$.root.date"}}}})
	// The supplier invoiced above the order's price: Wn 341 / Ma 300.
	activate("post-variance", 50, core.RuleSpec{Match: cascade("price_variance"),
		Effect: core.Effect{Postings: &core.PostingsTemplate{Currency: "PLN", Lines: []core.PostingLine{
			{Account: "341", Debit: "=$.state.amount"}, {Account: "300", Credit: "=$.state.amount"}}}}})

	submit("w", "item.created", "2026-09-01", `{"sku":"W"}`)
	submit("v", "item.created", "2026-09-01", `{"sku":"V"}`)
	for _, a := range []string{"300", "341"} {
		submit("acc-"+a, "account.created", "2026-09-01", `{"code":"`+a+`"}`)
	}
	submit("po1", "purchase.order.placed", "2026-09-01", `{"number":"PO/1","date":"2026-09-01","lines":[{"item":"W","qty":100,"price":"10.00"},{"item":"V","qty":100,"price":"5.00"}]}`)
	submit("pz1", "delivery.received", "2026-09-05", `{"document":"WZ 1","order":"PO/1","date":"2026-09-05","lines":[{"item":"W","qty":60},{"item":"V","qty":100}]}`)
	submit("pz2", "delivery.received", "2026-09-08", `{"document":"WZ 2","order":"PO/1","date":"2026-09-08","lines":[{"item":"W","qty":40}]}`)
	if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
		t.Fatal(errs)
	}
	// The invoice: W at 10.20 (within two percent of 10.00), V at 5.00.
	submit("inv1", "purchase.invoice.received", "2026-09-20", `{"number":"INV/1","order":"PO/1","date":"2026-09-20","lines":[{"item":"W","qty":100,"price":"10.20"},{"item":"V","qty":100,"price":"5.00"}]}`)
	if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
		t.Fatal(errs)
	}
	closed := func(typ string) int {
		t.Helper()
		objs, _ := x.Store.ObjectsByType(ctx, typ)
		n := 0
		for _, o := range objs {
			if o.State["status"] == "closed" && o.State["open_qty"].(float64) == 0 {
				n++
			}
		}
		return n
	}
	if closed("po_line") != 2 || closed("pz_line") != 3 || closed("invoice_line") != 2 {
		t.Fatalf("every line closed: orders %d, receipts %d, invoice %d", closed("po_line"), closed("pz_line"), closed("invoice_line"))
	}
	matches, _ := x.Store.ObjectsByType(ctx, "three_way_match")
	if len(matches) != 2 {
		t.Fatalf("one three-way match per invoice line: %d", len(matches))
	}
	variances, _ := x.Store.ObjectsByType(ctx, "price_variance")
	if len(variances) != 1 || variances[0].State["per_unit"].(float64) != 20 || variances[0].State["amount"].(float64) != 2000 || variances[0].State["direction"] != "over" {
		t.Fatalf("W invoiced 0.20 above the order on 100 units: %v", variances)
	}
	postings, _ := x.Store.ObjectsByType(ctx, "posting")
	if len(postings) != 2 || postings[0].State["amount"].(float64) != 2000 {
		t.Fatalf("Wn 341 / Ma 300 20.00: %v", postings)
	}
	derived, _ := x.Store.EventsByKind(ctx, core.KindDerived)
	var ms []core.MatchedObjects
	for _, ev := range derived {
		if ev.Type == core.EventObjectsMatched {
			var m core.MatchedObjects
			json.Unmarshal(ev.Payload, &m)
			ms = append(ms, m)
		}
	}
	// W: the exact step fails on the receipts (60 and 40), the subset finds
	// both; the join kept V's lines off W's sides.
	w := ms[0]
	if w.Strategy != "subset" || len(w.Sides["order"]) != 1 || len(w.Sides["receipt"]) != 2 || w.Sides["receipt"][0].Allocated["qty"].(float64) != 60 {
		t.Fatalf("W's match: %+v", w)
	}
	if !strings.Contains(w.Tried[0].Reason, "no single item on receipt opens at 100 (2 candidates)") {
		t.Fatalf("the exact step explains the receipts: %q", w.Tried[0].Reason)
	}
	if ids, _ := w.Inputs["candidates.order"].([]any); len(ids) != 1 || w.Inputs["join.order.item"] == nil {
		t.Fatalf("the join narrows the order's lines to W's: %v", w.Inputs)
	}
	if ms[1].Strategy != "exact" || len(ms[1].Differences) != 0 {
		t.Fatalf("V's match: %+v", ms[1])
	}
}

// The fifth customer, the generality test: intercompany settlement for
// group consolidation (KK, 2026-10-11) — three sides from a different
// corner, nothing added to the engine. Alfa's bank line from Beta settles
// Alfa's receivable and Beta's payable at once: the match is the
// elimination's evidence. The sides cross companies through refs at
// firing, as the brief's partition note says.
func TestMatchingIntercompany(t *testing.T) {
	ctx := context.Background()
	x, _ := bankWorld(t)
	for _, ot := range []core.ObjectType{
		{Name: "purchase_invoice", Version: 1, Domain: "finance", IsDocument: true, LabelField: "number",
			Fields: []core.FieldDef{
				{Name: "number", Type: "string", Required: true},
				{Name: "invoice_date", Type: "date", Required: true},
				{Name: "supplier", Type: "ref<company>", Required: true},
				{Name: "owner", Type: "ref<company>", Required: true},
				{Name: "gross", Type: "money", Required: true},
				{Name: "currency", Type: "string", Required: true},
				{Name: "open", Type: "money", Required: true},
				{Name: "status", Type: "enum", Required: true, Values: []string{"open", "paid"}}},
			Lifecycle: &core.LifecycleDef{Field: "status", Transitions: map[string][]string{"open": {"open", "paid"}}},
			Amendable: []string{"open"}},
		{Name: "bank_line", Version: 2, Domain: "finance", LabelField: "reference",
			Fields: []core.FieldDef{
				{Name: "date", Type: "date", Required: true},
				{Name: "amount", Type: "money", Required: true},
				{Name: "currency", Type: "string", Required: true},
				{Name: "reference", Type: "string", Required: true},
				{Name: "counterparty_account", Type: "string", Required: true},
				{Name: "owner", Type: "ref<company>", Required: true},
				{Name: "open", Type: "money", Required: true},
				{Name: "status", Type: "enum", Required: true, Values: []string{"open", "settled"}}},
			Lifecycle: &core.LifecycleDef{Field: "status", Transitions: map[string][]string{"open": {"open", "settled"}}},
			Amendable: []string{"open"}},
		{Name: "sales_invoice", Version: 2, Domain: "finance", IsDocument: true, LabelField: "number",
			Fields: []core.FieldDef{
				{Name: "number", Type: "string", Required: true},
				{Name: "issue_date", Type: "date", Required: true},
				{Name: "seller", Type: "ref<company>", Required: true},
				{Name: "buyer", Type: "ref<company>", Required: true},
				{Name: "gross", Type: "money", Required: true},
				{Name: "currency", Type: "string", Required: true},
				{Name: "open", Type: "money", Required: true},
				{Name: "status", Type: "enum", Required: true, Values: []string{"open", "paid"}}},
			Lifecycle: &core.LifecycleDef{Field: "status", Transitions: map[string][]string{"open": {"open", "paid"}}},
			Amendable: []string{"open"}},
		{Name: "intercompany_settlement", Version: 1, Domain: "group", Fields: []core.FieldDef{
			{Name: "payment", Type: "ref<bank_line>", Required: true},
			{Name: "receivable", Type: "ref<sales_invoice>", Required: true},
			{Name: "payable", Type: "ref<purchase_invoice>", Required: true},
			{Name: "amount", Type: "money", Required: true},
			{Name: "date", Type: "date", Required: true}}},
	} {
		if err := x.Store.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	submit := func(dedup, evType, date, payload string) {
		t.Helper()
		if _, err := x.Store.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: evType, OccurredAt: date,
			Payload: []byte(payload), DedupKey: dedup}); err != nil {
			t.Fatal(err)
		}
	}
	activate := func(id string, priority int, spec core.RuleSpec) {
		t.Helper()
		if _, err := x.Store.InsertRuleVersion(ctx, core.Rule{ID: id, Status: core.StatusDraft, Priority: priority,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: spec}); err != nil {
			t.Fatal(err)
		}
		if _, _, procErrs, err := x.ApproveRule(ctx, id, "krzysztof", "2026-09-30"); err != nil {
			t.Fatal(err)
		} else if len(procErrs) > 0 {
			t.Fatalf("approve %s: %v", id, procErrs)
		}
	}
	activate("register-company", 10, core.RuleSpec{Match: core.Match{EventType: "company.registered"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "company", Fields: map[string]string{"name": "=$.name", "iban": "=$.iban"}}}})
	activate("book-sales-invoice", 20, core.RuleSpec{Match: core.Match{EventType: "sales.invoice.issued"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "sales_invoice", Fields: map[string]string{"number": "=$.number",
			"issue_date": "=$.issue_date", "seller": "=ref(company, iban, $.seller_iban)", "buyer": "=ref(company, iban, $.buyer_iban)",
			"gross": "=$.gross", "currency": "=$.currency", "open": "=$.gross", "status": "open"}}}})
	activate("book-purchase-invoice", 20, core.RuleSpec{Match: core.Match{EventType: "purchase.invoice.received"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "purchase_invoice", Fields: map[string]string{"number": "=$.number",
			"invoice_date": "=$.invoice_date", "supplier": "=ref(company, iban, $.supplier_iban)", "owner": "=ref(company, iban, $.owner_iban)",
			"gross": "=$.gross", "currency": "=$.currency", "open": "=$.gross", "status": "open"}}}})
	activate("bank-line", 20, core.RuleSpec{Match: core.Match{EventType: "bank.statement.line"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "bank_line", Fields: map[string]string{"date": "=$.date", "amount": "=$.amount",
			"currency": "=$.currency", "reference": "=$.reference", "counterparty_account": "=$.counterparty_account",
			"owner": "=ref(company, iban, $.account)", "open": "=$.amount", "status": "open"}}}})
	// Three sides: the payment, our receivable from the payer, the payer's
	// payable to us — the payer's documents keyed by the payer, joined on us.
	activate("intercompany-settlement", 30, core.RuleSpec{
		Match: core.Match{EventType: core.EventObjectMaterialized, Where: []core.Condition{{Path: "$.object_type", Op: "eq", Value: "bank_line"}}},
		Effect: core.Effect{Matching: &core.MatchingTemplate{Name: "intercompany", Trigger: "payment",
			Sides: []core.MatchingSide{
				{Name: "payment", Type: "bank_line", Open: map[string]string{"money": "open"}, Currency: "currency", Date: "date", Reference: "reference"},
				{Name: "receivable", Type: "sales_invoice", Items: "=objects(sales_invoice, buyer, ref(company, iban, $.state.counterparty_account))",
					Join: map[string]string{"seller": "owner"}, Open: map[string]string{"money": "open"}, Currency: "currency", Date: "issue_date", Reference: "number"},
				{Name: "payable", Type: "purchase_invoice", Items: "=objects(purchase_invoice, owner, ref(company, iban, $.state.counterparty_account))",
					Join: map[string]string{"supplier": "owner"}, Open: map[string]string{"money": "open"}, Currency: "currency", Date: "invoice_date", Reference: "number"}},
			Policy: []core.MatchingStep{{Strategy: "reference"}, {Strategy: "exact"}}}}})
	onMatch := core.Match{EventType: core.EventObjectsMatched, Where: []core.Condition{{Path: "$.matching", Op: "eq", Value: "intercompany"}}}
	for _, side := range []struct{ name, typ, done string }{{"payment", "bank_line", "settled"}, {"receivable", "sales_invoice", "paid"}, {"payable", "purchase_invoice", "paid"}} {
		activate("settle-"+side.name, 40, core.RuleSpec{Match: onMatch, Effect: core.Effect{Amend: &core.AmendTemplate{Type: side.typ,
			Each: "=$.sides." + side.name + "[*]", Target: "=$.line.object_id", Set: map[string]string{"open": "=$.line.open_after.money",
				"status": `=if($.line.open_after.money = 0, "` + side.done + `", "open")`}}}})
	}
	activate("record-elimination", 40, core.RuleSpec{Match: onMatch, Effect: core.Effect{Object: core.ObjectTemplate{Type: "intercompany_settlement",
		Fields: map[string]string{"payment": "=$.trigger.object_id", "receivable": "=$.sides.receivable[0].object_id",
			"payable": "=$.sides.payable[0].object_id", "amount": "=$.sides.payment[0].allocated.money", "date": "=$.root.date"}}}})

	submit("alfa", "company.registered", "2026-09-01", `{"name":"Alfa","iban":"PL10"}`)
	submit("beta", "company.registered", "2026-09-01", `{"name":"Beta","iban":"PL20"}`)
	submit("gamma", "company.registered", "2026-09-01", `{"name":"Gamma","iban":"PL30"}`)
	// Alfa sells to Beta; Beta books the purchase; Gamma also invoices Beta with the same number.
	submit("a1", "sales.invoice.issued", "2026-09-02", `{"number":"FV/A1","issue_date":"2026-09-02","seller_iban":"PL10","buyer_iban":"PL20","gross":"1000.00","currency":"PLN"}`)
	submit("g1", "sales.invoice.issued", "2026-09-02", `{"number":"FV/A1","issue_date":"2026-09-02","seller_iban":"PL30","buyer_iban":"PL20","gross":"1000.00","currency":"PLN"}`)
	submit("b1", "purchase.invoice.received", "2026-09-05", `{"number":"FV/A1","invoice_date":"2026-09-05","supplier_iban":"PL10","owner_iban":"PL20","gross":"1000.00","currency":"PLN"}`)
	submit("b1g", "purchase.invoice.received", "2026-09-05", `{"number":"FV/A1","invoice_date":"2026-09-05","supplier_iban":"PL30","owner_iban":"PL20","gross":"1000.00","currency":"PLN"}`)
	if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
		t.Fatal(errs)
	}
	// Beta pays Alfa, naming the invoice: one match, three sides, Gamma's
	// documents kept off by the join.
	submit("l1", "bank.statement.line", "2026-09-10", `{"date":"2026-09-10","amount":"1000.00","currency":"PLN","reference":"FV/A1","counterparty_account":"PL20","account":"PL10"}`)
	if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
		t.Fatal(errs)
	}
	settlements, _ := x.Store.ObjectsByType(ctx, "intercompany_settlement")
	if len(settlements) != 1 || settlements[0].State["amount"].(float64) != 100000 {
		t.Fatalf("the elimination's evidence: %v", settlements)
	}
	paid := func(typ, number, sellerIBAN string) string {
		t.Helper()
		objs, _ := x.Store.ObjectsByType(ctx, typ)
		for _, o := range objs {
			if o.State["number"] == number {
				party := o.State["seller"]
				if typ == "purchase_invoice" {
					party = o.State["supplier"]
				}
				ids, _ := x.Store.FindObjectIDsByField(ctx, "company", "iban", sellerIBAN)
				if party == ids[0] {
					return o.State["status"].(string)
				}
			}
		}
		return "?"
	}
	if paid("sales_invoice", "FV/A1", "PL10") != "paid" || paid("purchase_invoice", "FV/A1", "PL10") != "paid" {
		t.Fatal("Alfa's receivable and Beta's payable from Alfa settle together")
	}
	if paid("sales_invoice", "FV/A1", "PL30") != "open" || paid("purchase_invoice", "FV/A1", "PL30") != "open" {
		t.Fatal("Gamma's documents, same number, same buyer, stay open: the join kept them off")
	}
	derived, _ := x.Store.EventsByKind(ctx, core.KindDerived)
	for _, ev := range derived {
		if ev.Type == core.EventObjectsMatched {
			var m core.MatchedObjects
			json.Unmarshal(ev.Payload, &m)
			if len(m.Sides) != 3 || m.Strategy != "reference" || len(m.Sides["receivable"]) != 1 || len(m.Sides["payable"]) != 1 {
				t.Fatalf("three sides by reference: %+v", m)
			}
		}
	}
}

// Unpairing (DIRECTION, the brief's invariant 5: "a wrong match is undone
// by a correcting event that restores the open amounts, never by
// deletion"). The unpairing verb seeds a reversal of the match — the
// allocations negated, the open amounts as they are now before and
// restored after, every difference flipped — into the chain under the
// match's own rule, and the same rules explain it: the invoice and the
// line open again (by the transition their types declare), a settlement
// for the negative amount, the flipped difference posting the opposite
// entry. Undone twice is refused; replay reproduces it.
func TestMatchingUnpairing(t *testing.T) {
	ctx := context.Background()
	x, _ := bankWorld(t)
	// The types consent to reopening.
	for name, done := range map[string]string{"sales_invoice": "paid", "bank_line": "settled"} {
		ot, _ := x.Store.GetObjectType(ctx, name)
		ot.Version++
		ot.Lifecycle.Transitions[done] = []string{"open"}
		if err := x.Store.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	submit := func(dedup, evType, date, payload string) {
		t.Helper()
		if _, err := x.Store.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: evType, OccurredAt: date,
			Payload: []byte(payload), DedupKey: dedup}); err != nil {
			t.Fatal(err)
		}
	}
	activate := func(id string, priority int, spec core.RuleSpec) {
		t.Helper()
		if _, err := x.Store.InsertRuleVersion(ctx, core.Rule{ID: id, Status: core.StatusDraft, Priority: priority,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: id, Spec: spec}); err != nil {
			t.Fatal(err)
		}
		if _, _, procErrs, err := x.ApproveRule(ctx, id, "krzysztof", "2026-09-30"); err != nil {
			t.Fatal(err)
		} else if len(procErrs) > 0 {
			t.Fatalf("approve %s: %v", id, procErrs)
		}
	}
	activate("register-company", 10, core.RuleSpec{Match: core.Match{EventType: "company.registered"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "company", Fields: map[string]string{"name": "=$.name", "iban": "=$.iban"}}}})
	activate("register-account", 10, core.RuleSpec{Match: core.Match{EventType: "account.created"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "account", Fields: map[string]string{"code": "=$.code", "name": "=$.name"}}}})
	activate("book-sales-invoice", 20, core.RuleSpec{Match: core.Match{EventType: "sales.invoice.issued"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "sales_invoice", Fields: map[string]string{"number": "=$.number",
			"issue_date": "=$.issue_date", "buyer": "=ref(company, iban, $.buyer_iban)", "gross": "=$.gross", "currency": "=$.currency",
			"open": "=$.gross", "status": "open"}}}})
	activate("bank-line", 20, core.RuleSpec{Match: core.Match{EventType: "bank.statement.line"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "bank_line", Fields: map[string]string{"date": "=$.date", "amount": "=$.amount",
			"currency": "=$.currency", "reference": "=$.reference", "counterparty_account": "=$.counterparty_account",
			"open": "=$.amount", "status": "open"}}}})
	activate("match-incoming-payments", 30, incomingPayments())
	onMatch := core.Match{EventType: core.EventObjectsMatched, Where: []core.Condition{{Path: "$.matching", Op: "eq", Value: "incoming-payments"}}}
	activate("settle-invoice", 40, core.RuleSpec{Match: onMatch, Effect: core.Effect{Amend: &core.AmendTemplate{Type: "sales_invoice",
		Each: "=$.sides.invoice[*]", Target: "=$.line.object_id", Set: map[string]string{"open": "=$.line.open_after.money",
			"status": `=if($.line.open_after.money = 0, "paid", "open")`}}}})
	activate("settle-line", 40, core.RuleSpec{Match: onMatch, Effect: core.Effect{Amend: &core.AmendTemplate{Type: "bank_line",
		Each: "=$.sides.payment[*]", Target: "=$.line.object_id", Set: map[string]string{"open": "=$.line.open_after.money",
			"status": `=if($.line.open_after.money = 0, "settled", "open")`}}}})
	activate("record-settlement", 40, core.RuleSpec{Match: onMatch, Effect: core.Effect{Object: core.ObjectTemplate{Type: "settlement",
		Each: "=$.sides.invoice[*]", Fields: map[string]string{"payment": "=$.doc.trigger.object_id", "invoice": "=$.line.object_id",
			"amount": "=$.line.allocated.money", "date": "=$.root.date"}}}})
	activate("record-difference", 40, core.RuleSpec{Match: core.Match{EventType: core.EventObjectsMatched, Where: []core.Condition{
		{Path: "$.matching", Op: "eq", Value: "incoming-payments"}, {Path: "$.differences[0]", Op: "exists"}}},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "matching_difference", Each: "=$.differences[*]", Fields: map[string]string{
			"matching": "=$.doc.matching", "kind": "=$.line.kind", "amount": "=$.line.amount", "currency": "=$.line.currency",
			"direction": "=$.line.direction", "account": "=$.line.account", "date": "=$.root.date"}}}})
	for _, dir := range []struct{ name, debit, credit string }{{"short", "=$.state.account", "201"}, {"over", "201", "=$.state.account"}} {
		activate("post-"+dir.name+"-difference", 50, core.RuleSpec{Match: core.Match{EventType: core.EventObjectMaterialized, Where: []core.Condition{
			{Path: "$.object_type", Op: "eq", Value: "matching_difference"}, {Path: "$.state.direction", Op: "eq", Value: dir.name}}},
			Effect: core.Effect{Postings: &core.PostingsTemplate{Currency: "=$.state.currency", Lines: []core.PostingLine{
				{Account: dir.debit, Debit: "=$.state.amount"}, {Account: dir.credit, Credit: "=$.state.amount"}}}}})
	}
	submit("nova", "company.registered", "2026-09-01", `{"name":"Nova","iban":"PL60"}`)
	for _, a := range [][2]string{{"201", "Należności"}, {"755", "Koszty finansowe"}} {
		submit("acc-"+a[0], "account.created", "2026-09-01", `{"code":"`+a[0]+`","name":"`+a[1]+`"}`)
	}
	submit("fv8", "sales.invoice.issued", "2026-09-09", `{"number":"FV/8","issue_date":"2026-09-09","buyer_iban":"PL60","gross":"99.99","currency":"PLN"}`)
	// Two grosz short, within the tolerance: matched, the difference posted.
	submit("l5", "bank.statement.line", "2026-09-16", `{"date":"2026-09-16","amount":"99.97","currency":"PLN","reference":"przelew","counterparty_account":"PL60"}`)
	if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
		t.Fatal(errs)
	}
	var matchID int64
	derived, _ := x.Store.EventsByKind(ctx, core.KindDerived)
	for _, ev := range derived {
		if ev.Type == core.EventObjectsMatched {
			matchID = ev.ID
		}
	}
	state := func(typ string) map[string]any {
		t.Helper()
		objs, _ := x.Store.ObjectsByType(ctx, typ)
		return objs[0].State
	}
	if state("sales_invoice")["status"] != "paid" || state("bank_line")["status"] != "settled" {
		t.Fatal("the match did not settle")
	}

	// The person undoes it, with a reason.
	res, err := x.TriggerActivity(ctx, "unpair_items", map[string]any{"match_event": matchID, "reason": "wrong customer"}, "shell:krzysztof", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) > 0 {
		t.Fatal(res.Errors)
	}
	inv, line := state("sales_invoice"), state("bank_line")
	if inv["status"] != "open" || inv["open"].(float64) != 9999 || line["status"] != "open" || line["open"].(float64) != 9997 {
		t.Fatalf("after the unpairing: invoice %v, line %v", inv, line)
	}
	settlements, _ := x.Store.ObjectsByType(ctx, "settlement")
	if len(settlements) != 2 || settlements[1].State["amount"].(float64) != -9999 {
		t.Fatalf("a settlement for the negative amount records the undoing: %v", settlements)
	}
	diffs, _ := x.Store.ObjectsByType(ctx, "matching_difference")
	if len(diffs) != 2 || diffs[1].State["direction"] != "over" || diffs[1].State["amount"].(float64) != 2 {
		t.Fatalf("the flipped difference: %v", diffs)
	}
	postings, _ := x.Store.ObjectsByType(ctx, "posting")
	var net int64
	for _, p := range postings {
		amt := int64(p.State["amount"].(float64))
		if p.State["side"] == "debit" {
			net += amt
		} else {
			net -= amt
		}
	}
	if len(postings) != 4 || net != 0 {
		t.Fatalf("the opposite entry reverses the first: %d postings, net %d", len(postings), net)
	}
	// The reversal names what it undoes and who undid it.
	var reversal core.MatchedObjects
	var reversalEvent core.Event
	derived, _ = x.Store.EventsByKind(ctx, core.KindDerived)
	for _, ev := range derived {
		if ev.Type == core.EventObjectsMatched && ev.ID != matchID {
			json.Unmarshal(ev.Payload, &reversal)
			reversalEvent = ev
		}
	}
	if reversal.ReversalOf != matchID || reversal.Strategy != core.StrategyReversal || reversal.Reason != "wrong customer" ||
		reversalEvent.RuleID != "match-incoming-payments" || *reversalEvent.CauseEventID != res.EventID {
		t.Fatalf("the reversal: %+v caused by %v", reversal, reversalEvent.CauseEventID)
	}
	// Undone twice is refused; undoing a reversal is refused; a non-match is refused.
	if _, err := x.TriggerActivity(ctx, "unpair_items", map[string]any{"match_event": matchID, "reason": "again"}, "shell:krzysztof", ""); err == nil || !strings.Contains(err.Error(), "already undone") {
		t.Fatalf("twice: %v", err)
	}
	if _, err := x.TriggerActivity(ctx, "unpair_items", map[string]any{"match_event": reversalEvent.ID, "reason": "undo the undo"}, "shell:krzysztof", ""); err == nil || !strings.Contains(err.Error(), "itself a reversal") {
		t.Fatalf("a reversal: %v", err)
	}
	if _, err := x.TriggerActivity(ctx, "unpair_items", map[string]any{"match_event": res.EventID, "reason": "x"}, "shell:krzysztof", ""); err == nil || !strings.Contains(err.Error(), "not a match") {
		t.Fatalf("a non-match: %v", err)
	}
	// Pairing again by hand is the way forward: the same items, matched once more.
	ids, _ := x.Store.FindObjectIDsByField(ctx, "sales_invoice", "number", "FV/8")
	lines, _ := x.Store.ObjectsByType(ctx, "bank_line")
	if _, err := x.TriggerActivity(ctx, "pair_items", map[string]any{"pairing": map[string]any{
		"matching": "match-incoming-payments", "trigger": lines[0].ID, "items": []any{ids[0]}}}, "shell:krzysztof", ""); err != nil {
		t.Fatal(err)
	}
	if inv := state("sales_invoice"); inv["open"].(float64) != 2 || inv["status"] != "open" {
		t.Fatalf("paired again by hand, 99.97 of 99.99: %v", inv)
	}
	before, _ := x.Store.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := x.Store.AllObjects(ctx)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("replay does not reproduce the reversal")
	}
}
