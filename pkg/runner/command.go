package runner

import (
	"io"
	"os"
	"os/exec"
	"sync"
)

// RunCommand executes a shell command and forwards its output streams.
//
// The command runs through the user's configured shell when available so the
// behavior matches normal terminal usage for pipelines, redirects, and quoting.
// When a highlighter is present, stdout and stderr are processed line by line.
func RunCommand(command string, stdout io.Writer, stderr io.Writer, highlighter *Highlighter) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	cmd := exec.Command(shell, "-lc", command)
	cmd.Stdin = os.Stdin
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		errs <- processStream(stdoutPipe, stdout, highlighter)
	}()
	go func() {
		defer wg.Done()
		errs <- processStream(stderrPipe, stderr, highlighter)
	}()
	wg.Wait()
	close(errs)

	var streamErr error
	for err := range errs {
		if err != nil && streamErr == nil {
			streamErr = err
		}
	}
	if streamErr != nil {
		_ = cmd.Wait()
		return streamErr
	}

	return cmd.Wait()
}
