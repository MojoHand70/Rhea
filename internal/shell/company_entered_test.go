package shell_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

// TestACompanyIsEnteredByItsNIP: the Poland pack's enter_company door (v44)
// is how a company comes into the books when a person founds one or an
// office adds a client (DIRECTION 2026-10-22, "The desk in order", step 4):
// the company record takes the name, kind, NIP, address, account and legal
// form as entered; the register's REGON, KRS and its mark of the answer stay
// in the fact as its source; the same company of the same kind entered
// twice is the same company, while another company keeps it in its own
// books. It is its own event, never the company.registered a house may
// have taught a rule for, so nothing is entered twice by two rules.
func TestACompanyIsEnteredByItsNIP(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	for _, f := range []string{"finance.json", "finance_v2.json", "finance_v3.json", "finance_v4.json",
		"finance_v5.json", "finance_v6.json", "finance_v7.json", "finance_v8.json"} {
		seedFromFile(t, s, f)
	}
	x := &exec.Executor{Store: s}
	for _, p := range []string{"common/people.json", "common/periods.json", "common/documents.json", "pl/pack.json"} {
		sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", p), "test")
		if err != nil {
			t.Fatal(err)
		}
		if _, _, _, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-09-01"); err != nil {
			t.Fatal(err)
		}
	}
	a, b := store.WithCompany(ctx, "PL5260001246"), store.WithCompany(ctx, "PL7011998877")
	enter := func(ctx context.Context, in map[string]any) (exec.Triggered, error) {
		return x.TriggerActivity(ctx, "enter_company", in, "krzysztof", "")
	}
	tr, err := enter(a, map[string]any{"name": "Biuro Rachunkowe Plus sp. z o.o.", "kind": "self", "vat_id": "5260001246",
		"address": "ul. Prosta 1, 00-850 Warszawa", "iban": "PL61109010140000071219812874", "legal_form": "legal_entity",
		"regon": "146234567", "krs": "0000123456", "source": "gus,wykaz-vat", "source_ref": "GxrAG-993bbbf", "occurred_at": "2026-10-23"})
	if err != nil || tr.Booked != 1 || len(tr.Errors) > 0 {
		t.Fatalf("the office entered by its NIP: %+v %v", tr, err)
	}
	mine, _ := s.ObjectsByType(a, "company")
	if len(mine) != 1 {
		t.Fatalf("one company record: %d", len(mine))
	}
	c := mine[0].State
	if c["name"] != "Biuro Rachunkowe Plus sp. z o.o." || c["kind"] != "self" || c["vat_id"] != "5260001246" || c["country"] != "PL" ||
		c["address"] != "ul. Prosta 1, 00-850 Warszawa" || c["iban"] != "PL61109010140000071219812874" || c["legal_form"] != "legal_entity" || mine[0].Company != "PL5260001246" {
		t.Fatalf("the record as entered: %+v (%s)", c, mine[0].Company)
	}
	ev, err := s.GetEvent(a, tr.EventID)
	var said map[string]any
	if err == nil {
		err = json.Unmarshal(ev.Payload, &said)
	}
	if err != nil || said["regon"] != "146234567" || said["krs"] != "0000123456" || said["source_ref"] != "GxrAG-993bbbf" || ev.Type != "company.entered" {
		t.Fatalf("the register's answer stays in the fact as its source: %s %s %v", ev.Type, ev.Payload, err)
	}
	if _, err := enter(a, map[string]any{"name": "Biuro Rachunkowe Plus", "kind": "self", "vat_id": "5260001246"}); !store.IsDuplicate(err) {
		t.Fatalf("the same company of the same kind entered twice is the same company: %v", err)
	}
	if tr, err := enter(a, map[string]any{"name": "Kawiarnia Bez Nazwy", "kind": "customer", "vat_id": "7010001234"}); err != nil || tr.Booked != 1 {
		t.Fatalf("a customer with no legal form: %+v %v", tr, err)
	}
	if tr, err := enter(b, map[string]any{"name": "Kawiarnia Bez Nazwy", "kind": "customer", "vat_id": "7010001234"}); err != nil || tr.Booked != 1 {
		t.Fatalf("another company keeps the same customer in its own books: %+v %v", tr, err)
	}
	all, _ := s.ObjectsByType(a, "company")
	if len(all) != 2 {
		t.Fatalf("the office's books: its own record and its customer, nothing doubled: %d", len(all))
	}
	for _, o := range all {
		if o.State["kind"] == "customer" && o.State["legal_form"] != nil && o.State["legal_form"] != "" {
			t.Fatalf("a legal form no one gave: %+v", o.State)
		}
	}
}
