#!/bin/sh
# Entrypoint script for ActaLog Docker container
# Starts the application and optionally imports seed data on first run

set -e

# Start the main application in the background
/app/actalog &
APP_PID=$!

# Wait for app to be ready (max 30 seconds)
echo "Waiting for ActaLog to start..."
for i in $(seq 1 30); do
    if wget -q -O- http://localhost:8080/health > /dev/null 2>&1; then
        echo "ActaLog is ready!"
        break
    fi
    sleep 1
done

# Run seed import script (only if ADMIN_EMAIL and ADMIN_PASSWORD are set)
if [ -n "$ADMIN_EMAIL" ] && [ -n "$ADMIN_PASSWORD" ]; then
    echo "Admin credentials provided. Running seed import script..."
    /app/scripts/init-seeds.sh || echo "Warning: Seed import failed or was skipped"
else
    echo "No admin credentials provided. Skipping automatic seed import."
    echo "You can manually import seeds via the web UI or API after creating an admin account."
fi

# Bring the application to the foreground
wait $APP_PID
