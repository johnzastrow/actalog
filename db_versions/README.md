# Database Version Snapshots

This directory contains SQLite database snapshots for each version of ActaLog. These snapshots are used for:

1. **Migration Testing** - Test forward and backward migrations between versions
2. **Regression Testing** - Verify data integrity across version upgrades
3. **Development** - Quick access to database states at different schema versions
4. **Documentation** - Reference schema evolution over time

## Version Management

### Creating a New Snapshot

When releasing a new version, create a snapshot of the current production database:

```bash
# After running migrations and verifying the new version
cp actalog.db db_versions/actalog_X.Y.Z.db

# Verify the snapshot
sqlite3 db_versions/actalog_X.Y.Z.db "SELECT version FROM schema_migrations ORDER BY applied_at DESC LIMIT 1;"
```

### Testing Migrations

To test a migration from one version to another:

```bash
# Copy an old version snapshot to test with
cp db_versions/actalog_0.13.0.db test_migration.db

# Run the application against the test database
DB_DRIVER=sqlite3 DB_NAME=test_migration.db ./bin/actalog

# The migration will run automatically
# Verify the migration was successful
sqlite3 test_migration.db "SELECT version FROM schema_migrations ORDER BY applied_at DESC LIMIT 1;"
```

### Sample Data

For testing with realistic data, snapshots should include:
- Multiple users (regular and admin)
- Various workouts, movements, WODs
- User workout logs with performance data
- Organizations and memberships
- Subscriptions (for v0.14.0+)

Sample data can be imported from the MariaDB `acta` database using the export/import functionality.

## Available Snapshots

| Version | File | Description | Date | Records |
|---------|------|-------------|------|---------|
| 0.14.0 | `actalog_0.14.0.db` | Subscription billing system | 2024-12-16 | 4 users, 222 movements, 101 WODs, 9 workouts, 407 user_workouts, 4 subscriptions |

## Quick Reference

```bash
# Create snapshot for current version
./scripts/create-db-snapshot.sh

# Test migration from 0.13.0 to 0.14.0
cp db_versions/actalog_0.13.0.db test.db
DB_DRIVER=sqlite3 DB_NAME=test.db ./bin/actalog
# Should show: Migration 0.14.0 applied

# Verify subscription seeding worked
sqlite3 test.db "SELECT COUNT(*) FROM user_subscriptions;"
# Should match user count
```

## Schema Evolution

### Version 0.14.0
- Added `user_subscriptions` table
- Added `organization_subscriptions` table
- Seeded all existing users with permanent free subscriptions
- Added subscription audit event types

### Version 0.13.0
- Added multi-organization support
- Created `organizations` and `user_organizations` tables

### Earlier Versions
See `internal/repository/migrations.go` for complete migration history.

## Notes

- Snapshots should be committed to the repository for team access
- Keep at minimum the last 3 major version snapshots
- Snapshots are SQLite format for portability
- Do NOT include sensitive production data in committed snapshots
- For production testing, create sanitized copies with fake user data
