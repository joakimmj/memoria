package cli

import (
	"errors"
	"flag"
	"maps"
	"slices"
	"testing"
)

// argsCmd exercises every kind of argument parse knows.
var argsCmd = Command{
	Name: "demo",
	Args: []Arg{{Name: "file"}, {Name: "line", Optional: true}, {Name: "col", Optional: true}},
	Flags: []Flag{
		{Name: "title", Value: "T"},
		{Name: "dir", Value: "dir", Required: true},
		{Name: "force"},
	},
}

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		argv      []string
		wantArgs  []string
		wantFlags map[string]string
		wantCode  string // "" means no error
		wantHelp  bool
	}{
		{
			name:      "positional then flag",
			argv:      []string{"a.md", "--dir", "d"},
			wantArgs:  []string{"a.md"},
			wantFlags: map[string]string{"dir": "d"},
		},
		{
			name:      "flags first, equals form",
			argv:      []string{"--dir=d", "--title", "Hello world", "a.md", "12"},
			wantArgs:  []string{"a.md", "12"},
			wantFlags: map[string]string{"dir": "d", "title": "Hello world"},
		},
		{
			name:      "interleaved, boolean flag",
			argv:      []string{"a.md", "--force", "12", "--dir", "d"},
			wantArgs:  []string{"a.md", "12"},
			wantFlags: map[string]string{"dir": "d", "force": "true"},
		},
		{
			name:      "double dash ends flags",
			argv:      []string{"--dir", "d", "--", "--weird.md", "-f"},
			wantArgs:  []string{"--weird.md", "-f"},
			wantFlags: map[string]string{"dir": "d"},
		},
		{
			// Second Parse consumes only "--"; a missed "--" leaves "-y" as a flag.
			name:      "double dash right after a positional",
			argv:      []string{"--dir", "d", "a.md", "--", "-x", "-y"},
			wantArgs:  []string{"a.md", "-x", "-y"},
			wantFlags: map[string]string{"dir": "d"},
		},
		{
			name:     "missing positional",
			argv:     []string{"--dir", "d"},
			wantCode: codeUsage,
		},
		{
			name:     "extra positional",
			argv:     []string{"a.md", "12", "3", "extra", "--dir", "d"},
			wantCode: codeUsage,
		},
		{
			name:     "missing required flag",
			argv:     []string{"a.md"},
			wantCode: codeUsage,
		},
		{
			name:     "unknown flag",
			argv:     []string{"a.md", "--dir", "d", "--nope"},
			wantCode: codeUsage,
		},
		{
			name:     "flag without its value",
			argv:     []string{"a.md", "--dir"},
			wantCode: codeUsage,
		},
		{
			name:     "help wins over missing arguments",
			argv:     []string{"--help"},
			wantCode: codeUsage,
			wantHelp: true,
		},
		{
			name:     "short help after a positional",
			argv:     []string{"a.md", "-h"},
			wantCode: codeUsage,
			wantHelp: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv, err := parse(argsCmd, tt.argv)
			if tt.wantCode != "" {
				if err == nil {
					t.Fatalf("parse(%q) = nil error, want code %q", tt.argv, tt.wantCode)
				}
				if got := codeOf(err); got != tt.wantCode {
					t.Errorf("code = %q, want %q (err: %v)", got, tt.wantCode, err)
				}
				if got := errors.Is(err, flag.ErrHelp); got != tt.wantHelp {
					t.Errorf("errors.Is(err, flag.ErrHelp) = %v, want %v (err: %v)", got, tt.wantHelp, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parse(%q): %v", tt.argv, err)
			}
			if !slices.Equal(inv.Args, tt.wantArgs) {
				t.Errorf("Args = %q, want %q", inv.Args, tt.wantArgs)
			}
			if !maps.Equal(inv.Flags, tt.wantFlags) {
				t.Errorf("Flags = %v, want %v", inv.Flags, tt.wantFlags)
			}
		})
	}
}
