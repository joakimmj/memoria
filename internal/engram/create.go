// Package engram creates engrams, the markdown files of a brain.
package engram

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joakimmj/memoria/internal/brain"
	"github.com/joakimmj/memoria/internal/config"
)

// Errors a caller can act on; wrapped with the brain's name.
var (
	ErrEmptySlug = errors.New("the title needs a letter or digit")
	ErrCollision = errors.New("already exists")
	ErrPrefix    = errors.New("filename prefix not supported")
	ErrNoTitle   = errors.New("a title is required")
	ErrTemplate  = errors.New("content_template is not valid")
)

// Options are what a new engram is made from.
type Options struct {
	Title string
	Body  string
}

// Position is a place in a file: Line from 1, Col in bytes from 0.
type Position struct {
	Line int `json:"line"`
	Col  int `json:"col"`
}

// Created is the answer of Create.
type Created struct {
	Brain   string   `json:"brain"`
	Path    string   `json:"path"`
	Body    Position `json:"body"`
	Created bool     `json:"created"`
}

// Create writes a new engram into b. now is the creation time.
func Create(b brain.Brain, cfg config.Engrams, opts Options, now time.Time) (Created, error) {
	title := strings.TrimSpace(opts.Title)
	name, err := filename(cfg.Filename, title, now)
	if err != nil {
		return Created{}, fmt.Errorf("(%s) %w", b.Name, err)
	}
	content, pos, err := render(cfg.ContentTemplate, vars{title: title, date: formatDate(now, cfg.DateFormat)}, opts.Body)
	if err != nil {
		return Created{}, fmt.Errorf("(%s) %w", b.Name, err)
	}
	path := filepath.Join(b.Location, name)
	if err := writeNew(path, content); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return Created{}, fmt.Errorf("(%s) %s %w", b.Name, name, ErrCollision)
		}
		return Created{}, fmt.Errorf("(%s) %w", b.Name, err)
	}
	return Created{Brain: b.Name, Path: path, Body: pos, Created: true}, nil
}

// O_EXCL fails when the file exists, with no gap between check and write.
func writeNew(path, content string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
