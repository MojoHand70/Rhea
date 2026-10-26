// Package packs carries the packs the kernel ships in its binary: the core
// and the common packs every installation runs (DIRECTION 2026-10-23, step 1).
// A market's packs are not here — they are data a server is given.
package packs

import "embed"

//go:embed core/core.json core/apps common/people.json common/periods.json common/documents.json common/apps
var FS embed.FS

// Core is what rhea init loads, in the order they are approved: Ludzie
// first (a task names a person), the founding vocabulary, then Okresy (its
// rule writes the period lock) and Dokumenty.
var Core = []string{"common/people.json", "core/core.json", "common/periods.json", "common/documents.json"}
