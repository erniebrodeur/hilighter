package cli

import "testing"

func TestFormattedVersionUsesSourceVersion(t *testing.T) {
	if Version == "" {
		t.Fatal("source version must not be empty")
	}
	if got := formattedVersion(); got != "hilighter-"+Version {
		t.Fatalf("expected source version output, got %q", got)
	}
}
