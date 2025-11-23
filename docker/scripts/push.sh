#!/bin/bash
# Push script for ActaLog Docker image to GitHub Container Registry
# Usage: ./docker/scripts/push.sh [tag]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
REGISTRY="ghcr.io"
IMAGE_NAME="${GITHUB_REPOSITORY:-yourusername/actalog}"
TAG="${1:-dev}"

# Print banner
echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}   ActaLog Docker Push Script${NC}"
echo -e "${GREEN}============================================${NC}"
echo ""

echo -e "${YELLOW}Registry: ${REGISTRY}${NC}"
echo -e "${YELLOW}Image: ${IMAGE_NAME}${NC}"
echo -e "${YELLOW}Tag: ${TAG}${NC}"
echo ""

# Check if logged in to GitHub Container Registry
echo -e "${GREEN}Checking authentication...${NC}"
if ! docker info 2>/dev/null | grep -q "${REGISTRY}"; then
    echo -e "${YELLOW}Not logged in to ${REGISTRY}${NC}"
    echo ""
    echo "To log in, run:"
    echo -e "  ${YELLOW}echo \$GITHUB_TOKEN | docker login ${REGISTRY} -u \$GITHUB_USERNAME --password-stdin${NC}"
    echo ""
    echo "Or use a Personal Access Token:"
    echo -e "  ${YELLOW}docker login ${REGISTRY}${NC}"
    echo ""
    read -p "Do you want to log in now? (y/n) " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker login "${REGISTRY}"
    else
        echo -e "${RED}Aborted${NC}"
        exit 1
    fi
fi

# Extract build number
if [ -f "pkg/version/version.go" ]; then
    BUILD=$(grep -E "^\s*Build\s*=\s*[0-9]+" pkg/version/version.go | awk '{print $3}')
fi

# Push the images
echo -e "${GREEN}Pushing Docker image...${NC}"
docker push "${REGISTRY}/${IMAGE_NAME}:${TAG}"

if [ ! -z "${BUILD}" ]; then
    echo -e "${GREEN}Pushing build-specific tag...${NC}"
    docker push "${REGISTRY}/${IMAGE_NAME}:build-${BUILD}"
fi

echo ""
echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}   Push Complete!${NC}"
echo -e "${GREEN}============================================${NC}"
echo ""
echo -e "Image pushed: ${GREEN}${REGISTRY}/${IMAGE_NAME}:${TAG}${NC}"
if [ ! -z "${BUILD}" ]; then
    echo -e "Also pushed: ${GREEN}${REGISTRY}/${IMAGE_NAME}:build-${BUILD}${NC}"
fi
echo ""
echo "To pull this image:"
echo -e "  ${YELLOW}docker pull ${REGISTRY}/${IMAGE_NAME}:${TAG}${NC}"
echo ""
echo "To make this image public, go to:"
echo -e "  ${YELLOW}https://github.com/${IMAGE_NAME}/pkgs/container/actalog/settings${NC}"
echo ""
