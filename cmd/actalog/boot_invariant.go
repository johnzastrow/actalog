package main

import (
	"database/sql"

	"github.com/johnzastrow/actalog/internal/protectedusers"
)

// VerifyProtectedUserInvariant delegates to the importable package so that
// cmd/actalog/main.go and the integration test suite can both call it.
func VerifyProtectedUserInvariant(db *sql.DB, driver string) (*protectedusers.InvariantReport, error) {
	return protectedusers.VerifyProtectedUserInvariant(db, driver)
}
