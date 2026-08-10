package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/erniebrodeur/hilighter/pkg/config"
	"github.com/erniebrodeur/hilighter/pkg/runner"
)

type reportedError struct {
	err error
}

func (e *reportedError) Error() string { return e.err.Error() }
func (e *reportedError) Unwrap() error { return e.err }

// ErrorReported reports whether an error was already written to stderr.
func ErrorReported(err error) bool {
	var reported *reportedError
	return errors.As(err, &reported)
}

// Main is the top-level CLI entrypoint.
func Main() error {
	return run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, config.DefaultDir())
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer, configDir string) error {
	opts, err := parseOptions(args)
	if err != nil {
		return err
	}
	if err := config.EnsureLayout(configDir); err != nil {
		return err
	}
	if opts.ShowHelp {
		printHelp(stdout)
		return nil
	}
	if opts.ShowVersion {
		_, err := fmt.Fprintln(stdout, formattedVersion())
		return err
	}

	rulesPath, themePath, err := resolveOptions(opts, configDir)
	if err != nil {
		return err
	}
	highlighter, err := runner.NewHighlighterWithExpressions(opts.Expressions, rulesPath, themePath)
	if err != nil {
		return err
	}
	defer highlighter.Close()

	err = runner.RunInputs(stdin, stdout, stderr, highlighter, opts.Inputs)
	if errors.Is(err, runner.ErrInput) {
		return &reportedError{err: err}
	}
	return err
}

func resolveOptions(opts Options, configDir string) (rulesPath, themePath string, err error) {
	if len(opts.Expressions) == 0 {
		rulesPath = config.DefaultRulesPath(configDir)
		nonEmpty, readErr := fileHasContent(rulesPath)
		if readErr != nil {
			return "", "", readErr
		}
		if !nonEmpty {
			rulesPath = ""
		}
	}

	themePath = opts.ThemePath
	if themePath == "" {
		themePath, err = config.LoadTheme(filepath.Join(configDir, "config.yaml"))
		if err != nil {
			return "", "", err
		}
	}
	if themePath == "monokai" {
		themePath = ""
	} else if themePath != "" && !filepath.IsAbs(themePath) {
		themePath = filepath.Join(configDir, themePath)
	}
	return rulesPath, themePath, nil
}

func fileHasContent(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	return len(bytes.TrimSpace(data)) > 0, nil
}

func printHelp(out io.Writer) {
	_, _ = fmt.Fprint(out, `Usage:
  hilighter [-e <pattern> ...] [-t <theme>] [--] [file ...]
  hilighter --help
  hilighter --version
`)
}
