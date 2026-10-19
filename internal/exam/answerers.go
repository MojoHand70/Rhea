package exam

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/statute"
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
	// Library, when set, is the statute in front of the host: before
	// answering it names the provisions it would consult, is handed their
	// text, and may then answer "statute" — only citing a provision it was
	// handed; anything else stays "memory".
	Library *statute.Library
}

// Provision is one article the host asked for and was handed.
type Provision struct {
	Act     string `json:"act"`
	Article string `json:"article"`
	Unit    string `json:"unit,omitempty"`
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

You may answer only in one of these ways, and you must say which:
- "rule": a rule in the rulebook settles the question — its description or its warrant's
  citation decides between the options. Name the rule_id and quote the citation you relied
  on. Only rules in the rulebook count; never invent one, never stretch one that merely
  touches the topic.
- "statute": no rule covers it, but a provision whose text is in front of you (under "The
  statute in front of you") decides it. Answer, and in "citation" name the act and article
  and quote the decisive words from that text. Only provisions handed to you count.
- "memory": no rule and no provision in front of you covers it, but you know the statute or
  standard yourself. Answer, name the act and article in "citation", and say in "note", in
  one line, which lesson the rulebook lacks. A right answer earns twice what a wrong one
  costs: answer from memory whenever you are more likely right than wrong.
- "abstain": nothing covers it and you cannot tell which option is right. Leave "answer"
  empty and say in "note", in one line, which lesson the rulebook lacks.

Reply with exactly one JSON object and nothing else:
{"answer": "a" | "b" | "c" | "", "source": "rule" | "statute" | "memory" | "abstain", "rule_id": "", "citation": "", "note": ""}`

const consultSystem = `You are about to answer a single-choice examination question with a library of
consolidated acts at hand. Name the provisions you would open to decide it: up to six, each
an act from the list given and an article number as the act numbers it (e.g. "28", "106e"),
optionally the unit ("ust. 1 pkt 5"). Name only acts from the list. If none applies, give an
empty list. Reply with exactly one JSON object and nothing else:
{"consult": [{"act": "", "article": "", "unit": ""}]}`

// maxProvisionChars caps what one article and all of them may put in front
// of the host; an article longer than that is cut, and says so.
const (
	maxProvisionChars = 14000
	maxStatuteChars   = 40000
)

// consult asks the host which provisions it would open and hands it their
// text. It returns the text block and the provisions actually handed.
func (h *Host) consult(ctx context.Context, q Question) (string, []Provision, error) {
	acts := h.Library.Acts()
	var names strings.Builder
	for _, a := range acts {
		fmt.Fprintf(&names, "- %s (%s)\n", a.Names[0], a.Basis)
	}
	raw, err := h.Agent.Complete(ctx, consultSystem, "The acts at hand:\n"+names.String()+"\nThe question:\n"+q.Prompt())
	if err != nil {
		return "", nil, err
	}
	var r struct {
		Consult []Provision `json:"consult"`
	}
	if err := decodeFirst(raw, &r); err != nil {
		return "", nil, err
	}
	var sb strings.Builder
	var handed []Provision
	seen := map[string]bool{}
	for _, p := range r.Consult {
		act, art, ok := h.Library.Find(p.Act, p.Article)
		if !ok || seen[act.Key+"/"+art.N] {
			continue
		}
		seen[act.Key+"/"+art.N] = true
		text := art.Text
		if len(text) > maxProvisionChars {
			text = text[:maxProvisionChars] + "\n[… the article continues; cut here]"
		}
		if sb.Len()+len(text) > maxStatuteChars {
			break
		}
		fmt.Fprintf(&sb, "[%s, art. %s — %s]\n%s\n\n", act.Names[0], art.N, act.Basis, text)
		handed = append(handed, Provision{Act: act.Names[0], Article: art.N, Unit: p.Unit})
	}
	if sb.Len() == 0 {
		return "", nil, nil
	}
	return "The statute in front of you (consolidated texts, each with its basis):\n" + sb.String(), handed, nil
}

// handed reports whether a citation names a provision the host was given.
func handed(citation string, provisions []Provision, lib *statute.Library) bool {
	c, ok := statute.ParseCitation(citation)
	if !ok {
		return false
	}
	act, ok := lib.Act(c.Act)
	if !ok {
		return false
	}
	for _, p := range provisions {
		if pa, ok := lib.Act(p.Act); ok && pa.Key == act.Key && strings.EqualFold(p.Article, c.Article) {
			return true
		}
	}
	return false
}

func (h *Host) Answer(ctx context.Context, q Question) (Answer, error) {
	var provisions []Provision
	statuteText := ""
	if h.Library != nil {
		var err error
		statuteText, provisions, err = h.consult(ctx, q)
		if err != nil {
			return Answer{}, err
		}
	}
	raw, err := h.Agent.Complete(ctx, hostSystem, h.rulebook+"\n"+statuteText+"\nThe question:\n"+q.Prompt())
	if err != nil {
		return Answer{}, err
	}
	r, err := decodeReply(raw)
	if err != nil {
		return Answer{}, err
	}
	ans := Answer{Letter: r.Answer, Source: strings.ToLower(strings.TrimSpace(r.Source)), RuleID: r.RuleID, Citation: r.Citation, Note: r.Note}
	if len(provisions) > 0 {
		b, _ := json.Marshal(provisions)
		ans.Consulted = b
	}
	switch {
	case strings.TrimSpace(ans.Letter) == "":
		ans.Source = SourceAbstain
	case ans.Source == SourceRule && h.known[ans.RuleID]:
		// a rule it was taught
	case ans.Source == SourceRule:
		ans.Source = SourceMemory
		ans.Note = strings.TrimSpace("named a rule not in the rulebook (" + ans.RuleID + "); counted as memory. " + ans.Note)
		ans.RuleID = ""
	case ans.Source == SourceStatute && h.Library != nil && handed(ans.Citation, provisions, h.Library):
		// a provision it was handed and quoted
	case ans.Source == SourceStatute:
		ans.Source = SourceMemory
		ans.Note = strings.TrimSpace("cited a provision it was not handed; counted as memory. " + ans.Note)
	default:
		ans.Source = SourceMemory
	}
	return ans, nil
}
