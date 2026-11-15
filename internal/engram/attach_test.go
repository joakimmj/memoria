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

func block(items ...string) string {
	return "<!-- SYNAPSES -->\n" + strings.Join(append(items, "***"), "\n") + "\n<!-- /SYNAPSES -->\n\n"
}

func TestAttach(t *testing.T) {
	defaults := map[string]config.Synapse{
		"down": {Inverse: "up", List: true, ShowEmpty: true},
		"up":   {Inverse: "down", List: true, ShowEmpty: true},
	}
	// next and prev hold one value each.
	single := map[string]config.Synapse{
		"next": {Inverse: "prev", ShowEmpty: true},
		"prev": {Inverse: "next", ShowEmpty: true},
	}
	// one parent, many children.
	tree := map[string]config.Synapse{
		"children": {Inverse: "parent", List: true, ShowEmpty: true},
		"parent":   {Inverse: "children", ShowEmpty: true},
	}
	tests := []struct {
		name                  string
		fields                map[string]config.Synapse
		files                 map[string]string
		source, field, target string
		want                  map[string]string // after; files left out must not change
		wantErr               error
	}{
		{
			name:   "both sides",
			fields: defaults,
			files:  map[string]string{"a.md": "# A\n", "b.md": "# B\n"},
			source: "a.md", field: "down", target: "b.md",
			want: map[string]string{
				"a.md": block("- **down:** [b](b.md)", "- **up:**") + "# A\n",
				"b.md": block("- **down:**", "- **up:** [a](a.md)") + "# B\n",
			},
		},
		{
			name:   "appends to existing values",
			fields: defaults,
			files: map[string]string{
				"a.md": block("- **down:** [c](c.md)", "- **up:**") + "# A\n",
				"b.md": "# B\n",
			},
			source: "a.md", field: "down", target: "b.md",
			want: map[string]string{
				"a.md": block("- **down:** [c](c.md), [b](b.md)", "- **up:**") + "# A\n",
				"b.md": block("- **down:**", "- **up:** [a](a.md)") + "# B\n",
			},
		},
		{
			name:   "already attached changes nothing",
			fields: defaults,
			files: map[string]string{
				"a.md": block("- **down:** [b](b.md)") + "# A\n",
				"b.md": block("- **up:** [a](a.md)") + "# B\n",
			},
			source: "a.md", field: "down", target: "b.md",
		},
		{
			name:   "one-sided synapse gets its inverse",
			fields: defaults,
			files: map[string]string{
				"a.md": block("- **down:** [b](b.md)") + "# A\n",
				"b.md": "# B\n",
			},
			source: "a.md", field: "down", target: "b.md",
			want: map[string]string{"b.md": block("- **down:**", "- **up:** [a](a.md)") + "# B\n"},
		},
		{
			name:   "no inverse",
			fields: map[string]config.Synapse{"see": {List: true, ShowEmpty: true}},
			files:  map[string]string{"a.md": "# A\n", "b.md": "# B\n"},
			source: "a.md", field: "see", target: "b.md",
			want: map[string]string{"a.md": block("- **see:** [b](b.md)") + "# A\n"},
		},
		{
			name:   "inverse not configured",
			fields: map[string]config.Synapse{"see": {Inverse: "gone", List: true, ShowEmpty: true}},
			files:  map[string]string{"a.md": "# A\n", "b.md": "# B\n"},
			source: "a.md", field: "see", target: "b.md",
			want: map[string]string{"a.md": block("- **see:** [b](b.md)") + "# A\n"},
		},
		{
			name:   "single field replaces, and the old target loses its inverse",
			fields: single,
			files: map[string]string{
				"a.md": block("- **next:** [c](c.md)", "- **prev:**") + "# A\n",
				"b.md": "# B\n",
				"c.md": block("- **next:**", "- **prev:** [a](a.md)") + "# C\n",
			},
			source: "a.md", field: "next", target: "b.md",
			want: map[string]string{
				"a.md": block("- **next:** [b](b.md)", "- **prev:**") + "# A\n",
				"b.md": block("- **next:**", "- **prev:** [a](a.md)") + "# B\n",
				"c.md": block("- **next:**", "- **prev:**") + "# C\n",
			},
		},
		{
			name:   "single inverse replaces, and the old one loses target",
			fields: tree,
			files: map[string]string{
				"a.md": "# A\n",
				"b.md": block("- **children:**", "- **parent:** [c](c.md)") + "# B\n",
				"c.md": block("- **children:** [b](b.md), [d](d.md)", "- **parent:**") + "# C\n",
			},
			source: "a.md", field: "children", target: "b.md",
			want: map[string]string{
				"a.md": block("- **children:** [b](b.md)", "- **parent:**") + "# A\n",
				"b.md": block("- **children:**", "- **parent:** [a](a.md)") + "# B\n",
				"c.md": block("- **children:** [d](d.md)", "- **parent:**") + "# C\n",
			},
		},
		{
			name:   "single field already holding target changes nothing",
			fields: single,
			files: map[string]string{
				"a.md": block("- **next:** [c](c.md), [b](b.md)", "- **prev:**") + "# A\n",
				"b.md": block("- **next:**", "- **prev:** [a](a.md)") + "# B\n",
				"c.md": block("- **next:**", "- **prev:** [a](a.md)") + "# C\n",
			},
			source: "a.md", field: "next", target: "b.md",
		},
		{
			name:   "replaced target whose file is gone",
			fields: single,
			files: map[string]string{
				"a.md": block("- **next:** [gone](gone.md)", "- **prev:**") + "# A\n",
				"b.md": "# B\n",
			},
			source: "a.md", field: "next", target: "b.md",
			want: map[string]string{
				"a.md": block("- **next:** [b](b.md)", "- **prev:**") + "# A\n",
				"b.md": block("- **next:**", "- **prev:** [a](a.md)") + "# B\n",
			},
		},
		{
			name:   "hand-added field kept, stray line dropped",
			fields: defaults,
			files: map[string]string{
				"a.md": block("- **see:** [s](s.md)", "a stray line", "- **up:**") + "# A\n",
				"b.md": block("- **down:**", "- **up:** [a](a.md)") + "# B\n",
			},
			source: "a.md", field: "down", target: "b.md",
			want: map[string]string{
				"a.md": block("- **down:** [b](b.md)", "- **see:** [s](s.md)", "- **up:**") + "# A\n",
			},
		},
		{
			name:   "under frontmatter",
			fields: defaults,
			files:  map[string]string{"a.md": "---\nk: v\n---\n# A\n", "b.md": "# B\n"},
			source: "a.md", field: "down", target: "b.md",
			want: map[string]string{
				"a.md": "---\nk: v\n---\n" + block("- **down:** [b](b.md)", "- **up:**") + "# A\n",
				"b.md": block("- **down:**", "- **up:** [a](a.md)") + "# B\n",
			},
		},
		{
			name:   "field not configured",
			fields: defaults,
			files:  map[string]string{"a.md": "# A\n", "b.md": "# B\n"},
			source: "a.md", field: "sideways", target: "b.md",
			wantErr: ErrNoField,
		},
		{
			name:   "to itself",
			fields: defaults,
			files:  map[string]string{"a.md": "# A\n"},
			source: "a.md", field: "down", target: "a.md",
			wantErr: ErrSelf,
		},
		{
			name:   "target missing",
			fields: defaults,
			files:  map[string]string{"a.md": "# A\n"},
			source: "a.md", field: "down", target: "b.md",
			wantErr: brain.ErrNoEngram,
		},
		{
			name:   "source missing",
			fields: defaults,
			files:  map[string]string{"b.md": "# B\n"},
			source: "a.md", field: "down", target: "b.md",
			wantErr: brain.ErrNoEngram,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := brain.Brain{Name: "work", Location: t.TempDir()}
			for f, content := range tt.files {
				if err := os.WriteFile(filepath.Join(b.Location, f), []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			cfg := config.Config{Synapses: tt.fields}
			source := brain.Ref{Brain: b, File: tt.source}
			target := brain.Ref{Brain: b, File: tt.target}

			got, err := Attach(cfg, source, tt.field, target)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Fatal(err)
			} else if want := (Attached{Brain: "work", Source: source.Path(), Field: tt.field, Target: target.Path()}); got != want {
				t.Errorf("Attach = %+v, want %+v", got, want)
			}

			entries, _ := os.ReadDir(b.Location)
			if len(entries) != len(tt.files) {
				t.Errorf("brain holds %v, want only %d files", entries, len(tt.files))
			}
			for f, before := range tt.files {
				want, changed := tt.want[f]
				if !changed {
					want = before
				}
				content, err := os.ReadFile(filepath.Join(b.Location, f))
				if err != nil {
					t.Fatal(err)
				}
				if string(content) != want {
					t.Errorf("%s =\n%s\nwant\n%s", f, content, want)
				}
			}
		})
	}
}

func TestReplaceFileKeepsMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a.md")
	if err := os.WriteFile(path, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := replaceFile(path, "new"); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(path)
	if string(content) != "new" || info.Mode().Perm() != 0o600 {
		t.Errorf("content %q, mode %v; want new, 0600", content, info.Mode().Perm())
	}
}
