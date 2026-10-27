package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"rhea/internal/core"
)

// MT940 is the reader of the bank channel (DIRECTION 2026-10-10, step 5):
// every received statement in the store becomes one bank.statement.line per
// entry — the value date, the amount and its direction, the title and the
// counterparty as the bank wrote them — each naming the file, and a
// document.read beside them. The reading is exact. A statement that
// contradicts itself (the opening balance and the entries do not give the
// closing balance) is reported as document.unreadable for a person, and so
// is a file that came through the bank door but is not a statement.
//
// Every account a statement is of is said once, the first time a statement
// of it is read: bank.account.seen, keyed by the account (code version 2,
// KK 2026-10-27: Rachunki bankowe stayed empty while statements were in the
// books) — a pack proposes the account from it, for a person to approve.
//
// A file may carry several statements (one per :20:), each over one or more
// balance pages (:60F:/:62F:, :60M:/:62M: at the page breaks); every page
// is checked on its own and the line numbers run through the file. A
// reversal (RC/RD in :61:) keeps its direction's effect and says so.
//
// The :86: line is the one trap: every bank writes it differently. The
// reader keeps it raw on every line (raw86) and understands one dialect,
// "sim-bank" (kind|title|counterparty|account) with a known kind in front;
// any other line is the title whole. A real bank's dialect is a second
// parser behind the same channel, versioned like the channel.
type MT940 struct {
	Today func() string
}

const mt940CodeVersion = 2

func init() {
	Register(Implemented{Name: "bank_mt940", CodeVersion: mt940CodeVersion,
		Format: "MT940 statements from the document store; :86: kept raw, the sim-bank dialect read"})
	readers["bank_mt940"] = func() Adapter { return MT940{} }
}

// EventBankAccountSeen: an account a statement is of, said once.
const EventBankAccountSeen = "bank.account.seen"

func (MT940) Name() string     { return "bank_mt940" }
func (MT940) CodeVersion() int { return mt940CodeVersion }

func (m MT940) Pass(ctx context.Context, reads Reads) ([]core.Event, error) {
	docs, err := reads.ObjectsByType(ctx, ReceivedDocumentType)
	if err != nil {
		return nil, err
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].ID < docs[j].ID })
	// Lines logged before the key named the statement's day were keyed by
	// account, the year and statement number and line, or without the year:
	// such a line read again from another file is the one already logged
	// when its date is the same.
	legacy := map[string]bool{}
	if er, ok := reads.(interface {
		EventsOfTypes(ctx context.Context, types []string) ([]core.Event, error)
	}); ok {
		logged, err := er.EventsOfTypes(ctx, []string{"bank.statement.line"})
		if err != nil {
			return nil, err
		}
		for _, e := range logged {
			legacy[e.DedupKey+"@"+e.OccurredAt] = true
		}
	}
	var out []core.Event
	seen := map[string]bool{} // the accounts this pass has said
	for _, d := range docs {
		status, _ := d.State["status"].(string)
		source, _ := d.State["source"].(string)
		hash, _ := d.State["hash"].(string)
		name, _ := d.State["name"].(string)
		if status != "received" || hash == "" || source != "bank" {
			continue // the bank channel reads what came through the bank door
		}
		b, err := reads.GetBlob(ctx, hash)
		if err != nil {
			return out, fmt.Errorf("%s names file %s: %w", d.ID, hash, err)
		}
		statements, err := ParseMT940(b.Bytes)
		if err != nil {
			payload, _ := json.Marshal(map[string]any{
				"document": hash, "name": name, "channel": m.Name(), "reason": err.Error(),
			})
			out = append(out, core.Event{Type: EventDocumentUnreadable, OccurredAt: m.today(),
				Payload: payload, DedupKey: "bank_mt940/unreadable/" + hash})
			continue
		}
		lines := 0
		for _, st := range statements {
			if !seen[st.Account] {
				seen[st.Account] = true
				said, _ := json.Marshal(map[string]any{"account": st.Account, "currency": st.Currency, "date": st.From,
					"statement": st.Number, "document": hash})
				out = append(out, core.Event{Type: EventBankAccountSeen, OccurredAt: st.From, Payload: said,
					DedupKey: "bank_mt940/account/" + st.Account})
			}
			for _, l := range st.Lines {
				lines++
				fact := map[string]any{
					"date": l.Date, "amount": core.FormatMoney(l.Amount), "currency": st.Currency,
					"direction": l.Direction, "reference": l.Reference, "counterparty": l.Counterparty,
					"counterparty_account": l.CounterpartyAccount, "account": st.Account,
					"statement": st.Number, "line": lines, "raw86": l.Raw86, "document": hash,
					"code": l.Code,
				}
				if l.Reversal {
					fact["reversal"] = true
				}
				payload, err := json.Marshal(fact)
				if err != nil {
					return out, err
				}
				// a re-downloaded statement is the same statement: keyed on
				// the account, the day the statement opens, its number and
				// the line. Banks number statements within the year, and
				// without the year January 2027's statement 38 was January
				// 2026's (review 2026-10-21, D1); some write the same number
				// on every daily statement, and with only the year the
				// second day's lines were the first day's (review
				// 2026-10-27). A line logged under an older key — the
				// year's, or neither — is known by that key and its date.
				if legacy[fmt.Sprintf("bank_mt940/%s/%s/%d", st.Account, st.Number, l.Seq)+"@"+l.Date] ||
					legacy[fmt.Sprintf("bank_mt940/%s/%.4s/%s/%d", st.Account, st.From, st.Number, l.Seq)+"@"+l.Date] {
					continue
				}
				out = append(out, core.Event{Type: "bank.statement.line", OccurredAt: l.Date,
					Payload: payload, DedupKey: fmt.Sprintf("bank_mt940/%s/%s/%s/%d", st.Account, st.From, st.Number, l.Seq)})
			}
		}
		read, _ := json.Marshal(map[string]any{
			"document": hash, "channel": m.Name(), "event_type": "bank.statement.line", "lines": lines,
		})
		out = append(out, core.Event{Type: EventDocumentRead, OccurredAt: statements[len(statements)-1].To,
			Payload: read, DedupKey: "bank_mt940/read/" + hash})
	}
	return out, nil
}

func (m MT940) today() string {
	if m.Today != nil {
		return m.Today()
	}
	return core.Today()
}

// Statement is what the reader takes from one :20: of an MT940: the first
// page's opening balance, the last page's closing balance, every entry.
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
	Direction           string // "credit" | "debit": the effect on the account
	Reversal            bool   // RC/RD in :61:: the entry reverses an earlier one
	Code                string // the transaction type in :61: — NTRF a transfer, NCHG a charge, NINT interest
	Amount              int64  // minor units, never negative
	Reference           string // the title, as the dialect gives it; else the whole :86:
	Counterparty        string
	CounterpartyAccount string
	Raw86               string
}

type mt940Tag struct{ tag, value string }

// IsMT940 reports whether bytes are a statement: the first tag is :20:,
// after any SWIFT block header ({1:…}) or :940: line.
func IsMT940(b []byte) bool {
	tags, err := mt940Tags(b)
	return err == nil && len(tags) > 0
}

// mt940Tags splits a file into its tags with their continuation lines; the
// first must be :20:. SWIFT block lines ({…}), the "-" terminator and a
// :940: header carry no statement data and are dropped.
// mt940TagStart is how a tag line begins: a colon, two digits, an optional
// letter, a colon.
var mt940TagStart = regexp.MustCompile(`^:([0-9]{2}[A-Z]?):`)

func mt940Tags(b []byte) ([]mt940Tag, error) {
	var tags []mt940Tag
	if !utf8.Valid(b) {
		// Polish banks commonly send CP1250 or ISO-8859-2: read as UTF-8
		// the titles would arrive mangled with no warning.
		return nil, fmt.Errorf("the file is not UTF-8 (banks often send CP1250 or ISO-8859-2) — convert it before reading")
	}
	text := strings.ReplaceAll(string(b), "\r\n", "\n")
	for _, ln := range strings.Split(text, "\n") {
		ln = strings.TrimRight(ln, " \t")
		if ln == "" || ln == "-" || ln == "-}" || strings.HasPrefix(ln, "{") || ln == ":940:" {
			continue
		}
		if len(tags) == 0 && !strings.HasPrefix(ln, ":20:") {
			return nil, fmt.Errorf("not an MT940 statement: the file starts with %q, not :20:", ln)
		}
		// A tag is :NN: or :NNa:; a continuation line of free text may start
		// with a colon too and is not one.
		if m := mt940TagStart.FindStringSubmatch(ln); m != nil {
			tags = append(tags, mt940Tag{m[1], ln[len(m[0]):]})
			continue
		}
		if len(tags) == 0 {
			return nil, fmt.Errorf("line %q: a tag without its closing colon", ln)
		}
		tags[len(tags)-1].value += "\n" + ln // a continuation line
	}
	return tags, nil
}

var (
	// :61: value date YYMMDD, optional entry date MMDD, C/D (RC/RD for
	// reversals), optional funds code, amount with a comma, then the
	// transaction type (N + 3) and the references.
	entryPattern   = regexp.MustCompile(`^(\d{6})(\d{4})?(R?[CD])([A-Z])?(\d+,\d{0,2})([A-Z][A-Z0-9]{3})(.*)$`)
	balancePattern = regexp.MustCompile(`^([CD])(\d{6})([A-Z]{3})(\d+,\d{0,2})$`)
)

// ParseMT940 reads a file of statements, one per :20:, and checks that each
// agrees with itself page by page: the page's opening balance plus the
// credits minus the debits is its closing balance, and a page opens where
// the one before closed.
func ParseMT940(b []byte) ([]Statement, error) {
	tags, err := mt940Tags(b)
	if err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return nil, fmt.Errorf("not an MT940 statement: an empty file")
	}
	var out []Statement
	for start := 0; start < len(tags); {
		end := start + 1
		for end < len(tags) && tags[end].tag != "20" {
			end++
		}
		var prev *Statement
		if len(out) > 0 {
			prev = &out[len(out)-1]
		}
		st, err := parseStatement(tags[start:end], prev)
		if err != nil {
			if len(out) > 0 {
				return nil, fmt.Errorf("statement %d (%s): %w", len(out)+1, tags[start].value, err)
			}
			return nil, err
		}
		out = append(out, st)
		start = end
	}
	return out, nil
}

// parseStatement reads the tags of one :20: as a run of balance pages. A
// :60M: at its start continues the statement before it in the file and
// must open where that one closed.
func parseStatement(tags []mt940Tag, prev *Statement) (Statement, error) {
	var st Statement
	var pending *StatementLine
	var pageOpen bool      // between an opening balance and its closing
	var opening, sum int64 // the page's opening balance and its entries so far
	pages := 0
	for _, t := range tags {
		if t.tag != "86" && t.tag != "61" {
			pending = nil // an :86: after a balance line is the statement's, not the last entry's
		}
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
			if pageOpen {
				return st, fmt.Errorf(":%s: a page opens before the one before it closed", t.tag)
			}
			if pages == 0 && t.tag == "60M" && prev != nil && bal != prev.Closing {
				return st, fmt.Errorf(":60M: the page opens at %s, the statement before closed at %s",
					core.FormatMoney(bal), core.FormatMoney(prev.Closing))
			}
			if pages == 0 {
				st.Opening, st.From, st.Currency = bal, date, cur
			} else if bal != st.Closing {
				return st, fmt.Errorf(":%s: the page opens at %s, the one before closed at %s",
					t.tag, core.FormatMoney(bal), core.FormatMoney(st.Closing))
			}
			opening, sum, pageOpen = bal, 0, true
			pages++
		case "62F", "62M":
			bal, date, _, err := parseBalance(t.value)
			if err != nil {
				return st, fmt.Errorf(":%s: %w", t.tag, err)
			}
			if !pageOpen {
				return st, fmt.Errorf(":%s: a closing balance without an opening", t.tag)
			}
			if opening+sum != bal {
				return st, fmt.Errorf("the statement contradicts itself: opening %s and the entries give %s, the closing says %s",
					core.FormatMoney(opening), core.FormatMoney(opening+sum), core.FormatMoney(bal))
			}
			st.Closing, st.To, pageOpen = bal, date, false
		case "61":
			if !pageOpen {
				return st, fmt.Errorf(":61:%s: an entry outside a balance page", t.value)
			}
			m := entryPattern.FindStringSubmatch(strings.SplitN(t.value, "\n", 2)[0])
			if m == nil {
				return st, fmt.Errorf(":61:%s is not an entry", t.value)
			}
			day, err := time.Parse("060102", m[1])
			if err != nil {
				return st, fmt.Errorf(":61: value date %q: %w", m[1], err)
			}
			amount, err := mt940Amount(m[5])
			if err != nil {
				return st, fmt.Errorf(":61: amount %q: %w", m[5], err)
			}
			if amount >= 1e16 || sum >= 1e17 || sum <= -1e17 {
				return st, fmt.Errorf(":61: amount %q is past what the books hold", m[5])
			}
			line := StatementLine{Seq: len(st.Lines) + 1, Date: day.Format("2006-01-02"), Amount: amount,
				Reversal: strings.HasPrefix(m[3], "R"), Code: m[6]}
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
	if st.Account == "" || pages == 0 || pageOpen {
		return st, fmt.Errorf("a statement needs :25:, :60F: and :62F:")
	}
	if st.Number == "" {
		return st, fmt.Errorf("a statement needs its number (:28C:): a line's identity is the account, the statement and the line")
	}
	return st, nil
}

// mt940Amount reads an amount as the format writes it: digits, a comma,
// up to two decimals ("1500,5" is 1500.50).
func mt940Amount(s string) (int64, error) {
	whole, frac, _ := strings.Cut(s, ",")
	for len(frac) < 2 {
		frac += "0"
	}
	return core.ParseMoney(whole + "." + frac)
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
	amount, err := mt940Amount(m[4])
	if err != nil {
		return 0, "", "", err
	}
	if m[1] == "D" {
		amount = -amount
	}
	return amount, day.Format("2006-01-02"), m[3], nil
}

// simBankKinds are the kinds the sim-bank dialect writes first in :86:;
// a line that does not start with one is another bank's and stays whole.
var simBankKinds = map[string]bool{"PRZELEW PRZYCHODZACY": true, "PRZELEW WYCHODZACY": true}

// parse86 reads the sim-bank dialect — kind|title|counterparty|account,
// exactly four fields behind a known kind — and otherwise keeps the whole
// line as the title, which is what a person would read anyway.
func parse86(v string) (reference, counterparty, account string) {
	flat := strings.TrimSpace(strings.ReplaceAll(v, "\n", " "))
	if parts := strings.Split(flat, "|"); len(parts) == 4 && simBankKinds[strings.TrimSpace(parts[0])] {
		return strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2]), strings.TrimSpace(parts[3])
	}
	return flat, "", ""
}
