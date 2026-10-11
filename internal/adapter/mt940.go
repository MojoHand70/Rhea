package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"rhea/internal/core"
)

// MT940 is the reader of the bank channel (DIRECTION 2026-10-10, step 5):
// every received statement in the store becomes one bank.statement.line per
// entry — the value date, the amount and its direction, the title and the
// counterparty as the bank wrote them — each naming the file, and a
// document.read beside them. The reading is exact. A statement that
// contradicts itself (the opening balance and the entries do not give the
// closing balance) is reported as document.unreadable for a person.
//
// The :86: line is the one trap: every bank writes it differently. The
// reader keeps it raw on every line (raw86) and understands one dialect,
// "sim-bank" (kind|title|counterparty|account); a real bank's dialect is a
// second parser behind the same channel, versioned like the channel.
type MT940 struct {
	Today func() string
}

const mt940CodeVersion = 1

func init() {
	Register(Implemented{Name: "bank_mt940", CodeVersion: mt940CodeVersion,
		Format: "MT940 statements from the document store; :86: kept raw, the sim-bank dialect read"})
	readers["bank_mt940"] = func() Adapter { return MT940{} }
}

func (MT940) Name() string     { return "bank_mt940" }
func (MT940) CodeVersion() int { return mt940CodeVersion }

func (m MT940) Pass(ctx context.Context, reads Reads) ([]core.Event, error) {
	docs, err := reads.ObjectsByType(ctx, ReceivedDocumentType)
	if err != nil {
		return nil, err
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].ID < docs[j].ID })
	var out []core.Event
	for _, d := range docs {
		status, _ := d.State["status"].(string)
		media, _ := d.State["media_type"].(string)
		source, _ := d.State["source"].(string)
		hash, _ := d.State["hash"].(string)
		name, _ := d.State["name"].(string)
		if status != "received" || hash == "" {
			continue
		}
		if !strings.HasPrefix(media, "text/") && source != "bank" {
			continue
		}
		b, err := reads.GetBlob(ctx, hash)
		if err != nil {
			return out, fmt.Errorf("%s names file %s: %w", d.ID, hash, err)
		}
		if !IsMT940(b.Bytes) {
			continue
		}
		st, err := ParseMT940(b.Bytes)
		if err != nil {
			payload, _ := json.Marshal(map[string]any{
				"document": hash, "name": name, "channel": m.Name(), "reason": err.Error(),
			})
			out = append(out, core.Event{Type: EventDocumentUnreadable, OccurredAt: m.today(),
				Payload: payload, DedupKey: "bank_mt940/unreadable/" + hash})
			continue
		}
		for _, l := range st.Lines {
			fact := map[string]any{
				"date": l.Date, "amount": core.FormatMoney(l.Amount), "currency": st.Currency,
				"direction": l.Direction, "reference": l.Reference, "counterparty": l.Counterparty,
				"counterparty_account": l.CounterpartyAccount, "account": st.Account,
				"statement": st.Number, "line": l.Seq, "raw86": l.Raw86, "document": hash,
			}
			payload, err := json.Marshal(fact)
			if err != nil {
				return out, err
			}
			out = append(out, core.Event{Type: "bank.statement.line", OccurredAt: l.Date,
				Payload: payload, DedupKey: fmt.Sprintf("bank_mt940/%s/%d", hash, l.Seq)})
		}
		readOn := st.To
		if readOn == "" {
			readOn = m.today()
		}
		read, _ := json.Marshal(map[string]any{
			"document": hash, "channel": m.Name(), "event_type": "bank.statement.line", "lines": len(st.Lines),
		})
		out = append(out, core.Event{Type: EventDocumentRead, OccurredAt: readOn,
			Payload: read, DedupKey: "bank_mt940/read/" + hash})
	}
	return out, nil
}

func (m MT940) today() string {
	if m.Today != nil {
		return m.Today()
	}
	return time.Now().UTC().Format("2006-01-02")
}

// Statement is what the reader takes from an MT940.
type Statement struct {
	Reference string // :20:
	Account   string // :25:
	Number    string // :28C:
	Currency  string
	Opening   int64  // :60F:, signed
	Closing   int64  // :62F:, signed
	From, To  string // the balance dates
	Lines     []StatementLine
}

type StatementLine struct {
	Seq                 int
	Date                string // value date, YYYY-MM-DD
	Direction           string // "credit" | "debit"
	Amount              int64  // minor units, never negative
	Reference           string // the title, as the dialect gives it; else the whole :86:
	Counterparty        string
	CounterpartyAccount string
	Raw86               string
}

// IsMT940 reports whether bytes are a statement: the first tag is :20:.
func IsMT940(b []byte) bool {
	return strings.HasPrefix(strings.TrimLeft(string(b), "\r\n\t "), ":20:")
}

var (
	// :61: value date YYMMDD, optional entry date MMDD, C/D (RC/RD for
	// reversals), optional funds code, amount with a comma, then the
	// transaction type (N + 3) and the references.
	entryPattern   = regexp.MustCompile(`^(\d{6})(\d{4})?(R?[CD])([A-Z])?(\d+,\d{0,2})([A-Z][A-Z0-9]{3})(.*)$`)
	balancePattern = regexp.MustCompile(`^([CD])(\d{6})([A-Z]{3})(\d+,\d{0,2})$`)
)

// ParseMT940 reads one statement and checks that it agrees with itself: the
// opening balance plus the credits minus the debits is the closing balance.
func ParseMT940(b []byte) (Statement, error) {
	var st Statement
	text := strings.ReplaceAll(string(b), "\r\n", "\n")
	var tags []struct{ tag, value string }
	for _, ln := range strings.Split(text, "\n") {
		ln = strings.TrimRight(ln, " \t")
		if ln == "" || ln == "-" {
			continue
		}
		if strings.HasPrefix(ln, ":") {
			end := strings.Index(ln[1:], ":")
			if end < 0 {
				return st, fmt.Errorf("line %q: a tag without its closing colon", ln)
			}
			tags = append(tags, struct{ tag, value string }{ln[1 : end+1], ln[end+2:]})
			continue
		}
		if len(tags) == 0 {
			return st, fmt.Errorf("line %q before the first tag", ln)
		}
		tags[len(tags)-1].value += "\n" + ln // a continuation line
	}
	var pending *StatementLine
	var sum int64
	for _, t := range tags {
		switch t.tag {
		case "20":
			st.Reference = t.value
		case "25":
			st.Account = strings.TrimPrefix(t.value, "/")
		case "28C", "28":
			st.Number = t.value
		case "60F", "60M":
			bal, date, cur, err := parseBalance(t.value)
			if err != nil {
				return st, fmt.Errorf(":%s: %w", t.tag, err)
			}
			st.Opening, st.From, st.Currency = bal, date, cur
		case "62F", "62M":
			bal, date, _, err := parseBalance(t.value)
			if err != nil {
				return st, fmt.Errorf(":%s: %w", t.tag, err)
			}
			st.Closing, st.To = bal, date
		case "61":
			m := entryPattern.FindStringSubmatch(strings.SplitN(t.value, "\n", 2)[0])
			if m == nil {
				return st, fmt.Errorf(":61:%s is not an entry", t.value)
			}
			day, err := time.Parse("060102", m[1])
			if err != nil {
				return st, fmt.Errorf(":61: value date %q: %w", m[1], err)
			}
			amount, err := core.ParseMoney(strings.ReplaceAll(strings.TrimSuffix(m[5], ","), ",", "."))
			if err != nil {
				return st, fmt.Errorf(":61: amount %q: %w", m[5], err)
			}
			line := StatementLine{Seq: len(st.Lines) + 1, Date: day.Format("2006-01-02"), Amount: amount}
			switch m[3] {
			case "C", "RD":
				line.Direction, sum = "credit", sum+amount
			case "D", "RC":
				line.Direction, sum = "debit", sum-amount
			}
			st.Lines = append(st.Lines, line)
			pending = &st.Lines[len(st.Lines)-1]
		case "86":
			if pending == nil {
				continue // a statement-level :86:, not an entry's
			}
			pending.Raw86 = t.value
			pending.Reference, pending.Counterparty, pending.CounterpartyAccount = parse86(t.value)
			pending = nil
		}
	}
	if st.Account == "" || st.From == "" || st.To == "" {
		return st, fmt.Errorf("a statement needs :25:, :60F: and :62F:")
	}
	if st.Opening+sum != st.Closing {
		return st, fmt.Errorf("the statement contradicts itself: opening %s and the entries give %s, the closing says %s",
			core.FormatMoney(st.Opening), core.FormatMoney(st.Opening+sum), core.FormatMoney(st.Closing))
	}
	return st, nil
}

func parseBalance(v string) (int64, string, string, error) {
	m := balancePattern.FindStringSubmatch(strings.TrimSpace(v))
	if m == nil {
		return 0, "", "", fmt.Errorf("%q is not a balance", v)
	}
	day, err := time.Parse("060102", m[2])
	if err != nil {
		return 0, "", "", err
	}
	amount, err := core.ParseMoney(strings.ReplaceAll(strings.TrimSuffix(m[4], ","), ",", "."))
	if err != nil {
		return 0, "", "", err
	}
	if m[1] == "D" {
		amount = -amount
	}
	return amount, day.Format("2006-01-02"), m[3], nil
}

// parse86 reads the sim-bank dialect — kind|title|counterparty|account —
// and falls back to the whole line as the title, which is what a person
// would read anyway.
func parse86(v string) (reference, counterparty, account string) {
	parts := strings.Split(strings.ReplaceAll(v, "\n", " "), "|")
	if len(parts) >= 2 {
		reference = strings.TrimSpace(parts[1])
		if len(parts) >= 3 {
			counterparty = strings.TrimSpace(parts[2])
		}
		if len(parts) >= 4 {
			account = strings.TrimSpace(parts[3])
		}
		return
	}
	return strings.TrimSpace(v), "", ""
}
