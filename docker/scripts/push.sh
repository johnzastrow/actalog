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
IMAGE_NAME="${GITHUB_REPOSITORY:-johnzastrow/actalog}"
TAG="${1:-latest}"

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

    # Try to use GitHub CLI authentication automatically
    if command -v gh &> /dev/null; then
        echo -e "${GREEN}Found GitHub CLI (gh), checking authentication...${NC}"

        if gh auth status &> /dev/null; then
            echo -e "${GREEN}GitHub CLI is authenticated, using token for Docker login...${NC}"

            # Get username from gh
            GH_USERNAME=$(gh api user --jq '.login' 2>/dev/null)

            if [ -z "$GH_USERNAME" ]; then
                echo -e "${YELLOW}Could not get GitHub username, using default...${NC}"
                GH_USERNAME="${USER}"
            fi

            # Use gh auth token to log in to GHCR
            if gh auth token | docker login "${REGISTRY}" -u "${GH_USERNAME}" --password-stdin; then
                echo -e "${GREEN}Successfully logged in to ${REGISTRY} using GitHub CLI token${NC}"
            else
                echo -e "${RED}Failed to log in using GitHub CLI token${NC}"
                exit 1
            fi
        else
            echo -e "${YELLOW}GitHub CLI is not authenticated${NC}"
            echo -e "Run: ${GREEN}gh auth login${NC}"
            exit 1
        fi
    else
        # Fall back to manual authentication
        echo ""
        echo "GitHub CLI (gh) not found. To install:"
        echo -e "  ${YELLOW}https://cli.github.com/${NC}"
        echo ""
        echo "Or log in manually with a Personal Access Token:"
        echo -e "  ${YELLOW}docker login ${REGISTRY}${NC}"
        echo ""
        echo "Or use environment variables:"
        echo -e "  ${YELLOW}echo \$GITHUB_TOKEN | docker login ${REGISTRY} -u \$GITHUB_USERNAME --password-stdin${NC}"
        echo ""
        read -p "Do you want to log in manually now? (y/n) " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            docker login "${REGISTRY}"
        else
            echo -e "${RED}Aborted${NC}"
            exit 1
        fi
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
echo "To view image labels:"
echo -e "  ${YELLOW}docker inspect ${REGISTRY}/${IMAGE_NAME}:${TAG} --format '{{json .Config.Labels}}' | jq${NC}"
echo ""
echo "To make this image public, go to:"
echo -e "  ${YELLOW}https://github.com/${IMAGE_NAME}/pkgs/container/actalog/settings${NC}"
echo ""
echo -e "${YELLOW}Note: The image description label (org.opencontainers.image.description)${NC}"
echo -e "${YELLOW}      will appear on the GitHub package page after push.${NC}"
echo ""
