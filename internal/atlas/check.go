package atlas

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/joakimmj/memoria/internal/brain"
	"github.com/joakimmj/memoria/internal/config"
	"github.com/joakimmj/memoria/internal/engram"
	"github.com/joakimmj/memoria/internal/markdown"
)

// Problem kinds.
const (
	BrokenSynapse  = "broken_synapse"
	MissingInverse = "missing_inverse"
	BrokenLink     = "broken_link"
)

// Problem is one thing wrong in a brain, on a line of a file.
type Problem struct {
	Brain  string `json:"brain"`
	Kind   string `json:"kind"`
	Engram string `json:"engram"`
	File   string `json:"file"`
	Line   int    `json:"line"`
	Text   string `json:"text"`

	field, target string
}

// Check answers everything wrong in b, by engram, then line.
func Check(b brain.Brain, a Atlas, cfg config.Config) []Problem {
	problems := find(a, cfg)
	locate(b, problems)
	slices.SortFunc(problems, func(x, y Problem) int {
		return cmp.Or(cmp.Compare(x.Engram, y.Engram), cmp.Compare(x.Line, y.Line), cmp.Compare(x.Kind, y.Kind))
	})
	return problems
}

// find works on the atlas alone; Brain, File and Line are locate's.
func find(a Atlas, cfg config.Config) []Problem {
	problems := []Problem{}
	for file, e := range a.Engrams {
		for field, targets := range e.Synapses {
			inverse, hasInverse := inverseOf(cfg, field)
			for _, target := range targets {
				linked, exists := a.Engrams[target]
				switch {
				case !exists:
					problems = append(problems, Problem{
						Kind: BrokenSynapse, Engram: file, field: field, target: target,
						Text: fmt.Sprintf("broken synapse: %s → %s", field, target),
					})
				case hasInverse && !slices.Contains(linked.Synapses[inverse], file):
					problems = append(problems, Problem{
						Kind: MissingInverse, Engram: file, field: field, target: target,
						Text: fmt.Sprintf("missing inverse: %s has no %s → %s", target, inverse, file),
					})
				}
			}
		}
		for _, target := range e.Links {
			if _, exists := a.Engrams[target]; !exists {
				problems = append(problems, Problem{
					Kind: BrokenLink, Engram: file, target: target,
					Text: "broken link: " + target,
				})
			}
		}
	}
	return problems
}

func inverseOf(cfg config.Config, field string) (string, bool) {
	inverse := cfg.Synapses[field].Inverse
	_, configured := cfg.Synapses[inverse]
	return inverse, inverse != "" && configured
}

// locate sets each problem's file and line, reading each engram once. A line
// that cannot be found is 1.
func locate(b brain.Brain, problems []Problem) {
	lines := map[string][]string{}
	for i := range problems {
		p := &problems[i]
		p.Brain, p.File, p.Line = b.Name, filepath.Join(b.Location, p.Engram), 1
		if _, read := lines[p.Engram]; !read {
			data, _ := os.ReadFile(p.File)
			lines[p.Engram] = markdown.Lines(string(data))
		}
		if i := lineOf(lines[p.Engram], *p); i >= 0 {
			p.Line = i + 1
		}
	}
}

func lineOf(lines []string, p Problem) int {
	if p.Kind != BrokenLink {
		return engram.FieldLine(lines, p.field)
	}
	for i, inBody := range body(lines) {
		if !inBody {
			continue
		}
		for _, l := range markdown.Links(lines[i]) {
			if target, ok := markdown.EngramTarget(l.Dest); ok && target == p.target {
				return i
			}
		}
	}
	return -1
}
