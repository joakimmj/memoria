// Package markdown edits markdown line by line, leaving the rest untouched.
package markdown

import "strings"

// Lines splits content at "\n"; a final newline gives a last "".
func Lines(content string) []string {
	return strings.Split(content, "\n")
}

// Join is the inverse of Lines.
func Join(lines []string) string {
	return strings.Join(lines, "\n")
}

// A "\r" left by a CRLF file is not part of a marker.
func trimmed(line string) string {
	return strings.TrimSpace(line)
}
