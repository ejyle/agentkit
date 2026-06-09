package version

import "fmt"

// Version, GOOS, and GOARCH are set at build time by GoReleaser via ldflags:
//
//	-X github.com/ejyle/agentkit/internal/version.Version={{.Version}}
//	-X github.com/ejyle/agentkit/internal/version.GOOS={{.Os}}
//	-X github.com/ejyle/agentkit/internal/version.GOARCH={{.Arch}}
//
// Local builds without ldflags will use the default "dev"/"unknown" values.
var (
	Version = "dev"
	GOOS    = "unknown"
	GOARCH  = "unknown"
)

// String returns the version string in the format "agentkit/<version> (<os>/<arch>)".
func String() string {
	return fmt.Sprintf("agentkit/%s (%s/%s)", Version, GOOS, GOARCH)
}
