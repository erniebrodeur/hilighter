package runner

import (
	"io"
)

// RunStdin copies an input stream to an output stream.
//
// The current scaffold is a transparent passthrough. Later slices will insert
// rule evaluation and ANSI rendering while preserving the original text content.
func RunStdin(in io.Reader, out io.Writer) error {
	_, err := io.Copy(out, in)
	return err
}
