package repository

import (
	"strings"
	"testing"
)

// TestSplitProtectedTriggerSQL_SQLite verifies the sqlite3 constant splits into exactly
// 2 statements (one trigger each) and contains no LOCKSTEP markers.
func TestSplitProtectedTriggerSQL_SQLite(t *testing.T) {
	stmts := SplitProtectedTriggerSQL(SQLiteProtectedTriggers)

	if len(stmts) != 2 {
		t.Errorf("sqlite3: expected 2 statements, got %d: %v", len(stmts), stmts)
	}

	for _, s := range stmts {
		if strings.Contains(s, "LOCKSTEP") {
			t.Errorf("sqlite3: LOCKSTEP marker leaked into statement: %q", s)
		}
		if strings.HasPrefix(strings.TrimSpace(s), "--") {
			t.Errorf("sqlite3: comment line leaked into statement: %q", s)
		}
	}

	// Verify the two expected trigger names appear.
	if len(stmts) >= 1 && !strings.Contains(stmts[0], "protected_users_no_update") {
		t.Errorf("sqlite3: statement[0] missing protected_users_no_update: %q", stmts[0])
	}
	if len(stmts) >= 2 && !strings.Contains(stmts[1], "protected_users_no_delete") {
		t.Errorf("sqlite3: statement[1] missing protected_users_no_delete: %q", stmts[1])
	}
}

// TestSplitProtectedTriggerSQL_SQLite_NoIfNotExists confirms the sqlite3 constant uses
// plain CREATE TRIGGER (not CREATE TRIGGER IF NOT EXISTS) — per the spec, idempotency is
// provided by the migration system's applied-tracking, not by DDL qualifiers.
func TestSplitProtectedTriggerSQL_SQLite_NoIfNotExists(t *testing.T) {
	if strings.Contains(SQLiteProtectedTriggers, "CREATE TRIGGER IF NOT EXISTS") {
		t.Error("sqlite3 constant must use plain CREATE TRIGGER, not CREATE TRIGGER IF NOT EXISTS")
	}
}

// TestSplitProtectedTriggerSQL_SQLite_BeginEndNotSplit ensures the BEGIN...END block
// inside the sqlite3 trigger is not treated as multiple statements.
func TestSplitProtectedTriggerSQL_SQLite_BeginEndNotSplit(t *testing.T) {
	stmts := SplitProtectedTriggerSQL(SQLiteProtectedTriggers)
	// Each trigger statement should contain BEGIN and END.
	for i, s := range stmts {
		upper := strings.ToUpper(s)
		if !strings.Contains(upper, "BEGIN") {
			t.Errorf("sqlite3: statement[%d] missing BEGIN: %q", i, s)
		}
		if !strings.Contains(upper, "END") {
			t.Errorf("sqlite3: statement[%d] missing END: %q", i, s)
		}
	}
}

// TestSplitProtectedTriggerSQL_Postgres verifies the postgres constant splits into
// exactly 5 statements: 1 function, 2 DROP TRIGGER, 2 CREATE TRIGGER.
func TestSplitProtectedTriggerSQL_Postgres(t *testing.T) {
	stmts := SplitProtectedTriggerSQL(PostgresProtectedTriggers)

	if len(stmts) != 5 {
		t.Errorf("postgres: expected 5 statements, got %d", len(stmts))
		for i, s := range stmts {
			t.Logf("  [%d]: %q", i, s)
		}
	}

	for _, s := range stmts {
		if strings.Contains(s, "LOCKSTEP") {
			t.Errorf("postgres: LOCKSTEP marker leaked into statement: %q", s)
		}
	}
}

// TestSplitProtectedTriggerSQL_Postgres_DollarQuoteAtomic checks that semicolons inside
// the postgres dollar-quoted function body are not treated as statement delimiters.
func TestSplitProtectedTriggerSQL_Postgres_DollarQuoteAtomic(t *testing.T) {
	stmts := SplitProtectedTriggerSQL(PostgresProtectedTriggers)
	// The first statement should be the full CREATE OR REPLACE FUNCTION body,
	// including the $$ ... $$ dollar-quoted block.
	if len(stmts) == 0 {
		t.Fatal("postgres: no statements returned")
	}
	first := stmts[0]
	if !strings.Contains(first, "CREATE OR REPLACE FUNCTION") {
		t.Errorf("postgres: first statement should be function definition, got: %q", first)
	}
	if !strings.Contains(first, "$$") {
		t.Errorf("postgres: first statement should contain dollar-quote markers: %q", first)
	}
	// Should also contain LANGUAGE plpgsql (end of function), proving it wasn't split mid-body.
	if !strings.Contains(first, "plpgsql") {
		t.Errorf("postgres: function body was split — 'plpgsql' missing from first statement: %q", first)
	}
}

// TestSplitProtectedTriggerSQL_MySQL verifies the mysql constant splits into exactly
// 4 statements: 2 DROP TRIGGER + 2 CREATE TRIGGER.
func TestSplitProtectedTriggerSQL_MySQL(t *testing.T) {
	stmts := SplitProtectedTriggerSQL(MySQLProtectedTriggers)

	if len(stmts) != 4 {
		t.Errorf("mysql: expected 4 statements, got %d", len(stmts))
		for i, s := range stmts {
			t.Logf("  [%d]: %q", i, s)
		}
	}

	for _, s := range stmts {
		if strings.Contains(s, "LOCKSTEP") {
			t.Errorf("mysql: LOCKSTEP marker leaked into statement: %q", s)
		}
	}
}

// TestSplitProtectedTriggerSQL_MySQL_BeginEndAtomic checks that BEGIN...END in mysql
// triggers is treated as an atomic block (semicolons inside don't split).
func TestSplitProtectedTriggerSQL_MySQL_BeginEndAtomic(t *testing.T) {
	stmts := SplitProtectedTriggerSQL(MySQLProtectedTriggers)
	// Statements at index 1 and 3 should be the CREATE TRIGGER bodies containing BEGIN/END.
	for _, idx := range []int{1, 3} {
		if idx >= len(stmts) {
			break
		}
		s := strings.ToUpper(stmts[idx])
		if !strings.Contains(s, "BEGIN") || !strings.Contains(s, "END") {
			t.Errorf("mysql: statement[%d] should contain full BEGIN/END block, got: %q", idx, stmts[idx])
		}
	}
}

// TestSplitProtectedTriggerSQL_LockstepMarkersAbsent verifies that LOCKSTEP markers
// are not included in any returned statement across all dialects.
func TestSplitProtectedTriggerSQL_LockstepMarkersAbsent(t *testing.T) {
	cases := []struct {
		name     string
		constant string
	}{
		{"sqlite3", SQLiteProtectedTriggers},
		{"postgres", PostgresProtectedTriggers},
		{"mysql", MySQLProtectedTriggers},
	}
	for _, tc := range cases {
		for i, s := range SplitProtectedTriggerSQL(tc.constant) {
			if strings.Contains(s, "LOCKSTEP-START") || strings.Contains(s, "LOCKSTEP-END") {
				t.Errorf("%s statement[%d]: contains LOCKSTEP marker: %q", tc.name, i, s)
			}
		}
	}
}

// TestSplitProtectedTriggerSQL_EdgeCases confirms that empty, comment-only, marker-only,
// and whitespace-only inputs all return zero statements.
func TestSplitProtectedTriggerSQL_EdgeCases(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"empty input", ""},
		{"only comments", "-- one\n-- two\n"},
		{"only LOCKSTEP markers", "-- LOCKSTEP-START sqlite\n-- LOCKSTEP-END sqlite\n"},
		{"whitespace only", "   \n\t\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := SplitProtectedTriggerSQL(tc.in)
			if len(out) != 0 {
				t.Errorf("expected empty result, got %d statements: %v", len(out), out)
			}
		})
	}
}

// TestSplitProtectedTriggerSQL_ErrorMessagePresent confirms the L4 contract string is
// present in at least one statement per dialect.
func TestSplitProtectedTriggerSQL_ErrorMessagePresent(t *testing.T) {
	const l4Contract = "protected user: writes blocked at db layer"
	cases := []struct {
		name     string
		constant string
	}{
		{"sqlite3", SQLiteProtectedTriggers},
		{"postgres", PostgresProtectedTriggers},
		{"mysql", MySQLProtectedTriggers},
	}
	for _, tc := range cases {
		found := false
		for _, s := range SplitProtectedTriggerSQL(tc.constant) {
			if strings.Contains(s, l4Contract) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s: L4 contract string %q not found in any statement", tc.name, l4Contract)
		}
	}
}
