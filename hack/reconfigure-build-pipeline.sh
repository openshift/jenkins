#!/bin/bash

set -euo pipefail

#################################################################################
# CONFIGURATION - Modify these values as needed
#################################################################################

# Additional files to add to pipelinesascode CEL expression (space-separated)
ADDITIONAL_PATHCHANGED_FILES="rpms.lock.yaml rpms.in.yaml ubi.repo artifacts.lock.yaml"

# Additional parameters to add to the PipelineRun
BUILD_SOURCE_IMAGE="true"
HERMETIC="true"
PREFETCH_INPUT='{"packages": [{"type": "generic"},{"type": "rpm"}]}'

# Build arguments (array items, one per line)
BUILD_ARGS_ITEMS=(
    "CI_UPSTREAM_COMMIT={{revision}}"
    "CI_UPSTREAM_URL={{source_url}}"
)

# Build args file path
BUILD_ARGS_FILE="konflux-build-args.env"

# Multi-arch platforms (in addition to linux/x86_64)
MULTI_ARCH_PLATFORMS=(
    "linux/arm64"
    "linux/ppc64le"
    "linux/s390x"
)

#################################################################################
# END CONFIGURATION
#################################################################################

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
    echo "Configuration:"
    echo "  Edit the CONFIGURATION section at the top of this script to customize:"
    echo "  - Additional files for CEL expression"
    echo "  - Build parameters (hermetic, prefetch-input, etc.)"
    echo "  - Build arguments"
    echo "  - Multi-arch platforms"
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

# Configuration - Output to .tekton directory
OUTPUT_DIR=".tekton"
BUILD_PIPELINE_FILE="${OUTPUT_DIR}/build-pipeline.yaml"
PIPELINE_RUN_FILE="${OUTPUT_DIR}/${FILENAME_WITHOUT_EXT}.yaml"
TEMP_FILE="temp-source.yaml"

# Create .tekton directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

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
echo "  Additional Files: ${ADDITIONAL_PATHCHANGED_FILES}"
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

# Build the build-args array from configuration
BUILD_ARGS_JSON="["
for i in "${!BUILD_ARGS_ITEMS[@]}"; do
    if [ $i -gt 0 ]; then
        BUILD_ARGS_JSON="${BUILD_ARGS_JSON},"
    fi
    BUILD_ARGS_JSON="${BUILD_ARGS_JSON}\"${BUILD_ARGS_ITEMS[$i]}\""
done
BUILD_ARGS_JSON="${BUILD_ARGS_JSON}]"

# Write PREFETCH_INPUT to temp file (complex JSON value)
echo -n "$PREFETCH_INPUT" > "${TEMP_FILE}.prefetch-input"

# Add the additional parameters using configuration values
yq eval -i '
  .spec.params += [
    {
      "name": "build-source-image",
      "value": "'"${BUILD_SOURCE_IMAGE}"'"
    },
    {
      "name": "hermetic",
      "value": "'"${HERMETIC}"'"
    },
    {
      "name": "prefetch-input",
      "value": load_str("'"${TEMP_FILE}"'.prefetch-input")
    },
    {
      "name": "build-args",
      "value": '"${BUILD_ARGS_JSON}"'
    },
    {
      "name": "build-args-file",
      "value": "'"${BUILD_ARGS_FILE}"'"
    }
  ]
' "$PIPELINE_RUN_FILE"

# Clean up temp file
rm -f "${TEMP_FILE}.prefetch-input"

# Add additional platforms if multi-arch
if [ "$ARCH_TYPE" = "multi-arch" ]; then
    echo -e "${BLUE}Adding multi-arch platforms...${NC}"
    
    # Build the platforms array from configuration
    PLATFORMS_JSON="["
    for i in "${!MULTI_ARCH_PLATFORMS[@]}"; do
        if [ $i -gt 0 ]; then
            PLATFORMS_JSON="${PLATFORMS_JSON},"
        fi
        PLATFORMS_JSON="${PLATFORMS_JSON}\"${MULTI_ARCH_PLATFORMS[$i]}\""
    done
    PLATFORMS_JSON="${PLATFORMS_JSON}]"
    
    yq eval -i "
      (.spec.params[] | select(.name == \"build-platforms\") | .value) += ${PLATFORMS_JSON}
    " "$PIPELINE_RUN_FILE"
    echo -e "${GREEN}✓ Added platforms: ${MULTI_ARCH_PLATFORMS[*]}${NC}"
fi

# Update pipelinesascode on-cel-expression annotation if it exists
echo -e "${BLUE}Updating pipelinesascode CEL expression...${NC}"
if yq eval '.metadata.annotations["pipelinesascode.tekton.dev/on-cel-expression"]' "$PIPELINE_RUN_FILE" | grep -q "pathChanged"; then
    # Get the current expression
    CURRENT_CEL=$(yq eval '.metadata.annotations["pipelinesascode.tekton.dev/on-cel-expression"]' "$PIPELINE_RUN_FILE")
    
    # Build the CEL expression addition from the file list
    CEL_ADDITION=""
    for file in $ADDITIONAL_PATHCHANGED_FILES; do
        CEL_ADDITION="${CEL_ADDITION} || \"${file}\".pathChanged()"
    done
    
    # Remove trailing ) and add new files, then add ) back
    NEW_CEL="${CURRENT_CEL% )}${CEL_ADDITION} )"
    
    # Write the new expression to a temporary file and use yq to load it
    echo "$NEW_CEL" > "${TEMP_FILE}.cel"
    yq eval -i ".metadata.annotations[\"pipelinesascode.tekton.dev/on-cel-expression\"] = load_str(\"${TEMP_FILE}.cel\")" "$PIPELINE_RUN_FILE"
    rm -f "${TEMP_FILE}.cel"
    
    echo -e "${GREEN}✓ Added files to CEL expression: ${ADDITIONAL_PATHCHANGED_FILES}${NC}"
else
    echo -e "${YELLOW}⚠ No pipelinesascode CEL expression found, skipping${NC}"
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
echo "Modifications to $PIPELINE_RUN_FILE:"
echo ""
echo "Additional parameters:"
echo "  - build-source-image: ${BUILD_SOURCE_IMAGE}"
echo "  - hermetic: ${HERMETIC}"
echo "  - prefetch-input: ${PREFETCH_INPUT}"
echo "  - build-args: [$(IFS=, ; echo "${BUILD_ARGS_ITEMS[*]}")]"
echo "  - build-args-file: ${BUILD_ARGS_FILE}"
if [ "$ARCH_TYPE" = "multi-arch" ]; then
    echo "  - build-platforms: linux/x86_64, ${MULTI_ARCH_PLATFORMS[*]// /, }"
else
    echo "  - build-platforms: linux/x86_64"
fi
echo ""
echo "CEL expression updated with additional files:"
for file in $ADDITIONAL_PATHCHANGED_FILES; do
    echo "  - ${file}"
done

