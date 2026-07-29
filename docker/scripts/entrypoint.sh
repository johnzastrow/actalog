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

    # Check if the health endpoint responds.
    # Use 127.0.0.1, not "localhost": on hosts where localhost resolves to ::1
    # first, wget connects to [::1] while the server binds IPv4 only (SERVER_HOST
    # defaults to 0.0.0.0 and is commonly set to 127.0.0.1), so the probe can
    # never succeed. Honour SERVER_PORT so non-default ports (e.g. a beta
    # instance on 8081) probe themselves rather than another instance on 8080.
    if wget -q -O- "http://127.0.0.1:${SERVER_PORT:-8080}/health" > /dev/null 2>&1; then
        echo "ActaLog is ready!"
        READY=1
        break
    fi
    sleep 1
done

if [ $READY -eq 0 ]; then
    # Do NOT kill the app here. This readiness poll exists only to gate the
    # optional seed import below; it is not a liveness supervisor. A probe that
    # kills a healthy process turns any probe defect into a permanent crash-loop
    # (with restart: unless-stopped this produced 25,946 restarts and a 9-day
    # prod outage). The process-died check inside the loop above still exits 1
    # for a genuinely dead app, and Docker's HEALTHCHECK reports health state.
    echo "WARNING: ActaLog did not answer the readiness probe within 30 seconds"
    echo "Continuing anyway - the process is running. Skipping seed import."
    echo "Check Docker's HEALTHCHECK status and the logs above if the app is unreachable."
fi

# Run seed import script (only if the app answered the readiness probe and
# ADMIN_EMAIL and ADMIN_PASSWORD are set) - seeding drives the HTTP API, so it
# cannot work while the app is unreachable.
if [ $READY -eq 0 ]; then
    echo "Skipping seed import: app did not answer the readiness probe."
elif [ -n "$ADMIN_EMAIL" ] && [ -n "$ADMIN_PASSWORD" ]; then
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
