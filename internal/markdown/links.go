package markdown

import (
	"regexp"
	"strings"
)

// Link is a markdown link, [Text](Dest).
type Link struct {
	Text, Dest string
}

var scheme = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]*:`)

// EngramTarget returns the engram a link destination names: a .md file in the
// same folder, without "./" or "#…". ok is false for any other link.
func EngramTarget(dest string) (file string, ok bool) {
	if scheme.MatchString(dest) {
		return "", false
	}
	file, _, _ = strings.Cut(dest, "#")
	file = strings.TrimPrefix(file, "./")
	if strings.Contains(file, "/") || !strings.HasSuffix(file, ".md") || file == ".md" {
		return "", false
	}
	return file, true
}

// Links returns the links on line, skipping images and code spans.
func Links(line string) []Link {
	var out []Link
	for i := 0; i < len(line); {
		switch {
		case line[i] == '`':
			i = afterCodeSpan(line, i)
		case line[i] == '!' && strings.HasPrefix(line[i+1:], "["):
			if _, end, ok := linkAt(line, i+1); ok {
				i = end
			} else {
				i++
			}
		case line[i] == '[':
			if l, end, ok := linkAt(line, i); ok {
				out = append(out, l)
				i = end
			} else {
				i++
			}
		default:
			i++
		}
	}
	return out
}

func linkAt(line string, i int) (l Link, end int, ok bool) {
	textEnd := strings.IndexByte(line[i+1:], ']') + i + 1
	if textEnd == i || strings.ContainsRune(line[i+1:textEnd], '[') {
		return Link{}, 0, false
	}
	if !strings.HasPrefix(line[textEnd+1:], "(") {
		return Link{}, 0, false
	}
	destStart := textEnd + 2
	destEnd := strings.IndexByte(line[destStart:], ')') + destStart
	if destEnd < destStart {
		return Link{}, 0, false
	}
	return Link{Text: line[i+1 : textEnd], Dest: strings.TrimSpace(line[destStart:destEnd])}, destEnd + 1, true
}

// A code span closes at the next backtick run of the same length; unclosed,
// the backticks are literal.
func afterCodeSpan(line string, i int) int {
	n := backticks(line, i)
	for j := i + n; j < len(line); {
		k := strings.IndexByte(line[j:], '`')
		if k < 0 {
			break
		}
		j += k
		m := backticks(line, j)
		if m == n {
			return j + m
		}
		j += m
	}
	return i + n
}

func backticks(line string, i int) int {
	n := 0
	for i+n < len(line) && line[i+n] == '`' {
		n++
	}
	return n
}
