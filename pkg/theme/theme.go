// Package theme owns semantic style resolution for output rendering.
//
// Themes map semantic labels such as "error" or "test-name" onto concrete
// terminal styles. The default theme direction for hilighter is Monokai-like
// foreground styling without background colors.
package theme

import (
	"os"

	"go.yaml.in/yaml/v3"
)

// Theme is the YAML shape for a hilighter theme file.
type Theme struct {
	// Styles maps semantic labels to concrete terminal styling.
	Styles map[string]Style `yaml:"styles"`
}

// Style is one resolved terminal style.
type Style struct {
	// FG is the foreground color name.
	FG string `yaml:"fg,omitempty"`
	// BG is the background color name.
	BG string `yaml:"bg,omitempty"`
	// Bold toggles ANSI bold output.
	Bold bool `yaml:"bold,omitempty"`
}

// Load reads a theme YAML file from disk.
func Load(path string) (Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Theme{}, err
	}

	var th Theme
	if err := yaml.Unmarshal(data, &th); err != nil {
		return Theme{}, err
	}

	return th, nil
}

// Resolve looks up a semantic style label.
func (t Theme) Resolve(label string) (Style, bool) {
	style, ok := t.Styles[label]
	return style, ok
}
