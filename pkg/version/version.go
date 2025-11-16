// Package version provides version information for ActaLog
package version

import "fmt"

const (
	// Major version number
	Major = 0
	// Minor version number
	Minor = 4
	// Patch version number
	Patch = 6
	// PreRelease identifier (e.g., "alpha", "beta", "rc1")
	PreRelease = "beta"
	// Build number - increment this with each code change
	Build = 1
)

// Version returns the full semantic version string
func Version() string {
	if PreRelease != "" {
		return fmt.Sprintf("%d.%d.%d-%s", Major, Minor, Patch, PreRelease)
	}
	return fmt.Sprintf("%d.%d.%d", Major, Minor, Patch)
}

// FullVersion returns the version with build number
func FullVersion() string {
	if PreRelease != "" {
		return fmt.Sprintf("%d.%d.%d-%s+build.%d", Major, Minor, Patch, PreRelease, Build)
	}
	return fmt.Sprintf("%d.%d.%d+build.%d", Major, Minor, Patch, Build)
}

// BuildNumber returns just the build number
func BuildNumber() int {
	return Build
}

// String returns the version with application name
func String() string {
	return fmt.Sprintf("ActaLog v%s", Version())
}

// FullString returns the version with application name and build
func FullString() string {
	return fmt.Sprintf("ActaLog v%s", FullVersion())
}
