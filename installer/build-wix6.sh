#!/bin/bash
# Build script for PinShare Windows Installer using WiX 6
# Requires: .NET SDK 6+ and WiX .NET tool

set -e

echo "==============================================="
echo "Building PinShare Windows Installer (WiX 6)"
echo "==============================================="
echo ""

# Check if .NET is installed
if ! command -v dotnet &> /dev/null; then
    echo "ERROR: .NET SDK not found"
    echo "Please install .NET SDK 6.0 or later from https://dotnet.microsoft.com/download"
    exit 1
fi

echo ".NET SDK found: $(dotnet --version)"

# Check if WiX tool is installed
if ! command -v wix &> /dev/null; then
    echo "WiX .NET tool not found. Installing..."
    dotnet tool install --global wix
fi

echo "WiX tool installed: $(wix --version)"
echo ""

# Check if dist directory exists
if [ ! -d "../dist/windows" ]; then
    echo "ERROR: Build directory ../dist/windows does not exist"
    echo "Please run the build process first to create binaries"
    exit 1
fi

# Check for required files
for file in pinsharesvc.exe pinshare.exe pinshare-tray.exe ipfs.exe; do
    if [ ! -f "../dist/windows/$file" ]; then
        echo "ERROR: $file not found in ../dist/windows"
        exit 1
    fi
done

if [ ! -f "../dist/windows/ui/index.html" ]; then
    echo "ERROR: UI files not found in ../dist/windows/ui"
    echo "Please build the React UI first"
    exit 1
fi

echo "All required files found!"
echo ""

# Build the MSI using dotnet build
echo "Building MSI package..."
dotnet build PinShare.wixproj -c Release

if [ $? -eq 0 ]; then
    echo ""
    echo "==============================================="
    echo "Build completed successfully!"
    echo "==============================================="
    echo "Installer: bin/Release/PinShare-Setup.msi"
    echo ""
else
    echo "ERROR: Failed to build MSI package"
    exit 1
fi
