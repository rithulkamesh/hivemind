#!/usr/bin/env bash
# Seed the registry with the test plugin (hivemind-plugin-demo) so you can
# run: pip install --index-url http://localhost:3000/simple/ hivemind-plugin-demo
#
# Requires: API running with storage (S3 or local) configured, and a built wheel.
# Usage:
#   ./registry/scripts/seed-dev.sh
#   ./registry/scripts/seed-dev.sh /path/to/hivemind_plugin_demo-0.1.0-py3-none-any.whl
#
# Env: REGISTRY_URL (default http://localhost:8080), REGISTRY_EMAIL, REGISTRY_PASSWORD

set -e
BASE="${REGISTRY_URL:-http://localhost:8080}"
EMAIL="${REGISTRY_EMAIL:-seed@localhost}"
PASS="${REGISTRY_PASSWORD:-seedpass123}"
PKG_NAME="hivemind-plugin-demo"
WHEEL="${1:-}"

if [ -z "$WHEEL" ]; then
  WHEEL="registry/test-plugin/dist/hivemind_plugin_demo-0.1.0-py3-none-any.whl"
  if [ ! -f "$WHEEL" ]; then
    echo "Build the test plugin first: just test-plugin-build"
    echo "Or pass the wheel path: $0 /path/to/wheel.whl"
    exit 1
  fi
fi

echo "Register (or 409)..."
curl -s -X POST "$BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"username\":\"seeduser\",\"password\":\"$PASS\"}" || true

echo "Login..."
RESP=$(curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASS\"}")
TOKEN=$(echo "$RESP" | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')
if [ -z "$TOKEN" ]; then
  echo "Login failed. Response: $RESP"
  exit 1
fi

echo "Create package..."
curl -s -X POST "$BASE/api/v1/packages" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"$PKG_NAME\",\"display_name\":\"Hivemind Plugin Demo\",\"description\":\"Demo plugin for registry testing\"}" || true

VERSION=$(basename "$WHEEL" | sed -n 's/.*-\([0-9][0-9.]*\)-py.*/\1/p')
if [ -z "$VERSION" ]; then
  VERSION="0.1.0"
fi

echo "Upload wheel (version=$VERSION)..."
UPLOAD=$(curl -s -w "%{http_code}" -X POST "$BASE/api/v1/packages/$PKG_NAME/upload" \
  -H "Authorization: Bearer $TOKEN" \
  -F "version=$VERSION" \
  -F "file=@$WHEEL")
CODE="${UPLOAD: -3}"
if [ "$CODE" != "201" ]; then
  echo "Upload failed (HTTP $CODE). Response: ${UPLOAD%???}"
  exit 1
fi

echo "Done. Install with: pip install --index-url $BASE/simple/ $PKG_NAME"
