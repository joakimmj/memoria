// Package config loads memoria's configuration: the built-in defaults, the
// user's config.json and a brain's .mia_dna.json, merged in that order.
package config

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
)

// DNAFile is the brain tier's filename, at the brain's root.
const DNAFile = ".mia_dna.json"

//go:embed defaults.json
var defaultsJSON []byte

// defaultTree is shared; never modify it.
var defaultTree = mustParse(defaultsJSON)

// Config is the merged configuration.
type Config struct {
	Engrams  Engrams            `json:"engrams"`
	Synapses map[string]Synapse `json:"synapses"`
}

// Engrams shapes new engrams.
type Engrams struct {
	DateFormat      string   `json:"date_format"`
	Filename        Filename `json:"filename"`
	ContentTemplate string   `json:"content_template"`
}

// Filename shapes an engram's filename.
type Filename struct {
	Prefix     string `json:"prefix"`
	DateFormat string `json:"date_format"`
}

// Synapse is a synapse field, keyed by its label.
type Synapse struct {
	Inverse   string `json:"inverse"`
	List      bool   `json:"list"`
	ShowEmpty bool   `json:"show_empty"`
}

// UnmarshalJSON makes a left-out list or show_empty true.
func (s *Synapse) UnmarshalJSON(data []byte) error {
	// plain has no methods, so decoding into it does not recurse.
	type plain Synapse
	p := plain{List: true, ShowEmpty: true}
	if err := strictUnmarshal(data, &p); err != nil {
		return err
	}
	*s = Synapse(p)
	return nil
}

var label = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

// Defaults returns the built-in defaults alone.
func Defaults() (Config, error) {
	return decode(defaultTree)
}

// Load merges the defaults, the user's config and, when brainDir is set, its
// .mia_dna.json. A missing file is skipped; an invalid one is an error naming it.
func Load(brainDir string) (Config, error) {
	paths, err := Locations()
	if err != nil {
		return Config{}, err
	}
	files := []string{paths.Config}
	if brainDir != "" {
		files = append(files, filepath.Join(brainDir, DNAFile))
	}

	tree := defaultTree
	for _, f := range files {
		tier, err := readTier(f)
		if err != nil {
			return Config{}, err
		}
		if tier != nil {
			tree = removeNullEntries(merge(tree, tier, defaultTree).(map[string]any), tier)
		}
	}

	c, err := decode(tree)
	if err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}
	return c, nil
}

// readTier returns nil, nil for a missing file.
func readTier(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	var tree map[string]any
	if err := strictUnmarshal(data, &tree); err != nil {
		return nil, fmt.Errorf("config: %s: %w", path, err)
	}
	if err := check(data); err != nil {
		return nil, fmt.Errorf("config: %s: %w", path, err)
	}
	return tree, nil
}

// check validates one tier on its own, so an error can name its file.
func check(data []byte) error {
	var c Config
	if err := strictUnmarshal(data, &c); err != nil {
		return err
	}
	for name := range c.Synapses {
		if !label.MatchString(name) {
			return fmt.Errorf("synapse label %q: only letters, digits and _", name)
		}
	}
	return nil
}

// mustParse panics: broken embedded JSON is a build bug.
func mustParse(data []byte) map[string]any {
	var tree map[string]any
	if err := json.Unmarshal(data, &tree); err != nil {
		panic("config: built-in defaults: " + err.Error())
	}
	return tree
}

func decode(tree any) (Config, error) {
	data, err := json.Marshal(tree)
	if err != nil {
		return Config{}, err
	}
	var c Config
	err = strictUnmarshal(data, &c)
	return c, err
}

// strictUnmarshal also rejects unknown keys and trailing data.
func strictUnmarshal(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return errors.New("unexpected data after the JSON object")
	}
	return nil
}
