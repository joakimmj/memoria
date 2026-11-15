package markdown

import (
	"regexp"
	"strings"
)

// ListItem is a list item: "- [x] Send" is Prefix "- ", Marker "[x]", Text "Send".
type ListItem struct {
	Prefix, Marker, Text string
}

var listPrefix = regexp.MustCompile(`^\s*(?:[-*+]|\d+[.)])\s+`)

// ParseListItem reads line as a list item. A bracketed run is a marker only
// when whitespace or the end of the line follows it.
func ParseListItem(line string) (ListItem, bool) {
	prefix := listPrefix.FindString(line)
	if prefix == "" {
		return ListItem{}, false
	}
	rest := line[len(prefix):]
	end := strings.IndexByte(rest, ']')
	if !strings.HasPrefix(rest, "[") || end < 0 || strings.Contains(rest[1:end], "[") {
		return ListItem{Prefix: prefix, Text: rest}, true
	}
	marker, after := rest[:end+1], rest[end+1:]
	text := strings.TrimLeft(after, " \t")
	if after != "" && text == after {
		return ListItem{Prefix: prefix, Text: rest}, true
	}
	return ListItem{Prefix: prefix, Marker: marker, Text: text}, true
}

// String writes the item back, one space between marker and text.
func (l ListItem) String() string {
	switch {
	case l.Marker == "":
		return l.Prefix + l.Text
	case l.Text == "":
		return l.Prefix + l.Marker
	default:
		return l.Prefix + l.Marker + " " + l.Text
	}
}
