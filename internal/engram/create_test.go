package engram

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joakimmj/memoria/internal/brain"
	"github.com/joakimmj/memoria/internal/config"
)

func TestCreate(t *testing.T) {
	defaults, err := config.Defaults()
	if err != nil {
		t.Fatal(err)
	}
	withPrefix := func(prefix string) config.Config {
		c := defaults
		c.Engrams.Filename.Prefix = prefix
		return c
	}
	noSynapses := defaults
	noSynapses.Synapses = nil
	hiddenSynapses := defaults
	hiddenSynapses.Synapses = map[string]config.Synapse{"next": {List: true}}
	const header = "<!-- SYNAPSES -->\n- **down:**\n- **up:**\n***\n<!-- /SYNAPSES -->\n\n"
	tests := []struct {
		name     string
		cfg      config.Config
		opts     Options
		wantFile string
		want     string
		wantBody Position
		wantErr  error
	}{
		{name: "date prefix", cfg: defaults, opts: Options{Title: "Project X"}, wantFile: "2026-08-01_project_x.md", want: header + "# Project X\n\n", wantBody: Position{Line: 9}},
		{name: "date alone without a title", cfg: defaults, wantFile: "2026-08-01.md", want: header + "# \n\n", wantBody: Position{Line: 9}},
		{name: "title trimmed", cfg: withPrefix("none"), opts: Options{Title: "  Project X  "}, wantFile: "project_x.md", want: header + "# Project X\n\n", wantBody: Position{Line: 9}},
		{name: "body", cfg: withPrefix("none"), opts: Options{Title: "x", Body: "hello\n"}, wantFile: "x.md", want: header + "# x\n\nhello\n", wantBody: Position{Line: 9}},
		{name: "no synapse fields, no header", cfg: noSynapses, opts: Options{Title: "x"}, wantFile: "2026-08-01_x.md", want: "# x\n\n", wantBody: Position{Line: 3}},
		{name: "no field to show, no header", cfg: hiddenSynapses, opts: Options{Title: "x"}, wantFile: "2026-08-01_x.md", want: "# x\n\n", wantBody: Position{Line: 3}},
		{name: "none needs a title", cfg: withPrefix("none"), wantErr: ErrNoTitle},
		{name: "empty slug", cfg: defaults, opts: Options{Title: "???"}, wantErr: ErrEmptySlug},
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
			if got.Path != path || got.Brain != "work" || !got.Created || got.Body != tt.wantBody {
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
	first, err := Create(b, defaults, Options{Title: "x", Body: "first"}, testNow)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Create(b, defaults, Options{Title: "X", Body: "second"}, testNow); !errors.Is(err, ErrCollision) {
		t.Fatalf("err = %v, want ErrCollision", err)
	}
	content, _ := os.ReadFile(first.Path)
	if !strings.HasSuffix(string(content), "# x\n\nfirst") {
		t.Errorf("collision changed the file: %q", content)
	}
}
