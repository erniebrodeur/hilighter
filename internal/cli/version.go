package cli

import "fmt"

// Version is the build-time application version marker.
//
// Release builds can override this with:
// go build -ldflags "-X github.com/erniebrodeur/hilighter/internal/cli.Version=1.0.0"
var Version = "dev"

func formattedVersion() string {
	return fmt.Sprintf("hilighter-%s", Version)
}
