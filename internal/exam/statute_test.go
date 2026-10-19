package exam

import (
	"context"
	"strings"
	"testing"

	"rhea/internal/agent"
	"rhea/internal/statute"
)

// A scripted model: the consult call names provisions, the answer call
// answers as told; what it was handed is checked by the test.
func scriptedHost(t *testing.T, consult, answer string) (*Host, *string) {
	lib := statute.New()
	lib.Add(statute.Act{Key: "uor", Names: []string{"ustawa o rachunkowości", "UoR"}, Basis: "t.j. Dz. U. z 2026 r. poz. 522",
		Articles: []statute.Article{{N: "28", Text: "Art. 28. 1. Aktywa i pasywa wycenia się nie rzadziej niż na dzień bilansowy…"}}})
	var handedText string
	a := &agent.Agent{Complete: func(_ context.Context, system, user string) (string, error) {
		if strings.Contains(system, "Name the provisions you would open") {
			if !strings.Contains(user, "ustawa o rachunkowości (t.j. Dz. U. z 2026 r. poz. 522)") {
				t.Errorf("the consult call must list the acts at hand: %q", user)
			}
			return consult, nil
		}
		handedText = user
		return answer, nil
	}}
	h := NewHost(a, nil, nil)
	h.Library = lib
	return h, &handedText
}

func TestHostAnswersFromAProvisionItWasHanded(t *testing.T) {
	h, handed := scriptedHost(t,
		`{"consult": [{"act": "UoR", "article": "28", "unit": "ust. 1"}]}`,
		`{"answer": "a", "source": "statute", "citation": "ustawa o rachunkowości art. 28 ust. 1: „wycenia się nie rzadziej niż na dzień bilansowy”", "note": ""}`)
	ans, err := h.Answer(context.Background(), Question{N: 1, Stem: "Kiedy wycenia się aktywa?", Options: []Option{{"a", "na dzień bilansowy"}, {"b", "nigdy"}, {"c", "co tydzień"}}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(*handed, "The statute in front of you") || !strings.Contains(*handed, "[ustawa o rachunkowości, art. 28 — t.j. Dz. U. z 2026 r. poz. 522]") || !strings.Contains(*handed, "Art. 28. 1. Aktywa") {
		t.Errorf("the answer call was not handed the provision: %q", *handed)
	}
	if ans.Source != SourceStatute || ans.Letter != "a" || !strings.Contains(string(ans.Consulted), `"article":"28"`) {
		t.Errorf("%+v", ans)
	}
}

func TestAProvisionNotHandedIsMemory(t *testing.T) {
	h, _ := scriptedHost(t,
		`{"consult": [{"act": "UoR", "article": "28"}]}`,
		`{"answer": "b", "source": "statute", "citation": "ustawa o rachunkowości art. 35 ust. 3", "note": ""}`)
	ans, err := h.Answer(context.Background(), Question{N: 1, Stem: "?", Options: []Option{{"a", "x"}, {"b", "y"}, {"c", "z"}}})
	if err != nil {
		t.Fatal(err)
	}
	if ans.Source != SourceMemory || !strings.Contains(ans.Note, "not handed") {
		t.Errorf("%+v", ans)
	}
}

func TestAnUnknownActHandsNothing(t *testing.T) {
	h, handed := scriptedHost(t,
		`{"consult": [{"act": "ustawa o CIT", "article": "15"}]}`,
		`{"answer": "c", "source": "statute", "citation": "ustawa o CIT art. 15 ust. 1", "note": ""}`)
	ans, err := h.Answer(context.Background(), Question{N: 1, Stem: "?", Options: []Option{{"a", "x"}, {"b", "y"}, {"c", "z"}}})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(*handed, "The statute in front of you") {
		t.Errorf("nothing should be handed for an act the library lacks: %q", *handed)
	}
	if ans.Source != SourceMemory {
		t.Errorf("%+v", ans)
	}
}
