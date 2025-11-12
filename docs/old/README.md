# Outdated Documentation Archive

This directory contains documentation that is no longer accurate for the current implementation.

## Why These Documents Were Moved

As of 2025-11-12, it was determined that the project will **not be performing database migrations** for the foreseeable future. All new database instances will use the current schema defined in `internal/repository/database.go`.

These documents were moved here to prevent confusion and ensure that only current, accurate documentation is read during development.

## Archived Documents

### Migration Planning Documents (v0.3.0/v0.4.0)
- `MIGRATION_v0.4.0_PROGRESS.md` - Progress tracking for cancelled v0.4.0 migration
- `MIGRATION_v0.4.0_STATUS.md` - Status report for cancelled v0.4.0 migration
- `SCHEMA_MIGRATION_v0.4.0.md` - Detailed migration plan that won't be executed
- `SCHEMA_ANALYSIS.md` - Schema analysis for planned migrations

### Outdated Schema Documentation
- `DATABASE_SCHEMA_CURRENT.md` - Duplicate/outdated schema documentation (superseded by `DATABASE_SCHEMA.md`)

### Outdated Test Documentation
- `TEST_PLAN_v0.3.0.md` - Test plan for v0.3.0 (superseded by current testing practices)
- `TEST_RESULTS.md` - Old test results (superseded by current test suite)

## Current Documentation

For accurate, up-to-date documentation, refer to:

- **`../DATABASE_SCHEMA.md`** - Current database schema (as implemented in database.go)
- **`../REQUIREMENTS_VS_IMPLEMENTATION.md`** - Analysis of requirements vs current implementation
- **`../ARCHITECTURE.md`** - Current architecture patterns
- **`../TESTING.md`** - Current testing practices
- **`../CHANGELOG.md`** - Version history

## Historical Value

These documents are preserved for historical reference and may contain useful context about design decisions, but they should **not** be used as a guide for current development.

## Date Archived

2025-11-12

## Related Decision

See `../REQUIREMENTS_VS_IMPLEMENTATION.md` for the architectural analysis that led to this decision.
