// Command gen-protected-emails generates web/src/utils/protectedUsers.js from
// the Go source of truth in pkg/security/protected_users.go.
//
// Run via:
//
//	make gen-protected-emails
//	go run ./cmd/gen-protected-emails/
//
// The output file is committed to the repository. CI enforces that the JS file
// stays in sync with the Go source; if they drift the build fails.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/johnzastrow/actalog/pkg/security"
)

// outputPath is the path to the generated JS file, relative to the project root.
const outputPath = "web/src/utils/protectedUsers.js"

// generate converts a pre-sorted slice of protected email addresses into the
// text of an ES6 module that exports a Set and an isProtectedEmail helper.
// The caller is responsible for sorting; generate is a mechanical conversion.
func generate(emails []string) string {
	var b strings.Builder

	b.WriteString("// Code generated from pkg/security/protected_users.go — do not edit by hand;\n")
	b.WriteString("// run `make gen-protected-emails` to regenerate.\n")
	b.WriteString("// CI fails on drift between this file and the Go source.\n")
	b.WriteString("\n")
	b.WriteString("export const PROTECTED_EMAILS = new Set([\n")
	for _, email := range emails {
		b.WriteString("  '")
		b.WriteString(email)
		b.WriteString("',\n")
	}
	b.WriteString("])\n")
	b.WriteString("\n")
	b.WriteString("export function isProtectedEmail(email) {\n")
	b.WriteString("  if (!email) return false\n")
	b.WriteString("  return PROTECTED_EMAILS.has(String(email).trim().toLowerCase())\n")
	b.WriteString("}\n")

	return b.String()
}

// main calls security.ProtectedEmailsList() and writes the generated JS module
// to web/src/utils/protectedUsers.js relative to the current working directory.
func main() {
	emails := security.ProtectedEmailsList()
	out := generate(emails)

	if err := os.WriteFile(outputPath, []byte(out), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "gen-protected-emails: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("wrote %s (%d bytes)\n", outputPath, len(out))
}
