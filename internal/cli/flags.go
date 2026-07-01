package cli

import (
	"flag"

	"github.com/erniebrodeur/hilighter/pkg/config"
)

// Options carries CLI configuration into the application layer.
type Options struct {
	// RulesPath points at a YAML file containing ordered highlighting rules.
	RulesPath string
	// ThemePath points at a YAML file containing theme style definitions.
	ThemePath string
	// Command is an optional shell command to execute instead of reading stdin.
	Command string
	// ConfigDir is the base directory for default config discovery.
	ConfigDir string
}

// parseOptions reads CLI flags into an Options value.
//
// The current scaffold exposes direct rule and theme paths plus a config root.
// Later slices can layer config-file loading on top of these values without
// changing the external CLI shape.
func parseOptions() Options {
	opts := Options{}

	flag.StringVar(&opts.RulesPath, "rules", "", "path to a rules YAML file")
	flag.StringVar(&opts.ThemePath, "theme", "", "path to a theme YAML file")
	flag.StringVar(&opts.Command, "cmd", "", "command to execute and stream through hilighter")
	flag.StringVar(&opts.ConfigDir, "config-dir", config.DefaultDir(), "config directory (default: ~/.hilighter)")
	flag.Parse()

	return opts
}
