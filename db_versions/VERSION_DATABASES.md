# Version Database Snapshots

This document describes the version-specific database copies maintained for migration testing across all supported database engines.

## Overview

For each version, we maintain database snapshots in three formats:

1. **SQLite** - Single file in `db_versions/` directory
2. **PostgreSQL** - Schema in `jcz` database on 192.168.1.143
3. **MariaDB** - Database on 192.168.1.234

These snapshots contain identical test data and allow testing migrations from any version to the current version.

## Version 0.14.0

**Release Date:** 2024-12-16
**Features:** Subscription billing system with dual-level (user + organization) subscriptions

### SQLite
**Location:** `db_versions/actalog_0.14.0.db`
**Size:** 564 KB

```bash
# Access
sqlite3 db_versions/actalog_0.14.0.db

# Test migration to current version
cp db_versions/actalog_0.14.0.db test.db
DB_DRIVER=sqlite3 DB_NAME=test.db ./bin/actalog
```

**Contents:**
- 4 users (all with permanent free subscriptions)
- 222 movements
- 101 WODs
- 9 workout templates
- 407 user workout logs
- 4 user subscriptions (all permanent free)
- 0 organizations

### PostgreSQL
**Location:** `192.168.1.143:5432/jcz` schema: `actalog_0_14_0`

```bash
# Access
PGPASSWORD='yub.miha' psql -h 192.168.1.143 -U jcz -d jcz

# Switch to version schema
SET search_path TO actalog_0_14_0;

# Test migration to current version
DB_DRIVER=postgres DB_HOST=192.168.1.143 DB_PORT=5432 \
DB_NAME=jcz DB_USER=jcz DB_PASSWORD=yub.miha \
DB_SCHEMA=actalog_0_14_0_test ./bin/actalog
```

**Contents:**
- 4 users (user1@actalog.test, user2@actalog.test, admin@actalog.test, user3@actalog.test)
- 4 user subscriptions (all permanent free, active)
- 10 movements
- 10 WODs
- 1 workout template
- 1 organization (Test Gym)

**Test Users:**
- Email: `user1@actalog.test`, `user2@actalog.test`, `user3@actalog.test`, `admin@actalog.test`
- Password Hash: `$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5yvFy/F4GhRju`
- All users: Permanent free subscription, active status

### MariaDB
**Location:** `192.168.1.234:3306` database: `actalog_0_14_0`

```bash
# Access
mysql -h 192.168.1.234 -u jcz -p'yub.miha' actalog_0_14_0

# Test migration to current version
DB_DRIVER=mysql DB_HOST=192.168.1.234 DB_PORT=3306 \
DB_NAME=actalog_0_14_0_test DB_USER=jcz DB_PASSWORD=yub.miha \
./bin/actalog
```

**Contents:**
- 4 users (same as PostgreSQL)
- 4 user subscriptions (all permanent free, active)
- 32 movements
- 10 WODs
- 1 organization (Test Gym)
- 4 user-organization links (all users assigned to Test Gym)

## Test Data Specifications

### Users
All version databases contain the same 4 test users:

| Email | Name | Role | Password (hashed) |
|-------|------|------|-------------------|
| user1@actalog.test | Test User 1 | user | bcrypt hash |
| user2@actalog.test | Test User 2 | user | bcrypt hash |
| admin@actalog.test | Admin User | admin | bcrypt hash |
| user3@actalog.test | Test User 3 | user | bcrypt hash |

**Note:** Password hash is `$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5yvFy/F4GhRju` (hashes "password123")

### Subscriptions (v0.14.0+)
All users have permanent free subscriptions:
- `subscription_type`: 'free'
- `status`: 'active'
- `is_permanent_free`: TRUE
- `end_date`: NULL
- `next_billing_date`: NULL

### Sample Movements
Core movements included in test data:
- Back Squat (strength)
- Deadlift (strength)
- Pull-ups (gymnastics)
- Box Jump (metcon)
- Double Unders (metcon)
- Thruster (strength)
- Burpees (metcon)
- Wall Ball (metcon)
- Kettlebell Swing (metcon)
- Row (cardio)

### Sample WODs
Classic CrossFit benchmark WODs:
- Fran (21-15-9 Thrusters/Pull-ups)
- Cindy (20 min AMRAP)
- Helen (3 rounds)
- Murph (with 1 mile runs)
- Grace (30 Clean & Jerks)

### Organizations
- Test Gym (Sample CrossFit gym for testing)
- All test users are members

## Creating New Version Snapshots

When releasing a new version (e.g., 0.15.0):

### 1. SQLite Snapshot
```bash
# Using the script
./scripts/create-db-snapshot.sh

# Manual
cp actalog.db db_versions/actalog_X.Y.Z.db
git add db_versions/actalog_X.Y.Z.db
```

### 2. PostgreSQL Schema
```bash
# Create new schema
PGPASSWORD='yub.miha' psql -h 192.168.1.143 -U jcz -d jcz << EOF
CREATE SCHEMA actalog_X_Y_Z;
EOF

# Run migrations
DB_DRIVER=postgres DB_HOST=192.168.1.143 DB_PORT=5432 \
DB_NAME=jcz DB_USER=jcz DB_PASSWORD=yub.miha \
DB_SCHEMA=actalog_X_Y_Z ./bin/actalog

# Populate with test data (see scripts below)
```

### 3. MariaDB Database
```bash
# Create new database
mysql -h 192.168.1.234 -u jcz -p'yub.miha' << EOF
CREATE DATABASE actalog_X_Y_Z;
EOF

# Run migrations
DB_DRIVER=mysql DB_HOST=192.168.1.234 DB_PORT=3306 \
DB_NAME=actalog_X_Y_Z DB_USER=jcz DB_PASSWORD=yub.miha \
./bin/actalog

# Populate with test data
```

### 4. Update Documentation
- Add entry to this file
- Update `db_versions/README.md` table
- Document new features/schema changes

## Testing Migration Paths

### From 0.14.0 to Current

#### SQLite
```bash
cp db_versions/actalog_0.14.0.db test_migration.db
DB_DRIVER=sqlite3 DB_NAME=test_migration.db ./bin/actalog
# Verify new migrations applied
sqlite3 test_migration.db "SELECT version FROM schema_migrations ORDER BY applied_at DESC LIMIT 1;"
```

#### PostgreSQL
```bash
# Create test schema from 0.14.0 snapshot
PGPASSWORD='yub.miha' psql -h 192.168.1.143 -U jcz -d jcz << EOF
CREATE SCHEMA actalog_migration_test;
-- Copy structure from actalog_0_14_0
-- (use pg_dump/restore or manual table recreation)
EOF

# Run current version
DB_DRIVER=postgres DB_HOST=192.168.1.143 DB_PORT=5432 \
DB_NAME=jcz DB_USER=jcz DB_PASSWORD=yub.miha \
DB_SCHEMA=actalog_migration_test ./bin/actalog
```

#### MariaDB
```bash
# Create test database from 0.14.0 snapshot
mysql -h 192.168.1.234 -u jcz -p'yub.miha' << EOF
CREATE DATABASE actalog_migration_test;
-- Use mysqldump to copy from actalog_0_14_0
EOF

# Run current version
DB_DRIVER=mysql DB_HOST=192.168.1.234 DB_PORT=3306 \
DB_NAME=actalog_migration_test DB_USER=jcz DB_PASSWORD=yub.miha \
./bin/actalog
```

## Maintenance

### Cleanup Old Versions
Keep the last 3-5 major version snapshots. Archive older versions:

```bash
# SQLite - move to archive
mkdir -p db_versions/archive
mv db_versions/actalog_0.11.0.db db_versions/archive/

# PostgreSQL - drop old schemas
PGPASSWORD='yub.miha' psql -h 192.168.1.143 -U jcz -d jcz << EOF
DROP SCHEMA IF EXISTS actalog_0_11_0 CASCADE;
EOF

# MariaDB - drop old databases
mysql -h 192.168.1.234 -u jcz -p'yub.miha' << EOF
DROP DATABASE IF EXISTS actalog_0_11_0;
EOF
```

### Verify Snapshot Integrity
```bash
# Check all snapshots have correct schema version
for db in db_versions/actalog_*.db; do
    version=$(basename "$db" .db | sed 's/actalog_//')
    schema_version=$(sqlite3 "$db" "SELECT version FROM schema_migrations ORDER BY applied_at DESC LIMIT 1;")
    echo "$db: schema_version=$schema_version (expected=$version)"
done
```

## Connection Information

### PostgreSQL
- **Host:** 192.168.1.143
- **Port:** 5432
- **Database:** jcz
- **User:** jcz
- **Password:** yub.miha
- **Schemas:** actalog, actalog_0_14_0, etc.

### MariaDB
- **Host:** 192.168.1.234
- **Port:** 3306
- **User:** jcz
- **Password:** yub.miha
- **Databases:** actalog, actalog_0_14_0, etc.

## Security Notes

⚠️ **Important:**
- These are TEST databases with FAKE DATA only
- Never commit real production data to version control
- Test credentials are for development environment only
- Production databases use different credentials and network isolation
