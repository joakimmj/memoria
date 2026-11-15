package engram

import (
	"slices"
	"testing"

	"github.com/joakimmj/memoria/internal/config"
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
