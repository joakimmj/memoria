package engram

import (
	"maps"
	"slices"
	"testing"

	"github.com/joakimmj/memoria/internal/config"
	"github.com/joakimmj/memoria/internal/markdown"
)

func TestFormatBlock(t *testing.T) {
	shown := config.Synapse{List: true, ShowEmpty: true}
	hidden := config.Synapse{List: true}
	tests := []struct {
		name   string
		fields map[string]config.Synapse
		values map[string][]string
		want   []string
	}{
		{
			name:   "empty fields, sorted",
			fields: map[string]config.Synapse{"up": shown, "down": shown},
			want:   []string{"- **down:**", "- **up:**", "***"},
		},
		{
			name:   "values as links in order",
			fields: map[string]config.Synapse{"down": shown},
			values: map[string][]string{"down": {"b.md", "a.md"}},
			want:   []string{"- **down:** [b](b.md), [a](a.md)", "***"},
		},
		{
			name:   "show_empty false hides an empty field",
			fields: map[string]config.Synapse{"up": shown, "next": hidden},
			want:   []string{"- **up:**", "***"},
		},
		{
			name:   "show_empty false still shows values",
			fields: map[string]config.Synapse{"next": hidden},
			values: map[string][]string{"next": {"n.md"}},
			want:   []string{"- **next:** [n](n.md)", "***"},
		},
		{
			name:   "hand-added field kept, sorted in",
			fields: map[string]config.Synapse{"down": shown, "up": shown},
			values: map[string][]string{"see": {"s.md"}, "ask": nil},
			want:   []string{"- **ask:**", "- **down:**", "- **see:** [s](s.md)", "- **up:**", "***"},
		},
		{
			name:   "nothing to show is no block",
			fields: map[string]config.Synapse{"next": hidden},
		},
		{name: "no fields is no block"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBlock(tt.fields, tt.values)
			if !slices.Equal(got, tt.want) || (tt.want == nil) != (got == nil) {
				t.Errorf("formatBlock =\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}

func TestParseBlock(t *testing.T) {
	tests := []struct {
		name string
		body []string
		want map[string][]string
	}{
		{name: "empty block", body: []string{"***"}, want: map[string][]string{}},
		{
			name: "fields in order",
			body: []string{"- **down:** [b](b.md), [a](a.md)", "- **up:**", "***"},
			want: map[string][]string{"down": {"b.md", "a.md"}, "up": {}},
		},
		{
			name: "marker dropped",
			body: []string{"- [ ] **down:** [a](a.md)", "- [x] **up:**"},
			want: map[string][]string{"down": {"a.md"}, "up": {}},
		},
		{
			name: "not field items ignored",
			body: []string{"some text", "- plain item", "- **bad label:** [a](a.md)", "* **down:** [a](a.md)", "  - **up:** [u](u.md)  "},
			want: map[string][]string{"up": {"u.md"}},
		},
		{
			name: "text that is not a link ignored",
			body: []string{"- **down:** [a](a.md), loose text, `[c](c.md)`"},
			want: map[string][]string{"down": {"a.md"}},
		},
		{
			name: "links that are not engrams ignored, targets normalised",
			body: []string{"- **down:** [a](./a.md), [s](https://x.org), [n](sub/n.md), [h](a.md#top), [b](b.md#x)"},
			want: map[string][]string{"down": {"a.md", "b.md"}},
		},
		{
			name: "label twice merges without duplicates",
			body: []string{"- **down:** [a](a.md)", "- **down:** [b](b.md), [a](a.md)"},
			want: map[string][]string{"down": {"a.md", "b.md"}},
		},
		{name: "no block", body: nil, want: map[string][]string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseBlock(tt.body)
			if !maps.EqualFunc(got, tt.want, slices.Equal) {
				t.Errorf("parseBlock = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBlockRoundTrip(t *testing.T) {
	fields := map[string]config.Synapse{"down": {ShowEmpty: true}, "up": {ShowEmpty: true}}
	values := map[string][]string{"down": {"b.md", "a.md"}, "up": {}, "see": {"s.md"}}
	if got := parseBlock(formatBlock(fields, values)); !maps.EqualFunc(got, values, slices.Equal) {
		t.Errorf("parseBlock(formatBlock(v)) = %q, want %q", got, values)
	}
}

func TestWriteSynapses(t *testing.T) {
	fields := map[string]config.Synapse{"down": {ShowEmpty: true}, "up": {ShowEmpty: true}}
	values := map[string][]string{"down": {"a.md"}}
	const block = "<!-- SYNAPSES -->\n- **down:** [a](a.md)\n- **up:**\n***\n<!-- /SYNAPSES -->\n"
	tests := []struct {
		name, content, want string
		fields              map[string]config.Synapse
	}{
		{name: "replace", content: "<!-- SYNAPSES -->\n- **up:**\n***\n<!-- /SYNAPSES -->\n\n# A\n", want: block + "\n# A\n"},
		{name: "new, no frontmatter", content: "# A\n", want: block + "\n# A\n"},
		{name: "new, blank line already there", content: "\n# A\n", want: block + "\n# A\n"},
		{name: "new, under frontmatter", content: "---\nk: v\n---\n# A\n", want: "---\nk: v\n---\n" + block + "\n# A\n"},
		{name: "new, empty file", content: "", want: block},
		{name: "new, no final newline", content: "# A", want: block + "\n# A"},
		{name: "replace keeps the rest", content: "---\nk: v\n---\n<!-- SYNAPSES -->\nold\n<!-- /SYNAPSES -->\n\n# A\ntext\n", want: "---\nk: v\n---\n" + block + "\n# A\ntext\n"},
		{name: "nothing to show leaves the file", content: "# A\n", fields: map[string]config.Synapse{}, want: "# A\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fields
			v := values
			if tt.fields != nil {
				f, v = tt.fields, nil
			}
			lines := markdown.Lines(tt.content)
			before := slices.Clone(lines)
			if got := markdown.Join(writeSynapses(lines, f, v)); got != tt.want {
				t.Errorf("writeSynapses =\n%q\nwant\n%q", got, tt.want)
			}
			if !slices.Equal(lines, before) {
				t.Errorf("lines modified: %q", lines)
			}
		})
	}
}
