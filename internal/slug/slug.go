// Package slug turns text into the words memoria uses as filenames and
// concept keys: lowercase ASCII letters, digits and single underscores.
package slug

import (
	"strings"
)

// latin transliterates the lowercase Latin letters slugs know to ASCII.
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

func transliterate(r rune) string {
	if val, ok := latin[r]; ok {
		return val
	}
	if isAlphanumeric(r) {
		return string(r)
	}
	return "_"
}

// Make returns the slug of text, or "" when nothing of it survives.
func Make(text string) string {
	s := strings.TrimSpace(strings.ToLower(text))
	prev := ""

	var out strings.Builder
	out.Grow(len(s))

	for _, r := range s {
		cur := transliterate(r)
		if cur == "_" && cur == prev {
			continue
		}
		out.WriteString(cur)
		prev = cur
	}

	return strings.Trim(out.String(), "_")
}
