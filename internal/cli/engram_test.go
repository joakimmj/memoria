package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCreateEngram(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(dir, "data"))
	work := filepath.Join(dir, "work")
	if _, stderr, status := run(t, "register-brain", work); status != exitOK {
		t.Fatalf("register-brain: %s", stderr)
	}
	if err := os.WriteFile(filepath.Join(work, ".mia_dna.json"), []byte(`{"engrams": {"filename": {"prefix": "none"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	steps := []struct {
		name     string
		args     []string
		stdin    string
		wantCode string // "" means ok
		wantFile string // relative to the brain; checked when set
		want     string
	}{
		{name: "no brain", args: []string{"create-engram", "--title", "x"}, wantCode: codeBrain},
		{name: "unknown brain", args: []string{"create-engram", "--brain", "nope", "--title", "x"}, wantCode: codeNotFound},
		{name: "create", args: []string{"create-engram", "--brain", "work", "--title", "Project X"}, wantFile: "project_x.md", want: "# Project X\n\n"},
		{name: "collision", args: []string{"create-engram", "--brain", "work", "--title", "project x"}, wantCode: codeCollision},
		{name: "body from stdin", args: []string{"create-engram", "--brain", "work", "--title", "piped", "--body", "-"}, stdin: "hello\n", wantFile: "piped.md", want: "# piped\n\nhello\n"},
		{name: "no title", args: []string{"create-engram", "--brain", "work"}, wantCode: codeUsage},
		{name: "empty slug", args: []string{"create-engram", "--brain", "work", "--title", "???"}, wantCode: codeEmptySlug},
	}
	for _, st := range steps {
		var out, errOut bytes.Buffer
		Run(st.args, strings.NewReader(st.stdin), &out, &errOut)
		env := decode(t, out.String())
		code, _ := env["code"].(string)
		if code != st.wantCode || (st.wantCode == "" && env["ok"] != true) {
			t.Fatalf("%s: code = %q, want %q\nstdout: %s\nstderr: %s", st.name, code, st.wantCode, out.String(), errOut.String())
		}
		if st.wantFile != "" {
			got, err := os.ReadFile(filepath.Join(work, st.wantFile))
			if err != nil || string(got) != st.want {
				t.Errorf("%s: %s = %q, %v; want %q", st.name, st.wantFile, got, err, st.want)
			}
		}
	}
}
