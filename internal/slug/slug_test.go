package slug

import (
	"testing"
	"unicode"
)

func TestMake(t *testing.T) {
	tests := []struct {
		text, want string
	}{
		{"Note about Java", "note_about_java"},
		{"  Alice   Smith  ", "alice_smith"},
		{"Mary-Jane Watson", "mary_jane_watson"},
		{"Ærlig Østers", "aerlig_osters"},
		{"proj/web", "proj_web"},
		{"Node.js", "node_js"},
		{"O'Brien", "o_brien"},
		{"C++", "c"},
		{"C#", "c"},
		{"東京", ""},
		{"???", ""},
		{"", ""},
		{"Straße", "strasse"},
		{"ÆØÅ", "aeoa"},
		{"Crème brûlée", "creme_brulee"},
		{"Łódź", "lodz"},
		{"2026-08-01 Plan", "2026_08_01_plan"},
		{"a東b", "a_b"},
		{"__x__", "x"},
		{"tab\tand\nnewline", "tab_and_newline"},
		{"Ωmega", "mega"},
		{"İSTANBUL ŸVES ẞ", "istanbul_yves_ss"},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			if got := Make(tt.text); got != tt.want {
				t.Errorf("Make(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}

// TestLatin: keys are lowercase (Make lowercases first), values are
// lowercase ASCII.
func TestLatin(t *testing.T) {
	for r, ascii := range latin {
		if unicode.ToLower(r) != r {
			t.Errorf("latin key %q is not lowercase", r)
		}
		for _, c := range ascii {
			if c < 'a' || c > 'z' {
				t.Errorf("latin[%q] = %q, not lowercase ASCII", r, ascii)
			}
		}
	}
}
