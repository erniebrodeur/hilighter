package main

import (
	"fmt"
	"os"

	"github.com/erniebrodeur/hilighter/internal/cli"
)

func main() {
	if err := cli.Main(); err != nil {
		if !cli.ErrorReported(err) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}
