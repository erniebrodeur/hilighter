package cli

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/erniebrodeur/hilighter/pkg/config"
	"github.com/erniebrodeur/hilighter/pkg/rules"
	"github.com/erniebrodeur/hilighter/pkg/runner"
)

// Main is the top-level CLI entrypoint.
func Main() error {
	opts := parseOptions()
	resolved, err := resolveOptions(opts)
	if err != nil {
		return err
	}

	highlighter, err := runner.NewHighlighter(resolved.RulesPath, resolved.App, resolved.ThemePath)
	if err != nil {
		return err
	}
	defer highlighter.Close()

	if shouldRunCommand(opts, resolved, stdinMode(os.Stdin)) {
		return runner.RunCommand(resolved.Command, os.Stdout, os.Stderr, highlighter)
	}

	return runner.RunStdin(os.Stdin, os.Stdout, highlighter)
}

func shouldRunCommand(original, resolved Options, mode fs.FileMode) bool {
	if original.Command != "" {
		return true
	}

	if resolved.Command == "" {
		return false
	}

	return mode&os.ModeCharDevice != 0
}

func stdinMode(file *os.File) fs.FileMode {
	info, err := file.Stat()
	if err != nil {
		return os.ModeCharDevice
	}

	return info.Mode()
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

	if opts.Command == "" {
		if opts.RulesPath != "" {
			ruleFile, err := rules.Load(opts.RulesPath)
			if err != nil {
				return Options{}, err
			}
			opts.Command = ruleFile.Command
		} else if opts.App != "" {
			ruleFile, ok := rules.Builtin(opts.App)
			if !ok {
				return Options{}, rules.ErrUnknownBuiltin(opts.App)
			}
			opts.Command = ruleFile.Command
		}
	}

	return opts, nil
}
