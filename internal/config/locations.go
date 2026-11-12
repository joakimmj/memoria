package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Paths are the per-user files outside any brain.
type Paths struct {
	Config   string
	Registry string
}

// Locations returns the per-user file paths, via XDG on every OS.
func Locations() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, fmt.Errorf("locating config: %w", err)
	}
	configHome := xdg("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	dataHome := xdg("XDG_DATA_HOME", filepath.Join(home, ".local", "share"))
	return Paths{
		Config:   filepath.Join(configHome, "memoria", "config.json"),
		Registry: filepath.Join(dataHome, "memoria", "brains.json"),
	}, nil
}

// Relative values are ignored, per the XDG spec.
func xdg(name, fallback string) string {
	if v := os.Getenv(name); filepath.IsAbs(v) {
		return v
	}
	return fallback
}
