// Package statute is the library of acts the host reads from: consolidated
// texts split into articles, each act dated by the consolidation it was
// made from, found by the name a citation uses and the article's number.
// The library is data; a market fetches and parses its own acts (the Sejm's
// ELI service for Poland) and writes them here as JSON. The host consults
// it by citation — the article it would look up — never by similarity: an
// answer that rests on the statute names the provision and quotes it, and
// the version it quoted from is on the record.
package statute

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Act is one act's consolidated text.
type Act struct {
	Key   string   `json:"key"`   // short file key, e.g. "uor"
	Names []string `json:"names"` // the names citations use: "ustawa o rachunkowości", "UoR"
	Title string   `json:"title"`
	ELI   string   `json:"eli"` // "DU/1994/591"
	// Basis is the consolidation the text was made from, as the publisher
	// states it: "t.j. Dz. U. z 2026 r. poz. 522, 640, 644, 1206".
	Basis    string    `json:"basis"`
	Changed  string    `json:"changed"` // the publisher's last change stamp
	Fetched  string    `json:"fetched"` // the day the text was fetched
	Articles []Article `json:"articles"`
}

// Article is one article, whole: its units (ust., pkt, §) inside the text.
type Article struct {
	N    string `json:"n"` // "28", "28a"
	Text string `json:"text"`
}

// Library holds the acts of a directory.
type Library struct {
	acts   []Act
	byName map[string]*Act
	index  map[string]map[string]*Article
}

// Load reads every act file (*.json) of a directory.
func Load(dir string) (*Library, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	l := &Library{byName: map[string]*Act{}, index: map[string]map[string]*Article{}}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") || strings.HasSuffix(e.Name(), ".meta.json") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var a Act
		if err := json.Unmarshal(b, &a); err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		if a.Key == "" || len(a.Articles) == 0 {
			continue
		}
		l.Add(a)
	}
	return l, nil
}

// New makes an empty library; Add puts an act in it.
func New() *Library {
	return &Library{byName: map[string]*Act{}, index: map[string]map[string]*Article{}}
}

func (l *Library) Add(a Act) {
	l.acts = append(l.acts, a)
	act := &l.acts[len(l.acts)-1]
	// re-point every name at the new backing array
	l.byName = map[string]*Act{}
	l.index = map[string]map[string]*Article{}
	for i := range l.acts {
		p := &l.acts[i]
		for _, n := range append([]string{p.Key}, p.Names...) {
			l.byName[norm(n)] = p
		}
		idx := map[string]*Article{}
		for j := range p.Articles {
			idx[strings.ToLower(p.Articles[j].N)] = &p.Articles[j]
		}
		l.index[p.Key] = idx
	}
	_ = act
}

// Acts lists the acts, by key.
func (l *Library) Acts() []Act {
	out := make([]Act, len(l.acts))
	copy(out, l.acts)
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

// Act finds an act by any of its names (case and spacing do not matter;
// a trailing "(t.j. …)" is ignored).
func (l *Library) Act(name string) (*Act, bool) {
	n := norm(name)
	if a, ok := l.byName[n]; ok {
		return a, true
	}
	// the longest known name the citation starts with
	var best *Act
	bestLen := 0
	for known, a := range l.byName {
		if strings.HasPrefix(n, known) && len(known) > bestLen {
			best, bestLen = a, len(known)
		}
	}
	return best, best != nil
}

// Find returns an article of an act by the names a citation uses.
func (l *Library) Find(act, article string) (*Act, *Article, bool) {
	a, ok := l.Act(act)
	if !ok {
		return nil, nil, false
	}
	art, ok := l.index[a.Key][strings.ToLower(strings.TrimSpace(article))]
	if !ok {
		return a, nil, false
	}
	return a, art, true
}

// Citation is a provision as a citation names it.
type Citation struct {
	Act     string // the act's name as written
	Article string // "28", "106e"
	Unit    string // "ust. 1 pkt 5", "§ 2" — informational
}

var reCitation = regexp.MustCompile(`(?i)^(.*?)\s*\bart\.?\s*(\d+[a-z]*)\b\.?\s*(.*)$`)
var reParen = regexp.MustCompile(`\([^)]*\)`)

// ParseCitation reads "ustawa o rachunkowości (t.j. Dz.U. 2026 poz. 522)
// art. 30 ust. 2 pkt 2" into its act, article and unit. An act with no
// "art." in the text is not a citation of a provision.
func ParseCitation(s string) (Citation, bool) {
	s = strings.TrimSpace(s)
	m := reCitation.FindStringSubmatch(s)
	if m == nil {
		return Citation{}, false
	}
	act := strings.TrimSpace(reParen.ReplaceAllString(m[1], ""))
	act = strings.TrimRight(act, " ,;:–—-")
	return Citation{Act: act, Article: strings.ToLower(m[2]), Unit: strings.TrimSpace(strings.Split(m[3], ":")[0])}, act != ""
}

// Lookup resolves a citation string against the library.
func (l *Library) Lookup(citation string) (*Act, *Article, Citation, bool) {
	c, ok := ParseCitation(citation)
	if !ok {
		return nil, nil, c, false
	}
	a, art, ok := l.Find(c.Act, c.Article)
	return a, art, c, ok
}

func norm(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = reParen.ReplaceAllString(s, "")
	return strings.Join(strings.Fields(strings.NewReplacer(".", "", ",", "").Replace(s)), " ")
}

// Write saves an act as the library's JSON file.
func Write(dir string, a Act) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(a, "", " ")
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, a.Key+".json")
	return path, os.WriteFile(path, b, 0o644)
}
