@echo off
REM Build script for PinShare Windows Installer
REM Requires WiX Toolset 3.x or 4.x installed

setlocal

echo ===============================================
echo Building PinShare Windows Installer
echo ===============================================
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

REM Harvest UI files using heat.exe
echo Harvesting UI files...
heat.exe dir "..\dist\windows\ui" -cg UIComponents -dr UIFolder -gg -g1 -sf -srd -var var.UISourceDir -out UIComponents.wxs
if errorlevel 1 (
    echo ERROR: Failed to harvest UI files
    exit /b 1
)
echo UI files harvested successfully
echo.

REM Compile WiX sources
echo Compiling WiX sources...
candle.exe -ext WixUIExtension -ext WixUtilExtension -dUISourceDir="..\dist\windows\ui" Product.wxs UIComponents.wxs
if errorlevel 1 (
    echo ERROR: Failed to compile WiX sources
    exit /b 1
)
echo WiX sources compiled successfully
echo.

REM Link to create MSI
echo Linking MSI package...
light.exe -ext WixUIExtension -ext WixUtilExtension -out PinShare-Setup.msi Product.wixobj UIComponents.wixobj
if errorlevel 1 (
    echo ERROR: Failed to link MSI package
    exit /b 1
)
echo MSI package created successfully
echo.

REM Move to dist folder
if not exist "..\dist" mkdir "..\dist"
move /Y PinShare-Setup.msi ..\dist\
echo.

echo ===============================================
echo Build completed successfully!
echo ===============================================
echo Installer: ..\dist\PinShare-Setup.msi
echo.

REM Cleanup
del *.wixobj
del *.wixpdb
del UIComponents.wxs

endlocal
