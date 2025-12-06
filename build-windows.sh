#!/bin/bash
# Build PinShare for Windows - Cross-platform build script
# Works on: macOS, Linux, Windows (Git Bash/WSL)
#
# On Windows, this script delegates to build-windows.bat for native builds.
# On macOS/Linux, this script cross-compiles Windows binaries.

set -e

echo "=========================================="
echo "Building PinShare for Windows"
echo "=========================================="
echo ""

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DIST_DIR="$SCRIPT_DIR/dist/windows"

# Detect platform
detect_platform() {
    case "$(uname -s)" in
        CYGWIN*|MINGW*|MSYS*)
            echo "windows"
            ;;
        Darwin*)
            echo "darwin"
            ;;
        Linux*)
            echo "linux"
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

PLATFORM=$(detect_platform)
echo "Detected platform: $PLATFORM"

# On Windows, delegate to the batch file for native builds
if [ "$PLATFORM" = "windows" ]; then
    echo "Running native Windows build via build-windows.bat..."
    echo ""
    # Use cmd.exe to run the batch file
    cmd.exe //c "$(cygpath -w "$SCRIPT_DIR/build-windows.bat")" "$@"
    exit $?
fi

# Cross-compilation from macOS/Linux
echo "Cross-compiling Windows binaries..."
echo ""

# Get version from git tag or use default
if git describe --tags --match "v[0-9]*" --abbrev=0 >/dev/null 2>&1; then
    GIT_TAG=$(git describe --tags --match "v[0-9]*" --abbrev=0)
    GIT_VERSION=$(git describe --tags --match "v[0-9]*" 2>/dev/null || echo "$GIT_TAG")
else
    GIT_VERSION="1.0.0"
fi

# Clean up version string (remove 'v' prefix if present)
VERSION="${GIT_VERSION#v}"

# Convert git describe format (1.0.0-5-gabcdef) to MSI-compatible (1.0.0.5)
BASE_VERSION=$(echo "$VERSION" | cut -d'-' -f1)
COMMITS=$(echo "$VERSION" | cut -d'-' -f2)

# Check if COMMITS is numeric (means we have commits after tag)
if [[ "$COMMITS" =~ ^[0-9]+$ ]] && [ "$COMMITS" != "$BASE_VERSION" ]; then
    VERSION="${BASE_VERSION}.${COMMITS}"
else
    VERSION="$BASE_VERSION"
fi

echo "Version: $VERSION"
echo ""

# Create dist directory
mkdir -p "$DIST_DIR"

# Set cross-compilation environment
export GOOS=windows
export GOARCH=amd64

# Check for CGO requirements
# Note: CGO is disabled for cross-compilation by default
# If CGO is needed, you'll need a Windows cross-compiler toolchain
if [ "${CGO_ENABLED:-}" != "1" ]; then
    export CGO_ENABLED=0
    echo "Note: CGO disabled for cross-compilation"
    echo ""
fi

# Build PinShare backend
echo "Building PinShare backend..."
go build -ldflags "-s -w" -o "$DIST_DIR/pinshare.exe" "$SCRIPT_DIR"
echo "[OK] Built: $DIST_DIR/pinshare.exe"
echo ""

# Build Windows service wrapper
echo "Building Windows service wrapper..."
go build -ldflags "-s -w" -o "$DIST_DIR/pinsharesvc.exe" "$SCRIPT_DIR/cmd/pinsharesvc"
echo "[OK] Built: $DIST_DIR/pinsharesvc.exe"
echo ""

# Build system tray application
echo "Building system tray application..."
# Note: -H windowsgui is a Windows linker flag, may not work in cross-compilation
# The binary will still work, but may show a console window briefly
go build -ldflags "-s -w -H windowsgui" -o "$DIST_DIR/pinshare-tray.exe" "$SCRIPT_DIR/cmd/pinshare-tray" 2>/dev/null || \
go build -ldflags "-s -w" -o "$DIST_DIR/pinshare-tray.exe" "$SCRIPT_DIR/cmd/pinshare-tray"
echo "[OK] Built: $DIST_DIR/pinshare-tray.exe"
echo ""

# Copy tray application resources
echo "Copying tray application resources..."
mkdir -p "$DIST_DIR/resources"
cp -r "$SCRIPT_DIR/cmd/pinshare-tray/resources/"* "$DIST_DIR/resources/" 2>/dev/null || true
echo "[OK] Copied: $DIST_DIR/resources/"
echo ""

# Build React UI (if present and has package.json)
echo "Building React UI..."
if [ ! -f "$SCRIPT_DIR/pinshare-ui/package.json" ]; then
    echo "[SKIP] pinshare-ui/package.json not found - UI build skipped"
    echo ""
else
    pushd "$SCRIPT_DIR/pinshare-ui" > /dev/null

    if [ ! -d "node_modules" ]; then
        echo "Installing npm dependencies..."
        npm install
    fi

    npm run build

    # Copy UI files
    rm -rf "$DIST_DIR/ui"
    cp -r dist "$DIST_DIR/ui"
    popd > /dev/null
    echo "[OK] Built: $DIST_DIR/ui/"
    echo ""
fi

# Download IPFS if not present
IPFS_VERSION="v0.31.0"
if [ ! -f "$DIST_DIR/ipfs.exe" ]; then
    echo "Downloading IPFS Kubo $IPFS_VERSION..."
    TEMP_DIR=$(mktemp -d)
    curl -L "https://dist.ipfs.tech/kubo/${IPFS_VERSION}/kubo_${IPFS_VERSION}_windows-amd64.zip" -o "$TEMP_DIR/kubo.zip"
    unzip -q "$TEMP_DIR/kubo.zip" -d "$TEMP_DIR"
    cp "$TEMP_DIR/kubo/ipfs.exe" "$DIST_DIR/ipfs.exe"
    rm -rf "$TEMP_DIR"
    echo "[OK] Downloaded: $DIST_DIR/ipfs.exe"
else
    echo "[OK] IPFS already present: $DIST_DIR/ipfs.exe"
fi
echo ""

echo "=========================================="
echo "All Windows components built successfully!"
echo "=========================================="
echo ""
echo "Binaries:"
ls -1 "$DIST_DIR"/*.exe
echo ""

# Note about MSI installer
echo "Note: MSI installer must be built on Windows using WiX."
echo "To build the installer, copy dist/windows to a Windows machine and run:"
echo "  cd installer"
echo "  ./build-wix6.bat $VERSION"
echo ""
