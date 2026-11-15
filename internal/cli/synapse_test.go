package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAttachSynapse(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(dir, "data"))
	work, personal := filepath.Join(dir, "work"), filepath.Join(dir, "personal")
	for _, b := range []string{work, personal} {
		if _, stderr, status := run(t, "register-brain", b); status != exitOK {
			t.Fatalf("register-brain: %s", stderr)
		}
	}
	for _, f := range []string{"work/a.md", "work/b.md", "personal/b.md", "personal/p.md"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("# x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	steps := []struct {
		name     string
		args     []string
		wantCode string // "" means ok
	}{
		{name: "names", args: []string{"attach-synapse", "a", "down", "b"}},
		{name: "path and name", args: []string{"attach-synapse", filepath.Join(personal, "p.md"), "down", "b"}},
		{name: "names in two brains", args: []string{"attach-synapse", "b", "down", "b.md"}, wantCode: codeAmbiguous},
		{name: "names in no one brain", args: []string{"attach-synapse", "a", "down", "p"}, wantCode: codeNotFound},
		{name: "paths in two brains", args: []string{"attach-synapse", filepath.Join(work, "a.md"), "down", filepath.Join(personal, "p.md")}, wantCode: codeCrossBrain},
		{name: "path outside --brain", args: []string{"attach-synapse", "--brain", "personal", filepath.Join(work, "a.md"), "down", "b"}, wantCode: codeNotInBrain},
		{name: "path in a subfolder", args: []string{"attach-synapse", filepath.Join(work, "sub", "x.md"), "down", "b"}, wantCode: codeNotAnEngram},
		{name: "unknown field", args: []string{"attach-synapse", "--brain", "work", "a", "sideways", "b"}, wantCode: codeNotFound},
		{name: "to itself", args: []string{"attach-synapse", "--brain", "work", "a", "down", "a"}, wantCode: codeUsage},
		{name: "missing target", args: []string{"attach-synapse", "a", "down"}, wantCode: codeUsage},
	}
	for _, st := range steps {
		stdout, stderr, _ := run(t, st.args...)
		env := decode(t, stdout)
		code, _ := env["code"].(string)
		if code != st.wantCode || (st.wantCode == "" && env["ok"] != true) {
			t.Fatalf("%s: code = %q, want %q\nstdout: %s\nstderr: %s", st.name, code, st.wantCode, stdout, stderr)
		}
	}

	want := map[string]string{
		"work/a.md":     "- **down:** [b](b.md)",
		"work/b.md":     "- **up:** [a](a.md)",
		"personal/p.md": "- **down:** [b](b.md)",
		"personal/b.md": "- **up:** [p](p.md)",
	}
	for _, b := range []string{work, personal} {
		if _, err := os.Stat(filepath.Join(b, ".mia_atlas.json")); err != nil {
			t.Errorf("atlas not refreshed: %v", err)
		}
	}
	for f, item := range want {
		content, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil || !strings.Contains(string(content), item) {
			t.Errorf("%s = %q, %v; want it to hold %q", f, content, err, item)
		}
	}
}
