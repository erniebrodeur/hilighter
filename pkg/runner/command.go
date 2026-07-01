package runner

import (
	"io"
	"os"
	"os/exec"
)

// RunCommand executes a shell command and forwards its output streams.
func RunCommand(command string, stdout io.Writer, stderr io.Writer) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	cmd := exec.Command(shell, "-lc", command)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
