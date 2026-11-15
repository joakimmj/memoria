package brain

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// work/        a.md  shared.md  notes.txt  deep/n.md
// work/inner/  i.md  (brain "inner")
// personal/    p.md  shared.md
// outside/     o.md
func newBrains(t *testing.T) (*Registry, string) {
	t.Helper()
	dir := t.TempDir()
	files := []string{
		"work/a.md", "work/shared.md", "work/notes.txt", "work/deep/n.md",
		"work/inner/i.md",
		"personal/p.md", "personal/shared.md",
		"outside/o.md",
	}
	for _, f := range files {
		path := filepath.Join(dir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	r := newRegistry(t, dir, map[string]string{"work": "work", "inner": "work/inner", "personal": "personal"})
	return r, dir
}

type engramArg struct {
	arg  string
	path bool
}

func (a engramArg) resolve(dir string) string {
	if a.path {
		return filepath.Join(dir, a.arg)
	}
	return a.arg
}

func name(s string) engramArg { return engramArg{arg: s} }
func path(s string) engramArg { return engramArg{arg: s, path: true} }

func TestResolve(t *testing.T) {
	tests := []struct {
		name      string
		arg       engramArg
		chdir     string // relative to the temp dir
		brain     string
		wantBrain string
		wantFile  string
		wantErr   error
	}{
		{name: "name", arg: name("a"), wantBrain: "work", wantFile: "a.md"},
		{name: "name with .md", arg: name("a.md"), wantBrain: "work", wantFile: "a.md"},
		{name: "name in a nested brain", arg: name("i"), wantBrain: "inner", wantFile: "i.md"},
		{name: "name in two brains", arg: name("shared"), wantErr: ErrAmbiguous},
		{name: "name in the brain given", arg: name("shared"), brain: "personal", wantBrain: "personal", wantFile: "shared.md"},
		{name: "name nowhere", arg: name("missing"), wantErr: ErrNoEngram},
		{name: "name not in the brain given", arg: name("p"), brain: "work", wantErr: ErrNoEngram},
		{name: "a folder is not an engram name", arg: name("deep"), wantErr: ErrNoEngram},
		{name: "unknown brain", arg: name("a"), brain: "nope", wantErr: ErrNotFound},
		{name: "path", arg: path("work/a.md"), wantBrain: "work", wantFile: "a.md"},
		{name: "relative path", arg: name("./a.md"), chdir: "work", wantBrain: "work", wantFile: "a.md"},
		{name: "unclean path", arg: path("personal/../work/a.md"), wantBrain: "work", wantFile: "a.md"},
		{name: "path not checked for existence", arg: path("work/gone.md"), wantBrain: "work", wantFile: "gone.md"},
		{name: "path in a nested brain", arg: path("work/inner/i.md"), wantBrain: "inner", wantFile: "i.md"},
		{name: "path in a subfolder", arg: path("work/deep/n.md"), wantErr: ErrNotAnEngram},
		{name: "path not .md", arg: path("work/notes.txt"), wantErr: ErrNotAnEngram},
		{name: "path to the brain itself", arg: path("work/"), wantErr: ErrNotAnEngram},
		{name: "path outside every brain", arg: path("outside/o.md"), wantErr: ErrNotInBrain},
		{name: "path escaping a brain", arg: path("work/../outside/o.md"), wantErr: ErrNotInBrain},
		{name: "path outside the brain given", arg: path("personal/p.md"), brain: "work", wantErr: ErrNotInBrain},
		{name: "path in the brain given", arg: path("personal/p.md"), brain: "personal", wantBrain: "personal", wantFile: "p.md"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, dir := newBrains(t)
			if tt.chdir != "" {
				t.Chdir(filepath.Join(dir, tt.chdir))
			}
			got, err := r.Resolve(tt.arg.resolve(dir), tt.brain)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.Brain.Name != tt.wantBrain || got.File != tt.wantFile {
				t.Errorf("Resolve = %s %s, want %s %s", got.Brain.Name, got.File, tt.wantBrain, tt.wantFile)
			}
		})
	}
}

func TestResolvePair(t *testing.T) {
	tests := []struct {
		name         string
		a, b         engramArg
		brain        string
		wantBrain    string
		wantA, wantB string
		wantErr      error
	}{
		{name: "names: the one brain holding both", a: name("a"), b: name("shared"), wantBrain: "work", wantA: "a.md", wantB: "shared.md"},
		{name: "names: two brains hold both", a: name("shared"), b: name("shared.md"), wantErr: ErrAmbiguous},
		{name: "names: no brain holds both", a: name("a"), b: name("p"), wantErr: ErrNoEngram},
		{name: "names: one nowhere", a: name("a"), b: name("missing"), wantErr: ErrNoEngram},
		{name: "names in the brain given", a: name("shared"), b: name("p"), brain: "personal", wantBrain: "personal", wantA: "shared.md", wantB: "p.md"},
		{name: "names not both in the brain given", a: name("a"), b: name("p"), brain: "work", wantErr: ErrNoEngram},
		{name: "path then name", a: path("work/a.md"), b: name("shared"), wantBrain: "work", wantA: "a.md", wantB: "shared.md"},
		{name: "name then path", a: name("shared"), b: path("personal/p.md"), wantBrain: "personal", wantA: "shared.md", wantB: "p.md"},
		{name: "name not in the path's brain", a: path("work/a.md"), b: name("p"), wantErr: ErrNoEngram},
		{name: "two paths, one brain", a: path("work/a.md"), b: path("work/shared.md"), wantBrain: "work", wantA: "a.md", wantB: "shared.md"},
		{name: "two paths, two brains", a: path("work/a.md"), b: path("personal/p.md"), wantErr: ErrCrossBrain},
		{name: "path outside every brain", a: path("outside/o.md"), b: name("a"), wantErr: ErrNotInBrain},
		{name: "path outside the brain given", a: path("personal/p.md"), b: name("a"), brain: "work", wantErr: ErrNotInBrain},
		{name: "path in a subfolder", a: name("a"), b: path("work/deep/n.md"), wantErr: ErrNotAnEngram},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, dir := newBrains(t)
			a, b, err := r.ResolvePair(tt.a.resolve(dir), tt.b.resolve(dir), tt.brain)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if a.Brain != b.Brain || a.Brain.Name != tt.wantBrain || a.File != tt.wantA || b.File != tt.wantB {
				t.Errorf("ResolvePair = %s %s, %s %s; want %s %s %s", a.Brain.Name, a.File, b.Brain.Name, b.File, tt.wantBrain, tt.wantA, tt.wantB)
			}
		})
	}
}

func TestWithin(t *testing.T) {
	tests := []struct {
		dir, path string
		wantRel   string
		wantOK    bool
	}{
		{"/b", "/b/x.md", "x.md", true},
		{"/b", "/b/sub/x.md", filepath.Join("sub", "x.md"), true},
		{"/b", "/b", ".", true},
		{"/b", "/c/x.md", "", false},
		{"/b", "/bb/x.md", "", false},
		{"/b/", "/b/./x.md", "x.md", true},
	}
	for _, tt := range tests {
		rel, ok := Within(tt.dir, tt.path)
		if rel != tt.wantRel || ok != tt.wantOK {
			t.Errorf("Within(%q, %q) = %q, %v; want %q, %v", tt.dir, tt.path, rel, ok, tt.wantRel, tt.wantOK)
		}
	}
}
