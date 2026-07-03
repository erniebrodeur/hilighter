package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
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
	"github.com/erniebrodeur/hilighter/pkg/theme"
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
	if isControlMode(opts.Mode) {
		return runControlMode(opts, os.Stdout, os.Stderr)
	}

	resolved, err := resolveOptions(opts)
	if err != nil {
		return err
	}
	if opts.DebugDetect && resolved.DetectionNote != "" {
		_, _ = fmt.Fprintf(os.Stderr, "detect: %s\n", resolved.DetectionNote)
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

func printHelp(out io.Writer) {
	_, _ = fmt.Fprintf(out, `hilighter

Usage:
  hilighter --app <name>
  hilighter --rules <file>
  hilighter --cmd "<command>"
  hilighter tail <profile> [file]
  hilighter cat <profile> [file]
  hilighter head <profile> [file]
  hilighter validate
  hilighter list apps|profiles
  hilighter show app|profile <name>
  hilighter --version

Flags:
  --app         built-in profile to use
  --rules       path to a rules YAML file
  --theme       path to a theme YAML file
  --cmd         command to run through hilighter
  --config-dir  config directory (default: ~/.hilighter)
  --no-detect   disable file-mode auto-detection
  --debug-detect print file-mode detection decisions
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

func isControlMode(mode string) bool {
	switch mode {
	case "validate", "list", "show":
		return true
	default:
		return false
	}
}

func resolveOptions(opts Options) (Options, error) {
	cfg, _, err := loadConfigIfPresent(opts.ConfigDir)
	if err != nil {
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
				if opts.DebugDetect {
					opts.DetectionNote = fmt.Sprintf("direct file target %q", resolveFilePath(opts.FilePath))
				}
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
				if err := validateProfileAsset(opts.Profile, "rules", opts.RulesPath); err != nil {
					return Options{}, err
				}
			}
			if opts.ThemePath == "" {
				opts.ThemePath = profile.ThemePath
				if err := validateProfileAsset(opts.Profile, "theme", opts.ThemePath); err != nil {
					return Options{}, err
				}
			}
			if opts.FilePath == "" {
				opts.FilePath = profile.FilePath
			}
			if opts.DebugDetect {
				opts.DetectionNote = fmt.Sprintf("explicit profile %q", opts.Profile)
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
		if opts.NoDetect {
			if opts.DebugDetect {
				opts.DetectionNote = "auto-detection disabled"
			}
		} else if opts.Profile == "" && opts.App == "" && opts.RulesPath == "" {
			detected := autoDetectFileModeHighlighting(opts.FilePath, cfg, opts.ConfigDir)
			if opts.App == "" {
				opts.App = detected.App
			}
			if opts.RulesPath == "" {
				opts.RulesPath = detected.RulesPath
			}
			if opts.DebugDetect {
				opts.DetectionNote = detected.Message
			}
		} else if opts.DebugDetect && opts.DetectionNote == "" {
			opts.DetectionNote = "auto-detection not needed"
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

func loadConfigIfPresent(configDir string) (config.Config, bool, error) {
	configPath := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		cfg, loadErr := config.Load(configPath)
		if loadErr != nil {
			return config.Config{}, false, loadErr
		}
		return cfg, true, nil
	} else if err != nil && !os.IsNotExist(err) {
		return config.Config{}, false, err
	}

	return config.Config{}, false, nil
}

func runControlMode(opts Options, stdout, stderr io.Writer) error {
	cfg, cfgFound, err := loadConfigIfPresent(opts.ConfigDir)
	if err != nil {
		return err
	}

	switch opts.Mode {
	case "validate":
		return runValidate(opts, cfg, cfgFound, stdout)
	case "list":
		return runList(opts, cfg, stdout)
	case "show":
		return runShow(opts, cfg, stdout)
	default:
		_, _ = fmt.Fprintf(stderr, "unknown control mode %q\n", opts.Mode)
		return fmt.Errorf("unknown control mode %q", opts.Mode)
	}
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

func validateProfileAsset(profileName, assetKind, path string) error {
	if path == "" {
		return nil
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("profile %q references missing %s file %q", profileName, assetKind, path)
		}
		return err
	}

	return nil
}

type detectionResult struct {
	App       string
	RulesPath string
	Message   string
}

func autoDetectFileModeHighlighting(path string, cfg config.Config, configDir string) detectionResult {
	resolvedPath := resolveFilePath(path)

	if name, profile, ok := detectProfileForFile(resolvedPath, cfg.Profiles); ok {
		return detectionResult{
			App:       profile.App,
			RulesPath: profile.RulesPath,
			Message:   fmt.Sprintf("matched profile file %q for %q", name, resolvedPath),
		}
	}

	if rulesPath, ok := detectRulesFileForPath(resolvedPath, configDir); ok {
		return detectionResult{
			RulesPath: rulesPath,
			Message:   fmt.Sprintf("matched rules file %q for %q", rulesPath, resolvedPath),
		}
	}

	if appName, ok := detectBuiltinForPath(resolvedPath); ok {
		return detectionResult{
			App:     appName,
			Message: fmt.Sprintf("matched built-in app %q for %q", appName, resolvedPath),
		}
	}

	return detectionResult{Message: fmt.Sprintf("no highlight match for %q", resolvedPath)}
}

func detectProfileForFile(path string, profiles map[string]config.Profile) (string, config.Profile, bool) {
	if len(profiles) == 0 {
		return "", config.Profile{}, false
	}

	target := normalizeMatchPath(path)
	for name, profile := range profiles {
		if profile.FilePath == "" || (profile.App == "" && profile.RulesPath == "") {
			continue
		}

		candidate := normalizeMatchPath(profile.FilePath)
		if candidate == "" {
			continue
		}

		if target == candidate || strings.HasSuffix(target, "/"+candidate) {
			return name, profile, true
		}
	}

	targetBase := filepath.Base(target)
	var matched config.Profile
	var matchedName string
	matches := 0
	for name, profile := range profiles {
		if profile.FilePath == "" || (profile.App == "" && profile.RulesPath == "") {
			continue
		}

		if filepath.Base(normalizeMatchPath(profile.FilePath)) == targetBase {
			matched = profile
			matchedName = name
			matches++
		}
	}
	if matches == 1 {
		return matchedName, matched, true
	}

	return "", config.Profile{}, false
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

func runList(opts Options, cfg config.Config, out io.Writer) error {
	switch opts.Subject {
	case "apps":
		names := builtinNames()
		for _, name := range names {
			_, _ = fmt.Fprintln(out, name)
		}
		return nil
	case "profiles":
		names := profileNames(cfg.Profiles)
		if len(names) == 0 {
			_, _ = fmt.Fprintln(out, "no profiles configured")
			return nil
		}
		for _, name := range names {
			_, _ = fmt.Fprintln(out, name)
		}
		return nil
	default:
		return fmt.Errorf("unknown list subject %q", opts.Subject)
	}
}

func runShow(opts Options, cfg config.Config, out io.Writer) error {
	switch opts.Subject {
	case "app":
		file, ok := rules.Builtin(opts.Name)
		if !ok {
			return rules.ErrUnknownBuiltin(opts.Name)
		}
		_, _ = fmt.Fprintf(out, "app: %s\n", opts.Name)
		if file.Command != "" {
			_, _ = fmt.Fprintf(out, "default command: %s\n", file.Command)
		} else {
			_, _ = fmt.Fprintln(out, "default command: none")
		}
		_, _ = fmt.Fprintf(out, "rules: %d\n", len(file.Rules))
		return nil
	case "profile":
		profile, ok := cfg.Profiles[opts.Name]
		if !ok {
			return fmt.Errorf("unknown profile %q", opts.Name)
		}
		_, _ = fmt.Fprintf(out, "profile: %s\n", opts.Name)
		if profile.App != "" {
			_, _ = fmt.Fprintf(out, "app: %s\n", profile.App)
		}
		if profile.RulesPath != "" {
			_, _ = fmt.Fprintf(out, "rules: %s\n", profile.RulesPath)
		}
		if profile.ThemePath != "" {
			_, _ = fmt.Fprintf(out, "theme: %s\n", profile.ThemePath)
		}
		if profile.FilePath != "" {
			_, _ = fmt.Fprintf(out, "file: %s\n", profile.FilePath)
		}
		return nil
	default:
		return fmt.Errorf("unknown show subject %q", opts.Subject)
	}
}

func runValidate(opts Options, cfg config.Config, cfgFound bool, out io.Writer) error {
	var issues []string
	var checks []string

	if cfgFound {
		checks = append(checks, filepath.Join(opts.ConfigDir, "config.yaml"))
		issues = append(issues, validateConfigAssets(cfg)...)
	} else if opts.App == "" && opts.RulesPath == "" && opts.ThemePath == "" {
		_, _ = fmt.Fprintln(out, "ok: no config file found and no explicit app/rules/theme provided")
		return nil
	}

	if opts.App != "" {
		checks = append(checks, "app:"+opts.App)
		issues = append(issues, validateBuiltin(opts.App)...)
	}
	if opts.RulesPath != "" {
		checks = append(checks, opts.RulesPath)
		issues = append(issues, validateRulesFile(opts.RulesPath)...)
	}
	if opts.ThemePath != "" {
		checks = append(checks, opts.ThemePath)
		issues = append(issues, validateThemeFile(opts.ThemePath)...)
	}

	if len(issues) > 0 {
		sort.Strings(issues)
		return errors.New(strings.Join(issues, "\n"))
	}

	if len(checks) == 0 {
		_, _ = fmt.Fprintln(out, "ok")
		return nil
	}

	sort.Strings(checks)
	_, _ = fmt.Fprintf(out, "ok: %s\n", strings.Join(checks, ", "))
	return nil
}

func validateConfigAssets(cfg config.Config) []string {
	var issues []string

	if cfg.RulesPath != "" {
		issues = append(issues, validateRulesFile(cfg.RulesPath)...)
	}
	if cfg.ThemePath != "" {
		issues = append(issues, validateThemeFile(cfg.ThemePath)...)
	}

	for name, profile := range cfg.Profiles {
		if profile.App != "" {
			for _, issue := range validateBuiltin(profile.App) {
				issues = append(issues, fmt.Sprintf("profile %q: %s", name, issue))
			}
		}
		if profile.RulesPath != "" {
			for _, issue := range validateRulesFile(profile.RulesPath) {
				issues = append(issues, fmt.Sprintf("profile %q: %s", name, issue))
			}
		}
		if profile.ThemePath != "" {
			for _, issue := range validateThemeFile(profile.ThemePath) {
				issues = append(issues, fmt.Sprintf("profile %q: %s", name, issue))
			}
		}
	}

	return issues
}

func validateBuiltin(name string) []string {
	file, ok := rules.Builtin(name)
	if !ok {
		return []string{fmt.Sprintf("unknown built-in app profile %q", name)}
	}

	compiled, err := rules.Compile(file.Rules)
	if err != nil {
		return []string{fmt.Sprintf("app %q: %v", name, err)}
	}
	rules.Close(compiled)
	return nil
}

func validateRulesFile(path string) []string {
	file, err := rules.Load(path)
	if err != nil {
		return []string{fmt.Sprintf("rules %q: %v", path, err)}
	}

	compiled, err := rules.Compile(file.Rules)
	if err != nil {
		return []string{fmt.Sprintf("rules %q: %v", path, err)}
	}
	rules.Close(compiled)
	return nil
}

func validateThemeFile(path string) []string {
	if _, err := theme.Load(path); err != nil {
		return []string{fmt.Sprintf("theme %q: %v", path, err)}
	}
	return nil
}

func builtinNames() []string {
	names := make([]string, 0, len(rules.Builtins()))
	for name := range rules.Builtins() {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func profileNames(profiles map[string]config.Profile) []string {
	names := make([]string, 0, len(profiles))
	for name := range profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
