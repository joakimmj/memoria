package markdown

const frontmatterDelimiter = "---"

// FrontmatterEnd returns the index after the closing "---", or 0 if none.
func FrontmatterEnd(lines []string) int {
	if len(lines) == 0 || trimmed(lines[0]) != frontmatterDelimiter {
		return 0
	}
	for i := 1; i < len(lines); i++ {
		if trimmed(lines[i]) == frontmatterDelimiter {
			return i + 1
		}
	}
	return 0
}
