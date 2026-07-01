package config

import "os"

const defaultDirName = ".hilighter"
const defaultConfigName = "config.yaml"

// DefaultDir returns the default per-user configuration directory.
func DefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return defaultDirName
	}

	return home + string(os.PathSeparator) + defaultDirName
}

// DefaultConfigPath returns the default per-user config file path.
func DefaultConfigPath() string {
	return DefaultDir() + string(os.PathSeparator) + defaultConfigName
}

// Load is a scaffold placeholder until the YAML loader is implemented.
func Load(_ string) (Config, error) {
	return Config{}, nil
}
