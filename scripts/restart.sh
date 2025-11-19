#!/bin/bash

echo "=== ActaLog Restart Script ==="
echo ""

# Get the absolute path of the project directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PROJECT_NAME="actionlog"

echo "Project directory: $PROJECT_ROOT"
echo ""

# Kill backend processes
echo "Stopping backend..."

# Kill processes with "actalog" binary name
pkill -9 -f "bin/actalog" 2>/dev/null || true

# Kill go run processes in this directory
for pid in $(ps aux | grep "[g]o run" | grep "$PROJECT_ROOT" | awk '{print $2}'); do
    echo "  Killing go run process: $pid"
    kill -9 $pid 2>/dev/null || true
done

# Kill any actalog binary processes
for pid in $(ps aux | grep "[a]ctalog" | grep -v "grep" | awk '{print $2}'); do
    # Check if it's running from our directory
    if [ -d "/proc/$pid" ]; then
        cwd=$(readlink -f /proc/$pid/cwd 2>/dev/null || echo "")
        if [[ "$cwd" == *"$PROJECT_NAME"* ]] || [[ "$cwd" == *"$PROJECT_ROOT"* ]]; then
            echo "  Killing actalog process: $pid (from $cwd)"
            kill -9 $pid 2>/dev/null || true
        fi
    fi
done

# Also check for processes listening on any port from our binary
for pid in $(lsof -ti -c actalog 2>/dev/null); do
    echo "  Killing actalog listener: $pid"
    kill -9 $pid 2>/dev/null || true
done

echo "✓ Backend processes stopped"

# Kill frontend processes
echo "Stopping frontend..."

# Kill npm/node processes from the web directory
for pid in $(ps aux | grep "[n]pm run dev" | awk '{print $2}'); do
    if [ -d "/proc/$pid" ]; then
        cwd=$(readlink -f /proc/$pid/cwd 2>/dev/null || echo "")
        if [[ "$cwd" == *"$PROJECT_NAME"* ]] || [[ "$cwd" == *"/web"* ]] || [[ "$cwd" == *"$PROJECT_ROOT/web"* ]]; then
            echo "  Killing npm dev process: $pid (from $cwd)"
            kill -9 $pid 2>/dev/null || true
        fi
    fi
done

# Kill vite processes from the web directory
for pid in $(ps aux | grep "[v]ite" | grep -v "grep" | awk '{print $2}'); do
    if [ -d "/proc/$pid" ]; then
        cwd=$(readlink -f /proc/$pid/cwd 2>/dev/null || echo "")
        if [[ "$cwd" == *"$PROJECT_NAME"* ]] || [[ "$cwd" == *"/web"* ]] || [[ "$cwd" == *"$PROJECT_ROOT/web"* ]]; then
            echo "  Killing vite process: $pid (from $cwd)"
            kill -9 $pid 2>/dev/null || true
        fi
    fi
done

# Kill node processes from the web directory
for pid in $(ps aux | grep "[n]ode" | grep "vite" | awk '{print $2}'); do
    if [ -d "/proc/$pid" ]; then
        cwd=$(readlink -f /proc/$pid/cwd 2>/dev/null || echo "")
        if [[ "$cwd" == *"$PROJECT_NAME/web"* ]] || [[ "$cwd" == *"$PROJECT_ROOT/web"* ]]; then
            echo "  Killing node/vite process: $pid (from $cwd)"
            kill -9 $pid 2>/dev/null || true
        fi
    fi
done

echo "✓ Frontend processes stopped"

# Give processes a moment to fully terminate
sleep 1

echo ""
echo "Building and starting backend..."
# Ensure Go is installed before attempting to build
if ! command -v go >/dev/null 2>&1; then
    cat <<MSG
❌ 'go' not found in PATH. The backend build requires Go to be installed.

Install Go and ensure the 'go' binary is on your PATH. Common install commands:

  # Debian/Ubuntu (may provide older Go package):
  sudo apt update && sudo apt install -y golang

  # Fedora/CentOS/RHEL (dnf):
  sudo dnf install golang

  # Arch Linux:
  sudo pacman -S go

  # Or download latest from https://go.dev/dl and follow the Linux tarball install instructions.

After installing, reopen your shell or ensure the 'go' binary is available, then re-run this script.
MSG
    exit 1
fi

# Build backend (run Make in project root)
make -C "$PROJECT_ROOT" build
if [ $? -ne 0 ]; then
    echo "❌ Backend build failed!"
    exit 1
fi

# Start backend in background (make -C so we don't change cwd)
make -C "$PROJECT_ROOT" run > "$PROJECT_ROOT/backend.log" 2>&1 &
BACKEND_PID=$!
echo "✓ Backend started (PID: $BACKEND_PID, logs: $PROJECT_ROOT/backend.log)"


# Wait a moment for backend to start
sleep 2

# Check if backend is running
if ! ps -p $BACKEND_PID > /dev/null; then
    echo "❌ Backend failed to start! Check backend.log for errors"
    tail -20 backend.log
    exit 1
fi

# Detect backend port from the running process
BACKEND_PORT=$(lsof -Pan -p $BACKEND_PID -i 2>/dev/null | grep LISTEN | awk '{print $9}' | cut -d: -f2 | head -1)
if [ -z "$BACKEND_PORT" ]; then
    BACKEND_PORT="8080"
fi

echo ""
echo "Starting frontend..."
# Recommend running the full setup script on first use
if [ ! -f "$PROJECT_ROOT/bin/actalog" ]; then
    echo "\nNote: If this is the first time running ActaLog on this machine,"
    echo "consider running the full setup script to install tools and dependencies:"
    echo "  ./scripts/build.sh"
fi

# Verify Node/NPM are available before attempting to start the frontend
if ! command -v node >/dev/null 2>&1; then
    echo "❌ 'node' not found in PATH. Install Node.js (16+) and retry. See: https://nodejs.org/"
    exit 1
fi
if ! command -v npm >/dev/null 2>&1; then
    echo "❌ 'npm' not found in PATH. Install npm (bundled with Node.js) and retry."
    exit 1
fi

# Install frontend dependencies if needed (uses npm --prefix so we don't change cwd)
if [ ! -d "$PROJECT_ROOT/web/node_modules" ]; then
    echo "Installing frontend dependencies..."
    npm --prefix "$PROJECT_ROOT/web" install
fi

# Start frontend in background using npm --prefix to avoid changing cwd
npm --prefix "$PROJECT_ROOT/web" run dev > "$PROJECT_ROOT/frontend.log" 2>&1 &
FRONTEND_PID=$!
echo "✓ Frontend started (PID: $FRONTEND_PID, logs: $PROJECT_ROOT/frontend.log)"

# Optional: warn if mkcert might be useful for local HTTPS testing
if ! command -v mkcert >/dev/null 2>&1; then
    echo "\nNote: 'mkcert' not found in PATH. If you plan to test HTTPS locally, install mkcert: https://github.com/FiloSottile/mkcert"
fi

# Wait for frontend to start and detect its port
sleep 3
FRONTEND_PORT=$(lsof -Pan -p $FRONTEND_PID -i 2>/dev/null | grep LISTEN | awk '{print $9}' | cut -d: -f2 | head -1)
if [ -z "$FRONTEND_PORT" ]; then
    # Try to find any vite process port
    FRONTEND_PORT=$(lsof -Pan -c node -i 2>/dev/null | grep LISTEN | grep "$PROJECT_ROOT/web" | awk '{print $9}' | cut -d: -f2 | head -1)
    if [ -z "$FRONTEND_PORT" ]; then
        FRONTEND_PORT="5173 or 3000"
    fi
fi

echo ""
echo "=== Services Running ==="
echo "Backend:  http://localhost:$BACKEND_PORT (PID: $BACKEND_PID)"
echo "Frontend: http://localhost:$FRONTEND_PORT (PID: $FRONTEND_PID)"
echo ""
echo "Log files:"
echo "  Backend:  $PROJECT_ROOT/backend.log"
echo "  Frontend: $PROJECT_ROOT/frontend.log"
echo ""
echo "To view logs:"
echo "  tail -f $PROJECT_ROOT/backend.log"
echo "  tail -f $PROJECT_ROOT/frontend.log"
echo ""
echo "To stop services:"
echo "  ./stop.sh"
echo "  or: kill $BACKEND_PID $FRONTEND_PID"
echo ""
echo "✓ Restart complete!"
