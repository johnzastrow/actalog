#!/usr/bin/env bash
# Last-resort recovery: restore protected-user triggers via raw SQL.
# Use ONLY when the actalog binary itself is unrunnable. Prefer
# `actalog admin reapply-protected-migrations --confirm` first.
set -euo pipefail

usage() {
  cat <<EOF
Usage: $0 --driver=<sqlite3|postgres|mysql> --conn=<connection-string>

  --driver  one of: sqlite3 | postgres | mysql
  --conn    DB connection (file path for sqlite3, URL for postgres/mysql)

Examples:
  $0 --driver=sqlite3 --conn=./data/actalog.db
  $0 --driver=postgres --conn='postgres://user:pass@host:5432/db?sslmode=require'
  $0 --driver=mysql --conn='user:pass@tcp(host:3306)/db'
EOF
  exit 1
}

DRIVER=""
CONN=""
for arg in "$@"; do
  case "$arg" in
    --driver=*) DRIVER="${arg#*=}";;
    --conn=*) CONN="${arg#*=}";;
    -h|--help) usage;;
    *) echo "unknown arg: $arg" >&2; usage;;
  esac
done

[[ -z "$DRIVER" || -z "$CONN" ]] && usage

case "$DRIVER" in
  sqlite3)  DIALECT=sqlite ;;
  postgres) DIALECT=postgres ;;
  mysql)    DIALECT=mysql ;;
  *) echo "unsupported driver: $DRIVER" >&2; exit 2 ;;
esac

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SQL_FILE="$SCRIPT_DIR/sql/$DIALECT/protected_users.sql"
[[ -f "$SQL_FILE" ]] || { echo "missing SQL file: $SQL_FILE" >&2; exit 2; }

case "$DRIVER" in
  sqlite3)  sqlite3 "$CONN" < "$SQL_FILE" ;;
  postgres) psql "$CONN" -f "$SQL_FILE" ;;
  mysql)    mysql -h "$(echo "$CONN" | sed 's/.*@tcp(\([^:]*\):.*/\1/')" --defaults-extra-file=<(printf '[client]\nuser=%s\npassword=%s\n' \
              "$(echo "$CONN" | cut -d: -f1)" "$(echo "$CONN" | cut -d: -f2 | cut -d@ -f1)") \
              "$(echo "$CONN" | sed 's/.*\///')" < "$SQL_FILE" ;;
  *) echo "unsupported driver: $DRIVER" >&2; exit 2 ;;
esac

echo "Recovery SQL applied via $DRIVER. Restart actalog and verify with:"
echo "    ./bin/actalog admin verify-protected-users --verbose"
