#!/usr/bin/env bash
# Generate a changelog from git commits since the last tag.
# Usage: ./scripts/generate-changelog.sh [output-file]
set -euo pipefail

OUT="${1:-CHANGELOG.md}"

PREV_TAG=$(git describe --tags --abbrev=0 HEAD^ 2>/dev/null || echo "")

if [ -n "$PREV_TAG" ]; then
    echo "## Changes since $PREV_TAG" > "$OUT"
    echo "" >> "$OUT"
    git log --pretty=format:"- %s" "${PREV_TAG}..HEAD" >> "$OUT"
else
    echo "## Initial Release" > "$OUT"
    echo "" >> "$OUT"
    git log --pretty=format:"- %s" -20 >> "$OUT"
fi

echo "" >> "$OUT"
echo "Changelog: $OUT ($(wc -l < "$OUT") lines)" >&2
