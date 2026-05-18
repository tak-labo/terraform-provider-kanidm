#!/usr/bin/env bash
# Downloads the latest Kanidm OpenAPI spec from GitHub releases.

set -euo pipefail

SPEC_PATH="internal/spec/kanidm-openapi.json"
REPO="kanidm/kanidm"

echo "Fetching latest Kanidm release..."
TAG=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | python3 -c "import sys,json; print(json.load(sys.stdin)['tag_name'])")
echo "Latest: $TAG"

URL="https://github.com/$REPO/releases/download/$TAG/kanidm-openapi.json"
echo "Downloading spec from $URL..."

if curl -fsSL "$URL" -o "$SPEC_PATH" 2>/dev/null; then
    VERSION=$(python3 -c "import json; print(json.load(open('$SPEC_PATH'))['info']['version'])")
    echo "Updated $SPEC_PATH to v$VERSION"
else
    echo "Direct download failed. Spec may not be included in release assets."
    echo "Run Kanidm locally and use: curl -sk https://localhost:8443/docs/v1/openapi.json > $SPEC_PATH"
    exit 1
fi
