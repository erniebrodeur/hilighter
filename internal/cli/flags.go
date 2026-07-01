package cli

import (
	"flag"

	"github.com/erniebrodeur/hilighter/pkg/config"
)

// Options carries CLI configuration into the application layer.
type Options struct {
	RulesPath string
	ThemePath string
	Command   string
	ConfigDir string
}

func parseOptions() Options {
	opts := Options{}

	flag.StringVar(&opts.RulesPath, "rules", "", "path to a rules YAML file")
	flag.StringVar(&opts.ThemePath, "theme", "", "path to a theme YAML file")
	flag.StringVar(&opts.Command, "cmd", "", "command to execute and stream through hilighter")
	flag.StringVar(&opts.ConfigDir, "config-dir", config.DefaultDir(), "config directory (default: ~/.hilighter)")
	flag.Parse()

	return opts
}
