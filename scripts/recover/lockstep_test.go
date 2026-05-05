// Package recovery_test verifies that the standalone per-dialect recovery SQL
// files under scripts/recover/sql/ stay byte-for-byte in sync with the
// LOCKSTEP-bracketed bodies inside the migration source constant.
//
// If this test fails it means someone edited one side without updating the
// other. Fix: re-extract the SQL body from internal/repository/protected_triggers_sql.go
// and paste it into the matching scripts/recover/sql/<dialect>/protected_users.sql.
package recovery_test

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// dialects lists every dialect that has a LOCKSTEP block in the migration
// source and a corresponding recovery script.
var dialects = []string{"sqlite", "postgres", "mysql"}

func TestLockstep_RecoveryScriptsMatchMigration(t *testing.T) {
	root := findProjectRoot(t)

	migrationSrc := mustReadProjectFile(t, root, "internal/repository/protected_triggers_sql.go")

	for _, dialect := range dialects {
		dialect := dialect // capture for t.Run closure
		t.Run(dialect, func(t *testing.T) {
			migrationBody := extractLockstepBody(migrationSrc, dialect)
			if migrationBody == "" {
				t.Fatalf("dialect %q: no LOCKSTEP block found in migration source", dialect)
			}

			scriptPath := filepath.Join("sql", dialect, "protected_users.sql")
			scriptSrc := mustReadProjectFile(t, root, filepath.Join("scripts", "recover", scriptPath))

			normMigration := normalise(migrationBody)
			normScript := normalise(scriptSrc)

			if normMigration != normScript {
				t.Errorf("dialect %q: recovery script does not match migration constant.\n"+
					"--- migration body (normalised) ---\n%s\n"+
					"--- recovery script (normalised) ---\n%s",
					dialect, normMigration, normScript)
			}
		})
	}
}

// findProjectRoot walks up from the directory of this test file until it
// finds a directory containing go.mod, which is the project root.
func findProjectRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find project root (go.mod) starting from %s", filepath.Dir(filename))
		}
		dir = parent
	}
}

// mustReadProjectFile reads a file at path relative to projectRoot and returns
// its contents as a string. It calls t.Fatal on any error.
func mustReadProjectFile(t *testing.T, projectRoot, rel string) string {
	t.Helper()
	full := filepath.Join(projectRoot, rel)
	data, err := os.ReadFile(full)
	if err != nil {
		t.Fatalf("mustReadProjectFile: %v", err)
	}
	return string(data)
}

// extractLockstepBody extracts the text between
//
//	-- LOCKSTEP-START <dialect>
//	-- LOCKSTEP-END <dialect>
//
// in source. The marker lines themselves are not included in the result.
func extractLockstepBody(source, dialect string) string {
	startMarker := "-- LOCKSTEP-START " + dialect
	endMarker := "-- LOCKSTEP-END " + dialect

	startIdx := strings.Index(source, startMarker)
	if startIdx == -1 {
		return ""
	}
	// Advance past the start-marker line.
	afterStart := source[startIdx+len(startMarker):]
	newline := strings.Index(afterStart, "\n")
	if newline != -1 {
		afterStart = afterStart[newline+1:]
	}

	endIdx := strings.Index(afterStart, endMarker)
	if endIdx == -1 {
		return ""
	}
	return afterStart[:endIdx]
}

var (
	reBlankLine   = regexp.MustCompile(`(?m)^\s*$`)
	reCommentLine = regexp.MustCompile(`(?m)^\s*--[^\n]*$`)
	reWhitespace  = regexp.MustCompile(`[ \t]+`)
)

// normalise strips blank lines, strips SQL line-comment lines, collapses
// runs of horizontal whitespace to a single space, and trims surrounding
// whitespace. This makes the comparison immune to indentation style and
// trailing-space differences.
func normalise(s string) string {
	// Strip comment-only lines.
	s = reCommentLine.ReplaceAllString(s, "")
	// Strip blank / whitespace-only lines.
	s = reBlankLine.ReplaceAllString(s, "")
	// Collapse runs of horizontal whitespace.
	s = reWhitespace.ReplaceAllString(s, " ")
	// Normalise line endings and trim.
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	var out []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}
