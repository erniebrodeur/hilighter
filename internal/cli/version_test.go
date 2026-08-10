package cli

import "testing"

func TestFormattedVersionUsesSourceVersion(t *testing.T) {
	if Version != "1.0.0" {
		t.Fatalf("expected source version 1.0.0, got %q", Version)
	}
	if got := formattedVersion(); got != "hilighter-1.0.0" {
		t.Fatalf("expected source version output, got %q", got)
	}
}
