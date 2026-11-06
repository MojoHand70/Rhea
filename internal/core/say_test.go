package core

import (
	"encoding/json"
	"strings"
	"testing"
)

// The pack says, in its language, what a rule's step did (KK 2026-11-05,
// the stream: "each event should be an event"; the rule's words must not
// "sound like a Robocop"): a sentence template per rule in the words, its
// placeholders filled from the step's facts, rendered as parts a face sets
// — text, a figure in the monospace, money with its currency, a date. The
// face knows no business and no Polish; the pack knows both.
func TestWordsSayAStep(t *testing.T) {
	var w WordsSpec
	if err := json.Unmarshal([]byte(`{"lang": "pl",
		"values": {"bank_operation.kind": {"transfer": "Zapłata kontrahenta", "fee": "opłata bankowa"}},
		"steps": {
			"pl-match-incoming": {"says": "dopasowana do {other|f} po {strategy|value:matching.strategy}", "area": "Rozrachunki"},
			"pl-post-incoming-payment": {"says": "zaksięgowana: {n|f} {n|plural:zapis,zapisy,zapisów}", "area": "Księgi"},
			"pl-bank-operation": {"says": "rozpoznana jako {kind|value:bank_operation.kind}", "area": "Bank"},
			"pl-settle-vat": {"says": "VAT za {period|month} rozliczony: {vat_due|m} do zapłaty", "area": "VAT"},
			"pl-record-incoming-settlement": {"says": ""},
			"pl-ksef-rejection-task": {"says": "odrzucona przez KSeF: {reason}", "area": "KSeF", "state": "wait"},
			"event:bank.statement.line": {"says": "{direction|value:bank_operation.direction} {amount|m} · {counterparty}, tytułem {reference|f}"},
			"type:posting": {"says": "zapis w księgach"}}}`), &w); err != nil {
		t.Fatal(err)
	}
	if err := w.Check(); err != nil {
		t.Fatalf("words with steps: %v", err)
	}
	if w.Steps["pl-ksef-rejection-task"].State != "wait" || w.Steps["pl-record-incoming-settlement"].Says != "" {
		t.Fatalf("steps as read: %+v", w.Steps)
	}
	look := func(m map[string]any) func(string) (any, bool) {
		return func(k string) (any, bool) { v, ok := m[k]; return v, ok }
	}
	flat := func(parts []SayPart) string {
		var b strings.Builder
		for _, p := range parts {
			switch {
			case p.T != "":
				b.WriteString(p.T)
			case p.F != "":
				b.WriteString("[" + p.F + "]")
			case p.M != "":
				b.WriteString("<" + p.M + " " + p.C + ">")
			case p.D != "":
				b.WriteString("(" + p.D + ":" + p.As + ")")
			}
		}
		return b.String()
	}
	cases := []struct {
		tpl  string
		in   map[string]any
		want string
	}{
		{w.Steps["pl-match-incoming"].Says, map[string]any{"other": "FS 6/2026", "strategy": "reference"}, "dopasowana do [FS 6/2026] po reference"},
		{w.Steps["pl-post-incoming-payment"].Says, map[string]any{"n": 2}, "zaksięgowana: [2] zapisy"},
		{w.Steps["pl-post-incoming-payment"].Says, map[string]any{"n": 1}, "zaksięgowana: [1] zapis"},
		{w.Steps["pl-post-incoming-payment"].Says, map[string]any{"n": 12}, "zaksięgowana: [12] zapisów"},
		{w.Steps["pl-post-incoming-payment"].Says, map[string]any{"n": 23}, "zaksięgowana: [23] zapisy"},
		{w.Steps["pl-bank-operation"].Says, map[string]any{"kind": "fee"}, "rozpoznana jako opłata bankowa"},
		{w.Steps["pl-bank-operation"].Says, map[string]any{"kind": "transfer"}, "rozpoznana jako Zapłata kontrahenta"},
		// a value's word lowered for the middle of a sentence; an abbreviation keeps its capitals
		{"to {kind|lvalue:bank_operation.kind} i {x|lower}", map[string]any{"kind": "transfer", "x": "ZUS"}, "to zapłata kontrahenta i ZUS"},
		{w.Steps["pl-settle-vat"].Says, map[string]any{"period": "2026-09", "vat_due": float64(509000)}, "VAT za (2026-09:month) rozliczony: <509000 > do zapłaty"},
		// a raw payload carries money as a decimal string: it is parsed, never guessed; the currency rides along
		{w.Steps["event:bank.statement.line"].Says, map[string]any{"direction": "credit", "amount": "795.81", "currency": "PLN", "counterparty": "Hostingowo", "reference": "HOS/1/2026"}, "credit <79581 PLN> · Hostingowo, tytułem [HOS/1/2026]"},
		// a name the facts do not have renders as nothing, never as its own brackets
		{"{nothing} tu", map[string]any{}, " tu"},
	}
	for _, c := range cases {
		if got := flat(w.Say(c.tpl, look(c.in))); got != c.want {
			t.Errorf("say %q over %v:\n got %q\nwant %q", c.tpl, c.in, got, c.want)
		}
	}
	// English counts one and many; the plural's third form is for languages that have it
	en := WordsSpec{Lang: "en"}
	if got := flat(en.Say("{n|f} {n|plural:entry,entries}", look(map[string]any{"n": 3}))); got != "[3] entries" {
		t.Fatalf("english plural: %q", got)
	}
	// a later pack's step words lay over the earlier's by key, like every other word
	w.Merge(WordsSpec{Lang: "pl", Steps: map[string]StepWords{"pl-bank-operation": {Says: "to {kind}", Area: "Bank"}, "pl-new": {Says: "nowy krok"}}})
	if w.Steps["pl-bank-operation"].Says != "to {kind}" || w.Steps["pl-new"].Says != "nowy krok" || w.Steps["pl-settle-vat"].Area != "VAT" {
		t.Fatalf("merged steps: %+v", w.Steps)
	}
	for what, bad := range map[string]string{
		"an unclosed placeholder":  `{"pl-x": {"says": "zapis {n"}}`,
		"an unknown format":        `{"pl-x": {"says": "{amount|dollars}"}}`,
		"a plural with one form":   `{"pl-x": {"says": "{n|plural:zapis}"}}`,
		"a value with no field":    `{"pl-x": {"says": "{kind|value}"}}`,
		"a state that is no state": `{"pl-x": {"says": "x", "state": "maybe"}}`,
		"a silent event title":     `{"event:bank.statement.line": {"says": ""}}`,
		"a silent type fallback":   `{"type:posting": {"says": " "}}`,
		"a key naming nothing":     `{"": {"says": "x"}}`,
	} {
		var ws WordsSpec
		if err := json.Unmarshal([]byte(`{"lang": "pl", "steps": `+bad+`}`), &ws); err != nil {
			t.Fatalf("%s: %v", what, err)
		}
		if err := ws.Check(); err == nil {
			t.Errorf("%s: not refused", what)
		}
	}
}
