package main

import (
	"strings"
	"testing"
)

// TestGenerate_ProducesEs6ModuleWithSetAndHelper verifies that generate returns
// a well-formed ES6 module containing the expected header comment, a named Set
// export, all seeded emails, and the isProtectedEmail helper.
func TestGenerate_ProducesEs6ModuleWithSetAndHelper(t *testing.T) {
	emails := []string{"alice@example.com", "bob@example.com"}
	out := generate(emails)

	checks := []struct {
		name string
		want string
	}{
		{"header source path", "pkg/security/protected_users.go"},
		{"header do-not-edit warning", "do not edit by hand"},
		{"header regen instruction", "make gen-protected-emails"},
		{"header CI drift note", "CI fails on drift"},
		{"Set export declaration", "export const PROTECTED_EMAILS = new Set(["},
		{"alice email entry", "'alice@example.com',"},
		{"bob email entry", "'bob@example.com',"},
		{"Set closing bracket", "])"},
		{"function export", "export function isProtectedEmail(email)"},
		{"falsy guard", "if (!email) return false"},
		{"set lookup", "PROTECTED_EMAILS.has("},
		{"trim lowercase", "String(email).trim().toLowerCase()"},
	}

	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if !strings.Contains(out, c.want) {
				t.Errorf("output missing %q\n\nFull output:\n%s", c.want, out)
			}
		})
	}

	if !strings.HasSuffix(out, "\n") {
		t.Error("output must end with a newline")
	}
}

// TestGenerate_DeterministicOrder verifies that calling generate 20 times with
// the same input slice produces byte-identical output every time.
func TestGenerate_DeterministicOrder(t *testing.T) {
	emails := []string{"charlie@example.com", "alice@example.com", "bob@example.com"}
	// Three sample emails in arbitrary order — the determinism property holds
	// regardless of input ordering (same input → same output on every call).
	first := generate(emails)
	for i := 1; i < 20; i++ {
		got := generate(emails)
		if got != first {
			t.Fatalf("call %d produced different output than call 0", i)
		}
	}
}

// TestGenerate_PreservesInputOrder verifies that generate is a pure pass-through
// with respect to ordering — it must NOT sort its input. Sorting responsibility
// belongs to the security package (pkg/security), not this generator.
func TestGenerate_PreservesInputOrder(t *testing.T) {
	// Deliberately non-sorted: zebra before alpha.
	emails := []string{"zebra@example.com", "alpha@example.com"}
	out := generate(emails)

	zebraPos := strings.Index(out, "'zebra@example.com'")
	alphaPos := strings.Index(out, "'alpha@example.com'")
	if zebraPos == -1 || alphaPos == -1 {
		t.Fatalf("emails missing from output: zebra=%d alpha=%d", zebraPos, alphaPos)
	}
	if zebraPos > alphaPos {
		t.Errorf("generate() must preserve input order; got alpha at %d, zebra at %d", alphaPos, zebraPos)
	}
}
