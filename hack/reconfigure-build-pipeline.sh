#!/bin/bash

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Usage function
usage() {
    echo "Usage: $0 <arch-type> <github-url>"
    echo ""
    echo "Arguments:"
    echo "  <arch-type>    Architecture type: 'single-arch' or 'multi-arch'"
    echo "  <github-url>   GitHub URL to the file (blob or raw URL)"
    echo ""
    echo "Examples:"
    echo "  $0 single-arch https://github.com/openshift/jenkins/blob/7bd0d72b84262d7e1851e8a1ec924309f7f4d560/.tekton/jenkins-rhel9-push.yaml"
    echo "  $0 multi-arch https://raw.githubusercontent.com/openshift/jenkins/7bd0d72b84262d7e1851e8a1ec924309f7f4d560/.tekton/jenkins-rhel9-push.yaml"
    exit 1
}

# Check if arguments are provided
if [ $# -lt 2 ]; then
    echo -e "${RED}Error: Both arch-type and github-url are required${NC}"
    echo ""
    usage
fi

ARCH_TYPE="$1"
GITHUB_URL="$2"

# Validate arch-type
if [[ "$ARCH_TYPE" != "single-arch" && "$ARCH_TYPE" != "multi-arch" ]]; then
    echo -e "${RED}Error: Invalid arch-type '${ARCH_TYPE}'${NC}"
    echo "Must be either 'single-arch' or 'multi-arch'"
    echo ""
    usage
fi

# Parse GitHub URL to extract components
echo -e "${BLUE}Parsing GitHub URL...${NC}"

# Check if it's a raw.githubusercontent.com URL or github.com URL
if [[ "$GITHUB_URL" =~ ^https://raw\.githubusercontent\.com/([^/]+)/([^/]+)/([^/]+)/(.+)$ ]]; then
    # Raw URL format: https://raw.githubusercontent.com/owner/repo/commit/path
    GITHUB_OWNER="${BASH_REMATCH[1]}"
    GITHUB_REPO_NAME="${BASH_REMATCH[2]}"
    COMMIT_ID="${BASH_REMATCH[3]}"
    SOURCE_FILE_PATH="${BASH_REMATCH[4]}"
    SOURCE_URL="$GITHUB_URL"
elif [[ "$GITHUB_URL" =~ ^https://github\.com/([^/]+)/([^/]+)/blob/([^/]+)/(.+)$ ]]; then
    # Blob URL format: https://github.com/owner/repo/blob/commit/path
    GITHUB_OWNER="${BASH_REMATCH[1]}"
    GITHUB_REPO_NAME="${BASH_REMATCH[2]}"
    COMMIT_ID="${BASH_REMATCH[3]}"
    SOURCE_FILE_PATH="${BASH_REMATCH[4]}"
    SOURCE_URL="https://raw.githubusercontent.com/${GITHUB_OWNER}/${GITHUB_REPO_NAME}/${COMMIT_ID}/${SOURCE_FILE_PATH}"
else
    echo -e "${RED}Error: Invalid GitHub URL format${NC}"
    echo "URL must be either:"
    echo "  - https://github.com/owner/repo/blob/commit/path"
    echo "  - https://raw.githubusercontent.com/owner/repo/commit/path"
    exit 1
fi

GITHUB_REPO="${GITHUB_OWNER}/${GITHUB_REPO_NAME}"

# Extract filename from path
SOURCE_FILENAME=$(basename "$SOURCE_FILE_PATH")
FILENAME_WITHOUT_EXT="${SOURCE_FILENAME%.yaml}"

# Configuration
BUILD_PIPELINE_FILE="build-pipeline.yaml"
PIPELINE_RUN_FILE="${FILENAME_WITHOUT_EXT}.yaml"
TEMP_FILE="temp-source.yaml"

# Check if yq is installed
if ! command -v yq &> /dev/null; then
    echo -e "${RED}Error: yq is not installed.${NC}"
    echo "Please install yq first:"
    echo "  - macOS: brew install yq"
    echo "  - Linux: wget https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 -O /usr/bin/yq && chmod +x /usr/bin/yq"
    exit 1
fi

echo -e "${GREEN}✓ Parsed successfully${NC}"
echo ""
echo -e "${BLUE}Configuration:${NC}"
echo "  Architecture: ${ARCH_TYPE}"
echo "  Repository: ${GITHUB_REPO}"
echo "  Commit ID: ${COMMIT_ID}"
echo "  Source File: ${SOURCE_FILE_PATH}"
echo "  Output File: ${PIPELINE_RUN_FILE}"
echo ""
echo -e "${BLUE}Downloading source YAML from GitHub...${NC}"
if ! curl -sSL "$SOURCE_URL" -o "$TEMP_FILE"; then
    echo -e "${RED}Error: Failed to download file from URL${NC}"
    echo "URL: $SOURCE_URL"
    exit 1
fi

echo -e "${BLUE}Extracting Pipeline definition...${NC}"

# Create Pipeline from pipelineSpec
yq eval '
  {
    "apiVersion": "tekton.dev/v1",
    "kind": "Pipeline",
    "metadata": {
      "name": "build-pipeline"
    },
    "spec": .spec.pipelineSpec
  }
' "$TEMP_FILE" > "$BUILD_PIPELINE_FILE"

echo -e "${GREEN}✓ Created $BUILD_PIPELINE_FILE${NC}"

echo -e "${BLUE}Extracting PipelineRun definition...${NC}"

# Create PipelineRun with pipelineRef instead of pipelineSpec
# Delete pipelineSpec and add pipelineRef
yq eval '
  del(.spec.pipelineSpec) |
  .spec.pipelineRef = {
    "name": "build-pipeline"
  }
' "$TEMP_FILE" > "$PIPELINE_RUN_FILE"

echo -e "${BLUE}Adding additional parameters...${NC}"

# Add the three additional parameters
yq eval -i '
  .spec.params += [
    {
      "name": "build-source-image",
      "value": "true"
    },
    {
      "name": "hermetic",
      "value": "true"
    },
    {
      "name": "prefetch-input",
      "value": "{\"packages\": [{\"type\": \"generic\"},{\"type\": \"rpm\"}]}"
    }
  ]
' "$PIPELINE_RUN_FILE"

# Add additional platforms if multi-arch
if [ "$ARCH_TYPE" = "multi-arch" ]; then
    echo -e "${BLUE}Adding multi-arch platforms...${NC}"
    yq eval -i '
      (.spec.params[] | select(.name == "build-platforms") | .value) += [
        "linux/arm64",
        "linux/ppc64le",
        "linux/s390x"
      ]
    ' "$PIPELINE_RUN_FILE"
    echo -e "${GREEN}✓ Added arm64, ppc64le, and s390x platforms${NC}"
fi

echo -e "${GREEN}✓ Created $PIPELINE_RUN_FILE with additional parameters${NC}"

# Clean up
rm -f "$TEMP_FILE"

echo ""
echo -e "${GREEN}✓ Successfully extracted and created both files!${NC}"
echo ""
echo "Source:"
echo "  Architecture: ${ARCH_TYPE}"
echo "  Repository: ${GITHUB_REPO}"
echo "  Commit: ${COMMIT_ID}"
echo "  URL: ${GITHUB_URL}"
echo ""
echo "Generated files:"
echo "  1. $BUILD_PIPELINE_FILE - Reusable Pipeline definition"
echo "  2. $PIPELINE_RUN_FILE - PipelineRun with additional parameters"
echo ""
echo "Additional parameters added to $PIPELINE_RUN_FILE:"
echo "  - build-source-image: true"
echo "  - hermetic: true"
echo "  - prefetch-input: {\"packages\": [{\"type\": \"generic\"},{\"type\": \"rpm\"}]}"
if [ "$ARCH_TYPE" = "multi-arch" ]; then
    echo "  - build-platforms: linux/x86_64, linux/arm64, linux/ppc64le, linux/s390x"
else
    echo "  - build-platforms: linux/x86_64"
fi

