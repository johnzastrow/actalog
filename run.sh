#!/bin/bash
# Convenience script to run ActaLog in production mode

# Navigate to project directory
cd "$(dirname "$0")"

# Export Go paths
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin

# Run the application
echo "Starting ActaLog..."
echo "Backend will run on http://localhost:8080"
echo "Press Ctrl+C to stop"
echo ""

./bin/actalog
