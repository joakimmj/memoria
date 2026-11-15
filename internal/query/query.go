// Package query answers questions over brains' atlases: listings, single
// engrams, tasks.
package query

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/joakimmj/memoria/internal/atlas"
	"github.com/joakimmj/memoria/internal/brain"
	"github.com/joakimmj/memoria/internal/config"
	"github.com/joakimmj/memoria/internal/engram"
	"github.com/joakimmj/memoria/internal/markdown"
)

// Source is a brain and its atlas, brought up to date.
type Source struct {
	Brain brain.Brain
	Atlas atlas.Atlas
}

// Summary is one engram in a listing.
type Summary struct {
	Brain    string    `json:"brain"`
	File     string    `json:"file"`
	Title    string    `json:"title"`
	Modified time.Time `json:"modified"`
}

// Engrams lists every engram of sources, newest first.
func Engrams(sources []Source) []Summary {
	out := []Summary{}
	for _, s := range sources {
		for file, e := range s.Atlas.Engrams {
			out = append(out, Summary{Brain: s.Brain.Name, File: file, Title: e.Title, Modified: e.Modified})
		}
	}
	slices.SortFunc(out, func(a, b Summary) int {
		return cmp.Or(b.Modified.Compare(a.Modified), cmp.Compare(a.Brain, b.Brain), cmp.Compare(a.File, b.File))
	})
	return out
}

// Detail is one engram in full.
type Detail struct {
	Brain              string      `json:"brain"`
	File               string      `json:"file"`
	Path               string      `json:"path"`
	Entry              atlas.Entry `json:"entry"`
	Backlinks          []string    `json:"backlinks"`
	Content            string      `json:"content"`
	DateFormat         string      `json:"date_format"`
	FilenameDateFormat string      `json:"filename_date_format"`
}

// Engram answers ref from its brain's atlas and its file. With bodyOnly the
// content leaves out the header.
func Engram(ref brain.Ref, a atlas.Atlas, cfg config.Engrams, bodyOnly bool) (Detail, error) {
	entry, ok := a.Engrams[ref.File]
	if !ok {
		return Detail{}, fmt.Errorf("(%s) %s: %w", ref.Brain.Name, ref.File, brain.ErrNoEngram)
	}
	data, err := os.ReadFile(ref.Path())
	if err != nil {
		return Detail{}, fmt.Errorf("(%s) %w", ref.Brain.Name, err)
	}
	content := string(data)
	if bodyOnly {
		content = markdown.Join(engram.Body(markdown.Lines(content)))
	}
	backlinks := a.Backlinks[ref.File]
	if backlinks == nil {
		backlinks = []string{}
	}
	return Detail{
		Brain:              ref.Brain.Name,
		File:               ref.File,
		Path:               ref.Path(),
		Entry:              entry,
		Backlinks:          backlinks,
		Content:            content,
		DateFormat:         cfg.DateFormat,
		FilenameDateFormat: cfg.Filename.DateFormat,
	}, nil
}

// Filter narrows Tasks; a field left empty filters nothing.
type Filter struct {
	State  string
	Engram string // a filename, in every source
}

// Task is one task, with where it is.
type Task struct {
	Brain  string `json:"brain"`
	Engram string `json:"engram"`
	Line   int    `json:"line"`
	Text   string `json:"text"`
	State  string `json:"state"`
}

// TaskList is the answer of Tasks; Done and Total count the tasks answered.
type TaskList struct {
	Tasks []Task `json:"tasks"`
	Done  int    `json:"done"`
	Total int    `json:"total"`
}

// Tasks lists the tasks of sources that pass f: by brain, then engram newest
// first, then line.
func Tasks(sources []Source, f Filter) TaskList {
	type dated struct {
		Task
		modified time.Time
	}
	var found []dated
	for _, s := range sources {
		for state, tasks := range map[string][]atlas.Task{config.NotDone: s.Atlas.Tasks.NotDone, config.Done: s.Atlas.Tasks.Done} {
			for _, t := range tasks {
				if (f.State != "" && f.State != state) || (f.Engram != "" && f.Engram != t.Engram) {
					continue
				}
				task := Task{Brain: s.Brain.Name, Engram: t.Engram, Line: t.Line, Text: t.Text, State: state}
				found = append(found, dated{task, s.Atlas.Engrams[t.Engram].Modified})
			}
		}
	}
	slices.SortFunc(found, func(a, b dated) int {
		return cmp.Or(
			cmp.Compare(a.Brain, b.Brain),
			b.modified.Compare(a.modified),
			cmp.Compare(a.Engram, b.Engram),
			cmp.Compare(a.Line, b.Line),
		)
	})
	out := TaskList{Tasks: make([]Task, len(found)), Total: len(found)}
	for i, d := range found {
		out.Tasks[i] = d.Task
		if d.State == config.Done {
			out.Done++
		}
	}
	return out
}
