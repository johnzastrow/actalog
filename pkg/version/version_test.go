package version

import (
	"fmt"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	version := Version()

	// Version should contain major.minor.patch
	expected := fmt.Sprintf("%d.%d.%d", Major, Minor, Patch)
	if !strings.HasPrefix(version, expected) {
		t.Errorf("Version() = %q, want prefix %q", version, expected)
	}

	// If PreRelease is set, it should be included
	if PreRelease != "" {
		expectedWithPreRelease := fmt.Sprintf("%d.%d.%d-%s", Major, Minor, Patch, PreRelease)
		if version != expectedWithPreRelease {
			t.Errorf("Version() = %q, want %q", version, expectedWithPreRelease)
		}
	}
}

func TestFullVersion(t *testing.T) {
	fullVersion := FullVersion()

	// Should contain version info
	if !strings.Contains(fullVersion, fmt.Sprintf("%d.%d.%d", Major, Minor, Patch)) {
		t.Errorf("FullVersion() = %q, should contain version numbers", fullVersion)
	}

	// Should contain build number
	expectedBuild := fmt.Sprintf("build.%d", Build)
	if !strings.Contains(fullVersion, expectedBuild) {
		t.Errorf("FullVersion() = %q, should contain %q", fullVersion, expectedBuild)
	}

	// If PreRelease is set, it should be included
	if PreRelease != "" {
		if !strings.Contains(fullVersion, PreRelease) {
			t.Errorf("FullVersion() = %q, should contain prerelease %q", fullVersion, PreRelease)
		}
	}
}

func TestBuildNumber(t *testing.T) {
	buildNum := BuildNumber()

	if buildNum != Build {
		t.Errorf("BuildNumber() = %d, want %d", buildNum, Build)
	}

	// Build number should be positive
	if buildNum < 0 {
		t.Errorf("BuildNumber() = %d, should be non-negative", buildNum)
	}
}

func TestString(t *testing.T) {
	str := String()

	// Should start with "ActaLog v"
	if !strings.HasPrefix(str, "ActaLog v") {
		t.Errorf("String() = %q, should start with 'ActaLog v'", str)
	}

	// Should contain the version
	if !strings.Contains(str, Version()) {
		t.Errorf("String() = %q, should contain version %q", str, Version())
	}
}

func TestFullString(t *testing.T) {
	fullStr := FullString()

	// Should start with "ActaLog v"
	if !strings.HasPrefix(fullStr, "ActaLog v") {
		t.Errorf("FullString() = %q, should start with 'ActaLog v'", fullStr)
	}

	// Should contain the full version
	if !strings.Contains(fullStr, FullVersion()) {
		t.Errorf("FullString() = %q, should contain full version %q", fullStr, FullVersion())
	}
}

func TestVersionFormat(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		checkFn  func(string) bool
		wantPass bool
	}{
		{
			name:    "Version is semver format",
			version: Version(),
			checkFn: func(v string) bool {
				// Basic semver pattern check
				parts := strings.Split(strings.Split(v, "-")[0], ".")
				return len(parts) == 3
			},
			wantPass: true,
		},
		{
			name:    "FullVersion has build suffix",
			version: FullVersion(),
			checkFn: func(v string) bool {
				return strings.Contains(v, "+build.")
			},
			wantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.checkFn(tt.version); got != tt.wantPass {
				t.Errorf("check for %q = %v, want %v", tt.version, got, tt.wantPass)
			}
		})
	}
}

func TestVersionConstants(t *testing.T) {
	// Version constants should be reasonable values
	if Major < 0 {
		t.Errorf("Major = %d, should be non-negative", Major)
	}
	if Minor < 0 {
		t.Errorf("Minor = %d, should be non-negative", Minor)
	}
	if Patch < 0 {
		t.Errorf("Patch = %d, should be non-negative", Patch)
	}
	if Build < 0 {
		t.Errorf("Build = %d, should be non-negative", Build)
	}
}
