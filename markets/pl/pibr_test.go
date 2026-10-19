package pl

import (
	"strings"
	"testing"
)

// A paper in the Commission's layout as pdftohtml -xml renders it:
// letter-spaced headings, the question number and its stem as separate
// elements on one baseline, the right option in bold — on the letter's
// element or on the whole line — and a task with its model solution.
const samplePaperXML = `<?xml version="1.0" encoding="UTF-8"?>
<pdf2xml producer="poppler" version="26.01.0">
<page number="1" position="absolute" top="0" left="0" height="1263" width="893">
<text top="35" left="85" width="200" height="18" font="0"><b>R a c h u n k o w o ś ć   f i n a n s o w a</b></text>
<text top="35" left="710" width="80" height="18" font="0"><b>R o k   2 0 2 5</b></text>
<text top="65" left="253" width="300" height="18" font="0"><b>R O Z W I Ą Z A N I A   W Z O R C O W E</b></text>
<text top="102" left="85" width="200" height="18" font="0"><b>B L O K   T E M A T Y C Z N Y   I</b></text>
<text top="123" left="85" width="200" height="18" font="0"><b>P y t a n i a   t e s t o w e</b></text>
<text top="157" left="85" width="10" height="18" font="1">1.</text>
<text top="157" left="138" width="500" height="18" font="1">Pierwsze pytanie brzmi, czy ustawa</text>
<text top="176" left="85" width="500" height="18" font="1">o rachunkowości pozwala:</text>
<text top="214" left="85" width="10" height="18" font="1"><b>a)</b></text>
<text top="214" left="139" width="500" height="18" font="1"><b>odpowiedź pierwsza, w dwóch</b></text>
<text top="233" left="139" width="500" height="18" font="1"><b>wierszach.</b></text>
<text top="271" left="85" width="10" height="18" font="1">b)</text>
<text top="271" left="139" width="500" height="18" font="1">odpowiedź druga.</text>
<text top="309" left="85" width="10" height="18" font="1">c)</text>
<text top="309" left="139" width="500" height="18" font="1">odpowiedź trzecia.</text>
<text top="340" left="85" width="10" height="18" font="1">2.</text>
<text top="340" left="138" width="500" height="18" font="1">Drugie pytanie:</text>
<text top="380" left="85" width="500" height="18" font="1">a) tak,</text>
<text top="400" left="85" width="500" height="18" font="1">b) nie,</text>
<text top="420" left="85" width="500" height="18" font="1"><b>c) to zależy.</b></text>
<text top="1194" left="776" width="40" height="18" font="1">str. 1</text>
<text top="1214" left="350" width="200" height="18" font="1">K o m i s j a   E g z a m i n a c y j n a</text>
</page>
<page number="2" position="absolute" top="0" left="0" height="1263" width="893">
<text top="35" left="85" width="200" height="18" font="0"><b>R a c h u n k o w o ś ć   f i n a n s o w a</b></text>
<text top="35" left="710" width="80" height="18" font="0"><b>R o k   2 0 2 5</b></text>
<text top="64" left="85" width="200" height="18" font="0"><b>Z a d a n i a   s y t u a c y j n e</b></text>
<text top="99" left="85" width="200" height="18" font="0"><b>Z a d a n i e   1</b></text>
<text top="131" left="85" width="200" height="18" font="1">Założenia:</text>
<text top="150" left="85" width="10" height="18" font="1">1.</text>
<text top="150" left="112" width="500" height="18" font="1">Spółka X jest mała.</text>
<text top="200" left="323" width="300" height="18" font="0"><b>R o z w i ą z a n i e   w z o r c o w e :</b></text>
<text top="220" left="64" width="40" height="18" font="0"><b>Ad 1</b></text>
<text top="240" left="64" width="300" height="18" font="1">Dt 751 Ct 145 10.000</text>
<text top="300" left="85" width="200" height="18" font="0"><b>B L O K   T E M A T Y C Z N Y   I I</b></text>
<text top="320" left="85" width="10" height="18" font="1">1.</text>
<text top="320" left="139" width="500" height="18" font="1">Pytanie w bloku drugim:</text>
<text top="340" left="85" width="500" height="18" font="1">a) x,</text>
<text top="360" left="85" width="500" height="18" font="1"><b>b) y,</b></text>
<text top="380" left="85" width="500" height="18" font="1">c) z.</text>
<text top="1194" left="776" width="40" height="18" font="1">str. 2</text>
</page>
</pdf2xml>`

func TestParsePIBRPaper(t *testing.T) {
	p, err := ParsePIBRPaper([]byte(samplePaperXML), "sample.xml")
	if err != nil {
		t.Fatal(err)
	}
	if p.Subject != "Rachunkowość finansowa" || p.Year != "2025" {
		t.Errorf("header: %q %q", p.Subject, p.Year)
	}
	if len(p.Questions) != 3 {
		t.Fatalf("questions: %d (%+v)", len(p.Questions), p.Questions)
	}
	q := p.Questions[0]
	if q.Block != "I" || q.N != 1 || q.Key != "a" || q.Stem != "Pierwsze pytanie brzmi, czy ustawa o rachunkowości pozwala:" {
		t.Errorf("q1: %+v", q)
	}
	if len(q.Options) != 3 || q.Options[0].Text != "odpowiedź pierwsza, w dwóch wierszach." || q.Options[2].Letter != "c" {
		t.Errorf("q1 options: %+v", q.Options)
	}
	if q := p.Questions[1]; q.N != 2 || q.Key != "c" || q.Options[2].Text != "to zależy." {
		t.Errorf("q2: %+v", q)
	}
	if q := p.Questions[2]; q.Block != "II" || q.N != 1 || q.Key != "b" {
		t.Errorf("block II: %+v", q)
	}
	if len(p.Tasks) != 1 {
		t.Fatalf("tasks: %+v", p.Tasks)
	}
	task := p.Tasks[0]
	if task.Block != "I" || task.N != 1 || !strings.Contains(task.Text, "1. Spółka X jest mała.") || strings.Contains(task.Text, "Dt 751") {
		t.Errorf("task text: %+v", task)
	}
	if !strings.Contains(task.Solution, "Dt 751 Ct 145 10.000") || !strings.HasPrefix(task.Solution, "Ad 1") {
		t.Errorf("task solution: %q", task.Solution)
	}
	rendered := p.Questions[1].Prompt()
	if !strings.Contains(rendered, "2. Drugie pytanie:\na) tak,\nb) nie,\nc) to zależy.\n") {
		t.Errorf("prompt: %q", rendered)
	}
}

func TestUnspaceCollapsesLetterSpacedHeadingsOnly(t *testing.T) {
	for in, want := range map[string]string{
		"B L O K   T E M A T Y C Z N Y   I I":                          "BLOK TEMATYCZNY II",
		"Z a d a n i e   1":                                            "Zadanie 1",
		"R a c h u n k o w o ś ć   f i n a n s o w a  R o k   2 0 2 5": "Rachunkowość finansowa Rok 2025",
		"1. W przeddzień przekazania jednostki":                        "1. W przeddzień przekazania jednostki",
		"a) tak,":                                                      "a) tak,",
	} {
		if got := unspace(in); got != want {
			t.Errorf("unspace(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPIBRScoringIsTheRegulations(t *testing.T) {
	if PIBRScoring.Correct != 2 || PIBRScoring.Wrong != -1 || PIBRScoring.Blank != 0 || PIBRScoring.PassShare != 0.6 {
		t.Errorf("%+v", PIBRScoring)
	}
}
