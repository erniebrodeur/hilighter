package cli

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

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
		if opts.Profile == "" && opts.App == "" && opts.RulesPath == "" {
			autoApp, autoRules := autoDetectFileModeHighlighting(opts.FilePath, cfg, opts.ConfigDir)
			if opts.App == "" {
				opts.App = autoApp
			}
			if opts.RulesPath == "" {
				opts.RulesPath = autoRules
			}
		}

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

func autoDetectFileModeHighlighting(path string, cfg config.Config, configDir string) (string, string) {
	resolvedPath := resolveFilePath(path)

	if profile, ok := detectProfileForFile(resolvedPath, cfg.Profiles); ok {
		return profile.App, profile.RulesPath
	}

	if rulesPath, ok := detectRulesFileForPath(resolvedPath, configDir); ok {
		return "", rulesPath
	}

	if appName, ok := detectBuiltinForPath(resolvedPath); ok {
		return appName, ""
	}

	return "", ""
}

func detectProfileForFile(path string, profiles map[string]config.Profile) (config.Profile, bool) {
	if len(profiles) == 0 {
		return config.Profile{}, false
	}

	target := normalizeMatchPath(path)
	for _, profile := range profiles {
		if profile.FilePath == "" || (profile.App == "" && profile.RulesPath == "") {
			continue
		}

		candidate := normalizeMatchPath(profile.FilePath)
		if candidate == "" {
			continue
		}

		if target == candidate || strings.HasSuffix(target, "/"+candidate) {
			return profile, true
		}
	}

	targetBase := filepath.Base(target)
	var matched config.Profile
	matches := 0
	for _, profile := range profiles {
		if profile.FilePath == "" || (profile.App == "" && profile.RulesPath == "") {
			continue
		}

		if filepath.Base(normalizeMatchPath(profile.FilePath)) == targetBase {
			matched = profile
			matches++
		}
	}
	if matches == 1 {
		return matched, true
	}

	return config.Profile{}, false
}

func detectRulesFileForPath(path, configDir string) (string, bool) {
	rulesDir := filepath.Join(configDir, "rules")
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		return "", false
	}

	targetTokens := pathTokens(path)
	if len(targetTokens) == 0 {
		return "", false
	}

	type candidate struct {
		name string
		path string
	}

	var matches []candidate
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		stem := strings.TrimSuffix(name, filepath.Ext(name))
		normalizedStem := normalizeToken(stem)
		if normalizedStem == "" {
			continue
		}

		for _, token := range targetTokens {
			if token == normalizedStem {
				matches = append(matches, candidate{
					name: normalizedStem,
					path: filepath.Join(rulesDir, name),
				})
				break
			}
		}
	}

	if len(matches) == 0 {
		return "", false
	}

	sort.Slice(matches, func(i, j int) bool {
		if len(matches[i].name) == len(matches[j].name) {
			return matches[i].path < matches[j].path
		}
		return len(matches[i].name) > len(matches[j].name)
	})

	return matches[0].path, true
}

func detectBuiltinForPath(path string) (string, bool) {
	targetTokens := pathTokens(path)
	if len(targetTokens) == 0 {
		return "", false
	}

	var matches []string
	for name := range rules.Builtins() {
		normalizedName := normalizeToken(name)
		if normalizedName == "" {
			continue
		}

		for _, token := range targetTokens {
			if token == normalizedName {
				matches = append(matches, name)
				break
			}
		}
	}

	if len(matches) == 0 {
		return "", false
	}

	sort.Slice(matches, func(i, j int) bool {
		if len(matches[i]) == len(matches[j]) {
			return matches[i] < matches[j]
		}
		return len(matches[i]) > len(matches[j])
	})

	return matches[0], true
}

func pathTokens(path string) []string {
	normalized := strings.ToLower(filepath.Clean(strings.TrimPrefix(path, "./")))
	parts := strings.FieldsFunc(normalized, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})

	seen := make(map[string]struct{}, len(parts))
	tokens := make([]string, 0, len(parts))
	for _, part := range parts {
		token := normalizeToken(part)
		if token == "" {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		tokens = append(tokens, token)
	}

	return tokens
}

func normalizeMatchPath(path string) string {
	normalized := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(path, "./")))
	return strings.ToLower(normalized)
}

func normalizeToken(value string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(value) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
