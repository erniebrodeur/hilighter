package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

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
// config.yaml is intended to hold file-location overrides and later global
// defaults that should apply without repeating flags on every invocation.
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

// Load reads a hilighter YAML config file.
//
// When path is empty, Load reads from the default ~/.hilighter/config.yaml
// location. Tilde-prefixed rule and theme paths are expanded to the user's home
// directory on load.
func Load(path string) (Config, error) {
	if path == "" {
		path = DefaultConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	cfg.RulesPath = expandHome(cfg.RulesPath)
	cfg.ThemePath = expandHome(cfg.ThemePath)
	for name, profile := range cfg.Profiles {
		profile.RulesPath = expandHome(profile.RulesPath)
		profile.ThemePath = expandHome(profile.ThemePath)
		profile.FilePath = expandHome(profile.FilePath)
		cfg.Profiles[name] = profile
	}

	return cfg, nil
}

func expandHome(path string) string {
	if path == "" || !strings.HasPrefix(path, "~") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}

	if path == "~" {
		return home
	}

	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}

	return path
}
