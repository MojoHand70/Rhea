package shell_test

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"rhea/internal/core"
)

// TestFinanceLeaseSplitsEachInstalment (KSR 5 on the same method as the
// loans, KK 2026-11-04; EXAM.md 6.1.8 b, 6.1.6 b, 7.1.9): the office enters a
// lease contract once and Rhea computes the lease's internal rate of return
// and splits every instalment into interest and capital. The asset enters
// the fixed assets against the lease liability (not construction), it
// depreciates like any other, the month's interest enters the books as the
// clock opens the next month, and the lessor's invoices — booked to the
// liability account, which the contract sets as the lessor's cost account —
// bring the liability down, so its balance is the schedule's. A lease kept
// operating in the books (UoR art. 3 ust. 6) books nothing new; a finance
// lease that is operating for tax raises a task; the same contract entered
// twice books once; replay reproduces it all.
func TestFinanceLeaseSplitsEachInstalment(t *testing.T) {
	w := newCITWorld(t)
	ctx, s, x, door := w.ctx, w.s, w.x, w.door
	door("declare_company_form", map[string]any{"company": w.self, "vat_status": "active", "books": "full", "income_tax": "cit",
		"from": "2026-01-01", "occurred_at": "2026-01-01"})
	w.submit("company.registered", "2026-01-02", "lessor", map[string]any{"name": "Leasing Polska S.A.", "kind": "supplier", "vat_id": "5280001112", "country": "PL"})
	var lessor core.Object
	for _, c := range mustObjects(t, s, ctx, "company") {
		if c.State["name"] == "Leasing Polska S.A." {
			lessor = c
		}
	}
	sums := func() map[string]int64 {
		t.Helper()
		accounts, _ := s.ObjectsByType(ctx, "account")
		code := map[string]string{}
		for _, a := range accounts {
			code[a.ID] = fmt.Sprint(a.State["code"])
		}
		ps, _ := s.ObjectsByType(ctx, "posting")
		out := map[string]int64{}
		for _, p := range ps {
			n, _ := p.State["amount"].(float64)
			out[code[fmt.Sprint(p.State["account"])]+" "+fmt.Sprint(p.State["side"])] += int64(n)
		}
		return out
	}
	monthOpens := func(date string) {
		t.Helper()
		if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: core.EventMonthOpened, OccurredAt: date,
			Payload: []byte(fmt.Sprintf(`{"month":%q,"date":%q}`, date[:7], date)), DedupKey: "time/month/" + date[:7]}); err != nil {
			t.Fatal(err)
		}
		if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
			t.Fatalf("month opens %s: %v", date, errs)
		}
	}
	money := func(o core.Object, field string) int64 {
		n, _ := o.State[field].(float64)
		return int64(n)
	}
	lease := func(number string) core.Object {
		t.Helper()
		for _, l := range mustObjects(t, s, ctx, "lease_contract") {
			if l.State["number"] == number {
				return l
			}
		}
		t.Fatalf("no lease %q", number)
		return core.Object{}
	}
	instalments := func(leaseID string) []core.Object {
		t.Helper()
		var out []core.Object
		for _, i := range mustObjects(t, s, ctx, "lease_instalment") {
			if i.State["lease"] == leaseID {
				out = append(out, i)
			}
		}
		sort.Slice(out, func(a, b int) bool { return money(out[a], "n") < money(out[b], "n") })
		return out
	}
	interests := func(leaseID string) []core.Object {
		var out []core.Object
		all, _ := s.ObjectsByType(ctx, "lease_interest")
		for _, i := range all {
			if i.State["lease"] == leaseID {
				out = append(out, i)
			}
		}
		sort.Slice(out, func(a, b int) bool { return fmt.Sprint(out[a].State["date"]) < fmt.Sprint(out[b].State["date"]) })
		return out
	}
	row := func(i core.Object) string {
		return fmt.Sprintf("%d %d %d %d %d %d %v", money(i, "n"), money(i, "opening"), money(i, "interest"), money(i, "capital"),
			money(i, "payment"), money(i, "closing"), i.State["period"])
	}

	// A car on an operating lease for tax, a finance lease in the books:
	// 100 000 net, 10 000 on handover on 15 January, 35 monthly instalments
	// of 2 700 from 15 February, a 1 000 buy-out, depreciated at 20%.
	door("record_lease", map[string]any{"number": "L/2026/1", "lessor": lessor.ID, "description": "Samochód osobowy Toyota Corolla",
		"kind": "operational", "books": "finance", "asset_value": "100000.00", "initial_fee": "10000.00", "instalment": "2700.00",
		"months": 35, "every": 1, "residual": "1000.00", "first_instalment": "2026-02-15", "start": "2026-01-15",
		"depreciation_rate": "20", "date": "2026-01-15", "occurred_at": "2026-01-15"})
	l := lease("L/2026/1")
	if l.State["books"] != "finance" || l.State["effective_rate"] != "4.0321" || l.State["next_period"] != "2026-02" ||
		money(l, "net") != 10550000 || l.State["end"] != "2028-12-15" || l.State["status"] != "active" {
		t.Fatalf("the contract: %v", l.State)
	}
	rows := instalments(l.ID)
	if len(rows) != 35 || row(rows[0]) != "1 9000000 29696 240304 270000 8759696 2026-02" || row(rows[34]) != "35 368784 1216 368784 370000 0 2028-12" {
		t.Fatalf("the split: %d rows, first %s, last %s", len(rows), row(rows[0]), row(rows[len(rows)-1]))
	}
	// The asset: among the fixed assets, from the lease, at the contract's
	// value, depreciating from February; booked against the lease liability.
	assets := mustObjects(t, s, ctx, "fixed_asset")
	if len(assets) != 1 || assets[0].State["source"] != "lease" || assets[0].State["lease"] != l.ID || money(assets[0], "initial_value") != 10000000 ||
		assets[0].State["annual_rate"] != "20" || assets[0].State["start"] != "2026-02-01" || assets[0].State["name"] != "Samochód osobowy Toyota Corolla" {
		t.Fatalf("the leased asset: %v", assets)
	}
	after := sums()
	if after["010 debit"] != 10000000 || after["242 credit"] != 10000000 || after["080 credit"] != 0 {
		t.Fatalf("the asset against the lease liability, Wn 010 / Ma 242: %v", after)
	}
	// The lessor's invoices will reduce the liability: its cost account is 242.
	for _, c := range mustObjects(t, s, ctx, "company") {
		if c.ID == lessor.ID && c.State["cost_account"] != "242" {
			t.Fatalf("the lessor's cost account: %v", c.State)
		}
	}
	// Operating for tax: the tax side waits for the next version, as a task.
	tasks, _ := s.ObjectsByType(ctx, "task")
	if len(tasks) != 1 || tasks[0].State["category"] != "taxes" || tasks[0].State["description"] != "L/2026/1" {
		t.Fatalf("the tax split task: %v", tasks)
	}
	// The handover invoice: the initial fee lowers the liability.
	w.buy("lz-0", "LZ/0/2026", "2026-01-15", "5280001112", "10000.00", "2300.00", "12300.00")
	after = sums()
	if after["242 debit"] != 1000000 {
		t.Fatalf("the initial fee against the liability, Wn 242: %v", after)
	}

	// February opens: nothing yet — the asset depreciates from February,
	// the first instalment falls in February.
	monthOpens("2026-02-01")
	after = sums()
	if after["401 debit"] != 0 || after["751 debit"] != 0 {
		t.Fatalf("January: no depreciation, no interest: %v", after)
	}
	// March opens: February's depreciation 1 666.67 and the first
	// instalment's interest 296.96 on 90 000 owed, dated the instalment.
	monthOpens("2026-03-01")
	after = sums()
	if after["401 debit"] != 166667 || after["071 credit"] != 166667 {
		t.Fatalf("a month's depreciation, Wn 401 / Ma 071: %v", after)
	}
	got := interests(l.ID)
	if len(got) != 1 || money(got[0], "amount") != 29696 || got[0].State["date"] != "2026-02-15" || got[0].State["period"] != "2026-02" {
		t.Fatalf("February's interest: %v", got)
	}
	if after["751 debit"] != 29696 || after["242 credit"] != 10000000+29696 {
		t.Fatalf("the interest, Wn 751 / Ma 242: %v", after)
	}
	if l = lease("L/2026/1"); l.State["next_period"] != "2026-03" {
		t.Fatalf("March is next: %v", l.State)
	}
	// The instalment's invoice: the liability stands at the schedule's
	// closing amount for the first period.
	w.buy("lz-1", "LZ/1/2026", "2026-02-15", "5280001112", "2700.00", "621.00", "3321.00")
	after = sums()
	if after["242 credit"]-after["242 debit"] != 8759696 {
		t.Fatalf("the liability after the first instalment is the schedule's: %d", after["242 credit"]-after["242 debit"])
	}

	// The same contract entered again is nobody's rule.
	before := sums()
	tr, err := x.TriggerActivity(ctx, "record_lease", map[string]any{"number": "L/2026/1", "lessor": lessor.ID, "description": "Samochód osobowy Toyota Corolla",
		"kind": "operational", "books": "finance", "asset_value": "100000.00", "initial_fee": "10000.00", "instalment": "2700.00",
		"months": 35, "every": 1, "residual": "1000.00", "first_instalment": "2026-02-15", "start": "2026-01-15",
		"depreciation_rate": "20", "date": "2026-03-02", "occurred_at": "2026-03-02"}, "krzysztof", "")
	if err != nil || tr.Booked != 0 {
		t.Fatalf("the contract twice: %+v %v", tr, err)
	}
	if after = sums(); !reflect.DeepEqual(before, after) || len(instalments(l.ID)) != 35 {
		t.Fatalf("the contract twice changed the books: %v", after)
	}

	// A printer on an operating lease, kept operating in the books too: the
	// contract is recorded with its rate, nothing else moves — the
	// instalments will be the lessor's invoices, costs as they come.
	door("record_lease", map[string]any{"number": "L/2026/2", "description": "Drukarka", "kind": "operational", "books": "operating",
		"asset_value": "6000.00", "instalment": "200.00", "months": 36, "every": 1, "first_instalment": "2026-02-01", "start": "2026-01-20",
		"date": "2026-01-20", "occurred_at": "2026-01-20"})
	p := lease("L/2026/2")
	if p.State["next_period"] != "" || p.State["effective_rate"] != "12.9606" || len(instalments(p.ID)) != 0 {
		t.Fatalf("an operating lease in the books: %v", p.State)
	}
	if assets = mustObjects(t, s, ctx, "fixed_asset"); len(assets) != 1 {
		t.Fatalf("an operating lease brings no asset: %d", len(assets))
	}
	if tasks, _ = s.ObjectsByType(ctx, "task"); len(tasks) != 1 {
		t.Fatalf("an operating lease in the books raises no task: %d", len(tasks))
	}

	// Replay reproduces the split, the asset, the interest and the liability.
	all, _ := s.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	again, _ := s.AllObjects(ctx)
	if !reflect.DeepEqual(all, again) {
		t.Fatal("replay does not reproduce the lease")
	}
}
