package adapter_test

import (
	"bytes"
	"context"
	"encoding/json"
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

	p, _, err := sim.Load(filepath.Join("..", "..", "testdata", "customers", "nordwind.json"))
	if err != nil {
		t.Fatal(err)
	}
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
	// every line, a read a file, and the company's one account said once
	// across both months' statements (code version 2: the pack proposes
	// the account from it; the count changed with the reader)
	if res.Channel != "bank_mt940" || res.Submitted != lines+len(statements)+1 || res.Booked != len(statements) {
		t.Fatalf("pass = %+v (lines %d, statements %d)", res, lines, len(statements))
	}
	raws, _ := s.EventsByKind(ctx, core.KindRaw)
	var said []map[string]any
	for _, ev := range raws {
		if ev.Type == adapter.EventBankAccountSeen {
			var m map[string]any
			jsonUnmarshal(ev.Payload, &m)
			said = append(said, m)
			if ev.DedupKey != "bank_mt940/account/"+m["account"].(string) || ev.OccurredAt != m["date"] {
				t.Fatalf("the account said: %+v", ev)
			}
		}
	}
	if len(said) != 1 || said[0]["account"] != f.Events[statements[0].Events[0]].Payload["account"] || said[0]["currency"] != "PLN" {
		t.Fatalf("the account, said once: %v", said)
	}
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
	if wl, _ := s.UnmatchedRawEvents(ctx); len(wl) != lines+1 {
		t.Fatalf("worklist = %d, want the %d lines and the account said, no pack here to read them", len(wl), lines)
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

	// A file through the bank door that is no statement is reported, not
	// skipped; a statement through another door is not this channel's.
	receive(t, x, "export.csv", "text/csv", "bank", "2026-03-01", []byte("date,amount\n2026-01-02,10.00\n"))
	forwarded := bytes.Replace(statements[1].Bytes, []byte(":20:STMT"), []byte(":20:FWD"), 1)
	receive(t, x, "forwarded.sta", "text/plain", "email", "2026-03-01", forwarded)
	res, err = adapter.Run(ctx, s, x, adapter.MT940{Today: func() string { return "2026-03-02" }})
	if err != nil || res.Submitted != 1 {
		t.Fatalf("csv pass = %+v, %v", res, err)
	}
	ev, ok, _ = s.LatestEventOfType(ctx, adapter.EventDocumentUnreadable)
	if !ok || !bytes.Contains(ev.Payload, []byte("not an MT940 statement")) || !bytes.Contains(ev.Payload, []byte("export.csv")) {
		t.Fatalf("unreadable = %+v", ev)
	}

	// Two months in one file: the lines number through the file, the read
	// is dated by the last closing balance. (New statement numbers: a line
	// is keyed on its account, statement and number, so the statements
	// already read would be duplicates, which is the next check.)
	two := bytes.ReplaceAll(append(append([]byte{}, statements[0].Bytes...), statements[1].Bytes...), []byte(":28C:"), []byte(":28C:Q-"))
	receive(t, x, "q1.sta", "text/plain", "bank", "2026-03-01", two)
	res, err = adapter.Run(ctx, s, x, adapter.MT940{})
	if err != nil || res.Submitted != lines+1 || res.Booked != 1 {
		t.Fatalf("two-statement pass = %+v, %v (lines %d)", res, err, lines)
	}
	read, ok, _ := s.LatestEventOfType(ctx, adapter.EventDocumentRead)
	if !ok || read.OccurredAt != statements[1].Date || !bytes.Contains(read.Payload, []byte(core.HashBytes(two))) {
		t.Fatalf("read = %+v, want dated %s", read, statements[1].Date)
	}
	last, _, _ := s.LatestEventOfType(ctx, "bank.statement.line")
	var m map[string]any
	jsonUnmarshal(last.Payload, &m)
	if m["line"] != float64(lines) || m["statement"] != "Q-2/2026" {
		t.Fatalf("last line = %v, want line %d of statement Q-2/2026", m, lines)
	}

	// A statement downloaded again in other bytes is the same statement:
	// its reading is new, its lines are duplicates, nothing books twice.
	linesBefore, _ := s.EventsOfTypes(ctx, []string{"bank.statement.line"})
	receive(t, x, "again.sta", "text/plain", "bank", "2026-03-03", bytes.ReplaceAll(statements[0].Bytes, []byte("\r\n"), []byte("\n")))
	if res, err := adapter.Run(ctx, s, x, adapter.MT940{}); err != nil || res.Submitted != 1 || res.Duplicates < 1 {
		t.Fatalf("the same statement again = %+v, %v", res, err)
	}
	if linesAfter, _ := s.EventsOfTypes(ctx, []string{"bank.statement.line"}); len(linesAfter) != len(linesBefore) {
		t.Fatalf("the same statement in other bytes booked its lines again: %d → %d", len(linesBefore), len(linesAfter))
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
	sts, err := adapter.ParseMT940([]byte(text))
	if err != nil || len(sts) != 1 {
		t.Fatal(err, len(sts))
	}
	st := sts[0]
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
	// A continuation line of free text may start with a colon: it is not a
	// tag, and the title keeps it.
	colon := strings.Replace(text, "about FAB/1/2026\r\n", "about FAB/1/2026\r\n:tel. 123\r\n", 1)
	if sts, err := adapter.ParseMT940([]byte(colon)); err != nil || !strings.Contains(sts[0].Lines[1].Reference, ":tel. 123") {
		t.Fatalf("a colon in free text: %v %+v", err, sts)
	}
	// A statement without its number (:28C:) is refused: a line's identity
	// is the account, the statement and the line within it.
	if _, err := adapter.ParseMT940([]byte(strings.Replace(text, ":28C:1/2026\r\n", "", 1))); err == nil || !strings.Contains(err.Error(), ":28C:") {
		t.Fatalf("a statement without a number parsed: %v", err)
	}
	// CP1250, which Polish banks commonly send, is refused with its reason
	// rather than read as mangled titles.
	cp1250 := strings.Replace(text, "Hurtownia", "Hurtownia \xb3", 1)
	if _, err := adapter.ParseMT940([]byte(cp1250)); err == nil || !strings.Contains(err.Error(), "not UTF-8") {
		t.Fatalf("CP1250 read as UTF-8: %v", err)
	}
}

// A real bank's export: a SWIFT header, several statements in one file,
// a statement over two balance pages, a one-decimal amount, a reversal,
// and a free-text :86: that happens to contain a bar.
func TestParseMT940Pages(t *testing.T) {
	text := "{1:F01BANKPLPWAXXX0000000000}{2:O9400000260301BANKPLPWAXXX00000000002603010000N}{4:\r\n:940:\r\n" +
		":20:STMT202601\r\n:25:PL61109010140000071219812874\r\n:28C:1/1\r\n:60F:C260101PLN10000,00\r\n" +
		":61:2601190119C362,85NTRFNONREF//1\r\n:86:PRZELEW PRZYCHODZACY|FS 6/2026|Mazowiecka Hurtownia S.A.|PL10105000997603123456789123\r\n" +
		":62M:C260115PLN10362,85\r\n-\r\n" +
		":20:STMT202601\r\n:25:PL61109010140000071219812874\r\n:28C:1/2\r\n:60M:C260115PLN10362,85\r\n" +
		":61:2601200120D1500,5NTRFNONREF//2\r\n:86:TRANSFER | ref FAB/1/2026\r\n" +
		":62F:C260131PLN8862,35\r\n-\r\n" +
		":20:STMT202602\r\n:25:PL61109010140000071219812874\r\n:28C:2/1\r\n:60F:C260201PLN8862,35\r\n" +
		":61:2602030203RD1500,5NTRFNONREF//3\r\n:86:PRZELEW PRZYCHODZACY|ZWROT FAB/1/2026|Fabryka Sp. z o.o.|PL10105000997603123456789123\r\n" +
		":62F:C260228PLN10362,85\r\n-}\r\n{5:{CHK:000000000000}}\r\n"
	if !adapter.IsMT940([]byte(text)) {
		t.Fatal("IsMT940 refuses a SWIFT header")
	}
	sts, err := adapter.ParseMT940([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	if len(sts) != 3 {
		t.Fatalf("%d statements, want 3 (one per :20:)", len(sts))
	}
	jan, page2, feb := sts[0], sts[1], sts[2]
	if jan.Number != "1/1" || jan.Opening != 1_000_000 || jan.Closing != 1_036_285 || jan.To != "2026-01-15" || len(jan.Lines) != 1 {
		t.Fatalf("page 1 = %+v", jan)
	}
	if page2.Number != "1/2" || page2.Opening != 1_036_285 || page2.Closing != 886_235 || len(page2.Lines) != 1 {
		t.Fatalf("page 2 = %+v", page2)
	}
	// "1500,5" is 1500.50; a bar in another bank's text does not make it the dialect.
	if l := page2.Lines[0]; l.Amount != 150050 || l.Direction != "debit" || l.Reversal ||
		l.Reference != "TRANSFER | ref FAB/1/2026" || l.Counterparty != "" {
		t.Fatalf("line = %+v", l)
	}
	// A reversed debit comes back as a credit and says it is a reversal.
	if l := feb.Lines[0]; l.Amount != 150050 || l.Direction != "credit" || !l.Reversal || l.Reference != "ZWROT FAB/1/2026" {
		t.Fatalf("reversal = %+v", l)
	}

	for _, bad := range []struct{ name, from, to, reason string }{
		{"a page opening elsewhere than the last closed", ":60M:C260115PLN10362,85", ":60M:C260115PLN10362,86", "opens at"},
		{"an entry after the closing balance", ":62F:C260228PLN10362,85\r\n", ":62F:C260228PLN10362,85\r\n:61:2602040204C1,00NTRFNONREF//4\r\n", "outside a balance page"},
		{"a statement never closed", ":62F:C260228PLN10362,85\r\n", "", "needs :25:, :60F: and :62F:"},
		{"a wrong page total", "8862,35", "8862,36", "contradicts itself"},
	} {
		_, err := adapter.ParseMT940([]byte(strings.Replace(text, bad.from, bad.to, 1)))
		if err == nil || !strings.Contains(err.Error(), bad.reason) {
			t.Errorf("%s: %v", bad.name, err)
		}
	}
	if _, err := adapter.ParseMT940([]byte(":25:PL61109010140000071219812874\r\n:20:X\r\n")); err == nil || !strings.Contains(err.Error(), "not an MT940 statement") {
		t.Errorf("a file starting elsewhere than :20: parsed: %v", err)
	}
}

// A bank line is keyed by its statement's year (review 2026-10-21, D1):
// Polish banks number statements within the year, and a line was keyed by
// account, statement number and position alone — January 2027's statement
// 38 had the keys of January 2026's, so its lines counted as already
// logged, the file was marked read, and the payments never reached the
// books. Now the key names the year of the statement's opening balance; a
// line logged before, under the key without a year, is still known when
// the same statement comes again in another file.
func TestMT940LinesOfTwoYearsAreTwoLines(t *testing.T) {
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
	account := "PL61109010140000071219812874"
	statement := func(yy, day, amount, closing string) []byte {
		return []byte(":20:STMT\r\n:25:" + account + "\r\n:28C:38/1\r\n:60F:C" + yy + "0102PLN1000,00\r\n" +
			":61:" + yy + "01" + day + "01" + day + "C" + amount + "NTRFNONREF//1\r\n:86:PRZELEW PRZYCHODZACY|FS 1/" + yy + "|Mazowiecka Hurtownia S.A.|PL10105000997603123456789123\r\n" +
			":62F:C" + yy + "0131PLN" + closing + "\r\n-\r\n")
	}
	// A line read before the key named the year: its statement 38/1 of 2026,
	// as the reader logged it then.
	old, _ := json.Marshal(map[string]any{"date": "2026-01-05", "amount": "362.85", "currency": "PLN", "direction": "credit",
		"reference": "FS 1/26", "account": account, "statement": "38/1", "line": 1})
	if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: "bank.statement.line", OccurredAt: "2026-01-05", Payload: old,
		DedupKey: "bank_mt940/" + account + "/38/1/1", Actor: "adapter:bank_mt940", ActivityName: "bank_mt940", ActivityVersion: 1}); err != nil {
		t.Fatal(err)
	}
	receive(t, x, "january-2026-again.sta", "text/plain", "bank", "2026-02-01", statement("26", "05", "362,85", "1362,85"))
	receive(t, x, "january-2027.sta", "text/plain", "bank", "2027-02-01", statement("27", "05", "99,00", "1099,00"))
	if res, err := adapter.Run(ctx, s, x, adapter.MT940{}); err != nil || res.PassErr != nil {
		t.Fatalf("pass: %+v %v", res, err)
	}
	lines, err := s.EventsOfTypes(ctx, []string{"bank.statement.line"})
	if err != nil {
		t.Fatal(err)
	}
	dates := map[string]int{}
	for _, l := range lines {
		dates[l.OccurredAt]++
	}
	if len(lines) != 2 || dates["2026-01-05"] != 1 || dates["2027-01-05"] != 1 {
		t.Fatalf("2026's line once and 2027's line once: %v", dates)
	}
}

// A bank line is keyed by its statement's opening day (review 2026-10-27,
// high): the key named the statement's year and number, and a bank that
// writes the same number on every daily statement ("1", "00001/001") gave
// the second day's lines the first day's keys — they were taken for lines
// already logged, the file was marked read, and the payments never reached
// the books. Now the key names the day the statement opens; a line logged
// before under the year's key, or under the key with neither, is still
// known when the same statement comes again.
func TestMT940ADailyStatementNumberIsEachDaysOwn(t *testing.T) {
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
	account := "PL61109010140000071219812874"
	statement := func(day, amount, opening, closing string) []byte {
		return []byte(":20:STMT\r\n:25:" + account + "\r\n:28C:1\r\n:60F:C2609" + day + "PLN" + opening + "\r\n" +
			":61:2609" + day + "09" + day + "C" + amount + "NTRFNONREF//1\r\n:86:PRZELEW PRZYCHODZACY|FS " + day + "/26|Mazowiecka Hurtownia S.A.|PL10105000997603123456789123\r\n" +
			":62F:C2609" + day + "PLN" + closing + "\r\n-\r\n")
	}
	// The 14th's line, read before the key named the day: under the year's.
	old, _ := json.Marshal(map[string]any{"date": "2026-09-14", "amount": "10.00", "currency": "PLN", "direction": "credit",
		"reference": "FS 14/26", "account": account, "statement": "1", "line": 1})
	if _, err := s.AppendEvent(ctx, core.Event{Kind: core.KindRaw, Type: "bank.statement.line", OccurredAt: "2026-09-14", Payload: old,
		DedupKey: "bank_mt940/" + account + "/2026/1/1", Actor: "adapter:bank_mt940", ActivityName: "bank_mt940", ActivityVersion: 1}); err != nil {
		t.Fatal(err)
	}
	receive(t, x, "14.sta", "text/plain", "bank", "2026-09-14", statement("14", "10,00", "1000,00", "1010,00"))
	receive(t, x, "15.sta", "text/plain", "bank", "2026-09-15", statement("15", "100,00", "1010,00", "1110,00"))
	receive(t, x, "16.sta", "text/plain", "bank", "2026-09-16", statement("16", "250,00", "1110,00", "1360,00"))
	if res, err := adapter.Run(ctx, s, x, adapter.MT940{}); err != nil || res.PassErr != nil {
		t.Fatalf("pass: %+v %v", res, err)
	}
	lines, err := s.EventsOfTypes(ctx, []string{"bank.statement.line"})
	if err != nil {
		t.Fatal(err)
	}
	dates := map[string]int{}
	for _, l := range lines {
		dates[l.OccurredAt]++
	}
	if len(lines) != 3 || dates["2026-09-14"] != 1 || dates["2026-09-15"] != 1 || dates["2026-09-16"] != 1 {
		t.Fatalf("each day's line once: %v", dates)
	}
}
