package config

// Config is the root YAML shape for hilighter configuration.
type Config struct {
	// Rules is the ordered list of highlighting rules to evaluate for each line.
	Rules []Rule `yaml:"rules"`
}

// Rule describes one ordered highlighting rule.
//
// Rules are evaluated in declaration order. When multiple rules overlap, the
// earlier rule wins. Scope defaults to substring matching when omitted.
type Rule struct {
	// Name is a stable identifier for diagnostics and future rule references.
	Name string `yaml:"name"`
	// Pattern is the PCRE-compatible regular expression to evaluate.
	Pattern string `yaml:"pattern"`
	// Scope selects whether the match affects only the substring or the line.
	Scope string `yaml:"scope,omitempty"`
	// Style is the semantic label applied to the whole match or full line.
	Style string `yaml:"style,omitempty"`
	// Groups maps capture group indexes to semantic labels for rendering.
	Groups map[string]string `yaml:"groups,omitempty"`
}
