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
)

// DNAFile is the brain tier's filename, at the brain's root.
const DNAFile = ".mia_dna.json"

//go:embed defaults.json
var defaultsJSON []byte

// defaultTree is shared; never modify it.
var defaultTree = mustParse(defaultsJSON)

// Config is the merged configuration.
type Config struct {
	Engrams Engrams `json:"engrams"`
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

	var tree any = defaultTree
	for _, f := range files {
		tier, err := readTier(f)
		if err != nil {
			return Config{}, err
		}
		if tier != nil {
			tree = merge(tree, tier, defaultTree)
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
	// Typed check per file, so the error names it.
	if err := strictUnmarshal(data, &Config{}); err != nil {
		return nil, fmt.Errorf("config: %s: %w", path, err)
	}
	return tree, nil
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
