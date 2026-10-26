// Package slug turns text into the words memoria uses as filenames and
// concept keys: lowercase ASCII letters, digits and single underscores.
package slug

import (
	"strings"
)

// latin transliterates the lowercase Latin letters slugs know to ASCII.
// Text is lowercased first, so uppercase letters need no entries. A letter
// not in it is dropped, so a slug is the same on every system.
var latin = expand(map[string]string{
	"àáâãäåāăą":  "a",
	"æ":          "ae",
	"çćĉċč":      "c",
	"ðďđ":        "d",
	"èéêëēĕėęě":  "e",
	"ĝğġģ":       "g",
	"ĥħ":         "h",
	"ìíîïĩīĭįı":  "i",
	"ĳ":          "ij",
	"ĵ":          "j",
	"ķ":          "k",
	"ĺļľŀł":      "l",
	"ñńņň":       "n",
	"òóôõöøōŏő":  "o",
	"œ":          "oe",
	"ŕŗř":        "r",
	"śŝşš":       "s",
	"ß":          "ss",
	"ţťŧ":        "t",
	"þ":          "th",
	"ùúûüũūŭůűų": "u",
	"ŵ":          "w",
	"ýÿŷ":        "y",
	"źżž":        "z",
})

// expand turns {"àá": "a"} into {'à': "a", 'á': "a"}.
func expand(groups map[string]string) map[rune]string {
	m := map[rune]string{}
	for letters, ascii := range groups {
		for _, r := range letters {
			m[r] = ascii
		}
	}
	return m
}

func isAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}

func slugify(r rune) string {
	if val, ok := latin[r]; ok {
		return val
	}
	if isAlphanumeric(r) {
		return string(r)
	}
	// if unicode.IsSpace(r) {
	// 	return "_"
	// }
	// return ""
	return "_"
}

// Make returns the slug of text, or "" when nothing of it survives.
//
// TODO(human): implement. Lowercase comes first so latin needs only
// lowercase keys:
//  1. trim surrounding whitespace; lowercase
//  2. replace each rune found in latin by its ASCII string (other runes stay)
//  3. every run of characters outside [a-z0-9] becomes one "_"
//  4. strip "_" at either end
//
// slug_test.go holds the cases. Steps 2–4 can be one pass over the runes.
//
// Hints:
//   - for _, r := range s gives runes (code points), not bytes:
//     len("ø") == 2, but the loop sees one rune 'ø'.
//   - strings.Builder for building the result:
//     https://pkg.go.dev/strings#Builder
//   - Unknown non-ASCII letters (ω) must end up outside [a-z0-9]: test the
//     range yourself, 'a' <= r && r <= 'z', not unicode.IsLetter.
//   - One "_" per run: remember whether the last thing written was "_".
func Make(text string) string {
	s := strings.TrimSpace(strings.ToLower(text))
	prev := ""

	var out strings.Builder
	out.Grow(len(s))

	for _, r := range s {
		cur := slugify(r)
		if cur == "_" && cur == prev {
			continue
		}
		out.WriteString(cur)
		prev = cur
	}

	return strings.Trim(out.String(), "_")
}
