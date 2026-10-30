package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"rhea/internal/core"
)

// ParameterAsk is a kept answer to be given its parameters (KK 2026-10-30:
// "every time Rhea creates a report she should think for what set of
// parameters user will want to run that report"): the question, the query
// that answered it and what it counts.
type ParameterAsk struct {
	Question, SQL, Reading, Today string
	Words                         *core.WordsSpec
}

// ParameterAnswer is the author's proposal: the query reading its
// parameters, and the parameters — none when nothing should vary.
type ParameterAnswer struct {
	SQL        string           `json:"sql"`
	Parameters []core.Parameter `json:"parameters,omitempty"`
}

const parameterPrompt = `You are the reporting author of Rhea, an accounting system. A person asked a question about
their business, you answered it with one DuckDB SELECT, and the person is keeping that answer as a
report they will run again. Think what they will want to change when they run it: the period, the
firm, an account, a value among a column's values, a day, a text. Make each a parameter the query
reads, and keep everything the question did not ask to vary exactly as it is - never change what
the query counts, its columns or their names.

HOW A QUERY READS A PARAMETER
- "period": the days {{<name>_from}} and {{<name>_to}} (DATEs), e.g. d BETWEEN {{period_from}} AND
  {{period_to}}; name the one period "period". Its default is the month the question asked for
  ("2026-09"), two days ("2026-09-01..2026-09-30"), or "open-month" when the question named none.
- "firm": the firm's id compared as = {{<name>}} on the column that holds a firm's id (buyer,
  seller, counterparty...). Its default is the firm's name as the question wrote it; the kernel
  finds the id. A name in the query ("c.name = 'X'") becomes the firm's id compared on the link.
- "account": the account's id as = {{<name>}}; default the account's code or name.
- "choice": a value among the values a column holds, as = {{<name>}}; give "values", the ones the
  books hold as you saw them, and the default the question asked for.
- "date": one day as {{<name>}} (a DATE); default the day the question named.
- "text": a text as {{<name>}}, quoted by the kernel; use it rarely.
A name is lowercase letters, digits and underscores; a label is 1-3 words in the question's
language. A question for one month with no firm gives one parameter, the period; "sales to X in
September" gives two.

Reply with only the JSON object: {"sql": "...", "parameters": [{"name": "...", "kind": "...",
"label": "...", "default": "...", "values": [...]}]}. When nothing should vary, reply with the query
unchanged and "parameters": [].`

// Parameterise asks the author which parameters a kept report takes, and
// how its query reads them.
func (a *Agent) Parameterise(ctx context.Context, ask ParameterAsk) (ParameterAnswer, error) {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Today: %s\nQuestion: %s\nWhat the query counts: %s\nQuery:\n%s\n", ask.Today, ask.Question, ask.Reading, ask.SQL)
	text, err := a.Complete(ctx, parameterPrompt, sb.String())
	if err != nil {
		return ParameterAnswer{}, err
	}
	var out ParameterAnswer
	if err := json.Unmarshal([]byte(extractAnswer(text)), &out); err != nil {
		return ParameterAnswer{}, fmt.Errorf("parameters: %w", err)
	}
	out.SQL = strings.TrimSpace(out.SQL)
	if out.SQL == "" {
		return ParameterAnswer{}, fmt.Errorf("parameters: no query")
	}
	seen := map[string]bool{}
	for _, p := range out.Parameters {
		if err := p.Check(); err != nil {
			return ParameterAnswer{}, fmt.Errorf("parameters: %w", err)
		}
		if seen[p.Name] {
			return ParameterAnswer{}, fmt.Errorf("parameters: %s named twice", p.Name)
		}
		seen[p.Name] = true
	}
	return out, nil
}
