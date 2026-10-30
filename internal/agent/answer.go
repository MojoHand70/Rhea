package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"rhea/internal/core"
)

// AnswerAsk is a question answered from the books now (DIRECTION 2026-10-17,
// "Reporting"): the author writes one read-only query over the collection
// and says plainly what it counts; the kernel runs it locked and shows the
// rows. She never states a figure herself.
type AnswerAsk struct {
	Question string
	Today    string
	Tables   []CollectionTable
	Words    *core.WordsSpec // the market's names for types, fields and values; nil without a words view
	Books    []string        // the books the active rules post to
	// Reports are the company's kept reports (KK 2026-10-30): a question
	// asking for one of them, for a period it takes, is answered by naming
	// it, never by writing it again.
	Reports []AnswerReport
	// Tried is the earlier query of this very question and what DuckDB said
	// of it: the retry fixes exactly that.
	Tried *AnswerTry
}

// AnswerReport is a kept report as the author sees it: its id, its title,
// what its parts count, and whether it takes a period.
type AnswerReport struct {
	ID, Title  string
	Parts      []string
	Period     bool
	Parameters []core.Parameter // what a person sets before it is computed (KK 2026-10-30)
}

// CollectionTable is one table of the collection as the author sees it.
type CollectionTable struct {
	Schema, Name, Type string
	Columns            []CollectionColumn
}

// CollectionColumn is one column: its name, DuckDB type, the field's declared
// kind (money, date, ref<company>…), and for a link the table it points to.
type CollectionColumn struct {
	Name, SQL, Kind, Links string
}

// AnswerTry is a query that failed and why.
type AnswerTry struct {
	SQL, Error string
}

// Answer is the author's answer: the query, what it counts, how to show it;
// or why the books cannot answer.
type Answer struct {
	SQL     string      `json:"sql,omitempty"`
	Reading string      `json:"reading,omitempty"`
	Visual  core.Visual `json:"visual"`
	Cannot  string      `json:"cannot,omitempty"`
	// Report names a kept report that answers the question, with the
	// period it is asked for where the report takes one (KK 2026-10-30).
	Report string `json:"report,omitempty"`
	From   string `json:"from,omitempty"`
	To     string `json:"to,omitempty"`
	// Parameters are the values asked for the report's parameters, by name:
	// a period as a month or two days, a firm or an account by name or id,
	// a choice among its values, a day, a text.
	Parameters map[string]string `json:"parameters,omitempty"`
	// Lesson is the kind of question in general words — no names, digits or
	// dates — what reaches the Rhea team if she cannot answer it.
	Lesson string `json:"lesson,omitempty"`
}

const answerPrompt = `You are the reporting author of Rhea, an accounting system. A person asks a question about
their business; you answer it from the books by writing ONE DuckDB SELECT over the business's
collection, and you say plainly what it counts. You never state a figure: the kernel runs your
query in a locked, read-only session and shows the rows it returns.

THE COLLECTION - one schema per domain, one table per kind of object (shown below with each
column's type and the name the business uses for it). Every table has "id", a column per
field, and provenance columns (_event, _rule, _occurred_at, _root_event, and _state: the whole
object as JSON). Money is DECIMAL(18,2): exact - add it, never cast it to a float. A link column
holds the linked object's id: join it to that table's "id". Bookings are the postings table:
each line has an account (join the account table for its code, name and type), a side (debit or
credit) and an unsigned amount - signed, a debit is plus for assets and costs, a credit is plus
for revenue and liabilities. Keep bookings to the books named below. What was paid is money that
left the bank: the bank operations going out, or the debit side of the account the payment
settled (a liability, such as social security or tax, is paid on its debit side, owed on its
credit side) - never the credit side of a liability. What was received is the reverse.

THE QUERY
- One SELECT: WITH, joins, GROUP BY, window functions, PIVOT ... FOR ... IN (...) are all fine.
  Nothing but SELECT; only the collection's tables. Quote names that need it ("date", "number").
- It is DuckDB, not Postgres: no to_char (a month is strftime(d, '%Y-%m'), a quarter is
  'Q' || quarter(d)), no ::money; ORDER BY after a UNION names only the output columns — to put
  a total row last, UNION ALL inside a subquery with a sort column and order the outer select
  by it, or use GROUP BY ROLLUP.
- Today is given below; "this month", "last month", "this year", "so far" are relative to it.
  Write it as DATE '<today>'; never CURRENT_DATE, now() or any clock of the machine.
- Money stays exact: an average, a share or a ratio is ROUND(..., 2)::DECIMAL(18,2) (a percent
  ROUND(... * 100, 1)::DECIMAL(9,1)); never let a column come back as a float (DOUBLE, REAL).
- Name every column you show in the language of the question, with AS "...".
- When rows add up or list objects, add a column _of naming them: string_agg(t.id, ',' ORDER BY
  t.id) AS _of for a total, t.id AS _of for a listing - the person opens "why?" from it. _of is
  never shown.
- A total over nothing is zero, not empty: COALESCE(sum(...), 0).
- Order rows the way a reader expects (by month, by amount descending for a ranking). Keep a
  listing to what was asked; a ranking without a number given shows the top ten.

THE READING - "reading": one plain sentence, in the language of the question, of what the query
counts: which objects, which period, which amount, how grouped. No figures, no SQL words.

THE VISUAL - "visual": how the rows are shown.
  {"kind": "kpi", "value": "<col>", "label": "<col>", "compare": "<col>"}  one number (label and
     compare optional; compare is the figure it is set against, e.g. last month)
  {"kind": "table", "columns": ["<col>", ...]}  a listing (columns optional: all but _of)
  {"kind": "statement", "line": "<col>", "amount": "<col>", "op": "<col>", "section": "<col>"}
     lines of a financial statement; op holds '', '+', '-' or '=' (op and section optional)
  {"kind": "bar" | "line", "x": "<col>", "y": ["<col>", ...], "series": "<col>"}  a chart
  {"kind": "pivot", "columns": [...]}  a cross table made with PIVOT; the first column heads rows
  Add "title" (short, in the question's language) and, where a column's type does not say it,
  "formats": {"<col>": "money" | "percent" | "int" | "date" | "month" | "text"}.

THE LESSON - "lesson": the kind of question in general words, in the question's language, with no
names of companies or people and no digits, dates or amounts (e.g. "wpłaty do ZUS w kwartale,
miesiąc po miesiącu, z sumą"). If the question cannot be answered, it is what reaches the Rhea
team so they can teach it.

A KEPT REPORT - the company's reports already kept and approved are listed below, each with what
its parts count and the parameters a person sets before it is computed (a period, a firm, an
account, a choice among values, a date, a text; each with its default). When the question asks for
one of them - the same report, for other values of its parameters or as it is - do not write a
query: name it. {"report": "<its id>", "parameters": {"<name>": "<value>", ...}, "reading": "<one
plain sentence: which report, for what>"}: a period as a month "2026-08" or two days
"2026-08-01..2026-08-31", a firm or an account by its name as the question wrote it, a choice as one
of its values, a date as a day; leave out a parameter the question does not change, and give none
for a report with no parameters. Prefer the kept report to a query of your own whenever it is what
was asked.

When the collection holds nothing that could answer the question, say so instead:
{"cannot": "<why, in the question's language, without figures>", "lesson": "..."}.

Reply with only the JSON object: {"sql": "...", "reading": "...", "visual": {...}, "lesson": "..."},
{"report": "...", "parameters": {...}, "reading": "..."} or {"cannot": "...", "lesson": "..."}.`

// Answer asks the author for a query that answers the question.
func (a *Agent) Answer(ctx context.Context, ask AnswerAsk) (Answer, error) {
	text, err := a.Complete(ctx, answerPrompt, answerMessage(ask))
	if err != nil {
		return Answer{}, err
	}
	var out Answer
	if err := json.Unmarshal([]byte(extractAnswer(text)), &out); err != nil {
		return Answer{}, fmt.Errorf("answer: %w", err)
	}
	out.SQL, out.Cannot, out.Report = strings.TrimSpace(out.SQL), strings.TrimSpace(out.Cannot), strings.TrimSpace(out.Report)
	switch {
	case out.Report != "" && out.SQL == "":
		if !slices.ContainsFunc(ask.Reports, func(r AnswerReport) bool { return r.ID == out.Report }) {
			return Answer{}, fmt.Errorf("answer: names a kept report %q the company has not", out.Report)
		}
		return out, nil
	case out.Cannot != "" && out.SQL == "":
		return out, nil
	case out.SQL == "":
		return Answer{}, fmt.Errorf("answer: no query and no reason why not")
	case strings.TrimSpace(out.Reading) == "":
		return Answer{}, fmt.Errorf("answer: the query says nothing of what it counts")
	}
	return out, nil
}

func answerMessage(ask AnswerAsk) string {
	var sb strings.Builder
	writeCollection(&sb, ask.Tables, ask.Words)
	if len(ask.Books) > 0 {
		fmt.Fprintf(&sb, "\nBooks the active rules post to: %s\n", strings.Join(ask.Books, ", "))
	}
	if len(ask.Reports) > 0 {
		sb.WriteString("\nKept reports (approved; name one when it is what was asked):\n")
		for _, r := range ask.Reports {
			params := "no parameters, as kept"
			if len(r.Parameters) > 0 {
				var ps []string
				for _, p := range r.Parameters {
					d := p.Name + " (" + p.Kind
					if p.Label != "" {
						d += ", " + p.Label
					}
					if p.Default != "" {
						d += ", default " + p.Default
					}
					if len(p.Values) > 0 {
						d += ", values: " + strings.Join(p.Values, " | ")
					}
					ps = append(ps, d+")")
				}
				params = "parameters: " + strings.Join(ps, "; ")
			} else if r.Period {
				params = "takes a period"
			}
			fmt.Fprintf(&sb, "- %s \u201e%s\u201d - %s - %s\n", r.ID, r.Title, strings.Join(r.Parts, "; "), params)
		}
	}
	fmt.Fprintf(&sb, "\nToday: %s\n", ask.Today)
	if ask.Tried != nil {
		fmt.Fprintf(&sb, "\nYour previous query for this question failed:\n%s\nDuckDB said: %s\nFix exactly that.\n", ask.Tried.SQL, ask.Tried.Error)
	}
	fmt.Fprintf(&sb, "\nQuestion: %s\n", ask.Question)
	return sb.String()
}

// writeCollection describes the collection table by table, in the market's
// own names: what a query of hers may read; then the market's notes on how
// its books are read.
func writeCollection(sb *strings.Builder, tables []CollectionTable, words *core.WordsSpec) {
	defer func() {
		if words == nil || len(words.Notes) == 0 {
			return
		}
		sb.WriteString("\nHow this market's books are read (the market pack's notes - follow them):\n")
		for _, n := range words.Notes {
			fmt.Fprintf(sb, "- %s\n", n)
		}
	}()
	label := func(key string) string {
		if words == nil {
			return ""
		}
		return words.Fields[key]
	}
	sb.WriteString("The collection:\n")
	for _, t := range tables {
		name := ""
		if words != nil {
			if tw, ok := words.Types[t.Type]; ok {
				name = " — " + tw.Label
			}
		}
		fmt.Fprintf(sb, "%s.%s%s\n", t.Schema, t.Name, name)
		for _, c := range t.Columns {
			if strings.HasPrefix(c.Name, "_") {
				continue // the provenance is the same everywhere, said once above
			}
			fmt.Fprintf(sb, "  %s %s", c.Name, c.SQL)
			if c.Kind != "" && c.Kind != "string" {
				fmt.Fprintf(sb, " (%s)", c.Kind)
			}
			if c.Links != "" {
				fmt.Fprintf(sb, " → %s.id", c.Links)
			}
			l := label(t.Type + "." + c.Name)
			if l == "" {
				l = label(c.Name)
			}
			if l != "" {
				fmt.Fprintf(sb, " — %s", l)
			}
			if words != nil {
				if vals := words.Values[t.Type+"."+c.Name]; len(vals) > 0 {
					var vs []string
					for k, v := range vals {
						vs = append(vs, k+"="+v)
					}
					fmt.Fprintf(sb, " [%s]", strings.Join(vs, ", "))
				}
			}
			sb.WriteByte('\n')
		}
	}
}

// extractAnswer is the object the model meant: the largest in the text that
// carries a query or a reason, else the text from its first brace.
func extractAnswer(s string) string {
	first := strings.IndexByte(s, '{')
	if first < 0 {
		return s
	}
	best := ""
	for i := first; i < len(s); i++ {
		if s[i] != '{' {
			continue
		}
		var v json.RawMessage
		dec := json.NewDecoder(strings.NewReader(s[i:]))
		if err := dec.Decode(&v); err != nil {
			continue
		}
		if (strings.Contains(string(v), `"sql"`) || strings.Contains(string(v), `"cannot"`)) && len(v) > len(best) {
			best = string(v)
		}
		i += int(dec.InputOffset()) - 1
	}
	if best != "" {
		return best
	}
	return s[first:]
}
