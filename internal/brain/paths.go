package brain

import "path/filepath"

// SamePath reports whether two absolute paths name the same location. Every
// comparison of brain locations goes through it.
func SamePath(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}
