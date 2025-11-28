#!/bin/bash
# Build script for ActaLog Docker image
# Usage: ./docker/scripts/build.sh [tag]
# Can be run from anywhere in the repository

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# ============================================================================
# USER-EDITABLE LABELS - Customize these for your deployment
# ============================================================================
# Description: A short text description of the image (max 512 characters)
# This will appear on the GitHub package page below the package name
IMAGE_DESCRIPTION="ActaLog - A mobile-first CrossFit workout tracker. Track WODs, strength training, personal records, and workout history. Built with Go backend (Chi router, SQLite/MariaDB/PostgreSQL) and Vue.js 3 frontend with Vuetify 3."

# Vendor/Organization name
IMAGE_VENDOR="John Zastrow"

# License identifier (SPDX format recommended)
IMAGE_LICENSE="MIT"

# Authors (comma-separated if multiple)
IMAGE_AUTHORS="John Zastrow"

# Documentation URL (optional)
IMAGE_DOCUMENTATION=""
# ============================================================================

# Default values
REGISTRY="ghcr.io"
IMAGE_NAME="${GITHUB_REPOSITORY:-johnzastrow/actalog}"
TAG="${1:-latest}"
PLATFORM="${DOCKER_PLATFORM:-linux/amd64}"

# Print banner
echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}   ActaLog Docker Build Script${NC}"
echo -e "${GREEN}============================================${NC}"
echo ""

# Find repository root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Change to repository root
if [ ! -f "${REPO_ROOT}/docker/Dockerfile" ]; then
    echo -e "${RED}Error: Could not find repository root${NC}"
    echo "Script location: ${SCRIPT_DIR}"
    echo "Expected repo root: ${REPO_ROOT}"
    echo "docker/Dockerfile not found at expected location"
    exit 1
fi

echo -e "${YELLOW}Repository: ${REPO_ROOT}${NC}"
cd "${REPO_ROOT}"

# Extract version from pkg/version/version.go if available
if [ -f "pkg/version/version.go" ]; then
    VERSION_MAJOR=$(grep -E "^\s*Major\s*=\s*[0-9]+" pkg/version/version.go | awk '{print $3}')
    VERSION_MINOR=$(grep -E "^\s*Minor\s*=\s*[0-9]+" pkg/version/version.go | awk '{print $3}')
    VERSION_PATCH=$(grep -E "^\s*Patch\s*=\s*[0-9]+" pkg/version/version.go | awk '{print $3}')
    VERSION_PRERELEASE=$(grep -E "^\s*PreRelease\s*=" pkg/version/version.go | sed 's/.*"\(.*\)".*/\1/')
    BUILD=$(grep -E "^\s*Build\s*=\s*[0-9]+" pkg/version/version.go | awk '{print $3}')

    # Construct full version string
    if [ -n "$VERSION_PRERELEASE" ]; then
        FULL_VERSION="${VERSION_MAJOR}.${VERSION_MINOR}.${VERSION_PATCH}-${VERSION_PRERELEASE}"
    else
        FULL_VERSION="${VERSION_MAJOR}.${VERSION_MINOR}.${VERSION_PATCH}"
    fi
fi

# Extract dynamic build information
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_COMMIT_FULL=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
GIT_DIRTY=$(git diff --quiet 2>/dev/null || echo "-dirty")
SOURCE_URL="https://github.com/${IMAGE_NAME}"

echo -e "${YELLOW}Version: ${FULL_VERSION} (Build ${BUILD})${NC}"
echo -e "${YELLOW}Registry: ${REGISTRY}${NC}"
echo -e "${YELLOW}Image: ${IMAGE_NAME}${NC}"
echo -e "${YELLOW}Tag: ${TAG}${NC}"
echo -e "${YELLOW}Platform: ${PLATFORM}${NC}"
echo ""
echo -e "${CYAN}Build Metadata:${NC}"
echo -e "${CYAN}  Git Commit: ${GIT_COMMIT}${GIT_DIRTY}${NC}"
echo -e "${CYAN}  Git Branch: ${GIT_BRANCH}${NC}"
echo -e "${CYAN}  Build Date: ${BUILD_DATE}${NC}"
echo ""
echo -e "${CYAN}Image Labels (edit in script header):${NC}"
echo -e "${CYAN}  Description: ${IMAGE_DESCRIPTION:0:60}...${NC}"
echo -e "${CYAN}  Vendor: ${IMAGE_VENDOR}${NC}"
echo -e "${CYAN}  License: ${IMAGE_LICENSE}${NC}"
echo ""

# Build the image with OCI labels
echo -e "${GREEN}Building Docker image...${NC}"
docker buildx build \
    --platform "${PLATFORM}" \
    --tag "${REGISTRY}/${IMAGE_NAME}:${TAG}" \
    --tag "${REGISTRY}/${IMAGE_NAME}:build-${BUILD}" \
    --label "org.opencontainers.image.title=ActaLog" \
    --label "org.opencontainers.image.description=${IMAGE_DESCRIPTION}" \
    --label "org.opencontainers.image.version=${FULL_VERSION}+build.${BUILD}" \
    --label "org.opencontainers.image.created=${BUILD_DATE}" \
    --label "org.opencontainers.image.revision=${GIT_COMMIT_FULL}" \
    --label "org.opencontainers.image.source=${SOURCE_URL}" \
    --label "org.opencontainers.image.url=${SOURCE_URL}" \
    --label "org.opencontainers.image.vendor=${IMAGE_VENDOR}" \
    --label "org.opencontainers.image.licenses=${IMAGE_LICENSE}" \
    --label "org.opencontainers.image.authors=${IMAGE_AUTHORS}" \
    --label "org.opencontainers.image.ref.name=${GIT_BRANCH}" \
    --label "org.opencontainers.image.base.name=alpine:latest" \
    --label "build.number=${BUILD}" \
    --label "build.git.commit=${GIT_COMMIT}${GIT_DIRTY}" \
    --label "build.git.branch=${GIT_BRANCH}" \
    --file docker/Dockerfile \
    --load \
    .

echo ""
echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}   Build Complete!${NC}"
echo -e "${GREEN}============================================${NC}"
echo ""
echo -e "Image: ${GREEN}${REGISTRY}/${IMAGE_NAME}:${TAG}${NC}"
echo -e "Also tagged: ${GREEN}${REGISTRY}/${IMAGE_NAME}:build-${BUILD}${NC}"
echo ""
echo "To run locally:"
echo -e "  ${YELLOW}docker run -p 8080:8080 ${REGISTRY}/${IMAGE_NAME}:${TAG}${NC}"
echo ""
echo "To push to registry:"
echo -e "  ${YELLOW}./docker/scripts/push.sh ${TAG}${NC}"
echo ""
echo "To view image labels:"
echo -e "  ${YELLOW}docker inspect ${REGISTRY}/${IMAGE_NAME}:${TAG} --format '{{json .Config.Labels}}' | jq${NC}"
echo ""
echo -e "${CYAN}Tip: Edit the USER-EDITABLE LABELS section at the top of this script${NC}"
echo -e "${CYAN}     to customize the image description, vendor, license, and authors.${NC}"
echo ""
