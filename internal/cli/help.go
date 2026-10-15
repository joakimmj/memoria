package cli

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/joakimmj/memoria/docs"
)

// topicPattern keeps glob metacharacters out of findPage.
var topicPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

// findPage maps a help topic to its docs file: "" → mia.1.md,
// x → mia-x.<n>.md or x.<n>.md.
func findPage(topic string) (string, error) {
	if topic == "" {
		return "mia.1.md", nil
	}
	if !topicPattern.MatchString(topic) {
		return "", notFoundErrorf("no manual page for %q", topic)
	}
	for _, pattern := range []string{"mia-" + topic + ".*.md", topic + ".*.md"} {
		matches, err := fs.Glob(docs.FS, pattern)
		if err != nil {
			return "", err
		}
		if len(matches) > 0 {
			return matches[0], nil
		}
	}
	return "", notFoundErrorf("no manual page for %q", topic)
}

// loadPage reads a page and adds its generated sections.
func loadPage(filename string) (page, error) {
	src, err := fs.ReadFile(docs.FS, filename)
	if err != nil {
		return page{}, fmt.Errorf("reading manual page: %w", err)
	}
	p, err := parsePage(filename, string(src))
	if err != nil {
		return page{}, err
	}
	addGenerated(&p)
	return p, nil
}

// writeHelp writes the page for topic to w as text.
func writeHelp(w io.Writer, topic string) error {
	filename, err := findPage(topic)
	if err != nil {
		return err
	}
	p, err := loadPage(filename)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, renderText(p))
	return err
}

func runHelp(inv Invocation) (any, error) {
	topic := ""
	if len(inv.Args) > 0 {
		topic = inv.Args[0]
	}
	return nil, writeHelp(inv.Stdout, topic)
}

// manResult is the result of `mia man`.
type manResult struct {
	Dir   string   `json:"dir"`
	Pages []string `json:"pages"` // relative to Dir
}

func runMan(inv Invocation) (any, error) {
	dir, err := filepath.Abs(inv.Flags["dir"])
	if err != nil {
		return nil, fmt.Errorf("man: %w", err)
	}
	files, err := fs.Glob(docs.FS, "*.md")
	if err != nil {
		return nil, err
	}
	res := manResult{Dir: dir, Pages: []string{}}
	for _, f := range files {
		p, err := loadPage(f)
		if err != nil {
			return nil, err
		}
		rel := filepath.Join("man"+p.section, p.name+"."+p.section)
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("man: %w", err)
		}
		if err := os.WriteFile(path, []byte(renderRoff(p)), 0o644); err != nil {
			return nil, fmt.Errorf("man: %w", err)
		}
		res.Pages = append(res.Pages, rel)
	}
	return res, nil
}

func runCommands(Invocation) (any, error) {
	return table, nil
}
