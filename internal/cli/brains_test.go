package cli

import (
	"path/filepath"
	"testing"
)

// Steps share one registry and run in order.
func TestRunBrains(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(dir, "data"))
	work := filepath.Join(dir, "notes", "work")

	steps := []struct {
		name      string
		args      []string
		wantCode  string // "" means ok
		wantCount int    // brains listed; -1 skips the check
	}{
		{name: "empty registry", args: []string{"brains"}, wantCount: 0},
		{name: "register", args: []string{"register-brain", work}, wantCount: -1},
		{name: "same folder again", args: []string{"register-brain", work, "--name", "other"}, wantCode: codeCollision, wantCount: -1},
		{name: "listed", args: []string{"brains"}, wantCount: 1},
		{name: "config with brain", args: []string{"config", "--brain", "work"}, wantCount: -1},
		{name: "config with unknown brain", args: []string{"config", "--brain", "nope"}, wantCode: codeNotFound, wantCount: -1},
		{name: "config --defaults --brain", args: []string{"config", "--defaults", "--brain", "work"}, wantCode: codeUsage, wantCount: -1},
		{name: "deregister", args: []string{"deregister-brain", "work"}, wantCount: -1},
		{name: "deregister again", args: []string{"deregister-brain", "work"}, wantCode: codeNotFound, wantCount: -1},
		{name: "empty again", args: []string{"brains"}, wantCount: 0},
	}
	for _, st := range steps {
		stdout, stderr, _ := run(t, st.args...)
		env := decode(t, stdout)
		code, _ := env["code"].(string)
		if code != st.wantCode || (st.wantCode == "" && env["ok"] != true) {
			t.Fatalf("%s: code = %q, want %q\nstdout: %s\nstderr: %s", st.name, code, st.wantCode, stdout, stderr)
		}
		if st.wantCount >= 0 {
			list, _ := env["result"].([]any)
			if len(list) != st.wantCount {
				t.Fatalf("%s: %d brains, want %d: %s", st.name, len(list), st.wantCount, stdout)
			}
		}
	}
}
