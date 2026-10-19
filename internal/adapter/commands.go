package adapter

import (
	"context"
	"io"
	"sort"
)

// A Command is a market's tool a person runs by name (rhea run NAME …): work
// that talks to a statutory system on the owner's behalf but writes nothing
// to the log — asking for a certificate, writing a key the owner keeps. The
// market registers it from its own package at init; the kernel and the CLI
// know none by name (invariant 8).
type Command struct {
	Name    string
	Summary string
	Run     func(ctx context.Context, args []string, out io.Writer) error
}

var commands = map[string]Command{}

// ProvideCommand registers a market's command.
func ProvideCommand(c Command) { commands[c.Name] = c }

// LookupCommand returns a provided command by its name.
func LookupCommand(name string) (Command, bool) {
	c, ok := commands[name]
	return c, ok
}

// Commands lists the provided commands by name.
func Commands() []Command {
	out := make([]Command, 0, len(commands))
	for _, c := range commands {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
