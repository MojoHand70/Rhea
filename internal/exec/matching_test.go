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
