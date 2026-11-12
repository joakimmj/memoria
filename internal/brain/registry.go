// Package brain is the registry of brains: which exist on this machine, and
// where.
package brain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/joakimmj/memoria/internal/config"
)

// Errors a caller can act on; wrapped with the brain's name.
var (
	ErrCollision   = errors.New("already registered")
	ErrNotFound    = errors.New("not registered")
	ErrInvalidName = errors.New("invalid brain name")
)

// Brain is a registered brain.
type Brain struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

// Exists reports whether the brain's folder exists.
func (b Brain) Exists() bool {
	info, err := os.Stat(b.Location)
	return err == nil && info.IsDir()
}

// HasDNA reports whether the brain has its own config file.
func (b Brain) HasDNA() bool {
	_, err := os.Stat(filepath.Join(b.Location, config.DNAFile))
	return err == nil
}

// Registry is the brain registry file, loaded.
type Registry struct {
	path   string
	brains map[string]entry
}

type entry struct {
	Location string `json:"location"`
}

// Load reads the registry at path. A missing file is an empty registry.
func Load(path string) (*Registry, error) {
	r := &Registry{path: path, brains: map[string]entry{}}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return r, nil
	}
	if err != nil {
		return nil, fmt.Errorf("registry: %w", err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&r.brains); err != nil {
		return nil, fmt.Errorf("registry: %s: %w", path, err)
	}
	return r, nil
}

// Save writes the registry. It replaces the file in one rename, so a crash
// never leaves it half-written.
func (r *Registry) Save() error {
	data, err := json.MarshalIndent(r.brains, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("registry: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".brains-*.json")
	if err != nil {
		return fmt.Errorf("registry: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return fmt.Errorf("registry: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("registry: %w", err)
	}
	if err := os.Rename(tmp.Name(), r.path); err != nil {
		return fmt.Errorf("registry: %w", err)
	}
	return nil
}

// List returns every brain, sorted by name.
func (r *Registry) List() []Brain {
	var out []Brain
	for _, name := range slices.Sorted(maps.Keys(r.brains)) {
		out = append(out, Brain{Name: name, Location: r.brains[name].Location})
	}
	return out
}

// Get returns the brain named name.
func (r *Registry) Get(name string) (Brain, error) {
	e, ok := r.brains[name]
	if !ok {
		return Brain{}, fmt.Errorf("(%s) %w", name, ErrNotFound)
	}
	return Brain{Name: name, Location: e.Location}, nil
}

// Register adds the folder at path, creating it when missing. name defaults
// to the folder's basename.
func (r *Registry) Register(path, name string) (Brain, error) {
	loc, err := filepath.Abs(path)
	if err != nil {
		return Brain{}, fmt.Errorf("register: %w", err)
	}
	if name == "" {
		name = filepath.Base(loc)
	}
	if err := checkName(name); err != nil {
		return Brain{}, err
	}
	if _, taken := r.brains[name]; taken {
		return Brain{}, fmt.Errorf("(%s) %w", name, ErrCollision)
	}
	for other, e := range r.brains {
		if SamePath(e.Location, loc) {
			return Brain{}, fmt.Errorf("(%s) %s %w", other, loc, ErrCollision)
		}
	}
	if err := os.MkdirAll(loc, 0o755); err != nil {
		return Brain{}, fmt.Errorf("(%s) %w", name, err)
	}
	r.brains[name] = entry{Location: loc}
	return Brain{Name: name, Location: loc}, nil
}

// A leading "-" would read as a flag in --brain -x.
func checkName(name string) error {
	if name == "" || strings.ContainsRune(name, filepath.Separator) || strings.HasPrefix(name, "-") {
		return fmt.Errorf("%q: %w", name, ErrInvalidName)
	}
	return nil
}

// Deregister removes the brain named name. The folder is left alone.
func (r *Registry) Deregister(name string) (Brain, error) {
	b, err := r.Get(name)
	if err != nil {
		return Brain{}, err
	}
	delete(r.brains, name)
	return b, nil
}
