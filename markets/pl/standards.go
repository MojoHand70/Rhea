package pl

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"rhea/internal/adapter"
	"rhea/internal/statute"
)

// The Krajowe Standardy Rachunkowości as the Ministry publishes them
// (gov.pl/web/finanse/krajowe-standardy-rachunkowosci): each standard a
// communiqué in the Ministry's journal with the standard as its annex,
// numbered in points ("6.18."), which is how the profession and the exam
// cite them. Where the Ministry lists two versions, the updated one.
var polishStandards = []struct {
	n, title, id string
}{
	{"1", "Rachunek przepływów pieniężnych", "c21830fc-9205-4f33-8ddd-a0cdd5d30099"},
	{"2", "Podatek dochodowy", "ab289f43-2138-4903-a88d-41db3342c656"},
	{"3", "Niezakończone usługi budowlane", "8b0eaa9b-1295-4751-99d2-5611fe36dc56"},
	{"4", "Utrata wartości aktywów", "4c74806b-00ac-4a6b-a00c-27c5f6cc9646"},
	{"5", "Leasing, najem i dzierżawa", "07da18b3-1ae8-4fde-9239-7e264ed39e24"},
	{"6", "Rezerwy, bierne rozliczenia międzyokresowe kosztów, zobowiązania warunkowe", "bc48b7e4-7102-4458-bdd9-545997b2e38b"},
	{"7", "Zmiany zasad (polityki) rachunkowości, wartości szacunkowych, poprawianie błędów, zdarzenia następujące po dniu bilansowym", "9da792cf-ca83-4ace-9779-17aa810a8548"},
	{"8", "Działalność deweloperska", "585caf0f-71da-4bd0-9601-90179ba93884"},
	{"9", "Sprawozdanie z działalności", "0b95026a-9043-4682-9067-77ecd9f985f8"},
	{"10", "Umowy o partnerstwie publiczno-prywatnym oraz umowy koncesji na roboty budowlane lub usługi", "33853a90-f477-46db-b54e-6a69d9fa87aa"},
	{"11", "Środki trwałe", "92b030f0-544b-4114-b877-f050ea186ebf"},
	{"12", "Działalność rolnicza", "fce99e23-d840-4108-a855-eab81b8f05ee"},
	{"13", "Koszt wytworzenia jako podstawa wyceny produktów", "5761f089-bda5-4a79-9d16-f27b204ce989"},
	{"14", "Kontynuacja działalności oraz rachunkowość jednostek przy braku kontynuowania działalności", "69f68d95-61f2-4bc2-a72e-7632a076d655"},
	{"15", "Przychody ze sprzedaży wyrobów, półproduktów, towarów i materiałów", "143e3deb-6668-4506-b191-b0efcce109f7"},
	{"16", "Uproszczenia dopuszczone ustawą o rachunkowości", "feba9252-201d-4357-93b3-b3dbeae1c59a"},
}

func init() {
	adapter.ProvideCommand(adapter.Command{
		Name:    "pl-standards",
		Summary: "the Krajowe Standardy Rachunkowości from the Ministry: fetch | list (rhea run pl-standards for help)",
		Run:     runPLStandards,
	})
}

const plStandardsHelp = `rhea run pl-standards fetch [-dir data/standards/pl] [-only 6,11]
rhea run pl-standards list  [-dir data/standards/pl]

  fetch   downloads each KSR from gov.pl, reads the PDF (pdftotext) into its points and writes
          ksr<N>.json to -dir, with the journal item it was published as and the day of fetching
  list    shows the standards in -dir with their points`

func runPLStandards(ctx context.Context, args []string, out io.Writer) error {
	if len(args) == 0 {
		fmt.Fprintln(out, plStandardsHelp)
		return nil
	}
	fs := flag.NewFlagSet("pl-standards "+args[0], flag.ContinueOnError)
	fs.SetOutput(out)
	dir := fs.String("dir", "data/standards/pl", "where the standards live")
	only := fs.String("only", "", "fetch only these numbers, comma-separated")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	switch args[0] {
	case "fetch":
		want := map[string]bool{}
		for _, k := range strings.Split(*only, ",") {
			if k = strings.TrimSpace(k); k != "" {
				want[k] = true
			}
		}
		for _, st := range polishStandards {
			if len(want) > 0 && !want[st.n] {
				continue
			}
			act, err := fetchStandard(ctx, st.n, st.title, st.id, *dir)
			if err != nil {
				fmt.Fprintf(out, "KSR %-3s FAILED: %v\n", st.n, err)
				continue
			}
			path, err := statute.Write(*dir, act)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "KSR %-3s %-60.60s — %d points, %s (%s)\n", st.n, st.title, len(act.Articles), act.Basis, filepath.Base(path))
		}
		return nil
	case "list":
		lib, err := statute.Load(*dir)
		if err != nil {
			return err
		}
		for _, a := range lib.Acts() {
			fmt.Fprintf(out, "%-6s %-70.70s %4d points  %s  fetched %s\n", a.Key, a.Title, len(a.Articles), a.Basis, a.Fetched)
		}
		return nil
	}
	return fmt.Errorf("pl-standards: %q is neither fetch nor list", args[0])
}

func fetchStandard(ctx context.Context, n, title, id, dir string) (statute.Act, error) {
	a := statute.Act{Key: "ksr" + n, Title: "Krajowy Standard Rachunkowości nr " + n + " „" + title + "”",
		ELI:   "gov.pl/attachment/" + id,
		Names: []string{"KSR " + n, "KSR nr " + n, "Krajowy Standard Rachunkowości nr " + n, "KSR " + n + " „" + title + "”", "KSR nr " + n + " „" + title + "”"}}
	pdfPath := filepath.Join(dir, a.Key+".pdf")
	if err := download(ctx, "https://www.gov.pl/attachment/"+id, pdfPath); err != nil {
		return a, err
	}
	text, err := exec.CommandContext(ctx, "pdftotext", "-layout", pdfPath, "-").Output()
	if err != nil {
		return a, fmt.Errorf("pdftotext %s: %w (poppler-utils needed)", pdfPath, err)
	}
	a.Basis, a.Articles = ParseStandard(string(text))
	a.Fetched = time.Now().UTC().Format("2006-01-02")
	if len(a.Articles) == 0 {
		return a, fmt.Errorf("no points read from the text")
	}
	return a, nil
}

var (
	rePoint       = regexp.MustCompile(`^\s*(\d{1,2}\.\d{1,3}[a-z]?)(?:\.\s+|\s{2,})(\S.*)$`) // "6.18. …" or, as KSR 15 has it, "1.1      …"
	reStdHeader   = regexp.MustCompile(`^\s*(Dziennik Urzędowy|–\s*\d+\s*–\s*$|.*\bPoz\.\s*\d+\s*$)`)
	reStdChapter  = regexp.MustCompile(`^\s*[IVX]+\.\s+\S`)
	reStdPoz      = regexp.MustCompile(`Poz\.\s*(\d+)`)
	reStdYear     = regexp.MustCompile(`\b(20\d\d)\b`)
	reStdUnit     = regexp.MustCompile(`^\s*(\d+[a-z]?\)|[a-z]\)|[–-]\s)`)
	reStdBlanks   = regexp.MustCompile(`[ \t]+`)
	reStdPrelude  = regexp.MustCompile(`^\s*1\.1(?:\.\s|\s{2,})`)
	reStdExampleH = regexp.MustCompile(`^\s*Przykład\b`)
)

// ParseStandard reads a KSR as pdftotext lays the Ministry's journal out:
// the item it was published as from the prelude, then every point from
// its "N.N." to the next point or a chapter heading, the page headers
// dropped. Examples stay inside the point they illustrate.
func ParseStandard(text string) (basis string, points []statute.Article) {
	lines := strings.Split(text, "\n")
	// the prelude, up to the first point, names the journal item
	var prelude strings.Builder
	for _, ln := range lines {
		if reStdPrelude.MatchString(ln) {
			break
		}
		prelude.WriteString(ln)
		prelude.WriteString(" ")
		if prelude.Len() > 4000 {
			break
		}
	}
	if m := reStdPoz.FindStringSubmatch(prelude.String()); m != nil {
		year := ""
		if y := reStdYear.FindString(prelude.String()); y != "" {
			year = " z " + y + " r."
		}
		basis = "Dz. Urz. Min. Fin." + year + " poz. " + m[1]
	}
	var cur *statute.Article
	var sb strings.Builder
	flush := func() {
		if cur != nil {
			cur.Text = strings.TrimSpace(sb.String())
			points = append(points, *cur)
		}
		cur = nil
		sb.Reset()
	}
	started := false
	for _, raw := range lines {
		line := strings.TrimRight(strings.TrimLeft(raw, "\f"), " \t")
		if reStdHeader.MatchString(line) {
			continue
		}
		if m := rePoint.FindStringSubmatch(line); m != nil {
			started = true
			flush()
			cur = &statute.Article{N: m[1]}
			sb.WriteString(m[1] + ". " + strings.TrimSpace(reStdBlanks.ReplaceAllString(m[2], " ")))
			continue
		}
		if !started || cur == nil {
			continue
		}
		if reStdChapter.MatchString(line) && !reStdExampleH.MatchString(line) {
			flush()
			continue
		}
		t := strings.TrimSpace(reStdBlanks.ReplaceAllString(line, " "))
		if t == "" {
			continue
		}
		if reStdUnit.MatchString(t) {
			sb.WriteString("\n")
		} else {
			sb.WriteString(" ")
		}
		sb.WriteString(t)
	}
	flush()
	return basis, points
}
