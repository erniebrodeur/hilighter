package runner

import (
	"io"
)

// RunStdin copies stdin to stdout. Highlighting is wired in later slices.
func RunStdin(in io.Reader, out io.Writer) error {
	_, err := io.Copy(out, in)
	return err
}
