#!/bin/bash

# Build script for Harness Remote Migrator
# Usage: ./build.sh [version] [commit] [date]

set -e

# Default values
VERSION=${1:-"dev"}
GIT_COMMIT=${2:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}
BUILD_DATE=${3:-$(date -u +'%Y-%m-%dT%H:%M:%SZ')}

echo "Building Harness Remote Migrator..."
echo "Version: $VERSION"
echo "Git Commit: $GIT_COMMIT"
echo "Build Date: $BUILD_DATE"
echo ""

# Build flags
LDFLAGS="-X 'main.Version=${VERSION}' -X 'main.GitCommit=${GIT_COMMIT}' -X 'main.BuildDate=${BUILD_DATE}'"

# Create build directory
mkdir -p build

echo "Building for current platform..."
PLATFORM_OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
PLATFORM_ARCH="$(uname -m | sed 's/x86_64/amd64/')"
BINARY_NAME="harness-remote-migrator-${VERSION}-${PLATFORM_OS}-${PLATFORM_ARCH}"
go build -ldflags="$LDFLAGS" -o "$BINARY_NAME" .

echo "Testing version output..."
./"$BINARY_NAME" --version

echo ""
echo "Creating archive with clean binary name..."
CLEAN_BINARY="harness-remote-migrator"
ARCHIVE_NAME="harness-remote-migrator-${VERSION}-${PLATFORM_OS}-${PLATFORM_ARCH}.tar.gz"

cp "$BINARY_NAME" "$CLEAN_BINARY"
tar -czf "$ARCHIVE_NAME" "$CLEAN_BINARY"
rm "$CLEAN_BINARY"

echo ""
echo "✅ Build completed successfully!"
echo "Binary: ./$BINARY_NAME"
echo "Archive: ./$ARCHIVE_NAME (contains 'harness-remote-migrator')"
echo ""
echo "To build for multiple platforms, use the GitHub Actions workflow or run:"
echo "GOOS=linux GOARCH=amd64 go build -ldflags=\"$LDFLAGS\" -o build/harness-remote-migrator-${VERSION}-linux-amd64 ."