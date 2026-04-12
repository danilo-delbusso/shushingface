#!/bin/bash
set -e

# Generate THIRD_PARTY_LICENSES.md and compare with committed version.
# Usage: ./scripts/check-licenses.sh [--update]

command -v go-licenses &>/dev/null || go install github.com/google/go-licenses@latest

OUT=$(mktemp)

cat > "$OUT" << 'HEADER'
# Third-Party Licenses

## Assets

- **OpenMoji** — [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/)
  All emojis designed by [OpenMoji](https://openmoji.org) — the open-source emoji and icon project.

## Go Dependencies

HEADER

go-licenses csv --tags webkit2_41 ./... 2>/dev/null | grep -v "codeberg.org/dbus/shushingface" | sort | while IFS=, read -r mod url license; do
  echo "- **$mod** — $license" >> "$OUT"
done

echo "" >> "$OUT"
echo "## Frontend Dependencies" >> "$OUT"
echo "" >> "$OUT"

cd frontend
bun x license-checker --production --csv --excludePackages "frontend@0.0.0" 2>/dev/null | tail -n +2 | sort | while IFS=, read -r mod license repo; do
  mod=$(echo "$mod" | tr -d '"')
  license=$(echo "$license" | tr -d '"')
  echo "- **$mod** — $license" >> "$OUT"
done
cd ..

if [ "$1" = "--update" ]; then
  cp "$OUT" THIRD_PARTY_LICENSES.md
  echo "Updated THIRD_PARTY_LICENSES.md ($(grep -c '^\-' THIRD_PARTY_LICENSES.md) entries)"
else
  if ! diff -q THIRD_PARTY_LICENSES.md "$OUT" &>/dev/null; then
    echo "THIRD_PARTY_LICENSES.md is out of date. Run 'just licenses' and commit."
    diff THIRD_PARTY_LICENSES.md "$OUT" || true
    rm "$OUT"
    exit 1
  fi
  echo "Licenses up to date."
fi

rm -f "$OUT"
