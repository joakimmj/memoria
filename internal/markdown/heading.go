package markdown

import "strings"

// Heading reads an ATX heading: one to six "#", then a space or the end of
// the line. text is trimmed.
func Heading(line string) (level int, text string, ok bool) {
	level = len(line) - len(strings.TrimLeft(line, "#"))
	rest := line[level:]
	if level == 0 || level > 6 || (rest != "" && rest[0] != ' ' && rest[0] != '\t') {
		return 0, "", false
	}
	return level, strings.TrimSpace(rest), true
}
