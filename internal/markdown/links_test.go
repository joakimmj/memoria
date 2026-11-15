package markdown

import (
	"slices"
	"testing"
)

func TestEngramTarget(t *testing.T) {
	tests := []struct {
		dest, want string
		wantOK     bool
	}{
		{"a.md", "a.md", true},
		{"./a.md", "a.md", true},
		{"a.md#heading", "a.md", true},
		{"2026-08-01_x.md", "2026-08-01_x.md", true},
		{"sub/a.md", "", false},
		{"../a.md", "", false},
		{"/abs/a.md", "", false},
		{"https://x.org/a.md", "", false},
		{"mailto:a.md", "", false},
		{"a.txt", "", false},
		{"#heading", "", false},
		{".md", "", false},
		{"", "", false},
	}
	for _, tt := range tests {
		got, ok := EngramTarget(tt.dest)
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("EngramTarget(%q) = %q, %v; want %q, %v", tt.dest, got, ok, tt.want, tt.wantOK)
		}
	}
}

func TestLinks(t *testing.T) {
	tests := []struct {
		name, line string
		want       []Link
	}{
		{name: "none", line: "plain text"},
		{name: "one", line: "see [a](a.md).", want: []Link{{"a", "a.md"}}},
		{name: "several", line: "[a](a.md), [b](b.md)", want: []Link{{"a", "a.md"}, {"b", "b.md"}}},
		{name: "dest trimmed", line: "[a]( a.md )", want: []Link{{"a", "a.md"}}},
		{name: "empty text", line: "[](a.md)", want: []Link{{"", "a.md"}}},
		{name: "image skipped", line: "![pic](p.png) [a](a.md)", want: []Link{{"a", "a.md"}}},
		{name: "code span skipped", line: "`[a](a.md)` [b](b.md)", want: []Link{{"b", "b.md"}}},
		{name: "double backtick span", line: "``x ` [a](a.md)`` [b](b.md)", want: []Link{{"b", "b.md"}}},
		{name: "unclosed backtick is literal", line: "` [a](a.md)", want: []Link{{"a", "a.md"}}},
		{name: "brackets without dest", line: "[a] (a.md) [b]", want: nil},
		{name: "unclosed dest", line: "[a](a.md", want: nil},
		{name: "inner bracket starts over", line: "[x [a](a.md)", want: []Link{{"a", "a.md"}}},
		{name: "task marker is not a link", line: "- [x] [a](a.md)", want: []Link{{"a", "a.md"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Links(tt.line); !slices.Equal(got, tt.want) {
				t.Errorf("Links(%q) = %q, want %q", tt.line, got, tt.want)
			}
		})
	}
}
