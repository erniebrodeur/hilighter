package cli

import (
	"runtime/debug"
	"testing"
)

func TestFormattedVersionPrefersLinkerOverride(t *testing.T) {
	previousVersion := Version
	previousReadBuildInfo := readBuildInfo
	t.Cleanup(func() {
		Version = previousVersion
		readBuildInfo = previousReadBuildInfo
	})
	Version = "2.0.0"
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "v1.0.0"}}, true
	}

	if got := formattedVersion(); got != "hilighter-2.0.0" {
		t.Fatalf("expected linker version, got %q", got)
	}
}

func TestFormattedVersionUsesEmbeddedModuleVersion(t *testing.T) {
	previousVersion := Version
	previousReadBuildInfo := readBuildInfo
	t.Cleanup(func() {
		Version = previousVersion
		readBuildInfo = previousReadBuildInfo
	})
	Version = "dev"
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "v1.2.3"}}, true
	}

	if got := formattedVersion(); got != "hilighter-1.2.3" {
		t.Fatalf("expected tagged module version, got %q", got)
	}
}

func TestFormattedVersionFallsBackToDev(t *testing.T) {
	previousVersion := Version
	previousReadBuildInfo := readBuildInfo
	t.Cleanup(func() {
		Version = previousVersion
		readBuildInfo = previousReadBuildInfo
	})
	Version = "dev"
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, true
	}

	if got := formattedVersion(); got != "hilighter-dev" {
		t.Fatalf("expected development version, got %q", got)
	}
}
