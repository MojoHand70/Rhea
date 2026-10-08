package exec_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store/storetest"
)

// The formula language's first customers (DIRECTION 2026-10-08): the PZ
// line valued at quantity × price, the PZ entry Wn 330 / Ma 300 at the
// item's standard cost read through the line's link, VAT from net × rate
// with statutory rounding, a due date from a date plus payment days, and
// the book quantity of an item in a location as a fold over its movements.
// Every computed value is baked into its derived event with the formula and
// the inputs it read, replay reproduces the state exactly, and whatever
// cannot be computed lawfully refuses into the worklist naming its law.
func TestFormulasFirstCustomers(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	for _, ot := range []core.ObjectType{
		{Name: "item", Version: 1, Domain: "warehouse", LabelField: "sku", Fields: []core.FieldDef{
			{Name: "sku", Type: "string", Required: true}, {Name: "std_cost", Type: "decimal", Required: true}}},
		{Name: "location", Version: 1, Domain: "warehouse", LabelField: "code", Fields: []core.FieldDef{
			{Name: "code", Type: "string", Required: true}}},
		{Name: "account", Version: 1, Domain: "finance", Fields: []core.FieldDef{
			{Name: "code", Type: "string", Required: true}, {Name: "name", Type: "string"}}},
		{Name: "posting", Version: 1, Domain: "finance", Fields: []core.FieldDef{
			{Name: "entry", Type: "string", Required: true}, {Name: "line", Type: "int", Required: true},
			{Name: "book", Type: "string"}, {Name: "account", Type: "ref<account>", Required: true},
			{Name: "side", Type: "enum", Values: []string{"debit", "credit"}, Required: true},
			{Name: "amount", Type: "money", Required: true}, {Name: "currency", Type: "string", Required: true},
			{Name: "date", Type: "date", Required: true}}},
		{Name: "vat_rate", Version: 1, Domain: "finance", LabelField: "code", Fields: []core.FieldDef{
			{Name: "code", Type: "string", Required: true}, {Name: "rate", Type: "int", Required: true}}},
		{Name: "pz", Version: 1, Domain: "warehouse", IsDocument: true, LabelField: "supplier_document", Fields: []core.FieldDef{
			{Name: "supplier_document", Type: "string", Required: true},
			{Name: "location", Type: "ref<location>", Required: true}, {Name: "date", Type: "date", Required: true}}},
		{Name: "pz_line", Version: 1, Domain: "warehouse", Fields: []core.FieldDef{
			{Name: "document", Type: "ref<pz>", Required: true}, {Name: "item", Type: "ref<item>", Required: true},
			{Name: "qty", Type: "int", Required: true}, {Name: "price", Type: "decimal", Required: true},
			{Name: "value", Type: "money", Required: true}}},
		{Name: "stock_movement", Version: 1, Domain: "warehouse", Fields: []core.FieldDef{
			{Name: "item", Type: "ref<item>", Required: true}, {Name: "location", Type: "ref<location>", Required: true},
			{Name: "direction", Type: "enum", Values: []string{"in", "out"}, Required: true},
			{Name: "qty", Type: "int", Required: true}, {Name: "date", Type: "date", Required: true}}},
		{Name: "sales_invoice", Version: 1, Domain: "finance", IsDocument: true, LabelField: "number", Fields: []core.FieldDef{
			{Name: "number", Type: "string", Required: true}, {Name: "net", Type: "money", Required: true},
			{Name: "vat_rate", Type: "ref<vat_rate>", Required: true}, {Name: "vat", Type: "money", Required: true},
			{Name: "gross", Type: "money", Required: true}, {Name: "issue_date", Type: "date", Required: true},
			{Name: "due_date", Type: "date", Required: true}}},
		{Name: "stock_level", Version: 1, Domain: "warehouse", Fields: []core.FieldDef{
			{Name: "item", Type: "ref<item>", Required: true}, {Name: "location", Type: "ref<location>", Required: true},
			{Name: "counted", Type: "int", Required: true}, {Name: "book", Type: "int", Required: true},
			{Name: "difference", Type: "int", Required: true}}},
	} {
		if err := s.InsertObjectType(ctx, ot); err != nil {
			t.Fatal(err)
		}
	}
	materialized := func(typ string) core.Match {
		return core.Match{EventType: core.EventObjectMaterialized,
			Where: []core.Condition{{Path: "$.object_type", Op: "eq", Value: typ}}}
	}
	book := `sum(m in objects(stock_movement, item, ref(item, sku, $.item)) where m.location = ref(location, code, $.location): if(m.direction = "in", m.qty, -m.qty))`
	rules := []struct {
		id   string
		prio int
		spec core.RuleSpec
	}{
		{"register-item", 10, core.RuleSpec{Match: core.Match{EventType: "item.created"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "item", Fields: map[string]string{"sku": "=$.sku", "std_cost": "=$.std_cost"}}}}},
		{"register-location", 10, core.RuleSpec{Match: core.Match{EventType: "location.created"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "location", Fields: map[string]string{"code": "=$.code"}}}}},
		{"register-account", 10, core.RuleSpec{Match: core.Match{EventType: "account.created"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "account", Fields: map[string]string{"code": "=$.code", "name": "=$.name"}}}}},
		{"register-vat-rate", 10, core.RuleSpec{Match: core.Match{EventType: "vat_rate.defined"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "vat_rate", Fields: map[string]string{"code": "=$.code", "rate": "=$.rate"}}}}},
		{"pz-from-delivery", 100, core.RuleSpec{Match: core.Match{EventType: "delivery.received"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "pz", Fields: map[string]string{
				"supplier_document": "=$.document", "location": "=ref(location, code, $.location)", "date": "=$.date"}}}}},
		// the delivery states the price: value = qty × price, exact at two places
		{"pz-lines", 200, core.RuleSpec{Match: materialized("pz"),
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "pz_line", Each: "=$.root.lines[*]", Fields: map[string]string{
				"document": "=$.doc.object_id", "item": "=ref(item, sku, $.line.item)",
				"qty": "=$.line.qty", "price": "=$.line.price", "value": "=$.line.qty * $.line.price"}}}}},
		{"pz-line-moves-stock", 300, core.RuleSpec{Match: materialized("pz_line"),
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "stock_movement", Fields: map[string]string{
				"item": "=$.state.item", "qty": "=$.state.qty", "direction": "in",
				"location": "=ref(location, code, $.root.location)", "date": "=$.root.date"}}}}},
		// the PZ entry at the item's standard cost, read through the line's link
		{"pz-line-posts", 310, core.RuleSpec{Match: materialized("pz_line"),
			Effect: core.Effect{Postings: &core.PostingsTemplate{Book: "pl-stat", Currency: "PLN", Lines: []core.PostingLine{
				{Account: "330", Debit: "=round($.state.qty * $.state.item.std_cost, 2, half_up)"},
				{Account: "300", Credit: "=round($.state.qty * $.state.item.std_cost, 2, half_up)"}}}}}},
		// VAT from net × rate, the rate read after a lookup, rounded to the grosz
		{"sales-invoice", 100, core.RuleSpec{Match: core.Match{EventType: "sales.invoice.issued"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "sales_invoice", Fields: map[string]string{
				"number": "=$.number", "net": "=$.net", "issue_date": "=$.issue_date",
				"vat_rate": "=ref(vat_rate, code, $.vat_rate)",
				"vat":      "=round($.net * ref(vat_rate, code, $.vat_rate).rate / 100, 2, half_up)",
				"gross":    "=$.net + round($.net * ref(vat_rate, code, $.vat_rate).rate / 100, 2, half_up)",
				"due_date": "=$.issue_date + $.payment_days"}}}}},
		// the stock count: book quantity as a fold over the item's movements here
		{"stock-count", 100, core.RuleSpec{Match: core.Match{EventType: "stock.counted"},
			Effect: core.Effect{Object: core.ObjectTemplate{Type: "stock_level", Fields: map[string]string{
				"item": "=ref(item, sku, $.item)", "location": "=ref(location, code, $.location)",
				"counted": "=$.counted", "book": "=" + book, "difference": "=$.counted - " + book}}}}},
	}
	catalog := func(name string) (core.ObjectType, bool) {
		ot, err := s.GetObjectType(ctx, name)
		return ot, err == nil
	}
	for _, r := range rules {
		if err := r.spec.Check(catalog); err != nil {
			t.Fatalf("%s: %v", r.id, err)
		}
		if _, err := s.InsertRuleVersion(ctx, core.Rule{ID: r.id, Status: core.StatusActive, Priority: r.prio,
			EffectiveFrom: "2026-01-01", CreatedBy: "test", Description: r.id, Spec: r.spec}); err != nil {
			t.Fatal(err)
		}
	}
	submit := func(typ, payload string) {
		t.Helper()
		var p map[string]any
		if err := json.Unmarshal([]byte(payload), &p); err != nil {
			t.Fatal(err)
		}
		tr, err := x.TriggerActivity(ctx, "submit_event", map[string]any{
			"event_type": typ, "occurred_at": "2026-09-14", "payload": p}, "test", "")
		if err != nil || len(tr.Errors) != 0 {
			t.Fatalf("%s: %v %v", typ, err, tr.Errors)
		}
	}
	submit("item.created", `{"sku":"WID-1","std_cost":"12.345"}`)
	submit("item.created", `{"sku":"GAD-1","std_cost":"7.1234"}`)
	submit("location.created", `{"code":"MAIN"}`)
	submit("location.created", `{"code":"OTHER"}`)
	submit("account.created", `{"code":"330","name":"Towary"}`)
	submit("account.created", `{"code":"300","name":"Rozliczenie zakupu"}`)
	submit("vat_rate.defined", `{"code":"23","rate":23}`)
	submit("delivery.received", `{"document":"WZ 77/2026","location":"MAIN","date":"2026-09-14",
		"lines":[{"item":"WID-1","qty":10,"price":"12.50"},{"item":"GAD-1","qty":4,"price":"7.00"}]}`)
	submit("sales.invoice.issued", `{"number":"FV 1/2026","net":"100.10","vat_rate":"23","issue_date":"2026-09-14","payment_days":14}`)

	// state read back from the cache is JSON: numbers are float64
	num := func(o core.Object, field string) float64 { n, _ := o.State[field].(float64); return n }
	idOf := func(typ, field, value string) string {
		ids, _ := s.FindObjectIDsByField(ctx, typ, field, value)
		if len(ids) != 1 {
			t.Fatalf("%s with %s = %s: %v", typ, field, value, ids)
		}
		return ids[0]
	}
	widget, gadget, vat23 := idOf("item", "sku", "WID-1"), idOf("item", "sku", "GAD-1"), idOf("vat_rate", "code", "23")

	// the lines: value = qty × delivered price
	lines, _ := s.ObjectsByType(ctx, "pz_line")
	if len(lines) != 2 || num(lines[0], "value") != 12500 || num(lines[1], "value") != 2800 {
		t.Fatalf("lines = %v", states(lines))
	}
	// the entry: at standard cost through the link, rounded on purpose —
	// 10 × 12.345 = 123.45; 4 × 7.1234 = 28.4936 → 28.49
	postings, _ := s.ObjectsByType(ctx, "posting")
	amounts := map[string]float64{}
	for _, p := range postings {
		amounts[p.State["entry"].(string)+"/"+p.State["side"].(string)] = num(p, "amount")
	}
	if len(postings) != 4 || amounts["entry-"+lines[0].ID+"-pz-line-posts/debit"] != 12345 || amounts["entry-"+lines[1].ID+"-pz-line-posts/credit"] != 2849 {
		t.Fatalf("postings = %v", amounts)
	}
	// VAT: 100.10 × 23% = 23.023 → 23.02; gross 123.12; due in fourteen days
	inv, _ := s.ObjectsByType(ctx, "sales_invoice")
	if len(inv) != 1 || num(inv[0], "vat") != 2302 || num(inv[0], "gross") != 12312 || inv[0].State["due_date"] != "2026-09-28" {
		t.Fatalf("invoice = %v", inv[0].State)
	}

	// the explanation is in the log: formula and inputs, baked at firing
	derived, _ := s.EventsByKind(ctx, core.KindDerived)
	var explained int
	for _, d := range derived {
		if d.Type != core.EventObjectMaterialized {
			continue
		}
		var mat core.MaterializedObject
		json.Unmarshal(d.Payload, &mat)
		switch {
		case mat.ObjectType == "posting" && mat.State["line"] == float64(1) && strings.HasPrefix(mat.ObjectID, "posting-"+lines[1].ID+"-"):
			c := mat.Calc["amount"]
			if c.Formula != "=round($.state.qty * $.state.item.std_cost, 2, half_up)" ||
				c.Inputs["$.state.qty"] != float64(4) || c.Inputs["$.state.item"] != gadget || c.Inputs["$.state.item.std_cost"] != "7.1234" {
				t.Fatalf("posting calc = %+v", mat.Calc)
			}
			explained++
		case mat.ObjectType == "sales_invoice":
			if c := mat.Calc["vat"]; c.Inputs["$.net"] != "100.10" || c.Inputs["ref(vat_rate, code, 23)"] != vat23 || c.Inputs[vat23+".rate"] != float64(23) {
				t.Fatalf("vat calc = %+v", c)
			}
			if _, copied := mat.Calc["net"]; copied {
				t.Fatalf("a copy needs no explanation: %+v", mat.Calc)
			}
			explained++
		case mat.ObjectType == "pz_line":
			if c := mat.Calc["value"]; c.Formula != "=$.line.qty * $.line.price" || len(c.Inputs) != 2 {
				t.Fatalf("line calc = %+v", c)
			}
			explained++
		}
	}
	if explained != 4 {
		t.Fatalf("explained %d computed objects, want 4", explained)
	}

	// the count: 14 in at MAIN minus nothing; OTHER holds nothing of it
	submit("stock.counted", `{"item":"WID-1","location":"MAIN","counted":9}`)
	levels, _ := s.ObjectsByType(ctx, "stock_level")
	if len(levels) != 1 || num(levels[0], "book") != 10 || num(levels[0], "difference") != -1 {
		t.Fatalf("stock level = %v", states(levels))
	}

	// what cannot be computed lawfully waits in the worklist, naming its law
	submitRaw(t, x, "wz-78", "delivery.received", "2026-09-14", `{"document":"WZ 78/2026","location":"MAIN","date":"2026-09-14",
		"lines":[{"item":"WID-1","qty":3,"price":"1.005"}]}`)
	_, errs := x.ProcessPending(ctx)
	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "does not fit a money field (two places) — declare round") {
		t.Fatalf("three-place price: %v", errs)
	}
	wl, _ := s.UnmatchedRawEvents(ctx)
	if len(wl) != 1 {
		t.Fatalf("worklist = %d", len(wl))
	}

	// the collection cap is the kernel's; above it the read refuses itself —
	// a second delivery makes two movements of the widget, one too many
	// (the unlawful delivery stays in the worklist, so the door reports it
	// on every pass from here on; the raw door and a count keep it apart)
	submitRaw(t, x, "wz-79", "delivery.received", "2026-09-15", `{"document":"WZ 79/2026","location":"MAIN","date":"2026-09-15",
		"lines":[{"item":"WID-1","qty":5,"price":"1.00"}]}`)
	x.ProcessPending(ctx)
	if moves, _ := s.ObjectsByType(ctx, "stock_movement"); len(moves) != 3 {
		t.Fatalf("movements = %d, want 3", len(moves))
	}
	x.MaxCollection = 1
	submitRaw(t, x, "count-2", "stock.counted", "2026-09-15", `{"item":"WID-1","location":"MAIN","counted":9}`)
	_, errs = x.ProcessPending(ctx)
	var capped bool
	for _, e := range errs {
		capped = capped || strings.Contains(e.Error(), "objects(stock_movement, item, "+widget+") holds 2 objects, above the cap of 1")
	}
	if !capped {
		t.Fatalf("cap: %v", errs)
	}
	x.MaxCollection = 0

	assertReplayIdentical(t, x)
}

func states(objs []core.Object) []map[string]any {
	out := make([]map[string]any, 0, len(objs))
	for _, o := range objs {
		out = append(out, o.State)
	}
	return out
}
