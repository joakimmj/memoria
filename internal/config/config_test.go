package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setup points the XDG variables at a temp dir and returns it.
func setup(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(dir, "data"))
	return dir
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		user    string // "" means no file
		brain   string // "" means no file
		check   func(t *testing.T, c Config)
		wantErr string // substring; "" means no error
	}{
		{
			name: "no files is the defaults",
			check: func(t *testing.T, c Config) {
				if c.Engrams.Filename.Prefix != "date" || c.Engrams.ContentTemplate != "# %title%\n\n%cursor%" {
					t.Errorf("not the defaults: %+v", c)
				}
			},
		},
		{
			name: "user tier over defaults",
			user: `{"engrams": {"date_format": "DD.MM.YYYY"}}`,
			check: func(t *testing.T, c Config) {
				if c.Engrams.DateFormat != "DD.MM.YYYY" || c.Engrams.Filename.DateFormat != "YYYY-MM-DD" {
					t.Errorf("engrams = %+v", c.Engrams)
				}
			},
		},
		{
			name:  "brain tier over user tier",
			user:  `{"engrams": {"filename": {"prefix": "none"}}}`,
			brain: `{"engrams": {"filename": {"prefix": "date"}}}`,
			check: func(t *testing.T, c Config) {
				if c.Engrams.Filename.Prefix != "date" {
					t.Errorf("prefix = %q", c.Engrams.Filename.Prefix)
				}
			},
		},
		{
			name:  "brain null skips the user tier",
			user:  `{"engrams": {"date_format": "DD.MM.YYYY"}}`,
			brain: `{"engrams": {"date_format": null}}`,
			check: func(t *testing.T, c Config) {
				if c.Engrams.DateFormat != "YYYY-MM-DD" {
					t.Errorf("date_format = %q", c.Engrams.DateFormat)
				}
			},
		},
		{name: "invalid JSON names the file", user: `{"engrams": `, wantErr: "config.json"},
		{name: "trailing data", user: `{} {}`, wantErr: "config.json"},
		{name: "unknown key names file and key", brain: `{"engrams": {"date_fromat": "x"}}`, wantErr: `.mia_dna.json: json: unknown field "date_fromat"`},
		{name: "key of a later feature is unknown", brain: `{"synapses": {"up": null}}`, wantErr: `unknown field "synapses"`},
		{name: "wrong type", user: `{"engrams": {"date_format": 3}}`, wantErr: "config.json"},
		{name: "not an object", user: `[]`, wantErr: "config.json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := setup(t)
			brainDir := filepath.Join(dir, "brain")
			if tt.user != "" {
				write(t, filepath.Join(dir, "config", "memoria", "config.json"), tt.user)
			}
			if tt.brain != "" {
				write(t, filepath.Join(brainDir, DNAFile), tt.brain)
			}
			c, err := Load(brainDir)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want one containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			tt.check(t, c)
		})
	}
}

func TestLoadUnreadable(t *testing.T) {
	dir := setup(t)
	path := filepath.Join(dir, "config", "memoria", "config.json")
	write(t, path, `{}`)
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatal(err)
	}
	if _, err := os.ReadFile(path); err == nil {
		t.Skip("file still readable (running as root?)")
	}
	if _, err := Load(""); err == nil || !strings.Contains(err.Error(), "config.json") {
		t.Errorf("err = %v, want one naming config.json", err)
	}
}

func TestDefaults(t *testing.T) {
	c, err := Defaults()
	if err != nil {
		t.Fatal(err)
	}
	want := Engrams{
		DateFormat:      "YYYY-MM-DD",
		Filename:        Filename{Prefix: "date", DateFormat: "YYYY-MM-DD"},
		ContentTemplate: "# %title%\n\n%cursor%",
	}
	if c.Engrams != want {
		t.Errorf("Defaults().Engrams = %+v, want %+v", c.Engrams, want)
	}
}

func TestLocations(t *testing.T) {
	tests := []struct {
		name, configHome, dataHome string
		wantConfig, wantRegistry   string // relative to $HOME when not absolute
	}{
		{
			name:         "XDG set",
			configHome:   "/x/config",
			dataHome:     "/x/data",
			wantConfig:   "/x/config/memoria/config.json",
			wantRegistry: "/x/data/memoria/brains.json",
		},
		{
			name:         "XDG unset",
			wantConfig:   ".config/memoria/config.json",
			wantRegistry: ".local/share/memoria/brains.json",
		},
		{
			name:         "relative XDG ignored",
			configHome:   "rel/config",
			dataHome:     "rel/data",
			wantConfig:   ".config/memoria/config.json",
			wantRegistry: ".local/share/memoria/brains.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("XDG_CONFIG_HOME", tt.configHome)
			t.Setenv("XDG_DATA_HOME", tt.dataHome)
			p, err := Locations()
			if err != nil {
				t.Fatal(err)
			}
			abs := func(s string) string {
				if filepath.IsAbs(s) {
					return filepath.FromSlash(s)
				}
				return filepath.Join(home, filepath.FromSlash(s))
			}
			if p.Config != abs(tt.wantConfig) || p.Registry != abs(tt.wantRegistry) {
				t.Errorf("Locations() = %+v", p)
			}
		})
	}
}
