package config

import "os"

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
// The scaffold currently returns an empty config until the rule-engine slice is
// implemented. The signature is stable so callers and tests can be built first.
func Load(_ string) (Config, error) {
	return Config{}, nil
}
