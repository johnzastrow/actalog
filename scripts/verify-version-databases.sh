#!/bin/bash
# Verify all version database snapshots

set -e

echo "=== ActaLog Version Database Verification ==="
echo ""

# Get current version
MAJOR=$(grep 'Major =' pkg/version/version.go | grep -oP '\d+')
MINOR=$(grep 'Minor =' pkg/version/version.go | grep -oP '\d+')
PATCH=$(grep 'Patch =' pkg/version/version.go | grep -oP '\d+')
VERSION="${MAJOR}.${MINOR}.${PATCH}"

echo "Current Version: $VERSION"
echo ""

# Check SQLite snapshots
echo "SQLite Snapshots:"
if [ -d "db_versions" ]; then
    for db in db_versions/actalog_*.db; do
        if [ -f "$db" ]; then
            filename=$(basename "$db")
            version=$(echo "$filename" | sed 's/actalog_//' | sed 's/\.db//')
            schema_version=$(sqlite3 "$db" "SELECT version FROM schema_migrations ORDER BY applied_at DESC LIMIT 1;" 2>/dev/null || echo "error")
            user_count=$(sqlite3 "$db" "SELECT COUNT(*) FROM users;" 2>/dev/null || echo "0")
            sub_count=$(sqlite3 "$db" "SELECT COUNT(*) FROM user_subscriptions;" 2>/dev/null || echo "0")

            if [ "$schema_version" = "$version" ]; then
                echo "  ✓ $filename - v$schema_version ($user_count users, $sub_count subscriptions)"
            else
                echo "  ✗ $filename - schema v$schema_version (expected v$version)"
            fi
        fi
    done
else
    echo "  ⚠ db_versions directory not found"
fi
echo ""

# Check PostgreSQL schemas
echo "PostgreSQL Schemas (192.168.1.143):"
if command -v psql &> /dev/null; then
    PGPASSWORD='yub.miha' psql -h 192.168.1.143 -U jcz -d jcz -t -A -c "
        SELECT schema_name
        FROM information_schema.schemata
        WHERE schema_name LIKE 'actalog_%'
        ORDER BY schema_name;
    " 2>/dev/null | while read -r schema; do
        if [ -n "$schema" ]; then
            version=$(echo "$schema" | sed 's/actalog_//' | tr '_' '.')
            schema_version=$(PGPASSWORD='yub.miha' psql -h 192.168.1.143 -U jcz -d jcz -t -A -c "
                SELECT version FROM $schema.schema_migrations ORDER BY applied_at DESC LIMIT 1;
            " 2>/dev/null | tr -d ' ')
            user_count=$(PGPASSWORD='yub.miha' psql -h 192.168.1.143 -U jcz -d jcz -t -A -c "
                SELECT COUNT(*) FROM $schema.users;
            " 2>/dev/null | tr -d ' ')

            if [ "$schema_version" = "$version" ]; then
                echo "  ✓ $schema - v$schema_version ($user_count users)"
            else
                echo "  ✗ $schema - schema v$schema_version (expected v$version)"
            fi
        fi
    done
else
    echo "  ⚠ psql not available, skipping PostgreSQL verification"
fi
echo ""

# Check MariaDB databases
echo "MariaDB Databases (192.168.1.234):"
if command -v mysql &> /dev/null; then
    mysql -h 192.168.1.234 -u jcz -p'yub.miha' -N -e "
        SHOW DATABASES LIKE 'actalog_%';
    " 2>/dev/null | while read -r database; do
        if [ -n "$database" ]; then
            version=$(echo "$database" | sed 's/actalog_//' | tr '_' '.')
            schema_version=$(mysql -h 192.168.1.234 -u jcz -p'yub.miha' -N "$database" -e "
                SELECT version FROM schema_migrations ORDER BY applied_at DESC LIMIT 1;
            " 2>/dev/null)
            user_count=$(mysql -h 192.168.1.234 -u jcz -p'yub.miha' -N "$database" -e "
                SELECT COUNT(*) FROM users;
            " 2>/dev/null)

            if [ "$schema_version" = "$version" ]; then
                echo "  ✓ $database - v$schema_version ($user_count users)"
            else
                echo "  ✗ $database - schema v$schema_version (expected v$version)"
            fi
        fi
    done
else
    echo "  ⚠ mysql not available, skipping MariaDB verification"
fi
echo ""

echo "✓ Verification complete"
echo ""
echo "For detailed access information, see: db_versions/VERSION_DATABASES.md"
