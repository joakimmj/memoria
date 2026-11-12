package config

import (
	"encoding/json"
	"reflect"
	"testing"
)

func tree(t *testing.T, s string) any {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("bad test JSON %s: %v", s, err)
	}
	return v
}

func TestMerge(t *testing.T) {
	defaults := `{
		"engrams": {"date_format": "YYYY-MM-DD", "filename": {"prefix": "date", "date_format": "YYYY-MM-DD"}},
		"markers": ["[x]", "[X]"]
	}`
	tests := []struct {
		name       string
		base, tier string
		want       string
	}{
		{
			name: "empty tier keeps base",
			base: defaults,
			tier: `{}`,
			want: defaults,
		},
		{
			name: "scalar replaces",
			base: `{"engrams": {"date_format": "YYYY-MM-DD"}}`,
			tier: `{"engrams": {"date_format": "DD.MM.YYYY"}}`,
			want: `{"engrams": {"date_format": "DD.MM.YYYY"}}`,
		},
		{
			name: "maps merge by key",
			base: `{"engrams": {"filename": {"prefix": "date", "date_format": "YYYY-MM-DD"}}}`,
			tier: `{"engrams": {"filename": {"prefix": "none"}}}`,
			want: `{"engrams": {"filename": {"prefix": "none", "date_format": "YYYY-MM-DD"}}}`,
		},
		{
			name: "list replaces wholesale",
			base: `{"markers": ["[x]", "[X]"]}`,
			tier: `{"markers": ["[x]"]}`,
			want: `{"markers": ["[x]"]}`,
		},
		{
			name: "null on a setting restores the built-in default, skipping the user tier",
			base: `{"engrams": {"date_format": "DD.MM.YYYY"}}`,
			tier: `{"engrams": {"date_format": null}}`,
			want: `{"engrams": {"date_format": "YYYY-MM-DD"}}`,
		},
		{
			name: "null on a whole map restores the default map",
			base: `{"engrams": {"filename": {"prefix": "none"}}}`,
			tier: `{"engrams": {"filename": null}}`,
			want: `{"engrams": {"filename": {"prefix": "date", "date_format": "YYYY-MM-DD"}}}`,
		},
		{
			name: "key only in tier is added",
			base: `{"engrams": {"date_format": "YYYY-MM-DD"}}`,
			tier: `{"engrams": {"filename": {"prefix": "none"}}}`,
			want: `{"engrams": {"date_format": "YYYY-MM-DD", "filename": {"prefix": "none"}}}`,
		},
		{
			name: "null with no default removes the key",
			base: `{"extra": {"a": 1, "b": 2}}`,
			tier: `{"extra": {"a": null}}`,
			want: `{"extra": {"b": 2}}`,
		},
		{
			name: "scalar replaces a map",
			base: `{"engrams": {"filename": {"prefix": "date"}}}`,
			tier: `{"engrams": {"filename": "x"}}`,
			want: `{"engrams": {"filename": "x"}}`,
		},
		{
			name: "null inside a map that replaces a scalar takes the default",
			base: `{"engrams": {"filename": "x"}}`,
			tier: `{"engrams": {"filename": {"prefix": null}}}`,
			want: `{"engrams": {"filename": {"prefix": "date"}}}`,
		},
		{
			name: "map replaces a scalar",
			base: `{"engrams": {"filename": "x"}}`,
			tier: `{"engrams": {"filename": {"prefix": "none"}}}`,
			want: `{"engrams": {"filename": {"prefix": "none"}}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := merge(tree(t, tt.base), tree(t, tt.tier), tree(t, defaults))
			if want := tree(t, tt.want); !reflect.DeepEqual(got, want) {
				g, _ := json.Marshal(got)
				w, _ := json.Marshal(want)
				t.Errorf("merge =\n  %s\nwant\n  %s", g, w)
			}
		})
	}
}

func TestMergeDoesNotAlias(t *testing.T) {
	base := tree(t, `{"engrams": {"filename": {"prefix": "date"}}, "markers": ["[x]"]}`)
	tier := tree(t, `{"extra": {"x": true}, "list": ["a"]}`)
	defaults := tree(t, `{"engrams": {"date_format": "YYYY-MM-DD"}}`)
	nullTier := tree(t, `{"engrams": {"date_format": null}}`)

	got := merge(base, tier, defaults).(map[string]any)
	got["engrams"].(map[string]any)["filename"].(map[string]any)["prefix"] = "changed"
	got["markers"].([]any)[0] = "changed"
	got["extra"].(map[string]any)["x"] = false
	got["list"].([]any)[0] = "changed"

	restored := merge(base, nullTier, defaults).(map[string]any)
	restored["engrams"].(map[string]any)["date_format"] = "changed"

	want := `{"engrams": {"filename": {"prefix": "date"}}, "markers": ["[x]"]}`
	if !reflect.DeepEqual(base, tree(t, want)) {
		t.Errorf("base was modified: %v", base)
	}
	if !reflect.DeepEqual(tier, tree(t, `{"extra": {"x": true}, "list": ["a"]}`)) {
		t.Errorf("tier was modified: %v", tier)
	}
	if !reflect.DeepEqual(defaults, tree(t, `{"engrams": {"date_format": "YYYY-MM-DD"}}`)) {
		t.Errorf("defaults were modified: %v", defaults)
	}
}
