#!/bin/bash
# Build script for PinShare Windows Installer using WiX 6
# Works on: Windows (Git Bash/WSL), macOS, Linux
#
# On Windows, this script delegates to build-wix6.bat for native builds.
# On macOS/Linux, WiX is not available - the script will provide instructions.
#
# Requires: .NET SDK 6+ and WiX .NET tool (Windows only)
# Usage: ./build-wix6.sh [version]
#   version: Optional version string (e.g., 1.2.3). Defaults to 1.0.0

set -e

# Get version from command line or use default
VERSION="${1:-1.0.0}"

echo "==============================================="
echo "Building PinShare Windows Installer (WiX 6)"
echo "Version: $VERSION"
echo "==============================================="
echo ""

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

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
echo ""

# On Windows, delegate to the batch file
if [ "$PLATFORM" = "windows" ]; then
    echo "Running native Windows build via build-wix6.bat..."
    echo ""
    # Use cmd.exe to run the batch file
    cmd.exe //c "$(cygpath -w "$SCRIPT_DIR/build-wix6.bat")" "$VERSION"
    exit $?
fi

# On macOS/Linux, WiX cannot run natively
echo "ERROR: WiX installer tools are only available on Windows."
echo ""
echo "The MSI installer must be built on a Windows machine."
echo ""
echo "Options:"
echo ""
echo "1. Copy the built binaries to Windows and build there:"
echo "   - Copy dist/windows/* to a Windows machine"
echo "   - Copy installer/* to the same machine"
echo "   - Run: ./build-wix6.bat $VERSION"
echo ""
echo "2. Use a Windows VM or CI/CD pipeline (e.g., GitHub Actions):"
echo "   - The GitHub Actions workflow builds the MSI on Windows runners"
echo ""
echo "3. Use Wine with .NET (experimental, not recommended):"
echo "   - Install Wine and .NET SDK under Wine"
echo "   - This is fragile and not officially supported"
echo ""

# Check if binaries exist
if [ -d "$SCRIPT_DIR/../dist/windows" ]; then
    echo "Built binaries found in dist/windows/:"
    ls -la "$SCRIPT_DIR/../dist/windows/"*.exe 2>/dev/null || echo "  (no .exe files found)"
    echo ""
else
    echo "No built binaries found. Run build-windows.sh first."
    echo ""
fi

exit 1
