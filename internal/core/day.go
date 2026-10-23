package core

import (
	"os"
	"sync"
	"time"
)

// The books' day (review 2026-10-21, E4): a fact dated "today" takes the
// date where the books are kept, not the date in UTC — a document received
// in Warsaw at half past midnight on the 1st was dated the last day of the
// month before, and fell into the VAT period before. The installation names
// its zone (RHEA_TIMEZONE, an IANA name such as Europe/Warsaw); one naming
// none keeps UTC. The clock opens one day for every company at once, so the
// zone is the installation's.

// BooksDay is an instant's date where the books are kept.
func BooksDay(t time.Time) string {
	loc, err := BooksZone()
	if err != nil {
		loc = time.UTC
	}
	return t.In(loc).Format("2006-01-02")
}

// Today is the books' day now, by the wall clock — the reading for code
// that is handed no clock of its own.
func Today() string { return BooksDay(time.Now()) }

// BooksZone is the installation's time zone; an error names a zone the
// time zone database does not know (rhea serve refuses to start on it).
func BooksZone() (*time.Location, error) {
	name := os.Getenv("RHEA_TIMEZONE")
	if name == "" {
		return time.UTC, nil
	}
	zones.Lock()
	defer zones.Unlock()
	if loc, ok := zones.m[name]; ok {
		return loc, nil
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, err
	}
	zones.m[name] = loc
	return loc, nil
}

var zones = struct {
	sync.Mutex
	m map[string]*time.Location
}{m: map[string]*time.Location{}}
