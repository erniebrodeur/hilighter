package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/erniebrodeur/hilighter/pkg/config"
)

// Options carries CLI configuration into the application layer.
type Options struct {
	// Mode selects a higher-level CLI mode such as "tail".
	Mode string
	// Profile selects a named user profile from config.yaml.
	Profile string
	// App selects a built-in rule pack such as "syslog".
	App string
	// RulesPath points at a YAML file containing ordered highlighting rules.
	RulesPath string
	// ThemePath points at a YAML file containing theme style definitions.
	ThemePath string
	// Command is an optional shell command to execute instead of reading stdin.
	Command string
	// FilePath points at a file target for modes such as `tail`.
	FilePath string
	// ConfigDir is the base directory for default config discovery.
	ConfigDir string
}

// parseOptions reads CLI flags into an Options value.
//
// Rules and themes can be set explicitly with flags or discovered through the
// config.yaml file under the selected config directory.
func parseOptions() (Options, error) {
	opts := Options{}

	flag.StringVar(&opts.App, "app", "", "built-in app profile to use (for example: syslog)")
	flag.StringVar(&opts.RulesPath, "rules", "", "path to a rules YAML file")
	flag.StringVar(&opts.ThemePath, "theme", "", "path to a theme YAML file")
	flag.StringVar(&opts.Command, "cmd", "", "command to execute and stream through hilighter")
	flag.StringVar(&opts.ConfigDir, "config-dir", config.DefaultDir(), "config directory (default: ~/.hilighter)")
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		return Options{}, err
	}

	args := flag.Args()
	if len(args) == 0 {
		return opts, nil
	}

	if args[0] != "tail" {
		return Options{}, fmt.Errorf("unknown command %q", args[0])
	}

	if len(args) < 2 {
		return Options{}, fmt.Errorf("tail requires a profile name")
	}

	opts.Mode = "tail"
	opts.Profile = args[1]
	if len(args) >= 3 {
		opts.FilePath = args[2]
	}
	if len(args) > 3 {
		return Options{}, fmt.Errorf("tail accepts at most a profile name and one file path")
	}

	return opts, nil
}
