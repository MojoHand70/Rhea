package exam

import (
	"context"
	"fmt"
	"strings"

	"rhea/internal/agent"
	"rhea/internal/core"
)

// Bare is the model with nothing in front of it: the baseline every claim
// is measured against. It may leave a question blank, as a candidate may.
type Bare struct{ Agent *agent.Agent }

func (Bare) Name() string { return "bare model" }

const bareSystem = `You are sitting a professional examination paper: single-choice questions,
exactly one option is right. A wrong answer costs points and a blank answer costs nothing,
so answer only when you are confident; otherwise leave it blank.
Reply with exactly one JSON object and nothing else:
{"answer": "a" | "b" | "c" | "", "note": "one sentence: why"}`

func (b Bare) Answer(ctx context.Context, q Question) (Answer, error) {
	raw, err := b.Agent.Complete(ctx, bareSystem, q.Prompt())
	if err != nil {
		return Answer{}, err
	}
	r, err := decodeReply(raw)
	if err != nil {
		return Answer{}, err
	}
	src := SourceModel
	if strings.TrimSpace(r.Answer) == "" {
		src = SourceAbstain
	}
	return Answer{Letter: r.Answer, Source: src, Note: r.Note}, nil
}

// Host is the host with the pack's rulebook open: every approved rule with
// its description and its warrant. It must say where each answer came from
// — a named rule, its own memory of the statute the pack was never taught,
// or nothing (it abstains and names the missing lesson). A rule it names
// that is not in the rulebook is not a rule: the answer is re-labelled as
// memory, so the share "from rules" is never inflated.
type Host struct {
	Agent    *agent.Agent
	rulebook string
	known    map[string]bool
}

// NewHost opens the rulebook once for every question.
func NewHost(a *agent.Agent, rules []core.Rule, types []core.ObjectType) *Host {
	h := &Host{Agent: a, known: map[string]bool{}}
	var sb strings.Builder
	sb.WriteString("The rulebook — the pack's rules, each as rule_id: description [basis: citations]:\n")
	for _, r := range rules {
		h.known[r.ID] = true
		fmt.Fprintf(&sb, "- %s: %s", r.ID, strings.TrimSpace(r.Description))
		if r.Warrant != nil {
			fmt.Fprintf(&sb, " [%s", r.Warrant.Basis)
			if len(r.Warrant.Citations) > 0 {
				fmt.Fprintf(&sb, ": %s", strings.Join(r.Warrant.Citations, " | "))
			}
			sb.WriteString("]")
		}
		sb.WriteString("\n")
	}
	if len(types) > 0 {
		sb.WriteString("The objects the pack knows: ")
		names := make([]string, 0, len(types))
		for _, t := range types {
			names = append(names, t.Name)
		}
		sb.WriteString(strings.Join(names, ", "))
		sb.WriteString("\n")
	}
	h.rulebook = sb.String()
	return h
}

func (*Host) Name() string { return "host" }

const hostSystem = `You are the host of Rhea, a rule-driven accounting system, sitting a professional
examination with the rulebook you were taught open in front of you: the approved rules of a
market pack, each with its warrant — the article or standard it rests on. Single-choice
questions, exactly one option is right. A wrong answer costs points, a blank costs nothing.

You may answer only in one of three ways, and you must say which:
- "rule": a rule in the rulebook settles the question — its description or its warrant's
  citation decides between the options. Name the rule_id and quote the citation you relied
  on. Only rules in the rulebook count; never invent one, never stretch one that merely
  touches the topic.
- "memory": no rule covers it, but you know the statute or standard yourself. Answer, name
  the act and article in "citation", and say in "note", in one line, which lesson the
  rulebook lacks. A right answer earns twice what a wrong one costs: answer from memory
  whenever you are more likely right than wrong.
- "abstain": no rule covers it and you cannot tell which option is right. Leave "answer"
  empty and say in "note", in one line, which lesson the rulebook lacks.

Reply with exactly one JSON object and nothing else:
{"answer": "a" | "b" | "c" | "", "source": "rule" | "memory" | "abstain", "rule_id": "", "citation": "", "note": ""}`

func (h *Host) Answer(ctx context.Context, q Question) (Answer, error) {
	raw, err := h.Agent.Complete(ctx, hostSystem, h.rulebook+"\nThe question:\n"+q.Prompt())
	if err != nil {
		return Answer{}, err
	}
	r, err := decodeReply(raw)
	if err != nil {
		return Answer{}, err
	}
	ans := Answer{Letter: r.Answer, Source: strings.ToLower(strings.TrimSpace(r.Source)), RuleID: r.RuleID, Citation: r.Citation, Note: r.Note}
	switch {
	case strings.TrimSpace(ans.Letter) == "":
		ans.Source = SourceAbstain
	case ans.Source == SourceRule && h.known[ans.RuleID]:
		// a rule it was taught
	case ans.Source == SourceRule:
		ans.Source = SourceMemory
		ans.Note = strings.TrimSpace("named a rule not in the rulebook (" + ans.RuleID + "); counted as memory. " + ans.Note)
		ans.RuleID = ""
	default:
		ans.Source = SourceMemory
	}
	return ans, nil
}
