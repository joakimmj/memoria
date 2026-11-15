package engram

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/joakimmj/memoria/internal/brain"
	"github.com/joakimmj/memoria/internal/config"
	"github.com/joakimmj/memoria/internal/markdown"
)

// Errors attaching a synapse.
var (
	ErrNoField = errors.New("not a synapse field")
	ErrSelf    = errors.New("an engram cannot be attached to itself")
)

// Attached is the answer of Attach.
type Attached struct {
	Brain  string `json:"brain"`
	Source string `json:"source"`
	Field  string `json:"field"`
	Target string `json:"target"`
}

// Attach adds target to source's field, and source to target's inverse
// field. Values already there are left alone, so attaching twice is a no-op.
func Attach(cfg config.Config, source brain.Ref, field string, target brain.Ref) (Attached, error) {
	b := source.Brain
	if _, ok := cfg.Synapses[field]; !ok {
		return Attached{}, fmt.Errorf("(%s) %s: %w", b.Name, field, ErrNoField)
	}
	if source.File == target.File {
		return Attached{}, fmt.Errorf("(%s) %s: %w", b.Name, source.File, ErrSelf)
	}
	s := &session{brain: b, fields: cfg.Synapses, docs: map[string]*doc{}}
	for _, file := range []string{source.File, target.File} {
		if _, err := s.load(file); err != nil {
			return Attached{}, err
		}
	}
	if err := s.attach(source.File, field, target.File); err != nil {
		return Attached{}, err
	}
	if err := s.save(); err != nil {
		return Attached{}, fmt.Errorf("(%s) %w", b.Name, err)
	}
	return Attached{Brain: b.Name, Source: source.Path(), Field: field, Target: target.Path()}, nil
}

func (s *session) attach(source, field, target string) error {
	if err := s.link(source, field, target); err != nil {
		return err
	}
	if inverse, ok := s.inverse(field); ok {
		return s.link(target, inverse, source)
	}
	return nil
}

// A single-value field gives up its old values; each loses its link back.
func (s *session) link(from, field, to string) error {
	if s.fields[field].List {
		return s.add(from, field, to)
	}
	replaced, err := s.set(from, field, to)
	if err != nil {
		return err
	}
	inverse, ok := s.inverse(field)
	if !ok {
		return nil
	}
	for _, old := range replaced {
		if err := s.remove(old, inverse, from); err != nil {
			return err
		}
	}
	return nil
}

func (s *session) inverse(field string) (string, bool) {
	inverse := s.fields[field].Inverse
	_, configured := s.fields[inverse]
	return inverse, inverse != "" && configured
}

type doc struct {
	lines    []string
	synapses map[string][]string
	changed  bool
}

// session holds every file touched, so all are read before any is written.
type session struct {
	brain  brain.Brain
	fields map[string]config.Synapse
	docs   map[string]*doc
}

func (s *session) load(file string) (*doc, error) {
	if d, ok := s.docs[file]; ok {
		return d, nil
	}
	data, err := os.ReadFile(filepath.Join(s.brain.Location, file))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("(%s) %s: %w", s.brain.Name, file, brain.ErrNoEngram)
	}
	if err != nil {
		return nil, fmt.Errorf("(%s) %w", s.brain.Name, err)
	}
	lines := markdown.Lines(string(data))
	d := &doc{lines: lines, synapses: readSynapses(lines)}
	s.docs[file] = d
	return d, nil
}

func (s *session) add(file, field, value string) error {
	d, err := s.load(file)
	if err != nil {
		return err
	}
	if !slices.Contains(d.synapses[field], value) {
		d.synapses[field] = append(d.synapses[field], value)
		d.changed = true
	}
	return nil
}

// set returns the values it replaced; a value already there changes nothing.
func (s *session) set(file, field, value string) ([]string, error) {
	d, err := s.load(file)
	if err != nil {
		return nil, err
	}
	old := d.synapses[field]
	if slices.Contains(old, value) {
		return nil, nil
	}
	d.synapses[field] = []string{value}
	d.changed = true
	return old, nil
}

// A file that is gone has nothing to remove.
func (s *session) remove(file, field, value string) error {
	d, err := s.load(file)
	if errors.Is(err, brain.ErrNoEngram) {
		return nil
	}
	if err != nil {
		return err
	}
	if i := slices.Index(d.synapses[field], value); i >= 0 {
		d.synapses[field] = slices.Delete(slices.Clone(d.synapses[field]), i, i+1)
		d.changed = true
	}
	return nil
}

func (s *session) save() error {
	for _, file := range slices.Sorted(maps.Keys(s.docs)) {
		d := s.docs[file]
		if !d.changed {
			continue
		}
		lines := writeSynapses(d.lines, s.fields, d.synapses)
		if err := replaceFile(filepath.Join(s.brain.Location, file), markdown.Join(lines)); err != nil {
			return err
		}
	}
	return nil
}
