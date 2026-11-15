// Package atlas is a brain's derived index, .mia_atlas.json: rebuildable from
// the engrams at any time, and brought up to date on every read.
package atlas

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/joakimmj/memoria/internal/brain"
	"github.com/joakimmj/memoria/internal/config"
	"github.com/joakimmj/memoria/internal/engram"
	"github.com/joakimmj/memoria/internal/markdown"
)

// File is the atlas's filename, at the brain's root.
const File = ".mia_atlas.json"

// Atlas is a brain's index.
type Atlas struct {
	Config    string              `json:"config"`
	Engrams   map[string]Entry    `json:"engrams"`
	Backlinks map[string][]string `json:"backlinks"`
	Tasks     Tasks               `json:"tasks"`
}

// Entry is one engram, by filename.
type Entry struct {
	Title    string              `json:"title"`
	Synapses map[string][]string `json:"synapses"`
	Links    []string            `json:"links"`
	Modified time.Time           `json:"modified"`
	Hash     string              `json:"hash"`
}

// Tasks are every task of the brain, by state.
type Tasks struct {
	NotDone []Task `json:"not_done"`
	Done    []Task `json:"done"`
}

// Task is one task line.
type Task struct {
	Engram string `json:"engram"`
	Line   int    `json:"line"`
	Text   string `json:"text"`
}

type stateTask struct {
	Task
	state string
}

// Refresh brings b's atlas up to date with its files, writing it only when
// something changed.
func Refresh(b brain.Brain, cfg config.Config) (Atlas, error) {
	return refresh(b, cfg, false)
}

// Rebuild parses every engram again, ignoring the stored atlas.
func Rebuild(b brain.Brain, cfg config.Config) (Atlas, error) {
	return refresh(b, cfg, true)
}

func refresh(b brain.Brain, cfg config.Config, full bool) (Atlas, error) {
	files, err := engramFiles(b.Location)
	if err != nil {
		return Atlas{}, fmt.Errorf("(%s) atlas: %w", b.Name, err)
	}
	old, stored := read(b.Location)
	fp := fingerprint(cfg)
	stale := full || old.Config != fp

	atlas := Atlas{Config: fp, Engrams: map[string]Entry{}}
	parsed := map[string][]stateTask{}
	changed := stale || !stored
	for _, file := range files {
		data, err := os.ReadFile(filepath.Join(b.Location, file))
		if err != nil {
			continue
		}
		h := hash(data)
		if prev, ok := old.Engrams[file]; ok && prev.Hash == h && !stale {
			atlas.Engrams[file] = prev
			continue
		}
		entry, tasks := index(file, markdown.Lines(string(data)), cfg)
		entry.Hash, entry.Modified = h, modified(filepath.Join(b.Location, file))
		atlas.Engrams[file], parsed[file] = entry, tasks
		changed = true
	}
	// Kept entries are a subset of old ones, so the same count means none vanished.
	if !changed && len(atlas.Engrams) == len(old.Engrams) {
		return old, nil
	}
	atlas.Tasks = tasks(old.Tasks, atlas.Engrams, parsed)
	atlas.Backlinks = backlinks(atlas.Engrams)
	if err := write(b.Location, atlas); err != nil {
		return Atlas{}, fmt.Errorf("(%s) atlas: %w", b.Name, err)
	}
	return atlas, nil
}

// index reads one engram. Title, links and tasks come from its body alone.
func index(file string, lines []string, cfg config.Config) (Entry, []stateTask) {
	e := Entry{
		Title:    strings.TrimSuffix(file, ".md"),
		Synapses: engram.Synapses(lines),
		Links:    []string{},
	}
	var tasks []stateTask
	titled := false
	for i, inBody := range body(lines) {
		if !inBody {
			continue
		}
		line := lines[i]
		if _, text, ok := markdown.Heading(line); ok && text != "" && !titled {
			e.Title, titled = text, true
		}
		for _, l := range markdown.Links(line) {
			if target, ok := markdown.EngramTarget(l.Dest); ok && !slices.Contains(e.Links, target) {
				e.Links = append(e.Links, target)
			}
		}
		if item, ok := markdown.ParseListItem(line); ok {
			if state := cfg.Engrams.TaskMarkers.State(item.Marker); state != "" {
				tasks = append(tasks, stateTask{Task{Engram: file, Line: i + 1, Text: item.Text}, state})
			}
		}
	}
	return e, tasks
}

// body marks the lines after the frontmatter that are outside the SYNAPSES
// block and fenced code.
func body(lines []string) []bool {
	in := markdown.FencedCode(lines)
	for i := range in {
		in[i] = !in[i]
	}
	for i := range markdown.FrontmatterEnd(lines) {
		in[i] = false
	}
	if open, close, ok := markdown.SectionBounds(lines, engram.Section); ok {
		for i := open; i <= close; i++ {
			in[i] = false
		}
	}
	return in
}

// Sorted by engram, then line.
func tasks(old Tasks, engrams map[string]Entry, parsed map[string][]stateTask) Tasks {
	keep := func(ts []Task) []Task {
		return slices.DeleteFunc(slices.Clone(ts), func(t Task) bool {
			_, kept := engrams[t.Engram]
			_, reparsed := parsed[t.Engram]
			return !kept || reparsed
		})
	}
	out := Tasks{NotDone: keep(old.NotDone), Done: keep(old.Done)}
	for _, ts := range parsed {
		for _, t := range ts {
			if t.state == config.Done {
				out.Done = append(out.Done, t.Task)
			} else {
				out.NotDone = append(out.NotDone, t.Task)
			}
		}
	}
	byPlace := func(a, b Task) int {
		if c := strings.Compare(a.Engram, b.Engram); c != 0 {
			return c
		}
		return a.Line - b.Line
	}
	slices.SortFunc(out.NotDone, byPlace)
	slices.SortFunc(out.Done, byPlace)
	if out.NotDone == nil {
		out.NotDone = []Task{}
	}
	if out.Done == nil {
		out.Done = []Task{}
	}
	return out
}

// A target is listed whether its file exists or not.
func backlinks(engrams map[string]Entry) map[string][]string {
	out := map[string][]string{}
	add := func(target, from string) {
		if !slices.Contains(out[target], from) {
			out[target] = append(out[target], from)
		}
	}
	for _, file := range slices.Sorted(maps.Keys(engrams)) {
		e := engrams[file]
		for _, field := range slices.Sorted(maps.Keys(e.Synapses)) {
			for _, target := range e.Synapses[field] {
				add(target, file)
			}
		}
		for _, target := range e.Links {
			add(target, file)
		}
	}
	return out
}

func engramFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if e.Type().IsRegular() && strings.HasSuffix(e.Name(), ".md") {
			files = append(files, e.Name())
		}
	}
	return files, nil
}

// A corrupt atlas counts as none.
func read(dir string) (Atlas, bool) {
	data, err := os.ReadFile(filepath.Join(dir, File))
	if errors.Is(err, fs.ErrNotExist) {
		return Atlas{}, false
	}
	var a Atlas
	if err != nil || json.Unmarshal(data, &a) != nil || a.Engrams == nil {
		return Atlas{}, false
	}
	return a, true
}

func write(dir string, a Atlas) error {
	data, err := json.Marshal(a)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".mia_atlas-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), filepath.Join(dir, File))
}

// A change to the config an entry is parsed with re-parses every engram.
func fingerprint(cfg config.Config) string {
	data, _ := json.Marshal(struct {
		Synapses    map[string]config.Synapse `json:"synapses"`
		TaskMarkers config.TaskMarkers        `json:"task_markers"`
	}{cfg.Synapses, cfg.Engrams.TaskMarkers})
	return hash(data)
}

func hash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])[:16]
}

func modified(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime().UTC().Truncate(time.Second)
}
