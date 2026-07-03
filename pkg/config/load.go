package config

import (
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

const defaultDirName = ".hilighter"
const defaultConfigName = "config.yaml"

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
