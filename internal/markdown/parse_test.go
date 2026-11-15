package markdown

import (
	"slices"
	"testing"
)

func TestParseListItem(t *testing.T) {
	tests := []struct {
		line   string
		want   ListItem
		wantOK bool
	}{
		{"- [x] Send follow-up", ListItem{"- ", "[x]", "Send follow-up"}, true},
		{"- [ ] todo", ListItem{"- ", "[ ]", "todo"}, true},
		{"* item", ListItem{"* ", "", "item"}, true},
		{"+ item", ListItem{"+ ", "", "item"}, true},
		{"  - nested", ListItem{"  - ", "", "nested"}, true},
		{"1. first", ListItem{"1. ", "", "first"}, true},
		{"12) [x] done", ListItem{"12) ", "[x]", "done"}, true},
		{"- [x]", ListItem{"- ", "[x]", ""}, true},
		{"- [Note](note.md) matters", ListItem{"- ", "", "[Note](note.md) matters"}, true},
		{"- [x]y", ListItem{"- ", "", "[x]y"}, true},
		{"- [unclosed", ListItem{"- ", "", "[unclosed"}, true},
		{"- [a [b] c", ListItem{"- ", "", "[a [b] c"}, true},
		{"-item", ListItem{}, false},
		{"text", ListItem{}, false},
		{"1.5 apples", ListItem{}, false},
		{"", ListItem{}, false},
	}
	for _, tt := range tests {
		got, ok := ParseListItem(tt.line)
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("ParseListItem(%q) = %+v, %v; want %+v, %v", tt.line, got, ok, tt.want, tt.wantOK)
		}
	}
}

func TestListItemRoundTrip(t *testing.T) {
	for _, line := range []string{"- [x] Send", "* plain", "  1) [ ]", "- [Note](n.md) x", "+ [x]y"} {
		item, _ := ParseListItem(line)
		if got := item.String(); got != line {
			t.Errorf("ParseListItem(%q).String() = %q", line, got)
		}
	}
}

func TestHeading(t *testing.T) {
	tests := []struct {
		line      string
		wantLevel int
		wantText  string
		wantOK    bool
	}{
		{"# Title", 1, "Title", true},
		{"### Deep  ", 3, "Deep", true},
		{"###### Six", 6, "Six", true},
		{"#", 1, "", true},
		{"####### Seven", 0, "", false},
		{"#word", 0, "", false},
		{" # indented", 0, "", false},
		{"text", 0, "", false},
	}
	for _, tt := range tests {
		level, text, ok := Heading(tt.line)
		if level != tt.wantLevel || text != tt.wantText || ok != tt.wantOK {
			t.Errorf("Heading(%q) = %d, %q, %v", tt.line, level, text, ok)
		}
	}
}

func TestFencedCode(t *testing.T) {
	tests := []struct {
		name, content string
		want          []bool
	}{
		{name: "none", content: "a\nb", want: []bool{false, false}},
		{name: "backticks", content: "a\n```go\nx\n```\nb", want: []bool{false, true, true, true, false}},
		{name: "tildes", content: "~~~\nx\n~~~", want: []bool{true, true, true}},
		{name: "longer close", content: "```\nx\n`````\nb", want: []bool{true, true, true, false}},
		{name: "shorter run does not close", content: "````\n```\nx\n````\nb", want: []bool{true, true, true, true, false}},
		{name: "other char does not close", content: "```\n~~~\n```\nb", want: []bool{true, true, true, false}},
		{name: "closing fence holds only the run", content: "```\n``` x\n```\nb", want: []bool{true, true, true, false}},
		{name: "unclosed runs to the end", content: "a\n```\nx", want: []bool{false, true, true}},
		{name: "two backticks are no fence", content: "``\nx", want: []bool{false, false}},
		{name: "indented fence", content: "  ```\nx\n  ```\nb", want: []bool{true, true, true, false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FencedCode(Lines(tt.content)); !slices.Equal(got, tt.want) {
				t.Errorf("FencedCode = %v, want %v", got, tt.want)
			}
		})
	}
}
