package brain

import "path/filepath"

// SamePath reports whether two absolute paths name the same location. Every
// comparison of brain locations goes through it.
func SamePath(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}

// Within returns path relative to dir, and whether path lies inside dir.
func Within(dir, path string) (string, bool) {
	rel, err := filepath.Rel(dir, path)
	if err != nil || !filepath.IsLocal(rel) {
		return "", false
	}
	return rel, true
}
