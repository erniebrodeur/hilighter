package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/erniebrodeur/hilighter/pkg/config"
)

// Options carries CLI configuration into the application layer.
type Options struct {
	// ShowHelp requests CLI usage output.
	ShowHelp bool
	// ShowVersion requests the app version marker.
	ShowVersion bool
	// Mode selects a higher-level CLI mode such as "tail", "cat", or "head".
	Mode string
	// Profile selects a named user profile from config.yaml when a file mode
	// argument resolves to a saved profile name.
	Profile string
	// App selects a built-in rule pack such as "syslog".
	App string
	// RulesPath points at a YAML file containing ordered highlighting rules.
	RulesPath string
	// ThemePath points at a YAML file containing theme style definitions.
	ThemePath string
	// Command is an optional shell command to execute instead of reading stdin.
	Command string
	// FilePath points at a file target for modes such as `tail`, `cat`, or `head`.
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
	flag.BoolVar(&opts.ShowVersion, "version", false, "print version information")
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		return Options{}, err
	}

	args := flag.Args()
	if len(args) == 0 {
		return opts, nil
	}

	if len(args) == 1 && args[0] == "version" {
		opts.ShowVersion = true
		return opts, nil
	}

	if !isFileMode(args[0]) {
		return Options{}, fmt.Errorf("unknown command %q", args[0])
	}

	if len(args) < 2 {
		return Options{}, fmt.Errorf("%s requires a profile name or file path", args[0])
	}

	opts.Mode = args[0]
	opts.Profile = args[1]
	if len(args) >= 3 {
		opts.FilePath = args[2]
	}
	if len(args) > 3 {
		return Options{}, fmt.Errorf("%s accepts at most a profile name and one file path", args[0])
	}

	return opts, nil
}

func isFileMode(mode string) bool {
	switch mode {
	case "tail", "cat", "head":
		return true
	default:
		return false
	}
}
