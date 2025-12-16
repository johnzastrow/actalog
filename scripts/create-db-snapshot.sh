#!/bin/bash
# Create a versioned database snapshot for migration testing

set -e

# Get the current version from version.go
MAJOR=$(grep 'Major =' pkg/version/version.go | grep -oP '\d+')
MINOR=$(grep 'Minor =' pkg/version/version.go | grep -oP '\d+')
PATCH=$(grep 'Patch =' pkg/version/version.go | grep -oP '\d+')
VERSION="${MAJOR}.${MINOR}.${PATCH}"

# Check if version was extracted correctly
if [ -z "$VERSION" ]; then
    echo "Error: Could not extract version from pkg/version/version.go"
    exit 1
fi

# Remove "beta" suffix if present
VERSION=$(echo "$VERSION" | sed 's/-beta//')

# Create db_versions directory if it doesn't exist
mkdir -p db_versions

# Determine source database
if [ -f "actalog.db" ]; then
    SOURCE_DB="actalog.db"
elif [ -f "actalog_test.db" ]; then
    SOURCE_DB="actalog_test.db"
else
    echo "Error: No database file found (actalog.db or actalog_test.db)"
    exit 1
fi

SNAPSHOT_FILE="db_versions/actalog_${VERSION}.db"

# Check if snapshot already exists
if [ -f "$SNAPSHOT_FILE" ]; then
    echo "Warning: Snapshot $SNAPSHOT_FILE already exists"
    read -p "Overwrite? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted"
        exit 1
    fi
fi

# Create snapshot
echo "Creating snapshot: $SNAPSHOT_FILE"
cp "$SOURCE_DB" "$SNAPSHOT_FILE"

# Verify snapshot
SNAPSHOT_VERSION=$(sqlite3 "$SNAPSHOT_FILE" "SELECT version FROM schema_migrations ORDER BY applied_at DESC LIMIT 1;" 2>/dev/null || echo "unknown")

echo ""
echo "✓ Snapshot created successfully"
echo "  File: $SNAPSHOT_FILE"
echo "  Size: $(du -h "$SNAPSHOT_FILE" | cut -f1)"
echo "  Schema version: $SNAPSHOT_VERSION"
echo ""

# Count records
echo "Database contents:"
sqlite3 "$SNAPSHOT_FILE" << 'EOF'
.mode column
.headers on
SELECT
    'users' AS table_name, COUNT(*) AS records FROM users
UNION ALL
SELECT 'movements', COUNT(*) FROM movements
UNION ALL
SELECT 'wods', COUNT(*) FROM wods
UNION ALL
SELECT 'workouts', COUNT(*) FROM workouts
UNION ALL
SELECT 'user_workouts', COUNT(*) FROM user_workouts
UNION ALL
SELECT 'user_subscriptions', COUNT(*) FROM user_subscriptions
UNION ALL
SELECT 'organizations', COUNT(*) FROM organizations WHERE EXISTS (SELECT 1 FROM organizations)
ORDER BY records DESC;
EOF

echo ""
echo "Next steps:"
echo "  1. Test the migration: DB_DRIVER=sqlite3 DB_NAME=$SNAPSHOT_FILE ./bin/actalog"
echo "  2. Commit the snapshot: git add $SNAPSHOT_FILE"
echo "  3. Update db_versions/README.md with version details"
