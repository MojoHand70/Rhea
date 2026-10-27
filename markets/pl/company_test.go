package pl_test

import (
	"context"
	"encoding/json"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/rights"
	"rhea/internal/store"
)

// selfCo is the company the Poland pack's tests keep the books of — its own
// NIP, as registerParties names it.
const selfCo = "PL8058083055"

// appendAt appends a fact of a company and books what it starts, as a
// channel's fact of that company is booked.
func appendAt(t *testing.T, s *store.Store, x *exec.Executor, company, typ, date, dedup string, payload map[string]any) {
	t.Helper()
	ctx := store.WithCompany(context.Background(), company)
	raw, _ := json.Marshal(payload)
	if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: typ, OccurredAt: date, Payload: raw, DedupKey: dedup, Company: company}); err != nil {
		t.Fatal(err)
	}
	if _, errs := x.ProcessPending(ctx); len(errs) > 0 {
		t.Fatalf("%s: %v", typ, errs)
	}
}

// registerPartiesAt is registerParties in a company: itself and a customer.
func registerPartiesAt(t *testing.T, s *store.Store, x *exec.Executor, company string) string {
	t.Helper()
	appendAt(t, s, x, company, "company.registered", "2026-09-01", "self", map[string]any{"name": "Rhea Test sp. z o.o.", "kind": "self",
		"vat_id": "PL8058083055", "country": "PL", "iban": "PL83101010230000261395100000", "address": "ul. Testowa 1, 00-001 Warszawa"})
	appendAt(t, s, x, company, "company.registered", "2026-09-01", "buyer", map[string]any{"name": "Nabywca S.A.", "kind": "customer",
		"vat_id": "PL5265877635", "country": "PL", "iban": "PL60102010260000042270201111", "address": "ul. Próbna 2, 00-002 Warszawa"})
	companies, _ := s.ObjectsByType(store.WithCompany(context.Background(), company), "company")
	for _, c := range companies {
		if c.State["kind"] == "self" {
			return c.ID
		}
	}
	t.Fatal("no company of kind self")
	return ""
}

// ownAccount makes a company's bank account its own as its manager does it:
// proposed, then approved — Bankowość is nobody's yet, so the manager is
// the one. Statement lines of an account not approved wait (Poland pack
// v54), so a story that books bank lines first approves its account.
func ownAccount(t *testing.T, s *store.Store, x *exec.Executor, company, number, currency string) {
	t.Helper()
	ctx := context.Background()
	at := store.WithCompany(ctx, company)
	if ids, _ := s.FindObjectIDsByField(ctx, "person", "handle", "kk"); len(ids) == 0 {
		if _, err := x.TriggerActivity(ctx, "register_person", map[string]any{"name": "Krzysztof Kowal", "handle": "kk", "role": "user"}, "operator", ""); err != nil {
			t.Fatal(err)
		}
	}
	ids, _ := s.FindObjectIDsByField(ctx, "person", "handle", "kk")
	if _, err := x.TriggerActivity(at, "grant_membership", map[string]any{"person": ids[0], "company": company, "role": "company_manager", "from": "2026-01-01", "granted_by": "operator"}, "operator", ""); err != nil {
		t.Fatal(err)
	}
	day := core.Today()
	if x.Today != nil {
		day = x.Today()
	}
	p, err := rights.Resolve(ctx, s, "kk", day)
	if err != nil {
		t.Fatal(err)
	}
	as := rights.WithPrincipal(at, p)
	acc, _ := s.FindObjectIDsByField(at, "bank_account", "number", number)
	if len(acc) == 0 { // a statement may have proposed it already
		tr, err := x.TriggerActivity(as, "propose_bank_account", map[string]any{"name": "Rachunek bieżący", "number": number, "currency": currency, "proposed_by": "kk"}, "kk", "")
		if err != nil || tr.Booked < 1 {
			t.Fatalf("propose the account: %+v %v", tr, err)
		}
		acc, _ = s.FindObjectIDsByField(at, "bank_account", "number", number)
	}
	if tr, err := x.TriggerActivity(as, "approve_bank_account", map[string]any{"bank_account": acc[0], "approved_by": "kk"}, "kk", ""); err != nil || tr.Booked < 1 {
		t.Fatalf("approve the account: %+v %v", tr, err)
	}
}
