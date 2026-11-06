package core

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// The stream says what happened in the pack's own words (KK 2026-11-05:
// "each event should be an event", and a rule's words must not "sound like
// a Robocop"). A pack gives, per rule, the sentence its step reads as, with
// placeholders filled from the step's facts; the kernel renders it into
// parts a face sets without knowing the language or the business: text, a
// figure for the monospace, money with its currency, a date or a month.

// StepWords is how one step of the stream reads: keyed in WordsSpec.Steps
// by the rule's id; by "event:<type>" for the title of a raw event's card;
// by "type:<object type>" as the sentence for any rule that makes that
// type and has no words of its own.
type StepWords struct {
	// Says is the sentence template. Empty on a rule's key, the step is
	// silent: said by another step of the chain (a settlement record by the
	// settled document's step). An event's title and a type's fallback say
	// something.
	Says string `json:"says"`
	// Area names who handled the step, in the pack's words: Bank, Księgi,
	// Rozrachunki, KSeF, VAT. A face shows it as the step's chip.
	Area string `json:"area,omitempty"`
	// State: "" or "ok" — done; "wait" — a person must act on it; "no" —
	// refused. The card takes the gravest state of its steps.
	State string `json:"state,omitempty"`
}

// SayPart is one piece of a rendered sentence. Exactly one of T, F, M, D is
// set: text; a figure (a number, a code, a count — set in the monospace);
// money in minor units as a string with its currency; a date YYYY-MM-DD,
// or a month YYYY-MM when As is "month". The face formats figures, money
// and dates by its locale; the words are the pack's.
type SayPart struct {
	T  string `json:"t,omitempty"`
	F  string `json:"f,omitempty"`
	M  string `json:"m,omitempty"`
	C  string `json:"c,omitempty"`
	D  string `json:"d,omitempty"`
	As string `json:"as,omitempty"`
}

var sayFormats = []string{"f", "m", "d", "month", "value", "plural"}

// Say renders a template over the facts a lookup answers for. Placeholders:
// {name} the value as text; {name|f} a figure; {name|m} money (a decimal
// string is parsed, a number is minor units; the currency is looked up as
// "currency"); {name|d} a date; {name|month} a month; {name|value:type.field}
// the words' label of an enum value; {name|plural:one,few,many} the noun
// form for the count by the words' language. A name the facts lack renders
// as nothing.
func (w WordsSpec) Say(tpl string, lookup func(string) (any, bool)) []SayPart {
	var parts []SayPart
	text := func(s string) {
		if s == "" {
			return
		}
		if n := len(parts); n > 0 && parts[n-1].T != "" {
			parts[n-1].T += s
			return
		}
		parts = append(parts, SayPart{T: s})
	}
	rest := tpl
	for {
		i := strings.IndexByte(rest, '{')
		if i < 0 {
			text(rest)
			return parts
		}
		j := strings.IndexByte(rest[i:], '}')
		if j < 0 {
			text(rest)
			return parts
		}
		text(rest[:i])
		ph := rest[i+1 : i+j]
		rest = rest[i+j+1:]
		name, format, _ := strings.Cut(ph, "|")
		format, arg, _ := strings.Cut(format, ":")
		v, ok := lookup(strings.TrimSpace(name))
		if !ok || v == nil {
			continue
		}
		switch format {
		case "":
			text(fmt.Sprint(v))
		case "f":
			parts = append(parts, SayPart{F: figure(v)})
		case "m":
			minor, ok := moneyOf(v)
			if !ok {
				text(fmt.Sprint(v))
				continue
			}
			p := SayPart{M: strconv.FormatInt(minor, 10)}
			if c, ok := lookup("currency"); ok {
				p.C = fmt.Sprint(c)
			}
			parts = append(parts, p)
		case "d":
			parts = append(parts, SayPart{D: fmt.Sprint(v)})
		case "month":
			s := fmt.Sprint(v)
			if len(s) > 7 {
				s = s[:7]
			}
			parts = append(parts, SayPart{D: s, As: "month"})
		case "value":
			s := fmt.Sprint(v)
			if lbl := w.valueWord(arg, s); lbl != "" {
				s = lbl
			}
			text(s)
		case "plural":
			n, ok := countOf(v)
			if !ok {
				continue
			}
			text(pluralForm(w.Lang, n, strings.Split(arg, ",")))
		default:
			text(fmt.Sprint(v))
		}
	}
}

// valueWord is the words' label for an enum value: the field's own ("type.field"),
// then the bare field's.
func (w WordsSpec) valueWord(field, v string) string {
	if m, ok := w.Values[field]; ok {
		if s := m[v]; s != "" {
			return s
		}
	}
	if _, f, ok := strings.Cut(field, "."); ok {
		if m, ok := w.Values[f]; ok {
			return m[v]
		}
	}
	return ""
}

func figure(v any) string {
	switch x := v.(type) {
	case float64:
		if x == math.Trunc(x) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	default:
		return fmt.Sprint(v)
	}
}

// moneyOf reads money as the log carries it: a number is minor units (an
// object's state), a decimal string is an amount to parse (a raw payload).
func moneyOf(v any) (int64, bool) {
	switch x := v.(type) {
	case int64:
		return x, true
	case int:
		return int64(x), true
	case float64:
		return int64(math.Round(x)), true
	case string:
		m, err := ParseMoney(x)
		return m, err == nil
	}
	return 0, false
}

func countOf(v any) (int, bool) {
	switch x := v.(type) {
	case int:
		return x, true
	case int64:
		return int(x), true
	case float64:
		return int(x), true
	case string:
		n, err := strconv.Atoi(x)
		return n, err == nil
	}
	return 0, false
}

// pluralForm picks the noun's form for a count by the language: Polish and
// its kin count one, a few (2–4 but not 12–14) and many; the rest one and
// many. A language's number words are language, not market: the kernel
// imports no market by knowing how Polish counts.
func pluralForm(lang string, n int, forms []string) string {
	for i := range forms {
		forms[i] = strings.TrimSpace(forms[i])
	}
	one, many := forms[0], forms[len(forms)-1]
	few := many
	if len(forms) >= 3 {
		few = forms[1]
	}
	if n < 0 {
		n = -n
	}
	switch lang {
	case "pl", "cs", "sk", "uk", "ru":
		switch {
		case n == 1:
			return one
		case n%10 >= 2 && n%10 <= 4 && (n%100 < 12 || n%100 > 14):
			return few
		default:
			return many
		}
	default:
		if n == 1 {
			return one
		}
		return many
	}
}

// checkSteps refuses step words a face could not render: a key naming
// nothing, an unclosed placeholder, a format the engine lacks, a plural
// with one form, a value with no field, a state that is none, a silent
// title or fallback.
func (w WordsSpec) checkSteps() error {
	for key, sw := range w.Steps {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("words: a step keyed by nothing")
		}
		if strings.TrimSpace(sw.Says) == "" && (strings.HasPrefix(key, "event:") || strings.HasPrefix(key, "type:")) {
			return fmt.Errorf("words: step %s says nothing; only a rule's step may be silent", key)
		}
		switch sw.State {
		case "", "ok", "wait", "no":
		default:
			return fmt.Errorf("words: step %s: state %q is not ok, wait or no", key, sw.State)
		}
		rest := sw.Says
		for {
			i := strings.IndexByte(rest, '{')
			if i < 0 {
				break
			}
			j := strings.IndexByte(rest[i:], '}')
			if j < 0 {
				return fmt.Errorf("words: step %s: a placeholder is not closed in %q", key, sw.Says)
			}
			ph := rest[i+1 : i+j]
			rest = rest[i+j+1:]
			name, format, has := strings.Cut(ph, "|")
			if strings.TrimSpace(name) == "" {
				return fmt.Errorf("words: step %s: a placeholder names nothing in %q", key, sw.Says)
			}
			if !has {
				continue
			}
			format, arg, _ := strings.Cut(format, ":")
			switch format {
			case "f", "m", "d", "month":
			case "value":
				if strings.TrimSpace(arg) == "" {
					return fmt.Errorf("words: step %s: {%s|value} names no field", key, name)
				}
			case "plural":
				if len(strings.Split(arg, ",")) < 2 {
					return fmt.Errorf("words: step %s: {%s|plural} needs the noun's forms, at least two", key, name)
				}
			default:
				return fmt.Errorf("words: step %s: format %q is none of %v", key, format, sayFormats)
			}
		}
	}
	return nil
}
