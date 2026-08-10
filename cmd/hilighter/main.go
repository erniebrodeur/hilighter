package main

import (
	"fmt"
	"io"
	"os"

	"github.com/erniebrodeur/hilighter/internal/cli"
)

func main() {
	os.Exit(runMain(cli.Main, os.Stderr))
}

func runMain(run func() error, stderr io.Writer) int {
	err := run()
	if err == nil {
		return 0
	}
	if !cli.SuppressError(err) {
		_, _ = fmt.Fprintln(stderr, err)
	}
	return 1
}
