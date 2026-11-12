package cli

import (
	"flag"
	"io"
)

// parse lets flags and positionals interleave; `--` ends flags. -h/--help →
// error wrapping flag.ErrHelp.
func parse(cmd Command, argv []string) (Invocation, error) {
	fs := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	for _, f := range cmd.Flags {
		if f.Value == "" {
			fs.Bool(f.Name, false, f.Summary)
		} else {
			fs.String(f.Name, "", f.Summary)
		}
	}

	var pos []string
	input := argv

	for {
		if err := fs.Parse(input); err != nil {
			return Invocation{}, usageErrorf("%s: %w", cmd.Name, err)
		}
		rest := fs.Args()

		if len(rest) == 0 {
			break
		}

		consumed := len(input) - len(rest)
		if consumed > 0 && input[consumed-1] == "--" {
			pos = append(pos, rest...) // after "--": all positional
			break
		}

		pos = append(pos, rest[0]) // Parse stops at a positional; take it, go on
		input = rest[1:]
	}

	// Visit skips flags not given.
	flags := map[string]string{}
	fs.Visit(func(f *flag.Flag) {
		flags[f.Name] = f.Value.String()
	})

	if err := checkCounts(cmd, pos, flags); err != nil {
		return Invocation{}, err
	}
	return Invocation{Args: pos, Flags: flags}, nil
}

func checkCounts(cmd Command, pos []string, flags map[string]string) error {
	for i, a := range cmd.Args {
		if i >= len(pos) && !a.Optional {
			return usageErrorf("%s: missing <%s>", cmd.Name, a.Name)
		}
	}
	if len(pos) > len(cmd.Args) {
		return usageErrorf("%s: unexpected argument %q", cmd.Name, pos[len(cmd.Args)])
	}
	for _, f := range cmd.Flags {
		if _, given := flags[f.Name]; f.Required && !given {
			return usageErrorf("%s: missing --%s", cmd.Name, f.Name)
		}
	}
	return nil
}
