package security_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/johnzastrow/actalog/pkg/security"
)

func TestIsProtectedEmail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "ExactMatch",
			input: "br8kwall@gmail.com",
			want:  true,
		},
		{
			name:  "CaseInsensitive_AllUpper",
			input: "BR8KWALL@GMAIL.COM",
			want:  true,
		},
		{
			name:  "CaseInsensitive_MixedCase",
			input: "Br8kWall@Gmail.com",
			want:  true,
		},
		{
			name:  "CaseInsensitive_UpperDomain",
			input: "br8kwall@GMAIL.COM",
			want:  true,
		},
		{
			name:  "TrimsWhitespace",
			input: "  br8kwall@gmail.com  ",
			want:  true,
		},
		{
			name:  "DoesNotMatchPlusAddressing",
			input: "br8kwall+test@gmail.com",
			want:  false,
		},
		{
			name:  "NonProtectedReturnsFalse",
			input: "alice@example.com",
			want:  false,
		},
		{
			name:  "EmptyReturnsFalse",
			input: "",
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := security.IsProtectedEmail(tc.input)
			if got != tc.want {
				t.Errorf("IsProtectedEmail(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

// TestLowercaseInvariant verifies every entry returned by ProtectedEmailsList
// is lowercase, ensuring the ToLower-on-input lookup strategy cannot miss a
// typo in the registry.
func TestLowercaseInvariant(t *testing.T) {
	for _, email := range security.ProtectedEmailsList() {
		if email != strings.ToLower(email) {
			t.Errorf("protected email %q is not all lowercase", email)
		}
	}
}

// TestProtectedEmailsList_Sorted verifies that ProtectedEmailsList returns a
// non-empty slice in lexicographic order. Sorting is part of the contract so
// that callers (e.g. the Task 3 frontend-guard generator) can rely on
// deterministic output without sorting on every call.
func TestProtectedEmailsList_Sorted(t *testing.T) {
	list := security.ProtectedEmailsList()

	if len(list) == 0 {
		t.Fatal("ProtectedEmailsList() returned empty slice; at least one protected email must be registered")
	}

	// Verify the known entry is present.
	found := false
	for _, e := range list {
		if e == "br8kwall@gmail.com" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ProtectedEmailsList() does not contain expected entry br8kwall@gmail.com")
	}

	// Verify lexicographic order (contract: callers must not need to sort).
	if !sort.StringsAreSorted(list) {
		t.Errorf("ProtectedEmailsList() is not sorted: %v", list)
	}
}
