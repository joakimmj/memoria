package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// run runs a command line in-process.
func run(t *testing.T, args ...string) (stdout, stderr string, status int) {
	t.Helper()
	var out, errOut bytes.Buffer
	status = Run(args, strings.NewReader(""), &out, &errOut)
	return out.String(), errOut.String(), status
}

// decode parses stdout as exactly one envelope.
func decode(t *testing.T, stdout string) map[string]any {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(stdout))
	var env map[string]any
	if err := dec.Decode(&env); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout)
	}
	if dec.More() {
		t.Fatalf("stdout holds more than one JSON object:\n%s", stdout)
	}
	return env
}

func TestRunEnvelope(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantStatus int
		wantOK     bool
		wantCode   string
	}{
		{name: "version", args: []string{"version"}, wantStatus: exitOK, wantOK: true},
		{name: "commands", args: []string{"commands"}, wantStatus: exitOK, wantOK: true},
		{name: "no command", args: nil, wantStatus: exitFail, wantCode: codeUsage},
		{name: "unknown command", args: []string{"nope"}, wantStatus: exitFail, wantCode: codeUsage},
		{name: "extra argument", args: []string{"version", "x"}, wantStatus: exitFail, wantCode: codeUsage},
		{name: "man without --dir", args: []string{"man"}, wantStatus: exitFail, wantCode: codeUsage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _, status := run(t, tt.args...)
			if status != tt.wantStatus {
				t.Errorf("status = %d, want %d", status, tt.wantStatus)
			}
			env := decode(t, stdout)
			if env["ok"] != tt.wantOK {
				t.Errorf("ok = %v, want %v", env["ok"], tt.wantOK)
			}
			if code, _ := env["code"].(string); code != tt.wantCode {
				t.Errorf("code = %q, want %q", code, tt.wantCode)
			}
			if !tt.wantOK && env["error"] == "" {
				t.Errorf("error message is empty")
			}
		})
	}
}

func TestRunHelp(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantStatus int
		wantIn     string // a substring of stdout
	}{
		{name: "help", args: []string{"help"}, wantStatus: exitOK, wantIn: "mia - the memoria command line"},
		{name: "--help", args: []string{"--help"}, wantStatus: exitOK, wantIn: "COMMANDS"},
		{name: "help topic", args: []string{"help", "man"}, wantStatus: exitOK, wantIn: "mia man --dir <dir>"},
		{name: "command --help", args: []string{"man", "--help"}, wantStatus: exitOK, wantIn: "mia man --dir <dir>"},
		{name: "unknown topic", args: []string{"help", "nope"}, wantStatus: exitFail},
		{name: "glob in topic", args: []string{"help", "*"}, wantStatus: exitFail},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, status := run(t, tt.args...)
			if status != tt.wantStatus {
				t.Fatalf("status = %d, want %d (stderr: %s)", status, tt.wantStatus, stderr)
			}
			if !strings.Contains(stdout, tt.wantIn) {
				t.Errorf("stdout does not contain %q:\n%s", tt.wantIn, stdout)
			}
			if status != exitOK && stdout != "" {
				t.Errorf("failed help wrote to stdout: %q", stdout)
			}
		})
	}
}

func TestRunMan(t *testing.T) {
	dir := t.TempDir()
	stdout, stderr, status := run(t, "man", "--dir", dir)
	if status != exitOK {
		t.Fatalf("status = %d, stdout %s, stderr %s", status, stdout, stderr)
	}
	env := decode(t, stdout)
	result, _ := env["result"].(map[string]any)
	pages, _ := result["pages"].([]any)
	if len(pages) == 0 {
		t.Fatalf("no pages written: %s", stdout)
	}
	got, err := os.ReadFile(filepath.Join(dir, "man1", "mia-man.1"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(got), `.TH "MIA-MAN" "1"`) {
		t.Errorf("mia-man.1 does not start with .TH:\n%s", got)
	}
}
