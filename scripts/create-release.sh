#!/usr/bin/env bash
# Create a Forgejo/Codeberg release and upload assets.
# Usage: ./scripts/create-release.sh <tag> <token> [url]
set -euo pipefail

TAG="${1:?Usage: create-release.sh <tag> <token> [url]}"
TOKEN="${2:?Token required}"
URL="${3:-https://codeberg.org}"
REPO="dbus/shushingface"
API="${URL}/api/v1"

# Read changelog
BODY=""
if [ -f CHANGELOG.md ]; then
    BODY=$(cat CHANGELOG.md)
fi

IS_PRE="false"
if echo "$TAG" | grep -q "rc\|alpha\|beta"; then
    IS_PRE="true"
fi

echo "Creating release ${TAG}..."

# Create the release
RELEASE=$(curl -s -X POST "${API}/repos/${REPO}/releases" \
    -H "Authorization: token ${TOKEN}" \
    -H "Content-Type: application/json" \
    -d "{
        \"tag_name\": \"${TAG}\",
        \"name\": \"${TAG}\",
        \"body\": $(echo "$BODY" | python3 -c 'import json,sys; print(json.dumps(sys.stdin.read()))'),
        \"prerelease\": ${IS_PRE}
    }")

RELEASE_ID=$(echo "$RELEASE" | python3 -c "import json,sys; print(json.load(sys.stdin).get('id',''))" 2>/dev/null)

if [ -z "$RELEASE_ID" ] || [ "$RELEASE_ID" = "None" ]; then
    echo "Failed to create release:"
    echo "$RELEASE"
    exit 1
fi

echo "Release created (ID: ${RELEASE_ID})"

# Upload assets from dist/
for FILE in dist/*.tar.gz dist/*.deb; do
    [ -f "$FILE" ] || continue
    FILENAME=$(basename "$FILE")
    echo "Uploading ${FILENAME}..."
    curl -s -X POST "${API}/repos/${REPO}/releases/${RELEASE_ID}/assets?name=${FILENAME}" \
        -H "Authorization: token ${TOKEN}" \
        -H "Content-Type: application/octet-stream" \
        --data-binary "@${FILE}"
    echo " Done."
done

echo "Release ${TAG} published."
