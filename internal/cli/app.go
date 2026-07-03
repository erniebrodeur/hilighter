package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/erniebrodeur/hilighter/pkg/config"
	"github.com/erniebrodeur/hilighter/pkg/rules"
	"github.com/erniebrodeur/hilighter/pkg/runner"
)

// Main is the top-level CLI entrypoint.
func Main() error {
	opts, err := parseOptions()
	if err != nil {
		return err
	}
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
	if original.Mode == "tail" {
		return true
	}

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
	var cfg config.Config
	configPath := filepath.Join(opts.ConfigDir, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		cfg, err = config.Load(configPath)
		if err != nil {
			return Options{}, err
		}
	} else if err != nil && !os.IsNotExist(err) {
		return Options{}, err
	}

	if opts.Profile != "" {
		profile, ok := cfg.Profiles[opts.Profile]
		if !ok {
			return Options{}, fmt.Errorf("unknown profile %q", opts.Profile)
		}

		if opts.App == "" {
			opts.App = profile.App
		}
		if opts.RulesPath == "" {
			opts.RulesPath = profile.RulesPath
			if opts.RulesPath != "" {
				if _, err := os.Stat(opts.RulesPath); err != nil {
					if os.IsNotExist(err) {
						return Options{}, fmt.Errorf("profile %q references missing rules file %q", opts.Profile, opts.RulesPath)
					}
					return Options{}, err
				}
			}
		}
		if opts.ThemePath == "" {
			opts.ThemePath = profile.ThemePath
			if opts.ThemePath != "" {
				if _, err := os.Stat(opts.ThemePath); err != nil {
					if os.IsNotExist(err) {
						return Options{}, fmt.Errorf("profile %q references missing theme file %q", opts.Profile, opts.ThemePath)
					}
					return Options{}, err
				}
			}
		}
		if opts.FilePath == "" {
			opts.FilePath = profile.FilePath
		}
	}

	if opts.RulesPath == "" {
		opts.RulesPath = cfg.RulesPath
	}
	if opts.ThemePath == "" {
		opts.ThemePath = cfg.ThemePath
	}

	if opts.Mode == "tail" {
		if opts.FilePath == "" {
			return Options{}, fmt.Errorf("profile %q does not define a default file and no file argument was provided", opts.Profile)
		}

		opts.FilePath = resolveTailPath(opts.FilePath)
		if _, err := os.Stat(opts.FilePath); err != nil {
			if os.IsNotExist(err) {
				return Options{}, fmt.Errorf("tail target %q does not exist", opts.FilePath)
			}
			return Options{}, err
		}

		opts.Command = "tail -f " + strconv.Quote(opts.FilePath)
		return opts, nil
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

func resolveTailPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	if path == "." || path == ".." {
		return path
	}

	if len(path) >= 2 && path[:2] == "./" {
		return path
	}

	return "." + string(os.PathSeparator) + path
}
