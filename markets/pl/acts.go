package pl

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"rhea/internal/adapter"
	"rhea/internal/statute"
)

// The acts the Poland pack cites, from the Sejm's ELI service
// (api.sejm.gov.pl/eli): each act's consolidated text ("tekst
// ujednolicony", type U) as the Chancellery publishes it, with the
// consolidation it was made from. The names are the ones the pack's
// warrants and the host's citations use.
var polishActs = []statute.Act{
	{Key: "uor", ELI: "DU/1994/591", Names: []string{"ustawa o rachunkowości", "UoR", "u.o.r.", "ustawa z dnia 29 września 1994 r. o rachunkowości"}},
	{Key: "vat", ELI: "DU/2004/535", Names: []string{"ustawa o VAT", "ustawa o podatku od towarów i usług", "u.p.t.u.", "ustawa z dnia 11 marca 2004 r. o podatku od towarów i usług"}},
	{Key: "cit", ELI: "DU/1992/86", Names: []string{"ustawa o CIT", "ustawa o podatku dochodowym od osób prawnych", "u.p.d.o.p.", "updop"}},
	{Key: "pit", ELI: "DU/1991/350", Names: []string{"ustawa o PIT", "ustawa o podatku dochodowym od osób fizycznych", "u.p.d.o.f.", "updof"}},
	{Key: "ordynacja", ELI: "DU/1997/926", Names: []string{"Ordynacja podatkowa", "o.p.", "ustawa - Ordynacja podatkowa"}},
	{Key: "ksh", ELI: "DU/2000/1037", Names: []string{"Kodeks spółek handlowych", "KSH", "k.s.h."}},
	{Key: "kc", ELI: "DU/1964/93", Names: []string{"Kodeks cywilny", "KC", "k.c."}},
	{Key: "kp", ELI: "DU/1974/141", Names: []string{"Kodeks pracy", "KP", "k.p."}},
	{Key: "biegli", ELI: "DU/2017/1089", Names: []string{"ustawa o biegłych rewidentach", "ustawa o biegłych rewidentach, firmach audytorskich oraz nadzorze publicznym"}},
	{Key: "pcc", ELI: "DU/2000/959", Names: []string{"ustawa o PCC", "ustawa o podatku od czynności cywilnoprawnych"}},
}

const eliBase = "https://api.sejm.gov.pl/eli/acts/"

func init() {
	adapter.ProvideCommand(adapter.Command{
		Name:    "pl-acts",
		Summary: "the acts the Poland pack cites, from the Sejm's ELI service: fetch | list (rhea run pl-acts for help)",
		Run:     runPLActs,
	})
}

const plActsHelp = `rhea run pl-acts fetch [-dir data/acts/pl] [-only uor,vat]
rhea run pl-acts list  [-dir data/acts/pl]

  fetch   downloads each act's consolidated text (the Chancellery's "tekst ujednolicony") and
          its metadata from api.sejm.gov.pl/eli, reads the PDF (pdftotext) into articles and
          writes <key>.json to -dir; the consolidation the text was made from, the publisher's
          change stamp and the day of fetching are kept with it
  list    shows the acts in -dir with their basis and article counts`

func runPLActs(ctx context.Context, args []string, out io.Writer) error {
	if len(args) == 0 {
		fmt.Fprintln(out, plActsHelp)
		return nil
	}
	fs := flag.NewFlagSet("pl-acts "+args[0], flag.ContinueOnError)
	fs.SetOutput(out)
	dir := fs.String("dir", "data/acts/pl", "where the acts live")
	only := fs.String("only", "", "fetch only these keys, comma-separated")
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
		for _, a := range polishActs {
			if len(want) > 0 && !want[a.Key] {
				continue
			}
			act, err := fetchAct(ctx, a, *dir)
			if err != nil {
				fmt.Fprintf(out, "%-10s FAILED: %v\n", a.Key, err)
				continue
			}
			path, err := statute.Write(*dir, act)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "%-10s %s — %d articles, %s (%s)\n", act.Key, act.Title, len(act.Articles), act.Basis, filepath.Base(path))
		}
		return nil
	case "list":
		lib, err := statute.Load(*dir)
		if err != nil {
			return err
		}
		for _, a := range lib.Acts() {
			fmt.Fprintf(out, "%-10s %-70.70s %4d articles  %s  fetched %s\n", a.Key, a.Title, len(a.Articles), a.Basis, a.Fetched)
		}
		return nil
	}
	return fmt.Errorf("pl-acts: %q is neither fetch nor list", args[0])
}

// fetchAct downloads one act's metadata and consolidated text and reads it
// into articles.
func fetchAct(ctx context.Context, a statute.Act, dir string) (statute.Act, error) {
	meta, err := getJSON(ctx, eliBase+a.ELI)
	if err != nil {
		return a, err
	}
	a.Title, _ = meta["title"].(string)
	a.Changed, _ = meta["changeDate"].(string)
	var uFile string
	if texts, ok := meta["texts"].([]any); ok {
		for _, t := range texts {
			if m, ok := t.(map[string]any); ok && m["type"] == "U" {
				uFile, _ = m["fileName"].(string)
			}
		}
	}
	if uFile == "" {
		return a, errors.New("the ELI service offers no consolidated text (type U) for this act")
	}
	pdfPath := filepath.Join(dir, a.Key+".pdf")
	if err := download(ctx, eliBase+a.ELI+"/text/U/"+uFile, pdfPath); err != nil {
		return a, err
	}
	if b, err := json.MarshalIndent(meta, "", " "); err == nil {
		_ = os.WriteFile(filepath.Join(dir, a.Key+".meta.json"), b, 0o644)
	}
	text, err := exec.CommandContext(ctx, "pdftotext", "-layout", pdfPath, "-").Output()
	if err != nil {
		return a, fmt.Errorf("pdftotext %s: %w (poppler-utils needed)", pdfPath, err)
	}
	a.Basis, a.Articles = ParseAct(string(text))
	a.Fetched = time.Now().UTC().Format("2006-01-02")
	if len(a.Articles) == 0 {
		return a, errors.New("no articles read from the text")
	}
	return a, nil
}

var (
	reArt       = regexp.MustCompile(`^\s*Art\.\s+(\d+[a-z]*)\.(\s|$)`)
	reHeader    = regexp.MustCompile(`^\s*(©Kancelaria Sejmu|Dz\. U\. \d{4} Nr|\d{4}-\d{2}-\d{2}\s*$)`)
	reDivision  = regexp.MustCompile(`^\s*(DZIAŁ|Dział|Rozdział|Oddział|TYTUŁ|Tytuł|Księga|KSIĘGA)\s+[IVXLC\d]+`)
	reUnitStart = regexp.MustCompile(`^\s*(\d+[a-z]*\)|\d+[a-z]*\.|[a-z]\)|§\s*\d)`)
	reBlanks    = regexp.MustCompile(`[ \t]+`)
)

// ParseAct reads the Chancellery's consolidated text as pdftotext lays it
// out: the consolidation basis from the first page's margin, then every
// article from its "Art. N." to the next one or a division heading, the
// page headers dropped. Units (ust., pkt, §) stay inside the article's text,
// each on its own line.
func ParseAct(text string) (basis string, articles []statute.Article) {
	// The basis sits in the first page's right margin ("Opracowano na
	// podstawie: t.j. Dz. U. z 2026 r. poz. 522, …"), which the text layout
	// interleaves with the title: read the margin by its column.
	lines := strings.Split(text, "\n")
	head := lines[:min(60, len(lines))]
	col := -1
	var margin strings.Builder
	for _, ln := range head {
		if reArt.MatchString(ln) {
			break // the margin block sits in the title area, before art. 1
		}
		r := []rune(ln)
		if col < 0 {
			if i := strings.Index(ln, "Opracowano na"); i >= 0 {
				col = len([]rune(ln[:i]))
			}
			continue
		}
		if len(r) > col {
			margin.WriteString(string(r[col:]))
			margin.WriteString(" ")
		}
	}
	// the block reads "podstawie: t.j. Dz. U. z 2024 r. poz. 18, 96; z 2025 r. poz. 20." — kept whole
	if text := strings.Join(strings.Fields(margin.String()), " "); strings.Contains(text, "t.j.") {
		basis = strings.TrimRight(strings.TrimSpace(text[strings.Index(text, "t.j."):]), ".")
		basis = strings.ReplaceAll(basis, " ,", ",")
	}
	var cur *statute.Article
	var sb strings.Builder
	flush := func() {
		if cur != nil {
			cur.Text = strings.TrimSpace(sb.String())
			articles = append(articles, *cur)
		}
		cur = nil
		sb.Reset()
	}
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimRight(raw, " \t\f")
		if strings.HasPrefix(line, "\f") {
			line = strings.TrimLeft(line, "\f")
		}
		if reHeader.MatchString(line) {
			continue
		}
		if m := reArt.FindStringSubmatch(line); m != nil {
			flush()
			cur = &statute.Article{N: strings.ToLower(m[1])}
			sb.WriteString(strings.TrimSpace(line))
			continue
		}
		if cur == nil {
			continue
		}
		if reDivision.MatchString(line) {
			flush()
			continue
		}
		t := strings.TrimSpace(reBlanks.ReplaceAllString(line, " "))
		if t == "" {
			continue
		}
		if reUnitStart.MatchString(t) {
			sb.WriteString("\n")
		} else {
			sb.WriteString(" ")
		}
		sb.WriteString(t)
	}
	flush()
	return basis, articles
}

func getJSON(ctx context.Context, url string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	res, err := fetchClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", url, res.Status)
	}
	var m map[string]any
	return m, json.NewDecoder(res.Body).Decode(&m)
}

// fetchClient reads the Ministry's journal and the standards with a
// deadline; a transfer that stalls ends instead of hanging the command.
var fetchClient = &http.Client{Timeout: 3 * time.Minute}

// download writes a fetched file whole or not at all: a failed transfer
// leaves no truncated PDF for the next run to parse as an act.
func download(ctx context.Context, url, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	res, err := fetchClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", url, res.Status)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, res.Body); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}
