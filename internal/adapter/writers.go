package adapter

import (
	"context"
	"fmt"
	"sort"
)

// A writer is the second seam beside the adapter (DIRECTION 2026-10-15,
// "Month end"): where an adapter talks to an outside system and returns
// events, a writer reads objects and returns a file — a statutory filing,
// a statement, an invoice's visualization — for a person to hand on. The
// kernel renders any provided writer by name and never knows one; a market
// provides its own (markets/<cc>), as it provides its channels. A file
// handed over is a fact through the documents pack's issue_document door,
// kept under its hash; the writer itself writes nothing.

// File is what a writer returns.
type File struct {
	Name      string
	MediaType string
	Bytes     []byte
}

// Writer renders a file from the projected objects and the parameters a
// person named (a period, a book). It is deterministic in the objects:
// the same state and parameters give the same bytes.
type Writer func(ctx context.Context, reads Reads, params map[string]string) (File, error)

var writers = map[string]Writer{}

// ProvideWriter registers a writer by name; markets call it at init.
func ProvideWriter(name string, w Writer) { writers[name] = w }

// LookupWriter returns a provided writer by name.
func LookupWriter(name string) (Writer, bool) {
	w, ok := writers[name]
	return w, ok
}

// WriterNames lists what the binary can render, sorted.
func WriterNames() []string {
	out := make([]string, 0, len(writers))
	for n := range writers {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// Render runs a writer by name.
func Render(ctx context.Context, reads Reads, name string, params map[string]string) (File, error) {
	w, ok := writers[name]
	if !ok {
		return File{}, fmt.Errorf("no writer named %q — the binary renders: %v", name, WriterNames())
	}
	return w(ctx, reads, params)
}
