package shell

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The host's law, applied to the recorded morning (DECISIONS 2026-10-15):
// the kernel refuses a number the facts do not carry, and it refuses
// silence about a fact they do. The morning is the Poland story's — both
// VAT registers with amounts, both payment grids with one bank operation
// of 100.00 waiting on each — and the words are the shapes the model
// measurement produced: the small model that denied the waiting payment
// is now refused, the one that named every amount is accepted, and an
// empty grid may be called empty.
func TestHostLaw(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "testdata", "eval", "briefing.json"))
	if err != nil {
		t.Fatal(err)
	}
	var morning struct {
		Facts json.RawMessage `json:"facts"`
	}
	if err := json.Unmarshal(raw, &morning); err != nil {
		t.Fatal(err)
	}
	verdict := func(words string, facts []byte) string {
		if foreign := foreignNumbers(words, facts); len(foreign) > 0 {
			return "foreign: " + strings.Join(foreign, ", ")
		}
		if unspoken := unspokenFacts(words, facts); len(unspoken) > 0 {
			return "unspoken: " + strings.Join(unspoken, "; ")
		}
		return "accepted"
	}

	// The denial: registers named, the waiting payment called absent.
	denied := "Sprzedaż we wrześniu: 1230,00 zł brutto (VAT 230,00 zł); zakupy 615,00 zł. Brak wpłat i wypłat czekających na rozliczenie."
	if v := verdict(denied, morning.Facts); v != "unspoken: wpłaty bez rozliczenia (100); wypłaty bez rozliczenia (100)" {
		t.Fatalf("a denied fact is refused by name: %s", v)
	}
	// The truth: every fact carries one of its amounts in the words.
	truthful := "Zaksięgowana sprzedaż 1230,00 zł brutto, zakupy 615,00 zł. Jeden przelew na 100,00 zł z 27 września czeka na przypisanie do faktury — do decyzji."
	if v := verdict(truthful, morning.Facts); v != "accepted" {
		t.Fatalf("the truthful morning is accepted: %s", v)
	}
	// Whole złoty for a whole amount is the same value, not a number the
	// facts do not carry: "1230 zł" stands for 1230.00. Decimals that were
	// there and are dropped are another matter — 1234 is not 1234.56.
	if v := verdict(strings.ReplaceAll(truthful, ",00", ""), morning.Facts); v != "accepted" {
		t.Fatalf("whole złoty for a whole amount is the amount: %s", v)
	}
	cents := []byte(`[{"as":"fee","kind":"analysis","data":{"columns":[{"type":"money"}],"rows":[[{"v":"1234.56"}]]}}]`)
	if v := verdict("The fee came to 1234 zł.", cents); v != "foreign: 1234" {
		t.Fatalf("decimals that were there and are dropped are a rounding: %s", v)
	}
	if v := verdict("The fee came to 1 234,56 zł.", cents); v != "accepted" {
		t.Fatalf("the Polish spelling of the same value: %s", v)
	}
	// Silence about a register is a denial too, even with the payment named.
	if v := verdict("Jeden przelew na 100,00 zł czeka na rozliczenie.", morning.Facts); v != "unspoken: sprzedaż tego miesiąca (rejestr VAT) (1000, 230, 1230); zakupy tego miesiąca (rejestr VAT) (500, 115, 615)" {
		t.Fatalf("a register passed over is named with its amounts, the typed columns only: %s", v)
	}

	// An empty grid asks nothing: strip the waiting payment, and "nothing
	// waits" is the truth.
	var facts []map[string]any
	if err := json.Unmarshal(morning.Facts, &facts); err != nil {
		t.Fatal(err)
	}
	for _, f := range facts {
		if f["kind"] != "matching" {
			continue
		}
		for _, side := range f["data"].(map[string]any)["sides"].([]any) {
			side.(map[string]any)["items"] = []any{}
		}
	}
	empty, _ := json.Marshal(facts)
	if v := verdict(denied, empty); v != "accepted" {
		t.Fatalf("an empty grid may be called empty: %s", v)
	}

	// A register that is all zeros asks nothing either; an untyped table
	// counts every number its cells carry.
	zeros := []byte(`[{"as":"sales","kind":"analysis","data":{"columns":[{"type":"string"},{"type":"money"}],"rows":[[{"v":"2026-09"},{"v":"0.00"}]]}}]`)
	if v := verdict("No sales this month.", zeros); v != "accepted" {
		t.Fatalf("zeros ask nothing: %s", v)
	}
	untyped := []byte(`[{"as":"stock","kind":"analysis","data":{"columns":[{"label":"Item"},{"label":"Qty"}],"rows":[[{"v":"bolt"},{"v":"150"}]]}}]`)
	if v := verdict("Bolts are in stock.", untyped); v != "unspoken: stock (150)" {
		t.Fatalf("an untyped table counts every number: %s", v)
	}
	if v := verdict("150 bolts are in stock.", untyped); v != "accepted" {
		t.Fatalf("the untyped number spoken: %s", v)
	}
}

// The third law (KK's Helios morning, 2026-10-18): she named "HOS/2/2026 z
// 14 stycznia", an invoice in her earlier words but not in today's facts —
// every number in it passed. A document's name must stand in the facts as
// written; amounts with units, dates and article numbers are not names.
func TestHostNamesOnlyWhatTheFactsCarry(t *testing.T) {
	facts := []byte(`[{"as":"terminy","kind":"analysis","data":{"rows":[[{"v":"2026-10-26"},{"v":"VAT i JPK_V7M"},{"v":"2026-09"},{"v":"9220.00"}]]}},
		{"as":"faktury","kind":"analysis","data":{"rows":[[{"v":"FV 7/09/2026"},{"v":"Tech Park Gdańsk S.A."}]]}}]`)
	for words, want := range map[string]string{
		"Najstarsza faktura zakupu HOS/2/2026 z 14 stycznia czeka.":                   "HOS/2/2026",
		"Faktura FV1 nadal otwarta.":                                                  "FV1",
		"VAT i JPK_V7M: 9220.00 zł do 26 października; faktura 7/09/2026 Tech Parku.": "",
		"jpk_v7m złożony; 100.00zł bez rozliczenia, art. 19 ust. 1e, 20/10/2026.":     "",
	} {
		if got := strings.Join(foreignNames(words, facts), ","); got != want {
			t.Fatalf("%q: foreign names %q, want %q", words, got, want)
		}
	}
}
