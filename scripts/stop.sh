#!/bin/bash

echo "=== Stopping ActaLog Services ==="
echo ""

# Get the absolute path of the project directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_NAME="actionlog"

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

echo "✓ Backend stopped"

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

echo "✓ Frontend stopped"

echo ""
echo "✓ All services stopped"
