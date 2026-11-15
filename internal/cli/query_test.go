package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunQueries(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(dir, "data"))
	work, home, gone := filepath.Join(dir, "work"), filepath.Join(dir, "home"), filepath.Join(dir, "gone")
	for _, b := range []string{work, home, gone} {
		if _, stderr, status := run(t, "register-brain", b); status != exitOK {
			t.Fatalf("register-brain: %s", stderr)
		}
	}
	if err := os.Remove(gone); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"work/a.md": "<!-- SYNAPSES -->\n- **up:**\n***\n<!-- /SYNAPSES -->\n\n# A\n- [ ] one\n- [x] two\n",
		"home/h.md": "# H\n- [ ] three\n",
	}
	for f, content := range files {
		if err := os.WriteFile(filepath.Join(dir, f), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	steps := []struct {
		name       string
		args       []string
		wantCode   string // "" means ok, unless wantFail
		wantFail   bool   // an error without a code
		wantCount  int    // list length, or total tasks; -1 skips
		wantWarn   bool
		wantResult string // substring of stdout
	}{
		{name: "engrams: every brain, missing one skipped", args: []string{"engrams"}, wantCount: 2, wantWarn: true},
		{name: "engrams: named brains", args: []string{"engrams", "--brain", "home"}, wantCount: 1},
		{name: "engrams: named missing folder", args: []string{"engrams", "--brain", "gone"}, wantFail: true, wantCount: -1},
		{name: "engrams: unknown brain", args: []string{"engrams", "--brain", "nope"}, wantCode: codeNotFound, wantCount: -1},
		{name: "engram", args: []string{"engram", "a"}, wantCount: -1, wantResult: `"title":"A"`},
		{name: "engram --body", args: []string{"engram", "a", "--body"}, wantCount: -1, wantResult: `"content":"# A\n- [ ] one\n- [x] two\n"`},
		{name: "engram: none", args: []string{"engram", "zzz"}, wantCode: codeNotFound, wantCount: -1},
		{name: "engram: not indexed", args: []string{"engram", filepath.Join(work, "never.md")}, wantCode: codeNotFound, wantCount: -1},
		{name: "tasks", args: []string{"tasks"}, wantCount: 3, wantWarn: true},
		{name: "tasks --state", args: []string{"tasks", "--state", "done"}, wantCount: 1, wantWarn: true, wantResult: `"done":1`},
		{name: "tasks --engram", args: []string{"tasks", "--engram", "a"}, wantCount: 2},
		{name: "tasks --engram not in --brain", args: []string{"tasks", "--engram", filepath.Join(work, "a.md"), "--brain", "home"}, wantCode: codeNotInBrain, wantCount: -1},
		{name: "tasks --engram with two brains", args: []string{"tasks", "--engram", "a", "--brain", "home", "--brain", "work"}, wantCode: codeUsage, wantCount: -1},
		{name: "tasks bad state", args: []string{"tasks", "--state", "open"}, wantCode: codeUsage, wantCount: -1},
	}
	for _, st := range steps {
		stdout, stderr, _ := run(t, st.args...)
		env := decode(t, stdout)
		code, _ := env["code"].(string)
		if code != st.wantCode || (st.wantCode == "" && (env["ok"] == true) == st.wantFail) {
			t.Fatalf("%s: code = %q, want %q\nstdout: %s\nstderr: %s", st.name, code, st.wantCode, stdout, stderr)
		}
		if st.wantCount >= 0 {
			n := -1
			switch r := env["result"].(type) {
			case []any:
				n = len(r)
			case map[string]any:
				n = int(r["total"].(float64))
			}
			if n != st.wantCount {
				t.Errorf("%s: count %d, want %d: %s", st.name, n, st.wantCount, stdout)
			}
		}
		if st.wantWarn != strings.Contains(stderr, "warning") {
			t.Errorf("%s: stderr = %q, want warning %v", st.name, stderr, st.wantWarn)
		}
		if st.wantResult != "" && !strings.Contains(stdout, st.wantResult) {
			t.Errorf("%s: stdout %s, want it to hold %s", st.name, stdout, st.wantResult)
		}
	}
}
