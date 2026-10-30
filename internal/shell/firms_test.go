package shell_test

import "testing"

// TestOneFirmPerNIP (KK 2026-10-30: "we need to be able to add companies by
// NIP"): a firm is one record — a firm entered by a NIP the books already
// hold, in another role or by two people at once, is that firm, never a
// second one. What the registers say of it now (its address, account, legal
// form) updates it, what they leave out stays as it was, and its invoices
// still find it.
func TestOneFirmPerNIP(t *testing.T) {
	w := newCITWorld(t)
	ctx, s := w.ctx, w.s
	firms := func(nip string) []map[string]any {
		t.Helper()
		all, err := s.ObjectsByType(ctx, "company")
		if err != nil {
			t.Fatal(err)
		}
		var out []map[string]any
		for _, o := range all {
			if o.State["vat_id"] == nip {
				out = append(out, o.State)
			}
		}
		return out
	}
	// Nabywca S.A. is in the books as a customer; it is entered again by NIP,
	// as a supplier, with the address the register gives now.
	w.door("enter_company", map[string]any{"name": "NABYWCA SPÓŁKA AKCYJNA", "kind": "supplier", "vat_id": "5260001246",
		"address": "ul. Nowa 1, 00-001 Warszawa", "legal_form": "legal_entity", "regon": "123456789", "source": "gus", "occurred_at": "2026-10-01"})
	if got := firms("5260001246"); len(got) != 1 || got[0]["address"] != "ul. Nowa 1, 00-001 Warszawa" || got[0]["legal_form"] != "legal_entity" || got[0]["name"] != "Nabywca S.A." {
		t.Fatalf("one firm per NIP, updated by what the register says now: %v", got)
	}
	// entered once more with no address: the address stays
	w.door("enter_company", map[string]any{"name": "NABYWCA SPÓŁKA AKCYJNA", "kind": "self", "vat_id": "5260001246", "occurred_at": "2026-10-02"})
	if got := firms("5260001246"); len(got) != 1 || got[0]["address"] != "ul. Nowa 1, 00-001 Warszawa" {
		t.Fatalf("what an entry leaves out stays as it was: %v", got)
	}
	// its invoices still find it
	w.buy("fz-1", "FZ 1/10/2026", "2026-10-05", "5260001246", "100.00", "23.00", "123.00")
	// a NIP the books do not hold is a new firm
	w.door("enter_company", map[string]any{"name": "Nowa Firma sp. z o.o.", "kind": "customer", "vat_id": "5250009999", "occurred_at": "2026-10-03"})
	if got := firms("5250009999"); len(got) != 1 || got[0]["name"] != "Nowa Firma sp. z o.o." {
		t.Fatalf("a new NIP is a new firm: %v", got)
	}
}
