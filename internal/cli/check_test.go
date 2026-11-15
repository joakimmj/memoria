package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCheckRebuild(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(dir, "data"))
	work, home := filepath.Join(dir, "work"), filepath.Join(dir, "home")
	for _, b := range []string{work, home} {
		if _, stderr, status := run(t, "register-brain", b); status != exitOK {
			t.Fatalf("register-brain: %s", stderr)
		}
	}
	files := map[string]string{
		"work/a.md": "<!-- SYNAPSES -->\n- **down:** [b](b.md), [x](x.md)\n- **up:**\n***\n<!-- /SYNAPSES -->\n\n# A\n",
		"work/b.md": "# B\n",
		"home/h.md": "# H\n",
	}
	for f, content := range files {
		if err := os.WriteFile(filepath.Join(dir, f), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	steps := []struct {
		name      string
		args      []string
		wantOK    bool
		wantCode  string
		wantKinds []string
	}{
		{name: "check", args: []string{"check"}, wantOK: true, wantKinds: []string{"broken_synapse", "missing_inverse"}},
		{name: "check a clean brain", args: []string{"check", "--brain", "home", "--strict"}, wantOK: true, wantKinds: []string{}},
		{name: "check --strict", args: []string{"check", "--strict"}, wantCode: codeProblems, wantKinds: []string{"broken_synapse", "missing_inverse"}},
		{name: "rebuild needs a brain", args: []string{"rebuild"}, wantCode: codeBrain},
		{name: "rebuild", args: []string{"rebuild", "--brain", "work"}, wantOK: true, wantKinds: []string{"broken_synapse", "missing_inverse"}},
		{name: "rebuild --fix", args: []string{"rebuild", "--brain", "work", "--fix"}, wantOK: true, wantKinds: []string{"broken_synapse"}},
		{name: "fixed for check too", args: []string{"check", "--brain", "work"}, wantOK: true, wantKinds: []string{"broken_synapse"}},
	}
	for _, st := range steps {
		stdout, stderr, _ := run(t, st.args...)
		env := decode(t, stdout)
		code, _ := env["code"].(string)
		if (env["ok"] == true) != st.wantOK || code != st.wantCode {
			t.Fatalf("%s: ok %v, code %q; want %v, %q\nstdout: %s\nstderr: %s", st.name, env["ok"], code, st.wantOK, st.wantCode, stdout, stderr)
		}
		if st.wantKinds == nil {
			continue
		}
		list, ok := env["result"].([]any)
		if !ok {
			t.Fatalf("%s: result is not a list: %s", st.name, stdout)
		}
		kinds := []string{}
		for _, p := range list {
			kinds = append(kinds, p.(map[string]any)["kind"].(string))
		}
		if strings.Join(kinds, ",") != strings.Join(st.wantKinds, ",") {
			t.Errorf("%s: kinds %v, want %v", st.name, kinds, st.wantKinds)
		}
	}
	if b, _ := os.ReadFile(filepath.Join(work, "b.md")); !strings.Contains(string(b), "- **up:** [a](a.md)") {
		t.Errorf("b.md = %q, want the inverse written by --fix", b)
	}
}
