package exec_test

import (
	"context"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

// The matching attempt (DIRECTION 2026-10-11, "Matching: a deterministic
// engine with N sides"), kept as executable evidence. The ladder's rule: try
// the bank story with what exists — rules, lookups, folds, amendments,
// postings — and let it fail where it must. What carries: settling by
// reference (the rule of today) and, once, by a unique exact amount. What
// fails: an exact amount that is not unique refuses the bank line itself;
// a partial payment can move one open amount but cannot leave the status
// alone; a tolerance needs a posting line that exists only sometimes; a
// subset that sums and an allocation across invoices have no vocabulary
// at all — an amendment has one target, and no fold enumerates subsets.
func fresh(errs []error) []error {
	var out []error
	for _, e := range errs {
		if !strings.Contains(e.Error(), "ref is ambiguous") && !strings.Contains(e.Error(), "open→open") {
			out = append(out, e)
		}
	}
	return out
}

func TestMatchingAttempt(t *testing.T) {
	ctx := context.Background()
	x := &exec.Executor{Store: storetest.New(t)}
	for _, ot := range []core.ObjectType{
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
			Lifecycle: &core.LifecycleDef{Field: "status", Transitions: map[string][]string{"open": {"paid"}}},
			Amendable: []string{"open"}},
		{Name: "bank_line", Version: 1, Domain: "finance", LabelField: "reference",
			Fields: []core.FieldDef{
				{Name: "date", Type: "date", Required: true},
				{Name: "amount", Type: "money", Required: true},
				{Name: "currency", Type: "string", Required: true},
				{Name: "reference", Type: "string", Required: true},
				{Name: "counterparty_account", Type: "string", Required: true}}},
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
		} else if errs := fresh(procErrs); len(errs) > 0 {
			t.Fatalf("approve %s: %v", id, errs)
		}
	}
	// A refused event waits in the worklist and is retried on every pass:
	// once the ambiguous payment (step 2) and the partial one (step 3) are
	// in, their refusals recur. The later steps look at the errors that
	// are new.
	process := func() []error {
		t.Helper()
		_, errs := x.ProcessPending(ctx)
		return fresh(errs)
	}
	onBankLine := func(where ...core.Condition) core.Match {
		return core.Match{EventType: core.EventObjectMaterialized,
			Where: append([]core.Condition{{Path: "$.object_type", Op: "eq", Value: "bank_line"}}, where...)}
	}
	byNumber := func(number string) string {
		t.Helper()
		ids, err := x.Store.FindObjectIDsByField(ctx, "sales_invoice", "number", number)
		if err != nil || len(ids) != 1 {
			t.Fatalf("invoice %s: %v %v", number, ids, err)
		}
		return ids[0]
	}
	status := func(number string) (string, int64) {
		t.Helper()
		o, err := x.Store.GetObject(ctx, byNumber(number))
		if err != nil {
			t.Fatal(err)
		}
		open, _ := o.State["open"].(float64)
		return o.State["status"].(string), int64(open)
	}
	bankLines := func() int {
		t.Helper()
		n, err := x.Store.CountObjectsByType(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return n["bank_line"]
	}

	activate("register-company", 10, core.RuleSpec{Match: core.Match{EventType: "company.registered"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "company",
			Fields: map[string]string{"name": "=$.name", "iban": "=$.iban"}}}})
	activate("register-account", 10, core.RuleSpec{Match: core.Match{EventType: "account.created"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "account",
			Fields: map[string]string{"code": "=$.code", "name": "=$.name"}}}})
	// The open amount is a field the booking rule fills: at birth, the gross.
	activate("book-sales-invoice", 20, core.RuleSpec{Match: core.Match{EventType: "sales.invoice.issued"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "sales_invoice",
			Fields: map[string]string{"number": "=$.number", "issue_date": "=$.issue_date",
				"buyer": "=ref(company, iban, $.buyer_iban)", "gross": "=$.gross", "currency": "=$.currency",
				"open": "=$.gross", "status": "open"}}}})
	activate("bank-line", 20, core.RuleSpec{Match: core.Match{EventType: "bank.statement.line"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "bank_line",
			Fields: map[string]string{"date": "=$.date", "amount": "=$.amount", "currency": "=$.currency",
				"reference": "=$.reference", "counterparty_account": "=$.counterparty_account"}}}})
	submit("nova", "company.registered", "2026-09-01", `{"name":"Nova","iban":"PL60"}`)
	submit("orion", "company.registered", "2026-09-01", `{"name":"Orion","iban":"PL61"}`)
	for _, a := range [][2]string{{"130", "Bank"}, {"201", "Należności"}, {"755", "Koszty finansowe"}} {
		submit("acc-"+a[0], "account.created", "2026-09-01", `{"code":"`+a[0]+`","name":"`+a[1]+`"}`)
	}
	invoice := func(dedup, number, date, iban, gross string) {
		submit(dedup, "sales.invoice.issued", date, `{"number":"`+number+`","issue_date":"`+date+
			`","buyer_iban":"`+iban+`","gross":"`+gross+`","currency":"PLN"}`)
	}
	invoice("fv1", "FV/1", "2026-09-02", "PL60", "123.00")
	invoice("fv2", "FV/2", "2026-09-03", "PL60", "200.00")
	invoice("fv3", "FV/3", "2026-09-04", "PL60", "300.00")
	invoice("fv4", "FV/4", "2026-09-05", "PL61", "123.00") // Orion owes the same 123.00
	if n, errs := x.ProcessPending(ctx); n != 9 || len(errs) != 0 {
		t.Fatalf("world: booked %d, errs %v", n, errs)
	}
	line := func(dedup, date, amount, reference, iban string) {
		submit(dedup, "bank.statement.line", date, `{"date":"`+date+`","amount":"`+amount+
			`","currency":"PLN","reference":"`+reference+`","counterparty_account":"`+iban+`"}`)
	}

	// 1. By reference: the rule of today. A line naming an invoice settles
	// it. Carries — as the personas' bank-payments bundle already proves.
	activate("pay-by-reference", 30, core.RuleSpec{
		Match: onBankLine(core.Condition{Path: "$.state.reference", Op: "ne", Value: ""}),
		Effect: core.Effect{Amend: &core.AmendTemplate{Type: "sales_invoice",
			Target: "=ref(sales_invoice, number, $.state.reference)",
			Set:    map[string]string{"status": "paid", "open": "=ref(sales_invoice, number, $.state.reference).open - $.state.amount"}}}})
	line("l1", "2026-09-10", "200.00", "FV/2", "PL60")
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("by reference: booked %d, errs %v", n, errs)
	}
	if st, open := status("FV/2"); st != "paid" || open != 0 {
		t.Fatalf("FV/2 after its payment: %s, open %d", st, open)
	}

	// 2. By exact amount: a lookup keys on one field, so "the open invoice
	// of this customer with this amount" is "any invoice with this amount".
	// Unique once, by luck; then Orion's identical 123.00 makes the lookup
	// ambiguous, and atomicity makes the gap total: the bank line itself
	// does not materialize. Nor can a paid invoice ever leave the key.
	activate("pay-by-exact-amount", 30, core.RuleSpec{
		Match: onBankLine(core.Condition{Path: "$.state.reference", Op: "eq", Value: ""}),
		Effect: core.Effect{Amend: &core.AmendTemplate{Type: "sales_invoice",
			Target: "=ref(sales_invoice, gross, $.state.amount)",
			Set:    map[string]string{"status": "paid", "open": "0.00"}}}})
	line("l2", "2026-09-11", "300.00", "", "PL60")
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("unique exact amount: booked %d, errs %v", n, errs)
	}
	if st, _ := status("FV/3"); st != "paid" {
		t.Fatalf("FV/3 after an unnamed 300.00: %s", st)
	}
	line("l3", "2026-09-12", "123.00", "", "PL60")
	_, errs := x.ProcessPending(ctx)
	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "2 sales_invoice objects have gross = 12300, ref is ambiguous") {
		t.Fatalf("exact amount shared by two customers: want the ambiguity refusal, got %v", errs)
	}
	if bankLines() != 2 {
		t.Fatalf("the ambiguous payment's bank line materialized although its chain was refused: %d lines", bankLines())
	}

	// 3. A partial payment: one named invoice's open amount can move by an
	// amendment reading through the link — but with the lifecycle as
	// declared (open → paid) the status cannot be left where it is. That
	// much is the type's to consent to, as data: a lifecycle that declares
	// open → open admits "paid when nothing stays open, else as it is".
	// What no data expresses is the allocation across invoices — the
	// remainder settling the next oldest: an amendment has exactly one
	// target.
	activate("pay-part-by-reference", 25, core.RuleSpec{
		Match: onBankLine(core.Condition{Path: "$.state.reference", Op: "eq", Value: "FV/1"}),
		Effect: core.Effect{Amend: &core.AmendTemplate{Type: "sales_invoice",
			Target: "=ref(sales_invoice, number, $.state.reference)",
			Set: map[string]string{
				"open":   "=ref(sales_invoice, number, $.state.reference).open - $.state.amount",
				"status": `=if(ref(sales_invoice, number, $.state.reference).open - $.state.amount = 0, "paid", "open")`}}}})
	line("l4", "2026-09-13", "100.00", "FV/1", "PL60")
	_, errs = x.ProcessPending(ctx)
	if len(errs) != 2 || !strings.Contains(errs[1].Error(), "status open→open is not a declared transition") {
		t.Fatalf("partial payment: want the self-transition refusal, got %v", errs)
	}
	if st, open := status("FV/1"); st != "open" || open != 12300 {
		t.Fatalf("FV/1 after the refused partial payment: %s, open %d — nothing may half-apply", st, open)
	}

	// 4. A tolerance: the two-grosz difference needs a posting line that
	// exists only when there is a difference. Lines are not conditional
	// and amounts must be positive: the rule books a short payment and
	// refuses every exact or over one.
	activate("post-payment-with-difference", 40, core.RuleSpec{
		Match: onBankLine(core.Condition{Path: "$.state.reference", Op: "eq", Value: "FV/4"}),
		Effect: core.Effect{Postings: &core.PostingsTemplate{Currency: "PLN", Lines: []core.PostingLine{
			{Account: "130", Debit: "=$.state.amount"},
			{Account: "755", Debit: "=ref(sales_invoice, number, $.state.reference).gross - $.state.amount"},
			{Account: "201", Credit: "=ref(sales_invoice, number, $.state.reference).gross"}}}}})
	line("l5", "2026-09-14", "123.00", "FV/4", "PL61")
	errs = process()
	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "amount must be positive, got 0") {
		t.Fatalf("exact payment under a difference line: want 'amount must be positive', got %v", errs)
	}

	// 5. A subset that sums (one line paying three invoices) and 6. a
	// match across N sides (order, receipt, invoice) cannot be attempted:
	// no fold enumerates subsets, and no effect touches N objects by an
	// allocation. The attempt stops here; DECISIONS records the verdict.
}
