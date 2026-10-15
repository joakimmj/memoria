package cli

import (
	"io"
	"strings"
)

// Command is one entry of the command table (§17.2).
type Command struct {
	Name    string                            `json:"name"`
	Summary string                            `json:"summary"`
	Args    []Arg                             `json:"args,omitempty"`
	Flags   []Flag                            `json:"flags,omitempty"`
	Text    bool                              `json:"text,omitempty"` // prints text to Stdout, not JSON
	Run     func(inv Invocation) (any, error) `json:"-"`
}

// Arg is a positional argument.
type Arg struct {
	Name     string `json:"name"`
	Summary  string `json:"summary"`
	Optional bool   `json:"optional,omitempty"`
}

// Flag is a --name flag; boolean when Value is empty.
type Flag struct {
	Name     string `json:"name"`
	Value    string `json:"value,omitempty"`
	Summary  string `json:"summary"`
	Required bool   `json:"required,omitempty"`
}

// Invocation is one parsed command line.
type Invocation struct {
	Args   []string          // positionals given
	Flags  map[string]string // flags given; booleans are "true"
	Stdin  io.Reader
	Stdout io.Writer // Text commands only
}

// table is the ordered command list (§22). Filled in init: Run funcs read
// it, and a self-referencing initializer is an initialization cycle.
var table []Command

func init() {
	table = []Command{
		{
			Name:    "version",
			Summary: "the version and commit mia was built from",
			Run:     runVersion,
		},
		{
			Name:    "help",
			Summary: "a manual page rendered as text",
			Args:    []Arg{{Name: "topic", Summary: "a command or a page name; mia(1) when omitted", Optional: true}},
			Text:    true,
			Run:     runHelp,
		},
		{
			Name:    "man",
			Summary: "the manual pages written as roff",
			Flags:   []Flag{{Name: "dir", Value: "dir", Summary: "directory to write man1/, man5/ and man7/ into; created if missing", Required: true}},
			Run:     runMan,
		},
		{
			Name:    "commands",
			Summary: "every command with its arguments, as a schema",
			Run:     runCommands,
		},
	}
}

// lookup returns the command named name.
func lookup(name string) (Command, bool) {
	for _, c := range table {
		if c.Name == name {
			return c, true
		}
	}
	return Command{}, false
}

// synopsis is the usage line, e.g. "mia man --dir <dir>".
func (c Command) synopsis() string {
	parts := []string{"mia", c.Name}
	for _, a := range c.Args {
		if a.Optional {
			parts = append(parts, "["+a.Name+"]")
		} else {
			parts = append(parts, "<"+a.Name+">")
		}
	}
	for _, f := range c.Flags {
		parts = append(parts, f.usage(!f.Required))
	}
	return strings.Join(parts, " ")
}

// usage formats the flag for a synopsis.
func (f Flag) usage(bracket bool) string {
	s := "--" + f.Name
	if f.Value != "" {
		s += " <" + f.Value + ">"
	}
	if bracket {
		s = "[" + s + "]"
	}
	return s
}
