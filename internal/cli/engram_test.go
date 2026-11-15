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
	if err := os.WriteFile(filepath.Join(work, ".mia_dna.json"), []byte(`{"engrams": {"filename": {"prefix": "none"}}, "synapses": {"up": null}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	const header = "<!-- SYNAPSES -->\n- **down:**\n***\n<!-- /SYNAPSES -->\n\n"
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
		{name: "create", args: []string{"create-engram", "--brain", "work", "--title", "Project X"}, wantFile: "project_x.md", want: header + "# Project X\n\n"},
		{name: "collision", args: []string{"create-engram", "--brain", "work", "--title", "project x"}, wantCode: codeCollision},
		{name: "body from stdin", args: []string{"create-engram", "--brain", "work", "--title", "piped", "--body", "-"}, stdin: "hello\n", wantFile: "piped.md", want: header + "# piped\n\nhello\n"},
		{name: "no title", args: []string{"create-engram", "--brain", "work"}, wantCode: codeUsage},
		{name: "empty slug", args: []string{"create-engram", "--brain", "work", "--title", "???"}, wantCode: codeEmptySlug},
	}
	defer func() {
		if _, err := os.Stat(filepath.Join(work, ".mia_atlas.json")); err != nil {
			t.Errorf("atlas not refreshed: %v", err)
		}
	}()
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

func TestCreateEngramAtlasFailureWarns(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(dir, "data"))
	work := filepath.Join(dir, "work")
	if _, stderr, status := run(t, "register-brain", work); status != exitOK {
		t.Fatalf("register-brain: %s", stderr)
	}
	// A folder where the atlas goes makes writing it fail.
	if err := os.MkdirAll(filepath.Join(work, ".mia_atlas.json", "x"), 0o755); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, status := run(t, "create-engram", "--brain", "work", "--title", "x")
	if status != exitOK || decode(t, stdout)["ok"] != true {
		t.Fatalf("status %d, stdout %s", status, stdout)
	}
	if !strings.Contains(stderr, "warning") {
		t.Errorf("stderr = %q, want a warning", stderr)
	}
}
