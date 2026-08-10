package cli

import (
	"fmt"
	"runtime/debug"
	"strings"
)

// Version is the build-time application version marker.
//
// Release builds can override this with:
// go build -ldflags "-X github.com/erniebrodeur/hilighter/internal/cli.Version=1.0.0"
var Version = "dev"

var readBuildInfo = debug.ReadBuildInfo

func formattedVersion() string {
	return fmt.Sprintf("hilighter-%s", resolvedVersion())
}

func resolvedVersion() string {
	if Version != "" && Version != "dev" {
		return strings.TrimPrefix(Version, "v")
	}

	info, ok := readBuildInfo()
	if ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return strings.TrimPrefix(info.Main.Version, "v")
	}
	return "dev"
}
