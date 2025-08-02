#!/bin/bash
set -e

# Prompt for version
read -p "Enter new version (e.g., v1.2.3): " VERSION
if [[ -z "$VERSION" ]]; then
  echo "Version cannot be empty."
  exit 1
fi

PLATFORMS=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")
RELEASE_DIR="release/$VERSION"
mkdir -p "$RELEASE_DIR"

for PLATFORM in "${PLATFORMS[@]}"; do
  OS="${PLATFORM%%/*}"
  ARCH="${PLATFORM##*/}"
  EXT=""
  if [ "$OS" == "windows" ]; then EXT=".exe"; fi
  OUTPUT="$RELEASE_DIR/colade-$OS-$ARCH$EXT"
  echo "Building $OUTPUT..."
  GOOS=$OS GOARCH=$ARCH go build -ldflags "-X main.version=$VERSION" -o "$OUTPUT"
done

echo "Tagging release in git..."
git tag "$VERSION"
git push origin "$VERSION"

echo "Creating GitHub release and uploading assets..."
gh release create "$VERSION" $RELEASE_DIR/* --title "$VERSION" --notes "Release $VERSION"

echo "Release $VERSION complete!"
