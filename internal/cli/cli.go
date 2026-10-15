// Package cli is the `mia <command>` surface.
package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
)

// Exit statuses, see mia(1).
const (
	exitOK   = 0
	exitFail = 1
)

// envelope is the one JSON object on stdout.
type envelope struct {
	OK     bool   `json:"ok"`
	Result any    `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
	Code   string `json:"code,omitempty"`
}

// Run runs one command line (without the program name) and returns the exit
// status.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "mia: no command; see `mia help`")
		return writeEnvelope(stdout, stderr, nil, usageErrorf("no command given"))
	}
	if args[0] == "--help" || args[0] == "-h" {
		return showHelp(stdout, stderr, "")
	}

	cmd, ok := lookup(args[0])
	if !ok {
		fmt.Fprintln(stderr, "mia: see `mia help` for the command list")
		return writeEnvelope(stdout, stderr, nil, usageErrorf("unknown command %q", args[0]))
	}

	inv, err := parse(cmd, args[1:])
	if errors.Is(err, flag.ErrHelp) {
		return showHelp(stdout, stderr, cmd.Name)
	}
	if err != nil {
		return finish(cmd, stdout, stderr, nil, err)
	}
	inv.Stdin = stdin
	inv.Stdout = stdout

	result, err := cmd.Run(inv)
	return finish(cmd, stdout, stderr, result, err)
}

// finish reports the outcome: envelope, or stderr for Text commands.
func finish(cmd Command, stdout, stderr io.Writer, result any, err error) int {
	if !cmd.Text {
		return writeEnvelope(stdout, stderr, result, err)
	}
	if err != nil {
		fmt.Fprintf(stderr, "mia %s: %v\n", cmd.Name, err)
		return exitFail
	}
	return exitOK
}

// writeEnvelope writes result or err and returns the exit status.
func writeEnvelope(stdout, stderr io.Writer, result any, err error) int {
	env := envelope{OK: true, Result: result}
	status := exitOK
	if err != nil {
		env = envelope{Error: err.Error(), Code: codeOf(err)}
		status = exitFail
	}
	if encErr := json.NewEncoder(stdout).Encode(env); encErr != nil {
		fmt.Fprintf(stderr, "mia: writing output: %v\n", encErr)
		return exitFail
	}
	return status
}

// showHelp prints the page for --help.
func showHelp(stdout, stderr io.Writer, topic string) int {
	if err := writeHelp(stdout, topic); err != nil {
		fmt.Fprintf(stderr, "mia: %v\n", err)
		return exitFail
	}
	return exitOK
}
