package exec_test

// E3 of the enterprise-structure ladder (DIRECTION): foreign-currency events
// post in transaction and functional currency. The pure-data attempt failed
// as the hollow-explanation contortion (DECISIONS 2026-10-04, git history);
// this is the earned form: the postings sub-language's `convert` clause.
// Rates are master data (fx_rate objects from rate events, append-only),
// conversion runs at firing time with integer arithmetic and a declared
// rounding stance, the rounding residue books to a declared account as an
// explicit plug line, identity conversion makes one rule explain domestic
// and foreign documents alike, a missing rate refuses the event into the
// worklist, simulation shares the path verbatim, and replay reproduces the
// converted ledger without ever re-converting.

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

func TestConvert(t *testing.T) {
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
		{Name: "posting", Version: 3, Domain: "finance",
			Fields: []core.FieldDef{
				{Name: "entry", Type: "string", Required: true},
				{Name: "line", Type: "int", Required: true},
				{Name: "book", Type: "string", Required: true},
				{Name: "account", Type: "ref<account>", Required: true},
				{Name: "side", Type: "enum", Values: []string{"debit", "credit"}, Required: true},
				{Name: "amount", Type: "money", Required: true},
				{Name: "currency", Type: "string", Required: true},
				{Name: "date", Type: "date", Required: true},
				{Name: "tx_amount", Type: "money", Required: false},
				{Name: "tx_currency", Type: "string", Required: false},
			}},
		{Name: "fx_rate", Version: 1, Domain: "finance", LabelField: "code",
			Fields: []core.FieldDef{
				{Name: "code", Type: "string", Required: true}, // "EUR/PLN/2026-09-21"
				{Name: "base", Type: "string", Required: true},
				{Name: "quote", Type: "string", Required: true},
				{Name: "date", Type: "date", Required: true},
				{Name: "rate", Type: "string", Required: true},
			}},
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
	// Rates are ordinary master data: append-only facts from rate events.
	activate("register-fx-rate", core.RuleSpec{
		Match: core.Match{EventType: "fx.rate.published"},
		Effect: core.Effect{Object: core.ObjectTemplate{Type: "fx_rate",
			Fields: map[string]string{"code": "=$.code", "base": "=$.base",
				"quote": "=$.quote", "date": "=$.date", "rate": "=$.rate"}}},
	})
	for dedup, p := range map[string]string{
		"acc-201": `{"code":"201","name":"Rozrachunki","type":"debtor"}`,
		"acc-700": `{"code":"700","name":"Sprzedaż","type":"revenue"}`,
		"acc-222": `{"code":"222","name":"VAT należny","type":"tax"}`,
		"acc-756": `{"code":"756","name":"Różnice kursowe i zaokrąglenia","type":"expense"}`,
	} {
		submit(dedup, "account.created", "2026-09-01", p)
	}
	submit("nbp-0921", "fx.rate.published", "2026-09-21",
		`{"code":"EUR/PLN/2026-09-21","base":"EUR","quote":"PLN","date":"2026-09-21","rate":"4.3215"}`)
	submit("nbp-0922", "fx.rate.published", "2026-09-22",
		`{"code":"EUR/PLN/2026-09-22","base":"EUR","quote":"PLN","date":"2026-09-22","rate":"4.3333"}`)
	if n, errs := x.ProcessPending(ctx); n != 6 || len(errs) != 0 {
		t.Fatalf("master data: booked %d, errs %v", n, errs)
	}

	// One rule, any currency: lines in transaction currency, the ledger in
	// PLN. Simulate the draft first — the dry run shares the conversion path.
	postSpec := core.RuleSpec{
		Match: core.Match{EventType: "sale.recorded"},
		Effect: core.Effect{Postings: &core.PostingsTemplate{
			Book: "pl-stat", Currency: "=$.currency",
			Convert: &core.ConvertTemplate{
				To: "PLN", Date: "=$.date", Rounding: "half_up", RoundingAccount: "756"},
			Lines: []core.PostingLine{
				{Account: "201", Debit: "=$.gross"},
				{Account: "700", Credit: "=$.net"},
				{Account: "222", Credit: "=$.vat"},
			},
		}},
	}
	submit("sale-eur", "sale.recorded", "2026-09-21",
		`{"date":"2026-09-21","currency":"EUR","net":"200.00","vat":"46.00","gross":"246.00"}`)
	if _, err := x.Store.InsertRuleVersion(ctx, core.Rule{
		ID: "pl-stat-post", Status: core.StatusDraft, Priority: 100,
		EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: "post", Spec: postSpec,
	}); err != nil {
		t.Fatal(err)
	}
	diff, err := x.SimulateRule(ctx, "pl-stat-post")
	if err != nil {
		t.Fatal(err)
	}
	if len(diff.Added) != 3 {
		t.Fatalf("simulated postings = %+v", diff.Added)
	}
	simAmounts := map[string]bool{}
	for _, o := range diff.Added {
		simAmounts[o.State["account"].(string)] = true
		if o.State["currency"] != "PLN" {
			t.Fatalf("simulated posting in %v", o.State["currency"])
		}
	}

	// Approve: 246.00 EUR at 4.3215 books 1063.09 PLN — line by line
	// (864.30 + 198.79 = 1063.09, balanced without a plug), every posting
	// carrying its transaction amount.
	if _, booked, procErrs, err := x.ApproveRule(ctx, "pl-stat-post", "krzysztof", "2026-09-30"); err != nil || booked != 1 || len(procErrs) != 0 {
		t.Fatalf("approve: booked %d, %v, %v", booked, procErrs, err)
	}
	wantEUR := map[string][3]int64{ // account code → functional, tx, line
		"201": {106309, 24600, 1},
		"700": {86430, 20000, 2},
		"222": {19879, 4600, 3},
	}
	assertEntry := func(dedupTag string, want map[string][3]int64, wantCount int) []core.Object {
		t.Helper()
		postings, err := x.Store.ObjectsByType(ctx, "posting")
		if err != nil {
			t.Fatal(err)
		}
		var checked int
		for _, p := range postings {
			acc, _ := x.Store.GetObject(ctx, p.State["account"].(string))
			w, ok := want[acc.State["code"].(string)]
			if !ok || int64(p.State["line"].(float64)) != w[2] {
				continue
			}
			if int64(p.State["amount"].(float64)) != w[0] || p.State["currency"] != "PLN" {
				t.Fatalf("%s %s: amount %v %v, want %d PLN", dedupTag, acc.State["code"], p.State["amount"], p.State["currency"], w[0])
			}
			if w[1] > 0 && (int64(p.State["tx_amount"].(float64)) != w[1] || p.State["tx_currency"] != "EUR") {
				t.Fatalf("%s %s: tx %v %v, want %d EUR", dedupTag, acc.State["code"], p.State["tx_amount"], p.State["tx_currency"], w[1])
			}
			checked++
		}
		if checked != wantCount {
			t.Fatalf("%s: checked %d of %d expected lines", dedupTag, checked, wantCount)
		}
		return postings
	}
	assertEntry("sale-eur", wantEUR, 3)
	for acc := range wantEUR {
		ids, _ := x.Store.FindObjectIDsByField(ctx, "account", "code", acc)
		if !simAmounts[ids[0]] {
			t.Fatalf("simulation missed account %s — dry run drifted from live", acc)
		}
	}

	// Per-line rounding that breaks the functional balance books its
	// residue to 756 as an explicit plug line: 100.01 at 4.3333 rounds to
	// 433.37 debit vs 216.67 + 216.71 = 433.38 credit — one grosz, visible.
	submit("sale-plug", "sale.recorded", "2026-09-22",
		`{"date":"2026-09-22","currency":"EUR","net":"50.00","vat":"50.01","gross":"100.01"}`)
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("plug sale: booked %d, errs %v", n, errs)
	}
	postings, _ := x.Store.ObjectsByType(ctx, "posting")
	var plug *core.Object
	for i, p := range postings {
		if int64(p.State["line"].(float64)) == 4 {
			plug = &postings[i]
		}
	}
	if plug == nil {
		t.Fatalf("no plug line among %d postings", len(postings))
	}
	acc756, _ := x.Store.FindObjectIDsByField(ctx, "account", "code", "756")
	if plug.State["account"] != acc756[0] || plug.State["side"] != "debit" ||
		int64(plug.State["amount"].(float64)) != 1 {
		t.Fatalf("plug line = %v", plug.State)
	}
	if _, has := plug.State["tx_amount"]; has {
		t.Fatal("the plug line has no transaction-currency counterpart")
	}

	// A large amount at a rate with eight places: 250 000 000.00 EUR at
	// 4.32150000 is 1 080 375 000.00 PLN. Integer mantissa arithmetic
	// overflowed int64 here and booked a negative amount; the conversion is
	// exact rational arithmetic now, and what leaves int64 refuses.
	submit("nbp-0924", "fx.rate.published", "2026-09-24",
		`{"code":"EUR/PLN/2026-09-24","base":"EUR","quote":"PLN","date":"2026-09-24","rate":"4.32150000"}`)
	submit("sale-big", "sale.recorded", "2026-09-24",
		`{"date":"2026-09-24","currency":"EUR","net":"203252032.52","vat":"46747967.48","gross":"250000000.00"}`)
	if n, errs := x.ProcessPending(ctx); n != 2 || len(errs) != 0 {
		t.Fatalf("big sale: booked %d, errs %v", n, errs)
	}
	bigWant := map[string]int64{"201": 108037500000, "700": 87835365854, "222": 20202134146}
	bigPostings, _ := x.Store.ObjectsByType(ctx, "posting")
	bigSeen := 0
	for _, p := range bigPostings {
		if p.State["date"] != "2026-09-24" {
			continue
		}
		acc, _ := x.Store.GetObject(ctx, p.State["account"].(string))
		if got := int64(p.State["amount"].(float64)); got != bigWant[acc.State["code"].(string)] {
			t.Fatalf("sale-big %s: amount %d PLN, want %d", acc.State["code"], got, bigWant[acc.State["code"].(string)])
		}
		bigSeen++
	}
	if bigSeen != 3 {
		t.Fatalf("sale-big: %d lines, want 3 and no plug", bigSeen)
	}

	// Identity conversion: the same rule explains a domestic PLN invoice —
	// currency does not route, it converts (to itself).
	submit("sale-pln", "sale.recorded", "2026-09-23",
		`{"date":"2026-09-23","currency":"PLN","net":"100.00","vat":"23.00","gross":"123.00"}`)
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 0 {
		t.Fatalf("pln sale: booked %d, errs %v", n, errs)
	}
	postings, _ = x.Store.ObjectsByType(ctx, "posting")
	var identity bool
	for _, p := range postings {
		if p.State["date"] == "2026-09-23" && p.State["side"] == "debit" {
			identity = int64(p.State["amount"].(float64)) == 12300 &&
				p.State["currency"] == "PLN" &&
				int64(p.State["tx_amount"].(float64)) == 12300 &&
				p.State["tx_currency"] == "PLN"
		}
	}
	if !identity {
		t.Fatal("identity conversion did not book 123.00 PLN with its own tx trace")
	}

	// A missing rate refuses the event into the worklist: unexplainable
	// beats guessed. The error names the exact rate the world owes us.
	submit("sale-norate", "sale.recorded", "2026-09-25",
		`{"date":"2026-09-25","currency":"EUR","net":"10.00","vat":"2.30","gross":"12.30"}`)
	if n, errs := x.ProcessPending(ctx); n != 0 || len(errs) == 0 ||
		!strings.Contains(errs[0].Error(), "no fx_rate with code = EUR/PLN/2026-09-25") {
		t.Fatalf("missing rate: booked %d, errs %v", n, errs)
	}

	// A product that leaves int64 refuses into the worklist too, named in
	// full, instead of wrapping into a negative amount.
	submit("nbp-0926", "fx.rate.published", "2026-09-26",
		`{"code":"EUR/PLN/2026-09-26","base":"EUR","quote":"PLN","date":"2026-09-26","rate":"1000000000.0"}`)
	submit("sale-overflow", "sale.recorded", "2026-09-26",
		`{"date":"2026-09-26","currency":"EUR","net":"80000000000.00","vat":"18400000000.00","gross":"98400000000.00"}`)
	if n, errs := x.ProcessPending(ctx); n != 1 || len(errs) != 2 || !strings.Contains(errs[1].Error(), "overflows money") {
		t.Fatalf("overflowing sale: booked %d, errs %v", n, errs)
	}

	// Determinism: the conversion is baked into the derived events at
	// firing time, so replay reproduces the converted ledger — plug lines
	// and all — without ever reading a rate again.
	before, _ := x.Store.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := x.Store.AllObjects(ctx)
	if len(before) == 0 || !reflect.DeepEqual(normalize(t, before), normalize(t, after)) {
		t.Fatal("replay diverged")
	}
}
