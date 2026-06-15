#!/bin/sh
set -e

# Universal CLI build script for cross-platform distribution

DIST_DIR="dist"
echo "Cleaning and preparing target directory: ${DIST_DIR}/..."
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# Target matrices
# format: GOOS/GOARCH/OUTPUT_NAME
targets="
linux/amd64/pass_analy_linux_amd64
linux/arm64/pass_analy_linux_arm64
darwin/amd64/pass_analy_darwin_amd64
darwin/arm64/pass_analy_darwin_arm64
windows/amd64/pass_analy_windows_amd64.exe
windows/arm64/pass_analy_windows_arm64.exe
"

echo "Beginning multi-compilation pipelines..."
for target in $targets; do
  GOOS=$(echo "$target" | cut -d'/' -f1)
  GOARCH=$(echo "$target" | cut -d'/' -f2)
  OUT_NAME=$(echo "$target" | cut -d'/' -f3)

  echo "  -> Building for ${GOOS}/${GOARCH}..."
  # -ldflags="-s -w" strips debug information and symbols to yield smaller static binaries
  env GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="-s -w" -o "${DIST_DIR}/${OUT_NAME}" .
done

echo "Cross-compilation pipelines successfully finished. Binaries ready in ${DIST_DIR}/:"
ls -lh "$DIST_DIR"
