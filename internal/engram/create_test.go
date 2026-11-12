package engram

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/joakimmj/memoria/internal/brain"
	"github.com/joakimmj/memoria/internal/config"
)

func TestCreate(t *testing.T) {
	defaults, err := config.Defaults()
	if err != nil {
		t.Fatal(err)
	}
	withPrefix := func(prefix string) config.Engrams {
		e := defaults.Engrams
		e.Filename.Prefix = prefix
		return e
	}
	tests := []struct {
		name     string
		cfg      config.Engrams
		opts     Options
		wantFile string
		want     string
		wantErr  error
	}{
		{name: "date prefix", cfg: defaults.Engrams, opts: Options{Title: "Project X"}, wantFile: "2026-08-01_project_x.md", want: "# Project X\n\n"},
		{name: "date alone without a title", cfg: defaults.Engrams, wantFile: "2026-08-01.md", want: "# \n\n"},
		{name: "title trimmed", cfg: withPrefix("none"), opts: Options{Title: "  Project X  "}, wantFile: "project_x.md", want: "# Project X\n\n"},
		{name: "body", cfg: withPrefix("none"), opts: Options{Title: "x", Body: "hello\n"}, wantFile: "x.md", want: "# x\n\nhello\n"},
		{name: "none needs a title", cfg: withPrefix("none"), wantErr: ErrNoTitle},
		{name: "empty slug", cfg: defaults.Engrams, opts: Options{Title: "???"}, wantErr: ErrEmptySlug},
		{name: "concept prefix not yet", cfg: withPrefix("concept"), opts: Options{Title: "x"}, wantErr: ErrPrefix},
		{name: "unknown prefix", cfg: withPrefix("weekly"), opts: Options{Title: "x"}, wantErr: ErrPrefix},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := brain.Brain{Name: "work", Location: t.TempDir()}
			got, err := Create(b, tt.cfg, tt.opts, testNow)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				if entries, _ := os.ReadDir(b.Location); len(entries) > 0 {
					t.Errorf("failed Create wrote %v", entries)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(b.Location, tt.wantFile)
			if got.Path != path || got.Brain != "work" || !got.Created {
				t.Errorf("Create = %+v, want path %s", got, path)
			}
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(content) != tt.want {
				t.Errorf("content = %q, want %q", content, tt.want)
			}
		})
	}
}

func TestCreateCollision(t *testing.T) {
	defaults, err := config.Defaults()
	if err != nil {
		t.Fatal(err)
	}
	b := brain.Brain{Name: "work", Location: t.TempDir()}
	first, err := Create(b, defaults.Engrams, Options{Title: "x", Body: "first"}, testNow)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Create(b, defaults.Engrams, Options{Title: "X", Body: "second"}, testNow); !errors.Is(err, ErrCollision) {
		t.Fatalf("err = %v, want ErrCollision", err)
	}
	content, _ := os.ReadFile(first.Path)
	if string(content) != "# x\n\nfirst" {
		t.Errorf("collision changed the file: %q", content)
	}
}
