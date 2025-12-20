@echo off
REM Test script to verify build fixes
setlocal enabledelayedexpansion

echo ==========================================
echo Testing PinShare Build Fixes
echo ==========================================
echo.

set SCRIPT_DIR=%~dp0
set DIST_DIR=%SCRIPT_DIR%dist\windows

echo [TEST 1] Checking .NET SDK installation and PATH
echo.

REM Try direct PATH first
dotnet --version >nul 2>&1
if errorlevel 1 (
    echo .NET SDK not in PATH, checking common locations...
    if exist "C:\Program Files\dotnet\dotnet.exe" (
        set "PATH=C:\Program Files\dotnet;%PATH%"
        echo [OK] Found in Program Files, adding to PATH
    ) else if exist "%USERPROFILE%\.dotnet\dotnet.exe" (
        set "PATH=%USERPROFILE%\.dotnet;%PATH%"
        echo [OK] Found in user profile, adding to PATH
    ) else (
        echo [FAIL] .NET SDK not found!
        goto :test_failed
    )
)

REM Try again
dotnet --version >nul 2>&1
if errorlevel 1 (
    echo [FAIL] .NET SDK still not accessible
    goto :test_failed
)

for /f "tokens=*" %%i in ('dotnet --version') do set DOTNET_VERSION=%%i
echo [OK] .NET SDK version: %DOTNET_VERSION%
echo.

echo [TEST 2] Checking WiX tool
echo.
wix --version >nul 2>&1
if errorlevel 1 (
    echo WiX tool not installed, attempting install...
    dotnet tool install --global wix
    if errorlevel 1 (
        echo [FAIL] Could not install WiX tool
        goto :test_failed
    )
    echo [OK] WiX tool installed
) else (
    for /f "tokens=*" %%i in ('wix --version') do set WIX_VERSION=%%i
    echo [OK] WiX version: %WIX_VERSION%
)
echo.

echo [TEST 3] Checking PowerShell download script
echo.
if not exist "%SCRIPT_DIR%installer\download-ipfs.ps1" (
    echo [FAIL] PowerShell script not found: %SCRIPT_DIR%installer\download-ipfs.ps1
    goto :test_failed
)
echo [OK] PowerShell script exists
echo.

echo [TEST 4] Testing IPFS download (will delete and re-download)
echo.

REM Clean up old IPFS
if exist "%DIST_DIR%\ipfs.exe" (
    echo Removing old ipfs.exe for clean test...
    del /q "%DIST_DIR%\ipfs.exe"
)

REM Test download
powershell -ExecutionPolicy Bypass -File "%SCRIPT_DIR%installer\download-ipfs.ps1" -DestDir "%DIST_DIR%" -Version "v0.31.0"
if errorlevel 1 (
    echo [FAIL] IPFS download failed!
    goto :test_failed
)

if not exist "%DIST_DIR%\ipfs.exe" (
    echo [FAIL] ipfs.exe not found after download
    goto :test_failed
)

echo [OK] IPFS downloaded successfully
echo.

echo [TEST 5] Verifying IPFS executable
echo.
for %%A in ("%DIST_DIR%\ipfs.exe") do set IPFS_SIZE=%%~zA
echo IPFS file size: %IPFS_SIZE% bytes

if %IPFS_SIZE% LSS 10000000 (
    echo [WARN] IPFS file seems small, expected ~17-36 MB
) else (
    echo [OK] IPFS file size looks good
)
echo.

echo ==========================================
echo All Tests Passed!
echo ==========================================
echo.
echo You can now run:
echo   .\build-windows.bat
echo.
echo The build should complete successfully including the MSI installer.
echo.
goto :end

:test_failed
echo.
echo ==========================================
echo Tests Failed!
echo ==========================================
echo.
echo Please review the errors above.
echo.
exit /b 1

:end
endlocal
