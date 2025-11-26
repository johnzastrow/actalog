#!/bin/bash
# Build script for ActaLog Docker image
# Usage: ./docker/scripts/build.sh [tag]
# Can be run from anywhere in the repository

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

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
    VERSION=$(grep -E "^\s*Major\s*=\s*[0-9]+" pkg/version/version.go | awk '{print $3}')
    MINOR=$(grep -E "^\s*Minor\s*=\s*[0-9]+" pkg/version/version.go | awk '{print $3}')
    PATCH=$(grep -E "^\s*Patch\s*=\s*[0-9]+" pkg/version/version.go | awk '{print $3}')
    BUILD=$(grep -E "^\s*Build\s*=\s*[0-9]+" pkg/version/version.go | awk '{print $3}')
    echo -e "${YELLOW}Version: ${VERSION}.${MINOR}.${PATCH} (Build ${BUILD})${NC}"
fi

echo -e "${YELLOW}Registry: ${REGISTRY}${NC}"
echo -e "${YELLOW}Image: ${IMAGE_NAME}${NC}"
echo -e "${YELLOW}Tag: ${TAG}${NC}"
echo -e "${YELLOW}Platform: ${PLATFORM}${NC}"
echo ""

# Build the image
echo -e "${GREEN}Building Docker image...${NC}"
docker buildx build \
    --platform "${PLATFORM}" \
    --tag "${REGISTRY}/${IMAGE_NAME}:${TAG}" \
    --tag "${REGISTRY}/${IMAGE_NAME}:build-${BUILD}" \
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
