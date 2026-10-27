package shell_test

import (
	"strings"
	"testing"

	"rhea/internal/pack"
)

// TestTheCoresListsSpeakTheMarketsLanguage (KK's screenshot, 2026-10-27:
// Wszystkie firmy headed NAME, KIND, COUNTRY): the core and the common
// packs bring no words of their own — their lists and reports are labelled
// in English for any market — so a market's words name their columns where
// they name the field: a list's by its kind's field, a report's by the view
// and its column. A pack with words of its own keeps the labels it gave its
// lists: the Poland pack's "Nr inw." is shorter than its words' "Numer" on
// purpose.
func TestTheCoresListsSpeakTheMarketsLanguage(t *testing.T) {
	w := newCITWorld(t)
	ctx, s, x := w.ctx, w.s, w.x
	sums, err := pack.LoadCore(ctx, s, "init")
	if err != nil {
		t.Fatal(err)
	}
	for _, sum := range sums {
		if sum.Bundle == "" {
			continue
		}
		if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-09-01"); err != nil || len(errs) != 0 {
			t.Fatalf("approve %s: %v %v", sum.Bundle, err, errs)
		}
	}
	headers := func(view string) string {
		t.Helper()
		var v struct {
			Columns []struct {
				Label string `json:"label"`
			} `json:"columns"`
		}
		w.get("/api/views/"+view, &v)
		var out []string
		for _, c := range v.Columns {
			out = append(out, c.Label)
		}
		return strings.Join(out, " | ")
	}
	for view, want := range map[string]string{
		"company-list":           "Nazwa | Rola | Kraj",
		"journal":                "Zapis | Linia | Księga | Konto | Strona | Kwota | Waluta | Data",
		"account-list":           "Kod | Nazwa | Typ",
		"fx-rates":               "Data | Z waluty | Na walutę | Kurs",
		"purchase-invoice-list":  "Numer | Data faktury | Dostawca | Nabywca | Brutto | Waluta",
		"issued-documents":       "Nazwa | Rodzaj pliku | Parametry | Rozmiar | Skrót",
		"group-trial-balance":    "Konto | Nazwa konta | BO Wn | BO Ma | Obroty Wn | Obroty Ma | Narastająco Wn | Narastająco Ma | Saldo Wn | Saldo Ma",
		"intercompany-positions": "Dokument | Należność | Zobowiązanie | Różnica",
		"invoice-by-customer":    "Odbiorca | Liczba faktur | Razem",
		"trial-balance":          "Księga | Konto | Nazwa konta | BO Wn | BO Ma | Obroty Wn | Obroty Ma | Narastająco Wn | Narastająco Ma | Saldo Wn | Saldo Ma",
		// the Poland pack's own list keeps its own short labels
		"srodki-trwale": "Nr inw.",
	} {
		if got := headers(view); !strings.HasPrefix(got, want) {
			t.Errorf("%s: %s, want %s", view, got, want)
		}
	}
}
