package atlas

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/joakimmj/memoria/internal/config"
)

func entry(synapses map[string][]string, links ...string) Entry {
	if synapses == nil {
		synapses = map[string][]string{}
	}
	return Entry{Synapses: synapses, Links: append([]string{}, links...)}
}

func TestFind(t *testing.T) {
	cfg := defaults(t)
	cfg.Synapses["see"] = config.Synapse{Inverse: "gone", List: true}
	tests := []struct {
		name    string
		engrams map[string]Entry
		want    []string // kind engram field target text, sorted
	}{
		{
			name: "all well",
			engrams: map[string]Entry{
				"a.md": entry(map[string][]string{"down": {"b.md"}}, "b.md"),
				"b.md": entry(map[string][]string{"up": {"a.md"}}),
			},
		},
		{
			name:    "broken synapse",
			engrams: map[string]Entry{"a.md": entry(map[string][]string{"down": {"x.md"}})},
			want:    []string{"broken_synapse a.md down x.md broken synapse: down → x.md"},
		},
		{
			name: "missing inverse",
			engrams: map[string]Entry{
				"a.md": entry(map[string][]string{"down": {"b.md"}}),
				"b.md": entry(nil),
			},
			want: []string{"missing_inverse a.md down b.md missing inverse: b.md has no up → a.md"},
		},
		{
			name: "inverse not configured, or field added by hand",
			engrams: map[string]Entry{
				"a.md": entry(map[string][]string{"see": {"b.md"}, "hand": {"b.md", "x.md"}}),
				"b.md": entry(nil),
			},
			want: []string{"broken_synapse a.md hand x.md broken synapse: hand → x.md"},
		},
		{
			name:    "broken link",
			engrams: map[string]Entry{"a.md": entry(nil, "x.md", "a.md")},
			want:    []string{"broken_link a.md  x.md broken link: x.md"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := find(Atlas{Engrams: tt.engrams}, cfg)
			if got == nil {
				t.Fatal("find = nil, want a list")
			}
			var lines []string
			for _, p := range got {
				lines = append(lines, strings.Join([]string{p.Kind, p.Engram, p.field, p.target, p.Text}, " "))
			}
			slices.Sort(lines)
			if !slices.Equal(lines, tt.want) {
				t.Errorf("find =\n%q\nwant\n%q", lines, tt.want)
			}
		})
	}
}

func TestCheck(t *testing.T) {
	cfg := defaults(t)
	b := newBrain(t, map[string]string{
		"a.md": "<!-- SYNAPSES -->\n- **down:** [b](b.md), [x](x.md)\n- **up:**\n***\n<!-- /SYNAPSES -->\n\n# A\n" +
			"```\n[x](x.md)\n```\nsee [x](x.md) and [y](y.md)\n",
		"b.md": "# B\n",
	})
	a, err := Rebuild(b, cfg)
	if err != nil {
		t.Fatal(err)
	}
	got := Check(b, a, cfg)
	path := filepath.Join(b.Location, "a.md")
	want := []Problem{
		{Brain: "work", Kind: BrokenSynapse, Engram: "a.md", File: path, Line: 2, Text: "broken synapse: down → x.md", field: "down", target: "x.md"},
		{Brain: "work", Kind: MissingInverse, Engram: "a.md", File: path, Line: 2, Text: "missing inverse: b.md has no up → a.md", field: "down", target: "b.md"},
		{Brain: "work", Kind: BrokenLink, Engram: "a.md", File: path, Line: 11, Text: "broken link: x.md", target: "x.md"},
		{Brain: "work", Kind: BrokenLink, Engram: "a.md", File: path, Line: 11, Text: "broken link: y.md", target: "y.md"},
	}
	slices.SortStableFunc(got, func(x, y Problem) int { return strings.Compare(x.Text, y.Text) })
	slices.SortStableFunc(want, func(x, y Problem) int { return strings.Compare(x.Text, y.Text) })
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Check =\n%+v\nwant\n%+v", got, want)
	}
}

func TestFix(t *testing.T) {
	cfg := defaults(t)
	cfg.Synapses["see"] = config.Synapse{List: true, ShowEmpty: true}
	b := newBrain(t, map[string]string{
		"a.md": "<!-- SYNAPSES -->\n- **down:** [b](b.md), [x](x.md)\n***\n<!-- /SYNAPSES -->\n\n# A\n",
		"b.md": "# B\n",
		"c.md": "# C, no block\n",
	})
	a, err := Rebuild(b, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if errs := Fix(b, cfg, Check(b, a, cfg)); len(errs) > 0 {
		t.Fatal(errs)
	}
	read := func(f string) string {
		data, err := os.ReadFile(filepath.Join(b.Location, f))
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	if got, want := read("a.md"), "<!-- SYNAPSES -->\n- **down:** [b](b.md), [x](x.md)\n- **see:**\n- **up:**\n***\n<!-- /SYNAPSES -->\n\n# A\n"; got != want {
		t.Errorf("a.md =\n%s\nwant\n%s", got, want)
	}
	if got := read("b.md"); !strings.Contains(got, "- **up:** [a](a.md)") {
		t.Errorf("b.md = %q, want the inverse written", got)
	}
	if got := read("c.md"); got != "# C, no block\n" {
		t.Errorf("c.md changed: %q", got)
	}
	a, err = Rebuild(b, cfg)
	if err != nil {
		t.Fatal(err)
	}
	left := Check(b, a, cfg)
	if len(left) != 1 || left[0].Kind != BrokenSynapse {
		t.Errorf("left after Fix = %+v, want the broken synapse alone", left)
	}
}
