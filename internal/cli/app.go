package cli

import (
	"errors"
	"flag"
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
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	mode := stdinMode(os.Stdin)
	if opts.ShowVersion {
		_, _ = fmt.Fprintln(os.Stdout, formattedVersion())
		return nil
	}
	if shouldShowHelp(opts, mode) {
		printHelp(os.Stdout)
		return nil
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

	if shouldRunCommand(opts, resolved, mode) {
		return runner.RunCommand(resolved.Command, os.Stdout, os.Stderr, highlighter)
	}

	return runner.RunStdin(os.Stdin, os.Stdout, highlighter)
}

func shouldShowHelp(opts Options, mode fs.FileMode) bool {
	return opts.Mode == "" &&
		!opts.ShowVersion &&
		opts.Profile == "" &&
		opts.App == "" &&
		opts.RulesPath == "" &&
		opts.ThemePath == "" &&
		opts.Command == "" &&
		opts.FilePath == "" &&
		mode&os.ModeCharDevice != 0
}

func printHelp(out *os.File) {
	_, _ = fmt.Fprintf(out, `hilighter

Usage:
  hilighter --app <name>
  hilighter --rules <file>
  hilighter --cmd "<command>"
  hilighter tail <profile> [file]
  hilighter cat <profile> [file]
  hilighter head <profile> [file]
  hilighter --version

Flags:
  --app         built-in profile to use
  --rules       path to a rules YAML file
  --theme       path to a theme YAML file
  --cmd         command to run through hilighter
  --config-dir  config directory (default: ~/.hilighter)
  --version     print version information
`)
}

func shouldRunCommand(original, resolved Options, mode fs.FileMode) bool {
	if isFileMode(original.Mode) {
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
			if isFileMode(opts.Mode) {
				if opts.FilePath != "" {
					return Options{}, fmt.Errorf("%s accepts either a file path alone or a profile name with an optional file path", opts.Mode)
				}

				opts.FilePath = opts.Profile
				opts.Profile = ""
			} else {
				return Options{}, fmt.Errorf("unknown profile %q", opts.Profile)
			}
		}

		if ok {
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
	}

	if opts.RulesPath == "" {
		opts.RulesPath = cfg.RulesPath
	}
	if opts.ThemePath == "" {
		opts.ThemePath = cfg.ThemePath
	}

	if isFileMode(opts.Mode) {
		filePath, command, err := resolveFileMode(opts.Mode, opts.Profile, opts.FilePath)
		if err != nil {
			return Options{}, err
		}
		opts.FilePath = filePath
		opts.Command = command
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

func resolveFileMode(mode, profile, path string) (string, string, error) {
	if path == "" {
		return "", "", fmt.Errorf("profile %q does not define a default file and no file argument was provided", profile)
	}

	resolvedPath := resolveFilePath(path)
	if _, err := os.Stat(resolvedPath); err != nil {
		if os.IsNotExist(err) {
			return "", "", fmt.Errorf("%s target %q does not exist", mode, resolvedPath)
		}
		return "", "", err
	}

	switch mode {
	case "tail":
		return resolvedPath, "tail -f " + strconv.Quote(resolvedPath), nil
	case "cat":
		return resolvedPath, "cat " + strconv.Quote(resolvedPath), nil
	case "head":
		return resolvedPath, "head " + strconv.Quote(resolvedPath), nil
	default:
		return "", "", fmt.Errorf("unknown file mode %q", mode)
	}
}

func resolveFilePath(path string) string {
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
