@echo off
REM Build script for PinShare Windows Installer using WiX 6
REM Requires: .NET SDK 6+ and WiX .NET tool

setlocal

echo ===============================================
echo Building PinShare Windows Installer (WiX 6)
echo ===============================================
echo.

REM Check if .NET is installed
dotnet --version >nul 2>&1
if errorlevel 1 (
    echo ERROR: .NET SDK not found
    echo Please install .NET SDK 6.0 or later from https://dotnet.microsoft.com/download
    exit /b 1
)

REM Check if WiX tool is installed
wix --version >nul 2>&1
if errorlevel 1 (
    echo WiX .NET tool not found. Installing...
    dotnet tool install --global wix
    if errorlevel 1 (
        echo ERROR: Failed to install WiX tool
        exit /b 1
    )
)

echo WiX tool installed
echo.

REM Check if dist directory exists
if not exist "..\dist\windows" (
    echo ERROR: Build directory ..\dist\windows does not exist
    echo Please run the build process first to create binaries
    exit /b 1
)

REM Check for required files
if not exist "..\dist\windows\pinsharesvc.exe" (
    echo ERROR: pinsharesvc.exe not found in ..\dist\windows
    exit /b 1
)

if not exist "..\dist\windows\pinshare.exe" (
    echo ERROR: pinshare.exe not found in ..\dist\windows
    exit /b 1
)

if not exist "..\dist\windows\pinshare-tray.exe" (
    echo ERROR: pinshare-tray.exe not found in ..\dist\windows
    exit /b 1
)

if not exist "..\dist\windows\ipfs.exe" (
    echo ERROR: ipfs.exe not found in ..\dist\windows
    echo Please download IPFS Kubo from https://dist.ipfs.tech/kubo/
    exit /b 1
)

if not exist "..\dist\windows\ui\index.html" (
    echo ERROR: UI files not found in ..\dist\windows\ui
    echo Please build the React UI first
    exit /b 1
)

echo All required files found!
echo.

REM Build the MSI using dotnet build
echo Building MSI package...
dotnet build PinShare.wixproj -c Release
if errorlevel 1 (
    echo ERROR: Failed to build MSI package
    exit /b 1
)

echo.
echo ===============================================
echo Build completed successfully!
echo ===============================================
echo Installer: bin\Release\PinShare-Setup.msi
echo.

endlocal
