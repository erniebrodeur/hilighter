package runner

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

// ErrInput indicates that one or more input operands failed after their errors
// were already reported.
var ErrInput = errors.New("one or more input operands failed")

type readError struct {
	err error
}

func (e *readError) Error() string { return e.err.Error() }
func (e *readError) Unwrap() error { return e.err }

// RunStdin streams an input stream to an output stream.
//
// When a highlighter is provided, RunStdin preserves the original text content
// while applying ANSI styling to matching lines and spans.
func RunStdin(in io.Reader, out io.Writer, highlighter *Highlighter) error {
	return processStream(in, out, highlighter)
}

// RunInputs streams text operands in order, using stdin for each "-" operand.
// With no operands, it reads stdin. Input failures are reported immediately;
// later operands are still processed and ErrInput is returned afterward.
func RunInputs(stdin io.Reader, out, errOut io.Writer, highlighter *Highlighter, operands []string) error {
	if len(operands) == 0 {
		operands = []string{"-"}
	}

	failed := false
	for _, operand := range operands {
		var input io.ReadCloser
		if operand == "-" {
			input = io.NopCloser(stdin)
		} else {
			file, err := os.Open(operand)
			if err != nil {
				if reportErr := reportInputError(errOut, operand, err); reportErr != nil {
					return reportErr
				}
				failed = true
				continue
			}
			input = file
		}

		streamErr := processStream(input, out, highlighter)
		closeErr := input.Close()
		if streamErr != nil {
			var sourceErr *readError
			if !errors.As(streamErr, &sourceErr) {
				return streamErr
			}
			if reportErr := reportInputError(errOut, operand, sourceErr.err); reportErr != nil {
				return reportErr
			}
			failed = true
		}
		if closeErr != nil {
			if reportErr := reportInputError(errOut, operand, closeErr); reportErr != nil {
				return reportErr
			}
			failed = true
		}
	}

	if failed {
		return ErrInput
	}
	return nil
}

func reportInputError(out io.Writer, operand string, err error) error {
	_, writeErr := fmt.Fprintf(out, "%s: %v\n", operand, err)
	return writeErr
}

func processStream(in io.Reader, out io.Writer, highlighter *Highlighter) error {
	reader := bufio.NewReader(in)
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			hasNewline := line[len(line)-1] == '\n'
			content := line
			if hasNewline {
				content = line[:len(line)-1]
			}

			rendered := content
			if highlighter != nil {
				rendered = highlighter.ProcessLine(content)
			}
			if _, writeErr := io.WriteString(out, rendered); writeErr != nil {
				return writeErr
			}
			if hasNewline {
				if _, writeErr := io.WriteString(out, "\n"); writeErr != nil {
					return writeErr
				}
			}
		}

		if err == io.EOF {
			return nil
		}
		if err != nil {
			return &readError{err: err}
		}
	}
}
