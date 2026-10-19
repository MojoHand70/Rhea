package statute

import "testing"

func library() *Library {
	l := New()
	l.Add(Act{Key: "uor", Names: []string{"ustawa o rachunkowości", "UoR"}, Basis: "t.j. Dz. U. z 2026 r. poz. 522",
		Articles: []Article{{N: "28", Text: "Art. 28. 1. Aktywa i pasywa wycenia się…"}, {N: "28a", Text: "Art. 28a. …"}}})
	l.Add(Act{Key: "vat", Names: []string{"ustawa o VAT", "ustawa o podatku od towarów i usług"},
		Articles: []Article{{N: "106e", Text: "Art. 106e. 1. Faktura powinna zawierać…"}}})
	return l
}

func TestFindByAnyNameAndArticle(t *testing.T) {
	l := library()
	for _, c := range []struct{ act, art string }{{"UoR", "28"}, {"ustawa o rachunkowości", "28a"}, {"Ustawa o Rachunkowości (t.j. Dz.U. 2026 poz. 522)", "28"}, {"ustawa o podatku od towarów i usług", "106e"}} {
		if _, art, ok := l.Find(c.act, c.art); !ok || art.N != c.art {
			t.Errorf("Find(%q, %q): %v %v", c.act, c.art, art, ok)
		}
	}
	if _, _, ok := l.Find("ustawa o CIT", "15"); ok {
		t.Error("an act the library lacks must not be found")
	}
	if a, art, ok := l.Find("UoR", "99"); !ok && (a == nil || art != nil) {
		t.Error("a missing article of a known act returns the act")
	}
}

func TestParseCitation(t *testing.T) {
	for in, want := range map[string]Citation{
		"ustawa o rachunkowości art. 30 ust. 2 pkt 2":                                 {"ustawa o rachunkowości", "30", "ust. 2 pkt 2"},
		"ustawa o rachunkowości (t.j. Dz.U. 2026 poz. 522) art. 39 ust. 1: „koszty…”": {"ustawa o rachunkowości", "39", "ust. 1"},
		"Ordynacja podatkowa art. 63 § 1":                                             {"Ordynacja podatkowa", "63", "§ 1"},
		"ustawa o VAT art. 106e ust. 11":                                              {"ustawa o VAT", "106e", "ust. 11"},
	} {
		got, ok := ParseCitation(in)
		if !ok || got != want {
			t.Errorf("ParseCitation(%q) = %+v %v, want %+v", in, got, ok, want)
		}
	}
	if _, ok := ParseCitation("wzorcowy plan kont"); ok {
		t.Error("no article, no citation")
	}
}

func TestLookupResolvesACitation(t *testing.T) {
	l := library()
	a, art, c, ok := l.Lookup("ustawa o rachunkowości art. 28 ust. 1 pkt 5")
	if !ok || a.Key != "uor" || art.N != "28" || c.Unit != "ust. 1 pkt 5" {
		t.Fatalf("%v %v %+v %v", a, art, c, ok)
	}
}
