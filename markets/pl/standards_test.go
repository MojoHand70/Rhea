package pl

import (
	"strings"
	"testing"
)

const sampleStandardText = `MINISTRA ROZWOJU I FINANSÓW
Warszawa, dnia 29 maja 2017 r.
Poz. 105
K O M U N I K AT
w sprawie ogłoszenia uchwały Komitetu Standardów Rachunkowości
Dziennik Urzędowy Ministra Rozwoju i Finansów             –2–                                                 Poz. 105
I. Cel i zakres standardu
1.1.    Celem niniejszego Krajowego Standardu Rachunkowości (zwanego dalej standardem) jest
        wyjaśnienie zasad.
II. Definicje
3.1.    środek trwały – składnik aktywów, który:
        a) jest kompletny,
        b) jest zdatny do użytku.
VI. Wycena
6.18. Przedpłaty/zaliczki udzielone na poczet pozyskania środka trwałego – wyrażone w walucie obcej –
        przelicza się na walutę polską jedynie na moment ich przekazania.
Dziennik Urzędowy Ministra Rozwoju i Finansów   – 31 –                                     Poz. 105
        Ze względu na ich treść ekonomiczną nie dokonuje się ponownej wyceny.
Przykład 6.1. Spółka przekazała zaliczkę.
6.19. Kolejny punkt.
`

func TestParseStandardReadsPointsAndTheJournalItem(t *testing.T) {
	basis, pts := ParseStandard(sampleStandardText)
	if basis != "Dz. Urz. Min. Fin. z 2017 r. poz. 105" {
		t.Errorf("basis %q", basis)
	}
	if len(pts) != 4 || pts[0].N != "1.1" || pts[1].N != "3.1" || pts[2].N != "6.18" || pts[3].N != "6.19" {
		t.Fatalf("points: %+v", pts)
	}
	if !strings.Contains(pts[1].Text, "\na) jest kompletny,\nb) jest zdatny do użytku.") {
		t.Errorf("units on their own lines: %q", pts[1].Text)
	}
	p := pts[2].Text
	if !strings.HasPrefix(p, "6.18. Przedpłaty/zaliczki") || !strings.Contains(p, "nie dokonuje się ponownej wyceny.") || !strings.Contains(p, "Przykład 6.1.") || strings.Contains(p, "Dziennik") {
		t.Errorf("6.18: %q", p)
	}
}
