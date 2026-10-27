package pl_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/rights"
	"rhea/internal/store"
	"rhea/internal/store/storetest"
)

// A company's bank account is approved before it is the company's (KK,
// 2026-10-27: "such important metadata like bank accounts should be
// approved by someone responsible for bank accounts"; Poland pack v54): the
// first statement of an account nobody registered proposes it; a person may
// propose one, its IBAN checked; Bankowość approves — while nobody was put
// in it, the company's manager; once someone is, they alone, and never the
// one who proposed it while another could. One account per number. Closed
// with its day. Replay gives the same accounts.
func TestABankAccountIsApprovedByBankowosc(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s, Today: func() string { return "2026-10-27" }}
	installPoland(t, s, x)
	const co, first, second = "PL8058083055", "PL61109010140000071219812874", "PL27114020040000300201355387"
	at := store.WithCompany(ctx, co)
	door := func(ctx context.Context, name string, inputs map[string]any) (exec.Triggered, error) {
		return x.TriggerActivity(ctx, name, inputs, "operator", "")
	}
	must := func(ctx context.Context, name string, inputs map[string]any) exec.Triggered {
		t.Helper()
		tr, err := door(ctx, name, inputs)
		if err != nil || tr.Booked != 1 || len(tr.Errors) != 0 {
			t.Fatalf("%s: %+v %v", name, tr, err)
		}
		return tr
	}
	id := func(handle string) string {
		ids, _ := s.FindObjectIDsByField(ctx, "person", "handle", handle)
		return ids[0]
	}
	for _, h := range []string{"kk", "anna", "ewa"} {
		must(ctx, "register_person", map[string]any{"name": h, "handle": h, "role": "user"})
	}
	must(at, "grant_membership", map[string]any{"person": id("kk"), "company": co, "role": "company_manager", "from": "2026-01-01", "granted_by": "operator"})
	for _, h := range []string{"anna", "ewa"} {
		must(at, "grant_membership", map[string]any{"person": id(h), "company": co, "role": "user", "from": "2026-01-01", "granted_by": "kk"})
		must(at, "join_group", map[string]any{"group": "accountant", "person": id(h), "company": co, "from": "2026-01-01", "granted_by": "kk"})
	}
	as := func(handle string) context.Context {
		p, err := rights.Resolve(ctx, s, handle, "2026-10-27")
		if err != nil {
			t.Fatal(err)
		}
		return rights.WithPrincipal(at, p)
	}
	refused := func(err error, why string) {
		t.Helper()
		if err == nil || !strings.Contains(err.Error(), why) {
			t.Fatalf("want refused for %q, got %v", why, err)
		}
	}
	account := func(number string) core.Object {
		t.Helper()
		ids, err := s.FindObjectIDsByField(at, "bank_account", "number", number)
		if err != nil || len(ids) != 1 {
			t.Fatalf("the account %s: %v %v", number, ids, err)
		}
		o, _ := s.GetObject(at, ids[0])
		return o
	}
	said := func(number, date, key string) {
		t.Helper()
		raw, _ := json.Marshal(map[string]any{"account": number, "currency": "PLN", "date": date, "statement": "1/2026", "document": "sha256:x", "market": "pl"})
		if _, err := s.AppendEvent(at, core.Event{Kind: core.KindRaw, Type: "bank.account.seen", OccurredAt: date, Payload: raw, DedupKey: key, Company: co}); err != nil {
			t.Fatal(err)
		}
		if _, errs := x.ProcessPending(at); len(errs) > 0 {
			t.Fatal(errs)
		}
	}

	// the first statement of an account nobody registered proposes it
	said(first, "2026-09-01", "bank_mt940/account/"+first)
	if a := account(first); a.State["status"] != "proposed" || a.State["source"] != "statement" || a.State["name"] != first || a.State["seen_on"] != "2026-09-01" {
		t.Fatalf("proposed by the statement: %v", a.State)
	}
	// nobody is in Bankowość yet: an accountant, who may press every door, is not the one to approve
	_, err := door(as("anna"), "approve_bank_account", map[string]any{"bank_account": account(first).ID, "approved_by": "x"})
	refused(err, "a person in banking makes this move, and you are not in it")
	// … its manager is, until he hands it over, and may name the account
	must(as("kk"), "approve_bank_account", map[string]any{"bank_account": account(first).ID, "name": "PKO — bieżący", "approved_by": "x"})
	if a := account(first); a.State["status"] != "active" || a.State["approved_by"] != "kk" || a.State["name"] != "PKO — bieżący" {
		t.Fatalf("approved by the manager: %v", a.State)
	}

	// a person proposes one: the IBAN checked, one account per number
	_, err = door(as("anna"), "propose_bank_account", map[string]any{"name": "mBank EUR", "number": "PL27 1140 2004 0000 3002 0135 5388", "currency": "EUR", "proposed_by": "x"})
	refused(err, "check digits")
	_, err = door(as("anna"), "propose_bank_account", map[string]any{"name": "PKO", "number": first, "currency": "PLN", "proposed_by": "x"})
	refused(err, "registered already")
	must(as("anna"), "propose_bank_account", map[string]any{"name": "mBank EUR", "number": "pl27 1140 2004 0000 3002 0135 5387", "currency": "EUR", "bank": "mBank", "proposed_by": "x"})
	if a := account(second); a.State["status"] != "proposed" || a.State["source"] != "person" || a.State["proposed_by"] != "anna" {
		t.Fatalf("proposed by anna: %v", a.State)
	}
	// a statement of an account already proposed only notes its first day on one
	said(second, "2026-10-01", "bank_mt940/account/"+second)
	if a := account(second); a.State["status"] != "proposed" || a.State["seen_on"] != "2026-10-01" {
		t.Fatalf("a proposed account on a statement: %v", a.State)
	}

	// kk hands Bankowość over to anna and ewa: anna proposed it, so ewa
	// approves; kk is no longer the one
	for _, h := range []string{"anna", "ewa"} {
		must(at, "join_group", map[string]any{"group": "banking", "person": id(h), "company": co, "from": "2026-10-01", "granted_by": "kk"})
	}
	_, err = door(as("kk"), "approve_bank_account", map[string]any{"bank_account": account(second).ID, "approved_by": "x"})
	refused(err, "you are not in it")
	_, err = door(as("anna"), "approve_bank_account", map[string]any{"bank_account": account(second).ID, "approved_by": "x"})
	refused(err, "you proposed it — another person in banking makes this move")
	must(as("ewa"), "approve_bank_account", map[string]any{"bank_account": account(second).ID, "approved_by": "x"})
	if a := account(second); a.State["status"] != "active" || a.State["approved_by"] != "ewa" {
		t.Fatalf("approved by ewa: %v", a.State)
	}
	// a person only in Bankowość, no accountant, still approves: the group gives its buttons
	must(ctx, "register_person", map[string]any{"name": "Olga", "handle": "olga", "role": "user"})
	must(at, "grant_membership", map[string]any{"person": id("olga"), "company": co, "role": "user", "from": "2026-01-01", "granted_by": "kk"})
	must(at, "join_group", map[string]any{"group": "banking", "person": id("olga"), "company": co, "from": "2026-10-01", "granted_by": "kk"})
	must(as("olga"), "close_bank_account", map[string]any{"bank_account": account(first).ID, "closed_on": "2026-10-31", "closed_by": "x"})
	if a := account(first); a.State["status"] != "closed" || a.State["closed_on"] != "2026-10-31" {
		t.Fatalf("closed: %v", a.State)
	}
	groups, err := rights.Groups(ctx, s, co, "2026-10-27")
	if err != nil {
		t.Fatal(err)
	}
	for _, g := range groups {
		if g.Name != "banking" {
			continue
		}
		var hs []string
		for _, m := range g.Members {
			hs = append(hs, m.Handle)
		}
		if fmt.Sprint(hs) != "[anna ewa olga]" || !g.Everyones {
			t.Fatalf("Bankowość, every company's group, handed over: %v %+v", hs, g)
		}
	}
	before, _ := s.AllObjects(ctx)
	if _, err := x.Replay(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := s.AllObjects(ctx)
	if fmt.Sprint(len(before)) != fmt.Sprint(len(after)) || account(first).State["status"] != "closed" || account(second).State["approved_by"] != "ewa" {
		t.Fatalf("replay: %d → %d objects", len(before), len(after))
	}
}

// A statement line of an account Bankowość has not approved waits, and the
// approval releases every line that waited (KK, 2026-10-27, answer a:
// "nothing gets booked to an account nobody has checked, and the wait is
// one approval").
func TestStatementLinesWaitForTheirAccount(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s, Today: func() string { return "2026-10-27" }}
	installPoland(t, s, x)
	registerPartiesAt(t, s, x, selfCo)
	at := store.WithCompany(ctx, selfCo)
	const own = "PL83101010230000261395100000"
	said, _ := json.Marshal(map[string]any{"account": own, "currency": "PLN", "date": "2026-09-01", "statement": "1/2026", "document": "sha256:x", "market": "pl"})
	if _, err := s.AppendEvent(at, core.Event{Kind: core.KindRaw, Type: "bank.account.seen", OccurredAt: "2026-09-01", Payload: said, DedupKey: "bank_mt940/account/" + own, Company: selfCo}); err != nil {
		t.Fatal(err)
	}
	for i, amount := range []string{"100.00", "250.00"} {
		appendAt(t, s, x, selfCo, "bank.statement.line", "2026-09-20", fmt.Sprintf("line-%d", i), map[string]any{
			"date": "2026-09-20", "amount": amount, "currency": "PLN", "direction": "credit", "reference": fmt.Sprintf("pozycja %d", i),
			"counterparty": "Ktoś", "counterparty_account": "PL27114020040000300201355387", "account": own, "code": "NTRF", "market": "pl"})
	}
	ops := func() int {
		o, _ := s.ObjectsByType(at, "bank_operation")
		return len(o)
	}
	waiting, _ := s.UnmatchedRawEvents(at)
	if ops() != 0 || len(waiting) != 2 {
		t.Fatalf("before the approval: %d operations, %d facts waiting", ops(), len(waiting))
	}
	ownAccount(t, s, x, selfCo, own, "PLN") // the statement proposed it; its manager approves it
	if ops() != 2 {
		t.Fatalf("the approval released %d of the 2 lines", ops())
	}
	if d := debits(t, s); d["130"] != 35000 {
		t.Fatalf("booked to the bank: %v", d)
	}
}
