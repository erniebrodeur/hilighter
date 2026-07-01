package config

// Config is the root YAML shape for hilighter configuration.
type Config struct {
	Rules []Rule `yaml:"rules"`
}

// Rule describes one ordered highlighting rule.
type Rule struct {
	Name    string            `yaml:"name"`
	Pattern string            `yaml:"pattern"`
	Scope   string            `yaml:"scope,omitempty"`
	Style   string            `yaml:"style,omitempty"`
	Groups  map[string]string `yaml:"groups,omitempty"`
}
