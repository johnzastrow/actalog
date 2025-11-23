#!/bin/sh
# Initialize seed data for ActaLog Docker deployment
# This script imports movements and WODs from CSV files on first run

set -e

SEED_MARKER="/app/data/.seeds_imported"
API_URL="${API_URL:-http://localhost:8080}"

echo "Checking if seed data needs to be imported..."

# Check if already imported
if [ -f "$SEED_MARKER" ]; then
    echo "Seed data already imported. Skipping."
    exit 0
fi

# Wait for API to be ready
echo "Waiting for API to be ready..."
for i in $(seq 1 30); do
    if wget -q -O- "${API_URL}/health" > /dev/null 2>&1; then
        echo "API is ready!"
        break
    fi
    echo "Waiting... ($i/30)"
    sleep 2
done

# Check if seed files exist
if [ ! -f "/app/seeds/movements.csv" ] || [ ! -f "/app/seeds/wods.csv" ]; then
    echo "WARNING: Seed CSV files not found. Skipping seed import."
    echo "The application will start with minimal hardcoded movements/WODs."
    exit 0
fi

echo "Seed files found. Checking for admin user..."

# Wait for first user to be created (admin)
# This script should run AFTER the first user registers
# For automated deployment, you can set ADMIN_EMAIL and ADMIN_PASSWORD

if [ -z "$ADMIN_EMAIL" ] || [ -z "$ADMIN_PASSWORD" ]; then
    echo "ADMIN_EMAIL and ADMIN_PASSWORD not set."
    echo "Please manually import seeds via the web UI or API after creating an admin account."
    exit 0
fi

echo "Attempting to login as admin..."

# Login and get token
TOKEN=$(wget -q -O- --post-data="{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\"}" \
    --header="Content-Type: application/json" \
    "${API_URL}/api/auth/login" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "ERROR: Failed to login. Cannot import seeds automatically."
    echo "Please import seeds manually via the web UI."
    exit 0
fi

echo "Login successful. Importing seed data..."

# Import movements
echo "Importing movements..."
wget -q -O- --header="Authorization: Bearer $TOKEN" \
    --post-file="/app/seeds/movements.csv" \
    "${API_URL}/api/import/movements/confirm?skip_duplicates=true" > /dev/null

# Import WODs
echo "Importing WODs..."
wget -q -O- --header="Authorization: Bearer $TOKEN" \
    --post-file="/app/seeds/wods.csv" \
    "${API_URL}/api/import/wods/confirm?skip_duplicates=true" > /dev/null

# Mark as imported
touch "$SEED_MARKER"

echo "Seed data imported successfully!"
echo "  - Movements: $(wc -l < /app/seeds/movements.csv) rows"
echo "  - WODs: $(wc -l < /app/seeds/wods.csv) rows"
