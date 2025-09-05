# Versioning Guide

This document explains how versioning works in the Harness Remote Migrator.

## Version Information

The binary includes version information that can be displayed using the `--version` flag:

```bash
./harness-remote-migrator --version
```

This displays:
- **Version**: The release version (e.g., v1.2.3)
- **Git Commit**: The git commit hash when built
- **Build Date**: When the binary was built
- **Go Version**: The Go version used to build
- **OS/Arch**: The target operating system and architecture

## Local Development

### Quick Build
```bash
# Build with default development version
go build -o harness-remote-migrator .

# Build with custom version using the build script
./build.sh v1.0.0-beta
```

### Manual Build with Version
```bash
# Set version variables
VERSION="v1.2.3"
GIT_COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

# Build with version information
go build -ldflags="-X 'main.Version=${VERSION}' -X 'main.GitCommit=${GIT_COMMIT}' -X 'main.BuildDate=${BUILD_DATE}'" -o harness-remote-migrator .
```

## GitHub Releases

### Automatic Release Process

1. **Create a GitHub Release** through the GitHub UI or using `gh` CLI:
   ```bash
   # Using GitHub CLI
   gh release create v1.2.3 --title "Release v1.2.3" --notes "Release notes here"
   
   # Or create manually through GitHub web interface
   ```

2. **GitHub Actions automatically**:
   - Triggers on release creation (`release: types: [created]`)
   - Uses standard `actions/setup-go@v5` for reliable Go setup
   - Builds binaries for multiple platforms using matrix strategy
   - Injects comprehensive version information (version, git commit, build date)
   - Uploads binaries and SHA256 checksums as release assets
   - Tests the Linux AMD64 binary to verify version injection works

### Supported Platforms

The release workflow builds for:
- Linux AMD64
- Linux ARM64
- macOS AMD64
- macOS ARM64 (Apple Silicon)
- Windows AMD64

**Note**: Windows ARM64 is excluded from the matrix as it's not commonly needed.

### Workflow Architecture

The release workflow uses a reliable, standard approach:

```yaml
# Key workflow steps:
1. Checkout code with full git history
2. Set up Go 1.21 using actions/setup-go@v5
3. Extract version info (tag, commit, build date)
4. Build binary with ldflags injection
5. Generate SHA256 checksums
6. Upload binary and checksum to release
7. Test version output (Linux AMD64 only)
```

### Release Assets

Each release includes (per platform):
- **Binary**: `harness-remote-migrator-{version}-{os}-{arch}[.exe]`
- **Binary Checksum**: `harness-remote-migrator-{version}-{os}-{arch}[.exe].sha256`
- **Compressed Archive**: `harness-remote-migrator-{version}-{os}-{arch}.tar.gz` (Unix) or `.zip` (Windows)
- **Archive Checksum**: `harness-remote-migrator-{version}-{os}-{arch}.tar.gz.sha256` or `.zip.sha256`

**Archive Contents**:
- Unix archives (`.tar.gz`) contain: `harness-remote-migrator`
- Windows archives (`.zip`) contain: `harness-remote-migrator.exe`

**Example for v1.5.5**:
- `harness-remote-migrator-v1.5.5-linux-amd64` (binary)
- `harness-remote-migrator-v1.5.5-linux-amd64.sha256`
- `harness-remote-migrator-v1.5.5-linux-amd64.tar.gz` (contains `harness-remote-migrator`)
- `harness-remote-migrator-v1.5.5-linux-amd64.tar.gz.sha256`
- `harness-remote-migrator-v1.5.5-windows-amd64.exe` (binary)
- `harness-remote-migrator-v1.5.5-windows-amd64.zip` (contains `harness-remote-migrator.exe`)

**Total**: 4 files per platform (20 files for 5 platforms)

**Benefits of archives**:
- ~46% smaller download size
- Platform-appropriate format (tar.gz for Unix, zip for Windows)
- Clean binary name inside archive for easy deployment
- Easy extraction:
  - Unix: `tar -xzf archive.tar.gz`
  - Windows: `unzip archive.zip` or double-click

### Manual Release Testing

To test the release process locally:
```bash
# Test multi-platform build
VERSION="v1.2.3"
GIT_COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS="-X 'main.Version=${VERSION}' -X 'main.GitCommit=${GIT_COMMIT}' -X 'main.BuildDate=${BUILD_DATE}'"

# Build for different platforms with version in filename
GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o build/harness-remote-migrator-${VERSION}-linux-amd64 .
GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -o build/harness-remote-migrator-${VERSION}-windows-amd64.exe .
GOOS=darwin GOARCH=arm64 go build -ldflags="$LDFLAGS" -o build/harness-remote-migrator-${VERSION}-darwin-arm64 .
```

## Version Schema

We follow [Semantic Versioning](https://semver.org/):
- `v1.0.0` - Major release
- `v1.1.0` - Minor release (new features)
- `v1.1.1` - Patch release (bug fixes)
- `v1.0.0-beta.1` - Pre-release

## Troubleshooting

### Common Issues

1. **Version shows "dev"**: Binary was built without ldflags injection
   ```bash
   # Fix: Use build script or manual ldflags
   ./build.sh v1.0.0
   ```

2. **GitHub Actions fails**: Check the Actions tab for detailed logs
   ```bash
   # Common fixes:
   - Ensure go.mod exists in repository root
   - Verify release was created properly
   - Check repository permissions
   ```

3. **Binary won't run**: Download the correct platform binary
   ```bash
   # Check your platform
   uname -m    # Architecture (amd64, arm64)
   uname -s    # OS (Linux, Darwin, Windows)
   ```

### Verification

Verify downloaded releases (example with v1.5.5):

**Direct Binary**:
```bash
# Download and verify binary directly
sha256sum -c harness-remote-migrator-v1.5.5-linux-amd64.sha256
./harness-remote-migrator-v1.5.5-linux-amd64 --version
```

**Archive (recommended for faster download)**:
```bash
# Unix/Linux/macOS
sha256sum -c harness-remote-migrator-v1.5.5-linux-amd64.tar.gz.sha256
tar -xzf harness-remote-migrator-v1.5.5-linux-amd64.tar.gz
./harness-remote-migrator --version

# Windows
# Verify checksum and extract harness-remote-migrator-v1.5.5-windows-amd64.zip
# Contains: harness-remote-migrator.exe (clean name for easy deployment)
```

## Files

- **`.github/workflows/release.yml`** - GitHub Actions workflow for releases
- **`build.sh`** - Local build script  
- **`main.go`** - Contains version variables and `--version` flag
- **`VERSIONING.md`** - This documentation
