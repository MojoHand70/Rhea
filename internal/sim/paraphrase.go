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
	for _, tok := range strings.FieldsFunc(original, func(r rune) bool {
		return r == ' ' || r == ',' || r == ';' || r == ':' || r == '(' || r == ')' || r == '\n'
	}) {
		t := strings.Trim(tok, ".'\"")
		if t == "" || seen[t] {
			continue
		}
		if !factual(t) {
			continue
		}
		seen[t] = true
		if !strings.Contains(lower, strings.ToLower(t)) {
			missing = append(missing, t)
		}
	}
	return missing
}

// factual: a token that carries a fact rather than wording — it has a
// digit, a code's dash, slash or underscore, or is a known vocabulary word
// of the language.
func factual(t string) bool {
	if strings.ContainsAny(t, "0123456789_-/") {
		return true
	}
	switch strings.ToLower(t) {
	case "open", "paid", "resolved", "half", "up", "grosz", "net", "vat", "gross", "debit", "credit":
		return true
	}
	return false
}

// SaveVoices writes the paraphrases beside the persona.
func SaveVoices(personaPath string, v Voices) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(VoicesPath(personaPath), append(b, '\n'), 0o644)
}
