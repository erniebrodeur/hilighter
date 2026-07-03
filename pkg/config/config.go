package config

// Config is the root YAML shape for hilighter's user defaults file.
//
// This file lives at ~/.hilighter/config.yaml by default and points the CLI at
// the rule and theme files it should use when flags do not override them.
// It may also define named profiles for workflows like `hilighter tail`.
type Config struct {
	// RulesPath points at the YAML file containing ordered highlighting rules.
	RulesPath string `yaml:"rules,omitempty"`
	// ThemePath points at the YAML file containing semantic theme definitions.
	ThemePath string `yaml:"theme,omitempty"`
	// Profiles stores named rule/theme/file presets for common workflows.
	Profiles map[string]Profile `yaml:"profiles,omitempty"`
}

// Profile is one named configuration preset.
type Profile struct {
	// App selects a built-in app profile.
	App string `yaml:"app,omitempty"`
	// RulesPath points at a YAML file containing ordered highlighting rules.
	RulesPath string `yaml:"rules,omitempty"`
	// ThemePath points at a YAML file containing theme style definitions.
	ThemePath string `yaml:"theme,omitempty"`
	// FilePath is an optional default file path for subcommands like `tail`.
	FilePath string `yaml:"file,omitempty"`
}
