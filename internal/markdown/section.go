package markdown

import "slices"

func markers(name string) (open, close string) {
	return "<!-- " + name + " -->", "<!-- /" + name + " -->"
}

func find(lines []string, name string) (open, close int, ok bool) {
	openMarker, closeMarker := markers(name)
	is := func(marker string) func(string) bool {
		return func(line string) bool { return trimmed(line) == marker }
	}
	open = slices.IndexFunc(lines, is(openMarker))
	if open < 0 {
		return 0, 0, false
	}
	after := slices.IndexFunc(lines[open+1:], is(closeMarker))
	if after < 0 {
		return 0, 0, false
	}
	return open, open + 1 + after, true
}

// Section returns a copy of the body of section name.
func Section(lines []string, name string) ([]string, bool) {
	open, close, ok := find(lines, name)
	if !ok {
		return nil, false
	}
	return slices.Clone(lines[open+1 : close]), true
}

// SetSection returns lines with section name's body set to body. A missing
// section is inserted before index at, or at the end when at is out of range.
func SetSection(lines []string, name string, body []string, at int) []string {
	if open, close, ok := find(lines, name); ok {
		return slices.Concat(lines[:open+1], body, lines[close:])
	}
	if at < 0 || at >= len(lines) {
		at = len(lines)
		if at > 0 && lines[at-1] == "" {
			at--
		}
	}
	openMarker, closeMarker := markers(name)
	return slices.Concat(lines[:at], []string{openMarker}, body, []string{closeMarker}, lines[at:])
}
