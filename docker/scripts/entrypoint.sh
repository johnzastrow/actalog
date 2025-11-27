#!/bin/sh
# Entrypoint script for ActaLog Docker container
# Starts the application and optionally imports seed data on first run

set -e

# Trap signals and forward them to the application
trap 'kill -TERM $APP_PID 2>/dev/null' TERM INT

# Start the main application in the background
echo "Starting ActaLog..."
/app/actalog &
APP_PID=$!

# Wait for app to be ready (max 30 seconds)
echo "Waiting for ActaLog to start (PID: $APP_PID)..."
READY=0
for i in $(seq 1 30); do
    # Check if the process is still running
    if ! kill -0 $APP_PID 2>/dev/null; then
        echo "ERROR: ActaLog process died during startup!"
        echo "Check the logs above for error messages."
        exit 1
    fi

    # Check if the health endpoint responds
    if wget -q -O- http://localhost:8080/health > /dev/null 2>&1; then
        echo "ActaLog is ready!"
        READY=1
        break
    fi
    sleep 1
done

if [ $READY -eq 0 ]; then
    echo "ERROR: ActaLog failed to become healthy within 30 seconds"
    echo "The process is running but not responding. Check logs for details."
    kill $APP_PID 2>/dev/null || true
    exit 1
fi

# Run seed import script (only if ADMIN_EMAIL and ADMIN_PASSWORD are set)
if [ -n "$ADMIN_EMAIL" ] && [ -n "$ADMIN_PASSWORD" ]; then
    echo "Admin credentials provided. Running seed import script..."
    /app/scripts/init-seeds.sh || echo "Warning: Seed import failed or was skipped"
else
    echo "No admin credentials provided. Skipping automatic seed import."
    echo "You can manually import seeds via the web UI or API after creating an admin account."
fi

# Bring the application to the foreground
echo "ActaLog is now running. Monitoring process..."
wait $APP_PID
EXIT_CODE=$?
echo "ActaLog exited with code: $EXIT_CODE"
exit $EXIT_CODE
