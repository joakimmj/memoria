package markdown

import "testing"

func TestLinesRoundTrip(t *testing.T) {
	for _, s := range []string{"", "a", "a\n", "a\nb", "a\n\n", "a\r\nb\r\n", "\n"} {
		if got := Join(Lines(s)); got != s {
			t.Errorf("Join(Lines(%q)) = %q", s, got)
		}
	}
}

func TestFrontmatterEnd(t *testing.T) {
	tests := []struct {
		name, content string
		want          int
	}{
		{"none", "# Title\n", 0},
		{"empty file", "", 0},
		{"closed", "---\na: 1\n---\n# Title\n", 3},
		{"empty frontmatter", "---\n---\n", 2},
		{"CRLF", "---\r\na: 1\r\n---\r\n", 3},
		{"never closed", "---\na: 1\n", 0},
		{"not on line 1", "\n---\na: 1\n---\n", 0},
	}
	for _, tt := range tests {
		if got := FrontmatterEnd(Lines(tt.content)); got != tt.want {
			t.Errorf("%s: FrontmatterEnd = %d, want %d", tt.name, got, tt.want)
		}
	}
}
