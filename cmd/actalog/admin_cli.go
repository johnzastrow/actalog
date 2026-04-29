package main

import (
	"database/sql"
	"io"

	"github.com/johnzastrow/actalog/internal/protectedusers"
)

// AdminVerifyProtectedUsers delegates to the importable package.
// Prints a per-check report to out; returns shell-style exit code (0 = pass, 1 = fail).
func AdminVerifyProtectedUsers(db *sql.DB, driver string, verbose bool, out io.Writer) int {
	return protectedusers.AdminVerifyProtectedUsers(db, driver, verbose, out)
}

// AdminReapplyProtectedMigrations delegates to the importable package.
// Drops + recreates the L3 triggers, then re-verifies the invariant.
func AdminReapplyProtectedMigrations(db *sql.DB, driver string, confirm bool, out io.Writer) error {
	return protectedusers.AdminReapplyProtectedMigrations(db, driver, confirm, out)
}
