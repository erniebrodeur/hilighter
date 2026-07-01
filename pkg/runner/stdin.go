package runner

import (
	"bufio"
	"io"
)

// RunStdin streams an input stream to an output stream.
//
// When a highlighter is provided, RunStdin preserves the original text content
// while applying ANSI styling to matching lines and spans.
func RunStdin(in io.Reader, out io.Writer, highlighter *Highlighter) error {
	return processStream(in, out, highlighter)
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
			return err
		}
	}
}
