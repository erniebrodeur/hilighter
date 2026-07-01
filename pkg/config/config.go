package config

// Config is the root YAML shape for hilighter's user defaults file.
//
// This file lives at ~/.hilighter/config.yaml by default and points the CLI at
// the rule and theme files it should use when flags do not override them.
type Config struct {
	// RulesPath points at the YAML file containing ordered highlighting rules.
	RulesPath string `yaml:"rules,omitempty"`
	// ThemePath points at the YAML file containing semantic theme definitions.
	ThemePath string `yaml:"theme,omitempty"`
}
