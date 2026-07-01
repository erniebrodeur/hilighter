package cli

import (
	"os"

	"github.com/erniebrodeur/hilighter/pkg/runner"
)

// Main is the top-level CLI entrypoint.
func Main() error {
	opts := parseOptions()

	if opts.Command != "" {
		return runner.RunCommand(opts.Command, os.Stdout, os.Stderr)
	}

	return runner.RunStdin(os.Stdin, os.Stdout)
}
