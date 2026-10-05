// Package web embeds the shell's static assets.
package web

import "embed"

//go:embed index.html app.js app.css tokens.css
var FS embed.FS
