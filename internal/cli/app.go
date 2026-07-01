package cli

import (
	"os"
	"path/filepath"

	"github.com/erniebrodeur/hilighter/pkg/config"
	"github.com/erniebrodeur/hilighter/pkg/runner"
)

// Main is the top-level CLI entrypoint.
func Main() error {
	opts := parseOptions()
	resolved, err := resolveOptions(opts)
	if err != nil {
		return err
	}

	highlighter, err := runner.NewHighlighter(resolved.RulesPath, resolved.ThemePath)
	if err != nil {
		return err
	}
	defer highlighter.Close()

	if resolved.Command != "" {
		return runner.RunCommand(resolved.Command, os.Stdout, os.Stderr, highlighter)
	}

	return runner.RunStdin(os.Stdin, os.Stdout, highlighter)
}

func resolveOptions(opts Options) (Options, error) {
	configPath := filepath.Join(opts.ConfigDir, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		cfg, err := config.Load(configPath)
		if err != nil {
			return Options{}, err
		}
		if opts.RulesPath == "" {
			opts.RulesPath = cfg.RulesPath
		}
		if opts.ThemePath == "" {
			opts.ThemePath = cfg.ThemePath
		}
	} else if err != nil && !os.IsNotExist(err) {
		return Options{}, err
	}

	return opts, nil
}
