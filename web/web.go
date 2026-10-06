// Package web embeds the shell's static assets.
package web

import "embed"

//go:embed index.html app.js app.css tokens.css logo.svg logo-dark.svg logo-lockup.svg logo-lockup-dark.svg favicon.svg
var FS embed.FS
