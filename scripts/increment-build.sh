#!/bin/bash
# Script to increment the build number in pkg/version/version.go

VERSION_FILE="pkg/version/version.go"

if [ ! -f "$VERSION_FILE" ]; then
  echo "Error: $VERSION_FILE not found!"
  exit 1
fi

# Extract current build number
CURRENT_BUILD=$(grep "Build = " "$VERSION_FILE" | sed 's/.*Build = \([0-9]*\)/\1/')

if [ -z "$CURRENT_BUILD" ]; then
  echo "Error: Could not find Build number in $VERSION_FILE"
  exit 1
fi

# Increment build number
NEW_BUILD=$((CURRENT_BUILD + 1))

# Update the file (macOS/BSD compatible)
if [[ "$OSTYPE" == "darwin"* ]]; then
  # macOS
  sed -i '' "s/Build = $CURRENT_BUILD/Build = $NEW_BUILD/" "$VERSION_FILE"
else
  # Linux
  sed -i "s/Build = $CURRENT_BUILD/Build = $NEW_BUILD/" "$VERSION_FILE"
fi

echo "Build number incremented: $CURRENT_BUILD -> $NEW_BUILD"
echo "Updated $VERSION_FILE"
