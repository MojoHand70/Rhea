package adapter_test

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"rhea/internal/adapter"
	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/sim"
	"rhea/internal/store/storetest"
)

// The bank reader (DIRECTION 2026-10-10, step 5): every entry of a stored
// statement becomes a bank.statement.line saying what the generator's fact
// says — date, amount, direction, title, counterparty — naming the file;
// the statement moves to read; a statement whose balances do not add up is
// reported, not read.
func TestMT940ReadsTheStore(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	installIntake(t, s, x)
	sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", "pl", "bank-mt940.json"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-01-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve: %v %v", err, errs)
	}

	p, _, _ := sim.Load(filepath.Join("..", "..", "testdata", "customers", "nordwind.json"))
	p.Months = 2
	f, err := sim.Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	var statements []sim.Document
	for _, d := range f.Documents {
		if d.Source == "bank" {
			statements = append(statements, d)
			receive(t, x, filepath.Base(d.Path), d.MediaType, d.Source, d.Date, d.Bytes)
		}
	}
	lines := 0
	for _, d := range statements {
		lines += len(d.Events)
	}
	res, err := adapter.Run(ctx, s, x, adapter.MT940{})
	if err != nil || res.PassErr != nil {
		t.Fatalf("pass: %v %v", err, res.PassErr)
	}
	if res.Channel != "bank_mt940" || res.Submitted != lines+len(statements) || res.Booked != len(statements) {
		t.Fatalf("pass = %+v (lines %d, statements %d)", res, lines, len(statements))
	}
	raws, _ := s.EventsByKind(ctx, core.KindRaw)
	byKey := map[string]map[string]any{}
	for _, ev := range raws {
		if ev.Type != "bank.statement.line" {
			continue
		}
		var m map[string]any
		jsonUnmarshal(ev.Payload, &m)
		byKey[m["reference"].(string)+"|"+m["date"].(string)] = m
		if ev.ActivityName != "bank_mt940" || ev.OccurredAt != m["date"] {
			t.Fatalf("stamp: %+v", ev)
		}
	}
	debits := 0
	for _, d := range statements {
		hash := core.HashBytes(d.Bytes)
		for _, i := range d.Events {
			fact := f.Events[i]
			m, ok := byKey[fact.Payload["reference"].(string)+"|"+fact.OccurredAt]
			if !ok {
				t.Fatalf("line %v on %s not read", fact.Payload["reference"], fact.OccurredAt)
			}
			if m["amount"] != fact.Payload["amount"] || m["direction"] != fact.Payload["direction"] ||
				m["counterparty"] != fact.Payload["counterparty"] || m["counterparty_account"] != fact.Payload["counterparty_account"] ||
				m["account"] != fact.Payload["account"] || m["currency"] != "PLN" || m["document"] != hash ||
				!strings.Contains(m["raw86"].(string), "|"+fact.Payload["reference"].(string)+"|") {
				t.Fatalf("read %v, generated %v", m, fact.Payload)
			}
			if m["direction"] == "debit" {
				debits++
			}
		}
		ids, _ := s.FindObjectIDsByField(ctx, "received_document", "hash", hash)
		if o, err := s.GetObject(ctx, ids[0]); err != nil || o.State["status"] != "read" {
			t.Fatalf("statement status = %v %v", o.State["status"], err)
		}
	}
	if debits == 0 {
		t.Fatal("no purchase payment in two months — the generator pays nobody")
	}
	if wl, _ := s.UnmatchedRawEvents(ctx); len(wl) != lines {
		t.Fatalf("worklist = %d, want the %d lines", len(wl), lines)
	}
	if res, err := adapter.Run(ctx, s, x, adapter.MT940{}); err != nil || res.Submitted != 0 {
		t.Fatalf("second pass = %+v, %v", res, err)
	}

	// A statement whose closing balance does not follow is reported.
	lying := bytes.Replace(statements[0].Bytes, []byte(":62F:C"), []byte(":62F:D"), 1)
	receive(t, x, "lying.sta", "text/plain", "bank", "2026-03-01", lying)
	res, err = adapter.Run(ctx, s, x, adapter.MT940{Today: func() string { return "2026-03-02" }})
	if err != nil || res.Submitted != 1 {
		t.Fatalf("lying pass = %+v, %v", res, err)
	}
	ev, ok, _ := s.LatestEventOfType(ctx, adapter.EventDocumentUnreadable)
	if !ok || !bytes.Contains(ev.Payload, []byte("contradicts itself")) {
		t.Fatalf("unreadable = %+v", ev)
	}
}

func TestParseMT940(t *testing.T) {
	text := ":20:STMT202601\r\n:25:PL61109010140000071219812874\r\n:28C:1/2026\r\n:60F:C260101PLN10000,00\r\n" +
		":61:2601190119C362,85NTRFNONREF//1\r\n:86:PRZELEW PRZYCHODZACY|FS 6/2026|Mazowiecka Hurtownia S.A.|PL10105000997603123456789123\r\n" +
		":61:2601200120D1500,00NTRFNONREF//2\r\n:86:some other bank's free text about FAB/1/2026\r\n" +
		":62F:C260131PLN8862,85\r\n-\r\n"
	if !adapter.IsMT940([]byte(text)) || adapter.IsMT940([]byte("<Faktura/>")) {
		t.Fatal("IsMT940")
	}
	st, err := adapter.ParseMT940([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	if st.Account != "PL61109010140000071219812874" || st.Number != "1/2026" || st.Currency != "PLN" ||
		st.Opening != 1_000_000 || st.Closing != 886_285 || st.From != "2026-01-01" || st.To != "2026-01-31" || len(st.Lines) != 2 {
		t.Fatalf("statement = %+v", st)
	}
	a, b := st.Lines[0], st.Lines[1]
	if a.Direction != "credit" || a.Amount != 36285 || a.Date != "2026-01-19" || a.Reference != "FS 6/2026" ||
		a.Counterparty != "Mazowiecka Hurtownia S.A." || a.CounterpartyAccount != "PL10105000997603123456789123" {
		t.Fatalf("line 1 = %+v", a)
	}
	// Another bank's dialect: the whole line is the title, nothing is guessed.
	if b.Direction != "debit" || b.Amount != 150000 || b.Reference != "some other bank's free text about FAB/1/2026" || b.Counterparty != "" {
		t.Fatalf("line 2 = %+v", b)
	}
	if _, err := adapter.ParseMT940([]byte(strings.Replace(text, "8862,85", "8862,86", 1))); err == nil || !strings.Contains(err.Error(), "contradicts") {
		t.Fatalf("a wrong closing balance parsed: %v", err)
	}
}
