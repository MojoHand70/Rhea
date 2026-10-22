package store

import (
	"net/url"
	"regexp"
)

// RedactDSN is a database address fit for a log: its password masked, in
// the URL form and the key=value form alike — a failed connection wrote
// the password into the log (review 2026-10-21, G6).
func RedactDSN(dsn string) string {
	if u, err := url.Parse(dsn); err == nil && u.Scheme != "" && u.User != nil {
		if _, has := u.User.Password(); has {
			u.User = url.UserPassword(u.User.Username(), "xxxxx")
		}
		return u.String()
	}
	return passwordField.ReplaceAllString(dsn, "${1}xxxxx")
}

var passwordField = regexp.MustCompile(`(password=)\S+`)
