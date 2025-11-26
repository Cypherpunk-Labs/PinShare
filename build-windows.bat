@echo off
REM Build PinShare for Windows - Simple batch script
REM No make required!

setlocal enabledelayedexpansion

echo ==========================================
echo Building PinShare for Windows
echo ==========================================
echo.

REM Get the directory where this script is located
set SCRIPT_DIR=%~dp0
set DIST_DIR=%SCRIPT_DIR%dist\windows

REM Create dist directory
if not exist "%DIST_DIR%" mkdir "%DIST_DIR%"

REM Build PinShare backend
echo Building PinShare backend...
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64

go build -ldflags "-s -w" -o "%DIST_DIR%\pinshare.exe" "%SCRIPT_DIR%."
if errorlevel 1 (
    echo ERROR: Failed to build pinshare.exe
    exit /b 1
)
echo [OK] Built: %DIST_DIR%\pinshare.exe
echo.

REM Build Windows service wrapper
echo Building Windows service wrapper...
go build -ldflags "-s -w" -o "%DIST_DIR%\pinsharesvc.exe" "%SCRIPT_DIR%cmd\pinsharesvc"
if errorlevel 1 (
    echo ERROR: Failed to build pinsharesvc.exe
    exit /b 1
)
echo [OK] Built: %DIST_DIR%\pinsharesvc.exe
echo.

REM Build system tray application
echo Building system tray application...
go build -ldflags "-s -w -H windowsgui" -o "%DIST_DIR%\pinshare-tray.exe" "%SCRIPT_DIR%cmd\pinshare-tray"
if errorlevel 1 (
    echo ERROR: Failed to build pinshare-tray.exe
    exit /b 1
)
echo [OK] Built: %DIST_DIR%\pinshare-tray.exe
echo.

REM Build React UI
echo Building React UI...
pushd "%SCRIPT_DIR%pinshare-ui"
if errorlevel 1 (
    echo ERROR: pinshare-ui directory not found
    exit /b 1
)

if not exist "node_modules" (
    echo Installing npm dependencies...
    call npm install
    if errorlevel 1 (
        echo ERROR: Failed to install npm dependencies
        popd
        exit /b 1
    )
)

call npm run build
if errorlevel 1 (
    echo ERROR: Failed to build UI
    popd
    exit /b 1
)

REM Copy UI files
if exist "%DIST_DIR%\ui" rmdir /s /q "%DIST_DIR%\ui"
xcopy /E /I /Q dist "%DIST_DIR%\ui"
popd
echo [OK] Built: %DIST_DIR%\ui\
echo.

REM Download IPFS if not present
if not exist "%DIST_DIR%\ipfs.exe" (
    echo Downloading IPFS Kubo...
    powershell -Command "& {Invoke-WebRequest -Uri 'https://dist.ipfs.tech/kubo/v0.31.0/kubo_v0.31.0_windows-amd64.zip' -OutFile '%TEMP%\kubo.zip'; Expand-Archive -Path '%TEMP%\kubo.zip' -DestinationPath '%TEMP%' -Force; Copy-Item '%TEMP%\kubo\ipfs.exe' '%DIST_DIR%\ipfs.exe'; Remove-Item '%TEMP%\kubo.zip'; Remove-Item '%TEMP%\kubo' -Recurse}"
    echo [OK] Downloaded: %DIST_DIR%\ipfs.exe
) else (
    echo [OK] IPFS already present: %DIST_DIR%\ipfs.exe
)
echo.

echo ==========================================
echo All Windows components built successfully!
echo ==========================================
echo.
echo Binaries:
dir /b "%DIST_DIR%\*.exe"
echo.

REM Ask about building installer
echo.
echo Would you like to build the MSI installer now? (Y/N)
set /p BUILD_INSTALLER=
if /i "%BUILD_INSTALLER%"=="Y" (
    echo.
    echo Building MSI installer...
    pushd "%SCRIPT_DIR%installer"
    if errorlevel 1 (
        echo ERROR: Failed to change to installer directory at %SCRIPT_DIR%installer
        exit /b 1
    )

    call build-wix6.bat
    if errorlevel 1 (
        echo ERROR: Installer build failed
        popd
        exit /b 1
    )

    popd
    echo.
    echo ==========================================
    echo Build Complete!
    echo ==========================================
    echo.
    echo Installer: %SCRIPT_DIR%installer\bin\Release\PinShare-Setup.msi
    echo.
    echo To install, run:
    echo   msiexec /i "%SCRIPT_DIR%installer\bin\Release\PinShare-Setup.msi"
) else (
    echo.
    echo Skipping installer build. To build later, run:
    echo   cd "%SCRIPT_DIR%installer"
    echo   build-wix6.bat
)

endlocal
