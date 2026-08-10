package main

import (
	"bytes"
	"errors"
	"syscall"
	"testing"
)

func TestRunMainSilencesBrokenPipeAndReturnsFailure(t *testing.T) {
	var stderr bytes.Buffer

	status := runMain(func() error { return syscall.EPIPE }, &stderr)

	if status != 1 {
		t.Fatalf("expected status 1, got %d", status)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no diagnostic, got %q", stderr.String())
	}
}

func TestRunMainReportsOrdinaryErrorsAndReturnsFailure(t *testing.T) {
	var stderr bytes.Buffer
	writeErr := errors.New("write failed")

	status := runMain(func() error { return writeErr }, &stderr)

	if status != 1 {
		t.Fatalf("expected status 1, got %d", status)
	}
	if stderr.String() != "write failed\n" {
		t.Fatalf("expected ordinary diagnostic, got %q", stderr.String())
	}
}

func TestRunMainReturnsSuccessWithoutDiagnostic(t *testing.T) {
	var stderr bytes.Buffer

	status := runMain(func() error { return nil }, &stderr)

	if status != 0 {
		t.Fatalf("expected status 0, got %d", status)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no diagnostic, got %q", stderr.String())
	}
}
