package config

import (
	"bytes"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

const defaultDirName = ".hilighter"
const defaultConfigName = "config.yaml"
const defaultRulesName = "rules.yaml"
const themesDirName = "themes"

const defaultConfig = "theme: monokai\n"

// DefaultDir returns the default per-user configuration directory.
//
// The project convention is to keep user-managed defaults in ~/.hilighter so
// rules, themes, and config overrides live under one predictable root.
func DefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return defaultDirName
	}

	return home + string(os.PathSeparator) + defaultDirName
}

// DefaultConfigPath returns the default per-user config file path.
//
// config.yaml stores the persistent theme selector.
func DefaultConfigPath() string {
	return DefaultDir() + string(os.PathSeparator) + defaultConfigName
}

// EnsureLayout creates the user customization files without overwriting
// anything already present.
func EnsureLayout(dir string) error {
	if dir == "" {
		dir = DefaultDir()
	}

	if err := os.MkdirAll(filepath.Join(dir, themesDirName), 0o755); err != nil {
		return err
	}
	if err := createFileIfMissing(filepath.Join(dir, defaultConfigName), defaultConfig); err != nil {
		return err
	}
	return createFileIfMissing(filepath.Join(dir, defaultRulesName), "")
}

func createFileIfMissing(path, contents string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if os.IsExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if _, err := file.WriteString(contents); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

// DefaultRulesPath returns the conventional user rules file.
func DefaultRulesPath(dir string) string {
	if dir == "" {
		dir = DefaultDir()
	}
	return filepath.Join(dir, defaultRulesName)
}

// LoadTheme reads the strict, theme-only user config.
func LoadTheme(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var cfg struct {
		Theme string `yaml:"theme,omitempty"`
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return "", err
	}
	return cfg.Theme, nil
}
