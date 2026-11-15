package markdown

import "strings"

// FencedCode marks the lines of fenced code blocks, fences included. A fence
// closes at a run of its character at least as long, or at the end of file.
func FencedCode(lines []string) []bool {
	in := make([]bool, len(lines))
	var open string
	for i, line := range lines {
		t := trimmed(line)
		run := fenceRun(t)
		switch {
		case open == "" && run != "":
			open = run
			in[i] = true
		case open != "":
			in[i] = true
			if strings.HasPrefix(run, open) && run == t {
				open = ""
			}
		}
	}
	return in
}

func fenceRun(line string) string {
	for _, c := range "`~" {
		run := line[:len(line)-len(strings.TrimLeft(line, string(c)))]
		if len(run) >= 3 {
			return run
		}
	}
	return ""
}
