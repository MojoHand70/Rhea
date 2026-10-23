package shell_test

import (
	"testing"

	"rhea/internal/core"
	"rhea/internal/store"
)

// A change of Rhea's rules is announced to every company (KK, 2026-10-22: a
// global rule reaches every company at once, announced to each): staff's
// approval of Rhea's rule is the installation's fact, which every company
// reads, and each company lists what Rhea changed since a day. A company's
// approval of its own rule stays its own, and no other company hears of it.
func TestRheasChangesAreAnnouncedToEveryCompany(t *testing.T) {
	w := newGrantsWorld(t)
	a, b := alfaKey, betaKey
	kk := w.person("Krzysztof Kowal", "kk", "", "owner")
	w.mustDoor("", core.House, "grant_membership", map[string]any{"person": kk, "company": core.House, "role": "owner", "from": "2026-01-01", "granted_by": "operator"})
	w.person("Alfa Owner", "alfa", a, "owner")
	w.person("Beta Owner", "beta", b, "owner")
	draft := func(id, company string) {
		t.Helper()
		if _, err := w.s.InsertRuleVersion(w.ctx, core.Rule{ID: id, Company: company, Status: core.StatusDraft, Priority: 100,
			EffectiveFrom: w.today, CreatedBy: "agent:test", Description: "memos as notes: " + id, Spec: core.RuleSpec{Match: core.Match{EventType: "memo.taken"},
				Effect: core.Effect{Object: core.ObjectTemplate{Type: "note", Fields: map[string]string{"text": "=$.text", "amount": "=$.amount"}}}}}); err != nil {
			t.Fatal(err)
		}
	}
	draft("memo", "")
	draft("memo@"+a, a)
	if code := w.call("kk", core.House, "POST", "/api/rules/memo/approve", map[string]any{}, nil); code != 200 {
		t.Fatalf("staff approving Rhea's rule: %d", code)
	}
	if code := w.call("alfa", a, "POST", "/api/rules/memo@"+a+"/approve", map[string]any{}, nil); code != 200 {
		t.Fatalf("alfa approving its own: %d", code)
	}

	approvals := func(scope string) map[string]string {
		t.Helper()
		evs, err := w.s.EventsOfTypes(store.WithCompany(w.ctx, scope), []string{"rule.approved"})
		if err != nil {
			t.Fatal(err)
		}
		out := map[string]string{}
		for _, ev := range evs {
			out[ev.Company] = ev.Actor
		}
		return out
	}
	if got := approvals(b); len(got) != 1 || got[""] != "kk" {
		t.Fatalf("beta hears of Rhea's change and of nothing of alfa's: %v", got)
	}
	if got := approvals(a); len(got) != 2 || got[a] != "alfa" || got[""] != "kk" {
		t.Fatalf("alfa's own approval and Rhea's change: %v", got)
	}

	// Beta reads what Rhea changed today — the people pack loaded this
	// morning, and the memo rule — and nothing that is alfa's own.
	var changes []core.Rule
	if code := w.call("beta", b, "GET", "/api/changes?since="+w.today, nil, &changes); code != 200 {
		t.Fatalf("what Rhea changed, told to beta: %d", code)
	}
	memo := false
	for _, c := range changes {
		if c.Company != "" || c.Status != core.StatusActive {
			t.Fatalf("not one of Rhea's changes: %+v", c)
		}
		memo = memo || c.ID == "memo"
	}
	if !memo {
		t.Fatalf("Rhea's memo rule is not among today's changes: %+v", changes)
	}
	if code := w.call("beta", b, "GET", "/api/changes?since="+w.dayFrom(1), nil, &changes); code != 200 || len(changes) != 0 {
		t.Fatalf("nothing changed since tomorrow: %d %+v", code, changes)
	}
}
