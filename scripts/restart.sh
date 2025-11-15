#!/bin/bash

echo "=== ActaLog Restart Script ==="
echo ""

# Get the absolute path of the project directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_NAME="actionlog"

echo "Project directory: $SCRIPT_DIR"
echo ""

# Kill backend processes
echo "Stopping backend..."

# Kill processes with "actalog" binary name
pkill -9 -f "bin/actalog" 2>/dev/null || true

# Kill go run processes in this directory
for pid in $(ps aux | grep "[g]o run" | grep "$SCRIPT_DIR" | awk '{print $2}'); do
    echo "  Killing go run process: $pid"
    kill -9 $pid 2>/dev/null || true
done

# Kill any actalog binary processes
for pid in $(ps aux | grep "[a]ctalog" | grep -v "grep" | awk '{print $2}'); do
    # Check if it's running from our directory
    if [ -d "/proc/$pid" ]; then
        cwd=$(readlink -f /proc/$pid/cwd 2>/dev/null || echo "")
        if [[ "$cwd" == *"$PROJECT_NAME"* ]]; then
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
        if [[ "$cwd" == *"$PROJECT_NAME"* ]] || [[ "$cwd" == *"/web"* ]]; then
            echo "  Killing npm dev process: $pid (from $cwd)"
            kill -9 $pid 2>/dev/null || true
        fi
    fi
done

# Kill vite processes from the web directory
for pid in $(ps aux | grep "[v]ite" | grep -v "grep" | awk '{print $2}'); do
    if [ -d "/proc/$pid" ]; then
        cwd=$(readlink -f /proc/$pid/cwd 2>/dev/null || echo "")
        if [[ "$cwd" == *"$PROJECT_NAME"* ]] || [[ "$cwd" == *"/web"* ]]; then
            echo "  Killing vite process: $pid (from $cwd)"
            kill -9 $pid 2>/dev/null || true
        fi
    fi
done

# Kill node processes from the web directory
for pid in $(ps aux | grep "[n]ode" | grep "vite" | awk '{print $2}'); do
    if [ -d "/proc/$pid" ]; then
        cwd=$(readlink -f /proc/$pid/cwd 2>/dev/null || echo "")
        if [[ "$cwd" == *"$PROJECT_NAME/web"* ]]; then
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
# Build backend
make build
if [ $? -ne 0 ]; then
    echo "❌ Backend build failed!"
    exit 1
fi

# Start backend in background
make run > backend.log 2>&1 &
BACKEND_PID=$!
echo "✓ Backend started (PID: $BACKEND_PID, logs: backend.log)"

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
cd web

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    echo "Installing frontend dependencies..."
    npm install
fi

# Start frontend in background
npm run dev > ../frontend.log 2>&1 &
FRONTEND_PID=$!
cd ..
echo "✓ Frontend started (PID: $FRONTEND_PID, logs: frontend.log)"

# Wait for frontend to start and detect its port
sleep 3
FRONTEND_PORT=$(lsof -Pan -p $FRONTEND_PID -i 2>/dev/null | grep LISTEN | awk '{print $9}' | cut -d: -f2 | head -1)
if [ -z "$FRONTEND_PORT" ]; then
    # Try to find any vite process port
    FRONTEND_PORT=$(lsof -Pan -c node -i 2>/dev/null | grep LISTEN | grep "$SCRIPT_DIR/web" | awk '{print $9}' | cut -d: -f2 | head -1)
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
echo "  Backend:  backend.log"
echo "  Frontend: frontend.log"
echo ""
echo "To view logs:"
echo "  tail -f backend.log"
echo "  tail -f frontend.log"
echo ""
echo "To stop services:"
echo "  ./stop.sh"
echo "  or: kill $BACKEND_PID $FRONTEND_PID"
echo ""
echo "✓ Restart complete!"
