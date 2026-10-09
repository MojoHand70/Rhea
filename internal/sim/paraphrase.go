package sim

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// The model's one role in a simulated customer: saying the owner's answers
// the way different people would. Facts, numbers, codes and names stay
// exactly as scripted — the paraphrase is checked for that — so the eval
// measures the agent against human phrasing with ground truth untouched.

const paraphraseSystem = `You rephrase what a Polish business owner says to the person implementing
their accounting and operations system. Keep EVERY fact exactly: every number,
amount, account code, document name, type name, field name, status word, and
rounding rule. Change only wording and sentence order, the way a different
person would say the same thing — one terse, one talkative, one colloquial.
Never add a fact, never drop one, never explain. Answer with ONE JSON array of
strings and nothing else.`

// Paraphrase asks the model for n rephrasings of every interview answer and
// returns them as Voices. complete is the model call (agent.Agent.Complete).
func Paraphrase(ctx context.Context, complete func(ctx context.Context, system, user string) (string, error), model string, p Persona, n int) (Voices, error) {
	v := Voices{Model: model}
	for i, a := range p.Interview {
		user := fmt.Sprintf("Rephrase this answer %d different ways:\n\n%s", n, a.Answer)
		raw, err := complete(ctx, paraphraseSystem, user)
		if err != nil {
			return v, fmt.Errorf("answer %d: %w", i+1, err)
		}
		variants, err := decodeVariants(raw)
		if err != nil {
			return v, fmt.Errorf("answer %d: %w", i+1, err)
		}
		if len(variants) < n {
			return v, fmt.Errorf("answer %d: model gave %d variants, wanted %d", i+1, len(variants), n)
		}
		for k, s := range variants[:n] {
			if missing := Preserved(a.Answer, s); len(missing) > 0 {
				return v, fmt.Errorf("answer %d, variant %d drops %v — a paraphrase keeps every fact", i+1, k+1, missing)
			}
		}
		v.Answers = append(v.Answers, variants[:n])
	}
	return v, nil
}

// decodeVariants tolerates prose or fences around the JSON array.
func decodeVariants(raw string) ([]string, error) {
	start := strings.IndexByte(raw, '[')
	if start < 0 {
		return nil, fmt.Errorf("model returned no JSON array")
	}
	var out []string
	if err := json.NewDecoder(strings.NewReader(raw[start:])).Decode(&out); err != nil {
		return nil, fmt.Errorf("model returned unparseable JSON: %w", err)
	}
	return out, nil
}

// Preserved lists the tokens of the original that carry facts — numbers,
// codes, quoted names, identifiers with digits or underscores — missing from
// the paraphrase. Empty means every fact survived.
func Preserved(original, paraphrase string) []string {
	var missing []string
	lower := strings.ToLower(paraphrase)
	seen := map[string]bool{}
	prev := ""
	for _, tok := range strings.FieldsFunc(original, func(r rune) bool {
		return r == ' ' || r == ',' || r == ';' || r == ':' || r == '(' || r == ')' || r == '\n'
	}) {
		t := strings.Trim(tok, ".'\"")
		if t == "" {
			continue
		}
		isFact := factual(t, prev)
		prev = strings.ToLower(t)
		if !isFact || seen[t] {
			continue
		}
		seen[t] = true
		if !present(lower, strings.ToLower(t)) {
			missing = append(missing, t)
		}
	}
	orig := strings.ToLower(original)
	for _, forms := range roundingPhrases {
		if has(orig, forms) && !has(lower, forms) {
			missing = append(missing, forms[0])
		}
	}
	return missing
}

func has(s string, forms []string) bool {
	for _, f := range forms {
		if strings.Contains(s, f) {
			return true
		}
	}
	return false
}

// present finds a fact in the paraphrase; a hyphenated token ("follow-up",
// "pl-stat") may be written with its dash, a space, or joined.
func present(lower, t string) bool {
	if strings.Contains(lower, t) {
		return true
	}
	if strings.Contains(t, "-") {
		return strings.Contains(lower, strings.ReplaceAll(t, "-", " ")) || strings.Contains(lower, strings.ReplaceAll(t, "-", ""))
	}
	return false
}

// factual: a token that carries a fact rather than wording — it has a
// digit, a slash or an underscore, it names a book, type or thing ("the
// book pl-stat", "a new type named pz"), or it is a vocabulary word of the
// language. A hyphen alone is wording ("follow-up").
func factual(t, prev string) bool {
	if strings.ContainsAny(t, "0123456789_/") {
		return true
	}
	if strings.Contains(t, "-") && (prev == "book" || prev == "type" || prev == "named") {
		return true // a hyphenated name: pl-stat
	}
	switch strings.ToLower(t) {
	case "open", "paid", "resolved", "grosz", "net", "vat", "gross", "debit", "credit":
		return true
	}
	return false
}

// roundingPhrases are the rounding stances, checked as phrases: "half up"
// may be written with a space, a dash or an underscore.
var roundingPhrases = [][]string{{"half up", "half-up", "half_up"}, {"half even", "half-even", "half_even"}}

// SaveVoices writes the paraphrases beside the persona.
func SaveVoices(personaPath string, v Voices) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(VoicesPath(personaPath), append(b, '\n'), 0o644)
}
