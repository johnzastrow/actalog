// Package version provides version information for ActaLog
package version

import "fmt"

const (
	// Major version number
	Major = 0
	// Minor version number
	Minor = 2
	// Patch version number
	Patch = 0
	// PreRelease identifier (e.g., "alpha", "beta", "rc1")
	PreRelease = "beta"
)

// Version returns the full semantic version string
func Version() string {
	if PreRelease != "" {
		return fmt.Sprintf("%d.%d.%d-%s", Major, Minor, Patch, PreRelease)
	}
	return fmt.Sprintf("%d.%d.%d", Major, Minor, Patch)
}

// String returns the version with application name
func String() string {
	return fmt.Sprintf("ActaLog v%s", Version())
}
