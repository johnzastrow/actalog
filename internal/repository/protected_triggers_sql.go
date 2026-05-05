package repository

import "strings"

// Package repository — protected_triggers_sql.go
//
// This file holds the per-dialect SQL that installs BEFORE UPDATE / BEFORE DELETE
// triggers on the users table to block writes to protected accounts at the database
// layer (L3 in the security model).
//
// CONTRACT — the error message text must be byte-identical across all three dialects:
//
//	"protected user: writes blocked at db layer"
//
// This exact string is the L4 contract. The service-layer error wrapper (Task 11)
// pattern-matches on this substring to detect trigger rejections and translate them
// into domain errors. Any drift breaks L4 detection.
//
// LOCKSTEP markers allow Task 7's CI check to verify that the SQL bodies in this
// file match the standalone recovery scripts under docs/security/.

// SplitProtectedTriggerSQL splits a per-dialect protected-trigger constant into individual
// statements that can be executed one at a time. It handles:
//   - SQLite: BEGIN/END blocks (semicolons inside are not statement delimiters)
//   - PostgreSQL: dollar-quoted strings ($$...$$) are atomic; semicolons inside don't split
//   - MySQL: BEGIN/END blocks; "END IF" and "END LOOP" etc. do NOT close a BEGIN block
//   - SQL line comments (-- ...) including LOCKSTEP markers are stripped from output
//
// Returns a slice of trimmed, non-empty statements ready for db.Exec().
func SplitProtectedTriggerSQL(constant string) []string {
	var stmts []string
	var buf strings.Builder
	depth := 0             // BEGIN/END nesting depth
	inDollarQuote := false // inside $$ ... $$ block (postgres)
	lines := strings.Split(constant, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip pure comment lines (including LOCKSTEP markers) — they are not statements.
		if strings.HasPrefix(trimmed, "--") {
			continue
		}

		// Walk character by character to track state and split on top-level semicolons.
		i := 0
		for i < len(line) {
			ch := line[i]

			// Detect inline comment start (outside dollar-quotes) — skip rest of line.
			if !inDollarQuote && i+1 < len(line) && ch == '-' && line[i+1] == '-' {
				break
			}

			// Detect dollar-quote toggle (postgres $$).
			// Dollar-quoting: this implementation handles only bare $$ pairs, which
			// is sufficient for the current postgres constant. If a future change
			// introduces named dollar-quoting ($tag$...$tag$), this needs to be
			// extended to track the tag name.
			if !inDollarQuote && i+1 < len(line) && ch == '$' && line[i+1] == '$' {
				inDollarQuote = true
				buf.WriteByte(ch)
				buf.WriteByte(line[i+1])
				i += 2
				continue
			}
			if inDollarQuote && i+1 < len(line) && ch == '$' && line[i+1] == '$' {
				inDollarQuote = false
				buf.WriteByte(ch)
				buf.WriteByte(line[i+1])
				i += 2
				continue
			}

			// Track BEGIN/END depth (case-insensitive word-boundary match) outside dollar-quotes.
			// "END IF", "END LOOP", "END CASE" etc. are compound keywords that do NOT close
			// a BEGIN block — only bare "END" (possibly followed by ; or whitespace) does.
			if !inDollarQuote {
				rest := strings.ToUpper(line[i:])
				if strings.HasPrefix(rest, "BEGIN") && isWordBoundary(line, i, 5) {
					depth++
				} else if strings.HasPrefix(rest, "END") && isWordBoundary(line, i, 3) {
					// Look ahead past whitespace to see if it's "END IF/LOOP/CASE/etc."
					afterEnd := strings.TrimLeft(rest[3:], " \t")
					compound := strings.HasPrefix(afterEnd, "IF") ||
						strings.HasPrefix(afterEnd, "LOOP") ||
						strings.HasPrefix(afterEnd, "CASE") ||
						strings.HasPrefix(afterEnd, "REPEAT") ||
						strings.HasPrefix(afterEnd, "WHILE")
					if !compound && depth > 0 {
						depth--
					}
				}
			}

			// At depth 0 and outside dollar-quotes, a semicolon ends a statement.
			if !inDollarQuote && depth == 0 && ch == ';' {
				stmt := strings.TrimSpace(buf.String())
				if stmt != "" {
					stmts = append(stmts, stmt)
				}
				buf.Reset()
				i++
				continue
			}

			buf.WriteByte(ch)
			i++
		}

		// Add a newline to preserve formatting between lines (the line loop stripped them).
		buf.WriteByte('\n')
	}

	// Emit any remaining content (statement without trailing semicolon).
	if stmt := strings.TrimSpace(buf.String()); stmt != "" {
		stmts = append(stmts, stmt)
	}

	return stmts
}

// isWordBoundary returns true if the keyword of length kwLen starting at pos is followed
// by a non-letter/non-digit/non-underscore character (or end-of-string).
func isWordBoundary(s string, pos, kwLen int) bool {
	end := pos + kwLen
	if end >= len(s) {
		return true
	}
	next := s[end]
	return !((next >= 'a' && next <= 'z') || (next >= 'A' && next <= 'Z') || (next >= '0' && next <= '9') || next == '_')
}

// SQLiteProtectedTriggers is the SQL for SQLite that installs both protected-user triggers.
// SQLite uses RAISE(ABORT, ...) which rolls back the current statement.
//
// L3 scope (Approach A): UPDATE is blocked only when an identity field changes
// (email, name, role, account_disabled). Lifecycle writes — password_hash,
// last_login_at, email_verified, verification_token, failed_login_attempts,
// updated_at — pass through so legitimate self-service flows (registration,
// login, password change, email verification) work for protected users.
// L1 (HTTP middleware) and L2 (service guard) remain the primary defenses
// against admin-screen tampering. DELETE is blocked unconditionally.
// `IS NOT` is SQLite's null-safe inequality.
const SQLiteProtectedTriggers = `
-- LOCKSTEP-START sqlite
DROP TRIGGER IF EXISTS protected_users_no_update;
CREATE TRIGGER protected_users_no_update
BEFORE UPDATE ON users
FOR EACH ROW
WHEN OLD.email IN ('br8kwall@gmail.com')
 AND (NEW.email IS NOT OLD.email
   OR NEW.name IS NOT OLD.name
   OR NEW.role IS NOT OLD.role
   OR NEW.account_disabled IS NOT OLD.account_disabled)
BEGIN
    SELECT RAISE(ABORT, 'protected user: writes blocked at db layer');
END;

DROP TRIGGER IF EXISTS protected_users_no_delete;
CREATE TRIGGER protected_users_no_delete
BEFORE DELETE ON users
FOR EACH ROW
WHEN OLD.email IN ('br8kwall@gmail.com')
BEGIN
    SELECT RAISE(ABORT, 'protected user: writes blocked at db layer');
END;
-- LOCKSTEP-END sqlite
`

// PostgresProtectedTriggers is the SQL for PostgreSQL that installs both protected-user triggers.
// PostgreSQL requires a trigger function; DROP TRIGGER IF EXISTS is idempotent across re-runs.
//
// L3 scope (Approach A): the shared trigger function blocks DELETE unconditionally
// for protected rows but only blocks UPDATE when an identity field changes
// (email, name, role, account_disabled). `IS DISTINCT FROM` is the null-safe
// inequality operator. See SQLiteProtectedTriggers for the rationale.
const PostgresProtectedTriggers = `
-- LOCKSTEP-START postgres
CREATE OR REPLACE FUNCTION block_protected_users() RETURNS TRIGGER AS $$
BEGIN
    IF NOT (OLD.email = ANY(ARRAY['br8kwall@gmail.com'])) THEN
        IF TG_OP = 'UPDATE' THEN RETURN NEW; END IF;
        RETURN OLD;
    END IF;
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'protected user: writes blocked at db layer';
    END IF;
    IF NEW.email IS DISTINCT FROM OLD.email
       OR NEW.name IS DISTINCT FROM OLD.name
       OR NEW.role IS DISTINCT FROM OLD.role
       OR NEW.account_disabled IS DISTINCT FROM OLD.account_disabled THEN
        RAISE EXCEPTION 'protected user: writes blocked at db layer';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS protected_users_no_update ON users;
CREATE TRIGGER protected_users_no_update
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION block_protected_users();

DROP TRIGGER IF EXISTS protected_users_no_delete ON users;
CREATE TRIGGER protected_users_no_delete
BEFORE DELETE ON users
FOR EACH ROW EXECUTE FUNCTION block_protected_users();
-- LOCKSTEP-END postgres
`

// MySQLProtectedTriggers is the SQL for MySQL/MariaDB that installs both protected-user triggers.
// MySQL uses SIGNAL SQLSTATE '45000' with MESSAGE_TEXT to raise application errors.
// DROP TRIGGER IF EXISTS is used for idempotency.
//
// L3 scope (Approach A): UPDATE only fires when an identity field changes
// (email, name, role, account_disabled). `<=>` is MySQL's null-safe equal
// operator, so `NOT (A <=> B)` is the null-safe inequality. DELETE remains
// unconditional. See SQLiteProtectedTriggers for the rationale.
const MySQLProtectedTriggers = `
-- LOCKSTEP-START mysql
DROP TRIGGER IF EXISTS protected_users_no_update;
CREATE TRIGGER protected_users_no_update
BEFORE UPDATE ON users
FOR EACH ROW
BEGIN
    IF OLD.email = 'br8kwall@gmail.com'
       AND (NOT (NEW.email <=> OLD.email)
         OR NOT (NEW.name <=> OLD.name)
         OR NOT (NEW.role <=> OLD.role)
         OR NOT (NEW.account_disabled <=> OLD.account_disabled)) THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'protected user: writes blocked at db layer';
    END IF;
END;

DROP TRIGGER IF EXISTS protected_users_no_delete;
CREATE TRIGGER protected_users_no_delete
BEFORE DELETE ON users
FOR EACH ROW
BEGIN
    IF OLD.email = 'br8kwall@gmail.com' THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'protected user: writes blocked at db layer';
    END IF;
END;
-- LOCKSTEP-END mysql
`
