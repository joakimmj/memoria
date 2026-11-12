package brain

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func newRegistry(t *testing.T, dir string, brains map[string]string) *Registry {
	t.Helper()
	r, err := Load(filepath.Join(dir, "brains.json"))
	if err != nil {
		t.Fatal(err)
	}
	for name, loc := range brains {
		r.brains[name] = entry{Location: filepath.Join(dir, loc)}
	}
	return r
}

var errAny = errors.New("any error")

func TestRegister(t *testing.T) {
	tests := []struct {
		name     string
		existing map[string]string // name → location, relative to the temp dir
		file     string            // a file to create first, relative to the temp dir
		path     string            // relative to the temp dir
		chdir    bool              // pass path as given, from inside the temp dir
		brain    string            // --name; "" means the default
		wantName string
		wantLoc  string // relative to the temp dir
		wantErr  error  // errAny: any error
	}{
		{name: "name defaults to the basename", path: "notes/work", wantName: "work", wantLoc: "notes/work"},
		{name: "given name", path: "notes/work", brain: "w", wantName: "w", wantLoc: "notes/work"},
		{name: "unclean path", path: "notes/./x/../work/", wantName: "work", wantLoc: "notes/work"},
		{name: "relative path", path: "rel", chdir: true, wantName: "rel", wantLoc: "rel"},
		{name: "separator in name", path: "notes/work", brain: "a/b", wantErr: ErrInvalidName},
		{name: "leading dash", path: "notes/work", brain: "-x", wantErr: ErrInvalidName},
		{name: "root has no basename", path: "/", chdir: true, wantErr: ErrInvalidName},
		{name: "name taken", existing: map[string]string{"work": "elsewhere"}, path: "notes/work", wantErr: ErrCollision},
		{name: "folder taken", existing: map[string]string{"old": "notes/work"}, path: "notes/work/", brain: "new", wantErr: ErrCollision},
		{name: "path is a file", file: "f", path: "f", wantErr: errAny},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			r := newRegistry(t, dir, tt.existing)
			if tt.file != "" {
				if err := os.WriteFile(filepath.Join(dir, tt.file), nil, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			path := filepath.Join(dir, tt.path)
			if tt.chdir {
				t.Chdir(dir)
				path = tt.path
			}
			_, statErr := os.Stat(path)
			existedBefore := statErr == nil

			b, err := r.Register(path, tt.brain)
			if tt.wantErr != nil {
				if err == nil || (tt.wantErr != errAny && !errors.Is(err, tt.wantErr)) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				if _, err := os.Stat(path); !existedBefore && err == nil {
					t.Errorf("rejected Register created %s", path)
				}
				if len(r.List()) != len(tt.existing) {
					t.Errorf("rejected Register changed the registry: %+v", r.List())
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := Brain{Name: tt.wantName, Location: filepath.Join(dir, tt.wantLoc)}
			if b != want {
				t.Errorf("Register = %+v, want %+v", b, want)
			}
			if !b.Exists() {
				t.Errorf("folder %s was not created", b.Location)
			}
			if got, err := r.Get(b.Name); err != nil || got != want {
				t.Errorf("Get(%q) = %+v, %v", b.Name, got, err)
			}
		})
	}
}

func TestSaveLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "memoria", "brains.json")
	r, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	r.brains["work"] = entry{Location: "/w"}
	r.brains["personal"] = entry{Location: "/p"}
	if err := r.Save(); err != nil {
		t.Fatal(err)
	}
	again, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []Brain{{Name: "personal", Location: "/p"}, {Name: "work", Location: "/w"}}
	if got := again.List(); !reflect.DeepEqual(got, want) {
		t.Errorf("List = %+v, want %+v", got, want)
	}
	leftovers, _ := filepath.Glob(filepath.Join(dir, "memoria", ".brains-*"))
	if len(leftovers) > 0 {
		t.Errorf("temp files left behind: %v", leftovers)
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name, content string
		wantErr       string
	}{
		{name: "missing file is empty"},
		{name: "valid", content: `{"work": {"location": "/w"}}`},
		{name: "invalid JSON names the file", content: `{"work": `, wantErr: "brains.json"},
		{name: "unknown key", content: `{"work": {"path": "/w"}}`, wantErr: `unknown field "path"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "brains.json")
			if tt.content != "" {
				if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			_, err := Load(path)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("err = %v, want one containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestDeregister(t *testing.T) {
	dir := t.TempDir()
	r := newRegistry(t, dir, map[string]string{"work": "work"})
	if err := os.Mkdir(filepath.Join(dir, "work"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Deregister("work"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Get("work"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get after Deregister: err = %v, want ErrNotFound", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "work")); err != nil {
		t.Errorf("folder touched: %v", err)
	}
	if _, err := r.Deregister("work"); !errors.Is(err, ErrNotFound) {
		t.Errorf("second Deregister: err = %v, want ErrNotFound", err)
	}
}

func TestBrainStatus(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".mia_dna.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	b := Brain{Name: "b", Location: dir}
	if !b.Exists() || !b.HasDNA() {
		t.Errorf("Exists = %v, HasDNA = %v, want both true", b.Exists(), b.HasDNA())
	}
	gone := Brain{Name: "g", Location: filepath.Join(dir, "gone")}
	if gone.Exists() || gone.HasDNA() {
		t.Errorf("missing folder: Exists = %v, HasDNA = %v", gone.Exists(), gone.HasDNA())
	}
}

func TestSamePath(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"/a/b", "/a/b", true},
		{"/a/b/", "/a/b", true},
		{"/a/./b", "/a/c/../b", true},
		{"/a/b", "/a/c", false},
	}
	for _, tt := range tests {
		if got := SamePath(tt.a, tt.b); got != tt.want {
			t.Errorf("SamePath(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
