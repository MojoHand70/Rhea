package shell_test

import (
	"fmt"
	"testing"
)

// TestInventories (the auditor's exam, 2026-10-20, lessons I-6 and I-29 of
// financial accounting 2025 drafted by the author and read against the
// text; EXAM.md 6.1.2 b, 6.1.3 d–e, 6.1.4 b–c): a sales invoice says what
// was sold. Services go to 700 as before; goods to 730 with their purchase
// value off 330 onto 731 on the day of sale; materials and fixed assets are
// other operating revenue on 760, materials' cost off 310 onto 761. Stock
// whose net selling price fell below its value is written down to it against
// a correcting account, and the write-down returns to revenue when the
// cause ends.
func TestInventories(t *testing.T) {
	w := newCITWorld(t)
	ctx, s, door, submit := w.ctx, w.s, w.door, w.submit
	door("declare_company_form", map[string]any{"company": w.self, "vat_status": "active", "books": "full", "income_tax": "cit",
		"from": "2026-01-01", "occurred_at": "2026-08-01"})
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
	base := map[string]any{"market": "pl", "intragroup": "no", "seller_nip": "5250001111", "buyer_nip": "5260001246", "vat_rate": "23", "currency": "PLN"}
	invoice := func(key, no, date, kind, net, vat, gross string, extra map[string]any) {
		t.Helper()
		p := map[string]any{"number": no, "issue_date": date, "net": net, "vat": vat, "gross": gross}
		for k, v := range base {
			p[k] = v
		}
		if kind != "" {
			p["kind"] = kind
		}
		for k, v := range extra {
			p[k] = v
		}
		submit("sales.invoice.issued", date, key, p)
	}
	// services as before, nothing said: 700
	invoice("fv-1", "FV 1/10/2026", "2026-10-05", "", "1000.00", "230.00", "1230.00", nil)
	// goods bought for 600, sold for 900: 730 and the cost off the shelf
	invoice("fv-2", "FV 2/10/2026", "2026-10-06", "goods", "900.00", "207.00", "1107.00", map[string]any{"carrying_value": "600.00"})
	// materials sold off: other operating revenue, their cost to other operating costs
	invoice("fv-3", "FV 3/10/2026", "2026-10-07", "materials", "200.00", "46.00", "246.00", map[string]any{"carrying_value": "150.00"})
	// an old laptop sold: other operating revenue; the asset itself leaves through the disposal door
	invoice("fv-4", "FV 4/10/2026", "2026-10-08", "fixed_asset", "500.00", "115.00", "615.00", nil)
	after := sums()
	if after["700 credit"] != 100000 || after["730 credit"] != 90000 || after["760 credit"] != 70000 || after["222 credit"] != 59800 || after["201 debit"] != 319800 {
		t.Fatalf("revenues by kind: %v", after)
	}
	if after["731 debit"] != 60000 || after["330 credit"] != 60000 || after["761 debit"] != 15000 || after["310 credit"] != 15000 {
		t.Fatalf("the cost of what was sold: %v", after)
	}
	docs, _ := s.ObjectsByType(ctx, "sales_invoice")
	kinds := map[string]string{}
	for _, d := range docs {
		kinds[d.State["number"].(string)] = fmt.Sprint(d.State["kind"])
	}
	if kinds["FV 1/10/2026"] != "services" || kinds["FV 2/10/2026"] != "goods" || kinds["FV 4/10/2026"] != "fixed_asset" {
		t.Fatalf("kinds: %v", kinds)
	}

	// Stock of goods worth 5 000 now sells for 3 800: a write-down of 1 200 against the correcting account.
	door("write_down_inventory", map[string]any{"title": "Towary sezonowe", "stock_account": "330", "carrying_value": "5000.00", "net_selling_price": "3800.00",
		"date": "2026-12-31", "occurred_at": "2026-12-31"})
	wds, _ := s.ObjectsByType(ctx, "inventory_write_down")
	if len(wds) != 1 || fmt.Sprint(wds[0].State["amount"]) != "120000" || wds[0].State["status"] != "open" {
		t.Fatalf("write-down: %+v", wds)
	}
	after = sums()
	if after["761 debit"] != 15000+120000 || after["349 credit"] != 120000 {
		t.Fatalf("write-down postings: %v", after)
	}
	// A price that is below value cannot be written "up": no object, nothing booked.
	tr, err := w.x.TriggerActivity(ctx, "write_down_inventory", map[string]any{"title": "x", "stock_account": "330", "carrying_value": "100.00", "net_selling_price": "150.00",
		"date": "2026-12-31", "occurred_at": "2026-12-31"}, "krzysztof", "")
	if err != nil || tr.Booked != 0 {
		t.Fatalf("a price above value is no write-down: %+v %v", tr, err)
	}
	// The season turns: the write-down returns to revenue.
	door("reverse_inventory_write_down", map[string]any{"write_down": wds[0].ID, "date": "2027-03-31", "occurred_at": "2027-03-31"})
	wds, _ = s.ObjectsByType(ctx, "inventory_write_down")
	if wds[0].State["status"] != "reversed" || wds[0].State["reversed_on"] != "2027-03-31" {
		t.Fatalf("reversed: %v", wds[0].State)
	}
	after = sums()
	if after["349 debit"] != 120000 || after["760 credit"] != 70000+120000 {
		t.Fatalf("reversal postings: %v", after)
	}
}
