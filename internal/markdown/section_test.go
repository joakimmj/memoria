package markdown

import (
	"slices"
	"testing"
)

func TestSection(t *testing.T) {
	tests := []struct {
		name, content string
		want          []string
		wantOK        bool
	}{
		{name: "none", content: "# Title\n"},
		{name: "body", content: "<!-- S -->\na\nb\n<!-- /S -->\n", want: []string{"a", "b"}, wantOK: true},
		{name: "empty", content: "<!-- S -->\n<!-- /S -->\n", want: []string{}, wantOK: true},
		{name: "markers trimmed", content: "  <!-- S -->\r\na\n\t<!-- /S -->  \n", want: []string{"a"}, wantOK: true},
		{name: "opening only", content: "<!-- S -->\na\n"},
		{name: "closing only", content: "a\n<!-- /S -->\n"},
		{name: "closing before opening is skipped", content: "<!-- /S -->\n<!-- S -->\na\n<!-- /S -->\n", want: []string{"a"}, wantOK: true},
		{name: "first section wins", content: "<!-- S -->\na\n<!-- /S -->\n<!-- S -->\nb\n<!-- /S -->\n", want: []string{"a"}, wantOK: true},
		{name: "other name", content: "<!-- SX -->\na\n<!-- /SX -->\n"},
		{name: "marker inside a line", content: "x <!-- S -->\na\n<!-- /S -->\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Section(Lines(tt.content), "S")
			if ok != tt.wantOK || !slices.Equal(got, tt.want) {
				t.Errorf("Section = %q, %v; want %q, %v", got, ok, tt.want, tt.wantOK)
			}
			if ok && got == nil {
				t.Errorf("Section of a present section is nil, want a non-nil body")
			}
		})
	}
}

func TestSectionReturnsCopy(t *testing.T) {
	lines := Lines("<!-- S -->\na\n<!-- /S -->\nafter\n")
	body, ok := Section(lines, "S")
	if !ok || len(body) != 1 {
		t.Fatalf("Section = %q, %v", body, ok)
	}
	body[0] = "changed"
	_ = append(body, "appended")
	if want := Lines("<!-- S -->\na\n<!-- /S -->\nafter\n"); !slices.Equal(lines, want) {
		t.Errorf("lines changed through Section's result: %q", lines)
	}
}

func TestSetSection(t *testing.T) {
	tests := []struct {
		name, content string
		body          []string
		at            int
		want          string
	}{
		{name: "replace", content: "x\n<!-- S -->\nold\n<!-- /S -->\ny\n", body: []string{"new1", "new2"}, want: "x\n<!-- S -->\nnew1\nnew2\n<!-- /S -->\ny\n"},
		{name: "replace with empty", content: "<!-- S -->\nold\n<!-- /S -->\n", body: nil, want: "<!-- S -->\n<!-- /S -->\n"},
		{name: "replace keeps marker lines as written", content: "  <!-- S -->\nold\n<!-- /S -->  \n", body: []string{"n"}, want: "  <!-- S -->\nn\n<!-- /S -->  \n"},
		{name: "insert at 0", content: "# T\n", body: []string{"a"}, at: 0, want: "<!-- S -->\na\n<!-- /S -->\n# T\n"},
		{name: "insert after frontmatter", content: "---\nk: v\n---\n# T\n", body: []string{"a"}, at: 3, want: "---\nk: v\n---\n<!-- S -->\na\n<!-- /S -->\n# T\n"},
		{name: "insert at end keeps final newline", content: "# T\n", body: []string{"a"}, at: -1, want: "# T\n<!-- S -->\na\n<!-- /S -->\n"},
		{name: "insert past the end", content: "# T\n", body: []string{"a"}, at: 99, want: "# T\n<!-- S -->\na\n<!-- /S -->\n"},
		{name: "insert at end without final newline", content: "# T", body: []string{"a"}, at: -1, want: "# T\n<!-- S -->\na\n<!-- /S -->"},
		{name: "insert into empty file", content: "", body: []string{"a"}, at: -1, want: "<!-- S -->\na\n<!-- /S -->\n"},
		{name: "opening only is no section", content: "<!-- S -->\nx\n", body: []string{"a"}, at: 0, want: "<!-- S -->\na\n<!-- /S -->\n<!-- S -->\nx\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := Lines(tt.content)
			before := slices.Clone(lines)
			got := Join(SetSection(lines, "S", tt.body, tt.at))
			if got != tt.want {
				t.Errorf("SetSection =\n%q\nwant\n%q", got, tt.want)
			}
			if !slices.Equal(lines, before) {
				t.Errorf("lines modified: %q", lines)
			}
			if again := Join(SetSection(Lines(got), "S", tt.body, tt.at)); again != got {
				t.Errorf("second SetSection changed the result:\n%q", again)
			}
		})
	}
}
