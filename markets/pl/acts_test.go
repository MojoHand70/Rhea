package pl

import (
	"strings"
	"testing"
)

const sampleActText = `©Kancelaria Sejmu                                                                         s. 1/3

                                Dz. U. 1994 Nr 121 poz. 591

                                                                                                     Opracowano na
                                                                                                     podstawie:   t.j.
                                         U S T AWA                                                   Dz. U. z 2026 r.
                                                                                                     poz. 522, 640,
                                  z dnia 29 września 1994 r.                                         644, 1206.

                                      o rachunkowości1)

                                           Rozdział 1

                                        Przepisy ogólne

       Art. 1. Ustawa określa zasady rachunkowości oraz zasady wykonywania
działalności w zakresie usługowego prowadzenia ksiąg rachunkowych.

       Art. 2. 1. Przepisy ustawy stosuje się do:
1)     spółek handlowych,
2)     osób fizycznych.
2. Osoby fizyczne mogą stosować zasady rachunkowości.
©Kancelaria Sejmu                                                                         s. 2/3
2026-10-08

                                           Rozdział 2

       Art. 28a. Jednostka mikro wycenia aktywa inaczej.
`

func TestParseActReadsBasisAndArticles(t *testing.T) {
	basis, arts := ParseAct(sampleActText)
	if basis != "t.j. Dz. U. z 2026 r. poz. 522, 640, 644, 1206" {
		t.Errorf("basis %q", basis)
	}
	if len(arts) != 3 || arts[0].N != "1" || arts[1].N != "2" || arts[2].N != "28a" {
		t.Fatalf("articles: %+v", arts)
	}
	if !strings.HasPrefix(arts[0].Text, "Art. 1. Ustawa określa") || !strings.Contains(arts[0].Text, "prowadzenia ksiąg rachunkowych.") {
		t.Errorf("art. 1: %q", arts[0].Text)
	}
	a2 := arts[1].Text
	if !strings.Contains(a2, "\n1) spółek handlowych,\n2) osób fizycznych.\n2. Osoby fizyczne") || strings.Contains(a2, "Kancelaria") || strings.Contains(a2, "2026-10-08") || strings.Contains(a2, "Rozdział") {
		t.Errorf("art. 2: %q", a2)
	}
}
