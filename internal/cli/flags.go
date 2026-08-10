package cli

import (
	"fmt"
	"strings"
)

// Options carries the complete public CLI grammar into the application layer.
type Options struct {
	ShowHelp    bool
	ShowVersion bool
	Expressions []string
	ThemePath   string
	Inputs      []string
}

func parseOptions(args []string) (Options, error) {
	opts := Options{}
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--":
			opts.Inputs = append([]string(nil), args[index+1:]...)
			return validateOptions(opts)
		case "--help":
			opts.ShowHelp = true
		case "--version":
			opts.ShowVersion = true
		case "-e":
			index++
			if index >= len(args) {
				return Options{}, fmt.Errorf("-e requires a pattern")
			}
			opts.Expressions = append(opts.Expressions, args[index])
		case "-t":
			index++
			if index >= len(args) {
				return Options{}, fmt.Errorf("-t requires a theme")
			}
			opts.ThemePath = args[index]
		default:
			if args[index] != "-" && strings.HasPrefix(args[index], "-") {
				return Options{}, fmt.Errorf("unknown option %q", args[index])
			}
			opts.Inputs = append([]string(nil), args[index:]...)
			return validateOptions(opts)
		}
	}
	return validateOptions(opts)
}

func validateOptions(opts Options) (Options, error) {
	if opts.ShowHelp && opts.ShowVersion {
		return Options{}, fmt.Errorf("--help and --version cannot be combined")
	}
	if (opts.ShowHelp || opts.ShowVersion) && (len(opts.Expressions) > 0 || opts.ThemePath != "" || len(opts.Inputs) > 0) {
		return Options{}, fmt.Errorf("--help and --version cannot be combined with processing options or file operands")
	}
	return opts, nil
}
